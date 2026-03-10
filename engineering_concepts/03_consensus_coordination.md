# Module 3: Consensus & Distributed Coordination

---

## 1. CAP Theorem

### Definition
In a distributed system experiencing a network **Partition**, you must choose between **Consistency** (all nodes see the same data) and **Availability** (every request gets a response).

### The Three Properties
```
    C ─── Consistency ──── Every read gets the most recent write
    A ─── Availability ── Every request gets a (non-error) response
    P ─── Partition      ── Network can drop/delay messages
           Tolerance

You can only guarantee 2 of 3.
In practice, P always happens, so you choose C or A.
```

### Visual
```
       ┌──── Consistency ─────┐
       │                      │
  CP Systems             CA Systems
  (Refuse requests       (Only works if
   during partition)      no partitions)
       │                      │
  HBase, MongoDB         Single-node RDBMS
  Zookeeper              (not distributed)
       │
       └──── Partition Tolerance ─────┐
                                      │
                                 AP Systems
                                 (Return stale data
                                  during partition)
                                      │
                                 Cassandra, DynamoDB
                                 CouchDB
```

### Reality Check: PACELC
```
If Partition → choose A or C
Else (normal operation) → choose Latency or Consistency

Example:
  DynamoDB: PA/EL (Available during P, Low latency normally)
  Spanner:  PC/EC (Consistent always, uses TrueTime to reduce latency cost)
```

### Interview Tip
"CAP isn't a binary switch — it's a spectrum. Systems like DynamoDB let you TUNE consistency per-request (eventual vs strong reads)."

### Summary
CAP states that during network partitions, distributed systems must choose consistency or availability. Real systems make nuanced per-operation tradeoffs along the PACELC spectrum.

---

## 2. Paxos

### Definition
A consensus protocol that allows a group of unreliable nodes to agree on a single value, tolerating up to ⌊(N-1)/2⌋ failures.

### Problem It Solves
How do distributed nodes agree on who is leader, what the next log entry is, or which transaction commits — even when nodes crash or messages are lost?

### Roles
```
Proposer  → suggests a value
Acceptor  → votes on proposals
Learner   → learns the decided value
```

### How It Works (Two Phases)
```
Phase 1: PREPARE
  Proposer → Acceptors: "Prepare(n)"     (n = proposal number)
  Acceptors: if n > highest_seen:
    Promise not to accept lower proposals
    Reply with any previously accepted value

Phase 2: ACCEPT
  Proposer → Acceptors: "Accept(n, value)"
  Acceptors: if no higher promise made:
    Accept the value
    Reply ACK

  If MAJORITY accept → value is CHOSEN
```

### Visual
```
Proposer         Acceptor1    Acceptor2    Acceptor3
   │                │            │            │
   │──Prepare(1)──→│            │            │
   │──Prepare(1)──→───────────→│            │
   │──Prepare(1)──→────────────────────────→│
   │                │            │            │
   │←──Promise(1)──│            │            │
   │←──Promise(1)──────────────│            │
   │←──Promise(1)──────────────────────────│
   │                │            │            │
   │──Accept(1,v)─→│            │            │
   │──Accept(1,v)─→───────────→│            │
   │──Accept(1,v)─→──────────────────────→│
   │                │            │            │
   │    MAJORITY ACCEPTED → v is chosen     │
```

### Tradeoffs

| Pros | Cons |
|------|------|
| Proven correct (Lamport) | Complex to understand and implement |
| Tolerates minority failures | Multiple proposers can livelock |
| Foundation of all consensus | Multi-Paxos needed for log replication |

### Real Systems
Google Chubby, Google Spanner (internally), Apache ZooKeeper (variant)

### Summary
Paxos is the foundational consensus algorithm using prepare/accept phases with majority quorums. Correct but complex — Raft was created as a more understandable alternative.

---

## 3. Raft

### Definition
A consensus algorithm designed to be understandable. Decomposes consensus into leader election, log replication, and safety — all using majority quorums.

