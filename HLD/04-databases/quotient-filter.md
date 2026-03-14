# Quotient Filter

## 1. Concept Overview

### Definition
A **quotient filter** is a space-efficient probabilistic data structure for approximate set membership testing, similar to a Bloom filter. Unlike a standard Bloom filter, it supports **deletion**, **merging**, and **resizing** while maintaining comparable space efficiency and false positive rates. It stores a compact representation of elements using quotient and remainder from a hash function.

### Purpose
- **Membership testing**: "Is element x in the set?" (with possible false positives)
- **Deletion support**: Remove elements (Bloom filter cannot)
- **Mergeability**: Combine two quotient filters efficiently
- **Resizability**: Grow or shrink the filter
- **Space efficiency**: Comparable to Bloom filter (~10 bits per element for 1% FPR)

### Problems It Solves
- **Bloom filter limitation**: Cannot delete elements
- **Dynamic sets**: Add and remove elements over time
- **Distributed systems**: Merge filters from different nodes
- **Storage engines**: SSTable-level filters that need deletion (compaction)

---

## 2. Real-World Motivation

### SSD Storage Engines
- **LSM trees**: Quotient filters used in some storage engines for level metadata
- **Deletion**: When SSTables are compacted/merged, need to remove keys from filter
- **RocksDB variants**: Some use quotient filter for block-level filtering with delete support

### Databases
- **Approximate membership**: Like Bloom filter; avoid disk reads when key not present
- **Mutable filters**: Tables that support DELETE; need filter that supports deletion
- **Merge during compaction**: Combine filters when merging SSTables

### Distributed Systems
- **Set reconciliation**: Merge quotient filters from different nodes to approximate union
- **Sync protocols**: Efficient representation of "what I have" for sync

### Research & Academia
- **Bender et al. (2012)**: Original quotient filter paper
- **Alternative to counting Bloom filter**: Simpler, no overflow issues
- **Better than Cuckoo for merge**: Quotient filter merges in linear time

---

## 3. Architecture Diagrams

### Quotient Filter Structure

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     QUOTIENT FILTER STRUCTURE                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   Hash function h(x) produces p-bit value                                 │
│   Split into: quotient q (high bits) + remainder r (low bits)             │
│                                                                          │
│   h(x) = 0b 10110 010011  (example: p=10, q=5 bits, r=5 bits)            │
│            \____/ \_____/                                                │
│            quotient  remainder                                           │
│            q=22      r=19                                                │
│                                                                          │
│   TABLE: 2^q slots (buckets), each slot stores:                          │
│   - Remainder r (or fingerprint)                                         │
│   - 3 metadata bits: is_occupied, is_continuation, is_shifted             │
│                                                                          │
│   Slot layout (per bucket):                                              │
│   ┌─────────────┬──────────────────┬─────────────────────────────────┐  │
│   │ is_occupied │ is_continuation  │ is_shifted  │  remainder (r)     │  │
│   │    (1 bit)  │    (1 bit)       │   (1 bit)   │  (r bits)          │  │
│   └─────────────┴──────────────────┴─────────────────────────────────┘  │
│                                                                          │
│   is_occupied:  Bucket q has at least one element                        │
│   is_continuation: This slot is continuation of previous (collision)    │
│   is_shifted:   This slot was shifted from its canonical bucket         │
│                                                                          │
│   Linear probing: If slot full, shift right; use metadata to decode      │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Insert and Lookup Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     QUOTIENT FILTER OPERATIONS                            │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   INSERT x:                                                               │
│   1. h(x) → q, r                                                         │
│   2. Find canonical slot for bucket q (first slot with is_shifted=0     │
│      or is_continuation=0 for bucket q)                                  │
│   3. If slot empty: store r, set metadata                                 │
│   4. If slot occupied: linear probe right; shift elements; insert r      │
│   5. Update is_occupied for bucket q                                     │
│                                                                          │
│   LOOKUP x:                                                               │
│   1. h(x) → q, r                                                         │
│   2. If is_occupied[q]=0 → definitely not in set                         │
│   3. Find run of elements for bucket q (linear scan using metadata)      │
│   4. Check if r in run → maybe in set (or false positive)                │
│                                                                          │
│   DELETE x:                                                               │
│   1. Find slot containing r for bucket q                                  │
│   2. Remove r; compact run (shift left)                                   │
│   3. Update metadata                                                      │
│   (Bloom filter cannot do this!)                                          │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Quotient Filter vs Bloom Filter

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     QUOTIENT vs BLOOM FILTER                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   BLOOM FILTER                    QUOTIENT FILTER                        │
│   ┌─────────────────┐            ┌─────────────────┐                    │
│   │  Bit array      │            │  Slot array     │                    │
│   │  [0,1,1,0,1,...] │            │  [r,metadata]   │                    │
│   │  No structure   │            │  Linear probing │                    │
│   │  No delete      │            │  Delete: YES    │                    │
│   └─────────────────┘            └─────────────────┘                    │
│                                                                          │
│   Space: ~10 bits/elem (1% FPR)   Space: ~10-12 bits/elem (1% FPR)       │
│   Delete: NO                     Delete: YES                            │
│   Merge: XOR (approximate)        Merge: Sort-merge (exact)               │
│   Resize: Rebuild                 Resize: Possible                       │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Core Mechanics

