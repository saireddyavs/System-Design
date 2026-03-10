# Load Balancing

## 1. Concept Overview

### Definition
Load balancing is the practice of distributing incoming network traffic across multiple servers (backends) to ensure no single server is overwhelmed, to maximize throughput, minimize response time, and avoid single points of failure.

### Purpose
- **High Availability**: If one server fails, others continue serving
- **Scalability**: Add servers to handle more traffic
- **Performance**: Distribute load to prevent overload
- **Flexibility**: Rolling updates, maintenance without downtime
- **Geographic distribution**: Route users to nearest/optimal datacenter

### Problems It Solves
1. **Single point of failure**: One server down = service down
2. **Overload**: Traffic spikes overwhelm single server
3. **Uneven utilization**: Some servers idle while others overloaded
4. **Maintenance**: Can't take down server for updates without outage
5. **Geographic latency**: Users far from single datacenter experience high latency

---

## 2. Real-World Motivation

### Google
- **Global Load Balancing**: 200+ datacenters, route to nearest
- **Maglev**: Consistent hashing load balancer, 10M+ connections
- **Internal**: Stubby/gRPC load balancing for microservices
- **Andromeda**: Network virtualization, load balancing

### Netflix
- **Zuul**: API gateway, L7 load balancing, routing, filtering
- **Eureka**: Service discovery + client-side load balancing
- **Ribbon**: Client-side LB library (deprecated, replaced by Spring Cloud LoadBalancer)
- **Chaos Monkey**: Tests resilience by killing instances—LB must handle

### Uber
- **Multi-region**: Load balance across US, EU, APAC
- **Ribbon/Envoy**: Service mesh for microservices
- **Geographic routing**: Route to nearest region

### Amazon
- **ELB**: Elastic Load Balancing (Classic, Application, Network)
- **ALB**: L7, path-based routing, WebSocket, HTTP/2
- **NLB**: L4, ultra-low latency, millions of connections
- **Multi-AZ**: Automatic distribution across availability zones

### Twitter
- **Finagle**: RPC framework with load balancing
- **Multiple datacenters**: Active-active, failover
- **Traffic management**: Canary, gradual rollouts

---

## 3. Architecture Diagrams

### Basic Load Balancer Topology

```
                    +------------------+
                    |   DNS / Client   |
                    +--------+---------+
                             |
                             v
                    +------------------+
                    |  LOAD BALANCER   |
                    |  (Virtual IP)    |
                    +--------+---------+
                             |
         +-------------------+-------------------+
         |                   |                   |
         v                   v                   v
+----------------+  +----------------+  +----------------+
|   Backend 1    |  |   Backend 2    |  |   Backend 3    |
|   (Healthy)    |  |   (Healthy)    |  |   (Healthy)    |
+----------------+  +----------------+  +----------------+
```

### L4 vs L7 Load Balancer

```
L4 (Transport Layer - TCP/UDP):
+----------+     +----------+     +----------+
|  Client  |---->|   L4 LB  |---->| Backend  |
+----------+     +----------+     +----------+
                 Sees: IP, Port
                 No HTTP awareness

L7 (Application Layer - HTTP):
+----------+     +----------+     +----------+
|  Client  |---->|   L7 LB  |---->| Backend  |
+----------+     +----------+     +----------+
                 Sees: URL, Headers, Cookies
                 Path-based routing, SSL termination
```

### Global Server Load Balancing (GSLB)

```
                    +------------------+
                    |   Global DNS     |
                    |   (GeoDNS)       |
                    +--------+---------+
                             |
              +--------------+--------------+
              |              |              |
       +------v------+ +-----v-----+ +-----v-----+
       |  US Region  | | EU Region | | APAC Region|
       |  +-------+  | | +-------+ | | +-------+ |
       |  | LB    |  | | | LB    | | | | LB    | |
       |  +---+---+  | | +---+---+ | | +---+---+ |
       |  | B1 | B2 | | | B1 | B2 | | | B1 | B2 | |
       |  +---+---+  | | +---+---+ | | +---+---+ |
       +-------------+ +-----------+ +-----------+
```

### Health Check Flow

