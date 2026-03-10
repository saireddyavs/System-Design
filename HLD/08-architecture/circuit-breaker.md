# Circuit Breaker Pattern

## 1. Concept Overview

### Definition
The Circuit Breaker pattern is a fault-tolerance design that prevents a client from repeatedly invoking a failing remote service. It acts like an electrical circuit breaker: when failures exceed a threshold, the circuit "opens" and blocks further calls for a period, allowing the failing service to recover. After a timeout, the circuit transitions to "half-open" to test if the service has recovered.

### Purpose
- **Prevent cascading failures**: Stop sending requests to a failing service
- **Fail fast**: Return immediately when circuit is open instead of waiting for timeouts
- **Allow recovery**: Give failing service time to recover
- **Protect resources**: Avoid exhausting threads, connections, or memory on doomed requests
- **Graceful degradation**: Provide fallback when dependency is unavailable

### Problems It Solves
- **Cascading failure**: One failing service causes callers to hang (waiting for timeouts), exhausting their resources, which then fail their callers
- **Resource exhaustion**: Thousands of threads blocked on slow/failing downstream
- **Amplification**: Retrying failed requests can overwhelm the failing service
- **Slow failure detection**: Without circuit breaker, each request times out (e.g., 30s); with it, open circuit returns immediately

---

## 2. Real-World Motivation

### Netflix
- **Hystrix**: Created circuit breaker library; open-sourced 2012
- **Cascading failure (2008)**: Database corruption caused cascading failure; led to Hystrix
- **Scale**: 400+ million API requests/day; one failing dependency could take down entire platform
- **Deprecated Hystrix**: Moved to Resilience4j, Envoy; resilience now in platform layer

### Uber
- **Service isolation**: Circuit breakers between 4000+ microservices
- **Use case**: Payment service down → Order service returns "payment unavailable" instead of hanging
- **Integration**: Envoy, custom middleware

### Amazon
- **Dependency isolation**: Each service uses circuit breakers for downstream calls
- **Bulkhead**: Isolate failure to specific dependency; don't let one failure exhaust all connections
- **Fallback**: Return cached data or degraded response

### Stripe
- **Payment API resilience**: Circuit breakers for bank/processor connections
- **Fallback**: Queue for retry; return "processing" to user

---

## 3. Architecture Diagrams

### Circuit Breaker State Machine

```
                    ┌──────────────────────────────────────────────────┐
                    │                 CIRCUIT BREAKER                   │
                    └──────────────────────────────────────────────────┘

    ┌─────────────┐     Failures < threshold      ┌─────────────┐
    │   CLOSED    │◀─────────────────────────────│  HALF-OPEN  │
    │ (Normal)    │                               │  (Testing)  │
    │             │     Success on test call     │             │
    └──────┬──────┘                               └──────▲──────┘
           │                                              │
           │ Failures >= threshold                        │
           │ (or slow call rate)                          │ Timeout elapsed;
           │                                              │ allow 1 test call
           ▼                                              │
    ┌─────────────┐                                       │
    │    OPEN     │───────────────────────────────────────┘
    │ (Failing)   │   After sleep/timeout window
    │             │
    │ Fail fast   │
    │ Return      │
    │ immediately │
    └─────────────┘
```

### Cascading Failure (Without Circuit Breaker)

```
    User Request
         │
         ▼
    ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
    │  Frontend   │────▶│  Order Svc   │────▶│ Payment Svc │
    │  (1000 req) │     │  (1000 req) │     │   (DOWN!)    │
    └─────────────┘     └─────────────┘     └─────────────┘
         │                     │                     │
         │                     │  Each request       │
         │                     │  waits 30s timeout  │
         │                     │  Threads exhausted  │
         │                     │  Order Svc fails    │
         │                     │                     │
         │  Frontend threads   │                     │
         │  exhausted waiting  │                     │
         │  for Order Svc      │                     │
         ▼                     ▼                     ▼
    CASCADING FAILURE - Entire system down
```

### With Circuit Breaker

