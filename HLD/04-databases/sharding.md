# Database Sharding

## Staff+ Engineer Deep Dive for FAANG Interviews

---

## 1. Concept Overview

### Definition
**Sharding** is a horizontal partitioning strategy that distributes data across multiple database instances (shards). Each shard holds a subset of the data and operates independently, allowing the system to scale beyond the limits of a single machine.

### Purpose
- **Scale storage**: Single node has finite disk (typically 1-2TB practical limit for OLTP)
- **Scale compute**: CPU and memory limits per machine
- **Scale throughput**: Distribute read/write load across nodes
- **Geographic distribution**: Place data closer to users

### Why It Exists
Vertical scaling (bigger machine) hits physical and economic limits. At web scale (billions of users, petabytes of data), horizontal scaling is the only path. Sharding is the primary mechanism for horizontal scaling of relational databases.

### Problems Solved
| Problem | Sharding Solution |
|---------|-------------------|
| Single node capacity | Distribute data across N nodes |
| Write bottleneck | N nodes = N× write throughput |
| Storage limit | Each shard holds 1/N of data |
| Single point of failure | Shard failure affects subset only |

---

## 2. Real-World Motivation

### Instagram
- Sharded PostgreSQL by user ID (initially); billions of users
- Each shard: subset of users and their data (photos, likes)
- Vitess-like patterns for MySQL at scale

### YouTube (Google)
- Shards by video ID; massive scale
- Each shard: videos, metadata, comments
- Consistent hashing for even distribution

### Uber
- Sharded by city/region and entity (trips, riders)
- Geographic sharding for latency
- Cross-shard for cross-city trips (complex)

### Facebook
- Sharded MySQL; user_id as shard key
- TAO (key-value) for social graph; different partitioning
- Billions of users; thousands of shards

### Twitter
- Manhattan sharded; tweet ID, user ID
- Timeline data partitioned for read scaling
- Fan-out vs fan-in strategies

### Vitess (YouTube/Google)
- MySQL sharding layer; handles routing, resharding
- VTGate: routing; Vttablet: per-shard
- Used by Slack, GitHub, Square

### CockroachDB
- Automatic sharding; range-based
- No manual shard key selection; system manages
- Global distribution

---

## 3. Architecture Diagrams

