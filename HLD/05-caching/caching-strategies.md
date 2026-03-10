# Caching Strategies: Staff+ Engineer Deep Dive

## 1. Concept Overview

### Definition
Caching strategies define the **read and write patterns** between an application, cache layer, and persistent data store. They answer: "When does data flow to/from cache? Who is responsible for cache population? What are the consistency guarantees?"

### Purpose
- **Reduce latency**: Serve data from memory (μs) instead of disk/network (ms)
- **Reduce load**: Offload reads from primary database
- **Improve throughput**: Handle more requests with same infrastructure
- **Cost optimization**: Reduce database compute and I/O costs

### Why It Exists
Databases are optimized for durability and consistency, not speed. A single disk seek takes ~10ms; memory access takes ~100ns. That's a **100,000x** difference. Caching bridges this gap by keeping hot data in fast storage.

### Problems It Solves
1. **Read amplification**: Repeated reads of same data overwhelm DB
2. **Write amplification**: Every write doesn't need immediate durability
3. **Consistency vs. performance**: Different strategies offer different tradeoffs
4. **Cache invalidation**: "One of the two hard problems in computer science"

---

## 2. Real-World Motivation

### Google
- **Bigtable/Megastore**: Read-through caches for tablet metadata
- **Spanner**: Multi-level caching (local → distributed → global)
- **YouTube**: Cache-aside for video metadata, write-behind for analytics

### Netflix
- **EVCache**: Memcached-based, cache-aside for user preferences, recommendations
- **Zuul**: Edge caching with write-through for session data
- **Cassandra**: Write-behind for viewing history (eventual consistency acceptable)

### Uber
- **Geolocation**: Cache-aside for driver/rider locations (high read, moderate write)
- **Pricing**: Read-through for fare estimates (must be fresh)
- **Trip history**: Write-behind for completed trips (async acceptable)

### Amazon
- **Product catalog**: Read-through (DynamoDB DAX)
- **Shopping cart**: Write-through (consistency critical)
- **Recommendations**: Cache-aside with TTL (stale acceptable)

### Twitter
- **Timeline**: Cache-aside with fan-out on write
- **User profiles**: Read-through
- **Tweet counts**: Write-behind (eventual consistency)

---

## 3. Architecture Diagrams

### Cache-Aside (Lazy Loading) Flow

```
                    READ FLOW
    ┌─────────┐     ┌─────────┐     ┌─────────┐
    │   App   │────▶│  Cache  │     │   DB    │
    └────┬────┘     └────┬────┘     └────▲────┘
         │               │               │
         │  1. get(key)   │               │
         │──────────────▶│               │
         │               │               │
         │  2. MISS      │               │
         │◀──────────────│               │
         │               │               │
         │  3. get(key)  │               │
         │──────────────────────────────▶│
         │               │               │
         │  4. data      │               │
         │◀──────────────────────────────│
         │               │               │
         │  5. set(key, data)            │
         │──────────────▶│               │
         │               │               │
         │  6. return data               │
         │◀──────────────│               │
         └───────────────┴───────────────┘

                    WRITE FLOW
    ┌─────────┐     ┌─────────┐     ┌─────────┐
    │   App   │────▶│  Cache  │     │   DB    │
    └────┬────┘     └────┬────┘     └────▲────┘
         │               │               │
         │  1. delete(key) or set(key)   │
         │──────────────▶│               │
         │               │               │
         │  2. write to DB               │
         │──────────────────────────────▶│
         │               │               │
         │  3. invalidate/update cache   │
         │──────────────▶│               │
         └───────────────┴───────────────┘
```

### Read-Through Flow

```
    ┌─────────┐     ┌─────────────────────────┐     ┌─────────┐
    │   App   │     │  Cache (with loader)    │     │   DB    │
    └────┬────┘     └───────────┬─────────────┘     └────▲────┘
         │                      │                        │
         │  1. get(key)          │                        │
         │─────────────────────▶│                        │
         │                      │  2. MISS: load(key)    │
         │                      │───────────────────────▶│
         │                      │                        │
         │                      │  3. data               │
         │                      │◀───────────────────────│
         │                      │                        │
         │                      │  4. store in cache     │
         │                      │  (internal)           │
         │                      │                        │
         │  5. return data      │                        │
         │◀─────────────────────│                        │
         └──────────────────────┴────────────────────────┘
```

