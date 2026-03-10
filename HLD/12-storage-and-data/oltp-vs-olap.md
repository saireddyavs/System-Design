# OLTP vs OLAP: Transactional vs Analytical Processing

## 1. Concept Overview

**OLTP** (Online Transaction Processing) and **OLAP** (Online Analytical Processing) represent two fundamentally different database workloads:

- **OLTP**: Optimized for **many small, fast transactions** — inserts, updates, point lookups. Powers real-time applications (e-commerce, banking, reservations). **Row-oriented** storage, **normalized** schema.
- **OLAP**: Optimized for **complex analytical queries** over large datasets — aggregations, joins, scans. Powers reporting, dashboards, BI. **Column-oriented** storage, **denormalized** schema (star/snowflake).

The **latency** and **throughput** requirements differ dramatically: OLTP targets <10ms per transaction; OLAP queries run seconds to minutes over petabytes.

---

## 2. Real-World Motivation

- **Stripe**: OLTP for millions of payment transactions per second. PostgreSQL for strong consistency.
- **Snowflake**: OLAP for analytics. Columnar storage, separation of compute and storage, scales to petabytes.
- **Uber**: OLTP (MySQL) for ride booking; OLAP (Hive, Presto) for analytics on billions of trips.
- **Netflix**: OLTP for playback state, recommendations; OLAP for content analytics, viewer behavior.
- **Amazon**: DynamoDB for cart/checkout (OLTP); Redshift for business intelligence (OLAP).

---

## 3. Architecture Diagrams

### 3.1 OLTP vs OLAP High-Level

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         OLTP (Online Transaction Processing)                      │
│                                                                                   │
│  Users/Apps ──► [API] ──► [OLTP DB] ◄── Many small transactions                   │
│                              │         - INSERT order                                                            │
│                              │         - UPDATE user balance                                                     │
│                              │         - SELECT * WHERE id = 123                                                 │
│                              │         Latency: < 10ms                                                            │
│                              │         Throughput: 10K–100K TPS                                                   │
│                              │         Storage: Row-oriented                                                     │
│                              │         Schema: Normalized (3NF)                                                   │
│                              ▼                                                                                   │
│  PostgreSQL | MySQL | Oracle | DynamoDB | Cassandra                                                             │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────┐
│                         OLAP (Online Analytical Processing)                       │
│                                                                                   │
│  Analysts/BI ──► [Query] ──► [OLAP / Data Warehouse] ◄── Few complex queries       │
│                              │         - SELECT region, SUM(revenue) GROUP BY region                              │
│                              │         - JOIN 10 tables, aggregate over 1B rows                                   │
│                              │         Latency: seconds to minutes                                                │
│                              │         Throughput: 10–100 concurrent queries                                     │
│                              │         Storage: Column-oriented                                                  │
│                              │         Schema: Denormalized (star/snowflake)                                      │
│                              ▼                                                                                   │
│  Snowflake | BigQuery | Redshift | ClickHouse | Apache Druid                                                      │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 3.2 Row vs Column Storage

```
ROW-ORIENTED (OLTP)                    COLUMN-ORIENTED (OLAP)
─────────────────────                  ─────────────────────

| id | name  | age | city |           | id | id | id | id |
|----|-------|-----|------|           | 1  | 2  | 3  | 4  |
| 1  | Alice | 30  | NYC  |           | name | name | name | name |
| 2  | Bob   | 25  | LA   |           | Alice| Bob | Carol| Dave|
| 3  | Carol | 35  | SF   |           | age | age | age | age |
| 4  | Dave  | 28  | NYC  |           | 30 | 25 | 35 | 28 |
                                    | city | city | city | city |
                                    | NYC | LA | SF | NYC |

Row stored together:                    Column stored together:
- Good for: fetch one row               - Good for: SUM(age), AVG(age)
- Bad for: SUM(age) (scan all rows)    - Good for: compression (same type)
- Bad for: fetch one row (many columns)
```

### 3.3 ETL Pipeline Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         SOURCE SYSTEMS (OLTP)                                    │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐                          │
│  │PostgreSQL│  │  MySQL   │  │  MongoDB │  │   Kafka  │                          │
│  │ (orders) │  │ (users)  │  │ (events) │  │ (logs)   │                          │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘                          │
└───────┼─────────────┼─────────────┼─────────────┼──────────────────────────────────────────────────┘
        │             │             │             │
        │    CDC / Batch / Stream  │             │
        ▼             ▼             ▼             ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         ETL LAYER (Extract, Transform, Load)                       │
