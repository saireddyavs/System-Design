# Design YouTube

## 1. Problem Statement & Requirements

### Problem Statement
Design a video sharing platform like YouTube that supports video upload, streaming, search, recommendations, comments, likes, and subscriptions at massive scale.

### Functional Requirements
- **Upload videos**: Support large files (up to 256GB), chunked upload, resumable
- **Watch videos**: Stream with adaptive bitrate (multiple resolutions)
- **Search**: Full-text search on title, description, tags
- **Recommendations**: Personalized home feed, related videos
- **Comments**: Threaded comments, replies
- **Likes/Dislikes**: Engagement tracking
- **Subscriptions**: Follow channels, subscription feed
- **Thumbnails**: Auto-generated and custom

### Non-Functional Requirements
- **Upload**: 500 hours of video uploaded per minute
- **Watch**: 1 billion hours watched per day
- **Latency**: Video start < 2s, search < 200ms
- **Availability**: 99.99%

### Out of Scope
- Live streaming
- YouTube Shorts (vertical video)
- Monetization/ads
- Copyright detection (Content ID)

---

## 2. Back-of-Envelope Estimation

### Scale Assumptions
- **Upload**: 500 hours/min = 720K hours/day
- **Avg video**: 10 min, 500MB (1080p)
- **Watch**: 1B hours/day
- **Videos**: 2B total, 50M new/month

### QPS Calculation
| Operation | Daily Volume | QPS |
|-----------|--------------|-----|
| Video upload (chunks) | 720K hrs × 50 chunks ≈ 36M | ~420 |
| Video metadata write | 720K × 6 = 4.3M | ~50 |
| Video stream requests | 1B hrs × 10 segments/hr ≈ 10B | ~115K |
| Search queries | 500M | ~5,800 |
| Recommendation requests | 2B | ~23,000 |

### Storage (5 years)
- **Video files**: 720K × 365 × 5 × 500MB ≈ 657 PB/year → 3.3 EB
- **Thumbnails**: 50M × 100KB × 5 ≈ 25 PB
- **Metadata**: 50M × 1KB × 60 ≈ 3 TB
- **Transcoded variants**: 5 formats × 5 resolutions = 25× multiplier on storage

### Bandwidth
- **Upload**: 720K × 500MB ≈ 360 PB/day
- **Download (streaming)**: 1B hours × 2 Mbps avg ≈ 216 PB/day

### Cache
- **Hot videos**: 20% of views = 80% traffic (Pareto)
- **CDN**: Serve 95% of video bytes from edge
- **Metadata cache**: Redis, 100M videos × 1KB ≈ 100 GB

---

## 3. API Design

### REST Endpoints

```
POST   /api/v1/videos/upload/init
Body: { "title": "...", "description": "...", "file_size": 1234567890 }
Response: { "upload_id": "u123", "chunk_urls": ["...", "..."] }

PUT    /api/v1/videos/upload/chunk/:upload_id/:chunk_index
Body: binary chunk
Response: { "etag": "..." }

POST   /api/v1/videos/upload/complete
Body: { "upload_id": "u123", "chunk_etags": [...] }
Response: { "video_id": "v123", "status": "processing" }

GET    /api/v1/videos/:id
Response: { "video_id", "title", "channel_id", "thumbnail_url", "duration", "formats": [...] }

GET    /api/v1/videos/:id/stream
Query: format=hls|dash, quality=360p|720p|1080p
Response: Redirect to CDN URL or manifest

GET    /api/v1/videos/:id/thumbnail
GET    /api/v1/search?q=query&limit=20&cursor=
Response: { "videos": [...], "next_cursor": "..." }

GET    /api/v1/recommendations
Response: { "videos": [...] }

GET    /api/v1/videos/:id/related
Response: { "videos": [...] }

POST   /api/v1/videos/:id/like
POST   /api/v1/videos/:id/dislike
POST   /api/v1/videos/:id/comments
Body: { "text": "...", "parent_id": "..." }

GET    /api/v1/videos/:id/comments?cursor=&limit=20

POST   /api/v1/channels/:id/subscribe
GET    /api/v1/feed/subscriptions
```

