# Database Types and Selection Guide

## Staff+ Engineer Deep Dive for FAANG Interviews

---

## 1. Concept Overview

### Definition
The **database landscape** encompasses multiple categories—relational, document, wide-column, key-value, graph, time-series, search, and NewSQL—each optimized for specific data models, access patterns, and scalability requirements. **Selection** is the process of matching database type to use case based on consistency, scale, and operational constraints.

### Purpose
- **Right tool for the job**: No single database excels at everything
- **Polyglot persistence**: Use multiple databases in one system
- **Informed tradeoffs**: Understand strengths/weaknesses before commitment

### Why It Exists
Different problems need different solutions. Relational databases excel at transactions; key-value at caching; graph at relationships; time-series at metrics. The proliferation of database types reflects the diversity of modern data challenges.

### Problems Solved
| Problem | Database Type | Solution |
|---------|---------------|----------|
| Complex transactions | Relational | ACID, JOINs |
| Flexible schema | Document | Schema-on-read |
| High write throughput | Wide-column | LSM, partition |
| Sub-ms lookup | Key-value | Hash index |
| Relationship queries | Graph | Index-free adjacency |
| Metrics, events | Time-series | Compression, retention |
| Full-text search | Search | Inverted index |
| Global scale + SQL | NewSQL | Distributed + SQL |

---

## 2. Real-World Motivation

### Relational (PostgreSQL, MySQL)
- **Instagram**: PostgreSQL for users, payments
- **Uber**: MySQL for trips, riders
- **Facebook**: MySQL for core data
- **Airbnb**: MySQL for bookings

### Document (MongoDB)
- **eBay**: Product catalogs, flexible attributes
- **Forbes**: Content management
- **SAP**: Flexible document storage

### Wide-Column (Cassandra)
- **Netflix**: Viewing history, recommendations
- **Apple**: iCloud metadata
- **Instagram**: Feed data (some)

### Key-Value (Redis, DynamoDB)
- **Twitter**: Manhattan for timelines
- **Amazon**: DynamoDB for shopping cart
- **Facebook**: TAO, Memcached for cache

### Graph (Neo4j)
- **LinkedIn**: People You May Know
- **Pinterest**: Recommendations
- **Uber**: Fraud detection (entity graphs)

### Time-Series (InfluxDB, TimescaleDB)
- **Tesla**: Vehicle telemetry
- **Bosch**: IoT sensor data
- **Robinhood**: Stock tick data

### NewSQL (Spanner, CockroachDB)
- **Google**: Spanner for Ads, Gmail metadata
- **Cockroach Labs**: CockroachDB for financial services
- **TiDB**: Used in China for scale

---

## 3. Architecture Diagrams

