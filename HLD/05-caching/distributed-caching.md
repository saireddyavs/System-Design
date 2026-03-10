# Distributed Caching: Staff+ Engineer Deep Dive

## 1. Concept Overview

### Definition
Distributed caching is a caching architecture where **cache data is spread across multiple nodes** in a cluster, providing horizontal scalability beyond single-machine memory limits. Clients access the cache through a unified interface while data is partitioned and optionally replicated across nodes.

### Purpose
- **Scale beyond single node**: Single machine RAM is limited (typically 64-512GB); distributed cache can reach TBs
- **High availability**: Replication and failover when nodes fail
- **Geographic distribution**: Edge caches closer to users
- **Load distribution**: Spread read/write load across many nodes

### Why It Exists
A single Redis instance can hold ~25GB (practical limit). Facebook serves 2B+ users—product catalog, sessions, and feeds require 100s of TB of cache. Distributed caching is the only way to achieve this scale.

### Problems It Solves
1. **Memory ceiling**: Single node limits
2. **Single point of failure**: Node failure = total cache loss
3. **Network bottleneck**: Single node I/O limit
4. **Geographic latency**: Users far from single datacenter

---

## 2. Real-World Motivation

### Twitter
- **Redis + Memcached**: Redis for real-time (timelines, counters), Memcached for general object cache
- **Timeline service**: Fan-out on write; cached per user
- **Manhattan**: Custom distributed key-value store for metadata

### Facebook
- **Memcached**: Trillions of operations per day (paper: "Scaling Memcache at Facebook")
- **TAO (The Associations and Objects)**: Read-through cache over MySQL, 99.99% hit rate
- **Mcrouter**: Proxy layer for Memcached, consistent hashing, replication

### Netflix
- **EVCache**: Memcached-based, multi-region replication
- **EVCache**: 100ms p99 latency, 30% reduction in Cassandra load
- **Zuul**: Edge caching for API responses

### Instagram
- **Redis**: Feed ranking, sessions, real-time features
- **Redis Cluster**: Sharded across hundreds of nodes

### Amazon
- **ElastiCache**: Managed Redis/Memcached
- **DynamoDB DAX**: Read-through cache for DynamoDB
- **CloudFront**: Edge cache (CDN) for static assets

---

## 3. Architecture Diagrams

### Distributed Cache Topology

```
                    DISTRIBUTED CACHE ARCHITECTURE
    ┌─────────────────────────────────────────────────────────────────┐
    │                         CLIENT LAYER                              │
    │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐             │
    │  │  App 1  │  │  App 2  │  │  App 3  │  │  App N  │             │
    │  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘             │
    │       │            │            │            │                    │
    │       └────────────┴─────┬──────┴────────────┘                    │
    │                         │                                         │
    │                    ┌────▼────┐                                    │
    │                    │  Proxy  │  (Optional: Mcrouter, Twemproxy)   │
    │                    │  Layer  │                                    │
    │                    └────┬────┘                                    │
    │                         │                                         │
    │       ┌─────────────────┼─────────────────┐                      │
    │       │                 │                 │                      │
    │  ┌────▼────┐       ┌────▼────┐       ┌────▼────┐                 │
    │  │ Node 1  │       │ Node 2  │       │ Node N  │   CACHE LAYER   │
    │  │ Redis   │       │ Redis   │       │ Redis   │                 │
    │  │ 32GB    │       │ 32GB    │       │ 32GB    │                 │
    │  └─────────┘       └─────────┘       └─────────┘                 │
    │       │                 │                 │                      │
    │       └─────────────────┼─────────────────┘                      │
    │                         │                                         │
    │                    ┌────▼────┐                                    │
    │                    │   DB   │         PERSISTENCE LAYER           │
    │                    └────────┘                                    │
    └─────────────────────────────────────────────────────────────────┘
```

### Consistent Hashing

