# Clocks & Ordering in Distributed Systems

> Staff+ Engineer Level — FAANG Interview Deep Dive

---

## 1. Concept Overview

### Definition

**Clocks and ordering** mechanisms allow distributed systems to reason about the **order of events** across nodes that have no shared memory or global clock. They enable causal ordering, conflict detection, and consistency guarantees.

### Purpose

- **Ordering**: Determine which event happened before another
- **Causality**: Preserve cause-effect relationships
- **Conflict detection**: Identify concurrent (unordered) events
- **Timestamping**: Assign unique, comparable identifiers to events

### Problems Solved

| Problem | Solution |
|---------|----------|
| "Which write is newer?" | Timestamps, version vectors |
| Concurrent updates | Vector clocks detect concurrency |
| Out-of-order delivery | Logical clocks order events |
| Clock skew | NTP, TrueTime, HLC |
| Distributed transactions | Timestamps for serialization |

---

## 2. Real-World Motivation

### Google Spanner

- **TrueTime**: GPS + atomic clocks; bounded uncertainty [earliest, latest]
- **External consistency**: If T1 commits before T2 starts, ts(T1) < ts(T2)
- **Commit wait**: Transaction waits until TT.after(commit_ts)

### CockroachDB

- **Hybrid Logical Clocks (HLC)**: Physical + logical component
- **No GPS**: Works without special hardware
- **Causal ordering**: HLC preserves happens-before

### DynamoDB

- **Vector clocks**: Detect concurrent writes; conflict resolution
- **Version vectors**: Per-item causality tracking

### Cassandra

- **Timestamp**: Per-write timestamp (client or server)
- **Last-writer-wins**: Conflict resolution by timestamp
- **Tombstones**: For deletes; timestamp-based expiry

### MongoDB

- **Logical clocks**: For causal consistency
- **Cluster time**: Monotonically increasing

---

## 3. Architecture Diagrams

### Physical Clock Drift

```
    Node A (fast)          Node B (slow)          Node C (accurate)
    |----|----|----|       |----|----|----|       |----|----|----|
    t=0  t=1  t=2  t=3     t=0  t=0.5 t=1  t=1.5   t=0  t=1  t=2  t=3
    
    Real time: 0 -------- 1 -------- 2 -------- 3
    Drift: A runs fast, B runs slow
    NTP tries to sync but has ~1-10ms error, worse across WAN
```

### Lamport Timestamp Propagation

```
    Process P1              Process P2              Process P3
    L1=0                    L2=0                    L3=0
       |                         |                         |
       | local event             |                         |
       | L1=1                     |                         |
       |---- send(m) ------------->|                         |
       |                     L2=max(0,1)+1=2                |
       |                     receive(m)                     |
       |                         |---- send(m) ------------->|
       |                         |                    L3=max(0,2)+1=3
       |                         |                    receive(m)
    
    Rule: on send, L = L+1. On receive, L = max(L, msg_ts)+1
```

### Vector Clock Example

```
    Process A    Process B    Process C    Event
    [1,0,0]      [0,1,0]      [0,0,1]      Initial
    [2,0,0]      [0,1,0]      [0,0,1]      A: local
    [2,1,0]      [0,1,0]      [0,0,1]      A receives from B
    [2,1,1]      [0,1,0]      [0,0,1]      A receives from C
    [3,1,1]      [0,1,0]      [0,0,1]      A: local
    
    Compare: V(a) < V(b) iff every component of V(a) <= V(b) and at least one strict
    Concurrent: neither V(a) <= V(b) nor V(b) <= V(a)
```

### Hybrid Logical Clock (HLC)

```
    HLC = (pt, l, c)
    pt = physical time (max of local clock, received pt)
    l  = logical component (incremented when pt unchanged)
    c  = counter (for same pt, l)
    
    On local event: pt = max(pt, local_clock); if pt unchanged, l++
    On receive(msg): pt = max(pt, msg.pt, local_clock)
                     if pt from msg or local, l = max(msg.l, l) + 1
                     else l = 0
    
    Invariant: HLC stays close to physical time (bounded drift)
```

### TrueTime (Spanner)

```
    TrueTime interval: [earliest, latest]
    - earliest: latest time we're sure we haven't passed
    - latest: earliest time we're sure we've passed
    
    Typical uncertainty: 1-7 ms
    
    Commit: assign timestamp s
    Commit wait: wait until TT.after(s) is true
    Then: release locks, respond to client
```

### Version Vector

```
    Replica A    Replica B    Replica C
    [3,0,0]     [0,2,0]      [0,0,1]
    
    After sync A-B: A has [3,2,0], B has [3,2,0]
    Detects: concurrent updates when vectors are incomparable
```

