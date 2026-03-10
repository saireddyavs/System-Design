# Change Data Capture (CDC) — Staff+ Engineer Deep Dive

## 1. Concept Overview

### Definition
**Change Data Capture (CDC)** is a pattern and set of technologies that capture changes made to a database (inserts, updates, deletes) and propagate them as events to downstream systems. Instead of applications explicitly publishing events, CDC observes the database's change stream (typically the write-ahead log) and emits change events automatically.

### Purpose
- **Single source of truth**: Database is the only write path; no dual-write
- **Consistency**: All consumers see the same change stream
- **Decoupling**: Downstream systems don't need to be modified when schema changes
- **Audit trail**: Complete history of data changes
- **Real-time sync**: Near real-time propagation to caches, search, analytics

### Problems It Solves
1. **Dual-write problem**: Avoid writing to DB + event bus; eliminates consistency bugs
2. **Out-of-sync systems**: Search index, cache, data warehouse stay in sync with DB
3. **Legacy integration**: Capture changes from systems that can't publish events
4. **Data pipeline**: Feed data lake, analytics, ML pipelines from transactional DB
5. **Microservices sync**: Keep service-owned data stores consistent with source of truth

---

## 2. Real-World Motivation

### Airbnb
- **Search indexing**: MySQL → CDC (Debezium) → Kafka → Elasticsearch
- **Single write**: Application writes to MySQL only; search index updated via CDC
- **Scale**: Millions of listings, real-time search

### Uber
- **Data pipeline**: MySQL/PostgreSQL → CDC → Kafka → data lake, analytics
- **Multi-region**: CDC for cross-region replication
- **Schema evolution**: Handle schema changes in CDC pipeline

### Netflix
- **Cache invalidation**: DB changes → CDC → invalidate/update caches
- **Recommendation**: User behavior in DB → CDC → recommendation pipeline
- **Billing**: Subscription changes → CDC → billing systems

### Stripe
- **Financial data**: Transaction DB → CDC → analytics, reporting
- **Compliance**: Audit trail of all changes
- **Real-time dashboards**: CDC feeds real-time metrics

### LinkedIn
- **Databus**: Internal CDC system (pre-Debezium era)
- **Search, analytics**: DB changes → event stream → consumers
- **Schema registry**: Evolve schemas across pipeline

---

## 3. Architecture Diagrams

### CDC Pipeline Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        SOURCE DATABASE                                      │
│  ┌─────────────┐     ┌─────────────────────────────────────────────────┐   │
│  │ Application │────▶│  MySQL / PostgreSQL / MongoDB / etc.             │   │
│  │   (writes)  │     │  WAL / binlog / oplog (change log)               │   │
│  └─────────────┘     └─────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        │ CDC reads WAL (non-invasive)
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                     CDC CONNECTOR (e.g., Debezium)                           │
│  - Parses WAL/binlog/oplog                                                  │
│  - Converts to canonical format (insert/update/delete)                      │
│  - Tracks offset (resume from last position)                                │
└─────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                     MESSAGE BROKER (e.g., Kafka)                             │
│  Topic: db.server.inventory.orders                                          │
│  [before][after][op][ts_ms][source]                                          │
└─────────────────────────────────────────────────────────────────────────────┘
                                        │
                    ┌───────────────────┼───────────────────┐
                    ▼                   ▼                   ▼
              ┌──────────┐       ┌──────────┐       ┌──────────┐
              │Elasticsearch│     │ Data Lake│       │  Cache   │
              │ (Search)   │     │(Analytics)│       │ Invalidate│
              └──────────┘       └──────────┘       └──────────┘
```

### Outbox Pattern Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  APPLICATION (Single Transaction)                                            │
│                                                                              │
│  1. BEGIN TRANSACTION                                                        │
│  2. INSERT INTO orders (...) VALUES (...)                                    │
│  3. INSERT INTO outbox (aggregate_type, aggregate_id, event_type, payload)   │
│     VALUES ('Order', 123, 'OrderCreated', '{"amount": 99.99}')               │
│  4. COMMIT                                                                   │
│                                                                              │
│  Both writes in same transaction → atomic                                    │
└─────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        │ CDC reads outbox table
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  CDC CONNECTOR                                                              │
│  - Watches outbox table                                                     │
│  - Publishes to Kafka topic                                                 │
│  - Deletes/marks outbox rows (optional)                                     │
└─────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  KAFKA: outbox.order.created                                                │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Log-Based vs Trigger-Based vs Polling

```
LOG-BASED (Preferred):
  DB WAL ──▶ CDC reads ──▶ Events
  Pros: Non-invasive, low latency, no schema changes
  Cons: Requires WAL access, DB-specific