### Write-Through Flow

```
    ┌─────────┐     ┌─────────┐     ┌─────────┐
    │   App   │     │  Cache  │     │   DB    │
    └────┬────┘     └────┬────┘     └────▲────┘
         │               │               │
         │  1. set(key, value)           │
         │──────────────▶│               │
         │               │  2. write to DB (sync)        │
         │               │──────────────▶│
         │               │               │
         │               │  3. ack       │
         │               │◀──────────────│
         │               │               │
         │               │  4. update cache              │
         │               │  (internal)   │
         │               │               │
         │  5. ack       │               │
         │◀──────────────│               │
         └───────────────┴───────────────┘
```

### Write-Behind (Write-Back) Flow

```
    ┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐
    │   App   │     │  Cache  │     │  Queue  │     │   DB    │
    └────┬────┘     └────┬────┘     └────┬────┘     └────▲────┘
         │               │               │               │
         │  1. set(key)  │               │               │
         │──────────────▶│               │               │
         │               │  2. update cache (immediate)  │
         │               │  (in-memory)   │               │
         │  3. ack       │               │               │
         │◀──────────────│               │               │
         │               │               │               │
         │               │  4. async: enqueue write      │
         │               │──────────────▶│               │
         │               │               │  5. batch write
         │               │               │──────────────▶│
         └───────────────┴───────────────┴───────────────┘
```

### Write-Around Flow

```
    ┌─────────┐     ┌─────────┐     ┌─────────┐
    │   App   │     │  Cache  │     │   DB    │
    └────┬────┘     └────┬────┘     └────▲────┘
         │               │               │
         │  WRITE: bypass cache          │
         │  1. write directly to DB      │
         │──────────────────────────────▶│
         │               │               │
         │  READ: populate on miss       │
         │  2. get(key)  │               │
         │──────────────▶│               │
         │  3. MISS      │  4. read from DB              │
         │               │───────────────▶│
         │               │  5. populate cache             │
         │               │  (optional)   │
         └───────────────┴───────────────┘
```

---

## 4. Core Mechanics

### Cache-Aside (Lazy Loading)

**Internal Workings:**
1. Application owns cache logic—cache is "dumb" storage
2. On read miss: app fetches from DB, then populates cache
3. On write: app writes to DB first, then invalidates (or updates) cache
4. **Critical**: Use "cache stampede" prevention (locking, probabilistic expiration)

**Consistency Model:** Best-effort. Cache can be stale until next invalidation. Race conditions possible: two requests miss, both fetch, both write—last write wins.

**Cache Population:** Only on read miss. Cold cache = all reads hit DB until warmed.

### Read-Through

**Internal Workings:**
1. Cache layer implements `CacheLoader` interface—knows how to fetch from source
2. Application only calls `cache.get(key)`—no DB knowledge
3. Cache transparently fetches on miss, stores, returns
4. Often combined with write-through in same cache (e.g., Guava, Caffeine)

**Consistency Model:** Same as cache-aside for reads. Writes typically handled separately (write-through or write-behind).

**Cache Population:** Transparent. Application code is simpler—no "if miss, fetch, put" logic.

### Write-Through

**Internal Workings:**
1. Every write goes to cache AND DB synchronously
2. Write completes only when both succeed
3. Cache and DB always consistent (for that write)
4. Read always returns cached value (which matches DB)

**Consistency Model:** Strong. Cache is always coherent with DB.

**Tradeoff:** Write latency = max(cache_write, db_write). DB is bottleneck.

### Write-Behind (Write-Back)

**Internal Workings:**
1. Write goes to cache immediately (returns fast)
2. Cache queues write for async flush to DB
3. Flush can be: time-based (every 5s), size-based (every 1000 writes), or both
4. Risk: cache failure before flush = data loss

**Consistency Model:** Eventual. Reads from cache are consistent; DB catches up asynchronously.

