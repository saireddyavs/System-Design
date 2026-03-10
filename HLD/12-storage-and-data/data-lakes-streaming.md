# Data Lakes & Streaming Architectures

## 1. Concept Overview

**Data Lakes** store raw data in its native format (structured, semi-structured, unstructured) at scale. **Schema-on-read** means structure is applied when querying, not at ingest. Data lakes enable flexibility but require governance.

**Streaming architectures** process data in motion. **Lambda architecture** combines batch (correctness) and stream (latency) layers. **Kappa architecture** uses a single stream-processing pipeline for both.

**Data Lakehouse** (Delta Lake, Apache Iceberg, Apache Hudi) combines lake flexibility with warehouse structure: ACID, schema evolution, time travel.

---

## 2. Real-World Motivation

- **Netflix**: Data lake on S3; 500B+ events/day. Keystone pipeline (Kafka) for streaming. Data mesh for domain ownership.
- **Uber**: Hudi for incremental processing; Kafka for real-time. Petabyte-scale analytics.
- **LinkedIn**: Samza/Kafka for stream processing. 7 trillion events/day.
- **Databricks**: Delta Lake as lakehouse. ACID, time travel, merge.
- **Snowflake**: Data lake + warehouse integration. Query S3 directly.

---

## 3. Architecture Diagrams

### 3.1 Data Lake Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         DATA SOURCES                                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐          │
│  │   Apps   │  │   DBs    │  │  Kafka   │  │   APIs   │  │   IoT    │          │
│  │  (JSON)  │  │  (SQL)   │  │ (events) │  │  (REST)  │  │ (binary) │          │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘          │
└───────┼─────────────┼─────────────┼─────────────┼─────────────┼─────────────────┘
        │             │             │             │             │
        │    Batch    │   CDC       │   Stream    │   Batch     │
        ▼             ▼             ▼             ▼             ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    INGESTION LAYER                                                │
│  Fluentd | Kafka Connect | Airflow | Glue | Custom                               │
└─────────────────────────────────────────────────────────────────────────────────┘
        │
        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    DATA LAKE (S3 / GCS / ADLS)                                    │
│  ┌─────────────────────────────────────────────────────────────────────────────┐ │
│  │  Raw zone:     /raw/events/2024/03/10/                                        │ │
│  │  Processed:    /processed/analytics/                                           │ │
│  │  Curated:      /curated/datamarts/                                             │ │
│  │  Formats:      Parquet, JSON, CSV, Avro                                       │ │
│  │  Schema:       Applied at read (schema-on-read)                               │ │
│  └─────────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────┘
        │
        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    QUERY / ANALYTICS                                              │
│  Athena | Presto | Spark | Snowflake (external tables) | Databricks              │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 3.2 Lambda Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         DATA SOURCES                                              │
│  Events stream (Kafka) ───────────────────────────────────────────────────────►  │
└─────────────────────────────────────────────────────────────────────────────────┘
        │                                     │
        │                                     │
        ▼                                     ▼
┌───────────────────────┐           ┌───────────────────────┐
│   BATCH LAYER         │           │   SPEED LAYER         │
│   (Correctness)       │           │   (Latency)            │
│                       │           │                       │
│  - Raw data (immutable)│          │  - Stream processing   │
│  - Batch jobs (Spark)  │           │  - Kafka Streams      │
│  - Master dataset      │           │  - Flink              │
│  - Recompute all      │           │  - Real-time views    │
│  - Latency: hours     │           │  - Latency: seconds   │
└───────────┬───────────┘           └───────────┬───────────┘
            │                                   │
            │         ┌─────────────────┐       │
            └────────►│  SERVING LAYER  │◄──────┘
                      │                 │
                      │  Query = batch  │
                      │  view + real-   │
                      │  time view      │
                      └────────┬────────┘
                               │
                               ▼
                      ┌─────────────────┐
                      │  Application    │
                      │  (merged view)  │
                      └─────────────────┘