---

## 4. Data Model / Database Schema

### Database Choice
- **Videos metadata**: Cassandra (read-heavy after upload)
- **Channels/Users**: MySQL
- **Comments**: Cassandra (partition by video_id)
- **Likes/Subscriptions**: Cassandra
- **Search**: Elasticsearch
- **Recommendations**: Pre-computed in Redis/Cassandra

### Schema

**Channels (MySQL)**
```sql
channels (
  channel_id BIGINT PRIMARY KEY,
  user_id BIGINT,
  name VARCHAR(100),
  description TEXT,
  subscriber_count BIGINT,
  created_at TIMESTAMP
)
```

**Videos (Cassandra)**
```sql
videos_by_id (
  video_id UUID PRIMARY KEY,
  channel_id BIGINT,
  title VARCHAR(200),
  description TEXT,
  duration_seconds INT,
  thumbnail_url VARCHAR(500),
  status VARCHAR(20),       -- processing, ready, failed
  created_at TIMESTAMP,
  view_count COUNTER,
  like_count COUNTER,
  comment_count COUNTER
)

videos_by_channel (
  channel_id BIGINT,
  video_id UUID,
  created_at TIMESTAMP,
  PRIMARY KEY (channel_id, created_at)
) WITH CLUSTERING ORDER BY (created_at DESC);
```

**Video Formats (Cassandra)**
```sql
video_formats (
  video_id UUID,
  format_id VARCHAR(20),    -- 360p_h264, 720p_h264, etc.
  storage_path VARCHAR(500),
  resolution VARCHAR(10),
  bitrate_kbps INT,
  codec VARCHAR(20),
  PRIMARY KEY (video_id, format_id)
)
```

**Comments (Cassandra)**
```sql
comments (
  video_id UUID,
  comment_id TIMEUUID,
  user_id BIGINT,
  text TEXT,
  parent_comment_id TIMEUUID,
  created_at TIMESTAMP,
  like_count COUNTER,
  PRIMARY KEY (video_id, comment_id)
) WITH CLUSTERING ORDER BY (comment_id DESC);
```

**Subscriptions (Cassandra)**
```sql
subscriptions (
  user_id BIGINT,
  channel_id BIGINT,
  created_at TIMESTAMP,
  PRIMARY KEY (user_id, channel_id)
)

subscribers (
  channel_id BIGINT,
  user_id BIGINT,
  created_at TIMESTAMP,
  PRIMARY KEY (channel_id, user_id)
)
```

---

## 5. High-Level Architecture

```
                                    ┌─────────────────────────────────────────────────────────────┐
                                    │                    CLIENT (Web/Mobile/TV)                   │
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
│ Upload  │      │ Streaming│      │  Search  │      │Recommend │      │  Social  │      │  Video   │
│ Service │      │ Service  │      │ Service  │      │ Service  │      │ Service  │      │ Metadata │
└────┬────┘      └────┬─────┘      └────┬─────┘      └────┬─────┘      └────┬─────┘      └────┬─────┘
     │                │                 │                 │                 │                 │
     ▼                │                 │                 │                 │                 │
┌─────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                        UPLOAD PIPELINE (DAG: Transcode, Thumbnail, Index)                                      │
│  Chunked Upload → S3 → Queue → Transcoding Workers (parallel) → S3 → CDN → Metadata DB                        │
└─────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
     │                │                 │                 │                 │                 │
     ▼                ▼                 ▼                 ▼                 ▼                 ▼
┌─────────┐      ┌──────────┐      ┌──────────┐      ┌──────────┐      ┌──────────┐      ┌──────────┐
│   S3    │      │   CDN    │      │Elastic   │      │  Redis   │      │Cassandra │      │Cassandra │
│ Raw +   │      │ (HLS/DASH│      │ Search   │      │ Pre-comp │      │ Comments │      │  Videos  │
│Transcoded│     │ segments)│      │          │      │ Recs     │      │  Likes   │      │ Metadata │
└─────────┘      └──────────┘      └──────────┘      └──────────┘      └──────────┘      └──────────┘
```

