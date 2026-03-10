# Module 8: Messaging & Streams

---

## 1. Log-Based Message Broker (Kafka)

### Definition
A message broker where the log (append-only sequence of records) IS the data structure. Messages are retained on disk; consumers track their own offset.

### Traditional Queue vs Log-Based
```
Traditional Queue (RabbitMQ):          Log-Based (Kafka):
  Message consumed → DELETED            Message consumed → RETAINED
  1 message → 1 consumer               1 message → many consumers
  No replay                            Full replay by resetting offset

┌─────────────────────────┐     ┌──────────────────────────────┐
│ Queue: [msg3][msg2]     │     │ Log: [msg1][msg2][msg3][msg4]│
│        (msg1 deleted)   │     │       ↑ Consumer A (offset=1)│
└─────────────────────────┘     │            ↑ Consumer B (o=3)│
                                └──────────────────────────────┘
```

### Kafka Architecture
```
┌─── Topic: "orders" ──────────────────────────┐
│                                               │
│  Partition 0: [msg0][msg1][msg2][msg3]        │
│  Partition 1: [msg0][msg1][msg2]              │
│  Partition 2: [msg0][msg1][msg2][msg3][msg4]  │
│                                               │
│  Each partition: ordered, append-only          │
│  Across partitions: no ordering guarantee      │
└───────────────────────────────────────────────┘

Producer → hash(key) → partition
Consumer Group: 1 partition → 1 consumer (parallel processing)
```

### Key Properties
- **Retention**: 7 days default (configurable, or forever)
- **Replay**: Reset consumer offset to re-process
- **Ordering**: Guaranteed within a partition
- **Throughput**: Millions of messages/sec (sequential I/O + zero-copy)

### Real Systems
LinkedIn (Kafka origin), Uber, Netflix, Confluent, Amazon MSK

### Summary
Kafka uses an append-only log with consumer-tracked offsets. Messages persist, enabling replay and multiple independent consumers. Foundation of modern event streaming.

---

## 2. Exactly-Once Semantics

### Definition
Guaranteeing that each message is processed exactly once — no duplicates, no losses.

### The Three Delivery Guarantees
```
┌─── AT-MOST-ONCE ───────────────┐
│ Fire and forget. No retry.      │
│ Message may be LOST.            │
│ Example: UDP, metrics logging   │
│ Fast but unreliable.            │
└─────────────────────────────────┘

┌─── AT-LEAST-ONCE ──────────────┐
│ Retry until ACK received.       │
│ Message may be DUPLICATED.      │
│ Example: Most message queues    │
│ Reliable but may double-process │
└─────────────────────────────────┘

┌─── EXACTLY-ONCE ───────────────┐
│ Each message processed 1 time.  │
│ Requires idempotency or         │
│ transactional guarantees.       │
│ Example: Kafka + Flink          │
│ Correct but complex/slower.     │
└─────────────────────────────────┘
```

### How Kafka Achieves Exactly-Once
```
1. Idempotent Producer:
   Producer gets a ProducerID + sequence number
   Broker deduplicates based on (ProducerID, Sequence)
   → No duplicate writes

2. Transactional Writes:
   Read from input topic → process → write to output topic
   All within an atomic transaction
   → Consume + produce is atomic

3. Consumer: read_committed isolation
   Only sees messages from committed transactions
```

### Visual
```
  At-Least-Once:
  Producer ──msg──→ Broker ──ACK lost──→ Producer retries
                   [msg][msg]  ← duplicate!

  Exactly-Once (Kafka):
  Producer ──msg(seq=5)──→ Broker: "seq=5 already received" → deduplicate
                          [msg]  ← exactly once
```

### Summary
Exactly-once requires idempotent producers (dedup by sequence number) and transactional consume-process-produce cycles. Most systems achieve "effectively once" via idempotency.

---

## 3. At-Least-Once Delivery

### How It Works
```
1. Producer sends message
2. Broker stores and ACKs
3. If ACK lost → Producer retries → duplicate on broker
4. Consumer processes message
5. Consumer ACKs broker
6. If ACK lost → Broker redelivers → duplicate processing
```

### Handling Duplicates
- **Idempotent consumers**: Processing same message twice has no side effect
- **Deduplication table**: Store processed message IDs
- **Exactly-once semantics**: Use transactional processing

