# Retries & Backoff Strategies

> Staff+ Engineer Level | FAANG Interview Deep Dive

---

## 1. Concept Overview

**Retries** are the practice of automatically re-attempting failed operations, typically for transient failures. **Backoff** refers to the strategy of waiting between retry attempts to avoid overwhelming failing systems. Together, they are fundamental to building resilient distributed systems.

### Key Definitions

| Term | Definition |
|------|------------|
| **Retry** | Re-execution of a failed operation |
| **Backoff** | Delay between retry attempts |
| **Jitter** | Randomization to prevent synchronized retries |
| **Retry Budget** | Limit on total retries to prevent amplification |
| **Idempotency** | Property that repeated execution has same effect as single execution |
| **Circuit Breaker** | Stops retries when failure rate exceeds threshold |

### The Retry Decision Tree

```
Request fails → Retryable? → Under budget? → Wait (backoff) → Retry
                    │              │
                    No             No
                    │              │
                    ▼              ▼
               Give up        Give up / DLQ
```

---

## 2. Real-World Motivation

### Why Retries Matter

- **Network reliability**: Packet loss, transient congestion, DNS resolution failures
- **Distributed systems**: Remote calls fail 0.1-1% of the time even in healthy systems
- **Resource contention**: Temporary overload, GC pauses, connection pool exhaustion
- **External dependencies**: Third-party APIs have variable reliability

### Failure Modes That Benefit from Retries

| Failure Type | Retryable? | Typical Duration |
|--------------|------------|------------------|
| Network timeout | Yes | Milliseconds to seconds |
| 503 Service Unavailable | Yes | Seconds to minutes |
| 429 Rate Limited | Yes (with backoff) | Seconds |
| 500 Internal Error | Maybe (transient) | Variable |
| 404 Not Found | No | N/A |
| 400 Bad Request | No | N/A |
| Connection refused | Yes | Until service restarts |

### The Cost of Not Retrying

- User-facing errors for transient issues
- Wasted capacity (retry from client often succeeds)
- Cascading failures (one timeout causes more timeouts)

---

## 3. Architecture Diagrams

### Retry Flow (Success Path)

```
Client          Service A         Service B
  │                 │                 │
  │  request        │                 │
  │───────────────►│  request        │
  │                 │───────────────►│
  │                 │                 │
  │                 │  [FAIL - timeout]│
  │                 │◄───────────────│
  │                 │                 │
  │                 │  wait (backoff)  │
  │                 │  ████████        │
  │                 │                 │
  │                 │  retry request   │
  │                 │───────────────►│
  │                 │                 │
  │                 │  success        │
  │                 │◄───────────────│
  │  response       │                 │
  │◄───────────────│                 │
```

### Retry Storm Amplification

```
                    Normal Load: 1000 req/s
                              │
                              ▼
        ┌─────────────────────────────────────────┐
        │           Service A (failing)             │
        │  Each request retries 3x = 3000 req/s    │
        └─────────────────────┬─────────────────────┘
                              │
                              ▼
        ┌─────────────────────────────────────────┐
        │           Service B (dependency)         │
        │  Receives 3x load, starts failing         │
        │  B's clients retry 3x = 9000 req/s        │
        └─────────────────────┬─────────────────────┘
                              │
                              ▼
                    CASCADING FAILURE
                    Total amplification: 9x
```

### Retry + Circuit Breaker Integration

```
                    ┌──────────────────┐
                    │   CLOSED         │
                    │   (normal)       │
                    └────────┬─────────┘
                             │
              failures < threshold
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
         ▼                   │                   ▼
┌─────────────────┐          │          ┌─────────────────┐
│ Retry with      │          │          │ failures >= N   │
│ backoff         │          │          │                 │
└────────┬────────┘          │          └────────┬────────┘
         │                   │                   │
         │ success            │                   ▼
         │                   │          ┌─────────────────┐
         └──────────────────►│          │   OPEN          │
                             │          │ (no retries)    │
                             │          └────────┬────────┘
                             │                   │
                             │                   │ timeout elapsed
                             │                   ▼
                             │          ┌─────────────────┐
                             │          │  HALF_OPEN      │
                             │          │ (probe)         │
                             └──────────│ 1 request       │
                                        └─────────────────┘
```

