# Redis Use Cases

## 1. Concept Overview

### Definition
**Redis** (Remote Dictionary Server) is an in-memory data structure store used as a database, cache, message broker, and streaming engine. It supports strings, hashes, lists, sets, sorted sets, bitmaps, hyperloglogs, geospatial indexes, and streams. Beyond simple caching, Redis powers session storage, rate limiting, leaderboards, pub/sub, distributed locks, and real-time analytics.

### Purpose
- **Speed**: Sub-millisecond latency; in-memory operations
- **Versatility**: Multiple data structures for different use cases
- **Persistence**: Optional RDB snapshots and AOF for durability
- **Replication**: Master-replica for HA and read scaling
- **Cluster**: Sharding for horizontal scaling

### Problems It Solves
- **Caching**: Reduce database load, lower latency
- **Session state**: Stateless app servers with shared session store
- **Rate limiting**: Throttle API requests per user/IP
- **Real-time**: Pub/sub, leaderboards, live dashboards
- **Distributed coordination**: Locks, queues, coordination

---

## 2. Real-World Motivation

### Uber
- **Geospatial**: `GEORADIUS` for "find drivers nearby"; real-time matching
- **Session store**: User/driver session state
- **Rate limiting**: API throttling
- **Scale**: Millions of geospatial queries per second

### Twitter
- **Timeline**: Redis for home timeline caching
- **Counters**: Tweet counts, follower counts
- **Session**: User sessions
- **Real-time**: Activity feeds

### GitHub
- **Rate limiting**: API rate limits per user
- **Caching**: Repository metadata, user data
- **Job queues**: Sidekiq (Redis-backed)
- **Real-time**: Live updates, notifications

### Stack Overflow
- **Caching**: Question/answer caching
- **Session**: User sessions
- **Leaderboards**: Reputation rankings (sorted sets)
- **Rate limiting**: Prevent abuse

### Netflix
- **Session store**: User preferences, watch state
- **Caching**: Recommendations, metadata
- **Distributed locks**: Coordinate across services
- **Pub/sub**: Event distribution

---

## 3. Architecture Diagrams

### Redis in Typical Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     REDIS IN TYPICAL ARCHITECTURE                         │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   CLIENTS (App Servers)                                                   │
│   ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐                        │
│   │  Web 1  │ │  Web 2  │ │  API 1  │ │  API 2  │                        │
│   └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘                        │
│        │           │           │           │                             │
│        └───────────┴───────────┴───────────┘                             │
│                            │                                              │
│                    ┌───────▼───────┐                                      │
│                    │  REDIS        │                                      │
│                    │  - Cache      │  (or Redis Cluster)                  │
│                    │  - Session    │                                      │
│                    │  - Rate limit │                                      │
│                    │  - Pub/Sub    │                                      │
│                    │  - Locks      │                                      │
│                    └───────┬───────┘                                      │
│                            │                                              │
│                    ┌───────▼───────┐                                      │
│                    │  DATABASE     │  (PostgreSQL, MySQL, etc.)          │
│                    │  (Source of   │                                      │
│                    │   truth)      │                                      │
│                    └───────────────┘                                      │
│                                                                          │
│   Redis: Sub-ms reads; offload DB; shared state across app servers       │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Redis Data Structures Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     REDIS DATA STRUCTURES → USE CASES                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   STRING      → Cache, counters, session (single value)                  │
│   HASH        → Object cache (user:123 → {name, email, ...})            │
│   LIST        → Queue, feed, recent items                                │
│   SET         → Unique items, tags, "who liked this"                    │
│   SORTED SET  → Leaderboard, priority queue, time-series                │
│   BITMAP      → Presence, feature flags, analytics                       │
│   HYPERLOGLOG → Cardinality (unique visitors)                            │
│   GEO         → Nearby, radius search (Uber)                             │
│   STREAM      → Message broker, event log                                │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Core Mechanics

### Use Case Deep Dives

| Use Case | Data Structure | Key Commands | Company Example |
|----------|----------------|--------------|-----------------|
| **Caching** | String, Hash | GET, SET, HSET, HGET | All |
| **Session store** | Hash, String | SETEX, HSET, EXPIRE | Netflix, GitHub |
| **Rate limiter** | String (INCR) or Sorted Set | INCR, EXPIRE, ZADD | GitHub, Stripe |
| **Leaderboard** | Sorted Set | ZADD, ZRANGE, ZREVRANK | Stack Overflow, gaming |
| **Pub/Sub** | - | PUBLISH, SUBSCRIBE | Real-time feeds |
| **Distributed lock** | String | SET NX EX | Netflix, distributed systems |
| **Message broker** | Stream | XADD, XREAD, XGROUP | Event sourcing |
| **Geospatial** | Geo | GEOADD, GEORADIUS | Uber, delivery apps |
| **Counters** | String, Hash | INCR, HINCRBY | Twitter, analytics |
| **Bit operations** | String (bitmap) | SETBIT, GETBIT, BITCOUNT | Feature flags, presence |

