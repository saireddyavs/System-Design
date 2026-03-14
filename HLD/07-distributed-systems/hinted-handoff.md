# Hinted Handoff

> Staff+ Engineer Level — FAANG Interview Deep Dive

---

## 1. Concept Overview

### Definition

**Hinted Handoff** (or **hinted handoff**) is a technique in eventually consistent distributed storage systems where writes intended for an **unavailable node** are temporarily stored on another node. When the target node recovers, the "hint" is replayed to deliver the missed writes, maintaining availability during failures.

### Purpose

- **Availability during failures**: Accept writes even when replica nodes are down
- **Eventual consistency**: Ensure all replicas eventually receive all writes
- **No write rejection**: Clients get success without waiting for all replicas

### Problems Solved

| Problem | Solution |
|---------|----------|
| Node unavailable | Store write on alternate node; replay later |
| Write unavailability | Don't fail writes; use hinted replica |
| Data loss risk | Hints ensure no write is lost during outage |
| Client latency | Return success after W replicas (including hints) |

---

## 2. Real-World Motivation

### Amazon Dynamo (Original Paper)

- **Source** of hinted handoff concept (2007)
- Used in DynamoDB's predecessor
- Key technique for "always writable" design

### Apache Cassandra

- **Hinted handoff** for writes when replica node is down
- Hints stored on coordinator or other replicas
- Replay when node recovers (or manually trigger)
- Configurable hint window (default 3 hours)

### Riak (Basho)

- **Hinted handoff** as core availability mechanism
- Fallback chain: primary → secondary → tertiary
- Handoff queue per unavailable node

### Amazon DynamoDB

- Internal use of similar techniques
- Multi-AZ replication with automatic failover
- Hinted handoff concepts in replication layer

---

## 3. Architecture Diagrams

### Normal Write Flow (All Replicas Up)

```
    Client                    Coordinator              Replicas
       │                           │                       │
       │  PUT key=K, value=V       │                       │
       │──────────────────────────>│                       │
       │                           │  Replicate to N nodes │
       │                           │──────────────────────>│
       │                           │     N1    N2    N3   │
       │                           │      ✓     ✓     ✓   │
       │                           │<──────────────────────│
       │  200 OK (W=2 ack)         │                       │
       │<──────────────────────────│                       │
```

### Write Flow When Node Is Down — Hinted Handoff

```
    Client                    Coordinator              Replicas
       │                           │                       │
       │  PUT key=K, value=V       │                       │
       │──────────────────────────>│                       │
       │                           │  Replicate            │
       │                           │──────────────────────>│
       │                           │     N1    N2   N3(DOWN)│
       │                           │      ✓     ✓    X     │
       │                           │                       │
       │                           │  Store HINT on N2     │
       │                           │  (for N3 when back)   │
       │                           │  ┌─────────────────┐  │
       │                           │  │ Hint: N3, K, V  │  │
       │                           │  └─────────────────┘  │
       │                           │                       │
       │  200 OK (W=2 satisfied)   │                       │
       │<──────────────────────────│                       │
```

### Handoff on Recovery

```
    N3 comes back online
         │
         │  N2 (or coordinator) detects N3 is alive
         │
         ▼
    ┌─────────────────────────────────────────────────────┐
    │  Replay hints for N3                                 │
    │  For each hint (key, value) intended for N3:        │
    │    - Send write to N3                                │
    │    - Delete hint after success                       │
    └─────────────────────────────────────────────────────┘
         │
         ▼
    N3 now has all missed writes
    Hints cleared from N2
```

### Hint Storage Structure

```
    Node N2 (holding hints for N3)
    ┌─────────────────────────────────────────────┐
    │  Hinted Handoff Queue (for N3)               │
    │  ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐           │
    │  │ K1  │ │ K2  │ │ K3  │ │ K4  │  ...      │
    │  │ V1  │ │ V2  │ │ V3  │ │ V4  │           │
    │  │ ts  │ │ ts  │ │ ts  │ │ ts  │           │
    │  └─────┘ └─────┘ └─────┘ └─────┘           │
    │  Replay in order when N3 recovers           │
    └─────────────────────────────────────────────┘
```

