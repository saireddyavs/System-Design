# System Design Interview Framework

## 1. Concept Overview

### Definition
**System design interviews** assess a candidate's ability to architect scalable, reliable, and maintainable systems. Candidates are given an open-ended problem (e.g., "Design Twitter") and expected to clarify requirements, estimate scale, propose a high-level design, and dive into critical components—all within 45-60 minutes.

### Purpose
- **Demonstrate architecture skills**: Show you can think at scale
- **Communication**: Explain tradeoffs clearly to interviewer
- **Structured thinking**: Follow a repeatable framework
- **Depth on demand**: Go deep when asked, stay high-level otherwise

### Problems It Solves
- **Chaos**: Without a framework, candidates ramble or get stuck
- **Incomplete designs**: Missing requirements, scale, or failure modes
- **Poor time management**: Spending 30 min on one component
- **Interviewer mismatch**: Not addressing what they care about

---

## 2. Real-World Motivation

### Why Companies Use This Format
- **FAANG/MANGA**: Standard for L4+ (senior) and above
- **Startups**: Staff+ roles expect system design
- **Real work**: Architects make these decisions daily
- **Signal**: Separates senior from junior thinking

### What Interviewers Evaluate
- **Clarification**: Do you ask the right questions?
- **Scale**: Can you estimate and reason with numbers?
- **Tradeoffs**: Do you understand pros/cons?
- **Depth**: Can you go deep on databases, caching, etc.?
- **Communication**: Clear, structured, collaborative

---

## 3. Architecture Diagrams

### Interview Timeline (45-60 min)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     SYSTEM DESIGN INTERVIEW TIMELINE                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   0 min    5 min    10 min   25 min   40 min   50 min   60 min           │
│   │         │         │         │         │         │         │          │
│   ├─────────┼─────────┼─────────┼─────────┼─────────┼─────────┤          │
│   │ Clarify │ Estimate│  HLD    │ Detailed│ Scale & │ Wrap-up │          │
│   │ Reqs    │ (B-o-E) │         │ Design  │ Follow  │         │          │
│   │ 3-5 min │ 3-5 min │ 10-15   │ 10-15   │ 5-10    │         │          │
│   │         │         │  min    │  min    │  min    │         │          │
│   └─────────┴─────────┴─────────┴─────────┴─────────┴─────────┘          │
│                                                                          │
│   Key: Don't skip clarification. Don't over-design. Leave time for Q&A.  │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### High-Level Design Template

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     TYPICAL HLD COMPONENTS                                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   ┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐            │
│   │ Clients │────►│   LB    │────►│  App    │────►│   DB    │            │
│   │ (Web,   │     │ / API   │     │ Servers │     │         │            │
│   │  Mobile)│     │ Gateway │     │         │     │         │            │
│   └─────────┘     └─────────┘     └────┬────┘     └─────────┘            │
│         │                │              │                │                │
│         │                │              │                │                │
│         │                ▼              ▼                │                │
│         │           ┌─────────┐    ┌─────────┐           │                │
│         │           │   CDN   │    │  Cache  │           │                │
│         │           │         │    │ (Redis) │           │                │
│         │           └─────────┘    └─────────┘           │                │
│         │                                                  │                │
│         └──────────────────► Message Queue ◄───────────────┘                │
│                             (async jobs)                                    │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Core Mechanics: The 5-Step Framework

### Step 1: Requirements Clarification (3-5 minutes)

**Functional Requirements**
- What does the system do? (Core features)
- Who are the users? (End users, admins, APIs)
- What are the main use cases? (Read, write, search, etc.)

**Non-Functional Requirements**
- Scale: DAU, QPS, storage growth
- Latency: p50, p99 targets
- Availability: 99.9%? 99.99%?
- Consistency: Strong vs eventual

**Scale Estimation**
- DAU (Daily Active Users)
- Reads vs writes per user per day
- Data size per entity
- Retention period

**Questions to Ask the Interviewer**

| Category | Example Questions |
|----------|-------------------|
| **Scope** | "Should I focus on the feed, or the entire Twitter?" |
| **Scale** | "What's the expected DAU—1M, 10M, or 100M?" |
| **Priorities** | "Is latency or consistency more important?" |
| **Constraints** | "Any technology constraints?" |
| **Out of scope** | "Should I exclude DMs, or include them?" |

**Template**
```
Let me clarify:
- Functional: [list 3-5 core features]
- Scale: DAU X, reads Y/user/day, writes Z/user/day
- Latency: p99 < 200ms for reads
- Availability: 99.9%
- Any priorities or constraints?
```

---

### Step 2: Back-of-Envelope Estimation (3-5 minutes)

**QPS Calculation**
```
Daily requests = DAU × requests per user per day
QPS (avg) = Daily requests / 86,400
QPS (peak) = QPS (avg) × 2-3 (peak multiplier)
```

