# Message Queues — Staff+ Engineer Deep Dive

## 1. Concept Overview

### Definition
A **message queue** is an asynchronous communication mechanism where producers send messages to a queue, and consumers retrieve and process them. Messages are stored in the queue until a consumer successfully processes them, enabling decoupling between services.

### Purpose
- **Decoupling**: Producers and consumers operate independently; neither needs to know about the other's availability
- **Load leveling**: Absorb traffic spikes by buffering messages
- **Reliability**: Messages persist until acknowledged; no message loss on consumer failure
- **Ordering**: FIFO queues guarantee processing order (when supported)
- **Scalability**: Multiple consumers can process from the same queue in parallel

### Problems It Solves
1. **Synchronous coupling**: Eliminates tight request-response dependencies
2. **Backpressure**: Queue depth acts as a natural backpressure signal
3. **Failure isolation**: Consumer crashes don't affect producers
4. **Temporal decoupling**: Producers can send when consumers are offline
5. **Work distribution**: Fan-out work across multiple workers

---

## 2. Real-World Motivation

### Uber
- **Trip event processing**: Driver location updates, ride requests, and payment events flow through queues
- **Async order processing**: Payment confirmation, receipt generation, and notification dispatch
- **Scale**: Millions of events per second during peak hours

### Airbnb
- **Search indexing pipeline**: Property updates, availability changes, and pricing updates flow through SQS
- **Async workflows**: Booking confirmation emails, host notifications, review processing
- **Dual-write avoidance**: Single source of truth → queue → multiple consumers

### Amazon
- **Order fulfillment**: Order placed → SQS → inventory, shipping, billing services
- **Recommendation updates**: User behavior events → queue → ML pipeline
- **SQS origins**: Amazon built SQS for internal use before offering it as a service

### Netflix
- **Video encoding**: Raw uploads → queue → encoding workers (multiple quality tiers)
- **Recommendation events**: View events, ratings → queue → recommendation engine
- **Billing events**: Subscription changes → queue → billing, notification services

### Twitter
- **Tweet delivery**: Fan-out to followers via queues (hybrid with pub/sub)
- **Analytics pipeline**: Engagement events → queue → real-time analytics
- **Media processing**: Image/video uploads → queue → transcoding workers

---

## 3. Architecture Diagrams

### Point-to-Point Messaging Model

```
┌─────────────┐     ┌─────────────────────────────────────┐     ┌─────────────┐
│  Producer 1 │────▶│                                     │     │ Consumer 1 │
└─────────────┘     │                                     │────▶└─────────────┘
                    │           MESSAGE QUEUE             │
┌─────────────┐     │   [msg1][msg2][msg3][msg4][msg5]    │     ┌─────────────┘
│  Producer 2 │────▶│                                     │────▶│ Consumer 2 │
└─────────────┘     │   FIFO ordering (if supported)       │     └─────────────┘
                    │                                     │
                    └─────────────────────────────────────┘
                                    │
                                    ▼
                            Each message consumed
                            by exactly ONE consumer
```

### Producer-Consumer with Multiple Workers

```
                    ┌──────────────────────────────────────────┐
                    │              MESSAGE QUEUE               │
                    │  [m1][m2][m3][m4][m5][m6][m7][m8]...     │
                    └──────────────────────────────────────────┘
                                        │
                    ┌───────────────────┼───────────────────┐
                    │                   │                   │
                    ▼                   ▼                   ▼
              ┌──────────┐       ┌──────────┐       ┌──────────┐
              │Consumer 1│       │Consumer 2│       │Consumer 3│
              │ (Worker) │       │ (Worker) │       │ (Worker) │
              └──────────┘       └──────────┘       └──────────┘
                    │                   │                   │
                    └───────────────────┼───────────────────┘
                                        │
                                        ▼
                              Load distributed evenly
                              (competing consumers)
```

### Dead-Letter Queue (DLQ) Flow