│  ┌─────────────────────────────────────────────────────────────────────────────┐ │
│  │  Extract: Pull data from sources (full dump, incremental, CDC)             │ │
│  │  Transform: Clean, join, aggregate, denormalize                               │ │
│  │  Load: Write to data warehouse                                               │ │
│  │  Tools: Airflow, dbt, Fivetran, Spark                                         │ │
│  └─────────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────┘
        │
        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         DATA WAREHOUSE (OLAP)                                     │
│  ┌─────────────────────────────────────────────────────────────────────────────┐ │
│  │  Star Schema:                                                                │
│  │    fact_sales (date_id, product_id, customer_id, revenue, quantity)           │
│  │    dim_date, dim_product, dim_customer                                        │
│  │  Columnar: Parquet, ORC, native (Snowflake, BigQuery)                         │
│  └─────────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────┘
        │
        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         BI / REPORTING                                            │
│  Tableau | Looker | Metabase | Superset | Custom dashboards                       │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 3.4 Star Schema vs Snowflake Schema

```
STAR SCHEMA                              SNOWFLAKE SCHEMA
──────────────                           ─────────────────

        fact_sales                               fact_sales
    ┌───────────────┐                       ┌───────────────┐
    │ date_id       │                       │ date_id       │
    │ product_id    │                       │ product_id    │
    │ customer_id   │                       │ customer_id   │
    │ revenue       │                       │ revenue       │
    │ quantity      │                       │ quantity      │
    └───────┬───────┘                       └───────┬───────┘
        │   │   │                                   │   │   │
        ▼   ▼   ▼                                   ▼   ▼   ▼
   dim_date  dim_product  dim_customer         dim_date  dim_product  dim_customer
   ┌─────┐   ┌─────────┐  ┌──────────┐        ┌─────┐   ┌─────────┐  ┌──────────┐
   │id   │   │id       │  │id        │        │id   │   │id       │  │id        │
   │date │   │name     │  │name      │        │date │   │name     │  │name      │
   │...  │   │category │  │region_id │        │...  │   │category │  │region_id │
   └─────┘   └─────────┘  └────┬─────┘        └─────┘   └────┬────┘  └────┬─────┘
                               │                              │            │
                        Denormalized                    dim_category   dim_region
                        (category in dim_product)       ┌─────────┐  ┌─────────┐
                                                        │id       │  │id       │
                                                        │name     │  │name     │
                                                        └─────────┘  └─────────┘
                                                        More normalized
```

---

## 4. Core Mechanics

### 4.1 OLTP Characteristics

- **ACID**: Strong consistency, transactions
- **Indexes**: B-tree for point lookups, range scans
- **Locks**: Row-level or page-level locking
- **Writes**: Optimized for insert/update
- **Concurrency**: Many short transactions

### 4.2 OLAP Characteristics

- **Eventually consistent** or batch-consistent: Analytics can tolerate slight staleness
- **Indexes**: Minimal; often full scan
- **Compression**: Columnar enables high compression (same type)
- **Vectorized processing**: Process columns in batches (SIMD)
- **Concurrency**: Fewer, longer-running queries

### 4.3 Column-Oriented Storage

**Why columns for analytics?**

1. **Compression**: Same data type in one column → run-length encoding, dictionary encoding. 10x compression possible.
2. **I/O efficiency**: Query needs only `age`, `revenue` → read only those columns, not entire rows.
3. **Vectorized execution**: Process 1024 values in one CPU instruction (SIMD).
4. **Late materialization**: Join/aggregate in column form, materialize row only at end.

**Parquet** (columnar format): Row groups → column chunks → pages. Metadata at footer for predicate pushdown.

**ORC** (Optimized Row Columnar): Similar, Hive-native. Indexes (min, max, bloom) for skip.

### 4.4 ETL vs ELT

| | ETL | ELT |
|---|-----|-----|
| **Transform** | Before load (in ETL tool) | After load (in warehouse) |
| **Where** | Spark, Airflow, Python | SQL in Snowflake/BigQuery |
| **Use case** | Legacy, complex transforms | Modern cloud DW (scales compute) |

**ELT** is favored when the warehouse has massive compute (Snowflake, BigQuery) — load raw, transform in SQL.

---

## 5. Numbers