### Database Type Landscape
```
┌─────────────────────────────────────────────────────────────────────────┐
│                    DATABASE TYPE LANDSCAPE                                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   RELATIONAL          DOCUMENT           WIDE-COLUMN                     │
│   ┌──────────┐       ┌──────────┐       ┌──────────┐                   │
│   │ Tables   │       │ Docs     │       │ Rows     │                   │
│   │ Rows     │       │ Nested   │       │ Cols     │                   │
│   │ Cols     │       │ Flexible │       │ Sparse   │                   │
│   │ JOINs    │       │ No JOINs │       │ Partition│                   │
│   └──────────┘       └──────────┘       └──────────┘                   │
│                                                                          │
│   KEY-VALUE           GRAPH              TIME-SERIES                      │
│   ┌──────────┐       ┌──────────┐       ┌──────────┐                   │
│   │ Key→Val  │       │ Nodes     │       │ Timestamp│                   │
│   │ O(1)     │       │ Edges     │       │ Metrics  │                   │
│   │ Simple   │       │ Traverse  │       │ Compress │                   │
│   └──────────┘       └──────────┘       └──────────┘                   │
│                                                                          │
│   SEARCH               NEWSQL                                              │
│   ┌──────────┐       ┌──────────┐                                       │
│   │ Inverted │       │ SQL +    │                                       │
│   │ Index    │       │ Distribute│                                      │
│   │ Full-text│       │ ACID     │                                       │
│   └──────────┘       └──────────┘                                       │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Decision Framework Flowchart
```
┌─────────────────────────────────────────────────────────────────────────┐
│                    DATABASE SELECTION FLOWCHART                           │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   START: What's your primary access pattern?                             │
│                    │                                                      │
│    ┌───────────────┼───────────────┬───────────────┐                    │
│    ▼               ▼               ▼               ▼                     │
│  Key lookup?   Relationships?   Time-ordered?   Full-text?                │
│    │               │               │               │                     │
│    ▼               ▼               ▼               ▼                     │
│  Key-Value      Graph          Time-Series      Search                   │
│  (Redis,        (Neo4j,        (InfluxDB,       (Elasticsearch)          │
│   DynamoDB)     Neptune)       TimescaleDB)                              │
│                                                                          │
│   Need ACID transactions + complex queries?                              │
│                    │                                                      │
│    ┌───────────────┴───────────────┐                                    │
│    ▼                               ▼                                     │
│  Single region?                Global scale?                             │
│    │                               │                                     │
│    ▼                               ▼                                     │
│  PostgreSQL, MySQL            Spanner, CockroachDB                       │
│                                                                          │
│   Flexible schema + document model?                                      │
│                    │                                                      │
│    ▼                                                                     │
│  MongoDB, CouchDB                                                         │
│                                                                          │
│   Write-heavy + partition by key?                                         │
│                    │                                                      │
│    ▼                                                                     │
│  Cassandra, ScyllaDB, HBase                                               │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Polyglot Persistence Example
```
┌─────────────────────────────────────────────────────────────────────────┐
│                    POLYGLOT PERSISTENCE (E-commerce)                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   ┌─────────────┐                                                        │
│   │ Application │                                                        │
│   └──────┬──────┘                                                        │
│          │                                                               │
│    ┌─────┼─────┬─────────┬─────────┬─────────┐                         │
│    ▼     ▼     ▼         ▼         ▼         ▼                          │
│  ┌────┐ ┌────┐ ┌────┐  ┌────┐  ┌────┐  ┌────┐                            │
│  │PG  │ │Mongo│ │Redis│ │Cass│ │ES  │  │S3  │                            │
│  │    │ │    │ │    │  │    │  │    │  │    │                            │
│  │Orders│ │Prod│ │Cart│ │Events│ │Search│ │Blobs│                         │
│  │Users│ │Cat │ │Sess│ │    │  │    │  │    │                            │
│  └────┘ └────┘ └────┘  └────┘  └────┘  └────┘                            │
│  ACID   Flexible  Cache   Logs   Full-text  Files                        │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Core Mechanics

### Relational
- Tables, rows, columns; normalized
- ACID via WAL, locks, MVCC
- SQL; JOINs, transactions
- Scale: Vertical + sharding

### Document
- Collections of documents (JSON/BSON)
- Schema-on-read; nested structures
- Indexes on any field; aggregation pipeline
- Scale: Horizontal; replica sets

### Wide-Column
- Partition key + clustering columns
- Sparse; column families
- LSM-tree; tunable consistency
- Scale: Add nodes; automatic distribution

### Key-Value
- Hash table; get/put/delete
- In-memory (Redis) or persistent (DynamoDB)
- O(1) lookup
- Scale: Partition by key

### Graph
- Nodes (vertices), edges (relationships)
- Index-free adjacency; pattern matching
- Cypher, Gremlin
- Scale: Horizontal (Neo4j Enterprise)

### Time-Series
- Timestamp + metrics
- Compression (delta encoding)
- Retention policies; downsampling
- Scale: Shard by time, metric

### Search
- Inverted index; tokenization
- Relevance scoring; faceting
- Scale: Shards, replicas

### NewSQL
- SQL interface + distributed storage
- ACID across nodes
- Scale: Automatic sharding

---

## 5. Numbers

| Type | Latency (p99) | Throughput | Scale |
|------|----------------|------------|-------|
| Relational | 1-20ms | 10K-50K TPS | TBs per instance |
| Document | 1-10ms | 50K-100K | TBs |
| Wide-column | 5-50ms | 100K+ | PBs |
| Key-value | <1ms | 100K+ | Depends |
| Graph | 1-100ms | 10K | Billions of edges |
| Time-series | 1-10ms | 1M+ writes/sec | PBs |
| Search | 10-100ms | 10K | PBs |
| NewSQL | 5-20ms | 50K+ | Global |

---

## 6. Tradeoffs

### Consistency vs Availability
| Type | Default Consistency | Availability |
|------|---------------------|--------------|
| Relational | Strong | Single leader |
| Document | Tunable | Replica sets |
| Wide-column | Tunable | High |
| Key-value | Tunable | High |
| Graph | Strong (Neo4j) | Depends |
| NewSQL | Strong | High (distributed) |

### Operational Complexity
| Type | Complexity | Managed Options |
|------|------------|-----------------|
| Relational | Medium | RDS, Cloud SQL |
| Document | Medium | Atlas, DocumentDB |
| Wide-column | High | Astra, Keyspaces |
| Key-value | Low | ElastiCache, DynamoDB |
| Graph | Medium | Neptune |
| NewSQL | High | Spanner, Cockroach Cloud |

---

## 7. Variants / Implementations

### By Type
| Type | Examples |
|------|----------|
| Relational | PostgreSQL, MySQL, Oracle, SQL Server |
| Document | MongoDB, CouchDB, Firestore, DocumentDB |
| Wide-column | Cassandra, HBase, ScyllaDB, Bigtable |
| Key-value | Redis, DynamoDB, etcd, Memcached |
| Graph | Neo4j, Amazon Neptune, JanusGraph |
| Time-series | InfluxDB, TimescaleDB, Prometheus |
| Search | Elasticsearch, OpenSearch, Solr |
| NewSQL | Spanner, CockroachDB, TiDB, YugabyteDB |

---

## 8. Scaling Strategies

| Type | Scale Strategy |
|------|----------------|
| Relational | Read replicas, sharding |
| Document | Sharding, replica sets |
| Wide-column | Add nodes |
| Key-value | Partition, cluster |
| Graph | Sharding (complex) |
| Time-series | Retention, downsampling, cluster |
| Search | Shards, replicas |
| NewSQL | Automatic |

---

## 9. Failure Scenarios

| Scenario | Relational | NoSQL | Mitigation |
|----------|------------|-------|------------|
| Node failure | Failover | Replica serves | Replication |
| Partition | 2PC blocks | AP or CP | Design choice |
| Wrong type | Poor fit | Migration | Prototype first |

---

## 10. Performance Considerations

- **Access pattern**: Match DB to query shape
- **Consistency level**: Lower = faster
- **Indexing**: Right indexes for workload
- **Connection**: Pooling, proximity

---

## 11. Use Cases

| Use Case | Recommended Type | Example |
|----------|------------------|---------|
| E-commerce orders | Relational | PostgreSQL |
| Product catalog | Document | MongoDB |
| User sessions | Key-value | Redis |
| Activity feed | Wide-column | Cassandra |
| Social graph | Graph | Neo4j |
| Metrics | Time-series | InfluxDB |
| Search | Search | Elasticsearch |
| Global ledger | NewSQL | Spanner |

---

## 12. Comparison Tables

### Comprehensive Comparison
| Type | Strengths | Weaknesses | Consistency | Scaling | Ideal Use |
|------|-----------|------------|-------------|---------|-----------|
| Relational | ACID, JOINs, mature | Scale complexity | Strong | Vertical, shard | Transactions |
| Document | Flexible, nested | No JOINs | Tunable | Horizontal | Content, catalogs |
| Wide-column | Write scale, partition | No JOINs | Tunable | Horizontal | Events, logs |
| Key-value | Fast, simple | No queries | Tunable | Horizontal | Cache, session |
| Graph | Relationships | Scale | Strong | Horizontal | Social, fraud |
| Time-series | Metrics, compression | Limited queries | Eventual | Horizontal | IoT, metrics |
| Search | Full-text | Not primary store | Eventual | Horizontal | Search |
| NewSQL | SQL + scale | Complexity | Strong | Global | Global apps |

### FAANG Usage Summary
| Company | Databases | Use |
|---------|-----------|-----|
| Google | Spanner, Bigtable, MySQL | Ads, Gmail, YouTube |
| Amazon | DynamoDB, Aurora, Redshift | Cart, relational, analytics |
| Meta | MySQL, TAO, Memcached | Core, graph, cache |
| Netflix | Cassandra, Redis, PostgreSQL | Catalog, cache, billing |
| Uber | MySQL, Cassandra, Redis | Trips, events, cache |

---

## 13. Code or Pseudocode

### Selection Pseudocode
```python
def select_database(requirements):
    if requirements.need_acid and requirements.complex_queries:
        if requirements.global_scale:
            return "Spanner/CockroachDB"
        return "PostgreSQL/MySQL"
    if requirements.access_pattern == "key_lookup":
        return "Redis/DynamoDB"
    if requirements.access_pattern == "relationships":
        return "Neo4j/Neptune"
    if requirements.access_pattern == "time_series":
        return "InfluxDB/TimescaleDB"
    if requirements.flexible_schema:
        return "MongoDB"
    if requirements.write_heavy and requirements.partition_key:
        return "Cassandra"
    return "PostgreSQL"  # Default
