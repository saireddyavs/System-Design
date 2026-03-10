# Bloom Filters

## Staff+ Engineer Deep Dive for FAANG Interviews

---

## 1. Concept Overview

### Definition
A **Bloom filter** is a probabilistic, space-efficient data structure used to test whether an element is a member of a set. It can produce **false positives** (incorrectly report that an element is in the set) but never **false negatives** (if it says "not in set," the element is definitely not in the set).

### Purpose
- **Fast membership testing**: O(k) where k = number of hash functions (typically 5-10)
- **Space efficiency**: Far less memory than storing full elements (e.g., 10 bits per element for 1% false positive rate)
- **Avoid expensive operations**: Skip disk reads, network calls when element definitely not present

### Why It Exists
When checking set membership is expensive (disk I/O, network request) and the common case is "not present," a Bloom filter provides a cheap pre-check. If the filter says "maybe," do the expensive check; if "definitely not," skip it and save cost.

### Problems Solved
| Problem | Bloom Filter Solution |
|---------|----------------------|
| Expensive membership check | Cheap probabilistic pre-check |
| Large set in memory | Compact representation |
| Avoid unnecessary disk reads | Filter before SSTable lookup |
| Malicious URL check | Fast "not in blocklist" |

---

## 2. Real-World Motivation

### Cassandra
- Each SSTable has a Bloom filter
- Before reading from disk, check Bloom filter
- If "not present," skip disk read entirely
- Saves millions of disk seeks at scale

### Google Chrome
- Malicious URL check: billions of URLs in blocklist
- Bloom filter: "definitely safe" → skip full check
- "Maybe malicious" → full lookup (rare)
- Enables fast safe browsing

### Akamai (CDN)
- Cache lookups: "Is this object in cache?"
- Bloom filter at edge: avoid upstream lookup if not in cache
- Reduces origin load, latency

### HBase
- Per HFile (similar to SSTable) Bloom filter
- Avoid reading HFile when key not present
- Critical for read performance

### Bitcoin (Simplified Payment Verification)
- Bloom filters for wallet: "Does this block have my transactions?"
- Light clients avoid downloading full blocks

### Database Systems
- PostgreSQL: Optional Bloom index for multiple columns
- Bigtable: Bloom filter per SSTable
- RocksDB: Bloom filter per SSTable block

---

## 3. Architecture Diagrams

### Bloom Filter Operation
```
┌─────────────────────────────────────────────────────────────────────────┐
│                    BLOOM FILTER OPERATION                                 │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   INSERT "x":                                                            │
│   ┌─────┐     h1(x)=2   h2(x)=5   h3(x)=9                               │
│   │ "x" │ ──────────────────────────────────────────▶                  │
│   └─────┘                                                               │
│                                                                          │
│   Bit array (m bits):                                                    │
│   Index:  0   1   2   3   4   5   6   7   8   9  10  11  ...            │
│   Before: 0   0   0   0   0   0   0   0   0   0   0   0                 │
│   After:  0   0   1   0   0   1   0   0   0   1   0   0                 │
│                 ▲           ▲           ▲                                │
│                 │           │           │                                │
│            Set bits at positions 2, 5, 9                                 │
│                                                                          │
│   LOOKUP "x":                                                            │
│   Check h1(x), h2(x), h3(x) → all 1? YES → "maybe in set"                │
│                                                                          │
│   LOOKUP "y" (not in set):                                               │
│   Check h1(y), h2(y), h3(y) → any 0? YES → "definitely not in set"      │
│                                                                          │
│   FALSE POSITIVE: "z" hashes to same positions as some "x"               │
│   All bits 1 → "maybe in set" (wrong!) → expensive check needed         │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Bloom Filter in LSM Tree (Cassandra/RocksDB)
```
┌─────────────────────────────────────────────────────────────────────────┐
│                    BLOOM FILTER IN LSM TREE                               │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   Lookup key "user:12345"                                                │
│        │                                                                 │
│        ▼                                                                 │
│   ┌─────────────┐                                                        │
│   │  Memtable   │  Check in-memory first                                 │
│   └──────┬──────┘                                                        │
│          │ Not found                                                     │
│          ▼                                                               │
│   ┌─────────────┐     ┌─────────────┐     ┌─────────────┐               │
│   │ SSTable 1   │     │ SSTable 2   │     │ SSTable 3   │               │
│   │ + Bloom     │     │ + Bloom     │     │ + Bloom     │               │
│   │   Filter    │     │   Filter    │     │   Filter    │               │
│   └──────┬──────┘     └──────┬──────┘     └──────┬──────┘               │
│          │                   │                   │                       │
│          │ "maybe"           │ "no"              │ "no"                  │
│          │ (disk read)       │ (skip!)           │ (skip!)               │
│          ▼                   │                   │                       │
│   Read from disk             │                   │                       │
│   (only if "maybe")          │                   │                       │
│                              │                   │                       │
│   Result: 1 disk read instead of 3 (or N)                                │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Counting Bloom Filter
```
┌─────────────────────────────────────────────────────────────────────────┐
│                    COUNTING BLOOM FILTER                                  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   Standard Bloom: Cannot delete (setting bit to 0 might remove another)  │
│   Counting Bloom: Use counters instead of bits                           │
│                                                                          │
│   Insert "x": Increment positions 2, 5, 9                                │
│   Delete "x": Decrement positions 2, 5, 9                                │
│                                                                          │
│   Index:  0   1   2   3   4   5   6   7   8   9                         │
│   Count:  0   0   2   0   0   1   0   0   0   1   (after inserts)        │
│                                                                          │
│   Tradeoff: 4x more space (4-bit counters vs 1 bit)                      │
│   Overflow: Counters can overflow; use Cuckoo filter for better delete   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Cuckoo Filter (Variant)
```
┌─────────────────────────────────────────────────────────────────────────┐
│                    CUCKOO FILTER (vs Bloom)                                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   - Supports deletions (unlike standard Bloom)                           │
│   - Lower false positive rate for same space                             │
│   - Uses fingerprint + bucket structure                                 │
│   - Insert: place fingerprint in bucket; if full, kick existing         │
│   - Lookup: check 2 possible buckets                                     │
│                                                                          │
│   Space: ~95% of Bloom for same FPR; supports delete                     │
│   Use case: When deletion needed (cache, etc.)                          │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Core Mechanics

