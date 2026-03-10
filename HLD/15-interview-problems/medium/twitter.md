# Design Twitter

## 1. Problem Statement & Requirements

### Problem Statement
Design a system like Twitter that allows users to post short messages (tweets), follow other users, view a personalized home timeline (news feed), search tweets, and see trending topics.

### Functional Requirements
- **Post tweets**: Users can post tweets (max 280 characters)
- **Follow users**: Users can follow/unfollow other users
- **Home timeline**: Users see a reverse-chronological feed of tweets from people they follow
- **Search**: Full-text search over tweets
- **Trending topics**: Display trending hashtags and topics in real-time
- **User profile**: View user profile and their tweets
- **Engagement**: Like, retweet, reply to tweets
- **Media**: Support images and videos in tweets

### Non-Functional Requirements
- **Latency**: Home timeline load < 200ms (p99)
- **Availability**: 99.99% uptime
- **Consistency**: Eventual consistency acceptable for timeline
- **Durability**: No tweet loss

### Out of Scope
- Direct messaging (DM)
- Video live streaming
- Tweet editing (assume immutable)
- Advanced analytics dashboard

---

## 2. Back-of-Envelope Estimation

### Scale Assumptions
- **DAU**: 500 million
- **Tweets/day**: 500 million
- **Avg followers per user**: 200
- **Avg following per user**: 300
- **Timeline reads**: 10x tweet writes = 5B reads/day

### QPS Calculation
| Operation | Daily Volume | QPS |
|-----------|--------------|-----|
| Tweet writes | 500M | ~5,800 |
| Timeline reads | 5B | ~58,000 |
| Search queries | 500M | ~5,800 |
| Follow operations | 100M | ~1,200 |

### Storage (5 years)
- **Tweets**: 500M × 280 bytes × 365 × 5 ≈ 255 TB (text only)
- **With metadata**: ~500 TB
- **Media**: Assume 20% tweets have media, avg 500KB → 45 PB
- **Social graph**: 500M × 300 × 8 bytes × 2 ≈ 2.4 TB

### Bandwidth
- **Timeline reads**: 58K QPS × 10 KB/response ≈ 580 MB/s
- **Media**: 10M media views/day × 500KB ≈ 5 TB/day

### Cache
- **Hot timeline cache**: 20% users active → 100M × 500 tweets × 1KB ≈ 50 TB (use LRU, cache top 20%)
- **User graph cache**: 500M × 300 × 50 bytes ≈ 7.5 TB

---

## 3. API Design

### REST Endpoints

```
POST   /api/v1/tweets
Body: { "content": "Hello world", "media_ids": ["m1", "m2"] }
Response: { "tweet_id": "t123", "created_at": "..." }

GET    /api/v1/tweets/:id
Response: { "tweet_id", "user_id", "content", "created_at", "likes", "retweets" }

DELETE /api/v1/tweets/:id

GET    /api/v1/users/:id/tweets?cursor=&limit=20
Response: { "tweets": [...], "next_cursor": "..." }

GET    /api/v1/timeline?cursor=&limit=20
Response: { "tweets": [...], "next_cursor": "..." }

POST   /api/v1/users/:id/follow
POST   /api/v1/users/:id/unfollow

GET    /api/v1/users/:id/followers?cursor=&limit=20
GET    /api/v1/users/:id/following?cursor=&limit=20

POST   /api/v1/tweets/:id/like
POST   /api/v1/tweets/:id/unlike
POST   /api/v1/tweets/:id/retweet

GET    /api/v1/search?q=hello&limit=20&cursor=
Response: { "tweets": [...], "next_cursor": "..." }

GET    /api/v1/trending?location=US
Response: { "trends": [{"hashtag": "#AI", "count": 1234567}, ...] }

POST   /api/v1/media/upload
Body: multipart/form-data
Response: { "media_id": "m123" }
```

---

## 4. Data Model / Database Schema

### Database Choice
- **Tweets**: Cassandra/DynamoDB (write-heavy, time-series)
- **Social graph**: Neo4j or Cassandra (graph queries)
- **User data**: MySQL/PostgreSQL (relational, ACID)
- **Search**: Elasticsearch (inverted index)
- **Trending**: Redis (real-time counters)

### Schema

**Users Table (MySQL)**
```sql
users (
  user_id BIGINT PRIMARY KEY,
  username VARCHAR(50) UNIQUE,
  display_name VARCHAR(100),
  bio TEXT,
  profile_image_url VARCHAR(500),
  created_at TIMESTAMP,
  updated_at TIMESTAMP
)
```

