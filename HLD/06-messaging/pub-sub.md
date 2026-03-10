# Publish/Subscribe Pattern — Staff+ Engineer Deep Dive

## 1. Concept Overview

### Definition
**Publish/Subscribe (Pub/Sub)** is a messaging pattern where message producers (publishers) send messages to a logical channel (topic) without knowing the recipients. Subscribers express interest in topics and receive messages matching their subscriptions. The key characteristic: **one message can be delivered to many subscribers** (fan-out).

### Purpose
- **Decoupling**: Publishers and subscribers are completely decoupled; neither knows the other exists
- **Fan-out**: Single publish reaches N subscribers without N separate sends
- **Dynamic topology**: Subscribers can join/leave without publisher changes
- **Event-driven architecture**: Enables reactive, event-sourced systems
- **Broadcast semantics**: Ideal for notifications, events, state propagation

### Problems It Solves
1. **N-to-N coupling**: Avoids publishers maintaining subscriber lists
2. **Synchronous fan-out**: One event → many consumers without blocking
3. **Temporal decoupling**: Subscribers can be offline at publish time
4. **Scalability**: Add subscribers without changing publishers
5. **Flexibility**: Filtering, routing at broker level (topic hierarchies, attributes)

---

## 2. Real-World Motivation

### Google
- **Internal Pub/Sub**: Google Cloud Pub/Sub is based on technology used internally for years
- **Event distribution**: Ads events, Search indexing, Gmail notifications
- **Scale**: Billions of messages per day across thousands of services
- **Dataflow**: Real-time stream processing built on Pub/Sub

### Netflix
- **Recommendation events**: View events published → multiple consumers (recommendation engine, analytics, A/B testing)
- **Content updates**: New titles, metadata changes → CDN invalidation, search index, UI caches
- **Device events**: Playback state → sync across devices, analytics

### Uber
- **Trip events**: Trip started, ended, driver location → dispatch, billing, analytics, ETA
- **Surge pricing**: Price change events → driver app, rider app, analytics
- **Geofence events**: Enter/exit zones → notifications, billing rules

### Amazon
- **SNS**: Notification service (email, SMS, push) built on pub/sub
- **Order events**: Order placed → inventory, shipping, recommendation, fraud detection
- **Inventory events**: Stock changes → search, recommendations, alerts

### Twitter
- **Tweet delivery**: Historically used fan-out (pub/sub) for timeline updates
- **Trending**: Engagement events → trending algorithm, analytics
- **Notifications**: Likes, retweets, mentions → push, email, in-app

---

## 3. Architecture Diagrams

### Basic Pub/Sub Model

```
                    ┌─────────────────────────────────────┐
                    │              TOPIC / CHANNEL         │
                    │         "order-events"               │
                    └─────────────────────────────────────┘
                                        │
        ┌──────────────────────────────┼──────────────────────────────┐
        │                               │                              │
        ▼                               ▼                              ▼
┌───────────────┐              ┌───────────────┐              ┌───────────────┐
│  Subscriber 1 │              │  Subscriber 2 │              │  Subscriber 3 │
│  (Inventory)  │              │  (Shipping)   │              │  (Analytics)  │
└───────────────┘              └───────────────┘              └───────────────┘

        One message published → All three receive (fan-out)
```

### Push vs Pull Subscription

```
PUSH SUBSCRIPTION:
┌──────────┐     ┌─────────────┐     ┌──────────────┐
│ Publisher│────▶│   Topic     │────▶│  Subscriber  │  Broker pushes to endpoint
└──────────┘     └─────────────┘     │  (HTTP)      │  (webhook, serverless)
                                     └──────────────┘

PULL SUBSCRIPTION:
┌──────────┐     ┌─────────────┐     ┌──────────────┐
│ Publisher│────▶│   Topic     │     │  Subscriber  │  Subscriber polls
└──────────┘     └──────┬──────┘     │  (pull)      │  for messages
                       │             └──────▲───────┘
                       │                     │
                       └─────────────────────┘
```

