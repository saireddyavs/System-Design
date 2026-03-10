# Module 15: Reliability & Operations

---

## 1. Chaos Engineering

### Definition
The discipline of experimenting on production systems to build confidence in their ability to withstand turbulent conditions.

### Netflix Chaos Monkey Family
```
Chaos Monkey:    Randomly kills EC2 instances
Latency Monkey:  Injects artificial delays
Chaos Kong:      Simulates entire AWS region failure
Chaos Gorilla:   Kills an entire Availability Zone
```

### Process
```
1. Define "steady state" (normal system behavior/metrics)
2. Hypothesize: "The system will remain in steady state when X fails"
3. Inject failure: Kill instance, inject latency, partition network
4. Observe: Did the system degrade gracefully?
5. Fix: If not, improve resilience and retest
```

### Principles
- Run in production (staging doesn't catch real issues)
- Minimize blast radius (start small)
- Automate experiments (GameDay exercises)

### Real Systems
Netflix (Chaos Monkey), AWS (FIS - Fault Injection Simulator), Gremlin, LitmusChaos

### Interview Tip
"I'd implement chaos engineering by first establishing SLOs, then running controlled experiments in production — killing instances, injecting latency, simulating region failures."

---

## 2. Blue/Green Deployment

### Definition
Running two identical production environments. Deploy new version to inactive (green) environment, then switch traffic instantly via load balancer.

### How It Works
```
Phase 1: Blue is LIVE
  ┌────────┐     ┌───────────┐
  │ Users  │ ──→ │ LB → Blue │  (v1, live)
  └────────┘     │    Green   │  (v1, idle)
                 └───────────┘

Phase 2: Deploy v2 to Green, test
  ┌────────┐     ┌───────────┐
  │ Users  │ ──→ │ LB → Blue │  (v1, live)
  └────────┘     │    Green   │  (v2, testing)
                 └───────────┘

Phase 3: Switch LB to Green
  ┌────────┐     ┌───────────┐
  │ Users  │ ──→ │ LB → Green│  (v2, live)
  └────────┘     │    Blue    │  (v1, standby/rollback)
                 └───────────┘

Rollback: Switch LB back to Blue (instant!)
```

### Tradeoffs

| Pros | Cons |
|------|------|
| Instant rollback | 2x infrastructure cost |
| Zero downtime | Database migrations are tricky |
| Full testing before go-live | Stateful services need careful handling |

---

## 3. Canary Deployment

### Definition
Gradually rolling out a new version to a small percentage of users, monitoring for issues before expanding to everyone.

### How It Works
```
Phase 1: 1% traffic → v2, 99% → v1
  Monitor: error rate, latency, CPU, user complaints

Phase 2: 10% → v2 (if metrics look good)

Phase 3: 50% → v2

Phase 4: 100% → v2 (full rollout)

At any phase: if metrics degrade → automatic rollback to v1
```

### Blue/Green vs Canary

| | Blue/Green | Canary |
|-|------------|--------|
| Rollout | All-at-once switch | Gradual (1%→10%→100%) |
| Risk | Full exposure on switch | Limited blast radius |
| Rollback | Instant (switch LB) | Instant (route to old) |
| Infrastructure | 2x needed | Can share infra |
| Detection | Post-switch monitoring | Progressive monitoring |

### Real Systems
Kubernetes (Argo Rollouts), AWS CodeDeploy, Spinnaker, Flagger, Google

---

## 4. The C10K Problem

### Definition
The challenge of handling 10,000 concurrent connections on a single server, solved by event-driven I/O instead of thread-per-connection.

### The Problem (Year ~2000)
```
Thread-per-connection (Apache):
  10,000 connections × 1MB stack = 10GB RAM
  10,000 threads × context switch cost = CPU thrashing
  Result: Server crashes at ~1,000 connections

Event-driven (Nginx):
  1 event loop thread handles ALL connections
  epoll/kqueue: O(1) per event notification
  Result: 10,000+ connections on single thread
```

### Evolution
```
C10K (2000s):   10,000 connections → solved by epoll/kqueue
C10M (2010s):   10,000,000 connections → kernel bypass (DPDK)
C100M:          Beyond kernel, specialized hardware

Key OS APIs:
  select():  O(N) scan — doesn't scale
  poll():    O(N) scan — slightly better
  epoll():   O(1) per event — Linux, scales to millions
  kqueue():  O(1) per event — BSD/macOS
  io_uring:  O(1), reduced syscalls — newest Linux
```

### Real Systems
Nginx, Node.js, Redis, Go runtime, HAProxy

---

## 5. CDN (Content Delivery Network)

### Definition
A globally distributed network of servers that cache and serve content from the nearest edge location to the user, reducing latency and origin server load.

### Architecture
```
                    User (Tokyo)
                         │
                    ┌────▼────┐
                    │Tokyo POP│ ← Cache HIT? Serve directly
                    │ (Edge)  │   Cache MISS? Fetch from origin
                    └────┬────┘
                         │ miss
                    ┌────▼────┐
                    │ Shield  │ ← "Origin Shield" (second layer cache)
                    │ (Tokyo) │   Prevents multiple edges hitting origin
                    └────┬────┘
                         │ miss
                    ┌────▼────┐
                    │ Origin  │ ← Your actual server
                    │ Server  │   (US-East)
                    └─────────┘
```

### What CDNs Cache
```
Static:    Images, CSS, JS, fonts, videos     (easy, long TTL)
Dynamic:   API responses, HTML pages           (short TTL, cache keys)
Streaming: Video chunks (HLS/DASH segments)    (CDN essential)
```

### Cache Headers
```
Cache-Control: public, max-age=86400     ← cache for 1 day
Cache-Control: private, no-cache         ← don't cache (user-specific)
Cache-Control: stale-while-revalidate=60 ← serve stale, refresh in background
ETag: "abc123"                           ← conditional revalidation
```

### Impact
90% of traffic served from edge → 10x lower latency, 90% less origin load.

### Real Systems
Cloudflare, Akamai, AWS CloudFront, Fastly, Google Cloud CDN

---

## 6. Health Checks

### Types
```
┌─── SHALLOW (Liveness) ──────────────┐
│ GET /health → 200 OK                │
│ "Is the process alive?"             │
│ Fast, lightweight                    │
│ K8s: livenessProbe                  │
└──────────────────────────────────────┘

┌─── DEEP (Readiness) ───────────────┐
│ GET /ready → checks DB, cache, deps│
│ "Can the service handle requests?" │
│ Slower, may fail if deps are down   │
│ K8s: readinessProbe                │
└──────────────────────────────────────┘
```

### Danger: Deep Health Check Cascading Failure
```
DB slows down (high load)
  → All web servers' deep health checks fail
  → LB removes ALL web servers from rotation
  → TOTAL OUTAGE (even though web servers are fine!)

Solution:
  - Separate liveness (am I alive?) from readiness (can I serve?)
  - Don't fail liveness on dependency issues
  - Use circuit breakers on dependency checks
```

### Kubernetes Health Probes
```yaml
livenessProbe:      # Restart container if fails
  httpGet:
    path: /health
    port: 8080
  periodSeconds: 10
  failureThreshold: 3

readinessProbe:     # Remove from Service (no traffic) if fails
  httpGet:
    path: /ready
    port: 8080
  periodSeconds: 5

startupProbe:       # Give slow-starting apps time to boot
  httpGet:
    path: /health
    port: 8080
  failureThreshold: 30
  periodSeconds: 10
```

---

## 7. Graceful Degradation

### Definition
When a system component fails, deliberately reduce functionality instead of crashing entirely. Users get a degraded but functional experience.

### Examples
```
┌─── Amazon ──────────────────────────────────────┐
│ Recommendation engine down?                      │
│ → Show "Trending items" instead of personalized  │
│ User barely notices                              │
└──────────────────────────────────────────────────┘

┌─── Netflix ─────────────────────────────────────┐
│ Personalization service down?                    │
│ → Show generic "Top 10" instead of "For You"     │
│ → Still show cached recommendations              │
└──────────────────────────────────────────────────┘

┌─── Google Search ───────────────────────────────┐
│ Spell check service down?                        │
│ → Show results without "Did you mean..."         │
│ → Core search still works                        │
└──────────────────────────────────────────────────┘
```

### Implementation
```
try:
    recommendations = recommendationService.get(userId)
except (Timeout, CircuitOpen):
    recommendations = cache.get("trending_items")  # fallback
    metrics.increment("recommendation_fallback")
```

### Key Principle
Every dependency should have a fallback. No single dependency failure should bring down the entire user experience.

---

## 8. Shuffle Sharding

### Definition
Assigning each customer to a random subset of servers. The probability of two customers sharing the exact same subset is extremely low, limiting blast radius.

### How It Works
```
Traditional sharding (1 shard per customer):
  Customer A, B, C → Shard 1
  Customer D, E, F → Shard 2

  Customer A attacks Shard 1 → B and C also affected!

Shuffle sharding (pick 2 of 8 servers per customer):
  Customer A → {Server 1, Server 5}
  Customer B → {Server 3, Server 7}
  Customer C → {Server 2, Server 5}

  Customer A attacks → only Servers 1, 5 affected
  Customer B: {3, 7} → completely unaffected!

  Probability both share same 2 servers = 1/C(8,2) = 1/28 = 3.6%
  With 2 of 100 servers: 1/C(100,2) = 1/4950 = 0.02%
```

### Visual
```
Servers: [1] [2] [3] [4] [5] [6] [7] [8]

Customer A: [1]          [5]           ← these 2
Customer B:      [3]               [7] ← these 2 (no overlap!)
Customer C:   [2]        [5]          ← shares Server 5 with A

A attacks: Server 1, 5 degraded
B: unaffected (different servers)
C: partially affected (shares 5, but 2 is fine)
```

### Real Systems
AWS Route 53 (DNS), AWS Lambda (isolation), Amazon internal services

### Summary
Shuffle sharding assigns customers to random server subsets. The combinatorial explosion of possible subsets means most customers share zero servers, drastically limiting noisy-neighbor impact.
