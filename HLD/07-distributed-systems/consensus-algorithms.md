# Consensus Algorithms in Distributed Systems

> Staff+ Engineer Level — FAANG Interview Deep Dive

---

## 1. Concept Overview

### Definition

**Consensus** is the process by which a group of distributed nodes agrees on a single value (or sequence of values) despite failures. All correct nodes must decide on the same value, and that value must have been proposed by some node.

### Purpose

- **Replicated state machines**: Keep multiple copies of state identical
- **Leader election**: Agree on who is the leader
- **Distributed locking**: Agree on lock ownership
- **Configuration management**: Agree on cluster membership

### The Consensus Problem

Given N nodes, up to F of which may fail:

1. **Agreement**: All correct nodes decide on the same value
2. **Validity**: The decided value was proposed by some node
3. **Termination**: All correct nodes eventually decide (liveness)
4. **Integrity**: A node decides at most once

---

## 2. Real-World Motivation

### etcd (Kubernetes)

- Uses **Raft** for distributed key-value store
- Stores cluster state, service discovery, config
- Kubernetes API server uses etcd as backing store

### ZooKeeper

- Uses **ZAB** (ZooKeeper Atomic Broadcast)
- Coordination for Kafka, HBase, Hadoop
- Leader handles all writes; followers replicate

### Google Chubby

- Uses **Paxos** (Multi-Paxos)
- Lock service for GFS, Bigtable
- Used for leader election, configuration

### CockroachDB

- Uses **Raft** per range
- Distributed SQL with strong consistency
- Each key range has its own Raft group

### Amazon DynamoDB

- Uses **Paxos-like** for leader election
- Multi-AZ replication with quorum

---

## 3. Architecture Diagrams

### Paxos — Roles and Phases

```
    PROPOSER              ACCEPTORS              LEARNERS
        |                      |                      |
        |--- Phase 1a: PREPARE(n) ------------------>|
        |                      |                      |
        |<-- Phase 1b: PROMISE(n, n', v') -----------|
        |    (ignore if n <= n')                      |
        |                      |                      |
        |--- Phase 2a: ACCEPT(n, v) ----------------->|
        |    (v = highest v' or own value)            |
        |                      |                      |
        |<-- Phase 2b: ACCEPTED(n, v) ---------------|
        |                      |                      |
        |                      |---- LEARN(v) -------->|
```

### Raft — Leader Election

```
    Node A (follower)    Node B (follower)    Node C (follower)
         |                      |                      |
         |  timeout, become      |                      |
         |  candidate            |                      |
         |---------------- RequestVote(term=2) -------->|
         |<--------------------- vote -----------------|
         |---------------- RequestVote(term=2) -------->|
         |<--------------------- vote -----------------|
         |  majority: become LEADER                    |
         |---------------- AppendEntries (heartbeat) ->|
         |<--------------------- ack ------------------|
```

### Raft — Log Replication

```
    LEADER                    FOLLOWER 1              FOLLOWER 2
       |                           |                        |
       |  log: [1,2,3,4]           |  log: [1,2,3]          |  log: [1,2]
       |                           |                        |
       |  AppendEntries(4, term=2) |                        |
       |------------------------->|                        |
       |------------------------->| AppendEntries(4)       |
       |<-------- ack ------------|------------------------>|
       |<-------- ack --------------------------------------|
       |  commit index = 4        |                        |
```

### ZAB — ZooKeeper Flow

```
    LEADER                         FOLLOWERS
       |                               |
       |  PROPOSAL(1, "set /a 1")     |
       |----------------------------->|
       |  ACK                          |
       |<------------------------------|
       |  COMMIT(1)                     |
       |----------------------------->|
       |  (all apply to state machine) |
```

---

## 4. Core Mechanics

### Paxos

**Roles**: Proposer, Acceptor, Learner

**Phase 1 — Prepare/Promise**:
1. Proposer picks proposal number n, sends PREPARE(n) to acceptors
2. Acceptor: if n > highest n seen, promise not to accept any n' < n; reply with (n, highest accepted n', v')
3. Proposer: if majority promise, proceed to Phase 2

**Phase 2 — Accept/Accepted**:
1. Proposer sends ACCEPT(n, v) where v = highest v' from promises (or own value if none)
2. Acceptor: if n >= promised n, accept and reply ACCEPTED(n, v)
3. Learner: when majority accepted, value is chosen

**Key insight**: Proposal numbers order rounds; acceptors reject older rounds.

### Multi-Paxos

- Elect a *stable leader* (via Paxos) to be the sole proposer
- Skip Phase 1 for subsequent proposals (leader already has promises)
- Batch proposals for efficiency
- Used in practice (Chubby)

### Raft

**Leader Election**:
1. Follower timeout → become candidate, increment term
2. Request votes from all; vote for first candidate in term
3. Candidate with majority → leader
4. Leader sends heartbeats (AppendEntries with no entries)

