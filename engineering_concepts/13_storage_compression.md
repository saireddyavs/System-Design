# Module 13: Storage & Compression

---

## 1. Delta Encoding

### Definition
Storing only the difference (delta) between consecutive values instead of full values. Extremely effective for time-series and sorted data.

### How It Works
```
Raw timestamps:     [1000, 1001, 1002, 1005, 1010]
Delta encoded:      [1000, 1, 1, 3, 5]

Delta-of-delta:     [1000, 1, 0, 2, 2]   ← even smaller!

Raw:    64 bits × 5 = 320 bits
Delta:  64 + 4×3 = 76 bits (~4x compression)
```

### Facebook Gorilla (Time-Series Compression)
```
Timestamps:
  T1 = 1609459200
  T2 = T1 + 60         → delta = 60
  T3 = T2 + 60         → delta-of-delta = 0 (1 bit!)
  T4 = T3 + 61         → delta-of-delta = 1 (few bits)

Values (XOR encoding):
  V1 = 72.5 (full)
  V2 = 72.3 → XOR(V1, V2) = small number → few bits
```

### Real Systems
Prometheus, Facebook Gorilla, InfluxDB, Parquet (integer columns)

### Summary
Delta encoding stores differences between consecutive values. Combined with delta-of-delta and XOR encoding, it achieves 10x+ compression for time-series data.

---

## 2. Run-Length Encoding (RLE)

### Definition
Replacing consecutive identical values with a single value and a count.

### How It Works
```
Raw:     [USA, USA, USA, USA, UK, UK, USA, USA]
RLE:     [(USA, 4), (UK, 2), (USA, 2)]

Raw:     8 × 3 bytes = 24 bytes
RLE:     3 × (3 + 4) bytes = 21 bytes  (better with longer runs)

Extreme case:
  [USA × 1,000,000]  → [(USA, 1000000)]  = 7 bytes!
```

### When RLE Works Well
```
✓ Sorted columnar data (country column sorted by country)
✓ Bitmap indexes with long runs of 0s or 1s
✓ Images with large uniform regions

✗ Random data (no runs = no compression = overhead!)
✗ High-cardinality columns (many unique values)
```

### Real Systems
Apache Parquet, Amazon Redshift, video encoding (I-frames), BMP/TIFF images

---

## 3. Erasure Coding

### Definition
A storage technique that breaks data into N data chunks + K parity chunks. Any N chunks out of N+K are sufficient to reconstruct the original data.

### Problem It Solves
```
3x Replication:
  Data: 1TB → Store 3 copies = 3TB (200% overhead)
  Can lose 2 copies and survive

Erasure Coding (Reed-Solomon 10+4):
  Data: 1TB → 10 data + 4 parity = 1.4TB (40% overhead)
  Can lose ANY 4 chunks and survive
  Same durability, 60% less storage!
```

### How It Works
```
Original file split into 10 data chunks:
  D1 D2 D3 D4 D5 D6 D7 D8 D9 D10

Compute 4 parity chunks (linear algebra over Galois field):
  P1 P2 P3 P4

Store 14 chunks across 14 different servers.

Lose any 4 servers (e.g., D3, D7, P1, P4 fail):
  Remaining 10 chunks → solve system of equations → reconstruct all data
```

### Tradeoffs

| | 3x Replication | Erasure Coding |
|-|----------------|----------------|
| Storage overhead | 200% | 40-60% |
| Recovery speed | Instant (read copy) | Slow (computation) |
| Read latency | Low (read any copy) | Higher (may need computation) |
| Best for | Hot data, low-latency | Cold/warm data, archival |
| Durability | Same | Same or better |

### Real Systems
Google Colossus (GFS successor), HDFS (EC support), MinIO, Azure Blob, Ceph

---

## 4. Zstandard (Zstd)

### Definition
A compression algorithm by Meta that provides excellent compression ratios at very high speed, replacing older algorithms like gzip.

### Comparison
```
┌─────────────┬───────────┬─────────────┬──────────────┐
│ Algorithm   │ Compress  │ Decompress  │ Ratio        │
├─────────────┼───────────┼─────────────┼──────────────┤
│ gzip        │ 30 MB/s   │ 300 MB/s    │ 3.0x         │
│ Snappy      │ 500 MB/s  │ 1500 MB/s   │ 1.8x         │
│ LZ4         │ 750 MB/s  │ 3500 MB/s   │ 2.1x         │
│ Zstd        │ 400 MB/s  │ 1200 MB/s   │ 3.1x         │
│ Zstd (fast) │ 600 MB/s  │ 1200 MB/s   │ 2.5x         │
└─────────────┴───────────┴─────────────┴──────────────┘
```

### Key Innovation: Dictionary Compression
```
Regular Zstd: Each file compressed independently.
Dictionary:   Train on sample data → shared dictionary.
              Small files (1KB logs) compress much better
              because the dictionary contains common patterns.
```

### When to Use What
```
Need speed over ratio:     LZ4 or Snappy (Kafka default)
Need balance:              Zstd (best overall)
Need max compression:      Zstd level 19+ or gzip -9
Legacy compatibility:      gzip
```

### Real Systems
Linux kernel, Kafka, Hadoop, HTTP (Accept-Encoding: zstd), pkg managers (apt)

---

## 5. Vectorized Execution

### Definition
Processing data in batches (vectors/columns of ~1000 values) instead of row-by-row, enabling CPU SIMD instructions and cache-friendly access.