```
                    +------------------+
                    |  Load Balancer   |
                    +--------+---------+
                             |
         +-------------------+-------------------+
         |                   |                   |
         v                   v                   v
    +--------+          +--------+          +--------+
    | B1     |          | B2     |          | B3     |
    | Healthy|          | Unhealthy|        | Healthy|
    +--------+          +--------+          +--------+
         |                   X                   |
         |              (Removed from           |
         |               pool)                  |
         +-------------------+-------------------+
                             |
                    Traffic only to B1, B3
```

### Consistent Hashing (Simplified)

```
        Hash Ring (0 to 2^32-1)
              |
     B1 -----+----- B2
              |
              +----- B3
              |
    Key "user123" hashes to point X
    -> Assigned to B2 (next node clockwise)
    
    When B2 is removed: user123 -> B3
    Other keys unchanged (minimal reassignment)
```

---

## 4. Core Mechanics

### L4 Load Balancing
- **Operates on**: IP, port (TCP/UDP)
- **Connection-based**: Forwards entire TCP connection to one backend
- **No content inspection**: Cannot route based on URL, headers
- **State**: Maintains connection table (client IP:port <-> backend IP:port)
- **NAT**: Often performs NAT (client sees LB IP, backend sees LB IP)

### L7 Load Balancing
- **Operates on**: HTTP/HTTPS (headers, body, URL)
- **Request-based**: Can route each request differently (e.g., /api -> backend1, /static -> backend2)
- **SSL termination**: Decrypts at LB, forwards plaintext to backend (or re-encrypt)
- **Features**: Path routing, host routing, header rewrite, rate limiting

### Load Balancing Algorithms

#### Round Robin
- Rotate through backends in order: B1, B2, B3, B1, B2, B3...
- **Pros**: Simple, even distribution
- **Cons**: Ignores server load, capacity

#### Weighted Round Robin
- Assign weight to each backend (e.g., B1:3, B2:1, B3:1)
- B1 gets 3 requests for every 1 to B2/B3
- **Pros**: Account for capacity differences
- **Cons**: Static weights

#### Least Connections
- Route to backend with fewest active connections
- **Pros**: Adapts to varying request duration
- **Cons**: Connection count != actual load

#### Least Response Time
- Route to backend with lowest latency
- **Pros**: Optimal for latency-sensitive
- **Cons**: Requires active probing or feedback

#### IP Hash
- Hash(client IP) mod N -> backend
- **Pros**: Same client always to same backend (sticky)
- **Cons**: Uneven if IP distribution skewed

#### Consistent Hashing
- Hash backends and keys onto ring; key maps to nearest backend
- **Pros**: Minimal reassignment when backends added/removed
- **Cons**: Can be uneven (virtual nodes solve this)

### Health Checks
- **Active**: LB periodically sends request to backend (HTTP GET /health, TCP connect)
- **Passive**: LB observes backend response (success/failure of real requests)
- **Interval**: e.g., every 5s
- **Threshold**: N failures before marking unhealthy

### Connection Draining
- When removing backend: stop new connections, wait for existing to drain
- **Drain timeout**: e.g., 300s

---

## 5. Numbers

### Capacity
| Load Balancer | Connections | Throughput | Latency |
|---------------|-------------|------------|---------|
| **Nginx** | ~50K concurrent | 10K+ RPS | <1ms |
| **HAProxy** | Millions | 100K+ RPS | <1ms |
| **AWS ALB** | Millions | 100K+ RPS | <1ms |
| **AWS NLB** | Millions | Millions | <1ms |
| **Envoy** | 100K+ | 50K+ RPS | <1ms |

### Health Check
- **Typical interval**: 5-30 seconds
- **Unhealthy threshold**: 2-5 consecutive failures
- **Healthy threshold**: 2-3 consecutive successes

### Connection Draining
- **Typical timeout**: 30-300 seconds
- **Best practice**: Match to longest request duration

### Latency Overhead
- **L4**: ~0.1-0.5ms
- **L7**: ~0.5-2ms (SSL termination, parsing)

---

## 6. Tradeoffs

### L4 vs L7