**Log Replication**:
1. Leader appends to local log
2. Sends AppendEntries to followers
3. Follower: append if log matches (prevLogIndex, prevLogTerm)
4. Leader: when majority ack, commit; notify followers

**Safety**: Election restriction — only vote for candidate whose log is at least as up-to-date

### ZAB (ZooKeeper Atomic Broadcast)

- **Phase 1**: Leader election (similar to Raft)
- **Phase 2**: Atomic broadcast — leader proposes, followers ack, leader commits
- **Difference from Raft**: ZAB guarantees order of delivery; optimized for ZooKeeper's use case
- **Epoch numbers**: Like Raft terms; each leader has unique epoch

### Byzantine Fault Tolerance (PBFT)

- Tolerates **Byzantine** (arbitrary) failures — malicious nodes
- Requires 3f+1 nodes to tolerate f Byzantine nodes
- **Pre-prepare, Prepare, Commit** phases
- Used when adversaries may exist (blockchain, some financial systems)

---

## 5. Numbers

| Algorithm | Nodes for f Failures | Messages per Consensus | Latency (RTT) |
|-----------|---------------------|------------------------|---------------|
| Paxos | 2f+1 | 2 (with stable leader) | 1-2 RTT |
| Raft | 2f+1 | 1 (leader→followers) | 1 RTT |
| ZAB | 2f+1 | 1 (broadcast) | 1 RTT |
| PBFT | 3f+1 | O(n²) | 2 RTT |

### Scale Numbers

- **etcd**: ~1000s of key-value operations/sec per cluster
- **ZooKeeper**: ~10K writes/sec, ~100K reads/sec (reads are local)
- **CockroachDB**: Raft per range; 100s of ranges per node

---

## 6. Tradeoffs

### Paxos vs. Raft

| Aspect | Paxos | Raft |
|--------|-------|------|
| Understandability | Complex | Designed for understandability |
| Leader | Implicit (Multi-Paxos) | Explicit, always |
| Log holes | Possible | No holes (leader has complete log) |
| Membership change | Complex | Joint consensus |
| Industry adoption | Chubby, DynamoDB | etcd, CockroachDB, TiKV |

### Consensus vs. No Consensus

| With Consensus | Without |
|----------------|---------|
| Strong consistency | Eventual consistency |
| CP in CAP | AP in CAP |
| Higher latency | Lower latency |
| Single leader (usually) | Multi-leader possible |

---

## 7. Variants / Implementations

### Viewstamped Replication (VR)

- Similar to Raft; view = term
- Primary (leader) + backups
- Used in some systems

### EPaxos (Egalitarian Paxos)

- No leader; any node can propose
- Conflict-free commands fast-path
- Used in some research systems

### Fast Paxos

- Optimistic path: 1 RTT when no conflicts
- Fallback to classic Paxos on conflict

---

## 8. Scaling Strategies

1. **Sharding**: Multiple Raft groups (e.g., CockroachDB ranges)
2. **Batching**: Batch multiple proposals in one round
3. **Pipeline**: Don't wait for commit before next proposal
4. **Read replicas**: Followers serve reads (Raft read-only)
5. **Leader placement**: Co-locate leader with hot shard

---

## 9. Failure Scenarios

| Scenario | Paxos | Raft |
|----------|-------|------|
| Leader crash | New proposer; may need full Phase 1 | Election; new leader |
| Network partition | Blocks if no majority | Blocks if minority partition |
| Split vote | Retry with new proposal number | Retry with new term (randomized timeout) |
| Duplicate leader | Impossible (safety) | Impossible (term + log) |

### FLP Impossibility

- In asynchronous system with one faulty node, consensus is **impossible** to guarantee in finite time
- **Practical implication**: All real systems assume partial synchrony (timeouts, failure detectors)

---

## 10. Performance Considerations

- **Latency**: Typically 1-2 RTT for commit
- **Throughput**: Limited by leader; batching helps
- **Disk**: Sync to disk before ack (durability)
- **Network**: Minimize cross-datacenter traffic; put leader in same region as clients

---

## 11. Use Cases

| Use Case | Algorithm | Why |
|----------|------------|-----|
| Kubernetes state | Raft (etcd) | Strong consistency, well-understood |
| Kafka controller | ZAB (ZooKeeper) | Coordination |
| Distributed DB | Raft/Paxos | Replicated log |
| Lock service | Paxos (Chubby) | Single source of truth |
| Blockchain | PBFT (some) | Byzantine tolerance |

---

## 12. Comparison Tables

### Algorithm Summary