**Storage Calculation**
```
Storage/day = Operations/day × Bytes per operation
Total = Daily storage × Retention × Replication factor (e.g., 3)
```

**Bandwidth Calculation**
```
Bandwidth = QPS × Avg request/response size
```

**Common Numbers to Know**

| Item | Value |
|------|-------|
| Seconds per day | 86,400 |
| 1M DAU, 100 req/user/day | ~1,200 QPS avg |
| 100M DAU, 100 req/user/day | ~120K QPS avg |
| L1 cache | ~0.5 ns |
| RAM | ~100 ns |
| SSD | ~100 μs |
| HDD | ~10 ms |
| Same-DC RTT | ~0.5 ms |
| Cross-continent RTT | ~150 ms |

**Example (Twitter-like)**
```
DAU = 100M
Reads: 100/user/day → 10B/day → ~115K read QPS (avg), ~350K peak
Writes: 10/user/day → 1B/day → ~12K write QPS (avg), ~36K peak
Tweet = 300 bytes → 1B × 300B = 300 GB/day → 55 TB/year (×3 replication = 165 TB)
```

---

### Step 3: High-Level Design (10-15 minutes)

**API Design**
- REST or gRPC? (REST for simplicity in interview)
- Key endpoints with request/response shape
- Idempotency for writes
- Pagination, filtering

**Example API**
```
POST /tweets
  Body: { user_id, content, media_ids[] }
  Response: { tweet_id, created_at }

GET /feed?user_id=X&cursor=Y&limit=20
  Response: { tweets[], next_cursor }

GET /tweet/:id
  Response: { tweet, author, engagement }
```

**Data Model**
- Core entities and relationships
- Key indexes
- Normalized vs denormalized (tradeoffs)

**Example Data Model**
```
Users: user_id (PK), username, created_at
Tweets: tweet_id (PK), user_id (FK), content, created_at
Follows: follower_id, followee_id (composite PK)
Feed: user_id, tweet_id, created_at (for pre-computed feed)
```

**Architecture Diagram**
- Client → Load Balancer → App Servers
- App Servers → Cache (Redis) → Database
- Async: Message Queue → Workers
- CDN for static/media
- Draw boxes, label, show data flow

---

### Step 4: Detailed Design (10-15 minutes)

**Deep Dive into Critical Components**

| Component | What to Cover |
|-----------|---------------|
| **Database** | Schema, indexes, sharding strategy, replication |
| **Cache** | What to cache, TTL, invalidation, cache-aside vs write-through |
| **Load Balancer** | Algorithm (round-robin, least connections) |
| **Message Queue** | Why async, ordering, exactly-once, dead letter |
| **Storage** | Blob vs block, CDN for media |

**Algorithms & Data Structures**
- Feed: Fan-out on write vs fan-out on read
- Search: Inverted index, Elasticsearch
- Rate limiting: Token bucket, sliding window
- Deduplication: Idempotency keys, bloom filters

**Identify Bottlenecks**
- "At 350K read QPS, single DB can't handle it → need caching"
- "Feed generation is CPU-heavy → fan-out on write, pre-compute"

---

### Step 5: Scaling & Follow-ups (5-10 minutes)

**Bottleneck Identification**
- Database: Sharding, read replicas, connection pooling
- Cache: Partitioning, multi-level cache
- App servers: Horizontal scaling, stateless design
- Network: CDN, compression, HTTP/2

**Scaling Strategies**
- Horizontal: Add more machines
- Vertical: Bigger machines (limited)
- Caching: Reduce DB load
- Async: Decouple, batch
- Partitioning: Shard by user_id, etc.

**Edge Cases**
- What if a user has 100M followers? (Celebrity problem)
- What if cache fails? (Fallback to DB, degrade gracefully)
- What if a region goes down? (Multi-region, failover)

---

## 5. Numbers

### Estimation Cheat Sheet

| DAU | Read QPS (100 req/user) | Write QPS (10 req/user) |
|-----|-------------------------|--------------------------|
| 1M | ~1,200 | ~120 |
| 10M | ~12,000 | ~1,200 |
| 100M | ~120,000 | ~12,000 |
| 1B | ~1.2M | ~120,000 |

### Latency Targets

| Operation | Typical Target |
|-----------|----------------|
| API read | p99 < 100-200 ms |
| API write | p99 < 200-500 ms |
| Search | p99 < 500 ms |
| Real-time | < 100 ms |

### Storage Sizes

| Entity | Size |
|--------|------|
| User profile | 1 KB |
| Tweet/post | 300-500 bytes |
| Chat message | 200-500 bytes |
| Short URL | 100 bytes |
| Photo metadata | 500 bytes |
| Photo (full) | 3-5 MB |

---

## 6. Tradeoffs

### Read-Heavy vs Write-Heavy

