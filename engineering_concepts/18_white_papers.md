# Module 18: Software White Papers

---

> **Purpose:** Deep-dive analysis of foundational systems papers that shaped modern distributed systems.
> **Format:** Each paper: Definition → Problem → How It Works → Diagram → Tradeoffs → Interview Tips

---

## 1. Amazon Dynamo (2007)

### Definition

**"Dynamo: Amazon's Highly Available Key-value Store"** — The 2007 paper that launched the NoSQL revolution. Dynamo is a distributed key-value store designed for **always-writable** workloads, prioritizing availability over strong consistency.

### Problem It Solves

```
Amazon's shopping cart must ALWAYS be writable, even during:
  - Network partitions
  - Datacenter failures
  - Node crashes

Rejecting a write = lost sale. Better to accept and resolve conflicts later.
```

**Core design choice:** Availability over strong consistency. "Shopping cart must never reject a write."

### How It Works

#### Key Techniques Introduced

```
┌─── Techniques from Dynamo Paper ─────────────────────────────┐
│                                                               │
│  1. Consistent Hashing (with virtual nodes)                   │
│     → Data distribution, minimal rebalancing on scale         │
│                                                               │
│  2. Vector Clocks                                             │
│     → Conflict detection without synchronized clocks          │
│                                                               │
│  3. Sloppy Quorum (W + R > N)                                 │
│     → Availability: don't require all N replicas              │
│                                                               │
│  4. Hinted Handoff                                             │
│     → Temporary failure handling: store writes for down nodes  │
│                                                               │
│  5. Merkle Trees                                              │
│     → Anti-entropy repair: efficient replica sync             │
│                                                               │
│  6. Gossip Protocol                                           │
│     → Membership detection without central coordinator        │
│                                                               │
└───────────────────────────────────────────────────────────────┘
```

#### Tunable Consistency (N, R, W Parameters)

```
N = Replication factor (how many nodes hold each key)
R = Read quorum (minimum nodes to read from)
W = Write quorum (minimum nodes to acknowledge write)

Guarantees:
  W + R > N  →  At least one overlap between read and write
                → Read sees at least one latest write

Examples:
  N=3, R=2, W=2  →  Quorum read/write, 1 overlap
  N=3, R=1, W=3  →  Strong write, weak read (fast reads)
  N=3, R=3, W=1  →  Strong read, weak write (fast writes)
```

#### Read Path (Detail)

```
1. Client requests key K
2. Coordinator (or client) hashes K → consistent hashing ring
3. Identifies N successor nodes (preference list)
4. Sends read to R nodes in parallel
5. Waits for R responses
6. If multiple versions (vector clock conflict) → return all to client
7. Client resolves conflict (application-specific)
8. Optionally: read repair (push latest to stale replicas)
```

#### Write Path (Detail)

```
1. Client sends put(key, value) with vector clock
2. Coordinator hashes key → preference list of N nodes
3. Sends write to N nodes (or first healthy N)
4. If node down → hinted handoff: store on alternate node
5. Wait for W acknowledgments (hints count as acks in sloppy quorum)
6. Return success to client
7. Background: hinted node replays to target when it recovers
```

#### Client-Side Conflict Resolution

```
Dynamo does NOT resolve conflicts server-side.
Returns all conflicting versions to the application.

Application strategies:
  - Last-writer-wins (LWW) — use timestamp
  - Merge (e.g., shopping cart: union of items)
  - Custom business logic (e.g., highest version wins)
```

### Diagrams

#### Consistent Hashing Ring (with Virtual Nodes)

```
            0°
            │
    330° ── C ── 90°
   /    \        /  \
  /  A-2 \      / A-0 \
 /        \    /       \
C          A ─────────── A
 \   B-1  /    \  B-2  /
  \      /      \     /
   \ B-0/        \   /
    ─────────────────
        180°

Physical: A, B, C
Virtual:  A-0, A-1, A-2, B-0, B-1, B-2, C-0, C-1, C-2
Key K → hash(K) → walk CW to next vnode → owner
```

#### Read/Write Quorum (N=3, R=2, W=2)

```
         N1 ──── N2 ──── N3
          │       │       │
    Write │ ✓     │ ✓     │ (hint on N2 for N3 if down)
          │       │       │
    Read  │ ✓     │ ✓     │ (R=2, overlap with write)
          │       │       │
    Overlap: at least 1 node has latest write
```