```
┌──────────┐     ┌─────────────┐     ┌──────────┐
│ Producer │────▶│ Main Queue  │────▶│ Consumer │
└──────────┘     └─────────────┘     └────┬─────┘
                                         │
                              Process succeeds?
                              │           │
                         Yes  │           │ No (nack / timeout)
                              │           │
                              ▼           ▼
                         [ACK]      ┌─────────────┐
                                    │ Retry Queue │
                                    │ (or retry)  │
                                    └──────┬──────┘
                                           │
                              After N retries (e.g., 3)
                                           │
                                           ▼
                                    ┌─────────────┐
                                    │     DLQ     │
                                    │ (dead msgs) │
                                    └─────────────┘
```

### RabbitMQ Architecture (Exchanges, Bindings)

```
┌─────────────┐
│  Producer   │
└──────┬──────┘
       │ publish(routing_key="order.created")
       ▼
┌─────────────────────────────────────────────────────────┐
│                    EXCHANGE (e.g., topic)                │
│  Routes based on: routing_key + binding patterns         │
└─────────────────────────────────────────────────────────┘
       │                    │                    │
       │ binding:            │ binding:           │ binding:
       │ order.*             │ *.created          │ order.created
       ▼                     ▼                    ▼
┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│ Queue:      │      │ Queue:      │      │ Queue:      │
│ orders      │      │ created     │      │ audit       │
└──────┬──────┘      └──────┬──────┘      └──────┬──────┘
       │                    │                    │
       ▼                    ▼                    ▼
┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│ Consumer A  │      │ Consumer B  │      │ Consumer C  │
└─────────────┘      └─────────────┘      └─────────────┘
```

---

## 4. Core Mechanics

### Message Acknowledgment (Ack/Nack)

**Ack (Acknowledge)**: Consumer signals successful processing. Message is removed from queue.

**Nack (Negative Acknowledge)**: Consumer signals failure. Message can be:
- Requeued immediately (back to queue)
- Sent to DLQ after max retries
- Discarded (at-most-once)

**Pre-fetch / QoS**: Consumer requests N messages at a time. Until acked, they're "in flight" and invisible to other consumers (visibility timeout in SQS).

### Visibility Timeout (SQS)

When a consumer receives a message, it becomes **invisible** for a configurable duration (0 sec to 12 hours). If the consumer doesn't delete it before timeout:
- Message becomes visible again
- Another consumer (or same) can process it
- Risk: **Duplicate processing** if original consumer was slow but eventually succeeded

**Best practice**: Set visibility timeout > typical processing time. Use heartbeat extension for long-running tasks.

### Delivery Guarantees

| Guarantee | Description | Use Case |
|-----------|-------------|----------|
| **At-most-once** | Message may be lost, never duplicated | Metrics, logs (best-effort) |
| **At-least-once** | Message never lost, may be duplicated | Most business logic (idempotent consumers) |
| **Exactly-once** | No loss, no duplicates | Financial transactions (complex, often via dedup) |

**At-least-once** is the default: ack after processing. If consumer crashes before ack, message is redelivered.

**Exactly-once** requires: idempotent consumers + deduplication (e.g., unique message ID + DB constraint).

### Backpressure

- **Queue depth** = natural backpressure signal
- **Producer**: Can slow down or reject when queue is full (if bounded)
- **Consumer**: Processing rate limits throughput
- **Circuit breaker**: Stop consuming when downstream is failing

### Queue Depth Monitoring

- **CloudWatch** (SQS): `ApproximateNumberOfMessages`, `ApproximateNumberOfMessagesNotVisible`, `ApproximateNumberOfMessagesDelayed`
- **RabbitMQ**: `queue_depth`, `consumer_count`, `message_rates`
- **Alerts**: Queue depth > threshold, consumer lag, DLQ growth

---

## 5. Numbers

| System | Throughput | Latency | Scale |
|--------|------------|---------|-------|
| **Amazon SQS** | 3,000 msg/sec (standard), 300 msg/sec (FIFO) per queue | < 10 ms (send), variable (receive) | Unlimited queues |
| **RabbitMQ** | 50K-100K msg/sec per node | Sub-millisecond | Clustered, 100K+ msg/sec |
| **ActiveMQ** | 10K-50K msg/sec | Low ms | Clustered |
| **ZeroMQ** | 1M+ msg/sec (in-memory) | Microseconds | No broker, peer-to-peer |

