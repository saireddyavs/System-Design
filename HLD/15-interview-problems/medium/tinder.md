# Design Tinder

## 1. Problem Statement & Requirements

### Problem Statement
Design a dating app like Tinder that enables users to create profiles with photos, discover nearby users based on location and preferences, swipe left/right to express interest, get notified of mutual matches, and chat with matches.

### Functional Requirements
- **User profiles**: Photos, bio, age, gender, interests
- **Location-based matching**: Show users within X km
- **Preferences**: Filter by age range, distance, gender
- **Swipe**: Swipe right (like) or left (pass)
- **Mutual match**: When both swipe right; notification
- **Chat**: Message after match
- **Unmatch**: Remove match, stop chat

### Non-Functional Requirements
- **Scale**: 75M MAU, 2B swipes per day
- **Latency**: Profile load < 200ms, swipe < 100ms
- **Availability**: 99.9%
- **Real-time**: Match notification < 2 seconds

### Out of Scope
- Video profiles
- Super Like / Boost (premium features)
- Block/report (assume simple implementation)
- Verification (blue check)

---

## 2. Back-of-Envelope Estimation

### Scale Assumptions
- **Users**: 75M MAU, 10M DAU
- **Swipes/day**: 2B = ~23,000 QPS average, 100K QPS peak
- **Profiles**: 75M; each has 5-6 photos
- **Matches/day**: 50M (2% of swipes)
- **Messages/day**: 500M
- **Avg profile size**: 5 photos × 200KB = 1MB, metadata 5KB

### QPS Calculation
| Operation | Daily Volume | Peak QPS |
|-----------|--------------|----------|
| Swipe (left/right) | 2B | ~25,000 |
| Profile fetch (discovery) | 500M | ~6,000 |
| Match check | 2B | ~25,000 |
| Chat send | 500M | ~6,000 |
| Chat read | 1B | ~12,000 |
| Photo upload | 5M | ~60 |
| Location update | 100M | ~1,200 |

### Storage (5 years)
- **Profiles**: 75M × 5KB ≈ 375 GB
- **Photos**: 75M × 5 × 200KB ≈ 75 TB
- **Swipes**: 2B × 365 × 5 × 50B ≈ 180 TB (who swiped whom)
- **Matches**: 50M × 365 × 5 × 100B ≈ 90 TB
- **Messages**: 500M × 365 × 5 × 200B ≈ 1.8 PB
- **Indexes**: 2x for queries

### Bandwidth
- **Photo serving**: 500M profile views × 1MB ≈ 500 TB/day
- **Chat**: 500M × 500B ≈ 250 GB/day
- **Swipe writes**: 2B × 100B ≈ 200 GB/day

### Cache
- **Profile cache**: Redis, 10M active × 10KB ≈ 100 GB
- **Swipe cache**: Bloom filter or Redis (already swiped) - 2B × 1 bit ≈ 250 MB
- **Discovery stack**: Pre-computed "next N profiles" per user - 10M × 50 × 1KB ≈ 500 GB

---

## 3. API Design

### REST Endpoints

```
# Auth
POST   /api/v1/auth/login
Body: { "phone" | "email", "otp" } or OAuth (Facebook, Apple)
Response: { "access_token", "user_id", "expires_in" }

POST   /api/v1/auth/verify
Body: { "phone", "otp" }
Response: { "access_token" }

# Profile
GET    /api/v1/me
Response: { "user_id", "name", "bio", "birth_date", "gender", "photos", "preferences" }

PUT    /api/v1/me
Body: { "name", "bio", "photos", "preferences" }

POST   /api/v1/me/photos
Body: multipart/form-data (image file)
Response: { "photo_id", "url" }

DELETE /api/v1/me/photos/:photo_id

# Location
PUT    /api/v1/me/location
Body: { "lat": 37.77, "lng": -122.42 }
Response: { "updated": true }

# Discovery (Get next profiles to show)
GET    /api/v1/discovery
Query: limit=10
Response: { "profiles": [{ "user_id", "name", "age", "photos", "distance_km" }] }

# Swipe
POST   /api/v1/swipes
Body: { "target_user_id": "xxx", "direction": "left" | "right" }
Response: { "swipe_id", "is_match": false }

# When is_match=true, also return match info
Response: { "swipe_id", "is_match": true, "match_id": "m_xxx" }

# Matches
GET    /api/v1/matches
Query: limit=20, cursor=
Response: { "matches": [{ "match_id", "user", "last_message", "created_at" }], "next_cursor": "..." }

GET    /api/v1/matches/:match_id
DELETE /api/v1/matches/:match_id   # Unmatch

# Chat
GET    /api/v1/matches/:match_id/messages
Query: limit=50, before=message_id
Response: { "messages": [...], "has_more": true }

POST   /api/v1/matches/:match_id/messages
Body: { "text": "Hello!" }
Response: { "message_id", "created_at" }

# Real-time (WebSocket)
WS     /api/v1/ws
# Events: match, new_message, typing
```