| Aspect | L4 | L7 |
|--------|-----|-----|
| **Layer** | Transport (TCP/UDP) | Application (HTTP) |
| **Routing** | IP, port | URL, headers, host |
| **SSL** | Pass-through or terminate | Typically terminate |
| **Overhead** | Lower | Higher |
| **Use case** | High throughput, TCP | HTTP routing, path-based |

### Algorithm Comparison

| Algorithm | Use Case | Pros | Cons |
|-----------|----------|------|------|
| **Round Robin** | Equal capacity | Simple | Ignores load |
| **Weighted RR** | Unequal capacity | Flexible | Static |
| **Least Connections** | Variable request time | Adaptive | Connection != load |
| **Least Response Time** | Latency-critical | Optimal | Complex |
| **IP Hash** | Sticky sessions | Simple | Uneven |
| **Consistent Hash** | Cache affinity | Minimal reassignment | Uneven (use virtual nodes) |

### Hardware vs Software

| Aspect | Hardware LB | Software LB |
|--------|-------------|-------------|
| **Performance** | Very high | High (scales with instances) |
| **Cost** | High | Lower |
| **Flexibility** | Limited | High |
| **Scaling** | Replace/upgrade | Add instances |
| **Examples** | F5, Citrix | Nginx, HAProxy, Envoy |

---

## 7. Variants / Implementations

### Cloud Load Balancers
- **AWS**: ALB (L7), NLB (L4), CLB (Classic, legacy)
- **GCP**: Global HTTP(S) LB, Regional LB, Network LB
- **Azure**: Application Gateway (L7), Load Balancer (L4)

### Software Load Balancers
- **Nginx**: L7, reverse proxy, high performance
- **HAProxy**: L4/L7, very mature
- **Envoy**: L4/L7, service mesh, dynamic config
- **Traefik**: L7, Kubernetes-native, automatic Let's Encrypt

### Client-Side Load Balancing
- **Ribbon** (Netflix): Client chooses backend from list
- **gRPC**: Built-in round-robin, pick-first
- **Service mesh**: Envoy sidecar does client-side LB

### DNS Load Balancing
- **Round-robin DNS**: Multiple A records
- **Limitation**: Client caching, no health awareness
- **Use**: Often combined with other LB (e.g., GeoDNS + LB per region)

---

## 8. Scaling Strategies

### Horizontal Scaling
- **Add more LB instances**: Behind DNS or anycast
- **Stateless**: LB should not maintain state (or use shared state)

### Connection Pooling
- **L7 LB**: Maintain pool of connections to backends
- **Reduce**: Connection overhead per request

### Auto Scaling
- **Backend pool**: Add/remove backends based on load
- **LB**: Must support dynamic backend registration (e.g., AWS Target Groups)

### Multi-Region
- **GSLB**: DNS routes to region
- **Per-region LB**: Each region has own LB + backends

---

## 9. Failure Scenarios

### Load Balancer Failure
- **Mitigation**: Multiple LB instances, active-passive or active-active
- **Health check**: External monitor (e.g., Route 53 health check)
- **Failover**: DNS failover, or floating IP (VIP)

### Backend Failure
- **Mitigation**: Health checks, remove unhealthy
- **Cascading**: Too many backends down -> LB overloaded

### Thundering Herd
- **Scenario**: All backends come back; LB sends all traffic to first healthy
- **Mitigation**: Stagger health check recovery, gradual add

### Real Incidents
- **AWS ELB (2011)**: Single AZ outage affected ELB
- **Mitigation**: Multi-AZ deployment
- **Netflix**: Chaos Monkey kills instances; LB must handle
- **Lesson**: Design for failure; test failover

---

## 10. Performance Considerations

### SSL/TLS Termination
- **CPU-intensive**: Offload at LB or use hardware acceleration
- **Session resumption**: Reduce handshake overhead

### Connection Limits
- **File descriptors**: LB needs FD per connection
- **Tuning**: ulimit, kernel parameters

### Caching
- **L7 LB**: Can cache responses (e.g., Nginx proxy_cache)
- **Reduces**: Backend load

### Monitoring
- **Metrics**: Request latency, error rate, backend health
- **Alerts**: Backend down, high latency

