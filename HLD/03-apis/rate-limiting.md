# Rate Limiting

## 1. Concept Overview

### Definition
Rate limiting is a technique to control the rate at which requests are processed. It restricts the number of requests a client can make in a given time window.

### Purpose
- **Protect resources**: Prevent overload, ensure fair usage
- **Prevent abuse**: Mitigate DoS, brute force, scraping
- **Ensure fairness**: Multi-tenant systems; no single client monopolizes
- **Cost control**: API costs (e.g., LLM calls) bounded per user

### Problems It Solves
- **Resource exhaustion**: One client can exhaust CPU, DB, or bandwidth
- **Cascading failures**: Unbounded load can take down entire system
- **Abuse**: Bots, scrapers, credential stuffing

---

## 2. Real-World Motivation

### Stripe
- **Tiered limits**: Different limits per API key (test vs live). Returns `429` with `Retry-After` header. Rate limit info in response headers.

### Twitter
- **Twitter API v2**: 300 tweets/15min (user), 2M tweets/month (app). Per-endpoint limits. `x-rate-limit-remaining`, `x-rate-limit-reset`.

### GitHub
- **GitHub API**: 5,000 req/hour (authenticated), 60 (unauthenticated). OAuth apps have separate limits. GraphQL has point-based limits.

### AWS
- **API throttling**: Per-service limits (e.g., DynamoDB read/write capacity). Returns `ThrottlingException`; client should retry with backoff.

---

## 3. Architecture Diagrams

### Token Bucket

```
         ┌─────────────────┐
         │   Token Bucket  │
    ┌───▶│  Capacity: 10   │
    │    │  Refill: 2/sec   │
    │    └────────┬────────┘
    │             │ 1 token per request
    │             ▼
    │    ┌─────────────────┐
    │    │    Request      │
    │    │  Token available?│──Yes──▶ Process
    │    └────────┬────────┘
    │             │ No
    │             ▼
    │         Reject 429
    │
    └── Refill at fixed rate
```

### Leaky Bucket

```
    Requests ──▶ ┌─────────────┐
                 │   Bucket   │
                 │  (queue)   │
                 │  Leak rate │──▶ Process at fixed rate
                 │  = 5/sec   │
                 └─────────────┘
                 (requests leave at constant rate)
```

### Fixed Window Counter (Boundary Burst Problem)

```
Window 1: [0-60s]          Window 2: [60-120s]
|--------100 reqs---------||--------100 reqs---------|
                           ^
                    At 59s: 100 requests
                    At 60s: 100 more (200 in 1 second!)
```

### Sliding Window Log

```
Current time: 65s
Track: [10, 15, 20, 55, 60, 62]  (timestamps of requests)
Window: 60s → count requests after 5s → 4 requests
New request at 65s → add 65, remove < 5 → allow if count < limit
```

### Sliding Window Counter (Hybrid)

```
count_prev = requests in previous window
count_curr = requests in current window
elapsed = time into current window
count = count_prev * (1 - elapsed/window) + count_curr
Allow if count < limit
```

---

## 4. Core Mechanics

### Token Bucket
- Bucket holds tokens (max capacity)
- Tokens added at fixed rate (e.g., 10/sec)
- Each request consumes 1 token (or N for weighted)
- If no tokens, reject

### Leaky Bucket
- Requests enter queue
- Process at fixed "leak" rate
- Excess either queued (bounded) or dropped

### Fixed Window Counter
- Counter per (user, window). Window = e.g., 1 minute.
- Increment on request; reject if counter >= limit
- Reset at window boundary
- **Boundary burst**: User gets 2x limit at window boundary

### Sliding Window Log
- Store timestamp of each request
- Count requests in last N seconds
- Accurate but memory-heavy (O(requests))

### Sliding Window Counter
- Approximate sliding window using fixed window stats
- Lower memory; slight approximation

---

## 5. Numbers

| Algorithm | Memory | Accuracy | Boundary burst |
|-----------|--------|----------|----------------|
| Fixed Window | O(1) | Low | Yes (2x) |
| Sliding Log | O(n) | High | No |
| Sliding Counter | O(1) | Medium | No |
| Token Bucket | O(1) | High | No (smooth) |
| Leaky Bucket | O(n) or O(1) | High | No |

---

## 6. Tradeoffs

