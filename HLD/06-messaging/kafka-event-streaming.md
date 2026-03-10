# Apache Kafka & Event Streaming — Staff+ Engineer Deep Dive

## 1. Concept Overview

### Definition
**Apache Kafka** is a distributed event streaming platform that provides high-throughput, fault-tolerant, durable storage and processing of event streams. It is built around a **distributed commit log** abstraction: producers append records to topics, and consumers read from topics at their own pace. Kafka blurs the line between messaging and storage—it retains data for configurable periods, enabling replay and multiple consumers.

### Purpose
- **Event streaming**: Capture events (orders, clicks, logs) as they occur
- **Durability**: Events are persisted; multiple consumers can read independently
- **Replay**: Consumers can re-read from any offset (time-travel debugging, new consumers)
- **Decoupling**: Producers and consumers are fully decoupled
- **Scalability**: Partition-based parallelism; linear scaling with brokers

### Problems It Solves
1. **Dual-write consistency**: Single write to Kafka → multiple consumers (CDC, search, analytics)
2. **Backpressure**: Consumers control read rate; no push overload
3. **Replay**: New consumers can process historical data
4. **High throughput**: Millions of messages per second
5. **Event sourcing**: Complete history of events for rebuilding state

---

## 2. Real-World Motivation

### LinkedIn
- **Origin**: Kafka was created at LinkedIn in 2011
- **Scale**: 7+ trillion messages per day (2020+)
- **Use cases**: Activity tracking, metrics, log aggregation, stream processing
- **Kafka Connect**: Built for data integration (DB → Kafka → data lake)

### Netflix
- **Scale**: 700+ billion events per day
- **Use cases**: Playback events, recommendation events, device events, A/B testing
- **Kafka + Flink**: Real-time stream processing for personalization
- **Schema Registry**: Avro schemas for evolution

### Uber
- **Real-time data**: Trip events, driver location, surge pricing
- **Kafka + Flink**: Real-time ETA, fraud detection
- **Multi-datacenter**: Kafka replication across regions

### Airbnb
- **Search indexing**: CDC from MySQL → Kafka → Elasticsearch
- **Event pipeline**: User actions, booking events, analytics
- **Schema evolution**: Protobuf/Avro for compatibility

### Amazon
- **Amazon MSK**: Managed Kafka service
- **Kinesis**: Competing product (similar model)
- **Internal**: Event-driven microservices, analytics pipelines

---

## 3. Architecture Diagrams

### Kafka Cluster Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         KAFKA CLUSTER (Brokers)                              │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐                        │
│  │Broker 0 │  │Broker 1 │  │Broker 2 │  │Broker 3 │  ...                     │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘                          │
│       │            │            │            │                               │
│  ┌────┴────────────┴────────────┴────────────┴────┐                          │
│  │         ZooKeeper / KRaft (Metadata)          │                          │
│  └────────────────────────────────────────────────┘                          │
└─────────────────────────────────────────────────────────────────────────────┘
         ▲                    │                    ▲
         │                    │                    │
    ┌────┴────┐           ┌────┴────┐         ┌────┴────┐
    │Producer │           │  Topic  │         │Consumer │
    │         │           │ Partitions       │  Group  │
    └─────────┘           └─────────┘         └─────────┘
```

### Topic, Partitions, Segments

```
TOPIC: orders
├── Partition 0  [Leader: Broker 0]  [msg0][msg1][msg2][msg3]...  offset 0,1,2,3...
├── Partition 1  [Leader: Broker 1]  [msg0][msg1][msg2]...
├── Partition 2  [Leader: Broker 2]  [msg0][msg1][msg2][msg3][msg4]...
└── Partition 3  [Leader: Broker 3]  [msg0][msg1]...

Each partition = ordered, immutable log
Segments: partition split into files (e.g., 1GB each)
  segment_0.log, segment_1.log, ...
  .index files for offset lookup
