# Design Reddit

## 1. Problem Statement & Requirements

### Problem Statement
Design a social news aggregation platform like Reddit with subreddits, posts, nested comments (tree structure), voting, feed generation (hot/new/top/controversial), karma system, and moderation tools.

### Functional Requirements
- **Subreddits**: Create, browse, subscribe to communities
- **Posts**: Create post (link/text); upvote/downvote; sort by hot, new, top, controversial
- **Comments**: Tree structure (nested replies); upvote/downvote
- **Feed generation**: Home feed (subscribed subreddits); subreddit feed; ranking algorithms
- **Karma system**: User karma from post/comment votes
- **Moderation**: Mods can remove posts/comments; ban users; subreddit rules
- **Search**: Full-text search across posts

### Non-Functional Requirements
- **Scale**: 1.7B monthly visits, 100K+ subreddits
- **Latency**: Feed load < 200ms; vote count eventually consistent
- **Availability**: 99.9%
- **Vote counting**: High write volume; Redis counters + batch persist

### Out of Scope
- Reddit Premium / coins
- Live chat
- Video upload (focus on link/text)
- Recommendation algorithm (simplified)

---

## 2. Back-of-Envelope Estimation

### Scale Assumptions
- **Monthly visits**: 1.7B
- **DAU**: 50M
- **Subreddits**: 100K
- **Posts**: 1B total; 10M new/day
- **Comments**: 50B total; 500M new/day
- **Votes**: 10x comments = 5B/day
- **Peak**: 3x average

### QPS Calculation
| Operation | Daily Volume | Peak QPS |
|-----------|--------------|----------|
| Feed read | 500M | ~17,000 |
| Post read | 1B | ~35,000 |
| Vote (post/comment) | 5B | ~175,000 |
| Comment read | 2B | ~70,000 |
| Post create | 10M | ~350 |
| Comment create | 500M | ~17,000 |
| Search | 100M | ~3,500 |

### Storage (5 years)
- **Posts**: 18B × 1KB ≈ 18 TB
- **Comments**: 900B × 300B ≈ 270 TB (tree structure)
- **Votes**: 9T × 20B ≈ 180 TB (or aggregated only)
- **Subreddits**: 100K × 5KB ≈ 500 MB
- **Users**: 500M × 1KB ≈ 500 GB

### Bandwidth
- **Feed**: 17K × 50KB ≈ 850 MB/s
- **Votes**: 175K × 100B ≈ 17.5 MB/s write
- **Comments**: 70K × 20KB ≈ 1.4 GB/s

### Cache
- **Vote counts**: Redis; 1B posts × 8B ≈ 8 GB; 50B comments × 8B ≈ 400 GB (hot only)
- **Feed**: Pre-computed; cache top N per subreddit/sort
- **Hot posts**: Redis sorted set; score as key

---

## 3. API Design

### REST Endpoints

```
# Subreddits
GET    /api/v1/subreddits
GET    /api/v1/subreddits/:name
POST   /api/v1/subreddits/:name/subscribe
POST   /api/v1/subreddits/:name/unsubscribe

# Posts
GET    /api/v1/subreddits/:name/posts
Query: sort=hot|new|top|controversial|rising, limit=25, after=token
GET    /api/v1/posts/:id
POST   /api/v1/subreddits/:name/posts
Body: { "title": "...", "content": "...", "type": "text|link" }
PUT    /api/v1/posts/:id/vote
Body: { "direction": 1|-1|0 }   # upvote, downvote, unvote

# Comments
GET    /api/v1/posts/:id/comments
Query: sort=best|top|new|controversial, limit=100, depth=5
POST   /api/v1/posts/:id/comments
Body: { "content": "...", "parent_id": "..." }   # parent_id for replies
PUT    /api/v1/comments/:id/vote
Body: { "direction": 1|-1|0 }

# Feed
GET    /api/v1/feed
Query: sort=hot|new|top, limit=25, after=token
Response: Posts from subscribed subreddits

# User
GET    /api/v1/me
GET    /api/v1/me/karma
GET    /api/v1/users/:username

# Search
GET    /api/v1/search
Query: q=query, subreddit=name, sort=relevance|new
```

---

## 4. Data Model / Database Schema

### Database Choice
- **Posts**: PostgreSQL or Cassandra (partition by subreddit)
- **Comments**: PostgreSQL with adjacency list or closure table; or Cassandra
- **Votes**: Redis (counters) + Kafka → batch to DB
- **Subreddits, Users**: PostgreSQL
- **Feed**: Pre-computed; Redis sorted sets; Cassandra
- **Search**: Elasticsearch

### Schema

