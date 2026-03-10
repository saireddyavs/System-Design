# Microservices Architecture

## 1. Concept Overview

### Definition
Microservices architecture is an approach to developing software systems as a suite of small, independently deployable services, each running in its own process and communicating via lightweight mechanisms (typically HTTP/REST or messaging). Each service is built around a specific business capability and can be developed, deployed, and scaled independently.

### Purpose
- **Independent deployment**: Deploy services without coordinating with other teams
- **Technology diversity**: Use different tech stacks per service (polyglot persistence, polyglot programming)
- **Fault isolation**: Failure in one service doesn't bring down the entire system
- **Team autonomy**: Small teams own end-to-end responsibility for their services (Conway's Law alignment)
- **Scalability**: Scale individual services based on their specific load patterns

### Problems It Solves
- **Monolith deployment bottleneck**: One change requires full system redeployment
- **Scaling inefficiency**: Must scale entire application for one component's load
- **Organizational coupling**: Large teams create coordination overhead
- **Technology lock-in**: Entire system tied to one language/framework
- **Release cycle friction**: Long release cycles due to integration testing of everything

---

## 2. Real-World Motivation

### Netflix
- **Journey**: Started as DVD rental monolith; decomposed to 700+ microservices for streaming
- **Scale**: 200M+ subscribers, 15% of global internet traffic during peak
- **Key decisions**: Chaos Engineering (Chaos Monkey), Zuul API gateway, Eureka service discovery, Hystrix circuit breakers
- **Decomposition**: By domain—recommendations, playback, billing, content delivery, user preferences

### Amazon
- **2-Pizza Team Rule**: Teams small enough to be fed by 2 pizzas (~6-10 people)
- **Service-oriented from 2002**: Mandate that all teams expose data via service interfaces
- **Result**: Thousands of services; each team owns their service end-to-end
- **API-first**: Internal services communicate via well-defined APIs

### Uber
- **Domain-oriented decomposition**: Rider, Driver, Trip, Billing, Maps, Notifications
- **Scale**: 100M+ monthly active users, millions of trips daily
- **Event-driven**: Kafka for real-time event streaming between services
- **Geographic distribution**: Services deployed per region for latency

### Google
- **Borg/Kubernetes**: Container orchestration enabling microservice deployment at scale
- **gRPC**: High-performance RPC for internal service communication
- **Protocol Buffers**: Schema-first API contracts

---

## 3. Architecture Diagrams

### Monolith vs Microservices

```
MONOLITHIC ARCHITECTURE
┌─────────────────────────────────────────────────────────────────┐
│                     SINGLE DEPLOYMENT UNIT                        │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐    │
│  │  Auth   │ │  User   │ │  Order  │ │Payment  │ │Inventory│    │
│  │ Module  │ │ Module  │ │ Module  │ │ Module  │ │ Module  │    │
│  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘    │
│       │           │           │           │           │         │
│       └───────────┴───────────┴───────────┴───────────┘         │
│                              │                                   │
│                    ┌─────────▼─────────┐                         │
│                    │  SHARED DATABASE  │                         │
│                    └──────────────────┘                         │
└─────────────────────────────────────────────────────────────────┘
```

```
MICROSERVICES ARCHITECTURE
┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
│  Auth    │  │  User    │  │  Order   │  │ Payment  │  │Inventory │
│ Service  │  │ Service  │  │ Service  │  │ Service  │  │ Service  │
└────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘
     │             │             │             │             │
     │    HTTP/gRPC/Messaging (Service Mesh)   │             │
     │             │             │             │             │
┌────▼─────┐  ┌────▼─────┐  ┌────▼─────┐  ┌────▼─────┐  ┌────▼─────┐
│ Auth DB  │  │ User DB  │  │ Order DB │  │Payment DB│  │Inventory  │
│          │  │          │  │          │  │          │  │   DB     │
└──────────┘  └──────────┘  └──────────┘  └──────────┘  └──────────┘
     Each service owns its data - Database per service
```

### API Composition (BFF Pattern)

```
                    ┌─────────────────┐
                    │   API Gateway   │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
     ┌────────▼──────┐ ┌─────▼─────┐ ┌─────▼─────┐
     │ Mobile BFF    │ │ Web BFF   │ │ Admin BFF │
     └────────┬──────┘ └─────┬─────┘ └─────┬─────┘
              │              │              │
     ┌────────┴──────────────┴──────────────┴────────┐
     │         Backend Microservices                  │
     │  [User] [Order] [Product] [Recommendation]    │
     └──────────────────────────────────────────────┘
```

---

## 4. Core Mechanics

### Decomposition Strategies

**By Business Domain (Domain-Driven Design)**
- Identify bounded contexts from DDD
- Each bounded context becomes a candidate microservice
- Example: Order context → Order Service, Payment context → Payment Service

**Strangler Fig Pattern**
- Gradually replace monolith by building new services around it
- Route new features to new services; migrate functionality incrementally
- Monolith shrinks over time until deprecated

**Decomposition by Use Case**
- Split by distinct user journeys (e.g., checkout vs. browse)
- Risk: May create services that share too much data

**Decomposition by Volatility**
- Services that change frequently vs. stable core
- Isolate high-change areas (e.g., pricing rules) from stable (e.g., product catalog)

### Inter-Service Communication

**Synchronous**
- **REST**: HTTP/JSON, simple, widely supported, higher latency
- **gRPC**: HTTP/2, Protocol Buffers, streaming, lower latency, binary
- **GraphQL**: Client-specified queries, reduces over-fetching

**Asynchronous**
- **Message Queues**: RabbitMQ, SQS—point-to-point, at-least-once
- **Event Streaming**: Kafka, Kinesis—pub/sub, replay, high throughput
- **Event-Driven**: Loose coupling, eventual consistency

### Data Ownership
- **Database per service**: Each service owns its data; no direct DB access from other services
- **API as contract**: Data accessed only via service APIs
- **Eventual consistency**: Acceptable for cross-service data; use events/Saga for coordination

### Saga Pattern (Distributed Transactions)
- **Choreography**: Each service listens for events and publishes its own; no central coordinator
- **Orchestration**: Central saga orchestrator sends commands to each service
- **Compensation**: Each step has compensating action for rollback

---

## 5. Numbers

| Metric | Monolith | Microservices |
|--------|----------|---------------|
| Deployment frequency | Weekly/Monthly | Multiple times/day |
| Team size per service | 50-200+ | 2-10 (2-pizza) |
| Service count (large org) | 1 | 100-1000+ |
| P99 latency overhead | N/A | +5-50ms (network hops) |
| Operational complexity | Low | High (observability, deployment) |
| Cold start (containers) | N/A | 1-30 seconds |
| Network calls per request | 0 | 5-50+ |

### Netflix Scale (Reference)
- 700+ microservices
- 100+ deployments per day
- 2.5+ billion API requests/day
- 500+ TB data through recommendation engine daily

---

## 6. Tradeoffs

### Monolith vs SOA vs Microservices

| Aspect | Monolith | SOA | Microservices |
|--------|----------|-----|---------------|
| Granularity | Coarse | Medium | Fine |
| Communication | In-process | ESB-centric | Direct/lightweight |
| Data | Shared DB | Shared or federated | DB per service |
| Deployment | Single unit | Service units | Independent services |
| Governance | Centralized | ESB governance | Decentralized |
| Technology | Single stack | Often single | Polyglot |
| Team structure | Large teams | Medium | Small autonomous |

### Sync vs Async Communication

| Factor | Synchronous | Asynchronous |
|--------|-------------|--------------|
| Coupling | Tight (requestor waits) | Loose |
| Consistency | Strong (immediate) | Eventual |
| Failure handling | Timeout, retry | Retry, dead letter |
| Complexity | Lower | Higher (ordering, idempotency) |
| Latency | Additive | Non-blocking |
| Use case | Need immediate response | Fire-and-forget, events |

---

## 7. Variants / Implementations

### Microservice Chassis
Pre-built framework with cross-cutting concerns:
- **Spring Cloud**: Config, Discovery, Circuit Breaker, Gateway
- **Micronaut**: Low memory, fast startup
- **Quarkus**: Native compilation, container-optimized

### Shared Libraries vs Service Mesh
- **Shared libraries**: Common code (logging, metrics) in libraries; requires redeployment to update
- **Service mesh**: Sidecar proxies handle cross-cutting; no code changes for updates
- **Recommendation**: Use mesh for infrastructure concerns; libraries for business logic

### Database Patterns
- **Database per service**: Strict isolation, schema independence
- **Shared database (anti-pattern)**: Coupling, defeats purpose
- **Schema per service (same DB)**: Compromise; still some coupling

---

## 8. Scaling Strategies

### Horizontal Scaling
- Scale service instances based on load
- Stateless services scale linearly
- Use load balancer + service discovery

### Independent Scaling
- Scale Order Service 10x during Black Friday; User Service 1x
- Kubernetes HPA, custom metrics (requests/sec, queue depth)

### Caching
- Per-service caches (Redis, Memcached)
- CDN for static/read-heavy data
- Cache-aside, write-through patterns

### Database Scaling
- Read replicas for read-heavy services
- Sharding per service (already isolated)
- CQRS for read/write scaling separation

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|------------|
| Single service down | Partial functionality | Circuit breaker, fallback, graceful degradation |
| Cascading failure | System-wide outage | Circuit breaker, bulkhead, timeouts |
| Network partition | Split-brain | Retry with backoff, idempotency |
| Data inconsistency | Wrong state | Saga compensation, eventual consistency |
| Deployment failure | Rollback needed | Blue-green, canary, feature flags |
| Dependency outage | Dependent services fail | Circuit breaker, cache, fallback |

### Netflix Chaos Engineering
- Chaos Monkey: Randomly terminates production instances
- Latency Monkey: Introduces artificial delay
- Consequence Monkey: Shuts down entire availability zone
- **Goal**: Ensure system survives failures; validate resilience

---

## 10. Performance Considerations

- **Network latency**: Each hop adds 1-10ms; minimize service calls (BFF, batch APIs)
- **Serialization**: gRPC/Protobuf faster than JSON
- **Connection pooling**: Reuse connections; avoid connection per request
- **Timeouts**: Set at every layer; prevent hung requests
- **Async where possible**: Use messaging for non-critical path
- **Caching**: Cache at API gateway, BFF, service level
- **Database**: Connection pooling, connection limits per service

---

## 11. Use Cases

**Good fit:**
- Large organizations with multiple teams
- High deployment frequency requirements
- Different scaling needs per component
- Polyglot technology requirements
- Independent release cycles

**Poor fit:**
- Small team (<10 developers)
- Simple CRUD application
- Strong transactional consistency requirements
- Limited DevOps maturity
- Tight budget (operational overhead)

---

## 12. Comparison Tables

### When to Use Each Communication Style

| Scenario | Recommendation |
|----------|----------------|
| Need immediate response | Sync (REST/gRPC) |
| Fire-and-forget notification | Async (queue/event) |
| High throughput, replay needed | Kafka/event stream |
| Request-reply with retry | Sync with circuit breaker |
| Cross-service workflow | Saga (orchestration or choreography) |

### Service Granularity Heuristics

| Too coarse | Just right | Too fine |
|------------|------------|----------|
| Monolith | Single business capability | Single DB table as service |
| Multiple bounded contexts | One bounded context | Chatty services |
| 50+ developers on one service | 2-pizza team | Network overhead dominates |

---

## 13. Code or Pseudocode

### Service Communication (gRPC Client)

```python
# Order Service calling Payment Service
import grpc
from payment_service_pb2 import ChargeRequest
from payment_service_pb2_grpc import PaymentServiceStub

def charge_customer(order_id: str, amount: float, token: str) -> bool:
    with grpc.insecure_channel('payment-service:50051') as channel:
        stub = PaymentServiceStub(channel)
        try:
            response = stub.Charge(
                ChargeRequest(order_id=order_id, amount=amount, token=token),
                timeout=5.0
            )
            return response.success
        except grpc.RpcError as e:
            if e.code() == grpc.StatusCode.UNAVAILABLE:
                # Circuit breaker would trip here
                raise ServiceUnavailableError("Payment service down")
            raise
```

### Saga Orchestrator (Pseudocode)

```python
class OrderSagaOrchestrator:
    def execute(self, order: Order):
        try:
            # Step 1: Reserve inventory
            inventory_result = inventory_service.reserve(order.items)
            if not inventory_result.success:
                return self.compensate([])
            
            # Step 2: Charge payment
            payment_result = payment_service.charge(order.total, order.customer_id)
            if not payment_result.success:
                return self.compensate([('inventory', inventory_result.reservation_id)])
            
            # Step 3: Create shipment
            shipment_result = shipping_service.create(order, order.shipping_address)
            if not shipment_result.success:
                return self.compensate([
                    ('inventory', inventory_result.reservation_id),
                    ('payment', payment_result.transaction_id)
                ])
            
            return Success(order.id)
        except Exception as e:
            return self.compensate(all_completed_steps)
    
    def compensate(self, steps: List[Tuple[str, str]]):
        for step_type, step_id in reversed(steps):
            if step_type == 'inventory':
                inventory_service.release(step_id)
            elif step_type == 'payment':
                payment_service.refund(step_id)
```

---

## 14. Interview Discussion

### Key Points to Articulate
1. **Start with monolith** unless you have clear scaling/team needs—microservices add complexity
2. **Decomposition**: DDD bounded contexts are the best starting point
3. **Data**: Database per service is non-negotiable for true microservices
4. **Communication**: Prefer async for cross-cutting; sync for request-reply
5. **Failure**: Design for failure—circuit breakers, timeouts, fallbacks
6. **Observability**: Distributed tracing, centralized logging, metrics are essential

### Common Interview Questions
- **"How would you decompose a monolith?"** → Strangler fig, identify bounded contexts, prioritize by volatility
- **"How do you handle distributed transactions?"** → Saga pattern, eventual consistency, avoid 2PC when possible
- **"What's the biggest challenge with microservices?"** → Operational complexity, distributed debugging, eventual consistency
- **"When would you NOT use microservices?"** → Small team, simple app, need for strong consistency, limited DevOps

### Red Flags to Avoid
- Suggesting shared database
- Ignoring failure scenarios
- Over-granular services (nanoservices)
- No mention of observability or deployment strategy

---

## Appendix: Additional Deep Dives

### API Gateway Responsibilities
- **Routing**: Route requests to appropriate backend services
- **Authentication**: JWT validation, API keys
- **Rate limiting**: Protect backends from overload
- **Transformation**: Request/response modification
- **Caching**: Cache responses for read-heavy endpoints
- **Examples**: Kong, AWS API Gateway, Zuul, Nginx

### Database per Service - Deep Dive
- **Schema ownership**: Each service owns its schema; no shared tables
- **Data duplication**: Acceptable for read models; use events to sync
- **Transactions**: Avoid distributed transactions; use Saga for cross-service
- **Migration**: Schema changes are service-internal; no coordination with other services

### Strangler Fig Pattern - Steps
1. Identify a bounded context to extract
2. Create new service with API
3. Add routing layer (gateway) that can route to monolith or new service
4. Migrate functionality incrementally; switch routing
5. Decommission monolith code when fully migrated

### Service Decomposition Anti-Patterns
- **Shared database**: Defeats independence; creates coupling
- **Distributed monolith**: Services deployed separately but tightly coupled
- **Nanoservices**: Too fine-grained; network overhead dominates
- **God service**: One service does too much; extract subdomains

### Communication Protocol Selection Guide
| Scenario | Protocol | Rationale |
|----------|----------|------------|
| Public API | REST/JSON | Wide compatibility, easy debugging |
| Internal high-performance | gRPC | Binary, streaming, low latency |
| Event notification | Kafka/SQS | Async, replay, high throughput |
| Request-reply with retry | gRPC + circuit breaker | Reliable, fast |
