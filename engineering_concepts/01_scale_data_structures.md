# Module 1: Scale, Data Structures & Consistency

---

## 1. Operational Transformation (OT)

### Definition
An algorithm that enables real-time collaborative editing by transforming concurrent operations so they converge to the same state regardless of execution order.

### Problem It Solves
User A types "Hello" at position 0. User B simultaneously deletes character at position 0. Without coordination, documents diverge permanently.

### How It Works
```
1. Each client generates operations locally (Insert, Delete)
2. Operations are sent to a central server
3. Server transforms incoming ops against already-executed ops
4. T(op1, op2) adjusts indices so both apply correctly
```

### Visual
```
  User A: Insert('X', pos=0)     User B: Delete(pos=0)
       \                           /
        \                         /
         ▼                       ▼
       ┌─────── SERVER ─────────┐
       │ Transform(InsA, DelB)  │
       │ InsA' = Insert('X', 1) │  ← adjusted index
       │ DelB' = Delete(0)      │
       └────────────────────────┘
         Both arrive at same doc
```

### Core Algorithm
```
function transform(op1, op2):
    if op1.type == INSERT and op2.type == DELETE:
        if op1.pos <= op2.pos:
            return (op1, Delete(op2.pos + 1))
        else:
            return (Insert(op1.pos - 1, op1.char), op2)
```

### Tradeoffs

| Pros | Cons |
|------|------|
| Battle-tested (Google Docs since 2006) | Requires central server |
| Low bandwidth (send ops, not full doc) | Complex to implement correctly |
| Real-time feel | Correctness proofs are notoriously hard |

### Comparison: OT vs CRDTs

| | OT | CRDTs |
|-|----|----|
| Server | Required (central authority) | Optional (peer-to-peer) |
| Offline | Limited | Full support |
| Complexity | Algorithm complexity | Data structure complexity |
| Memory | Low | Higher (unique IDs per character) |

### Real Systems
Google Docs, Apache Wave, ShareDB, Microsoft Office Online

### Interview Tip
Say: "OT transforms non-commutative operations into commutative ones by adjusting indices based on causally preceding operations."

### Summary
OT enables real-time collaboration by transforming concurrent edits through a central server so all clients converge to identical state. It's the backbone of Google Docs.

---

## 2. CRDTs (Conflict-Free Replicated Data Types)

### Definition
Data structures that can be modified independently on different replicas and merged automatically without conflicts, guaranteeing eventual convergence.

### Problem It Solves
OT needs a central server. What if you need offline-first editing, or peer-to-peer collaboration without any coordinator?

### Types
```
┌─────────────── CRDTs ───────────────┐
│                                      │
│  State-based (CvRDT)                │
│  ├── G-Counter (grow-only counter)   │
│  ├── PN-Counter (add/subtract)       │
│  ├── G-Set (grow-only set)           │
│  └── OR-Set (observed-remove set)    │
│                                      │
│  Operation-based (CmRDT)            │
│  └── Requires causal delivery        │
│                                      │
│  Sequence CRDTs                      │
│  ├── RGA (Replicated Growable Array) │
│  └── Fractional Indexing (Figma)     │
└──────────────────────────────────────┘
```

### How It Works (G-Counter Example)
```
Node A: [A:3, B:0, C:0]   value = 3
Node B: [A:0, B:5, C:0]   value = 5
Node C: [A:0, B:0, C:2]   value = 2

Merge = element-wise max:
Result: [A:3, B:5, C:2]   value = 10
```

### Sequence CRDT (Figma's Approach)
```
Character positions use fractional indices:
  'H' = 0.1,  'e' = 0.2,  'l' = 0.3

Insert 'X' between 'H' and 'e':
  'X' = 0.15  ← always room between any two fractions

No conflicts possible — every insert has a unique position.
```

### Tradeoffs

| Pros | Cons |
|------|------|
| No central server needed | Higher memory (metadata per element) |
| Works offline/P2P | Tombstones accumulate (deleted items linger) |
| Mathematically proven to converge | Limited operation set per type |