### Consistency Level Interaction

```
    Write: CL=QUORUM, N=3, W=2
    Replicas: N1, N2, N3 (N3 down)

    Option A (no hinted handoff):
      - Only N1, N2 receive → W=2 ✓ → Success
      - N3 misses write when it comes back

    Option B (with hinted handoff):
      - N1, N2 receive
      - Hint stored for N3 on N1 or N2
      - W=2 ✓ → Success
      - When N3 recovers: replay hint → N3 has data
```

---

## 4. Core Mechanics

### Write Path with Hinted Handoff

1. **Coordinator** receives write for key K (replicas: N1, N2, N3)
2. **Send** to all replicas; N1 and N2 ack; N3 is down
3. **Store hint** on N1 or N2: (target=N3, key=K, value=V, timestamp)
4. **Return success** to client (if W replicas ack, including "hint" as ack in some implementations)
5. **Background**: Periodically check if N3 is up; replay hints

### Hint Replay

- **Trigger**: Node recovery detected (gossip, heartbeat)
- **Process**: For each hint targeting recovered node, send write
- **Ordering**: Typically by timestamp to preserve causality
- **Cleanup**: Delete hint after successful delivery

### Hint Window (Cassandra)

- **Max hint duration**: Default 3 hours
- **Reason**: Prevent unbounded hint growth; old hints may conflict with newer writes
- **Configurable**: `max_hint_window_in_ms`

### Who Stores Hints?

| System | Hint Storage |
|--------|--------------|
| Cassandra | Coordinator or another replica in same DC |
| Riak | Next node in preference list |
| Dynamo | Any healthy node in ring |

---

## 5. Numbers

| Metric | Typical Value |
|--------|---------------|
| Cassandra hint window | 3 hours (default) |
| Hint replay throttle | Configurable (e.g., 100K hints/sec) |
| Hint storage | On disk (persistent across restarts) |
| Hint size | Same as write (key + value + metadata) |

---

## 6. Tradeoffs

### Hinted Handoff vs Read Repair

| Aspect | Hinted Handoff | Read Repair |
|--------|----------------|-------------|
| **When** | During write (node down) | During read (stale detected) |
| **Proactive** | Yes (store for later) | Reactive (fix on read) |
| **Write path** | Affects write latency | No write path impact |
| **Coverage** | Only keys written during outage | Any divergent key |

### Hinted Handoff vs Anti-Entropy (Merkle Trees)

| Aspect | Hinted Handoff | Anti-Entropy |
|--------|----------------|--------------|
| **Scope** | Writes during outage | Full replica comparison |
| **Trigger** | Node recovery | Scheduled/background |
| **Efficiency** | Only missed writes | Scans entire dataset |
| **Use together** | Yes — hints for recent; Merkle for full sync |

### Hinted Handoff vs Quorum Write

| Aspect | Hinted Handoff | Strict Quorum |
|--------|----------------|---------------|
| **Node down** | Accept write; hint | Reject if < W replicas |
| **Availability** | Higher | Lower |
| **Consistency** | Eventual | Stronger (when all up) |

---

## 7. Variants / Implementations

### Cassandra

- **hinted_handoff_enabled**: Config option
- **Hint storage**: `hints` directory per node
- **Replay**: `nodetool repair` or automatic on recovery
- **Throttle**: `hinted_handoff_throttle_in_kb`

### Riak

- **Fallback chain**: Primary, secondary, tertiary nodes
- **Handoff queue**: Per unavailable node
- **Transfer**: When node returns, transfer hinted data

### Dynamo (Paper)

- **Sloppy quorum**: W nodes may include "proxy" nodes holding hints
- **Hinted handoff**: Part of "always writable" design

---

## 8. Scaling Strategies

1. **Hint throttling**: Limit replay rate to avoid overwhelming recovered node
2. **Hint compaction**: Merge/compress hints for same key
3. **Hint expiry**: Drop hints older than window (Cassandra)
4. **Parallel replay**: Replay from multiple holders if hints distributed

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|------------|
| Node down > hint window | Hints expire; data loss for that node | Reduce window or use read repair |
| Hint holder crashes | Hints lost | Replicate hints (Cassandra: single holder) |
| Replay storm | Recovered node overloaded | Throttle replay rate |
| Conflicting hints | Same key, different values | Use timestamps; last-writer-wins or vector clocks |
| Multiple nodes down | Many hints; storage pressure | Limit hints per node; prioritize |