TRIGGER-BASED:
  INSERT/UPDATE/DELETE ──▶ Trigger ──▶ changelog table ──▶ CDC polls
  Pros: Works with any DB
  Cons: Overhead on every write, schema changes

POLLING-BASED:
  CDC polls: SELECT * FROM t WHERE updated_at > last_poll
  Pros: Simple, no WAL
  Cons: High latency, missed changes (clock skew), load on DB
```

---

## 4. Core Mechanics

### Log-Based CDC
- **MySQL**: Read binlog (row-based); Debezium MySQL connector
- **PostgreSQL**: Read WAL via logical replication slot; pgoutput or wal2json
- **MongoDB**: Read oplog
- **SQL Server**: Change Data Capture or Change Tracking
- **Oracle**: LogMiner

### Trigger-Based CDC
- **Triggers**: ON INSERT/UPDATE/DELETE → write to changelog table
- **Changelog table**: id, table_name, op, old_values, new_values, ts
- **CDC polls** changelog table
- **Trade-off**: Trigger overhead, schema coupling

### Polling-Based CDC
- **Query**: `SELECT * FROM t WHERE updated_at > :last_ts`
- **Timestamp column**: Must be indexed, monotonic
- **Issues**: Miss rapid updates (same second), clock skew, full table scan if not indexed

### Timestamp-Based CDC
- **Variant of polling**: Use `updated_at` or `modified_at`
- **Incremental**: Only fetch changed rows
- **Gaps**: Batch updates may have same timestamp

### Outbox Pattern
- **Table**: `outbox` (id, aggregate_type, aggregate_id, event_type, payload, created_at)
- **Application**: Writes business data + outbox row in same transaction
- **CDC**: Reads outbox, publishes to Kafka, optionally deletes
- **Benefit**: Domain events, not raw DB changes; application controls event shape

### Debezium
- **Open-source**: Apache 2.0
- **Connectors**: MySQL, PostgreSQL, MongoDB, SQL Server, Oracle, Db2
- **Kafka Connect**: Runs as Kafka Connect source connector
- **Format**: Envelope (before, after, source, op, ts_ms)
- **Offset**: Stored in Kafka Connect offset topic
- **Snapshot**: Initial full snapshot, then incremental from WAL

---

## 5. Numbers

| Metric | Value |
|--------|-------|
| **Latency** | Sub-second (log-based) to seconds (polling) |
| **Throughput** | 10K-100K+ events/sec (Debezium + Kafka) |
| **MySQL binlog** | Row-based required for CDC |
| **PostgreSQL** | WAL level = logical, replication slot |
| **Debezium** | Single connector per DB (or per table with filters) |

---

## 6. Tradeoffs

### Log-Based vs Trigger vs Polling

| Approach | Latency | Overhead | Complexity | DB Support |
|----------|---------|----------|------------|------------|
| Log-based | Low | Minimal | Medium | MySQL, PG, etc. |
| Trigger | Low | High (every write) | Low | Any |
| Polling | High | Query load | Low | Any |

### Outbox vs Direct CDC
- **Outbox**: Application-defined events; extra table; transactional
- **Direct CDC**: Raw DB changes; no app changes; schema-coupled

---

## 7. Variants / Implementations

### Debezium
- Kafka Connect connectors
- Snapshot + streaming
- Schema evolution (Schema Registry)
- Transformations (SMT)

### AWS DMS (Database Migration Service)
- Managed CDC
- Homogeneous and heterogeneous migration
- Kinesis, S3, Redshift targets

### Fivetran, Airbyte
- Managed ETL/ELT
- CDC for data warehouse sync
- SaaS connectors

### Maxwell (MySQL)
- Lightweight MySQL CDC
- Kafka, Kinesis, S3, stdout
- No Kafka Connect (standalone)

### Canal (Alibaba)
- MySQL binlog parser
- Kafka, RocketMQ
- Used in Alibaba ecosystem

---

## 8. Scaling Strategies

1. **Partitioning**: Kafka topic partitioned by table or key
2. **Parallel connectors**: One connector per DB (or shard)
3. **Filtering**: Include/exclude tables to reduce volume
4. **Snapshot tuning**: Parallel snapshot for large tables
5. **Kafka partitioning**: Partition by entity ID for ordering

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|------------|
| WAL retention exceeded | Missed changes | Increase WAL retention, snapshot restart |
| Connector crash | Resume from offset | Kafka Connect stores offset |
| Schema change | Break parsing | Schema Registry, backward compatibility |
| High volume | Connector lag | Scale Kafka partitions, tune batch size |
| DB failover | Replication slot | Handle slot on new primary |

---

## 10. Performance Considerations

- **WAL retention**: Ensure enough for connector catch-up
- **Snapshot**: Can be slow for large tables; use parallel snapshot
- **Batch size**: Tune for throughput vs latency
- **Schema Registry**: Cache schemas to reduce lookups
- **Filtering**: Exclude high-churn tables if not needed

---

## 11. Use Cases

- Search indexing (DB → Elasticsearch)
- Cache invalidation
- Data warehouse ETL
- Event sourcing (outbox)
- Microservices data sync
- Audit logging
- Real-time analytics

---

## 12. Comparison Tables

### CDC Approaches

| Approach | Latency | Overhead | Use Case |
|----------|---------|----------|----------|
| Log-based | < 1 sec | Low | Production, high volume |
| Trigger | < 1 sec | High | Legacy DBs without WAL |
| Polling | Secs-min | Medium | Simple, low volume |
| Outbox | < 1 sec | Medium | Domain events |

### Debezium vs DMS vs Fivetran

| Feature | Debezium | AWS DMS | Fivetran |
|---------|----------|----------|----------|
| **Managed** | Self | AWS | SaaS |
| **Targets** | Kafka | Kinesis, S3, etc. | Warehouses |
| **Cost** | Free | Per hour | Per connector |
| **Flexibility** | High | Medium | Lower |

---

## 13. Code or Pseudocode

### Debezium MySQL Connector Config

```json
{
  "name": "mysql-connector",
  "config": {
    "connector.class": "io.debezium.connector.mysql.MySqlConnector",
    "database.hostname": "mysql",
    "database.port": "3306",
    "database.user": "debezium",
    "database.password": "secret",
    "database.server.id": "184054",
    "topic.prefix": "dbserver1",
    "database.include.list": "inventory",
    "table.include.list": "inventory.orders,inventory.customers",
    "key.converter": "org.apache.kafka.connect.json.JsonConverter",
    "value.converter": "io.confluent.connect.avro.AvroConverter",
    "value.converter.schema.registry.url": "http://schema-registry:8081"
  }
}
```

### Outbox Table Schema

```sql
CREATE TABLE outbox (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  aggregate_type VARCHAR(255) NOT NULL,
  aggregate_id VARCHAR(255) NOT NULL,
  event_type VARCHAR(255) NOT NULL,
  payload JSON NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_created (created_at)
);
```

### Application Outbox Write

```python
def create_order(order_data):
    with db.transaction():
        order = Order.create(**order_data)
        Outbox.create(
            aggregate_type='Order',
            aggregate_id=str(order.id),
            event_type='OrderCreated',
            payload=json.dumps({'order_id': order.id, 'amount': order.amount})
        )
    # CDC will pick up outbox row and publish to Kafka