### Real Systems
Figma, Apple Notes, Redis CRDTs, Riak, Automerge, Yjs

### Interview Tip
"CRDTs guarantee convergence by making the merge function commutative, associative, and idempotent — the mathematical properties of a join semilattice."

### Summary
CRDTs are data structures where any replica can be modified independently and all replicas always converge upon merge. They enable offline-first and P2P collaboration without a central server.

---

## 3. Gossip Protocol

### Definition
A peer-to-peer protocol where nodes randomly exchange state information, spreading updates like an epidemic through the cluster.

### Problem It Solves
How do 10,000 servers learn cluster membership (who is alive/dead) without a single-point-of-failure central registry?

### How It Works
```
Every T seconds:
  1. Node A picks a RANDOM peer B
  2. A sends its state table to B
  3. B merges A's info with its own
  4. B sends merged state back to A

Information spreads exponentially:
  Round 1: 1 node knows  → 2
  Round 2: 2 nodes know  → 4
  Round 3: 4 nodes know  → 8
  ...
  Round log₂(N): ALL nodes know
```

### Visual
```
  Round 0:    [A*] [B] [C] [D] [E]     (* = has info)
  Round 1:    [A*] [B*] [C] [D] [E]    A tells B
  Round 2:    [A*] [B*] [C*] [D*] [E]  A→C, B→D
  Round 3:    [A*] [B*] [C*] [D*] [E*] full propagation
```

### Variants
- **Push**: Send your state to random peer
- **Pull**: Ask random peer for their state
- **Push-Pull**: Exchange state both ways (fastest)

### Complexity
- Propagation time: **O(log N)** rounds
- Message load per node: **O(1)** per round
- Total messages per round: **O(N)**

### Tradeoffs

| Pros | Cons |
|------|------|
| No single point of failure | Eventually consistent (not instant) |
| Scales to thousands of nodes | Redundant messages (bandwidth) |
| Simple to implement | Convergence time is probabilistic |

### Real Systems
Amazon DynamoDB, Cassandra, Consul (HashiCorp), Uber Ringpop, Redis Cluster

### Interview Tip
"Gossip provides O(log N) convergence with O(1) per-node cost, making it ideal for failure detection and membership in large clusters."

### Summary
Gossip is a decentralized protocol where nodes randomly share state, achieving eventual consistency across large clusters with O(log N) propagation time and no central coordinator.

---

## 4. Consistent Hashing

### Definition
A hashing strategy that maps both keys and servers to a circular ring, minimizing key redistribution when servers join or leave.

### Problem It Solves
With `hash(key) % N`, adding 1 server changes N, remapping nearly ALL keys. This causes massive cache misses (cache stampede).

### How It Works
```
1. Hash servers onto ring:  hash(Server_A) = 90°
                            hash(Server_B) = 210°
                            hash(Server_C) = 330°

2. Hash keys onto ring:     hash("user_42") = 120°

3. Walk clockwise to find owner:
   120° → next server clockwise = Server_B (210°)

4. Add Server_D at 150°:
   Only keys between 90°-150° move. Everything else stays.
```

### Visual
```
            0°
            │
    330° ── C ── 90°
   /                  \
  /     key=120°       \
 /        ↓ walks CW    \
C                    A (90°)
 \                    /
  \    D(150°)  ←NEW /
   \     B(210°)    /
    ─────────────────
          180°

Adding D: only keys in (90°, 150°] move from B → D
```

### Virtual Nodes
Physical servers get multiple hash positions to balance load:
```
Server_A → hash("A-0")=30°, hash("A-1")=120°, hash("A-2")=270°
```

### Complexity
- Lookup: **O(log N)** with sorted ring, O(1) with jump hash
- Add/Remove node: **O(K/N)** keys moved (K=total keys, N=nodes)

### Tradeoffs

| Pros | Cons |
|------|------|
| Minimal disruption on scale events | Non-uniform distribution without vnodes |
| Simple mental model | Hotspots possible with unlucky hashing |
| Works for caching, DB routing, CDNs | Rebalancing still needed for vnodes |

