# Consistency Models in Distributed Systems

> Staff+ Engineer Level — FAANG Interview Deep Dive

---

## 1. Concept Overview

### Definition

**Consistency models** define the guarantees a distributed system provides about the visibility and ordering of read and write operations across multiple replicas. They answer: *"When a client reads data, what value will it see relative to writes performed by itself or other clients?"*

### Purpose

- **Predictability**: Clients can reason about system behavior without understanding internal replication
- **Correctness**: Applications can rely on invariants (e.g., "balance never goes negative")
- **Performance vs. Correctness tradeoff**: Stronger consistency = more coordination = higher latency

### Problems Solved

| Problem | Solution |
|---------|----------|
| Stale reads | Strong consistency guarantees |
| Lost updates | Linearizability / serializability |
| Conflicting concurrent writes | Causal consistency, CRDTs |
| High latency from coordination | Tunable/eventual consistency |
| Split-brain scenarios | Quorum-based reads/writes |

---

## 2. Real-World Motivation

### Google Spanner

- **Linearizability** via TrueTime (GPS + atomic clocks)
- External consistency: if T1 commits before T2 starts, T1's timestamp < T2's
- Used for: Ads, Gmail, YouTube metadata — where financial correctness matters

### Netflix

- **Eventual consistency** for recommendations, watch history
- Accepts temporary inconsistencies (e.g., "Continue Watching" may lag)
- Strong consistency only for billing/subscription state

### Uber

- **Causal consistency** for ride state (driver location, ETA updates)
- Eventual for non-critical data (driver ratings, heat maps)
- Strong for payment processing

### Amazon DynamoDB

- **Strong consistency** (optional): read-after-write consistency
- **Eventual consistency** (default): lower latency, lower cost
- Tunable per-request: `ConsistentRead=true` doubles read cost

### Twitter/X

- **Eventual consistency** for timeline, follower counts
- Counters may be stale; "like" counts eventually converge
- Strong for DMs, account security

---

## 3. Architecture Diagrams

### Linearizability — Single Logical Copy

```
Client A          Replica 1          Replica 2          Client B
   |                   |                   |                 |
   |-- W(x=1) -------->|                   |                 |
   |                   |---- replicate ---->|                 |
   |<-- ack -----------|                   |                 |
   |                   |                   |<-- R(x) --------|
   |                   |                   |-- return 1 ----->|
   |                   |                   |                 |
   |  (B's read must see A's write; operations appear atomic)
```

### Sequential Consistency — Preserve Program Order

```
Process P1              Process P2              Process P3
   |                         |                         |
   | W(x)=1                   |                         |
   | W(x)=2                   | R(x)=?  (sees 1 or 2)   |
   |                          | R(x)=?  (sees 1 or 2)   |
   |                          | W(x)=3                  |
   | R(x)=?  (must see 3)     |                         |
   |
   All processes agree on SOME total order; each process sees its own order
```

### Causal Consistency — Happens-Before Preserved

```
Client A                    Client B                    Client C
   |                            |                            |
   | W(x=1) ------------------> |                            |
   |                            | W(y=2)  (causally after)   |
   |                            | -------------------------->| R(y)=2
   |                            |                            | R(x)=?
   |                            |                            | (may see 0 or 1)
   |
   C sees y=2 (caused by B) but x may not have propagated yet
```

### Quorum Read/Write (W + R > N)

```
        N=5 Replicas
    ┌───┬───┬───┬───┬───┐
    │ R1│ R2│ R3│ R4│ R5│
    └───┴───┴───┴───┴───┘
         ▲   ▲   ▲
         │   │   │
    Write: W=3 (quorum)     Read: R=3 (quorum)
    W + R = 6 > 5  →  At least one replica has latest write
```

### Eventual Consistency — Convergence

```
    t=0:  R1: x=1   R2: x=1   R3: x=1
    t=1:  W(x=2) → R1
    t=2:  R1: x=2   R2: x=1   R3: x=1   (divergence)
    t=3:  gossip/replication propagates
    t=4:  R1: x=2   R2: x=2   R3: x=2   (converged)
```