### Hash Function
- **Single hash**: h(x) produces p-bit value (e.g., p=32 or 64)
- **Split**: High bits = quotient q (index into 2^q buckets), low bits = remainder r (stored)
- **Quotient**: q = h(x) >> r_bits
- **Remainder**: r = h(x) & ((1 << r_bits) - 1)

### Metadata Bits (3 bits per slot)
- **is_occupied**: This bucket has at least one element
- **is_continuation**: This slot is part of a run that started in a previous slot (collision)
- **is_shifted**: This slot's element hashed to a bucket to the left (linear probing)

### Linear Probing
- Collisions: Multiple elements can hash to same quotient
- Store in contiguous "run"; use linear probing to find empty slot
- Metadata allows correct decoding during lookup

### Deletion
- Find slot with matching remainder
- Remove it; shift subsequent elements in run left
- Update metadata
- **No overflow**: Unlike counting Bloom filter, no counter overflow

### Merging
- Two quotient filters with same parameters (q, r)
- Sort-merge: Walk both in order; output union
- **Exact**: Merged filter has no extra false positives from merge process
- **Linear time**: O(n + m)

### Resizing
- Create new filter with different size (more buckets)
- Rehash all elements from old filter into new
- **Amortized**: Can do incremental resize

---

## 5. Numbers

| Metric | Quotient Filter | Bloom Filter |
|--------|-----------------|--------------|
| Bits per element (1% FPR) | ~10-12 | ~10 |
| Delete support | Yes | No (standard) |
| Merge | Yes (exact) | Approximate (OR) |
| Resize | Yes | Rebuild |
| Lookup time | O(1) expected | O(k) |
| Insert time | O(1) expected | O(k) |
| Space overhead | 3 metadata bits/slot | 0 |

### Space Breakdown
- **Remainder**: r bits (e.g., r=8 → 256 possible remainders per bucket)
- **Metadata**: 3 bits per slot
- **Load factor**: ~0.75 for good performance; more = longer runs
- **Total**: (r + 3) bits per slot; multiple slots per "element" due to load factor

---

## 6. Tradeoffs

### Quotient Filter vs Bloom Filter

| Aspect | Quotient Filter | Bloom Filter |
|--------|-----------------|--------------|
| **Delete** | Yes | No |
| **Merge** | Exact | OR (more FPR) |
| **Resize** | Yes | Rebuild |
| **Space** | Slightly more (metadata) | Slightly less |
| **Complexity** | Higher | Lower |
| **Implementation** | More complex | Simple |

### Quotient Filter vs Cuckoo Filter

| Aspect | Quotient Filter | Cuckoo Filter |
|--------|-----------------|--------------|
| **Delete** | Yes | Yes |
| **Merge** | Yes (easier) | Harder |
| **Space** | ~10-12 bits/elem | ~8-10 bits/elem |
| **FPR** | Similar | Slightly lower |
| **Implementation** | Medium | Complex |

### Quotient Filter vs Count-Min Sketch

| Aspect | Quotient Filter | Count-Min Sketch |
|--------|-----------------|------------------|
| **Use case** | Membership | Frequency estimation |
| **Delete** | Yes | Decrement |
| **False positive** | Yes | Over-estimation |
| **Retrieval** | No (membership only) | Approximate count |

---

## 7. Variants / Implementations

### Variants

**Basic Quotient Filter**
- Original design; 3 metadata bits
- Supports insert, lookup, delete

**Rank-Select Quotient Filter**
- Optimized for cache performance
- Uses rank-select structure for faster run finding

**Saturated Quotient Filter**
- When load factor too high; performance degrades
- May need resize or rebuild

### Implementations
- **Research implementations**: C++ in papers
- **Rust**: `qfilter` crate
- **Python**: Various academic implementations
- **Database engines**: Some LSM implementations use quotient filter
- **Less common than Bloom**: Bloom filter more widely used; quotient filter when delete/merge needed

---

## 8. Scaling Strategies

- **Resize**: Double buckets when load factor high; rehash
- **Sharding**: Partition by key prefix; multiple quotient filters
- **Merge**: Combine filters from different shards for global membership
- **Tiered**: Different filters for hot vs cold data

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|------------|
| **False positive** | Unnecessary expensive check | Tune parameters; same as Bloom |
| **High load factor** | Long runs; slow lookup | Resize; keep load < 0.75 |
| **Corruption** | Undefined behavior | Checksum; rebuild from source |
| **Merge with different params** | Invalid filter | Ensure same q, r |

---

## 10. Performance Considerations

- **Load factor**: Keep < 0.75 for O(1) expected operations
- **Cache locality**: Runs are contiguous; good for cache
- **Hash quality**: Critical; use good hash (Murmur, xxHash)
- **Metadata overhead**: 3 bits per slot; ~30% overhead for small r

---

## 11. Use Cases