**Tweets Table (Cassandra)**
```sql
-- Partition by user_id for user timeline
tweets (
  user_id BIGINT,
  tweet_id TIMEUUID,
  content TEXT,
  media_ids LIST<UUID>,
  created_at TIMESTAMP,
  likes_count COUNTER,
  retweets_count COUNTER,
  reply_to_tweet_id TIMEUUID,
  PRIMARY KEY (user_id, tweet_id)
) WITH CLUSTERING ORDER BY (tweet_id DESC);

-- Global tweet lookup
tweets_by_id (
  tweet_id TIMEUUID PRIMARY KEY,
  user_id BIGINT,
  content TEXT,
  ... 
)
```

**Follow Graph (Cassandra)**
```sql
followers (
  user_id BIGINT,        -- user being followed
  follower_id BIGINT,    -- user who follows
  created_at TIMESTAMP,
  PRIMARY KEY (user_id, follower_id)
)

following (
  user_id BIGINT,
  following_id BIGINT,
  created_at TIMESTAMP,
  PRIMARY KEY (user_id, following_id)
)
```

**Timeline Cache (Redis)**
```
timeline:{user_id} -> Sorted Set (score=timestamp, member=tweet_id)
```

---

## 5. High-Level Architecture

```
                                    ┌─────────────────────────────────────────────────────────────┐
                                    │                      CLIENT (Web/Mobile)                     │
                                    └─────────────────────────────────────────────────────────────┘
                                                              │
                                                              ▼
┌─────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                          API GATEWAY / LOAD BALANCER                                          │
└─────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
                                                              │
                    ┌─────────────────────────────────────────┼─────────────────────────────────────────┐
                    ▼                                         ▼                                         ▼
         ┌──────────────────┐                      ┌──────────────────┐                      ┌──────────────────┐
         │   Tweet Service  │                      │   User Service   │                      │  Timeline Service│
         │  ─────────────── │                      │  ──────────────  │                      │  ─────────────── │
         │  • Post tweet    │                      │  • Follow/Unfollow│                     │  • Get home feed │
         │  • Like/Retweet  │                      │  • Followers list│                     │  • Merge feeds   │
         │  • Media upload  │                      │  • User profile  │                     │  • Fan-out logic │
         └────────┬─────────┘                      └────────┬─────────┘                      └────────┬─────────┘
                  │                                         │                                         │
                  │         ┌───────────────────────────────┼───────────────────────────────┐          │
                  │         │                               │                               │          │
                  ▼         ▼                               ▼                               ▼          ▼
         ┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
         │                              MESSAGE QUEUE (Kafka)                                                    │
         │  Topics: tweet-events, follow-events, timeline-events                                                │
         └─────────────────────────────────────────────────────────────────────────────────────────────────────┘
                  │                                         │                                         │
    ┌─────────────┼─────────────┐              ┌────────────┼────────────┐              ┌────────────┼────────────┐
    ▼             ▼             ▼              ▼            ▼            ▼              ▼            ▼            ▼
┌────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
│Tweet   │  │ Social   │  │  Media   │  │ Timeline │  │  Search  │  │ Trending │  │  Cache   │  │  CDN     │  │  Push    │
│Store   │  │ Graph    │  │  Store   │  │ Generator│  │  Index   │  │  Service │  │  (Redis) │  │          │  │  Service │
│(Cass)  │  │ (Cass)   │  │ (S3/CDN) │  │ (Workers)│  │(Elastic) │  │(CountMin)│  │          │  │          │  │          │
└────────┘  └──────────┘  └──────────┘  └──────────┘  └──────────┘  └──────────┘  └──────────┘  └──────────┘  └──────────┘
```

### Timeline Generation Flow (ASCII Architecture)

```
                    USER POSTS TWEET
                           │
                           ▼
              ┌────────────────────────┐
              │   Tweet Service        │
              │   • Validate tweet     │
              │   • Store in DB        │
              │   • Publish to Kafka   │
              └────────────┬───────────┘
                           │
                           ▼
              ┌────────────────────────┐
              │   Fan-out Checker      │
              │   • Get follower count │
              └────────────┬───────────┘
                           │
              ┌────────────┴────────────┐
              │                         │
         < 100K followers          > 100K followers
         (Normal user)             (Celebrity)
              │                         │
              ▼                         ▼
    ┌─────────────────┐       ┌─────────────────┐
    │ FAN-OUT ON WRITE│       │ FAN-OUT ON READ │
    │  ─────────────  │       │  ─────────────  │
    │  • Fetch all    │       │  • Store tweet  │
    │    follower IDs │       │    in DB only   │
    │  • Push to each │       │  • Merge at     │
    │    follower's   │       │    read time    │
    │    timeline     │       │  • Query celeb  │
    │  • Update Redis │       │    tweets when  │
    │    sorted sets  │       │    loading feed │
    └─────────────────┘       └─────────────────┘
              │                         │
              └────────────┬────────────┘
                           │
                           ▼
              ┌────────────────────────┐
              │   Timeline Merge       │
              │   • Pre-computed (norm) │
              │   • Merge celeb tweets  │
              │   • Sort by recency    │
              │   • Return top 20       │
              └────────────────────────┘
```