**Optimization:** Batch writes, coalesce updates to same key.

### Write-Around

**Internal Workings:**
1. Writes bypass cache entirely—go straight to DB
2. Cache only populated on read miss (lazy) or never for write-heavy data
3. Use case: write-heavy, read-rarely data (logs, analytics events)
4. Avoids polluting cache with data that won't be read

**Consistency Model:** Reads may be stale (cache not updated on write). Good when writes >> reads for specific keys.

---

## 5. Numbers

| Metric | Cache-Aside | Read-Through | Write-Through | Write-Behind | Write-Around |
|--------|-------------|--------------|---------------|--------------|--------------|
| Read latency (hit) | ~100μs | ~100μs | ~100μs | ~100μs | ~100μs |
| Read latency (miss) | DB + cache put | DB + cache put | N/A (always hit) | N/A | DB + optional put |
| Write latency | DB + invalidate | DB + cache | DB (sync) | Cache only (~1ms) | DB only |
| DB load (reads) | Reduced | Reduced | Minimal | Minimal | Reduced |
| DB load (writes) | Same | Same | Same (every write) | Batched (90%+ reduction) | Same |
| Cache memory | On-demand | On-demand | Full working set | Full + queue | Sparse |
| Data loss risk | Low | Low | None | **High** (crash before flush) | Low |

**Real-World Scale:**
- **Facebook TAO**: 1B+ objects, read-through, 99.99% cache hit rate
- **Netflix EVCache**: 100ms p99 for cache-aside, 30% DB load reduction
- **Amazon ElastiCache**: Sub-millisecond latency, 100K+ ops/sec per node

---

## 6. Tradeoffs

### Consistency vs. Performance

| Strategy | Consistency | Performance | When to Choose |
|----------|--------------|-------------|----------------|
| Write-Through | Strong | Lower (sync writes) | Banking, inventory |
| Write-Behind | Eventual | Highest | Analytics, logs |
| Cache-Aside | Best-effort | High | General purpose |
| Read-Through | Best-effort | High | Simpler app code |
| Write-Around | Stale reads | Medium | Write-heavy, read-rare |

### Complexity vs. Control

| Strategy | App Complexity | Cache Complexity | Control |
|----------|----------------|------------------|---------|
| Cache-Aside | High (app owns logic) | Low | Full |
| Read-Through | Low | High (loader logic) | Medium |
| Write-Through | Low | Medium | Low |
| Write-Behind | Low | High (queue, flush) | Low |
| Write-Around | Medium | Low | Medium |

### Failure Modes

| Strategy | Cache Down | DB Down | Partial Failure |
|----------|-------------|---------|-----------------|
| Cache-Aside | Degrades to DB | App fails | Stale cache possible |
| Read-Through | Degrades to DB | App fails | Same |
| Write-Through | Can't write | Can't write | N/A |
| Write-Behind | Can't write, queue at risk | Writes queued | **Data loss** if cache dies |
| Write-Around | Reads hit DB | App fails | Cache never updated |

---

## 7. Variants / Implementations

### Cache-Aside Variants

**1. Double-checked locking:**
```
lock(key)
  if cache.get(key) == null
    value = db.get(key)
    cache.set(key, value)
  return cache.get(key)
unlock(key)
```

**2. Probabilistic early expiration (stampede prevention):**
```
if cache.get(key) == null or (random() < 0.01 and age > TTL/2)
  value = db.get(key)
  cache.set(key, value, TTL)
return cache.get(key)
```

**3. Refresh-ahead:** Background refresh before expiry (Netflix pattern)

### Read-Through Implementations

- **Guava Cache**: `CacheLoader.load(key)` 
- **Caffeine**: `LoadingCache` with `CacheLoader`
- **Redis with RedisGears**: Lua scripts for read-through
- **DynamoDB DAX**: Transparent read-through for DynamoDB

### Write-Behind Implementations

- **MySQL with Debezium**: CDC to Kafka, async apply
- **CQRS/Event Sourcing**: Write to event store, async project to read model
- **Kafka Connect**: Sink connectors with batching
- **Custom**: Redis + background worker + DB batch inserts

