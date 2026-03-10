# Data Replication

## Staff+ Engineer Deep Dive for FAANG Interviews

---

## 1. Concept Overview

### Definition
**Replication** is the process of copying and maintaining data across multiple database instances to ensure redundancy, improve availability, and enable read scaling. Replicas can be synchronous (strong consistency) or asynchronous (eventual consistency).

### Purpose
- **Fault tolerance**: If primary fails, replica can take over
- **Read scaling**: Distribute read load across replicas
- **Latency reduction**: Place replicas geographically close to users
- **Disaster recovery**: Replica in different datacenter/region

### Why It Exists
Single-node databases are single points of failure. Replication provides high availability (HA) and enables horizontal read scaling without the complexity of sharding for write scaling.

### Problems Solved
| Problem | Replication Solution |
|---------|----------------------|
| Single point of failure | Replica failover |
| Read bottleneck | Route reads to replicas |
| High latency (distant users) | Regional replicas |
| Data loss (disk failure) | Multiple copies |

---

## 2. Real-World Motivation

### PostgreSQL (Streaming Replication)
- Primary-replica; WAL shipping (physical) or logical replication
- Used by Instagram, Spotify, Discord
- Sync replication for critical data; async for scale

### MySQL (Replication)
- Binary log (binlog) shipping; async by default
- Semi-sync option for durability
- Used by Facebook, Uber, Airbnb

### Cassandra (Leaderless)
- No primary; all nodes equal; tunable consistency (ONE, QUORUM, ALL)
- Used by Netflix, Apple, Instagram (for some workloads)

### DynamoDB (Multi-Region)
- Global tables: active-active replication across regions
- Conflict resolution: last-writer-wins
- Used by Amazon, Lyft

### MongoDB (Replica Set)
- Primary + secondaries; automatic failover
- Read preference: primary, primaryPreferred, secondary, nearest

### Google Spanner
- Synchronous replication across datacenters
- Paxos for consensus; external consistency via TrueTime

---

## 3. Architecture Diagrams

