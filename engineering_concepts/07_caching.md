# Module 7: Caching

---

## 1. Cache-Aside (Lazy Loading)

### Definition
Application checks cache first. On miss, it fetches from DB, then populates the cache. The application manages the cache explicitly.

### Flow
```
READ:
  App → Cache: GET user:123
  Cache: MISS
  App → DB: SELECT * FROM users WHERE id=123
  App → Cache: SET user:123 = {data}  (with TTL)
  App → Client: return {data}

  Next read: Cache HIT → skip DB

WRITE:
  App → DB: UPDATE users SET name='Alice' WHERE id=123
  App → Cache: DELETE user:123  (invalidate)
```

### Visual
```
  ┌────────┐   1. GET    ┌───────┐
  │ Client │ ──────────→ │ Cache │ → HIT? return
  └───┬────┘             └───┬───┘
      │                      │ MISS
      │  4. Return           │ 2. Query DB
      │                      ▼
      │                 ┌──────┐
      │                 │  DB  │
      │                 └──┬───┘
      │                    │ 3. Populate cache
      └────────────────────┘
```

### Tradeoffs

| Pros | Cons |
|------|------|
| Only caches what's actually read | First request always slow (cold miss) |
| Cache failure = fallback to DB | Stale data until TTL expires or invalidation |
| Simple to implement | Application manages cache logic |

### Real Systems
Most web applications, Memcached pattern, Redis pattern

---

## 2. Write-Through Cache

### Definition
Every write goes to BOTH cache and database synchronously. Cache is always up-to-date.

### Flow
```
WRITE:
  App → Cache: SET user:123 = {data}
  Cache → DB: Write to database
  Both succeed → ACK to App

READ:
  App → Cache: GET user:123 → always HIT (if written before)
```

### Visual
```
  ┌────────┐  Write  ┌───────┐  Sync Write  ┌──────┐
  │ Client │ ──────→ │ Cache │ ───────────→ │  DB  │
  └────────┘         └───────┘              └──────┘
                     (both updated simultaneously)
```

### Tradeoffs

| Pros | Cons |
|------|------|
| Cache always consistent with DB | Higher write latency (2 writes) |
| No stale reads | Caches data that may never be read |
| Simple consistency model | Cache churn for write-heavy data |

---

## 3. Write-Back (Write-Behind) Cache

### Definition
Writes go to cache only. Cache asynchronously flushes to DB in the background. Fastest writes, but risk of data loss.

### Flow
```
WRITE:
  App → Cache: SET user:123 = {data}  → ACK immediately
  Cache → DB: async batch write (later)

READ:
  App → Cache: GET user:123 → HIT (latest data in cache)
```

### Visual
```
  ┌────────┐  Write  ┌───────┐  Async Flush  ┌──────┐
  │ Client │ ──────→ │ Cache │ ─ ─ ─ ─ ─ ─→ │  DB  │
  └────────┘  (fast) └───────┘  (background)  └──────┘
```

### Tradeoffs

| Pros | Cons |
|------|------|
| Fastest writes (cache only) | DATA LOSS if cache crashes before flush |
| Batch writes reduce DB load | Complex to implement correctly |
| Absorbs write spikes | Inconsistency window |

### Real Systems
CPU L1/L2 caches, Linux page cache, some game servers

---

## 4. Comparison: Cache Strategies

```
┌──────────────┬──────────────┬───────────────┬──────────────┐
│              │ Cache-Aside  │ Write-Through │ Write-Back   │
├──────────────┼──────────────┼───────────────┼──────────────┤
│ Read miss    │ Load from DB │ N/A (pre-pop) │ N/A          │
│ Write path   │ DB then inv. │ Cache+DB sync │ Cache only   │
│ Consistency  │ Eventual     │ Strong        │ Weak         │
│ Write speed  │ DB speed     │ Slower        │ Fastest      │
│ Data loss    │ No           │ No            │ Possible     │
│ Complexity   │ Low          │ Medium        │ High         │
│ Best for     │ Read-heavy   │ Read-heavy    │ Write-heavy  │
└──────────────┴──────────────┴───────────────┴──────────────┘
```

---

## 5. Cache Invalidation

### Definition
The process of removing or updating stale cached data. "There are only two hard things in CS: cache invalidation and naming things."

### Strategies
```
1. TTL (Time-To-Live)
   SET key value EX 300    ← expires in 300 seconds
   Simple but stale for TTL duration

2. Event-Based Invalidation
   DB write → publish event → consumer deletes cache key
   Consistent but complex

3. Write-Through
   Writes update cache simultaneously
   Always fresh but higher write cost

4. Versioning
   Cache key includes version: "user:123:v7"
   New version = new key, old naturally expires
```

### Common Bugs
```
Race Condition:
  Thread 1: Read DB (old value)
  Thread 2: Write DB (new value)
  Thread 2: Delete cache
  Thread 1: Set cache (OLD value!)  ← stale cache!

Solution: Use "cache-aside with DELETE" not "SET"
         Or use versioned keys
         Or use distributed locks
```

