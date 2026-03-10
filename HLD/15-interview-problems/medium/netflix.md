# Design Netflix

## 1. Problem Statement & Requirements

### Problem Statement
Design a video streaming platform like Netflix that enables subscribers to browse a catalog, stream video content with high quality, receive personalized recommendations, and watch across multiple devices—including offline downloads.

### Functional Requirements
- **Browse catalog**: Browse movies, TV shows, categories, search by title/actor/genre
- **Stream video**: Adaptive bitrate streaming (ABR), multiple resolutions (SD/HD/4K)
- **Personalized recommendations**: "Because you watched X", trending, continue watching
- **Multi-device support**: Watch on TV, mobile, tablet, web; sync watch progress
- **Download for offline**: Download titles for offline viewing (encrypted)
- **User profiles**: Multiple profiles per account (family sharing)
- **Playback controls**: Play, pause, seek, skip intro/credits

### Non-Functional Requirements
- **Scale**: 250M+ subscribers, 17% of global internet traffic
- **Latency**: Video start < 2 seconds, catalog load < 500ms
- **Availability**: 99.99% (streaming is mission-critical)
- **Quality**: Seamless playback, minimal buffering, 4K HDR support

### Out of Scope
- Content creation/production
- Live streaming (sports, events)
- Ad-supported tier (focus on subscription model)
- Content licensing negotiations

---

## 2. Back-of-Envelope Estimation

### Scale Assumptions
- **Subscribers**: 250M active
- **Concurrent streams**: ~15% peak = 37.5M simultaneous streams
- **Catalog**: ~15,000 titles, ~500K episodes
- **Avg watch session**: 2 hours, 5 Mbps average bitrate

### QPS Calculation
| Operation | Daily Volume | Peak QPS |
|-----------|--------------|----------|
| Catalog browse | 500M | ~6,000 |
| Video stream (segment requests) | 37.5M × 4 segments/min × 60 × 24 ≈ 216B | ~2.5M |
| Search | 100M | ~1,200 |
| Recommendations | 300M | ~3,500 |
| Playback state updates | 500M | ~6,000 |
| Auth/session | 50M logins | ~600 |

### Storage (5 years)
- **Video files**: 15K titles × 100GB avg (all encodes) ≈ 1.5 PB
- **Per-title encodes**: 5 bitrates × 3 codecs (H.264, HEVC, AV1) × regions ≈ 15 variants
- **Metadata**: 500K items × 5KB ≈ 2.5 GB
- **User data**: 250M × 10KB ≈ 2.5 TB
- **Thumbnails/artwork**: 500K × 500KB ≈ 250 GB

### Bandwidth
- **Peak egress**: 37.5M × 5 Mbps ≈ 187.5 Tbps (served by CDN)
- **Ingest (new content)**: ~100 titles/week × 100GB ≈ 10 TB/week

### Cache
- **CDN**: 95%+ of video bytes from edge (Open Connect)
- **Metadata cache**: Redis, 500K titles × 5KB ≈ 2.5 GB
- **Recommendation cache**: Pre-computed, 250M × 1KB ≈ 250 GB
- **Session cache**: 50M active × 2KB ≈ 100 GB

---

## 3. API Design

### REST Endpoints

```
# Authentication & Profiles
POST   /api/v1/auth/login
Body: { "email": "...", "password": "..." }
Response: { "session_token": "...", "profiles": [...] }

GET    /api/v1/profiles/:profile_id
PUT    /api/v1/profiles/:profile_id
Body: { "name": "...", "avatar": "...", "maturity_level": "..." }

# Catalog & Browse
GET    /api/v1/browse
Query: profile_id, region, language
Response: { "rows": [{ "title": "...", "titles": [...] }] }

GET    /api/v1/titles/:title_id
Response: { "title_id", "name", "synopsis", "seasons", "episodes", "artwork", "rating" }

GET    /api/v1/search
Query: q=query, type=movie|series|both, limit=20
Response: { "results": [...] }

# Playback
POST   /api/v1/playback/start
Body: { "profile_id", "title_id", "episode_id", "position_ms" }
Response: { "playback_id", "manifest_url", "license_url" }

PUT    /api/v1/playback/:playback_id/position
Body: { "position_ms" }

POST   /api/v1/playback/:playback_id/stop

GET    /api/v1/playback/state
Query: profile_id
Response: { "continue_watching": [...], "recently_watched": [...] }

# Recommendations
GET    /api/v1/recommendations
Query: profile_id, category=home|similar|trending
Response: { "rows": [{ "title": "...", "titles": [...] }] }

# Downloads (Offline)
POST   /api/v1/downloads
Body: { "profile_id", "title_id", "episode_id", "quality" }
Response: { "download_id", "manifest_url", "license" }

GET    /api/v1/downloads
DELETE /api/v1/downloads/:download_id
```