---

## 4. Core Mechanics

### Physical Clocks

- **Wall clock**: System clock (NTP-synced)
- **Drift**: ~1-10ms with NTP; worse across WAN
- **Skew**: Difference between two clocks at same instant
- **Problem**: Cannot order events across nodes reliably

### Lamport Timestamps

- **Single counter per process**
- **Rules**: On local event: L = L + 1. On send: L = L + 1, attach to message. On receive: L = max(L, msg_L) + 1
- **Property**: If a → b (happens-before), then L(a) < L(b)
- **Limitation**: L(a) < L(b) does NOT imply a → b (only one direction)

### Vector Clocks

- **Vector of N counters** (one per process)
- **Rules**: On local event: V[i]++. On send: V[i]++, attach V. On receive: V = max(V, msg_V) element-wise, then V[i]++
- **Compare**: V(a) ≤ V(b) iff every component of V(a) ≤ V(b)
- **Concurrent**: Neither V(a) ≤ V(b) nor V(b) ≤ V(a)
- **Property**: V(a) < V(b) iff a happens-before b

### Hybrid Logical Clocks (HLC)

- **Format**: (physical_time, logical, counter)
- **Goal**: Stay close to physical time; preserve causality
- **Used by**: CockroachDB, MongoDB
- **Advantage**: No special hardware; comparable to NTP-synced clocks

### TrueTime

- **Hardware**: GPS receivers + atomic clocks in each datacenter
- **Output**: Interval [earliest, latest]
- **TT.before(t)**: Now is definitely before t
- **TT.after(t)**: Now is definitely after t
- **Commit wait**: Ensures external consistency

### Version Vectors

- **Per-replica counters** (like vector clocks for replicas)
- **Used for**: Conflict detection in eventually consistent systems
- **Dynamo-style**: (replica_id, counter) pairs

---

## 5. Numbers

| Clock Type | Size | Precision | Sync Required |
|------------|------|-----------|---------------|
| Lamport | 1 counter | N/A | No |
| Vector | N counters | N/A | No |
| HLC | (pt, l, c) | ~ms | NTP |
| TrueTime | interval | 1-7ms | GPS+atomic |
| NTP | 64-bit | 1-10ms | Yes |

### Scale

- **Spanner**: TrueTime in every datacenter; 7ms typical uncertainty
- **CockroachDB**: HLC; works with NTP (ms-level)
- **DynamoDB**: Vector clock per item; O(replicas) size

---

## 6. Tradeoffs

### Lamport vs. Vector Clocks

| Aspect | Lamport | Vector |
|--------|---------|--------|
| Size | O(1) | O(N) |
| Concurrency | Cannot detect | Can detect |
| Causality | One direction | Full |
| Use case | Ordering | Conflict detection |

### TrueTime vs. HLC

| Aspect | TrueTime | HLC |
|--------|----------|-----|
| Hardware | GPS + atomic | None |
| Guarantee | External consistency | Causal |
| Drift | Bounded | Bounded by NTP |
| Cost | High | Low |

---

## 7. Variants / Implementations

### Logical Clocks (Leslie Lamport, 1978)

- Original Lamport timestamps
- Foundation for distributed ordering

### Vector Clocks (Fidge, Mattern, 1988-89)

- Independent discovery
- Used in Dynamo, Riak, Voldemort

### Hybrid Logical Clocks (Kulkarni et al., 2014)

- Combines physical + logical
- CockroachDB, MongoDB

### TrueTime (Google, 2012)

- Spanner paper
- Requires specialized hardware

### Interval Tree Clocks (ITC)

- Compact; can fork and merge
- Alternative to version vectors

---

## 8. Scaling Strategies

1. **Server-side timestamps**: Avoid client clock skew (Cassandra)
2. **Compression**: Version vectors can be compressed (dotted versions)
3. **Sharding**: Different shards can use independent clocks
4. **Hybrid**: Use physical when possible, logical when needed (HLC)

---

## 9. Failure Scenarios

| Scenario | Lamport | Vector | HLC | TrueTime |
|----------|---------|--------|-----|----------|
| Clock skew | OK | OK | Bounded | Bounded |
| Message loss | Order lost | Order lost | Order lost | N/A |
| Node crash | Restart counter | Restart vector | Restart HLC | N/A |
| GPS failure | N/A | N/A | N/A | Fallback to atomic |

---

## 10. Performance Considerations

- **Vector clock size**: Grows with replicas; can use dotted version vectors
- **Comparison cost**: O(N) for vector clocks
- **Storage**: Timestamps stored with data; affects storage
- **TrueTime**: Commit wait adds latency (up to 7ms)

---

## 11. Use Cases

