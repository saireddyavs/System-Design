# API Gateway vs Load Balancer vs Reverse Proxy

> Staff+ Engineer Level — FAANG Interview Deep Dive

---

## 1. Concept Overview

### Definition

| Component | Definition |
|-----------|------------|
| **Reverse Proxy** | Server that sits in front of backend servers, forwarding client requests and optionally modifying them. Clients connect to the proxy; proxy connects to origin. |
| **Load Balancer** | Distributes incoming traffic across multiple backend instances to improve availability and scalability. Can operate at L4 (TCP) or L7 (HTTP). |
| **API Gateway** | Entry point for APIs that handles routing, authentication, rate limiting, protocol translation, and aggregation. API-specific logic. |

### Purpose

- **Reverse Proxy**: Hide origin, SSL termination, caching, compression
- **Load Balancer**: Distribute load, health checks, failover
- **API Gateway**: Auth, rate limit, routing, aggregation, API management

### Overlap

All three can forward requests to backends. They differ in **focus** and **capabilities**:
- **Proxy**: Generic forward + transform
- **LB**: Distribution + high availability
- **API Gateway**: API-specific (auth, versioning, etc.)

---

## 2. Real-World Motivation

### AWS Stack

- **CloudFront**: CDN + reverse proxy (cache, SSL)
- **ALB (Application Load Balancer)**: L7 load balancer (path routing, health checks)
- **API Gateway**: API management (auth, throttling, REST/WebSocket)

### Kong

- **API Gateway** with reverse proxy and load balancing
- Plugins: auth, rate limit, logging
- Can replace nginx for API use cases

### Nginx

- **Reverse proxy** and **load balancer**
- Not full API gateway (no built-in auth, rate limit as module)

### Envoy

- **Proxy** and **load balancer**
- Used by Istio (service mesh); also as edge proxy
- Dynamic config via xDS

### HAProxy

- **Load balancer** (L4, L7)
- High performance; less API-specific features

---

## 3. Architecture Diagrams

### Typical Architecture — All Three

```
                    ┌─────────────────────────────────────────┐
                    │              Internet                    │
                    └────────────────────┬────────────────────┘
                                         │
                                         ▼
                    ┌─────────────────────────────────────────┐
                    │         Reverse Proxy / CDN              │
                    │  (CloudFront, nginx)                    │
                    │  - SSL termination                      │
                    │  - Caching, compression                 │
                    │  - DDoS protection                      │
                    └────────────────────┬────────────────────┘
                                         │
                                         ▼
                    ┌─────────────────────────────────────────┐
                    │           API Gateway                    │
                    │  (AWS API Gateway, Kong)                 │
                    │  - Auth (JWT, API key)                   │
                    │  - Rate limiting                        │
                    │  - Routing by path, version             │
                    │  - Request/response transform           │
                    └────────────────────┬────────────────────┘
                                         │
                                         ▼
                    ┌─────────────────────────────────────────┐
                    │         Load Balancer                    │
                    │  (ALB, nginx, HAProxy)                   │
                    │  - Distribute across instances           │
                    │  - Health checks                         │
                    │  - Sticky sessions (optional)            │
                    └────────────────────┬────────────────────┘
                                         │
              ┌──────────────────────────┼──────────────────────────┐
              │                          │                          │
              ▼                          ▼                          ▼
        ┌───────────┐              ┌───────────┐              ┌───────────┐
        │ Service A │              │ Service A │              │ Service B │
        │ Instance 1│              │ Instance 2│              │ Instance 1│
        └───────────┘              └───────────┘              └───────────┘
```

### Reverse Proxy — Basic Flow

```
    Client                Reverse Proxy              Origin
       │                        │                        │
       │  Request               │                        │
       │───────────────────────>│                        │
       │                        │  Forward (maybe modify)│
       │                        │───────────────────────>│
       │                        │                        │
       │                        │  Response             │
       │                        │<──────────────────────│
       │  Response (maybe       │                        │
       │  cached, compressed)   │                        │
       │<───────────────────────│                        │
```

### Load Balancer — Distribution

```
                    Clients
                        │
                        ▼
                ┌───────────────┐
                │ Load Balancer │
                │ (round-robin, │
                │  least conn)  │
                └───────┬───────┘
                        │
        ┌───────────────┼───────────────┐
        │               │               │
        ▼               ▼               ▼
    ┌───────┐       ┌───────┐       ┌───────┐
    │Backend│       │Backend│       │Backend│
    │  1    │       │  2    │       │  3    │
    └───────┘       └───────┘       └───────┘
```

### API Gateway — Routing + Auth

