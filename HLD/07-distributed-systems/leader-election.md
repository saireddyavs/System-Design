# Leader Election in Distributed Systems

> Staff+ Engineer Level — FAANG Interview Deep Dive

---

## 1. Concept Overview

### Definition

**Leader election** is the process by which a group of distributed nodes selects exactly one node to act as the **leader** (coordinator, master) while others become **followers** (workers, replicas). The leader typically has exclusive authority for certain operations.

### Purpose

- **Coordination**: Single point for scheduling, partitioning, or decision-making
- **Single writer**: Avoid conflicting writes; leader serializes updates
- **Resource management**: Leader allocates work (e.g., Kafka partition assignment)
- **Simplified logic**: Easier than coordinating among equals

### Problems Solved

| Problem | Solution |
|---------|----------|
| Split-brain (multiple leaders) | Leader election + fencing |
| Coordination overhead | Single leader makes decisions |
| Write conflicts | Leader is sole writer |
| Work distribution | Leader assigns partitions/tasks |

---

## 2. Real-World Motivation

### Kafka

- **Broker controller**: One broker is "controller" — manages partition assignment, leader election for partitions, cluster metadata
- Uses ZooKeeper (pre-KRaft) or KRaft (Kafka Raft) for controller election
- Controller failure triggers re-election; new controller takes over

### Elasticsearch

- **Master node**: Elected via Zen Discovery (gossip-based) or newer coordination layer
- Master manages cluster state, index creation, shard allocation
- Split-brain prevented by `discovery.zen.minimum_master_nodes`

### Redis Sentinel

- **Leader Sentinel**: Elected among Sentinels to perform failover
- Monitors masters and replicas; promotes replica to master on failure
- Uses Raft-like protocol for agreement

### HDFS (Hadoop)

- **NameNode**: Single active NameNode (with standby for HA)
- ZooKeeper used for failover; fencing via shared storage
- Leader holds entire namespace; critical single point

### etcd / Kubernetes

- **Raft leader**: etcd uses Raft; leader handles writes
- Kubernetes control plane relies on etcd leader for API persistence

---

## 3. Architecture Diagrams

### Bully Algorithm

```
    Nodes: 1, 2, 3, 4, 5 (ordered by ID)
    
    Node 3 detects Node 5 (leader) has failed
         |
         v
    Node 3 sends ELECTION to 4, 5
         |
         v
    Node 4 responds (alive); Node 5 doesn't
         |
         v
    Node 4 sends ELECTION to 5
         |
         v
    No response from 5
         |
         v
    Node 4 declares itself LEADER, sends COORDINATOR to all
```

### Ring Algorithm

```
    N1 -----> N2 -----> N3 -----> N4 -----> N1
     |         |         |         |
     |    Election message passes around ring
     |    Each node adds self to list, forwards
     |    When message returns to initiator: highest ID wins
     v
    N4 is leader
```

### ZooKeeper-Based Leader Election

```
    /election
        /node_0000000001  (ephemeral, sequential)
        /node_0000000002
        /node_0000000003  <-- lowest sequence = leader
        /node_0000000004
        /node_0000000005
    
    Each node creates ephemeral sequential znode
    Node with lowest sequence number is leader
    Others watch the node before them; when it goes away, check if leader
```

### etcd Lease-Based Election

```
    Candidate                    etcd Cluster
        |                             |
        |  PUT /leader (with lease)   |
        |---------------------------->|
        |  lease TTL = 10s            |
        |<----------------------------|
        |  (periodic keepalive)       |
        |---------------------------->|
        |  (if keepalive fails, lease expires, key deleted)
        |                             |
    Follower detects /leader gone, tries to become leader
```

### Raft Leader Election

```
    [Follower] --timeout--> [Candidate] --majority votes--> [Leader]
         ^                         |                              |
         |                         |--no majority, new term-------|
         |<-------- AppendEntries (heartbeat) --------------------|
```

### Fencing Token Flow

```
    Client A (old leader)     Storage        Client B (new leader)
         |                       |                    |
         |  write (token=33)     |                    |
         |---------------------->|  reject (33 < 34)  |
         |                       |                    |
         |                       |<--- write (token=34)-|
         |                       |  accept            |
```

---

## 4. Core Mechanics

### Bully Algorithm

- **Assumption**: Each node has unique ID; all know all IDs
- **Process**: On coordinator failure, node with higher ID "bullies" others
- **Steps**: Send ELECTION to higher IDs; if no response, declare self leader
- **Messages**: O(n²) in worst case (cascading elections)

### Ring Algorithm

- **Topology**: Logical ring; each node knows successor
- **Process**: Election message circulates; each adds self, forwards
- **Leader**: Highest ID in message when it returns
- **Messages**: O(n) but slow (sequential)

