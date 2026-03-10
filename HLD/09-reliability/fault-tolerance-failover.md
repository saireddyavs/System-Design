# Fault Tolerance & Failover

> Staff+ Engineer Level | FAANG Interview Deep Dive

---

## 1. Concept Overview

**Fault tolerance** is the ability of a system to continue operating correctly in the presence of failures. **Failover** is the automatic or manual switching to a redundant or standby system when the primary system fails. Together, they form the backbone of highly available distributed systems.

### Key Definitions

| Term | Definition |
|------|------------|
| **Fault** | A defect or deviation that may cause a failure |
| **Error** | Incorrect internal state that may lead to failure |
| **Failure** | System deviation from specified behavior |
| **Fault Tolerance** | System's ability to continue despite faults |
| **Failover** | Process of switching to backup when primary fails |
| **Failback** | Process of returning to primary after recovery |

### The Fault-Failure Chain

```
Fault → Error → Failure → Degradation/Outage
```

---

## 2. Real-World Motivation

### Why Fault Tolerance Matters

- **Financial impact**: Amazon loses ~$66,000 per minute of downtime; Gartner estimates $5,600/minute average
- **User trust**: Netflix, GitHub, AWS outages make headlines and erode confidence
- **Regulatory**: SLAs often mandate 99.9%+ availability (8.76 hours downtime/year max)
- **Cascading failures**: Single component failure can take down entire systems (e.g., AWS us-east-1 2017 S3 outage)

### Notable Incidents

| Company | Year | Cause | Impact |
|---------|------|-------|--------|
| AWS S3 | 2017 | Human error during capacity expansion | 4+ hours, cascaded to many services |
| GitHub | 2018 | Database failover, split-brain risk | 24+ hours partial outage |
| Netflix | 2012 | AWS region failure | Prompted Chaos Engineering investment |
| Google Cloud | 2019 | Network configuration error | Multi-region impact |

---

## 3. Architecture Diagrams

### Active-Active Architecture

```
                    ┌─────────────────────────────────────────────────┐
                    │              Global Load Balancer                 │
                    │         (Route 53 / Cloudflare / GSLB)            │
                    └─────────────────────┬───────────────────────────┘
                                          │
                    ┌─────────────────────┼─────────────────────┐
                    │                     │                     │
                    ▼                     ▼                     ▼
            ┌───────────────┐     ┌───────────────┐     ┌───────────────┐
            │   Region A    │     │   Region B    │     │   Region C    │
            │  (us-east-1)  │     │  (us-west-2)  │     │  (eu-west-1)  │
            └───────┬───────┘     └───────┬───────┘     └───────┬───────┘
                    │                     │                     │
            ┌───────▼───────┐     ┌───────▼───────┐     ┌───────▼───────┐
            │  App Servers  │     │  App Servers  │     │  App Servers  │
            │  (Active)     │     │  (Active)     │     │  (Active)     │
            └───────┬───────┘     └───────┬───────┘     └───────┬───────┘
                    │                     │                     │
            ┌───────▼───────┐     ┌───────▼───────┐     ┌───────▼───────┐
            │  DB Primary   │◄────│  Replication  │◄────│  Replication  │
            │  (Region A)   │     │  (async)      │     │  (async)      │
            └───────────────┘     └───────────────┘     └───────────────┘
```

### Active-Passive Architecture

```
                    ┌─────────────────────────────────────────────────┐
                    │              Load Balancer / DNS                  │
                    └─────────────────────┬───────────────────────────┘
                                          │
                                          ▼
            ┌─────────────────────────────────────────────────────────┐
            │                    PRIMARY SITE (Active)                 │
            │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐   │
            │  │ App Server 1│  │ App Server 2│  │  Primary DB     │   │
            │  └─────────────┘  └─────────────┘  └────────┬────────┘   │
            └─────────────────────────────────────────────┼────────────┘
                                                          │
                                              Replication │ (sync/async)
                                                          │
            ┌─────────────────────────────────────────────┼────────────┐
            │                    STANDBY SITE (Passive)    │            │
            │  ┌─────────────┐  ┌─────────────┐  ┌────────▼────────┐   │
            │  │ App Server 1│  │ App Server 2│  │  Standby DB     │   │
            │  │ (idle/cold) │  │ (idle/cold) │  │  (replica)      │   │
            │  └─────────────┘  └─────────────┘  └─────────────────┘   │
            └─────────────────────────────────────────────────────────┘
```