**Subreddits (PostgreSQL)**
```sql
subreddits (
  subreddit_id BIGSERIAL PRIMARY KEY,
  name VARCHAR(100) UNIQUE,
  description TEXT,
  subscriber_count BIGINT,
  created_at TIMESTAMP
)

subreddit_subscriptions (
  user_id BIGINT,
  subreddit_id BIGINT,
  subscribed_at TIMESTAMP,
  PRIMARY KEY (user_id, subreddit_id)
)
```

**Posts (Cassandra)**
```sql
posts (
  post_id UUID,
  subreddit_id BIGINT,
  user_id BIGINT,
  title VARCHAR(500),
  content TEXT,
  type VARCHAR(20),
  score INT,              -- Denormalized; updated async
  num_comments INT,
  created_at TIMESTAMP,
  PRIMARY KEY (subreddit_id, created_at)
) WITH CLUSTERING ORDER BY (created_at DESC);

-- Alternative: partition by (subreddit_id, sort_bucket) for hot/top
```

**Comments (PostgreSQL - Adjacency List)**
```sql
comments (
  comment_id BIGSERIAL PRIMARY KEY,
  post_id UUID,
  parent_id BIGINT,       -- NULL for top-level
  user_id BIGINT,
  content TEXT,
  score INT,
  depth INT,              -- 0 = top-level
  path LTREE,             -- Materialized path: 1.2.5.12
  created_at TIMESTAMP
)

CREATE INDEX idx_comments_post ON comments(post_id);
CREATE INDEX idx_comments_parent ON comments(parent_id);
CREATE INDEX idx_comments_path ON comments USING GIST(path);
```

**Comments (Cassandra - for scale)**
```sql
comments (
  comment_id UUID,
  post_id UUID,
  parent_id UUID,
  user_id BIGINT,
  content TEXT,
  score INT,
  created_at TIMESTAMP,
  PRIMARY KEY (post_id, comment_id)
) WITH CLUSTERING ORDER BY (comment_id ASC);
-- Thread: Query by post_id; build tree in application
```

**Votes (Redis + Kafka)**
```
Key: vote:post:{post_id}
Value: score (incr/decr on vote)
Key: vote:comment:{comment_id}
Value: score
```

**Vote events (Kafka)**
```
Topic: votes
Message: { post_id, comment_id, user_id, direction, timestamp }
Consumer: Batch aggregate; update PostgreSQL/Cassandra
```

**Users (PostgreSQL)**
```sql
users (
  user_id BIGSERIAL PRIMARY KEY,
  username VARCHAR(50) UNIQUE,
  karma BIGINT,
  created_at TIMESTAMP
)
```

---

## 5. High-Level Architecture

### ASCII Architecture Diagram

```
                                    ┌─────────────────────────────────────────────────────────────┐
                                    │                    CLIENTS (Web, Mobile)                     │
                                    └───────────────────────────┬─────────────────────────────────┘
                                                                  │
                                                                  ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                              API GATEWAY                                                       │
└───────────────────────────┬──────────────────────────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┬───────────────────┬───────────────────┐
        ▼                   ▼                   ▼                   ▼                   ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│ Feed          │   │ Post          │   │ Comment       │   │ Vote          │   │ Search        │
│ Service       │   │ Service       │   │ Service       │   │ Service       │   │ Service       │
│               │   │               │   │               │   │               │   │               │
│ - Rank        │   │ - CRUD        │   │ - Tree        │   │ - Redis       │   │ - Elasticsearch│
│ - Merge       │   │ - Score       │   │ - Nesting     │   │   INCR/DECR   │   │               │
└───────┬───────┘   └───────┬───────┘   └───────┬───────┘   └───────┬───────┘   └───────┬───────┘
        │                   │                   │                   │                   │
        │                   │                   │                   │                   │
        ▼                   ▼                   ▼                   ▼                   ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│ Redis         │   │ Cassandra     │   │ PostgreSQL    │   │ Redis         │   │ Elasticsearch │
│ (Feed cache)  │   │ (Posts)       │   │ (Comments     │   │ (Vote counts) │   │ (Search)      │
│ Sorted sets   │   │               │   │  tree)        │   │               │   │               │
└───────────────┘   └───────────────┘   └───────────────┘   └───────┬───────┘   └───────────────┘
        │                   │                   │                   │
        │                   │                   │                   ▼
        │                   │                   │           ┌───────────────┐
        │                   │                   │           │ Kafka         │
        │                   │                   │           │ (Vote events) │
        │                   │                   │           └───────┬───────┘
        │                   │                   │                   │
        │                   │                   │                   ▼
        │                   │                   │           ┌───────────────┐
        │                   │                   │           │ Vote          │
        │                   │                   │           │ Aggregator    │
        │                   │                   │           │ (Batch write) │
        │                   │                   │           └───────┬───────┘
        │                   │                   │                   │
        │                   │                   └───────────────────┘
        │                   │                             (Update comment score)
        │                   └─────────────────────────────────────────────────┐
        │                                                                     │
        └─────────────────────────────────────────────────────────────────────┘
```

