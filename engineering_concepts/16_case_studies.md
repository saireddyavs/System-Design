# Module 16: Famous Case Studies

---

## 1. Amazon Prime Video: Microservices → Monolith

### The Story
Amazon Prime Video's video quality monitoring service was built as serverless microservices (AWS Step Functions + Lambda). They moved it BACK to a monolith.

### Why They Changed
```
Original Architecture (Microservices):
  Step Functions orchestrator
    → Lambda: extract frames
    → S3: store frames
    → Lambda: analyze quality
    → Lambda: aggregate results

Problems:
  - Step Functions: $0.025 per 1000 state transitions → massive cost
  - S3 round-trips for every frame → high latency
  - Lambda cold starts → inconsistent performance
  - Data passed between services via S3 → slow
```

### New Architecture (Monolith)
```
Single ECS service:
  Extract frames → analyze → aggregate
  All in-memory, no S3 round-trips
  No Step Functions orchestration cost

Result: 90% cost reduction
```

### Lesson
"Microservices add network overhead. If components are 'chatty' (high data exchange), put them in the same process. Microservices excel when teams work independently with minimal data coupling."

---

## 2. Facebook Haystack: Billions of Photos

### The Problem
```
Standard filesystem (POSIX):
  Each photo = 1 file
  Reading 1 photo = 3 disk seeks:
    1. Read directory inode
    2. Read file inode
    3. Read file data
  At 1 billion photos: inode metadata doesn't fit in RAM
```

### The Solution: Haystack
```
Haystack Architecture:
  Pack thousands of photos into large "Haystack files" (multi-GB)

  ┌─────────────────────────────────────────┐
  │  Haystack File (100GB)                   │
  │  [Photo1][Photo2][Photo3]...[Photo_N]    │
  │  Each photo: header + data               │
  └─────────────────────────────────────────┘

  In-memory index:
    photo_id → (haystack_file, offset, size)

  Read = 1 disk seek:
    lookup offset in memory → seek to position → read photo
```

### Impact
- Reduced disk seeks from 3 to 1 per photo
- All metadata in RAM (tiny per photo: ~40 bytes)
- Eliminated filesystem overhead entirely

---

## 3. Twitter Timeline Fanout

### The Problem
```
Justin Bieber tweets. 100 million followers need to see it.

PULL approach (Fan-out on Read):
  User opens app → query "get latest tweets from everyone I follow"
  → Scan millions of rows → SLOW for read (especially heavy followers)

PUSH approach (Fan-out on Write):
  Justin tweets → Insert into 100 million Redis timelines
  → 100 MILLION writes per tweet → SLOW for write!
```

### The Hybrid Solution
```
For normal users (< 10K followers):
  PUSH: Fan-out on write. Insert tweet into followers' timelines.
  Pre-computed → instant read.

For celebrities (> 10K followers):
  PULL: Don't fan out. When follower reads timeline:
  Timeline = pre-computed (pushed tweets) + merge (pull celebrity tweets)

┌────────┐  tweet   ┌──────────┐   push    ┌─────────────────┐
│ Normal │ ───────→ │ Fanout   │ ────────→ │ Follower Redis  │
│ User   │          │ Service  │            │ Timelines       │
└────────┘          └──────────┘            └────────┬────────┘
                                                     │ merge
┌────────┐  tweet   ┌──────────┐   pull     ┌───────▼────────┐
│Celebrity│ ───────→│ Celebrity │ ──on read──│  Timeline API  │
│         │         │ Tweet DB  │            └────────────────┘
└────────┘          └──────────┘
```

### Interview Tip
"For a social media timeline, I'd use a hybrid fanout: push for regular users (pre-computed), pull for celebrities (computed at read time), merged at query time."

---

## 4. WhatsApp: 2 Million Connections per Server