#### Vector Clock Example

```
Node A:  [1,0,0] ──write──→ [2,0,0]
                \
Node B:   [1,1,0] ──merge──→ [2,2,0]
                              ↑
Node C:        [0,0,1] ──write──→ [0,0,2]  ← CONCURRENT with A,B
```

#### Hinted Handoff Flow

```
    Client              Coordinator           Replicas
       │                      │                   │
       │  PUT key=K           │                   │
       │─────────────────────>│                   │
       │                      │  Replicate to N1,N2,N3
       │                      │──────────────────>│
       │                      │     ✓     ✓    X (N3 down)
       │                      │                   │
       │                      │  Store HINT on N2  │
       │                      │  (for N3 when back)│
       │                      │                   │
       │  200 OK (W=2 ack)     │                   │
       │<─────────────────────│                   │
       │                      │                   │
       │              ... N3 recovers ...         │
       │                      │  Replay hints ───>│ N3
       │                      │  from N2          │
```

### Tradeoffs

| Pros | Cons |
|------|------|
| Always writable (high availability) | Eventual consistency; conflicts possible |
| No single point of failure | Client must resolve conflicts |
| Tunable N, R, W per workload | Vector clocks don't scale past ~100s of nodes |
| Handles partitions gracefully | Complex operational model |

### Legacy & Real Systems

```
Dynamo (paper) → inspired:
  - Amazon DynamoDB (managed, simplified)
  - Apache Cassandra (wide-column, tunable consistency)
  - Riak (Basho) — key-value
  - LinkedIn Voldemort — key-value
```

### Interview Deep Dive

**"Explain any concept from the Dynamo paper"**

- **Consistent hashing:** Maps keys and nodes to a ring. Adding/removing a node only redistributes O(K/N) keys. Virtual nodes balance load.
- **Vector clocks:** Track causality per node. If V1 and V2 are incomparable, writes are concurrent → client must resolve.
- **Sloppy quorum:** W and R don't have to be strict subsets of N; we can accept hints. Ensures availability when nodes are down.
- **Hinted handoff:** When replica N3 is down, store write on N2 as "hint for N3." When N3 recovers, replay hints.
- **Merkle trees:** Anti-entropy: compare replicas by hash tree. Leaf = hash of key range. Mismatch → sync that subtree.

### Summary

Dynamo prioritizes availability over consistency. Its techniques (consistent hashing, vector clocks, sloppy quorum, hinted handoff, Merkle trees, gossip) became the foundation for distributed key-value stores. The paper is essential reading for understanding DynamoDB, Cassandra, and Riak.

---

## 2. Google Spanner (2012)

### Definition

**"Spanner: Google's Globally-Distributed Database"** — A globally distributed, strongly consistent database that supports SQL-like queries and external consistency across datacenters. The key innovation is **TrueTime**, which enables global transaction ordering without a single global clock.

### Problem It Solves

```
How do you run transactions across continents with STRONG consistency?

Traditional approaches fail:
  - 2PC across regions: high latency, blocking
  - Logical clocks: can't order across independent clusters
  - NTP: unreliable for ordering (ms-level drift)

Spanner's answer: TrueTime — hardware clocks with bounded uncertainty.
```

### How It Works

#### TrueTime API (In Detail)

```
Every Google datacenter has:
  - GPS receivers (primary time source)
  - Atomic clocks (backup when GPS unavailable)

TrueTime API:
  TT.now()  →  returns [earliest, latest]  (uncertainty interval)
  TT.after(t)   →  true if t has definitely passed
  TT.before(t)  →  true if t has definitely not arrived

Typical uncertainty: 1–7 ms
```

**Why uncertainty?** Even with GPS + atomic clocks, there's clock skew. The interval captures "we don't know exactly when now is, but it's definitely in this range."

#### Commit-Wait Rule (External Consistency)

```
External consistency guarantee:
  If transaction T1 commits before T2 starts,
  then timestamp(T1) < timestamp(T2). GUARANTEED.

How:
  1. Transaction chooses commit timestamp T.
  2. Commit-wait: WAIT until TT.now().earliest > T
  3. Then commit. Guaranteed: "now" is definitely in the future of T.
  4. All future transactions get timestamp > T.
```

**Visual:** Commit-wait ensures the commit timestamp is in the past before we commit.

