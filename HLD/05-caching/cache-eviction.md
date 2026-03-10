# Cache Eviction Policies: Staff+ Engineer Deep Dive

## 1. Concept Overview

### Definition
Cache eviction (or replacement) policies are algorithms that determine **which cached item to remove** when the cache is full and a new item must be stored. The goal: maximize cache hit ratio by retaining the "most valuable" items.

### Purpose
- **Bounded memory**: Caches have finite capacity; eviction enables bounded memory usage
- **Hit ratio optimization**: Keep items that will be accessed again; evict those that won't
- **Predictability**: Deterministic or statistically predictable behavior under load

### Why It Exists
Without eviction, caches would grow unbounded until OOM. Eviction transforms a cache from "unbounded storage" to "bounded working set approximation." The policy encodes assumptions about access patterns (temporal locality, frequency, etc.).

### Problems It Solves
1. **Memory pressure**: Prevent cache from consuming all available RAM
2. **Working set dynamics**: Adapt to changing access patterns (hot keys shift over time)
3. **Cost control**: Cloud caches (Redis, Memcached) charge by memory—smaller = cheaper
4. **Performance**: Eviction overhead must be low—O(1) or O(log n) per operation

---

## 2. Real-World Motivation

### Redis
- **Approximated LRU**: 24-bit timestamp, sample 5 keys, evict LRU among them (O(1))
- **LFU mode**: 8-bit counter, logarithmic decay, probabilistic eviction
- **Configurable**: `maxmemory-policy` = noeviction, allkeys-lru, volatile-lru, allkeys-lfu, volatile-lfu, volatile-ttl, allkeys-random, volatile-random

### Memcached
- **LRU per slab class**: Each slab (size bucket) has its own LRU list
- **Segmented LRU**: Hot/warm/cold segments to reduce eviction mistakes
- **No TTL eviction**: Purely size-based; TTL is expiration, not eviction

### Caffeine (Java)
- **Window TinyLFU (W-TinyLFU)**: Combines LRU (recent) + TinyLFU (frequency)
- **Count-Min Sketch**: Probabilistic frequency estimation with sub-linear space
- **99% hit ratio** on many workloads vs. 95% for LRU

### Varnish (HTTP cache)
- **LRU with TTL**: Evict expired first, then LRU among valid
- **ESI (Edge Side Includes)**: Per-fragment eviction

### Linux Page Cache
- **LRU lists**: Active/inactive, two-list approximation
- **Swap**: Evict to swap when memory pressure high

---

## 3. Architecture Diagrams

### LRU: Doubly Linked List + HashMap

```
                    LRU CACHE STRUCTURE
    ┌─────────────────────────────────────────────────────────┐
    │  HashMap: key → Node                                      │
    │  ┌─────┬─────┬─────┬─────┐                               │
    │  │ k1  │ k2  │ k3  │ k4  │  →  Node pointers             │
    │  └──┬──┴──┬──┴──┬──┴──┬──┘                               │
    │     │     │     │     │                                   │
    │     ▼     ▼     ▼     ▼                                   │
    │  Doubly Linked List (MRU ← → LRU)                         │
    │  ┌────┐   ┌────┐   ┌────┐   ┌────┐   ┌────┐               │
    │  │head│──▶│ k1 │──▶│ k2 │──▶│ k3 │──▶│ k4 │──▶│tail│     │
    │  └────┘   └──▲─┘   └────┘   └────┘   └─▲──┘   └────┘     │
    │       MRU   │                          │         LRU      │
    │             └──────────────────────────┘                  │
    │  On access: move to head (MRU)                            │
    │  On evict: remove tail (LRU)                              │
    └─────────────────────────────────────────────────────────┘
```

### LFU: Frequency Buckets + Doubly Linked Lists

```
                    LFU CACHE STRUCTURE
    ┌─────────────────────────────────────────────────────────┐
    │  freq_map: frequency → DoublyLinkedList of (key, value)  │
    │  key_map: key → (value, frequency, node_ref)             │
    │                                                          │
    │  freq=1: [k1] [k2] [k3]  ←  min_freq = 1                 │
    │  freq=2: [k4] [k5]                                       │
    │  freq=3: [k6]                                            │
    │                                                          │
    │  On access: move from freq bucket to freq+1 bucket       │
    │  On evict: remove from min_freq list (FIFO within freq)  │
    └─────────────────────────────────────────────────────────┘
```