```

### Producer Flow (acks, Batching)

```
┌─────────────┐
│  Producer   │
└──────┬──────┘
       │ 1. Serialize, partition (key hash or round-robin)
       │ 2. Add to batch (linger.ms, batch.size)
       │ 3. Optional: compress (snappy, lz4, gzip)
       ▼
┌─────────────────────────────────────────────────────────┐
│  acks=0: Fire and forget (no ack, highest throughput)   │
│  acks=1: Leader ack only (leader may lose before repl)  │
│  acks=all: Leader + all in-sync replicas ack            │
└─────────────────────────────────────────────────────────┘
       │
       ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ Partition 0 │     │ Partition 1 │     │ Partition 2 │
│   Leader    │     │   Leader    │     │   Leader    │
│   Follower  │     │   Follower  │     │   Follower  │
└─────────────┘     └─────────────┘     └─────────────┘
```

### Consumer Group

```
TOPIC: orders (3 partitions)

Consumer Group: order-processors
├── Consumer A  →  Partition 0
├── Consumer B  →  Partition 1
└── Consumer C  →  Partition 2

Each partition assigned to exactly one consumer in group
If Consumer D joins → rebalance (partition reassignment)
If Consumer B leaves → rebalance (partitions redistributed)
```

### ZooKeeper → KRaft Migration

```
TRADITIONAL (ZooKeeper):
┌─────────────┐     ┌─────────────┐
│   Brokers   │────▶│  ZooKeeper  │  Metadata, controller election
└─────────────┘     └─────────────┘