```
  Timeline:  ──────────────────────────────────────────────────────>
             T (commit ts)     [earliest, latest]
                                ↑
                    Wait until earliest > T, then commit
```

#### Paxos-Based Replication

```
Each shard (tablet) = replicated by Paxos group
  - Leader handles reads/writes
  - 2f+1 replicas for f failures
  - Paxos for log replication
```

#### Two-Phase Commit (Cross-Shard Transactions)

```
Transaction touches shards A, B, C (different Paxos groups):

  Coordinator (e.g., A's leader):
    Phase 1: Prepare — send to B, C leaders
    Phase 2: Commit — if all accept, commit with timestamp T
    Commit-wait: wait until TT.now().earliest > T
    Then commit at all participants
```

#### Schema: Semi-Relational, Interleaved Tables

```
Spanner is NOT a pure key-value store. It has:
  - Tables with rows and columns
  - Interleaved tables (child rows stored with parent)
  - Secondary indexes
  - SQL-like query language (F1 layer on top)
```

#### Read Types

```
1. Strong reads (current): 
   - Read latest committed data
   - May involve Paxos leader or lock

2. Stale reads:
   - Read at timestamp T (or "exact staleness")
   - No blocking, can read from replicas

3. Snapshot reads:
   - Read at timestamp T
   - Consistent view across shards
```

### Diagrams

#### TrueTime Uncertainty

```
  Time:  ──────────────────────────────────────────────────────>
         
  TT.now() = [earliest, latest]
  |<-------- 1-7 ms uncertainty -------->|
  
  We know "now" is somewhere in this interval.
  Commit-wait: wait until entire interval is past our commit timestamp.
```

#### Commit-Wait

```
  Transaction T1 gets commit timestamp T = 100
  
  TT.now() = [98, 105]  →  earliest=98 < 100 → WAIT
  TT.now() = [99, 106]  →  earliest=99 < 100 → WAIT
  TT.now() = [101, 108] →  earliest=101 > 100 → COMMIT
```

#### Cross-Shard Transaction (2PC)

```
  Coordinator (Shard A leader)
       │
       │  Phase 1: Prepare
       ├──────────────────> Shard B leader
       ├──────────────────> Shard C leader
       │
       │  Phase 2: All accept
       │  Choose timestamp T
       │  Commit-wait: TT.now().earliest > T
       │
       │  Phase 3: Commit
       ├──────────────────> B, C commit
       └──────────────────> A commits
```

#### Paxos Group

```
  Shard (tablet) replicated by Paxos:
  
  Leader ──> Replica 1
       └──> Replica 2
       └──> Replica 3 (2f+1 for f=1)
  
  Writes: Leader proposes, majority acks
  Reads: Leader or follower (for stale reads)
```

### Tradeoffs

| Pros | Cons |
|------|------|
| Strong consistency globally | Requires hardware (GPS, atomic clocks) |
| External consistency | Commit-wait adds latency (~7ms avg) |
| SQL-like, schema | Complex (Paxos + 2PC) |
| Scalable | Not all systems can deploy TrueTime |

### Impact & Real Systems

```
Google Spanner (internal) → Google Cloud Spanner (managed)

CockroachDB: Inspired by Spanner, uses NTP instead of atomic clocks
  - Larger uncertainty intervals → longer commit-wait
  - Or: HLC (Hybrid Logical Clocks) for some cases
```

### Interview Deep Dive

**"Explain TrueTime and why it matters"**

- TrueTime gives a bounded uncertainty interval for "now." We don't pretend to know exact time; we know a range.
- Commit-wait: wait until we're sure our commit timestamp is in the past. Then commit. This guarantees external consistency.
- Without TrueTime: we can't order transactions across datacenters. With it: we can run global transactions with strong consistency.
- Why hardware? NTP alone can't give tight enough bounds. GPS + atomic clocks give ~1–7ms uncertainty.

### Summary

Spanner proves you can have globally distributed transactions with strong consistency by solving the clock problem. TrueTime's bounded uncertainty + commit-wait enables external consistency. Paxos per shard + 2PC across shards complete the architecture.

---

## 3. Meta XFaaS: Hyperscale and Low Cost Serverless Functions (2023)

### Definition

**"XFaaS: Hyperscale and Low Cost Serverless Functions at Meta"** — A 2023 SOSP paper describing Meta's platform for running serverless functions at hyperscale (trillions of function calls per day) with high cost efficiency.

