# System Design Tradeoffs — Master Reference

> **Staff+ Engineer Level** — Comprehensive guide to every major system design tradeoff for FAANG interviews and production architecture decisions.

---

## Table of Contents

1. [SQL vs NoSQL](#1-sql-vs-nosql)
2. [Vertical vs Horizontal Scaling](#2-vertical-vs-horizontal-scaling)
3. [Consistency vs Availability (CAP)](#3-consistency-vs-availability-cap)
4. [Strong vs Eventual Consistency](#4-strong-vs-eventual-consistency)
5. [Latency vs Throughput](#5-latency-vs-throughput)
6. [Read-Through vs Write-Through Cache](#6-read-through-vs-write-through-cache)
7. [Push vs Pull Architecture](#7-push-vs-pull-architecture)
8. [Synchronous vs Asynchronous Communication](#8-synchronous-vs-asynchronous-communication)
9. [Long Polling vs WebSockets vs SSE](#9-long-polling-vs-websockets-vs-sse)
10. [Batch vs Stream Processing](#10-batch-vs-stream-processing)
11. [Stateful vs Stateless Design](#11-stateful-vs-stateless-design)
12. [Monolith vs Microservices](#12-monolith-vs-microservices)
13. [REST vs RPC (gRPC)](#13-rest-vs-rpc-grpc)
14. [Concurrency vs Parallelism](#14-concurrency-vs-parallelism)
15. [Replication vs Sharding](#15-replication-vs-sharding)

---

## 1. SQL vs NoSQL

### 1.1 Concept Overview

**SQL (Relational Databases):** Structured data stored in tables with rows and columns. Enforces ACID transactions, rigid schemas, and relationships via foreign keys. Examples: PostgreSQL, MySQL, Oracle.

**NoSQL (Non-Relational):** Flexible schema, optimized for specific access patterns. Categories: Document (MongoDB), Key-Value (Redis), Wide-Column (Cassandra), Graph (Neo4j).

### 1.2 Real-World Motivation

- **SQL:** Banking transactions require ACID—money transfer must be atomic. E-commerce order processing needs referential integrity.
- **NoSQL:** Social media feeds need horizontal scale. Session storage needs sub-millisecond reads. Recommendation engines need graph traversals.

### 1.3 Architecture Diagrams (ASCII)

```
SQL (Relational)                    NoSQL (Document)
┌─────────────────────┐            ┌─────────────────────┐
│  Users Table        │            │  users collection   │
│  id | name | email  │            │  { _id, name,       │
└─────────┬───────────┘            │    email, posts[] }  │
          │ FK                     └─────────────────────┘
          ▼
┌─────────────────────┐            ┌─────────────────────┐
│  Orders Table       │            │  orders collection  │
│  id | user_id | amt │            │  { _id, user_id,     │
└─────────────────────┘            │    items[], total }  │
                                   └─────────────────────┘
```

### 1.4 Core Mechanics

| Aspect | SQL | NoSQL |
|--------|-----|-------|
| **Schema** | Fixed, defined upfront | Flexible, schema-on-read |
| **Joins** | Native JOIN operations | Application-level (denormalization) |
| **Transactions** | ACID, multi-row | Limited (single-document or eventual) |
| **Scaling** | Vertical first, sharding complex | Horizontal by design |

### 1.5 Numbers

| Metric | SQL (PostgreSQL) | NoSQL (MongoDB) | NoSQL (Cassandra) |
|--------|------------------|-----------------|-------------------|
| Read latency (simple) | 0.5-2ms | 1-5ms | 1-10ms |
| Write latency | 1-5ms | 1-5ms | 1-15ms |
| Max connections (single node) | ~500-1000 | 10K+ | 10K+ |
| Replication lag | <10ms (sync) | 10-100ms (async) | Tunable |

### 1.6 Tradeoffs — Comparison Table

| Criterion | SQL | NoSQL | When to Choose SQL | When to Choose NoSQL |
|-----------|-----|-------|-------------------|---------------------|
| **Consistency** | Strong (ACID) | Eventual (typically) | Financial, inventory | Social feeds, analytics |
| **Schema evolution** | Migrations required | Flexible | Stable domain model | Rapid iteration |
| **Scaling** | Vertical + complex sharding | Horizontal native | Moderate scale | Massive scale |
| **Query flexibility** | Ad-hoc SQL, complex joins | Predefined access patterns | Complex reporting | Simple key/document lookups |
| **Operational complexity** | Lower (mature tooling) | Higher (tuning per type) | Small team | Large SRE team |

### 1.7 Variants/Implementations

- **SQL:** PostgreSQL (JSONB for hybrid), MySQL (Vitess for sharding), Spanner (distributed SQL)
- **NoSQL:** MongoDB (document), Cassandra (wide-column, AP), DynamoDB (managed key-value), Redis (in-memory)

### 1.8 Scaling Strategies

- **SQL:** Read replicas → read/write split → sharding (by user_id, tenant_id)
- **NoSQL:** Add nodes (Cassandra), partition keys (DynamoDB), replica sets (MongoDB)

### 1.9 Failure Scenarios

- **SQL:** Single point of failure (primary), replication lag causes stale reads
- **NoSQL:** Split-brain in AP systems, eventual consistency can return stale data

### 1.10 Performance Considerations

- **SQL:** Index design critical; N+1 queries kill performance; connection pooling essential
- **NoSQL:** Denormalization increases write cost; hot partitions (DynamoDB) throttle

### 1.11 Use Cases

| SQL | NoSQL |
|-----|-------|
| Banking, payments | Session store, cache |
| E-commerce orders | User profiles, preferences |
| ERP, CRM | Time-series (sensors) |
| Reporting, analytics (OLAP) | Social graph, recommendations |

### 1.12 Comparison Table (Summary)

| | SQL | NoSQL |
|-|-----|-------|
| **Pros** | ACID, mature, joins, tooling | Scale, flexibility, low latency |
| **Cons** | Scaling hard, schema rigidity | No joins, eventual consistency |
| **Real-world** | Stripe (payments), Shopify | Netflix (catalog), Uber (sessions) |

### 1.13 Code/Pseudocode

```python
# SQL: Transaction for money transfer
BEGIN;
  UPDATE accounts SET balance = balance - 100 WHERE id = 'A';
  UPDATE accounts SET balance = balance + 100 WHERE id = 'B';
COMMIT;

# NoSQL: Document update (single document atomic)
db.users.update_one(
  {"_id": user_id},
  {"$push": {"posts": new_post}}
)
```

### 1.14 Interview Discussion

**Key points:** "I choose SQL when I need ACID and complex joins—payments, orders. NoSQL when I need horizontal scale and flexible schema—feeds, sessions. Hybrid is common: PostgreSQL for transactions, Redis for cache, Cassandra for time-series."

---

## 2. Vertical vs Horizontal Scaling

### 2.1 Concept Overview

**Vertical Scaling (Scale Up):** Add more CPU, RAM, disk to a single machine. Simpler but hits physical limits.

**Horizontal Scaling (Scale Out):** Add more machines to distribute load. More complex but theoretically unlimited.

### 2.2 Real-World Motivation

- **Vertical:** Quick fix for traffic spike; small startups; databases that don't shard easily.
- **Horizontal:** Web tier behind load balancer; microservices; distributed databases.

### 2.3 Architecture Diagrams (ASCII)

```
Vertical Scaling              Horizontal Scaling
┌──────────────────┐         ┌─────┐ ┌─────┐ ┌─────┐
│   BIG SERVER     │         │ S1  │ │ S2  │ │ S3  │
│   64 CPU         │         └──┬──┘ └──┬──┘ └──┬──┘
│   256GB RAM      │            └──────┼──────┘
│   2TB SSD        │                   │
└──────────────────┘              ┌────▼────┐
                                  │   LB   │
                                  └────────┘
```

### 2.4 Core Mechanics

- **Vertical:** Upgrade instance type (e.g., m5.large → m5.4xlarge). Single point of failure.
- **Horizontal:** Add nodes; load balancer distributes; stateless design required for app tier.

### 2.5 Numbers

| Scaling Type | Cost (2x capacity) | Complexity | Limit |
|--------------|---------------------|------------|-------|
| Vertical | ~1.5-2x (non-linear) | Low | ~TB RAM, ~100s vCPU |
| Horizontal | ~2x (linear) | High | Effectively unlimited |

### 2.6 Tradeoffs — Comparison Table

| Criterion | Vertical | Horizontal | Choose Vertical | Choose Horizontal |
|-----------|-----------|------------|-----------------|-------------------|
| **Cost** | Diminishing returns | Linear | Small scale | Large scale |
| **Complexity** | Low | High (LB, discovery, state) | MVP, small team | Production at scale |
| **Single point of failure** | Yes | No (with redundancy) | Acceptable for dev | Never for prod |
| **Database** | Easier | Sharding, replication | Monolithic DB | Distributed DB |
| **Deployment** | Downtime for upgrade | Rolling, zero-downtime | Low traffic | High availability |

### 2.7 Variants/Implementations

- **Vertical:** AWS RDS resize, GCP machine type change
- **Horizontal:** Kubernetes HPA, auto-scaling groups, database read replicas

### 2.8 Scaling Strategies

- Start vertical for simplicity; plan horizontal when approaching limits.
- Database: vertical for primary, horizontal for read replicas, then sharding.

### 2.9 Failure Scenarios

- **Vertical:** Machine failure = total outage; upgrade requires downtime.
- **Horizontal:** Partial failure; must handle node failures gracefully.

### 2.10 Performance Considerations

- **Vertical:** NUMA, memory bandwidth can bottleneck before CPU.
- **Horizontal:** Network latency between nodes; data locality matters.

### 2.11 Use Cases

| Vertical | Horizontal |
|----------|------------|
| Dev/staging | Web servers |
| Small DB | API servers |
| Legacy apps | Microservices |
| Quick capacity fix | Production at scale |

### 2.12 Comparison Table (Summary)

| | Vertical | Horizontal |
|-|----------|------------|
| **Pros** | Simple, no code changes | Unlimited scale, fault tolerance |
| **Cons** | Limit, SPOF, downtime | Complexity, distributed systems |
| **Real-world** | Small RDS instance | Netflix, Uber (thousands of nodes) |

### 2.13 Code/Pseudocode

```yaml
# Vertical: Change instance type
instance_type: m5.4xlarge  # was m5.large

# Horizontal: Kubernetes HPA
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
spec:
  minReplicas: 3
  maxReplicas: 100
  targetCPUUtilizationPercentage: 70
```

### 2.14 Interview Discussion

**Key points:** "Vertical first for speed; horizontal when we hit limits or need HA. Web tier is always horizontal; database often vertical until we need sharding. Cost: vertical has diminishing returns—doubling CPU doesn't double throughput."

---

## 3. Consistency vs Availability (CAP)

### 3.1 Concept Overview

**CAP Theorem:** In a distributed system under partition, you can have at most 2 of: **C**onsistency, **A**vailability, **P**artition tolerance. Partition tolerance is unavoidable in real networks, so the choice is effectively **CP** vs **AP**.

**CP Systems:** Sacrifice availability during partition to maintain consistency (e.g., ZooKeeper, etcd, HBase).

**AP Systems:** Sacrifice consistency during partition to remain available (e.g., Cassandra, DynamoDB, CouchDB).

### 3.2 Real-World Motivation

- **CP:** Configuration store (etcd), leader election—wrong config could cause split-brain.
- **AP:** Social feed, product catalog—better to show slightly stale data than error page.

### 3.3 Architecture Diagrams (ASCII)

```
CP (Consistency + Partition Tolerance)     AP (Availability + Partition Tolerance)
┌─────────┐     P     ┌─────────┐           ┌─────────┐     P     ┌─────────┐
│ Node A  │◄─────────►│ Node B  │           │ Node A  │◄─────────►│ Node B  │
│ Primary │  (split)  │ Replica │           │ v=2     │  (split)  │ v=1     │
└────┬────┘           └────┬────┘           └────┬────┘           └────┬────┘
     │                     │                     │                     │
     │ Reject writes       │ Reject reads        │ Accept writes       │ Serve reads
     │ until sync          │ (stale)              │ (divergence OK)     │ (stale OK)
```

### 3.4 Core Mechanics

- **CP:** On partition, one partition continues (usually with majority); minority blocks or returns errors.
- **AP:** All nodes accept reads/writes; conflict resolution (last-write-wins, vector clocks) when partition heals.

### 3.5 Numbers

| System | CAP Choice | Partition Behavior | Typical Latency |
|--------|------------|-------------------|-----------------|
| etcd | CP | Minority unavailable | 1-10ms |
| Cassandra | AP | All nodes available | 5-20ms |
| MongoDB (default) | CP | Primary only | 1-5ms |
| DynamoDB | Configurable | Tunable | 1-10ms |

### 3.6 Tradeoffs — Comparison Table

| Criterion | CP | AP | Choose CP | Choose AP |
|-----------|-----|-----|------------|------------|
| **During partition** | Unavailable (minority) | Available, possibly stale | Config, locks, coordination | User-facing reads |
| **Conflict resolution** | Not needed (single source) | Required (LWW, CRDTs) | Critical correctness | Best-effort OK |
| **Use case** | Distributed lock, config | Feed, catalog, session |
| **Examples** | etcd, ZooKeeper, HBase | Cassandra, DynamoDB, CouchDB |

### 3.7 Variants/Implementations

- **CP:** Raft (etcd), Paxos (Chubby), ZAB (ZooKeeper)
- **AP:** Dynamo-style (Cassandra), CRDTs (Riak), Conflict-free replicated types

### 3.8 Scaling Strategies

- **CP:** Add nodes to quorum; odd number (3, 5, 7) for majority.
- **AP:** Add nodes freely; more replicas = more availability, more divergence risk.

### 3.9 Failure Scenarios

- **CP:** Network partition → minority partition rejects requests; possible unavailability.
- **AP:** Partition → both sides accept writes; merge conflicts when healed.

### 3.10 Performance Considerations

- **CP:** Latency = round-trip to quorum; cross-region adds latency.
- **AP:** Local reads fast; writes may need quorum; read-your-writes not guaranteed.

### 3.11 Use Cases

| CP | AP |
|----|-----|
| Service discovery | Social feed |
| Distributed locks | Product catalog |
| Leader election | Session store |
| Configuration | Analytics |

### 3.12 Comparison Table (Summary)

| | CP | AP |
|-|-----|-----|
| **Pros** | Strong consistency, no conflicts | High availability, low latency |
| **Cons** | Unavailable during partition | Stale reads, conflict resolution |
| **Real-world** | Kubernetes (etcd), HBase | Cassandra (Netflix), DynamoDB (Amazon) |

### 3.13 Code/Pseudocode

```python
# CP: etcd - block until consensus
client.put("config", value, timeout=5)  # Fails if partition

# AP: Cassandra - always accept
session.execute("INSERT INTO users ...")  # Succeeds, may diverge
```

### 3.14 Interview Discussion

**Key points:** "CAP is about partition—we always have P. So we choose CP or AP. CP for coordination (locks, config); AP for user-facing data where availability beats perfect consistency. Many systems offer tunable consistency (DynamoDB's read consistency levels)."

---

## 4. Strong vs Eventual Consistency

### 4.1 Concept Overview

**Strong Consistency:** Every read returns the most recent write. Linearizability: operations appear to occur in a total order.

**Eventual Consistency:** If no new writes, all replicas eventually converge. Reads may return stale data temporarily.

### 4.2 Real-World Motivation

- **Strong:** Bank balance, inventory count—stale read = wrong decision.
- **Eventual:** Social like count, view count—eventual convergence acceptable.

### 4.3 Architecture Diagrams (ASCII)

```
Strong Consistency                 Eventual Consistency
Write → Replicate → Read           Write → Async Replicate
┌─────┐    sync     ┌─────┐        ┌─────┐   async    ┌─────┐
│  A  │────────────►│  B  │        │  A  │───────────►│  B  │
│ v=2 │   ack       │ v=2 │        │ v=2 │  (later)   │ v=1 │
└─────┘             └─────┘        └─────┘            └─────┘
Read returns v=2                   Read may return v=1
```

### 4.4 Core Mechanics

- **Strong:** Synchronous replication; read from primary or quorum; higher latency.
- **Eventual:** Asynchronous replication; read from any replica; lower latency, possible staleness.

### 4.5 Numbers

| Model | Read Latency | Write Latency | Replication Lag |
|-------|--------------|---------------|-----------------|
| Strong | 1-2 RTT (quorum) | 2 RTT (sync) | 0 |
| Eventual | 1 RTT (local) | 1 RTT | 10ms-seconds |

### 4.6 Tradeoffs — Comparison Table

| Criterion | Strong | Eventual | Choose Strong | Choose Eventual |
|-----------|--------|----------|----------------|-----------------|
| **Correctness** | Always latest | May be stale | Money, inventory | Counts, feeds |
| **Latency** | Higher (sync) | Lower (async) | Critical correctness | User experience |
| **Availability** | Lower (sync blocks) | Higher | Acceptable | Must stay up |
| **Complexity** | Simpler (no conflicts) | Conflict resolution | - | - |

### 4.7 Variants/Implementations

- **Strong:** PostgreSQL sync replication, Spanner, CockroachDB
- **Eventual:** Cassandra, DynamoDB (default), MongoDB (secondary reads)

### 4.8 Scaling Strategies

- **Strong:** Quorum reads/writes; more replicas = more latency.
- **Eventual:** Add replicas; read from nearest; monitor replication lag.

### 4.9 Failure Scenarios

- **Strong:** Replica failure can block writes if below quorum.
- **Eventual:** Replication lag can cause prolonged staleness; split-brain divergence.

### 4.10 Performance Considerations

- **Strong:** Cross-region = high latency (150ms+ RTT).
- **Eventual:** Local reads; optimize for read path.

### 4.11 Use Cases

| Strong | Eventual |
|--------|----------|
| Payments | Social likes |
| Inventory | View counts |
| Configuration | Caches |
| Leader election | CDN metadata |

### 4.12 Comparison Table (Summary)

| | Strong | Eventual |
|-|--------|----------|
| **Pros** | Correctness, no stale reads | Low latency, high availability |
| **Cons** | Higher latency, lower availability | Stale reads, conflicts |
| **Real-world** | Stripe (payments) | Instagram (likes) |

### 4.13 Code/Pseudocode

```python
# Strong: Read from primary
balance = db.query("SELECT balance FROM accounts WHERE id=?", read_from="primary")

# Eventual: Read from replica
likes = db.query("SELECT COUNT(*) FROM likes WHERE post_id=?", read_from="replica")
```

### 4.14 Interview Discussion

**Key points:** "Strong when correctness is critical—payments, inventory. Eventual when we optimize for latency and availability—counts, feeds. Many systems offer both: DynamoDB strong vs eventual read; PostgreSQL primary vs replica."

---

## 5. Latency vs Throughput

### 5.1 Concept Overview

**Latency:** Time for a single operation to complete (e.g., 10ms per request).

**Throughput:** Operations per unit time (e.g., 10,000 QPS).

### 5.2 Real-World Motivation

- **Latency-critical:** Search (user waits), trading (milliseconds matter), real-time games.
- **Throughput-critical:** Batch processing, log ingestion, analytics—total work matters more than per-item speed.

### 5.3 Architecture Diagrams (ASCII)

```
Latency-Optimized                    Throughput-Optimized
┌─────────────────────────┐         ┌─────────────────────────────────┐
│ Request → Cache → DB    │         │ Request → Queue → Workers (N)    │
│ (minimize path)         │         │ (maximize parallelism)           │
│ p99: 10ms               │         │ 100K msg/sec                      │
└─────────────────────────┘         └─────────────────────────────────┘
```

### 5.4 Core Mechanics

- **Latency:** Reduce round-trips, use cache, optimize critical path, avoid blocking.
- **Throughput:** Batch operations, parallelize, pipeline, increase concurrency.

### 5.5 Numbers

| Optimization | Latency Impact | Throughput Impact |
|---------------|----------------|-------------------|
| Add cache | -90% (cache hit) | +10x (fewer DB hits) |
| Batch 100 writes | +50ms (wait) | +50x (amortize overhead) |
| Add 10 workers | Minimal | +10x |
| Connection pooling | -5ms | +2x |

### 5.6 Tradeoffs — Comparison Table

| Criterion | Latency Focus | Throughput Focus | Optimize Latency | Optimize Throughput |
|-----------|---------------|------------------|------------------|---------------------|
| **Metric** | p50, p99, p999 | QPS, messages/sec | User-facing APIs | Batch, background |
| **Techniques** | Cache, reduce hops | Batching, parallelism | Real-time | Data pipelines |
| **Tradeoff** | May sacrifice throughput | May increase latency | Search, trading | ETL, logging |
| **Example** | Google Search <100ms | Kafka 1M msg/sec | - | - |

### 5.7 Variants/Implementations

- **Latency:** CDN, edge compute, read replicas, connection pooling
- **Throughput:** Kafka, batch inserts, worker pools, async processing

### 5.8 Scaling Strategies

- **Latency:** Geographic distribution, edge caching, reduce serialization.
- **Throughput:** Horizontal scaling, partitioning, backpressure handling.

### 5.9 Failure Scenarios

- **Latency:** Cache stampede can spike latency; slow DB query blocks all.
- **Throughput:** Queue backup; workers overwhelmed; backpressure needed.

### 5.10 Performance Considerations

- **Latency:** Tail latencies (p99) matter; one slow request affects UX.
- **Throughput:** Saturation point; need monitoring for queue depth.

### 5.11 Use Cases

| Latency-Critical | Throughput-Critical |
|------------------|---------------------|
| Search | Log aggregation |
| Trading | Video transcoding |
| Chat | Analytics ETL |
| Gaming | Event ingestion |

### 5.12 Comparison Table (Summary)

| | Latency | Throughput |
|-|---------|------------|
| **Pros** | Responsive UX | Process more data |
| **Cons** | May limit parallelism | May add delay |
| **Real-world** | Google (<100ms) | Kafka (millions/sec) |

### 5.13 Code/Pseudocode

```python
# Latency: Single fast path
result = cache.get(key) or db.get(key)

# Throughput: Batch
batch = collect_batch(timeout=100ms, max_size=1000)
db.bulk_insert(batch)
```

### 5.14 Interview Discussion

**Key points:** "Latency for user-facing—p99 matters. Throughput for batch—total volume. Often trade off: batching improves throughput but adds latency. Need to know which SLA we're optimizing for."

---

## 6. Read-Through vs Write-Through Cache

### 6.1 Concept Overview

**Read-Through:** Cache sits in front of DB. On miss, cache fetches from DB, stores, returns. Application treats cache as transparent.

**Write-Through:** On write, update cache and DB together. Cache always has latest. Application writes to cache; cache writes to DB.

**Write-Back (Write-Behind):** Write to cache only; async flush to DB. Higher risk of data loss.

### 6.2 Real-World Motivation

- **Read-Through:** Read-heavy (product catalog); cache populated on demand.
- **Write-Through:** Strong consistency; cache and DB in sync.
- **Write-Back:** High write throughput; accept risk (e.g., session data).

### 6.3 Architecture Diagrams (ASCII)

```
Read-Through                         Write-Through
┌─────┐  miss   ┌─────┐  fetch   ┌─────┐    ┌─────┐  write   ┌─────┐
│ App │────────►│Cache│─────────►│ DB  │    │ App │─────────►│Cache│
└─────┘  hit    └─────┘  store   └─────┘    └─────┘  sync    └──┬──┘
         ◄──────         ◄──────                    write       │
         return         return                     ┌──▼──┐      │
                                                   │ DB  │◄─────┘
                                                   └─────┘
```

### 6.4 Core Mechanics

- **Read-Through:** Lazy loading; cache miss triggers DB read; good for sparse access.
- **Write-Through:** Every write goes to both; cache never stale; higher write latency.

### 6.5 Numbers

| Strategy | Read Latency (hit) | Write Latency | Consistency |
|----------|-------------------|---------------|-------------|
| Read-Through | 0.1ms | N/A | Eventual |
| Write-Through | 0.1ms | 2x (cache + DB) | Strong |
| Write-Back | 0.1ms | 0.1ms (cache only) | Eventual (risk) |

### 6.6 Tradeoffs — Comparison Table

| Criterion | Read-Through | Write-Through | Choose Read-Through | Choose Write-Through |
|-----------|--------------|---------------|---------------------|----------------------|
| **Data freshness** | Stale until TTL/miss | Always fresh | Read-heavy, TTL OK | Strong consistency |
| **Write load** | Low (lazy) | High (every write) | Few writes | Critical reads |
| **Cache population** | On demand | Proactive | Sparse access | Hot data |
| **Complexity** | Simpler | Cache invalidation | - | - |

### 6.7 Variants/Implementations

- **Read-Through:** Redis with DB fallback, Memcached
- **Write-Through:** Application writes to both; or cache library (e.g., Caffeine with loader)
- **Write-Back:** Kafka + consumer; async flush

### 6.8 Scaling Strategies

- **Read-Through:** Add cache nodes; TTL tuning; cache warming for known hot keys.
- **Write-Through:** Write coalescing; batch DB writes.

### 6.9 Failure Scenarios

- **Read-Through:** Cache miss storm (thundering herd) if many requests miss same key.
- **Write-Through:** Cache down = every read hits DB; write to cache fails = inconsistent.

### 6.10 Performance Considerations

- **Read-Through:** Cache hit ratio critical; 80%+ for benefit.
- **Write-Through:** Doubles write path; consider write-behind for high write volume.

### 6.11 Use Cases

| Read-Through | Write-Through |
|--------------|---------------|
| Product catalog | User session |
| API responses | Configuration |
| Search results | Feature flags |

### 6.12 Comparison Table (Summary)

| | Read-Through | Write-Through |
|-|--------------|---------------|
| **Pros** | Lazy load, simple | Strong consistency |
| **Cons** | Stale reads, miss penalty | Higher write latency |
| **Real-world** | CDN, API cache | Session store |

### 6.13 Code/Pseudocode

```python
# Read-Through
def get(key):
    val = cache.get(key)
    if val is None:
        val = db.get(key)
        cache.set(key, val, ttl=300)
    return val

# Write-Through
def set(key, val):
    cache.set(key, val)
    db.set(key, val)
```

### 6.14 Interview Discussion

**Key points:** "Read-through for read-heavy, lazy load. Write-through when cache must match DB. Write-back for high write volume but we accept risk. Often combine: read-through with TTL, write-through for critical data."

---

## 7. Push vs Pull Architecture

### 7.1 Concept Overview

**Push:** Producer sends data to consumer. Producer controls flow. Examples: WebSocket, server-sent events, Kafka push.

**Pull:** Consumer requests data from producer. Consumer controls flow. Examples: REST polling, Kafka consumer pull.

### 7.2 Real-World Motivation

- **Push:** Real-time notifications, chat—minimize delay.
- **Pull:** Consumer-paced processing, backpressure—consumer controls rate.

### 7.3 Architecture Diagrams (ASCII)

```
Push                                    Pull
┌────────┐     push      ┌────────┐     ┌────────┐   poll    ┌────────┐
│Producer│──────────────►│Consumer│     │Consumer│──────────►│Producer│
└────────┘                └────────┘     └────────┘   fetch   └────────┘
Producer controls flow                   Consumer controls rate
```

### 7.4 Core Mechanics

- **Push:** Producer initiates; consumer must handle burst; backpressure needed.
- **Pull:** Consumer pulls when ready; natural backpressure; may add latency.

### 7.5 Numbers

| Model | Latency | Backpressure | Scalability |
|-------|---------|--------------|-------------|
| Push | Lower (immediate) | Hard (producer may overwhelm) | Consumer scaling tricky |
| Pull | Higher (poll interval) | Natural | Consumer scales by pulling |

### 7.6 Tradeoffs — Comparison Table

| Criterion | Push | Pull | Choose Push | Choose Pull |
|-----------|------|------|--------------|-------------|
| **Latency** | Lower | Higher (poll delay) | Real-time | Batch processing |
| **Backpressure** | Must implement | Built-in | - | Variable consumer speed |
| **Consumer scaling** | Complex | Simple (add consumers) | Few consumers | Many consumers |
| **Producer load** | High (fan-out) | Low | - | High fan-out |
| **Example** | WebSocket, SSE | Kafka, SQS | Chat, notifications | Event processing |

### 7.7 Variants/Implementations

- **Push:** WebSocket, SSE, gRPC streaming, fan-out
- **Pull:** Kafka, SQS, REST polling, long polling

### 7.8 Scaling Strategies

- **Push:** Fan-out to multiple consumers; queue per consumer; rate limiting.
- **Pull:** Partition; each consumer pulls from partition; add partitions for parallelism.

### 7.9 Failure Scenarios

- **Push:** Consumer slow = producer blocks or drops; need buffering.
- **Pull:** Consumer down = no pull; messages wait (Kafka retention).

### 7.10 Performance Considerations

- **Push:** Producer must handle slow consumers; backpressure protocols (HTTP/2, gRPC).
- **Pull:** Polling overhead; batch pulls reduce round-trips (Kafka fetch.min.bytes).

### 7.11 Use Cases

| Push | Pull |
|------|------|
| Chat | Log processing |
| Notifications | Analytics pipeline |
| Live dashboards | Event sourcing |
| Gaming | Data lake ingestion |

### 7.12 Comparison Table (Summary)

| | Push | Pull |
|-|------|------|
| **Pros** | Low latency, immediate | Backpressure, scalable |
| **Cons** | Backpressure hard | Polling latency |
| **Real-world** | Slack (WebSocket) | Kafka (pull) |

### 7.13 Code/Pseudocode

```python
# Push: WebSocket
ws.send(json.dumps({"type": "message", "data": msg}))

# Pull: Kafka
messages = consumer.poll(timeout_ms=1000)
for msg in messages:
    process(msg)
```

### 7.14 Interview Discussion

**Key points:** "Push for real-time—chat, notifications. Pull for processing—Kafka, SQS. Pull gives natural backpressure; push needs careful design. Kafka chose pull: consumers control rate, easier to scale."

---

## 8. Synchronous vs Asynchronous Communication

### 8.1 Concept Overview

**Synchronous:** Caller blocks until response. Request-response pattern. Tight coupling.

**Asynchronous:** Caller sends and continues. Response via callback, future, or separate channel. Loose coupling.

### 8.2 Real-World Motivation

- **Sync:** Simple flow, need immediate result (auth check, validation).
- **Async:** Long-running tasks, decoupling, resilience (order processing, notifications).

### 8.3 Architecture Diagrams (ASCII)

```
Synchronous                              Asynchronous
┌─────┐ request  ┌─────┐ response ┌─────┐  ┌─────┐  msg   ┌─────┐
│ A   │────────►│ B   │────────►│ A   │  │ A   │───────►│Queue│
└─────┘  wait   └─────┘  return └─────┘  └─────┘        └──┬──┘
  blocks                                              │
                                                      ▼
                                                 ┌─────┐
                                                 │ B   │ process
                                                 └─────┘
```

### 8.4 Core Mechanics

- **Sync:** HTTP request, gRPC unary; caller holds connection.
- **Async:** Message queue, event bus; fire-and-forget or callback.

### 8.5 Numbers

| Model | Latency (per call) | Throughput | Coupling |
|-------|-------------------|------------|----------|
| Sync | RTT + processing | Limited by slowest | High |
| Async | Queue + process | High (decoupled) | Low |

### 8.6 Tradeoffs — Comparison Table

| Criterion | Sync | Async | Choose Sync | Choose Async |
|-----------|------|-------|--------------|--------------|
| **Coupling** | Tight (A knows B) | Loose (A knows queue) | Simple flow | Microservices |
| **Latency** | Blocks on B | A continues | Need immediate result | Can defer |
| **Failure** | B down = A fails | B down = queue buffers | - | Resilience |
| **Complexity** | Simple | Retry, ordering, idempotency | - | - |
| **Example** | REST, gRPC | Kafka, SQS, events | Auth, validation | Order processing |

### 8.7 Variants/Implementations

- **Sync:** REST, gRPC, GraphQL
- **Async:** Kafka, RabbitMQ, SQS, SNS, event-driven

### 8.8 Scaling Strategies

- **Sync:** Load balance; circuit breaker; timeout.
- **Async:** Scale consumers; partition queue; dead-letter queue.

### 8.9 Failure Scenarios

- **Sync:** Cascading failure; B slow = A blocks.
- **Async:** Message loss (if not durable); ordering; poison messages.

### 8.10 Performance Considerations

- **Sync:** Connection pooling; keep-alive; avoid N+1 sync calls.
- **Async:** Batch messages; consumer prefetch; backpressure.

### 8.11 Use Cases

| Sync | Async |
|------|-------|
| API gateway → service | Order → fulfillment |
| Auth check | Email sending |
| Validation | Analytics events |
| Simple CRUD | Notification fan-out |

### 8.12 Comparison Table (Summary)

| | Sync | Async |
|-|------|-------|
| **Pros** | Simple, immediate | Decoupled, resilient |
| **Cons** | Coupling, blocking | Complexity, eventual |
| **Real-world** | REST APIs | Event-driven (Uber, Netflix) |

### 8.13 Code/Pseudocode

```python
# Sync
response = requests.post(url, json=payload)
result = response.json()

# Async
queue.send(message)
# ... later, consumer processes
```

### 8.14 Interview Discussion

**Key points:** "Sync when we need immediate result—auth, validation. Async for decoupling and resilience—order processing, notifications. Async adds complexity: retries, idempotency, ordering. Use sync for critical path, async for everything else."

---

## 9. Long Polling vs WebSockets vs SSE

### 9.1 Concept Overview

**Long Polling:** Client sends request; server holds until data or timeout; client reconnects. HTTP-based.

**WebSockets:** Full-duplex, persistent connection. Bidirectional. Protocol upgrade.

**SSE (Server-Sent Events):** Server pushes to client over HTTP. Unidirectional (server→client). Native EventSource API.

### 9.2 Real-World Motivation

- **Long Polling:** Fallback when WebSocket not available; simple.
- **WebSocket:** Chat, gaming, collaborative editing—bidirectional.
- **SSE:** Notifications, live feed—server push only, simpler than WebSocket.

### 9.3 Architecture Diagrams (ASCII)

```
Long Polling              WebSockets              SSE
┌────┐ request ┌────┐      ┌────┐ ◄══════► ┌────┐  ┌────┐ ──────► ┌────┐
│ C  │───────►│ S  │      │ C  │  full   │ S  │  │ C  │  push   │ S  │
└────┘        └────┘      └────┘  duplex  └────┘  └────┘  only    └────┘
   │ hold        │           │              │        │
   │◄────────────│ response   │              │        │
   │ reconnect   │           │              │        │
```

### 9.4 Core Mechanics

- **Long Polling:** Request → hold → response → repeat. Overhead per poll.
- **WebSocket:** HTTP upgrade → persistent TCP; frames both ways.
- **SSE:** HTTP connection; server sends `data:` lines; auto-reconnect.

### 9.5 Numbers

| Method | Overhead | Latency | Browser Support | Bidirectional |
|--------|----------|---------|------------------|---------------|
| Long Polling | High (new request each) | 1-2 RTT per poll | All | Yes (via new request) |
| WebSocket | Low (after upgrade) | 1 RTT | All modern | Yes |
| SSE | Low | 1 RTT | All modern (except IE) | No (client→server needs separate) |

### 9.6 Tradeoffs — Comparison Table

| Criterion | Long Polling | WebSocket | SSE | Choose Long Poll | Choose WebSocket | Choose SSE |
|-----------|--------------|-----------|-----|------------------|------------------|------------|
| **Overhead** | High | Low | Low | Legacy support | Real-time bidirectional | Server push only |
| **Latency** | Poll interval | Immediate | Immediate | - | Chat, gaming | Notifications |
| **Complexity** | Low | Medium | Low | - | - | Simpler than WS |
| **Firewall** | HTTP works | May block WS | HTTP works | - | - | - |
| **Example** | Old chat | Slack, Discord | Stock ticker | - | - | - |

### 9.7 Variants/Implementations

- **Long Polling:** Comet, BOSH (XMPP)
- **WebSocket:** Socket.io (fallback to polling), native WS
- **SSE:** EventSource API, polyfills for IE

### 9.8 Scaling Strategies

- **Long Polling:** Sticky sessions; connection limits per server.
- **WebSocket:** Sticky sessions; connection limits; Redis pub/sub for multi-server.
- **SSE:** Same as WebSocket for scaling.

### 9.9 Failure Scenarios

- **Long Polling:** Timeout too short = many requests; too long = stale.
- **WebSocket:** Connection drop; need reconnect logic.
- **SSE:** Auto-reconnect built-in; need id for resume.

### 9.10 Performance Considerations

- **Long Polling:** Each poll = new HTTP; headers overhead.
- **WebSocket:** One connection; efficient for frequent messages.
- **SSE:** One connection; text-based; efficient for push.

### 9.11 Use Cases

| Long Polling | WebSocket | SSE |
|--------------|-----------|-----|
| Legacy fallback | Chat | Notifications |
| Simple push | Gaming | Live feed |
| - | Collaborative edit | Stock ticker |

### 9.12 Comparison Table (Summary)

| | Long Polling | WebSocket | SSE |
|-|--------------|-----------|-----|
| **Pros** | Simple, works everywhere | Low latency, bidirectional | Simple, auto-reconnect |
| **Cons** | High overhead | More complex | One-way only |
| **Real-world** | Fallback | Slack, Zoom | Twitter feed |

### 9.13 Code/Pseudocode

```javascript
// Long Polling
async function poll() {
  const res = await fetch('/updates');
  const data = await res.json();
  handle(data);
  poll(); // recurse
}

// WebSocket
const ws = new WebSocket('wss://...');
ws.onmessage = (e) => handle(JSON.parse(e.data));

// SSE
const es = new EventSource('/stream');
es.onmessage = (e) => handle(JSON.parse(e.data));
```

### 9.14 Interview Discussion

**Key points:** "WebSocket for bidirectional—chat, gaming. SSE for server push only—simpler, notifications. Long polling as fallback. SSE has auto-reconnect; WebSocket needs custom. Consider sticky sessions for stateful connections."

---

## 10. Batch vs Stream Processing

### 10.1 Concept Overview

**Batch:** Process data in chunks at scheduled intervals. High throughput, higher latency. Examples: Hadoop, Spark batch.

**Stream:** Process data as it arrives. Low latency, more complex. Examples: Kafka Streams, Flink, Spark Streaming.

### 10.2 Real-World Motivation

- **Batch:** Nightly reports, ETL, analytics—delay acceptable.
- **Stream:** Fraud detection, alerting, real-time dashboards—immediate action.

### 10.3 Architecture Diagrams (ASCII)

```
Batch Processing                     Stream Processing
┌─────────┐     ┌─────────┐          ┌─────────┐     ┌─────────┐
│ Source  │────►│  Queue  │          │ Source  │────►│  Stream │
└─────────┘     └────┬────┘          └─────────┘     └────┬────┘
                    │  collect                              │
                    │  (e.g., hourly)                       │  process
                    ▼                                       │  per event
              ┌─────────┐                                   ▼
              │ Batch   │                            ┌─────────┐
              │ Job     │                            │ Sink    │
              └─────────┘                            └─────────┘
```

### 10.4 Core Mechanics

- **Batch:** Accumulate → process in bulk → output. Optimized for throughput.
- **Stream:** Process each event (or micro-batch); stateful windows; exactly-once semantics complex.

### 10.5 Numbers

| Model | Latency | Throughput | Complexity |
|-------|---------|------------|------------|
| Batch | Minutes-hours | Very high | Lower |
| Stream | Milliseconds-seconds | High | Higher (state, windows) |

### 10.6 Tradeoffs — Comparison Table

| Criterion | Batch | Stream | Choose Batch | Choose Stream |
|-----------|-------|--------|--------------|---------------|
| **Latency** | High (schedule) | Low | Reports, ETL | Fraud, alerts |
| **Throughput** | Very high | High | Large datasets | Real-time |
| **Complexity** | Lower | Higher (state, late data) | - | - |
| **Cost** | Burst (cluster) | Steady (always on) | - | - |
| **Example** | Hadoop, Spark | Flink, Kafka Streams | - | - |

### 10.7 Variants/Implementations

- **Batch:** Hadoop MapReduce, Spark batch, Airflow DAGs
- **Stream:** Kafka Streams, Flink, Spark Streaming, Samza

### 10.8 Scaling Strategies

- **Batch:** Add workers; partition input; tune batch size.
- **Stream:** Scale partitions; scale consumers; backpressure.

### 10.9 Failure Scenarios

- **Batch:** Job failure = reprocess batch; partial output handling.
- **Stream:** Late data; exactly-once; state recovery.

### 10.10 Performance Considerations

- **Batch:** I/O bound; optimize shuffle; compression.
- **Stream:** Checkpointing overhead; state size; watermark tuning.

### 10.11 Use Cases

| Batch | Stream |
|-------|--------|
| Nightly ETL | Fraud detection |
| Report generation | Real-time dashboards |
| ML training | Alerting |
| Data lake ingestion | CEP (complex event processing) |

### 10.12 Comparison Table (Summary)

| | Batch | Stream |
|-|-------|--------|
| **Pros** | High throughput, simpler | Low latency, real-time |
| **Cons** | High latency | Complex, state |
| **Real-world** | Netflix (recommendations batch) | Uber (fraud detection) |

### 10.13 Code/Pseudocode

```python
# Batch
for chunk in read_in_batches(path, size=10000):
    process(chunk)
    write(output, chunk)

# Stream
for event in kafka_consumer:
    result = process(event)
    emit(result)
```

### 10.14 Interview Discussion

**Key points:** "Batch for throughput when latency OK—ETL, reports. Stream when we need real-time—fraud, alerts. Lambda architecture combines both. Many systems do micro-batch (e.g., Spark Streaming) as compromise."

---

## 11. Stateful vs Stateless Design

### 11.1 Concept Overview

**Stateful:** Server stores client state (session, context). Request routing must be sticky.

**Stateless:** Server stores no client state. Each request self-contained. Any server can handle any request.

### 11.2 Real-World Motivation

- **Stateful:** WebSocket connections, in-memory sessions, some legacy apps.
- **Stateless:** REST APIs, microservices—scale freely, no sticky routing.

### 11.3 Architecture Diagrams (ASCII)

```
Stateful                              Stateless
┌─────┐     sticky      ┌─────┐       ┌─────┐  any server  ┌─────┐
│Client│───────────────►│ S1  │       │Client│────────────►│ S1  │
└─────┘  (session on S1)└─────┘       └─────┘              └─────┘
   │                        │            │  or
   │                        │            └────────────────►│ S2  │
   │                        │                              └─────┘
   │    must return to S1    │            any server works
```

### 11.4 Core Mechanics

- **Stateful:** Session stored in memory/DB; load balancer uses cookie/hash for sticky routing.
- **Stateless:** Token/session ID in request; state in DB/Redis; server looks up.

### 11.5 Numbers

| Design | Scaling | Failover | Load Balance |
|--------|---------|----------|--------------|
| Stateful | Hard (session migration) | Session loss | Sticky required |
| Stateless | Easy (add nodes) | No session loss | Round-robin OK |

### 11.6 Tradeoffs — Comparison Table

| Criterion | Stateful | Stateless | Choose Stateful | Choose Stateless |
|-----------|----------|-----------|-----------------|------------------|
| **Scaling** | Sticky, complex | Add nodes freely | WebSocket, legacy | APIs, microservices |
| **Failover** | Session loss | Seamless | - | - |
| **Memory** | Per-connection state | Shared (Redis/DB) | - | - |
| **Load balance** | Sticky | Any | - | - |
| **Example** | WebSocket server | REST API | - | - |

### 11.7 Variants/Implementations

- **Stateful:** Sticky sessions, session replication (Hazelcast), external store (Redis)
- **Stateless:** JWT, session in Redis, database-backed session

### 11.8 Scaling Strategies

- **Stateful:** Session replication; or move state to Redis (becomes stateless from server POV).
- **Stateless:** Horizontal scaling; auto-scaling; no special routing.

### 11.9 Failure Scenarios

- **Stateful:** Server crash = lost sessions; reconnection required.
- **Stateless:** Redis/DB down = auth fails; but no in-memory state loss.

### 11.10 Performance Considerations

- **Stateful:** In-memory fast; but limits scaling.
- **Stateless:** Redis lookup per request (~0.1ms); negligible.

### 11.11 Use Cases

| Stateful | Stateless |
|----------|-----------|
| WebSocket server | REST API |
| Legacy app | Microservices |
| In-memory cache (per node) | CDN, API gateway |
| Gaming server | Serverless |

### 11.12 Comparison Table (Summary)

| | Stateful | Stateless |
|-|----------|-----------|
| **Pros** | Fast (in-memory) | Scalable, resilient |
| **Cons** | Scaling, failover | External state lookup |
| **Real-world** | WebSocket (sticky) | Netflix, Uber APIs |

### 11.13 Code/Pseudocode

```python
# Stateful
session = server_sessions[client_id]  # in-memory
session["last_action"] = action

# Stateless
session = redis.get(f"session:{token}")
session["last_action"] = action
redis.set(f"session:{token}", session)
```

### 11.14 Interview Discussion

**Key points:** "Stateless for scalability—any server can handle any request. Stateful when we must—WebSocket. Move state to Redis/DB to make servers stateless. Sticky sessions are an anti-pattern for scale."

---

## 12. Monolith vs Microservices

### 12.1 Concept Overview

**Monolith:** Single deployable unit. All components in one codebase, one process. Shared database.

**Microservices:** Independently deployable services. Each owns its domain. Communicate via API/events.

### 12.2 Real-World Motivation

- **Monolith:** Startup speed, small team, simple deployment.
- **Microservices:** Large team, independent scaling, fault isolation, polyglot.

### 12.3 Architecture Diagrams (ASCII)

```
Monolith                              Microservices
┌─────────────────────────┐           ┌─────┐ ┌─────┐ ┌─────┐
│  ┌─────┐ ┌─────┐ ┌─────┐│           │ A   │ │ B   │ │ C   │
│  │Auth │ │Order│ │User ││           └──┬──┘ └──┬──┘ └──┬──┘
│  └──┬──┘ └──┬──┘ └──┬──┘│              │       │       │
│     └───────┼───────┘    │              ▼       ▼       ▼
│            │            │           ┌─────┐ ┌─────┐ ┌─────┐
│      ┌─────▼─────┐      │           │ DB1 │ │ DB2 │ │ DB3 │
│      │   DB      │      │           └─────┘ └─────┘ └─────┘
│      └───────────┘      │
└─────────────────────────┘
```

### 12.4 Core Mechanics

- **Monolith:** Single build, single deploy; function calls; shared transactions.
- **Microservices:** Independent deploy; network calls; distributed transactions (avoid).

### 12.5 Numbers

| Aspect | Monolith | Microservices |
|--------|-----------|---------------|
| Deploy frequency | Weekly | Daily/hourly |
| Team size | 1-10 | 10-100s |
| Latency (in-process) | μs | ms (network) |
| Operational cost | Lower | Higher (observability, mesh) |

### 12.6 Tradeoffs — Comparison Table

| Criterion | Monolith | Microservices | Choose Monolith | Choose Microservices |
|-----------|----------|---------------|-----------------|----------------------|
| **Complexity** | Low | High | Small team | Large team |
| **Deployment** | All or nothing | Independent | - | - |
| **Scaling** | Vertical/whole app | Per service | - | - |
| **Fault isolation** | Process crash = all down | Service crash isolated | - | - |
| **Team** | Single team | Multiple teams | <10 engineers | 50+ engineers |
| **Example** | Shopify (started mono) | Netflix, Uber | - | - |

### 12.7 Variants/Implementations

- **Monolith:** Modular monolith (bounded contexts in one codebase)
- **Microservices:** Service mesh (Istio), API gateway, event-driven

### 12.8 Scaling Strategies

- **Monolith:** Scale entire app; or extract hot paths to services (strangler fig).
- **Microservices:** Scale each service independently; use async for decoupling.

### 12.9 Failure Scenarios

- **Monolith:** Bug in one module can crash entire app.
- **Microservices:** Network failures; cascading; need circuit breaker, retries.

### 12.10 Performance Considerations

- **Monolith:** In-process calls; single DB connection pool.
- **Microservices:** Network hop adds latency; serialization; N+1 service calls.

### 12.11 Use Cases

| Monolith | Microservices |
|----------|---------------|
| MVP, startup | Large scale |
| Single team | Multiple teams |
| Simple domain | Complex domain |
| - | Polyglot needs |

### 12.12 Comparison Table (Summary)

| | Monolith | Microservices |
|-|----------|---------------|
| **Pros** | Simple, fast iteration | Scale, fault isolation |
| **Cons** | Scaling, deployment | Complexity, ops |
| **Real-world** | Basecamp, early Shopify | Netflix, Amazon |

### 12.13 Code/Pseudocode

```python
# Monolith
def create_order(user_id, items):
    user = get_user(user_id)  # in-process
    order = Order.create(user, items)
    inventory.decrement(items)  # in-process
    return order

# Microservices
def create_order(user_id, items):
    user = user_service.get(user_id)  # HTTP/gRPC
    order = order_service.create(user, items)
    inventory_service.decrement(items)  # async event
    return order
```

### 12.14 Interview Discussion

**Key points:** "Start monolith; extract services when team or scale demands. Microservices need: observability, service mesh, distributed tracing. Don't microservice too early—complexity cost is real. Modular monolith is a middle ground."

---

## 13. REST vs RPC (gRPC)

### 13.1 Concept Overview

**REST:** Resource-oriented over HTTP. Verbs (GET, POST), JSON/XML. Stateless. Cacheable.

**RPC (gRPC):** Procedure-oriented. Call remote functions. HTTP/2, Protocol Buffers. Strongly typed.

### 13.2 Real-World Motivation

- **REST:** Public APIs, browser clients, caching, wide tooling.
- **gRPC:** Internal services, performance, streaming, polyglot with codegen.

### 13.3 Architecture Diagrams (ASCII)

```
REST                                    gRPC
GET /users/123                    UserService.GetUser(123)
POST /orders {items: [...]}        OrderService.CreateOrder(order)
┌─────┐  HTTP/1.1  ┌─────┐        ┌─────┐  HTTP/2   ┌─────┐
│Client│───────────►│Server│        │Client│─────────►│Server│
└─────┘  JSON      └─────┘        └─────┘  Protobuf └─────┘
```

### 13.4 Core Mechanics

- **REST:** URL = resource; method = action; status codes; HATEOAS (optional).
- **gRPC:** Service definition (.proto); binary serialization; streaming (client, server, bidirectional).

### 13.5 Numbers

| Metric | REST (JSON) | gRPC (Protobuf) |
|--------|-------------|-----------------|
| Payload size | ~2-3x larger | Compact binary |
| Latency | Higher (text) | Lower (binary) |
| Throughput | Good | Better (HTTP/2 multiplexing) |
| Browser support | Native | Need gRPC-Web proxy |

### 13.6 Tradeoffs — Comparison Table

| Criterion | REST | gRPC | Choose REST | Choose gRPC |
|-----------|------|------|-------------|-------------|
| **Performance** | Good | Better | Public API | Internal, performance |
| **Tooling** | Mature (Postman, curl) | Codegen, less UI | - | - |
| **Browser** | Native | gRPC-Web needed | Web clients | - |
| **Streaming** | Limited (SSE) | Native (bidirectional) | - | - |
| **Schema** | OpenAPI (optional) | .proto required | - | - |
| **Example** | Stripe API | Google internal | - | - |

### 13.7 Variants/Implementations

- **REST:** OpenAPI, JSON:API, HATEOAS
- **gRPC:** gRPC-Web, grpc-gateway (REST→gRPC)

### 13.8 Scaling Strategies

- **REST:** CDN caching, HTTP caching headers.
- **gRPC:** Connection reuse (HTTP/2); load balancing (client-side).

### 13.9 Failure Scenarios

- **REST:** Verbose errors; status codes.
- **gRPC:** Error codes; status details; streaming errors.

### 13.10 Performance Considerations

- **REST:** JSON parsing; consider MessagePack for internal.
- **gRPC:** Protobuf fast; HTTP/2 multiplexing; connection pooling.

### 13.11 Use Cases

| REST | gRPC |
|------|------|
| Public API | Internal microservices |
| Mobile/web clients | Service-to-service |
| CRUD | High-throughput |
| - | Streaming (e.g., logs) |

### 13.12 Comparison Table (Summary)

| | REST | gRPC |
|-|------|------|
| **Pros** | Simple, cacheable, tooling | Fast, streaming, typed |
| **Cons** | Verbose, no streaming | Browser support |
| **Real-world** | Stripe, GitHub API | Google, Netflix internal |

### 13.13 Code/Pseudocode

```python
# REST
GET /users/123
{"id": 123, "name": "Alice"}

# gRPC
service UserService {
  rpc GetUser(GetUserRequest) returns (User);
}
# Binary payload, typed
```

### 13.14 Interview Discussion

**Key points:** "REST for public APIs—browser, caching, tooling. gRPC for internal—performance, streaming, codegen. Many use both: gRPC internally, REST gateway for external. gRPC-Web bridges browser gap."

---

## 14. Concurrency vs Parallelism

### 14.1 Concept Overview

**Concurrency:** Multiple tasks in progress; may interleave on single core. About structure and design.

**Parallelism:** Multiple tasks executing simultaneously on multiple cores. About execution.

### 14.2 Real-World Motivation

- **Concurrency:** Handle many connections (async I/O); structure for responsiveness.
- **Parallelism:** Speed up computation (multi-core); process more data.

### 14.3 Architecture Diagrams (ASCII)

```
Concurrency (1 core)                 Parallelism (multi-core)
Task A ──┬──┬────┬──                 Core 1: Task A ─────────
Task B ──┴──┴────┴──                 Core 2: Task B ─────────
(interleaved)                        Core 3: Task C ─────────
```

### 14.4 Core Mechanics

- **Concurrency:** Async/await, goroutines, event loop—one thread, many tasks.
- **Parallelism:** Threads, processes, distributed—multiple cores/nodes.

### 14.5 Numbers

| Model | Speedup | Use Case |
|-------|---------|----------|
| Concurrency | No (I/O bound) | Network servers |
| Parallelism | ~N cores (CPU bound) | Image processing |

### 14.6 Tradeoffs — Comparison Table

| Criterion | Concurrency | Parallelism | Choose Concurrency | Choose Parallelism |
|-----------|-------------|-------------|-------------------|---------------------|
| **Goal** | Responsiveness | Throughput | I/O bound | CPU bound |
| **Cores** | 1 can suffice | Needs multiple | - | - |
| **Example** | Node.js, Go | MapReduce, Ray | Web server | Video encode |
| **Complexity** | Async patterns | Race conditions, sync | - | - |

### 14.7 Variants/Implementations

- **Concurrency:** async/await (Python, JS), goroutines (Go), green threads
- **Parallelism:** multiprocessing, Ray, Spark, GPU

### 14.8 Scaling Strategies

- **Concurrency:** More connections per process; epoll, kqueue.
- **Parallelism:** Add cores; partition work; reduce synchronization.

### 14.9 Failure Scenarios

- **Concurrency:** Deadlock (if blocking); callback hell.
- **Parallelism:** Race conditions; false sharing.

### 14.10 Performance Considerations

- **Concurrency:** Context switch overhead minimal for I/O.
- **Parallelism:** Amdahl's law; sync overhead.

### 14.11 Use Cases

| Concurrency | Parallelism |
|-------------|-------------|
| Web server | Video transcoding |
| Chat | ML training |
| API gateway | Batch processing |
| - | Scientific compute |

### 14.12 Comparison Table (Summary)

| | Concurrency | Parallelism |
|-|-------------|-------------|
| **Pros** | Handle many I/O | Speed up compute |
| **Cons** | No speedup for CPU | Sync complexity |
| **Real-world** | Nginx, Node | Spark, TensorFlow |

### 14.13 Code/Pseudocode

```python
# Concurrency (async I/O)
async def handle_requests():
    await asyncio.gather(
        fetch(url1),
        fetch(url2),
    )

# Parallelism (multi-core)
with ThreadPoolExecutor(8) as pool:
    results = pool.map(process, items)
```

### 14.14 Interview Discussion

**Key points:** "Concurrency is structure—many tasks in progress. Parallelism is execution—simultaneous. Concurrency enables parallelism but not vice versa. I/O bound: concurrency. CPU bound: parallelism. Go has both: goroutines (concurrency) + GOMAXPROCS (parallelism)."

---

## 15. Replication vs Sharding

### 15.1 Concept Overview

**Replication:** Copy data across nodes. Same data, multiple copies. For availability, read scaling.

**Sharding:** Partition data across nodes. Different data per node. For write scaling, storage.

### 15.2 Real-World Motivation

- **Replication:** Read replicas for read-heavy; failover; geo-distribution.
- **Sharding:** Write scaling; storage limits; reduce per-node load.

### 15.3 Architecture Diagrams (ASCII)

```
Replication (same data)              Sharding (partitioned data)
┌─────┐ ┌─────┐ ┌─────┐              ┌─────┐ ┌─────┐ ┌─────┐
│ P   │ │ R1  │ │ R2  │              │Shard1│ │Shard2│ │Shard3│
│ A,B │ │ A,B │ │ A,B │              │ A,D │ │ B,E │ │ C,F │
└─────┘ └─────┘ └─────┘              └─────┘ └─────┘ └─────┘
Same data everywhere                 Different data per shard
```

### 15.4 Core Mechanics

- **Replication:** Primary-replica; sync or async; read from replica.
- **Sharding:** Partition key (user_id, tenant_id); routing; rebalancing.

### 15.5 Numbers

| Strategy | Read scaling | Write scaling | Storage |
|----------|--------------|---------------|---------|
| Replication | +Nx (N replicas) | 1x | Same data × N |
| Sharding | +Nx | +Nx | Total / N per shard |

### 15.6 Tradeoffs — Comparison Table

| Criterion | Replication | Sharding | Choose Replication | Choose Sharding |
|-----------|-------------|----------|---------------------|-----------------|
| **Read scaling** | Add replicas | Add shards | Read-heavy | Write-heavy |
| **Write scaling** | No (single primary) | Yes | - | - |
| **Storage** | Duplicated | Partitioned | - | Large dataset |
| **Complexity** | Lower | Higher (routing, rebalance) | - | - |
| **Failover** | Promote replica | Per-shard replication | - | - |
| **Can combine?** | Yes | Yes | - | - |

### 15.7 Variants/Implementations

- **Replication:** MySQL replica, PostgreSQL streaming, Cassandra replication factor
- **Sharding:** Vitess, Citus, application-level (user_id % N)

### 15.8 Scaling Strategies

- **Replication:** Add read replicas; cross-region for DR.
- **Sharding:** Add shards; rebalance; consistent hashing for minimal movement.

### 15.9 Failure Scenarios

- **Replication:** Replication lag; split-brain if misconfigured.
- **Sharding:** Hot shard; rebalancing complexity; cross-shard queries hard.

### 15.10 Performance Considerations

- **Replication:** Read from replica = eventual consistency; sync = write latency.
- **Sharding:** Partition key design critical; avoid cross-shard joins.

### 15.11 Use Cases

| Replication | Sharding |
|-------------|----------|
| Read scaling | Write scaling |
| Failover | Storage scaling |
| Geo-distribution | Multi-tenant |
| - | Large tables |

### 15.12 Comparison Table (Summary)

| | Replication | Sharding |
|-|-------------|----------|
| **Pros** | HA, read scale | Write scale, storage |
| **Cons** | Write bottleneck | Complexity |
| **Real-world** | Every DB | Twitter, Instagram |
| **Combine** | Replicate each shard | Common pattern |

### 15.13 Code/Pseudocode

```python
# Replication: read from replica
db = replica_pool.get_connection()  # read
primary_db.execute("INSERT ...")    # write

# Sharding: route by partition key
shard_id = hash(user_id) % num_shards
db = shards[shard_id]
```

### 15.14 Interview Discussion

**Key points:** "Replication for availability and read scaling. Sharding for write scaling and storage. We often do both: shard for scale, replicate each shard for HA. Replication doesn't help writes; sharding does. Choose shard key carefully—avoid hot spots."

---

## Summary: Quick Reference

| # | Tradeoff | Quick Rule |
|---|----------|------------|
| 1 | SQL vs NoSQL | ACID/joins → SQL; scale/flex → NoSQL |
| 2 | Vertical vs Horizontal | Start vertical; horizontal for scale |
| 3 | CAP | Coordination → CP; user data → AP |
| 4 | Strong vs Eventual | Money → strong; counts → eventual |
| 5 | Latency vs Throughput | User-facing → latency; batch → throughput |
| 6 | Read vs Write cache | Read-heavy → read-through; consistency → write-through |
| 7 | Push vs Pull | Real-time → push; processing → pull |
| 8 | Sync vs Async | Immediate result → sync; decouple → async |
| 9 | Poll vs WS vs SSE | Bidirectional → WS; push only → SSE |
| 10 | Batch vs Stream | ETL → batch; real-time → stream |
| 11 | Stateful vs Stateless | WebSocket → stateful; API → stateless |
| 12 | Monolith vs Microservices | Small team → mono; scale → micro |
| 13 | REST vs gRPC | Public → REST; internal → gRPC |
| 14 | Concurrency vs Parallelism | I/O → concurrency; CPU → parallelism |
| 15 | Replication vs Sharding | Read/HA → replication; write/storage → sharding |

---

*Document: Staff+ System Design Tradeoffs — 800+ lines | FAANG Interview Reference*