### Summary
Cache invalidation is the hardest caching problem. TTL is simplest, event-based is most consistent. Watch for race conditions between DB writes and cache updates.

---

## 6. LRU Cache (Least Recently Used)

### Definition
Evicts the item that hasn't been accessed for the longest time. Most popular cache eviction policy.

### Data Structure
```
HashMap + Doubly Linked List = O(1) get and put

HashMap:  key → pointer to linked list node
LinkedList: Most Recent ←→ ... ←→ Least Recent

GET(key):
  1. Find node via HashMap → O(1)
  2. Move node to HEAD of list → O(1)

PUT(key, value):
  1. If exists: update + move to HEAD
  2. If new + full: evict TAIL (LRU item), insert at HEAD
  3. If new + space: insert at HEAD
```

### Visual
```
Access order: A, B, C, D, A, E (capacity = 4)

After A,B,C,D:  HEAD→[D]↔[C]↔[B]↔[A]←TAIL
After access A:  HEAD→[A]↔[D]↔[C]↔[B]←TAIL  (A moves to head)
After insert E:  HEAD→[E]↔[A]↔[D]↔[C]←TAIL  (B evicted from tail)
```

### Complexity
- GET: **O(1)**
- PUT: **O(1)**
- Space: **O(capacity)**

### Real Systems
Redis (allkeys-lru), Memcached, CPU caches, every browser cache

---

## 7. LFU Cache (Least Frequently Used)

### Definition
Evicts the item with the lowest access count. Keeps frequently accessed items even if not recently used.

### Data Structure
```
HashMap: key → (value, frequency)
Frequency Map: frequency → list of keys with that frequency
Min Frequency: tracks current minimum

GET(key):
  1. Get value, increment frequency
  2. Move key from freq_map[old_freq] to freq_map[old_freq + 1]
  3. Update min_freq if old list is empty

PUT(key, value):
  1. If full: evict from freq_map[min_freq] (LRU within same freq)
  2. Insert with frequency = 1, min_freq = 1
```

### LRU vs LFU

| | LRU | LFU |
|-|-----|-----|
| Evicts | Least recently used | Least frequently used |
| Burst resilience | Poor (scan evicts popular items) | Good (popular items protected) |
| New items | Equal chance | Vulnerable (low freq) |
| Implementation | Simple (HashMap + DLL) | Complex (freq buckets) |
| Cache pollution | Susceptible | Resistant |

### Real Systems
Redis (allkeys-lfu), some CDN caches

---

## 8. Clock Algorithm (CLOCK)

### Definition
An approximation of LRU that uses a circular buffer and a reference bit. Much cheaper than true LRU.

### How It Works
```
Circular buffer with "clock hand":
  Each page has a reference bit (0 or 1)

  Access: Set reference bit = 1

  Eviction (clock hand sweeps):
    If ref bit = 1 → set to 0, move to next (second chance)
    If ref bit = 0 → EVICT this page

┌───────────────────────────┐
│    ┌─[A:1]─[B:0]─┐       │
│    │              │       │
│  [E:1]    hand→[C:1]     │  hand at C: ref=1 → set 0, advance
│    │              │       │  hand at D: ref=0 → EVICT D
│    └─[D:0]───────┘       │
└───────────────────────────┘
```

### Why Not True LRU?
True LRU requires moving items on every access (expensive for OS page replacement). Clock approximates LRU with just a bit flip — O(1) amortized.

### Real Systems
Linux page replacement (variant), OS virtual memory managers, database buffer pools

---

## 9. Redis Eviction Policies

### Available Policies
```
┌─────────────────────────────────────────────────────────┐
│ Policy              │ Behavior                          │
├─────────────────────┼───────────────────────────────────┤
│ noeviction          │ Return error when memory full     │
│ allkeys-lru         │ Evict LRU key from ALL keys      │
│ volatile-lru        │ Evict LRU key with TTL set       │
│ allkeys-lfu         │ Evict LFU key from ALL keys      │
│ volatile-lfu        │ Evict LFU key with TTL set       │
│ allkeys-random      │ Evict random key                 │
│ volatile-random     │ Evict random key with TTL        │
│ volatile-ttl        │ Evict key with shortest TTL      │
└─────────────────────┴───────────────────────────────────┘
```

### Choosing the Right Policy
```
Pure cache (all data expendable):     allkeys-lru or allkeys-lfu
Cache + persistent data mixed:        volatile-lru (only evict TTL keys)
Memory-constrained, want control:     noeviction (app handles errors)
```

### Redis Approximated LRU
Redis doesn't do true LRU (too expensive). It samples 5 random keys and evicts the oldest. Increasing `maxmemory-samples` improves accuracy.

### Summary
Redis offers 8 eviction policies. `allkeys-lru` is the most common for pure caching. Redis uses approximated LRU (sampling) for performance.