KRaft (Kafka 3.x+):
┌─────────────────────────────────────┐
│  Kafka Cluster                       │
│  Some brokers = controllers (quorum)  │  No ZooKeeper dependency
│  Metadata in Kafka log                │
└─────────────────────────────────────┘
```

---

## 4. Core Mechanics

### Brokers, Topics, Partitions
- **Broker**: Kafka server; hosts partitions
- **Topic**: Logical stream; divided into partitions
- **Partition**: Ordered, immutable sequence of records; unit of parallelism
- **Segment**: Partition split into segment files (retention, compaction)

### Offsets
- **Offset**: Sequential ID per partition; consumer tracks "last read" offset
- **Commit**: Consumer commits offset to `__consumer_offsets` topic
- **Seek**: Jump to specific offset (replay)

### Leaders and Followers
- **Leader**: Handles reads/writes for partition
- **Follower**: Replicates from leader; can become leader on failover
- **ISR (In-Sync Replica)**: Replicas that are caught up; acks=all waits for ISR

### Producer
- **Partitioner**: `hash(key) % num_partitions` (default) or custom
- **acks**: 0 (none), 1 (leader), all (ISR)
- **Batching**: `batch.size`, `linger.ms` for throughput
- **Compression**: snappy, lz4, gzip
- **Idempotent producer**: Prevents duplicates on retry (Kafka 0.11+)
- **Transactional API**: Exactly-once across partitions

### Consumer
- **Consumer group**: Shared group ID; each partition → one consumer
- **Rebalancing**: Triggered by join/leave; pause during rebalance
- **Offset commit**: Auto (periodic) or manual
- **Fetch**: Pull-based; `fetch.min.bytes`, `fetch.max.wait.ms`

### Message Ordering
- **Per-partition only**: Order guaranteed within partition
- **No global order**: Across partitions, order is not guaranteed
- **Ordering key**: Use same key for related messages → same partition

### Log Compaction
- **Key-based retention**: Keep latest value per key; delete older
- **Use case**: Changelog, materialized views
- **Not time-based**: Compacted by key, not time

### Retention
- **Time-based**: e.g., 7 days
- **Size-based**: e.g., 1 TB per partition
- **Compaction**: For compacted topics

### Exactly-Once Semantics
- **Idempotent producer**: Dedup on producer retries
- **Transactional producer**: Atomic write to multiple partitions
- **Read-process-write**: Consumer reads offset, processes, writes to output topic in same transaction

### Kafka Streams
- **Library** (not separate cluster): Stream processing in your app
- **Stateful**: Windows, aggregations, joins
- **Exactly-once**: With transactional producer + consumer
- **Scaling**: Number of threads = number of partitions (or less)

### Kafka Connect
- **Source connectors**: DB, S3, etc. → Kafka
- **Sink connectors**: Kafka → DB, S3, Elasticsearch
- **Distributed mode**: Scalable, fault-tolerant
- **Single message transforms**: Modify records in flight

### Schema Registry
- **Avro, Protobuf, JSON Schema**: Schema evolution
- **Compatibility**: Backward, forward, full
- **Storage**: Schemas stored; producers/consumers fetch by ID

---

## 5. Numbers

| Metric | Value |
|--------|-------|
| **LinkedIn** | 7+ trillion messages/day |
| **Netflix** | 700+ billion events/day |
| **Throughput** | Millions msg/sec per cluster |
| **Latency** | Sub-10 ms (p99) for produce |
| **Retention** | Configurable (days to weeks) |
| **Message size** | Default 1 MB (configurable) |
| **Partitions** | 100K+ per cluster (practical) |
| **Replication** | Typically 2-3x |

---

## 6. Tradeoffs

### acks=0 vs 1 vs all

| acks | Throughput | Durability | Use case |
|------|------------|------------|----------|
| 0 | Highest | None | Metrics, logs |
| 1 | High | Leader only | General |
| all | Lower | Strong | Financial, critical |

### Partition Count
- **More partitions** = more parallelism, but more overhead (metadata, connections)
- **Fewer partitions** = less parallelism
- **Rule of thumb**: Start with # of consumers you expect; can increase (not decrease easily)

### Retention
- **Long retention** = more storage, replay capability
- **Short retention** = less storage, no replay

---

## 7. Variants / Implementations

### Apache Kafka (Open Source)
- Self-managed or Confluent Cloud
- Full feature set
- KRaft mode (no ZooKeeper in 3.x+)

### Confluent Platform
- Schema Registry, ksqlDB, connectors
- Managed: Confluent Cloud

### Amazon MSK
- Managed Kafka
- Serverless option (MSK Serverless)
- Integration with AWS services

### Apache Pulsar
- Multi-tenancy, tiered storage
- Separate compute and storage
- Geo-replication

### Amazon Kinesis
- Similar model (shards ≈ partitions)
- Tight AWS integration
- Different API

---

## 8. Scaling Strategies

1. **Add partitions**: More parallelism (cannot reduce)
2. **Add brokers**: Distribute partition leaders
3. **Add consumers**: Up to # of partitions
4. **Increase replication**: Higher durability
5. **Tune batch size, compression**: Higher throughput
6. **Multi-cluster**: Mirroring, replication for DR

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|------------|
| Broker down | Partition leader failover | Replication, unclean leader election (trade-off) |
| Consumer crash | Rebalance, lag | Heartbeat timeout, quick rebalance |
| Producer retry | Duplicates | Idempotent producer, transactional API |
| Disk full | Broker failure | Monitor disk, retention |
| Network partition | Split brain | Replication factor, min.insync.replicas |

---

## 10. Performance Considerations

- **Batching**: Increase `batch.size`, `linger.ms` for throughput
- **Compression**: lz4/snappy for throughput vs CPU
- **Partition count**: Match consumer count
- **Fetch size**: `fetch.min.bytes` for efficiency
- **Consumer lag**: Monitor; scale consumers if lag grows

---

## 11. Use Cases

- Event sourcing
- CDC (database changes → Kafka)
- Log aggregation
- Metrics
- Stream processing (Kafka Streams, Flink)
- Activity tracking
- Real-time analytics

---

## 12. Comparison Tables

### Kafka vs RabbitMQ vs Pulsar vs Kinesis

| Feature | Kafka | RabbitMQ | Pulsar | Kinesis |
|---------|-------|----------|--------|---------|
| **Model** | Log/stream | Queue | Log + multi-tenant | Stream |
| **Replay** | Yes | No | Yes | Yes (24h default) |
| **Ordering** | Per-partition | Per-queue | Per-partition | Per-shard |
| **Retention** | Configurable | Until consumed | Tiered storage | 24h-7d |
| **Throughput** | Very high | Medium | Very high | High |
| **Managed** | Self/Confluent/MSK | Self/CloudAMQP | Self/StreamNative | AWS |
| **Protocol** | Custom | AMQP | Custom | Custom |

---

## 13. Code or Pseudocode

### Producer (Java)

```java
Properties props = new Properties();
props.put("bootstrap.servers", "localhost:9092");
props.put("key.serializer", StringSerializer.class);
props.put("value.serializer", StringSerializer.class);
props.put("acks", "all");
props.put("enable.idempotence", true);