### Fan-In (Multiple Publishers → One Topic)

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ Publisher A │     │ Publisher B │     │ Publisher C │
│ (Orders)    │     │ (Inventory) │     │ (Payments)  │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                   │
       └───────────────────┼───────────────────┘
                           │
                           ▼
                  ┌─────────────────┐
                  │  Topic: events   │
                  └────────┬────────┘
                           │
                           ▼
                  ┌─────────────────┐
                  │   Subscribers   │
                  └─────────────────┘
```

### Message Filtering (Attribute-Based)

```
┌──────────────┐
│  Publisher   │  publish(msg, attributes={region: "us", type: "order"})
└──────┬───────┘
       │
       ▼
┌─────────────────────────────────────────────────────────────┐
│                      TOPIC                                   │
│  Messages with attributes: region, type, priority, etc.      │
└─────────────────────────────────────────────────────────────┘
       │                    │                    │
       │ Filter:             │ Filter:            │ Filter:
       │ region="us"         │ type="order"       │ region="eu"
       ▼                    ▼                    ▼
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│ Sub US only  │     │ Sub orders   │     │ Sub EU only  │
└──────────────┘     └──────────────┘     └──────────────┘
```

---

## 4. Core Mechanics

### Topics and Channels
- **Topic**: Logical channel; messages published to a topic
- **Subscription**: Consumer's interest in a topic; each subscription has independent message delivery
- **Message**: Payload + optional attributes (metadata for filtering)

### Fan-Out Semantics
- One publish → N subscribers (each gets a copy)
- Each subscription has its own cursor/offset (in pull model)
- Late-joining subscribers typically get messages from "now" (no replay unless retention)

### Message Ordering
- **Best-effort**: Most pub/sub systems don't guarantee global order
- **Per-key ordering**: Some (e.g., Pub/Sub with ordering key) guarantee order for messages with same key
- **Trade-off**: Ordering reduces parallelism (same key → same partition/shard)

### At-Least-Once Delivery
- Default: Message may be delivered more than once (retries on ack failure)
- Subscriber must be **idempotent**
- Exactly-once requires: ordering key + deduplication window + idempotent processing

### Message Filtering
- **Topic hierarchy**: e.g., `orders.us.east`, `orders.eu.west`
- **Attribute filters**: Subscription filter expression (e.g., `attributes.region = "us"`)
- **Content-based**: Filter on message body (less common, expensive)

### Push vs Pull Subscriptions

| Push | Pull |
|------|------|
| Broker delivers to endpoint (HTTP) | Subscriber polls for messages |
| Good for: serverless, webhooks | Good for: batch processing, high throughput |
| Must handle retries, backoff | Subscriber controls rate |
| Endpoint must be publicly reachable | No inbound connectivity needed |
| Auto-scaling (serverless) | Need to manage worker scaling |

---

## 5. Numbers

| System | Throughput | Latency | Scale |
|--------|------------|---------|-------|
| **Google Cloud Pub/Sub** | Millions msg/sec (per project) | < 100 ms (push), variable (pull) | 10K+ topics |
| **Amazon SNS** | 30M msg/sec (per topic) | < 100 ms | Unlimited topics |
| **Redis Pub/Sub** | 100K+ msg/sec (single node) | Sub-ms | In-memory, no persistence |
| **Azure Service Bus** | 2K msg/sec (standard tier) | Low ms | Topics + subscriptions |

**Google Cloud Pub/Sub**:
- Message size: 10 MB max
- Retention: 7 days (configurable)
- Ordering: With ordering key, per-key FIFO
- At-least-once delivery

**Amazon SNS**:
- Message size: 256 KB (standard), 1 MB (FIFO)
- Fan-out to SQS, Lambda, HTTP, etc.
- No message retention (fire-and-forget to endpoints)

---

## 6. Tradeoffs

### Pub/Sub vs Message Queue

| Aspect | Pub/Sub | Message Queue |
|--------|---------|---------------|
| **Delivery** | One → many | One → one |
| **Consumers** | Each gets copy | Competing consumers |
| **Use case** | Broadcast, events | Task distribution |
| **Ordering** | Per-topic or per-key | Per-queue FIFO |
| **Replay** | Limited (retention) | Until consumed |

### Push vs Pull

| Push | Pull |
|------|------|
| Simpler for serverless | More control for batch |
| Backpressure on endpoint | Backpressure on broker |
| Need retry logic | Polling overhead |
| Good for low volume | Good for high volume |

---

## 7. Variants / Implementations

### Google Cloud Pub/Sub
- **Topics**: Logical channels
- **Subscriptions**: Pull or push
- **Ordering**: Ordering key for per-key FIFO
- **Filtering**: Filter expressions on attributes
- **Exactly-once**: With subscription-level deduplication (experimental)

### Amazon SNS
- **Topics**: Fan-out to multiple endpoints
- **Protocols**: SQS, Lambda, HTTP/HTTPS, email, SMS
- **FIFO topics**: Ordering + deduplication
- **Message filtering**: Filter policies (JSON matching)

### Redis Pub/Sub
- **Channels**: Simple string channel names
- **No persistence**: Fire-and-forget
- **Pattern subscribe**: `PSUBSCRIBE news.*`
- **Use case**: Real-time, low-latency, ephemeral

### Azure Service Bus
- **Topics + Subscriptions**: Each subscription gets copy
- **Sessions**: Ordering within session
- **Dead-letter**: Failed messages to DLQ
- **Duplicate detection**: Time-window based

---

## 8. Scaling Strategies

1. **Horizontal subscribers**: Add more pull subscribers (competing consumers per subscription)
2. **Topic sharding**: Multiple topics (e.g., by region, entity type)
3. **Filtering**: Reduce message volume per subscription via filters
4. **Push → Pull**: For high volume, use pull to control rate
5. **Partitioning**: Ordering key → partition → ordered per key

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|-------------|
| Subscriber offline | Messages may be lost (no retention) or buffered | Retention (Pub/Sub), DLQ (SNS→SQS) |
| Push endpoint down | Retries, eventually DLQ | Exponential backoff, DLQ, alerting |
| Message too large | Reject | Chunking, external storage (GCS/S3) + reference |
| Filter too broad | Overload subscriber | Narrow filters, separate topics |
| Duplicate delivery | Double processing | Idempotent subscribers, dedup |

---

## 10. Performance Considerations

- **Batching**: Publish in batches (Pub/Sub: 1000 msgs, 10 MB)
- **Compression**: For large payloads
- **Filter selectivity**: Narrow filters reduce subscriber load
- **Push timeout**: Configure ack deadline for push
- **Connection pooling**: Reuse connections for pull

---

## 11. Use Cases

- Notification systems (email, SMS, push)
- Event broadcasting (order created → N services)
- Cache invalidation (invalidate → many cache nodes)
- Real-time dashboards (metrics → UI)
- Audit logging (events → audit service)
- Search index updates (document change → indexer)
- Multi-service orchestration (saga events)

---

## 12. Comparison Tables

### Pub/Sub vs Queue vs Kafka

| Feature | Pub/Sub | Message Queue | Kafka |
|---------|---------|---------------|-------|
| **Model** | Fan-out | Competing consumers | Log, consumer groups |
| **Replay** | Limited | No | Full (offset seek) |
| **Ordering** | Per-key (optional) | Per-queue | Per-partition |
| **Retention** | Days | Until consumed | Days/bytes |
| **Throughput** | Very high | Medium | Very high |
| **Use case** | Events, notifications | Task queues | Event streaming |

### GCP Pub/Sub vs SNS vs Redis

| Feature | GCP Pub/Sub | Amazon SNS | Redis Pub/Sub |
|---------|-------------|------------|----------------|
| **Persistence** | Yes (retention) | No (fire-and-forget) | No |
| **Pull** | Yes | No (push only) | Yes |
| **Filtering** | Yes (attributes) | Yes (filter policy) | Pattern match |
| **Ordering** | Yes (ordering key) | Yes (FIFO topic) | No |
| **Managed** | Yes | Yes | Self/ElastiCache |

---

## 13. Code or Pseudocode

### Google Cloud Pub/Sub Publisher

```python
from google.cloud import pubsub_v1
import json