---

## 4. Core Mechanics

### Retry Policies

| Policy | Formula | Use Case |
|--------|---------|----------|
| **Immediate** | delay = 0 | Low contention, quick failures |
| **Fixed Interval** | delay = constant | Predictable, simple |
| **Linear Backoff** | delay = k × attempt | Moderate growth |
| **Exponential Backoff** | delay = base × 2^attempt | Standard for APIs |
| **Exponential + Jitter** | delay = random(0, base × 2^attempt) | Production standard |

### Exponential Backoff Formula

```
delay = min(cap, base × 2^attempt)
```

Where:
- `base`: Initial delay (e.g., 100ms)
- `attempt`: 0-indexed retry number
- `cap`: Maximum delay (e.g., 30s)

### Jitter Variants

#### Full Jitter (AWS Recommendation)

```
delay = random(0, min(cap, base × 2^attempt))
```

- **Pro**: Prevents thundering herd, simple
- **Con**: Average delay is half of exponential

#### Equal Jitter

```
temp = min(cap, base × 2^attempt)
delay = temp/2 + random(0, temp/2)
```

- **Pro**: Guarantees minimum delay of half, bounds maximum
- **Con**: Slightly more complex

#### Decorrelated Jitter

```
sleep = min(cap, random(sleep, sleep × 3))
```

- **Pro**: Spreads retries across time
- **Con**: More complex, requires state

### Comparison of Jitter Strategies

| Strategy | Min Delay | Max Delay | Use Case |
|----------|-----------|-----------|----------|
| No Jitter | base × 2^n | base × 2^n | Low concurrency |
| Full Jitter | 0 | base × 2^n | High concurrency |
| Equal Jitter | base × 2^(n-1) | base × 2^n | Balanced |
| Decorrelated | variable | cap | Distributed systems |

---

## 5. Numbers

### Typical Retry Configuration

| Parameter | Typical Value | Rationale |
|-----------|---------------|-----------|
| Max retries | 3-5 | Balance success rate vs latency |
| Initial delay | 100ms - 1s | Allow recovery time |
| Max delay | 10s - 60s | Avoid excessive wait |
| Total timeout | 30s - 120s | User experience limit |

### Retry Amplification Math

```
Amplification = (1 + retries) × concurrent_clients
Example: 3 retries, 1000 clients = 4000 requests to failing service
```

### AWS SDK Default Retry Policy

- **Max attempts**: 3 (4 total tries)
- **Backoff**: Exponential with full jitter
- **Retryable errors**: Throttling, 5xx, connection errors
- **Non-retryable**: 4xx (except 429)

### gRPC Retry Configuration

```json
{
  "retryPolicy": {
    "maxAttempts": 4,
    "initialBackoff": "0.1s",
    "maxBackoff": "10s",
    "backoffMultiplier": 2,
    "retryableStatusCodes": ["UNAVAILABLE", "RESOURCE_EXHAUSTED"]
  }
}
```

---

## 6. Tradeoffs (Comparison Tables)

### Retry Policy Comparison

| Policy | Latency Impact | Load on Failing Service | Implementation |
|--------|----------------|-------------------------|----------------|
| No retry | None | Minimal | Trivial |
| Immediate | Low | High (amplification) | Simple |
| Fixed | Medium | Medium | Simple |
| Exponential | Higher | Low (spread out) | Medium |
| Exponential + Jitter | Medium | Lowest | Medium |

### When to Retry vs Not

| Scenario | Retry? | Reason |
|----------|--------|--------|
| Timeout | Yes | May have succeeded |
| 503 | Yes | Transient overload |
| 429 | Yes (with backoff) | Rate limit is temporary |
| 500 | Maybe | Could be transient |
| 404 | No | Resource doesn't exist |
| 400 | No | Client error |
| Connection reset | Yes | Network blip |

---

## 7. Variants/Implementations

### Retry Budgets

Limit total retries across all requests to prevent amplification:

```
budget = 0.2  # 20% of requests can retry
if (retries_this_second / requests_this_second) > budget:
    do_not_retry()
```

### Dead Letter Queues (DLQ)

For message-based systems, after max retries:

