# System Design: Distributed Cache

## 1. Problem Statement & Requirements

### Functional Requirements

- **Get**: Retrieve value by key; return null if not found or expired
- **Put**: Store key-value pair with optional TTL (time-to-live)
- **Delete**: Remove key from cache
- **TTL Support**: Keys expire after specified duration (e.g., 60 seconds, 1 hour)
- **LRU Eviction**: When cache is full, evict least recently used items
- **Bulk Operations** (optional): Get multiple keys, put multiple keys

### Non-Functional Requirements

- **High Throughput**: 100K+ QPS per node
- **Low Latency**: p99 < 5ms for get/put
- **High Availability**: 99.9% uptime; tolerate node failures
- **Scalability**: Add/remove nodes without full rebalancing
- **Consistency**: Eventual consistency acceptable; strong consistency not required for cache

### Out of Scope

- Persistence as primary feature (cache is ephemeral)
- Complex query support (no range queries, no secondary indexes)
- Multi-region replication with strong consistency
- Authentication/authorization

---

## 2. Back-of-Envelope Estimation

### Traffic Estimates

| Metric | Value | Calculation |
|--------|-------|-------------|
| Total QPS | 500K | Given (distributed across cluster) |
| Get QPS | 400K | 80% reads |
| Put QPS | 80K | 15% writes |
| Delete QPS | 20K | 5% |
| Per node (10 nodes) | 50K QPS | 500K / 10 |

### Storage Estimates

- **Assumptions**:
  - Avg key size: 50 bytes
  - Avg value size: 1 KB
  - 100M keys in cache
  - Total: 100M × (50 + 1024) ≈ 100 GB
  - Per node (10 nodes): ~10 GB + overhead

### Memory Estimates

- **Per node**: 16-32 GB RAM
- **Overhead**: 20% for metadata (LRU pointers, TTL, hash table)
- **Usable**: ~12-25 GB per node for key-value data

### Bandwidth Estimates

- **Get**: 400K × 1 KB ≈ 400 MB/s (read)
- **Put**: 80K × 1 KB ≈ 80 MB/s (write)
- **Total**: ~500 MB/s cluster-wide

---

## 3. API Design

### REST API Endpoints

#### Get (Single)

```
GET /cache/{key}
```

**Response (200 OK):**
```
Content-Type: application/octet-stream
X-TTL-Remaining: 45
X-Cache-Hit: true

<binary value>
```

**Response (404 Not Found):**
```json
{
  "error": "KEY_NOT_FOUND",
  "message": "Key does not exist or has expired"
}
```

#### Put (Single)

```
PUT /cache/{key}
```

**Headers:**
```
Content-Type: application/octet-stream
X-TTL: 3600                    // optional, seconds
X-If-Not-Exists: true           // optional, only set if not exists (NX)
```

**Request Body:** Binary value

**Response (200 OK):**
```json
{
  "status": "ok",
  "key": "user:123",
  "ttl": 3600
}
```

#### Delete

```
DELETE /cache/{key}
```

**Response (200 OK):**
```json
{
  "status": "ok",
  "key": "user:123",
  "deleted": true
}
```

#### Get Multiple (Batch)

```
POST /cache/batch/get
```

**Request:**
```json
{
  "keys": ["user:1", "user:2", "user:3"]
}
```

**Response (200 OK):**
```json
{
  "results": {
    "user:1": {"value": "<base64>", "ttl_remaining": 100},
    "user:2": null,
    "user:3": {"value": "<base64>", "ttl_remaining": 50}
  }
}
```

---

## 4. Data Model / Database Schema

### In-Memory Data Structures

**No persistent database** — cache is in-memory. Optional: persistence layer (Redis RDB/AOF style).

### Key-Value Structure

```
Key:   string (max 256 bytes)
Value: bytes (max 1 MB typical, configurable)
TTL:   int64 (seconds, 0 = no expiry)
```

### Internal Structures (Per Node)