### Hybrid Strategies

**Facebook TAO:** Read-through for reads, write-through for writes (strong consistency where needed)

**Twitter:** Cache-aside + write-behind for counts (eventual consistency OK)

**Netflix:** Cache-aside for metadata, write-around for playback events (write-heavy)

---

## 8. Scaling Strategies

### Horizontal Scaling
- **Cache-aside/Read-through**: Add cache nodes, consistent hashing for distribution
- **Write-through**: Cache scales with reads; DB is write bottleneck—consider sharding
- **Write-behind**: Queue (Kafka, SQS) absorbs write bursts; workers scale independently

### Cache Warming
- **Eager loading**: Pre-populate cache at startup (top N items)
- **Predictive**: ML to predict hot keys, pre-fetch
- **Stale-while-revalidate**: Serve stale, refresh in background

### Multi-Level Caching
```
Request → L1 (local, 1ms) → L2 (distributed, 5ms) → DB (50ms)
```
- L1: Process-local (Caffeine, Guava)
- L2: Redis/Memcached cluster
- Different strategies per level (e.g., L1 write-through to L2, L2 cache-aside to DB)

---

## 9. Failure Scenarios

### Cache Stampede (Thundering Herd)
**Scenario:** Cache expires for hot key. 10,000 requests miss simultaneously. All 10,000 hit DB.
**Solution:** Lock (single flight), probabilistic early expiration, or "request coalescing"

### Write-Behind Data Loss
**Scenario:** Cache node crashes with 1000 queued writes. Queue in memory—lost.
**Solution:** Durable queue (Kafka, Redis AOF), replicate cache, or accept loss for non-critical data

### Stale Reads
**Scenario:** Write-through with replication lag. Read from replica that hasn't received write.
**Solution:** Read from primary for critical reads, or use session consistency

### Cache Invalidation Storm
**Scenario:** Invalidate 1M keys. Next 1M reads all miss. DB overload.
**Solution:** Batch invalidation with delay, or version-based invalidation (invalidate on next read)

---

## 10. Performance Considerations

### Latency Budget
- Cache hit: 0.1-1ms (local), 1-5ms (distributed)
- Cache miss + DB: 10-100ms
- Write-through: 2x DB latency
- Write-behind: Cache latency only (~1ms)

### Memory vs. Hit Rate
- Larger cache = higher hit rate (diminishing returns)
- 80/20 rule: 20% of keys get 80% of reads
- Typical target: 90-99% hit rate for read-heavy workloads

### Write Amplification
- Write-through: 1 write = 1 DB write + 1 cache write
- Write-behind: 1 write = 1 cache write, N writes batched = 1 DB write
- Write-around: 1 write = 1 DB write, cache unaffected

---

## 11. Use Cases

| Use Case | Strategy | Rationale |
|----------|----------|-----------|
| User session | Write-through | Consistency critical |
| Product catalog | Read-through | Read-heavy, stale OK |
| Shopping cart | Write-through | Must not lose items |
| Video metadata | Cache-aside | High read, moderate write |
| Analytics events | Write-behind | Throughput critical |
| Search results | Cache-aside + TTL | Stale acceptable |
| Leaderboard | Write-through or -behind | Depends on consistency need |
| Rate limiting | Write-through | Must be accurate |
| Recommendations | Cache-aside | Stale OK, refresh periodically |

---

## 12. Comparison Table

| Aspect | Cache-Aside | Read-Through | Write-Through | Write-Behind | Write-Around |
|--------|------------|--------------|---------------|--------------|--------------|
| **App logic** | Complex | Simple | Simple | Simple | Medium |
| **Cache logic** | Simple | Complex | Medium | Complex | Simple |
| **Read miss** | App fetches | Cache fetches | Rare | Rare | App fetches |
| **Write path** | DB → invalidate | DB → cache | Cache → DB (sync) | Cache → queue | DB only |
| **Consistency** | Best-effort | Best-effort | Strong | Eventual | Stale |
| **Data loss risk** | Low | Low | None | **High** | Low |
| **DB write load** | Full | Full | Full | Batched | Full |
| **Implementations** | Most common | Guava, Caffeine, DAX | Many | Kafka, CQRS | Custom |
| **Best for** | General | Simplicity | Consistency | Throughput | Write-heavy |

