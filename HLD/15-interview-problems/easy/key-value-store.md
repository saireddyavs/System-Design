# System Design: Distributed Key-Value Store

## 1. Problem Statement & Requirements

### Functional Requirements

- **Put**: Store key-value pair; overwrite if key exists
- **Get**: Retrieve value by key; return null if not found
- **Delete**: Remove key from store
- **Tunable Consistency**: Support strong consistency (read-your-writes) or eventual consistency (better availability)
- **Partition Tolerance**: System continues operating despite network partitions (CAP theorem)

### Non-Functional Requirements

- **High Availability**: 99.99% uptime; tolerate node failures
- **Scalability**: Horizontal scaling; support billions of keys
- **Low Latency**: p99 < 10ms for get/put
- **Durability**: No data loss on node failure (replication)
- **Partition Tolerance**: Must tolerate network splits (CAP: choose CP or AP; typically AP with tunable C)

### Out of Scope

- Secondary indexes
- Range queries (scan by key prefix only)
- Transactions (multi-key ACID)
- Full-text search
- Complex data types (just bytes)

---

## 2. Back-of-Envelope Estimation

### Traffic Estimates

| Metric | Value | Calculation |
|--------|-------|-------------|
| Total QPS | 1M | Given |
| Read QPS | 700K | 70% reads |
| Write QPS | 300K | 30% writes |
| Per node (100 nodes) | 10K QPS | 1M / 100 |

### Storage Estimates

- **Assumptions**:
  - 10B keys total
  - Avg key: 50 bytes, avg value: 1 KB
  - Total: 10B × 1.05 KB ≈ 10 TB
  - With replication (3x): 30 TB
  - Per node (100 nodes): ~300 GB

### Bandwidth Estimates

- **Read**: 700K × 1 KB ≈ 700 MB/s
- **Write**: 300K × 1 KB × 3 (replication) ≈ 900 MB/s
- **Total**: ~1.6 GB/s cluster-wide

---

## 3. API Design

### REST API Endpoints

#### Put

```
PUT /v1/keys/{key}
```

**Request Headers:**
```
Content-Type: application/octet-stream
X-Consistency: strong | eventual
```

**Request Body:** Binary value

**Response (200 OK):**
```json
{
  "key": "user:123:profile",
  "version": "v5",
  "timestamp": 1710000000
}
```

#### Get

```
GET /v1/keys/{key}
```

**Query Params:**
```
?consistency=strong|eventual
```

**Response (200 OK):**
```
Content-Type: application/octet-stream
X-Version: v5
X-Timestamp: 1710000000

<binary value>
```

**Response (404 Not Found):**
```json
{
  "error": "KEY_NOT_FOUND",
  "message": "Key does not exist"
}
```

#### Delete

```
DELETE /v1/keys/{key}
```

**Response (200 OK):**
```json
{
  "key": "user:123:profile",
  "deleted": true
}
```

---

## 4. Data Model / Database Schema

### Key-Value Structure

```
Key:   bytes (max 256 bytes)
Value: bytes (max 1 MB)
Version: vector clock or timestamp (for conflict resolution)
```

### Storage Format (LSM-Tree Style)

**In-Memory (MemTable)**:
```
key -> (value, timestamp, tombstone?)
```

**On-Disk (SSTable - Sorted String Table)**:
- Immutable, sorted by key
- Blocks with index for binary search
- Bloom filter for "key not present" fast path

### Replication Metadata

```
Key: partition_id
Replicas: [node_1, node_2, node_3]  // N=3
Version: vector_clock
```

---

## 5. High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              CLIENT REQUEST                                       │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│   CLIENT / COORDINATOR                                                           │
│   - Partition routing (consistent hashing)                                       │
│   - Replica selection                                                            │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
                    ┌───────────────────┼───────────────────┐
                    ▼                   ▼                   ▼