### Problem It Solves
Same as Paxos (distributed agreement) but with a clear, implementable design.

### Three Sub-Problems
```
1. LEADER ELECTION   → Who is in charge?
2. LOG REPLICATION   → Leader sends entries to followers
3. SAFETY            → Only leaders with up-to-date logs win election
```

### Leader Election
```
States: Follower → Candidate → Leader

1. Followers listen for heartbeats from leader
2. No heartbeat for election_timeout → become Candidate
3. Candidate requests votes from all nodes
4. Majority votes → become Leader
5. Leader sends periodic heartbeats

┌──────────┐  timeout  ┌───────────┐  majority  ┌────────┐
│ Follower │ ────────→ │ Candidate │ ─────────→ │ Leader │
└──────────┘           └───────────┘            └────────┘
     ▲                       │                       │
     │    higher term        │     heartbeats        │
     └───────────────────────┘←──────────────────────┘
```

### Log Replication
```
Client → Leader: "SET x=5"

Leader:
  1. Append to own log
  2. Send AppendEntries RPC to all followers
  3. Wait for MAJORITY ack
  4. Commit entry
  5. Apply to state machine
  6. Respond to client
```

### Comparison: Raft vs Paxos

| | Raft | Paxos |
|-|------|-------|
| Understandability | High (designed for it) | Low (notoriously complex) |
| Leader | Strong leader required | Leader optional |
| Implementation | Straightforward | Many variants, subtle bugs |
| Performance | Good (single leader) | Slightly more flexible |
| Adoption | etcd, CockroachDB, TiKV | Chubby, Spanner |

### Edge Cases
- **Split brain**: Prevented by majority requirement (at most 1 leader per term)
- **Stale leader**: Term numbers detect and reject stale leaders
- **Network partition**: Minority partition cannot elect a leader

### Real Systems
etcd, Consul, CockroachDB, TiKV, RethinkDB, InfluxDB

### Interview Tip
"Raft guarantees at most one leader per term. A leader must have the most up-to-date log to win election, which ensures safety."

### Summary
Raft decomposes consensus into leader election, log replication, and safety. Its strong-leader model and clear design make it the go-to choice for building distributed systems.

---

## 4. Two-Phase Commit (2PC)

### Definition
A protocol to ensure all participants in a distributed transaction either ALL commit or ALL abort.

### How It Works
```
Phase 1: PREPARE (Voting)
  Coordinator → All Participants: "Can you commit?"
  Participants: Lock resources, write to WAL
  Participants → Coordinator: "YES" or "NO"

Phase 2: COMMIT/ABORT
  If ALL said YES:
    Coordinator → All: "COMMIT"
  If ANY said NO:
    Coordinator → All: "ABORT"
```

### Visual
```
  Coordinator          Participant A      Participant B
      │                     │                  │
      │──── PREPARE ──────→│                  │
      │──── PREPARE ──────→──────────────────→│
      │                     │                  │
      │←──── YES ──────────│                  │
      │←──── YES ────────────────────────────│
      │                     │                  │
      │──── COMMIT ───────→│                  │
      │──── COMMIT ───────→──────────────────→│
      │                     │                  │
      │      DONE           │    DONE          │
```

### The Blocking Problem
```
What if Coordinator crashes AFTER Phase 1, BEFORE Phase 2?

Participants are stuck:
  - They voted YES and locked resources
  - They don't know if others voted YES or NO
  - They CANNOT commit or abort safely
  - Resources remain LOCKED until coordinator recovers

This is the fundamental flaw of 2PC.
```

### Tradeoffs

| Pros | Cons |
|------|------|
| Guarantees atomicity | Blocking if coordinator crashes |
| Simple protocol | High latency (2 round trips) |
| Widely implemented | Locks held during entire protocol |

### Real Systems
MySQL (XA transactions), PostgreSQL (prepared transactions), Oracle, most RDBMS distributed transactions