```
    User Request
         │
         ▼
    ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
    │  Frontend   │────▶│  Order Svc   │────▶│ Payment Svc │
    │             │     │  [Circuit    │     │   (DOWN!)    │
    │             │     │   OPEN]      │     └─────────────┘
    └─────────────┘     └──────┬──────┘
         │                     │
         │                     │ Fail fast - return immediately
         │                     │ Use fallback: "Payment unavailable"
         │                     │
         ▼                     ▼
    User gets response; system stays up
```

### Bulkhead Pattern (Resource Isolation)

```
    ┌─────────────────────────────────────────────────────────┐
    │                    APPLICATION                           │
    │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐      │
    │  │  Pool A     │  │  Pool B     │  │  Pool C     │      │
    │  │  (Payment)  │  │  (Inventory)│  │  (Shipping) │      │
    │  │  10 threads │  │  10 threads │  │  10 threads │      │
    │  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘      │
    │         │                │                │              │
    │  Failure in Payment doesn't exhaust threads for others   │
    └─────────┼────────────────┼────────────────┼──────────────┘
              │                │                │
              ▼                ▼                ▼
         Payment Svc     Inventory Svc    Shipping Svc
```

---

## 4. Core Mechanics

### States
- **Closed**: Normal operation; requests pass through; failures are counted
- **Open**: Circuit has tripped; requests fail immediately (no call to downstream); after timeout, transition to half-open
- **Half-Open**: Allow limited (e.g., 1) test request; success → closed; failure → open

### Thresholds
- **Failure count**: Open after N consecutive failures (e.g., 5)
- **Failure rate**: Open when failure rate exceeds X% over window (e.g., 50% over 10s)
- **Slow call rate**: Open when slow calls exceed threshold (e.g., 50% take >2s)
- **Timeout**: How long to wait before half-open (e.g., 30s)

### Fallback Strategies
- **Default value**: Return cached or default (e.g., empty list)
- **Cache**: Return stale cached data
- **Degraded service**: Return partial response (e.g., recommendations without personalization)
- **Error response**: Return 503 with retry-after
- **Queue for retry**: Accept request, process async when service recovers

---

## 5. Numbers

| Parameter | Typical Range | Example |
|-----------|---------------|---------|
| Failure threshold | 5-10 consecutive | 5 failures → open |
| Failure rate | 50-80% | 50% over 10s window |
| Slow call threshold | 1-5 seconds | Calls >2s count as slow |
| Wait in open | 10-60 seconds | 30s before half-open |
| Half-open calls | 1-5 | Allow 1 test call |
| Success threshold (half-open) | 1-3 | 1 success → closed |

### Impact
- **Without circuit breaker**: 1000 req/s × 30s timeout = 30,000 threads blocked
- **With circuit breaker**: Open → 0 threads blocked; fail fast in <1ms
- **Recovery**: Failing service gets no traffic; can recover; half-open tests

---

## 6. Tradeoffs

### Circuit Breaker vs Retry

| Aspect | Retry Only | Retry + Circuit Breaker |
|--------|------------|-------------------------|
| Transient failure | Retry helps | Retry helps |
| Sustained failure | Keeps retrying; exhausts resources | Stops; protects system |
| Recovery | N/A | Gives downstream time to recover |
| Complexity | Lower | Higher |

### Fallback Strategies Comparison

| Strategy | Use Case | Risk |
|----------|----------|------|
| Default value | Non-critical data | Stale/wrong data shown |
| Cache | Read-heavy | Stale data |
| Degraded | Partial functionality | UX degradation |
| Error | Critical path | User sees error |
| Queue | Async processing | Delayed processing |

### Threshold Tuning

| Too sensitive | Balanced | Too insensitive |
|---------------|----------|-----------------|
| Opens on 1-2 failures | 5-10 failures, 50% rate | 100+ failures |
| Short open (5s) | 30-60s | Very long (5 min) |
| Risk: False opens | Good | Risk: Cascading failure |

---

## 7. Variants / Implementations

### Netflix Hystrix (Deprecated)
- **Status**: In maintenance mode; Netflix moved to Resilience4j, Envoy
- **Features**: Circuit breaker, fallback, bulkhead, request cache
- **Integration**: Java; Spring Cloud Netflix

