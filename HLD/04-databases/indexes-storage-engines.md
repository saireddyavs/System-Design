# Database Indexes & Storage Engines

## Staff+ Engineer Deep Dive for FAANG Interviews

---

## 1. Concept Overview

### Definition
An **index** is a data structure that improves the speed of data retrieval operations at the cost of additional storage and write overhead. **Storage engines** are the underlying components that manage how data is physically stored, retrieved, and updated on disk or in memory.

### Purpose
- **Indexes**: Reduce full table scans from O(n) to O(log n) or O(1); enable efficient range queries, sorting, and uniqueness enforcement
- **Storage engines**: Abstract physical storage; different engines optimize for different workloads (OLTP vs OLAP, read vs write heavy)

### Why It Exists
Without indexes, finding a row requires scanning every row (O(n)). With millions of rows, this is impractical. Indexes provide a "roadmap" to data. Storage engines exist because one size doesn't fit all: some workloads need durability, others need raw speed; some need B-trees, others LSM-trees.

### Problems Solved
| Problem | Solution |
|---------|----------|
| Slow lookups | B-tree index: O(log n) |
| Full table scan | Index scan |
| Slow range queries | B+ tree sorted order |
| Write amplification | LSM-tree (batch writes) |
| Different workloads | Pluggable storage engines |

---

## 2. Real-World Motivation

### Facebook (MySQL/InnoDB)
- InnoDB B+ tree indexes for user lookups, feed queries
- Billions of rows; index enables sub-ms lookups
- Covering indexes for hot query paths

### Google (Bigtable/LevelDB)
- LSM-tree based; optimized for write-heavy workloads
- Gmail, Search indexing: massive write throughput
- Compaction strategies tuned for SSD

### MongoDB (WiredTiger)
- B+ tree default; LSM option for write-heavy
- Document model: indexes on nested fields, arrays
- WiredTiger: document-level locking, compression

### Netflix (Cassandra)
- SSTable + Memtable (LSM); no secondary indexes on hot path
- Partition key is primary access path
- Secondary indexes: local to partition (avoid scatter)

### LinkedIn (Voldemort)
- Key-value; hash index for O(1) lookup
- Read-heavy; simple index structure

### Uber (MySQL + custom)
- InnoDB for trip data; composite indexes for (rider_id, created_at)
- Time-series access patterns

---

## 3. Architecture Diagrams

### B+ Tree Structure
```
┌─────────────────────────────────────────────────────────────────────────┐
│                         B+ TREE INDEX                                    │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│                    ┌─────────────────────┐                              │
│                    │   ROOT (internal)   │                              │
│                    │  [5] [20] [35] [50] │  ← Separator keys            │
│                    └─────────┬───────────┘                              │
│           ┌──────────────────┼──────────────────┐                      │
│           ▼                  ▼                   ▼                      │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                  │
│  │  INTERNAL   │    │  INTERNAL   │    │  INTERNAL   │                  │
│  │ [1][3][5]   │    │ [20][30][35]│    │ [50][60]    │                  │
│  └──────┬──────┘    └──────┬──────┘    └──────┬──────┘                  │
│         │                  │                  │                         │
│         ▼                  ▼                  ▼                          │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                  │
│  │   LEAF      │───▶│   LEAF      │───▶│   LEAF      │  ← Linked list   │
│  │ 1,2,3,4,5   │    │ 20,21..35   │    │ 50,51..     │    for range scan │
│  │ [ptr][ptr]  │    │ [ptr][ptr]  │    │ [ptr][ptr]  │                  │
│  └─────────────┘    └─────────────┘    └─────────────┘                  │
│                                                                          │
│  Properties: All data in leaves; leaves linked; O(log n) height          │
│  Typical: 3-4 levels for millions of rows (branching factor ~100-200)   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### LSM Tree Structure
```
┌─────────────────────────────────────────────────────────────────────────┐
│                         LSM TREE                                         │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   WRITES (incoming)                                                      │
│        │                                                                 │
│        ▼                                                                 │
│   ┌─────────────┐                                                        │
│   │  MEMTABLE   │  ← In-memory, sorted (e.g., skip list)                 │
│   │  (mutable)  │    Fast writes; flushed when full                      │
│   └──────┬──────┘                                                        │
│          │ Flush (immutable)                                              │
│          ▼                                                               │
│   ┌─────────────┐     ┌─────────────┐     ┌─────────────┐                │
│   │  SSTable 1  │     │  SSTable 2  │     │  SSTable 3  │                │
│   │  (L0)       │     │  (L1)       │     │  (L2)       │                │
│   └─────────────┘     └─────────────┘     └─────────────┘                │
│          │                    │                    │                    │
│          └────────────────────┼────────────────────┘                    │
│                               │                                          │
│                          COMPACTION                                      │
│                    Merge sorted runs; remove overwrites                  │
│                                                                          │
│   READS: Check memtable → L0 → L1 → L2 (Bloom filter avoids disk)        │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Hash Index
```
┌─────────────────────────────────────────────────────────────────────────┐
│                         HASH INDEX                                       │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   Key "user:123"  ──▶  hash("user:123") % N  ──▶  Bucket 3               │
│                                                                          │
│   ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐                        │
│   │ Bkt 0│  │ Bkt 1│  │ Bkt 2│  │ Bkt 3│  │ Bkt 4│  ...                  │
│   │      │  │      │  │      │  │user: │  │      │                        │
│   │      │  │      │  │      │  │ 123  │  │      │                        │
│   │      │  │      │  │      │  │ ────▶│  │      │                        │
│   └──────┘  └──────┘  └──────┘  └──────┘  └──────┘                        │
│                                                                          │
│   O(1) lookup; NO range queries; collision handling (chaining/open addr) │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Write Amplification Comparison
```
┌─────────────────────────────────────────────────────────────────────────┐
│                    WRITE AMPLIFICATION                                    │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   B-Tree: 1 write → 1 page update (maybe more if split)                  │
│   Typical: 2-4x (page split, WAL)                                        │
│                                                                          │
│   LSM (Size-Tiered): 1 write → memtable → SSTable → compaction           │
│   Typical: 10-30x (multiple compaction levels)                          │
│                                                                          │
│   LSM (Leveled): More controlled; 10-15x typical                          │
│                                                                          │
│   SSD wear: High write amplification = shorter SSD life                 │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Core Mechanics