### Sharding Topology
```
┌─────────────────────────────────────────────────────────────────────────┐
│                     SHARDED DATABASE ARCHITECTURE                         │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│                    ┌─────────────────────┐                               │
│                    │   APPLICATION       │                               │
│                    │   / Load Balancer   │                               │
│                    └──────────┬──────────┘                               │
│                               │                                          │
│                    ┌──────────▼──────────┐                               │
│                    │   SHARD ROUTER      │  ← Determines shard from key  │
│                    │   (Middleware)      │                               │
│                    └──────────┬──────────┘                               │
│           ┌───────────────────┼───────────────────┐                     │
│           │                   │                   │                     │
│           ▼                   ▼                   ▼                     │
│   ┌───────────────┐   ┌───────────────┐   ┌───────────────┐            │
│   │   SHARD 0     │   │   SHARD 1     │   │   SHARD 2     │            │
│   │  user_id 0-   │   │  user_id 33M- │   │  user_id 66M- │            │
│   │  33M          │   │  66M          │   │  100M         │            │
│   │  [Replica]    │   │  [Replica]    │   │  [Replica]    │            │
│   └───────────────┘   └───────────────┘   └───────────────┘            │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Sharding Strategies
```
┌─────────────────────────────────────────────────────────────────────────┐
│                    SHARDING STRATEGIES                                    │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  RANGE-BASED                    HASH-BASED                               │
│  ┌────────────────────┐        ┌────────────────────┐                  │
│  │ Shard 0: A-M        │        │ hash(key) % N       │                  │
│  │ Shard 1: N-S        │        │ Shard 0: keys→0      │                  │
│  │ Shard 2: T-Z        │        │ Shard 1: keys→1      │                  │
│  │                     │        │ Shard 2: keys→2      │                  │
│  │ Pro: Range queries  │        │ Pro: Even distribution│                │
│  │ Con: Hotspots       │        │ Con: No range scan   │                  │
│  └────────────────────┘        └────────────────────┘                  │
│                                                                          │
│  DIRECTORY-BASED                 CONSISTENT HASHING                      │
│  ┌────────────────────┐        ┌────────────────────┐                  │
│  │ Lookup table:      │        │ Ring: 0..2^32       │                  │
│  │ user_123 → Shard 2 │        │ hash(key)→ring pos  │                  │
│  │ user_456 → Shard 0 │        │ Add/remove node:    │                  │
│  │                     │        │ only K/N keys move │                  │
│  │ Pro: Flexibility   │        │ Pro: Minimal rebal  │                  │
│  │ Con: Lookup cost   │        │ Con: Possible skew  │                  │
│  └────────────────────┘        └────────────────────┘                  │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Hotspot Problem
```
┌─────────────────────────────────────────────────────────────────────────┐
│                         HOTSPOT PROBLEM                                   │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   BAD: Shard by date (range)                                             │
│   ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐                     │
│   │ Jan     │  │ Feb     │  │ Mar     │  │ Apr     │  ← Today: ALL       │
│   │ (cold)  │  │ (cold)  │  │ (cold)  │  │ (HOT!)  │    writes here      │
│   └─────────┘  └─────────┘  └─────────┘  └─────────┘                     │
│                                                                          │
│   BETTER: Hash(user_id) or composite (user_id, date)                      │
│   ┌─────────┐  ┌─────────┐  ┌─────────┐                                 │
│   │ Shard 0 │  │ Shard 1 │  │ Shard 2 │  ← Writes distributed           │
│   │ mixed   │  │ mixed   │  │ mixed   │                                 │
│   └─────────┘  └─────────┘  └─────────┘                                 │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Scatter-Gather Query
```
┌─────────────────────────────────────────────────────────────────────────┐
│                    SCATTER-GATHER (Cross-Shard Query)                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   Query: SELECT * FROM orders WHERE status = 'pending'                    │
│   (status is NOT shard key)                                              │
│                                                                          │
│        APPLICATION                                                       │
│             │                                                            │
│             │  Scatter: Send query to ALL shards                         │
│             ├──────────────┬──────────────┬──────────────┐              │
│             ▼              ▼              ▼              ▼              │
│        Shard 0        Shard 1        Shard 2        Shard 3              │
│             │              │              │              │               │
│             │  Gather: Merge results                                     │
│             ◀──────────────┴──────────────┴──────────────┘              │
│             │                                                            │
│        Return merged result (expensive! Avoid if possible)                 │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Consistent Hashing Ring
```
┌─────────────────────────────────────────────────────────────────────────┐
│                    CONSISTENT HASHING RING                                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│                        0 / 2^32                                          │
│                           │                                              │
│                    ┌──────┴──────┐                                       │
│                   ╱               ╲                                      │
│              Shard B              Shard A                                │
│                 │                    │                                   │
│                 │   Keys hashed      │                                   │
│                 │   to ring pos     │                                   │
│                 │   → assigned to   │                                   │
│                 │   clockwise shard │                                   │
│                 │                    │                                   │
│                    ╲               ╱                                      │
│                    ┌──────┬──────┐                                       │
│                           │                                              │
│                      Shard C                                             │
│                                                                          │
│   Add Shard D: Only keys between C and D move (~25% with 4 shards)       │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Core Mechanics

### Shard Key Selection
- **Cardinality**: High cardinality avoids hotspots (user_id > status)
- **Distribution**: Even distribution (hash) vs skewed (range)
- **Query pattern**: Most queries must include shard key to avoid scatter-gather
- **Write pattern**: Avoid sequential keys (timestamp) as sole shard key

### Routing Logic
```python
def get_shard(user_id):
    return shards[hash(user_id) % num_shards]
