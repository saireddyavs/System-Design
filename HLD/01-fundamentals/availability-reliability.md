# Availability & Reliability: Staff+ Engineer Deep Dive

## 1. Concept Overview

### Definition
**Availability** is the proportion of time a system is operational and able to perform its required function. It's measured as a percentage (e.g., 99.9%) or "nines."

**Reliability** is the probability that a system will perform its required function under stated conditions for a specified period of time without failure. It encompasses correctness, consistency, and durability—not just uptime.

### Purpose
- **User trust**: Downtime loses revenue, users, and reputation
- **SLA compliance**: Contractual obligations to customers
- **Operational excellence**: Proactive vs reactive incident management
- **Business continuity**: Critical systems must stay up (payment, healthcare, etc.)

### Why It Exists
- **Cost of downtime**: Amazon loses ~$66,240 per minute of outage; Fortune 1000 average $5,600/minute
- **Competitive pressure**: Users expect 24/7 availability (Netflix, Uber, banking)
- **Regulatory requirements**: Financial services (99.9%+), healthcare (HIPAA)
- **Distributed systems**: More components = more failure modes; must design for failure

### Problems It Solves
1. **Single points of failure (SPOF)**: One component failure takes down entire system
2. **Unplanned downtime**: Hardware failures, bugs, deployments
3. **Cascading failures**: One failure triggers others
4. **Slow recovery**: MTTR (Mean Time To Recovery) too high
5. **Unpredictable behavior**: System works until it doesn't

---

## 2. Real-World Motivation

### Google's Targets
- **Search, Gmail, YouTube**: 99.99% availability (52 minutes downtime/year)
- **Google Cloud**: 99.95% for most services (SLA with financial credits)
- **Approach**: Multi-region, automatic failover, chaos engineering (DiRT drills)

### Amazon
- **Prime Day**: Cannot afford downtime; billions in sales
- **AWS SLA**: 99.99% for EC2, S3; 99.95% for RDS
- **Availability Zones**: Each region has 3+ AZs; design for AZ failure
- **Well-Architected Framework**: Reliability pillar is first

### Netflix Chaos Monkey
- **Purpose**: Randomly terminates production instances to test resilience
- **Philosophy**: "The best way to avoid failure is to fail constantly"
- **Result**: System designed to withstand any single component failure
- **Chaos Engineering**: Simian Army (Chaos Kong, Chaos Gorilla for larger failures)

### Stripe
- **Payment processing**: 99.99%+ required; downtime = lost transactions
- **Idempotency**: Prevents duplicate charges on retries
- **Multi-region**: Active-active for critical path

### Uber
- **Real-time**: ETA, matching, payments must work
- **Circuit breakers**: Isolate failing services
- **Graceful degradation**: Show cached data if live data unavailable

---

## 3. Architecture Diagrams

### Availability "Nines" Visualization

```
AVAILABILITY LEVELS (per year)
==============================
99%      (Two nines)     ████████████████████ 3.65 days downtime
99.9%    (Three nines)   ████                  8.76 hours
99.95%   (3.5 nines)     ██                    4.38 hours  
99.99%   (Four nines)    █                     52.56 minutes
99.999%  (Five nines)    ▌                     5.26 minutes
99.9999% (Six nines)     ▏                     31.5 seconds
```

### Single Point of Failure (SPOF)

```
SYSTEM WITH SPOF
================
                    ┌─────────────┐
    All traffic ───►│ Load Balancer│◄─── SPOF! (single LB)
                    └──────┬──────┘
                           │
                    ┌──────┴──────┐
                    │  Database   │◄─── SPOF! (single DB)
                    └─────────────┘

RESILIENT SYSTEM (No SPOF)
==========================
                    ┌─────────────┐
                    │     LB 1    │
    Traffic ───────►├─────────────┤◄─── Redundant LBs
                    │     LB 2    │
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
        ┌─────────┐  ┌─────────┐  ┌─────────┐
        │ Primary │  │ Replica  │  │ Replica │◄─── DB redundancy
        │    DB   │  │    1     │  │    2    │
        └─────────┘  └─────────┘  └─────────┘
```