1. Move message to DLQ
2. Alert/notify
3. Manual or batch processing
4. Preserve original message for debugging

### Idempotency for Safe Retries

**Critical**: Retries are only safe if operation is idempotent.

| Operation | Idempotent? | Strategy |
|-----------|-------------|----------|
| GET | Yes | Safe to retry |
| PUT (full replace) | Yes | Safe to retry |
| POST (create) | No | Use idempotency key |
| DELETE | Yes | Safe to retry |
| PATCH | Maybe | Use conditional updates |

### Idempotency Key Pattern

```
Client sends: X-Idempotency-Key: uuid-v4
Server: First request with key → process, store result
        Duplicate request with same key → return stored result
```

---

## 8. Timeout Strategy

### Timeout Hierarchy

```
Total Request Timeout (user-facing)
├── Connect Timeout (TCP/TLS handshake)
├── Request Timeout (send request)
└── Read Timeout (wait for response)
```

### Recommended Values

| Timeout Type | Typical Value | Purpose |
|--------------|---------------|---------|
| Connect | 1-5 seconds | Detect unreachable hosts |
| Read | 5-30 seconds | Allow server processing |
| Total | 30-120 seconds | User experience bound |
| Retry total | 2-5 minutes | Across all retries |

### Timeout and Retry Interaction

```
Total timeout must be: (max_retries + 1) × (request_timeout + backoff)
Example: 4 tries × (10s + avg 5s backoff) = 60s total
```

---

## 9. Failure Scenarios

### Retry Storm

**Symptom**: Failing service receives 10x normal load due to retries from many clients.

**Prevention**:
- Exponential backoff with jitter
- Retry budgets
- Circuit breaker (stop retrying when failure rate high)
- Per-client retry limits

### Cascading Retries

```
A → B → C
A retries to B, B retries to C
Each retry multiplies: 3 × 3 = 9x load on C
```

**Prevention**: Retry only at the edge (client), not at every hop. Or use retry budgets.

### Duplicate Operations

**Symptom**: Retry after timeout, but original succeeded → duplicate charge, duplicate order.

**Prevention**: Idempotency keys, idempotent design, at-least-once with deduplication.

### Resource Exhaustion

**Symptom**: Retries hold connections, threads, memory → pool exhaustion.

**Prevention**: Timeouts on retries, connection limits, circuit breaker.

---

## 10. Performance Considerations

### Retry Overhead

- **CPU**: Minimal (scheduling, random number)
- **Memory**: Request held in memory during backoff
- **Connections**: May hold connection pool slots
- **Latency**: P50 unchanged, P99 increases significantly

### Optimizing for Latency

- **Fast fail**: Short connect timeout for unreachable hosts
- **Retry only retryable**: Don't retry 4xx
- **Exponential backoff**: Avoid immediate retry storm
- **Parallel retries**: For read-only, consider fan-out (with care)

### Backoff Tuning

| Scenario | Shorter Backoff | Longer Backoff |
|----------|-----------------|----------------|
| Fast recovery expected | ✓ | |
| Slow recovery (restart) | | ✓ |
| Rate limiting (429) | | ✓ |
| Network blip | ✓ | |

---

## 11. Use Cases

| Use Case | Retry Policy | Special Considerations |
|----------|--------------|-------------------------|
| User-facing API | 2-3 retries, short timeout | Don't block UI |
| Background job | 5+ retries, long backoff | Can afford delay |
| Payment processing | 1-2 retries, idempotency key | No duplicates |
| Cache miss | 1 retry | Low cost of retry |
| External API | 3-5 retries, exponential | Respect rate limits |

---

## 12. Comparison Tables

### Framework Retry Defaults

| Framework | Max Retries | Backoff | Jitter |
|-----------|-------------|---------|--------|
| AWS SDK | 3 | Exponential | Full |
| gRPC | 4 | Exponential | No (configurable) |
| Envoy | 3 | Exponential | Yes |
| Resilience4j | Configurable | Multiple | Yes |
| Polly (.NET) | Configurable | Multiple | Yes |

### Retryable HTTP Status Codes