┌───────────────────────┐   ┌───────────────────────┐   ┌───────────────────────┐
│   Node 1 (Replica)    │   │   Node 2 (Replica)    │   │   Node 3 (Replica)    │
│   Partition A         │   │   Partition A          │   │   Partition A          │
│   - MemTable           │   │   - MemTable           │   │   - MemTable           │
│   - SSTables           │   │   - SSTables           │   │   - SSTables          │
└───────────────────────┘   └───────────────────────┘   └───────────────────────┘
        │                               │                               │
        └───────────────────────────────┼───────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│   GOSSIP PROTOCOL (Failure detection, cluster membership)                       │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 6. Detailed Component Design

### 6.1 Partitioning: Consistent Hashing

- **Hash ring**: 0 to 2^64
- **Partition key**: hash(key) → position on ring
- **Virtual nodes**: Each physical node has V virtual nodes (e.g., 256)
- **Replica placement**: Next N nodes clockwise on ring

**Key placement**: Key K → partition P (first node clockwise) → replicas on P, P+1, P+2

### 6.2 Replication: Quorum

**N = replication factor** (e.g., 3)
**W = write quorum** (number of nodes that must acknowledge write)
**R = read quorum** (number of nodes to read from)

**Consistency rule**: `W + R > N` ensures read sees latest write

| Config | W | R | N | Consistency | Availability |
|--------|---|---|---|-------------|--------------|
| Strong | 3 | 3 | 3 | Strong | Low (all must be up) |
| Balanced | 2 | 2 | 3 | Strong | Medium |
| Eventual | 1 | 1 | 3 | Eventual | High |

**Write path**: Send to all N replicas; wait for W acks
**Read path**: Send to R replicas; return latest (by version)

### 6.3 Conflict Resolution

**Scenarios**: Multiple replicas have different values (e.g., partition + concurrent writes)

**Options**:

1. **Last-Write-Wins (LWW)**: Use timestamp; highest wins. Simple but can lose data.
2. **Vector Clocks**: Each node tracks (node_id, counter); merge on conflict.
3. **Version Vectors**: Similar to vector clocks; used in Dynamo.

**Vector Clock Example**:
```
Node A: {A:3, B:1, C:2}
Node B: {A:2, B:2, C:2}
```

If A's clock dominates B (all counters >= B's): A wins.
If concurrent (neither dominates): Conflict; return both; client resolves.

**Practical**: Use **LWW** for most systems; **vector clocks** for advanced.

### 6.4 Write Path (LSM-Tree)

1. **Write arrives** at coordinator
2. **Route** to partition (consistent hash)
3. **Send to N replicas** (or W for quorum)
4. **Each replica**:
   - Append to **Write-Ahead Log (WAL)** for durability
   - Insert into **MemTable** (in-memory sorted structure)
   - When MemTable full (~64 MB): **Flush** to SSTable on disk
   - SSTable is immutable

**WAL**: Sequential writes; fast; replay on crash to recover MemTable

### 6.5 Read Path

1. **Route** to partition
2. **Query R replicas** (or all for strong consistency)
3. **Each replica**:
   - Check **Bloom filter** (if key not in SSTables, skip disk)
   - Search **MemTable** first
   - Search **SSTables** (newest to oldest; merge)
   - Return (value, version)
4. **Coordinator**: Merge results; resolve conflicts (LWW or vector clock)
5. **Return** latest value

**Bloom filter**: Probabilistic; if "not present" → definitely not in SSTable; if "present" → maybe (need to check)

### 6.6 Compaction

**Problem**: SSTables accumulate; read checks many files; slow.

**Solution**: **Compaction** — merge SSTables into fewer, larger files.

- **Leveled compaction**: Level 0 (newest) → Level 1 → ... → Level N
- **Size-tiered**: Merge small SSTables into larger ones
- **Output**: New SSTable; delete old SSTables

### 6.7 Failure Detection: Gossip Protocol

- **Gossip**: Each node periodically sends state to random peers
- **State**: (node_id, heartbeat, generation)
- **Failure**: If no heartbeat from node for T seconds, mark dead
- **Dissemination**: Failure info propagates via gossip

**Usage**: Avoid routing to dead nodes; trigger re-replication

### 6.8 Hinted Handoff

**Scenario**: Write to replica A, B, C. B is temporarily down.

**Hinted handoff**: 
- Write to A and C (quorum may still be met)
- Store "hint" on A: "when B is back, send this write to B"
- When B recovers, A sends hinted writes to B
- B applies; B is now consistent