```

---

## 14. Interview Discussion

### Key Points to Cover

1. **Why CDC**: Avoid dual-write, single source of truth, consistency
2. **Log-based**: WAL/binlog, non-invasive, low latency
3. **Outbox pattern**: Transactional events, application-defined
4. **Debezium**: Connectors, snapshot, streaming, offset
5. **Failure handling**: WAL retention, offset, schema evolution

### Sample Questions

**Q: How does CDC solve the dual-write problem?**
A: Application writes only to DB. CDC reads WAL and publishes changes. No second write path; no consistency bugs between DB and event bus.

**Q: What is the outbox pattern and when to use it?**
A: Application writes business data + event to outbox table in same transaction. CDC reads outbox and publishes. Use when you need domain events (not raw DB changes) or when DB doesn't support WAL CDC.

**Q: How do you handle schema changes in a CDC pipeline?**
A: Schema Registry with backward compatibility. Debezium emits schema with each record. Consumers use schema evolution. For breaking changes: new topic, versioned consumers, or transformation.

---

## Appendix: Additional Deep Dives

### MySQL Binlog Configuration for CDC

```
# Required for Debezium/CDC
binlog_format = ROW          # Row-based, not statement-based
binlog_row_image = FULL      # Before + after values
server_id = 1                # Unique per server
log_bin = mysql-bin
expire_logs_days = 7         # Retain for connector catch-up
```

### PostgreSQL WAL Configuration

```
wal_level = logical           # Required for logical replication
max_replication_slots = 4     # One per connector
max_wal_senders = 4
```

### Debezium Snapshot Modes

- **Initial**: Full snapshot on first start, then streaming. Default.
- **Never**: No snapshot; streaming only. For existing data elsewhere.
- **Initial_only**: Snapshot only, no streaming. For one-time migration.
- **Schema_only**: Schema only, no data. For schema sync.

### Outbox Table Cleanup Strategies

1. **CDC deletes**: Connector deletes after publish (requires custom SMT or connector)
2. **Separate job**: Cron job deletes rows older than N hours
3. **Partitioning**: Partition by date; drop old partitions
4. **No delete**: Rely on retention; archive to cold storage

### CDC and Schema Evolution

- **New column**: Add nullable column — CDC captures; consumers ignore if not in schema
- **Remove column**: Deprecate first; remove after consumers updated
- **Rename**: Add new column, backfill, deprecate old
- **Type change**: Breaking; new topic or transformation

### Debezium Event Envelope Structure

```json
{
  "before": {"id": 1, "name": "old"},
  "after": {"id": 1, "name": "new"},
  "source": {"version": "1.9", "connector": "mysql", "ts_ms": 1234567890},
  "op": "u",
  "ts_ms": 1234567890123
}
```
`op`: c=create, u=update, d=delete, r=read (snapshot)

### Debezium Connector Offset Storage

- **Kafka Connect**: Offset stored in `connect-offsets` topic
- **Format**: `{connector_name, partition}: {offset, source info}`
- **Resume**: Connector resumes from last committed offset
- **Reset**: Delete offset to trigger full snapshot

### CDC and Data Quality

- **Duplicate events**: Possible on connector restart; consumer must deduplicate
- **Ordering**: Per partition (table/primary key); maintain order for same entity
- **Late-arriving**: Out-of-order possible in distributed systems; use watermarking for stream processing
- **Tombstones**: Debezium emits null value for deletes (Kafka log compaction)

### Trigger-Based CDC Implementation

```sql
CREATE TABLE changelog (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  table_name VARCHAR(255),
  op ENUM('I','U','D'),
  pk_value VARCHAR(255),
  old_json JSON,
  new_json JSON,
  ts TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER orders_after_insert
AFTER INSERT ON orders
FOR EACH ROW
INSERT INTO changelog (table_name, op, pk_value, new_json)
VALUES ('orders', 'I', NEW.id, JSON_OBJECT('id', NEW.id, 'amount', NEW.amount));
```

### Polling-Based CDC with Watermarking

```sql
-- Table needs updated_at column, indexed
SELECT * FROM orders WHERE updated_at > :last_watermark ORDER BY updated_at LIMIT 1000;
-- Update last_watermark to MAX(updated_at) from result
-- Gap: two updates in same second may have same updated_at
```

### CDC for Multi-Region

- **Active-active**: CDC from each region to central Kafka; conflict resolution needed
- **Active-passive**: Primary region CDC → Kafka; failover to secondary
- **Replication**: Kafka MirrorMaker 2 for cross-datacenter

### Debezium Transformations (SMT) for CDC

- **ExtractNewRecordState**: Flatten envelope to just `after` (or `before` for deletes)
- **Filter**: Drop records (e.g., by table, column value)
- **ContentBasedRouter**: Route to different topics by content
- **Mask**: Redact PII (e.g., mask email, credit card)
- **AddHeader**: Add metadata for downstream routing

### CDC vs Event Sourcing

| CDC | Event Sourcing |
|-----|----------------|
| Observes DB changes | Application emits events |
| Raw DB operations (I/U/D) | Domain events (OrderCreated) |
| Schema-coupled | Schema-agnostic |
| Single source = DB | Single source = event log |
| Use case: Sync, indexing | Use case: Audit, replay, CQRS |

### Monitoring CDC Pipelines

- **Connector lag**: Offset behind latest DB position
- **Snapshot progress**: For initial snapshot, % complete
- **Error rate**: Failed records, DLQ depth
- **Throughput**: Events/sec per connector
- **Schema compatibility**: Failed schema registrations