```
                    CONSISTENT HASHING RING
    ┌─────────────────────────────────────────────────────────┐
    │                     Hash Ring (0 to 2^32-1)              │
    │                                                           │
    │    N1 ●──────────● N2                                     │
    │      ╲            ╱                                        │
    │       ╲   K1    ╱   K2                                    │
    │        ╲  ●    ╱   ●                                       │
    │         ╲    ╱                                             │
    │          ╲  ╱                                              │
    │           ╲╱                                               │
    │            ● N3                                            │
    │           ╱ ╲                                              │
    │          ╱   ╲   K3                                        │
    │         ╱     ╲  ●                                         │
    │        ╱       ╲                                           │
    │       ╱         ╲                                          │
    │    N4 ●──────────● N5                                      │
    │                                                           │
    │  Key hashes to point on ring; clockwise to next node       │
    │  Virtual nodes (vnodes): N1a, N1b, N1c... for balance      │
    └─────────────────────────────────────────────────────────┘
```

### Redis Cluster Sharding

```
                    REDIS CLUSTER (16,384 SLOTS)
    ┌─────────────────────────────────────────────────────────────┐
    │  Slot 0-5460    Slot 5461-10922   Slot 10923-16383          │
    │  ┌─────────┐    ┌─────────┐      ┌─────────┐               │
    │  │Master 1 │    │Master 2 │      │Master 3 │               │
    │  │  M1     │    │  M2     │      │  M3     │               │
    │  └────┬────┘    └────┬────┘      └────┬────┘               │
    │       │              │                │                     │
    │  ┌────▼────┐    ┌────▼────┐      ┌────▼────┐               │
    │  │Replica 1│    │Replica 2│      │Replica 3│               │
    │  │  R1     │    │  R2     │      │  R3     │               │
    │  └─────────┘    └─────────┘      └─────────┘               │
    │                                                             │
    │  Key → CRC16(key) mod 16384 → Slot → Node                   │
    │  Hash tags: {user:123} → same slot for user:123:profile     │
    └─────────────────────────────────────────────────────────────┘
```

### Cache Stampede (Thundering Herd)

```
                    CACHE STAMPEDE SCENARIO
    ┌─────────────────────────────────────────────────────────────┐
    │  T=0: Cache key "user:123" expires                           │
    │                                                             │
    │  T=1ms: 10,000 concurrent requests for user:123             │
    │         ┌───┐ ┌───┐ ┌───┐ ┌───┐ ... ┌───┐                    │
    │         │ 1 │ │ 2 │ │ 3 │ │ 4 │     │10K│   All MISS        │
    │         └─┬─┘ └─┬─┘ └─┬─┘ └─┬─┘     └─┬─┘                    │
    │           │     │     │     │          │                      │
    │           └─────┴─────┴─────┴──────────┘                      │
    │                         │                                    │
    │  T=2ms: ALL 10,000 hit DB simultaneously  ← THUNDERING HERD  │
    │                    ┌────▼────┐                               │
    │                    │   DB    │  Overload!                    │
    │                    └─────────┘                               │
    │                                                             │
    │  SOLUTION: Lock (single flight) - only 1 fetches             │
    └─────────────────────────────────────────────────────────────┘
```

### Cache Penetration vs. Avalanche vs. Stampede

```
    CACHE PENETRATION          CACHE AVALANCHE         CACHE STAMPEDE
    (Non-existent keys)        (Mass expiration)      (Hot key expiry)

    Request for key that        Many keys expire       One hot key expires
    doesn't exist in DB         at same time           → many requests miss

    ┌─────┐                    ┌─────┐                ┌─────┐
    │ App │───miss──────────▶   │ App │───miss─────────▶│ App │
    └──┬──┘                    └──┬──┘                └──┬──┘
       │   ┌─────┐   miss         │   ┌─────┐  miss      │   ┌─────┐
       └──▶│Cache│───────────────┼──▶│Cache│────────────┼──▶│Cache│
           └──┬──┘               │   └──┬──┘            │   └──┬──┘
              │   always miss     │      │  all miss     │      │  all miss
              ▼                   ▼      ▼               ▼      ▼
           ┌─────┐             ┌─────┐ ┌─────┐         ┌─────┐ ┌─────┐
           │ DB  │             │ DB  │ │ DB  │         │ DB  │ │ DB  │
           └─────┘             └─────┘ └─────┘         └─────┘ └─────┘
           Key not in DB       DB overload             DB overload
           (attack or bug)     (bad TTL design)        (concurrent access)

    Fix: Bloom filter,         Fix: Randomize TTL,     Fix: Lock, probabilistic
         null caching               multi-level cache      early expiration
```

---

## 4. Core Mechanics

### Why Distributed Cache?

