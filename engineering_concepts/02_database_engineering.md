# Module 2: Database Engineering

---

## 1. Sharding (Range vs Hash)

### Definition
Splitting a database horizontally across multiple machines so each holds a subset of the data.

### Problem It Solves
A single database cannot store petabytes or handle millions of QPS. Sharding distributes both storage and load.

### Strategies

```
┌──────────── SHARDING STRATEGIES ────────────┐
│                                              │
│  RANGE-BASED          HASH-BASED             │
│  key: A-M → Shard 1   hash(key) % N → Shard │
│  key: N-Z → Shard 2                          │
│                                              │
│  DIRECTORY-BASED                             │
│  Lookup table maps key → shard               │
└──────────────────────────────────────────────┘
```

### Comparison

| | Range Sharding | Hash Sharding |
|-|---|---|
| Range queries | Efficient (adjacent keys on same shard) | Scatter-gather (slow) |
| Hotspots | Likely (popular ranges) | Unlikely (uniform distribution) |
| Rebalancing | Move range boundary | Consistent hashing helps |
| Example | DynamoDB (sort key), Bigtable | Cassandra, MongoDB |

### Instagram's Approach
- Pre-created thousands of logical shards mapped to fewer physical servers
- Future growth: split logical shards to new machines without data movement
- Shard key: `user_id` (all photos of a user on one shard)

### Edge Cases
- **Hotspots**: Celebrity user on one shard overloads it
- **Cross-shard queries**: JOINs across shards are expensive
- **Resharding**: Adding shards requires data migration

### Interview Tip
"I'd shard by user_id using hash-based sharding with consistent hashing. For range queries, I'd add a secondary index or use range-based sharding on the sort key."

### Summary
Sharding splits a database across machines. Hash sharding distributes evenly; range sharding enables efficient range queries. Pre-sharding (Instagram) avoids future migration pain.

---

## 2. Snowflake IDs

### Definition
A 64-bit distributed ID format: timestamp + machine ID + sequence number. Generates unique, time-sortable IDs without coordination.

### Problem It Solves
Auto-increment needs a single leader (bottleneck). UUIDs aren't sortable and are 128-bit (waste space in indexes).

### Format
```
 64 bits total:
 ┌─────────────────┬──────────┬──────────────┐
 │ 41 bits         │ 10 bits  │ 12 bits      │
 │ Timestamp (ms)  │ Machine  │ Sequence     │
 │ ~69 years       │ 1024     │ 4096/ms      │
 │                 │ machines │ per machine  │
 └─────────────────┴──────────┴──────────────┘
```

### Properties
- **4096 IDs/ms/machine** → 4M IDs/sec per machine
- **K-sortable**: IDs are roughly time-ordered (great for B-Tree indexes)
- **No coordination**: Each machine generates independently

### Comparison: ID Schemes

| | Snowflake | UUID v4 | ULID | Auto-increment |
|-|-----------|---------|------|----------------|
| Size | 64-bit | 128-bit | 128-bit | 32/64-bit |
| Sortable | Yes (time) | No | Yes (time) | Yes |
| Coordination | None | None | None | Single leader |
| Uniqueness | Machine+seq | Random | Random+time | Sequential |
| Index perf | Excellent | Poor (random) | Good | Excellent |

### Real Systems
Twitter, Discord, Instagram (modified), Baidu (uid-generator)

### Summary
Snowflake IDs encode time + machine + sequence in 64 bits, enabling distributed generation of sortable unique IDs at millions per second.

---

## 3. UUID / ULID

### UUID (Universally Unique Identifier)
```
Format: 550e8400-e29b-41d4-a716-446655440000  (128-bit, hex)

v1: Timestamp + MAC address (privacy concern)
v4: Random (most common) — 2^122 possible values
v7: Time-ordered + random (new, recommended)
```

### ULID (Universally Unique Lexicographically Sortable Identifier)
```
Format: 01ARZ3NDEKTSV4RRFFQ69G5FAV  (128-bit, Crockford Base32)

┌──────────────┬──────────────────┐
│ 48 bits      │ 80 bits          │
│ Timestamp    │ Randomness       │
│ (ms since    │                  │
│  epoch)      │                  │
└──────────────┴──────────────────┘
```

### When to Use What
- **Snowflake**: High-throughput systems needing 64-bit IDs (Twitter, Discord)
- **UUID v4**: Simple, no coordination needed, storage space not critical
- **UUID v7 / ULID**: Need time-sortability + 128-bit uniqueness
- **Auto-increment**: Only for single-leader setups