```

### 3.3 Kappa Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         DATA SOURCES                                              │
│  Events stream (Kafka) ───────────────────────────────────────────────────────►  │
│  (Single source of truth; replayable)                                            │
└───────────────────────────────────────────────────────────────────────────────┬─┘
                                                                                 │
                                                                                 ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    STREAM PROCESSING (single pipeline)                           │
│  ┌─────────────────────────────────────────────────────────────────────────────┐ │
│  │  Kafka Streams / Flink / Samza                                                │ │
│  │  - Real-time: process as events arrive                                        │ │
│  │  - Batch: same code, replay from offset 0 (or compacted topic)                 │ │
│  │  - Exactly-once semantics                                                     │ │
│  └─────────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────┘
        │
        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    SERVING LAYER                                                  │
│  - Kafka (materialized views, KTables)                                           │
│  - Database (for query)                                                          │
│  - Cache                                                                         │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 3.4 Data Lakehouse (Delta Lake / Iceberg)

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    DATA LAKE (S3) + LAKEHOUSE FORMAT                              │
│                                                                                   │
│  /data/events/                                                                    │
│    ├── _delta_log/     (Delta Lake)                                               │
│    │   └── 00000.json, 00001.json, ...  (transaction log)                         │
│    ├── part-00000.parquet                                                         │
│    ├── part-00001.parquet                                                         │
│    └── ...                                                                        │
│                                                                                   │
│  Features:                                                                        │
│  - ACID transactions (optimistic concurrency)                                     │
│  - Time travel (query as of version N)                                             │
│  - Schema evolution                                                               │
│  - Merge (upsert)                                                                 │
│  - Vacuum (retention)                                                             │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Core Mechanics

### 4.1 Data Lake

- **Schema-on-read**: Data stored as-is; schema applied at query time
- **Zones**: Raw → Processed → Curated (medallion architecture)
- **Formats**: Parquet (columnar), JSON, Avro
- **Governance**: Catalog (Hive, Glue), access control, lineage

### 4.2 Lambda Architecture

- **Batch layer**: Immutable master dataset; batch jobs recompute views
- **Speed layer**: Stream processing; real-time views (incremental)
- **Serving layer**: Merge batch + speed for query
- **Trade-off**: Two codebases (batch + stream) to maintain

### 4.3 Kappa Architecture

- **Single pipeline**: Stream processing for both real-time and batch (replay)
- **Requires**: Kafka retention long enough to replay; or compacted topics
- **Simpler**: One codebase; same logic for real-time and historical

### 4.4 Exactly-Once Processing

- **Idempotent writes**: Same key → same result
- **Transactional outbox**: Write to DB + outbox in same transaction
- **Kafka transactions**: Producer sends batch; consumer commits offset atomically
- **Flink**: Changelog + checkpoint; recover from checkpoint

---

## 5. Numbers

| Metric | Value |
|--------|-------|
| Kafka retention | 7–365 days (configurable) |
| Kafka throughput | Millions msg/sec per cluster |
| Flink checkpoint | Seconds to minutes |
| Lambda batch latency | Hours |
| Lambda speed latency | Seconds |
| Data lake scale | PB+ |
| Delta Lake compaction | Periodic (weekly) |

---

## 6. Tradeoffs (Comparison Tables)

### Data Lake vs Data Warehouse

| Aspect | Data Lake | Data Warehouse |
|--------|-----------|----------------|
| **Schema** | Schema-on-read | Schema-on-write |
| **Data** | Raw, any format | Curated, structured |
| **Users** | Data scientists, engineers | Analysts, BI |
| **Cost** | Low (object storage) | Higher (optimized) |
| **Query** | Flexible, ad-hoc | Optimized for SQL |

### Lambda vs Kappa

| Aspect | Lambda | Kappa |
|--------|--------|-------|
| **Pipelines** | Batch + Stream | Stream only |
| **Code** | Two codebases | One |
| **Complexity** | Higher | Lower |
| **Replay** | Batch recomputes | Stream replays |
| **Use case** | When batch logic differs | When same logic suffices |

### Delta Lake vs Iceberg vs Hudi

| Aspect | Delta Lake | Iceberg | Hudi |
|--------|------------|---------|------|
| **Ecosystem** | Spark-native | Spark, Flink, Presto | Spark, Flink |
| **Format** | Parquet + JSON log | Parquet + metadata | Parquet + metadata |
| **ACID** | Yes | Yes | Yes |
| **Time travel** | Yes | Yes | Yes |
| **Merge** | Yes | Yes | Yes |

---

## 7. Variants/Implementations

### Data Lakes

- **AWS**: S3 + Glue + Athena
- **GCP**: GCS + BigQuery (external tables)
- **Azure**: ADLS + Synapse
- **On-prem**: HDFS + Hive

### Stream Processing

- **Kafka Streams**: Lightweight, Kafka-native
- **Apache Flink**: True stream processing, exactly-once
- **Spark Streaming**: Micro-batch (DStream)
- **Samza**: LinkedIn, Kafka-based

### Lakehouse

- **Delta Lake**: Databricks
- **Apache Iceberg**: Netflix, Apple
- **Apache Hudi**: Uber

### Medallion Architecture (Data Lake Zones)

A common pattern for organizing data lake layers:

- **Bronze (Raw)**: Immutable, schema-on-read. Raw events, logs, CDC streams. Minimal transformation.
- **Silver (Cleaned)**: Deduplicated, validated, conformed to schema. Still detailed but queryable.
- **Gold (Curated)**: Aggregated, business-level. Ready for BI, ML features.

Data flows: Bronze → Silver → Gold. Each layer can be queried; Gold is optimized for end users.

### Netflix Data Pipeline (Real Example)

Netflix's Keystone pipeline:
- **Ingest**: 500B+ events/day from apps, devices, backend services
- **Transport**: Kafka (regional clusters)
- **Processing**: Flink, Spark for real-time and batch
- **Storage**: S3 data lake; Iceberg for tables
- **Governance**: Data mesh; domain ownership; schema registry

### Uber's Data Lake with Hudi

Uber uses Apache Hudi for:
- **Upserts**: Merge CDC from MySQL into S3. No full table rewrites.
- **Incremental processing**: Only process changed records. 10x faster than full scan.
- **Time travel**: Query data as of a point in time for reproducibility.

### LinkedIn Stream Processing (Samza)

LinkedIn processes 7 trillion events/day with Samza:
- **Kafka** as log; **Samza** as stream processor
- **Exactly-once** via Kafka transactions + idempotent sinks
- **State**: RocksDB local store; changelog topic for durability

### Hadoop MapReduce (Batch Processing)

The original batch paradigm:
- **Map**: Process each record; emit (key, value) pairs
- **Shuffle**: Group by key; send to reducers
- **Reduce**: Aggregate per key

Limitation: Disk I/O between stages. Spark improves with in-memory RDDs and DAG optimization. MapReduce is legacy but conceptually important.

---

## 8. Scaling Strategies

- **Data lake**: Add storage (S3 scales); partition by date/region
- **Kafka**: Add brokers; partition topics
- **Flink**: Scale parallelism; checkpoint to distributed storage
- **Lambda batch**: Scale Spark cluster; optimize shuffle

---

## 9. Failure Scenarios

| Failure | Impact | Mitigation |
|---------|--------|------------|
| Kafka broker down | Stream stalled | Replication, ISR |
| Flink job fail | State lost | Checkpoint, restart |
| Lambda batch fail | Stale serving | Retry, idempotency |
| S3 outage | Lake unavailable | Multi-region |
| Schema drift | Query break | Schema evolution, validation |

---

## 10. Performance Considerations

- **Partitioning**: Date, region; avoid too many small files
- **Compaction**: Delta/Iceberg compact small files
- **Kafka**: Tune batch size, linger; partition key for ordering
- **Flink**: Checkpoint interval vs overhead

---

## 11. Use Cases

| Pattern | Use Case |
|---------|----------|
| **Data lake** | Raw data storage, ML, exploration |
| **Lambda** | When batch and stream logic differ (e.g., ML vs real-time) |
| **Kappa** | Event-driven; same logic for real-time and batch |
| **Lakehouse** | ACID, upsert, time travel on lake |

---

## 12. Comparison Tables

### Batch vs Stream Processing

| Aspect | Batch | Stream |
|--------|-------|--------|
| **Data** | Bounded | Unbounded |
| **Latency** | Minutes–hours | Seconds |
| **Tools** | Spark, MapReduce | Flink, Kafka Streams |
| **Fault tolerance** | Re-run job | Checkpoint, replay |

---

## 13. Code/Pseudocode

### Delta Lake Write (Spark)

```python
df.write \
  .format("delta") \
  .mode("append") \
  .option("mergeSchema", "true") \
  .save("/data/events")