**Single Machine Limits:**
- **Memory**: 64-512GB typical; 2TB max for large instances
- **Network**: 10-100 Gbps per node
- **CPU**: Single-threaded (Redis) or multi-threaded (Memcached) bottleneck
- **Failure domain**: One node = total loss

**Distributed Benefits:**
- **Linear scaling**: 10 nodes = 10x memory, 10x throughput
- **Fault tolerance**: Replication; node failure = partial impact
- **Geographic**: Deploy in multiple regions

### Redis Architecture

**Single-Threaded Event Loop:**
- One thread handles all commands (no locking)
- Non-blocking I/O (epoll/kqueue)
- Blocking operations (KEYS, FLUSHDB) stall entire server
- Pipelining: batch commands, reduce RTT

**Data Structures:**
- Strings, Hashes, Lists, Sets, Sorted Sets, Bitmaps, HyperLogLog, Streams
- Each has O(1) or O(log n) operations
- Memory optimized: ziplist for small collections, intset for small sets

**Persistence:**
- **RDB (snapshot)**: Fork, copy-on-write, dump to file. Point-in-time backup. Restore fast.
- **AOF (append-only file)**: Log every write. Durable. Replay on restart. Can be large.
- **Hybrid**: RDB + AOF incremental (Redis 4.0+)

**Redis Cluster:**
- 16,384 hash slots
- Key → CRC16(key) mod 16384 → slot
- Hash tags: `{user:123}` ensures user:123:profile, user:123:settings on same node
- Gossip protocol for cluster state
- Redirect: Client gets MOVED/ASK, retries to correct node

**Redis Sentinel:**
- High availability for non-cluster setup
- Monitors master; promotes replica on failure
- Client discovers master via Sentinel

### Memcached Architecture

**Multi-Threaded:**
- One thread per core; lock per slab class (not global)
- Higher throughput for multi-core (vs. Redis single-threaded)
- No persistence (by design)—cache only