### When to Use
Most production systems. Losing messages is worse than processing duplicates. Make consumers idempotent.

---

## 4. At-Most-Once Delivery

### How It Works
```
1. Producer sends message (no retry on failure)
   OR
2. Consumer ACKs BEFORE processing
   → If consumer crashes during processing, message is lost

Fire-and-forget. Message delivered 0 or 1 times.
```

### When to Use
- Metrics/logging where losing a data point is acceptable
- Real-time gaming (stale position updates are useless)
- High-frequency sensor data (next reading arrives soon)

---

## 5. Pub/Sub vs Message Queuing

### Comparison
```
┌─── MESSAGE QUEUE (Point-to-Point) ────┐
│                                        │
│  Producer → [Queue] → Consumer A       │
│                    → Consumer B        │
│                                        │
│  Each message consumed by ONE consumer │
│  (load balancing)                      │
│  Examples: SQS, RabbitMQ queues        │
└────────────────────────────────────────┘

┌─── PUB/SUB (Fan-out) ────────────────┐
│                                       │
│  Publisher → [Topic] → Subscriber A   │
│                     → Subscriber B    │
│                     → Subscriber C    │
│                                       │
│  Each message delivered to ALL subs   │
│  (broadcast)                          │
│  Examples: SNS, Kafka topics, Redis   │
└───────────────────────────────────────┘
```

### When to Use
```
Queue: Work distribution (resize images, send emails)
  → 1 message = 1 worker processes it

Topic: Event notification (UserSignedUp)
  → Email service, analytics service, CRM all need to know
```

### Kafka = Both
```
Topic with 1 consumer group  = Pub/Sub (each group gets all messages)
Topic with N consumer groups = Queue within each group (partitions distributed)
```

---

## 6. Dead Letter Queue (DLQ)

### Definition
A queue where messages that repeatedly fail processing are sent instead of being retried forever.

### Flow
```
Main Queue: [msg1][msg2][msg3]
                    ↓
Consumer tries msg2 → FAILS
Consumer retries msg2 → FAILS (attempt 2)
Consumer retries msg2 → FAILS (attempt 3 = max retries)
                    ↓
DLQ: [msg2]  ← moved here, alert human

Main Queue continues: [msg1][msg3] ← no longer blocked
```

### Why DLQs Matter
- **Poison pill prevention**: Bad message doesn't block entire queue
- **Debugging**: Inspect failed messages in DLQ
- **Alerting**: Monitor DLQ depth for operational issues
- **Reprocessing**: Fix bug, replay DLQ messages

### Real Systems
Amazon SQS (native DLQ), RabbitMQ, Azure Service Bus, Kafka (custom implementation)

---

## 7. Change Data Capture (CDC)

### Definition
Capturing row-level changes (INSERT, UPDATE, DELETE) from a database's transaction log and streaming them as events.

### Problem It Solves
```
WRONG (Dual Write):
  App → Write to DB
  App → Write to Kafka/Elasticsearch
  (If one fails, they're inconsistent!)

RIGHT (CDC):
  App → Write to DB (single source of truth)
  CDC → Read DB transaction log → Stream to Kafka
  (DB log is the source of truth, always consistent)
```

### Visual
```
  ┌──────┐   SQL    ┌──────────┐
  │ App  │ ──────→ │ Database │
  └──────┘         │ (MySQL)  │
                   └────┬─────┘
                        │ Binlog
                        ▼
                   ┌──────────┐     ┌─────────────┐
                   │ Debezium │ ──→ │    Kafka     │
                   │  (CDC)   │     │              │
                   └──────────┘     └──┬──────┬───┘
                                      │      │
                                      ▼      ▼
                                   [Elastic] [Redis]
                                   (search)  (cache)
```

### Use Cases
- **Cache invalidation**: DB change → invalidate Redis key
- **Search indexing**: DB change → update Elasticsearch
- **Data replication**: Sync databases across regions
- **Event sourcing**: Turn DB into event stream

### Real Systems
Debezium (open source), AWS DMS, Maxwell (MySQL), Striim

### Summary
CDC reads the database transaction log to stream changes as events. It eliminates dual-write inconsistency and enables reliable cache invalidation, search indexing, and cross-system sync.

---

## 8. Actor Model

### Definition
A concurrency model where "actors" are isolated units of state that communicate exclusively through asynchronous messages. No shared memory, no locks.