---

## 4. Data Model / Database Schema

### Database Choice
- **Titles/Catalog**: Cassandra (read-heavy, partition by title_id)
- **User accounts/profiles**: MySQL (ACID for billing)
- **Playback state**: Cassandra (partition by profile_id)
- **Recommendations**: Pre-computed in Redis + Cassandra
- **Search**: Elasticsearch
- **Analytics**: Data warehouse (Snowflake/BigQuery)

### Schema

**Titles (Cassandra)**
```sql
titles_by_id (
  title_id UUID PRIMARY KEY,
  type VARCHAR(20),           -- movie, series
  name VARCHAR(200),
  synopsis TEXT,
  release_year INT,
  rating VARCHAR(10),
  duration_minutes INT,
  artwork_url VARCHAR(500),
  created_at TIMESTAMP
)

episodes_by_series (
  series_id UUID,
  season_number INT,
  episode_number INT,
  episode_id UUID,
  name VARCHAR(200),
  duration_minutes INT,
  PRIMARY KEY (series_id, season_number, episode_number)
)
```

**Users (MySQL)**
```sql
accounts (
  account_id BIGINT PRIMARY KEY,
  email VARCHAR(255) UNIQUE,
  password_hash VARCHAR(255),
  plan VARCHAR(20),
  status VARCHAR(20),
  created_at TIMESTAMP
)

profiles (
  profile_id BIGINT PRIMARY KEY,
  account_id BIGINT,
  name VARCHAR(50),
  avatar_url VARCHAR(500),
  maturity_level VARCHAR(20),
  created_at TIMESTAMP
)
```

**Playback State (Cassandra)**
```sql
playback_state_by_profile (
  profile_id BIGINT,
  title_id UUID,
  episode_id UUID,
  position_ms BIGINT,
  updated_at TIMESTAMP,
  PRIMARY KEY (profile_id, title_id, episode_id)
)
```

**Recommendations (Redis + Cassandra)**
```sql
recommendations_by_profile (
  profile_id BIGINT,
  category VARCHAR(50),
  title_ids LIST<UUID>,
  updated_at TIMESTAMP,
  PRIMARY KEY (profile_id, category)
)
```

---

## 5. High-Level Architecture

### ASCII Architecture Diagram