### Real Systems
Discord, Akamai CDN, DynamoDB, Cassandra, Memcached

### Interview Tip
Always mention virtual nodes. Say: "I'd use consistent hashing with virtual nodes to avoid hotspots, similar to DynamoDB's approach."

### Summary
Consistent hashing maps servers and keys to a ring so that adding/removing a server only redistributes O(K/N) keys. Virtual nodes improve balance.

---

## 5. Rendezvous Hashing (Highest Random Weight)

### Definition
Each key computes a weight for every server; the server with the highest weight wins. When a server leaves, only its keys redistribute.

### Problem It Solves
Same as consistent hashing but with simpler implementation and more uniform distribution without virtual nodes.

### How It Works
```
For key K, compute:
  weight(K, Server_i) = hash(K + Server_i)

Assign K to server with MAX weight.

Remove Server_2:
  Recompute only for keys that were on Server_2.
  They go to their 2nd-highest-weight server.
```

### Comparison: Consistent Hashing vs Rendezvous

| | Consistent Hashing | Rendezvous Hashing |
|-|---|---|
| Lookup time | O(log N) | O(N) — must check all servers |
| Distribution | Needs virtual nodes | Naturally uniform |
| Implementation | Ring + binary search | Simple weight comparison |
| Memory | Ring structure | None (stateless) |
| Best for | Large N (1000s of servers) | Small N (10s of servers) |

### Real Systems
GitHub load balancer, Microsoft Azure, some CDN routing

### Summary
Rendezvous hashing selects the server with the highest hash weight for each key. Simpler than consistent hashing but O(N) per lookup, so better for small server counts.

---

## 6. Vector Clocks

### Definition
A data structure (vector of counters, one per node) that tracks causality in distributed systems, allowing detection of concurrent vs. causally ordered events.

### Problem It Solves
Two users update the same shopping cart from different datacenters. Physical timestamps can't determine order (clocks drift). How do we detect conflicts?

### How It Works
```
3 nodes: A, B, C. Each maintains vector [A, B, C].

1. A writes:      A=[1,0,0]
2. A sends to B:  B receives, merges → B=[1,1,0] (B increments its own)
3. C writes:      C=[0,0,1]

Compare A=[1,0,0] vs C=[0,0,1]:
  A[0]>C[0] but A[2]<C[2] → CONCURRENT (conflict!)

Compare A=[1,0,0] vs B=[1,1,0]:
  A ≤ B on all positions → A happened-before B (no conflict)
```

### Ordering Rules
```
V1 < V2  (happened-before) if: all V1[i] ≤ V2[i] AND at least one V1[i] < V2[i]
V1 ∥ V2  (concurrent)      if: neither V1 < V2 nor V2 < V1
```

### Visual
```
Node A:  [1,0,0] ──→ [2,0,0]
              \
Node B:   [1,1,0] ──→ [1,2,0]
                           ↑ concurrent with C
Node C:        [0,0,1] ──→ [0,0,2]
```

### Tradeoffs

| Pros | Cons |
|------|------|
| Detects concurrent writes precisely | Vector grows with number of nodes |
| No need for synchronized clocks | Client must resolve conflicts |
| Foundation for conflict detection | Doesn't scale past ~100s of nodes |

### Real Systems
Amazon Dynamo (original), Riak, Voldemort

### Interview Tip
"Vector clocks detect causality without synchronized clocks. If vectors are incomparable, the writes are concurrent and need application-level resolution."

### Summary
Vector clocks track causal ordering across distributed nodes. Each node maintains a vector of counters. Incomparable vectors indicate concurrent (conflicting) writes.

---

## 7. Lamport Clocks

### Definition
A simple logical clock where each event gets a monotonically increasing counter. If event A happened-before event B, then `L(A) < L(B)`.

### Problem It Solves
Physical clocks drift. Lamport clocks provide a partial ordering of events in distributed systems without requiring synchronized time.

### How It Works
```
Rules:
1. Before each local event: counter++
2. When sending message: attach counter
3. When receiving: counter = max(local, received) + 1
```

