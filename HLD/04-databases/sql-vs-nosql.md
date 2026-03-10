# SQL vs NoSQL Databases

## Staff+ Engineer Deep Dive for FAANG Interviews

---

## 1. Concept Overview

### Definition
**SQL (Structured Query Language)** databases are relational database management systems (RDBMS) that store data in structured tables with rows and columns, enforcing relationships through foreign keys and supporting ACID transactions. **NoSQL (Not Only SQL)** databases encompass a variety of non-relational data models designed for different access patterns, scalability requirements, and data structures.

### Purpose
- **SQL**: Provide strong consistency, complex querying, and data integrity for structured data with well-defined schemas
- **NoSQL**: Offer horizontal scalability, schema flexibility, and optimized performance for specific access patterns (document lookups, key-value access, graph traversals, wide-column scans)

### Why It Exists
The relational model emerged in the 1970s (Codd's paper) to solve data redundancy and inconsistency. NoSQL emerged in the late 2000s when web-scale companies (Google, Amazon, Facebook) hit limits of traditional RDBMS: vertical scaling ceilings, rigid schemas for rapidly evolving data, and overkill for simple key-value patterns.

### Problems Solved
| Problem | SQL Solution | NoSQL Solution |
|---------|--------------|----------------|
| Data consistency | ACID transactions | Eventual consistency, tunable |
| Schema evolution | Migrations (expensive) | Schema-on-read (flexible) |
| Horizontal scale | Sharding (complex) | Built-in distribution |
| Complex joins | Native JOIN operations | Denormalization, application joins |
| Simple lookups | Full RDBMS overhead | Minimal key-value stores |

---

## 2. Real-World Motivation

### Instagram (PostgreSQL)
- Uses PostgreSQL for core data: users, photos, likes, comments
- Chose PostgreSQL for ACID guarantees on financial/payment data
- Sharded PostgreSQL when single instance hit limits (~100M users era)
- Uses Redis for caching and session data

### Uber (PostgreSQL → MySQL Migration)
- Originally PostgreSQL; migrated to MySQL for better ecosystem, tooling, and MySQL-specific optimizations
- MySQL's replication and sharding tooling (Vitess) more mature at scale
- Uses Schemaless (in-house) for document-like storage

### Facebook (MySQL + TAO)
- MySQL for primary user data, social graph metadata
- TAO (The Associations and Objects): custom key-value store for social graph
- Billions of reads/sec; MySQL for writes, TAO/Redis for read-through cache

### Twitter (Manhattan)
- Manhattan: distributed key-value store (in-house)
- Optimized for high read throughput, low latency
- Timeline data, user profiles, tweet metadata

### Netflix (Cassandra)
- Cassandra for viewing history, recommendations, metadata
- Multi-region deployment for global availability
- Handles 500M+ members, billions of events/day

### Amazon (DynamoDB + Aurora)
- DynamoDB: shopping cart, session state, metadata
- Aurora: relational workloads needing SQL
- Polyglot persistence: right tool per use case

---

## 3. Architecture Diagrams

### Relational (SQL) Model
```
┌─────────────────────────────────────────────────────────────────┐
│                    RELATIONAL DATABASE                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐     FK      ┌──────────────┐     FK            │
│  │   USERS      │────────────▶│   ORDERS     │──────────────┐   │
│  ├──────────────┤             ├──────────────┤              │   │
│  │ id (PK)      │             │ id (PK)      │              │   │
│  │ name         │             │ user_id (FK) │              │   │
│  │ email        │             │ product_id   │              ▼   │
│  │ created_at   │             │ amount       │    ┌──────────────┐
│  └──────────────┘             └──────────────┘    │  ORDER_ITEMS │
│         │                             │           ├──────────────┤
│         │ FK                          │ FK        │ order_id     │
│         ▼                             ▼           │ product_id   │
│  ┌──────────────┐             ┌──────────────┐    │ quantity     │
│  │   PROFILES   │             │   PRODUCTS   │    └──────────────┘
│  └──────────────┘             └──────────────┘                      │
│                                                                  │
│  Normalization: 3NF, BCNF | Joins at query time                   │
└─────────────────────────────────────────────────────────────────┘
```

### NoSQL Categories Architecture
```
┌─────────────────────────────────────────────────────────────────────────┐
│                        NoSQL DATABASE TYPES                               │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  DOCUMENT STORE (MongoDB)     KEY-VALUE (Redis/DynamoDB)                 │
│  ┌─────────────────────┐      ┌─────────────────────┐                   │
│  │ {                   │      │  key          value  │                   │
│  │   "_id": "user123", │      │  "user:123"   {...}  │                   │
│  │   "name": "Alice",  │      │  "session:xyz {...}  │                   │
│  │   "orders": [...]   │      │  "cart:456   [...]  │                   │
│  │ }                   │      └─────────────────────┘                   │
│  └─────────────────────┘      O(1) lookup by key                         │
│  Nested documents, arrays                                                │
│                                                                          │
│  WIDE-COLUMN (Cassandra)      GRAPH (Neo4j)                              │
│  ┌─────────────────────┐      ┌─────────────────────┐                   │
│  │ RowKey | Col1|Col2  │      │  (User)-[:FRIENDS]  │                   │
│  │ user1  | a   | b    │      │       -[:LIKES]->   │                   │
│  │ user1  | c   | d    │      │  (Post)-[:TAGGED]   │                   │
│  │ user2  | e   | f    │      │       -[:IN]->(Tag) │                   │
│  └─────────────────────┘      └─────────────────────┘                   │
│  Sparse, column families       Nodes + Edges + Properties               │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Schema-on-Write vs Schema-on-Read
```
SCHEMA-ON-WRITE (SQL)              SCHEMA-ON-READ (NoSQL)
┌──────────────────────┐          ┌──────────────────────┐
│ INSERT validated     │          │ INSERT any JSON      │
│ against schema      │          │ No validation        │
│      │               │          │      │               │
│      ▼               │          │      ▼               │
│ Reject if invalid    │          │ Always accepted      │
│      │               │          │      │               │
│      ▼               │          │      ▼               │
│ Query: schema known  │          │ Query: app interprets│
│ JOINs, types clear   │          │ Schema in app code   │
└──────────────────────┘          └──────────────────────┘
```

---

## 4. Core Mechanics

### Relational Model Mechanics
- **Tables**: Rows (tuples) and columns (attributes); each row has unique identifier (primary key)
- **Normalization**: Decompose to reduce redundancy (1NF → 2NF → 3NF → BCNF)
- **Foreign Keys**: Referential integrity; cascading updates/deletes
- **Joins**: Cartesian product + selection; implemented via nested loops, hash joins, or merge joins
- **Transactions**: WAL (Write-Ahead Log), lock manager, MVCC for isolation

### NoSQL Mechanics by Type

**Document Stores (MongoDB)**:
- BSON (Binary JSON) storage; documents in collections
- Indexes: B-tree, text, geospatial
- Aggregation pipeline: $match, $lookup (join), $group
- No cross-document transactions (until 4.0+ with limited support)

**Wide-Column (Cassandra)**:
- Partition key → determines node; clustering columns → sort order within partition
- SSTable + Memtable; LSM-tree based
- Tunable consistency: ONE, QUORUM, ALL
- No joins; denormalize for read path

**Key-Value (DynamoDB/Redis)**:
- Hash partition key or composite (partition + sort key)
- Redis: in-memory, single-threaded, persistence optional
- DynamoDB: SSD-backed, automatic partitioning

**Graph (Neo4j)**:
- Property graph: nodes (entities), edges (relationships), properties
- Index-free adjacency: each node points to neighbors
- Cypher query language: pattern matching

---

## 5. Numbers

| Metric | SQL (PostgreSQL) | MongoDB | Cassandra | Redis |
|--------|------------------|---------|-----------|-------|
| Max connections (typical) | 100-500/instance | 10K+ | 10K+ | 10K+ |
| Read latency (p99) | 1-10ms | 1-5ms | 1-15ms | <1ms |
| Write latency (p99) | 1-20ms | 1-10ms | 5-50ms | <1ms |
| Throughput (writes/sec) | 10K-50K | 50K-100K | 100K+ | 100K+ |
| Data size (single node) | TBs | TBs | PBs (distributed) | GBs (RAM) |
| Join cost | O(n*m) without index | N/A (denormalize) | N/A | N/A |

### Scale Examples
- **Instagram**: 1B+ users, PostgreSQL sharded, billions of photos
- **Netflix**: 500M+ members, Cassandra 1000+ nodes, 2B+ events/day
- **Twitter**: Manhattan handles 100M+ tweets/day, millions QPS
- **Uber**: MySQL clusters, 10M+ trips/day, global replication

---

## 6. Tradeoffs

### ACID vs BASE
| Aspect | ACID (SQL) | BASE (NoSQL) |
|--------|------------|--------------|
| Consistency | Strong, immediate | Eventual |
| Availability | May block on failure | Prioritize availability |
| Partition tolerance | 2PC can block | Typically CP or AP |
| Use case | Financial, inventory | Social, analytics |

### Schema Flexibility
| Schema-on-Write | Schema-on-Read |
|-----------------|----------------|
| Migrations required | Add fields freely |
| Type safety at DB | Type safety in app |
| Slower schema evolution | Faster iteration |
| Better for regulated data | Better for rapid prototyping |

### Denormalization Tradeoffs
| Normalized (SQL) | Denormalized (NoSQL) |
|------------------|----------------------|
| No redundancy | Data duplication |
| Single source of truth | Update multiple places |
| Joins at read time | Pre-joined in document |
| Smaller storage | Larger storage |
| Complex writes | Simple reads |

---

## 7. Variants / Implementations

### SQL Implementations
- **PostgreSQL**: Most feature-rich open source; JSON support, extensions
- **MySQL**: Widely used, simpler; InnoDB engine
- **CockroachDB**: Distributed SQL, Spanner-like
- **Google Spanner**: Globally distributed, strong consistency
- **Aurora**: AWS, MySQL/PostgreSQL compatible, storage separated

### NoSQL Implementations by Category
| Category | Examples |
|----------|----------|
| Document | MongoDB, CouchDB, DocumentDB, Firestore |
| Key-Value | Redis, DynamoDB, etcd, Memcached |
| Wide-Column | Cassandra, HBase, ScyllaDB, Bigtable |
| Graph | Neo4j, Amazon Neptune, JanusGraph |

---

## 8. Scaling Strategies

### SQL Scaling
1. **Vertical**: Bigger machine (hits ceiling ~1-2TB RAM)
2. **Read replicas**: Replicate for read scaling
3. **Sharding**: Partition by key (user_id, tenant_id); complex
4. **Connection pooling**: PgBouncer, ProxySQL

### NoSQL Scaling
1. **Horizontal**: Add nodes; data auto-distributed
2. **Partition key design**: Critical for even distribution
3. **Replication factor**: Cassandra RF=3 typical
4. **Caching layer**: Redis in front of any DB

---

## 9. Failure Scenarios

| Scenario | SQL Impact | NoSQL Impact |
|----------|------------|--------------|
| Single node failure | Failover to replica; brief unavailability | Replica serves; eventual consistency |
| Network partition | 2PC may block; split-brain risk | AP: serve stale; CP: reject writes |
| Corrupt data | WAL replay, backups | Repair, anti-entropy |
| Hot partition | Single shard overloaded | Same; partition key design critical |
| Schema migration failure | Rollback complex | Less common; schema flexible |

---

## 10. Performance Considerations

### SQL
- **Index selection**: B-tree for equality/range; avoid over-indexing (write cost)
- **Query planning**: EXPLAIN ANALYZE; avoid N+1
- **Connection pooling**: Essential at scale
- **Lock contention**: MVCC helps; watch for long transactions

### NoSQL
- **Partition key**: Must match query pattern; avoid hot keys
- **Denormalization**: Trade write complexity for read speed
- **Consistency level**: Lower = faster but weaker
- **Compaction**: Cassandra; can cause latency spikes

---

## 11. Use Cases

### Choose SQL When
- Complex transactions (banking, inventory)
- Complex queries with joins
- Strong consistency required
- Mature ecosystem, tooling
- Regulatory compliance (audit trails)

### Choose NoSQL When
- Simple access patterns (key-value, document by ID)
- Need horizontal scale from day one
- Schema evolves rapidly
- High write throughput
- Geographic distribution (multi-region)

---

## 12. Comparison Tables

### Comprehensive SQL vs NoSQL
| Dimension | SQL | NoSQL Document | NoSQL Key-Value | NoSQL Wide-Column | NoSQL Graph |
|-----------|-----|----------------|-----------------|-------------------|-------------|
| Data model | Tables, rows | Documents | Key → Value | Row → Columns | Nodes, edges |
| Schema | Rigid | Flexible | Schema-less | Flexible | Flexible |
| Query | SQL, JOINs | Query API, $lookup | Get by key | Partition + range | Graph traversal |
| Scaling | Vertical, sharding | Horizontal | Horizontal | Horizontal | Horizontal |
| Consistency | ACID | Tunable | Tunable | Tunable | ACID (Neo4j) |
| Best for | Transactions, reports | Content, catalogs | Cache, session | Time-series, logs | Social, fraud |

### When to Choose
| Requirement | Recommendation |
|-------------|----------------|
| Financial transactions | SQL |
| User profiles, preferences | Document |
| Session, cache | Key-Value |
| Time-series, events | Wide-column |
| Social graph, recommendations | Graph |
| Full-text search | Elasticsearch (NoSQL) |
| Need both SQL + scale | NewSQL (Spanner, CockroachDB) |

---

## 13. Code or Pseudocode

### SQL: Normalized Schema + Join
```sql
-- Normalized tables
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255),
    email VARCHAR(255) UNIQUE
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    total DECIMAL(10,2),
    created_at TIMESTAMP
);