### Summary
UUIDs provide universally unique 128-bit IDs. ULIDs add time-sortability. UUID v4 is random (bad for indexes), UUID v7 and ULID are time-ordered (good for indexes).

---

## 4. Write-Ahead Logging (WAL)

### Definition
A durability technique where changes are appended to a log file on disk BEFORE being applied to the actual data pages.

### Problem It Solves
Database crashes after updating memory but before writing to disk → data loss. WAL ensures recovery by replaying the log.

### How It Works
```
1. Transaction: UPDATE balance SET amount=50 WHERE id=1
2. FIRST: Append to WAL on disk (sequential write — fast)
3. THEN: Update in-memory data page
4. LATER: Background checkpoint writes dirty pages to disk
5. CRASH? → Replay WAL from last checkpoint on startup
```

### Visual
```
  Transaction
      │
      ▼
  ┌────────┐   sequential    ┌────────────┐
  │  WAL   │ ───────────────→│ WAL on Disk│
  │ Buffer │   write (fast)  └────────────┘
  └───┬────┘                       │
      │                            │ on crash: replay
      ▼                            ▼
  ┌────────┐                 ┌────────────┐
  │ Memory │ ── checkpoint ─→│ Data Files │
  │ Pages  │   (background)  └────────────┘
  └────────┘
```

### Why Sequential Writes?
```
Random write:  Head seeks to position → 10ms per write
Sequential:    Head stays in place   → 0.01ms per write
               1000x faster on HDD, 10x on SSD
```

### Real Systems
PostgreSQL, MySQL (InnoDB), Kafka (the log IS the database), SQLite, etcd

### Summary
WAL writes changes to a sequential log before applying them. On crash, replay the log. This guarantees durability while maintaining high write throughput.

---

## 5. Checkpointing

### Definition
Periodically writing all in-memory dirty pages to disk, creating a known-good recovery point so the WAL can be truncated.

### Problem It Solves
Without checkpoints, crash recovery must replay the ENTIRE WAL from the beginning of time — impractical for long-running databases.

### How It Works
```
WAL: [entry1][entry2][entry3][entry4][entry5][entry6]
                              ▲
                         CHECKPOINT
                    (all data up to here is on disk)

On crash: only replay entry4, entry5, entry6
Previous WAL entries can be deleted.
```

### Types
- **Sharp checkpoint**: Pause all writes, flush everything (simple but causes stall)
- **Fuzzy checkpoint**: Write dirty pages in background while accepting new writes (Postgres)
- **Incremental checkpoint**: Only write pages modified since last checkpoint

### Real Systems
PostgreSQL (checkpoint process), Redis (RDB snapshots), Flink (distributed checkpoints via Chandy-Lamport), Spark (RDD checkpointing)

### Summary
Checkpoints flush dirty pages to disk, creating a recovery point. They allow WAL truncation and bound crash recovery time.

---

## 6. B-Trees / B+ Trees

### Definition
A self-balancing tree where each node contains many keys and has many children, keeping the tree shallow for minimal disk seeks.

### Problem It Solves
Binary trees have O(log₂ N) depth — too many disk reads. B-Trees with branching factor 100+ have depth 3-4 for billions of keys.

### B-Tree vs B+ Tree

```
B-Tree: Data in ALL nodes          B+ Tree: Data ONLY in leaves
┌───────────────┐                  ┌───────────────┐
│  [10, 20]     │ ← has data      │  [10, 20]     │ ← keys only
├───┬───┬───────┤                  ├───┬───┬───────┤
│<10│10-20│>20  │                  │<10│10-20│>20  │
└───┴───┴───────┘                  └─┬─┴──┬──┴──┬──┘
                                     ▼    ▼     ▼
                                   ┌──┐ ┌──┐ ┌──┐
                                   │d1│→│d2│→│d3│ ← data + linked
                                   └──┘ └──┘ └──┘   (range scans!)
```

### Why B+ Trees Win for Databases
- **Range queries**: Leaves are linked → sequential scan
- **Higher fanout**: Internal nodes are smaller (no data) → more keys per node
- **Predictable I/O**: Always exactly `depth` disk reads per lookup

### Complexity
```
Branching factor b = ~500 (4KB page / 8-byte keys)
1 billion keys:  depth = log₅₀₀(10⁹) ≈ 3-4 levels

Lookup:  O(log_b N) disk reads → 3-4 reads for 1B keys
Insert:  O(log_b N) + potential page split
Scan:    O(K) for K consecutive keys (sequential)
```