```

### Cross-Shard Joins
- **Application-level**: Fetch from shard A, then shard B; join in app
- **Denormalization**: Embed related data to avoid joins
- **Global table**: Small reference data replicated to all shards
- **Avoid**: Scatter-gather join (very expensive)

### Resharding
- **Double write**: Write to old and new shards during migration
- **Background migration**: Copy data; catch up; switch
- **Vitess**: VReplication for resharding with minimal downtime

---

## 5. Numbers

| Metric | Value |
|--------|-------|
| Single MySQL instance limit | ~1-2TB data, 10K-50K TPS |
| Shards at Instagram scale | Hundreds to thousands |
| Resharding duration | Hours to days (depends on data size) |
| Cross-shard query cost | N× single-shard (N = num shards) |
| Consistent hashing rebalance | K/N keys move when adding N-th node |

### Scale Examples
- **Facebook**: Thousands of MySQL shards
- **Uber**: Shards per region; cross-region for trips
- **YouTube**: Vitess manages 1000s of shards

---

## 6. Tradeoffs

### Range vs Hash
| Aspect | Range | Hash |
|--------|-------|------|
| Distribution | Can be skewed | Even |
| Range queries | Efficient (single shard) | Scatter-gather |
| Hotspots | Likely (e.g., recent data) | Unlikely |
| Resharding | Split/merge ranges | Rehash all |

### Sharding vs Replication
| Sharding | Replication |
|----------|-------------|
| Splits data | Copies data |
| Scale writes | Scale reads |
| Complex | Simpler |
| Both used together | |

---

## 7. Variants / Implementations

### Sharding Implementations
- **Vitess**: MySQL; automatic sharding, resharding
- **CockroachDB**: Range sharding; automatic split/merge
- **MongoDB**: Sharding built-in; range or hash
- **Cassandra**: Partition key = shard; automatic
- **Spanner**: Automatic; global

### Strategies
- **Vertical sharding**: Split by table (users on one, orders on another)
- **Horizontal sharding**: Split rows by key (this document)
- **Geographic**: Shard by region for latency

---

## 8. Scaling Strategies

1. **Add shards**: Increase N; rebalance (hash) or split (range)
2. **Consistent hashing**: Minimal rebalance on add/remove
3. **Virtual shards**: Map multiple virtual to physical; easier rebalancing
4. **Read replicas per shard**: Scale reads within shard

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|-------------|
| Shard failure | Data unavailable | Replica failover |
| Hot shard | Overload | Reshard; better key |
| Cross-shard transaction | Complex | Avoid; saga pattern |
| Resharding failure | Data inconsistency | Double-write; verification |
| Wrong routing | Wrong/missing data | Consistent hashing; validation |

---

## 10. Performance Considerations

- **Co-location**: Keep related data on same shard (user + orders)
- **Avoid scatter-gather**: Design queries to include shard key
- **Connection pooling**: Per-shard pools
- **Caching**: Cache by shard key

---

## 11. Use Cases

| Use Case | Shard Key | Strategy |
|----------|-----------|----------|
| Multi-tenant SaaS | tenant_id | Hash |
| Social feed | user_id | Hash |
| Time-series | (metric_id, time_bucket) | Range + hash |
| E-commerce | user_id or order_id | Hash |
| Geographic | region/country | Directory or range |

---

## 12. Comparison Tables

### Sharding Strategy Comparison
| Strategy | Distribution | Range Query | Resharding | Use Case |
|----------|--------------|-------------|------------|----------|
| Range | Skewed | Good | Split/merge | Time-series |
| Hash | Even | Poor | Rehash | User data |
| Directory | Flexible | Depends | Update map | Complex |
| Consistent Hash | Even | Poor | Minimal | Dynamic cluster |

### Shard Key Selection
| Key Type | Cardinality | Distribution | Query Fit |
|----------|-------------|--------------|-----------|
| user_id | High | Good | Most queries by user |
| status | Low | Bad (hotspots) | Avoid |
| (user_id, date) | High | Good | User + time range |
| tenant_id | Medium | Good | Multi-tenant |

---

## 13. Code or Pseudocode

### Hash-Based Shard Routing
```python
def route_query(shard_key, query_type):
    shard_id = hash(shard_key) % NUM_SHARDS
    return get_connection(shard_id)
```

### Scatter-Gather
```python
def scatter_gather(query):
    results = []
    for shard in shards:
        results.append(shard.execute(query))
    return merge(results)  # Sort, dedupe, etc.
```

### Consistent Hashing (Simplified)
```python
class ConsistentHash:
    def __init__(self, nodes, replicas=100):
        self.ring = {}
        for node in nodes:
            for i in range(replicas):
                self.ring[hash(f"{node}:{i}")] = node
        self.sorted_keys = sorted(self.ring.keys())
    
    def get_node(self, key):
        h = hash(key)
        for k in self.sorted_keys:
            if h <= k:
                return self.ring[k]
        return self.ring[self.sorted_keys[0]]
