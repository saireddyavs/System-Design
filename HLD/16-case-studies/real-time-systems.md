# Real-Time System Case Studies

---

## 1. Real-Time Gaming Leaderboard

### The Problem
A multiplayer game needs a global leaderboard that updates in real time as players earn points. Millions of players; thousands of score updates per second. Must support: get top N, get rank by user_id, update score. Low latency (< 50ms) is critical for UX.

### The Architecture/Solution

**Redis Sorted Sets**
- Data structure: `ZADD leaderboard score member`
- `ZREVRANGE 0 99` → top 100 in O(log N + M)
- `ZREVRANK leaderboard user_id` → user's rank in O(log N)
- `ZADD` updates score in O(log N)
- In-memory; sub-millisecond latency

**Sharding**
- Shard by game/season: `leaderboard:game_123:season_5`
- Or by score range for very large leaderboards
- Consistent hashing for key distribution

**Persistence**
- RDB snapshots + AOF for durability
- Async replication to replicas for read scaling

### Key Numbers
| Metric | Value |
|--------|-------|
| Latency | < 5ms (Redis) |
| Updates | 10K+ QPS per shard |
| Top N query | O(log N + M) |
| Sorted set | O(log N) add/update |

### Architecture Diagram (ASCII)

```
    ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
    │  Game       │     │  Game       │     │  Game       │
    │  Server 1   │     │  Server 2   │     │  Server N   │
    └──────┬──────┘     └──────┬──────┘     └──────┬──────┘
           │                   │                   │
           │  ZADD score       │                   │
           └───────────────────┼───────────────────┘
                               │
                    ┌──────────▼──────────┐
                    │  Redis Cluster      │
                    │  (Sharded Sorted    │
                    │   Sets)             │
                    │  leaderboard:g:s    │
                    └──────────┬──────────┘
                               │
                    ┌──────────▼──────────┐
                    │  Read Replicas      │
                    │  (Top N, Rank)      │
                    └─────────────────────┘
```

### Key Takeaways
- **Sorted sets** are purpose-built for leaderboards; avoid implementing with SQL
- **Sharding** by game/season keeps each shard manageable
- **Redis** provides the latency; DB for historical/analytics only

### Interview Tip
"For a real-time leaderboard, I'd use Redis sorted sets. ZADD for updates, ZREVRANGE for top N, ZREVRANK for user rank. All O(log N). Shard by game or season. No need for a relational DB for the hot path."

---

## 2. Real-Time Live Comments

### The Problem
A live stream has millions of viewers. Each can post comments that must appear for all viewers in real time. At peak: 100K+ comments/second. Must be low latency (< 200ms) and handle fan-out to millions of connections.

### The Architecture/Solution

**WebSocket**
- Persistent connections for real-time bidirectional communication
- One connection per viewer; server pushes new comments

**Pub/Sub**
- When user posts: publish to channel `live:stream_123`
- All connected clients subscribed to that channel receive the message
- Redis Pub/Sub or Kafka for message bus

**Fan-Out Strategies**
- **Naive**: 1M viewers = 1M deliveries per comment → unsustainable
- **Sampling**: Send to subset; or aggregate (e.g., "1.2K people said 🔥")
- **Tiered**: Full comments for recent N; aggregated for rest
- **Edge PoPs**: Regional WebSocket servers; pub/sub across regions

**Rate Limiting**
- Throttle comments per user (e.g., 1/sec)
- Server-side validation; reject excess

### Key Numbers
| Metric | Value |
|--------|-------|
| Comments/sec (peak) | 100K+ |
| Viewers (large stream) | 1M+ |
| Target latency | < 200ms |
| WebSocket connections | 1 per viewer |

### Architecture Diagram (ASCII)

```
    ┌─────────────┐                    ┌─────────────┐
    │  Client 1   │                    │  Client N   │
    │  (WebSocket)│                    │  (WebSocket)│
    └──────┬──────┘                    └──────┬──────┘
           │                                  │
           │    subscribe live:stream_123      │
           └──────────────────┬───────────────┘
                              │
                    ┌─────────▼─────────┐
                    │  WebSocket Server │
                    │  (Sticky sessions)│
                    └─────────┬─────────┘
                              │
                    ┌─────────▼─────────┐
                    │  Redis Pub/Sub    │
                    │  or Kafka         │
                    └─────────┬─────────┘
                              │
                    ┌─────────▼─────────┐
                    │  Comment API      │
                    │  PUBLISH on post  │
                    └──────────────────┘
```

