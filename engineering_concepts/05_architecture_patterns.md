# Module 5: Architecture Patterns

---

## 1. Idempotency

### Definition
An operation is idempotent if performing it multiple times produces the same result as performing it once. `f(f(x)) = f(x)`.

### Problem It Solves
Network retries, duplicate messages, and user double-clicks can cause double-charges, duplicate orders, or corrupted state.

### Implementation: Idempotency Keys
```
Client generates UUID: "key-abc-123"

Request 1: POST /charge {amount: $50, idempotency_key: "key-abc-123"}
  Server: Process charge, store result keyed by "key-abc-123"
  Response: 200 OK {charge_id: "ch_1"}

Request 2 (retry): POST /charge {amount: $50, idempotency_key: "key-abc-123"}
  Server: "key-abc-123" already exists → return cached result
  Response: 200 OK {charge_id: "ch_1"}  ← same result, no double charge
```

### Naturally Idempotent Operations
```
✓ SET x = 5          (always same result)
✓ DELETE WHERE id=1   (deleting twice = same)
✓ PUT /resource/1     (replace entire resource)

✗ x = x + 1          (increment is NOT idempotent)
✗ POST /orders        (creates new order each time)
✗ INSERT INTO ...     (duplicate row)
```

### Real Systems
Stripe (payment idempotency keys), AWS (client tokens), Kafka (idempotent producer)

### Interview Tip
"For any non-idempotent write operation, I'd require a client-generated idempotency key stored server-side with the result, enabling safe retries."

### Summary
Idempotency ensures repeated operations have the same effect as one. Implement via server-side deduplication using client-provided unique keys.

---

## 2. Event Sourcing

### Definition
Instead of storing current state, store the sequence of events that produced that state. The current state is derived by replaying events.

### Traditional vs Event Sourcing
```
Traditional DB:           Event Store:
┌──────────────────┐      ┌────────────────────────────┐
│ account_id: 1    │      │ 1. AccountOpened(id=1)     │
│ balance: $140    │      │ 2. Deposited($100)         │
│ name: "Alice"    │      │ 3. Deposited($50)          │
└──────────────────┘      │ 4. Withdrawn($10)          │
                          └────────────────────────────┘
  "What is the balance?"    Replay: 0 + 100 + 50 - 10 = $140
```

### Benefits
```
1. Complete audit trail (every change recorded)
2. Time travel debugging (replay to any point)
3. Event replay for new projections
4. Natural fit with CQRS
```

### Challenges
- **Event schema evolution**: Old events must remain readable
- **Replay performance**: Millions of events → slow (use snapshots)
- **Storage growth**: Events accumulate forever

### Snapshot Optimization
```
Events: [1..1,000,000]
Snapshot at event 999,000: {balance: $5,000}
Current state = snapshot + replay events 999,001..1,000,000
```

### Real Systems
LMAX Exchange, banking/ledger systems, Axon Framework, EventStoreDB

### Summary
Event sourcing stores state changes as an immutable log of events. Replay events to reconstruct state. Provides perfect auditability and temporal debugging at the cost of complexity.

---

## 3. CQRS (Command Query Responsibility Segregation)

### Definition
Separate the write model (commands) from the read model (queries), allowing each to be independently optimized and scaled.

### How It Works
```
┌─────────┐   Command (Write)   ┌──────────────┐
│  Client  │ ──────────────────→│  Write Model  │
│          │                    │  (Normalized   │
│          │                    │   SQL/Events)  │
│          │                    └──────┬─────────┘
│          │                           │ Async event
│          │                           ▼
│          │   Query (Read)     ┌──────────────┐
│          │ ←─────────────────│  Read Model   │
└─────────┘                    │  (Denormalized │
                               │   Elasticsearch│
                               │   /Redis)      │
                               └───────────────┘
```

### Why Separate?
```
Write model: Normalized, enforces business rules, ACID
Read model:  Denormalized, pre-computed joins, fast queries

Example:
  Write: INSERT INTO orders (user_id, item_id, qty)
  Read:  Pre-joined view: {user_name, item_name, qty, total, status}
         Already computed, no JOINs needed at query time
```

### Tradeoffs

| Pros | Cons |
|------|------|
| Scale reads and writes independently | Eventual consistency between models |
| Optimize each model for its purpose | Increased complexity |
| Natural fit with Event Sourcing | More infrastructure (sync mechanism) |

### Real Systems
Uber (separate read/write stores), Twitter, most systems using Elasticsearch for search + RDBMS for writes

### Summary
CQRS splits read and write into separate models. Writes go to a normalized store; reads come from a denormalized, pre-computed store. They sync asynchronously.

