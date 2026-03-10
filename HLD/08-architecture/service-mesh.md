# Service Mesh

## 1. Concept Overview

### Definition
A service mesh is a dedicated infrastructure layer for handling service-to-service communication in a microservices architecture. It provides a uniform way to control traffic, enforce policies, and observe interactions between services without requiring application code changes. The mesh typically consists of a **data plane** (proxies that intercept traffic) and a **control plane** (manages configuration and policy).

### Purpose
- **Traffic management**: Load balancing, retries, timeouts, circuit breaking
- **Security**: mTLS (mutual TLS), encryption in transit, access control
- **Observability**: Distributed tracing, metrics, access logs
- **Resilience**: Automatic retries, circuit breaking, fault injection
- **Operational simplicity**: Centralize cross-cutting concerns; no code changes

### Problems It Solves
- **Repeated boilerplate**: Every service had to implement retry, timeout, circuit breaker
- **Inconsistent behavior**: Different teams implement differently
- **Security complexity**: mTLS rollout across hundreds of services
- **Observability gaps**: Hard to trace requests across service boundaries
- **Deployment complexity**: Canary, blue-green, traffic splitting without code changes

---

## 2. Real-World Motivation

### Uber
- **Scale**: 4000+ microservices, millions of RPCs per second
- **Mesh**: Custom uMesh (Envoy-based), migrated from homegrown
- **Benefits**: mTLS for all service-to-service traffic; canary deployments; traffic mirroring
- **Challenges**: Migration from monolithic to mesh; performance overhead

### Airbnb
- **Adoption**: Service mesh for traffic management and observability
- **Use case**: Canary deployments; A/B testing; gradual rollout
- **Integration**: With Kubernetes, Envoy-based data plane

### Netflix
- **Pre-mesh**: Zuul (API gateway), Hystrix (circuit breaker), Eureka (discovery)
- **Evolution**: Moved toward mesh; use Envoy for some workloads
- **Focus**: Resilience and observability at scale

### Lyft
- **Envoy**: Major contributor to Envoy proxy
- **Use case**: Envoy as sidecar for all services; traffic management, observability
- **Blog**: Documented Envoy adoption and benefits

### Google
- **Internal**: Stubby (RPC), Borg; internal mesh-like infrastructure
- **Istio**: Co-founded with IBM, Lyft; open-source service mesh
- **GKE**: Istio integration for managed Kubernetes

---

## 3. Architecture Diagrams

### Data Plane vs Control Plane

```
                    ┌─────────────────────────────────────────────────┐
                    │              CONTROL PLANE                        │
                    │  ┌─────────┐ ┌─────────┐ ┌─────────┐           │
                    │  │ Config  │ │  Pilot  │ │  Citadel │           │
                    │  │  Mgmt   │ │  (xDS)  │ │  (mTLS)  │           │
                    │  └────┬────┘ └────┬────┘ └────┬────┘           │
                    └───────┼───────────┼───────────┼─────────────────┘
                            │           │           │
                    Push config to data plane
                            │           │           │
    ┌───────────────────────┼───────────┼───────────┼───────────────────┐
    │                       │    DATA PLANE (Sidecar Proxies)            │
    │  ┌─────────┐    ┌────▼────┐    ┌────▼────┐    ┌─────────┐         │
    │  │Service A│◀──▶│ Proxy A │◀──▶│ Proxy B │◀──▶│Service B│         │
    │  └─────────┘    └─────────┘    └─────────┘    └─────────┘         │
    │       │              │              │              │                │
    │       │    All traffic flows through proxy (intercept)             │
    └───────┴──────────────┴──────────────┴──────────────┴──────────────┘
```

### Sidecar Pattern