### Key Takeaways
- **Pub/Sub** decouples producers (comment API) from consumers (WebSocket servers)
- **Fan-out** is the bottleneck; use aggregation or sampling for huge audiences
- **WebSocket** for low-latency push; fallback to long-polling for compatibility

### Interview Tip
"For live comments, use WebSocket for client connections and Redis Pub/Sub or Kafka for the message bus. When a user posts, publish to a channel; all subscribed WebSocket servers forward to their clients. For millions of viewers, aggregate or sample to reduce fan-out."

---

## 3. Distributed Counter

### The Problem
A social app needs "like" counts, view counts, share counts. Millions of increments per second. Must be highly available and eventually consistent. Strong consistency is expensive at scale.

### The Architecture/Solution

**Sharded Counters**
- Single counter = hot partition; one key gets all writes
- **Solution**: Shard the counter: `likes:post_123:0` … `likes:post_123:99`
- Each write goes to a random shard: `INCR likes:post_123:{random(0,99)}`
- Read: `SUM(gets)` across all shards
- Trade-off: Read cost = 100 gets; write throughput = 100x

**Eventual Consistency**
- Accept that count may lag by seconds
- Use async aggregation (e.g., Kafka consumer) for exact count if needed
- Display "10K+" for very large counts to hide small errors

**Redis INCR**
- Atomic increment; O(1)
- Persist via AOF; replicate to replicas

### Key Numbers
| Metric | Value |
|--------|-------|
| Write amplification | 100x with 100 shards |
| Read cost | 100 GETs (or 1 MGET) |
| Consistency | Eventual (seconds) |
| INCR latency | < 1ms |

### Architecture Diagram (ASCII)

```
    ┌─────────────┐     ┌─────────────┐
    │  Client A   │     │  Client B   │
    └──────┬──────┘     └──────┬──────┘
           │                   │
           │  INCR likes:p:42  │  INCR likes:p:17
           └─────────┬─────────┘
                     │
            ┌────────▼────────┐
            │  Redis Cluster   │
            │  likes:post_1:0  │
            │  likes:post_1:1  │
            │  ...             │
            │  likes:post_1:99 │
            └────────┬────────┘
                     │
            Read: SUM(likes:post_1:*) = total
```

### Key Takeaways
- **Hot partition** is the enemy; shard counters to spread writes
- **Eventual consistency** is acceptable for social counts
- **Approximate counts** (e.g., "10K+") reduce read cost and hide lag

### Interview Tip
"For a high-write counter like likes, shard it: 100 keys per counter, random shard on write. Read = sum all shards. This trades read cost for write throughput. Eventual consistency is fine for social counts."

---

## 4. Real-Time Presence Platform

### The Problem
Slack/Discord-style "online" status: who is online, typing indicators, last seen. Millions of users; status changes frequently. Must be real-time and handle connection churn (mobile app backgrounding, network drops).

### The Architecture/Solution

**Heartbeat**
- Client sends heartbeat every 30–60 seconds
- Server updates `user_id → last_seen` in Redis
- If no heartbeat for 90 seconds → mark offline

**Redis TTL**
- `SET presence:user_123 "online" EX 90`
- Client heartbeat = `SET presence:user_123 "online" EX 90` (refresh TTL)
- No heartbeat → key expires → user offline
- No cron needed; TTL handles cleanup

**Pub/Sub for Typing**
- `PUBLISH typing:channel_456 user_123`
- Subscribers get real-time typing indicators
- Ephemeral; no persistence

**Presence Query**
- `MGET presence:user_1 presence:user_2 ...` for channel member list
- Or maintain `channel_id → set of online user_ids` in Redis

### Key Numbers
| Metric | Value |
|--------|-------|
| Heartbeat interval | 30–60 sec |
| TTL | 90 sec |
| Redis key size | ~50 bytes per user |
| 1M online users | ~50 MB Redis |