### Problem It Solves

```
Run serverless functions at hyperscale:
  - Trillions of function calls per day
  - 100,000+ servers in Meta's private cloud
  - Must be cost-efficient (high CPU utilization)
  - Must avoid cold start latency

Traditional FaaS (AWS Lambda, etc.): 
  - Cold starts kill latency
  - Low utilization (over-provisioned for spikes)
  - Per-invocation overhead
```

### How It Works

#### Key Innovations

```
┌─── XFaaS Optimizations ──────────────────────────────────────┐
│                                                                 │
│  1. Locality-Aware Scheduling                                  │
│     → Schedule functions near data, reduce cold starts         │
│                                                                 │
│  2. Worker Pooling & Warm Container Reuse                      │
│     → Keep containers warm; reuse across invocations           │
│                                                                 │
│  3. Overcommit-Based Resource Management                       │
│     → Overcommit CPU/memory; higher utilization                │
│                                                                 │
│  4. Adaptive Concurrency Control                               │
│     → TCP-like congestion control; pace execution              │
│                                                                 │
│  5. Function Fusion                                            │
│     → Batch small functions; reduce per-call overhead           │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

#### Cold Start Elimination

```
Goal: Approximate "every worker can execute every function immediately"

Techniques:
  - Pre-warm containers for popular functions
  - Locality-aware scheduling: place function near where it's called
  - Worker pooling: shared pool of warm workers
  - Predictive warming based on traffic patterns
```

#### Load Spike Handling

```
  - Defer delay-tolerant functions to off-peak hours
  - Global dispatch: route function calls across datacenter regions
  - Avoid over-provisioning for peak spikes
```

#### Congestion Control

```
  - TCP-like mechanism: pace function execution
  - Prevent functions from overloading downstream services
  - Adaptive concurrency: throttle when backpressure detected
```

### Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                     XFaaS Architecture                            │
│                                                                 │
│  ┌────────┐    ┌──────────────┐    ┌─────────────────────────┐ │
│  │ Client │───>│  Dispatcher  │───>│  Locality-Aware         │ │
│  │ Apps   │    │  (routing)   │    │  Scheduler              │ │
│  └────────┘    └──────────────┘    └────────────┬────────────┘ │
│                                                 │              │
│                                                 ▼              │
│  ┌──────────────────────────────────────────────────────────────┐
│  │              Worker Pool (warm containers)                    │
│  │  ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐           │
│  │  │ W1  │ │ W2  │ │ W3  │ │ W4  │ │ W5  │ │ ... │           │
│  │  └─────┘ └─────┘ └─────┘ └─────┘ └─────┘ └─────┘           │
│  │  Overcommit: more "logical" workers than physical cores     │
│  └──────────────────────────────────────────────────────────────┘
│                                                                 │
│  ┌──────────────────────────────────────────────────────────────┐
│  │  Adaptive Concurrency Control                                 │
│  │  (pace execution, prevent downstream overload)                │
│  └──────────────────────────────────────────────────────────────┘
└─────────────────────────────────────────────────────────────────┘
```

#### Scheduling Flow

```
  Request arrives
       │
       ▼
  Dispatcher ──> Route to region/datacenter (locality)
       │
       ▼
  Scheduler ──> Worker pool lookup (warm container?)
       │
       ├─ HIT:  Execute immediately on warm worker
       │
       └─ MISS: Allocate new worker (or wait for one)
                    │
                    ▼
              Congestion control: throttle if overloaded
```

### Comparison with Public Cloud FaaS

| Aspect | AWS Lambda | Azure Functions | Google Cloud Functions | XFaaS (Meta) |
|--------|------------|-----------------|------------------------|--------------|
| Scale | Millions/sec | Similar | Similar | Trillions/day |
| Cold start | 100ms–10s | Similar | Similar | Near-eliminated |
| Utilization | Low (per-user isolation) | Low | Low | ~66% avg |
| Overcommit | No | No | No | Yes |
| Congestion control | Limited | Limited | Limited | TCP-like |

### Key Metrics from Paper

```
  - Trillions of function calls per day
  - 100,000+ servers
  - Daily average CPU utilization: ~66%
  - Cold start: effectively eliminated for common workloads
  - Cost: several times more efficient than typical FaaS
```

### Tradeoffs