| Metric | OLTP | OLAP |
|--------|------|------|
| **Latency** | < 10 ms | Seconds to minutes |
| **Throughput** | 10K–100K TPS | 10–100 concurrent queries |
| **Data size** | GB–TB | TB–PB |
| **Query complexity** | Simple (point lookup) | Complex (aggregations, joins) |
| **Write pattern** | Many small writes | Batch loads |
| **Schema** | Normalized | Denormalized |

---

## 6. Tradeoffs (Comparison Tables)

### OLTP vs OLAP

| Aspect | OLTP | OLAP |
|--------|------|------|
| **Purpose** | Run business | Analyze business |
| **Users** | Applications, end users | Analysts, BI |
| **Data** | Current, transactional | Historical, aggregated |
| **Schema** | Normalized | Star/snowflake |
| **Storage** | Row-oriented | Column-oriented |
| **Indexes** | Heavy (B-tree) | Light |
| **Consistency** | Strong (ACID) | Eventual |

### Star vs Snowflake Schema

| Aspect | Star | Snowflake |
|--------|------|-----------|
| **Normalization** | Denormalized dims | Normalized dims |
| **Joins** | Fewer | More |
| **Storage** | More (redundancy) | Less |
| **Query** | Simpler | More joins |
| **Maintenance** | Easier | More complex |

---

## 7. Variants/Implementations

### OLTP Databases

- **PostgreSQL, MySQL**: Relational, ACID
- **Oracle, SQL Server**: Enterprise RDBMS
- **DynamoDB, Cassandra**: NoSQL, distributed
- **CockroachDB, TiDB**: Distributed SQL

### OLAP / Data Warehouses

- **Snowflake**: Cloud, separation of compute/storage
- **BigQuery**: Serverless, pay per query
- **Redshift**: AWS, columnar
- **ClickHouse**: Open-source, columnar, OLAP
- **Apache Druid**: Real-time analytics
- **StarRocks**: High-performance OLAP

### Columnar Formats

- **Parquet**: Hadoop ecosystem, Spark
- **ORC**: Hive
- **Arrow**: In-memory, interchange

### Parquet File Structure

A Parquet file is organized as:
- **Row groups**: Horizontal partitions (e.g., 128 MB each)
- **Column chunks**: Within each row group, one per column
- **Pages**: Within column chunks; min/max statistics for predicate pushdown
- **Footer**: Metadata (schema, row group locations, column stats)

When Spark reads `SELECT SUM(revenue) FROM sales`, it reads only the `revenue` column chunks, not entire rows. Combined with compression (dictionary + run-length for low-cardinality columns), this yields 10–50x less I/O than row storage.

### HTAP (Hybrid Transactional/Analytical Processing)

Some systems aim to serve both OLTP and OLAP:
- **TiDB**: TiKV (row store) for OLTP; TiFlash (columnar replica) for OLAP. Raft for replication.
- **Citus**: PostgreSQL extension; sharding for scale; can run analytical queries across shards.
- **SingleStore**: Row store + column store in one system.

Trade-off: Complexity vs. operational simplicity of separate systems.

### Data Modeling: Fact vs Dimension

**Fact table**: Measures (revenue, quantity) + foreign keys to dimensions. Often append-only, very large.
**Dimension table**: Descriptive attributes (product name, region). Smaller, slowly changing.

Example: `fact_orders(order_id, date_id, product_id, customer_id, amount, quantity)`. Joins to `dim_date`, `dim_product`, `dim_customer` for filters and groupings.

---

## 8. Scaling Strategies

### OLTP Scaling

- **Vertical**: Bigger instance
- **Read replicas**: Offload reads
- **Sharding**: Partition by key (user_id)
- **CQRS**: Separate read/write models

### OLAP Scaling

- **Compute**: Scale-out clusters (Snowflake warehouses)
- **Storage**: Object storage (S3, GCS) — cheap, scalable
- **Partitioning**: By date, region
- **Materialized views**: Pre-aggregate for common queries

---

## 9. Failure Scenarios

| Failure | Impact | Mitigation |
|---------|--------|------------|
| OLTP down | App unavailable | Replication, failover |
| OLAP down | Reports delayed | Less critical; async |
| ETL failure | Stale warehouse | Retry, idempotency |
| Schema change | ETL break | Versioned schemas |
| Data quality | Bad analytics | Validation, monitoring |

---

## 10. Performance Considerations