### The Secret
```
Stack: Erlang/OTP on FreeBSD

Key factors:
  1. Erlang processes (lightweight actors, 2KB each)
     → 2 million processes = ~4GB RAM

  2. OS tuning:
     - Increased file descriptor limit to 10M+
     - Tuned TCP buffer sizes
     - Custom FreeBSD kernel tweaks

  3. Minimalist approach:
     - No unnecessary features
     - Simple protocol
     - ~50 engineers for 1 billion users

  4. Erlang's "Let it crash" philosophy:
     - Processes crash and restart in microseconds
     - Supervisor trees manage recovery
     - No defensive programming bloat
```

### Result
900 million users, ~50 engineers, ~$0.01 per user per year. Acquired by Facebook for $19 billion.

---

## 5. Discord: Go → Rust (GC Pauses)

### The Problem
```
Discord's "Read States" service (tracks what each user has read):
  - Millions of users
  - Millions of updates per second
  - Stored in-memory

Go's garbage collector:
  - Every ~2 minutes: GC pause
  - During pause: latency spikes from 1ms → 300ms
  - Users experience lag

  Latency Graph:
  ────────╱╲────────╱╲────────  ← regular GC spikes
          ↑ GC      ↑ GC
```

### The Solution
```
Rewrote critical service in Rust:
  - No garbage collector
  - Manual memory management (ownership system)
  - Zero GC pauses
  - Consistent low latency

Result:
  Go:   avg 1ms, p99 300ms (GC spikes)
  Rust: avg 0.3ms, p99 1ms (flat, no spikes)
```

### Instagram's Approach (Different)
```
Instagram disabled Python's CYCLIC garbage collector:
  - Their web server creates and destroys objects per request
  - No long-lived cycles form between requests
  - Disabling cyclic GC saved ~10% memory

They kept reference counting (for non-cyclic garbage).
```

### Lesson
Managed languages (Go, Java, Python) have a latency ceiling imposed by GC. For latency-critical paths, Rust/C++ eliminate this ceiling.

---

## 6. Amazon Dynamo Paper (2007)

### Legacy
The paper that launched the NoSQL revolution. Introduced key concepts:

```
┌─── Techniques from Dynamo Paper ─────────────────┐
│                                                    │
│  Consistent Hashing      → data distribution      │
│  Vector Clocks           → conflict detection      │
│  Sloppy Quorum           → availability            │
│  Hinted Handoff          → failure recovery        │
│  Merkle Trees            → anti-entropy/repair     │
│  Gossip Protocol         → membership detection    │
│                                                    │
│  Design choice: ALWAYS WRITABLE                    │
│  "Shopping cart must never reject a write"          │
│  Conflicts resolved by client on read              │
└────────────────────────────────────────────────────┘
```

### Influence
DynamoDB, Cassandra, Riak, Voldemort — all inspired by this paper.

---

## 7. Google Spanner (2012)

### The Breakthrough
Proved you CAN have distributed transactions with strong consistency at global scale.

### How TrueTime Works
```
Every Google datacenter has:
  GPS receivers + Atomic clocks

TrueTime API returns:
  TT.now() = [earliest, latest]  (uncertainty interval)

Commit rule:
  Transaction gets timestamp T
  WAIT until TT.now().earliest > T
  Then commit is guaranteed to be in the past
  All future transactions will have timestamp > T

This achieves EXTERNAL CONSISTENCY:
  If transaction T1 commits before T2 starts,
  then T1's timestamp < T2's timestamp. GUARANTEED.
```

### Impact
```
Without TrueTime: Cannot order transactions across continents
With TrueTime: Global transactions with ~7ms average wait

Spanner proves: You CAN have C + A (practically) if you solve
the clock synchronization problem.
```

### Real Systems
Google Cloud Spanner, CockroachDB (uses NTP instead of atomic clocks, with larger uncertainty intervals)

---

## 8. MapReduce → Spark → Flink