### Active-Passive Failover

```
ACTIVE-PASSIVE FAILOVER
=======================
Normal operation:
                    ┌─────────────┐
    Traffic ───────►│   Active    │─────► Response
                    │   (Primary) │
                    └──────┬──────┘
                           │ Sync (replication)
                    ┌──────┴──────┐
                    │   Passive   │  (Standby, no traffic)
                    │  (Secondary)│
                    └─────────────┘

After failover (Active fails):
                    ┌─────────────┐
                    │   Active    │  (Failed)
                    │     X       │
                    └─────────────┘
                           │
                    ┌──────┴──────┐
    Traffic ───────►│   Passive   │─────► Now Active
                    │  (Promoted) │
                    └─────────────┘
```

### Active-Active Architecture

```
ACTIVE-ACTIVE (Multi-Region)
============================
    Region US-East              Region EU-West
    ┌─────────────────┐        ┌─────────────────┐
    │  App Servers    │        │  App Servers    │
    │  (Active)       │◄──────►│  (Active)       │
    └────────┬────────┘  Sync  └────────┬────────┘
             │                          │
             ▼                          ▼
    ┌─────────────────┐        ┌─────────────────┐
    │  DB Primary     │◄──────►│  DB Primary    │
    │  (US)          │ Repl.  │  (EU)          │
    └─────────────────┘        └─────────────────┘
             │                          │
             └──────────┬────────────────┘
                        │
                 Global Load Balancer
                 (Geo-routing, health checks)
```

### Redundancy Types

```
HARDWARE REDUNDANCY
===================
┌─────────────────────────────────────────┐
│  Server with redundant components:       │
│  - Dual power supplies                   │
│  - RAID (redundant disks)                │
│  - Bonded NICs (redundant network)       │
│  - ECC RAM (error correction)            │
└─────────────────────────────────────────┘

SOFTWARE REDUNDANCY
===================
┌─────────┐ ┌─────────┐ ┌─────────┐
│ Instance│ │ Instance│ │ Instance│  Multiple instances
│    1    │ │    2    │ │    N    │  (any can fail)
└─────────┘ └─────────┘ └─────────┘

DATA REDUNDANCY
===============
Primary ──replication──► Replica 1
     └──replication──► Replica 2 (cross-region)
```

---

## 4. Core Mechanics

### Availability Calculation

**Basic formula:**
```
Availability = Uptime / (Uptime + Downtime)
            = MTBF / (MTBF + MTTR)

Where:
MTBF = Mean Time Between Failures
MTTR = Mean Time To Recovery (or Repair)
```

**Example:**
- MTBF = 720 hours (30 days)
- MTTR = 1 hour
- Availability = 720 / (720 + 1) = 99.86%

### Serial vs Parallel Components

**Serial (components in sequence):**
```
A ──► B ──► C
All must work for system to work.
Availability = Avail(A) × Avail(B) × Avail(C)
```
Example: A=99%, B=99%, C=99% → System = 97.03%

**Parallel (redundant components):**
```
     ┌─ A ─┐
───► │     ├──► System works if ANY works
     └─ B ─┘
Availability = 1 - (1-Avail(A)) × (1-Avail(B))
```
Example: A=99%, B=99% → System = 1 - (0.01 × 0.01) = 99.99%

### SLA, SLO, SLI Hierarchy

```
SLA (Service Level Agreement)
├── Contract with customers
├── Includes consequences (credits, penalties)
└── Example: "99.9% uptime or 10% monthly credit"

SLO (Service Level Objective)
├── Internal target (stricter than SLA)
├── What we promise ourselves
└── Example: "99.95% uptime" (buffer above SLA)

SLI (Service Level Indicator)
├── Actual measurement
├── Raw metric
└── Example: "Successful requests / Total requests"
```

### Error Budget

```
Error Budget = 1 - SLO
Example: 99.9% SLO → 0.1% error budget = 43.2 minutes/month

Use: Can we deploy? Is reliability improving?
- Under budget: Can take risks, deploy faster
- Over budget: Freeze releases, focus on reliability
```