### ZooKeeper Ephemeral Sequential

- **Create** `/election/node_` with SEQUENCE flag → get unique path like `node_0000000003`
- **Leader**: Smallest sequence number
- **Watch**: Non-leaders watch the znode just before theirs; on deletion, re-check
- **Automatic**: Leader failure → ephemeral node deleted → next in line notified

### etcd Lease Mechanism

- **Lease**: Time-bounded ownership; client must renew (keepalive)
- **Election**: Compete to create key with lease; winner holds until lease expires
- **Failure**: No keepalive → lease expires → key deleted → others can try

### Raft Election

- **Trigger**: Follower election timeout (150-300ms typical)
- **Term**: Increment term; vote for self; RequestVote to all
- **Vote**: Grant to first candidate in term with up-to-date log
- **Leader**: Majority votes → send heartbeats

### Fencing Tokens

- **Monotonic**: Each new leader gets higher token
- **Storage**: Rejects operations with stale token
- **Prevents**: Old leader (split-brain) from writing after new leader elected

---

## 5. Numbers

| System | Election Time | Heartbeat Interval | Typical Cluster Size |
|--------|---------------|-------------------|----------------------|
| ZooKeeper | ~1-5s | 2s (tick time) | 3, 5, 7 |
| etcd | ~1-2s | 100ms (Raft) | 3, 5, 7 |
| Kafka Controller | ~5-30s | 6s (session timeout) | 100s of brokers |
| Elasticsearch | ~1-3s | 1s | 100s of nodes |
| Redis Sentinel | ~10-30s | 1s | 3-5 Sentinels |

### Scale

- **Kafka**: 100K+ partitions; controller manages all
- **Elasticsearch**: 1000+ node clusters
- **etcd**: 1000s of keys; single leader for writes

---

## 6. Tradeoffs

### Bully vs. Ring vs. ZooKeeper

| Aspect | Bully | Ring | ZooKeeper |
|--------|-------|------|-----------|
| Messages | O(n²) | O(n) | O(1) per node |
| Latency | Low (parallel) | High (sequential) | Low |
| Dependencies | None | None | ZooKeeper |
| Split-brain | Possible | Possible | Prevented (ZK consistency) |

### Leader Lease

| Short lease | Long lease |
|------------|------------|
| Fast failover | Fewer renewals |
| More overhead | Slower failover |
| Lower split-brain window | Higher risk if clock skew |

---

## 7. Variants / Implementations

### Apache Curator (ZooKeeper)

- `LeaderSelector` recipe: automatic re-election on leadership loss
- `LeaderLatch`: block until leader

### etcd Campaign

- `campaign(key)` — blocks until elected
- Uses lease; competitor waits on key deletion

### Raft (built-in)

- Leader election is part of Raft; no separate mechanism
- etcd, Consul, CockroachDB use Raft

### Consul

- Raft-based; leader handles writes
- Session-based health checks

---

## 8. Scaling Strategies

1. **Sharding**: Multiple leaders (e.g., Kafka partition leaders; one controller for metadata)
2. **Hierarchy**: Meta-leader elects sub-leaders (e.g., Kafka controller → partition leaders)
3. **Caching**: Followers serve reads; reduce leader load
4. **Batching**: Leader batches updates to reduce RPCs

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|------------|
| Leader crash | Election triggered | Timeout-based detection |
| Network partition | Minority partition blocks | Quorum requirement |
| Split-brain | Two leaders | Fencing tokens, quorum |
| Slow leader | Delayed operations | Lease timeout, replace |
| ZooKeeper outage | No election | ZooKeeper HA; multiple nodes |

### Split-Brain Prevention

1. **Quorum**: Only majority partition can elect leader
2. **Fencing**: Storage rejects old leader's writes
3. **Lease**: Leader holds lease; expires if disconnected
4. **Minimum master nodes** (Elasticsearch): Prevents minority from electing

---

## 10. Performance Considerations

- **Election frequency**: Minimize (stable leader is good)
- **Heartbeat tuning**: Balance between failover speed and overhead
- **Watch notifications**: ZooKeeper watchers; avoid herd effect
- **Lease renewal**: etcd keepalive; network blips can cause unnecessary elections

---

## 11. Use Cases

| Use Case | Why Leader |
|----------|------------|
| Kafka controller | Partition assignment, metadata |
| Database primary | Single writer, replication |
| Job scheduler | Avoid duplicate job execution |
| Config master | Single source of truth |
| Lock service | Centralized lock management |

---

## 12. Comparison Tables

### Election Mechanism by System

