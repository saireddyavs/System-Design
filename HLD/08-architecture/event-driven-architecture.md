# Event-Driven Architecture

## 1. Concept Overview

### Definition
Event-Driven Architecture (EDA) is a software design pattern in which the flow of the application is determined by events—discrete occurrences that signify a change in state. Components communicate by producing and consuming events asynchronously, enabling loose coupling, scalability, and resilience.

### Purpose
- **Loose coupling**: Producers and consumers don't need to know each other; they only know the event schema
- **Asynchronous processing**: Non-blocking; systems can process at their own pace
- **Scalability**: Independent scaling of producers and consumers
- **Resilience**: Message brokers buffer events; consumers can replay on failure
- **Auditability**: Event log provides complete history of state changes

### Problems It Solves
- **Tight coupling**: Services don't need direct dependencies
- **Synchronous bottlenecks**: Eliminates request-reply chains
- **Data consistency**: Eventual consistency across distributed systems
- **Audit requirements**: Immutable event log for compliance
- **Real-time reactivity**: Multiple consumers can react to same event

---

## 2. Real-World Motivation

### Uber
- **Event streaming**: Kafka for real-time trip events, driver location, rider requests
- **Scale**: Billions of events daily across ride matching, pricing, ETA
- **Use case**: Trip lifecycle—RequestCreated → DriverAssigned → TripStarted → TripEnded
- **Consumers**: Billing, analytics, notifications, fraud detection all consume same events

### Netflix
- **Keystone pipeline**: Kafka-based event streaming for user activity
- **Events**: Play, pause, seek, device info, playback quality
- **Consumers**: Recommendations, A/B testing, content delivery optimization
- **Scale**: Millions of events per second during peak

### Walmart Labs
- **Event Sourcing**: Order management uses event sourcing for audit trail
- **Black Friday**: Event-driven architecture handles 10x traffic spikes
- **Inventory**: Real-time inventory updates via events across 11,000+ stores

### Amazon
- **Event-driven order fulfillment**: Order placed → Inventory reserved → Payment charged → Shipment created
- **SQS/SNS**: Early adopters of message queues and pub/sub
- **EventBridge**: Serverless event bus for application integration

---

## 3. Architecture Diagrams

### Event Notification vs Event-Carried State Transfer

```
EVENT NOTIFICATION (Reference only)
┌─────────────┐                    ┌─────────────┐
│  Producer  │─── OrderCreated ───▶│  Consumer   │
│  (Order)   │    {orderId: 123}   │  (Billing)  │
└─────────────┘                    └──────┬─────┘
       │                                 │
       │                                 │ Must call Order Service
       │                                 │ to get full order details
       │                                 ▼
       │                          ┌─────────────┐
       └─────────────────────────│Order Service│
                                 └─────────────┘
```

```
EVENT-CARRIED STATE TRANSFER (Self-contained)
┌─────────────┐                    ┌─────────────┐
│  Producer  │─── OrderCreated ───▶│  Consumer   │
│  (Order)   │    {orderId, items, │  (Billing)  │
│            │     total, customer}│             │
└─────────────┘                    └─────────────┘
       │                                 │
       │                    Consumer has all data needed
       │                    No additional service calls
       ▼
```

### Event Sourcing Architecture

```
                    ┌─────────────────────────────────────┐
                    │           EVENT STORE                │
                    │  ┌───────────────────────────────┐  │
                    │  │ Order-123: OrderCreated        │  │
                    │  │ Order-123: ItemAdded           │  │
                    │  │ Order-123: ItemAdded           │  │
                    │  │ Order-123: PaymentReceived     │  │
                    │  │ Order-123: OrderShipped        │  │
                    │  └───────────────────────────────┘  │
                    └──────────────▲─────────────────────┘
                                   │ Append-only
         ┌─────────────────────────┼─────────────────────────┐
         │                         │                         │
    ┌────▼────┐              ┌─────▼─────┐              ┌─────▼─────┐
    │ Order   │              │ Payment   │              │ Shipping  │
    │ Service │              │ Service   │              │ Service   │
    └────┬────┘              └─────┬─────┘              └─────┬─────┘
         │                         │                         │
         └─────────────────────────┼─────────────────────────┘
                                   │
                    ┌──────────────▼─────────────────────┐
                    │     READ MODELS (Projections)       │
                    │  OrderSummary | OrderHistory | ...  │
                    └────────────────────────────────────┘
```