---

## 5. Numbers

| Metric | Value |
|--------|-------|
| Latency | 0.1-1 ms (in-memory) |
| Throughput | 100K-1M ops/sec (single instance) |
| Max key size | 512 MB |
| Max value size | 512 MB |
| Persistence | RDB (snapshot), AOF (append-only) |
| Cluster | 16,384 hash slots |

### Uber Geospatial Scale
- Millions of driver locations updated per second
- GEORADIUS: Find drivers within 5km in <10ms
- Redis Cluster for sharding

### Twitter Scale
- Billions of keys (timelines, counters)
- Redis Cluster; multiple clusters per service
- Sub-ms cache hit latency

---

## 6. Tradeoffs

### Redis vs Memcached

| Aspect | Redis | Memcached |
|--------|-------|-----------|
| **Data structures** | Rich (sets, sorted sets, etc.) | Key-value only |
| **Persistence** | Yes | No |
| **Replication** | Yes | No |
| **Use cases** | Cache + much more | Pure cache |
| **Memory** | More features = more overhead | Minimal |

### Redis vs Database

| Aspect | Redis | Database |
|--------|-------|----------|
| **Durability** | Optional | Primary |
| **Query** | Key-based | SQL, indexes |
| **Scale** | Vertical + cluster | Sharding, replicas |
| **Use** | Cache, session, real-time | Source of truth |

---

## 7. Variants / Implementations

### Redis Variants
- **Redis OSS**: Standard Redis
- **Redis Stack**: + RedisJSON, RedisSearch, RedisGraph
- **Redis Enterprise**: Commercial; multi-tenant, persistence
- **KeyDB**: Redis fork; multi-threaded
- **Valkey**: Linux Foundation fork (post-Redis license change)

### Managed Services
- **AWS ElastiCache**: Redis managed
- **Redis Cloud**: Managed Redis
- **Azure Cache for Redis**: Managed

---

## 8. Scaling Strategies

- **Replication**: Master-replica for read scaling
- **Cluster**: Shard across nodes (16K slots)
- **Persistence**: RDB + AOF for durability
- **Connection pooling**: Reuse connections; avoid per-request connect

---

## 9. Failure Scenarios

| Scenario | Mitigation |
|----------|------------|
| **Redis down** | Replica failover; cache miss → DB |
| **Memory full** | Eviction policy (LRU, LFU); maxmemory |
| **Network partition** | Sentinel/Cluster; split-brain handling |
| **Slow commands** | Avoid KEYS *; use SCAN; monitor slowlog |

---

## 10. Performance Considerations

- **Pipelining**: Batch commands; reduce RTT
- **Connection pooling**: Avoid connection per request
- **Eviction**: Configure maxmemory-policy
- **Avoid**: KEYS *, large values, blocking commands in hot path

---

## 11. Use Cases (Detailed)

### 1. Caching
- **Structure**: String, Hash
- **Commands**: GET, SET, HSET, HGET, SETEX
- **Pattern**: Cache-aside; TTL for expiry
- **Example**: Product details, user profile

### 2. Session Store
- **Structure**: Hash (session_id → {user_id, data})
- **Commands**: HSET, HGETALL, EXPIRE
- **Example**: Netflix watch state, GitHub session

### 3. Rate Limiter
- **Structure**: String (INCR) or Sorted Set (sliding window)
- **Commands**: INCR, EXPIRE or ZADD, ZREMRANGEBYSCORE
- **Example**: GitHub API 5000 req/hr

### 4. Leaderboard
- **Structure**: Sorted Set (score = points)
- **Commands**: ZADD, ZREVRANGE, ZREVRANK
- **Example**: Stack Overflow reputation, game scores

### 5. Pub/Sub
- **Structure**: Channels
- **Commands**: PUBLISH, SUBSCRIBE
- **Example**: Real-time notifications, chat

### 6. Distributed Lock (Redlock)
- **Structure**: String
- **Commands**: SET key value NX EX 30
- **Example**: Netflix, prevent duplicate processing

### 7. Message Broker (Streams)
- **Structure**: Stream
- **Commands**: XADD, XREAD, XGROUP, XACK
- **Example**: Event sourcing, job queues

### 8. Geospatial
- **Structure**: Geo (sorted set internally)
- **Commands**: GEOADD, GEORADIUS, GEODIST
- **Example**: Uber "drivers nearby"

### 9. Counters
- **Structure**: String
- **Commands**: INCR, INCRBY, HINCRBY
- **Example**: Tweet counts, view counts