### Failover Sequence Diagram

```
  Client    LB/Proxy    Primary    Standby    Health Check
    │          │           │          │            │
    │  request │           │          │            │
    │─────────►│──────────►│          │            │
    │          │           │          │            │
    │          │           │          │    heartbeat
    │          │           │          │◄───────────│
    │          │           │          │   (OK)     │
    │          │           │          │            │
    │          │     [PRIMARY FAILS]  │            │
    │          │           X          │            │
    │          │           │          │   heartbeat
    │          │           │          │◄───────────│
    │          │           │          │   (FAIL)   │
    │          │           │          │            │
    │          │     [FAILOVER TRIGGERED]          │
    │          │           │          │            │
    │          │     promote standby  │            │
    │          │           │─────────►│            │
    │          │           │          │            │
    │          │     [STANDBY BECOMES PRIMARY]    │
    │          │           │          │            │
    │  request │           │          │            │
    │─────────►│─────────────────────►│            │
    │          │           │          │            │
    │  response│           │          │            │
    │◄─────────│◄────────────────────│            │
```

---

## 4. Core Mechanics

### Types of Faults

| Fault Type | Duration | Examples | Mitigation |
|------------|----------|----------|------------|
| **Transient** | Seconds | Network blip, brief CPU spike | Retry, timeout |
| **Intermittent** | Recurring | Flaky disk, memory leak | Restart, replace |
| **Permanent** | Until fix | Hardware failure, corrupted data | Failover, replacement |

### Fault Tolerance Techniques

#### Redundancy Patterns

| Pattern | Description | Use Case |
|---------|-------------|----------|
| **Active-Active** | All replicas serve traffic | Stateless services, read replicas |
| **Active-Passive** | Standby waits for failover | Databases, stateful services |
| **N+1** | N required + 1 spare | Capacity for single failure |
| **N+M** | N required + M spare | Higher availability |

### Failover Types

| Type | Cold Standby | Warm Standby | Hot Standby |
|------|--------------|--------------|-------------|
| **Resources** | Minimal/none | Scaled-down | Full capacity |
| **Data** | Restore from backup | Replicated | Real-time sync |
| **Failover Time** | Minutes-hours | Seconds-minutes | Seconds |
| **Cost** | Lowest | Medium | Highest |
| **Example** | DR site with backups | Scaled-down replica | Active-active DB |

---

## 5. Numbers

### Typical Failover Times

| Component | Failover Mechanism | Typical Time |
|-----------|---------------------|--------------|
| **DNS** | TTL-based, manual change | 1-15 minutes (TTL dependent) |
| **Load Balancer** | Health check failure | 5-30 seconds |
| **Application** | LB health check | 10-60 seconds |
| **Database** | Replication promotion | 30 seconds - 5 minutes |
| **Cache** | Replica promotion | 1-10 seconds |
| **Full Region** | Multi-region failover | 1-15 minutes |

### Availability Math

| SLA | Downtime/Year | Downtime/Month |
|-----|---------------|----------------|
| 99% | 3.65 days | 7.2 hours |
| 99.9% | 8.76 hours | 43.2 minutes |
| 99.95% | 4.38 hours | 21.6 minutes |
| 99.99% | 52.6 minutes | 4.32 minutes |
| 99.999% | 5.26 minutes | 26.3 seconds |

---

## 6. Tradeoffs (Comparison Tables)

### Active-Active vs Active-Passive

| Dimension | Active-Active | Active-Passive |
|-----------|---------------|----------------|
| **Cost** | Higher (all resources active) | Lower (standby can be scaled down) |
| **Failover Time** | Near-instant (traffic rerouted) | Seconds to minutes |
| **Complexity** | High (data consistency, conflict resolution) | Lower |
| **Data Consistency** | Hard (eventual consistency) | Easier (single source of truth) |
| **Utilization** | High | Lower (standby idle) |
| **Best For** | Stateless, read-heavy | Stateful, write-heavy |

### Health Check Depth

| Type | What It Checks | Latency | Use Case |
|------|----------------|---------|----------|
| **Shallow** | Process alive, port open | <1ms | Liveness |
| **Deep** | Dependency connectivity | 10-100ms | Readiness |
| **Dependency** | Full request path | 100ms-1s | Full stack |

---

## 7. Variants/Implementations

### Health Check Mechanisms

```text
Shallow:  TCP connect → OK/FAIL
Deep:    TCP + HTTP GET /health → 200 OK
Full:    HTTP GET /health → DB query → Cache ping → 200 OK
```