| Algorithm | Origin | Leader | Complexity | Used By |
|-----------|--------|--------|------------|---------|
| Paxos | Lamport 1998 | Implicit | High | Chubby, DynamoDB |
| Raft | Ongaro 2014 | Explicit | Lower | etcd, CockroachDB, TiKV |
| ZAB | Yahoo | Explicit | Medium | ZooKeeper |
| PBFT | Castro 1999 | Primary | High | Byzantine systems |
| VR | Oki 1988 | Primary | Medium | Research |

---

## 13. Code or Pseudocode

### Raft — RequestVote RPC Handler

```python
def handle_request_vote(self, term, candidate_id, last_log_index, last_log_term):
    # Reply false if term < currentTerm
    if term < self.current_term:
        return (self.current_term, False)
    
    # If term > currentTerm, step down to follower
    if term > self.current_term:
        self.current_term = term
        self.voted_for = None
        self.state = FOLLOWER
    
    # Vote for at most one candidate per term
    if self.voted_for is not None and self.voted_for != candidate_id:
        return (self.current_term, False)
    
    # Candidate's log must be at least as up-to-date as receiver's
    if last_log_term < self.log[-1].term:
        return (self.current_term, False)
    if last_log_term == self.log[-1].term and last_log_index < len(self.log) - 1:
        return (self.current_term, False)
    
    self.voted_for = candidate_id
    self.reset_election_timer()
    return (self.current_term, True)
```

### Raft — AppendEntries RPC Handler

```python
def handle_append_entries(self, term, leader_id, prev_log_index, prev_log_term, entries, leader_commit):
    if term < self.current_term:
        return (self.current_term, False)
    
    self.reset_election_timer()
    
    # Check log consistency
    if prev_log_index >= 0 and (prev_log_index >= len(self.log) or 
                                self.log[prev_log_index].term != prev_log_term):
        return (self.current_term, False)
    
    # Append new entries, delete conflicting
    for i, entry in enumerate(entries):
        idx = prev_log_index + 1 + i
        if idx < len(self.log):
            if self.log[idx].term != entry.term:
                self.log = self.log[:idx]
                self.log.append(entry)
        else:
            self.log.append(entry)
    
    # Update commit index
    if leader_commit > self.commit_index:
        self.commit_index = min(leader_commit, len(self.log) - 1)
    
    return (self.current_term, True)
```

### Paxos — Acceptor Logic

```python
def on_prepare(self, n):
    if n > self.promised_n:
        self.promised_n = n
        return (True, self.accepted_n, self.accepted_v)
    return (False, None, None)

def on_accept(self, n, v):
    if n >= self.promised_n:
        self.promised_n = n
        self.accepted_n = n
        self.accepted_v = v
        return True
    return False
```

---

## 14. Interview Discussion

### Key Points

1. **Consensus = agreement** — All correct nodes decide same value
2. **FLP**: Impossible in pure async; systems use timeouts
3. **Raft vs. Paxos**: Raft is easier to understand; both used in production
4. **Leader-based**: Most algorithms have a leader for efficiency

### Common Questions

- **"Explain Raft leader election"** — Timeout → candidate → RequestVote → majority → leader
- **"Why can't we have 2 leaders in Raft?"** — Term + log comparison; at most one can get majority
- **"What is FLP?"** — Impossibility of consensus in async system with one fault
- **"Paxos vs. Raft?"** — Raft: explicit leader, no log holes, easier to teach

### Red Flags

- Claiming consensus without majority
- Ignoring FLP
- Confusing consensus with 2PC (2PC is not consensus)

---

## 15. Deep Dive: FLP Impossibility (Fischer, Lynch, Paterson)

**Theorem**: In an asynchronous system where at least one node may fail by stopping, there is no consensus algorithm that is guaranteed to terminate.

**Proof sketch**: Assume all messages eventually delivered, but no bound on delay. Assume one node can crash. By contradiction: any protocol that terminates must have a "bivalent" initial state (could go 0 or 1) that leads to decision. Show that we can keep the system in bivalent state indefinitely by delaying messages.

**Practical impact**: Real systems use:
- **Failure detectors** (timeouts) — assumes partial synchrony
- **Randomization** — some protocols use random to achieve probabilistic termination
- **Leader-based** — simplifies; leader failure triggers re-election (with timeout)

---

## 16. Deep Dive: Raft Log Matching

**Log Matching Property**: If two logs have an entry with same index and term, they have the same command and all preceding entries.

**Why**: Leader creates at most one entry per index; never deletes or overwrites. Once entry committed, all future leaders have it.

**Election restriction**: Candidate requests vote with (lastLogIndex, lastLogTerm). Voter grants only if candidate's log is at least as up-to-date (higher term, or same term and longer log). Ensures committed entries are never overwritten.

---

## 17. Deep Dive: Multi-Paxos Optimization

**Stable leader**: First round of Paxos elects leader. Leader becomes sole proposer.

**Skip Phase 1**: Leader already has promises from majority. For subsequent values, go directly to Phase 2 (Accept).