### Summary
2PC ensures distributed atomicity through prepare and commit phases. Its critical flaw: participants block indefinitely if the coordinator crashes after prepare.

---

## 5. Three-Phase Commit (3PC)

### Definition
Extension of 2PC that adds a "Pre-Commit" phase to avoid blocking on coordinator failure.

### How It Works
```
Phase 1: CAN-COMMIT?     → "Can you commit?"
Phase 2: PRE-COMMIT      → "Prepare to commit (but don't yet)"
Phase 3: DO-COMMIT       → "Actually commit now"

The pre-commit phase ensures all participants KNOW the decision
before committing, allowing timeout-based recovery.
```

### Why It Helps
```
If coordinator crashes after Phase 2 (Pre-Commit):
  - Participants know everyone voted YES
  - After timeout, they can safely COMMIT

If coordinator crashes after Phase 1:
  - Participants haven't pre-committed
  - After timeout, they safely ABORT
```

### Tradeoffs

| Pros | Cons |
|------|------|
| Non-blocking (timeouts work) | 3 round trips (slower) |
| Participants can recover independently | Doesn't work with network partitions |
| Better than 2PC for crash recovery | Rarely used in practice |

### Why It's Rarely Used
Network partitions (not just crashes) can still cause inconsistency. Modern systems prefer Paxos/Raft-based approaches or Saga patterns.

### Summary
3PC adds a pre-commit phase to 2PC to enable timeout-based recovery. It solves the blocking problem for crashes but not for network partitions, so it's rarely used in practice.

---

## 6. Leader Election

### Definition
The process by which distributed nodes select one node to act as the coordinator/master for making decisions.

### Why It Matters
Many distributed algorithms (Raft, single-leader replication) require exactly ONE leader. Getting this wrong causes split-brain (two leaders accepting conflicting writes).

### Approaches
```
┌─── RAFT-BASED ────────────────┐
│ Random timeout → candidate    │
│ Majority vote → leader        │
│ Term numbers prevent stale    │
└───────────────────────────────┘

┌─── BULLY ALGORITHM ──────────┐
│ Highest ID wins               │
│ Node detects failure → starts │
│   election, higher IDs take   │
│   over                        │
└───────────────────────────────┘

┌─── ZOOKEEPER (Ephemeral Nodes)─┐
│ Nodes create sequential znodes  │
│ Lowest sequence number = leader │
│ Leader dies → znode disappears  │
│ Next in line becomes leader     │
└─────────────────────────────────┘
```

### Split Brain Prevention
```
WRONG:  Network split, both sides elect a leader
        → Two leaders accept writes → data corruption

RIGHT:  Require MAJORITY (N/2 + 1) votes
        → Only one partition can have majority
        → At most ONE leader at a time
```

### Real Systems
Kubernetes (etcd + Raft), Kafka (controller election), Redis Sentinel, ZooKeeper

### Summary
Leader election selects a single coordinator node. Majority-based voting (Raft, ZooKeeper) prevents split-brain. Fencing tokens and term numbers prevent stale leaders from acting.

---

## 7. Quorum

### Definition
The minimum number of nodes that must agree for an operation to succeed. For N replicas, a typical quorum requires W + R > N.

### How It Works
```
N = 3 replicas, W = 2, R = 2

Write "x=5": Must succeed on 2 of 3 nodes
Read:         Must read from 2 of 3 nodes

Since W + R > N (2+2 > 3), at least 1 node
in the read set must have the latest write.

┌──────┐   write "x=5"   ┌──────┐
│Node 1│ ←───────────────│Client│
│ x=5  │ ✓               └──┬───┘
├──────┤                    │
│Node 2│ ←──────────────────┘
│ x=5  │ ✓ (W=2 satisfied, ACK client)
├──────┤
│Node 3│ (stale, x=3)
│ x=3  │
└──────┘

Read: contact 2 nodes → sees x=5 from at least one → returns x=5
```

### Sloppy Quorum
When designated nodes are unreachable, write to ANY available nodes. Repair later via "hinted handoff."