**Benefit**: Faster recovery; no need to read from quorum to rebuild B

### 6.9 Read Repair

**Scenario**: Read from R replicas; get different values (e.g., one replica missed a write).

**Read repair**:
- Coordinator gets R responses
- If all same → return
- If different → pick latest (LWW or vector clock)
- **Async**: Send repair to stale replicas (update them with latest)

**Benefit**: Eventually consistent; no separate anti-entropy needed for frequently read keys

### 6.10 Anti-Entropy: Merkle Trees

**Problem**: Replicas can drift (e.g., missed writes, corruption).

**Merkle tree**:
- Each leaf = hash of key range
- Parent = hash(children)
- Root = hash of entire partition

**Process**:
- Compare Merkle roots between replicas
- If different, drill down to find differing ranges
- Sync those ranges

**Efficient**: O(log n) to find differences; used in Cassandra

---

## 7. Scaling

### Horizontal Scaling

- **Add node**: Add to hash ring; some partitions move from existing nodes
- **Remove node**: Remove from ring; partitions move to neighbors
- **Consistent hashing**: Minimizes data movement

### Sharding Strategy

- **Partition key**: hash(key) % num_partitions
- **Replication**: Each partition on N nodes
- **Rebalancing**: When adding node, move ~1/N of data

### Handling Hot Partitions

- **Problem**: One partition (e.g., celebrity profile) gets too much traffic
- **Solutions**:
  - Replicate hot partition to more nodes (read from any)
  - Split key (e.g., `user:123` → `user:123:shard0`, `user:123:shard1`) — application logic

### Caching

- **Read-through cache**: Cache layer in front (e.g., Redis)
- **Cache invalidation**: On write, invalidate cache; or short TTL

---

## 8. Failure Handling

| Component | Failure | Mitigation |
|-----------|---------|------------|
| Node | Down | Replicas serve; hinted handoff; re-replicate from others |
| Network partition | Split | Each partition serves; eventual consistency; possible conflicts |
| Disk | Full | Compaction; alert; add capacity |
| WAL | Corrupt | Replicate from other nodes; lose recent writes on that node |

### Redundancy

- **Replication**: N=3 (or more)
- **Multi-AZ**: Replicas in different availability zones
- **WAL**: Replicated or on durable storage

### Recovery

- **Node restart**: Replay WAL; rebuild MemTable
- **New node**: Stream data from replicas; join ring
- **Data loss**: Replicate from replicas; read repair; anti-entropy

---

## 9. Monitoring & Observability

### Key Metrics

| Metric | Target | Alert |
|--------|--------|-------|
| Get latency | p99 < 10ms | Critical |
| Put latency | p99 < 20ms | Warning |
| Error rate | < 0.1% | Critical |
| Compaction lag | Low | Warning if backlog grows |
| Replication lag | 0 | Warning if replicas behind |
| Disk usage | < 80% | Critical |

### Dashboards

- QPS (read/write)
- Latency percentiles
- Replication factor
- Compaction status
- Node health (gossip)

### Alerting

- Node down
- Replication lag
- High latency
- Disk full

---

## 10. Interview Tips

### Common Follow-Up Questions

1. **Why quorum W+R>N?** Ensures read overlaps with write; at least one replica has latest
2. **Vector clock vs LWW?** LWW simple; lose concurrent updates. Vector clock preserves; complex merge
3. **Why LSM-tree?** Write-optimized; sequential writes; good for write-heavy
4. **Hinted handoff vs read repair?** Hinted handoff for writes to temporarily down node; read repair for stale reads
5. **How to handle split brain?** Last-write-wins or vector clocks; client may need to resolve

### What Interviewers Look For

- CAP understanding (partition tolerance required)
- Quorum (W+R>N)
- Consistency trade-offs
- LSM-tree basics
- Failure handling (hinted handoff, read repair, anti-entropy)

### Common Mistakes

- Ignoring partition tolerance
- Not explaining quorum
- Confusing replication with sharding
- Forgetting compaction

---

## Appendix: Additional Design Considerations

### A. LSM-Tree Write Path