| Pros | Cons |
|------|------|
| Hyperscale (trillions/day) | Meta-specific; not open source |
| High utilization (66%) | Overcommit can cause contention |
| Cold start elimination | Complex scheduling logic |
| Cost-efficient | Requires tight integration with Meta infra |

### Interview Deep Dive

**"Serverless at scale challenges"**

- **Cold starts:** Latency spike when no warm container. Mitigate: pre-warm, locality-aware scheduling, worker pooling.
- **Utilization:** Per-isolation model = low utilization. Mitigate: overcommit, shared workers, batch small functions.
- **Spikes:** Over-provision for peak = waste. Mitigate: defer delay-tolerant work, global load balancing.
- **Downstream overload:** Functions can overwhelm DBs. Mitigate: adaptive concurrency, TCP-like pacing.

### Summary

XFaaS achieves hyperscale serverless by locality-aware scheduling, worker pooling, overcommit, adaptive concurrency, and function fusion. Key result: ~66% CPU utilization and near-eliminated cold starts at trillions of function calls per day.

---

## 4. How to Read a Systems Paper

### Structure of a Typical Systems Paper

```
1. Abstract
   - One paragraph: problem, approach, results

2. Introduction
   - Motivation (why does this matter?)
   - Problem statement
   - Key contributions (numbered list)

3. Background / Related Work
   - What existed before
   - Why it's insufficient

4. Design

   - System overview
   - Key components (each with subsection)
   - Algorithms (pseudocode, diagrams)
   - Failure handling

5. Implementation
   - Practical details (optional in some papers)

6. Evaluation
   - Workloads
   - Metrics
   - Comparison with baselines
   - Experiments (scaling, failure injection)

7. Discussion / Limitations
   - What doesn't work
   - Future work

8. Conclusion
   - Summary of contributions
```

### What to Focus On

```
┌─── Reading Strategy ──────────────────────────────────────────┐
│                                                                 │
│  Problem → What pain does this solve?                          │
│  Design  → Core idea (1–3 sentences)                            │
│  Evaluation → Do the numbers back the claims?                   │
│  Limitations → What are the tradeoffs?                         │
│                                                                 │
│  Skip on first pass:                                            │
│  - Lengthy related work                                         │
│  - Implementation minutiae                                      │
│  - Every experiment (focus on main ones)                       │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Common Patterns Across Papers

```
  - Problem → Solution → Evaluation
  - "Existing X does Y, but fails at Z. We do W."
  - Tradeoffs: almost always a table or diagram
  - Scale: papers often show "10x" or "100x" improvement
  - Failure injection: how does it behave when things break?
```

### Building Intuition for System Design

```
1. Extract the core technique (e.g., "TrueTime" from Spanner)
2. Relate to CAP/PACELC (where does it sit?)
3. Identify the tradeoff (what did they give up?)
4. Map to real systems (which products use this?)
5. Practice: "How would you design X?" — borrow techniques
```

### Recommended Papers Beyond These Three

```
┌─── Foundational Reading ───────────────────────────────────────┐
│                                                                 │
│  Consensus & Replication:                                        │
│  - "In Search of an Understandable Consensus Algorithm" (Raft)  │
│  - Paxos Made Simple (Lamport)                                  │
│                                                                 │
│  Distributed Storage:                                           │
│  - "The Google File System" (GFS)                               │
│  - "Bigtable: A Distributed Storage System"                    │
│  - "MapReduce: Simplified Data Processing"                     │
│                                                                 │
│  Messaging & Streaming:                                         │
│  - "Kafka: A Distributed Messaging System" (LinkedIn)          │
│                                                                 │
│  More Advanced:                                                  │
│  - "Chubby: Google's Lock Service"                              │
│  - "F1: A Distributed SQL Database" (Spanner's SQL layer)       │
│  - "MillWheel: Fault-Tolerant Stream Processing"               │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Interview Tip

"When discussing a system design, reference papers: 'Similar to Dynamo's approach, I'd use consistent hashing with virtual nodes for distribution.' Or: 'For global consistency, we'd need something like Spanner's TrueTime — hardware clocks with bounded uncertainty.'"

### Summary

Systems papers follow a consistent structure: Problem → Design → Evaluation → Limitations. Focus on the core idea, the tradeoff, and how it maps to real systems. Build a mental library of techniques from papers like Dynamo, Spanner, GFS, Raft, and MapReduce.