---

## 4. Saga Pattern

### Definition
A sequence of local transactions where each step publishes an event triggering the next step. If any step fails, compensating transactions undo previous steps.

### Problem It Solves
Distributed transactions (2PC) are slow and blocking. Sagas achieve eventual consistency without distributed locks.

### Types
```
┌─── CHOREOGRAPHY ────────────────┐   ┌─── ORCHESTRATION ───────────────┐
│ Each service listens for events  │   │ Central orchestrator directs     │
│ and triggers the next step       │   │ each step                        │
│                                  │   │                                  │
│ Order → Payment → Inventory      │   │ Orchestrator:                    │
│   ↓ event   ↓ event   ↓ event   │   │   1. Call Order Service           │
│                                  │   │   2. Call Payment Service         │
│ Pros: Loose coupling             │   │   3. Call Inventory Service       │
│ Cons: Hard to track flow         │   │                                  │
└──────────────────────────────────┘   │ Pros: Clear control flow         │
                                       │ Cons: Orchestrator = SPOF        │
                                       └──────────────────────────────────┘
```

### Compensating Transactions
```
Happy Path:
  1. CreateOrder ✓ → 2. ChargePayment ✓ → 3. ReserveInventory ✓

Failure at step 3:
  3. ReserveInventory ✗
  2. RefundPayment (compensate)
  1. CancelOrder (compensate)
```

### Visual
```
  CreateOrder ──→ ChargePayment ──→ ReserveInventory ──→ ShipOrder
       │               │                  │                  │
       ▼               ▼                  ▼                  ▼
  CancelOrder ←── RefundPayment ←── ReleaseInventory   (compensations)
```

### Real Systems
Uber (ride lifecycle), e-commerce checkout, Temporal (workflow engine), AWS Step Functions

### Summary
Sagas coordinate distributed transactions through a sequence of local transactions with compensating actions for rollback. Use choreography for loose coupling or orchestration for clear control flow.

---

## 5. Bulkhead Pattern

### Definition
Isolating system components into separate pools so that failure in one doesn't cascade to others — like watertight compartments in a ship.

### Problem It Solves
If the image-processing feature exhausts all DB connections, the login feature can't get connections either. Total outage from a non-critical feature.

### Implementation
```
┌──────────── BEFORE ────────────┐
│                                │
│  All features share 1 pool     │
│  ┌────────────────────────┐    │
│  │   Connection Pool (100) │   │
│  │  Images ▓▓▓▓▓▓▓▓▓▓     │   │
│  │  Login  ░░░ (starved!)  │   │
│  │  Search ░░░             │   │
│  └────────────────────────┘    │
└────────────────────────────────┘

┌──────────── AFTER (Bulkhead) ──┐
│                                │
│  ┌──────────┐ ┌──────────┐    │
│  │ Images   │ │ Login    │    │
│  │ Pool: 40 │ │ Pool: 40 │    │
│  │ ▓▓▓▓▓▓▓▓ │ │ ░░░░░░░░ │   │
│  └──────────┘ └──────────┘    │
│  ┌──────────┐                  │
│  │ Search   │  Images flood    │
│  │ Pool: 20 │  → Login SAFE   │
│  │ ░░░░░░░░ │                  │
│  └──────────┘                  │
└────────────────────────────────┘
```

### Types of Bulkheads
- **Thread pool isolation**: Separate thread pools per dependency
- **Connection pool isolation**: Separate DB connection pools
- **Process isolation**: Separate containers/pods
- **Semaphore isolation**: Limit concurrent requests per dependency

### Real Systems
Netflix Hystrix (thread pool isolation), Kubernetes (resource limits per pod), Amazon (cell-based architecture)

### Summary
Bulkheads isolate failures by giving each component its own resource pool. A flood in one pool cannot starve others.

---

## 6. Circuit Breaker

### Definition
A pattern that detects repeated failures to a downstream service and "opens" the circuit to fail fast, preventing resource exhaustion and cascading failures.

### States
```
┌──────────┐  failures > threshold  ┌────────┐
│  CLOSED  │ ─────────────────────→ │  OPEN  │
│(normal)  │                        │(reject │
└──────────┘                        │ all)   │
     ▲                              └───┬────┘
     │  success                         │ timeout
     │                                  ▼
     │                          ┌──────────────┐
     └──────────────────────────│  HALF-OPEN   │
        success on test request │(allow 1 test)│
                                └──────────────┘
```