---

## 4. Core Mechanics

### Linearizability

- **Definition**: Every operation appears to take effect atomically at some instant between its invocation and response
- **Implication**: There exists a global total order consistent with real-time ordering
- **Implementation**: Single leader + synchronous replication; or TrueTime (Spanner)

### Sequential Consistency

- **Definition**: All processes see the same order of operations; each process's operations appear in its program order
- **Weaker than linearizability**: No real-time constraint — operations can be reordered if not causally related

### Causal Consistency

- **Definition**: If operation A causally precedes B (A happens-before B), every process must see A before B
- **Concurrent operations** may be seen in different orders by different processes
- **Implementation**: Vector clocks, version vectors

### Read-Your-Writes Consistency

- **Definition**: A process always sees its own writes
- **Session-based**: Tied to client session; replicas must include session's writes in read path

### Monotonic Reads

- **Definition**: If a process reads x, subsequent reads never see older values of x
- **Prevents**: Time-travel reads (reading stale data after having seen fresh data)

### Eventual Consistency

- **Definition**: If no new updates, all replicas eventually converge to the same state
- **No guarantee** on when or in what order updates propagate

### Quorum: W + R > N

- **Write quorum W**: Must write to W replicas
- **Read quorum R**: Must read from R replicas
- **Guarantee**: R replicas intersect with W replicas → at least one read sees latest write
- **Common**: W=R=⌈(N+1)/2⌉ (majority)

---

## 5. Numbers

| System | Consistency | Read Latency (p99) | Write Latency (p99) | Replication |
|--------|-------------|-------------------|---------------------|-------------|
| Spanner | Linearizable | ~10-50ms | ~50-100ms | Synchronous, 2-phase |
| DynamoDB Strong | Strong | ~10ms | ~15ms | 3 replicas, quorum |
| DynamoDB Eventual | Eventual | ~5ms | ~10ms | Async replication |
| Cassandra (QUORUM) | Tunable | ~5-15ms | ~10-25ms | Tunable W/R |
| Cassandra (ONE) | Eventual | ~2-5ms | ~5-10ms | Single replica |
| ZooKeeper | Sequential | ~1-5ms | ~5-15ms | ZAB consensus |
| etcd | Linearizable | ~5-20ms | ~10-30ms | Raft |

### Quorum Configurations

| N | W | R | W+R | Fault Tolerance |
|---|---|---|-----|-----------------|
| 3 | 2 | 2 | 4 | 1 node |
| 5 | 3 | 3 | 6 | 2 nodes |
| 7 | 4 | 4 | 8 | 3 nodes |

---

## 6. Tradeoffs

### Consistency vs. Availability (CAP)

| Choice | Consistency | Availability | Example |
|--------|--------------|---------------|---------|
| CP | Strong | Sacrifice under partition | ZooKeeper, etcd |
| AP | Eventual | Always available | Cassandra, DynamoDB (eventual) |
| CA | Strong | Available (no partition) | Single-node DB |

### Consistency vs. Latency

| Model | Typical Latency | Coordination |
|-------|-----------------|--------------|
| Linearizable | High (cross-datacenter sync) | Synchronous replication |
| Causal | Medium | Vector clocks, metadata |
| Eventual | Low | Async, best-effort |

### Comparison Table: Consistency Models

| Model | Stale Reads? | Read-Your-Writes? | Causal Order? | Real-Time Order? | Use Case |
|-------|--------------|-------------------|---------------|------------------|----------|
| Linearizable | No | Yes | Yes | Yes | Financial, ads |
| Sequential | No | Yes | Yes | No | Distributed shared memory |
| Causal | Possible | Yes | Yes | No | Social feeds |
| Read-your-writes | No (own) | Yes | Partial | No | User sessions |
| Monotonic reads | No (monotonic) | No | No | No | Caching |
| Eventual | Yes | No | No | No | Counters, recommendations |

---

## 7. Variants / Implementations

### Tunable Consistency (Cassandra)