### W-TinyLFU: Window + Main Cache

```
                    W-TINYLFU STRUCTURE
    ┌─────────────────────────────────────────────────────────┐
    │  ┌─────────────────┐    ┌─────────────────────────────┐ │
    │  │  Window Cache   │    │  Main Cache (TinyLFU)        │ │
    │  │  (1% of space)  │───▶│  (99% of space)              │ │
    │  │  LRU            │    │  LFU with Count-Min Sketch   │ │
    │  └─────────────────┘    └─────────────────────────────┘ │
    │           │                          │                   │
    │           │    Admission: Window victim vs Main victim   │
    │           │    Winner (higher freq) stays                 │
    │           ▼                          ▼                   │
    │  ┌─────────────────────────────────────────────────────┐│
    │  │  Count-Min Sketch (frequency estimation)            ││
    │  │  - Increment on access                               ││
    │  │  - Periodic decay (aging)                            ││
    │  └─────────────────────────────────────────────────────┘│
    └─────────────────────────────────────────────────────────┘
```

### ARC: Adaptive Replacement Cache

```
                    ARC STRUCTURE
    ┌─────────────────────────────────────────────────────────┐
    │  T1: LRU of items seen once (recency)                    │
    │  T2: LRU of items seen twice+ (frequency)                │
    │  B1: Ghost entries for T1 (evicted, no data)             │
    │  B2: Ghost entries for T2                                │
    │                                                          │
    │  Target: |T1| = p, |T2| = c - p                         │
    │  p adapts: if B1 hit, increase p (favor recency)        │
    │            if B2 hit, decrease p (favor frequency)       │
    │                                                          │
    │  On miss: if in B1 or B2, adapt p and replace            │
    │  On evict: from T1 or T2 based on p                      │
    └─────────────────────────────────────────────────────────┘
```

---

## 4. Core Mechanics

### LRU (Least Recently Used)

**Internal Workings:**
1. **Data structure**: HashMap + Doubly Linked List
2. **HashMap**: key → Node (for O(1) lookup)
3. **List**: head = MRU (most recently used), tail = LRU (least recently used)
4. **On get**: Move node to head (remove from current position, insert at head)
5. **On put**: If full, evict tail; insert new at head
6. **Complexity**: O(1) per operation

**Assumption**: Temporal locality—recently accessed items will be accessed again soon.

**Weakness**: One-time scan (e.g., full table scan) poisons cache—evicts everything. No frequency awareness.

### LFU (Least Frequently Used)

**Internal Workings:**
1. **Data structure**: HashMap (key → Node) + frequency buckets (freq → list of nodes)
2. **min_freq**: Track minimum frequency among cached items
3. **On get**: Increment freq, move node from freq bucket to freq+1 bucket
4. **On put**: If full, evict from min_freq list (FIFO within freq)
5. **Complexity**: O(1) with careful implementation (see O(1) LFU below)

**Assumption**: Frequency matters—items accessed often should stay.

**Weakness**: Old items with high frequency can "stick" forever (no aging). New items evicted quickly (no chance to grow).

### O(1) LFU Implementation

**Key insight**: Use doubly linked list per frequency bucket. Node stores: key, value, freq, pointer to list. HashMap: key → node. On increment: O(1) remove from list, O(1) add to next freq list. Maintain min_freq; when min_freq list empties, min_freq++.

### FIFO (First In First Out)

**Internal Workings:**
1. Simple queue: evict oldest inserted item
2. No access tracking—once in, eviction order fixed
3. O(1) but poor hit ratio

**Use case**: When no locality (random access pattern)

### Random

**Internal Workings:**
1. On eviction, pick random key
2. O(1), no metadata overhead
3. Unpredictable; can evict hot keys by chance

**Use case**: When eviction cost must be minimal; acceptable to evict hot keys occasionally

### TTL-Based (Time To Live)