```python
# Hash map for O(1) lookup
cache: Dict[str, CacheEntry] = {}

# Doubly linked list for LRU ordering
# head = most recently used, tail = least recently used
class CacheEntry:
    key: str
    value: bytes
    ttl: int
    expires_at: int  # Unix timestamp
    prev: CacheEntry
    next: CacheEntry
```

### Persistence (Optional)

- **RDB**: Snapshot to disk periodically (e.g., every 5 min)
- **AOF**: Append-only log of writes; replay on restart
- **Hybrid**: RDB + incremental AOF

---

## 5. High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              CLIENT REQUEST                                       │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         CLIENT LIBRARY (Smart Client)                             │
│              Consistent hashing, connection pooling, local cache                   │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
                    ┌───────────────────┼───────────────────┐
                    ▼                   ▼                   ▼
┌───────────────────────┐   ┌───────────────────────┐   ┌───────────────────────┐
│   Cache Node 1        │   │   Cache Node 2         │   │   Cache Node N         │
│   - Hash ring slot 0  │   │   - Hash ring slot 1   │   │   - Hash ring slot N   │
│   - LRU eviction      │   │   - Replica of 0       │   │   - Replica of N-1     │
└───────────────────────┘   └───────────────────────┘   └───────────────────────┘
        │                               │                               │
        └───────────────────────────────┼───────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│   Data Store (Optional) — Primary source for cache miss / cache warming          │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 6. Detailed Component Design

### 6.1 Consistent Hashing

**Problem**: When using `hash(key) % N` for N nodes, adding/removing a node causes ~(N-1)/N keys to be remapped.

**Solution**: Consistent hashing
- Hash ring: 0 to 2^32-1 (or 2^64)
- Each node has multiple virtual nodes (e.g., 150) on the ring
- Key maps to first node clockwise from `hash(key)` on the ring
- Adding node: Only keys from next node move
- Removing node: Keys from removed node go to next node

**Virtual nodes**: Reduce load imbalance. Without them, one node might get disproportionate keys.

```python
def get_node(key, ring, num_virtual=150):
    hash_val = hash(key) % (2**32)
    # Find first node clockwise
    for node_hash, node_id in sorted(ring.items()):
        if node_hash >= hash_val:
            return node_id
    return ring[min(ring.keys())]  # wrap around
```

### 6.2 LRU Eviction Implementation

**Data structures**:
- **HashMap**: key → (value, ttl, pointer to list node)
- **Doubly Linked List**: MRU at head, LRU at tail

**Operations**:
- **Get**: Move node to head (if found)
- **Put**: Add to head; if over capacity, evict tail
- **Delete**: Remove from list and map

**Complexity**: O(1) for get, put, delete

```python
class LRUCache:
    def __init__(self, capacity):
        self.capacity = capacity
        self.cache = {}  # key -> Node
        self.head = Node(None, None)  # dummy head
        self.tail = Node(None, None)  # dummy tail
        self.head.next = self.tail
        self.tail.prev = self.head

    def get(self, key):
        if key not in self.cache:
            return None
        node = self.cache[key]
        self._move_to_head(node)
        return node.value

    def put(self, key, value):
        if key in self.cache:
            self.cache[key].value = value
            self._move_to_head(self.cache[key])
            return
        node = Node(key, value)
        self.cache[key] = node
        self._add_to_head(node)
        if len(self.cache) > self.capacity:
            lru = self.tail.prev
            self._remove(lru)
            del self.cache[lru.key]
```

### 6.3 TTL Handling

**Lazy expiration**: On get, check `expires_at`; if expired, delete and return null.

**Active expiration**: Background thread periodically scans and removes expired keys. Or use a **min-heap** (priority queue) keyed by `expires_at` for efficient removal.

**Hybrid**: Lazy + periodic sweep (e.g., every 1 second) to reclaim memory from expired keys.

### 6.4 Cache Replication

**Options**:
1. **No replication**: Simple; lose data on node failure
2. **Master-Replica**: Each key has primary + 1-2 replicas on different nodes
3. **Multi-master**: Complex; conflict resolution needed

**Recommended**: Each node is primary for its keys; replicate to next 1-2 nodes on the ring (successor nodes). On read: read from primary. On write: write to primary + replicas (async or sync).