| Read-Heavy | Write-Heavy |
|------------|-------------|
| Cache aggressively | Write-through cache |
| Read replicas | Queue writes |
| Denormalize for reads | Optimize write path |
| CDN for static | Async processing |
| Example: Feed, search | Example: Logging, analytics |

### Consistency vs Availability

| Strong Consistency | Eventual Consistency |
|--------------------|----------------------|
| Banking, inventory | Social feed, likes |
| ACID, 2PC | CRDTs, conflict resolution |
| Higher latency | Lower latency |
| CAP: Choose C | CAP: Choose A |

### Sync vs Async

| Synchronous | Asynchronous |
|-------------|--------------|
| User waits | Fire-and-forget |
| Simpler | More complex |
| Lower latency (when fast) | Better throughput |
| Example: Login | Example: Email send |

---

## 7. Variants / Implementations

### Common Problem Types

| Type | Key Components | Example |
|------|----------------|---------|
| **Feed/Timeline** | Fan-out, cache, ranking | Twitter, Instagram |
| **Chat/Messaging** | WebSocket, message queue, presence | Slack, WhatsApp |
| **Search** | Inverted index, Elasticsearch | Google, product search |
| **Storage/File** | Blob storage, CDN, dedup | Dropbox, Google Drive |
| **Rate Limiter** | Token bucket, Redis | API gateway |
| **URL Shortener** | Hash, KV store, redirect | bit.ly |
| **Recommendation** | ML pipeline, feature store | Netflix, Amazon |
| **Real-time** | WebSocket, pub/sub | Uber, live dashboard |

### Pattern Cheat Sheet

| Pattern | When to Use |
|---------|-------------|
| **Cache-aside** | Read-heavy; cache miss → DB → populate cache |
| **Write-through** | Need cache and DB in sync |
| **Fan-out on write** | Feed with moderate follow graph |
| **Fan-out on read** | Feed with huge follow graph (celebrity) |
| **Sharding** | DB too large for single node |
| **Consistent hashing** | Cache/distribution; minimal rebalancing |
| **Circuit breaker** | Prevent cascade failure |
| **Idempotency** | Exactly-once writes |
| **Event sourcing** | Audit trail, replay |
| **CQRS** | Read/write models differ |

---

## 8. Scaling Strategies

### Database Scaling
- **Read replicas**: Offload reads
- **Sharding**: Partition by user_id, etc.
- **Connection pooling**: Reduce connections
- **Indexing**: Right indexes for queries

### Caching Scaling
- **Multi-level**: L1 (in-process) → L2 (Redis) → DB
- **Partitioning**: Shard Redis by key
- **TTL**: Balance freshness vs hit rate

### Application Scaling
- **Stateless**: No session on server; scale horizontally
- **Async**: Queue heavy work
- **Batch**: Group operations

### Network Scaling
- **CDN**: Static assets, media
- **Compression**: Gzip, Brotli
- **HTTP/2**: Multiplexing

---

## 9. Failure Scenarios

| Scenario | Mitigation |
|----------|------------|
| **Single point of failure** | Replication, multi-AZ |
| **Database down** | Failover, read replicas |
| **Cache down** | Fallback to DB; degrade |
| **Network partition** | Retry, circuit breaker |
| **Thundering herd** | Cache stampede prevention (lock, probabilistic) |
| **Data loss** | Backups, WAL, replication |
| **Region outage** | Multi-region, failover |

---

## 10. Performance Considerations

- **Latency**: Cache, CDN, reduce round trips
- **Throughput**: Async, batch, connection pooling
- **Resource usage**: Connection limits, memory limits
- **Monitoring**: Latency, error rate, saturation

---

## 11. Use Cases

### Template for Common Problems

**Feed/Timeline**
1. Clarify: DAU, posts/user, follows/user, real-time?
2. Estimate: Read QPS >> Write QPS
3. HLD: API, fan-out (write vs read), cache
4. Detail: Feed table schema, cache key design, celebrity handling
5. Scale: Sharding, cache partitioning

**Chat**
1. Clarify: 1:1, group, presence, history retention?
2. Estimate: Messages/user/day, group size
3. HLD: WebSocket, message queue, DB for history
4. Detail: Message ordering, delivery guarantee, offline
5. Scale: Shard by conversation, cache recent

**Search**
1. Clarify: Full-text, filters, autocomplete?
2. Estimate: Queries/day, document volume
3. HLD: Inverted index, Elasticsearch
4. Detail: Indexing pipeline, ranking, typo tolerance
5. Scale: Shard indices, cache hot queries

**Storage (Dropbox-like)**
1. Clarify: Sync, versioning, sharing?
2. Estimate: Files/user, avg size, retention
3. HLD: Blob storage, metadata DB, sync protocol
4. Detail: Dedup, delta sync, conflict resolution
5. Scale: CDN, shard metadata