### CQRS Architecture

```
                    ┌─────────────────────────────────────┐
                    │              COMMANDS                 │
                    │  CreateOrder, AddItem, PlaceOrder     │
                    └────────────────┬────────────────────┘
                                     │
                    ┌────────────────▼────────────────────┐
                    │           WRITE MODEL                │
                    │     (Normalized, transactional)      │
                    │     Event Store / Domain DB           │
                    └────────────────┬────────────────────┘
                                     │
                          Publish domain events
                                     │
                    ┌────────────────▼────────────────────┐
                    │      PROJECTIONS / HANDLERS          │
                    │  Build read-optimized views          │
                    └────────────────┬────────────────────┘
                                     │
                    ┌────────────────▼────────────────────┐
                    │            READ MODEL                │
                    │  (Denormalized, query-optimized)     │
                    │  OrderListView | OrderDetailView     │
                    └─────────────────────────────────────┘
                                     │
                    ┌────────────────▼────────────────────┐
                    │              QUERIES                 │
                    │  GetOrderList, GetOrderDetails       │
                    └─────────────────────────────────────┘
```

### Saga: Choreography vs Orchestration

```
CHOREOGRAPHY (Decentralized)
┌─────────┐     OrderCreated      ┌─────────┐     InventoryReserved
│  Order  │─────────────────────▶│Inventory│─────────────────────┐
└─────────┘                       └─────────┘                      │
                                                                   │
┌─────────┐     PaymentReceived   ┌─────────┐     InventoryReserved │
│ Payment │◀─────────────────────│  ???    │◀─────────────────────┘
└────┬────┘                       └─────────┘
     │
     │ PaymentFailed
     ▼
┌─────────┐     ReleaseInventory
│Inventory│◀───────────────────── (Compensation)
└─────────┘
```

```
ORCHESTRATION (Centralized)
                    ┌─────────────────┐
                    │ Saga Orchestrator│
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
        ▼                    ▼                    ▼
   ┌─────────┐         ┌─────────┐         ┌─────────┐
   │ Reserve │         │  Charge │         │  Ship   │
   │Inventory│         │ Payment │         │  Order  │
   └─────────┘         └─────────┘         └─────────┘
        │                    │                    │
        └────────────────────┼────────────────────┘
                             │
                    Compensate on failure
```

---

## 4. Core Mechanics

### Event Notification
- Event contains minimal data (e.g., ID, type); consumer fetches full data
- Low payload, but creates coupling (consumer must call producer)
- Good when: Event is trigger only; data changes frequently

### Event-Carried State Transfer
- Event contains all data consumer needs
- Self-contained; no additional calls
- Good when: Consumer needs snapshot; reduce coupling
- Trade-off: Larger payloads; potential staleness if event delayed

### Event Sourcing
- **Store events as source of truth**; current state is derived by replaying
- **Event Store**: Append-only log; events immutable
- **Replay**: Rebuild state by applying events in order
- **Snapshots**: Periodic state snapshots to avoid full replay for old aggregates
- **Projections**: Materialized views built from events for querying

### CQRS (Command Query Responsibility Segregation)
- **Separate read and write models**
- **Commands**: Modify state; go to write model
- **Queries**: Read state; go to read model (can be eventually consistent)
- **Sync**: Projections update read model from write model events
- **Benefit**: Optimize read and write independently (e.g., read from denormalized view)

### Transactional Outbox
- **Problem**: Need to publish event and update DB atomically
- **Solution**: Write event to outbox table in same transaction as business data
- **Background process**: Polls outbox, publishes to message broker, marks published
- **Guarantees**: At-least-once delivery; no lost events

### Event Schema Evolution
- **Versioning**: Add version to event schema (e.g., OrderCreated.v2)
- **Backward compatibility**: Add optional fields; don't remove fields
- **Schema registry**: Confluent Schema Registry, AWS Glue
- **Compatibility modes**: Backward, forward, full

---

## 5. Numbers

| Metric | Typical Range |
|--------|---------------|
| Kafka throughput | 100K-1M+ msg/sec per cluster |
| Event size (recommended) | < 1MB (Kafka default 1MB) |
| Consumer lag (healthy) | < 1000 messages |
| Event retention (Kafka) | 7 days (default) to infinite |
| Projection rebuild time | Minutes to hours (depends on event count) |
| Saga compensation rate | < 1% in well-designed systems |
| Outbox poll interval | 100ms - 1s |