```
    Client                    API Gateway                 Backends
       │                           │                          │
       │  GET /api/v1/users        │                          │
       │  Authorization: Bearer X  │                          │
       │──────────────────────────>│                          │
       │                           │  Validate token          │
       │                           │  Rate limit check        │
       │                           │  Route: /api/v1/* →      │
       │                           │  User Service            │
       │                           │─────────────────────────>│
       │                           │                          │
       │                           │  Response                │
       │                           │<─────────────────────────│
       │  Response                 │                          │
       │<──────────────────────────│                          │
```

### AWS: ALB vs API Gateway vs CloudFront

```
    User
      │
      ▼
    CloudFront (CDN, cache, SSL)
      │
      ▼
    API Gateway (REST/WebSocket API, auth, throttling)
      │
      ▼
    ALB (path-based routing to target groups)
      │
      ├──> Lambda
      ├──> ECS/Fargate
      └──> EC2
```

---

## 4. Core Mechanics

### Reverse Proxy Responsibilities

| Responsibility | Description |
|----------------|-------------|
| **Forward** | Send request to origin; return response |
| **SSL termination** | Decrypt at proxy; optional re-encrypt to origin |
| **Caching** | Cache responses; serve from cache |
| **Compression** | gzip, brotli |
| **Rewrite** | Modify path, headers |
| **Hide origin** | Origin IP/ports not exposed |

### Load Balancer Responsibilities

| Responsibility | Description |
|----------------|-------------|
| **Distribution** | Round-robin, least connections, IP hash |
| **Health checks** | Mark unhealthy; stop sending traffic |
| **Session affinity** | Sticky sessions (optional) |
| **L4 vs L7** | L4: TCP/UDP; L7: HTTP (path, header) |

### API Gateway Responsibilities

| Responsibility | Description |
|----------------|-------------|
| **Authentication** | JWT, OAuth, API key |
| **Rate limiting** | Per user, per key |
| **Routing** | Path, header, version |
| **Protocol translation** | REST → gRPC |
| **Aggregation** | BFF; combine multiple backends |
| **API management** | Versioning, documentation |

---

## 5. Numbers

| Component | Typical Latency | Throughput |
|-----------|-----------------|------------|
| Reverse proxy (nginx) | <1ms | 10K+ req/s |
| Load balancer (ALB) | 1-2ms | 10K+ req/s |
| API Gateway (AWS) | 5-10ms | 10K req/s (default) |
| Kong | 1-5ms | 10K-50K req/s |

---

## 6. Tradeoffs

### When to Use What

| Need | Use |
|------|-----|
| **Distribute traffic** | Load Balancer |
| **SSL, cache, hide origin** | Reverse Proxy |
| **Auth, rate limit, API routing** | API Gateway |
| **All of the above** | Often combined (proxy → gateway → LB) |

### Overlap and Differences

| Capability | Reverse Proxy | Load Balancer | API Gateway |
|------------|---------------|---------------|-------------|
| **Forward requests** | ✓ | ✓ | ✓ |
| **Multiple backends** | ✓ | ✓ (primary) | ✓ |
| **SSL termination** | ✓ | ✓ (L7) | ✓ |
| **Health checks** | Limited | ✓ (primary) | Limited |
| **Auth** | Basic | No | ✓ (primary) |
| **Rate limiting** | Module | No | ✓ (primary) |
| **Protocol translation** | Limited | No | ✓ |

---

## 7. Variants / Implementations

### Reverse Proxy

- **Nginx**: Config-based; modules for cache, auth
- **Caddy**: Auto HTTPS; simple config
- **CloudFront**: CDN; global edge
- **Traefik**: Dynamic; Kubernetes-native

### Load Balancer

- **HAProxy**: L4/L7; high performance
- **AWS ALB**: L7; path routing; Lambda target
- **AWS NLB**: L4; ultra-low latency
- **Nginx**: `upstream` directive

### API Gateway

- **Kong**: Open-source; plugin-based
- **AWS API Gateway**: Managed; Lambda, HTTP
- **Apigee**: Enterprise; full API management
- **Ambassador**: Kubernetes; Envoy-based

---

## 8. Scaling Strategies

1. **Horizontal**: Add instances; stateless
2. **Caching**: At proxy/gateway to reduce backend load
3. **Connection pooling**: Reuse connections to backends
4. **Edge**: CDN/proxy at edge; gateway in region

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|------------|
| **Proxy/Gateway down** | All traffic blocked | Multiple instances; LB in front |
| **Backend unhealthy** | LB stops routing | Health checks; auto-recovery |
| **Auth service down** | Gateway can't validate | Cache tokens; degrade |
| **Misconfiguration** | Wrong routing | Canary; feature flags |

