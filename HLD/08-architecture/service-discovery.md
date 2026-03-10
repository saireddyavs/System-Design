# Service Discovery

## 1. Concept Overview

### Definition
Service discovery is the mechanism by which services in a distributed system find and communicate with each other. In dynamic environments (cloud, containers), service instances have ephemeral IPs and ports. Service discovery maintains a registry of available instances and their network locations, allowing clients to locate healthy instances without hardcoding addresses.

### Purpose
- **Dynamic addressing**: Services get new IPs on restart, scale-out; clients need current addresses
- **Load distribution**: Route requests across multiple instances
- **Health awareness**: Only route to healthy instances
- **Decoupling**: Clients don't need to know deployment topology
- **Resilience**: Automatically exclude failed instances

### Problems It Solves
- **Static configuration**: Hardcoded IPs break when instances move
- **Manual updates**: No need to update config when instances scale
- **Single point of failure**: Avoid routing to dead instances
- **Load balancing**: Distribute across available instances
- **Multi-environment**: Same service name across dev/staging/prod

---

## 2. Real-World Motivation

### Netflix
- **Eureka**: Open-source service registry; services register on startup, deregister on shutdown
- **Scale**: Hundreds of services; thousands of instances; 100+ deployments/day
- **Client-side discovery**: Application uses Eureka client to find instances; Ribbon for load balancing
- **Integration**: Spring Cloud; used with Zuul, Hystrix

### Uber
- **Multi-region**: Services discover instances in same region for latency
- **Consul/etcd**: Service registry with health checks
- **Dynamic scaling**: New instances auto-register; traffic flows automatically

### Amazon
- **AWS Cloud Map**: Managed service discovery for ECS, EKS, EC2
- **Internal**: Proprietary discovery for millions of services
- **Route 53**: DNS-based discovery for some workloads

### Google
- **Internal**: Borg/Kubernetes; built-in service discovery via DNS and environment variables
- **Kubernetes Services**: ClusterIP, NodePort; DNS name `service-name.namespace.svc.cluster.local`

### Airbnb
- **Consul**: Service discovery with health checks
- **Use case**: Microservices find each other; no hardcoded IPs

---

## 3. Architecture Diagrams

### Client-Side Discovery

```
                    ┌─────────────────────────────────────────┐
                    │           SERVICE REGISTRY                │
                    │  (Eureka, Consul, etcd, ZooKeeper)        │
                    │  ┌─────────────────────────────────────┐  │
                    │  │ order-service: [ip1:8080, ip2:8080] │  │
                    │  │ payment-service: [ip3:8080, ip4:8080]│  │
                    │  └─────────────────────────────────────┘  │
                    └────────────▲──────────────────▲───────────┘
                                 │                  │
                    Query        │                  │    Query
                    instances    │                  │    instances
                                 │                  │
    ┌────────────────────────────┴──┐    ┌──────────┴────────────────────┐
    │  CLIENT (Order Service)        │    │  CLIENT (Frontend)           │
    │  1. Query registry             │    │  1. Query registry           │
    │  2. Get list of instances      │    │  2. Get list of instances    │
    │  3. Load balance (client)      │    │  3. Load balance (client)    │
    │  4. Call instance directly     │    │  4. Call instance directly  │
    └───────────────────────────────┼────┴──────────────────────────────┘
                                   │
                                   ▼
                    ┌─────────────────────────────────────────┐
                    │  PAYMENT SERVICE INSTANCES                │
                    │  [ip3:8080] [ip4:8080]                    │
                    └─────────────────────────────────────────┘
```

### Server-Side Discovery

```
    ┌─────────────┐
    │   CLIENT    │
    │  (Order Svc)│
    └──────┬──────┘
           │
           │ Request to "payment-service" (DNS name or host)
           │
           ▼
    ┌─────────────────────────────────────────┐
    │           LOAD BALANCER                  │
    │  (AWS ALB, Kubernetes Service, etc.)    │
    │  1. Receives request                    │
    │  2. Queries registry for instances      │
    │  3. Load balances to healthy instance   │
    └────────────────┬──────────────────────┘
                      │
                      │ Queries registry
                      ▼
    ┌─────────────────────────────────────────┐
    │           SERVICE REGISTRY               │
    │  payment-service: [ip1, ip2, ip3]        │
    └────────────────┬──────────────────────┘
                      │
                      ▼
    ┌─────────────────────────────────────────┐
    │  PAYMENT SERVICE INSTANCES               │
    │  [ip1:8080] [ip2:8080] [ip3:8080]       │
    └─────────────────────────────────────────┘
```