### B-Tree / B+ Tree Mechanics
- **Structure**: Balanced tree; all leaves at same depth
- **Branching factor**: Typically 100-200 (fill factor ~70%)
- **Search**: Start at root; compare key; descend; O(log n) = O(log_100(1M)) ≈ 3-4 levels
- **Insert**: Find leaf; insert; split if overflow (median up)
- **Delete**: Find; remove; merge/redistribute if underflow
- **Range scan**: Traverse leaf linked list

### LSM Tree Mechanics
- **Memtable**: In-memory sorted structure (skip list, red-black tree)
- **Flush**: When memtable reaches threshold (e.g., 64MB), write to SSTable
- **SSTable**: Immutable, sorted by key; block-indexed for binary search
- **Compaction**: Merge SSTables; keep newest value for key
  - **Size-tiered**: Merge same-size runs
  - **Leveled**: L0 → L1 → L2; each level 10x previous; no overlap within level

### Hash Index Mechanics
- **Hash function**: Distribute keys across buckets
- **Collision**: Chaining (linked list) or open addressing
- **No order**: Range queries impossible
- **Use case**: Exact match only (e.g., cache lookup)

### Covering Index
- Index contains all columns needed by query
- No need to access base table ("index-only scan")
- Reduces I/O significantly

### Composite Index
- Index on (A, B, C): sorted by A, then B, then C
- Left-prefix rule: (A), (A,B), (A,B,C) usable; (B) or (B,C) not
- Order matters for range: (created_at, user_id) vs (user_id, created_at)

---

## 5. Numbers

| Metric | B-Tree | LSM Tree | Hash |
|--------|--------|----------|------|
| Read (point) | O(log n) ~3-4 seeks | O(k) k=levels | O(1) |
| Read (range) | O(log n + m) | O(n) worst | N/A |
| Write | O(log n) | O(1) amortized | O(1) |
| Write amplification | 2-4x | 10-30x | ~1x |
| Space overhead | 20-50% | 10-20% (compression) | Depends |

### Real Numbers
- **B-tree**: 1M rows, 100-byte rows → ~3-4 disk seeks for point lookup
- **LSM**: RocksDB ~500K writes/sec on SSD; compaction can cause latency spikes
- **Bloom filter**: 1% false positive ≈ 10 bits per element; saves disk reads
- **InnoDB page**: 16KB; ~100-200 keys per internal node

---

## 6. Tradeoffs

### B-Tree vs LSM
| Aspect | B-Tree | LSM |
|--------|--------|-----|
| Read latency | Predictable, low | Can be higher (multiple SSTables) |
| Write latency | Higher (random I/O) | Lower (sequential writes) |
| Write throughput | Moderate | Very high |
| Space | More (fragmentation) | Less (compression) |
| Compaction | None | Background; can cause spikes |

