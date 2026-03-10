# Design Instagram

## 1. Problem Statement & Requirements

### Problem Statement
Design a photo and video sharing platform like Instagram that allows users to upload media, follow other users, view a personalized feed, share ephemeral stories, explore content, and interact through likes and comments.

### Functional Requirements
- **Upload media**: Photos and short videos (max 60 sec)
- **Follow users**: Follow/unfollow, view follower/following lists
- **News feed**: Chronological feed of posts from followed users
- **Stories**: Ephemeral content visible for 24 hours
- **Explore**: Discover content (recommended, trending)
- **Engagement**: Like, comment on posts
- **Direct messaging**: 1-on-1 and group messaging
- **Search**: Search users, hashtags, locations

### Non-Functional Requirements
- **Latency**: Feed load < 300ms (p99)
- **Media delivery**: < 2s for first byte (CDN)
- **Availability**: 99.99%
- **Storage**: Durable, no data loss

### Out of Scope
- Live streaming
- Shopping features
- Reels algorithm (assume chronological)
- Advanced ML recommendations

---

## 2. Back-of-Envelope Estimation

### Scale Assumptions
- **MAU**: 2 billion
- **DAU**: 500 million
- **Photos/day**: 100 million
- **Videos/day**: 20 million
- **Feed reads**: 2B/day
- **Story views**: 500M/day

### QPS Calculation
| Operation | Daily Volume | QPS |
|-----------|--------------|-----|
| Photo upload | 100M | ~1,200 |
| Video upload | 20M | ~230 |
| Feed reads | 2B | ~23,000 |
| Story views | 500M | ~5,800 |
| Like/comment | 1B | ~11,600 |

### Storage (5 years)
- **Photos**: 100M × 365 × 5 × 2MB (avg) ≈ 365 PB
- **Videos**: 20M × 365 × 5 × 20MB ≈ 7.3 EB
- **Thumbnails**: 10% of original ≈ 40 PB
- **Metadata**: 120M × 365 × 5 × 1KB ≈ 220 TB

### Bandwidth
- **Upload**: 120M × 5MB avg ≈ 600 TB/day
- **Download**: 2B feeds × 20 images × 100KB ≈ 4 PB/day

### Cache
- **Feed cache**: 500M DAU × 100 posts × 2KB ≈ 100 TB (cache 10% hot)
- **User graph**: 2B × 200 × 50 bytes ≈ 20 TB

---

## 3. API Design

### REST Endpoints

```
POST   /api/v1/media/upload
Body: multipart/form-data (image/video)
Response: { "media_id": "m123", "urls": {"thumbnail": "...", "standard": "...", "hd": "..."} }

POST   /api/v1/posts
Body: { "media_ids": ["m1","m2"], "caption": "...", "location": {...} }
Response: { "post_id": "p123", "created_at": "..." }

GET    /api/v1/posts/:id
GET    /api/v1/users/:id/posts?cursor=&limit=20

GET    /api/v1/feed?cursor=&limit=20
Response: { "posts": [...], "next_cursor": "..." }

POST   /api/v1/stories
Body: { "media_id": "m123", "type": "photo|video" }
Response: { "story_id": "s123", "expires_at": "..." }

GET    /api/v1/stories/feed
Response: { "users_with_stories": [...], "stories": [...] }

GET    /api/v1/explore?cursor=&limit=20
Response: { "posts": [...], "next_cursor": "..." }

POST   /api/v1/posts/:id/like
POST   /api/v1/posts/:id/unlike
POST   /api/v1/posts/:id/comments
Body: { "text": "...", "parent_comment_id": "..." }

GET    /api/v1/posts/:id/comments?cursor=&limit=20

POST   /api/v1/users/:id/follow
POST   /api/v1/users/:id/unfollow

GET    /api/v1/search?q=query&type=users|hashtags|places&limit=20
```

---

## 4. Data Model / Database Schema

### Database Choice
- **Users, follows**: MySQL (relational)
- **Posts, likes**: Cassandra (write-heavy)
- **Comments**: Cassandra (nested structure)
- **Stories**: Redis + Cassandra (TTL)
- **Search**: Elasticsearch

### Schema