### Comparison: B+ Tree vs LSM Tree

| | B+ Tree | LSM Tree |
|-|---------|----------|
| Read | O(log N), single lookup | May check multiple levels |
| Write | Random I/O (page updates) | Sequential I/O (append) |
| Use case | Read-heavy OLTP | Write-heavy workloads |
| Examples | MySQL, Postgres | RocksDB, Cassandra |

### Real Systems
MySQL (InnoDB), PostgreSQL, SQLite, Oracle, MongoDB (WiredTiger)

### Summary
B+ Trees store keys in a wide, shallow tree with data only in linked leaf nodes. They provide O(log N) lookups with ~3 disk reads for billions of keys and excellent range scan performance.

---

## 7. SSTables (Sorted String Tables)

### Definition
Immutable, sorted key-value files on disk. The on-disk component of LSM Trees.

### How It Works
```
┌─────────────────────────────────┐
│ SSTable File                    │
│ ┌──────┬──────┬──────┬──────┐  │
│ │ Data │ Data │ Data │ Data │  │  ← Sorted key-value blocks
│ │Block1│Block2│Block3│Block4│  │
│ └──────┴──────┴──────┴──────┘  │
│ ┌──────────────────────────┐   │
│ │     Index Block          │   │  ← Sparse index (key→offset)
│ └──────────────────────────┘   │
│ ┌──────────────────────────┐   │
│ │     Bloom Filter         │   │  ← "Is key maybe here?"
│ └──────────────────────────┘   │
│ ┌──────────────────────────┐   │
│ │     Footer/Metadata      │   │
│ └──────────────────────────┘   │
└─────────────────────────────────┘
```

### Properties
- **Sorted**: Binary search or index-based lookup
- **Immutable**: Never modified after creation (simplifies concurrency)
- **Mergeable**: Two sorted files merge in O(N) time (merge sort)

### Real Systems
Google BigTable, LevelDB, RocksDB, Apache HBase

### Summary
SSTables are sorted, immutable on-disk files that serve as the building block of LSM trees. Their sorted nature enables efficient lookups and merges.

---

## 8. Inverted Index

### Definition
A data structure that maps content (words) to their locations (document IDs), enabling fast full-text search.

### How It Works
```
Documents:
  Doc1: "the quick brown fox"
  Doc2: "the lazy dog"
  Doc3: "quick fox jumps"

Inverted Index:
  "the"   → [Doc1, Doc2]
  "quick" → [Doc1, Doc3]
  "brown" → [Doc1]
  "fox"   → [Doc1, Doc3]
  "lazy"  → [Doc2]
  "dog"   → [Doc2]
  "jumps" → [Doc3]

Query "quick fox" → intersect([Doc1,Doc3], [Doc1,Doc3]) = [Doc1, Doc3]
```

### Enhancements
- **TF-IDF scoring**: Rank by term frequency × inverse doc frequency
- **Positional index**: Store positions for phrase queries ("quick fox" adjacent)
- **Skip lists**: Speed up posting list intersection

### Complexity
- Build: O(N × L) where N=docs, L=avg length
- Query: O(K) where K=posting list length
- Boolean AND: O(min(K₁, K₂)) with sorted lists

### Real Systems
Elasticsearch, Apache Lucene/Solr, Google Search, PostgreSQL (GIN index)

### Summary
Inverted indexes map words to document IDs, enabling O(1) keyword lookup. They are the foundation of every search engine.

---

## 9. Geohashing / Quadtrees / H3

### Geohashing
Converts 2D coordinates into a 1D string. Points sharing a prefix are nearby.

```
World divided into grids, each grid gets a character:
  (37.7749, -122.4194) → "9q8yy"

Precision levels:
  "9"      → ~5000km cell
  "9q"     → ~1250km
  "9q8"    → ~150km
  "9q8y"   → ~40km
  "9q8yy"  → ~5km

Nearby search: SELECT * WHERE geohash LIKE '9q8y%'
```

### Quadtree
```
Recursively divide space into 4 quadrants:
┌─────┬─────┐
│ NW  │ NE  │
│     │  •  │ ← point in NE
├─────┼─────┤
│ SW  │ SE  │
│  •  │     │ ← point in SW
└─────┴─────┘

Subdivide only cells that contain points.
Adaptive: dense areas get more subdivision.
```