### Visual
```
Node A:  1 ──→ 2 ──→ 3 ──────→ 7
                       \        ↑
Node B:       1 ──→ 2  \→ 4 → 5
                         \
Node C:            1 ──→ 4 ──→ 6
```

### Limitation
`L(A) < L(B)` does NOT mean A happened before B. The converse is guaranteed:
- If A → B, then L(A) < L(B) ✓
- If L(A) < L(B), A may or may not have caused B ✗

Vector clocks fix this limitation.

### Comparison

| | Lamport Clock | Vector Clock |
|-|---|---|
| Size | Single integer | N integers |
| Detects causality | One direction only | Both directions |
| Detects concurrency | No | Yes |
| Scalability | Excellent | Limited by N |

### Real Systems
Most distributed systems use Lamport clocks internally for log ordering. Used in TLA+ specifications.

### Summary
Lamport clocks assign a single increasing counter to events. They guarantee ordering for causal events but cannot detect concurrency — vector clocks extend them for that.

---

## 8. Bloom Filters

### Definition
A space-efficient probabilistic data structure that tests set membership. It can definitively say "NOT in set" but only probabilistically "MAYBE in set."

### Problem It Solves
Checking if a row exists in a database costs a disk seek (~10ms). For millions of non-existent keys, this is devastating. A Bloom filter answers "definitely not here" from memory.

### How It Works
```
Setup: Bit array of m bits, k hash functions.

Insert("apple"):
  h1("apple") = 3   → set bit 3
  h2("apple") = 7   → set bit 7
  h3("apple") = 11  → set bit 11

Query("banana"):
  h1("banana") = 3   → bit 3 = 1 ✓
  h2("banana") = 5   → bit 5 = 0 ✗ → DEFINITELY NOT IN SET

Query("cherry"):
  All bits set = 1   → MAYBE in set (could be false positive)
```

### Visual
```
Bit Array:  [0][0][0][1][0][0][0][1][0][0][0][1][0][0][0]
             0   1   2   3   4   5   6   7   8   9  10  11

Insert "apple" → bits 3, 7, 11 set to 1

Query "banana" → checks bits 3, 5, 11
                 bit 5 = 0 → NOT in set (100% certain)
```

### Complexity
- Space: **O(m)** bits (m ≈ -n·ln(p) / (ln2)² where p = false positive rate)
- Insert: **O(k)** — k hash computations
- Query: **O(k)**
- 1% false positive, 1M items → ~1.2 MB

### Tradeoffs

| Pros | Cons |
|------|------|
| Extremely memory-efficient | False positives (never false negatives) |
| O(k) constant-time operations | Cannot delete elements (use Counting BF) |
| No hash collisions to handle | Size must be tuned to expected cardinality |

### Variants
- **Counting Bloom Filter**: Counters instead of bits (supports delete)
- **Cuckoo Filter**: Better for deletion, slightly better space
- **Quotient Filter**: Cache-friendly, supports merge

### Real Systems
Google BigTable, Postgres, Cassandra (SSTable lookups), Chrome (malicious URL check), Medium (article recommendations)

### Interview Tip
"I'd place a Bloom filter in front of the database to avoid disk I/O for keys that definitely don't exist, reducing read latency significantly."

### Summary
Bloom filters answer "is X in the set?" using a bit array and k hash functions. False positives are possible but false negatives never occur. Used everywhere to avoid expensive lookups.

---

## 9. HyperLogLog (HLL)

### Definition
A probabilistic algorithm for cardinality estimation — counting unique elements in a stream using sub-linear space.

### Problem It Solves
Counting "unique visitors" exactly requires storing every visitor ID (e.g., 1 billion IPs = gigabytes). HLL does it in ~12 KB with 99% accuracy.

### How It Works
```
Intuition:
- Hash each element to a binary string
- Count leading zeros of hash
- If you see 5 leading zeros, you've likely seen ~2⁵ = 32 unique items

Algorithm:
1. Hash element → binary
2. Use first p bits to pick a "register" (bucket)
3. Count leading zeros in remaining bits
4. Store MAX leading zeros per register
5. Estimate = harmonic mean across all registers
```