### Heartbeat Configurations

| Parameter | Typical Value | Purpose |
|-----------|---------------|---------|
| Interval | 1-5 seconds | Detection speed vs overhead |
| Timeout | 3-5x interval | Avoid false positives |
| History | 3-5 consecutive | Confirm failure |

### Split-Brain Prevention

**Split-brain**: Both nodes believe they're primary after network partition.

| Technique | How It Works | Tradeoff |
|-----------|--------------|----------|
| **Quorum** | Majority must agree (e.g., 3 nodes, 2 must agree) | Requires odd N, adds latency |
| **STONITH** | Shoot The Other Node In The Head — fence failed node | Violent but effective |
| **Witness** | Third-party arbiter (cloud service) | External dependency |
| **Tie-breaker** | Shared disk, lock service | SPOF risk |

---

## 8. Scaling Strategies

### Horizontal Scaling for Fault Tolerance

- **Stateless design**: Any instance can serve any request
- **Session externalization**: Redis/Memcached for session state
- **Request affinity**: Optional for performance, not for correctness

### Graceful Degradation

| Technique | Description | Example |
|-----------|-------------|---------|
| **Feature Flags** | Disable non-critical features | Turn off recommendations |
| **Load Shedding** | Reject excess load | 503 for low-priority requests |
| **Circuit Breaker** | Stop calling failing dependency | Skip cache, use DB directly |
| **Degraded Mode** | Reduced functionality | Read-only mode |

---

## 9. Failure Scenarios

### Cascading Failure

```
DB slow → App threads blocked → Thread pool exhausted → App unresponsive
    → LB marks unhealthy → All instances removed → Total outage
```

**Prevention**: Timeouts, circuit breakers, bulkheads, connection limits.

### Retry Storm

```
Service A fails → B retries → B overloaded → B fails → C retries → ...
```

**Prevention**: Exponential backoff, retry budgets, circuit breakers.

### Split-Brain (Database)

```
Network partition: Primary | Standby
Both think they're primary → Data divergence → Corruption on heal
```

**Prevention**: Quorum, STONITH, fencing, witness node.

### Thundering Herd

```
Cache expires → 1000 requests hit DB simultaneously → DB overload
```

**Prevention**: Probabilistic early expiration, request coalescing, cache stampede protection.

---

## 10. Performance Considerations

### Failover Performance Impact

- **Connection draining**: Allow in-flight requests to complete (30-60s typical)
- **Session migration**: Stateful failover may drop sessions
- **Cache warming**: Cold standby has cold cache (higher latency initially)
- **Connection pool**: New connections to new primary add latency

### Health Check Overhead

- Shallow checks: Negligible
- Deep checks: 1-5% CPU if too frequent
- **Recommendation**: Liveness 10s, Readiness 1-5s, separate endpoints

---

## 11. Use Cases

| Use Case | Recommended Pattern | Rationale |
|----------|---------------------|-----------|
| Web API | Active-Active, multi-AZ | Stateless, high availability |
| Database | Active-Passive, sync replica | Consistency over availability |
| Cache | Active-Active, replica | Cache can rebuild |
| Message Queue | Active-Passive, mirroring | Message durability critical |
| File Storage | Active-Active, eventual consistency | S3-style object storage |

---

## 12. Deployment Strategies

### Blue-Green Deployment

```
        ┌─────────┐
        │  Router │
        └────┬────┘
             │
     ┌───────┴───────┐
     │               │
     ▼               ▼
┌─────────┐    ┌─────────┐
│  Blue   │    │  Green  │
│ (v1.0)  │    │ (v1.1)  │
│ ACTIVE  │    │ IDLE    │
└─────────┘    └─────────┘

Switch: Router → Green (instant)
```

### Canary Deployment

```
        ┌─────────┐
        │  Router │
        └────┬────┘
             │
     ┌───────┼───────┬─────────┐
     │       │       │         │
     ▼       ▼       ▼         ▼
  90% v1   5% v2   5% v2    (metrics)
```

### Rolling Update

```
Before: [v1][v1][v1][v1]
Step 1: [v2][v1][v1][v1]
Step 2: [v2][v2][v1][v1]
Step 3: [v2][v2][v2][v1]
After:  [v2][v2][v2][v2]
```

### Comparison