---

## 6. Detailed Component Design

### 6.1 Feed Generation & Ranking Algorithms

**Hot (Reddit's hot algorithm)**
- Formula: `score = (log10(|votes|) + sign(votes) * (seconds_since_post / 45000))`
- Or Wilson score for confidence
- Pre-compute: Cron every 5 min; update Redis sorted set
- Key: `hot:subreddit:{id}` or `hot:feed:{user_id}`

**New**
- Sort by created_at DESC
- Simple; no pre-computation

**Top**
- Sort by score DESC; time filter (hour, day, week, month, all)
- Pre-compute for each time window

**Controversial**
- Formula: `(upvotes + downvotes) / max(|upvotes - downvotes|, 1)`
- High engagement, split opinion

**Rising**
- Posts with recent vote velocity; trending

### 6.2 Vote Counting at Scale (Redis Counters)

- **Write**: On vote, `INCRBY vote:post:{id} 1` or `-1`; publish to Kafka
- **Read**: `GET vote:post:{id}` from Redis; fallback to DB if miss
- **Persistence**: Kafka consumer batches votes; updates PostgreSQL/Cassandra
- **Recovery**: Rebuild Redis from DB if Redis loses data; or replay Kafka

### 6.3 Comment Tree Structure

**Adjacency List**
- Each comment has parent_id
- Query: Get all for post_id; build tree in app (O(n))
- Depth limit: Load top 5 levels; "load more" for deeper

**Materialized Path (LTREE)**
- path = "1.2.5.12" for comment 12 under 5 under 2 under 1
- Query subtree: `WHERE path <@ '1.2'`
- Efficient for "load more" on branch

**Nested Sets**
- left, right values; good for reads; complex updates

**Reddit approach**: Flattened with parent_id; client builds tree; "continue thread" loads more

### 6.4 Wilson Score (Confidence Sorting)

- **Problem**: New post with 1 upvote ranks above old post with 100 upvotes
- **Wilson score**: Lower bound of binomial proportion confidence interval
- **Formula**: `(p + z²/2n - z*sqrt((p(1-p)+z²/4n)/n)) / (1+z²/n)`
- **Effect**: Fewer votes = lower confidence = lower score
- **Use**: "Best" comment sort

### 6.5 Karma System

- **Post karma**: Sum of post scores (capped per post)
- **Comment karma**: Sum of comment scores
- **Update**: Async from vote aggregator; batch update user karma
- **Storage**: users.karma; denormalized

### 6.6 Moderation Tools

- **Remove**: Soft delete; flag removed=true
- **Ban**: user_bans table; check on post/comment
- **Rules**: subreddit_rules table; display on submit
- **Mod queue**: Posts/comments flagged; mod reviews
- **Auto-mod**: Rules-based (keyword, account age); auto-remove

---

## 7. Scaling

### Vote Volume
- **Redis**: 175K QPS; Redis cluster; shard by post_id/comment_id
- **Kafka**: Buffer writes; batch consumer
- **DB**: Batch updates; avoid 175K writes/s to PostgreSQL

### Feed
- **Pre-compute**: Cron; populate Redis sorted sets
- **Per-user feed**: Merge subscribed subreddits' hot lists; cache
- **Sharding**: By subreddit_id

### Comments
- **Pagination**: Load top-level first; "load more" for replies
- **Depth limit**: Don't load 1000 levels; cap at 10
- **Cassandra**: Partition by post_id; wide partition for popular posts (millions of comments)

### Search
- **Elasticsearch**: Index posts; shard by subreddit
- **Sync**: Kafka from post/comment writes

---

## 8. Failure Handling

### Redis Vote Count Down
- Fallback: Read score from PostgreSQL (may be stale)
- Rebuild: Replay Kafka or scan DB

### Feed Pre-compute Failure
- Serve stale cached feed
- Fallback: Real-time merge (slower)

### Comment Tree Load
- Timeout: Return partial tree
- Circuit breaker: If DB slow, return cached

### Vote Duplicate
- Idempotency: user_id + post_id + direction; reject duplicate
- Or: Toggle (upvote → downvote = change by 2)

---

## 9. Monitoring & Observability

### Key Metrics
- **Feed latency**: p50, p99
- **Vote latency**: Redis INCR time
- **Comment load**: Tree build time
- **Ranking**: Pre-compute job duration
- **Vote lag**: Kafka consumer lag

### Alerts
- Vote Redis error rate > 0.1%
- Feed p99 > 500ms
- Kafka consumer lag > 1M
- Pre-compute job failure

### Tracing
- Trace: Vote → Redis → Kafka → Aggregator → DB
- Feed: Merge → Cache → Response

---

## 10. Interview Tips

### Follow-up Questions
- "How would you change the hot algorithm to reduce gaming?"
- "How do you handle a post with 1M comments?"
- "How would you add real-time vote updates?"
- "How do you prevent vote manipulation (bots)?"
- "How would you implement 'best' comment sort?"

### Common Mistakes
- **Synchronous vote to DB**: 175K writes/s will kill DB; use Redis + async
- **No ranking algorithm**: Hot/new/top are core; must explain
- **Flat comments**: Reddit has trees; adjacency list or path
- **Single feed**: Per-user feed = merge of subreddits; pre-compute

### Key Points to Emphasize
- **Vote counting**: Redis INCR + Kafka + batch persist
- **Ranking**: Hot algorithm, Wilson score, pre-computation
- **Comment tree**: Adjacency list, materialized path, depth limit
- **Scale**: 1.7B visits, 100K subreddits, 5B votes/day
- **Feed**: Pre-computed sorted sets; merge for user feed

---

## Appendix: Extended Design Details & Walkthrough Scenarios

### A. Reddit Hot Algorithm (Simplified)

```
score = log10(max(|votes|, 1)) + sign(votes) * (seconds_since_epoch - 1134028003) / 45000
```

- **log10**: Diminishing returns for upvotes
- **sign * time**: Newer = boost; older = decay
- **45000**: ~12.5 hours half-life
- **1134028003**: Reddit's epoch (Dec 8, 2005)

### B. Wilson Score Formula

For upvotes U, downvotes D, n = U+D, p = U/n:

```
z = 1.96  (95% confidence)
score = (p + z²/2n - z*sqrt((p(1-p)+z²/4n)/n)) / (1+z²/n)
```

- 1 up, 0 down: score ~0.21
- 100 up, 0 down: score ~0.96
- 10 up, 10 down: score ~0.34 (controversial)

### C. Vote Flow Walkthrough

1. User upvotes post P123
2. Vote Service: Check idempotency (user U1, post P123)
3. Redis: `INCRBY vote:post:P123 1`
4. Kafka: Produce `{ post_id: P123, user_id: U1, delta: 1 }`
5. Response: Return new score from Redis
6. Consumer: Every 5s, batch all votes for P123; UPDATE posts SET score = score + batch_sum
7. Invalidate feed cache for subreddit

### D. Comment Tree Query (Adjacency List)

```sql
-- Get all comments for post
SELECT * FROM comments WHERE post_id = ? ORDER BY created_at ASC;

-- Build tree in app:
-- Map comment_id -> comment
-- For each, parent_id -> append to parent's children
-- Root: parent_id IS NULL
```

### E. Feed Merge (User Subscribed)

- User subscribes to S1, S2, S3
- Each has Redis sorted set: `hot:subreddit:S1` (score -> post_id)
- Merge: ZUNIONSTORE or merge in app (merge k sorted lists)
- Cache: `feed:user:U1` with TTL 5 min
- Pagination: ZRANGE with cursor

### F. Pre-compute Job (Hot)

1. For each subreddit: Get posts from last 7 days
2. Compute hot score for each
3. ZADD to Redis: `hot:subreddit:{id}` (score, post_id)
4. Trim to top 1000
5. Run every 5 minutes

### G. Controversial Formula

```
controversial_score = (upvotes + downvotes) / max(|upvotes - downvotes|, 1)
```

- 50 up, 50 down: 100/1 = 100 (very controversial)
- 100 up, 0 down: 100/100 = 1 (not controversial)

### H. Search Index (Elasticsearch)

```json
{
  "mappings": {
    "properties": {
      "post_id": { "type": "keyword" },
      "subreddit_id": { "type": "keyword" },
      "title": { "type": "text" },
      "content": { "type": "text" },
      "score": { "type": "integer" },
      "created_at": { "type": "date" }
    }
  }
}
```

### I. Post with 1M Comments

- **Load**: Don't load all; paginate top-level
- **Sort**: Best/top; load top 100 top-level
- **Replies**: "Load more" on demand; lazy load
- **Storage**: Cassandra partition by post_id; wide partition
- **Cache**: Top comments cached; long tail from DB

### J. Karma Update

- On vote batch: Update post/comment score
- Trigger: Update user karma (post owner, comment owner)
- Batch: Group by user_id; single UPDATE per user
- Or: Separate Kafka topic for karma; consumer updates users table