```
BEFORE (Direct service-to-service)
┌─────────────┐                    ┌─────────────┐
│  Service A  │────────────────────▶│  Service B  │
│  (Order)    │                    │  (Payment)  │
└─────────────┘                    └─────────────┘
     │                                      │
     │  Each service implements:             │
     │  - Retry, timeout, circuit breaker    │
     │  - Metrics, tracing                   │
     │  - mTLS                               │
└──────────────────────────────────────────┘

AFTER (Sidecar / Service Mesh)
┌─────────────────────────────────┐    ┌─────────────────────────────────┐
│  ┌─────────────┐  ┌───────────┐  │    │  ┌───────────┐  ┌─────────────┐  │
│  │  Service A  │  │  Proxy A  │  │    │  │  Proxy B  │  │  Service B   │  │
│  │  (Order)    │◀──▶│ (Sidecar) │◀───▶│  │ (Sidecar) │◀──▶│  (Payment)   │  │
│  └─────────────┘  └───────────┘  │    │  └───────────┘  └─────────────┘  │
│         │              │         │    │         │              │          │
│  App only does         │         │    │         │              │          │
│  business logic        │         │    │         │              │          │
│                        │         │    │         │              │          │
│  Proxy handles: retry, timeout, circuit breaker, mTLS, metrics, tracing   │
└─────────────────────────────────┘    └─────────────────────────────────┘
```

### Traffic Splitting (Canary)

```
                    ┌─────────────────┐
                    │   Ingress       │
                    │   (Gateway)     │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │  VirtualService  │
                    │  90% → v1        │
                    │  10% → v2        │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
     ┌────────▼────┐  ┌──────▼──────┐      │
     │  Order v1   │  │  Order v2    │      │
     │  (stable)   │  │  (canary)    │      │
     └─────────────┘  └──────────────┘      │
```

---

## 4. Core Mechanics

### Data Plane
- **Sidecar proxy**: Deployed alongside each service pod/VM
- **Intercept**: All inbound and outbound traffic flows through proxy
- **Responsibilities**: Load balancing, retry, timeout, circuit breaking, mTLS, metrics
- **Protocols**: HTTP/1.1, HTTP/2, gRPC, TCP

### Control Plane
- **Configuration**: Pushes config to data plane (e.g., xDS protocol)
- **Service discovery**: Integrates with Kubernetes, Consul, etc.
- **Certificate management**: Issues and rotates mTLS certs
- **Policy**: Authorization, rate limiting

### Key Capabilities
- **Traffic management**: Load balancing (round-robin, least request, consistent hash), retries, timeouts
- **Resilience**: Circuit breaking, outlier detection, connection pooling
- **Security**: mTLS by default, RBAC, JWT validation
- **Observability**: Access logs, metrics (latency, throughput), distributed tracing
- **Advanced routing**: Canary, blue-green, traffic mirroring, fault injection

---

## 5. Numbers

| Metric | Typical Range |
|--------|---------------|
| Latency overhead (sidecar) | 1-5ms per hop |
| Memory per sidecar | 50-150 MB |
| CPU per sidecar | 0.1-0.5 cores (idle) |
| mTLS handshake | ~1-2ms additional |
| Connection overhead | ~50-100 connections per proxy |
| Max connections (Envoy) | 100K+ per proxy |

### Istio/Envoy Scale (Reference)
- Uber: 4000+ services, Envoy proxies
- Single Envoy: 100K+ connections, 100K+ RPS
- Control plane: Can manage 1000s of proxies

---

## 6. Tradeoffs

### Service Mesh vs API Gateway

| Aspect | Service Mesh | API Gateway |
|--------|--------------|-------------|
| Scope | Service-to-service (east-west) | Client-to-service (north-south) |
| Traffic | Internal | External (internet, mobile) |
| Auth | mTLS, service identity | OAuth, API keys, JWT |
| Use case | Resilience, observability | Rate limiting, auth, API versioning |
| Placement | Sidecar per service | Single entry point |

### Service Mesh vs Library (e.g., Hystrix)

| Aspect | Service Mesh | Library |
|--------|--------------|---------|
| Language | Any (proxy is language-agnostic) | Per-language |
| Update | Deploy new proxy; no app change | Redeploy app |
| Consistency | Unified across all services | Per-team implementation |
| Overhead | Network hop (sidecar) | In-process |
| Polyglot | Yes | No (each language needs lib) |