| System | Mechanism | Coordination Service |
|--------|-----------|------------------------|
| Kafka | ZooKeeper / KRaft | ZooKeeper / Kafka Raft |
| Elasticsearch | Zen Discovery / coordination | Built-in |
| Redis | Sentinel Raft-like | Sentinel cluster |
| HDFS | ZooKeeper failover | ZooKeeper |
| etcd | Raft | Self (etcd cluster) |
| ZooKeeper | ZAB (internal) | Self |

### Failure Detection Methods

| Method | Latency | False Positives | Use Case |
|--------|---------|-----------------|----------|
| Heartbeat timeout | Configurable | Possible | Most systems |
| Phi accrual | Adaptive | Lower | Cassandra, Akka |
| Lease expiry | TTL-based | Low | etcd, Chubby |
| ZooKeeper session | Session timeout | Low | ZooKeeper clients |

---

## 13. Code or Pseudocode

### ZooKeeper Leader Election (Curator-style)

```python
def elect_leader():
    zk = connect("localhost:2181")
    path = zk.create("/election/node_", ephemeral=True, sequential=True)
    sequence = int(path.split("_")[-1])
    
    children = zk.get_children("/election")
    sorted_children = sorted(children, key=lambda c: int(c.split("_")[-1]))
    
    if path.endswith(sorted_children[0]):
        return "LEADER"
    
    # Watch the node before me
    my_index = sorted_children.index(path.split("/")[-1])
    watch_node = "/election/" + sorted_children[my_index - 1]
    
    event = zk.exists(watch_node, watch=True)
    # When watch_node is deleted, re-run election check
    wait_for_event()
    return elect_leader()  # Recursive re-check
```

### etcd Lease-Based Election

```python
def campaign(etcd_client, key="/leader"):
    lease = etcd_client.lease(10)  # 10 second TTL
    try:
        etcd_client.put(key, self.id, lease=lease)
        # We won - we're the leader
        while True:
            lease.keepalive()  # Renew every ~3s
            do_leader_work()
    except etcd.exceptions.AlreadyExists:
        # Someone else is leader, watch for change
        watch = etcd_client.watch(key)
        for event in watch:
            if event.type == "DELETE":
                return campaign(etcd_client, key)  # Retry
```

### Bully Algorithm (Simplified)

```python
def bully_election(self, node_id, all_nodes):
    higher_nodes = [n for n in all_nodes if n > node_id]
    for n in higher_nodes:
        if send_election_to(n) and receive_ok_from(n):
            return  # Someone else will become leader
    # No higher node responded - we are leader
    for n in all_nodes:
        send_coordinator(n, node_id)
    self.state = LEADER
```

### Raft RequestVote (Leader Election Part)

```python
def start_election(self):
    self.current_term += 1
    self.voted_for = self.node_id
    self.state = CANDIDATE
    votes = 1
    
    for node in self.peers:
        last_log_index = len(self.log) - 1
        last_log_term = self.log[-1].term if self.log else 0
        reply = node.request_vote(self.current_term, self.node_id, 
                                  last_log_index, last_log_term)
        if reply.vote_granted:
            votes += 1
            if votes > len(self.peers) // 2:
                self.become_leader()
                return
    # No majority - wait for timeout, retry
```

---

## 14. Interview Discussion

### Key Points

1. **Why leader?** — Coordination, single writer, simpler semantics
2. **Split-brain** — Prevent with quorum, fencing, lease
3. **ZooKeeper** — Ephemeral sequential; lowest sequence = leader
4. **etcd** — Lease-based; key + lease = leadership

### Common Questions

- **"How does ZooKeeper leader election work?"** — Ephemeral sequential znodes; smallest = leader; others watch predecessor
- **"What is fencing?"** — Monotonic tokens; storage rejects stale leader
- **"How do you prevent split-brain?"** — Quorum, fencing tokens, leader lease
- **"Bully vs. Ring?"** — Bully: O(n²) messages, faster; Ring: O(n), slower

### Red Flags

- No fencing in critical systems
- Ignoring split-brain
- Single point of failure without failover

---

## 15. Deep Dive: Leader Lease Mechanics

**Lease**: Time-bounded guarantee that leader holds authority. If leader doesn't renew before expiry, others may take over.

**Renewal**: Leader sends periodic keepalive. Server extends lease on each keepalive.

**Failure**: If leader crashes or network partitions, keepalive stops. Lease expires. New leader can be elected.

**Clock skew**: Lease duration must account for clock skew. If skew > lease/2, risk of overlap.

**Chubby**: Uses 12-second lease; leader renews every 4 seconds. Clients get 12-second cache.

---

## 16. Deep Dive: ZooKeeper Session and Ephemeral Nodes

**Session**: Client connects; gets session ID. Session has timeout (e.g., 30s). If no heartbeat, session expires.