---

## 12. Comparison Tables

### Database Selection

| Use Case | DB Type | Example |
|----------|---------|---------|
| Relational, transactions | SQL | PostgreSQL, MySQL |
| Key-value, high QPS | NoSQL KV | Redis, DynamoDB |
| Document, flexible schema | Document | MongoDB |
| Wide column | NoSQL | Cassandra |
| Search | Search | Elasticsearch |
| Graph | Graph | Neo4j |

### Caching Strategy

| Strategy | Read Path | Write Path | Use Case |
|----------|-----------|------------|----------|
| Cache-aside | App checks cache; miss → DB → populate | App writes DB; invalidate cache | General |
| Write-through | Read from cache | Write to cache + DB | Strong consistency |
| Write-behind | Read from cache | Write to cache; async to DB | High write throughput |

### Load Balancing

| Algorithm | Use Case |
|-----------|----------|
| Round-robin | Equal capacity |
| Least connections | Variable request time |
| IP hash | Session affinity |
| Weighted | Unequal capacity |

---

## 13. Code / Pseudocode

### Rate Limiter (Token Bucket)

```python
def is_allowed(user_id: str) -> bool:
    key = f"ratelimit:{user_id}"
    now = time.time()
    bucket = redis.get(key)  # {tokens, last_update}
    if not bucket:
        bucket = {"tokens": 100, "last": now}
    else:
        elapsed = now - bucket["last"]
        bucket["tokens"] = min(100, bucket["tokens"] + elapsed * 10)  # refill rate
        bucket["last"] = now
    if bucket["tokens"] >= 1:
        bucket["tokens"] -= 1
        redis.set(key, bucket, ex=60)
        return True
    return False
```

### Feed Fan-Out on Write (Pseudocode)

```python
def post_tweet(user_id: int, content: str):
    tweet_id = db.insert_tweet(user_id, content)
    followers = db.get_followers(user_id)
    for follower_id in followers:
        db.insert_feed(follower_id, tweet_id, timestamp)
    return tweet_id
```

### Idempotency Check

```python
def create_order(idempotency_key: str, order_data: dict):
    existing = db.get_by_idempotency_key(idempotency_key)
    if existing:
        return existing  # Return cached response
    order = db.insert_order(order_data, idempotency_key)
    return order
```

---

## 14. Interview Discussion

### Common Pitfalls and How to Avoid Them

| Pitfall | Avoid By |
|---------|----------|
| **Jumping to solution** | Spend 3-5 min on clarification |
| **No scale numbers** | Always do back-of-envelope |
| **Over-engineering** | Start simple; add complexity when asked |
| **Ignoring failure** | Mention replication, failover, backups |
| **Monologue** | Ask "Should I go deeper here?" |
| **Running out of time** | Leave 5-10 min for follow-ups |
| **Vague tradeoffs** | "We choose X because Y; tradeoff is Z" |
| **Wrong priorities** | Ask what matters: latency vs consistency |

### Behavioral Tips During the Interview

| Do | Don't |
|----|-------|
| Think out loud | Stay silent for minutes |
| Draw diagrams | Only describe verbally |
| Ask clarifying questions | Assume requirements |
| State assumptions | Leave assumptions implicit |
| Prioritize | Try to cover everything |
| Admit unknowns | Fake expertise |
| Engage interviewer | Ignore hints |
| Manage time | Spend 30 min on one component |

### Cheat Sheet: Quick Reference

**Clarification**
- Functional, non-functional, scale, constraints

**Estimation**
- QPS = DAU × req/user/day / 86400
- Storage = ops × bytes × retention × replication

**HLD**
- Client → LB → App → Cache → DB
- Async: Queue → Workers
- CDN for static

**Scaling**
- DB: Replicas, sharding
- Cache: Multi-level, partition
- App: Stateless, horizontal

**Failure**
- Replication, failover, circuit breaker, backups

### Ideal Answer Flow (Summary)

1. **Clarify** (3-5 min): Features, scale, priorities
2. **Estimate** (3-5 min): QPS, storage, bandwidth
3. **HLD** (10-15 min): API, data model, diagram
4. **Detail** (10-15 min): Deep dive on 1-2 components
5. **Scale** (5-10 min): Bottlenecks, solutions, edge cases

### Red Flags to Avoid

- Skipping clarification
- No numbers or estimation
- Single point of failure in design
- Ignoring consistency/availability tradeoffs
- Not asking what to prioritize
- Defensive when challenged
- Giving up when stuck (say "Let me think..." and reason through)

### Follow-Up Questions to Expect

- "What if the cache goes down?"
- "How would you shard the database?"
- "How do you handle the celebrity problem?" (feed)
- "What about exactly-once delivery?"
- "How would you add search?"
- "Design the schema in more detail"
- "What's the bottleneck at 10x scale?"