### Self-Registration vs Third-Party Registration

```
SELF-REGISTRATION
┌─────────────┐                    ┌─────────────────┐
│  Service    │── Register ────────▶│    Registry     │
│  Instance   │   (on startup)      │  (Eureka,       │
│             │                    │   Consul)       │
│             │◀── Heartbeat ──────│                 │
│             │   (periodic)       │                 │
│             │                    │                 │
│             │── Deregister ──────▶│  (on shutdown)  │
└─────────────┘                    └─────────────────┘
```

```
THIRD-PARTY REGISTRATION
┌─────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Service    │     │  Registrar      │     │    Registry      │
│  Instance   │     │  (Kubernetes,   │     │  (Consul, etcd)  │
│             │     │   AWS)          │     │                  │
└──────┬──────┘     └────────┬────────┘     └────────┬─────────┘
       │                     │                       │
       │  Platform knows     │  Registrar polls      │
       │  about instance     │  platform & registers │
       │  (container, VM)    │  instances            │
       └────────────────────┴───────────────────────┘
```

### Kubernetes Service Discovery

```
    ┌─────────────────────────────────────────────────────────────┐
    │                    KUBERNETES CLUSTER                         │
    │                                                              │
    │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
    │  │  Pod        │  │  Pod        │  │  Pod        │         │
    │  │  order-xxx  │  │  order-yyy  │  │  order-zzz  │         │
    │  │  ip: 10.0.1 │  │  ip: 10.0.2 │  │  ip: 10.0.3 │         │
    │  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘         │
    │         │                │                │                  │
    │         └────────────────┼────────────────┘                  │
    │                          │                                   │
    │                  ┌───────▼───────┐                           │
    │                  │   Service     │                           │
    │                  │ order-service │  ClusterIP: 10.96.1.1     │
    │                  │ (selector:    │  DNS: order-service.default.svc
    │                  │  app=order)   │                           │
    │                  └───────────────┘                           │
    │                          │                                   │
    │  Client resolves "order-service" → gets ClusterIP → LB to pod │
    └─────────────────────────────────────────────────────────────┘
```

---

## 4. Core Mechanics

### Client-Side Discovery
- **Flow**: Client queries registry for service instances; client chooses instance (load balancing); client calls instance directly
- **Pros**: No extra hop; client can implement custom load balancing
- **Cons**: Client must implement discovery logic; coupling to registry
- **Examples**: Eureka + Ribbon, Consul client

### Server-Side Discovery
- **Flow**: Client sends request to load balancer (DNS name); load balancer queries registry; load balancer forwards to instance
- **Pros**: Client is simple; load balancer handles discovery
- **Cons**: Extra network hop; load balancer can be bottleneck
- **Examples**: AWS ALB + Cloud Map, Kubernetes Service

### Service Registry
- **Central database** of service instances: service name → list of (host, port, metadata)
- **Operations**: Register, deregister, heartbeat, query
- **Consistency**: CP (Consul, etcd) vs AP (Eureka) trade-off

### Registration Modes
- **Self-registration**: Service registers itself on startup; sends heartbeat; deregisters on shutdown
- **Third-party**: Platform (Kubernetes, ECS) or agent registers on behalf of service
- **Kubernetes**: Automatic; no explicit registration; Service + Endpoints

### Health Checking
- **Heartbeat**: Service sends "I'm alive" periodically; registry marks unhealthy if missed
- **TCP/HTTP check**: Registry proactively checks instance
- **Application health**: Service exposes /health; registry calls it

---

## 5. Numbers

| Metric | Typical Range |
|--------|---------------|
| Registry size | 1K-100K+ services |
| Instances per service | 1-1000+ |
| Heartbeat interval | 5-30 seconds |
| Unhealthy timeout | 15-90 seconds |
| DNS TTL (K8s) | 30 seconds (default) |
| Cache refresh (Eureka) | 30 seconds |