| Strategy | Risk | Rollback | Cost | Use Case |
|----------|------|----------|------|----------|
| Blue-Green | Low | Instant | 2x resources | Critical releases |
| Canary | Very low | Gradual | 1x + small | High-traffic |
| Rolling | Medium | Slow | 1x | Standard updates |

---

## 13. Chaos Engineering

### Netflix Chaos Monkey

- Randomly terminates production instances
- Verifies system survives single instance failure
- **Principle**: If you can't handle one instance dying, you have a problem

### Simian Army (Netflix)

| Tool | Purpose |
|------|----------|
| Chaos Monkey | Terminate instances |
| Chaos Gorilla | Kill entire AZ |
| Chaos Kong | Kill entire region |
| Latency Monkey | Introduce network delay |
| Conformity Monkey | Find non-compliant instances |

### Gremlin

- Commercial chaos engineering platform
- Controlled, scheduled chaos experiments
- Scenarios: CPU, memory, disk, network, time

### Chaos Engineering Principles

1. **Build hypothesis** around steady state
2. **Vary real-world events** (not just instance kill)
3. **Run in production** (or production-like)
4. **Automate** experiments
5. **Minimize blast radius**

---

## 14. Real-World Examples

### AWS Multi-AZ

- RDS: Synchronous replication to standby in different AZ
- Failover: 60-120 seconds typical
- DNS/endpoint unchanged (managed by AWS)

### Google Spanner

- Globally distributed, strongly consistent
- Multi-region replication
- Failover: Regional failure handled by routing

### Netflix Region Failover

- Multi-region active-active
- Zuul for routing, Eureka for discovery
- Can lose entire region, traffic shifts

### GitHub Incident Response

- 2018: Database failover revealed split-brain risk
- Now: Multiple replicas, automated failover, extensive testing
- Runbooks for every failure scenario

---

## 15. Code/Pseudocode

### Health Check Endpoint (Pseudocode)

```python
@app.get("/health/live")
def liveness():
    """Shallow: Process is running"""
    return {"status": "ok"}

@app.get("/health/ready")
async def readiness():
    """Deep: Can serve traffic"""
    checks = {
        "db": await check_db_connection(),
        "cache": await check_cache_connection(),
        "disk": check_disk_space() > 10%
    }
    if all(checks.values()):
        return {"status": "ready", "checks": checks}
    return {"status": "not_ready", "checks": checks}, 503
```

### Failover Decision Logic

```python
def should_failover(primary_status, standby_status, quorum):
    if primary_status == HEALTHY:
        return False
    if standby_status != HEALTHY:
        return False  # Don't failover to unhealthy standby
    if not quorum_reached(quorum):
        return False  # Split-brain prevention
    return True
```

### Circuit Breaker State Machine

```python
class CircuitBreaker:
    states = [CLOSED, OPEN, HALF_OPEN]
    
    def on_success(self):
        if self.state == HALF_OPEN:
            self.state = CLOSED
            self.failures = 0
    
    def on_failure(self):
        self.failures += 1
        if self.failures >= self.threshold:
            self.state = OPEN
            self.open_until = now() + self.timeout
    
    def can_execute(self):
        if self.state == CLOSED:
            return True
        if self.state == OPEN and now() > self.open_until:
            self.state = HALF_OPEN
            return True
        return False
```

---

## 16. Interview Discussion

### Key Talking Points

1. **Start with requirements**: What's the availability target? What can we afford to lose?
2. **Fault taxonomy**: Transient vs permanent drives different strategies
3. **Failover is not free**: Tradeoffs in cost, complexity, consistency
4. **Split-brain is deadly**: Always discuss quorum/fencing for stateful systems
5. **Chaos engineering**: Prove your design, don't assume

### Common Questions

**Q: How would you design failover for a database?**
- Sync vs async replication tradeoff (RPO vs latency)
- Quorum for split-brain (e.g., 3 nodes, 2 must agree)
- STONITH for fencing failed primary
- Automated vs manual failover (blast radius consideration)

**Q: What's the difference between active-active and active-passive?**
- Active-active: All serve traffic, better utilization, harder consistency
- Active-passive: Standby waits, simpler, lower cost, slower failover

**Q: How do you prevent split-brain?**
- Quorum (majority vote)
- Fencing (STONITH)
- Witness/tie-breaker
- Consensus (Paxos, Raft)

**Q: When would you use blue-green vs canary?**
- Blue-green: Low-risk, instant rollback, can afford 2x resources
- Canary: High-traffic, want gradual validation, minimize blast radius