| Code | Retry? | Backoff |
|------|--------|---------|
| 408 Request Timeout | Yes | Standard |
| 429 Too Many Requests | Yes | Longer (respect Retry-After) |
| 500 Internal Server Error | Yes | Standard |
| 502 Bad Gateway | Yes | Standard |
| 503 Service Unavailable | Yes | Standard |
| 504 Gateway Timeout | Yes | Standard |
| 4xx (other) | No | - |
| 2xx | N/A | Success |

---

## 13. Code/Pseudocode

### Exponential Backoff with Full Jitter

```python
import random

def exponential_backoff_full_jitter(
    attempt: int,
    base_delay: float = 1.0,
    max_delay: float = 60.0,
    multiplier: float = 2.0
) -> float:
    """
    Full jitter: delay = random(0, min(cap, base * 2^attempt))
    """
    exponential = min(max_delay, base_delay * (multiplier ** attempt))
    return random.uniform(0, exponential)

def retry_with_backoff(operation, max_retries=5):
    last_exception = None
    for attempt in range(max_retries + 1):
        try:
            return operation()
        except RetryableError as e:
            last_exception = e
            if attempt == max_retries:
                raise
            delay = exponential_backoff_full_jitter(attempt)
            time.sleep(delay)
    raise last_exception
```

### Equal Jitter Implementation

```python
def exponential_backoff_equal_jitter(
    attempt: int,
    base_delay: float = 1.0,
    max_delay: float = 60.0,
    multiplier: float = 2.0
) -> float:
    """
    Equal jitter: temp = min(cap, base * 2^n)
                 delay = temp/2 + random(0, temp/2)
    """
    temp = min(max_delay, base_delay * (multiplier ** attempt))
    return temp / 2 + random.uniform(0, temp / 2)
```

### Retry Budget (Token Bucket Style)

```python
class RetryBudget:
    def __init__(self, ratio=0.2, min_requests=10):
        self.ratio = ratio
        self.min_requests = min_requests
        self.retries_used = 0
        self.requests_total = 0
    
    def allow_retry(self) -> bool:
        if self.requests_total < self.min_requests:
            return True  # Allow until we have enough data
        return (self.retries_used / self.requests_total) < self.ratio
    
    def record_request(self, was_retry: bool):
        self.requests_total += 1
        if was_retry:
            self.retries_used += 1
```

### Envoy Retry Configuration (YAML)

```yaml
retry_policy:
  retry_on: 5xx,reset,connect-failure,refused-stream
  num_retries: 3
  per_try_timeout: 10s
  retry_host_predicate:
    - name: envoy.retry_host_predicates.previous_hosts
  retry_back_off:
    base_interval: 0.25s
    max_interval: 60s
  retry_priority:
    name: envoy.retry_priorities.previous_priorities
    typed_config:
      "@type": type.googleapis.com/envoy.config.retry.previous_priorities.PreviousPrioritiesConfig
      update_frequency: 2
```

---

## 14. Interview Discussion

### Key Talking Points

1. **Idempotency first**: Retries are dangerous without idempotency
2. **Jitter is critical**: Prevents thundering herd, retry storms
3. **Retry budgets**: Limit amplification in large systems
4. **Circuit breaker combo**: Retry + circuit breaker = resilience
5. **Where to retry**: Edge vs every hop — tradeoff

### Common Questions

**Q: Why add jitter to exponential backoff?**
- Without jitter: 1000 clients fail at T=0, all retry at T=1s, T=2s, etc. — synchronized waves overwhelm the recovering service
- With jitter: Retries spread across time, load is smoothed

**Q: When should you NOT retry?**
- Non-idempotent operations (without idempotency key)
- Client errors (4xx except 429)
- When retry budget exhausted
- When circuit breaker is open

**Q: How do you prevent retry storms?**
- Exponential backoff with jitter
- Retry budgets (limit % of requests that can retry)
- Circuit breaker (stop when failure rate high)
- Per-client/per-endpoint limits

**Q: What's the difference between connect and read timeout?**
- Connect: Time to establish TCP/TLS connection. Short (1-5s) — quick fail for unreachable
- Read: Time waiting for response after request sent. Longer (5-30s) — allow server processing

**Q: How do you make POST requests safe to retry?**
- Idempotency key: Client sends unique key, server deduplicates
- Store result keyed by idempotency key
- Duplicate request returns same result, doesn't re-execute
