# API Gateway

## 1. Concept Overview

### Definition
An API Gateway is a server that acts as a single entry point for all client requests to backend services. It sits between clients and microservices, handling cross-cutting concerns that would otherwise be duplicated across services.

### Purpose
- **Single Entry Point**: One place for clients to connect; backend topology hidden
- **Cross-Cutting Concerns**: Centralize auth, rate limiting, logging, routing
- **Protocol Translation**: Expose REST to clients while backends use gRPC
- **Aggregation**: Combine multiple backend calls into one response (BFF pattern)

### Problems It Solves
- **Duplication**: Every service implementing auth, rate limiting, logging
- **Client Complexity**: Clients would need to know and call many service endpoints
- **Security**: Centralized point for authentication, TLS termination
- **Observability**: Single place to log, trace, and monitor traffic

---

## 2. Real-World Motivation

### Netflix
- **Zuul / Zuul 2**: Edge gateway handling 2B+ requests/day. Zuul 2 uses Netty for async, non-blocking I/O. Routes to hundreds of microservices. Handles auth, routing, canary deployments.

### Amazon
- **AWS API Gateway**: Managed service for REST and WebSocket APIs. Integrates with Lambda, HTTP backends. Usage plans, API keys, throttling. Used by millions of APIs.

### Kong
- **Kong Gateway**: Open-source, plugin-based. Used by companies like Expedia, Nasdaq. Plugins for rate limiting, auth (JWT, OAuth2), logging (Datadog, Prometheus).

### Apigee (Google Cloud)
- **Apigee**: Full API management platform. Policy-based configuration. Used for monetization, analytics, developer portals.

### Ambassador / Emissary
- **Kubernetes-native**: Built on Envoy proxy. Used for service mesh edge, gRPC, canary routing.

---

## 3. Architecture Diagrams

### Basic API Gateway Architecture

```
                    ┌─────────────────────────────────────────┐
                    │              API Gateway                 │
                    │  ┌─────────────────────────────────────┐│
                    │  │ Auth │ Rate Limit │ Routing │ Log   ││
                    │  └─────────────────────────────────────┘│
                    └─────────────────────────────────────────┘
                                        │
        ┌───────────────────────────────┼───────────────────────────────┐
        │                               │                               │
        ▼                               ▼                               ▼
┌───────────────┐               ┌───────────────┐               ┌───────────────┐
│ User Service  │               │ Order Service │               │ Product Svc   │
└───────────────┘               └───────────────┘               └───────────────┘
```

### BFF (Backend for Frontend) Pattern

```
┌─────────────┐     ┌─────────────────┐     ┌─────────────────────────────────┐
│ Mobile App  │────▶│ Mobile BFF      │────▶│ User Svc │ Order Svc │ Cart Svc  │
└─────────────┘     │ (mobile-opt)    │     └─────────────────────────────────┘
                    └─────────────────┘
┌─────────────┐     ┌─────────────────┐     ┌─────────────────────────────────┐
│ Web App     │────▶│ Web BFF         │────▶│ User Svc │ Order Svc │ Cart Svc  │
└─────────────┘     │ (web-optimized) │     └─────────────────────────────────┘
                    └─────────────────┘
```

### API Gateway vs Service Mesh

```
WITH API GATEWAY (North-South):          WITH SERVICE MESH (East-West):
┌─────────┐     ┌──────────┐             ┌─────┐     ┌─────┐     ┌─────┐
│ Client  │────▶│ Gateway  │────▶ Svc A │     │     │     │     │     │
└─────────┘     └──────────┘             │ A   │◀───▶│ B   │◀───▶│ C   │
                    │                    │     │     │     │     │     │
                    ▼                    └─────┘     └─────┘     └─────┘
              ┌─────┐ ┌─────┐                   ▲         ▲
              │ A   │ │ B   │                   └─────────┘
              └─────┘ └─────┘                   Sidecar proxies
```

---

## 4. Core Mechanics

### Responsibilities

| Responsibility | Description | Example |
|----------------|-------------|---------|
| **Routing** | Route requests to correct backend based on path, header, method | `/api/users` → User Service |
| **Authentication** | Verify identity (JWT, API key, OAuth) | Validate Bearer token |
| **Authorization** | Check permissions | Role-based access |
| **Rate Limiting** | Throttle requests per client/API key | 100 req/min per user |
| **Request/Response Transform** | Modify headers, body, protocol | REST → gRPC translation |
| **Protocol Translation** | REST, GraphQL, gRPC, WebSocket | Client REST, backend gRPC |
| **Load Balancing** | Distribute across backend instances | Round-robin, least connections |
| **Circuit Breaking** | Stop forwarding when backend fails | Open circuit after 5 failures |
| **Caching** | Cache responses | Cache GET /products for 60s |
| **Logging** | Request/response logging | Structured logs to ELK |

### Gateway Aggregation

Single client request triggers multiple backend calls; gateway aggregates:

```
Client: GET /dashboard
Gateway:
  - GET /user/profile (User Service)
  - GET /user/orders (Order Service)
  - GET /recommendations (Rec Service)
  → Combine into single JSON response
```

---

## 5. Numbers

| Metric | Typical Value |
|--------|---------------|
| Gateway added latency | 1-5ms (optimized), 5-20ms (with plugins) |
| Throughput (Kong/Envoy) | 10K-50K req/s per instance |
| Netflix Zuul 2 | 2B+ requests/day |
| AWS API Gateway | Sub-10ms p50 latency |

---

## 6. Tradeoffs

### API Gateway vs Reverse Proxy

| Aspect | API Gateway | Reverse Proxy (nginx) |
|--------|-------------|------------------------|
| Focus | API-specific (auth, rate limit) | Generic HTTP routing |
| Configuration | Declarative, API-centric | Config file |
| Extensibility | Plugins, policies | Modules |

### API Gateway vs Service Mesh

| Aspect | API Gateway | Service Mesh |
|--------|-------------|--------------|
| Traffic | North-South (client → services) | East-West (service → service) |
| Placement | Edge | Per-pod sidecar |
| Use case | External API, BFF | Internal mTLS, retries, observability |

---

## 7. Variants / Implementations

### Edge Gateway
- Single gateway at network edge
- Handles all external traffic
- Examples: Kong, AWS API Gateway, Apigee

### BFF (Backend for Frontend)
- One BFF per client type (mobile, web, partner)
- Tailored response shape
- Examples: Netflix (different BFFs per device)

### Gateway Aggregation
- Gateway calls multiple backends, aggregates response
- Reduces client round-trips
- Tradeoff: Gateway becomes bottleneck, latency = max(backend latencies)

---

## 8. Scaling Strategies

1. **Horizontal scaling**: Stateless gateway; add instances behind load balancer
2. **Caching**: Cache responses at gateway to reduce backend load
3. **Connection pooling**: Reuse connections to backends
4. **Async I/O**: Non-blocking (Netty, Envoy) for high concurrency

---

## 9. Failure Scenarios

| Failure | Impact | Mitigation |
|---------|--------|------------|
| Gateway overload | All traffic blocked | Auto-scaling, circuit breaker to backends |
| Backend timeout | Client waits | Timeout configuration, fallback responses |
| Auth service down | No requests authenticated | Cache tokens, degrade to optional auth |
| Misconfiguration | Wrong routing | Canary deployments, feature flags |

---

## 10. Performance Considerations

- **Minimize plugins**: Each plugin adds latency
- **Connection pooling**: Reuse backend connections
- **Caching**: Cache GET responses where appropriate
- **Async**: Use async/non-blocking implementations (Zuul 2, Envoy)

---

## 11. Use Cases

| Use Case | Gateway Role |
|----------|--------------|
| Public API | Auth, rate limit, routing, versioning |
| Microservices front | Single entry, protocol translation |
| Mobile backend | BFF aggregates, mobile-optimized payloads |
| Partner integration | API keys, usage tracking, SLAs |

---

## 12. Comparison Tables

### Gateway Implementations

| Gateway | Type | Key Features |
|---------|------|--------------|
| Kong | Open-source/Enterprise | Lua plugins, DB-less mode |
| AWS API Gateway | Managed | Lambda integration, usage plans |
| Zuul 2 | OSS (Netflix) | Netty, async |
| Apigee | Enterprise | Full API management, monetization |
| Ambassador | K8s | Envoy-based, gRPC |

---

## 13. Code or Pseudocode

### Routing Rule (Kong-style)

```yaml
services:
  - name: user-service
    url: http://user-service:8080
    routes:
      - paths: ["/api/v1/users"]
        methods: ["GET", "POST"]
  - name: order-service
    url: http://order-service:8080
    routes:
      - paths: ["/api/v1/orders"]
```

### Rate Limiting Plugin

```yaml
plugins:
  - name: rate-limiting
    config:
      minute: 100
      policy: local  # or redis for distributed
```

### Circuit Breaker (Pseudocode)

```python
class CircuitBreaker:
    def __init__(self, failure_threshold=5, timeout=30):
        self.failures = 0
        self.state = "closed"  # closed, open, half-open
        self.last_failure_time = None
    
    def call(self, fn):
        if self.state == "open":
            if time.now() - self.last_failure_time > self.timeout:
                self.state = "half-open"
            else:
                raise CircuitOpenError()
        
        try:
            result = fn()
            if self.state == "half-open":
                self.state = "closed"
                self.failures = 0
            return result
        except Exception as e:
            self.failures += 1
            self.last_failure_time = time.now()
            if self.failures >= self.failure_threshold:
                self.state = "open"
            raise
```

---

## 14. Interview Discussion

### How to Explain
"An API Gateway is the single entry point for client requests. It centralizes auth, rate limiting, routing, and logging so each microservice doesn't have to implement them. It can also do protocol translation (REST to gRPC) and aggregation (BFF pattern)."