publisher = pubsub_v1.PublisherClient()
topic_path = publisher.topic_path('my-project', 'order-events')

# Publish single message
data = json.dumps({'order_id': '123', 'amount': 99.99}).encode('utf-8')
future = publisher.publish(topic_path, data, region='us', type='order')
message_id = future.result()

# Batch publish
futures = []
for i in range(100):
    future = publisher.publish(topic_path, f'msg-{i}'.encode())
    futures.append(future)
for f in futures:
    f.result()
```

### Google Cloud Pub/Sub Subscriber (Pull)

```python
from google.cloud import pubsub_v1

subscriber = pubsub_v1.SubscriberClient()
subscription_path = subscriber.subscription_path('my-project', 'order-events-sub')

def callback(message):
    try:
        process(message.data)
        message.ack()
    except Exception:
        message.nack()

flow_control = pubsub_v1.types.FlowControl(max_messages=100)
subscriber.subscribe(subscription_path, callback=callback, flow_control=flow_control)
```

### Amazon SNS Publish

```python
import boto3
sns = boto3.client('sns')
topic_arn = 'arn:aws:sns:us-east-1:123456789:order-events'

sns.publish(
    TopicArn=topic_arn,
    Message=json.dumps({'order_id': '123'}),
    MessageAttributes={
        'region': {'DataType': 'String', 'StringValue': 'us'},
        'type': {'DataType': 'String', 'StringValue': 'order'}
    }
)
```

### Redis Pub/Sub

```python
import redis
r = redis.Redis()