---

## 6. Detailed Component Design

### 6.1 Tweet Service
- **Responsibilities**: Accept tweet creation, validate (280 chars), store in Cassandra
- **Flow**: 
  1. Validate content (length, sanitize)
  2. Generate tweet_id (UUID)
  3. Write to `tweets` partition by user_id
  4. Write to `tweets_by_id` for global lookup
  5. Publish tweet event to Kafka
- **Media**: Upload to S3, store metadata, return media_ids

### 6.2 Social Graph Service
- **Followers/Following**: Stored in Cassandra with bidirectional indexes
- **Fan-out list**: For normal users, fetch follower list for fan-out
- **Celebrity detection**: Users with >100K followers use read-time merge
- **Caching**: Cache follower list in Redis (invalidate on follow/unfollow)

### 6.3 Timeline Generation (Hybrid Approach)
**Fan-out on Write (Normal users, <100K followers)**:
- When user A tweets, fetch A's followers
- For each follower F, push tweet_id to `timeline:F` in Redis (sorted set)
- Trim to last 1000 tweets per user
- Pros: Fast reads (O(1) from cache)
- Cons: Slow writes for users with many followers

**Fan-out on Read (Celebrities, >100K followers)**:
- When celebrity C tweets, only store in tweet DB
- At read time: fetch pre-computed timeline (from normal users) + fetch tweets from followed celebrities
- Merge and sort by timestamp
- Pros: Fast writes
- Cons: Slower reads (extra merge step)

**Implementation**: 
- Fan-out worker consumes Kafka, checks follower count
- If < 100K: batch push to Redis (parallel, 1000 at a time)
- If >= 100K: add to "celebrity_tweets" index for merge

### 6.4 Search Service
- **Index**: Elasticsearch inverted index on tweets
- **Fields**: content, hashtags, mentions, user_id
- **Pipeline**: Kafka consumer indexes new tweets
- **Ranking**: Recency + engagement (likes, retweets)

### 6.5 Trending Topics
- **Count-Min Sketch**: Probabilistic data structure for streaming counts
- **Flow**: Extract hashtags from tweets → increment sketch buckets
- **Time windows**: 1h, 24h trends
- **Storage**: Redis sorted sets for top-K

### 6.6 Media Storage
- **Upload**: Client → API → S3
- **CDN**: CloudFront/CloudFront for global distribution
- **Resolutions**: Original + thumbnail (for timeline preview)

### 6.7 Push Notifications
- **Trigger**: New tweet from followed user
- **Service**: FCM/APNs integration
- **Queue**: Separate queue for push, deduplication by (user, tweet)

---

## 7. Scaling

### Sharding
- **Tweets**: Shard by user_id (hash)
- **Social graph**: Shard by user_id
- **Timeline cache**: Shard Redis by user_id

### Caching
- **Timeline**: Redis sorted sets, LRU eviction
- **User profile**: Redis, 5 min TTL
- **Social graph**: Cache follower list for fan-out

### CDN
- **Media**: All images/videos on CDN
- **Edge caching**: Cache static assets

### Read Replicas
- MySQL read replicas for user data
- Cassandra read replicas

---

## 8. Failure Handling

### Component Failures
- **Tweet Service down**: Queue tweets, process when back
- **Timeline cache miss**: Rebuild from DB (slower)
- **Kafka partition**: Rebalance consumers
- **Cassandra node**: Replication factor 3, quorum reads

### Redundancy
- Multi-AZ deployment
- Database replication across regions
- Cache cluster (Redis Cluster)

### Degradation
- **Heavy load**: Disable fan-out for non-essential users
- **Search down**: Return cached results or error gracefully

---

## 9. Monitoring & Observability

### Key Metrics
- **Latency**: p50, p99, p999 for timeline load

- **Throughput**: Tweets/sec, timeline reads/sec
- **Error rate**: 4xx, 5xx by endpoint
- **Cache hit rate**: Redis timeline hit rate
- **Kafka lag**: Consumer lag per partition
- **Fan-out latency**: Time from tweet to timeline update