**Slab Allocation:**
- Pre-allocate memory in size classes (1.25x growth: 64B, 80B, 100B, ...)
- No fragmentation within slab
- Eviction per slab class (can't evict 64B to make room for 1KB)
- Slab reassignment possible but expensive

**Protocol:**
- Text and binary protocols
- Simple: get, set, delete, incr, decr
- No data structures (just key-value blobs)

### Consistent Hashing

**Problem:** Modulo hashing (key % N) causes mass rehashing when N changes (node add/remove).

**Solution:** Hash nodes and keys to a ring (0 to 2^32-1). Key maps to clockwise next node.

**Virtual Nodes (vnodes):** Each physical node has 100-200 virtual positions. Balances distribution when nodes have different capacities.

**Implementation:**
```
hash(key) → point on ring
clockwise_next(point) → node
```

### Cache Partitioning Strategies

| Strategy | Description | Rebalance on Add/Remove |
|----------|-------------|-------------------------|
| Modulo | key % N | High (1/N keys move) |
| Consistent Hashing | Ring | Low (~K/N keys move) |
| Redis Cluster | 16K slots | Low (only affected slots) |
| Range | key range per node | Medium |

---

## 5. Numbers

### Throughput

| System | Ops/Sec (per node) | Latency (p99) | Notes |
|--------|-------------------|---------------|-------|
| Redis | 100K-150K | 1-5ms | Single-threaded, pipelining |
| Memcached | 200K-500K | 0.5-2ms | Multi-threaded |
| Redis Cluster | 1M+ (10 nodes) | 1-5ms | Linear scaling |
| Facebook Memcached | Trillions/day | <1ms | Thousands of nodes |

### Memory

| System | Typical Node | Max Key Size | Max Value Size |
|--------|--------------|--------------|----------------|
| Redis | 4-32GB | 512MB | 512MB |
| Memcached | 4-64GB | 250B (default) | 1MB (default) |
| Redis Cluster | 32GB × N | 512MB | 512MB |

### Scale

- **Facebook**: 1.5TB+ Memcached, trillions of ops/day
- **Twitter**: Hundreds of Redis nodes
- **Netflix EVCache**: Multi-region, 100ms p99
- **Instagram**: Redis Cluster for feed

### Cache Hit Ratios

- **Target**: 90-99% for read-heavy
- **Facebook TAO**: 99.99%
- **Typical**: 95% with good key design and TTL

---

## 6. Tradeoffs

### Redis vs. Memcached

| Aspect | Redis | Memcached |
|--------|-------|-----------|
| **Threading** | Single | Multi |
| **Data structures** | Rich (hash, list, set, sorted set) | Key-value only |
| **Persistence** | RDB, AOF | None |
| **Replication** | Yes (async) | No (client-side) |
| **Eviction** | Configurable (LRU, LFU, etc.) | LRU per slab |
| **Max value** | 512MB | 1MB (configurable) |
| **Use case** | Sessions, queues, real-time | Simple object cache |
| **Throughput** | 100K/sec | 200K+/sec |

### Consistency vs. Availability

| Choice | Consistency | Availability |
|--------|--------------|--------------|
| Sync replication | Strong | Lower (wait for replica) |
| Async replication | Eventual | Higher |
| No replication | N/A | Single point of failure |

### Partitioning vs. Replication

- **Partitioning**: More memory, higher throughput, no redundancy
- **Replication**: Redundancy, read scaling, failover
- **Both**: Redis Cluster (partitioned + replicated)

---

## 7. Variants / Implementations

### Redis Variants

- **Redis Cluster**: Sharded, replicated, built-in
- **Redis Sentinel**: Master-replica with HA
- **Twemproxy**: Proxy for Redis/Memcached, consistent hashing
- **Codis**: Redis proxy with dynamic sharding
- **KeyDB**: Multi-threaded Redis fork

### Memcached Variants

- **Mcrouter**: Facebook's proxy, consistent hashing, replication
- **EVCache**: Netflix's Memcached, multi-region
- **Twemproxy**: Also supports Memcached

### Cache Proxies

| Proxy | Use Case | Features |
|-------|----------|----------|
| Twemproxy | Simple sharding | Consistent hashing, no persistence |
| Mcrouter | Facebook-scale | Replication, failover, pools |
| Envoy | Service mesh | Redis filter, dynamic config |

---

## 8. Scaling Strategies

### Horizontal Scaling
- Add nodes; consistent hashing minimizes rebalance
- Redis Cluster: Add node, reshard slots (manual or automatic)
- Memcached: Add node, clients rehash (or proxy handles)

### Vertical Scaling
- Larger instances (more RAM)
- Diminishing returns: single-threaded Redis doesn't use more cores
- Memcached benefits from more cores

### Multi-Level Caching
```
Request → L1 (local, 1ms) → L2 (Redis cluster, 5ms) → DB (50ms)
```
- L1: Process-local (Caffeine, Guava)
- L2: Distributed (Redis, Memcached)
- L1 reduces load on L2

### Cache Warming
- **Eager**: Pre-populate at startup (top N keys)
- **Lazy**: Populate on first access
- **Predictive**: ML to predict hot keys, pre-fetch

---

## 9. Failure Scenarios

### Cache Stampede (Thundering Herd)

**Scenario:** Hot key expires. 10,000 requests miss simultaneously. All hit DB.

**Solutions:**
1. **Lock (single flight):** First request acquires lock, fetches, populates cache. Others wait or retry.
2. **Probabilistic early expiration:** 1% chance to refresh when 50% through TTL. Spreads load.
3. **Request coalescing:** Merge concurrent requests for same key into one fetch.
4. **Stale-while-revalidate:** Return stale, refresh in background.

### Cache Avalanche

**Scenario:** Many keys expire at same time (e.g., all set with same TTL at midnight). Mass miss.

**Solutions:**
1. **Randomize TTL:** TTL = base + random(0, 300) seconds
2. **Stagger expiration:** Don't set all keys at once
3. **Multi-level cache:** L1 absorbs some load when L2 expires

### Cache Penetration

**Scenario:** Requests for keys that don't exist (attack or bug). Every request misses cache and hits DB.

**Solutions:**
1. **Bloom filter:** Check key existence before DB. False positives possible; false negatives impossible.
2. **Null caching:** Cache "null" or sentinel for missing keys. Short TTL (e.g., 60s).
3. **Validation:** Reject invalid key patterns at API layer.

### Hot Key Problem

**Scenario:** One key (e.g., celebrity profile) gets 1M reads/sec. Single Redis node can't handle.

**Solutions:**
1. **Local cache:** L1 cache in app (short TTL)
2. **Replication:** Multiple replicas for read; round-robin or consistent hashing with vnodes
3. **Key splitting:** Shard key (e.g., user:123:shard0, user:123:shard1); aggregate on read
4. **Write-through to multiple keys:** Replicate hot key to N keys; read from random

### Dogpile Effect

**Same as cache stampede.** Multiple names for same problem.

### Node Failure

**Scenario:** Cache node dies. Data lost (if no replication). Requests fail or rehash.

**Solutions:**
1. **Replication:** Redis replication, Memcached replication (Mcrouter)
2. **Failover:** Sentinel for Redis; proxy for Memcached
3. **Degradation:** Fall through to DB; cache will repopulate

### Split Brain

**Scenario:** Network partition. Two nodes think they're master. Data divergence.

**Solutions:**
1. **Quorum:** Redis Cluster requires majority for failover
2. **Fencing:** Ensure only one master can accept writes
3. **Consistent hashing:** Clients may route to wrong partition; eventual consistency

---

## 10. Performance Considerations

### Latency

| Operation | Local | Same DC | Cross-Region |
|-----------|-------|---------|--------------|
| Cache get | 0.1ms | 1-2ms | 50-100ms |
| Cache set | 0.1ms | 1-2ms | 50-100ms |
| DB get | 1-10ms | 5-20ms | 50-200ms |

### Pipelining

- **Without pipelining:** 1 RTT per command. 1000 gets = 1000 RTTs.
- **With pipelining:** Batch 1000 gets in one request. 1 RTT.
- **Redis:** PIPELINE or pipeline() in clients
- **Memcached:** get_multi, multi-get

### Connection Pooling

- Reuse TCP connections; avoid connect/disconnect per request
- Pool size: 2-4 × number of app threads
- Too large: memory; too small: connection wait

### Serialization

- **JSON:** Human-readable, slower
- **MessagePack:** Binary, faster
- **Protobuf:** Schema, compact
- **Java Kryo:** Fast, Java-specific

---

## 11. Use Cases

| Use Case | Cache | Strategy | Notes |
|----------|-------|----------|-------|
| Session | Redis | Write-through | Persistence, TTL |
| Product catalog | Redis/Memcached | Read-through | High read |
| API response | Redis | Cache-aside + TTL | Stale OK |
| Rate limiting | Redis | Increment + TTL | Atomic |
| Leaderboard | Redis Sorted Set | Write-through | Real-time |
| Feed | Redis | Cache-aside | Fan-out |
| Counters | Redis/Memcached | Increment | Atomic |
| Full-text search | Redis/Elasticsearch | Cache-aside | Complex queries |

---

## 12. Comparison Tables

### Cache Systems

| Feature | Redis | Memcached | Hazelcast | Couchbase |
|---------|-------|-----------|-----------|------------|
| Data model | Rich | Key-value | Key-value + compute | Key-value + N1QL |
| Persistence | Yes | No | Yes | Yes |
| Clustering | Built-in | Client/proxy | Built-in | Built-in |
| Replication | Async | No | Sync/async | Yes |
| Use case | General | Simple cache | In-memory compute | Document store |

### Problem-Solution Matrix

| Problem | Solution | Implementation |
|---------|----------|----------------|
| Stampede | Lock | Redis SETNX, lease |
| Stampede | Probabilistic | β*exp(-λ*t) for refresh prob |
| Penetration | Bloom filter | Redis Bloom module, client-side |
| Penetration | Null cache | Cache empty result, TTL 60s |
| Avalanche | Random TTL | TTL + random(0, 300) |
| Hot key | Local cache | Caffeine, Guava |
| Hot key | Replication | Multiple replicas |
| Node failure | Replication | Redis replica, Sentinel |

---

## 13. Code / Pseudocode

### Cache Stampede Prevention (Lock)

```python
def get_with_lock(key, ttl=3600):
    value = cache.get(key)
    if value is not None:
        return value

    lock_key = f"lock:{key}"
    if cache.set(lock_key, "1", nx=True, ex=10):  # Acquire lock
        try:
            value = db.get(key)
            cache.set(key, value, ex=ttl)
            return value
        finally:
            cache.delete(lock_key)
    else:
        # Another request has the lock; wait and retry
        time.sleep(0.1)
        return get_with_lock(key, ttl)
```

### Probabilistic Early Expiration

```python
def get_with_probabilistic_refresh(key, ttl=3600):
    value, timestamp = cache.get_with_metadata(key)
    if value is None:
        value = db.get(key)
        cache.set(key, value, ex=ttl)
        return value

    age = time.time() - timestamp
    # Probability increases as we approach TTL
    # β = 1: at ttl/2, ~13% chance to refresh
    if age > ttl * 0.5 and random.random() < 0.01:
        # Background refresh (don't block)
        threading.Thread(target=lambda: refresh(key, ttl)).start()

    return value
```

### Null Caching (Penetration Prevention)

```python
NULL_MARKER = "__NULL__"

def get_with_null_cache(key):
    value = cache.get(key)
    if value == NULL_MARKER:
        return None  # Known to not exist
    if value is not None:
        return value

    value = db.get(key)
    if value is None:
        cache.set(key, NULL_MARKER, ex=60)  # Short TTL for nulls
        return None
    cache.set(key, value, ex=3600)
    return value
```

### Bloom Filter (Penetration Prevention)

```python
# Using pybloom or redis bloom module
bloom = BloomFilter(capacity=10**7, error_rate=0.01)

def get_with_bloom(key):
    if not bloom.contains(key):
        return None  # Definitely not in DB
    value = cache.get(key)
    if value is not None:
        return value
    value = db.get(key)
    if value is not None:
        cache.set(key, value, ex=3600)
    return value
```

### Consistent Hashing (Simplified)

```python
import hashlib

class ConsistentHash:
    def __init__(self, nodes, vnodes=100):
        self.ring = {}
        self.sorted_keys = []
        for node in nodes:
            self.add_node(node, vnodes)

    def _hash(self, key):
        return int(hashlib.md5(key.encode()).hexdigest(), 16) % (2**32)

    def add_node(self, node, vnodes):
        for i in range(vnodes):
            key = self._hash(f"{node}:{i}")
            self.ring[key] = node
            self.sorted_keys.append(key)
        self.sorted_keys.sort()

    def get_node(self, key):
        h = self._hash(key)
        for k in self.sorted_keys:
            if h <= k:
                return self.ring[k]
        return self.ring[self.sorted_keys[0]]
```

### Cache Warming

```python
def warm_cache(keys):
    for key in keys:
        value = db.get(key)
        if value is not None:
            cache.set(key, value, ex=3600)

# At startup
hot_keys = db.query("SELECT id FROM products ORDER BY views DESC LIMIT 10000")
warm_cache([f"product:{k}" for k in hot_keys])
```

---

## 14. Interview Discussion

### Key Points to Articulate

1. **"Distributed cache scales beyond single-node memory."** Facebook has 1.5TB+ Memcached. Single node = 64-512GB. Distributed = sum of all nodes.

2. **"Consistent hashing minimizes rebalancing when nodes are added or removed."** Modulo causes (N-1)/N keys to move. Consistent hashing: ~K/N keys move (K = keys, N = nodes).

3. **"Cache stampede: many requests miss simultaneously and hit DB."** Solution: lock (single flight), probabilistic early expiration, or request coalescing.

4. **"Cache penetration: requests for non-existent keys."** Solution: Bloom filter (reject before DB) or null caching (cache empty result).

5. **"Redis is single-threaded; Memcached is multi-threaded."** Redis: 100K ops/sec. Memcached: 200K+ ops/sec. Redis has persistence and data structures.

6. **"Hot key: one key gets disproportionate traffic."** Solution: local cache, replication, or key splitting.

### Follow-Up Questions

- **"How does Redis Cluster shard?"** 16,384 slots. CRC16(key) mod 16384. Hash tags for co-location.
- **"What's the difference between cache stampede and avalanche?"** Stampede: one key expires, many concurrent requests. Avalanche: many keys expire at once (bad TTL design).
- **"How would you prevent cache penetration?"** Bloom filter or null caching. Bloom: O(1) check, false positives possible. Null: cache empty, short TTL.
- **"Redis vs. Memcached?"** Redis: persistence, data structures, single-threaded. Memcached: multi-threaded, higher throughput, no persistence, simpler.

### Red Flags to Avoid

- Not knowing cache stampede (very common interview topic)
- Confusing penetration (non-existent keys) with stampede (concurrent miss)
- Recommending Redis for simple key-value when Memcached might be better (throughput)
- Ignoring replication and failover in distributed design