### Scale (Reference)
- Netflix Eureka: 1000s of services, 10K+ instances
- Consul: Used at scale for 1000s of nodes
- Kubernetes: Single cluster 5000 nodes, 150K pods (limits)

---

## 6. Tradeoffs

### Client-Side vs Server-Side Discovery

| Aspect | Client-Side | Server-Side |
|--------|-------------|-------------|
| Latency | Lower (direct) | Higher (LB hop) |
| Client complexity | Higher | Lower |
| Load balancer | Not needed | Required |
| Failure mode | Registry down = no discovery | LB can cache |
| Examples | Eureka, Consul | Kubernetes, AWS ALB |
| Load balancing | Client implements | LB implements |

### Self vs Third-Party Registration

| Aspect | Self-Registration | Third-Party |
|--------|-------------------|-------------|
| Service code | Must include SDK | No change |
| Platform support | Any | K8s, ECS, etc. |
| Failure | Service crash may not deregister | Platform knows |
| Complexity | In service | In platform |

### CP vs AP Registry

| Aspect | CP (Consul, etcd) | AP (Eureka) |
|--------|-------------------|------------|
| Consistency | Strong | Eventual |
| Availability | May block on partition | Always readable |
| Use case | Critical config | Service discovery (stale OK) |
| Failure | Partition = no writes | Partition = stale reads |

---

## 7. Variants / Implementations

### Consul
- **Features**: Service discovery, health checks, KV store, multi-datacenter
- **Protocol**: HTTP, DNS
- **Health**: TCP, HTTP, gRPC, TTL, Docker
- **Consistency**: CP (Raft)
- **Use case**: Multi-datacenter; VM + container

### etcd
- **Features**: Distributed KV; used by Kubernetes
- **Protocol**: gRPC
- **Consistency**: CP (Raft)
- **Use case**: Kubernetes control plane; service discovery via K8s

### ZooKeeper
- **Features**: Distributed coordination; znodes; watches
- **Consistency**: CP (ZAB)
- **Use case**: Kafka (broker registration); older systems
- **Note**: Heavier than etcd/Consul for discovery

### Eureka (Netflix)
- **Features**: Service registry; client-side discovery
- **Consistency**: AP (eventual)
- **Integration**: Spring Cloud, Ribbon
- **Use case**: Java/Spring ecosystems

### Kubernetes
- **Built-in**: Services + Endpoints; DNS (CoreDNS)
- **No external registry**: K8s is the registry
- **DNS format**: `service-name.namespace.svc.cluster.local`
- **Use case**: Kubernetes-native workloads

### AWS Cloud Map
- **Features**: Managed service discovery; HTTP namespace; DNS
- **Integration**: ECS, EKS, EC2, Lambda
- **Use case**: AWS-native

---

## 8. Scaling Strategies

- **Registry scaling**: Consul/etcd cluster; multiple nodes
- **Caching**: Client caches instance list; refresh periodically
- **Sharding**: Large deployments may shard by service type or region
- **DNS caching**: Kubernetes DNS cached at node; TTL tuning
- **Read replicas**: Eureka peer replication for read scaling

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|-------------|
| Registry down | No new lookups | Client cache; multiple registry replicas |
| Stale entries | Route to dead instance | Health checks; short TTL; timeout |
| Network partition | Split registry | CP: block; AP: stale reads |
| Registration failure | Instance not discoverable | Retry; platform registration (K8s) |
| Heartbeat loss | False unhealthy | Tune timeout; avoid aggressive eviction |
| Thundering herd | All clients refresh at once | Stagger cache refresh; jitter |

---

## 10. Performance Considerations

- **Cache**: Client caches instance list; avoid registry call per request
- **Connection pooling**: Reuse connections to instances
- **DNS**: Kubernetes DNS can add latency; consider client-side discovery for hot path
- **Registry load**: Many clients × refresh rate; tune cache TTL
- **Health check frequency**: Balance freshness vs load on instances

---

## 11. Use Cases