### Alerts
- Timeline p99 > 500ms
- Kafka consumer lag > 1000
- Cache hit rate < 80%
- Error rate > 1%

### Tracing
- Distributed tracing (Jaeger) for request flow
- Trace tweet creation → fan-out → cache update

---

## 10. Interview Tips

### Follow-up Questions
1. **How do you handle celebrity tweets with 50M followers?** Fan-out on read, merge at query time
2. **What if a user follows 10,000 people?** Pagination, limit timeline merge to top 500 by recency
3. **How to handle tweet deletion?** Soft delete, propagate to timeline caches, remove from search index
4. **How to rank timeline?** Add recency + engagement score (ML model)
5. **Retweet storage?** Store as new tweet with reference to original, or store retweet metadata

### Common Mistakes
- **Fan-out only on write**: Doesn't scale for celebrities
- **Fan-out only on read**: Too slow for normal users
- **No caching**: Can't handle 58K reads/sec
- **Synchronous fan-out**: Blocks tweet creation
- **Single DB**: Cassandra for tweets, not MySQL (write-heavy)

### Key Points to Emphasize
1. **Hybrid fan-out** is the core design decision
2. **Kafka** decouples write path from fan-out
3. **Redis** for hot timeline cache
4. **Sharding** by user_id for scalability
5. **Eventual consistency** for timeline is acceptable

---

## Appendix A: Alternative Design Considerations

### A.1 Pull vs Push for Timeline
- **Pull (fan-out on read)**: Simpler, but N+1 queries for N followed users
- **Push (fan-out on write)**: Fast reads, expensive writes for celebrities
- **Hybrid**: Best of both - threshold at 100K followers

### A.2 Timeline Ranking Algorithms
- **Chronological**: Simple, fair, but may bury engagement
- **Engagement-weighted**: Score = recency × (1 + 0.1×likes + 0.2×retweets)
- **ML ranking**: Train model on implicit feedback (scroll, dwell, like)
- **Blend**: 70% chronological, 30% ranked for "Top" tab

### A.3 Retweet Storage Options
| Approach | Pros | Cons |
|----------|------|------|
| New tweet with ref | Simple, searchable | Duplicate content |
| Retweet table | No duplication | Extra join |
| Embedded in timeline | Fast | Hard to "un-retweet" |

### A.4 Count-Min Sketch for Trending
- **Width**: 2^20 buckets
- **Depth**: 5 hash functions
- **Error**: Overcount by ε with probability δ
- **Merge**: Add buckets for time window aggregation

### A.5 Sample Kafka Message Schema
```json
{
  "event_type": "tweet_created",
  "tweet_id": "uuid",
  "user_id": 123,
  "content": "...",
  "created_at": "ISO8601",
  "follower_count": 50000
}
```

### A.6 Redis Timeline Data Structure
```
ZADD timeline:user_456 1734567890 tweet_uuid_1
ZADD timeline:user_456 1734567891 tweet_uuid_2
ZRANGE timeline:user_456 0 19 REV  -- Get latest 20
ZREMRANGEBYRANK timeline:user_456 0 -1001  -- Keep last 1000
```

### A.7 Capacity Planning Summary
- **Tweet writes**: 5.8K QPS → 50 servers (100 QPS each)
- **Timeline reads**: 58K QPS → 500 servers + Redis cluster
- **Fan-out workers**: 100 workers for 5.8K tweets × 200 avg followers

---

## Appendix B: Deep Dive - Fan-out Worker Implementation

### B.1 Batch Processing Logic
```
For each tweet event from Kafka:
  1. Get follower_count from social graph cache
  2. If follower_count < 100000:
       followers = fetch_followers(user_id)  # Paginated, 1000 per batch
       for batch in followers:
           redis.pipeline()
           for f in batch:
               ZADD timeline:{f} timestamp tweet_id
           redis.execute()
  3. Else:
       INSERT INTO celebrity_tweets_index (user_id, tweet_id, timestamp)
```

### B.2 Timeline Read Path (with Celebrity Merge)
```
1. Get pre_computed = ZRANGE timeline:{user_id} 0 19 REV  # From Redis
2. celebrity_ids = get_followed_celebrities(user_id)
3. celebrity_tweets = query celebrity_tweets_index WHERE user_id IN celebrity_ids LIMIT 50
4. merged = merge_and_sort(pre_computed, celebrity_tweets)
5. top_20 = merged[:20]
6. Fetch tweet content for top_20 from cache/DB
7. Return
```

