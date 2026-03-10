# Module 6: Rate Limiting & Load Balancing

---

## 1. Token Bucket

### Definition
A rate limiting algorithm where tokens accumulate at a fixed rate. Each request consumes a token. If the bucket is empty, the request is rejected.

### How It Works
```
Parameters:
  - Bucket capacity: 10 tokens (max burst)
  - Refill rate: 2 tokens/second

Time 0:  Bucket [■■■■■■■■■■] 10 tokens
         5 requests → consume 5
         Bucket [■■■■■░░░░░] 5 tokens

Time 1:  +2 tokens refilled
         Bucket [■■■■■■■░░░] 7 tokens

Time 2:  Burst of 8 requests
         7 pass, 1 REJECTED
         Bucket [░░░░░░░░░░] 0 tokens
```

### Pseudocode
```
class TokenBucket:
    tokens = capacity
    last_refill = now()

    def allow_request():
        elapsed = now() - last_refill
        tokens = min(capacity, tokens + elapsed * refill_rate)
        last_refill = now()

        if tokens >= 1:
            tokens -= 1
            return ALLOW
        return REJECT
```

### Properties
- **Allows bursts** up to bucket capacity
- **Average rate** converges to refill rate
- Memory: **O(1)** per user/key

### Real Systems
AWS API Gateway, Stripe, NGINX, Linux traffic control (tc)

### Summary
Token bucket allows bursts up to a cap while enforcing an average rate. Simple, memory-efficient, and the most widely used rate limiting algorithm.

---

## 2. Leaky Bucket

### Definition
A rate limiter that processes requests at a constant rate regardless of burst. Excess requests queue up; if the queue is full, they're dropped.

### How It Works
```
Incoming requests → Queue (fixed size) → Process at constant rate

  ███ ██ ████ → [Queue: ■■■■■■░░] → ──■──■──■──■── (constant rate)
  (bursty)       (buffer)              (smooth output)

  If queue full: ███ → DROP
```

### Token Bucket vs Leaky Bucket

| | Token Bucket | Leaky Bucket |
|-|---|---|
| Burst handling | Allows bursts | Smooths bursts |
| Output rate | Variable (up to burst) | Constant |
| Queue | No queue | Queue required |
| Best for | API rate limiting | Traffic shaping, disk writes |

### Visual
```
Token Bucket (bursty output):
  In:  ████░░░░████░░░░
  Out: ████░░░░████░░░░  ← bursts pass through

Leaky Bucket (smooth output):
  In:  ████░░░░████░░░░
  Out: ──■─■─■─■─■─■─■─  ← constant rate
```

### Real Systems
NGINX (limit_req with burst), network traffic shaping, Cisco routers

### Summary
Leaky bucket outputs at a constant rate, smoothing bursts. Token bucket allows bursts. Use leaky for traffic shaping, token bucket for API rate limiting.

---

## 3. Sliding Window

### Definition
Rate limiting based on a moving time window, avoiding the boundary problem of fixed windows.

### Fixed Window Problem
```
Window: 100 requests per minute

Time:    |----Minute 1----|----Minute 2----|
Requests: ░░░░░░░░░░██████████████░░░░░░░░░
                    ↑ 100 at end  ↑ 100 at start
                    = 200 requests in 1 minute (2x limit!)
```

### Sliding Window Log
```
Store timestamp of every request in a sorted set.

On new request:
  1. Remove all timestamps older than (now - window)
  2. Count remaining entries
  3. If count < limit → ALLOW, add timestamp
     Else → REJECT

Memory: O(limit) per user — expensive for high limits
```

### Sliding Window Counter (Hybrid)
```
Combine fixed windows with weighted overlap:

Current window: 70% elapsed
Previous window count: 80 requests
Current window count: 30 requests

Estimated rate = 80 × 0.3 + 30 = 54  (under 100 → ALLOW)

Memory: O(1) per user — just two counters!
```

### Comparison

| | Fixed Window | Sliding Log | Sliding Counter |
|-|---|---|---|
| Accuracy | Low (boundary spike) | Perfect | Approximate |
| Memory | O(1) | O(N) | O(1) |
| Implementation | Simple | Complex | Medium |
| Use case | Low stakes | Billing, payments | API rate limiting |

### Real Systems
Redis (sorted sets for sliding log), Cloudflare, Kong

### Summary
Sliding window rate limiting avoids the fixed window boundary spike. Sliding window counter uses weighted overlap of two fixed windows — O(1) memory with good accuracy.

---

## 4. Load Balancing

### Definition
Distributing incoming requests across multiple servers to maximize throughput, minimize latency, and avoid overloading any single server.