```
Write → WAL (append) → MemTable (insert)
                            │
                            │ when full (e.g. 64MB)
                            ▼
                      Flush to SSTable (immutable)
                            │
                            │ compaction
                            ▼
                      Merge SSTables
```

### B. LSM-Tree Read Path

```
Read → Bloom filter (per SSTable) → if not present, skip
                │
                │ if present
                ▼
        MemTable → SSTables (newest to oldest)
                │
                ▼
        Return first found (or merge if multiple versions)
```

### C. Vector Clock Merge

```
Merge(A, B):
  For each node_id: merged[node_id] = max(A[node_id], B[node_id])
  Return merged
```

### D. Quorum Examples

```
N=3, W=2, R=2:
  Write: 2 of 3 replicas ack
  Read: 2 of 3 replicas queried
  Overlap: At least 1 replica has both read and write → strong consistency

N=5, W=3, R=3:
  Same logic; more fault tolerance
```

### E. Dynamo-Style Architecture Summary

| Component | Dynamo/Cassandra |
|-----------|------------------|
| Partitioning | Consistent hashing |
| Replication | N replicas, quorum |
| Consistency | Tunable (W, R) |
| Conflict | Vector clocks, LWW |
| Failure detection | Gossip |
| Write path | MemTable → SSTable |
| Read path | Bloom filter → MemTable → SSTable |
| Compaction | Leveled / Size-tiered |
| Recovery | Hinted handoff, read repair, Merkle |

### F. CAP Theorem

- **C**onsistency: All nodes see same data
- **A**vailability: Every request gets response
- **P**artition tolerance: System works despite network partitions

**Reality**: In distributed systems, partitions happen. Must choose **CP** (consistency over availability) or **AP** (availability over consistency). Dynamo-style: **AP** with tunable consistency (quorum for read-your-writes).

### G. Bloom Filter False Positive Rate

- **Formula**: `(1 - e^(-kn/m))^k` where k=hash functions, n=items, m=bits
- **Typical**: 1% false positive with ~10 bits per element
- **Benefit**: Avoid 99% of disk reads for non-existent keys

### H. WAL (Write-Ahead Log) Design

- **Append-only**: Sequential writes; fast on HDD/SSD
- **Format**: (key, value, timestamp, op_type)
- **Recovery**: Replay from beginning; rebuild MemTable
- **Segmentation**: Rotate WAL when size exceeds threshold; old segments deleted after MemTable flush

### I. Compaction Strategies

**Leveled (LevelDB, RocksDB)**:
- Level 0: SSTables from MemTable flushes (overlapping)
- Level 1+: Non-overlapping; each key in one file
- Compaction: Merge Level L into Level L+1

**Size-tiered (Cassandra)**:
- Merge small SSTables into larger ones
- Simpler; more write amplification

### J. DynamoDB vs Cassandra vs Our Design

| Aspect | DynamoDB | Cassandra | Our Design |
|--------|----------|-----------|------------|
| Model | Managed, serverless | Self-hosted | Self-hosted |
| Consistency | Tunable (strong/eventual) | Tunable | Tunable |
| Partition key | Required | Required | hash(key) |
| Replication | Multi-AZ | RF per DC | N replicas |
| Conflict | LWW | Timestamp | LWW or vector clock |

### K. Write Path Pseudocode

```python
def put(key, value):
    partition = consistent_hash(key) % num_partitions
    replicas = get_replicas(partition)
    version = vector_clock.increment(local_node_id)
    acks = 0
    for node in replicas:
        node.append_wal((key, value, version))
        node.insert_memtable(key, value, version)
        acks += 1
        if acks >= W:
            break
    return acks >= W
```

### L. Read Path with Conflict Resolution

```python
def get(key):
    replicas = get_replicas(partition(key))
    responses = [node.get(key) for node in random_sample(replicas, R)]
    responses = [r for r in responses if r]
    if not responses:
        return None
    latest = max(responses, key=lambda r: r.version)
    for r in responses:
        if r != latest:
            async_repair(r.node, key, latest)
    return latest
```

### M. Hinted Handoff Flow