**Users (MySQL)**
```sql
users (
  user_id BIGINT PRIMARY KEY,
  username VARCHAR(50) UNIQUE,
  full_name VARCHAR(100),
  profile_pic_url VARCHAR(500),
  bio TEXT,
  is_private BOOLEAN,
  created_at TIMESTAMP
)
```

**Posts (Cassandra)**
```sql
posts_by_user (
  user_id BIGINT,
  post_id TIMEUUID,
  media_urls MAP<TEXT,TEXT>,  -- thumbnail, standard, hd
  caption TEXT,
  location_id BIGINT,
  created_at TIMESTAMP,
  likes_count COUNTER,
  comments_count COUNTER,
  PRIMARY KEY (user_id, post_id)
) WITH CLUSTERING ORDER BY (post_id DESC);

posts_by_id (
  post_id TIMEUUID PRIMARY KEY,
  user_id BIGINT,
  ...
)
```

**Follows (Cassandra)**
```sql
followers (
  user_id BIGINT,
  follower_id BIGINT,
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

**Likes (Cassandra)**
```sql
likes (
  post_id TIMEUUID,
  user_id BIGINT,
  created_at TIMESTAMP,
  PRIMARY KEY (post_id, user_id)
)
```

**Comments (Cassandra)**
```sql
comments (
  post_id TIMEUUID,
  comment_id TIMEUUID,
  user_id BIGINT,
  text TEXT,
  parent_comment_id TIMEUUID,  -- for replies
  created_at TIMESTAMP,
  PRIMARY KEY (post_id, comment_id)
)
```

**Stories (Redis + Cassandra)**
```
Redis: story:{user_id} -> List of story_ids (TTL 24h)
Cassandra: stories (backup, TTL 24h)
```

---

## 5. High-Level Architecture

```
                                    ┌─────────────────────────────────────────────────────────────┐
                                    │                   CLIENT (iOS/Android/Web)                   │
                                    └─────────────────────────────────────────────────────────────┘
                                                              │
                                                              ▼
┌─────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                          API GATEWAY / LOAD BALANCER                                          │
└─────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
                                                              │
    ┌──────────────────┬──────────────────┬──────────────────┬──────────────────┬──────────────────┐
    ▼                  ▼                  ▼                  ▼                  ▼                  ▼
┌─────────┐      ┌──────────┐      ┌──────────┐      ┌──────────┐      ┌──────────┐      ┌──────────┐
│ Media   │      │  Post    │      │  Feed    │      │  Story   │      │ Explore  │      │  User    │
│ Upload  │      │ Service  │      │ Service  │      │ Service  │      │ Service  │      │ Service  │
│ Service │      │          │      │          │      │          │      │          │      │          │
└────┬────┘      └────┬─────┘      └────┬─────┘      └────┬─────┘      └────┬─────┘      └────┬─────┘
     │                │                 │                 │                 │                 │
     ▼                │                 │                 │                 │                 │
┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                        MEDIA PROCESSING PIPELINE (Resize, Thumbnail, Transcode)                       │
│  Upload → S3 → Queue → Worker (multiple resolutions) → S3 → CDN                                     │
└─────────────────────────────────────────────────────────────────────────────────────────────────────┘
     │                │                 │                 │                 │                 │
     ▼                ▼                 ▼                 ▼                 ▼                 ▼
┌─────────┐      ┌──────────┐      ┌──────────┐      ┌──────────┐      ┌──────────┐      ┌──────────┐
│   S3    │      │Cassandra │      │  Redis   │      │  Redis   │      │Elastic   │      │  MySQL   │
│  + CDN  │      │ (Posts)  │      │ (Feed)   │      │ (Stories)│      │ Search   │      │ (Users)  │
└─────────┘      └──────────┘      └──────────┘      └──────────┘      └──────────┘      └──────────┘
```

### Media Upload Pipeline (ASCII)

```
    CLIENT                    API                    S3                    QUEUE                 WORKERS
       │                       │                      │                      │                      │
       │  Upload (chunked)     │                      │                      │                      │
       │─────────────────────>│                      │                      │                      │
       │                       │  Store raw           │                      │                      │
       │                       │─────────────────────>│                      │                      │
       │                       │  Return upload_id    │                      │                      │
       │<─────────────────────│                      │                      │                      │
       │                       │                      │                      │                      │
       │  Create post          │                      │                      │                      │
       │─────────────────────>│                      │                      │                      │
       │                       │  Publish job         │                      │                      │
       │                       │────────────────────────────────────────────>│                      │
       │                       │                      │                      │  Consume job         │
       │                       │                      │                      │<─────────────────────│
       │                       │                      │                      │                      │
       │                       │                      │  Fetch raw            │  Resize/Transcode    │
       │                       │                      │<─────────────────────────────────────────────│
       │                       │                      │  Store thumbnail      │                      │
       │                       │                      │<─────────────────────────────────────────────│
       │                       │                      │  Store 480p          │                      │
       │                       │                      │<─────────────────────────────────────────────│
       │                       │                      │  Store 1080p         │                      │
       │                       │                      │<─────────────────────────────────────────────│
       │                       │  Post ready          │                      │                      │
       │                       │<─────────────────────────────────────────────────────────────────────│
       │  Post published       │                      │                      │                      │
       │<─────────────────────│                      │                      │                      │