```text
ConsistencyLevel: ONE, TWO, THREE, QUORUM, ALL, LOCAL_QUORUM, EACH_QUORUM
```

- **ONE**: Single replica — lowest latency, eventual
- **QUORUM**: Majority — strong for single-DC
- **LOCAL_QUORUM**: Majority within local DC — cross-DC eventual

### Spanner TrueTime

- **Bounded uncertainty**: Each timestamp has [earliest, latest] window
- **Commit wait**: Transaction waits until TT.after(commit_timestamp) is true
- **External consistency**: Global ordering without central timestamp oracle

### DynamoDB

- **StronglyConsistentRead**: Read from leader; 2x cost
- **EventuallyConsistentRead**: Read from any replica

### ZooKeeper

- **Sequential consistency**: All clients see same order of updates
- **Single leader**: All writes go through leader; reads can be local

---

## 8. Scaling Strategies

1. **Sharding**: Partition data; consistency within shard
2. **Read replicas**: Strong consistency from primary; eventual from replicas
3. **Multi-region**: Accept eventual for cross-region; use CRDTs for conflict resolution
4. **Caching**: Cache with TTL; accept staleness for hot data
5. **Async replication**: Write to primary; replicate asynchronously

---

## 9. Failure Scenarios

| Scenario | Linearizable | Eventual | Quorum |
|----------|--------------|----------|--------|
| Network partition | Blocks or fails | Continues | Blocks if < quorum |
| Node failure | Blocks if < quorum | Continues | Blocks if < quorum |
| Clock skew | Breaks TrueTime | No impact | No impact |
| Split-brain | Prevented | Possible | Prevented (quorum) |

### Split-Brain with Eventual Consistency

- Two partitions both accept writes
- On merge: conflict resolution (last-writer-wins, CRDTs, manual)

---

## 10. Performance Considerations

- **Strong consistency**: Cross-datacenter round-trips; commit protocols (2PC, Paxos)
- **Quorum**: Latency = slowest of quorum replicas
- **Eventual**: No coordination; lowest latency
- **Caching**: Invalidate on write; or accept staleness

---

## 11. Use Cases

| Use Case | Recommended Model | Reason |
|----------|-------------------|--------|
| Bank balance | Linearizable | No double-spend |
| Shopping cart | Eventual or causal | Merge conflicts OK |
| Leader election | Strong (CP) | Single leader |
| Social feed | Causal or eventual | Order matters, not real-time |
| Counters (likes) | Eventual | Approximate OK |
| Session data | Read-your-writes | User sees own actions |

---

## 12. Comparison Tables

### Consistency Model Hierarchy (Strong → Weak)

```
Linearizability
    └── Sequential Consistency
            └── Causal Consistency
                    └── Read-Your-Writes
                            └── Monotonic Reads
                                    └── Eventual Consistency
```

### System-by-System

| System | Primary Model | When to Use |
|--------|---------------|-------------|
| Spanner | Linearizable | Global transactions, financial |
| DynamoDB | Eventual/Strong | Tunable per request |
| Cassandra | Tunable | High write throughput, flexible |
| ZooKeeper | Sequential | Coordination, config |
| etcd | Linearizable | Kubernetes, service discovery |
| Redis Cluster | Eventual (replication) | Caching, sessions |

---

## 13. Code or Pseudocode

### Quorum Read

```python
def quorum_read(key, N, R):
    replicas = get_replicas(key, N)
    responses = []
    for replica in random.sample(replicas, R):
        responses.append(replica.read(key))
    # Return latest by version/timestamp
    return max(responses, key=lambda r: r.version)
```

### Quorum Write

```python
def quorum_write(key, value, N, W):
    replicas = get_replicas(key, N)
    version = generate_version()
    acks = 0
    for replica in replicas:
        if replica.write(key, value, version):
            acks += 1
            if acks >= W:
                return True
    return False  # Did not achieve quorum
```

### Tunable Consistency Check (Cassandra-style)