### Sidecar Overhead

| Cost | Impact |
|------|--------|
| Latency | +1-5ms per hop |
| Memory | 50-150 MB per pod |
| CPU | 0.1-0.5 cores (idle) |
| Complexity | More moving parts; debugging |

---

## 7. Variants / Implementations

### Istio
- **Data plane**: Envoy proxy
- **Control plane**: Pilot (config), Citadel (certs), Galley (config validation)
- **Features**: Full-featured; traffic management, mTLS, observability, canary
- **Complexity**: High; many CRDs
- **Adoption**: Widely used; Kubernetes-native

### Linkerd
- **Data plane**: Lightweight Rust proxy (linkerd2-proxy)
- **Control plane**: Simpler than Istio
- **Focus**: Simplicity, performance, low resource usage
- **Features**: mTLS, retries, timeouts, metrics
- **Adoption**: CNCF project; easier to deploy

### Consul Connect
- **Data plane**: Envoy proxy (default) or built-in proxy
- **Control plane**: Consul (service discovery, KV, health)
- **Integration**: Works with Consul service discovery; not just Kubernetes
- **Use case**: Multi-cloud, VM + container workloads

### AWS App Mesh
- **Data plane**: Envoy proxy
- **Control plane**: AWS-managed
- **Integration**: ECS, EKS, EC2
- **Use case**: AWS-native; integrates with X-Ray, Cloud Map

### Cilium Service Mesh
- **Approach**: eBPF-based; no sidecar
- **Mechanism**: Kernel-level networking; no proxy overhead
- **Use case**: High-performance; Kubernetes

---

## 8. Scaling Strategies

### Horizontal Scaling
- Scale service pods; each gets sidecar automatically
- Proxy scales with pod count

### Control Plane Scaling
- Istio: Multiple replicas of Pilot, Citadel
- Config propagation: Eventually consistent

### Performance Tuning
- **Connection pooling**: Limit connections per upstream
- **Concurrency**: Tune proxy worker threads
- **Tracing sampling**: Reduce sampling rate at scale (e.g., 1%)
- **Access logging**: Disable or sample for high traffic

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|------------|
| Sidecar crash | Pod loses connectivity | Restart policy; sidecar runs in same pod |
| Control plane down | No new config; existing proxies continue | Proxies cache config; control plane HA |
| mTLS cert expiry | Connection failures | Citadel auto-rotation; monitor expiry |
| Proxy overload | Latency increase | Resource limits; scale pods |
| Config push delay | Stale routing | xDS eventual consistency; health checks |
| Mesh upgrade | Compatibility | Canary proxy rollout |

---

## 10. Performance Considerations

- **Latency**: Sidecar adds 1-5ms; optimize for hot path
- **Memory**: 50-150 MB per sidecar; factor into pod sizing
- **Connection pooling**: Reuse connections; avoid connection per request
- **Tracing**: Sample at 1-10% for production
- **Access logs**: Disable or use sampling for high RPS
- **mTLS**: Use session resumption to reduce handshake overhead

---

## 11. Use Cases

**Good fit:**
- Large microservices deployment (10+ services)
- Need for consistent resilience (retry, circuit breaker)
- mTLS requirement across services
- Canary/blue-green without code changes
- Polyglot environment (different languages)
- Observability (tracing, metrics) across services

**Poor fit:**
- Small deployment (few services)
- Latency-sensitive (every ms counts)
- Resource-constrained (sidecar overhead)
- Simple CRUD; no resilience needs
- Team unfamiliar with mesh complexity

---

## 12. Comparison Tables

### Mesh Implementation Comparison

| Feature | Istio | Linkerd | Consul Connect |
|---------|-------|---------|----------------|
| Complexity | High | Low | Medium |
| Proxy | Envoy | linkerd2-proxy | Envoy |
| mTLS | Yes | Yes | Yes |
| Canary | Yes | Yes | Yes |
| K8s | Primary | Primary | Yes + VM |
| Resource usage | Higher | Lower | Medium |