### How It Works
```
1. CLOSED: All requests pass through. Track failure rate.
2. If failure rate > 50% in last 10 calls → switch to OPEN
3. OPEN: Immediately return error/fallback. No calls to downstream.
4. After 30s timeout → switch to HALF-OPEN
5. HALF-OPEN: Allow 1 test request
   - Success → CLOSED (service recovered)
   - Failure → OPEN (still broken)
```

### Example
```
Service A → Service B (database service)

Service B goes down:
  Without CB: A's threads all wait for B's timeout → A goes down too
  With CB:    After 5 failures, circuit opens → A returns cached data instantly

A stays alive. Users see slightly stale data instead of errors.
```

### Real Systems
Netflix Hystrix, Resilience4j, Polly (.NET), Envoy proxy

### Summary
Circuit breakers detect failing downstream services and stop calling them, returning fallback responses. This prevents thread exhaustion and cascading failures.

---

## 7. Backpressure

### Definition
A feedback mechanism where a slow consumer signals a fast producer to slow down, preventing buffer overflow and OOM crashes.

### Problem It Solves
Producer generates 50K events/sec, consumer processes 10K/sec. Without backpressure, the buffer between them grows until OOM.

### Strategies
```
┌─── DROP ──────────────┐  Drop newest/oldest messages
│ Simple, lossy          │  (acceptable for metrics)
└────────────────────────┘

┌─── BLOCK ─────────────┐  Producer blocks until consumer catches up
│ Bounded buffer         │  (Java BlockingQueue)
└────────────────────────┘

┌─── REACTIVE ──────────┐  Consumer requests N items at a time
│ consumer.request(100)  │  Producer only sends 100
│ (Reactive Streams)     │  (Netflix RxJava, Project Reactor)
└────────────────────────┘

┌─── TCP WINDOW ────────┐  TCP flow control (kernel-level)
│ Receiver shrinks       │  Sender automatically slows
│ window size            │
└────────────────────────┘
```

### Visual
```
Without backpressure:
  Producer ═══50K/s═══→ [Buffer grows...GROWS...OOM] → Consumer (10K/s)

With backpressure:
  Producer ───10K/s───→ [Buffer stable] → Consumer (10K/s)
       ↑                                        │
       └──── "slow down!" ─────────────────────┘
```

### Real Systems
Kafka (consumer lag), TCP, Reactive Streams (RxJava), Akka Streams, Twitter Heron

### Summary
Backpressure lets slow consumers signal fast producers to reduce speed. Without it, unbounded buffers lead to OOM. Implement via blocking, dropping, or reactive pull-based flow control.

---

## 8. Strangler Fig Pattern

### Definition
Incrementally migrating a monolith to microservices by routing new functionality to new services while slowly replacing old routes.

### How It Works
```
Phase 1: Proxy in front of monolith
  ┌────────┐     ┌───────────┐
  │ Client │ ──→ │   Proxy   │ ──→ Monolith (handles everything)
  └────────┘     └───────────┘

Phase 2: New feature as microservice
  ┌────────┐     ┌───────────┐ ──/api/v2/users──→ New User Service
  │ Client │ ──→ │   Proxy   │
  └────────┘     └───────────┘ ──/everything-else──→ Monolith

Phase 3: Gradually migrate all routes
  ┌────────┐     ┌───────────┐ ──→ User Service
  │ Client │ ──→ │   Proxy   │ ──→ Order Service
  └────────┘     └───────────┘ ──→ Payment Service
                                   Monolith (shrinking → dead)
```

### Key Principles
- Never "Big Bang" rewrite (always fails)
- Proxy enables routing at the URL level
- Rollback = route back to monolith
- Monolith naturally shrinks and dies

### Real Systems
Shopify, Best Buy, gov.uk, Amazon (started monolith → microservices)

### Summary
The Strangler Fig pattern incrementally replaces a monolith by routing traffic through a proxy to new microservices, one route at a time. Low risk, reversible.

---

## 9. Sidecar Pattern / Service Mesh

### Sidecar Pattern
```
┌────────────────────────────────┐
│          Pod / VM              │
│  ┌──────────┐  ┌───────────┐  │
│  │   App    │──│  Sidecar   │  │
│  │ (Go/Java)│  │  (Envoy)   │  │
│  │          │  │            │  │
│  │ Speaks   │  │ Handles:   │  │
│  │ localhost│  │ - mTLS     │  │
│  │ :8080    │  │ - Retries  │  │
│  └──────────┘  │ - Metrics  │  │
│                │ - Tracing  │  │
│                │ - Circuit  │  │
│                │   Breaker  │  │
│                └───────────┘  │
└────────────────────────────────┘
```