```python
def is_quorum_satisfied(acks, consistency_level, replication_factor):
    if consistency_level == "ONE":
        return acks >= 1
    elif consistency_level == "QUORUM":
        return acks >= (replication_factor // 2) + 1
    elif consistency_level == "ALL":
        return acks >= replication_factor
    return False
```

---

## 14. Interview Discussion

### Key Points to Articulate

1. **"What consistency does your system need?"** — Start with use case; don't default to strong
2. **CAP tradeoff** — Under partition: choose consistency or availability
3. **Quorum math** — W + R > N ensures overlap; W = R = majority is common
4. **Real systems** — Spanner (linearizable + TrueTime), DynamoDB (tunable), Cassandra (tunable)

### Common Questions

- **"How does Spanner achieve linearizability?"** — TrueTime + commit wait; bounded clock uncertainty
- **"When would you use eventual over strong?"** — High read throughput, non-critical data, multi-region
- **"What's the difference between sequential and linearizable?"** — Linearizable has real-time constraint; sequential only preserves per-process order
- **"How do you prevent split-brain?"** — Quorum, leader election, fencing tokens

### Red Flags to Avoid

- Saying "we need strong consistency" without justifying
- Ignoring latency/availability tradeoffs
- Confusing consistency with durability (they're different)

---

## 15. Deep Dive: Linearizability Proof Sketch

Linearizability requires that for every history H of operations, there exists a legal sequential history S such that:
1. S is equivalent to H (same operations, same responses)
2. S respects real-time ordering: if op1 completes before op2 starts, op1 precedes op2 in S

**Example**: If client A writes x=1 and completes, then client B reads x, B must see 1. The write "takes effect" at some instant between its start and end; that instant is before B's read.

---

## 16. Deep Dive: CAP Theorem Nuances

**CAP** (Brewer): In presence of network partition, choose 2 of 3: Consistency, Availability, Partition tolerance.

**Reality**: Partition tolerance is unavoidable (networks partition). So effectively: CP or AP.

**PACELC** (Abadi): Extension — when no partition: tradeoff between Latency (L) and Consistency (C). So: PA/EL or PC/EC.

**Examples**:
- DynamoDB: AP when partitioned; when not, can choose strong (PC) or eventual (PA) per request
- Cassandra: Tunes per-request; typically AP

---

## 17. Deep Dive: Session Guarantees

**Read-your-writes**: User sees own updates. Implemented by: routing reads to same replica that got the write, or by version vectors.

**Monotonic reads**: User never sees "older" data after seeing "newer". Implemented by: sticky sessions, or tracking "max version seen" per session.

**Causal consistency**: Stronger — if A's write caused B's write (e.g., B read A's data), everyone sees A before B.

---

## 18. Additional Diagrams: Read/Write Ordering

### Linearizability Timeline

```
Time --->
    |---- W(x=1) ----|     |-- R(x) --|
    A                 A    B          B
    [===== write =====]    [= read =]
    
    Valid linearization: W completes, then R starts. R must return 1.
```

### Eventual Consistency Timeline

```
    R1:  W(x=1) -------- W(x=2) --------
    R2:  W(x=1) -------- (delayed) ---- W(x=2)
    R3:  W(x=1) ----------------------- (not yet)
    
    Reads from R2, R3 may see x=1 even after R1 has x=2.
    Eventually all converge to x=2.
```

---

## 19. Tunable Consistency Decision Tree

```
    Need strong consistency?
    ├── Yes → Can you afford latency?
    │   ├── Yes → Use Linearizable (Spanner, etcd)
    │   └── No → Use Quorum (Cassandra QUORUM)
    └── No → Need read-your-writes?
        ├── Yes → Session-based or Causal
        └── No → Eventual (Cassandra ONE, DynamoDB eventual)
```

---

## 20. Interview Walkthrough: Designing a System

**Question**: "Design a distributed cache with configurable consistency."

**Answer structure**:
1. **Requirements**: Read/write latency, consistency level (strong vs eventual), fault tolerance
2. **Sharding**: Consistent hashing for key distribution
3. **Replication**: N replicas per key; quorum W, R
4. **Consistency**: Per-request CL: ONE (eventual), QUORUM (strong for single DC), ALL (strongest)
5. **Failure**: If < quorum available, fail or return stale (configurable)
6. **Real system**: DynamoDB, Cassandra as references

---

## 21. Serializability vs. Linearizability

**Serializability**: Transactions appear as if executed in some serial order. No guarantee on real-time.

**Linearizability**: Operations (including multi-object) appear atomic at some instant. Stronger than serializability for single operations.

**Strict serializability**: Serializability + real-time ordering. Equivalent to linearizability for transactions.

**Example**: T1: W(x) W(y). T2: R(x) R(y). Serializability: T1 then T2, or T2 then T1. Linearizability: Same, but if T1 commits before T2 starts, T1 must precede T2 in the order.

---

## 22. Consistency in Multi-Region

**Challenge**: Cross-region latency (50-200ms). Strong consistency requires synchronous cross-region replication.

**Options**:
1. **Primary in one region**: Writes go to primary; async replicate. Reads from primary = strong; from replica = eventual.
2. **Multi-primary**: Each region has primary for its partition. Cross-region eventual.
3. **Synchronous multi-region**: Spanner, CockroachDB. Use TrueTime/HLC. Higher latency.
4. **CRDTs**: No coordination; merge on read. Good for collaborative data.

---

## 23. Brewer's Conjecture and Formal CAP

**Brewer (2000)**: Choose 2 of 3: C, A, P. **Gilbert & Lynch (2002)**: Formalized. In presence of partition, cannot have both C and A.

**Partition**: Network split; some nodes cannot communicate. **Availability**: Every request gets response. **Consistency**: Every read sees latest write.

**Proof sketch**: Partition into {A} and {B,C}. Client writes to A. Client in other partition reads from B. B doesn't have write. Either return stale (violate C) or block (violate A).

---

## 24. Consistency Levels in Detail: Cassandra

| Level | Read | Write | Use Case |
|-------|------|-------|----------|
| ONE | Any 1 replica | Any 1 replica | Lowest latency |
| TWO | Any 2 | Any 2 | Slightly stronger |
| THREE | Any 3 | Any 3 | For RF=3 |
| QUORUM | Majority | Majority | Strong for single DC |
| ALL | All | All | Strongest, slowest |
| LOCAL_QUORUM | Majority in local DC | Same | Multi-DC, avoid cross-DC |
| EACH_QUORUM | Majority in each DC | Same | Strong per DC |

---

## 25. Summary Cheat Sheet

| Model | Guarantee | Implementation | Latency |
|-------|-----------|---------------|---------|
| Linearizable | Real-time atomic | Leader + sync, TrueTime | High |
| Sequential | Same order, no real-time | Leader | Medium |
| Causal | Happens-before | Vector clocks | Medium |
| RYW | Own writes visible | Session routing | Low |
| Monotonic | No time travel | Sticky session | Low |
| Eventual | Converge eventually | Async replication | Lowest |

---

## 26. Further Reading

- **Spanner paper**: "Spanner: Google's Globally Distributed Database" — TrueTime, external consistency
- **Dynamo paper**: "Dynamo: Amazon's Highly Available Key-value Store" — eventual consistency, vector clocks
- **CAP**: Gilbert & Lynch, "Brewer's Conjecture and the Feasibility of Consistent, Available, Partition-Tolerant Web Services"
- **PACELC**: Abadi, "Consistency Tradeoffs in Modern Distributed Database System Design"

**Interview tip**: Always tie consistency choice to business requirements. "We need linearizability because we're handling financial transactions" is stronger than "we need strong consistency."

**Quorum formula**: For N replicas, W + R > N ensures read-after-write consistency. Common: W = R = (N+1)/2 (majority).

**Example**: N=5, W=3, R=3. Write goes to 3 replicas. Read from 3 replicas. Any read quorum intersects any write quorum in at least one replica. That replica has the latest write. Thus, the read returns the most recent value. This is the mathematical foundation of quorum-based consistency.

---

**End of Document** — Use this as a reference for FAANG distributed systems interviews.