| Use Case | Clock Type | Why |
|----------|------------|-----|
| Distributed locking | Lamport | Order lock requests |
| Conflict detection | Vector | DynamoDB, Riak |
| Distributed DB | HLC | CockroachDB |
| Global transactions | TrueTime | Spanner |
| Caching | Version vector | Invalidation |
| Event sourcing | Lamport/Vector | Order events |

---

## 12. Comparison Tables

### Clock Hierarchy

```
    Physical (NTP)
         |
    TrueTime (GPS+atomic) — strongest, external consistency
         |
    HLC — causal + close to physical
         |
    Vector Clocks — full causality, detect concurrent
         |
    Lamport — partial order (happens-before)
```

### When to Use What

| Requirement | Recommendation |
|--------------|----------------|
| Order events in single process | Lamport |
| Detect concurrent writes | Vector clocks |
| Distributed DB, no GPS | HLC |
| Global serializability | TrueTime |
| Conflict resolution | Version vectors |

---

## 13. Code or Pseudocode

### Lamport Timestamp

```python
class LamportClock:
    def __init__(self):
        self.time = 0
    
    def local_event(self):
        self.time += 1
        return self.time
    
    def send_message(self):
        self.time += 1
        return self.time  # Attach to message
    
    def receive_message(self, msg_timestamp):
        self.time = max(self.time, msg_timestamp) + 1
        return self.time
```

### Vector Clock

```python
class VectorClock:
    def __init__(self, node_id, num_nodes):
        self.node_id = node_id
        self.vector = [0] * num_nodes
    
    def local_event(self):
        self.vector[self.node_id] += 1
        return self.vector[:]
    
    def send_message(self):
        self.vector[self.node_id] += 1
        return self.vector[:]
    
    def receive_message(self, msg_vector):
        for i in range(len(self.vector)):
            self.vector[i] = max(self.vector[i], msg_vector[i])
        self.vector[self.node_id] += 1
        return self.vector[:]
    
    def happens_before(self, other):
        return all(self.vector[i] <= other[i] for i in range(len(self.vector))) and \
               any(self.vector[i] < other[i] for i in range(len(self.vector)))
    
    def concurrent(self, other):
        return not self.happens_before(other) and not self.happens_before_rev(other)
```

### Hybrid Logical Clock

```python
class HLC:
    def __init__(self):
        self.pt = 0  # physical time
        self.l = 0   # logical
    
    def local_event(self, physical_now):
        if physical_now > self.pt:
            self.pt = physical_now
            self.l = 0
        else:
            self.l += 1
        return (self.pt, self.l)
    
    def receive_message(self, msg_pt, msg_l, physical_now):
        if physical_now > self.pt and physical_now > msg_pt:
            self.pt = physical_now
            self.l = 0
        elif msg_pt > self.pt:
            self.pt = msg_pt
            self.l = msg_l + 1
        elif msg_pt == self.pt:
            self.l = max(self.l, msg_l) + 1
        else:
            self.l += 1
        return (self.pt, self.l)
```

### TrueTime Commit Wait (Pseudocode)

```python
def commit_transaction(txn):
    s = choose_timestamp(txn)  # e.g., TT.now().latest
    # Wait until we're sure real time has passed s
    while not TT.after(s):
        sleep(1)  # ms
    release_locks(txn)
    return ok
```

---

## 14. Interview Discussion

### Key Points

1. **Lamport**: Partial order; if a→b then L(a)<L(b); converse false
2. **Vector**: Full causality; can detect concurrent events
3. **HLC**: Physical + logical; no GPS; CockroachDB
4. **TrueTime**: GPS + atomic; Spanner; external consistency

### Common Questions

- **"Lamport vs. Vector?"** — Lamport: O(1), no concurrency detection. Vector: O(N), detects concurrent
- **"How does Spanner get external consistency?"** — TrueTime + commit wait
- **"What is HLC?"** — Hybrid Logical Clock; (pt, l); stays near physical time, preserves causality
- **"When are two events concurrent?"** — When neither happens-before the other; vector clocks detect this

### Red Flags

- Using physical clocks for ordering across nodes (skew)
- Confusing Lamport: L(a)<L(b) does NOT mean a caused b
- Ignoring clock skew in distributed systems

---

## 15. Deep Dive: Happens-Before Relation (Lamport 1978)

**Definition**: a → b (a happens before b) if:
1. a and b on same process and a precedes b, or
2. a is send of message m and b is receive of m, or
3. Transitive: a → c and c → b implies a → b

**Concurrent**: Neither a → b nor b → a. Cannot order.

**Lamport clocks**: If a → b then L(a) < L(b). Converse false: L(a) < L(b) does not imply a → b.

**Vector clocks**: V(a) < V(b) iff a → b. Full characterization.