### Follow-up Questions
- "When would you use a BFF vs a single gateway?"
- "How does an API Gateway differ from a service mesh?"
- "How would you scale an API Gateway to 100K req/s?"
- "Design rate limiting at the gateway. What are the challenges with distributed rate limiting?"

---

## Appendix: Deep Dive

### Zuul 2 Architecture (Netflix)

```
Request → Netty Server → Filter Chain → Backend
Filters: Inbound (auth, routing) → Origin (backend call) → Outbound (response)
```

### Envoy Proxy (Used by Ambassador, Istio)

- C++ implementation, high performance
- Dynamic configuration via xDS (Discovery Service)
- Supports HTTP/1.1, HTTP/2, gRPC

### Protocol Translation Example

Client sends REST; Gateway translates to gRPC:

```
POST /api/v1/users
Body: {"name": "Alice", "email": "alice@example.com"}

Gateway: Translates to gRPC CreateUserRequest, calls UserService.CreateUser
Response: Translates User proto to JSON
```

---

## Appendix B: Authentication Flow at Gateway

```
Client ──▶ Gateway ──▶ Auth Service (validate token)
              │
              ├── Valid: Forward to backend with user context
              └── Invalid: 401 Unauthorized
```

**JWT validation**: Gateway can validate JWT signature locally (public key) without calling auth service for each request. Reduces latency.

---

## Appendix C: Request/Response Transformation

| Transformation | Example |
|----------------|---------|
| Header injection | Add `X-User-Id` from JWT to backend |
| Header removal | Strip `Authorization` before forwarding to internal service |
| Body modification | Add `request_id` to JSON body |
| Protocol | REST JSON → gRPC binary |

---

## Appendix D: Multi-Region API Gateway

```
                    ┌─────────────────┐
                    │   DNS (Route53)  │
                    │   Geo routing   │
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
        ▼                    ▼                    ▼
   ┌─────────┐         ┌─────────┐         ┌─────────┐
   │ Gateway │         │ Gateway │         │ Gateway │
   │ us-east │         │ eu-west │         │ ap-south │
   └────┬────┘         └────┬────┘         └────┬────┘
        │                    │                    │
        ▼                    ▼                    ▼
   Regional Services   Regional Services   Regional Services
```

---

## Appendix E: Canary Deployment via Gateway

```
                    ┌─────────────────┐
                    │  API Gateway    │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              │ 90%          │ 10%          │
              ▼              ▼              │
        ┌──────────┐  ┌──────────┐        │
        │ v1 (prod)│  │ v2 (canary)        │
        └──────────┘  └──────────┘        │
```

Gateway routes by header (`X-Canary: true`) or percentage of traffic.

---

## Appendix F: Request Validation at Gateway

Gateway can validate before forwarding:
- **Schema validation**: JSON schema for request body
- **Size limits**: Max body size (e.g., 1MB)
- **Required headers**: Reject if missing
- **Path parameters**: Validate format (e.g., UUID)

Reduces load on backend; fails fast.

---

## Appendix G: Logging and Observability

| Metric | Purpose |
|--------|---------|
| Request count | Traffic volume |
| Latency (p50, p99) | Performance |
| Error rate | Reliability |
| Rate limit hits | Abuse detection |
| Backend latency | Backend health |

Structured logs: `request_id`, `user_id`, `path`, `status`, `latency_ms`, `backend`.

---

## Appendix H: API Gateway Capacity Planning

- **Throughput**: 10K-50K req/s per instance (Kong, Envoy)
- **Latency budget**: 1-5ms for gateway; rest for backend
- **Scaling**: Horizontal; stateless; scale based on CPU or request rate

---

## Appendix I: API Gateway Security Checklist

- [ ] TLS termination (HTTPS only)
- [ ] Authentication (JWT, API key, OAuth)
- [ ] Rate limiting (per user, per IP)
- [ ] Request size limits
- [ ] CORS configuration
- [ ] Input validation (path, query, body)
- [ ] Audit logging (who, what, when)

---

## Appendix J: API Gateway vs Load Balancer

| Aspect | API Gateway | Load Balancer |
|--------|-------------|---------------|
| Layer | L7 (application) | L4 or L7 |
| Logic | Auth, rate limit, routing | Distribute traffic |
| Backend selection | Path, header, rule | Health, round-robin |

---

## Appendix K: Plugin Execution Order

Typical order in Kong/API Gateway:
1. Pre-auth (IP restriction)
2. Authentication
3. Authorization
4. Rate limiting
5. Request transform
6. Proxy (forward to backend)
7. Response transform
8. Logging

---

## Appendix L: API Gateway Deployment Topology

```
                    ┌─────────────────┐
                    │  Load Balancer  │
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
        ▼                    ▼                    ▼
   ┌─────────┐         ┌─────────┐         ┌─────────┐
   │ Gateway │         │ Gateway │         │ Gateway │
   │ Node 1  │         │ Node 2  │         │ Node 3  │
   └─────────┘         └─────────┘         └─────────┘
```