### B.3 Tweet Deletion Propagation
- Soft delete: Set deleted=true in tweets table
- Timeline: Remove from Redis ZREM timeline:{user_id} tweet_id
- Search: Delete from Elasticsearch index
- Async: Fan-out deletion job processes Kafka delete events

### B.4 Hashtag Extraction Pipeline
- Regex: #(\w+) from tweet content
- Normalize: lowercase, trim
- Store: tweet_hashtags (tweet_id, hashtag)
- Trending: Increment Count-Min Sketch per hashtag
- Top-K: Query sketch, sort by count, return top 10

---

## Appendix C: Walkthrough Scenarios

### C.1 Scenario: User with 50K Followers Posts Tweet
1. User submits tweet via POST /api/v1/tweets
2. Tweet Service validates (280 chars), generates UUID
3. Writes to Cassandra tweets table (user_id partition)
4. Writes to tweets_by_id for global lookup
5. Publishes to Kafka topic tweet-events
6. Fan-out worker consumes event
7. Queries follower count: 50,000 (< 100K threshold)
8. Fetches follower IDs in batches of 1000 (50 batches)
9. For each batch: Redis pipeline ZADD to 1000 timelines
10. Total time: ~2-5 seconds for full fan-out
11. Each follower's next timeline read gets the tweet from Redis

### C.2 Scenario: Celebrity (10M Followers) Posts Tweet
1. Same steps 1-6 as above
2. Fan-out worker queries follower count: 10,000,000
3. Exceeds threshold -> Skip fan-out on write
4. Insert into celebrity_tweets_index (user_id, tweet_id, timestamp)
5. Total time: < 100ms
6. When follower loads timeline: Fetch pre-computed (normal users) + query celebrity_tweets for followed celebs, merge, return

### C.3 Scenario: Timeline Read with Cache Miss
1. User requests GET /api/v1/timeline
2. Timeline Service checks Redis timeline:{user_id}
3. Cache miss (new user or evicted)
4. Fetch followed user IDs from social graph
5. For each followed: Fetch latest 20 tweets from tweets table
6. Merge all, sort by timestamp DESC
7. Take top 20, return to client
8. Backfill Redis cache for next time (async)
9. Latency: 200-500ms (slower path)

### C.4 Scenario: Search for "machine learning"
1. User submits GET /api/v1/search?q=machine+learning
2. Search Service queries Elasticsearch
3. Query: match on content, hashtags, with "machine learning"
4. Ranking: BM25 + boost(recency) + boost(engagement)
5. Return top 20, cursor for pagination
6. Latency: 50-150ms

---

## Appendix D: Technology Choices Rationale

### D.1 Why Cassandra for Tweets?
- **Write-heavy**: 5.8K writes/sec, Cassandra excels at writes
- **Time-series**: Tweets are append-only, partition by user_id
- **No joins**: Timeline built from cache/merge, not SQL joins
- **Horizontal scale**: Add nodes, linear scalability
- **Tunable consistency**: QUORUM for durability, ONE for reads

### D.2 Why Redis for Timeline?
- **Sorted sets**: Perfect for timeline (score=timestamp)
- **O(log N) insert**: Fast fan-out
- **O(log N + M) range**: Fast retrieval of top 20
- **In-memory**: Sub-millisecond latency
- **Persistence**: Optional AOF for durability

### D.3 Why Kafka for Fan-out?
- **Decoupling**: Tweet creation doesn't wait for fan-out
- **Reliability**: At-least-once, replay on failure
- **Ordering**: Per-partition ordering for tweet events
- **Backpressure**: Queue absorbs spikes
- **Replay**: Can rebuild timelines from event log

### D.4 Why Elasticsearch for Search?
- **Full-text**: Inverted index, relevance ranking
- **Scale**: Distributed, handles 5.8K QPS
- **Flexible**: Add fields without schema change
- **Real-time**: Near real-time indexing from Kafka

### D.5 Interview Discussion Points
- **Trade-offs**: Discuss why you chose each technology
- **Alternatives**: "We could use DynamoDB for tweets - similar write scaling"
- **Evolution**: "Start with MySQL, migrate to Cassandra at 10M tweets/day"

---

## Quick Reference Summary

| Component | Technology | Key Reason |
|-----------|------------|------------|
| Tweets | Cassandra | Write-heavy, time-series |
| Social graph | Cassandra | Follower/following queries |
| Timeline cache | Redis Sorted Set | O(log N) insert, O(log N+M) range |
| Search | Elasticsearch | Full-text, relevance |
| Trending | Count-Min Sketch + Redis | Streaming counts |
| Queue | Kafka | Decouple, replay |
| Media | S3 + CDN | Durable, global delivery |