### Architecture Diagram (ASCII)

```
    ┌─────────────┐                    ┌─────────────┐
    │  Client A   │                    │  Client B   │
    │  heartbeat  │                    │  heartbeat  │
    └──────┬──────┘                    └──────┬──────┘
           │ every 30s                        │
           └──────────────────┬───────────────┘
                              │
                    ┌─────────▼─────────┐
                    │  Presence API     │
                    │  SET presence:u  │
                    │  EX 90            │
                    └─────────┬─────────┘
                              │
                    ┌─────────▼─────────┐
                    │  Redis            │
                    │  presence:user_1  │
                    │  TTL=90s          │
                    │  (auto-expire)    │
                    └─────────┬─────────┘
                              │
                    ┌─────────▼─────────┐
                    │  Pub/Sub          │
                    │  typing:channel   │
                    └──────────────────┘
```

### Key Takeaways
- **TTL** eliminates need for explicit cleanup jobs
- **Heartbeat** is simple; tune interval vs. server load
- **Pub/Sub** for ephemeral events (typing); Redis for presence state

### Interview Tip
"For presence, use Redis with TTL. SET key EX 90 on each heartbeat. Client stops → key expires → offline. No cron. For typing, use Pub/Sub—ephemeral, no storage."

---

## 5. How Disney+ Hotstar Delivered 5 Billion Emojis in Real Time

### The Problem
During cricket matches, millions of viewers send emoji reactions (e.g., 🔥, 😍) in real time. These must appear on all viewers' screens with low latency. Peak: 5 billion emojis during a tournament; 55M cricket fans.

### The Architecture/Solution

**Receiving**
- Go-based API servers receive emojis via HTTP
- Buffer locally; async write to Kafka via Goroutines
- Non-blocking; high throughput

**Processing**
- Apache Spark consumes from Kafka
- Aggregates emojis over micro-batches (e.g., 1–5 sec windows)
- Normalized counts (e.g., "🔥: 12.5K") written back to Kafka
- Exactly-once processing to avoid duplicates

**Delivering**
- Python consumers read from Kafka
- MQTT-based Pub/Sub (EMQX brokers) for delivery
- 250K connections per broker
- Persistent connections to clients

**Design Choices**
- No caching (real-time requirement)
- Independent services for scalability
- Async throughout to avoid blocking

### Key Numbers
| Metric | Value |
|--------|-------|
| Emojis (tournament) | 5 billion |
| Cricket fans | 55 million |
| EMQX connections | 250K per broker |
| Processing | Micro-batch (Spark) |

### Architecture Diagram (ASCII)

```
    ┌─────────────┐                    ┌─────────────┐
    │  Client 1   │                    │  Client N   │
    │  (Emoji)    │                    │  (Emoji)    │
    └──────┬──────┘                    └──────┬──────┘
           │ HTTP POST                         │
           └──────────────────┬───────────────┘
                              │
                    ┌─────────▼─────────┐
                    │  Go API Servers   │
                    │  (Buffer+async)   │
                    └─────────┬─────────┘
                              │
                    ┌─────────▼─────────┐
                    │  Kafka             │
                    │  (Emoji stream)   │
                    └─────────┬─────────┘
                              │
              ┌───────────────┼───────────────┐
              │               │               │
         ┌────▼────┐    ┌─────▼─────┐   ┌────▼────┐
         │ Spark   │    │  Kafka    │   │  EMQX   │
         │(Aggregate)│   │(Aggregated)│   │(Pub/Sub)│
         └─────────────────────────┘   └────┬────┘
                                             │
                                    ┌────────▼────────┐
                                    │  Clients         │
                                    │  (MQTT)          │
                                    └──────────────────┘
```

### Key Takeaways
- **Aggregate before fan-out**; don't send 5B raw emojis to 55M clients
- **Kafka** for buffering and decoupling
- **MQTT** for efficient delivery to many clients
- **Async** at every stage

### Interview Tip
"For real-time reactions at scale, aggregate first (e.g., Spark streaming), then fan out aggregated counts. Avoid sending every raw event to every viewer—you'll blow up the network."