```
                                    ┌─────────────────────────────────────────────────────────────┐
                                    │                        CLIENTS                               │
                                    │  (Smart TV, Mobile, Web, Gaming Console)                     │
                                    └───────────────────────────┬─────────────────────────────────┘
                                                                  │
                                                                  │ HTTPS
                                                                  ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                         AWS CLOUD (us-east-1, eu-west-1, ap-southeast-1)                       │
│                                                                                                               │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐                     │
│  │   API Gateway   │    │  Auth Service   │    │ Catalog Service │    │ Playback Svc    │                     │
│  │   (Zuul/Kong)   │───▶│  (OAuth/JWT)    │    │  (Metadata)     │    │ (Manifest Gen)  │                     │
│  └────────┬────────┘    └────────┬────────┘    └────────┬────────┘    └────────┬────────┘                     │
│           │                      │                      │                      │                              │
│           │                      ▼                      ▼                      ▼                              │
│           │              ┌──────────────────────────────────────────────────────────────┐                      │
│           │              │              Recommendation Engine (Personalization)          │                      │
│           │              │  (Collaborative Filtering + Content-Based + Contextual)        │                      │
│           │              └──────────────────────────────────────────────────────────────┘                      │
│           │                                              │                                                      │
│           ▼                                              ▼                                                      │
│  ┌─────────────────────────────────────────────────────────────────────────────────────────────────────────┐  │
│  │                    DATA LAYER                                                                             │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐                     │  │
│  │  │ Cassandra│  │  MySQL   │  │  Redis   │  │Elasticsearch│ │  Kafka   │  │   S3     │                     │  │
│  │  │(Catalog, │  │(Accounts)│  │ (Cache,  │  │ (Search)  │  │ (Events) │  │(Metadata)│                     │  │
│  │  │ Playback)│  │          │  │  Recs)   │  │           │  │          │  │          │                     │  │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘  └──────────┘  └──────────┘                     │  │
│  └─────────────────────────────────────────────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
                                                                  │
                                                                  │ Manifest URL (CDN)
                                                                  ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                              NETFLIX OPEN CONNECT (Custom CDN)                                                │
│                                                                                                               │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐                        │
│  │  OC Appliances  │  │  OC Appliances  │  │  OC Appliances  │  │  OC Appliances  │  ... (ISPs worldwide)   │
│  │  (ISP A)        │  │  (ISP B)        │  │  (ISP C)        │  │  (ISP D)        │                        │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘                        │
│           │                    │                    │                    │                                   │
│           └────────────────────┴────────────────────┴────────────────────┘                                   │
│                                        │                                                                      │
│                                        │ BGP Anycast / DNS-based routing                                     │
│                                        ▼                                                                      │
│                              ┌─────────────────┐                                                              │
│                              │  Origin Storage │                                                              │
│                              │  (AWS S3)       │                                                              │
│                              └─────────────────┘                                                              │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                         VIDEO PROCESSING PIPELINE (Ingest → Encode → Package → Distribute)                    │
│                                                                                                               │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐               │
│  │  Ingest  │───▶│ Transcode│───▶│  Encode  │───▶│ Package  │───▶│  DRM     │───▶│  Upload  │               │
│  │ (Source) │    │ (Mezz)   │    │(H.264,   │    │(HLS/DASH)│    │(Widevine,│    │ to CDN   │               │
│  │          │    │          │    │ HEVC,AV1) │    │          │    │ PlayReady)│    │          │               │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘    └──────────┘    └──────────┘               │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

### Component Descriptions
- **API Gateway**: Routes requests, rate limiting, auth validation
- **Auth Service**: OAuth 2.0, JWT, device limits, session management
- **Catalog Service**: Metadata for titles, browse rows, search
- **Playback Service**: Generates manifest URLs, tracks position, DRM license
- **Recommendation Engine**: 500+ microservices for personalization
- **Open Connect**: Netflix's custom CDN—appliances in ISP networks
- **Video Pipeline**: Ingest → transcode → multi-bitrate encode → package → DRM → CDN

---

## 6. Detailed Component Design

### 6.1 Video Processing Pipeline

**Ingest**
- Content arrives via Aspera/Signiant (high-speed transfer) or physical drives
- Source files stored in S3 (mezzanine quality)
- Metadata extracted: duration, resolution, frame rate, audio tracks

**Transcode & Encode**
- **Mezzanine**: Master transcode to intermediate format (ProRes, DNxHD)
- **Multi-bitrate encoding**: 5-7 quality levels (240p to 4K)
  - 240p, 360p, 480p, 720p, 1080p, 4K
- **Multi-codec**: H.264 (compatibility), HEVC (efficiency), AV1 (future)
- **Per-region**: Different encodes for different licensing/availability
- **Parallel processing**: Each title split into chunks, encoded on GPU clusters

**Package**
- **HLS (HTTP Live Streaming)**: Apple devices, universal
- **DASH (Dynamic Adaptive Streaming)**: Android, web
- **Manifest files**: M3U8 (HLS) or MPD (DASH) with segment URLs
- **Segment duration**: 2-6 seconds per segment

**DRM (Digital Rights Management)**
- **Widevine**: Android, Chrome
- **PlayReady**: Windows, Xbox
- **FairPlay**: Apple devices
- **License server**: Issues time-limited decryption keys per playback session

**Distribution**
- Encoded packages uploaded to S3 (origin)
- Open Connect appliances pull popular content via Bittorrent-like P2P
- Pre-positioning: New releases pushed to edge before launch

### 6.2 Adaptive Bitrate Streaming (ABR)

- **Client-side logic**: Client measures bandwidth, requests next segment at appropriate quality
- **Manifest**: Contains URLs for each quality level's segments
- **Switching**: Can switch quality mid-stream (no re-buffering)
- **Startup**: Begin with lowest quality for fast start, ramp up

### 6.3 Open Connect CDN

- **Appliances**: Custom servers (Netflix Open Connect Appliance) placed in ISP data centers
- **Benefits**: Reduced transit costs, lower latency, higher throughput
- **Content placement**: Popular content cached; long-tail from origin
- **BGP Anycast**: Same IP announced from multiple locations; routing picks nearest
- **Pre-positioning**: "Popular in your region" pushed proactively

### 6.4 Recommendation Engine

**Collaborative Filtering**
- User-item matrix: which users watched which titles
- Matrix factorization (SVD, ALS) for latent factors
- "Users who watched X also watched Y"

**Content-Based**
- Metadata: genre, cast, director, keywords
- Embeddings from synopsis (NLP)
- "Similar to X" based on attributes

**Contextual**
- Time of day, device type, day of week
- "Trending now", "New releases"

**Hybrid**
- Ensemble of models, A/B tested
- Real-time ranking with candidate generation + scoring

### 6.5 Microservices Architecture

- **500+ microservices**: Bounded contexts (catalog, playback, billing, etc.)
- **Event-driven**: Kafka for async communication (playback events, recommendations)
- **Service mesh**: Istio/Envoy for traffic management, observability
- **Multi-region**: Active-active in us-east-1, eu-west-1, ap-southeast-1

### 6.6 User Profile & Playback Tracking

- **Playback events**: Position, pause, seek, quality switches → Kafka
- **Analytics**: Aggregated for recommendations, quality metrics
- **Continue watching**: Last position per (profile, title, episode) in Cassandra
- **Cross-device sync**: Playback state replicated across regions

---

## 7. Scaling

### Sharding
- **Cassandra**: Partition by profile_id (playback), title_id (catalog)
- **MySQL**: Shard accounts by account_id range
- **Elasticsearch**: Shards by title_id hash

### Caching
- **CDN**: 95%+ video bytes from edge
- **Redis**: Catalog metadata, recommendations, session
- **Cache-aside**: Catalog service checks Redis before Cassandra
- **Pre-compute**: Recommendations batch-computed nightly, served from cache

### CDN
- **Open Connect**: Primary for video
- **CloudFront**: Fallback for metadata, API (non-video)
- **Edge locations**: 1000+ globally

### Database
- **Cassandra**: Multi-DC replication, tunable consistency
- **Read replicas**: MySQL read replicas for browse-heavy operations
- **Connection pooling**: PgBouncer/ProxySQL for MySQL

---

## 8. Failure Handling

### Component Failures
- **API Gateway**: Multiple instances behind load balancer, health checks
- **Catalog Service**: Circuit breaker to Cassandra; fallback to cached data
- **Playback Service**: Stateless; any instance can serve
- **Recommendation Service**: Fallback to "trending" if personalization fails

### Redundancy
- **Multi-region active-active**: Traffic can route to any region
- **Cassandra**: RF=3, read/write at QUORUM
- **Open Connect**: If appliance fails, traffic routes to next nearest
- **Origin**: S3 cross-region replication for video origin

### Degradation
- **Recommendation failure**: Show generic browse rows
- **Search failure**: Show "Popular" as fallback
- **Playback**: CDN redundancy; license server fallback

### DRM
- **License server**: Multi-region, failover
- **Offline**: Licenses cached with time-bound validity; refresh on next online

---

## 9. Monitoring & Observability

### Key Metrics
- **Playback**: Start latency (p50, p99), buffering ratio, bitrate distribution
- **API**: Latency, error rate, QPS per endpoint
- **CDN**: Cache hit ratio, egress bandwidth, origin load
- **Recommendations**: Click-through rate, watch time from recommendations
- **Encoding pipeline**: Job completion rate, queue depth, encode time

### Alerts
- **Playback start > 3s** (p99)
- **Error rate > 0.1%**
- **CDN cache hit < 90%**
- **License server errors**

### Tracing
- **Distributed tracing**: Trace ID from API → Playback → CDN
- **Correlation**: Link playback events to user session

### Logging
- **Structured logs**: JSON, searchable (Elasticsearch)
- **PII**: Redact in logs; separate analytics pipeline

---

## 10. Interview Tips

### Follow-up Questions
- "How would you handle a viral new release (e.g., Wednesday) that causes 10x traffic spike?"
- "How does Netflix decide what to pre-position on Open Connect appliances?"
- "How would you design the 'Skip Intro' feature?"
- "How do you prevent account sharing at scale?"
- "How would you add live streaming to this architecture?"

### Common Mistakes
- **Over-engineering CDN**: Don't need to design Open Connect from scratch; mention it exists, focus on integration
- **Ignoring DRM**: Critical for studios; must be part of playback flow
- **Single region**: Netflix is global; multi-region is mandatory
- **Ignoring encoding**: Video doesn't stream raw; pipeline is core
- **Recommendations as afterthought**: It's a major differentiator; dedicate design time

### Key Points to Emphasize
- **Video pipeline**: Ingest → encode (multi-bitrate, multi-codec) → package (HLS/DASH) → DRM → CDN
- **Open Connect**: Custom CDN in ISP networks; reduces cost, improves QoE
- **ABR**: Client-driven quality selection; manifest with multiple bitrate URLs
- **Recommendations**: Collaborative + content-based + contextual; pre-computed + real-time ranking
- **Microservices**: 500+ services; event-driven; multi-region active-active

---

## Appendix: Deep Dive Topics

### A. HLS vs DASH Comparison
| Aspect | HLS | DASH |
|--------|-----|------|
| Origin | Apple | MPEG | 
| Container | MPEG-TS / fMP4 | fMP4 |
| Manifest | M3U8 | MPD (XML) |
| Browser support | Safari, Chrome | Chrome, Firefox, Edge |
| Adaptive | Yes | Yes |

### B. Encoding Bitrate Ladder (Typical)
- 240p: 0.5 Mbps
- 360p: 1 Mbps
- 480p: 2 Mbps
- 720p: 3-5 Mbps
- 1080p: 5-8 Mbps
- 4K: 15-25 Mbps

### C. Open Connect Appliance Specs (Conceptual)
- Storage: 100+ TB per appliance
- Network: 10-100 Gbps
- Placement: ISP data centers, IXPs
- Content: Popular titles pre-positioned; long-tail from origin

### D. Recommendation Pipeline (Simplified)
1. **Candidate generation**: Collaborative filtering (top 1000)
2. **Ranking**: Neural network scores each candidate
3. **Diversity**: Re-rank to avoid homogeneity
4. **Business rules**: Exclude watched, apply maturity filters
5. **Response**: Return top 50-100

### E. Offline Download Flow
1. User selects title for download
2. Client requests encrypted segments + license
3. License server issues time-bound key (e.g., 30 days)
4. Client stores encrypted segments locally
5. Playback: Decrypt with cached key
6. Refresh: When online, extend license

### F. Multi-Region Data Replication
- **Cassandra**: Multi-DC replication; each region has full copy
- **MySQL**: Cross-region async replication; read from local replica
- **Redis**: Redis Enterprise or custom replication
- **S3**: Cross-region replication for video origin
- **Failover**: DNS/Route53 health checks; route to healthy region

### G. Content Pre-Positioning Strategy
- **New releases**: Push to all Open Connect appliances 24-48h before launch
- **Trending**: Monitor view velocity; push when threshold exceeded
- **Regional**: "Popular in your country" pushed to local appliances
- **Storage limits**: LRU eviction when appliance full; keep most popular

### H. Playback Quality Metrics
- **Rebuffer ratio**: Time buffering / total watch time (target < 0.5%)
- **Start time**: Time to first frame (target < 2s p99)
- **Bitrate distribution**: % of time at each quality level
- **Abandonment**: % of sessions that stop in first 30s

### I. Device Limit Enforcement
- **Concurrent streams**: Plan-based (e.g., 2 for Standard, 4 for Premium)
- **Tracking**: Redis (device_id, profile_id) with TTL on stream start
- **On new stream**: Check count; reject if over limit
- **On stream end**: Decrement count (heartbeat or explicit stop)

### J. Skip Intro/Credits Implementation
- **Metadata**: Per-episode timestamps (intro_start, intro_end, credits_start)
- **Client**: Show "Skip Intro" button when position in range
- **Seek**: On tap, seek to intro_end
- **Storage**: In episode metadata; ~10 bytes per episode

### K. Account Sharing Detection (Conceptual)
- **Signals**: Login patterns, device diversity, concurrent streams
- **ML**: Anomaly detection; flag suspicious accounts
- **Action**: Prompt for verification; limit devices
- **Note**: Balance UX vs revenue; no perfect solution

### L. Viral Release Handling (e.g., Wednesday 10x Traffic)
- **Pre-positioning**: Push to all Open Connect before launch
- **Auto-scale**: API and backend scale on demand (Kubernetes HPA)
- **Queue**: If origin overloaded, queue requests; don't fail
- **Degradation**: Lower default bitrate to reduce bandwidth
- **Monitoring**: Real-time dashboards; on-call ready

### M. Service Mesh Benefits (Istio/Envoy)
- **Traffic management**: Canary, blue-green, circuit breaker
- **Observability**: Automatic tracing, metrics per service
- **Security**: mTLS between services
- **Resilience**: Retry, timeout, fault injection

### N. Event-Driven Playback Pipeline
- **Events**: play_start, play_progress, play_complete, quality_change
- **Producer**: Client or playback service
- **Consumer**: Analytics, recommendations, billing
- **Kafka**: Partition by user_id for ordering; retention 7 days

### O. Live Streaming Addition (Future)
- **Ingest**: RTMP/RTMPS from broadcast source
- **Transcode**: Real-time encoding (lower latency than VOD)
- **Package**: LL-HLS or low-latency DASH
- **CDN**: Same Open Connect; live segments have short TTL
- **DRM**: Same license server; live keys

### P. A/B Testing for Recommendations
- **Experiment**: New ranking model vs control
- **Assignment**: User hash → experiment bucket
- **Metrics**: CTR, watch time, session length
- **Statistical significance**: Run until confidence interval

### Q. Billing Integration
- **Plan**: Basic, Standard, Premium (streams, resolution)
- **Usage**: Playback events → billing service
- **Enforcement**: Device limit at playback start
- **Upgrade/downgrade**: Immediate or next cycle

### R. Search Architecture
- **Index**: Elasticsearch; title, cast, description, genre
- **Typo tolerance**: Fuzzy matching, edit distance
- **Facets**: Filter by type (movie/series), year, genre
- **Personalization**: Boost recently watched, similar to liked

### S. Profile Switching
- **Storage**: Multiple profiles per account in MySQL
- **Session**: JWT includes profile_id; switch updates token
- **Playback state**: Partitioned by profile_id in Cassandra
- **Recommendations**: Per profile; different watch history

### T. Thumbnail Generation
- **Source**: Extract frame at 10%, 50%, 90% of video
- **Processing**: Resize, compress; store in S3
- **CDN**: Serve from edge like video
- **A/B**: Test which thumbnail gets more clicks

### U. Rate Limiting
- **API**: Per user, per IP; 1000 req/min typical
- **Playback**: No limit (streaming is core)
- **Auth**: Stricter on login (prevent brute force)

### V. Content Delivery Security
- **Signed URLs**: Time-limited CDN URLs; prevent hotlinking
- **Token**: JWT in URL; validated at edge