### Video Upload Pipeline (ASCII)

```
    CLIENT                    UPLOAD SVC              S3 (Raw)              QUEUE              TRANSCODE WORKERS
       │                           │                      │                    │                      │
       │  Init upload               │                      │                    │                      │
       │───────────────────────────>│                      │                    │                      │
       │  upload_id, chunk_urls     │                      │                    │                      │
       │<───────────────────────────│                      │                    │                      │
       │                           │                      │                    │                      │
       │  Upload chunk 0            │                      │                    │                      │
       │───────────────────────────>│  Store chunk         │                    │                      │
       │                           │─────────────────────>│                    │                      │
       │  ... chunks 1..N           │                      │                    │                      │
       │                           │                      │                    │                      │
       │  Complete upload           │                      │                    │                      │
       │───────────────────────────>│                      │                    │                      │
       │                           │  Create job DAG      │                    │                      │
       │                           │─────────────────────────────────────────>│                      │
       │  video_id, status=proc     │                      │                    │  Consume job        │
       │<───────────────────────────│                      │                    │<─────────────────────│
       │                           │                      │                    │                      │
       │                           │                      │  Fetch raw          │  Transcode 360p      │
       │                           │                      │<─────────────────────────────────────────│
       │                           │                      │                    │  Transcode 720p      │
       │                           │                      │                    │  Transcode 1080p     │
       │                           │                      │  Store segments     │  Generate thumbnail  │
       │                           │                      │<─────────────────────────────────────────│
       │                           │  Update metadata     │                    │  Extract metadata   │
       │                           │<─────────────────────────────────────────────────────────────────│
       │                           │  Publish to CDN      │                    │                      │
       │                           │  Index in Search     │                    │                      │
```

### Video Streaming Flow (ASCII)

```
    CLIENT                    STREAMING SVC              CDN                    ORIGIN (S3)
       │                           │                      │                      │
       │  GET /stream?format=hls    │                      │                      │
       │───────────────────────────>│                      │                      │
       │                           │  Lookup video formats│                      │
       │                           │  (Cassandra)          │                      │
       │                           │  Select best quality  │                      │
       │                           │  (user bandwidth)     │                      │
       │                           │                      │                      │
       │                           │  Redirect to CDN URL  │                      │
       │  302 Redirect              │                      │                      │
       │<───────────────────────────│                      │                      │
       │                           │                      │                      │
       │  GET manifest.m3u8        │                      │                      │
       │─────────────────────────────────────────────────>│                      │
       │                           │                      │  Cache miss?         │
       │                           │                      │  Fetch from S3       │
       │                           │                      │─────────────────────>│
       │  manifest content          │                      │<─────────────────────│
       │<─────────────────────────────────────────────────│                      │
       │                           │                      │                      │
       │  GET segment_0.ts         │                      │                      │
       │─────────────────────────────────────────────────>│  (repeat for each)   │
       │  segment_1.ts              │                      │                      │
       │  ... (adaptive: switch quality based on buffer)   │                      │
```

---

## 6. Detailed Component Design

### 6.1 Video Upload Pipeline

**Chunked Upload**:
- **Chunk size**: 5-10 MB (resumable)
- **Init**: Create upload_id, generate pre-signed URLs for each chunk
- **Parallel upload**: Client uploads chunks in parallel (3-5 concurrent)
- **Complete**: Verify all chunks (etag), merge or trigger processing