### Tuning

| Config | Behavior |
|--------|----------|
| W=1, R=N | Fast writes, slow reads |
| W=N, R=1 | Slow writes, fast reads |
| W=⌈N/2⌉+1, R=⌈N/2⌉+1 | Balanced (standard) |
| W=1, R=1 | Fast but NO consistency guarantee |

### Real Systems
DynamoDB, Cassandra, Riak, CockroachDB

### Summary
Quorum requires W + R > N to guarantee reading the latest write. It provides tunable consistency — trade read latency for write latency and vice versa.

---

## 8. Eventual Consistency

### Definition
If no new updates are made, all replicas will eventually converge to the same state. There is no guarantee WHEN.

### Visual
```
Time ──→
Node A: [x=1] ─── [x=5] ─── [x=5] ─── [x=5]
Node B: [x=1] ─── [x=1] ─── [x=5] ─── [x=5]  ← replication lag
Node C: [x=1] ─── [x=1] ─── [x=1] ─── [x=5]  ← longer lag

All converge eventually. During lag, reads may return stale data.
```

### Variants
- **Causal consistency**: Preserves cause-effect ordering
- **Read-your-writes**: You always see your own writes
- **Monotonic reads**: You never see older data after seeing newer

### When to Use
- Social media feeds (stale timeline is acceptable)
- DNS propagation
- Shopping cart (merge conflicts later)
- CDN cache propagation

### When NOT to Use
- Banking transactions (balance must be accurate)
- Inventory management (overselling)
- Distributed locks

### Summary
Eventual consistency guarantees convergence without a time bound. It enables high availability and low latency at the cost of temporarily stale reads.

---

## 9. Strong Consistency

### Definition
Every read returns the most recent write. All nodes see the same data at the same time (linearizability).

### How It's Achieved
```
1. Single Leader Replication (sync)
   Client → Leader → wait for ALL replicas → ACK

2. Consensus (Raft/Paxos)
   Majority must agree before commit

3. TrueTime (Google Spanner)
   GPS + atomic clocks bound uncertainty
   Wait out uncertainty interval before commit
```

### Cost
```
Latency: Must wait for replication/consensus
Availability: Cannot serve reads during partition (CAP)
Throughput: Limited by slowest replica
```

### When to Use
- Financial transactions
- Leader election
- Distributed locks
- Configuration management (etcd, ZooKeeper)

### Summary
Strong consistency means every read sees the latest write. It requires synchronous replication or consensus, trading availability and latency for correctness.

---

## 10. Distributed Locks

### Definition
A mechanism to ensure only one process across multiple machines can access a shared resource at a time.

### Approaches

```
┌─── REDIS (Redlock) ──────────────┐
│ SET key value NX PX 30000         │
│ (set if not exists, 30s expiry)   │
│ Acquire on majority of N nodes    │
└───────────────────────────────────┘

┌─── ZOOKEEPER ────────────────────┐
│ Create ephemeral sequential znode │
│ Lowest sequence = lock holder     │
│ Node crash → znode auto-deleted   │
└───────────────────────────────────┘

┌─── ETCD ─────────────────────────┐
│ Lease-based: create key with TTL  │
│ Lease expires → lock released     │
│ Raft-backed → strongly consistent │
└───────────────────────────────────┘
```

### The Fencing Token Problem
```
Process A acquires lock, gets fencing token #33
Process A pauses (GC pause)
Lock expires
Process B acquires lock, gets fencing token #34
Process A resumes, thinks it has lock

Solution: Resource server rejects token #33 after seeing #34

Process A: ──lock(#33)──GC PAUSE──────write(#33)──→ REJECTED
Process B: ────────────lock(#34)──write(#34)──────→ ACCEPTED
```

### Tradeoffs