**Internal Workings:**
1. Each entry has expiration timestamp
2. Eviction: remove expired items first (lazy or periodic sweep)
3. If no expired items, fall back to LRU/LFU or random
4. Redis `volatile-ttl`: evict keys with shortest TTL

**Use case**: When data has natural expiration (sessions, API responses)

### ARC (Adaptive Replacement Cache)

**Internal Workings:**
1. Maintains 4 lists: T1 (recency), T2 (frequency), B1 (ghost T1), B2 (ghost T2)
2. Parameter p balances recency vs. frequency
3. On ghost hit (B1 or B2): adapt p—increase if B1 hit (favor recency), decrease if B2 hit (favor frequency)
4. Combines LRU and LFU benefits; adapts to workload

**Complexity**: O(1) per operation. More memory (ghost entries).

### LIRS (Low Inter-reference Recency Set)

**Internal Workings:**
1. Uses "recency" (time since last access) and "IRR" (inter-reference recency—interval between last two accesses)
2. HIR (high IRR): infrequently accessed; LIR (low IRR): frequently accessed
3. Evict HIR items; promote LIR on access
4. Better than LRU for scan-resistant workloads

### W-TinyLFU (Window TinyLFU)

**Internal Workings:**
1. **Window**: Small LRU cache (1% of capacity)—admits new items
2. **Main**: TinyLFU cache (99%)—frequency-based
3. **Count-Min Sketch**: Approximate frequency for all keys (seen or not)
4. **Admission**: On eviction from window, compare window victim's freq vs. main victim's freq. Higher stays.
5. **Aging**: Periodically decay sketch to reduce impact of old accesses
6. **TinyLFU**: "Tiny" because sketch is small—only approximate counts