**Processing DAG**:
1. **Copy to processing**: Move from upload bucket to processing
2. **Transcode**: FFmpeg to H.264/VP9, multiple resolutions (360p, 480p, 720p, 1080p)
3. **Segment**: HLS (m3u8 + .ts) or DASH (mpd + m4s)
4. **Thumbnail**: Extract frame at 10%, 50%, 90% or user-selected
5. **Metadata**: Extract duration, resolution, codec
6. **Upload to CDN origin**: S3/GCS
7. **Update DB**: video_formats, status=ready
8. **Index**: Add to Elasticsearch

**Parallelism**: Each resolution is independent task, run in parallel. Use workflow engine (Temporal, Airflow) or simple queue with dependencies.

### 6.2 Video Storage
- **Object storage**: S3/GCS for raw and transcoded
- **Structure**: `videos/{video_id}/{format}/segment_0.ts`
- **CDN**: CloudFront/CloudFront, cache at edge
- **Distribution**: Invalidate on upload, long TTL for immutable segments

### 6.3 Video Streaming (Adaptive Bitrate)
- **HLS**: Apple format, .m3u8 manifest + .ts segments
- **DASH**: MPEG-DASH, .mpd manifest + .m4s segments
- **Adaptive**: Client monitors buffer, switches quality (ABR algorithm)
- **Resolutions**: 144p, 240p, 360p, 480p, 720p, 1080p, 4K
- **CDN**: Serve segments from edge, 95% cache hit

### 6.4 Video Metadata Service
- **Storage**: Cassandra, partition by video_id
- **Cache**: Redis for hot videos (top 10%)
- **Counters**: view_count, like_count (Cassandra counters)

### 6.5 Search (Inverted Index)
- **Elasticsearch**: Index title, description, tags, channel name
- **Ranking**: TF-IDF + recency + engagement (views, likes)
- **Autocomplete**: Edge n-gram for suggestions
- **Pipeline**: Index on video status=ready

### 6.6 Recommendation Engine
**Collaborative filtering**:
- Users who watched X also watched Y
- Item-item similarity matrix
- Matrix factorization (SVD, ALS)

**Content-based**:
- Same channel, same tags, same category
- Embedding from title/description (ML)

**Hybrid**:
- Combine both, rank by predicted engagement
- **Pre-compute**: Batch job daily, store top-K per user in Redis
- **Real-time**: Merge with recent watches, exclude already seen

### 6.7 Comments (Tree Structure)
- **Storage**: Cassandra, partition by video_id
- **Structure**: parent_comment_id for replies (max depth 2-3)
- **Pagination**: Cursor-based, sort by created_at or like_count
- **Load**: Lazy load replies (expand on click)

### 6.8 CDN Placement
- **Edge locations**: 200+ globally
- **Cache**: Video segments (immutable), long TTL
- **Origin**: S3 in multiple regions
- **Routing**: GeoDNS to nearest edge

### 6.9 Video Deduplication
- **Content hash**: Perceptual hash (pHash) or frame sampling
- **Storage**: Store hash, detect duplicates on upload
- **Use case**: Re-upload same video, copyright
- **Implementation**: Optional, compute-intensive

---

## 7. Scaling

### Sharding
- **Videos**: Shard by video_id
- **Comments**: Partition by video_id (hot videos may need splitting)
- **Formats**: Colocated with video

### Caching
- **Metadata**: Redis, 10% hot videos
- **Recommendations**: Pre-computed in Redis
- **CDN**: 95% of video traffic

### Transcoding
- **Horizontal**: Add more workers
- **GPU**: Use GPU for faster transcoding (NVENC)
- **Priority**: Popular videos first (predict at upload)

---

## 8. Failure Handling

### Upload
- **Chunk failure**: Client retries chunk
- **Processing failure**: Retry from last successful step (checkpoint)
- **Partial transcode**: Serve available resolutions, continue in background

### Streaming
- **CDN failure**: Fallback to origin
- **Segment missing**: Skip or retry, client handles

### Redundancy
- **S3**: Multi-AZ, versioning
- **Cassandra**: RF=3
- **CDN**: Multi-provider failover