### Uber/Kafka Scale
- 1M+ messages/second
- 100+ topics
- Sub-second end-to-end latency for critical path

---

## 6. Tradeoffs

### Event Notification vs Event-Carried State

| Aspect | Notification | Carried State |
|--------|--------------|---------------|
| Payload size | Small | Larger |
| Coupling | Higher (consumer calls producer) | Lower |
| Staleness | Always fresh (fetch on read) | Possible if event delayed |
| Consumer complexity | Higher (extra call) | Lower |
| Use case | Reference data changes often | Snapshot sufficient |

### Event Sourcing Tradeoffs

| Pro | Con |
|-----|-----|
| Complete audit trail | Storage grows indefinitely |
| Time travel / replay | Replay can be slow |
| Natural event-driven | Querying current state requires projection |
| Debugging (replay to point in time) | Schema evolution complexity |
| Multiple read models from same source | Learning curve |

### Choreography vs Orchestration (Saga)

| Aspect | Choreography | Orchestration |
|--------|--------------|---------------|
| Coupling | Loose | Central coordinator |
| Complexity | Distributed logic | Centralized logic |
| Visibility | Hard to trace | Easy to trace |
| Compensation | Each service knows its compensation | Orchestrator drives compensation |
| Use case | Simple flows, few services | Complex flows, many steps |

---

## 7. Variants / Implementations

### Message Brokers
- **RabbitMQ**: AMQP, flexible routing, good for workflows
- **Apache Kafka**: Log-based, replay, high throughput, retention
- **AWS SQS/SNS**: Managed, at-least-once, no replay
- **Google Pub/Sub**: At-least-once, global

### Event Stores
- **EventStoreDB**: Purpose-built event store
- **Kafka**: Used as event store (topics as streams)
- **DynamoDB/CosmosDB**: With careful key design
- **PostgreSQL**: Outbox + JSONB for events

### CQRS Implementations
- **Axon Framework** (Java): CQRS + Event Sourcing
- **MediatR** (.NET): CQRS pattern
- **Custom**: Event handlers + read model updates

---

## 8. Scaling Strategies

### Producer Scaling
- Partition by key (e.g., orderId) for ordering
- Multiple producers to same topic
- Batch events to reduce overhead

### Consumer Scaling
- **Consumer groups**: Each partition consumed by one consumer in group
- Scale consumers up to number of partitions
- Rebalance on consumer join/leave

### Event Store Scaling
- Kafka: Add partitions; add brokers
- Shard by aggregate ID for event sourcing
- Snapshots reduce replay load

### Projection Scaling
- Multiple projection instances (consumer groups)
- Idempotent projections for safe replay
- Separate read model DBs per projection type

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|-------------|
| Broker down | Events not delivered | Replication, multi-AZ |
| Consumer crash | Unprocessed messages | Consumer group rebalance; reprocess |
| Duplicate events | Double processing | Idempotency keys |
| Out-of-order events | Wrong state | Partition by key; ordering guarantee |
| Schema change | Consumer break | Schema registry; compatibility |
| Projection lag | Stale read | Monitor lag; scale consumers |
| Saga compensation failure | Inconsistent state | Retry; manual intervention; idempotent compensation |

### Idempotency
- Store processed event IDs; skip if already processed
- Use event ID + consumer as idempotency key
- Critical for at-least-once delivery

---

## 10. Performance Considerations

- **Batching**: Produce/consume in batches (Kafka: batch.size)
- **Compression**: Snappy, LZ4, GZIP for large payloads
- **Partitioning**: Choose partition key for even distribution
- **Consumer fetch size**: Tune for throughput vs latency
- **Projection design**: Avoid N+1; batch updates to read model
- **Snapshot frequency**: Balance storage vs replay time

---

## 11. Use Cases

**Event Sourcing:**
- Audit/compliance requirements
- Complex domain with rich history
- Need for time-travel debugging
- Multiple read models from same data

**CQRS:**
- Read and write have different scaling needs
- Complex queries (reporting) vs simple writes
- Different consistency requirements

**Event-Driven (general):**
- Real-time notifications
- Multiple consumers for same event
- Loose coupling between systems
- Async workflows