---

## 16. Deep Dive: TrueTime API and Implementation

**API**:
- `TT.now()` → [earliest, latest]
- `TT.after(t)` → true if now > t (we're sure)
- `TT.before(t)` → true if now < t (we're sure)

**Hardware**: Each datacenter has GPS receivers and atomic clocks (cesium or rubidium). GPS gives ~100ns accuracy but can have outages. Atomic clocks drift ~1ms/day. Combine: when GPS available, sync. When not, use atomic. Uncertainty grows during GPS outage.

**Typical**: 1-7ms uncertainty. Commit wait adds up to 7ms latency.

---

## 17. Deep Dive: Version Vectors vs. Vector Clocks

**Vector clocks**: One entry per process. Used for causality between events.

**Version vectors**: One entry per replica. Used for causality between replica states. (Replica ≈ process in distributed DB.)

**Dotted version vector**: Optimized. Only store (replica, counter) for replicas that have updated. Saves space when few replicas active.

**DynamoDB**: Uses version vectors per item. Detect concurrent writes; return both to client for merge.

---

## 18. Logical Clocks for Distributed Debugging

**Use case**: Trace requests across services. Order log events.

**Trace ID**: Unique per request. Propagated in headers.

**Span**: Each service creates span with (trace_id, span_id, parent_span_id, timestamp). Timestamps can be local (Lamport) or physical. For debugging, physical is often used (approximate).

**Causality**: Parent span happens-before child. Lamport/vector can order spans across services.

---

## 19. Hybrid Logical Clock: Bounded Drift Property

**Invariant**: HLC never exceeds physical time by more than max_drift (configurable, e.g., 100ms).

**Why**: On receive, pt = max(local, msg_pt). If local is ahead, we use local. If msg is ahead, we use msg. Logical component only increments when pt unchanged. So HLC stays within max_drift of physical.

**Use**: CockroachDB uses HLC for transaction ordering. No commit wait; but causal consistency, not external.

---

## 20. Interview Walkthrough: Choosing a Clock

**Question**: "How would you order events in a distributed system?"

**Answer structure**:
1. **Single process**: Lamport (simple, O(1))
2. **Need to detect concurrent?**: Vector clocks (O(N) size)
3. **Distributed DB, want physical-ish timestamps?**: HLC (CockroachDB)
4. **Global serializability, have budget for hardware?**: TrueTime (Spanner)
5. **Conflict resolution in eventually consistent?**: Version vectors (DynamoDB)
6. **Avoid**: Raw physical clocks for ordering across nodes (skew)

---

## 21. Lamport Clock: Total Order Extension

**Problem**: Lamport gives partial order. Multiple events can have same timestamp (concurrent).

**Total order**: Append process ID: (L, process_id). Lexicographic order. (1, A) < (1, B) < (2, A). Unique total order.

**Use**: Distributed locking, logical timestamps for logs.

---

## 22. Vector Clock Size and Dotted Version Vectors

**Problem**: N replicas → N components. Can be large (1000s of replicas).

**Dotted version vector**: Only store (replica, counter) for replicas that have updated. Typically few. Saves space.

**Example**: Full [3,0,1,0,2] vs. dotted {(A:3), (C:1), (E:2)}. Same information, compact.

---

## 23. Spanner Commit Wait: Why It Works

**External consistency**: If T1 commits before T2 starts, ts(T1) < ts(T2).

**Commit**: Assign s = TT.now().latest (or similar). Replicate. Wait until TT.after(s). Then release locks.

**Why wait**: Before wait, we're not sure real time has passed s. After, we're sure. So any transaction that starts after our commit will get timestamp > s. Ordering preserved.

**Cost**: Up to 2*epsilon latency, where epsilon = clock uncertainty (~7ms).

---

## 24. NTP and Clock Sync

**NTP**: Network Time Protocol. Syncs to stratum 0 (atomic clock) or stratum 1 (connected to stratum 0). Typical accuracy: 1-10ms on LAN, worse on WAN.

**Drift**: Without NTP, clocks drift. ~1 second per day for cheap oscillators. NTP corrects periodically.

**Leap seconds**: Occasional 61-second minute. Can cause issues (e.g., 2012 leap second caused outages). Some systems use "smearing" (spread leap second over day).

---

## 25. Summary: Clock Selection Cheat Sheet

| Requirement | Clock | Example |
|-------------|-------|---------|
| Order in one process | Lamport | Logging |
| Detect concurrent | Vector | DynamoDB |
| Causal + physical-ish | HLC | CockroachDB |
| Global serializability | TrueTime | Spanner |
| Conflict resolution | Version vector | Riak |
| Avoid cross-node order | Physical | Don't use |