### Row-at-a-Time vs Vectorized
```
ROW-AT-A-TIME (Volcano model):
  for each row:
    filter(row)  → project(row)  → aggregate(row)
  Function call overhead per row. Branch mispredictions.
  ~1 million rows/sec per core.

VECTORIZED:
  filter(column_batch[1000])    ← SIMD: 8 comparisons per instruction
  project(column_batch[1000])   ← tight loop, CPU prefetch works
  aggregate(column_batch[1000]) ← no function call overhead per row
  ~1 billion rows/sec per core!
```

### Why It's Faster
```
1. SIMD:     Process 4-16 values per CPU instruction
2. Cache:    Column data is contiguous → L1/L2 cache efficient
3. No calls: Tight loops instead of virtual function dispatch per row
4. Branch:   Predication instead of branching (no misprediction)
```

### Real Systems
ClickHouse, DuckDB, Velox (Meta), DataFusion (Apache Arrow), Snowflake

### Summary
Vectorized execution processes data in column batches instead of row-by-row. Combined with SIMD, it achieves 100-1000x speedup for analytical queries.

---

## 6. Roaring Bitmaps

### Definition
A compressed bitmap data structure that adaptively uses different container types for sparse vs dense data, enabling extremely fast set operations.

### Problem with Standard Bitmaps
```
Tracking which users (out of 1 billion) clicked a button:
  Standard bitmap: 1 billion bits = 125 MB
  If only 1000 users clicked: 125 MB for 1000 bits → wasteful!

  Standard array: 1000 × 4 bytes = 4 KB
  But set intersection/union on arrays is slow.
```

### How Roaring Bitmaps Work
```
Split 32-bit integer space into chunks of 2^16 = 65536.
High 16 bits = chunk ID, Low 16 bits = position within chunk.

Each chunk uses the best container:

  Sparse chunk (< 4096 values): Sorted Array
    [12, 45, 99, 512]  ← compact

  Dense chunk (≥ 4096 values): Bitmap
    [1001110100...0110]  ← 8KB bitmap

  Run chunk (consecutive runs): RLE
    [(100, 500), (1000, 2000)]  ← start, length pairs
```

### Set Operations
```
AND (Intersection):
  Bitmap ∩ Bitmap:   bitwise AND (CPU-native, SIMD-able)
  Array ∩ Array:     galloping/binary search
  Bitmap ∩ Array:    check each array element in bitmap O(n)

Union, XOR, etc. all have optimized implementations per container type pair.
```

### Complexity
- Space: Adaptive (optimal for any density)
- AND/OR: O(min(n, m)) to O(n + m) depending on containers
- Contains: O(1) for bitmap containers, O(log n) for array containers

### Real Systems
Apache Druid, Elasticsearch, Apache Lucene, Apache Spark, Pilosa

### Summary
Roaring bitmaps adaptively use arrays (sparse), bitmaps (dense), or RLE (runs) per chunk, achieving both compact storage and blazing-fast set operations.

---

## 7. Trie (Prefix Tree)

### Definition
A tree data structure where each node represents a character. Paths from root to nodes form keys. Enables O(L) lookup where L is the key length.

### How It Works
```
Insert: "app", "apple", "api", "bat"

        (root)
       /      \
      a        b
     / \        \
    p   -        a
   / \            \
  p   i            t
  |
  l
  |
  e

Search "api":
  root → a → p → i → FOUND (3 steps regardless of data size)
  
Prefix search "ap":
  root → a → p → returns ["app", "apple", "api"]
```

### Complexity
- Insert: O(L) where L = key length
- Lookup: O(L)
- Prefix search: O(L + K) where K = number of matches
- Space: O(ALPHABET_SIZE × N × L) — can be large!

### Variants
- **Compressed/Radix Trie**: Merge single-child chains ("a-p-p" → one node)
- **Ternary Search Trie**: 3 children per node (less memory)
- **PATRICIA Trie**: Bit-level trie (used in IP routing)

### Real Systems
Google Autocomplete, IP routing tables, spell checkers, Ethereum (Merkle Patricia Trie)

---

## 8. Cuckoo Hashing

### Definition
A hashing scheme using two hash functions and two tables that guarantees worst-case O(1) lookup by displacing existing entries on collision.

### How It Works
```
Two tables T1, T2 with hash functions h1, h2.

Insert key K:
  1. Try T1[h1(K)]. If empty → insert. Done.
  2. If occupied by K': Evict K'. Place K in T1[h1(K)].
  3. Try T2[h2(K')]. If empty → insert K'. Done.
  4. If occupied: evict again, repeat.
  5. If cycle detected → rehash with new functions.

Lookup key K:
  Check T1[h1(K)] and T2[h2(K)]. One of them has it. O(1).
```

### Visual
```
Insert "cat":
  h1("cat")=2, h2("cat")=5

  T1: [_][_][cat][_][_]     ← placed at index 2
  T2: [_][_][_][_][_][_]

Insert "dog" where h1("dog")=2 (collision!):
  T1: [_][_][dog][_][_]     ← "dog" kicks out "cat"
  T2: [_][_][_][_][_][cat]  ← "cat" goes to T2[h2("cat")=5]
```

### Tradeoffs

| Pros | Cons |
|------|------|
| O(1) worst-case lookup | Insert can cascade (eviction chain) |
| Cache-friendly (2 memory accesses max) | Needs rehash if load factor > ~50% |
| Deterministic performance | More complex than linear probing |

### Real Systems
High-frequency trading, network packet processing, MemC3 (improved Memcached), CPU TLB

### Summary
Cuckoo hashing uses two hash functions and displaces existing entries on collision. Guarantees O(1) worst-case lookup with at most 2 memory accesses — ideal for latency-critical systems.