### 6.5 Cache Warming

- **On startup**: Preload hot keys from persistent store
- **On miss**: Fetch from DB, populate cache, return
- **Background warming**: Proactively load keys likely to be accessed (e.g., trending items)

### 6.6 Hot Key Handling

**Problem**: One key gets massive traffic (e.g., celebrity profile); single node overloaded.

**Solutions**:
1. **Local cache**: Application-level cache (e.g., 1K entries LRU) in each client; reduces network calls
2. **Key splitting**: Store `key:1`, `key:2`, ... `key:N`; client picks random; aggregate on read
3. **Replication**: Replicate hot key to multiple nodes; client randomly picks
4. **Request coalescing**: Multiple requests for same key → single fetch; broadcast result

### 6.7 Cache Stampede Prevention

**Problem**: Key expires; 1000 requests simultaneously miss; all 1000 fetch from DB.

**Solutions**:
1. **Probabilistic early expiration**: When TTL < 10% remaining, one request extends TTL and refetches; others use stale
2. **Lock/mutex**: First miss acquires lock; others wait; single DB fetch
3. **Background refresh**: Refresh before expiry in background
4. **Stale-while-revalidate**: Return stale; async refresh

```python
def get_with_stampede_prevention(key):
    value = cache.get(key)
    if value:
        if value.ttl_remaining < value.original_ttl * 0.1:
            # Probabilistic: 10% chance to refresh
            if random.random() < 0.1:
                async_refresh(key)
        return value
    with lock(key):
        value = cache.get(key)  # double-check
        if value:
            return value
        value = db.fetch(key)
        cache.set(key, value)
        return value
```

### 6.8 Persistence Options

- **None**: Pure cache; empty on restart
- **RDB**: Snapshot every N minutes; fast restart; lose recent writes
- **AOF**: Log every write; slower restart; durable
- **Hybrid**: RDB + AOF since last RDB

---

## 7. Scaling

### Horizontal Scaling

- Add nodes to ring; consistent hashing minimizes key migration
- Remove nodes; keys redistribute to neighbors

### Sharding Strategy

- **Consistent hashing**: Natural sharding; no manual shard assignment
- **Replication factor**: 2-3; each key on 2-3 nodes for availability

### Caching Strategy (Multi-Tier)

1. **L1 (Local)**: In-process cache (e.g., Caffeine, Guava); ~1ms
2. **L2 (Distributed)**: Redis/Memcached cluster; ~5ms
3. **L3 (DB)**: Primary data store; ~50ms

### Handling Hot Keys

- Local cache absorbs hot key traffic
- Replicate hot keys across more nodes
- Request coalescing

---

## 8. Failure Handling

| Component | Failure | Mitigation |
|-----------|---------|------------|
| Cache node | Down | Replica serves; consistent hashing routes to replica |
| Network partition | Split | Each partition serves; possible stale reads |
| Cache stampede | Many misses | Lock, probabilistic refresh, stale-while-revalidate |
| Memory full | OOM | LRU eviction; reject new puts with 507 |

### Redundancy

- Replication: Each key on 2-3 nodes
- No single point of failure

### Recovery

- Node restart: Empty cache; repopulate from DB on miss
- With persistence: Load RDB/AOF on startup

---

## 9. Monitoring & Observability

### Key Metrics

| Metric | Target | Alert |
|--------|--------|-------|
| Latency (get) | p99 < 5ms | Critical |
| Latency (put) | p99 < 10ms | Warning |
| Hit rate | > 80% | Warning |
| Eviction rate | Monitor | High eviction = need more memory |
| Memory usage | < 90% | Critical |
| Node failures | 0 | Critical |

### Dashboards

- QPS per node
- Latency percentiles
- Hit rate
- Eviction rate
- Memory usage
- Connection count

### Alerting

- Latency spike
- Hit rate drop
- Node down
- Memory > 90%

---

## 10. Interview Tips

### Common Follow-Up Questions