**SQS specifics**:
- Message size: 256 KB max
- Retention: 1 min to 14 days
- Long polling: Up to 20 sec (reduces empty receives)

**RabbitMQ**:
- Message size: No hard limit (practical ~128 MB)
- Persistence: Disk or memory
- Prefetch: Typically 1-100

---

## 6. Tradeoffs

### Standard vs FIFO (SQS)

| Aspect | Standard | FIFO |
|--------|----------|------|
| Ordering | Best-effort | Strict FIFO |
| Throughput | 3,000 msg/sec | 300 msg/sec |
| Duplicates | At-least-once (may duplicate) | Exactly-once (with dedup) |
| Cost | Lower | Higher |
| Use case | High throughput, order not critical | Order-critical (e.g., audit) |

### Sync vs Async Processing

| Sync | Async (Queue) |
|------|---------------|
| Simple, immediate feedback | Complex, eventual consistency |
| Tight coupling | Loose coupling |
| Blocking, latency sensitive | Non-blocking, absorb spikes |
| Hard to scale | Easy to scale (add workers) |

---

## 7. Variants / Implementations

### RabbitMQ (AMQP)

- **Exchanges**: `direct`, `topic`, `fanout`, `headers`
- **Bindings**: Routing rules from exchange to queue
- **VHosts**: Multi-tenancy
- **Durable queues**: Survive broker restart
- **Clustering**: Mirrored queues for HA

### Amazon SQS

- **Standard**: High throughput, at-least-once
- **FIFO**: Ordering, exactly-once with content-based dedup
- **Long polling**: Reduce empty receives (cost, latency)
- **Visibility timeout**: 0–43200 sec
- **Dead-letter queues**: Failed messages for analysis

### ActiveMQ

- JMS compliant
- Supports queues and topics
- Persistence: KahaDB, JDBC
- Clustering: Master-slave, network of brokers

### ZeroMQ

- **No broker**: Peer-to-peer
- **Patterns**: REQ/REP, PUB/SUB, PUSH/PULL
- **In-process**: No network for same-machine
- **Trade-off**: No persistence, no built-in reliability

---

## 8. Scaling Strategies

1. **Horizontal consumer scaling**: Add more consumers to same queue (competing consumers)
2. **Queue sharding**: Multiple queues, producers round-robin or hash
3. **Priority queues**: Separate queues per priority, consumers prefer high-priority
4. **Partitioning**: Hash key (e.g., user_id) to maintain ordering per key across partitions
5. **Batch processing**: SQS batch send (10 msgs), batch receive (10 msgs) — reduce API calls

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|-------------|
| Consumer crash before ack | Message redelivered | Idempotent processing, visibility timeout |
| Queue full | Producer blocked/rejected | Increase limit, add consumers, backpressure |
| Broker down | Messages unavailable | Clustering, replication, multi-AZ |
| Poison message | Infinite retries | DLQ after N retries, alert on DLQ |
| Duplicate delivery | Double processing | Idempotency keys, dedup in consumer |

---

## 10. Performance Considerations

- **Batching**: SQS batch send/receive reduces API calls 10x
- **Long polling**: Reduces empty receives, lowers cost
- **Prefetch**: Tune for consumer throughput vs memory
- **Compression**: For large payloads (application-level)
- **Connection pooling**: Reuse connections (AMQP, HTTP)

---

## 11. Use Cases

- Order processing pipelines
- Email/SMS/notification dispatch
- Search index updates (Airbnb)
- Video encoding (Netflix)
- Payment reconciliation
- Audit logging
- Async API offload (webhook processing)

---

## 12. Comparison Tables

### RabbitMQ vs SQS vs Kafka

| Feature | RabbitMQ | Amazon SQS | Kafka |
|---------|-----------|------------|-------|
| **Model** | Queue (point-to-point) | Queue | Log / stream |
| **Ordering** | Per-queue (FIFO) | FIFO queue type | Per-partition |
| **Retention** | Until consumed | 1 min–14 days | Configurable (days/bytes) |
| **Replay** | No | No | Yes (offset seek) |
| **Throughput** | 50K–100K/sec | 3K/sec (standard) | Millions/sec |
| **Protocol** | AMQP | HTTP/HTTPS | Custom |
| **Managed** | Self/cloud | Fully managed | Self/Confluent/MSK |
| **Use case** | Task queues, RPC | Simple async, serverless | Event streaming, analytics |