# Time travel
spark.read.format("delta").option("versionAsOf", 5).load("/data/events")

# Merge (upsert)
deltaTable.alias("t").merge(
  updates.alias("u"), "t.id = u.id"
).whenMatchedUpdateAll().whenNotMatchedInsertAll().execute()
```

### Kafka Streams (Java)

```java
StreamsBuilder builder = new StreamsBuilder();
KStream<String, Order> orders = builder.stream("orders");

orders
  .groupByKey()
  .windowedBy(TimeWindows.of(Duration.ofMinutes(5)))
  .aggregate(() -> 0.0, (key, order, total) -> total + order.getAmount(),
             Materialized.with(Serdes.String(), Serdes.Double()))
  .toStream()
  .to("order-totals-by-5min");

KafkaStreams streams = new KafkaStreams(builder.build(), config);
streams.start();
```

### Flink Exactly-Once (Scala)

```scala
val env = StreamExecutionEnvironment.getExecutionEnvironment
env.enableCheckpointing(60000) // 1 min
env.getCheckpointConfig.setCheckpointingMode(CheckpointingMode.EXACTLY_ONCE)

val stream = env.addSource(new FlinkKafkaConsumer(...))
stream
  .keyBy(_.userId)
  .window(TumblingEventTimeWindows.of(Time.minutes(5)))
  .reduce((a, b) => a.copy(amount = a.amount + b.amount))
  .addSink(new FlinkKafkaProducer(...))