-- Query with join
SELECT u.name, o.id, o.total
FROM users u
JOIN orders o ON u.id = o.user_id
WHERE u.email = 'alice@example.com';
```

### NoSQL: Denormalized Document (MongoDB)
```javascript
// Document with embedded orders
db.users.insertOne({
  _id: "user123",
  name: "Alice",
  email: "alice@example.com",
  orders: [
    { orderId: "ord1", total: 99.99, createdAt: ISODate("2024-01-01") },
    { orderId: "ord2", total: 149.99, createdAt: ISODate("2024-01-15") }
  ]
});

// Single query, no join
db.users.findOne(
  { email: "alice@example.com" },
  { name: 1, orders: 1 }
);
```

### NoSQL: Key-Value (Redis)
```python
# Session storage
redis.set("session:xyz123", json.dumps({"user_id": 456, "cart": [...]}), ex=3600)
session = redis.get("session:xyz123")
```

### NoSQL: Wide-Column (Cassandra CQL)
```sql
-- Denormalized for read path
CREATE TABLE user_orders (
    user_id UUID,
    order_date TIMESTAMP,
    order_id UUID,
    total DECIMAL,
    PRIMARY KEY (user_id, order_date, order_id)
) WITH CLUSTERING ORDER BY (order_date DESC);