```

### Polyglot Example
```python
# Order: PostgreSQL (ACID)
db.sql.execute("INSERT INTO orders ...")

# Product catalog: MongoDB (flexible)
db.mongo.products.insert_one({...})

# Cart: Redis (fast)
redis.set(f"cart:{user_id}", json.dumps(cart))

# Search: Elasticsearch
es.index(index="products", body={...})
```

---

## 14. Interview Discussion

### Key Points
1. **Polyglot persistence**: Standard; use multiple DBs
2. **Access pattern first**: Match DB to how you query
3. **Consistency requirements**: Drive ACID vs eventual choice
4. **Operational burden**: Consider managed services
5. **Migration cost**: Hard to change; choose carefully

### Common Questions
- **Q**: "How do you choose between SQL and NoSQL?"
  - **A**: SQL for transactions, complex queries, strong consistency; NoSQL for scale, flexibility, specific patterns
- **Q**: "When would you use a graph database?"
  - **A**: Multi-hop relationships (friends of friends), recommendations, fraud detection
- **Q**: "What's NewSQL?"
  - **A**: Distributed database with SQL and ACID; Spanner, CockroachDB
- **Q**: "How do you handle multiple databases in one system?"
  - **A**: Polyglot persistence; each service owns its DB; event-driven sync if needed

---

## 15. Anti-Patterns in Database Selection

### Anti-Pattern 1: Default to Relational
- Not everything needs JOINs and transactions
- Session data, cache: key-value is better
- Consider access pattern first

### Anti-Pattern 2: NoSQL for Everything
- "We need scale" doesn't mean NoSQL
- Many workloads fit single PostgreSQL
- Operational cost of NoSQL is real

### Anti-Pattern 3: Premature Polyglot
- Start with one DB; add when needed
- Each new DB = operational burden
- Prove need before adding

### Anti-Pattern 4: Ignoring Consistency
- "Eventual is fine" can cause bugs
- Read-after-write, inventory: need strong
- Design for consistency from start

---

## 16. Managed vs Self-Hosted

| Aspect | Managed (RDS, Atlas, etc.) | Self-Hosted |
|--------|---------------------------|-------------|
| Ops burden | Low | High |
| Cost | Higher per GB | Lower (if you have ops) |
| Customization | Limited | Full |
| Lock-in | Vendor | None |
| Use when | Focus on product | Need control, cost-sensitive |

---

## 17. Database as a Service (DBaaS) Options

### Relational
- AWS RDS (PostgreSQL, MySQL, Aurora)
- Google Cloud SQL
- Azure Database for PostgreSQL/MySQL

### NoSQL
- MongoDB Atlas
- AWS DynamoDB, DocumentDB
- Cassandra: DataStax Astra, AWS Keyspaces

### NewSQL
- Google Spanner
- CockroachDB Cloud
- TiDB Cloud

---

## 18. Cost Considerations

### Factors
- **Storage**: Per GB/month
- **Compute**: Instance size, replicas
- **I/O**: Some charge per request (DynamoDB)
- **Backup**: Retention, cross-region
- **Data transfer**: Egress fees

### Optimization
- Right-size instances
- Use reserved instances for predictable load
- Archive cold data to cheaper storage
- Consider serverless (Aurora Serverless, DynamoDB on-demand)

---

## 19. Time-Series Database Deep Dive

### Why Specialized?
- Append-only writes (mostly)
- Time-range queries dominant
- Compression: delta encoding, Gorilla (Facebook)
- Retention: Downsample old data; delete after period
- Examples: InfluxDB, TimescaleDB (PostgreSQL extension), Prometheus

### When to Use
- IoT sensor data
- Application metrics
- Financial tick data
- Log aggregation (with time component)

---

## 20. Graph Database Use Cases (Expanded)

| Use Case | Why Graph |
|----------|-----------|
| Social: "Friends of friends" | Multi-hop traversal; SQL requires recursive CTE or N joins |
| Recommendation | "Users who bought X also bought" = path finding |
| Fraud detection | Entity resolution; pattern matching across accounts |
| Knowledge graph | Wikipedia, Google; entities and relationships |
| Network topology | IT infrastructure; dependencies |