---

## 5. Numbers

### Downtime by Availability Level

| Nines | Availability | Downtime/Year | Downtime/Month |
|-------|--------------|---------------|----------------|
| 2 | 99% | 3.65 days | 7.2 hours |
| 3 | 99.9% | 8.76 hours | 43.8 min |
| 4 | 99.99% | 52.56 min | 4.38 min |
| 5 | 99.999% | 5.26 min | 26.3 sec |
| 6 | 99.9999% | 31.5 sec | 2.6 sec |

### Industry Benchmarks

| Company/Service | Target | Actual (typical) |
|-----------------|--------|------------------|
| Google Search | 99.99% | ~99.99% |
| AWS EC2 | 99.99% (SLA) | ~99.99% |
| Netflix | 99.99% | ~99.99% |
| Stripe | 99.99%+ | ~99.99% |
| Slack | 99.99% | ~99.9% (historical incidents) |
| GitHub | 99.95% (SLA) | ~99.9% |

### Cost of Downtime (Approximate)

| Industry | Cost per minute |
|----------|-----------------|
| Amazon | $66,240 |
| Apple | $25,000 |
| Google | $100,000+ |
| Average (Fortune 1000) | $5,600 |
| Small business | $427 |

### MTTR Targets

| Severity | Target MTTR | Example |
|----------|-------------|---------|
| P0 (Critical) | < 15 min | Site down |
| P1 (High) | < 1 hour | Major feature broken |
| P2 (Medium) | < 4 hours | Degraded performance |
| P3 (Low) | < 24 hours | Minor issues |

---

## 6. Tradeoffs

### Active-Passive vs Active-Active

| Aspect | Active-Passive | Active-Active |
|--------|----------------|---------------|
| **Cost** | Lower (standby idle) | Higher (2x resources) |
| **Failover time** | Minutes (promotion) | Seconds (instant) |
| **Complexity** | Lower | Higher (sync, conflict) |
| **Resource utilization** | 50% (standby unused) | 100% |
| **Data consistency** | Simpler | Complex (multi-write) |
| **Use case** | DR, databases | Stateless, read-heavy |

### Redundancy Level vs Cost

| Redundancy | Availability | Cost | Use Case |
|------------|--------------|------|----------|
| None | 95-99% | 1x | Dev, MVP |
| N+1 | 99.9% | 1.2x | Production |
| N+2 | 99.99% | 1.4x | Critical |
| Multi-region | 99.99%+ | 2-3x | Global, compliance |

### Consistency vs Availability (CAP)

| Choice | Availability | Consistency | Example |
|--------|---------------|-------------|---------|
| CP | Lower during partition | Strong | ZooKeeper, etcd |
| AP | Higher | Eventual | Cassandra, DynamoDB |
| CA | N/A (no partition) | Strong | Single-node DB |

---

## 7. Variants / Implementations

### Failover Types

1. **Manual failover**: Human triggers; slow but controlled
2. **Automatic failover**: Health check triggers; fast but can flapping
3. **Semi-automatic**: System recommends, human approves
4. **Zero-downtime**: Rolling deployment, blue-green

### Redundancy Patterns

1. **Hot standby**: Passive replica always synced, ready to promote
2. **Warm standby**: Replica exists but may need sync/catch-up
3. **Cold standby**: Backup exists, significant restore time
4. **Multi-active**: All replicas serve traffic

### Fault Tolerance Techniques

| Technique | Description | Example |
|-----------|-------------|---------|
| **Replication** | Copy data across nodes | DB replicas |
| **Retry** | Retry failed operations | Exponential backoff |
| **Circuit breaker** | Stop calling failing service | Hystrix, Resilience4j |
| **Bulkhead** | Isolate failure domains | Thread pools per service |
| **Timeout** | Fail fast, don't hang | 30s request timeout |
| **Graceful degradation** | Reduce functionality | Show cached data |
| **Fallback** | Alternative when primary fails | Static page, default value |

---