---

## 6. How Disney+ Hotstar Scaled to 25 Million Concurrent Users

### The Problem
ICC Cricket World Cup 2019 semi-final: 25.3 million concurrent viewers on a single live stream. Need to deliver video reliably with low latency and handle massive traffic spikes.

### The Architecture/Solution

**Auto-Scaling**
- Horizontal scaling of origin servers and processing pipelines
- Scale up before peak (predictable event); scale down after

**CDN**
- Video segments cached at edge; most traffic served from CDN
- Origin only handles cache misses and first-time requests
- Multi-CDN for redundancy

**Adaptive Bitrate Streaming**
- HLS/DASH; multiple quality levels
- Client adapts to network conditions
- Reduces buffering and improves QoE

**Stateless Origin**
- Any request can hit any server
- Session/state in Redis or similar

### Key Numbers
| Metric | Value |
|--------|-------|
| Concurrent viewers | 25.3 million |
| Event | ICC Cricket World Cup 2019 |
| CDN | Edge caching |
| Scaling | Pre-provisioned for peak |

### Architecture Diagram (ASCII)

```
                    ┌─────────────────────────────────┐
                    │  CDN (Edge Locations)           │
                    │  Video segments cached          │
                    └────────────┬────────────────────┘
                                 │ cache miss
                    ┌────────────▼────────────────────┐
                    │  Load Balancer                   │
                    └────────────┬────────────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │                       │                       │
    ┌────▼────┐             ┌────▼────┐             ┌────▼────┐
    │ Origin  │             │ Origin  │             │ Origin  │
    │ Server  │             │ Server  │             │ Server  │
    └─────────┘             └─────────┘             └─────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌────────────▼────────────────────┐
                    │  Origin Storage / Transcoding    │
                    └─────────────────────────────────┘
```

### Key Takeaways
- **CDN** absorbs the vast majority of traffic
- **Predictable events** allow pre-scaling; don't wait for load
- **Adaptive bitrate** improves QoE under variable network

### Interview Tip
"For 25M concurrent viewers, CDN is non-negotiable. Origin serves cache misses only. Pre-scale for known peaks like sports events. Use adaptive bitrate (HLS/DASH) for resilience."

---

## 7. How Disney+ Scaled to 11 Million Users on Launch Day

### The Problem
Disney+ launched in November 2019 with massive demand. 11 million users signed up on day one. Need to handle signup, auth, and streaming without outages.

### The Architecture/Solution

**AWS Infrastructure**
- Multi-region deployment for redundancy
- Auto-scaling groups for compute
- RDS, ElastiCache, S3 for storage and caching

**Microservices**
- Signup, auth, catalog, playback, billing as separate services
- Independent scaling; failure isolation
- API Gateway for routing and rate limiting

**Queue-Based Signup**
- Signup spikes can overwhelm DB
- Queue signup requests; process asynchronously
- Return "we're setting up your account" immediately

**Caching**
- Catalog, user sessions, entitlements cached
- Reduce DB load during peak

### Key Numbers
| Metric | Value |
|--------|-------|
| Launch day signups | 11 million |
| Platform | Disney+ |
| Cloud | AWS |
| Architecture | Microservices |

### Key Takeaways
- **Queue for spikes**; don't let signup burst hit DB directly
- **Microservices** allow per-service scaling
- **Multi-region** for launch-day resilience

### Interview Tip
"For a big launch, use queues to absorb signup spikes. Don't synchronously write 11M rows to DB at once—queue and process in batches. Scale each microservice independently."

---

## 8. How Facebook Scaled Live Video to a Billion Users

### The Problem
Facebook Live streams to billions of users. Need low-latency, high-quality video that adapts to each viewer's network. Must handle viral streams (millions of concurrent viewers) and global distribution.

### The Architecture/Solution

**Adaptive Bitrate Streaming**
- Multiple quality levels (e.g., 360p to 1080p)
- Client selects based on bandwidth
- Reduces buffering; improves QoE

**Edge PoPs**
- Transcoding and packaging at the edge
- Reduces latency; content closer to viewers
- Scale transcoding horizontally