---

## 4. Data Model / Database Schema

### Database Choice
- **Users/Profiles**: PostgreSQL (ACID, complex queries)
- **Swipes**: Cassandra (write-heavy, partition by user_id)
- **Matches**: Cassandra (partition by user_id)
- **Messages**: Cassandra (partition by match_id)
- **Location index**: Redis (Geohash) or PostgreSQL (PostGIS)
- **Photos**: S3 + CDN
- **Search/Discovery**: Elasticsearch or custom (geohash + filters)

### Schema

**Users (PostgreSQL)**
```sql
users (
  user_id UUID PRIMARY KEY,
  phone VARCHAR(20) UNIQUE,
  email VARCHAR(255),
  name VARCHAR(100),
  bio TEXT,
  birth_date DATE,
  gender VARCHAR(20),
  last_active_at TIMESTAMP,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
)

user_photos (
  photo_id UUID PRIMARY KEY,
  user_id UUID,
  url VARCHAR(500),
  position INT,
  created_at TIMESTAMP
)

user_preferences (
  user_id UUID PRIMARY KEY,
  age_min INT,
  age_max INT,
  max_distance_km INT,
  gender_preference VARCHAR(20)
)
```

**Swipes (Cassandra)**
```sql
swipes_by_user (
  user_id UUID,
  target_user_id UUID,
  direction VARCHAR(10),    -- left, right
  created_at TIMESTAMP,
  PRIMARY KEY (user_id, target_user_id)
)
-- Query: Get users I've swiped on (for "already seen")
-- Query: Get users who swiped right on me (for match check)

swipes_by_target (
  target_user_id UUID,
  user_id UUID,
  direction VARCHAR(10),
  created_at TIMESTAMP,
  PRIMARY KEY (target_user_id, user_id)
)
-- Query: Who swiped on me? For match detection when I swipe right
```

**Matches (Cassandra)**
```sql
matches_by_user (
  user_id UUID,
  match_id UUID,
  other_user_id UUID,
  created_at TIMESTAMP,
  last_message_at TIMESTAMP,
  PRIMARY KEY (user_id, match_id)
)
-- Both users have row (symmetric)

matches_by_id (
  match_id UUID PRIMARY KEY,
  user1_id UUID,
  user2_id UUID,
  created_at TIMESTAMP
)
```

**Messages (Cassandra)**
```sql
messages_by_match (
  match_id UUID,
  message_id TIMEUUID,
  sender_id UUID,
  text TEXT,
  created_at TIMESTAMP,
  PRIMARY KEY (match_id, message_id)
) WITH CLUSTERING ORDER BY (message_id DESC)
```

**Location (Redis or PostgreSQL)**
```sql
-- Redis: user_id -> geohash
user_location:{user_id} = "9q8yy"  -- Geohash
-- Or PostgreSQL with PostGIS
user_locations (
  user_id UUID PRIMARY KEY,
  lat DOUBLE PRECISION,
  lng DOUBLE PRECISION,
  geohash VARCHAR(12),
  updated_at TIMESTAMP
)
CREATE INDEX idx_geohash ON user_locations(geohash);
```

---

## 5. High-Level Architecture

### ASCII Architecture Diagram