```

### Lambda Merge (Pseudocode)

```python
# Serving layer: merge batch + speed
def get_user_revenue(user_id):
    batch_revenue = batch_store.get(user_id)  # From batch layer
    realtime_delta = speed_store.get(user_id)  # From speed layer
    return batch_revenue + realtime_delta
```

---

## 14. Interview Discussion

### Key Points

1. **Data lake = flexibility**: Schema-on-read, any format. Warehouse = structure, performance.
2. **Lambda = two pipelines**: Batch (correct) + stream (fast). Merge at serve.
3. **Kappa = one pipeline**: Stream only; replay for batch. Simpler if logic is same.
4. **Lakehouse**: ACID, upsert, time travel on object storage. Delta, Iceberg, Hudi.
5. **Exactly-once**: Idempotency, transactions, checkpointing.

### Common Questions

**Q: When would you choose Lambda over Kappa?**  
A: When batch logic is fundamentally different (e.g., complex ML feature computation in batch vs simple aggregations in stream). Or when Kafka retention is insufficient for full replay.

**Q: What is schema-on-read?**  
A: Data stored without enforcing schema. When querying, you apply a schema (e.g., "this column is a timestamp"). Flexible but requires discipline.

**Q: How does Delta Lake achieve ACID?**  
A: Transaction log (_delta_log) records all changes. Optimistic concurrency: multiple writers; conflict resolution at commit.

**Q: How does Flink achieve exactly-once?**  
A: Checkpoints (distributed snapshot of state); on failure, restore from checkpoint and replay from last checkpoint. Two-phase commit for sinks.

### Kafka Retention and Compaction

- **Time-based retention**: Delete messages older than 7 days. Default for event streams.
- **Size-based retention**: Delete when log exceeds N GB.
- **Compaction**: For keyed topics, keep only latest value per key. Enables "replay from beginning" for Kappa without infinite storage. Used for changelog topics, materialized views.

### Stream Processing Semantics

| Semantics | Guarantee | Implementation |
|-----------|-----------|----------------|
| **At-most-once** | May lose, no duplicate | Fire-and-forget |
| **At-least-once** | No lose, may duplicate | Retry; idempotent sink |
| **Exactly-once** | No lose, no duplicate | Transactions, checkpoint |

Exactly-once requires: idempotent sink or transactional sink + consumer offset in same transaction.

### Spark Streaming: Micro-Batch vs Structured Streaming

- **DStream (legacy)**: Micro-batches (e.g., every 1 second). Each batch is a Spark job. Not true streaming; higher latency.
- **Structured Streaming**: Same DataFrame API as batch. Continuous processing mode (Flink-like) or micro-batch. Exactly-once with checkpointing.

### Data Lake Governance

- **Catalog**: Hive Metastore, AWS Glue, Unity Catalog. Tables, schemas, partitions.
- **Lineage**: Track data flow (source → transform → sink). Critical for compliance.
- **Access control**: Row-level, column-level. Integrate with IAM/LDAP.
- **Quality**: Validation rules, monitoring. Great Expectations, dbt tests.

### Kafka Consumer Groups

- **Consumer group**: Set of consumers sharing workload. Each partition assigned to one consumer in group.
- **Rebalance**: When consumer joins/leaves, partitions reassigned. Pauses consumption briefly.
- **Offset**: Each consumer tracks its position per partition. Stored in `__consumer_offsets` topic (or external store).

### Flink Checkpointing

1. **Barrier**: Special record injected in stream. Divides "before" and "after" checkpoint.
2. **Align**: Operators wait for barriers from all inputs before snapshotting state.
3. **Snapshot**: State backend (RocksDB, heap) writes to distributed storage (S3, HDFS).
4. **Recovery**: On failure, restore state from last checkpoint; replay source from that point.

### Event Sourcing and CQRS in Streaming

- **Event sourcing**: Store events (state changes), not current state. Replay to reconstruct. Fits stream processing.
- **CQRS**: Separate write model (commands) from read model (queries). Stream builds read model from events.
- **Kafka**: Event log. Consumers build materialized views. Exactly-once enables correct read models.

### Watermarks in Stream Processing

For event-time processing (vs processing-time):
- **Watermark**: "No more events with timestamp < T will arrive." Enables closing windows.
- **Late data**: Events after watermark. Side output or drop.
- **Allowed lateness**: Flink allows late data within bound; updates window result.

### Choosing Lambda vs Kappa: Decision Framework

| Factor | Prefer Lambda | Prefer Kappa |
|--------|---------------|--------------|
| Batch logic differs from stream | Yes | No |
| Kafka retention sufficient | N/A | Yes |
| Team expertise | Batch + Stream | Stream only |
| Operational complexity tolerance | Higher | Lower |

### Delta Lake Transaction Log

The `_delta_log` directory contains JSON files (00000.json, 00001.json, ...). Each file = one transaction. Contents: AddFile (new parquet), RemoveFile (delete), Metadata (schema). Readers read log to get current snapshot. Writers use optimistic concurrency: commit if no conflicting writes.

### Iceberg Table Format

Apache Iceberg (from Netflix) provides: (1) Hidden partitioning (partition by date without storing date column; partition pruning automatic), (2) Schema evolution (add column, rename), (3) Snapshot isolation, (4) Time travel. Works with Spark, Flink, Presto. Metadata in manifest files; no central metastore for partition info.

### Hudi Write Types

- **Copy on Write (CoW)**: Update = rewrite parquet file. Simple, good for read-heavy.
- **Merge on Read (MoR)**: Updates go to log file; merge on read or compaction. Better for write-heavy, higher read latency.

### Summary: Data Lake vs Warehouse vs Lakehouse

| | Data Lake | Warehouse | Lakehouse |
|---|-----------|-----------|-----------|
| **Schema** | On read | On write | On read + evolution |
| **Format** | Any | Optimized | Parquet + log |
| **ACID** | No | Yes | Yes |
| **Use case** | Raw, exploration | BI, reporting | Unified |

The lakehouse trend (Delta, Iceberg, Hudi) addresses the gap: ACID and structure on cheap object storage, without the rigidity of traditional warehouses. Choose based on ecosystem (Spark → Delta), multi-engine (Iceberg), or write-heavy upserts (Hudi).

### Final Checklist for Interview

- Explain Lambda vs Kappa and when to choose each.
- Describe exactly-once semantics and how Flink achieves it.
- Know medallion (bronze/silver/gold) architecture.
- Articulate Delta Lake or Iceberg transaction model.

---