**Avoid when:**
- Strong consistency required
- Simple CRUD; no audit need
- Team unfamiliar with eventual consistency

---

## 12. Comparison Tables

### Event vs Message vs Command

| Type | Direction | Expectation |
|------|-----------|-------------|
| Event | One-to-many | Something happened; consumers react |
| Message | Point-to-point | Task for specific consumer |
| Command | One-to-one | Instruction to do something |

### Broker Comparison

| Feature | Kafka | RabbitMQ | SQS |
|---------|-------|----------|-----|
| Ordering | Per partition | Per queue | Best-effort |
| Replay | Yes (retention) | No | No |
| Throughput | Very high | High | High |
| Protocol | Custom | AMQP | HTTP |
| Retention | Configurable | Until consumed | 14 days max |

---

## 13. Code or Pseudocode

### Event Sourcing - Append and Replay

```python
# Event Store Interface
class EventStore:
    def append(self, aggregate_id: str, events: List[Event]) -> None:
        """Append events atomically. Events are immutable."""
        for event in events:
            self._write(aggregate_id, event)
    
    def get_events(self, aggregate_id: str, from_version: int = 0) -> List[Event]:
        """Retrieve all events for aggregate, optionally from version."""
        return self._read(aggregate_id, from_version)
    
    def get_snapshot(self, aggregate_id: str) -> Optional[Snapshot]:
        """Get latest snapshot to avoid full replay."""
        return self._get_latest_snapshot(aggregate_id)

# Rebuilding state
def rebuild_aggregate(store: EventStore, aggregate_id: str) -> Order:
    snapshot = store.get_snapshot(aggregate_id)
    if snapshot:
        order = snapshot.state
        from_version = snapshot.version + 1
    else:
        order = Order()
        from_version = 0
    
    for event in store.get_events(aggregate_id, from_version):
        order.apply(event)
    
    return order
```

### Transactional Outbox

```python
# Same transaction: business data + outbox
def place_order(order: Order) -> None:
    with db.transaction():
        # 1. Persist order
        db.orders.insert(order)
        
        # 2. Write to outbox (same transaction)
        db.outbox.insert(OutboxEvent(
            aggregate_id=order.id,
            event_type='OrderPlaced',
            payload=json.dumps(order.to_event_payload()),
            created_at=now()
        ))

# Outbox publisher (separate process)
def publish_outbox():
    events = db.outbox.get_unpublished(limit=100)
    for event in events:
        try:
            kafka.send('orders', event.payload)
            db.outbox.mark_published(event.id)
        except Exception as e:
            log.error(f"Publish failed: {e}")
            # Retry later
```

### Saga Orchestrator

```python
class OrderSagaOrchestrator:
    def execute(self, order: Order) -> SagaResult:
        steps = [
            ('reserve_inventory', inventory_service.reserve, inventory_service.release),
            ('charge_payment', payment_service.charge, payment_service.refund),
            ('create_shipment', shipping_service.create, shipping_service.cancel),
        ]
        completed = []
        
        for step_name, execute_fn, compensate_fn in steps:
            try:
                result = execute_fn(order)
                completed.append((step_name, result, compensate_fn))
            except Exception as e:
                # Compensate in reverse order
                for name, res, comp in reversed(completed):
                    comp(res)
                return SagaResult.failed(e)
        
        return SagaResult.success()
```

---

## 14. Interview Discussion

### Key Points
1. **Event vs Command**: Events are facts (past tense); commands are instructions
2. **Event Sourcing**: When you need full history, audit, or multiple read models
3. **CQRS**: When reads and writes have different patterns
4. **Saga**: For distributed transactions; prefer choreography for simplicity, orchestration for complex flows
5. **Transactional Outbox**: Critical for reliable event publishing
6. **Idempotency**: Essential with at-least-once delivery

### Common Questions
- **"Event Sourcing vs traditional DB?"** → Event sourcing for audit, replay, multiple views; traditional for simple CRUD
- **"How handle schema evolution?"** → Version events; backward-compatible changes; schema registry
- **"Choreography vs Orchestration?"** → Choreography: loose coupling, harder to trace; Orchestration: centralized, easier to debug
- **"How ensure exactly-once?"** → Difficult; usually at-least-once + idempotency

### Red Flags
- Ignoring idempotency
- No discussion of failure/compensation
- Overusing event sourcing for simple domains
- Ignoring consumer lag monitoring