# Publisher
r.publish('orders', json.dumps({'order_id': '123'}))

# Subscriber
pubsub = r.pubsub()
pubsub.subscribe('orders')
for message in pubsub.listen():
    if message['type'] == 'message':
        process(message['data'])
```

---

## 14. Interview Discussion

### Key Points to Cover

1. **Pub/Sub vs Queue**: Fan-out vs competing consumers; when to use each
2. **Push vs Pull**: Trade-offs for serverless vs batch
3. **Ordering**: Per-key ordering, impact on parallelism
4. **At-least-once**: Idempotent consumers, deduplication
5. **Filtering**: Reduce load, topic design
6. **Failure handling**: Retries, DLQ, retention

### Sample Questions

**Q: When would you choose Pub/Sub over a message queue?**
A: When one event must reach multiple independent consumers (e.g., order placed → inventory, shipping, analytics). Queue is for distributing work among competing consumers (one consumer gets each message).

**Q: How does Google Pub/Sub achieve at-least-once delivery?**
A: Messages are stored until acked. If ack fails (timeout, nack), message is redelivered. Subscriber must be idempotent. For exactly-once, use ordering key + dedup window.

**Q: Design a notification system for 10M users.**
A: Pub/Sub topic per notification type. Push subscriptions to Lambda/Cloud Functions for email/SMS. For push notifications, fan-out to device-specific queues. Use filtering (user preferences). Consider batching for email (e.g., digest).

---

## Appendix: Additional Deep Dives

### Topic Hierarchy Design Patterns

**Flat topics**: `orders`, `inventory`, `payments` — simple, few topics.

**Hierarchical**: `events.orders.created`, `events.orders.updated`, `events.inventory.low` — enables wildcard subscriptions (`events.orders.*`).

**Attribute-based**: Single topic, filter by attributes (`region`, `type`, `priority`) — flexible, fewer topics.

### Message Filtering Implementation (GCP Pub/Sub)

```
Subscription filter: attributes.region = "us" AND attributes.priority >= "high"
- Only messages matching filter are delivered
- Reduces subscriber load
- Filter evaluated at publish time
```

### Ordering Key Semantics

With ordering key (e.g., `user_id`):
- Messages with same key go to same partition
- Order preserved for that key
- Throughput reduced (same key = same partition)
- Use for: user session events, entity lifecycle

### Push Subscription Retry Behavior

- **Exponential backoff**: 10s, 20s, 40s, ... (configurable)
- **Dead letter**: After max retries, move to DLQ (if configured)
- **Ack deadline**: Subscriber must ack within deadline or message redelivered
- **Best practice**: Process quickly, extend deadline for long tasks

### Fan-In vs Fan-Out

**Fan-out**: One publisher → many subscribers (broadcast). Use case: notifications, cache invalidation.

**Fan-in**: Many publishers → one topic → subscribers. Use case: aggregating events from multiple sources (orders, inventory, payments → event bus).

### Pub/Sub in Event-Driven Architecture

- **Event**: Something that happened (OrderCreated, UserSignedUp)
- **Publisher**: Emits events; doesn't know consumers
- **Subscriber**: Reacts to events; can be added/removed without publisher change
- **Saga**: Choreography via events (OrderCreated → ReserveInventory → ProcessPayment → ...)

### Cost Optimization (GCP Pub/Sub)

- **Batching**: Up to 1000 messages or 10 MB per publish request
- **Retention**: Shorter retention = less storage cost
- **Filtering**: Narrow filters reduce delivered messages = less compute
- **Pull vs Push**: Pull gives more control over rate; Push can cause cost spikes if subscriber is slow

### SNS + SQS Fan-Out Pattern

Common pattern: SNS topic → multiple SQS queues (each queue = subscriber type). Benefits:
- SQS provides persistence, retry, DLQ
- Decouples SNS (fire-and-forget) from consumer processing
- Each SQS queue can have different consumers, retention, visibility timeout

```
SNS Topic ──▶ SQS Queue 1 (email workers)
         ──▶ SQS Queue 2 (SMS workers)
         ──▶ SQS Queue 3 (push workers)
         ──▶ Lambda (immediate processing)