**Batching**: Leader batches multiple client requests into one consensus instance.

**Reordering**: Leader can reorder for efficiency; clients see order of commitment.

---

## 18. ZAB vs. Raft: Key Differences

| Aspect | ZAB | Raft |
|--------|-----|------|
| Ordering | All-or-nothing atomic broadcast | Log entries per index |
| Epoch | Leader has unique epoch | Term |
| Recovery | Leader syncs with followers | Leader forces its log |
| Design goal | ZooKeeper's use case | General-purpose, teachable |

**ZAB**: Optimized for ZooKeeper's primary-backup, in-memory state. **Raft**: General replicated log.

---

## 19. Byzantine Faults: Why 3f+1?

With f Byzantine nodes:
- **Total nodes**: 3f+1
- **Correct nodes**: 2f+1
- **Byzantine can lie**: Send different values to different nodes
- **Quorum**: Need 2f+1 to agree; any two quorums intersect in at least one correct node
- **3f+1**: Minimum for which 2f+1 > f (so correct majority exists)

**PBFT phases**: Pre-prepare (leader), Prepare (all), Commit (all). Two rounds of all-to-all.

---

## 20. Consensus in Practice: Operational Considerations

**Disk**: Must sync log to disk before acknowledging (durability). fsync cost.

**Snapshots**: Log grows unbounded. Periodically snapshot state, truncate log. Raft: snapshot includes last included index/term.

**Membership change**: Adding/removing nodes. Raft: joint consensus (old + new config) for transition. Risky: can cause split if done wrong.

**Leadership transfer**: Graceful handover. Raft: leader steps down, triggers election. Used for maintenance.

**Read-only queries**: Can followers serve reads? Raft: leader can respond; followers need to check they're not stale (lease or read-only request that gets committed).

---

## 21. Raft Membership Change: Joint Consensus

**Problem**: Adding/removing node. Cannot do atomically with single config (could cause split).

**Joint consensus**: Transition config C_old + C_new. Both configs must agree for commit. Leader replicates C_old,joint,C_new. Once committed, replicate C_new. Then leave joint.

**Safety**: During joint, quorum of C_old and quorum of C_new both required. Prevents split.

**Complexity**: Easy to get wrong. Many systems use external config management.

---

## 22. Consensus and Replicated State Machines

**Idea**: Log is sequence of commands. All nodes apply same commands in same order. State machine is deterministic. Result: identical state.

**Apply**: When entry committed, apply command to state machine. Order matters.

**Snapshot**: State can be large. Periodically snapshot state, truncate log. New node joins: receive snapshot + tail of log.

---

## 23. Paxos Variants: Fast Paxos, Multi-Paxos, Cheap Paxos

**Fast Paxos**: Optimistic. Proposer sends value directly to acceptors. If no conflict, 1 RTT. If conflict, fall back to classic.

**Cheap Paxos**: Reduce nodes for quorum when few failures. Uses "auxiliary" nodes.

**Flexible Paxos**: Quorum can be any set as long as intersection property holds. Enables smaller quorums for reads.

---

## 24. Raft in Production: etcd and TiKV

**etcd**: Key-value store. Raft for consensus. Used by Kubernetes. Tuned for: small values, watch API, lease.

**TiKV**: Distributed KV for TiDB. Raft per region. Optimized for: large values, multi-Raft, leader balance.

**Common**: Both use batching, pipeline, read-only from followers (with lease).

---

## 25. Summary: Consensus Algorithm Selection

| Need | Algorithm | Why |
|------|------------|-----|
| Understandable | Raft | Designed for teaching |
| Battle-tested | Paxos | Chubby, DynamoDB |
| ZooKeeper ecosystem | ZAB | Native to ZK |
| Byzantine faults | PBFT | Malicious nodes |
| No leader | EPaxos | Research |

---

## 26. Further Reading

- **Paxos**: Lamport, "The Part-Time Parliament" (1998)
- **Raft**: Ongaro & Ousterhout, "In Search of an Understandable Consensus Algorithm" (2014)
- **FLP**: Fischer, Lynch, Paterson, "Impossibility of Distributed Consensus with One Faulty Process" (1985)
- **ZAB**: Junqueira, Reed, Serafini, "Zab: High-performance broadcast for primary-backup systems" (2011)

---

## 27. Quick Reference: Consensus Properties

| Property | Meaning |
|----------|---------|
| Agreement | All correct nodes decide same value |
| Validity | Decided value was proposed |
| Termination | All correct nodes eventually decide |
| Integrity | No node decides twice |
| Leader election | Implicit (Paxos) or explicit (Raft, ZAB) |
| Fault tolerance | 2f+1 nodes tolerate f failures (crash); 3f+1 for Byzantine |

**Note**: In production, always run with odd cluster sizes (3, 5, 7) to ensure clear majority.