### Visual
```
hash("user_1")  = 00010110...  → register 0, leading zeros = 2
hash("user_2")  = 00001011...  → register 0, leading zeros = 3
hash("user_42") = 01101001...  → register 1, leading zeros = 0

Registers: [3, 0, ...]  ← max leading zeros per bucket

Estimate ≈ α · m² / Σ(2^(-register[i]))
```

### Complexity
- Space: **O(m · log log n)** — typically 12 KB for billions of items
- Standard error: **1.04 / √m** (m = number of registers)
- With m = 2¹⁴ (16384 registers): ~0.81% error

### Tradeoffs

| Pros | Cons |
|------|------|
| Constant memory regardless of cardinality | Cannot list the actual elements |
| Mergeable (union of two HLLs) | Not exact — ~1% error |
| Fast O(1) per insert | Cannot remove elements |

### Real Systems
Redis (`PFADD`, `PFCOUNT`), BigQuery, Presto, Reddit (unique view counts)

### Interview Tip
"For counting unique users across shards, I'd use HyperLogLog in Redis — it's mergeable across nodes and uses only 12KB per counter."

### Summary
HyperLogLog counts unique items in a stream using ~12KB of memory with ~1% error. It exploits the statistical properties of hash leading zeros.

---

## 10. Count-Min Sketch

### Definition
A probabilistic frequency estimation data structure for data streams. Answers "how many times did X appear?" with bounded overcount.

### Problem It Solves
Tracking exact frequency of every element in a high-volume stream (tweets, packets, clicks) requires unbounded memory.

### How It Works
```
Structure: d rows × w columns of counters (d hash functions)

Increment("cat"):
  h1("cat") = 3  → grid[0][3]++
  h2("cat") = 7  → grid[1][7]++
  h3("cat") = 1  → grid[2][1]++

Query("cat"):
  return MIN(grid[0][3], grid[1][7], grid[2][1])

Min reduces overcount from hash collisions.
```

### Visual
```
          col: 0  1  2  3  4  5  6  7
  h1 row0:   [0][0][0][3][0][0][0][0]   ← h1("cat")=3
  h2 row1:   [0][0][0][0][0][0][0][3]   ← h2("cat")=7
  h3 row2:   [0][3][0][0][0][0][0][0]   ← h3("cat")=1

  freq("cat") = min(3, 3, 3) = 3  (exact or slight overcount)
```

### Complexity
- Space: O(w × d) — typically a few KB
- Update: O(d) — d hash computations
- Query: O(d)
- Error: overestimates by at most ε·N with probability 1-δ

### Tradeoffs

| Pros | Cons |
|------|------|
| Bounded memory for infinite streams | Only overestimates, never underestimates |
| Parallelizable (merge by summing grids) | Cannot enumerate stored items |
| Sub-linear space | Accuracy degrades without tuning w, d |

### Real Systems
Twitter (trending topics), Cloudflare (DDoS detection), Google Analytics, network monitoring

### Summary
Count-Min Sketch estimates element frequency in a stream using a 2D array of counters and multiple hash functions. Takes the minimum across rows to reduce overcounting.

---

## 11. Merkle Trees

### Definition
A binary tree where every leaf is a hash of a data block and every non-leaf is the hash of its children. Any change in data propagates up, changing the root hash.

### Problem It Solves
How do two nodes verify they have identical data without transferring all of it? Merkle trees let you find differences in O(log N) comparisons.

### How It Works
```
Data Blocks:  [A]    [B]    [C]    [D]

Leaves:      H(A)   H(B)   H(C)   H(D)
               \     /       \     /
Internal:    H(AB)            H(CD)
                 \           /
Root:           H(ABCD)

If root hashes match → data is identical.
If not → descend to find differing subtree.
```