## 8. Scaling Strategies (for Reliability)

### Eliminating SPOFs
1. **Load balancers**: Multiple LBs (e.g., AWS NLB with multiple AZs)
2. **Databases**: Primary + replicas, or distributed (Cassandra)
3. **Application**: Multiple instances across AZs
4. **DNS**: Multiple DNS providers (Route53 + Cloudflare)

### Multi-AZ Deployment

```
AWS Region
├── AZ-1 (us-east-1a)
│   ├── App instances
│   ├── DB primary
│   └── Cache nodes
├── AZ-2 (us-east-1b)
│   ├── App instances
│   ├── DB replica
│   └── Cache nodes
└── AZ-3 (us-east-1c)
    ├── App instances
    └── Cache nodes
```

### Chaos Engineering Maturity

| Level | Practice | Example |
|-------|----------|---------|
| 1 | Manual testing | Restart service, check |
| 2 | Chaos Monkey | Random instance termination |
| 3 | Game days | Planned failure injection |
| 4 | Continuous | Automated chaos in staging |
| 5 | Production | Controlled production chaos |

---

## 9. Failure Scenarios

### Real Production Failures

**AWS us-east-1 Outage (2017)**
- **Cause**: S3 team ran command that removed more servers than intended
- **Impact**: 4 hours, affected S3, Lambda, many dependent services
- **Lesson**: Blast radius containment, automation, runbooks

**GitHub 24-hour Outage (2018)**
- **Cause**: Database failover, then split-brain (two primaries)
- **Impact**: Data inconsistency, 24 hours to resolve
- **Lesson**: Failover procedures, consistency verification

**Google Cloud Multi-Region Outage (2019)**
- **Cause**: Configuration change caused capacity exhaustion
- **Impact**: Multiple regions affected
- **Lesson**: Change management, canary deployments

**Facebook/Meta 6-Hour Outage (2021)**
- **Cause**: BGP configuration error; entire network unreachable
- **Impact**: Facebook, Instagram, WhatsApp down globally
- **Lesson**: Network-level redundancy, automation risks

### Mitigation Strategies

| Failure Type | Mitigation |
|--------------|------------|
| **Hardware failure** | Multi-AZ, replication |
| **Software bug** | Canary, feature flags, rollback |
| **Traffic spike** | Auto-scaling, rate limiting |
| **Dependency failure** | Circuit breaker, fallback |
| **Data corruption** | Backups, checksums, audit |
| **Human error** | Automation, approval gates |
| **Network partition** | Multi-path, retry, timeout |

---

## 10. Performance Considerations

### Impact of Redundancy on Performance
- **Replication lag**: Read replicas may be seconds behind
- **Cross-AZ latency**: 1-5ms additional per AZ hop
- **Health check overhead**: Frequent checks add load
- **Failover latency**: 30s-5min typical for DB failover

### Monitoring for Reliability
- **Uptime checks**: External probes from multiple locations
- **Synthetic monitoring**: Simulate user flows
- **Error rate**: 4xx, 5xx, timeouts
- **Latency percentiles**: p50, p95, p99
- **Dependency health**: Downstream service status

### Runbook Requirements
- **Clear steps**: Numbered, unambiguous
- **Decision trees**: If X, do Y
- **Contact info**: Escalation paths
- **Tested**: Run in drills, kept current

---

## 11. Use Cases

| System | Availability Need | Approach |
|--------|-------------------|----------|
| **YouTube** | 99.99% | Multi-region, CDN, stateless |
| **Uber** | 99.99% | Multi-AZ, circuit breakers, graceful degradation |
| **Netflix** | 99.99% | Chaos Engineering, redundancy, regional |
| **WhatsApp** | 99.9%+ | Erlang reliability, distributed |
| **Stripe** | 99.99%+ | Idempotency, multi-region, strong consistency |
| **AWS** | 99.99% SLA | Multi-AZ, automated failover |
| **Banking** | 99.99%+ | Regulatory, active-passive, audit |

---

## 12. Comparison Tables

### Availability Calculation Examples