---

## 9. Monitoring & Observability

### Key Metrics
- **Upload**: Time to first byte, time to processing complete
- **Streaming**: Start time (TTFB), buffer health, quality switches
- **Search**: Latency p99, result relevance
- **Recommendations**: CTR, watch time
- **Transcoding**: Queue depth, processing time per resolution

### Alerts
- Transcoding lag > 1 hour
- CDN cache hit rate < 90%
- Search p99 > 500ms
- Upload failure rate > 1%

---

## 10. Interview Tips

### Follow-up Questions
1. **How to handle 4K/8K?** Same pipeline, add resolution, consider lazy transcode (on first request)
2. **Live streaming?** Different architecture (ingest → transcode → distribute)
3. **Video deletion?** Soft delete, async purge from S3, remove from index
4. **Recommendation cold start?** Use trending, category, popular in region
5. **Cost optimization?** Transcode on-demand for long-tail, use spot instances

### Common Mistakes
- **Synchronous transcode**: Blocks upload, use async
- **No chunked upload**: Fails for large files, no resume
- **Single quality**: Need adaptive bitrate
- **No CDN**: Can't serve 115K stream QPS
- **Real-time recommendations**: Too slow, pre-compute

### Key Points
1. **Chunked upload** for large files, resumable
2. **DAG for processing** (transcode, thumbnail, index)
3. **Adaptive bitrate** (HLS/DASH) for streaming
4. **CDN** for 95% of video delivery
5. **Pre-compute recommendations** for latency
6. **Sharding** by video_id
7. **Object storage** (S3) for video files

---

## Appendix A: Extended Design Details

### A.1 HLS Manifest Example
```
#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
#EXTINF:10.0,
segment_0.ts
#EXTINF:10.0,
segment_1.ts
#EXT-X-ENDLIST
```

### A.2 Transcoding Job DAG (Temporal/Airflow)
```
Task 1: Download raw from S3
Task 2a: Transcode 360p (parallel)
Task 2b: Transcode 720p (parallel)
Task 2c: Transcode 1080p (parallel)
Task 3: Generate thumbnail
Task 4: Upload all to S3
Task 5: Update metadata DB
Task 6: Index in Elasticsearch
Dependencies: 1 -> 2a,2b,2c,3 -> 4 -> 5,6
```

### A.3 Recommendation Pre-compute Pipeline
```
1. Extract watch history (user, video, watch_time)
2. Compute item-item similarity (cosine on co-watch matrix)
3. For each user: get recent watches -> similar videos
4. Rank by: similarity * popularity * recency
5. Store top 100 in Redis: recs:user:{id}
6. Refresh: Daily batch + real-time exclude recent watches
```

### A.4 Comment Tree Traversal
```
Root comments: SELECT * FROM comments WHERE post_id=? AND parent_id=null
Replies: SELECT * FROM comments WHERE parent_id=?
Max depth: 2 (comment -> reply, no nested)
Sort: Root by like_count DESC, replies by created_at ASC
```

### A.5 Video Segment URL Structure
```
https://cdn.youtube.com/v/{video_id}/360p/segment_0.ts
https://cdn.youtube.com/v/{video_id}/720p/segment_0.ts
Immutable: Same URL forever, long cache TTL
```

### A.6 View Count Update Strategy
- **Real-time**: Too expensive (115K QPS)
- **Batch**: Aggregate every 5 min, update counter
- **Approximate**: HyperLogLog for unique viewers
- **Exact**: Cassandra counter, async increment

### A.7 Search Index Mapping (Elasticsearch)
```json
{
  "mappings": {
    "properties": {
      "title": {"type": "text", "analyzer": "standard"},
      "description": {"type": "text"},
      "tags": {"type": "keyword"},
      "channel_name": {"type": "text"},
      "created_at": {"type": "date"},
      "view_count": {"type": "long"}
    }
  }
}
```