| Approach | Consistency | Performance | Failure Mode |
|----------|------------|-------------|--------------|
| Redis (single) | Weak | Fast | Data loss on crash |
| Redlock | Debated | Medium | Clock issues |
| ZooKeeper | Strong | Slower | Complex setup |
| etcd | Strong (Raft) | Medium | Lease expiry |

### Interview Tip
"Distributed locks need fencing tokens to handle GC pauses and clock skew. Without them, two processes can hold the lock simultaneously."

### Summary
Distributed locks coordinate access across machines. Use ZooKeeper/etcd for correctness-critical locks. Always use fencing tokens. Redis is fast but not safe for critical sections.

---

## 11. ZooKeeper

### Definition
A centralized coordination service for distributed systems providing: configuration management, naming, distributed locks, and leader election.

### Core Abstractions
```
Data Model: Hierarchical namespace (like a filesystem)
/
├── /config
│   ├── /config/db_host = "10.0.0.1"
│   └── /config/cache_ttl = "300"
├── /locks
│   ├── /locks/resource_1
│   └── /locks/resource_2
└── /election
    ├── /election/node_0001  (ephemeral)
    └── /election/node_0002  (ephemeral)
```

### Key Features
- **Ephemeral nodes**: Deleted when session disconnects (crash detection)
- **Sequential nodes**: Auto-incrementing suffix (for ordering)
- **Watches**: Clients get notified on changes (event-driven)

### Leader Election with ZooKeeper
```
1. Each node creates /election/node_ (ephemeral + sequential)
2. Node with lowest sequence number is leader
3. Others watch the node JUST before them
4. Leader crashes → ephemeral node deleted → next node notified → becomes leader
```

### Real Systems
Kafka (broker coordination), HBase (master election), Hadoop (YARN), Solr

### Summary
ZooKeeper provides primitives (ephemeral nodes, watches, sequential nodes) for building distributed locks, leader election, and configuration management.

---

## 12. etcd

### Definition
A strongly consistent, distributed key-value store using Raft consensus. The backbone of Kubernetes.

### Comparison: etcd vs ZooKeeper

| | etcd | ZooKeeper |
|-|------|-----------|
| Consensus | Raft | ZAB (Zookeeper Atomic Broadcast) |
| API | gRPC + HTTP/JSON | Custom TCP protocol |
| Data model | Flat key-value | Hierarchical (znodes) |
| Watch | Reliable (revision-based) | Can miss events |
| Language | Go | Java |
| Kubernetes | Native integration | Possible but not standard |

### Use Cases
- Kubernetes: stores all cluster state
- Service discovery: services register themselves
- Distributed configuration
- Leader election via lease-based locks

### Summary
etcd is a Raft-based strongly consistent KV store. It's simpler than ZooKeeper, has reliable watches, and is the standard coordination service for Kubernetes.

---

## 13. Lease-Based Locks

### Definition
A distributed lock with a built-in expiration time (lease). If the holder crashes, the lock automatically releases after the lease expires.

### How It Works
```
1. Client requests lock with TTL (e.g., 30 seconds)
2. Server grants lock + lease
3. Client must RENEW lease before expiry
4. If client crashes → no renewal → lease expires → lock released
5. Other clients can now acquire
```

### Visual
```
Time ──→
Client A: ──[LOCK]──────────[RENEW]──────────[RENEW]──[RELEASE]
                   lease=30s        lease=30s

Client A crash scenario:
Client A: ──[LOCK]──────────── X (crash, no renewal)
                              │
                              ▼ 30s later
Client B: ──────────────────────────────[LOCK]── (acquired!)
```

### Edge Case: Process Pause
```
Client A gets lock (30s lease)
Client A enters GC pause for 35 seconds
Lease expires → Client B gets lock
Client A wakes up, thinks it still has lock → DANGER

Solution: Use fencing tokens (monotonically increasing IDs)
```

### Real Systems
etcd leases, DynamoDB lock client, Google Chubby, Consul sessions

### Summary
Lease-based locks auto-expire to prevent deadlocks from crashed holders. They must be paired with fencing tokens to handle process pauses safely.