---

## 13. Code or Pseudocode

### Producer (SQS-style)

```python
import boto3
sqs = boto3.client('sqs')
queue_url = 'https://sqs.us-east-1.amazonaws.com/123456789/my-queue'

# Send message
response = sqs.send_message(
    QueueUrl=queue_url,
    MessageBody=json.dumps({'order_id': '123', 'amount': 99.99}),
    MessageGroupId='order-123',  # FIFO: ordering within group
    MessageDeduplicationId='order-123-v1',  # FIFO: dedup
    DelaySeconds=0
)

# Batch send (up to 10)
sqs.send_message_batch(
    QueueUrl=queue_url,
    Entries=[
        {'Id': '1', 'MessageBody': '...'},
        {'Id': '2', 'MessageBody': '...'},
    ]
)
```

### Consumer (SQS-style with visibility timeout)

```python
while True:
    response = sqs.receive_message(
        QueueUrl=queue_url,
        MaxNumberOfMessages=10,
        WaitTimeSeconds=20,  # Long polling
        VisibilityTimeout=30
    )
    messages = response.get('Messages', [])
    for msg in messages:
        try:
            process(msg['Body'])
            sqs.delete_message(
                QueueUrl=queue_url,
                ReceiptHandle=msg['ReceiptHandle']
            )
        except Exception:
            # Don't delete - will reappear after visibility timeout
            pass
```

### RabbitMQ Producer (Python, pika)

```python
import pika
connection = pika.BlockingConnection(pika.ConnectionParameters('localhost'))
channel = connection.channel()
channel.queue_declare(queue='orders', durable=True)

channel.basic_publish(
    exchange='',
    routing_key='orders',
    body=json.dumps({'order_id': '123'}),
    properties=pika.BasicProperties(delivery_mode=2)  # persistent
)
connection.close()
```

### RabbitMQ Consumer (with ack)

```python
def callback(ch, method, properties, body):
    try:
        process(body)
        ch.basic_ack(delivery_tag=method.delivery_tag)
    except Exception:
        ch.basic_nack(delivery_tag=method.delivery_tag, requeue=True)

channel.basic_qos(prefetch_count=1)
channel.basic_consume(queue='orders', on_message_callback=callback)
channel.start_consuming()
```

---

## 14. Interview Discussion

### Key Points to Cover

1. **When to use queues**: Async processing, decoupling, load leveling, reliability
2. **Delivery guarantees**: Explain at-least-once, idempotency, exactly-once trade-offs
3. **Visibility timeout**: Why it exists, duplicate risk, heartbeat for long tasks
4. **DLQ**: Purpose, monitoring, replay strategy
5. **Scaling**: Competing consumers, queue sharding, priority queues
6. **SQS vs RabbitMQ vs Kafka**: Queue vs stream, retention, replay, throughput

### Sample Questions

**Q: How do you ensure exactly-once processing with SQS?**
A: Use FIFO queues with content-based deduplication (5-min window). Consumer must be idempotent (e.g., unique constraint on message ID). For standard queues, implement application-level deduplication.

**Q: What happens if visibility timeout is too short?**
A: Message becomes visible before processing completes. Another consumer gets it → duplicate processing. Set timeout > p99 processing time; use heartbeat/extension for long tasks.

**Q: How would you design a system to process 100K orders/sec?**
A: Shard queues by order_id hash. Multiple SQS queues or RabbitMQ with consistent-hash exchange. Scale consumers horizontally. Use batching. Consider Kafka if replay or event sourcing is needed.

---

## Appendix: Additional Deep Dives

### Priority Queue Implementation Strategies

**Multiple Queue Approach**: Maintain separate queues per priority level (high, medium, low). Consumers poll high-priority queue first, then medium, then low. Simple but can starve low-priority messages.

**Single Queue with Priority Field**: All messages in one queue with priority metadata. Consumer sorts or broker supports priority. RabbitMQ supports `x-max-priority` (0-255) for priority queues.