### Hash Functions
- **k hash functions**: Typically 5-10; can use double hashing: h_i(x) = h1(x) + i * h2(x)
- **Uniform distribution**: Critical for low false positive rate
- **Independent**: Minimize correlation between hash functions

### Bit Array Size
- **m bits**: Larger m = lower false positive rate, more space
- **Optimal m**: m = -n * ln(p) / (ln(2)^2) where n = elements, p = desired FPR
- **Example**: 1M elements, 1% FPR → m ≈ 9.6M bits ≈ 1.2 MB

### Optimal Number of Hash Functions
- **k = (m/n) * ln(2)**: Minimizes false positive rate for given m, n
- **Example**: m/n = 10 → k ≈ 7 hash functions
- **Tradeoff**: More hashes = more computation, but better FPR

### False Positive Rate Formula
- **FPR ≈ (1 - e^(-kn/m))^k**
- With optimal k: FPR ≈ 0.6185^(m/n)
- **1% FPR**: m/n ≈ 9.6 bits per element
- **0.1% FPR**: m/n ≈ 14.4 bits per element

---

## 5. Numbers

| Metric | Value |
|--------|-------|
| 1% false positive rate | ~10 bits per element |
| 0.1% FPR | ~14.4 bits per element |
| 0.01% FPR | ~19.2 bits per element |
| Optimal k (1% FPR) | ~7 hash functions |
| Space vs hash table | 10-100x smaller (no keys stored) |
| Lookup time | O(k) = O(1) for constant k |

### Real Scale
- **Cassandra SSTable**: Bloom filter ~1-2% of SSTable size
- **Chrome**: Billions of URLs; Bloom filter in MBs
- **1M elements, 1% FPR**: ~1.2 MB Bloom vs ~10+ MB for hash table (with keys)

---

## 6. Tradeoffs

### Bloom Filter Tradeoffs
| Aspect | Benefit | Cost |
|--------|---------|------|
| Space | Very compact | Cannot store keys |
| False negatives | Never | N/A |
| False positives | Acceptable | Extra lookup when occurs |
| Deletion | Not supported (standard) | Use counting/cuckoo |
| Speed | O(k) lookup | k hash computations |

### Bloom vs Hash Table
| Aspect | Bloom Filter | Hash Table |
|--------|--------------|------------|
| Space | O(n * bits) | O(n * key_size) |
| False positives | Yes | No |
| False negatives | No | No |
| Deletion | No (standard) | Yes |
| Retrieval | Membership only | Full key-value |