### 10. Bit Operations
- **Structure**: String (bitmap)
- **Commands**: SETBIT, GETBIT, BITCOUNT, BITOP
- **Example**: Feature flags, daily active users

---

## 12. Comparison Tables

### Use Case → Data Structure Matrix

| Use Case | Structure | Commands | TTL |
|----------|-----------|----------|-----|
| **Cache** | String/Hash | GET/SET/HSET | Yes |
| **Session** | Hash | HSET/HGET | Yes |
| **Rate limit** | String/Sorted Set | INCR/ZADD | Yes |
| **Leaderboard** | Sorted Set | ZADD/ZRANGE | Optional |
| **Pub/Sub** | - | PUBLISH/SUBSCRIBE | No |
| **Lock** | String | SET NX EX | Yes |
| **Stream** | Stream | XADD/XREAD | No |
| **Geo** | Geo | GEOADD/GEORADIUS | Optional |
| **Counters** | String | INCR | Optional |
| **Bitmap** | String | SETBIT/BITCOUNT | Optional |

### Company → Use Case Mapping

| Company | Primary Use Cases |
|---------|-------------------|
| **Uber** | Geospatial, session, rate limit |
| **Twitter** | Timeline cache, counters, session |
| **GitHub** | Rate limit, caching, Sidekiq (queues) |
| **Stack Overflow** | Cache, leaderboard, session |
| **Netflix** | Session, cache, distributed lock |
| **Instagram** | Feed cache, session |
| **Pinterest** | Feed, caching |

---

## 13. Code or Pseudocode

### Rate Limiter (Sliding Window)

```redis
# User 123: 100 requests per minute
# Key: ratelimit:123
# ZADD with score = timestamp
ZADD ratelimit:123 1615123456.789 "req1"
ZREMRANGEBYSCORE ratelimit:123 -inf (now - 60)
ZCARD ratelimit:123
# If ZCARD > 100 → reject
```

### Leaderboard

```redis
ZADD leaderboard 1500 "user:alice"
ZADD leaderboard 2300 "user:bob"
ZREVRANGE leaderboard 0 9 WITHSCORES   # Top 10
ZREVRANK leaderboard "user:alice"       # Alice's rank
```

### Distributed Lock (Redlock)

```redis
# Acquire
SET lock:resource "unique_id" NX EX 30

# Release (Lua script for atomicity)
if redis.call("get", KEYS[1]) == ARGV[1] then
  return redis.call("del", KEYS[1])
else
  return 0
end
```

### Geospatial (Uber-style)

```redis
GEOADD drivers 13.4050 52.5200 "driver:1"
GEOADD drivers 13.4100 52.5210 "driver:2"
GEORADIUS drivers 13.4 52.52 5 km WITHDIST
# Returns drivers within 5km of (13.4, 52.52)
```

### Session Store

```redis
HSET session:abc123 user_id 456 created_at 1615123456
HSET session:abc123 cart "item1,item2"
EXPIRE session:abc123 86400

HGETALL session:abc123
```

### Pub/Sub

```redis
# Publisher
PUBLISH notifications "user:123:new_message"

# Subscriber
SUBSCRIBE notifications
```

---

## 14. Interview Discussion

### Key Points to Articulate
1. **Beyond cache**: Session, rate limit, leaderboard, pub/sub, locks, geo, streams
2. **Data structures**: Choose right structure for use case
3. **Real examples**: Uber geo, GitHub rate limit, Twitter timeline
4. **Tradeoffs**: In-memory = fast but volatile; persistence optional
5. **Scaling**: Replication, cluster

### Common Interview Questions
- **"How would you implement rate limiting?"** → Redis INCR + EXPIRE (fixed window) or ZADD + ZREMRANGEBYSCORE (sliding)
- **"How does Uber find nearby drivers?"** → Redis GEORADIUS; drivers' locations in Geo sorted set
- **"How would you implement a leaderboard?"** → Sorted Set; ZADD score; ZREVRANGE for top N
- **"Distributed lock with Redis?"** → SET NX EX; unique value; Lua script for atomic release
- **"Redis vs Memcached?"** → Redis: more data structures, persistence, replication; Memcached: simple cache only

### Red Flags to Avoid
- Suggesting Redis for primary database (use for cache/session)
- Not considering eviction (maxmemory)
- Using KEYS * in production
- Ignoring persistence (when durability needed)

### Ideal Answer Structure
1. Define Redis (in-memory data structure store)
2. List use cases: cache, session, rate limit, leaderboard, pub/sub, locks, geo
3. Map use case → data structure → commands
4. Give company examples (Uber, GitHub, Twitter)
5. Discuss tradeoffs (speed vs durability, scaling)