```

### Resharding: Double Write
```python
def write_with_migration(key, value):
    old_shard = old_routing(key)
    new_shard = new_routing(key)
    old_shard.write(key, value)
    new_shard.write(key, value)
    # Background: verify; eventually switch routing
```

---

## 14. Interview Discussion

### Key Points
1. **Shard key is critical**: Determines distribution and query ability
2. **Avoid scatter-gather**: Design so most queries include shard key
3. **Hotspots**: Sequential keys (timestamp) or low cardinality = bad
4. **Cross-shard transactions**: Avoid; use saga if necessary
5. **Consistent hashing**: Minimal rebalance when adding nodes

### Common Questions
- **Q**: "How do you choose a shard key?"
  - **A**: High cardinality, even distribution, matches query pattern (most queries filter by it)
- **Q**: "What's the hotspot problem?"
  - **A**: One shard gets disproportionate load; e.g., sharding by date puts all today's writes on one shard
- **Q**: "How do you do cross-shard joins?"
  - **A**: Avoid; denormalize, or application-level (fetch from each shard, join in app)
- **Q**: "How does consistent hashing help?"
  - **A**: Adding/removing nodes moves only K/N keys, not all keys

---

## 15. Shard Sizing and Capacity Planning

### Sizing Guidelines
- **Target**: 50-80% capacity per shard to allow growth
- **Avoid**: Shards > 2TB (recovery, backup complexity)
- **Sweet spot**: 100-500GB per shard for MySQL/PostgreSQL
- **Replication**: Each shard has 1-2 replicas minimum

### Capacity Formula
```
Shards needed = (Total data × Growth factor) / (Target shard size × Replication factor)
Example: 10TB × 1.5 / (200GB × 2) ≈ 38 shards
```

---

## 16. Cross-Shard Transaction Patterns

### Pattern 1: Avoid
- Design so transaction stays within one shard
- Co-locate related data (user + orders by user_id)

### Pattern 2: Application-Level 2PC
- Coordinator in application
- Prepare each shard; commit or abort all
- Risk: Coordinator failure leaves inconsistent state

### Pattern 3: Saga
- Sequence of local transactions
- Compensating transactions on failure
- Accept eventual consistency

### Pattern 4: Distributed DB
- Use CockroachDB, Spanner for automatic distributed transactions
- Higher latency; more complex

---

## 17. Monitoring and Observability

### Key Metrics
- **Per-shard**: QPS, latency, error rate, replication lag
- **Hot shards**: Identify skewed load
- **Connection pool**: Utilization per shard
- **Disk**: Space per shard; growth rate

### Alerts
- Shard down
- Replication lag > threshold
- Hot shard (QPS 2x average)
- Disk > 80% full

---

## 18. Vitess and CockroachDB Comparison

| Aspect | Vitess | CockroachDB |
|--------|--------|-------------|
| Underlying | MySQL | Custom (RocksDB) |
| Sharding | Manual key choice | Automatic range |
| Resharding | VReplication | Automatic split |
| SQL | MySQL compatible | PostgreSQL-like |
| Use case | Existing MySQL at scale | New system, global |

---

## 19. Shard Key Design Patterns

### Pattern: Entity ID
- user_id, tenant_id, order_id
- Pro: Even distribution; co-locates entity data
- Con: Cross-entity queries need scatter-gather

### Pattern: Composite (Entity + Time)
- (user_id, created_at) or (tenant_id, date)
- Pro: Time-range queries within entity
- Con: Hot partition if all writes for "today"

### Pattern: Hash of Natural Key
- hash(email) % N
- Pro: Even distribution
- Con: Cannot range query by email

### Pattern: Geographic
- region, country, datacenter
- Pro: Data locality; compliance
- Con: Uneven distribution (e.g., US >> others)

---

## 20. Resharding Checklist

1. **Plan**: New shard count; key mapping
2. **Dual-write**: Write to old and new during migration
3. **Backfill**: Copy historical data
4. **Verification**: Checksum; spot checks
5. **Cutover**: Switch reads to new; stop old writes
6. **Cleanup**: Remove old shards; update config
7. **Rollback plan**: Keep old shards until verified