```

---

## 6. Detailed Component Design

### 6.1 Media Upload Pipeline
**Flow**:
1. **Chunked upload**: Client uploads in 5MB chunks (resumable)
2. **Raw storage**: Store in S3 temporarily
3. **Processing queue**: Publish to SQS/Kafka
4. **Worker tasks**:
   - Generate thumbnail (320x320)
   - Resize to 480p, 1080p
   - For video: transcode to H.264, multiple resolutions
   - Extract metadata (dimensions, duration)
5. **Output**: Store in S3, register in CDN
6. **Post creation**: Link media URLs to post

**Resolutions**: Thumbnail (320px), Standard (480p), HD (1080p)

### 6.2 CDN for Media
- **CloudFront/CloudFront**: Global edge locations
- **Cache**: Aggressive caching (1 year for immutable URLs)
- **URL structure**: `cdn.instagram.com/{user_id}/{media_id}_{resolution}.jpg`

### 6.3 News Feed Generation
- **Fan-out on write**: Similar to Twitter
- **Pre-compute**: Push post_id to follower timelines in Redis
- **Celebrity handling**: Merge on read for users with >100K followers
- **Feed structure**: Sorted set by timestamp

### 6.4 Database Sharding
- **Shard key**: user_id (consistent hashing)
- **Posts**: Partition by user_id
- **Feed cache**: Shard Redis by user_id

### 6.5 Explore/Recommendation Service
- **Sources**: 
  - Posts from friends-of-friends
  - Popular in user's region
  - Hashtags user follows
- **Ranking**: Engagement score (likes, comments, recency)
- **Implementation**: Pre-compute explore feed, refresh hourly

### 6.6 Stories (24hr TTL)
- **Storage**: Redis with 24h TTL
- **Structure**: `story:{user_id}` → list of story_ids
- **Persistence**: Cassandra with TTL for durability
- **View tracking**: Separate table for "who viewed"
- **Cleanup**: Background job removes expired stories

### 6.7 Search
- **Elasticsearch**: Index users, posts (caption, hashtags), locations
- **User search**: By username, full name
- **Hashtag search**: Inverted index
- **Location search**: Geospatial index

### 6.8 Likes and Comments
- **Likes**: Cassandra, partition by post_id
- **Counters**: Denormalized like_count on post
- **Comments**: Tree structure with parent_comment_id for replies

---

## 7. Scaling

### Sharding
- **By user_id**: Posts, follows, feed
- **By post_id**: Likes, comments

### Caching
- **Feed**: Redis, 70% hit rate target
- **User profile**: Redis, 5 min TTL
- **Media**: CDN handles 95% of requests

### CDN
- **All media** served via CDN
- **Multiple regions** for low latency

### Async Processing
- **Media processing**: Fully async, don't block upload
- **Feed fan-out**: Kafka consumers

---

## 8. Failure Handling

### Media Pipeline
- **Worker failure**: Retry from queue
- **Partial upload**: Client retries chunks
- **S3 failure**: Multi-AZ, retry

### Feed
- **Cache miss**: Rebuild from DB (slower path)
- **Fan-out delay**: Eventual consistency OK

### Stories
- **Redis failure**: Rebuild from Cassandra
- **TTL drift**: Background reconciliation

---

## 9. Monitoring & Observability

### Key Metrics
- **Upload latency**: Time to first byte, time to post ready
- **Feed latency**: p99 load time
- **Media delivery**: CDN cache hit rate, latency
- **Processing queue**: Lag, processing time
- **Story TTL**: Expiry accuracy

### Alerts
- Media processing lag > 1 hour
- Feed p99 > 500ms
- CDN error rate > 0.1%

---

## 10. Interview Tips

### Follow-up Questions
1. **How to handle viral posts?** CDN, cache popular posts at edge
2. **Story view count at scale?** Approximate counting, or sampled
3. **Explore ranking?** Collaborative filtering, content-based, hybrid
4. **Private accounts?** Filter at feed generation, don't include in explore

### Common Mistakes
- **Synchronous media processing**: Blocks user
- **No CDN**: Can't serve media at scale
- **Single resolution**: Mobile needs thumbnails
- **Stories in main DB**: Need TTL, Redis better

### Key Points
1. **Media pipeline** is critical (resize, CDN)
2. **Fan-out** similar to Twitter (hybrid)
3. **Stories** need ephemeral storage (Redis TTL)
4. **Sharding** by user_id
5. **Explore** is separate from feed (different ranking)

---

## Appendix A: Extended Design Details

### A.1 Media Processing Worker Configuration
```
Photo: thumbnail (320px), standard (1080px), hd (original)
Video: 240p, 360p, 480p, 720p, 1080p (H.264)
Codec: libx264, CRF 23 for balance
Thumbnail: Extract at 1s for video, resize for photo
```

### A.2 Story View Tracking Schema
```sql
story_views (
  story_id UUID,
  viewer_id BIGINT,
  viewed_at TIMESTAMP,
  PRIMARY KEY (story_id, viewer_id)
)
-- TTL 24h to auto-expire with story
```

### A.3 Explore Ranking Formula
```
score = 0.3 * recency_score + 0.4 * engagement_score + 0.2 * relevance_score + 0.1 * diversity_score
engagement_score = log(1 + likes + 2*comments)
relevance = cosine_similarity(user_embedding, post_embedding)
```

### A.4 S3 Bucket Structure
```
/uploads/{user_id}/{upload_id}/chunk_0, chunk_1, ...
/processed/{user_id}/{post_id}/thumbnail.jpg
/processed/{user_id}/{post_id}/standard.jpg
/processed/{user_id}/{post_id}/hd.jpg
/stories/{user_id}/{story_id}.jpg  (TTL 24h)
```

### A.5 Feed Cache Invalidation
- **On new post**: Fan-out worker pushes to follower timelines
- **On unfollow**: Remove posts from timeline (lazy or eager)
- **On post delete**: Remove from all follower timelines (async job)
- **Cache TTL**: No TTL for timeline (updated on write), 5 min for user profile

### A.6 Direct Messaging (Out of Scope but Architecture)
- Similar to WhatsApp: WebSocket, message queue, store-and-forward
- Media: Same pipeline as posts
- Read receipts: last_read_message_id per chat per user

### A.7 Sample API Response - Feed
```json
{
  "posts": [
    {
      "post_id": "uuid",
      "user_id": 123,
      "username": "johndoe",
      "media_urls": {"thumbnail": "...", "standard": "..."},
      "caption": "...",
      "likes_count": 150,
      "comments_count": 12,
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "next_cursor": "base64_encoded_cursor"
}
```

---

## Appendix B: Story Architecture Deep Dive

### B.1 Story Storage Flow
```
1. User uploads photo/video -> Media pipeline (same as post)
2. Create story: INSERT stories (user_id, story_id, media_url, expires_at)
3. Redis: RPUSH story:{user_id} story_id
4. Redis: EXPIRE story:{user_id} 86400  # 24 hours
5. Cassandra: TTL 24h on story row
```

### B.2 Story Feed Aggregation
```
1. Get users_with_stories = users I follow who have story:{user_id} key in Redis
2. For each: Get story list from Redis LRANGE story:{user_id} 0 -1
3. Return: [{user_id, username, profile_pic, stories: [...]}]
4. Client displays as "Story rings" on feed
```

### B.3 Story View Count
- Option A: Exact count - Cassandra counter (expensive at scale)
- Option B: Approximate - Redis INCR with TTL (good enough)
- Option C: Sampled - 1% of views recorded for analytics

### B.4 Explore Cold Start
- New user: Show trending (most likes last 24h)
- After 10 likes: Use content-based (tags from liked posts)
- After 50 interactions: Collaborative filtering kicks in

---

## Appendix C: Walkthrough Scenarios

### C.1 Scenario: User Uploads Photo and Creates Post
1. Client uploads photo (5MB) in single request or chunks
2. API stores in S3 uploads bucket, returns media_id
3. Client calls POST /api/v1/posts with media_ids, caption
4. Post Service creates post record (status=processing)
5. Publishes to SQS/Kafka media-processing queue
6. Worker picks up job, fetches from S3
7. Generates thumbnail (320px), standard (1080px), hd (original)
8. Stores processed images in S3 processed bucket
9. Updates post with media URLs, status=ready
10. Fan-out worker pushes to follower timelines (Redis)
11. Client can now see post in feed (polling or push)
12. Total time: 10-30 seconds (async, user sees "Posting...")

### C.2 Scenario: User Views Stories Feed
1. Client requests GET /api/v1/stories/feed
2. Story Service gets list of followed user IDs
3. For each: Check Redis EXISTS story:{user_id}
4. Users with key have active stories
5. For each: LRANGE story:{user_id} 0 -1 to get story IDs
6. Fetch story metadata from Cassandra (or Redis)
7. Return: [{user: {...}, stories: [...]}, ...]
8. Client displays story rings, user taps to view
9. On view: Increment view count (Redis INCR), record viewer

### C.3 Scenario: Explore Page for Returning User
1. Client requests GET /api/v1/explore
2. Check Redis for pre-computed explore:user:{id}
3. If exists: Return cached (refreshed hourly)
4. If not: Compute on-the-fly
   - Get user's liked posts, extract tags
   - Find posts with overlapping tags (friends-of-friends)
   - Rank by engagement score
   - Exclude already seen, return top 20
5. Cache result for 1 hour
6. Latency: 50ms (cached) or 200ms (compute)

### C.4 Scenario: Viral Post (1M likes in 1 hour)
- **Feed**: Already in follower timelines, no change
- **Explore**: High engagement score, surfaces to more users
- **CDN**: Media URLs cached at edge, handles load
- **Likes**: Cassandra counter, handles high write rate
- **Comments**: May need read replicas if comment load spikes

---

## Appendix D: Technology Choices Rationale

### D.1 Why S3 for Media?
- **Durability**: 11 nines, no data loss
- **Scale**: Unlimited storage, pay per use
- **CDN integration**: CloudFront origin
- **Cost**: Cheaper than block storage for large files
- **Versioning**: Optional for rollback

### D.2 Why Redis for Stories?
- **TTL**: Native 24h expiry, no cron needed
- **Fast**: Sub-ms read/write for story feed
- **List operations**: RPUSH for append, LRANGE for fetch
- **Memory**: Stories are small (metadata + URL)
- **Fallback**: Cassandra with TTL for durability

### D.3 Why Cassandra for Posts?
- **Write pattern**: Append-only, partition by user
- **Counters**: Native like_count, comment_count
- **Scale**: Linear with nodes
- **Flexible schema**: Add columns without migration
- **Multi-DC**: Built-in replication for geo

### D.4 Why Async Media Processing?
- **User experience**: Don't block upload, show "Posting..."
- **Resilience**: Retry on failure, no user retry needed
- **Scale**: Add workers independently
- **Cost**: Spot instances for batch processing

### D.5 Interview Discussion Points
- **Media first**: Instagram is media-heavy, pipeline is critical
- **Stories differentiation**: Ephemeral = different storage (Redis TTL)
- **Explore vs Feed**: Different ranking, different cache keys

---

## Quick Reference Summary

| Component | Technology | Key Reason |
|-----------|------------|------------|
| Media storage | S3 + CDN | Scale, durability |
| Posts | Cassandra | Write-heavy, counters |
| Feed | Redis | Fan-out on write |
| Stories | Redis + Cassandra | TTL 24h |
| Explore | Pre-computed + Redis | Ranking, cache |
| Search | Elasticsearch | Full-text |

*Note: All files follow the 10-section structure: Problem, Estimation, API, Data Model, Architecture, Components, Scaling, Failure, Monitoring, Interview Tips.*

*Appendices provide deep dives, walkthrough scenarios, and technology rationale.*