**Weighted Fair Queuing**: Virtual queues with weights; higher priority gets more processing time. Complex to implement correctly.

### Message Deduplication Strategies

1. **Content-based dedup (SQS FIFO)**: SHA-256 of body; 5-minute dedup window
2. **Application-level**: Producer assigns unique ID; consumer checks DB before processing
3. **Idempotency keys**: Same key within window → treat as duplicate
4. **Bloom filter**: Probabilistic dedup for high volume (false positives possible)

### Backpressure Handling Patterns

- **Queue depth threshold**: Producer slows when depth > N
- **Reject with retry**: Return 503, client backs off exponentially
- **Circuit breaker**: Stop sending when consumer error rate high
- **Shedding**: Drop low-priority messages under load
- **Rate limiting**: Token bucket at producer

### RabbitMQ Exchange Types Deep Dive

| Type | Routing | Use Case |
|------|---------|----------|
| **Direct** | routing_key exact match | Point-to-point, simple routing |
| **Topic** | routing_key pattern (e.g., `order.*.created`) | Multi-subscriber, hierarchical |
| **Fanout** | No routing, broadcast to all bound queues | Broadcast, pub/sub |
| **Headers** | Match message headers | Complex routing logic |

### SQS Long Polling vs Short Polling

**Short polling**: Immediate response; may return empty (empty receive still counts as request). Higher cost, more API calls.

**Long polling**: Wait up to 20 sec for messages; reduces empty receives. Set `WaitTimeSeconds=20`. Lower cost, better latency for consumers.

### At-Least-Once Implementation Checklist

1. Consumer processes message
2. Consumer performs side effects (DB write, API call)
3. Consumer acks message **after** side effects
4. If crash between 2 and 3: message redelivered, consumer must be idempotent
5. Idempotency: Unique constraint on (message_id, entity_id) or idempotency key in DB

### ZeroMQ Patterns for Message Queues

**PUSH/PULL**: Distributed task queue. Pusher sends to multiple pullers; load balanced (round-robin). No persistence.

**REQ/REP**: Request-reply; synchronous. Not a queue but RPC pattern.

**DEALER/ROUTER**: Async request-reply; allows multiple outstanding requests.

**When to use ZeroMQ**: Ultra-low latency, no broker, in-process or same-datacenter. Not for durability.

### ActiveMQ vs RabbitMQ

| Feature | ActiveMQ | RabbitMQ |
|---------|----------|----------|
| **Protocol** | OpenWire, AMQP, MQTT, STOMP | AMQP, MQTT, STOMP |
| **Persistence** | KahaDB, JDBC | Mnesia, plugin (e.g., PostgreSQL) |
| **Clustering** | Master-slave, network of brokers | Mirrored queues, quorum queues |
| **JMS** | Full JMS 1.1/2.0 | Via plugin |
| **Ecosystem** | Apache ecosystem | Pivotal/VMware, broad adoption |

### Queue Depth Alerting Best Practices

- **Warning**: Depth > 1000 (or 10% of daily volume)
- **Critical**: Depth > 10000 or growing 10% per minute
- **DLQ**: Alert on any message in DLQ
- **Consumer lag**: If using Kafka-style offset, alert on lag > N
- **Visibility timeout**: Alert if many messages reappearing (receive count > 2)

### Message Serialization Formats

| Format | Size | Speed | Schema | Use Case |
|--------|------|-------|--------|----------|
| JSON | Large | Slow | No | Debugging, interoperability |
| Protobuf | Small | Fast | Yes | Production, schema evolution |
| Avro | Small | Fast | Yes | Kafka, Schema Registry |
| MessagePack | Small | Fast | No | Binary JSON alternative |

### Redelivery and Retry Policies

- **Immediate requeue**: Nack with requeue=true — back to queue immediately
- **Delayed requeue**: Send to delay queue, TTL, then back to main queue (RabbitMQ)
- **Exponential backoff**: Retry after 1s, 2s, 4s, 8s...
- **Max retries**: After N failures, move to DLQ
- **Dead letter handling**: Manual inspection, replay, or discard