---

## 10. Performance Considerations

- **Proxy**: Minimal logic; fast
- **LB**: Health check interval; connection reuse
- **API Gateway**: Plugins add latency; minimize

---

## 11. Use Cases

| Use Case | Primary Component |
|----------|-------------------|
| **Static + API** | Reverse proxy (cache static); gateway for API |
| **Microservices** | API Gateway (edge) + LB (per service) |
| **Public API** | API Gateway (auth, rate limit) |
| **Internal only** | LB (no gateway needed) |

---

## 12. Comparison Tables

### Decision Matrix

| Requirement | Reverse Proxy | Load Balancer | API Gateway |
|-------------|---------------|---------------|-------------|
| **Multiple backends** | ✓ | ✓ | ✓ |
| **SSL termination** | ✓ | ✓ (L7) | ✓ |
| **Health-based routing** | Limited | ✓ | Limited |
| **Authentication** | Basic | ✗ | ✓ |
| **Rate limiting** | Module | ✗ | ✓ |
| **API versioning** | Rewrite | ✗ | ✓ |
| **Protocol translation** | Limited | ✗ | ✓ |
| **BFF aggregation** | ✗ | ✗ | ✓ |

### Tool Comparison

| Tool | Primary Role | Also Does |
|------|--------------|-----------|
| **Nginx** | Reverse proxy, LB | SSL, cache, basic auth |
| **HAProxy** | Load balancer | L4, L7, health checks |
| **Kong** | API Gateway | Proxy, LB, plugins |
| **Envoy** | Proxy | LB, observability, xDS |
| **AWS ALB** | Load balancer | Path routing, Lambda |
| **AWS API Gateway** | API Gateway | Auth, throttling |
| **CloudFront** | CDN, reverse proxy | Cache, SSL, DDoS |

### AWS: When to Use Which

| Scenario | Use |
|----------|-----|
| **Public website + API** | CloudFront → API Gateway → ALB |
| **Internal microservices** | ALB only (or NLB) |
| **Serverless API** | API Gateway → Lambda |
| **EC2/ECS behind API** | API Gateway or ALB → EC2/ECS |

---

## 13. Code / Pseudocode

### Nginx — Reverse Proxy + Load Balancer

```nginx
upstream backend {
    least_conn;
    server 10.0.1.1:8080;
    server 10.0.1.2:8080;
    server 10.0.1.3:8080;
}

server {
    listen 443 ssl;
    ssl_certificate /path/to/cert.pem;
    location / {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Kong — API Gateway Config

```yaml
services:
  - name: user-service
    url: http://user-service:8080
    routes:
      - paths: ["/api/v1/users"]
        methods: ["GET", "POST"]
    plugins:
      - name: rate-limiting
        config:
          minute: 100
      - name: jwt
```

### Decision Logic (Pseudocode)

```python
def choose_component(requirements):
    if needs_auth or needs_rate_limit or needs_api_versioning:
        return "API Gateway"
    if needs_distribution_across_instances and health_checks:
        return "Load Balancer"
    if needs_ssl_termination or caching or hide_origin:
        return "Reverse Proxy"
    # Often use all: Proxy (edge) -> Gateway (API) -> LB (backends)
    return "Combined"
```

---

## 14. Interview Discussion

### Key Points

1. **Reverse proxy**: Forward + transform; SSL, cache, hide origin
2. **Load balancer**: Distribute traffic; health checks; high availability
3. **API Gateway**: API-specific; auth, rate limit, routing, aggregation
4. **Overlap**: All forward; differ in focus. Often used together.

### Common Questions

- **"Difference between API Gateway and Load Balancer?"** — LB: distribute traffic, health checks. Gateway: auth, rate limit, API routing, protocol translation.
- **"When would you use an API Gateway?"** — Public API, need auth/rate limit, multiple backends, versioning, BFF aggregation.
- **"Can one tool do all three?"** — Kong does gateway + proxy + LB. Nginx does proxy + LB. AWS uses separate services (CloudFront, API Gateway, ALB).
- **"Typical architecture?"** — CDN/Proxy (edge) → API Gateway (auth, routing) → Load Balancer → Backend instances.
- **"AWS: ALB vs API Gateway?"** — API Gateway: REST/WebSocket API, auth, throttling, Lambda. ALB: path routing to targets (EC2, ECS, Lambda); no built-in API auth.

### Red Flags

- Using API Gateway for simple internal LB (overkill)
- No load balancer in front of gateway (SPOF)
- Putting all logic in gateway (can become bottleneck)