### A.8 Lazy Transcode for Long-Tail
- **Popular videos**: Transcode all resolutions at upload
- **Long-tail**: Transcode 360p only, others on first request
- **Queue**: On-demand transcode queue, lower priority

---

## Appendix B: Streaming Protocol Deep Dive

### B.1 Adaptive Bitrate (ABR) Client Logic
```
Buffer low (< 5s) -> Request lower quality (360p)
Buffer healthy (10-30s) -> Request current or higher
Buffer full (> 30s) -> Request highest quality
Quality switch: Seamless, same segment index different resolution
```

### B.2 CDN Cache Strategy
- **Manifest (.m3u8)**: Short TTL 5 min (may update)
- **Segments (.ts)**: Long TTL 1 year (immutable)
- **Thumbnail**: 1 year TTL
- **Invalidation**: Rarely needed (version in URL)

### B.3 Subscription Feed Generation
- **Pre-compute**: When user subscribes, add channel to list
- **Feed**: Fetch latest video from each subscribed channel
- **Merge**: Sort by published_at, return top 20
- **Cache**: Redis, key=sub_feed:{user_id}, invalidate on new sub/unsub

### B.4 Video Metadata Enrichment
- **On upload**: Title, description, tags from user
- **From file**: Duration, resolution, codec (FFprobe)
- **Auto**: Thumbnail selection (frame at 10%, 50%, 90%)
- **Search**: Index all text fields, tags as keywords

---

## Appendix C: Walkthrough Scenarios

### C.1 Scenario: User Uploads 1GB Video
1. Client calls POST /upload/init with title, description, file_size=1GB
2. Server creates upload_id, divides into 100 chunks (10MB each)
3. Generates 100 pre-signed S3 URLs
4. Returns upload_id and chunk_urls to client
5. Client uploads chunks in parallel (5 concurrent)
6. Each chunk: PUT to pre-signed URL, get etag
7. After all complete: POST /upload/complete with chunk_etags
8. Server verifies etags, merges in S3 (or triggers multipart complete)
9. Creates transcode job, publishes to queue
10. Returns video_id, status=processing
11. Transcode workers: Download, FFmpeg to 360p/720p/1080p
12. Segment to HLS, upload to S3, update DB status=ready
13. Index in Elasticsearch
14. Total: 10-30 min depending on video length and worker capacity

### C.2 Scenario: User Watches Video (Streaming)
1. Client requests GET /videos/v123/stream?format=hls
2. Streaming Service looks up video_formats in Cassandra
3. Finds: 360p, 720p, 1080p segment URLs (CDN)
4. Returns 302 redirect to manifest URL: cdn/v123/720p/playlist.m3u8
5. Client fetches manifest from CDN (cached)
6. Manifest lists segment_0.ts, segment_1.ts, ...
7. Client fetches segment_0 from CDN
8. Plays in video player, buffers
9. ABR: If buffer low, next request for 360p; if healthy, stay 720p
10. Continues until video end or user stops
11. Time to first byte: < 2s (CDN cache hit)

### C.3 Scenario: Recommendation for New User
1. New user (no watch history) requests GET /recommendations
2. Recommendation Service: No Redis cache
3. Fallback: Fetch trending videos (most views last 24h)
4. Or: Popular in user's region (from IP)
5. Or: Default "starter pack" (curated list)
6. Return 20 videos
7. Cache in Redis for 5 min (recs:user:new_user_id)
8. As user watches: Record in watch_history
9. Next day: Batch job computes recommendations
10. Store in Redis for fast subsequent loads

### C.4 Scenario: Search for "python tutorial"
1. User types in search box
2. GET /search?q=python+tutorial&limit=20
3. Search Service queries Elasticsearch
4. Match: title, description, tags containing "python" AND "tutorial"
5. Boost: Recency (newer videos rank higher)
6. Boost: Engagement (views, likes)
7. Return top 20 with thumbnails, channel info
8. Latency: 100-200ms