- **OLTP**: Index design, connection pooling, query optimization
- **OLAP**: Partition pruning, columnar scan, compression
- **Materialized views**: Trade off freshness for speed
- **Incremental ETL**: Only process changed data

---

## 11. Use Cases

| Type | Use Case |
|------|----------|
| **OLTP** | E-commerce checkout, banking transfers, inventory |
| **OLAP** | Revenue dashboards, user behavior analysis, forecasting |
| **Hybrid** | Some systems use both (e.g., Citus for HTAP) |

---

## 12. Comparison Tables

### When to Use Each

| Need | Choose |
|------|--------|
| Real-time transactions | OLTP |
| Historical analytics | OLAP |
| Both (HTAP) | Hybrid (Citrus, TiDB) or separate systems |

### Data Warehouse Comparison

| | Snowflake | BigQuery | Redshift |
|---|-----------|----------|----------|
| **Model** | Compute/storage separation | Serverless | Cluster |
| **Pricing** | Per compute-second | Per query | Per cluster |
| **Scale** | Auto | Auto | Manual resize |

---

## 13. Code/Pseudocode

### OLTP Query (PostgreSQL)

```sql
-- Point lookup, update
BEGIN;
SELECT * FROM orders WHERE id = 12345 FOR UPDATE;
UPDATE orders SET status = 'shipped' WHERE id = 12345;
INSERT INTO order_events (order_id, event) VALUES (12345, 'shipped');
COMMIT;
```

### OLAP Query (Snowflake)

```sql
-- Aggregation over millions of rows
SELECT
  d.region,
  d.month,
  SUM(f.revenue) AS total_revenue,
  COUNT(DISTINCT f.customer_id) AS unique_customers
FROM fact_sales f
JOIN dim_date d ON f.date_id = d.id
JOIN dim_region r ON f.region_id = r.id
WHERE d.year = 2024
GROUP BY d.region, d.month
ORDER BY total_revenue DESC;
```

### Parquet Write (Spark)

```python
df.write \
  .partitionBy("month") \
  .format("parquet") \
  .mode("overwrite") \
  .save("s3://bucket/data/sales/")
```

### Materialized View (PostgreSQL)

```sql
CREATE MATERIALIZED VIEW mv_daily_revenue AS
SELECT
  date_trunc('day', created_at) AS day,
  SUM(amount) AS revenue
FROM orders
GROUP BY 1;

-- Refresh periodically
REFRESH MATERIALIZED VIEW mv_daily_revenue;
```

---

## 14. Interview Discussion

### Key Points

1. **Different workloads**: OLTP = many small, OLAP = few large. Storage and schema design must match.
2. **Columnar for analytics**: Read only needed columns, compress, vectorize. 10–100x faster for aggregations.
3. **Star schema**: Fact table (measures) + dimension tables. Denormalized for fewer joins.
4. **ETL vs ELT**: ETL transforms before load; ELT loads raw then transforms in warehouse. ELT scales with cloud DW.
5. **Materialized views**: Pre-compute for common queries; trade freshness for speed.

### Common Questions

**Q: Why can't I run analytics on my OLTP database?**  
A: OLTP is optimized for writes and point lookups. Full scan + aggregation would lock tables and slow transactions. OLAP is optimized for scans.

**Q: When would you use a star vs snowflake schema?**  
A: Star when you want simpler queries and can tolerate redundancy. Snowflake when storage is a concern and dimensions are large.

**Q: What is vectorized processing?**  
A: Process a column of 1024 values in one CPU instruction (SIMD) instead of 1024 scalar operations. Columnar storage enables this.

**Q: How does Snowflake separate compute and storage?**  
A: Data in object storage (S3/GCS); compute clusters are stateless. Scale compute independently; pay per second of compute.

### Incremental ETL Patterns

1. **Full refresh**: Truncate and reload. Simple but expensive for large tables.
2. **Incremental by timestamp**: `WHERE updated_at > last_run`. Requires monotonic updates.
3. **CDC (Change Data Capture)**: Debezium, Kafka Connect; stream only changed rows from DB log.
4. **Merge/upsert**: Match on key; update if exists, insert if not. Delta Lake, Iceberg support this.

### Compression in Columnar Storage

| Technique | When Effective | Example |
|----------|----------------|---------|
| Dictionary encoding | Low cardinality | `region` column: 10 unique values → store as int |
| Run-length encoding | Sorted columns | `[1,1,1,2,2]` → `(1,3),(2,2)` |
| Bit packing | Small integers | Values 0–100 fit in 7 bits |
| Delta encoding | Sequential values | Timestamps: store deltas |