| Use Case | Why Quotient Filter |
|----------|---------------------|
| **LSM compaction** | Delete keys when SSTable merged |
| **Mutable cache filter** | Cache eviction = delete |
| **Set reconciliation** | Merge filters from nodes |
| **Dynamic blocklists** | Add/remove URLs, IPs |
| **SSTable metadata** | Per-level filter with compaction |
| **When Bloom won't work** | Need deletion |

---

## 12. Comparison Tables

### Probabilistic Structures Comparison

| Structure | Membership | Delete | Merge | Space (1% FPR) | Use Case |
|-----------|------------|--------|-------|----------------|----------|
| **Bloom filter** | Yes | No | OR | ~10 bits | SSTable, cache |
| **Quotient filter** | Yes | Yes | Yes | ~10-12 bits | Mutable sets |
| **Cuckoo filter** | Yes | Yes | Hard | ~8-10 bits | Cache, delete |
| **Count-Min Sketch** | No | N/A | Yes | Varies | Frequency |
| **HyperLogLog** | No | N/A | Yes | O(log log n) | Cardinality |

### When to Use Quotient Filter

| Use Quotient Filter | Use Bloom Filter |
|---------------------|------------------|
| Need deletion | No deletion |
| Need merge | Simple OR merge OK |
| Need resize | Fixed size OK |
| Mutable set | Immutable set |
| LSM with compaction | Static SSTable |

---

## 13. Code or Pseudocode

### Quotient Filter (Simplified)

```python
class QuotientFilter:
    def __init__(self, q, r):
        # q = quotient bits (2^q buckets), r = remainder bits
        self.q = q
        self.r = r
        self.num_slots = 2 ** q
        self.slots = [None] * (self.num_slots * 2)  # Oversized for probing
        # Each slot: (remainder, is_occupied, is_continuation, is_shifted)
    
    def _hash(self, x):
        h = hash(x) & ((1 << (self.q + self.r)) - 1)
        quotient = h >> self.r
        remainder = h & ((1 << self.r) - 1)
        return quotient, remainder
    
    def insert(self, x):
        q, r = self._hash(x)
        # Find empty slot in run for bucket q (simplified)
        slot = self._find_slot(q, r)
        self.slots[slot] = (r, True, False, False)
        # Update metadata; handle linear probing...
    
    def contains(self, x):
        q, r = self._hash(x)
        if not self._is_occupied(q):
            return False
        run = self._get_run(q)
        return r in run  # Maybe in set (false positive possible)
    
    def delete(self, x):
        q, r = self._hash(x)
        slot = self._find_slot_with_remainder(q, r)
        if slot is not None:
            self._remove_slot(slot)
            # Compact run
```

### Merge Two Quotient Filters

```python
def merge_qf(qf1: QuotientFilter, qf2: QuotientFilter) -> QuotientFilter:
    """Merge two quotient filters with same parameters."""
    assert qf1.q == qf2.q and qf1.r == qf2.r
    result = QuotientFilter(qf1.q, qf1.r)
    # Walk both filters in slot order; add elements to result
    # Sort-merge: elements appear in order by (q, r)
    for q in range(2 ** qf1.q):
        run1 = qf1._get_run(q)
        run2 = qf2._get_run(q)
        for r in sorted(set(run1) | set(run2)):
            result._insert_remainder(q, r)
    return result
```

### Comparison with Bloom

```python
# Bloom: Cannot delete
bloom.add("x")
bloom.contains("x")  # True
# bloom.remove("x")  # NOT POSSIBLE

# Quotient: Can delete
qf.insert("x")
qf.contains("x")  # True
qf.delete("x")
qf.contains("x")  # False (or false positive from collision)
```

---

## 14. Interview Discussion

### Key Points to Articulate
1. **Definition**: Space-efficient probabilistic membership; supports delete, merge, resize
2. **Mechanics**: Quotient + remainder from hash; linear probing; 3 metadata bits
3. **Advantage over Bloom**: Deletion, exact merge, resize
4. **Use case**: When you need mutable set (LSM compaction, cache eviction)
5. **Tradeoff**: Slightly more complex and more space than Bloom

### Common Interview Questions
- **"What is a quotient filter?"** → Probabilistic membership structure; like Bloom but supports delete and merge
- **"When would you use quotient filter over Bloom?"** → When you need to delete elements or merge filters
- **"How does it support deletion?"** → Store remainder in slots; remove and compact run; Bloom can't (bits shared)
- **"Compare quotient filter and Cuckoo filter"** → Both support delete; quotient easier to merge; Cuckoo slightly less space
- **"Where is it used?"** → Storage engines with compaction, mutable caches, set reconciliation

### Red Flags to Avoid
- Saying quotient filter has no false positives (it does)
- Confusing with Bloom filter (no deletion)
- Not mentioning merge/resize advantages

### Ideal Answer Structure
1. Define quotient filter (probabilistic membership, quotient+remainder)
2. Key advantage: deletion, merge (vs Bloom)
3. How it works: hash split, linear probing, metadata
4. Compare with Bloom, Cuckoo
5. Use cases: LSM compaction, mutable sets