```
1. Write to replica B fails (B is down)
2. Replica A stores: hint = {key, value, target=B}
3. Replica A continues normal operation
4. When B recovers, A sends hinted writes to B
5. B applies; hints deleted from A
6. B is now consistent with A and C
```

### N. Merkle Tree for Anti-Entropy

```
Partition range: [a, m)
  Root = hash(L1 + L2)
  L1 = hash([a,g))  L2 = hash([g,m))
  [a,g) = hash([a,d)) + hash([d,g))
  ...
  Leaves = hash of key ranges in SSTable
```

Compare roots; if different, recurse to find differing ranges; sync.

### O. Complete Interview Walkthrough (45 min)

**0-5 min**: Clarify: put/get/delete, consistency, partition tolerance, scale.
**5-10 min**: Estimates. 1M QPS, 10B keys, 10 TB. Replication factor.
**10-15 min**: API. Simple key-value. Consistency parameter.
**15-25 min**: Partitioning (consistent hashing). Replication (quorum W+R>N).
**25-35 min**: Write path (WAL, MemTable, SSTable). Read path (bloom, MemTable, SSTable).
**35-40 min**: Conflict resolution. Hinted handoff. Read repair. Merkle trees.
**40-45 min**: CAP. Dynamo-style AP. Compaction. Failure handling.

### P. Quick Reference Cheat Sheet

| Topic | Key Points |
|-------|------------|
| Partitioning | Consistent hashing; N replicas per partition |
| Quorum | W+R>N for strong consistency |
| Write path | WAL → MemTable → SSTable (flush) |
| Read path | Bloom filter → MemTable → SSTables |
| Conflict | LWW or vector clocks |
| Recovery | Hinted handoff, read repair, Merkle anti-entropy |

### Q. Further Reading & Real-World Examples

- **Dynamo (Amazon)**: Original paper; inspired DynamoDB, Cassandra, Riak
- **Cassandra**: Wide-column; tunable consistency; gossip
- **Riak**: Erlang; CRDTs; eventually consistent
- **ScyllaDB**: C++; Cassandra-compatible; lower latency

### R. Design Alternatives Considered

| Decision | Alternative | Why Rejected |
|----------|-------------|--------------|
| LSM-tree | B-tree | LSM write-optimized; B-tree read-optimized |
| Quorum | Single replica | Single = no fault tolerance |
| Vector clock | LWW | LWW simpler; vector clock for complex |
| Hinted handoff | Full sync | Hinted handoff faster recovery |

### S. Quorum Math Examples

- N=3, W=2, R=2: Any 2 writes overlap with any 2 reads → 1 common node has latest
- N=5, W=3, R=3: 3+3-5=1 overlap minimum
- N=3, W=1, R=1: No overlap → eventual consistency only

### T. SSTable File Format

```
[Data Block 1][Data Block 2]...[Data Block N][Index Block][Footer]
- Data Block: (key, value) pairs, sorted by key, compressed
- Index Block: (key, offset) for binary search
- Footer: Offset to index, magic number
```

### U. Bloom Filter Tuning

- **False positive rate** = (1 - e^(-kn/m))^k
- **Optimal k** = (m/n) * ln(2) ≈ 0.7 * (m/n)
- **Example**: 1M keys, 1% FP → m ≈ 10M bits, k ≈ 7

### V. Node Addition/Removal Impact

- **Add node**: ~1/N of keys move to new node (consistent hashing)
- **Remove node**: Keys from removed node go to next node on ring
- **Virtual nodes**: Smoother distribution; 100-200 per physical node typical

### W. Summary

A Dynamo-style KV store: consistent hashing, quorum replication (W+R>N), LSM-tree (WAL, MemTable, SSTable), bloom filters, compaction, hinted handoff, read repair, Merkle trees for anti-entropy.

---
*End of Key-Value Store System Design Document*

This document covers a Dynamo-style distributed key-value store. Key takeaways: quorum, LSM-tree, bloom filters, hinted handoff, read repair, and Merkle trees for anti-entropy. Understand the CAP theorem and why partition tolerance is non-negotiable. In practice, distributed systems choose AP with tunable consistency.

**Document Version**: 1.0 | **Last Updated**: 2025-03-10 | **Target**: System Design Interview (Easy)