1. **Why consistent hashing?** Minimize key movement on node add/remove
2. **How to implement LRU?** HashMap + doubly linked list; O(1)
3. **How to prevent cache stampede?** Lock, probabilistic refresh, stale-while-revalidate
4. **How to handle hot keys?** Local cache, replication, request coalescing
5. **Redis vs Memcached?** Redis: persistence, data structures, single-threaded. Memcached: multi-threaded, simpler.

### What Interviewers Look For

- Consistent hashing understanding
- LRU implementation
- Cache stampede awareness
- Hot key handling
- Trade-offs (consistency vs availability)

### Common Mistakes

- Using mod-based sharding (bad for scaling)
- Ignoring cache stampede
- Not considering hot keys
- Overcomplicating persistence

---

## Appendix: Additional Design Considerations

### A. Consistent Hashing Visualization

```
         Ring (0 to 2^32)
             0
            / \
    Node C /   \ Node A
          /     \
         /       \
        /  Node B \
       /           \
      -------------- 
      
Key K hashes to point P → goes to Node A (first clockwise)
```

### B. Virtual Nodes Example

Without virtual nodes: 3 nodes might get 50%, 30%, 20% of keys (imbalance).
With 150 virtual nodes per physical node: Each physical node has 150 points on ring; more uniform distribution.

### C. Read-Through vs Write-Through

- **Read-through**: On miss, cache fetches from DB and caches
- **Write-through**: On put, cache writes to DB and cache
- **Write-behind**: On put, cache writes to cache; async flush to DB (risk of data loss)

### D. Cache Invalidation

- **TTL**: Simple; eventual consistency
- **Explicit delete**: On DB update, delete cache key
- **Event-driven**: Publish invalidation event; all nodes delete
- **Version in key**: `user:123:v2`; old version expires naturally

### E. Redis vs Memcached Comparison

| Feature | Redis | Memcached |
|---------|-------|-----------|
| Data structures | String, List, Set, Hash, Sorted Set | String only |
| Persistence | RDB, AOF | None |
| Threading | Single-threaded | Multi-threaded |
| Replication | Master-replica | None |
| Use case | Rich features, persistence | Simple, high throughput |

### F. Memory Management in LRU

- **Max memory policy**: `maxmemory 10gb`
- **Eviction policy**: `maxmemory-policy allkeys-lru` (evict any key) or `volatile-lru` (only keys with TTL)
- **Fragmentation**: Use `jemalloc`; monitor `mem_fragmentation_ratio`

### G. Client-Side Caching (Redis 6+)

Redis supports client-side caching: Server notifies client when key is modified; client invalidates local copy. Reduces network round-trips for hot keys.

### H. Cluster Mode vs Sentinel

- **Sentinel**: Master-replica; automatic failover; single shard
- **Cluster**: Sharding; 16384 slots; each node handles subset; horizontal scale

### I. Consistent Hashing Code Example

```python
import hashlib

class ConsistentHash:
    def __init__(self, nodes, virtual_nodes=150):
        self.ring = {}
        for node in nodes:
            for i in range(virtual_nodes):
                key = f"{node}:{i}"
                h = int(hashlib.md5(key.encode()).hexdigest(), 16) % (2**32)
                self.ring[h] = node
        self.sorted_keys = sorted(self.ring.keys())

    def get_node(self, key):
        h = int(hashlib.md5(key.encode()).hexdigest(), 16) % (2**32)
        for k in self.sorted_keys:
            if k >= h:
                return self.ring[k]
        return self.ring[self.sorted_keys[0]]
```

### J. Cache Stampede: Lock Implementation

```python
def get_with_lock(key):
    value = cache.get(key)
    if value:
        return value
    lock_key = f"lock:{key}"
    if redis.set(lock_key, "1", nx=True, ex=10):
        try:
            value = db.get(key)
            cache.set(key, value, ttl=3600)
            return value
        finally:
            redis.delete(lock_key)
    else:
        time.sleep(0.1)
        return get_with_lock(key)  # Retry
```

### K. TTL and Eviction Interaction

- **TTL**: Key expires after N seconds
- **LRU**: Evict when memory full
- **Order**: TTL expiration takes precedence; then LRU for eviction
- **Redis**: `volatile-lru` = evict among keys with TTL only; `allkeys-lru` = evict any