### H3 (Uber's Hexagonal Indexing)
```
Why hexagons over squares?
┌──┐┌──┐     ╱╲╱╲╱╲
│  ││  │    ╱  ╲  ╱  ╲
└──┘└──┘   ╲  ╱╲  ╱╲  ╱
┌──┐┌──┐    ╲╱  ╲╱  ╲╱
│  ││  │
└──┘└──┘
Squares: diagonal neighbors are √2 farther
Hexagons: ALL 6 neighbors equidistant → uniform
```

### Comparison

| | Geohash | Quadtree | H3 |
|-|---------|----------|----|
| Type | String prefix | Tree structure | Hierarchical hex |
| Neighbor distance | Non-uniform | Varies | Uniform |
| DB-friendly | Yes (string prefix query) | Needs custom index | Needs H3 library |
| Use case | Simple proximity | Adaptive density | Ride-sharing, pricing |
| Company | Yelp, Tinder | Game engines | Uber |

### Summary
Geohashing converts coordinates to sortable strings. Quadtrees adaptively subdivide space. H3 uses hexagons for uniform-distance neighbors. Choose based on uniformity needs and query patterns.

---

## 10. Replication

### Definition
Maintaining copies of data on multiple machines for fault tolerance and read scalability.

### Strategies
```
┌─── SINGLE-LEADER ───┐    ┌─── MULTI-LEADER ───┐    ┌─── LEADERLESS ───┐
│ Client → Leader      │    │ Client → Any Leader │    │ Client → Any Node│
│ Leader → Followers   │    │ Leaders sync        │    │ Quorum R+W > N   │
│                      │    │                     │    │                  │
│ Postgres, MySQL      │    │ CouchDB, Galera     │    │ Dynamo, Cassandra│
└──────────────────────┘    └─────────────────────┘    └──────────────────┘
```

### Sync vs Async Replication
```
Synchronous:
  Client → Leader → waits for Follower ACK → responds to Client
  ✓ Strong consistency   ✗ Higher latency

Asynchronous:
  Client → Leader → responds immediately → replicates later
  ✓ Low latency   ✗ Data loss if leader crashes before replication
```

### Tradeoffs

| | Single-Leader | Multi-Leader | Leaderless |
|-|---|---|---|
| Consistency | Strong (sync) or eventual | Eventual | Tunable (quorum) |
| Write throughput | Limited by leader | Higher (multi-writer) | Highest |
| Conflict handling | No conflicts | Must resolve | Must resolve |
| Failover | Leader election needed | Automatic | N/A |

### Summary
Replication copies data across nodes for durability and read scaling. Single-leader is simplest, leaderless is most available, and multi-leader supports multi-datacenter writes.

---

## 11. Columnar Storage vs Row Storage

### Row Storage
```
Table: Employees
Row 1: [1, "Alice", "Eng",  95000]
Row 2: [2, "Bob",   "Mktg", 80000]
Row 3: [3, "Carol", "Eng",  92000]

Stored sequentially: [1,Alice,Eng,95000][2,Bob,Mktg,80000]...
```

### Columnar Storage
```
Column "id":     [1, 2, 3]
Column "name":   [Alice, Bob, Carol]
Column "dept":   [Eng, Mktg, Eng]
Column "salary": [95000, 80000, 92000]

Each column stored in its own file/block.
```

### Why Columnar is Faster for Analytics
```
Query: SELECT AVG(salary) FROM employees

Row store:  Load ALL columns for every row → wasted I/O
            [1,Alice,Eng,95000][2,Bob,Mktg,80000]...
            ^^^^^^^^^^^^^^^^^ unnecessary data loaded

Columnar:   Load ONLY salary column → minimal I/O
            [95000, 80000, 92000]
```

### Compression Benefits
```
Column "dept": [Eng, Eng, Eng, Eng, Mktg, Mktg, Eng, Eng]

Dictionary encoding: Eng=0, Mktg=1 → [0,0,0,0,1,1,0,0]
Run-length encoding: [0×4, 1×2, 0×2]

Same-type data compresses 10x better than mixed-type rows.
```

### Comparison

| | Row Store | Columnar Store |
|-|-----------|---------------|
| Best for | OLTP (single row lookups) | OLAP (aggregations) |
| Write | Fast (append a row) | Slow (update multiple columns) |
| Read single row | Fast (one seek) | Slow (read from N column files) |
| Read single column | Slow (load all columns) | Fast (one file) |
| Compression | Low | Very high |
| Examples | MySQL, Postgres | Redshift, Parquet, ClickHouse |

### Summary
Row stores excel at transactional workloads (read/write whole rows). Columnar stores excel at analytics (read few columns across many rows) with 10-100x better compression and query performance.