| Components | Config | Calculation | Result |
|------------|--------|-------------|--------|
| 3 serial @ 99% | A→B→C | 0.99³ | 97.03% |
| 2 parallel @ 99% | A∥B | 1-(0.01)² | 99.99% |
| 2+1 (2 active, 1 standby) | A∥B, C standby | Complex | ~99.99% |
| 3 AZ @ 99.9% each | Any 2 of 3 | Binomial | ~99.999% |

### Failover Comparison

| Type | Detection | Failover Time | Data Loss Risk |
|------|------------|---------------|----------------|
| Manual | Human | 15-60 min | Low |
| Automatic (DB) | Health check | 30-120 sec | Possible (replication lag) |
| Automatic (LB) | Health check | 5-30 sec | None |
| Active-active | N/A | 0 sec | Conflict resolution |

---

## 13. Code or Pseudocode

### Health Check Implementation

```python
def health_check_endpoint():
    """Kubernetes/Docker health check"""
    checks = {
        'database': check_db_connection(),
        'redis': check_redis_connection(),
        'disk': check_disk_space() > 0.1,  # 10% free
    }
    
    all_healthy = all(checks.values())
    status_code = 200 if all_healthy else 503
    
    return {
        'status': 'healthy' if all_healthy else 'unhealthy',
        'checks': checks,
        'timestamp': now()
    }, status_code
```

### Circuit Breaker Pattern

```python
class CircuitBreaker:
    def __init__(self, failure_threshold=5, timeout=60):
        self.failures = 0
        self.last_failure = None
        self.state = 'closed'  # closed, open, half-open
    
    def call(self, func, *args, **kwargs):
        if self.state == 'open':
            if time.now() - self.last_failure > self.timeout:
                self.state = 'half-open'
            else:
                raise CircuitOpenError()
        
        try:
            result = func(*args, **kwargs)
            if self.state == 'half-open':
                self.state = 'closed'
                self.failures = 0
            return result
        except Exception as e:
            self.failures += 1
            self.last_failure = time.now()
            if self.failures >= self.failure_threshold:
                self.state = 'open'
            raise
```

### Availability Calculation

```python
def availability_serial(availabilities):
    """Serial components: all must work"""
    result = 1.0
    for a in availabilities:
        result *= a
    return result

def availability_parallel(availabilities):
    """Parallel components: any can work"""
    result = 1.0
    for a in availabilities:
        result *= (1 - a)
    return 1 - result

# Example: 3 components at 99% in series
print(availability_serial([0.99, 0.99, 0.99]))  # 0.9703

# Example: 2 components at 99% in parallel
print(availability_parallel([0.99, 0.99]))  # 0.9999
```

---

## 14. Interview Discussion

### How to Explain Availability
1. **Define**: "Availability is the % of time the system is operational"
2. **Quantify**: "99.9% = three nines = 8.76 hours downtime per year"
3. **Connect to design**: "We achieve this through redundancy, failover, and eliminating SPOFs"
4. **Give examples**: "Netflix uses Chaos Monkey to verify resilience"

### When Interviewers Expect It
- **System design**: "Design a highly available payment system"
- **Tradeoffs**: "How do you balance availability vs consistency?"
- **Incident**: "Tell me about a time you improved system reliability"
- **Deep dive**: "How would you achieve 99.99% availability?"

### Key Formulas to Know
- Availability = MTBF / (MTBF + MTTR)
- Serial: A_total = A1 × A2 × A3
- Parallel: A_total = 1 - (1-A1)(1-A2)
- Downtime/year = 365 × 24 × (1 - availability) hours

### Follow-Up Questions
- "How do you measure availability?"
- "What's the difference between SLA, SLO, and SLI?"
- "How would you achieve 99.999% availability?"
- "What's the tradeoff between active-passive and active-active?"
- "How do you handle failover without data loss?"

### Common Mistakes
- Confusing availability with reliability (uptime vs correctness)
- Not considering serial components (each reduces total)
- Ignoring human error and deployment as failure modes
- Over-engineering for availability not needed by the business