-- Query: partition by user_id
SELECT * FROM user_orders WHERE user_id = ?;
```

---

## 14. Interview Discussion

### Key Points to Articulate
1. **"It's not either/or"**: Polyglot persistence is standard; use SQL for transactions, NoSQL for specific patterns
2. **"Tradeoffs, not absolutes"**: ACID vs eventual consistency is a spectrum; some NoSQL offers ACID (Neo4j, DynamoDB transactions)
3. **"Schema-on-read"**: Explain with example: adding a field in MongoDB requires no migration; app handles it
4. **"Sharding complexity"**: SQL sharding is hard (cross-shard joins, distributed transactions); NoSQL often designed for it
5. **"Real examples"**: Instagram/PostgreSQL, Netflix/Cassandra, Uber/MySQL—cite specific choices

### Common Follow-Up Questions
- **Q**: "When would you use SQL over NoSQL for a new project?"
  - **A**: When you need transactions, complex queries, or strong consistency; when team knows SQL; when schema is stable
- **Q**: "How do you handle joins in NoSQL?"
  - **A**: Denormalize (embed related data); or application-level joins (fetch multiple documents); or use $lookup in MongoDB
- **Q**: "What's the difference between MongoDB and Cassandra?"
  - **A**: MongoDB: document model, flexible schema, good for hierarchical data; Cassandra: wide-column, partition key critical, optimized for write-heavy, time-series
- **Q**: "When does eventual consistency fail?"
  - **A**: Read-after-write; financial operations; inventory (overselling); need strong consistency

### Red Flags to Avoid
- Saying "NoSQL is always faster"
- Ignoring consistency requirements
- Over-normalizing in NoSQL
- Underestimating operational complexity of distributed NoSQL

---

## 15. CAP Theorem and Database Choice

### CAP Explained
- **Consistency**: Every read receives most recent write
- **Availability**: Every request receives response
- **Partition tolerance**: System works despite network partitions

### Theorem
In presence of network partition, you can only guarantee 2 of 3. In practice, partitions happen, so choice is CP vs AP.

### SQL vs NoSQL in CAP
| Type | Typical Choice | Rationale |
|------|----------------|------------|
| PostgreSQL, MySQL | CP | Prefer consistency; block on partition |
| Cassandra, DynamoDB | AP | Prefer availability; eventual consistency |
| Spanner, CockroachDB | CP + availability | Use consensus (Paxos/Raft) for both |

---

## 16. Migration: SQL to NoSQL (and Vice Versa)

### SQL → NoSQL
- **When**: Scale limits, schema flexibility needs
- **Challenges**: Denormalization design, consistency semantics, operational learning
- **Strategy**: Dual-write, backfill, cutover; or strangler pattern

### NoSQL → SQL
- **When**: Need transactions, complex queries, reporting
- **Challenges**: Schema design from documents, data migration
- **Strategy**: ETL pipeline; run both during transition

### Hybrid Approach
- Keep SQL for transactional core
- Add NoSQL for specific workloads (cache, search, events)
- Sync via events or dual-write

---

## 17. Normalization vs Denormalization (Deep Dive)

### Normalization Levels
- **1NF**: Atomic values, no repeating groups
- **2NF**: 1NF + no partial dependencies (non-key depends on full key)
- **3NF**: 2NF + no transitive dependencies
- **BCNF**: Every determinant is candidate key

### Denormalization in NoSQL
- **Embedding**: Put related data in same document (orders in user doc)
- **Reference**: Store ID; fetch separately (like foreign key but no JOIN)
- **Hybrid**: Embed frequently accessed; reference rarely accessed
- **Tradeoff**: Write amplification (update in multiple places) vs read efficiency

---

## 18. NewSQL: Best of Both Worlds?

### What is NewSQL
- SQL interface + ACID + horizontal scale
- Examples: Google Spanner, CockroachDB, TiDB, YugabyteDB
- Typically use distributed consensus (Paxos/Raft) for consistency

### When to Consider
- Need SQL + global distribution
- Outgrowing single-region RDBMS
- Willing to accept operational complexity
- Budget for managed service (Spanner, Cockroach Cloud)

---

## 19. Document Model: Embedding vs Referencing

### Embedding
- Put related data in same document
- Pro: Single read; atomic update
- Con: Document size limit (16MB MongoDB); duplication
- Use: One-to-few; data always accessed together

### Referencing
- Store ID; fetch in separate query
- Pro: No duplication; smaller documents
- Con: Multiple reads; no atomic cross-document
- Use: One-to-many; large collections

### Example
```javascript
// Embed: user with addresses (few)
{ _id: 1, name: "Alice", addresses: [{city: "NYC"}, {city: "LA"}] }

// Reference: user with orders (many)
{ _id: 1, name: "Alice" }
// Separate: orders collection with user_id
```

---

## 20. Consistency Tuning in NoSQL

### Cassandra
- **ONE**: Single node; fast; weak
- **QUORUM**: (N/2)+1; strong for reads and writes
- **ALL**: Every node; slowest; strongest
- **LOCAL_QUORUM**: Quorum within same DC; multi-DC

### DynamoDB
- **Eventually consistent**: Default for reads; cheaper
- **Strongly consistent**: Latest write; 2x cost
- **Transactions**: Multi-item ACID; 2x cost

### MongoDB
- **Read concern**: local, available, majority, linearizable
- **Write concern**: w:1 (fast), w:majority (durable)