### Resilience4j
- **Successor**: Lightweight; modular (circuit breaker, retry, rate limiter, bulkhead)
- **Usage**: `CircuitBreaker.of("payment", config)`
- **Metrics**: Micrometer integration
- **Adoption**: Widely used in Java ecosystem

### Polly (.NET)
- **Features**: Retry, circuit breaker, bulkhead, timeout, cache
- **Usage**: `Policy.Handle<HttpRequestException>().CircuitBreakerAsync(5, TimeSpan.FromSeconds(30))`
- **Adoption**: Standard for .NET resilience

### Envoy
- **Outlier detection**: Ejects unhealthy hosts; circuit breaker-like
- **Config**: `outlier_detection.consecutive_5xx`, `base_ejection_time`
- **Level**: Infrastructure; no app code changes

### Go (sony/gobreaker)
- **Interface**: `func() (interface{}, error)` with circuit breaker wrapper
- **States**: Closed, open, half-open
- **Lightweight**: Pure Go; no dependencies

---

## 8. Scaling Strategies

- **Per dependency**: Separate circuit breaker per downstream service
- **Per operation**: Different breakers for read vs write (e.g., DB read vs write)
- **Shared state**: In distributed system, circuit state can be local (each instance) or shared (Redis); usually local for simplicity
- **Observability**: Metrics for state transitions, failure rates; alert on open circuits

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|------------|
| False open | Unnecessary failures | Tune threshold; use failure rate vs count |
| Stuck open | Never recovers | Half-open with test; manual reset option |
| Slow recovery | Service up but circuit still open | Reduce wait-in-open; increase half-open calls |
| Thundering herd | All instances go half-open at once; flood recovered service | Randomize half-open timing; allow only 1 test per instance |
| No fallback | User sees generic error | Implement fallback (cache, default) |

### Retry + Circuit Breaker
- **Retry**: For transient failures (network blip)
- **Circuit breaker**: After retries exhausted and failures persist
- **Order**: Retry within closed circuit; circuit opens if retries keep failing

---

## 10. Performance Considerations

- **Fail fast**: Open circuit returns in <1ms vs timeout (e.g., 30s)
- **Overhead**: Minimal when closed (counter increment, check)
- **Half-open**: Only 1 call; avoid thundering herd
- **Metrics**: Track state transitions; don't log every rejected call (noise)

---

## 11. Use Cases

**Essential for:**
- Microservices with many dependencies
- External API calls (payment, third-party)
- Database connections
- Any remote call that can hang or fail

**Less critical for:**
- In-process calls
- Highly reliable dependencies
- Synchronous critical path with no fallback

---

## 12. Comparison Tables

### Circuit Breaker vs Timeout vs Retry

| Pattern | Purpose |
|---------|---------|
| Timeout | Don't wait forever for single request |
| Retry | Handle transient failures |
| Circuit breaker | Stop calling when dependency is down |
| **Together** | Timeout per request; retry a few times; circuit breaker stops all after sustained failure |

### Implementation Comparison

| Library | Language | Deprecated | Notes |
|---------|----------|------------|-------|
| Hystrix | Java | Yes | Netflix; use Resilience4j |
| Resilience4j | Java | No | Lightweight; modular |
| Polly | .NET | No | Full-featured |
| gobreaker | Go | No | Simple |
| Envoy | Any | N/A | Infrastructure level |

---

## 13. Code or Pseudocode

### Circuit Breaker Pseudocode

```python
class CircuitBreaker:
    def __init__(self, failure_threshold=5, timeout=30):
        self.failure_threshold = failure_threshold
        self.timeout = timeout
        self.state = "CLOSED"
        self.failure_count = 0
        self.last_failure_time = None
    
    def call(self, func, *args, **kwargs):
        if self.state == "OPEN":
            if time.time() - self.last_failure_time > self.timeout:
                self.state = "HALF_OPEN"
                self.failure_count = 0
            else:
                raise CircuitOpenError("Circuit is open")
        
        try:
            result = func(*args, **kwargs)
            if self.state == "HALF_OPEN":
                self.state = "CLOSED"
                self.failure_count = 0
            return result
        except Exception as e:
            self.failure_count += 1
            self.last_failure_time = time.time()
            if self.failure_count >= self.failure_threshold:
                self.state = "OPEN"
            elif self.state == "HALF_OPEN":
                self.state = "OPEN"
            raise
```