### Evolution of Big Data Processing
```
┌─── MapReduce (2004, Google/Hadoop) ──────────────┐
│ Map: Process data in parallel                     │
│ Shuffle: Group by key                             │
│ Reduce: Aggregate                                 │
│                                                    │
│ Problem: Writes intermediate results to DISK       │
│ between every step. Slow for iterative algorithms. │
└──────────────────────────────────────────────────┘
         ↓ evolution

┌─── Spark (2012, UC Berkeley) ────────────────────┐
│ Keeps data in MEMORY between steps (RDDs)         │
│ 10-100x faster than MapReduce for iterative work  │
│                                                    │
│ Micro-batch streaming (not true real-time)         │
│ Great for: ML training, ETL, batch analytics       │
└──────────────────────────────────────────────────┘
         ↓ evolution

┌─── Flink (2014, Berlin) ────────────────────────┐
│ TRUE streaming: processes events ONE AT A TIME    │
│ (not micro-batch like Spark)                      │
│                                                    │
│ Event time processing (handles out-of-order data) │
│ Exactly-once semantics via checkpointing          │
│ Great for: real-time analytics, CEP, fraud detect  │
└──────────────────────────────────────────────────┘
```

### Comparison

| | MapReduce | Spark | Flink |
|-|-----------|-------|-------|
| Processing | Batch only | Batch + micro-batch | Batch + true stream |
| Speed | Slow (disk I/O) | Fast (in-memory) | Fastest (streaming) |
| Latency | Minutes-hours | Seconds | Milliseconds |
| State | Stateless | RDD lineage | Managed state + checkpoints |
| Use case | ETL, large batch | ML, batch + "near real-time" | Real-time, event-driven |

---

## 9. The CAP Theorem Proof

### Brewer's Theorem (2000, proven 2002)
```
In a distributed system, during a network PARTITION,
you can guarantee either:
  - Consistency (all nodes see same data)
  - Availability (every request gets a response)
But NOT both.

Proof sketch:
  Two nodes A and B, partitioned (can't communicate).

  Client writes x=1 to A.
  Client reads x from B.

  Option 1 (Consistency): B refuses to answer (no availability)
           because it can't verify it has latest value.

  Option 2 (Availability): B returns stale x=0 (no consistency)
           because it hasn't received the update from A.

  Can't have both during partition. QED.
```

### Real-World Interpretation
```
Partitions ALWAYS happen (network is unreliable).
So the real choice is: during failures, do you want...

CP (Consistency + Partition tolerance):
  Refuse requests during partition.
  Banks, inventory: "Better to reject than give wrong answer"
  Systems: HBase, MongoDB, etcd, ZooKeeper

AP (Availability + Partition tolerance):
  Serve potentially stale data during partition.
  Social media, DNS: "Better to show old data than nothing"
  Systems: Cassandra, DynamoDB, CouchDB

In practice: Tunable per-operation, not a binary system-wide choice.
```

---

## 10. Data Locality & Hot Partitions

### Data Locality
```
Moving computation TO the data is cheaper than moving data TO computation.

MapReduce: Schedule map task on the node that has the data block.
Spark: Prefer executors on the same rack as the data.
CDN: Cache content at the edge nearest to users.

Anti-pattern: Shuffle phase in MapReduce moves data across network.
             → Minimize shuffles for performance.
```

### Hot Partition Problem
```
Sharded database, hash-based sharding:
  99% of queries go to key "Justin_Bieber"
  → That shard is HOT (overloaded)
  → Other shards idle

Solutions:
  1. Add random suffix: "Justin_Bieber_1", "Justin_Bieber_2", ...
     → Spread across shards. Aggregate on read.

  2. Dedicated shard for hot keys (special handling)

  3. Caching layer in front (absorb read heat)

  4. Rate limiting per key

Celebrity problem = Hot partition problem in EVERY system design interview.
```

### Interview Tip
"For hot partitions, I'd add a random suffix to the key (e.g., key_0 through key_99) to spread writes across 100 partitions, then aggregate on read. For reads, a cache layer absorbs the heat."