```

### Redis Pub/Sub Limitations

- **No persistence**: Messages lost if no subscriber connected
- **No replay**: Can't retrieve past messages
- **No acknowledgment**: Fire-and-forget
- **Use case**: Real-time notifications, presence, ephemeral events
- **Alternative**: Redis Streams (persistent, consumer groups, acknowledgment)

### GCP Pub/Sub Ordering Key Details

- Messages with same ordering key are ordered
- Throughput limited by key cardinality (fewer keys = more contention)
- Use for: user events, entity updates, session data
- Without ordering key: best-effort ordering, higher throughput

### Message Size Limits and Chunking

| Service | Limit | Strategy |
|---------|-------|----------|
| GCP Pub/Sub | 10 MB | Chunk + metadata for reassembly |
| SNS | 256 KB (standard) | Store in S3, send reference |
| Redis | 512 MB (practical) | Usually small payloads |

### Subscription Expiration and Cleanup

- **Push**: Endpoint must return 200 within ack deadline
- **Pull**: Messages held until ack or ack deadline
- **Expired**: Message redelivered to same or other subscriber
- **Poison**: Move to DLQ after N nacks (configurable)

### Event-Driven Architecture Patterns

**Event Notification**: Minimal payload; subscribers fetch full data if needed. Reduces message size.

**Event Carried State Transfer**: Full payload in event. Subscribers don't need to query source. Larger messages.

**Event Sourcing**: Events are source of truth; state derived by replaying. Pub/Sub can distribute events.

### Pub/Sub in Microservices

- **Service discovery**: Not needed for pub/sub (broker is discovery)
- **Contract**: Schema for events (Avro, Protobuf) + Schema Registry
- **Versioning**: New schema versions; backward compatibility
- **Choreography**: Services react to events; no central orchestrator