### Index Tradeoffs
| Add Index | Benefit | Cost |
|-----------|---------|------|
| Faster reads | Yes | Slower writes, more storage |
| Covering index | Index-only scan | Larger index |
| Composite | Multi-column queries | Order sensitivity |

---

## 7. Variants / Implementations

### Index Types
- **B-tree / B+ tree**: PostgreSQL, MySQL InnoDB, MongoDB WiredTiger
- **LSM**: RocksDB, LevelDB, Cassandra, HBase
- **Hash**: MySQL MEMORY, Redis (in-memory)
- **Full-text**: Inverted index; Elasticsearch, PostgreSQL GIN
- **Partial**: Index subset of rows (WHERE clause)
- **Bitmap**: For low-cardinality columns

### Storage Engines
| Engine | DB | Structure | Best For |
|--------|-----|-----------|----------|
| InnoDB | MySQL | B+ tree | OLTP, ACID |
| WiredTiger | MongoDB | B+ tree | Documents, compression |
| RocksDB | Many | LSM | Write-heavy, embedded |
| LevelDB | Chrome, etc | LSM | Simple key-value |
| MyRocks | MySQL | LSM | Write-heavy MySQL |

---

## 8. Scaling Strategies

- **Partitioning**: Reduce index size per partition
- **Read replicas**: Offload index scans
- **Covering indexes**: Reduce I/O
- **Partial indexes**: Index only hot data
- **Async indexing**: Elasticsearch; eventual consistency

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|------------|
| Index corruption | Queries fail or wrong results | Rebuild index; checksums |
| Compaction lag (LSM) | Read amplification | Tune compaction; more resources |
| Too many indexes | Slow writes | Remove unused; consolidate |
| Wrong index chosen | Slow query | ANALYZE; hints; query rewrite |

---

## 10. Performance Considerations

- **Selectivity**: High cardinality columns better for indexes
- **Index size**: Smaller = more cache hits
- **Write cost**: Each index updated on write
- **Lock contention**: B-tree page locks; LSM lock-free reads
- **Fragmentation**: B-tree can fragment; VACUUM/OPTIMIZE

---

## 11. Use Cases

| Use Case | Index Type | Engine |
|----------|------------|--------|
| Primary key lookup | B-tree / Hash | Any |
| Range query (time, ID) | B+ tree | InnoDB, WiredTiger |
| Full-text search | Inverted | Elasticsearch, GIN |
| Write-heavy log | LSM | RocksDB, Cassandra |
| Cache | Hash | Redis |

---

## 12. Comparison Tables

### Index Type Comparison
| Type | Lookup | Range | Insert | Space |
|------|--------|-------|--------|-------|
| B-tree | O(log n) | Yes | O(log n) | Medium |
| Hash | O(1) | No | O(1) | Low |
| LSM | O(log n) | Yes | O(1)* | Low |
| Full-text | O(1)* | No | O(1)* | High |

*Amortized or approximate

### Storage Engine Comparison
| Engine | Read | Write | Durability | Transactions |
|-------|------|-------|------------|---------------|
| InnoDB | Fast | Medium | WAL | Full ACID |
| WiredTiger | Fast | Fast | WAL | Document-level |
| RocksDB | Good | Very fast | WAL | Optional |
| LevelDB | Good | Fast | WAL | No |

---

## 13. Code or Pseudocode

### B-Tree Search
```python
def btree_search(node, key):
    if node.is_leaf():
        return binary_search(node.keys, key)
    i = find_child_index(node.keys, key)
    return btree_search(node.children[i], key)
```

### LSM Memtable Flush
```python
def flush_memtable(memtable):
    sstable = sorted(memtable.items())  # Already sorted
    write_sstable(sstable)
    memtable.clear()
    add_to_level0(sstable)
```

### Composite Index Usage
```sql
-- Index: (user_id, created_at)
-- Uses index: WHERE user_id = 1
-- Uses index: WHERE user_id = 1 AND created_at > '2024-01-01'
-- Does NOT use: WHERE created_at > '2024-01-01'
CREATE INDEX idx_user_orders ON orders(user_id, created_at);
```

### Covering Index
```sql
-- Query: SELECT user_id, COUNT(*) FROM orders WHERE status = 'paid' GROUP BY user_id
-- Covering index: (status, user_id) includes both columns
CREATE INDEX idx_orders_status_user ON orders(status, user_id);
-- Index-only scan; no table access
```

---

## 14. Interview Discussion