KafkaProducer<String, String> producer = new KafkaProducer<>(props);
ProducerRecord<String, String> record = new ProducerRecord<>(
    "orders", "order-123", "{\"amount\": 99.99}"
);
producer.send(record, (metadata, exception) -> {
    if (exception != null) log.error(exception);
});
producer.flush();
```

### Consumer (Java)

```java
Properties props = new Properties();
props.put("bootstrap.servers", "localhost:9092");
props.put("group.id", "order-processors");
props.put("enable.auto.commit", "false");
props.put("key.deserializer", StringDeserializer.class);
props.put("value.deserializer", StringDeserializer.class);

KafkaConsumer<String, String> consumer = new KafkaConsumer<>(props);
consumer.subscribe(Collections.singleton("orders"));

while (true) {
    ConsumerRecords<String, String> records = consumer.poll(Duration.ofMillis(100));
    for (ConsumerRecord<String, String> record : records) {
        process(record.value());
        consumer.commitSync(Collections.singletonMap(
            new TopicPartition(record.topic(), record.partition()),
            new OffsetAndMetadata(record.offset() + 1)
        ));
    }
}
```

### Kafka Streams (Word Count)

```java
StreamsBuilder builder = new StreamsBuilder();
KStream<String, String> stream = builder.stream("input-topic");
stream.flatMapValues(v -> Arrays.asList(v.toLowerCase().split("\\W+")))
     .groupBy((k, v) -> v)
     .count()
     .toStream()
     .to("output-topic", Produced.with(Serdes.String(), Serdes.Long()));