### Algorithm Comparison

| Algorithm | Pros | Cons |
|-----------|------|------|
| Fixed Window | Simple, low memory | Boundary burst |
| Sliding Log | Accurate | High memory |
| Sliding Counter | Good balance | Approximation |
| Token Bucket | Smooth, allows bursts | Slightly more complex |
| Leaky Bucket | Smooth output rate | Can delay requests |

---

## 7. Variants / Implementations

### Distributed Rate Limiting (Redis)

```python
# Token bucket in Redis
def rate_limit(user_id, limit=100, window=60):
    key = f"ratelimit:{user_id}"
    now = time.time()
    pipe = redis.pipeline()
    pipe.zadd(key, {now: now})
    pipe.zremrangebyscore(key, 0, now - window)
    pipe.zcard(key)
    pipe.expire(key, window)
    _, _, count, _ = pipe.execute()
    return count <= limit
```

**Race condition**: Multiple requests can increment before seeing each other. Use Lua script for atomicity.

### Sticky Sessions
- Route same client to same server
- Use local rate limit (no Redis)
- Simpler but less flexible (server failure loses state)

---

## 8. Scaling Strategies

1. **Redis cluster**: Shard by user ID for distributed limits
2. **Lua scripts**: Atomic check-and-increment in Redis
3. **Local + distributed**: Local cache for fast path; Redis for consistency

---

## 9. Failure Scenarios

| Failure | Impact | Mitigation |
|---------|--------|------------|
| Redis down | Rate limit fails open/closed? | Fail open (allow) or fail closed (reject); configurable |
| Clock skew | Fixed window boundaries wrong | Use Redis time; avoid client time |
| Race condition | Allow more than limit | Lua script for atomicity |

---

## 10. Performance Considerations

- **Redis latency**: ~1ms per check; use pipeline/Lua to reduce round-trips
- **In-memory**: Sub-millisecond for local limits

---

## 11. Use Cases

| Use Case | Algorithm |
|----------|-----------|
| API per-user limit | Token bucket or sliding window |
| DDoS mitigation | Fixed window (simpler at edge) |
| Fair queuing | Leaky bucket |
| Cost control (LLM) | Token bucket with cost per request |

---

## 12. Comparison Tables

### Rate Limit Headers

| Header | Meaning |
|--------|---------|
| X-RateLimit-Limit | Max requests per window |
| X-RateLimit-Remaining | Remaining in current window |
| X-RateLimit-Reset | Unix timestamp when window resets |
| Retry-After | Seconds after which to retry (429) |

---

## 13. Code or Pseudocode

### Token Bucket (Pseudocode)

```python
class TokenBucket:
    def __init__(self, capacity, refill_rate):
        self.capacity = capacity
        self.tokens = capacity
        self.refill_rate = refill_rate
        self.last_refill = time.time()
    
    def allow(self):
        now = time.time()
        self.tokens = min(
            self.capacity,
            self.tokens + (now - self.last_refill) * self.refill_rate
        )
        self.last_refill = now
        if self.tokens >= 1:
            self.tokens -= 1
            return True
        return False
```

### Sliding Window (Redis Lua)

```lua
-- KEYS[1]: rate limit key
-- ARGV[1]: window size (seconds)
-- ARGV[2]: limit
-- ARGV[3]: current timestamp
local key = KEYS[1]
local window = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

redis.call('ZREMRANGEBYSCORE', key, 0, now - window)
local count = redis.call('ZCARD', key)
if count < limit then
    redis.call('ZADD', key, now, now)
    redis.call('EXPIRE', key, window)
    return 1  -- allowed
else
    return 0  -- rejected
end
```

### Fixed Window (Simple)

```python
def fixed_window_check(user_id, limit=100, window=60):
    key = f"ratelimit:{user_id}:{time.time() // window}"
    count = redis.incr(key)
    if count == 1:
        redis.expire(key, window)
    return count <= limit
```

---

## 14. Interview Discussion

### How to Explain
"Rate limiting controls request rate. Token bucket allows bursts up to capacity; refill at fixed rate. Fixed window is simple but has boundary burst (2x at window edge). Sliding window is accurate but needs more memory. For distributed: Redis with Lua for atomicity."