```
                                    ┌─────────────────────────────────────────────────────────────┐
                                    │                    MOBILE / WEB CLIENTS                       │
                                    │  (iOS, Android, Web)                                          │
                                    └───────────────────────────┬─────────────────────────────────┘
                                                                  │
                                                                  │ HTTPS / WebSocket
                                                                  ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                         API GATEWAY / LOAD BALANCER                                           │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
                                                                  │
                                                                  ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                         MICROSERVICES                                                          │
│                                                                                                               │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐            │
│  │   Auth      │ │   Profile   │ │  Discovery  │ │   Swipe     │ │   Match     │ │   Chat      │            │
│  │   Service   │ │   Service   │ │  Service   │ │  Service   │ │  Service   │ │  Service   │            │
│  └──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └──────┬──────┘            │
│         │              │              │              │              │              │                       │
│         │              │              │              │              │              │                       │
│         ▼              ▼              ▼              ▼              ▼              ▼                       │
│  ┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐    │
│  │                    LOCATION SERVICE (Geohash / Quadtree)                                            │    │
│  │  - Find users within X km                                                                            │    │
│  │  - Filter by preferences (age, gender)                                                               │    │
│  │  - Exclude already swiped                                                                            │    │
│  └─────────────────────────────────────────────────────────────────────────────────────────────────────┘    │
│                                                                                                               │
│  ┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐    │
│  │                    REAL-TIME SERVICE (WebSocket / Push)                                              │    │
│  │  - Match notifications                                                                               │    │
│  │  - New message notifications                                                                         │    │
│  │  - Typing indicators                                                                                 │    │
│  └─────────────────────────────────────────────────────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
                                                                  │
                                                                  ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                         DATA LAYER                                                            │
│                                                                                                               │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐                         │
│  │PostgreSQL│  │ Cassandra│  │  Redis   │  │Elasticsearch│ │   S3     │  │  Kafka   │                         │
│  │ (Users)  │  │(Swipes,  │  │ (Cache,  │  │ (Search   │  │ (Photos) │  │ (Events) │                         │
│  │          │  │ Matches, │  │ Location,│  │  optional)│  │          │  │          │                         │
│  │          │  │ Messages)│  │  Match   │  │           │  │          │  │          │                         │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘  └──────────┘  └──────────┘                         │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
                                                                  │
                                                                  │ Photo URLs
                                                                  ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                         CDN (CloudFront)                                                      │
│                                    Profile photos, thumbnails                                                 │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                    SWIPE → MATCH FLOW                                                         │
│                                                                                                               │
│  User A                Swipe Service              Match Check                User B (WebSocket)                 │
│     │                        │                        │                           │                            │
│     │  POST /swipes           │                        │                           │                            │
│     │  { target: B, right }  │                        │                           │                            │
│     │───────────────────────▶                        │                           │                            │
│     │                        │  Store swipe           │                           │                            │
│     │                        │  (A swiped B, right)    │                           │                            │
│     │                        │────────────────────────│                           │                            │
│     │                        │                        │                           │                            │
│     │                        │  Check: Did B swipe    │                           │                            │
│     │                        │  right on A?            │                           │                            │
│     │                        │───────────────────────▶                            │                            │
│     │                        │                        │                           │                            │
│     │                        │  Yes: Create match     │                           │                            │
│     │                        │  Notify both           │                           │                            │
│     │                        │◀───────────────────────│                           │                            │
│     │                        │                        │  Push: "It's a match!"    │                            │
│     │                        │                        │──────────────────────────▶                            │
│     │  200 { is_match: true } │                        │                           │                            │
│     │◀───────────────────────                        │                           │                            │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

### Component Descriptions
- **Auth Service**: OTP, OAuth, session
- **Profile Service**: CRUD profile, photos (S3)
- **Discovery Service**: Get next profiles (location + preferences + not swiped)
- **Swipe Service**: Store swipe, check match, create match
- **Match Service**: List matches, unmatch
- **Chat Service**: Send/receive messages
- **Location Service**: Geohash-based proximity search
- **Real-time**: WebSocket for match alerts, chat

---

## 6. Detailed Component Design

### 6.1 Location Service (Proximity Search)

**Geohashing**
- Encode (lat, lng) into string (e.g., "9q8yy")
- Prefix match: Same prefix = nearby
- Precision: 5 chars ≈ 5km, 6 chars ≈ 1.2km, 7 chars ≈ 150m
- **Query**: Get geohash of user; fetch users in same and adjacent geohash cells
- **Filter**: Distance < max_distance_km (Haversine formula)

**Quadtree**
- Divide world into quadrants recursively
- Store user_id in leaf cells
- Query: Traverse tree for user's cell; get neighbors
- Alternative: Use PostGIS `ST_DWithin` for "within X km"

**Implementation**
- **Redis**: Geo commands (GEOADD, GEORADIUS) - stores (user_id, lat, lng)
- **PostgreSQL**: PostGIS extension, spatial index
- **Cassandra**: Geohash as partition key; range query on geohash prefix

**Update**
- Location updated on app open, or periodically (every 5 min)
- Don't need real-time; 100m precision sufficient

### 6.2 Matching Algorithm (Who to Show)

**Input**
- User's location, preferences (age, distance, gender)
- Exclude: Already swiped (left or right)
- Exclude: Already matched
- Exclude: Blocked users

**Ranking**
- **ELO-like score**: Attractiveness score (implicit from right swipe rate)
- **Recency**: Active users first
- **Distance**: Closer first
- **Diversity**: Don't show same "type" repeatedly

**Pre-computation**
- For each user, maintain "discovery stack" of next 50-100 profiles
- Background job: Refresh stack when low
- On swipe: Remove from stack; add next if needed

**Real-time**
- Simpler: Query on each discovery request
- Location + filters + NOT IN (swiped_ids)
- Paginate: Cursor-based, limit 10

### 6.3 Swipe Storage

**Schema**
- `swipes_by_user(user_id, target_user_id) -> direction`
- For match check: `swipes_by_target(target_user_id, user_id) -> direction`
- When A swipes right on B: Check if B has swiped right on A
- If yes: Create match; notify both

**Write path**
- A swipes B: Write to swipes_by_user(A, B), swipes_by_target(B, A)
- Single write with two tables (batch or async)

**Read path**
- Discovery: Get users in range; filter out swiped (check swipes_by_user)
- Match check: On A's right swipe on B, read swipes_by_target(B, A)

### 6.4 Match Detection

- **Trigger**: User A swipes right on B
- **Check**: SELECT direction FROM swipes_by_target WHERE target_user_id=B AND user_id=A
- **If direction='right'**: Match! Create match record; notify both
- **Match record**: matches_by_user (both users), matches_by_id
- **Notification**: Push to A (immediate), Push to B (if online via WebSocket)

### 6.5 Recommendation / Ranking (ELO-like)

- **Implicit score**: Users with high right-swipe rate are "attractive"
- **Update**: When user gets right swipe, increase score; left swipe, decrease
- **Use**: Rank discovery stack by score (show "hot" profiles more)
- **Fairness**: Don't only show top 10%; mix in variety
- **Cold start**: New users get random ordering initially

### 6.6 Media Storage (Photos)

- **Upload**: Client → API → S3 (presigned URL or direct upload)
- **Processing**: Resize, thumbnail; store multiple sizes
- **CDN**: CloudFront; serve from edge
- **URL**: https://cdn.tinder.com/photos/{user_id}/{photo_id}.jpg

### 6.7 Real-Time Notifications

**WebSocket**
- Connection: User connects to WebSocket server with auth
- Events: match, new_message, typing
- Scale: WebSocket servers behind load balancer; sticky sessions
- **Redis Pub/Sub**: When match created, publish to channels user_A, user_B; WebSocket servers subscribe and push to connected clients

**Push Notifications**
- When user offline: Send push (FCM, APNs)
- Match: "You have a new match!"
- Message: "John sent you a message"

### 6.8 Chat Service

- **Send**: POST message; store in Cassandra; publish to WebSocket/Push
- **Read**: GET messages; cursor-based pagination
- **Real-time**: WebSocket delivers new messages to online user
- **Offline**: Push notification; fetch on next open
- **Unmatch**: Delete match; hide chat; optionally soft-delete messages

---

## 7. Scaling

### Sharding
- **Swipes**: Partition by user_id (Cassandra)
- **Matches**: Partition by user_id
- **Messages**: Partition by match_id
- **Location**: Shard by geohash region

### Caching
- **Profile**: Redis; 10M active users
- **Discovery stack**: Pre-computed in Redis (next 50 profiles per user)
- **Swipe check**: Bloom filter "already swiped" to avoid DB hit
- **Photos**: CDN; 95% from edge

### CDN
- **Photos**: CloudFront; global edge
- **Thumbnails**: Smaller size for discovery cards

### Database
- **Cassandra**: Multi-DC, RF=3
- **PostgreSQL**: Read replicas for profile reads
- **Redis Cluster**: For location, cache

### WebSocket
- **Sticky sessions**: Route user to same server
- **Horizontal**: Add WebSocket servers; Redis Pub/Sub for cross-server messaging

---

## 8. Failure Handling

### Component Failures
- **Discovery down**: Return cached stack or "try again"
- **Swipe down**: Queue; process async; eventual consistency
- **Match notification**: Retry push; store for later
- **Chat**: Queue messages; deliver when service up

### Redundancy
- **Multi-region**: Active-active for API; data replicated
- **Cassandra**: RF=3
- **WebSocket**: Multiple servers; Redis for pub/sub
- **S3**: Cross-region replication for photos

### Degradation
- **Location stale**: Use last known; acceptable for discovery
- **Real-time delayed**: Fallback to push; user gets notification on next open
- **Discovery slow**: Return fewer profiles; paginate

### Edge Cases
- **Double swipe**: Idempotency; same swipe_id returns same result
- **Race**: A and B swipe right simultaneously; one match creation wins; both get notified

---

## 9. Monitoring & Observability

### Key Metrics
- **Swipe**: Latency (p50, p99), QPS, success rate
- **Discovery**: Latency, profiles returned, cache hit rate
- **Match**: Match rate (right swipes → matches)
- **Chat**: Message latency, delivery rate
- **WebSocket**: Connections, message delivery latency
- **Location**: Update frequency, query latency

### Alerts
- **Swipe latency > 500ms**
- **Discovery returns 0** (bug)
- **Match notification failure > 1%**
- **WebSocket disconnect rate > 5%**

### Tracing
- **Trace ID**: Swipe → Match check → Notification
- **User journey**: Profile view → Swipe → Match → Chat

### Analytics
- **Swipe rate**: Left vs right
- **Match rate**: By cohort, by region
- **Chat**: Messages per match, response time

---

## 10. Interview Tips

### Follow-up Questions
- "How would you prevent abuse (e.g., bots swiping right on everyone)?"
- "How do you handle users in sparse areas (few people nearby)?"
- "How would you add 'Super Like' (priority in other's stack)?"
- "How do you handle timezone for 'active now'?"
- "How would you design the 'Rewind' feature (undo last swipe)?"

### Common Mistakes
- **Naive location query**: Full table scan; need geohash or spatial index
- **Showing already swiped**: Must exclude; use swipes_by_user
- **Match check wrong**: Check B's swipe on A when A swipes B
- **Synchronous notification**: Block swipe response; use async
- **Real-time scale**: WebSocket doesn't scale; need Redis Pub/Sub for multi-server

### Key Points to Emphasize
- **Location**: Geohash or PostGIS for proximity; not full scan
- **Match detection**: When A swipes right on B, check if B swiped right on A
- **Discovery**: Location + preferences + exclude swiped + ranking
- **Swipe volume**: 2B/day; write-heavy; Cassandra
- **Real-time**: WebSocket + Redis Pub/Sub for match alerts
- **Pre-compute**: Discovery stack to reduce latency

---

## Appendix: Deep Dive Topics

### A. Geohash Precision Reference
| Precision | Cell size (lat) | Cell size (lng) | Use case |
|-----------|-----------------|-----------------|----------|
| 4 | ±0.02° (~2km) | ±0.02° | City-level |
| 5 | ±0.005° (~500m) | ±0.005° | Neighborhood |
| 6 | ±0.001° (~100m) | ±0.001° | Street |
| 7 | ±0.0001° (~10m) | ±0.0001° | Building |

### B. Redis GEORADIUS Example
```
GEOADD users 37.77 -122.42 user_123
GEORADIUS users 37.77 -122.42 10 km WITHDIST ASC COUNT 50
```
Returns users within 10km, sorted by distance, limit 50.

### C. ELO Score Update (Simplified)
- When A swipes right on B: B's score += k * (1 - expected)
- When A swipes left on B: B's score += k * (0 - expected)
- expected = 1 / (1 + 10^((B_score - A_score)/400))
- k = 32 (sensitivity)

### D. WebSocket Scaling with Redis Pub/Sub
```
User A connects to WS Server 1
User B connects to WS Server 2
Match created → Publish to channel:user_A, channel:user_B
Server 1 subscribes to channel:user_A → pushes to A
Server 2 subscribes to channel:user_B → pushes to B
```
Each WS server subscribes to channels for its connected users.

### E. Discovery Stack Refresh Logic
- **Trigger**: Stack has < 10 profiles
- **Query**: Location + filters + NOT IN (swiped) + ORDER BY score
- **Batch**: Fetch 50; store in Redis with 1h TTL
- **On swipe**: Remove from stack; if stack < 10, trigger refresh