---

## 10. Performance Considerations

- **Write path**: Storing hint adds minimal latency (async in many impls)
- **Disk**: Hints consume disk; monitor `hints` directory size
- **Replay**: Can cause load spike when node recovers; throttle
- **Network**: Replay generates write traffic

---

## 11. Use Cases

| Use Case | Why Hinted Handoff |
|----------|---------------------|
| Cassandra cluster | Node maintenance; brief outages |
| Riak KV | High availability; always accept writes |
| DynamoDB-style systems | Multi-AZ; transparent failover |
| Any AP system | Improve availability during failures |

---

## 12. Comparison Tables

### Replication Repair Mechanisms

| Mechanism | When | What | Proactive? |
|-----------|------|------|------------|
| **Hinted Handoff** | Write (node down) | Store for replay | Yes |
| **Read Repair** | Read (stale) | Push correct value | Reactive |
| **Anti-Entropy** | Background | Merkle tree sync | Yes |
| **Manual Repair** | Admin | Full repair | On-demand |

### Consistency Mechanisms Together

```
    Write path:     Hinted Handoff (node down)
    Read path:      Read Repair (detect stale)
    Background:     Anti-Entropy / Merkle (full sync)
    All three:      Used together in Cassandra, Riak
```

---

## 13. Code / Pseudocode

### Write with Hinted Handoff

```python
def put(key, value, replicas, W):
    acks = 0
    hints = []
    for node in replicas:
        try:
            node.send(Put(key, value))
            acks += 1
        except NodeUnavailable:
            # Store hint on self or another healthy node
            hint = Hint(target=node, key=key, value=value, ts=now())
            store_hint(hint)
            acks += 1  # Count hint as ack (sloppy quorum)
    
    if acks >= W:
        return Success
    return InsufficientReplicas
```

### Hint Replay

```python
def replay_hints_for_node(recovered_node):
    hints = get_hints_for_target(recovered_node)
    for hint in sorted(hints, key=lambda h: h.timestamp):
        try:
            recovered_node.send(Put(hint.key, hint.value))
            delete_hint(hint)
        except Exception:
            # Retry later
            break
```

### Hint Storage (Simplified)

```python
class HintStore:
    def __init__(self, max_window_sec=10800):  # 3 hours
        self.hints = {}  # target_node -> [(key, value, ts), ...]
        self.max_window = max_window_sec

    def add_hint(self, target, key, value):
        ts = time.time()
        if target not in self.hints:
            self.hints[target] = []
        self.hints[target].append((key, value, ts))
        self._evict_expired()

    def _evict_expired(self):
        cutoff = time.time() - self.max_window
        for target in self.hints:
            self.hints[target] = [(k, v, t) for k, v, t in self.hints[target] if t > cutoff]
```

---

## 14. Interview Discussion

### Key Points

1. **Hinted handoff = store writes for unavailable node, replay on recovery**
2. **Availability**: Accept writes even when replica is down
3. **Works with read repair and anti-entropy** — all three used together
4. **Cassandra, Riak, Dynamo** — from Amazon Dynamo paper

### Common Questions

- **"What is hinted handoff?"** — When a replica node is down, we store the write (hint) on another node. When the node recovers, we replay the hint so it gets the missed writes.
- **"Why not just reject the write?"** — To maintain availability. Clients get success; data eventually reaches all replicas.
- **"Hinted handoff vs read repair?"** — Hinted: proactive during write. Read repair: reactive when we detect stale data on read.
- **"What if the hint holder crashes?"** — Hints can be lost. Cassandra stores on one node; some systems replicate hints. Read repair and anti-entropy are backups.

### Red Flags

- Assuming hinted handoff alone ensures consistency (need read repair, anti-entropy)
- Ignoring hint window (old hints may conflict)
- No replay throttling (can overwhelm recovered node)