### Single-Leader (Leader-Follower)
```
┌─────────────────────────────────────────────────────────────────────────┐
│                    SINGLE-LEADER REPLICATION                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   WRITES                          READS (optional)                       │
│      │                                  │                                │
│      ▼                                  │                                │
│   ┌─────────────┐                       │                                │
│   │   LEADER    │  ──WAL/Binlog────▶   │                                │
│   │  (Primary)  │                       │                                │
│   └──────┬──────┘                       │                                │
│          │                              │                                │
│          │ Replication stream           │                                │
│          ├──────────────┬──────────────┼──────────────┐                 │
│          ▼              ▼              ▼              ▼                 │
│   ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐           │
│   │ FOLLOWER │   │ FOLLOWER │   │ FOLLOWER │   │ FOLLOWER │           │
│   │ (Replica)│   │ (Replica)│   │ (Replica)│   │ (Replica)│           │
│   └──────────┘   └──────────┘   └──────────┘   └──────────┘           │
│       ▲              ▲              ▲              ▲                     │
│       └──────────────┴──────────────┴──────────────┘                     │
│                    Reads can go to any replica                           │
│                                                                          │
│   Failover: Promote follower to leader (manual or automatic)             │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Multi-Leader
```
┌─────────────────────────────────────────────────────────────────────────┐
│                    MULTI-LEADER REPLICATION                               │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   Datacenter A              Datacenter B              Datacenter C       │
│   ┌─────────────┐          ┌─────────────┐          ┌─────────────┐    │
│   │  LEADER A   │◀────────▶│  LEADER B   │◀────────▶│  LEADER C   │    │
│   │  (writes)   │  async   │  (writes)   │  async   │  (writes)   │    │
│   └──────┬──────┘  sync    └──────┬──────┘          └──────┬──────┘    │
│          │                        │                         │            │
│          ▼                        ▼                         ▼            │
│   ┌──────────┐            ┌──────────┐            ┌──────────┐          │
│   │ Replicas │            │ Replicas │            │ Replicas │          │
│   └──────────┘            └──────────┘            └──────────┘          │
│                                                                          │
│   Writes to any leader; conflict resolution required                     │
│   Use case: Multi-region, offline-first                                 │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Leaderless (Dynamo-style)
```
┌─────────────────────────────────────────────────────────────────────────┐
│                    LEADERLESS REPLICATION                                 │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   Client writes to N nodes (W of N must ack)                              │
│   Client reads from N nodes (R of N must respond)                        │
│   Consistency: W + R > N → strong consistency                            │
│                                                                          │
│        ┌──────┐                                                          │
│        │Client│                                                          │
│        └──┬───┘                                                          │
│           │                                                              │
│     ┌─────┼─────┬─────┐                                                  │
│     ▼     ▼     ▼     ▼                                                  │
│  ┌────┐ ┌────┐ ┌────┐ ┌────┐  ┌────┐  ┌────┐                             │
│  │ N1 │ │ N2 │ │ N3 │ │ N4 │  │ N5 │  │ N6 │  ← All equal, no leader     │
│  └────┘ └────┘ └────┘ └────┘  └────┘  └────┘                             │
│     │     │     │     │                                                   │
│     └─────┴─────┴─────┘                                                   │
│           Gossip protocol for sync                                        │
│                                                                          │
│   N=6, W=2, R=2: Write 2, read 2; eventual consistency                   │
│   N=6, W=3, R=3: Quorum; strong consistency                              │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Replication Lag Problems
```
┌─────────────────────────────────────────────────────────────────────────┐
│                    REPLICATION LAG SCENARIOS                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   READ-AFTER-WRITE: User writes, immediately reads from stale replica    │
│   ┌────────┐  write   ┌────────┐     ┌────────┐  read (stale)            │
│   │ Client │─────────▶│ Leader │────▶│Replica│◀──────────│ Client       │
│   └────────┘          └────────┘     └────────┘  (not yet replicated)   │
│                                                                          │
│   MONOTONIC READS: User reads older then newer (replica A lagged)        │
│   Read 1 (replica A, stale) → Read 2 (replica B, fresh) = going "back"   │
│                                                                          │
│   CONSISTENT PREFIX: User sees conversation out of order                 │
│   Msg1, Msg3, Msg2 (Msg2 replicated before Msg3 to user's replica)       │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Synchronous vs Asynchronous
```
┌─────────────────────────────────────────────────────────────────────────┐
│              SYNCHRONOUS vs ASYNCHRONOUS REPLICATION                       │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   SYNCHRONOUS:                                                           │
│   Client ──write──▶ Leader ──sync replicate──▶ Replica ──ack──▶ Leader  │
│   Leader ──ack──▶ Client                                                    │
│   Latency: +1 RTT to replica; No data loss on leader fail                  │
│                                                                          │
│   ASYNCHRONOUS:                                                          │
│   Client ──write──▶ Leader ──ack──▶ Client                               │
│   Leader ──async replicate──▶ Replica (background)                       │
│   Latency: Lower; Risk: Leader fails before replicate = data loss       │
│                                                                          │
│   SEMI-SYNC: At least 1 replica ack before leader ack                     │
│   Balance: Some durability, acceptable latency                           │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Core Mechanics

### WAL (Write-Ahead Log) Shipping
- Leader writes to WAL before applying to data
- WAL records shipped to replicas
- Replicas apply WAL in same order (physical replication)
- Low overhead; byte-for-byte copy

### Logical Replication
- Replicate logical changes (INSERT, UPDATE, DELETE)
- Can filter, transform; cross-version possible
- Higher overhead than physical

### Conflict Resolution (Multi-Leader)
- **Last-write-wins (LWW)**: Timestamp; simple but can lose updates
- **Vector clocks**: Track causality; detect conflicts
- **CRDTs**: Conflict-free replicated data types; merge automatically
- **Application-level**: Custom merge logic

### Quorum
- W + R > N: Every read sees at least one up-to-date write
- W = R = (N+1)/2: Typical for strong consistency

---

## 5. Numbers

| Metric | Value |
|--------|-------|
| Replication lag (async) | 100ms - 5s typical |
| Sync replication latency add | 1-10ms (same region) |
| MySQL binlog size | ~1MB per 1K transactions |
| PostgreSQL WAL segment | 16MB default |
| Cassandra consistency ONE | 1 node; QUORUM = (N/2)+1 |

### Scale
- **Facebook**: MySQL replication across regions; seconds lag
- **Netflix**: Cassandra multi-dc; tunable consistency
- **Spanner**: Sync replication; <10ms p99 cross-region (same continent)

---

## 6. Tradeoffs

### Sync vs Async
| Aspect | Sync | Async |
|--------|------|-------|
| Consistency | Strong | Eventual |
| Latency | Higher | Lower |
| Data loss risk | Low | Higher |
| Throughput | Lower | Higher |

### Single vs Multi-Leader
| Aspect | Single-Leader | Multi-Leader |
|--------|---------------|--------------|
| Consistency | Strong (with sync) | Conflicts possible |
| Write availability | Single point | Multi-region |
| Complexity | Lower | Conflict resolution |
| Use case | Most systems | Multi-region active-active |

---

## 7. Variants / Implementations

### Replication Topologies
- **Single-leader**: PostgreSQL, MySQL, MongoDB
- **Multi-leader**: CouchDB, DynamoDB Global Tables
- **Leaderless**: Cassandra, DynamoDB, Voldemort

### Replication Methods
- **Statement-based**: Replicate SQL statements (MySQL legacy)
- **Row-based**: Replicate row changes (MySQL default)
- **Logical**: Replicate logical operations (PostgreSQL logical replication)

---

## 8. Scaling Strategies

- **Read replicas**: Route reads to replicas; scale reads
- **Cascading replicas**: Leader → Replica1 → Replica2 (reduces leader load)
- **Sync replicas**: For critical reads (read-your-writes)
- **Geographic distribution**: Replicas in each region

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|-------------|
| Leader failure | Writes fail | Failover to replica |
| Replica failure | Read capacity reduced | Multiple replicas |
| Replication lag | Stale reads | Read from leader for critical |
| Split-brain (multi-leader) | Divergent data | Conflict resolution; avoid |
| Network partition | Sync blocks | Async; accept eventual consistency |

---

## 10. Performance Considerations

- **Replication lag**: Monitor; alert if exceeds threshold
- **Read-your-writes**: Route to leader or sync replica
- **Connection pooling**: Separate pools for leader vs replicas
- **Batch replication**: Group changes to reduce overhead

---

## 11. Use Cases

| Use Case | Topology | Consistency |
|----------|----------|-------------|
| OLTP primary | Single-leader | Strong |
| Read scaling | Single-leader + replicas | Eventual for reads |
| Multi-region | Multi-leader or leaderless | Tunable |
| Offline-first | Multi-leader | Eventual + conflict resolution |

---

## 12. Comparison Tables

### Replication Topology Comparison
| Topology | Consistency | Availability | Complexity | Use Case |
|----------|-------------|--------------|------------|----------|
| Single-leader | Strong (sync) | Leader SPOF | Low | Most apps |
| Multi-leader | Eventual | High | High | Multi-region |
| Leaderless | Tunable | High | Medium | Cassandra, DynamoDB |

### Conflict Resolution
| Strategy | Pros | Cons |
|----------|------|------|
| LWW | Simple | Lost updates |
| Vector clocks | Detect conflicts | Complex |
| CRDTs | Auto-merge | Limited types |
| App-level | Flexible | Custom code |

---

## 13. Code or Pseudocode

### Read-Your-Writes Routing
```python
def get_connection(user_id, is_write=False):
    if is_write:
        return leader_connection
    if user_id in recent_writers:  # Wrote in last N seconds
        return leader_connection  # Ensure read-your-writes
    return replica_connection
```

### Quorum Read/Write
```python
def quorum_write(key, value, w):
    nodes = get_replica_nodes(key)
    acks = 0
    for node in nodes:
        if node.write(key, value):
            acks += 1
            if acks >= w:
                return True
    return False

def quorum_read(key, r):
    nodes = get_replica_nodes(key)
    responses = [node.read(key) for node in nodes]
    # Return latest by version
    return max(responses, key=lambda x: x.version)
```

### Conflict Resolution (LWW)
```python
def merge_conflicts(versions):
    return max(versions, key=lambda v: v.timestamp)
```

---

## 14. Interview Discussion

### Key Points
1. **Replication lag**: Inevitable with async; causes read-after-write, monotonic read issues
2. **Read-your-writes**: Route to leader or sync replica after write
3. **Quorum**: W + R > N for strong consistency
4. **Multi-leader conflicts**: Need resolution strategy
5. **Leaderless**: No single point of failure; tunable consistency

### Common Questions
- **Q**: "What's replication lag?"
  - **A**: Delay between write on leader and visibility on replica; causes stale reads
- **Q**: "How do you ensure read-your-writes?"
  - **A**: Route reads to leader for N seconds after write; or use sync replica
- **Q**: "What's split-brain?"
  - **A**: Two leaders in multi-leader; both accept writes; divergent data
- **Q**: "When use multi-leader?"
  - **A**: Multi-region with local writes; offline-first; need write availability everywhere

---

## 15. Failover Procedures

### Automatic Failover
- **Health check**: Ping primary; detect failure (e.g., 3 consecutive failures)
- **Promotion**: Promote replica to primary; update DNS/config
- **Replication**: New primary; other replicas follow
- **Risk**: Split-brain if network partition (two primaries)
- **Mitigation**: Quorum; fencing; witness node

### Manual Failover
- Operator decides; execute failover script
- Safer for planned maintenance
- Downtime: 30s - 5min typical

### Failover Metrics
- **RTO** (Recovery Time Objective): Target time to recover
- **RPO** (Recovery Point Objective): Max data loss (replication lag)
- **MTTR** (Mean Time To Recovery): Actual recovery time

---

## 16. Monitoring Replication Lag

### Lag Metrics
- **Seconds behind master**: How far replica is in applying log
- **Replication delay**: Time from write to visible on replica
- **Bytes behind**: Pending bytes in replication stream

### Alerting
- Lag > 10s: Warning
- Lag > 60s: Critical (consider read from primary)
- Lag growing: Replica overloaded or network issue

### Tools
- PostgreSQL: pg_stat_replication
- MySQL: SHOW SLAVE STATUS
- Custom: Application-level timestamp in writes; compare on read

---

## 17. Logical vs Physical Replication

### Physical (WAL/Binlog)
- Replicate raw bytes
- Same DB version required
- Low overhead
- Cannot filter or transform

### Logical
- Replicate logical changes (INSERT, UPDATE, DELETE)
- Cross-version possible
- Can filter tables, transform
- Higher overhead
- Use: CDC, data warehouse sync, multi-tenant

---

## 18. Conflict Resolution Examples

### Last-Write-Wins (LWW)
```
Region A: Update user name to "Alice" at T1
Region B: Update user name to "Bob" at T2
Resolution: T2 > T1 → "Bob" wins
Problem: Clock skew; "Alice" might be intended
```

### Vector Clocks
- Each node has vector of (node_id, counter)
- Detect concurrent updates (no causal order)
- Application decides merge or prompt user

### CRDT (Conflict-free Replicated Data Type)
- Mathematical guarantee: merge always converges
- Examples: G-Counter, LWW-Register, OR-Set
- Limited types; complex for nested structures

---

## 19. Read Scaling with Replicas

### Read Distribution
- **Round-robin**: Simple; may hit lagged replica
- **Least connections**: Balance load
- **Route by consistency**: Critical reads → primary; analytics → replica
- **Route by region**: Nearest replica for latency

### Stale Read Handling
- **Max lag**: Don't route to replica if lag > N seconds
- **Session stickiness**: Same replica for session (monotonic reads)
- **Version check**: Include "last write timestamp" in response; client can retry

---

## 20. Replication Topology Examples

### PostgreSQL: Cascading Replicas
```
Primary → Replica1 → Replica2
         → Replica3
```
- Reduces load on primary
- Replica2, Replica3 have higher lag
- Use for read-heavy, less critical reads

### MySQL: Semi-Sync
- Primary waits for at least 1 replica ack before commit
- Balance: Durability vs latency
- Protects against single-replica failure