### Core Principles
```
Each Actor:
  1. Has private state (no external access)
  2. Has a mailbox (message queue)
  3. Processes one message at a time
  4. Can create child actors
  5. Can send messages to other actors

No shared state → No locks → No race conditions
```

### Visual
```
  ┌────────────┐  message  ┌────────────┐
  │  Actor A   │ ────────→ │  Actor B   │
  │ [state=5]  │           │ [mailbox:  │
  │            │           │  msg1,msg2]│
  └────────────┘           │ [state=10] │
                           └────────────┘
                           Processes msg1, then msg2 (sequential)
```

### Supervision (Erlang/Akka)
```
       [Supervisor]
      /     |      \
  [Actor1] [Actor2] [Actor3]

Actor2 crashes → Supervisor restarts Actor2
"Let it crash" philosophy — actors are cheap to restart
```

### Real Systems
- **WhatsApp**: Erlang actors (2M connections per server)
- **Akka**: JVM actor framework (Lightbend)
- **Orleans**: .NET virtual actors (Xbox/Halo)
- **Elixir/Phoenix**: Web framework on Erlang VM

### Summary
The actor model eliminates shared state by isolating state in actors that communicate via messages. Enables massive concurrency (WhatsApp: 2M connections per server on Erlang).

---

## 9. Disruptor Pattern

### Definition
A lock-free ring buffer for inter-thread communication, achieving millions of operations per second by eliminating locks, cache misses, and GC overhead.

### Problem It Solves
Traditional queues use locks (mutex) for thread safety. Lock contention limits throughput to ~1M ops/sec. LMAX needed 6M+ TPS for trading.

### How It Works
```
Ring Buffer (pre-allocated array):
┌────┬────┬────┬────┬────┬────┬────┬────┐
│ s0 │ s1 │ s2 │ s3 │ s4 │ s5 │ s6 │ s7 │  ← fixed-size slots
└────┴────┴────┴────┴────┴────┴────┴────┘
       ↑ write cursor          ↑ read cursor

Key tricks:
  1. Pre-allocate all slots (no GC)
  2. Single writer (no write contention)
  3. CAS (Compare-And-Swap) for sequence claiming
  4. Cache-line padding (no false sharing)
  5. Busy-spin wait (no OS context switch)
```

### Why It's Fast
```
Traditional Queue:               Disruptor:
  Lock acquire                    No locks (CAS only)
  Allocate memory for message     Pre-allocated ring buffer
  Enqueue                         Write to slot
  Lock release                    Increment sequence
  Context switch possible         Busy spin (no OS involvement)
  ~1M ops/sec                     ~100M ops/sec
```

### Real Systems
LMAX Exchange (6M TPS single-threaded), Log4j2 (async logging), many HFT systems

### Summary
The Disruptor is a lock-free ring buffer achieving extreme throughput via pre-allocation, single-writer principle, CAS operations, and cache-line padding. Powers high-frequency trading systems.

---

## 10. Lambda Architecture

### Definition
A data processing architecture with parallel batch and speed layers, merged at query time for complete, accurate results.

### Architecture
```
                    ┌─── Speed Layer (Stream) ───┐
                    │ Real-time, approximate      │
  Data ─── Kafka ──→│ Flink/Storm                 │──→ Real-time View
  Stream            │ Low latency, partial data   │
                    └─────────────────────────────┘
                    │
                    │
                    ┌─── Batch Layer ─────────────┐
                    │ Complete, accurate           │──→ Batch View
                    │ Spark/MapReduce              │
                    │ High latency (hours)         │
                    └─────────────────────────────┘

  Query = merge(Batch View + Real-time View)
```

### Kappa Architecture (Modern Replacement)
```
  Everything is a stream:
  Data ──→ Kafka ──→ Flink (stream processing) ──→ Serving Layer

  Reprocessing: replay Kafka from beginning
  No separate batch layer needed!
```

### Lambda vs Kappa

| | Lambda | Kappa |
|-|--------|-------|
| Code paths | Two (batch + stream) | One (stream only) |
| Maintenance | Higher (dual logic) | Lower |
| Reprocessing | Batch re-run | Replay stream |
| Accuracy | Batch corrects stream | Stream must be correct |

### Summary
Lambda architecture combines batch (accurate, slow) and speed (fast, approximate) layers. Kappa simplifies this to stream-only processing with replay capability.