### Resilience4j (Java)

```java
CircuitBreakerConfig config = CircuitBreakerConfig.custom()
    .failureRateThreshold(50)
    .waitDurationInOpenState(Duration.ofSeconds(30))
    .slidingWindowSize(10)
    .build();

CircuitBreaker circuitBreaker = CircuitBreaker.of("paymentService", config);

String result = circuitBreaker.executeSupplier(() -> 
    paymentService.charge(order)
);

// With fallback
String result = circuitBreaker.executeSupplier(() -> 
    paymentService.charge(order)
).recover(throwable -> "Payment unavailable. Please try later.");
```

### Polly (.NET)

```csharp
var circuitBreakerPolicy = Policy
    .Handle<HttpRequestException>()
    .CircuitBreakerAsync(
        handledEventsAllowedBeforeBreaking: 5,
        durationOfBreak: TimeSpan.FromSeconds(30)
    );

var result = await circuitBreakerPolicy.ExecuteAsync(() =>
    httpClient.GetAsync("https://payment-service/charge")
);
```

### Envoy Outlier Detection

```yaml
outlier_detection:
  consecutive_5xx: 5
  interval: 30s
  base_ejection_time: 30s
  max_ejection_percent: 50
```

---

## 14. Interview Discussion

### Key Points
1. **Three states**: Closed → Open → Half-Open
2. **Purpose**: Prevent cascading failure; fail fast when dependency is down
3. **Thresholds**: Failure count, failure rate, slow call rate
4. **Fallback**: Always have fallback (cache, default, error message)
5. **With retry**: Retry for transient; circuit breaker for sustained failure
6. **Bulkhead**: Isolate resource pools; one failing dependency doesn't exhaust all

### Common Questions
- **"What is circuit breaker?"** → Stops calling failing service; fails fast; allows recovery
- **"When does it open?"** → After N failures or X% failure rate
- **"What's half-open?"** → Test with limited requests; success → closed; failure → open
- **"How does it prevent cascading failure?"** → Stops sending requests; callers don't block; return fallback immediately
- **"Circuit breaker vs retry?"** → Retry for transient; circuit breaker stops when sustained failure
- **"What's bulkhead?"** → Isolate thread/connection pools per dependency

### Red Flags
- No fallback strategy
- Ignoring half-open (stuck open)
- Too aggressive (opens on 1 failure)
- No metrics/alerting on circuit state

---

## Appendix: Related Patterns

### Timeout Pattern
- **Purpose**: Don't wait indefinitely for a response
- **Implementation**: Set timeout on every outbound call (e.g., 5s)
- **With circuit breaker**: Timeout counts as failure toward circuit opening
- **Tuning**: Balance between too short (false failures) and too long (resource exhaustion)

### Retry Pattern
- **Exponential backoff**: 1s, 2s, 4s, 8s between retries
- **Jitter**: Add randomness to avoid thundering herd
- **Idempotency**: Ensure retried operations are safe (e.g., PUT, not POST for create)
- **Max retries**: Typically 3-5; then fail or trigger circuit breaker

### Health Check Pattern
- **Liveness**: Is the process running? Restart if not.
- **Readiness**: Can the process accept traffic? (e.g., DB connected)
- **Circuit breaker**: May use health check results to mark instance unhealthy
- **Kubernetes**: livenessProbe, readinessProbe

### Bulkhead Pattern (Detailed)
- **Isolation**: Separate thread pools per dependency
- **Example**: 10 threads for Payment, 10 for Inventory, 10 for Shipping
- **Benefit**: Payment service down → exhausts only Payment pool; others continue
- **Implementation**: BoundedExecutor, semaphores, connection pools per dependency