### L. Network Topology for Low Latency

- **Same-region**: Cache nodes in same AZ as app servers; <1ms latency
- **Cross-region**: Replicate for DR; higher latency for read
- **Client-side**: Smart client with connection pooling; pipelining for batch

### M. Complete Interview Walkthrough (45 min)

**0-5 min**: Clarify: get/put/delete, TTL, LRU, scale, latency requirements.
**5-10 min**: Estimates. 500K QPS, 100M keys, 100 GB. Per-node capacity.
**10-15 min**: API. Simple get/put/delete. Optional batch operations.
**15-25 min**: Consistent hashing. Why? Virtual nodes. LRU implementation (hash + DLL).
**25-35 min**: Architecture. Client, hash ring, cache nodes. Replication strategy.
**35-40 min**: Cache stampede. Hot keys. Persistence. Failure handling.
**40-45 min**: Trade-offs. Redis vs Memcached. Consistency vs availability.

### N. Quick Reference Cheat Sheet

| Topic | Key Points |
|-------|------------|
| Partitioning | Consistent hashing; virtual nodes for balance |
| LRU | HashMap + doubly linked list; O(1) |
| TTL | Lazy expiration on get; periodic sweep |
| Stampede | Lock, probabilistic refresh, stale-while-revalidate |
| Hot keys | Local cache, replication, request coalescing |
| Replication | Each key on 2-3 nodes; read from any |

### O. Further Reading & Real-World Examples

- **Redis**: In-memory; persistence; data structures; cluster mode
- **Memcached**: Simple; multi-threaded; no persistence
- **Hazelcast**: Java; distributed; IMDG
- **Aerospike**: SSD-optimized; strong consistency options

### P. Design Alternatives Considered

| Decision | Alternative | Why Rejected |
|----------|-------------|--------------|
| Consistent hash | Mod-based | Mod causes full remap on node change |
| LRU | LFU | LRU simpler; LFU for access patterns |
| Redis | Memcached | Redis has persistence, structures |
| No replication | Replication | Single node = SPOF; fail = data loss |

### Q. LRU Eviction Complexity Analysis

- **Get**: O(1) - hash lookup + move to head
- **Put**: O(1) - hash insert + add to head + optional evict tail
- **Delete**: O(1) - hash delete + remove from list
- **Space**: O(n) for n keys; hash + list pointers

### R. Cache Sizing Guidelines

- **Rule of thumb**: Cache 20% of working set for 80% hit rate (Pareto)
- **Memory**: 1-2 GB per 100K keys (1KB values)
- **Connection pool**: 1 connection per 10K QPS to Redis

### S. When to Use Cache-Aside vs Read-Through

- **Cache-aside**: App manages cache; on miss, app fetches from DB and populates cache. Flexible.
- **Read-through**: Cache layer fetches from DB on miss. Simpler app logic; cache is responsible.

### T. Eviction Policy Comparison

| Policy | Description | Use Case |
|--------|-------------|----------|
| LRU | Least recently used | General purpose |
| LFU | Least frequently used | Access pattern matters |
| FIFO | First in first out | Simple; rarely optimal |
| TTL | Time to live | Expiring data |
| Random | Random eviction | When LRU cost too high |

### U. Summary

A distributed cache requires: consistent hashing for partitioning, LRU for eviction, TTL for expiration, replication for availability, and stampede prevention for reliability. Redis and Memcached are production implementations.

---
*End of Distributed Cache System Design Document*

This document covers the design of a distributed cache system suitable for system design interviews. Key takeaways: consistent hashing, LRU, TTL, stampede prevention, and hot key handling. Practice drawing the architecture diagram and explaining each component's role. Be ready to implement LRU on a whiteboard.

**Document Version**: 1.0 | **Last Updated**: 2025-03-10 | **Target**: System Design Interview (Easy)

**Key Interview Questions**: Consistent hashing benefits? LRU implementation? Cache stampede? Hot key handling? Redis vs Memcached?