---

## 7. Variants / Implementations

### Variants
- **Standard Bloom**: Insert, lookup; no delete
- **Counting Bloom**: Counters; supports delete (overflow risk)
- **Cuckoo Filter**: Supports delete; lower FPR; more complex
- **Blocked Bloom**: Cache-friendly; blocks of small Bloom filters
- **Scalable Bloom**: Grows dynamically; multiple filters

### Implementations
- **Guava** (Java): com.google.common.hash.BloomFilter
- **Redis**: BF.ADD, BF.EXISTS (RedisBloom module)
- **Cassandra**: Built-in per SSTable
- **RocksDB**: Built-in
- **PostgreSQL**: bloom extension for indexes

---

## 8. Scaling Strategies

- **Scalable Bloom**: Add new filter when FPR exceeds threshold
- **Sharded Bloom**: Partition by key prefix; smaller filters
- **Tiered**: Different FPR for hot vs cold data

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|-------------|
| False positive | Unnecessary expensive operation | Tune m, k; accept tradeoff |
| Filter corruption | Wrong "not in set" (false negative?) | Bloom never has false negative; corruption = undefined |
| Too small m | High FPR | Size for expected n |
| Hash collision (bad hash) | Higher FPR | Use quality hash (Murmur, xxHash) |

---

## 10. Performance Considerations

- **Hash quality**: Poor hash = higher FPR
- **Cache locality**: Blocked Bloom for better CPU cache use
- **k tradeoff**: More k = lower FPR but more CPU
- **Memory access**: Bit array is compact; cache-friendly

---

## 11. Use Cases

| Use Case | Why Bloom Filter |
|----------|------------------|
| SSTable lookup (Cassandra, HBase) | Avoid disk read when key not in file |
| Malicious URL (Chrome) | Fast "safe" check; skip full lookup |
| CDN cache (Akamai) | Avoid origin lookup when not in cache |
| Spell check | "Word in dictionary?" |
| Distributed systems | "Have we seen this request?" (dedup) |
| Database | Skip index/table scan when key not present |

---

## 12. Comparison Tables

### Bloom Filter Variants
| Variant | Delete | Space | FPR | Complexity |
|---------|--------|-------|-----|------------|
| Standard Bloom | No | 1x | p | Low |
| Counting Bloom | Yes | 4x | p | Medium |
| Cuckoo Filter | Yes | 0.95x | Lower | High |

### Bits Per Element for FPR
| False Positive Rate | Bits/Element | k (optimal) |
|---------------------|--------------|-------------|
| 1% | 9.6 | 7 |
| 0.1% | 14.4 | 10 |
| 0.01% | 19.2 | 14 |
| 0.001% | 28.8 | 20 |

---

## 13. Code or Pseudocode

### Bloom Filter Implementation
```python
import hashlib

class BloomFilter:
    def __init__(self, size, num_hashes):
        self.size = size
        self.num_hashes = num_hashes
        self.bit_array = [0] * size
    
    def _hashes(self, item):
        h1 = int(hashlib.md5(str(item).encode()).hexdigest(), 16) % self.size
        h2 = int(hashlib.sha1(str(item).encode()).hexdigest(), 16) % self.size
        for i in range(self.num_hashes):
            yield (h1 + i * h2) % self.size
    
    def add(self, item):
        for h in self._hashes(item):
            self.bit_array[h] = 1
    
    def contains(self, item):
        for h in self._hashes(item):
            if self.bit_array[h] == 0:
                return False  # Definitely not in set
        return True  # Maybe in set (could be false positive)
```

### Optimal Parameters
```python
import math

def optimal_bloom_params(n, p):
    """
    n = expected number of elements
    p = desired false positive rate
    Returns (m, k): bits and number of hash functions
    """
    m = -n * math.log(p) / (math.log(2) ** 2)
    k = (m / n) * math.log(2)
    return int(m), int(k)

# 1M elements, 1% FPR
m, k = optimal_bloom_params(1_000_000, 0.01)
print(f"m = {m} bits ({m/8/1024:.1f} KB), k = {k}")
# m = 9585059 bits (1170.4 KB), k = 6
```

### Cassandra-Style Usage
```python
def sstable_lookup(key, sstables):
    for sst in sstables:
        if sst.bloom_filter.contains(key):
            # Maybe in this SSTable; do actual disk read
            result = sst.get(key)
            if result is not None:
                return result
        # "Definitely not" in this SSTable; skip
    return None
```

