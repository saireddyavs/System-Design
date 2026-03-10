# CAP Theorem & Consistency Models: Staff+ Engineer Deep Dive

## 1. Concept Overview

### Definition
The **CAP theorem** (Brewer's theorem, 2000) states that a distributed system can provide at most two of the following three guarantees simultaneously:

1. **Consistency (C)**: Every read receives the most recent write or an error
2. **Availability (A)**: Every request receives a (non-error) response
3. **Partition tolerance (P)**: The system continues operating despite arbitrary network partitions

**Key insight**: In the presence of a network partition, you must choose between Consistency and Availability. You cannot have both.

### Purpose
- **Design decisions**: Guides choice of distributed database/system
- **Tradeoff clarity**: Makes explicit what you're sacrificing
- **Expectation setting**: Helps stakeholders understand system behavior
- **Troubleshooting**: Explains "why did I get stale data?" during partitions

### Why It Exists
- **Distributed systems reality**: Networks partition (cables cut, switches fail, datacenters isolate)
- **Fundamental tradeoff**: During partition, either wait for consistency (block) or respond with possibly stale data (available)
- **No free lunch**: You cannot have perfect consistency, perfect availability, AND partition tolerance

### Problems It Solves
1. **Unrealistic expectations**: "Why can't we have strong consistency and 99.99% availability during partitions?"
2. **Wrong tool selection**: Using CP system when AP needed, or vice versa
3. **Architecture confusion**: Understanding why systems behave differently under failure
4. **Design rationale**: Justifying consistency/availability choices to stakeholders

---

## 2. Real-World Motivation

### Google Spanner (CP)
- **Choice**: Consistency + Partition tolerance
- **Tradeoff**: During partition, may return errors or block until resolved
- **Use case**: Financial data, global transactions, strong consistency required
- **Mechanism**: TrueTime, 2PC, Paxos for replication

### Cassandra (AP)
- **Choice**: Availability + Partition tolerance
- **Tradeoff**: Always responds; may return stale data during partition
- **Use case**: High write throughput, can tolerate eventual consistency (social feeds, metrics)
- **Mechanism**: Tunable consistency (ONE, QUORUM, ALL); last-write-wins

### DynamoDB (Configurable)
- **Default**: Eventually consistent reads (AP)
- **Option**: Strongly consistent reads (CP) - may have higher latency, lower throughput
- **Use case**: E-commerce (eventual for recommendations, strong for checkout)

### MongoDB
- **Default**: CP (strong consistency within replica set)
- **Sharded**: Per-shard consistency; cross-shard eventually consistent
- **Use case**: Document store with flexible consistency needs

### Netflix
- **Choice**: AP for recommendations, playback state
- **Reason**: Better to show slightly stale recommendations than nothing
- **Fallback**: Client-side caching, graceful degradation

---

## 3. Architecture Diagrams

### CAP Triangle (Venn Diagram)

```
                    CONSISTENCY
                         *
                        /|\
                       / | \
                      /  |  \
                     /   |   \
                    /    |    \
                   /  CP | AP  \
                  /     |      \
                 /      |       \
                /   CA  |   CA   \
               /    (impossible) \
              /___________________\
         PARTITION              AVAILABILITY
         TOLERANCE

In practice: CA doesn't exist in distributed systems
(because partitions WILL happen)
Real choice: CP vs AP
```

### Network Partition Scenario

```
NORMAL OPERATION
================
    Region A                    Region B
    ┌─────────┐    Network      ┌─────────┐
    │ Node 1  │◄──────────────►│ Node 2  │
    │  (DB)   │    (healthy)   │  (DB)   │
    └─────────┘                └─────────┘
    Write: X=1                  Replicated: X=1
    Read: X=1 ✓                 Read: X=1 ✓

DURING PARTITION
================
    Region A                    Region B
    ┌─────────┐    XXXXX       ┌─────────┐
    │ Node 1  │    (split)     │ Node 2  │
    │  (DB)   │                │  (DB)   │
    └─────────┘                └─────────┘
    Write: X=2                  Write: X=3
    Read: X=2                   Read: X=3
    
    CP: One side blocks or errors (wait for sync)
    AP: Both respond (X=2, X=3) - inconsistency!
```

### CP System Behavior

```
CP: Consistency + Partition Tolerance
=====================================
    Client A ──► Node 1 (partitioned) ──X── Node 2
                      │
                      ▼
              [BLOCK or ERROR]
              "Cannot guarantee consistency"
              Wait for partition to heal
              OR return error
```

### AP System Behavior

```
AP: Availability + Partition Tolerance
======================================
    Client A ──► Node 1 ──► Response (X=2) ✓
    Client B ──► Node 2 ──► Response (X=3) ✓
    
    Both respond immediately
    Conflict resolved later (last-write-wins, vector clocks, etc.)
```

### PACELC Extension

```
PACELC: When Partition (P), choose A or C
         Else (E), when Latency (L), choose C or Consistency

         Partition?         No Partition?
              │                    │
              ▼                    ▼
         A or C?              Latency vs Consistency?
              │                    │
         ┌────┴────┐          ┌─────┴─────┐
         ▼         ▼          ▼           ▼
    Choose A   Choose C   Low latency  Strong consistency
    (available) (consistent) (fast)    (correct)
```

---

## 4. Core Mechanics

### Why You Can Only Have 2 of 3

**Proof sketch:**
1. **Partition happens**: Network splits nodes into two groups that cannot communicate
2. **Write occurs**: Client writes to Node A (in group 1)
3. **Read occurs**: Client reads from Node B (in group 2)
4. **Choice**:
   - **Consistency**: Node B cannot have the latest data (not replicated). Must either:
     - Block/wait (unavailable) until partition heals
     - Return error (unavailable)
   - **Availability**: Node B must respond. It can only return:
     - Stale data (inconsistent)
     - Or make something up (inconsistent)

**Conclusion**: During partition, C and A are mutually exclusive.

### Why CA Doesn't Exist in Practice

- **Distributed systems**: By definition, multiple nodes = network between them
- **Networks partition**: It's not "if" but "when" (Gilbert & Lynch proof, 2002)
- **Single-node "CA"**: Only single-machine DBs are CA (no distribution = no partition)
- **Real systems**: All distributed systems must tolerate partitions → must choose CP or AP

### Consistency Model Hierarchy (Strongest to Weakest)

```
STRONGEST
    │
    ├── Linearizability (strict serializability)
    │   └── Single copy illusion; real-time ordering
    │
    ├── Sequential consistency
    │   └── All see same order; order consistent with program order
    │
    ├── Causal consistency
    │   └── Causally related operations ordered; concurrent may differ
    │
    ├── Eventual consistency
    │   └── All replicas converge if no new writes
    │
    └── Read-your-writes, Monotonic reads (session guarantees)
    └── WEAKEST
```

---

## 5. Numbers

### CAP System Classification

| System | CAP | Consistency | Availability | Notes |
|--------|-----|-------------|--------------|-------|
| **ZooKeeper** | CP | Strong | No (during partition) | Blocks on partition |
| **etcd** | CP | Strong | No | Raft consensus |
| **Consul** | CP | Strong | No | Raft |
| **Cassandra** | AP | Eventual | Yes | Tunable per-query |
| **DynamoDB** | AP/CP | Configurable | Yes | Strong read option |
| **Riak** | AP | Eventual | Yes | CRDTs available |
| **MongoDB** | CP | Strong (replica set) | No | Per-shard |
| **HBase** | CP | Strong | No | |
| **Redis Cluster** | CP | Strong | No | |
| **Spanner** | CP | Strong | No | TrueTime |
| **CockroachDB** | CP | Strong | No | |

### Consistency vs Latency (Typical)

| Consistency Level | Read Latency | Write Latency | Use Case |
|-------------------|--------------|---------------|----------|
| Strong (linearizable) | 1-10ms | 5-50ms | Financial, inventory |
| Causal | 1-5ms | 2-20ms | Social, chat |
| Eventual | 1-2ms | 1-5ms | Feeds, analytics |
| Read-your-writes | 1-3ms | 2-10ms | User profile |

### Partition Frequency (Rough Estimates)

| Environment | Partition probability | Mitigation |
|-------------|------------------------|------------|
| Single datacenter | Low (0.01%/year) | Multi-rack |
| Multi-AZ same region | Low (0.1%/year) | AZ redundancy |
| Multi-region | Medium (1%/year) | Design for partition |
| Global | Higher | AP or careful CP |

---

## 6. Tradeoffs

### CP vs AP Decision Matrix

| Factor | Choose CP | Choose AP |
|--------|------------|-----------|
| **Data type** | Financial, inventory | Social feed, metrics |
| **Consistency critical** | Yes (double-spend bad) | No (stale OK) |
| **Availability critical** | Can tolerate downtime | Must always respond |
| **Conflict resolution** | Easy (one source of truth) | Complex (merge, LWW) |
| **Latency** | Higher (consensus) | Lower (local read) |
| **Complexity** | Consensus protocols | Conflict resolution |

### Consistency Model Tradeoffs

| Model | Pros | Cons |
|-------|------|------|
| **Linearizable** | Simple mental model | Higher latency, lower throughput |
| **Sequential** | Easier to reason | Weaker than linearizable |
| **Causal** | Good balance | Harder to implement |
| **Eventual** | Fast, available | Stale reads, conflicts |
| **Read-your-writes** | Good UX | Per-session only |

### PACELC Tradeoffs

| Scenario | Choice | Tradeoff |
|----------|--------|----------|
| Partition | A | Availability over consistency |
| Partition | C | Consistency over availability |
| No partition, low latency | Consistency | Correctness over speed |
| No partition, low latency | Latency | Speed over strong consistency |

---

## 7. Variants / Implementations

### PACELC Theorem (2012)
- **P**artition: A or C (same as CAP)
- **E**lse: **L**atency vs **C**onsistency
- **Insight**: Even without partition, there's a tradeoff (consensus costs latency)
- **Example**: DynamoDB - strong consistent read has higher latency than eventual

### Consistency Patterns

1. **Strong consistency**: 2PC, Paxos, Raft (CP systems)
2. **Eventual consistency**: Anti-entropy, gossip (AP systems)
3. **Causal consistency**: Vector clocks, version vectors
4. **Read-your-writes**: Session stickiness + replication
5. **Monotonic reads**: Same replica or newer for session

### Conflict Resolution (AP Systems)

| Strategy | Description | Use Case |
|----------|-------------|----------|
| **Last-write-wins (LWW)** | Timestamp decides | Simple, clock sync critical |
| **Vector clocks** | Partial ordering | Causal consistency |
| **CRDTs** | Math ensures merge | Collaborative editing |
| **Application-level** | Custom merge logic | Business rules |
| **Multi-version** | Keep all, user resolves | Document editing |

---

## 8. Scaling Strategies

### Scaling CP Systems
- **Sharding**: Each shard is CP; cross-shard eventually consistent
- **Read replicas**: Replicas may lag; "read from primary" for strong consistency
- **Caching**: Cache invalidates on write; consistency window
- **Limit scope**: Strong consistency only where needed (e.g., checkout)

### Scaling AP Systems
- **Add nodes**: Linear scalability (no consensus)
- **Tunable consistency**: QUORUM for important reads, ONE for speed
- **Conflict resolution**: Design for merge from start
- **Monitoring**: Track replication lag, conflict rate

### Hybrid Approaches
- **CQRS**: Write path CP (command), read path AP (query)
- **Saga**: Eventual consistency with compensation
- **Two-phase**: Strong for critical path, eventual for rest

---

## 9. Failure Scenarios

### Real Production Failures

**GitHub Split-Brain (2018)**
- **Cause**: Database failover created two primaries
- **Impact**: Writes to both; data inconsistency
- **Lesson**: CP systems need careful failover; consensus prevents split-brain

**Amazon Dynamo (Design)**
- **Choice**: AP (always available)
- **Tradeoff**: Shopping cart could have temporary inconsistencies
- **Mitigation**: Merge algorithms, last-write-wins for cart

**MongoDB Replica Set Issues**
- **Scenario**: Primary fails; election; possible rollback
- **CP behavior**: Blocks during election
- **Mitigation**: Multiple replicas, fast failover

### Partition Handling

| System | On Partition | Recovery |
|--------|--------------|----------|
| **ZooKeeper** | Stops serving (minority partition) | Rejoin when healed |
| **Cassandra** | Serves stale data | Hinted handoff, repair |
| **DynamoDB** | Serves (eventual) | Sync when connected |
| **Spanner** | May block | TrueTime helps |

---

## 10. Performance Considerations

### Consistency Cost
- **Strong consistency**: 2-5x latency vs eventual (consensus round-trips)
- **Cross-region strong**: 100-300ms (speed of light)
- **Eventual**: Single node read; 1-2ms

### When to Use What
- **Strong**: Money, inventory, identity
- **Eventual**: Feeds, recommendations, analytics
- **Causal**: Chat, social (reply after message)
- **Read-your-writes**: User profile, settings

### Monitoring
- **Replication lag**: How far behind are replicas?
- **Conflict rate**: How often do conflicts occur?
- **Consistency violations**: Detected anomalies

---

## 11. Use Cases

| System | Consistency | Reason |
|--------|-------------|--------|
| **YouTube** | Eventual (views, likes) | High scale, stale OK |
| **Uber** | Strong (ride state) | Can't double-assign driver |
| **Netflix** | Eventual (recommendations) | Availability over freshness |
| **WhatsApp** | Eventual (message delivery) | Offline support, merge |
| **Stripe** | Strong (payments) | No double charge |
| **Banking** | Strong (balances) | Regulatory, correctness |
| **Twitter** | Eventual (timeline) | Scale, viral events |
| **Amazon Cart** | Eventual + merge | Availability, merge logic |

---

## 12. Comparison Tables

### Database CAP Classification

| Database | CAP | Default Read | Strong Option |
|---------|-----|--------------|---------------|
| Cassandra | AP | Eventual | No (QUORUM is probabilistic) |
| DynamoDB | AP | Eventual | Yes (strongly consistent read) |
| MongoDB | CP | Strong | N/A |
| Redis | CP | Strong | N/A |
| PostgreSQL | CA/CP | Strong | N/A |
| CockroachDB | CP | Strong | N/A |
| Riak | AP | Eventual | No |
| HBase | CP | Strong | N/A |

### Consistency Model Comparison

| Model | Stale Read? | Ordering | Implementation |
|-------|-------------|----------|----------------|
| Linearizable | No | Real-time | 2PC, Paxos |
| Sequential | No | Per-process | Timestamps |
| Causal | Maybe | Causal only | Vector clocks |
| Eventual | Yes | None | Gossip |
| Read-your-writes | No (own) | Session | Sticky + sync |

---

## 13. Code or Pseudocode

### CAP Decision Logic

```python
def handle_read_request(key, node_id, partition_detected):
    """
    Simplified CAP decision during partition
    """
    if partition_detected:
        # Must choose: Consistency or Availability
        if system_type == 'CP':
            # Block or error - cannot guarantee fresh data
            raise ConsistencyError("Partition detected, cannot serve consistent read")
        elif system_type == 'AP':
            # Serve potentially stale data
            return local_read(key)  # May be stale
    else:
        # No partition - can have both (until next partition)
        return local_read(key)
```

### Vector Clock (Causal Consistency)

```python
class VectorClock:
    def __init__(self):
        self.clocks = {}  # node_id -> counter
    
    def increment(self, node_id):
        self.clocks[node_id] = self.clocks.get(node_id, 0) + 1
        return self.clone()
    
    def merge(self, other):
        for node, count in other.clocks.items():
            self.clocks[node] = max(self.clocks.get(node, 0), count)
    
    def happens_before(self, other):
        """True if self happened before other"""
        return (all(self.clocks.get(n, 0) <= other.clocks.get(n, 0) for n in other.clocks)
                and self.clocks != other.clocks)
```

### Last-Write-Wins Conflict Resolution

```python
def resolve_lww(versions):
    """Last-write-wins: highest timestamp wins"""
    return max(versions, key=lambda v: v.timestamp)

def resolve_vector_clock(versions):
    """Vector clock: concurrent = conflict, otherwise newer wins"""
    for v1 in versions:
        if all(v1.happens_before(v2) or v1 == v2 for v2 in versions):
            return v1  # v1 is latest
    return None  # Concurrent - conflict, need merge
```

---

## 14. Interview Discussion

### How to Explain CAP
1. **State the theorem**: "In distributed systems, during a network partition, you choose consistency OR availability"
2. **Clarify**: "Partition = nodes can't communicate; happens in real networks"
3. **Give examples**: "ZooKeeper is CP - blocks during partition; Cassandra is AP - always responds"
4. **Nuance**: "CA doesn't exist in distributed systems; it's really CP vs AP"

### When Interviewers Expect It
- **System design**: "Design a distributed cache" → CAP tradeoffs
- **Database choice**: "Why Cassandra over PostgreSQL for this?"
- **Deep dive**: "What happens during a network partition?"
- **Tradeoffs**: "How do you balance consistency and availability?"

### Key Points to Hit
- P is inevitable in distributed systems
- CP = consistency over availability (block/error during partition)
- AP = availability over consistency (stale data possible)
- PACELC extends to normal operation (latency vs consistency)
- Different consistency models for different needs

### Follow-Up Questions
- "How would you implement strong consistency?"
- "What's the difference between eventual and strong consistency?"
- "When would you choose AP over CP?"
- "What is PACELC?"
- "How does DynamoDB handle CAP?"

### Common Mistakes
- Saying "you can have 2 of 3" without explaining partition scenario
- Claiming a system is "CA" when it's distributed
- Confusing consistency models (eventual vs strong)
- Not connecting CAP to actual system design decisions