### Anti-Entropy (Data Repair)
```
Node 1 Root: abc123     Node 2 Root: abc123  → Match! Done.

Node 1 Root: abc123     Node 2 Root: xyz789  → Mismatch!
  Compare left children:   Match
  Compare right children:  Mismatch → drill down right
  Find: Block D differs → transfer only Block D
```

### Complexity
- Verification: **O(log N)** hashes to check
- Space: **O(N)** nodes
- Finding differences: **O(log N)** comparisons

### Tradeoffs

| Pros | Cons |
|------|------|
| Efficient diff detection | Tree must be rebuilt on updates |
| Tamper-proof (blockchain) | Memory overhead for hash storage |
| Logarithmic comparison cost | Hash computation has CPU cost |

### Real Systems
Git (content-addressable storage), Bitcoin/Ethereum, Amazon Dynamo (anti-entropy), IPFS, ZFS, Cassandra

### Interview Tip
"For replica synchronization, I'd use Merkle trees to detect and transfer only the data blocks that differ, reducing network overhead from O(N) to O(log N)."

### Summary
Merkle trees hash data hierarchically so that any change produces a different root. They enable efficient data integrity verification and replica synchronization in O(log N).

---

## 12. LSM Trees (Log-Structured Merge Trees)

### Definition
A write-optimized storage structure that buffers writes in memory, flushes them as sorted immutable files, and periodically merges files in the background.

### Problem It Solves
B-Trees require random disk I/O per write (slow on HDD/SSD). LSM trees convert random writes to sequential writes, increasing write throughput 10-100x.

### How It Works
```
1. WRITE → MemTable (in-memory sorted tree, e.g., Red-Black tree)
2. MemTable full → flush to disk as immutable SSTable (Level 0)
3. Background COMPACTION merges SSTables into larger sorted files
4. READ → check MemTable → check L0 SSTables → L1 → L2...
   (Use Bloom filters to skip SSTables that don't have the key)
```

### Visual
```
  Write Path:
  ┌────────────┐
  │  MemTable   │ ← fast in-memory writes
  │ (sorted)    │
  └─────┬──────┘
        │ flush when full
        ▼
  ┌──────────┐  ┌──────────┐  ┌──────────┐
  │ SSTable  │  │ SSTable  │  │ SSTable  │  Level 0
  │  (L0)    │  │  (L0)    │  │  (L0)    │  (may overlap)
  └────┬─────┘  └────┬─────┘  └────┬─────┘
       └──────┬──────┘──────┬──────┘
              │  COMPACTION  │
              ▼              ▼
        ┌──────────────────────┐
        │   SSTable (L1)       │  Level 1 (non-overlapping)
        └──────────┬───────────┘
                   ▼
        ┌──────────────────────┐
        │   SSTable (L2)       │  Level 2 (larger)
        └──────────────────────┘
```

### Compaction Strategies
- **Size-tiered**: Merge similarly-sized SSTables (write-optimized)
- **Leveled**: Each level is 10x larger, non-overlapping (read-optimized)

### Complexity

| Operation | B-Tree | LSM Tree |
|-----------|--------|----------|
| Write | O(log N) random I/O | O(1) amortized sequential |
| Read | O(log N) — 1 lookup | O(log N) — may check multiple levels |
| Space amplification | ~1x | 1.1-2x (compaction overhead) |
| Write amplification | ~2x | 10-30x (repeated compaction) |

### Tradeoffs

| Pros | Cons |
|------|------|
| Extremely fast writes (sequential I/O) | Read amplification (check multiple levels) |
| Great for write-heavy workloads | Write amplification from compaction |
| Compression-friendly (sorted blocks) | Background compaction uses CPU/IO |

### Real Systems
RocksDB (Facebook), LevelDB (Google), Cassandra, HBase, CockroachDB, TiKV

### Interview Tip
"For a write-heavy workload like metrics ingestion, I'd use an LSM-tree engine like RocksDB. For read-heavy OLTP, B-Trees in Postgres are better."

### Summary
LSM Trees buffer writes in memory and flush as sorted immutable files. Background compaction merges files. They trade read performance for dramatically faster writes — the engine behind most modern NoSQL databases.