Combined, these can achieve 10x compression vs. uncompressed row storage.

### Slowly Changing Dimensions (SCD)

When dimension attributes change (e.g., customer moves):
- **SCD Type 1**: Overwrite. Lose history. Simple.
- **SCD Type 2**: Add new row with effective dates. Full history. Larger dimension.
- **SCD Type 3**: Add "previous value" column. Limited history.

Analytics often use Type 2 for audit and point-in-time correctness.

### Query Optimization in OLAP

- **Partition pruning**: `WHERE date = '2024-03-10'` → skip partitions for other dates
- **Predicate pushdown**: Filter in storage layer before loading to memory
- **Column pruning**: Read only columns in SELECT
- **Aggregate pushdown**: SUM in storage if pre-aggregated (e.g., in materialized view)

### Index Types in OLTP

- **B-tree**: Default. Good for range and point lookups. PostgreSQL, MySQL.
- **Hash**: Point lookup only. No range. Faster for equality.
- **Bitmap**: For low-cardinality columns. Multiple conditions → bitwise AND.
- **GiST/GIN**: PostgreSQL. Full-text, JSON, geometric.

### Batch Window Strategies (ETL)

- **Daily**: Run at midnight. Process previous day. Simple, high latency.
- **Hourly**: Lower latency. More jobs.
- **Near real-time**: CDC + stream. Sub-minute latency. Complex.

### Data Warehouse vs Data Mart

- **Data warehouse**: Enterprise-wide, integrated, subject-oriented. Single source of truth.
- **Data mart**: Subset for specific department (e.g., sales mart, HR mart). Can be derived from warehouse or built separately. Faster for departmental queries but risk of inconsistency.

### OLAP Query Patterns

| Pattern | Example | Optimization |
|---------|---------|--------------|
| **Slice** | Fix one dimension (e.g., year=2024) | Partition pruning |
| **Dice** | Fix multiple dimensions | Multi-column filter |
| **Roll-up** | Aggregate (day → month → year) | Pre-aggregate in MV |
| **Drill-down** | Reverse of roll-up | Store at grain |
| **Pivot** | Swap rows/columns | Application layer |

### Denormalization Trade-offs

**Pros**: Fewer joins, faster queries, simpler SQL.
**Cons**: Redundancy, storage cost, update anomalies (if not careful), consistency challenges.

In OLAP, denormalization is common because: (1) reads dominate, (2) batch updates avoid consistency issues, (3) storage is cheap.

### BigQuery vs Snowflake: Architectural Differences

**BigQuery**: Serverless. No clusters. Queries auto-scale. Columnar (Capacitor). Dremel execution (tree of workers). Pay per byte scanned.

**Snowflake**: Virtual warehouses (compute clusters). Scale up/down. Storage in S3/GCS; compute separate. Pay per compute-second + storage. Better for predictable workloads.

### Redshift Spectrum and External Tables

Redshift Spectrum (and similar in Snowflake, BigQuery) allows querying data in S3 without loading into the warehouse. External table = metadata pointing to S3 path. Query pushes down to S3 scan. Pay per data scanned. Useful for ad-hoc exploration, data lake integration.

### dbt (Data Build Tool) in Modern ELT

dbt runs SQL transformations in the warehouse. Models = SQL files; dependencies form DAG. Features: incremental models, tests (unique, not_null), documentation, Jinja templating. Fits ELT: load raw with Fivetran/Airflow; transform with dbt in Snowflake/BigQuery.

### Summary: OLTP vs OLAP at a Glance

| | OLTP | OLAP |
|---|------|------|
| **Workload** | Transactions | Analytics |
| **Query** | Simple, indexed | Complex, full scan |
| **Storage** | Row | Column |
| **Latency** | <10ms | Seconds–minutes |
| **Scale** | GB–TB | TB–PB |

Modern architectures often use both: OLTP for real-time operations, replicate to OLAP for analytics. CDC (Change Data Capture) streams changes; ETL/ELT transforms for the warehouse. Tools like Debezium, Fivetran, and Airflow orchestrate this flow.

### Final Checklist for Interview

- Know row vs column storage and why columnar wins for analytics.
- Explain star schema with fact and dimension tables.
- Articulate ETL vs ELT and when to use each.
- Describe at least one lakehouse format (Delta, Iceberg, or Hudi).

---