**RTMP Ingest**
- Broadcasters push RTMP to ingest servers
- Transcode to HLS/DASH for playback
- Low-latency modes (e.g., 3–5 sec end-to-end)

**Fan-Out**
- CDN for video segments
- Origin handles only cache misses
- Multi-CDN for redundancy

### Key Numbers
| Metric | Value |
|--------|-------|
| Users | 1B+ |
| Latency | 3–5 sec (low-latency mode) |
| Ingest | RTMP |
| Playback | HLS/DASH |

### Architecture Diagram (ASCII)

```
    ┌─────────────┐
    │  Broadcaster│
    │  (RTMP)     │
    └──────┬──────┘
           │
    ┌──────▼──────┐
    │  Ingest     │
    │  (RTMP)     │
    └──────┬──────┘
           │
    ┌──────▼──────┐
    │  Transcode  │
    │  (Edge)     │
    └──────┬──────┘
           │
    ┌──────▼──────────────────────────┐
    │  CDN (Edge PoPs)                 │
    │  HLS/DASH segments               │
    └──────┬──────────────────────────┘
           │
    ┌──────▼──────┐
    │  Viewers    │
    │  (Adaptive) │
    └─────────────┘
```

### Key Takeaways
- **Adaptive bitrate** is table stakes for live video
- **Edge transcoding** reduces latency and origin load
- **CDN** for global distribution

### Interview Tip
"For live video at scale, use RTMP ingest, transcode at the edge to HLS/DASH, and serve via CDN. Adaptive bitrate is essential. Edge PoPs reduce latency for global viewers."

---

## 9. How Canva Supports Real-Time Collaboration for 135 Million Monthly Users

### The Problem
Canva allows multiple users to edit a design simultaneously. Cursors, object placement, and edits must sync in real time. Conflicts (two users moving the same object) must be resolved. 135M monthly users; 50+ users per design in some cases.

### The Architecture/Solution

**CRDTs (Conflict-free Replicated Data Types)**
- Data structures that merge without central coordination
- Each operation is commutative and idempotent
- No "last write wins" conflicts; automatic merge
- Used for: text, lists, sets

**Operational Transformation (OT)**
- Alternative to CRDTs; used by Google Docs
- Transform operations against concurrent edits
- Requires central server for transformation
- Canva uses a hybrid: OT for some, CRDT for others

**WebRTC for Cursors**
- Initially: WebSockets + Redis for mouse positions
- Evolved to WebRTC for peer-to-peer cursor updates
- Reduces server load; lower latency for cursor movement
- STUN/TURN for NAT traversal

**Backend Sync**
- Design state persisted in DB
- Real-time layer (WebSocket/WebRTC) for live updates
- Periodic sync to DB for durability

### Key Numbers
| Metric | Value |
|--------|-------|
| Monthly users | 135M (40M+ cited in some sources) |
| Users per design | Up to 50 |
| Cursor updates | 3/sec per user |
| Approach | CRDTs, OT, WebRTC |

### Architecture Diagram (ASCII)

```
    ┌─────────────┐                    ┌─────────────┐
    │  User A     │                    │  User B     │
    │  (Canva)    │                    │  (Canva)    │
    └──────┬──────┘                    └──────┬──────┘
           │                                  │
           │  WebRTC (cursors)                 │
           │◄─────────────────────────────────►│
           │                                  │
           │  WebSocket (design ops)           │
           └──────────────────┬───────────────┘
                              │
                    ┌─────────▼─────────┐
                    │  Sync Server      │
                    │  CRDT/OT merge    │
                    └─────────┬─────────┘
                              │
                    ┌─────────▼─────────┐
                    │  Database         │
                    │  (Persistent)     │
                    └───────────────────┘
```

### Key Takeaways
- **CRDTs** enable conflict-free merge without central coordination
- **OT** is an alternative; requires server for transform
- **WebRTC** for high-frequency, low-value data (cursors) reduces server load
- **Hybrid** approaches common: CRDT for some structures, OT for others

### Interview Tip
"For real-time collaboration, use CRDTs or OT. CRDTs merge automatically; OT transforms operations on the server. For cursors, WebRTC P2P can reduce server load. Canva and Google Docs are the canonical examples."