**Essential for:**
- Microservices (dynamic instances)
- Container orchestration (Kubernetes, ECS)
- Cloud deployments (ephemeral IPs)
- Auto-scaling (instances come and go)

**Less critical for:**
- Static deployments (fixed IPs)
- Monolith (single instance)
- Small fixed cluster

---

## 12. Comparison Tables

### Service Discovery Implementation Comparison

| Feature | Consul | etcd | ZooKeeper | Eureka | K8s DNS |
|---------|--------|------|-----------|--------|---------|
| Consistency | CP | CP | CP | AP | N/A (K8s) |
| Health checks | Yes | Via K8s | No (custom) | Yes | Yes |
| KV store | Yes | Yes | Yes | No | No |
| Multi-DC | Yes | Limited | Yes | Yes | No |
| Protocol | HTTP, DNS | gRPC | Custom | HTTP | DNS |
| Primary use | Discovery, config | K8s | Kafka, etc. | Java/Spring | K8s |

### DNS-Based vs Registry-Based

| Aspect | DNS-Based | Registry-Based |
|--------|-----------|----------------|
| Protocol | DNS | HTTP, gRPC |
| TTL | Cache by TTL | Client refresh |
| Load balancing | DNS round-robin (limited) | Client or LB |
| Health | DNS doesn't know health | Registry has health |
| Examples | K8s DNS, Route 53 | Consul, Eureka |

---

## 13. Code or Pseudocode

### Self-Registration (Eureka-style)

```python
# Service registers on startup
def register_with_eureka(service_name, host, port):
    eureka_client.register(
        instance_id=f"{host}:{port}",
        app=service_name,
        ip=host,
        port=port,
        health_url=f"http://{host}:{port}/health"
    )
    # Start heartbeat thread
    threading.Thread(target=heartbeat_loop, args=(service_name, host, port)).start()

def heartbeat_loop(service_name, host, port):
    while True:
        time.sleep(30)
        eureka_client.renew(service_name, f"{host}:{port}")

# On shutdown
def shutdown():
    eureka_client.deregister(service_name, f"{host}:{port}")
```

### Client-Side Discovery (Get Instances)

```python
def get_payment_service_url():
    instances = eureka_client.get_instances("payment-service")
    if not instances:
        raise ServiceUnavailableError("No payment service instances")
    
    # Client-side load balancing (round-robin)
    instance = load_balancer.select(instances)
    return f"http://{instance.ip}:{instance.port}"
```

### Kubernetes Service (YAML)

```yaml
apiVersion: v1
kind: Service
metadata:
  name: order-service
spec:
  selector:
    app: order
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
---
# Pods with label app=order are automatically discovered
# DNS: order-service.default.svc.cluster.local
```

### Consul Registration

```python
import consul

c = consul.Consul()

# Register service
c.agent.service.register(
    name='payment-service',
    service_id='payment-1',
    address='10.0.1.5',
    port=8080,
    check=consul.Check.http('http://10.0.1.5:8080/health', interval='10s')
)

# Discover services
_, services = c.health.service('payment-service', passing=True)
for s in services:
    print(f"{s['Service']['Address']}:{s['Service']['Port']}")
```

---

## 14. Interview Discussion

### Key Points
1. **Why needed**: Dynamic IPs in cloud/containers; instances scale, restart
2. **Client-side**: Client queries registry; load balances; calls directly
3. **Server-side**: LB queries registry; client talks to LB
4. **Registration**: Self (service registers) vs third-party (platform does it)
5. **Health**: Critical; don't route to dead instances
6. **Kubernetes**: Built-in; Service + DNS; no external registry needed

### Common Questions
- **"What is service discovery?"** → Mechanism for services to find each other's network addresses in dynamic environments
- **"Client-side vs server-side?"** → Client-side: client queries, load balances; Server-side: LB does it
- **"How does Kubernetes do it?"** → Services + Endpoints; DNS; selector matches pods
- **"What if registry goes down?"** → Client cache; use cached list; registry HA
- **"Consul vs Eureka?"** → Consul: CP, KV, multi-DC; Eureka: AP, Java/Spring, simpler

### Red Flags
- Hardcoding IPs in microservices
- No health checking
- Ignoring registry failure modes
- No caching (registry call per request)