**Ephemeral node**: Tied to session. When session ends, node is deleted. Perfect for "I'm alive" or leader.

**Sequential**: Append monotonic counter. Creates total order. Lowest = leader.

**Watch**: Client can watch a znode. On deletion (or change), watcher fires. Used for "wait for predecessor to go away."

---

## 17. Deep Dive: Elasticsearch Minimum Master Nodes

**Setting**: `discovery.zen.minimum_master_nodes = (N/2) + 1`

**Purpose**: Prevent split-brain. Only partition with majority can elect master.

**Example**: 3 nodes. minimum_master_nodes=2. If 1 node partitions off, it can't elect (needs 2). Remaining 2 can elect.

**Quorum**: Same as Raft — majority required.

---

## 18. Failure Detection: Phi Accrual vs. Fixed Timeout

**Fixed timeout**: Simple. If no heartbeat for T, consider dead. Problem: T too small → false positives; T too large → slow failover.

**Phi accrual**: Adaptive. Based on historical variance of inter-arrival times. Phi = suspicion level. Higher = more likely dead. Threshold (e.g., 8) triggers. Used by Cassandra, Akka.

**Tradeoff**: Phi accrual adapts to network conditions; fixed timeout is predictable.

---

## 19. Leader Election in Multi-Datacenter

**Challenge**: Cross-DC latency; partitions between DCs.

**Options**:
1. **Single DC leader**: Leader in one DC; others replicate. Failover to same DC preferred.
2. **Regional leaders**: Each DC has leader for its partition; no cross-DC election.
3. **Quorum across DC**: Requires majority across DCs; minority DC cannot elect. Higher latency.

**Kafka**: Controller in one broker; typically in same DC as clients. Failover to same DC.

---

## 20. Interview Walkthrough: Designing Leader Election

**Question**: "How would you implement leader election for a job scheduler?"

**Answer structure**:
1. **Requirements**: Single active scheduler; failover < 30s; no duplicate job runs
2. **Mechanism**: ZooKeeper ephemeral sequential or etcd lease
3. **Fencing**: Use fencing tokens when scheduler executes jobs (e.g., on storage)
4. **Failure detection**: Session timeout (ZK) or lease TTL (etcd)
5. **Split-brain**: Quorum ensures only one partition elects
6. **Recovery**: New leader loads state from persistent store; reconciles any in-flight jobs

---

## 21. Apache Curator LeaderSelector

**Curator**: ZooKeeper client library. **LeaderSelector**: Recipe for leader election.

**Usage**: `LeaderSelectorListener` — `takeLeadership()` called when elected. Block in that method while leader. When method returns, release. Curator handles re-election.

**AutoRequeue**: Option to automatically re-enter election when leadership lost.

**Use case**: Kafka controller (pre-KRaft), custom coordinators.

---

## 22. Raft Election Timeout Tuning

**Range**: 150ms-300ms typical. Random between min and max to avoid split votes.

**Too short**: Unnecessary elections; churn. **Too long**: Slow failover.

**Heartbeat**: Leader sends AppendEntries every heartbeat_interval (e.g., 50ms). Follower timeout > several heartbeats (e.g., 150-300ms).

**Split vote**: Two candidates get 2 votes each (in 4-node). No majority. Both timeout. Retry with new term. Randomization reduces probability of repeated split.

---

## 23. Leader Election in Kafka (KRaft)

**KRaft**: Kafka Raft. Replaces ZooKeeper for metadata. Controller is Raft leader.

**Election**: Standard Raft. Controller holds cluster metadata. All brokers know controller.

**Failover**: Controller crash → Raft election → new controller. Brokers reconnect to new controller. Partition assignment, topic config, etc. managed by controller.

---

## 24. HDFS NameNode HA

**Active/Standby**: Two NameNodes. One active, one standby. Shared edit log (NFS or JournalNodes).

**Failover**: ZooKeeper used. Active holds ephemeral znode. On failure, znode deleted. Standby takes over. Fencing: prevent old active from writing (shared storage fencing, or "fence" command).

**Observer NameNode**: Read-only NameNode for scaling reads. Not in election.

---

## 25. Summary: Leader Election by System

| System | Mechanism | Coordination | Failover Time |
|--------|-----------|--------------|---------------|
| ZooKeeper | ZAB internal | Self | ~1-5s |
| etcd | Raft | Self | ~1-2s |
| Kafka | ZK or KRaft | ZK or Raft | ~5-30s |
| Elasticsearch | Zen/coordination | Gossip | ~1-3s |
| Redis Sentinel | Raft-like | Sentinel cluster | ~10-30s |
| HDFS | ZK failover | ZK | ~30-60s |