**Why it works**: Window captures recent items (burst); main captures frequency. Scan-resistant (one-time access doesn't pollute main).

---

## 5. Numbers

### Hit Ratio Targets

| Workload | LRU | LFU | ARC | W-TinyLFU | Target |
|----------|-----|-----|-----|-----------|--------|
| Zipf (α=0.9) | 90% | 95% | 95% | 98% | 95%+ |
| Loop (scan) | 0% | 95% | 90% | 95% | Scan-resistant |
| Search (LRU-friendly) | 95% | 90% | 95% | 95% | 95%+ |
| Mixed | 85% | 88% | 92% | 95% | 90%+ |

**Industry targets**: 90-99% for read-heavy caches. Below 90% = cache may not be worth the complexity.

### Memory Overhead

| Policy | Per-Entry Overhead | Notes |
|--------|-------------------|-------|
| LRU | 2 pointers (16B) + hash | Prev/next for list |
| LFU | 2 pointers + freq (4B) | ~20B |
| ARC | 4 pointers + 2 bits | T1/T2/B1/B2 |
| W-TinyLFU | Sketch + 2 pointers | Sketch ~4-8B per key (amortized) |
| FIFO | 1 pointer | Queue link |
| Random | 0 | Minimal |

### Redis Memory Management

- **maxmemory**: Set limit (e.g., 4GB)
- **maxmemory-policy**: Eviction policy when limit reached
- **Approximated LRU**: 24-bit timestamp per key (3B), sample 5 keys, evict LRU—99% of theoretical LRU accuracy with 10% memory overhead
- **LFU**: 8-bit counter, logarithmic decay (every 1M accesses), decay time 16 minutes

### Memcached Slab Allocation

- **Slab classes**: 1.25x growth (e.g., 64B, 80B, 100B, ...)
- **LRU per slab**: Each slab class has its own LRU list
- **Eviction**: Within slab class only—can't evict 64B item to make room for 1KB item
- **Slab reassignment**: Possible but expensive

---

## 6. Tradeoffs

### Consistency vs. Performance

| Policy | Eviction Cost | Hit Ratio | Scan Resistance | Memory |
|--------|---------------|-----------|------------------|--------|
| LRU | O(1) | Good | Poor | Low |
| LFU | O(1) | Good | Good | Medium |
| FIFO | O(1) | Poor | Poor | Lowest |
| Random | O(1) | Variable | N/A | Lowest |
| TTL | O(log n) or O(1) | Depends | N/A | Low |
| ARC | O(1) | Excellent | Good | High |
| W-TinyLFU | O(1) | Excellent | Excellent | Medium |

### When to Choose

| Policy | Best For | Avoid When |
|--------|----------|------------|
| LRU | Temporal locality, simple | Scans, one-time access |
| LFU | Frequency skew (Zipf) | Burst, new items |
| FIFO | No locality | Most cases |
| Random | Minimal overhead | Predictability needed |
| TTL | Expiring data | Size-based eviction |
| ARC | Mixed workloads | Memory constrained |
| W-TinyLFU | General purpose, scans | Very simple workloads |

---

## 7. Variants / Implementations

### LRU Variants

**1. Segmented LRU (SLRU):**
- Protected segment (hot): 80% of cache
- Probation segment (cold): 20%
- New items go to probation; promoted to protected on second access
- Evict from probation

**2. 2Q (Two Queue):**
- FIFO for first access
- LRU for second access
- Reduces scan pollution

**3. LRU-K:**
- Track last K access times
- Evict based on K-th order statistic (LRU-2 is common)
- More resistant to scans

### LFU Variants

**1. LFU with Aging:**
- Periodically decay all frequencies
- Prevents old items from sticking forever

**2. TinyLFU:**
- Count-Min Sketch for frequency (approximate)
- Periodic reset (aging)
- Sub-linear space

**3. LFU with Admission:**
- Only admit if new item's estimated freq > victim's
- W-TinyLFU uses this

### Redis Implementation

**Approximated LRU:**
- Each key has 24-bit "timestamp" (actually seconds since last access, with precision)
- On eviction: sample 5 random keys (or all if < 5), evict one with smallest timestamp
- Configurable: `maxmemory-samples 5` (higher = more accurate, slower)

**LFU:**
- 8-bit counter (0-255)
- Increment: probabilistic (1/(2^counter) chance to increment)
- Decay: every 1M accesses, counter halved
- Evict: smallest counter

---

## 8. Scaling Strategies

### Large Caches
- **Sharding**: Partition by key hash; each shard has own eviction
- **Tiered**: L1 (hot) small LRU, L2 (warm) large W-TinyLFU
- **Distributed**: Consistent hashing; eviction per node

### Dynamic Sizing
- **Adaptive**: Monitor hit ratio; adjust cache size or policy
- **Cost-aware**: Evict based on fetch cost (expensive to fetch = keep longer)

### Multi-Tenant
- **Per-tenant**: Separate eviction per tenant (fairness)
- **Weighted**: Evict from low-priority tenant first

---

## 9. Failure Scenarios

### Cache Pollution (Scan)
**Scenario**: Full table scan. LRU evicts all hot data. Hit ratio drops to 0%.
**Solution**: W-TinyLFU, LFU, or ARC. Scan-resistant policies.

### One-Hit Wonders
**Scenario**: Many unique keys accessed once. LFU evicts new items; cache never warms.
**Solution**: W-TinyLFU (window gives new items a chance). LRU. LRU-K with K=2.

### Frequency Bloat
**Scenario**: LFU—old item accessed 1000 times. Stays forever. New hot item evicted.
**Solution**: Aging (periodic decay). TinyLFU reset. LFU with max frequency cap.

### Eviction Storm
**Scenario**: Sudden traffic spike. Many evictions. Eviction overhead impacts latency.
**Solution**: Batch eviction. Async eviction. Pre-warm cache.

### Wrong Policy
**Scenario**: Using LRU for Zipf workload. LRU for scan workload.
**Solution**: Profile workload. Use adaptive (ARC) or general-purpose (W-TinyLFU).

---

## 10. Performance Considerations

### Eviction Overhead
- **LRU**: O(1)—pointer update
- **LFU**: O(1)—bucket move
- **ARC**: O(1)—list operations
- **W-TinyLFU**: O(1)—sketch increment + list
- **Random**: O(1)—random number + lookup

### Concurrency
- **LRU**: Need locks per operation, or lock-free with careful design
- **ConcurrentLinkedHashMap**: LRU with striping for concurrency
- **Caffeine**: Non-blocking for concurrent access

### Memory vs. Hit Ratio
- **Doubling cache**: ~10-20% hit ratio improvement (diminishing returns)
- **Policy change**: Can improve hit ratio 5-10% vs. LRU at same size
- **W-TinyLFU**: Often matches 2x cache size LRU with same memory

---

## 11. Use Cases

| Use Case | Policy | Rationale |
|----------|--------|-----------|
| Session cache | TTL + LRU | Sessions expire; LRU for active |
| User profile | LRU | Temporal locality |
| API response | TTL | Natural expiration |
| Search results | W-TinyLFU | Scan-resistant; mixed patterns |
| Database query | LRU or W-TinyLFU | Query patterns vary |
| CDN | TTL + LRU | Expiration + popularity |
| Object cache | LRU or LFU | Depends on access pattern |
| Rate limiting | Sliding window | Not eviction, but TTL-like |

---

## 12. Comparison Table

| Policy | Get | Put | Evict | Hit Ratio | Scan Resistance | Memory | Complexity |
|--------|-----|-----|-------|-----------|------------------|--------|------------|
| LRU | O(1) | O(1) | O(1) | Good | Poor | Low | Low |
| LFU | O(1) | O(1) | O(1) | Good | Good | Medium | Medium |
| FIFO | O(1) | O(1) | O(1) | Poor | Poor | Lowest | Lowest |
| Random | O(1) | O(1) | O(1) | Variable | N/A | Lowest | Lowest |
| TTL | O(1) | O(1) | O(log n) | Depends | N/A | Low | Low |
| ARC | O(1) | O(1) | O(1) | Excellent | Good | High | High |
| LIRS | O(1) | O(1) | O(1) | Excellent | Excellent | High | High |
| W-TinyLFU | O(1) | O(1) | O(1) | Excellent | Excellent | Medium | High |

---

## 13. Code / Pseudocode

### LRU Cache Implementation (HashMap + Doubly Linked List)

```python
class LRUCache:
    def __init__(self, capacity: int):
        self.capacity = capacity
        self.cache = {}  # key -> Node
        self.head = Node(0, 0)  # Sentinel
        self.tail = Node(0, 0)  # Sentinel
        self.head.next = self.tail
        self.tail.prev = self.head

    def _add_to_head(self, node):
        node.next = self.head.next
        node.prev = self.head
        self.head.next.prev = node
        self.head.next = node

    def _remove_node(self, node):
        node.prev.next = node.next
        node.next.prev = node.prev

    def get(self, key: int) -> int:
        if key not in self.cache:
            return -1
        node = self.cache[key]
        self._remove_node(node)
        self._add_to_head(node)

        return node.value

    def put(self, key: int, value: int):
        if key in self.cache:
            node = self.cache[key]
            node.value = value
            self._remove_node(node)
            self._add_to_head(node)
            return

        if len(self.cache) >= self.capacity:
            lru = self.tail.prev
            self._remove_node(lru)
            del self.cache[lru.key]

        node = Node(key, value)
        self.cache[key] = node
        self._add_to_head(node)

class Node:
    def __init__(self, key, value):
        self.key = key
        self.value = value
        self.prev = None
        self.next = None
```

### LFU Cache Implementation (O(1))

```python
class LFUCache:
    def __init__(self, capacity: int):
        self.capacity = capacity
        self.min_freq = 0
        self.key_to_node = {}  # key -> (value, freq, node)
        self.freq_to_list = defaultdict(OrderedDict)  # freq -> {key: node}

    def get(self, key: int) -> int:
        if key not in self.key_to_node:
            return -1
        value, freq, _ = self.key_to_node[key]
        self._update_freq(key, value, freq)
        return value

    def put(self, key: int, value: int):
        if self.capacity == 0:
            return
        if key in self.key_to_node:
            _, freq, _ = self.key_to_node[key]
            self._update_freq(key, value, freq)
            return

        if len(self.key_to_node) >= self.capacity:
            self._evict()

        self.min_freq = 1
        self.freq_to_list[1][key] = (value, 1)
        self.key_to_node[key] = (value, 1, None)

    def _update_freq(self, key, value, freq):
        del self.freq_to_list[freq][key]
        if not self.freq_to_list[freq] and freq == self.min_freq:
            self.min_freq += 1
        new_freq = freq + 1
        self.freq_to_list[new_freq][key] = (value, new_freq)
        self.key_to_node[key] = (value, new_freq, None)

    def _evict(self):
        od = self.freq_to_list[self.min_freq]
        key_to_remove = next(iter(od))
        del od[key_to_remove]
        del self.key_to_node[key_to_remove]
```

### W-TinyLFU (Simplified Pseudocode)

```python
class WTinyLFUCache:
    def __init__(self, capacity: int):
        self.capacity = capacity
        self.window_size = capacity // 100  # 1% window
        self.main_size = capacity - self.window_size
        self.window = LRUCache(self.window_size)
        self.main = LFUCache(self.main_size)
        self.sketch = CountMinSketch()  # Approximate frequency

    def get(self, key):
        self.sketch.increment(key)
        if key in self.window:
            return self.window.get(key)
        if key in self.main:
            return self.main.get(key)
        return None

    def put(self, key, value):
        self.sketch.increment(key)
        if key in self.window or key in self.main:
            # Update existing
            ...
        elif self.window.size() < self.window_size:
            self.window.put(key, value)
        else:
            # Evict from window
            victim_key = self.window.get_lru_key()
            victim_freq = self.sketch.estimate(victim_key)
            main_victim_key = self.main.get_lfu_key()
            main_victim_freq = self.sketch.estimate(main_victim_key)

            if victim_freq > main_victim_freq:
                self.main.evict(main_victim_key)
                self.main.put(victim_key, self.window.get(victim_key))
            self.window.evict(victim_key)
            self.window.put(key, value)
```

### Count-Min Sketch (Simplified)

```python
class CountMinSketch:
    def __init__(self, width=1000, depth=5):
        self.width = width
        self.depth = depth
        self.table = [[0] * width for _ in range(depth)]
        self.hash_seeds = [random.randint(0, 2**32) for _ in range(depth)]

    def _hash(self, key, seed):
        return (hash(key) ^ seed) % self.width

    def increment(self, key):
        for i in range(self.depth):
            j = self._hash(key, self.hash_seeds[i])
            self.table[i][j] += 1

    def estimate(self, key):
        return min(self.table[i][self._hash(key, self.hash_seeds[i])]
                   for i in range(self.depth))
```

---

## 14. Interview Discussion

### Key Points to Articulate

1. **"LRU is simple and works well for temporal locality, but fails on scans."** One full table scan evicts everything. Use W-TinyLFU or LFU for scan-resistant workloads.

2. **"LRU is O(1) with a doubly linked list and hash map."** HashMap for lookup; list for ordering. On access: move to head. On evict: remove tail.

3. **"LFU has a problem: old items with high frequency stick forever."** Solution: aging (periodic decay) or TinyLFU with reset.

4. **"W-TinyLFU combines recency (window) and frequency (main) with a Count-Min Sketch."** Used in Caffeine. 99% hit ratio on many workloads. Scan-resistant.

5. **"Redis uses approximated LRU—samples 5 keys, evicts LRU."** Tradeoff: 10% memory overhead for 99% of theoretical LRU accuracy. O(1) eviction.

### Follow-Up Questions

- **"Implement LRU cache."** HashMap + doubly linked list. Get: lookup, move to head. Put: evict tail if full, add to head.
- **"What's the difference between LRU and LFU?"** LRU = recency; LFU = frequency. LRU: "recently used"; LFU: "used most often."
- **"How does W-TinyLFU work?"** Window (LRU) + main (LFU) + Count-Min Sketch. Admission: new item competes with victim; higher estimated frequency wins.
- **"Why does Redis use approximated LRU?"** True LRU needs per-key metadata (24B+). Approximated: 24-bit timestamp, sample 5 keys. Same memory, 99% accuracy.

### Red Flags to Avoid

- Saying "LRU is O(n)" (it's O(1) with proper data structures)
- Ignoring scan resistance (LRU fails on full scans)
- Not knowing LFU's "sticky" problem (old items never evicted)
- Not mentioning W-TinyLFU or Caffeine for modern workloads