---

## 14. Interview Discussion

### Key Points
1. **No false negatives**: "Not in set" is definitive
2. **False positives possible**: "In set" might be wrong; do expensive check
3. **Space efficient**: ~10 bits per element for 1% FPR
4. **Cannot delete**: Standard Bloom; use counting or cuckoo
5. **Use when**: Expensive check + common case is "not present"

### Common Questions
- **Q**: "What's the difference between false positive and false negative?"
  - **A**: FP = says "in set" when not; FN = says "not in set" when it is. Bloom has FP, never FN
- **Q**: "Why can't you delete from a Bloom filter?"
  - **A**: Setting a bit to 0 might remove another element that hashed to same position
- **Q**: "How do you choose m and k?"
  - **A**: m from desired FPR and n: m = -n*ln(p)/(ln2)^2; k = (m/n)*ln2
- **Q**: "When would you use a Bloom filter?"
  - **A**: When membership check is expensive (disk, network) and "not present" is common; e.g., Cassandra SSTable lookup, Chrome URL blocklist

### Red Flags to Avoid
- Saying Bloom filter has false negatives
- Using Bloom when you need to retrieve the element (it only tests membership)
- Ignoring false positive rate in design

---

## 15. Mathematical Derivation of False Positive Rate

### Assumptions
- k hash functions
- m bits
- n elements
- Hash functions are independent and uniform

### Derivation
- Probability a specific bit is 0 after n inserts: (1 - 1/m)^(kn) ≈ e^(-kn/m)
- For a new element (not in set), all k positions must be 1
- Probability all k positions are 1: (1 - e^(-kn/m))^k
- This is the false positive rate

### Optimal k
- Minimize (1 - e^(-kn/m))^k with respect to k
- k_opt = (m/n) * ln(2) ≈ 0.693 * (m/n)
- At optimal k: FPR = 0.6185^(m/n)

---

## 16. Blocked Bloom Filter

### Concept
- Divide bit array into blocks (e.g., 256 bits each)
- Each block is a small Bloom filter
- Key hashes to block index; then k bits within block
- **Benefit**: Better cache locality; one block fits in cache line
- **Tradeoff**: Slightly higher FPR for same m (block boundaries)

### Use Case
- When lookup latency is critical
- L1/L2 cache friendly
- Used in some database implementations

---

## 17. Scalable Bloom Filter

### Problem
- Standard Bloom: fixed size; if you insert more than n, FPR increases
- Don't know n in advance

### Solution
- Chain of Bloom filters
- When FPR of current filter exceeds threshold, add new larger filter
- Lookup: check all filters; "in set" if any says yes
- **Growth**: Each new filter has lower FPR; overall FPR bounded
- **Space**: Slightly more than single filter for same n

---

## 18. Redis Bloom Module Example

```redis
# Create filter with 1% FPR, 1M capacity
BF.RESERVE myfilter 0.01 1000000

# Add elements
BF.ADD myfilter "user:12345"
BF.ADD myfilter "user:67890"

# Check membership
BF.EXISTS myfilter "user:12345"   # 1 (maybe in set)
BF.EXISTS myfilter "user:99999"   # 0 (definitely not)

# Bulk add
BF.MADD myfilter "a" "b" "c"
```

### Use Case
- Deduplication: "Have we processed this ID?"
- Cache warming: "Is this in cache?" before lookup
- Rate limiting: "Have we seen this IP in window?"

---

## 19. Comparison: Bloom vs Other Structures

| Structure | Membership | Space | False Positive | False Negative | Delete |
|-----------|------------|-------|----------------|----------------|--------|
| Hash table | O(1) | O(n × key_size) | No | No | Yes |
| Bloom filter | O(k) | O(n × bits) | Yes | No | No |
| Cuckoo filter | O(1) | ~0.95× Bloom | Yes | No | Yes |
| HyperLogLog | Cardinality | O(log log n) | N/A | N/A | No |

---

## 20. Production Considerations

### Memory
- Bloom filter lives in memory (or disk for Cassandra SSTable)
- Size = m bits; plan for growth
- Redis: BF.RESERVE with error rate and capacity

### Hash Function Selection
- **MurmurHash**: Fast, good distribution
- **xxHash**: Very fast, good for large data
- **Double hashing**: h_i = h1 + i * h2 reduces to 2 hash functions