KafkaStreams streams = new KafkaStreams(builder.build(), config);
streams.start();
```

---

## 14. Interview Discussion

### Key Points to Cover

1. **Log vs queue**: Kafka is a log; retention, replay, multiple consumers
2. **Partitioning**: Key-based routing, ordering per partition
3. **Consumer groups**: Competing consumers, rebalancing
4. **Exactly-once**: Idempotent producer, transactional API
5. **Scaling**: Partitions, brokers, consumers
6. **When to use**: Event streaming, CDC, high throughput, replay

### Sample Questions

**Q: How does Kafka achieve high throughput?**
A: Sequential disk I/O (log append), batching, compression, zero-copy, partition parallelism. OS page cache for reads.

**Q: Why is message ordering only per-partition?**
A: Partitions enable parallelism. Global order would require single partition = bottleneck. Use same key for related messages.

**Q: How do you ensure exactly-once processing?**
A: Idempotent producer + transactional consumer. Read from Kafka, process, write to output topic in same transaction. Enable `isolation.level=read_committed` for consumer.

---

## Appendix: Additional Deep Dives

### Partition Assignment Strategies

- **Range**: Assign consecutive partitions to consumers (e.g., C1: P0,P1; C2: P2,P3). Can cause imbalance.
- **RoundRobin**: Distribute evenly. Better balance.
- **Sticky**: Minimize movement during rebalance; stick to previous assignment when possible. Default in newer Kafka.
- **Cooperative Sticky**: Incremental rebalancing; fewer pauses.

### Kafka Storage Internals

- **Segment files**: `*.log` (data), `*.index` (offset→position), `*.timeindex` (timestamp→offset)
- **Append-only**: Sequential writes; OS page cache for reads
- **Compaction**: For log-compacted topics; keeps latest per key
- **Retention**: Delete segments older than retention or beyond size limit

### Producer Batching Deep Dive

```
batch.size = 16384 (16 KB default)
linger.ms = 0 (send immediately) or 5 (wait 5ms for more messages)
- Larger batch = higher throughput, higher latency
- linger.ms > 0 = more batching, trade latency for throughput
- Compression happens per batch (snappy, lz4, gzip)
```

### Consumer Lag and Monitoring

- **Lag**: Difference between latest offset and consumer's committed offset
- **Critical**: Lag grows = consumer can't keep up
- **Actions**: Add consumers (up to partition count), optimize consumer, increase partition count
- **Tools**: Kafka Manager, Confluent Control Center, custom metrics (lag per partition)

### Exactly-Once Semantics Implementation

1. **Idempotent producer**: `enable.idempotence=true` — deduplicates retries via producer ID + sequence number
2. **Transactional producer**: `initTransactions()`, `beginTransaction()`, `sendOffsetsToTransaction()`, `commitTransaction()`
3. **Consumer**: `isolation.level=read_committed` — only read committed messages
4. **Kafka Streams**: `processing.guarantee=exactly_once_v2` — end-to-end exactly-once

### Schema Registry and Evolution

- **Backward**: New schema can read old data (add optional fields)
- **Forward**: Old schema can read new data (remove optional fields)
- **Full**: Both backward and forward
- **Compatibility**: Set at topic level; Schema Registry validates

### Kafka vs Kinesis Comparison

| Aspect | Kafka | Kinesis |
|--------|-------|---------|
| **Shard/Partition** | Partition | Shard |
| **Retention** | Configurable (days) | 24h default, up to 7 days |
| **Scaling** | Add partitions (manual) | Split/merge shards |
| **Pricing** | Per broker/hour | Per shard/hour + PUT |
| **Ecosystem** | Rich (Connect, Streams) | Kinesis Data Analytics, Firehose |

### Kafka Connect Single Message Transforms (SMT)

- **Extract**: Extract field from value
- **Insert**: Add header or field
- **Replace**: Modify field value
- **Mask**: Redact sensitive data
- **Filter**: Drop records matching condition
- **Chain**: Multiple SMTs in pipeline

### Kafka Streams State Stores

- **RocksDB**: Default; persistent, local disk
- **In-memory**: Fast, lost on restart
- **Windowed**: For windowed aggregations (tumbling, hopping, session)
- **Changelog**: Backed by Kafka topic for fault tolerance
- **Interactive queries**: Query state store via RPC (ReadYourWrites)

### Unclean Leader Election

- **Unclean**: Allow non-ISR replica to become leader (may lose data)
- **Use case**: Availability over consistency; single replica scenario
- **Default**: Disabled (prefer consistency)
- **Trade-off**: With unclean disabled, partition unavailable if all ISR replicas fail

### Kafka Tiered Storage (KIP-405)

- **Hot tier**: Fast local storage (SSD)
- **Cold tier**: Object storage (S3, GCS)
- **Benefit**: Reduce broker storage cost; retain long history
- **Status**: In development / early adoption

### Kafka Multi-Region Replication

- **MirrorMaker 2**: Topic config sync, offset translation, ACL sync
- **Active-active**: Both regions produce; conflict resolution needed
- **Active-passive**: Primary → secondary; failover on disaster
- **Latency**: Cross-region replication adds latency (10-100ms+)