### When to Use What

| Need | Recommendation |
|------|----------------|
| Simple resilience | Linkerd |
| Full feature set | Istio |
| Multi-cloud, VM | Consul Connect |
| AWS-only | App Mesh |
| Zero sidecar overhead | Cilium (eBPF) |

---

## 13. Code or Pseudocode

### Istio VirtualService (Canary)

```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: order-service
spec:
  hosts:
  - order-service
  http:
  - match:
    - headers:
        version:
          exact: "v2"
    route:
    - destination:
        host: order-service
        subset: v2
  - route:
    - destination:
        host: order-service
        subset: v1
      weight: 90
    - destination:
        host: order-service
        subset: v2
      weight: 10
```

### Istio DestinationRule (Subsets)

```yaml
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: order-service
spec:
  host: order-service
  trafficPolicy:
    connectionPool:
      tcp:
        maxConnections: 100
      http:
        http1MaxPendingRequests: 50
        http2MaxRequests: 100
        maxRequestsPerConnection: 2
    outlierDetection:
      consecutiveErrors: 5
      interval: 30s
      baseEjectionTime: 30s
      maxEjectionPercent: 50
  subsets:
  - name: v1
    labels:
      version: v1
  - name: v2
    labels:
      version: v2
```

### Envoy Retry Configuration (xDS)

```yaml
retry_policy:
  retry_on: "5xx,reset,connect-failure,refused-stream"
  num_retries: 3
  per_try_timeout: 2s
  retry_back_off:
    base_interval: 0.25s
    max_interval: 30s
```

---

## 14. Interview Discussion

### Key Points
1. **Data plane**: Proxies; control plane: config
2. **Sidecar**: Every pod gets proxy; traffic flows through it
3. **Benefits**: Resilience, observability, mTLS without code changes
4. **Cost**: Latency, memory, complexity
5. **When**: Many services; need consistency; polyglot
6. **vs API Gateway**: Mesh = east-west; Gateway = north-south

### Common Questions
- **"What is a service mesh?"** → Infrastructure layer for service-to-service communication; data plane (proxies) + control plane (config)
- **"What problems does it solve?"** → Retry, timeout, circuit breaker, mTLS, tracing—without code changes
- **"What's the overhead?"** → 1-5ms latency, 50-150 MB memory per sidecar
- **"When would you NOT use a mesh?"** → Small deployment; latency-critical; resource-constrained
- **"Mesh vs API Gateway?"** → Mesh = internal; Gateway = external entry point

### Red Flags
- Suggesting mesh for 2-3 services
- Ignoring overhead
- No mention of control plane
- Confusing mesh with API gateway

---

## Appendix: Service Mesh Deep Dives

### xDS Protocol
- **Discovery service**: Envoy's configuration API
- **Types**: LDS (listeners), RDS (routes), CDS (clusters), EDS (endpoints)
- **Flow**: Control plane pushes config to Envoy via gRPC/HTTP
- **Incremental**: Only changed config pushed; reduces load

### mTLS in Service Mesh
- **Mutual authentication**: Both client and server present certificates
- **Automatic**: Citadel/Istio issues certs to each pod
- **Rotation**: Short-lived certs (e.g., 24h); auto-rotated
- **Zero-trust**: Encrypt all service-to-service traffic by default

### Traffic Mirroring (Shadowing)
- **Use case**: Test new version with production traffic; no user impact
- **Flow**: Copy traffic to shadow cluster; compare results
- **Istio**: `mirror` destination in VirtualService
- **Caution**: Double load on shadow; ensure it can handle

### Fault Injection
- **Purpose**: Test resilience; chaos engineering
- **Types**: Delay (latency), abort (error), limit bandwidth
- **Istio**: `fault` in VirtualService; inject to percentage of traffic
- **Example**: 10% of requests get 5s delay; verify timeout handling