### Follow-up Questions
- "How would you implement distributed rate limiting?"
- "What's the boundary burst problem? How do you fix it?"
- "Design multi-tier rate limiting (e.g., per-user and per-IP)."

---

## Appendix A: Multi-Tier Rate Limiting

```
Request ──▶ Tier 1: Per-IP (1000/min) ──▶ Tier 2: Per-User (100/min) ──▶ Tier 3: Per-API-Key (10000/min)
              │                              │                              │
              └── Reject 429                  └── Reject 429                 └── Reject 429
```

Apply in order: IP (anonymous), then user (authenticated), then plan (API key tier).

---

## Appendix B: Rate Limit Response Format

```json
{
  "error": {
    "code": "rate_limit_exceeded",
    "message": "Too many requests. Retry after 60 seconds."
  },
  "retry_after": 60
}
```

Headers:
```
HTTP/1.1 429 Too Many Requests
Retry-After: 60
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1640000000
```

---

## Appendix C: Token Bucket vs Leaky Bucket Detail

**Token Bucket**: Allows bursts. If bucket full (10 tokens), 10 requests can go through immediately. Then refill at 2/sec. Good for APIs where occasional burst is OK.

**Leaky Bucket**: Smooths output. Requests leave at fixed rate (5/sec). No burst; more predictable. Good for protecting downstream with fixed capacity.

---

## Appendix D: Redis Lua Script for Atomic Sliding Window

```lua
-- Prevents race: check and add in single atomic operation
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])

redis.call('ZREMRANGEBYSCORE', key, '-inf', now - window)
local current = redis.call('ZCARD', key)

if current < limit then
    redis.call('ZADD', key, now, now .. ':' .. math.random())
    redis.call('PEXPIRE', key, window * 1000)
    return 1
else
    return 0
end
```

---

## Appendix E: Rate Limiting by Cost (e.g., LLM API)

Instead of request count, limit by "cost" (tokens, compute):

```python
def check_cost_limit(user_id, cost):
    key = f"cost:{user_id}"
    total = redis.incrby(key, cost)
    if redis.ttl(key) == -1:
        redis.expire(key, 3600)  # 1 hour window
    return total <= COST_LIMIT_PER_HOUR
```

---

## Appendix F: Sticky Session vs Distributed Tradeoff

| Approach | Pros | Cons |
|----------|------|------|
| Sticky sessions | No Redis, fast | Uneven load, state on server |
| Redis distributed | Even load, consistent | Redis dependency, ~1ms latency |

---

## Appendix G: Rate Limit Bypass (Internal Services)

Some traffic should bypass limits:
- **Health checks**: No limit
- **Internal services**: Higher or no limit (trusted)
- **Admin API**: Separate limits

Implementation: Check `X-Forwarded-For`, `X-Internal-Service` header; apply different limits.

---

## Appendix H: Sliding Window Counter Formula

```
count = prev_count * (1 - elapsed / window) + curr_count

Example: 60s window, 100 limit
- Prev window had 80 requests
- Current window (30s elapsed) has 50 requests
- count = 80 * (1 - 30/60) + 50 = 40 + 50 = 90
- 90 < 100 → allow
```

---

## Appendix I: Real-World Rate Limit Numbers

| Service | Limit |
|---------|-------|
| Stripe API | 100 req/sec (standard) |
| GitHub API | 5,000 req/hour (authenticated) |
| Twitter API | 300 tweets/15min (user) |
| AWS | Varies by service (e.g., DynamoDB 40K read/sec) |

---

## Appendix J: Rate Limiting at CDN Edge

Cloudflare, Fastly, AWS CloudFront can rate limit at edge:
- **Pros**: Block before hitting origin; reduce load
- **Cons**: Less context (no user ID until auth); IP-based only for anonymous
- **Use**: DDoS mitigation, basic abuse prevention

---

## Appendix K: Graceful Degradation When Redis Fails

```python
def rate_limit(user_id):
    try:
        return redis_rate_limit(user_id)
    except RedisError:
        if config.FAIL_OPEN:
            return True   # Allow when Redis down
        else:
            return False  # Reject when Redis down (safer)
```

Fail-open: availability. Fail-closed: consistency/safety.

---

## Appendix L: Rate Limit Response Caching

When rate limited, some APIs return cached/stale data with header:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1640000060
Retry-After: 60
```
Client should wait `Retry-After` seconds before retrying.