---

## 11. Use Cases

| Use Case | LB Type | Algorithm |
|----------|---------|-----------|
| **Web app** | L7 | Least connections |
| **API** | L7 | Round robin, least connections |
| **WebSocket** | L7 | Sticky (IP hash or cookie) |
| **TCP gaming** | L4 | Least connections |
| **gRPC** | L7 | Round robin |
| **Static content** | L7 | Round robin |
| **Database read replicas** | L4 | Least connections |
| **Multi-region** | GSLB | Geo-based |

---

## 12. Comparison Tables

### AWS Load Balancer Types

| Type | Layer | Use Case | Key Feature |
|------|-------|----------|-------------|
| **ALB** | L7 | HTTP/HTTPS | Path routing, WebSocket |
| **NLB** | L4 | TCP/UDP | Ultra-low latency, static IP |
| **GLB** | L7 | Global | Geo routing, anycast |
| **CLB** | L4/L7 | Legacy | Legacy, use ALB/NLB |

### Nginx vs HAProxy vs Envoy

| Aspect | Nginx | HAProxy | Envoy |
|--------|-------|---------|-------|
| **L7** | Yes | Yes | Yes |
| **L4** | Yes | Yes | Yes |
| **Dynamic config** | Limited | Limited | Yes (xDS) |
| **Service mesh** | No | No | Yes |
| **Observability** | Good | Good | Excellent |

---

## 13. Code or Pseudocode

### Round Robin Algorithm

```python
class RoundRobinLB:
    def __init__(self, backends):
        self.backends = backends
        self.index = 0
    
    def next_backend(self):
        backend = self.backends[self.index % len(self.backends)]
        self.index += 1
        return backend
```

### Least Connections

```python
class LeastConnectionsLB:
    def __init__(self, backends):
        self.backends = {b: 0 for b in backends}
    
    def next_backend(self):
        return min(self.backends, key=self.backends.get)
    
    def on_connect(self, backend):
        self.backends[backend] += 1
    
    def on_disconnect(self, backend):
        self.backends[backend] -= 1
```

### Consistent Hashing

```python
import hashlib

class ConsistentHashLB:
    def __init__(self, backends, virtual_nodes=100):
        self.ring = {}
        for b in backends:
            for i in range(virtual_nodes):
                h = int(hashlib.md5(f"{b}:{i}".encode()).hexdigest(), 16)
                self.ring[h] = b
        self.sorted_keys = sorted(self.ring.keys())
    
    def next_backend(self, key):
        h = int(hashlib.md5(key.encode()).hexdigest(), 16)
        for k in self.sorted_keys:
            if h <= k:
                return self.ring[k]
        return self.ring[self.sorted_keys[0]]
```

### Health Check (Pseudocode)

```python
async def health_check_loop(backend, interval=5):
    while True:
        try:
            response = await http_get(f"http://{backend}/health")
            if response.status == 200:
                backend.set_healthy()
            else:
                backend.record_failure()
        except Exception:
            backend.record_failure()
        
        if backend.failures >= 3:
            backend.set_unhealthy()
        elif backend.successes >= 2:
            backend.set_healthy()
        
        await asyncio.sleep(interval)
```

---

## 14. Interview Discussion

### Key Points to Cover
1. **L4 vs L7**: Transport vs application layer, routing capabilities
2. **Algorithms**: Round robin, least connections, consistent hashing—when to use each
3. **Health checks**: Active, passive, thresholds
4. **Sticky sessions**: When needed (stateful), how (cookie, IP hash)
5. **Connection draining**: Graceful removal

### Common Follow-ups
- **"How would you design load balancing for 10M users?"** → Multiple LB tiers, geographic distribution, auto scaling
- **"What if one backend is slow?"** → Least connections, least response time, circuit breaker
- **"How to do zero-downtime deployment?"** → Drain, add new, remove old
- **"L4 vs L7 for your use case?"** → API: L7; raw TCP: L4

### Red Flags to Avoid
- Ignoring health checks
- Not considering sticky sessions for stateful apps
- Single LB instance (no HA)
- Forgetting connection draining