### Algorithms
```
┌─── ROUND ROBIN ─────────────────┐
│ Server 1 → 2 → 3 → 1 → 2 → 3  │
│ Simple. Ignores server load.     │
└──────────────────────────────────┘

┌─── WEIGHTED ROUND ROBIN ────────┐
│ Server 1 (weight 3): ■■■        │
│ Server 2 (weight 1): ■          │
│ Server 3 (weight 2): ■■         │
│ More traffic to more capable.    │
└──────────────────────────────────┘

┌─── LEAST CONNECTIONS ───────────┐
│ Send to server with fewest       │
│ active connections. Adaptive.    │
│ Best for varied request latency. │
└──────────────────────────────────┘

┌─── POWER OF TWO CHOICES ────────┐
│ Pick 2 random servers.           │
│ Send to the one with fewer       │
│ connections.                     │
│                                  │
│ Mathematically: exponential      │
│ improvement over random.         │
│ O(1) decision with near-optimal  │
│ balance.                         │
└──────────────────────────────────┘

┌─── CONSISTENT HASHING ──────────┐
│ Hash(request_key) → server       │
│ Same key always hits same server │
│ Good for caching. See Module 1.  │
└──────────────────────────────────┘
```

### Layers
```
L4 (Transport): Route by IP/port. Fast. No request inspection.
  Examples: AWS NLB, HAProxy (TCP mode), IPVS

L7 (Application): Route by URL, headers, cookies. Flexible.
  Examples: AWS ALB, NGINX, Envoy, HAProxy (HTTP mode)
```

### Health Checks
```
Active: LB periodically pings servers (GET /health)
Passive: LB monitors response codes/latency from real traffic
```

### Summary
Load balancers distribute traffic using algorithms from simple (round robin) to smart (least connections, power of two choices). L4 is fast, L7 is flexible.

---

## 5. Consistent Load Balancing

### Definition
Using consistent hashing as a load balancing strategy so that the same client/key always routes to the same server (unless servers change).

### Why It Matters
```
Without consistent LB:
  User A → Server 1 (has cached session)
  User A → Server 3 (cache miss — slow!)

With consistent LB:
  hash(User A) always → Server 1
  Server 1 removed → User A → Server 2 (only A moves)
```

### Use Cases
- **Stateful services**: Session data, in-memory cache
- **Database routing**: Same user always hits same shard
- **CDN**: Same content always cached on same edge node

### Real Systems
Envoy proxy (ring hash LB), Maglev (Google), Discord

### Summary
Consistent load balancing uses hashing to pin keys to servers, maximizing cache hit rates and minimizing disruption during scaling.

---

## 6. Sticky Sessions

### Definition
Ensuring all requests from the same client go to the same backend server, typically via cookies or IP hashing.

### Methods
```
1. COOKIE-BASED: LB sets cookie "server=backend-2"
   Client sends cookie → LB routes to backend-2

2. IP HASH: hash(client_IP) % N → server
   Same IP always hits same server

3. SESSION ID ROUTING: Encode server ID in session token
```

### Tradeoffs

| Pros | Cons |
|------|------|
| Simple session management | Uneven load distribution |
| Works with in-memory sessions | Server death = lost sessions |
| No shared session store needed | Prevents effective auto-scaling |

### Better Alternative
Store sessions in Redis/Memcached → any server can handle any request → true stateless architecture.

### Summary
Sticky sessions route the same client to the same server. Simple but fragile. Prefer externalized session storage (Redis) for production systems.

---

## 7. Service Discovery

### Definition
The mechanism by which microservices find the network locations (IP:port) of other services, even as instances scale up/down.

### Approaches
```
┌─── CLIENT-SIDE DISCOVERY ───────┐
│ Client queries registry directly │
│ Client does load balancing       │
│                                  │
│ [Client] → [Registry] → IP list │
│ [Client] → picks one → [Server] │
│                                  │
│ Examples: Netflix Eureka         │
└──────────────────────────────────┘

┌─── SERVER-SIDE DISCOVERY ───────┐
│ LB queries registry              │
│ Client doesn't know about it     │
│                                  │
│ [Client] → [LB] → [Registry]   │
│                 → [Server]       │
│                                  │
│ Examples: AWS ALB, Kubernetes    │
└──────────────────────────────────┘

┌─── DNS-BASED ───────────────────┐
│ Service registers DNS record     │
│ Client resolves DNS              │
│                                  │
│ user-service.internal → 10.0.1.5│
│                                  │
│ Examples: Consul DNS, CoreDNS    │
└──────────────────────────────────┘
```

### Kubernetes Service Discovery
```
Service "user-svc" creates:
  - ClusterIP: 10.96.0.1 (virtual IP)
  - DNS: user-svc.namespace.svc.cluster.local
  - kube-proxy routes to healthy pods via iptables/IPVS
```

### Real Systems
Consul, Eureka (Netflix), etcd + CoreDNS (Kubernetes), AWS Cloud Map

### Summary
Service discovery lets microservices find each other dynamically. Client-side (Eureka), server-side (K8s), or DNS-based (Consul). Essential for auto-scaling environments.