### Service Mesh
A fleet of sidecars controlled by a central control plane:
```
┌─── Data Plane (sidecars) ───┐    ┌─── Control Plane ───┐
│                              │    │                     │
│  [App A]↔[Envoy]            │    │   Istio / Linkerd   │
│  [App B]↔[Envoy]            │←──→│   - Config          │
│  [App C]↔[Envoy]            │    │   - Cert management │
│                              │    │   - Routing rules   │
└──────────────────────────────┘    └─────────────────────┘
```

### Real Systems
Envoy (Lyft), Istio (Google), Linkerd, Consul Connect

### Summary
Sidecar proxies handle cross-cutting networking concerns (TLS, retries, metrics) outside the application. A service mesh is a fleet of sidecars with centralized control.

---

## 10. API Gateway

### Definition
A single entry point for all client requests that handles routing, authentication, rate limiting, and protocol translation.

### Architecture
```
┌──────────┐     ┌──────────────┐     ┌──────────────┐
│  Mobile  │────→│              │────→│ User Service  │
│  Web     │────→│  API Gateway │────→│ Order Service │
│  Partner │────→│              │────→│ Auth Service  │
└──────────┘     │  - Auth      │     └──────────────┘
                 │  - Rate limit│
                 │  - Route     │
                 │  - Transform │
                 │  - Cache     │
                 └──────────────┘
```

### Responsibilities
- **Routing**: `/users/*` → User Service, `/orders/*` → Order Service
- **Authentication**: Validate JWT/API keys at the edge
- **Rate limiting**: Protect backends from abuse
- **Protocol translation**: REST→gRPC, HTTP→WebSocket
- **Response aggregation**: Combine multiple backend calls into one response

### Real Systems
Kong, AWS API Gateway, Netflix Zuul, NGINX, Envoy

### Summary
An API gateway is the single entry point for all client traffic, handling auth, routing, rate limiting, and protocol translation before forwarding to backend services.

---

## 11. Thundering Herd Problem

### Definition
When a popular cache key expires, thousands of requests simultaneously hit the database to regenerate it, causing a massive load spike.

### Visual
```
Normal:
  Request → Cache HIT → return cached value

Cache expires:
  1000 requests ──→ Cache MISS ──→ ALL hit DB simultaneously
                                    DB: 5% CPU → 100% CPU → crash

Solution (Lease/Mutex):
  1000 requests ──→ Cache MISS
  Request #1: Gets "lease" → queries DB → sets cache
  Request #2-1000: See lease → wait → get cached value
```

### Solutions
```
1. MUTEX/LEASE: Only 1 request regenerates cache
2. PROBABILISTIC EARLY EXPIRY: Randomly refresh before TTL
   expires_at = ttl - random(0, delta)
3. STALE-WHILE-REVALIDATE: Serve stale data while refreshing
4. PRE-WARMING: Refresh cache before it expires (background job)
```

### Real Systems
Facebook (Memcache lease), Slack, any high-traffic application

### Summary
The thundering herd occurs when many requests simultaneously try to regenerate an expired cache key. Solve with mutex/lease (one writer), probabilistic early expiry, or stale-while-revalidate.

---

## 12. Distributed Tracing

### Definition
Tracking a single user request as it propagates through dozens of microservices, recording timing at each hop.

### How It Works
```
User Request (Trace ID: abc-123)
  │
  ├─→ API Gateway        [Span 1: 2ms]
  │    │
  │    ├─→ Auth Service   [Span 2: 5ms]
  │    │
  │    ├─→ User Service   [Span 3: 50ms]  ← bottleneck!
  │    │    │
  │    │    └─→ Database  [Span 4: 45ms]  ← root cause
  │    │
  │    └─→ Cache          [Span 5: 1ms]
  │
  Total: 58ms
```

### Implementation
```
1. Generate Trace-ID at entry point (API Gateway)
2. Pass Trace-ID via HTTP header (e.g., X-Request-ID)
3. Each service creates a "Span" (service name, start, duration, parent)
4. Spans sent to collector (async, out-of-band)
5. Collector assembles spans into trace visualization
```

### Sampling
At scale, tracing 100% of requests is too expensive. Sample:
- 1% random sampling
- 100% for errors
- Head-based vs tail-based sampling

### Real Systems
Uber (Jaeger), Twitter (Zipkin), Google (Dapper paper), AWS X-Ray, Datadog APM, OpenTelemetry

### Summary
Distributed tracing follows requests across microservices using Trace IDs and Spans. It enables pinpointing latency bottlenecks and understanding system behavior.