---

## 13. Code / Pseudocode

### Cache-Aside (Full Implementation)

```python
def get_user(user_id: int) -> User:
    cache_key = f"user:{user_id}"
    
    # 1. Check cache first
    cached = cache.get(cache_key)
    if cached is not None:
        return deserialize(cached)
    
    # 2. Cache miss - fetch from DB
    # Optional: lock to prevent stampede
    with distributed_lock(f"lock:{cache_key}", ttl=5):
        # Double-check after acquiring lock
        cached = cache.get(cache_key)
        if cached is not None:
            return deserialize(cached)
        
        user = db.query("SELECT * FROM users WHERE id = ?", user_id)
        if user is None:
            # Cache null to prevent penetration
            cache.set(cache_key, NULL_MARKER, ttl=60)
            return None
        
        cache.set(cache_key, serialize(user), ttl=3600)
        return user

def update_user(user_id: int, data: dict):
    db.execute("UPDATE users SET ... WHERE id = ?", user_id, data)
    cache.delete(f"user:{user_id}")  # Invalidate, not update
```

### Read-Through (Pseudocode)

```python
class UserCacheLoader(CacheLoader):
    def load(self, key: str) -> User:
        user_id = extract_id(key)
        return db.query("SELECT * FROM users WHERE id = ?", user_id)

cache = LoadingCache(
    loader=UserCacheLoader(),
    max_size=10000,
    ttl=3600
)

# Application code - no DB logic
user = cache.get("user:123")  # Cache fetches on miss
```

### Write-Through (Pseudocode)

```python
def set_user(user_id: int, user: User):
    # Both must succeed - use transaction or sequential
    db.execute("INSERT OR UPDATE users ...", user)
    cache.set(f"user:{user_id}", serialize(user), ttl=3600)
    # Return only after both complete
```

### Write-Behind (Pseudocode)

```python
write_queue = Queue()

def set_analytics_event(event: Event):
    cache.set(f"event:{event.id}", event)  # Immediate
    write_queue.put(event)  # Async flush
    return  # Fast return

# Background worker
def flush_worker():
    batch = []
    while True:
        batch.extend(write_queue.drain(timeout=5s, max=1000))
        if batch:
            db.batch_insert("events", batch)
            batch.clear()
```

---

## 14. Interview Discussion

### Key Points to Articulate

1. **"Cache-aside is the most common because the application has full control."** Explain the tradeoff: more code, but flexibility for invalidation, stampede prevention, and partial failures.

2. **"Write-through gives strong consistency but doubles write latency."** When would you use it? Banking, inventory—where stale data is unacceptable.

3. **"Write-behind is the highest performance but has data loss risk."** Mitigate with durable queues. Use for analytics, logs, metrics.

4. **"Read-through simplifies application code—the cache is an abstraction over the data source."** Tradeoff: cache layer needs to implement loader logic; less control over fetch behavior.

5. **"Write-around avoids polluting cache with write-heavy, read-rare data."** Example: Clickstream events—millions of writes, few reads (until batch analytics).

### Follow-Up Questions

- **"How would you prevent cache stampede?"** Locking, probabilistic early expiration, request coalescing.
- **"When would you choose write-behind over write-through?"** When write throughput is the bottleneck and eventual consistency is acceptable (e.g., view counts).
- **"How does Facebook TAO work?"** Read-through cache, objects stored by (type, id), denormalized for common access patterns, 99.99% hit rate.
- **"What's the consistency model of cache-aside?"** Best-effort. Cache can be stale. Use TTL or explicit invalidation. Race conditions on concurrent miss + write possible.

### Red Flags to Avoid

- Saying "cache-aside and read-through are the same" (they're not—ownership of fetch logic differs)
- Recommending write-behind for financial data
- Ignoring cache stampede in high-traffic scenarios
- Not considering cache failure mode (degradation vs. data loss)