### Key Points
1. **B+ tree**: All data in leaves; leaves linked for range; 3-4 levels for millions of rows
2. **LSM**: Write-optimized; sequential writes; compaction is the cost
3. **Composite index order**: (A, B) supports A and (A,B); not B alone
4. **Covering index**: Avoid table lookup; index-only scan
5. **Write amplification**: LSM trades read for write; B-tree balanced

### Common Questions
- **Q**: "Why B+ tree over B-tree?"
  - **A**: Data only in leaves; internal nodes have more keys (higher fanout); range scans efficient via leaf links
- **Q**: "When would you use LSM over B-tree?"
  - **A**: Write-heavy (logs, metrics, events); SSD (sequential writes); can tolerate read variability
- **Q**: "What's write amplification?"
  - **A**: One logical write causes multiple physical writes; LSM has high amplification due to compaction
- **Q**: "How does a composite index work?"
  - **A**: Sorted by first column, then second; left-prefix rule for usability

---

## 15. Compaction Strategies (LSM) Deep Dive

### Size-Tiered Compaction
- Merge SSTables of similar size
- Level 0: 4x 10MB → 1x 40MB
- Level 1: 4x 40MB → 1x 160MB
- **Pro**: Simple; **Con**: Space amplification; read amplification (many files to check)

### Leveled Compaction
- L0: No size limit; can overlap
- L1, L2, ...: 10x size of previous; no overlap within level
- **Pro**: Bounded read amplification; **Con**: More write amplification
- Used by RocksDB, Cassandra (LeveledCompactionStrategy)

### Hybrid
- L0: Size-tiered; L1+: Leveled
- Balance between write and read amplification

---

## 16. Index Maintenance and Fragmentation

### B-Tree Fragmentation
- Deletes leave holes; inserts cause page splits
- **VACUUM** (PostgreSQL): Reclaim space; doesn't compact
- **VACUUM FULL**: Rewrites table; reclaims all space; locks table
- **OPTIMIZE TABLE** (MySQL): Rebuilds table
- **REINDEX**: Rebuild index from scratch

### When to Rebuild
- Fragmentation > 30%
- After bulk deletes
- After major schema change
- Monitor: pg_stat_user_tables.n_dead_tup (PostgreSQL)

---

## 17. Partial and Expression Indexes

### Partial Index (PostgreSQL)
```sql
-- Index only rows where status = 'active'
CREATE INDEX idx_active_orders ON orders(user_id) WHERE status = 'active';
-- Smaller index; faster for filtered queries
```

### Expression Index
```sql
-- Index on lower(email) for case-insensitive lookup
CREATE INDEX idx_users_email_lower ON users(LOWER(email));
```

### Use Cases
- Hot/cold data separation (index only recent)
- Conditional uniqueness
- Function-based lookups

---

## 18. Benchmark Numbers (Reference)

| Operation | InnoDB (1M rows) | RocksDB (1M rows) |
|-----------|------------------|-------------------|
| Point read | ~0.1ms | ~0.05ms |
| Range scan (10K rows) | ~10ms | ~5ms |
| Random write | ~0.2ms | ~0.01ms |
| Sequential write | ~0.05ms | ~0.005ms |
| *Numbers approximate; hardware dependent |

---

## 19. Full-Text Index Internals

### Inverted Index
- Term → List of (document_id, position)
- "hello" → [(doc1, 3), (doc2, 1), (doc3, 5)]
- Query "hello world": Intersect postings lists
- Ranking: TF-IDF, BM25

### Elasticsearch/Lucene
- Segments: Immutable; merged in background
- FST (Finite State Transducer): Compact term dictionary
- Skip lists: Fast intersection of postings

---

## 20. Index Selection Heuristics

1. **Equality on column**: B-tree or hash
2. **Range on column**: B-tree (hash cannot)
3. **Multiple columns in WHERE**: Composite index; order by selectivity
4. **ORDER BY column**: Index supports sort; avoid filesort
5. **Covering**: Include all SELECT columns in index
6. **Low cardinality**: Consider bitmap or skip index

---

## 21. Summary: Choosing Index and Engine

| Workload | Index | Engine |
|----------|-------|--------|
| OLTP, mixed read/write | B-tree | InnoDB, WiredTiger |
| Write-heavy, append | LSM | RocksDB, Cassandra |
| Read-heavy, point lookup | Hash | Redis, Memcached |
| Full-text search | Inverted | Elasticsearch, GIN |
| Analytics, scan | Columnar | ClickHouse, Parquet |

### Key Takeaway
- **B-tree**: General purpose; predictable; use by default
- **LSM**: When writes dominate; accept read variability
- **Covering index**: Eliminate table access for hot queries
