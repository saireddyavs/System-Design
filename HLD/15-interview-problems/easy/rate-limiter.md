# System Design: Rate Limiter

## 1. Problem Statement & Requirements

### Functional Requirements

- **Limit by Identifier**: Rate limit by user ID, API key, or IP address
- **Configurable Rules**: Different limits for different endpoints (e.g., 100/min for search, 10/min for write)
- **Multiple Tiers**: Support per-second, per-minute, per-hour, per-day limits
- **Return Clear Response**: When rate limited, return 429 with `Retry-After` header
- **Whitelist/Blacklist**: Allow certain users to bypass or have stricter limits

### Non-Functional Requirements

- **Distributed**: Must work across multiple API server instances (shared state)
- **Low Latency**: Rate limit check adds < 5ms (p99) to request path
- **High Availability**: Rate limiter failure should not block requests (fail open vs fail closed)
- **Accuracy**: Reasonable accuracy; slight over/under limit acceptable
- **Scalability**: Support millions of unique identifiers

### Out of Scope

- Per-tenant custom limits (assume fixed tiers)
- Machine learning for anomaly detection
- Geographic rate limiting
- Complex rule engine (AND/OR conditions)

---

## 2. Back-of-Envelope Estimation

### Traffic Estimates

| Metric | Value | Calculation |
|--------|-------|-------------|
| API QPS | 100K | Given |
| Unique users/min | 50K | Assume 50% of QPS from unique users |
| Rate limit checks | 100K/sec | Every request checked |
| Redis operations | 200K/sec | 2 ops per check (get + set) for sliding window |

### Storage Estimates

- **Per user state**: ~100 bytes (counters for different windows)
- **Active users**: 1M concurrent (sliding 1-hour window)
- **Total**: 1M × 100 bytes ≈ 100 MB in Redis
- **With overhead**: ~200 MB

### Latency Budget

- **Target**: < 5ms for rate limit check
- **Redis round-trip**: ~1-2ms (same datacenter)
- **Algorithm**: O(1) operations
- **Network**: Minimal

---

## 3. API Design

### Rate Limiter as Middleware

Rate limiter is typically **middleware** or **sidecar**, not a standalone API. It intercepts requests.

### Response When Rate Limited

```
HTTP/1.1 429 Too Many Requests
Retry-After: 45
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1710000000

{
  "error": "RATE_LIMIT_EXCEEDED",
  "message": "Rate limit exceeded. Try again in 45 seconds.",
  "retry_after": 45
}
```

### Response Headers (Every Request)

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1710000060
```

### Configuration API (Admin)

```
GET /admin/rate-limits
POST /admin/rate-limits
```

**Config structure:**
```json
{
  "endpoint": "/api/v1/search",
  "limits": [
    {"window": "1s", "max_requests": 10},
    {"window": "1m", "max_requests": 100},
    {"window": "1h", "max_requests": 1000}
  ],
  "identifier": "user_id"
}
```

---

## 4. Data Model / Database Schema

### Redis Data Structures

**Key format**: `ratelimit:{identifier}:{endpoint}:{window}`

**Example**: `ratelimit:user_123:/api/search:1m`

### Sliding Window Log

```
Key: ratelimit:user_123:/api/search:1m
Type: Sorted Set (ZSET)
Score: Timestamp of request
Member: Request ID or timestamp (unique)
```

**Operations**:
- Add current request: `ZADD key now member`
- Remove expired: `ZREMRANGEBYSCORE key -inf (now - window)`
- Count: `ZCARD key`
- Compare to limit

### Token Bucket (Alternative)

```
Key: ratelimit:user_123:/api/search:tokens
Type: String (or Hash)
Value: {tokens: 95, last_refill: 1710000000}
```

### Fixed Window Counter

```
Key: ratelimit:user_123:/api/search:1m:17100000
Type: String
Value: 42
TTL: 60
```

---

## 5. High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              CLIENT REQUEST                                       │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│   LAYER 1: API Gateway / Load Balancer (Optional - Coarse rate limit by IP)       │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│   LAYER 2: Application (Rate Limiter Middleware)                                  │
│   - Extract identifier (user_id, API key, IP)                                     │
│   - Check Redis for current count                                                 │
│   - Allow or reject (429)                                                         │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
                    ┌───────────────────┼───────────────────┐
                    ▼                   ▼                   ▼
┌───────────────────────┐   ┌───────────────────────┐   ┌───────────────────────┐
│   API Server 1        │   │   API Server 2         │   │   API Server N        │
│   + Rate Limit MW     │   │   + Rate Limit MW      │   │   + Rate Limit MW     │
└───────────────────────┘   └───────────────────────┘   └───────────────────────┘
                    │                   │                   │
                    └───────────────────┼───────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│   Redis Cluster (Shared state for distributed rate limiting)                      │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 6. Detailed Component Design

### 6.1 Rate Limiting Algorithms

#### A. Fixed Window Counter

- **Idea**: Count requests in fixed windows (e.g., minute 0-60, 60-120)
- **Pros**: Simple, low memory
- **Cons**: Burst at boundary (e.g., 100 at 59s, 100 at 61s = 200 in 2 seconds)

```
Window 1: [0--------60)     Window 2: [60-------120)
          |___100 reqs___|            |___100 reqs___|
                    ^^ Burst: 200 in 2 sec
```

#### B. Sliding Window Log

- **Idea**: Store timestamp of each request; count requests in last N seconds
- **Pros**: Accurate, no boundary burst
- **Cons**: Memory grows with request count (mitigate: use sliding window counter)

#### C. Sliding Window Counter (Hybrid)

- **Idea**: Approximate sliding window using previous window count
- **Formula**: `count = prev_count * (1 - elapsed/window) + current_count`
- **Pros**: Low memory, O(1), reasonably accurate
- **Cons**: Approximation

#### D. Token Bucket

- **Idea**: Bucket has tokens; each request consumes 1 token; tokens refill at constant rate
- **Params**: Capacity (max tokens), refill rate (tokens per second)
- **Pros**: Allows bursts up to capacity; smooth rate
- **Cons**: Slightly more complex

#### E. Leaky Bucket

- **Idea**: Requests enter queue; processed at fixed rate (leak)
- **Pros**: Smooth output rate
- **Cons**: Can cause latency; less common for API rate limiting

**Recommendation**: **Sliding Window Log** (accurate) or **Token Bucket** (allows bursts). For distributed: **Sliding Window Log with Redis ZSET** or **Lua script for atomicity**.

### 6.2 Redis Implementation: Sliding Window Log

**Lua Script** (atomic; avoids race conditions):

```lua
-- KEYS[1]: rate limit key
-- ARGV[1]: window size (seconds)
-- ARGV[2]: max requests
-- ARGV[3]: current timestamp
-- ARGV[4]: unique request id

local key = KEYS[1]
local window = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local request_id = ARGV[4]

-- Remove expired entries
redis.call('ZREMRANGEBYSCORE', key, 0, now - window)

-- Count current requests
local count = redis.call('ZCARD', key)

if count < limit then
    redis.call('ZADD', key, now, request_id)
    redis.call('EXPIRE', key, window)
    return {0, count + 1}  -- 0 = allowed, remaining
else
    return {1, count}  -- 1 = rejected
end
```

**Pseudocode (Application)**:

```python
def is_rate_limited(user_id, endpoint, limit=100, window=60):
    key = f"ratelimit:{user_id}:{endpoint}:{window}"
    now = time.time()
    request_id = str(uuid.uuid4())
    
    result = redis.eval(LUA_SCRIPT, 1, key, window, limit, now, request_id)
    allowed = result[0] == 0
    remaining = limit - result[1] if allowed else 0
    
    return not allowed, remaining
```

### 6.3 Token Bucket (Redis)

```lua
-- KEYS[1]: bucket key
-- ARGV[1]: capacity
-- ARGV[2]: refill rate (tokens per second)
-- ARGV[3]: now

local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local data = redis.call('HMGET', key, 'tokens', 'last_refill')
local tokens = tonumber(data[1]) or capacity
local last_refill = tonumber(data[2]) or now

-- Refill tokens based on elapsed time
local elapsed = now - last_refill
tokens = math.min(capacity, tokens + elapsed * rate)

if tokens >= 1 then
    tokens = tokens - 1
    redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
    redis.call('EXPIRE', key, 3600)  -- 1 hour TTL
    return {0, math.floor(tokens)}  -- allowed
else
    return {1, 0}  -- rejected
end
```

### 6.4 Race Condition Handling

**Problem**: Two requests arrive simultaneously; both read count=99; both think they can proceed; count becomes 101.

**Solution**: Use **Lua scripts** — Redis executes script atomically. No interleaving.

**Alternative**: Redis INCR with WATCH (optimistic locking) — more complex.

### 6.5 Multi-Tier Limits

**Example**: 10/sec, 100/min, 1000/hour

- Check all three; if any exceeded, reject
- Use separate keys: `...:1s`, `...:1m`, `...:1h`
- **Optimization**: Check most restrictive first (1s); if pass, check 1m; if pass, check 1h

### 6.6 Identifier Extraction

- **User ID**: From JWT or session (authenticated)
- **API Key**: From header `X-API-Key`
- **IP**: From `X-Forwarded-For` or `X-Real-IP` (last proxy)
- **Fallback**: IP if no user (anonymous)

### 6.7 Rate Limiting at Different Layers

| Layer | Pros | Cons |
|-------|------|------|
| **API Gateway** | Centralized, offloads app | Less flexible, coarse |
| **Load Balancer** | Early rejection | Usually IP-only |
| **Application** | Fine-grained, per-user | Adds latency, every server needs it |
| **Sidecar (Envoy)** | Decoupled, consistent | Extra hop |

**Recommendation**: Application-level for flexibility; optionally API Gateway for coarse IP limits (DDoS protection).

### 6.8 Fail Open vs Fail Closed

- **Fail open**: If Redis down, allow all requests (availability over protection)
- **Fail closed**: If Redis down, reject all (strict; use for critical APIs)

**Recommendation**: **Fail open** for most APIs; log failure for alerting.

---

## 7. Scaling

### Redis Scaling

- **Single Redis**: Up to ~100K ops/sec
- **Redis Cluster**: Shard by `identifier` (e.g., `hash(user_id) % 16384`); each shard handles subset
- **Replication**: Read replicas if needed (rate limit is write-heavy; replicas help for read-your-writes)

### Reducing Redis Load

- **Local cache**: Cache "allowed" for 1 second per user (optimistic); reduce Redis calls
- **Batch**: Not applicable for rate limit (each request is real-time)
- **Sliding window counter**: Fewer keys than sliding log (one counter per window)

### Distributed Deployment

- All API servers share same Redis
- Redis Cluster for horizontal scaling
- Consistent hashing for key distribution

---

## 8. Failure Handling

| Component | Failure | Mitigation |
|-----------|---------|------------|
| Redis | Down | Fail open (allow) or fail closed (reject); circuit breaker |
| Network | Latency | Timeout (e.g., 10ms); fail open on timeout |
| Lua script | Error | Fallback to non-atomic check (allow on error) |

### Redundancy

- Redis Sentinel or Cluster for HA
- Multiple Redis nodes

### Recovery

- Redis restart: Empty state; all users start fresh (acceptable for rate limit)
- Backfill: Not needed; rate limit is ephemeral

---

## 9. Monitoring & Observability

### Key Metrics

| Metric | Target | Alert |
|--------|--------|-------|
| Rate limit check latency | p99 < 5ms | Critical |
| Redis latency | p99 < 2ms | Warning |
| 429 rate | Monitor | Spike = abuse or limit too low |
| Redis connection errors | 0 | Critical |
| Fail-open events | Monitor | Alert if frequent |

### Dashboards

- 429 responses over time
- Rate limit check latency
- Redis ops/sec
- Top rate-limited users/IPs

### Alerting

- Redis down
- Rate limit latency > 10ms
- Sudden spike in 429s

---

## 10. Interview Tips

### Common Follow-Up Questions

1. **Sliding window vs fixed window?** Sliding is accurate; fixed has boundary burst
2. **How to handle race conditions?** Lua script for atomicity
3. **Fail open or closed?** Trade-off: availability vs protection
4. **How to scale Redis?** Cluster; shard by user
5. **Token bucket vs sliding window?** Token bucket allows bursts; sliding window is stricter

### What Interviewers Look For

- Algorithm choice and trade-offs
- Distributed state (Redis)
- Atomicity (Lua)
- Fail open/closed decision
- Multi-tier limits

### Common Mistakes

- Using fixed window (boundary burst)
- Not handling race conditions
- Ignoring fail open/closed
- Storing too much data (sliding log can grow)

---

## Appendix: Additional Design Considerations

### A. Sliding Window Counter Formula

```
previous_count = count in previous window
current_count = count in current window
window_start = start of current window
elapsed = now - window_start

sliding_count = previous_count * (1 - elapsed / window) + current_count
```

Example: Window=60s, previous=80, current=30, elapsed=30s
→ 80 * 0.5 + 30 = 70. If limit=100, allowed.

### B. Retry-After Header

- **Fixed window**: `Retry-After = window_end - now`
- **Sliding window**: `Retry-After = oldest_request_timestamp + window - now`
- **Token bucket**: `Retry-After = (1 - tokens) / refill_rate` (seconds to next token)

### C. Whitelist/Blacklist

- **Whitelist**: Skip rate limit for certain API keys (internal services)
- **Blacklist**: Always reject (blocked users)
- Store in Redis: `whitelist:{key}`, `blacklist:{key}`; check before rate limit

### D. Rate Limit by Endpoint

Different endpoints, different limits:

```json
{
  "/api/search": {"1m": 100},
  "/api/write": {"1m": 10, "1h": 100},
  "/api/export": {"1h": 5}
}
```

Key includes endpoint: `ratelimit:{user}:{endpoint}:{window}`

### E. Distributed Rate Limiter (No Redis)

If Redis not available:
- **In-memory per node**: Inaccurate (user can hit different nodes)
- **Consistent hashing**: Route same user to same node — possible but complex
- **Approximation**: Allow 2x limit when distributed (conservative)

**Best**: Use Redis or similar shared store for accuracy.

### F. Rate Limit Bypass for Internal Services

- **API key with bypass**: `X-RateLimit-Bypass: internal-service-secret`
- **IP whitelist**: Internal IPs (10.x, 172.16.x) skip rate limit
- **Header check**: Validate bypass token; log usage for audit

### G. Gradual Rollout of Rate Limits

- **Shadow mode**: Count violations but don't reject; measure impact
- **Percentage rollout**: 10% of users get rate limit; increase to 100%
- **Per-endpoint**: Start with write endpoints; add read limits later

### H. Rate Limit Headers Best Practices

```
X-RateLimit-Limit: 100        # Max requests in window
X-RateLimit-Remaining: 95      # Remaining in current window
X-RateLimit-Reset: 1710000060 # Unix timestamp when window resets
Retry-After: 45               # Only in 429 response
```

### I. Sliding Window Memory Optimization

Instead of storing every request timestamp (ZSET can grow large):
- **Sliding window counter**: Store only (prev_window_count, prev_window_start, current_count)
- **Memory**: O(1) per key
- **Trade-off**: Slight inaccuracy at window boundaries

### J. Rate Limiter Integration with API Gateway

```yaml
# Example: Kong API Gateway rate limiting
plugins:
  - name: rate-limiting
    config:
      minute: 100
      policy: redis
      redis_host: redis.example.com
      redis_port: 6379
      identifier: consumer
```

### K. Testing Rate Limits

- **Unit test**: Mock Redis; verify 429 when limit exceeded
- **Load test**: Simulate 1000 req/s from same user; verify 429 after limit
- **Chaos**: Kill Redis; verify fail-open behavior

### L. Rate Limit by User Tier

```json
{
  "free": {"1m": 60, "1h": 500},
  "pro": {"1m": 300, "1h": 10000},
  "enterprise": {"1m": 1000, "1h": 100000}
}
```

Key includes tier: `ratelimit:{user_id}:{tier}:{endpoint}:{window}`

### M. Observability for Rate Limiting

- **Metrics**: 429 rate by endpoint, by user, by IP
- **Logs**: Log every 429 with user_id, endpoint, timestamp
- **Alerts**: Spike in 429s may indicate abuse or misconfigured limit

### N. Complete Interview Walkthrough (45 min)

**0-5 min**: Clarify: limit by user/IP? configurable? distributed? multi-tier?
**5-10 min**: Estimates. 100K QPS, 50K unique users/min. Redis ops.
**10-15 min**: Algorithms. Fixed vs sliding window. Token bucket. Trade-offs.
**15-25 min**: Redis implementation. Lua script for atomicity. Sliding window log.
**25-35 min**: Architecture. Middleware. Redis. Fail open vs closed.
**35-40 min**: Scaling. Redis Cluster. Multi-tier limits. Race conditions.
**40-45 min**: Trade-offs. Accuracy vs memory. Fail open vs closed.

### O. Quick Reference Cheat Sheet

| Topic | Key Points |
|-------|------------|
| Algorithms | Sliding window (accurate) vs fixed (burst) vs token bucket |
| Atomicity | Lua script in Redis; no race conditions |
| Storage | Redis ZSET for sliding log; or counter for fixed |
| Fail | Open (availability) vs closed (strict) |
| Multi-tier | 10/s, 100/m, 1000/h; check all |

### P. Further Reading & Real-World Examples

- **Redis INCR**: Simple fixed-window; `INCR key` + `EXPIRE`
- **Kong**: API gateway with rate limiting plugin
- **Envoy**: Sidecar; local rate limit + Redis for distributed
- **AWS API Gateway**: Built-in throttling; per-account, per-stage

### Q. Design Alternatives Considered

| Decision | Alternative | Why Rejected |
|----------|-------------|--------------|
| Sliding window | Fixed window | Fixed has boundary burst |
| Redis | In-memory per node | In-memory doesn't work distributed |
| Lua script | Multiple Redis calls | Race condition without atomicity |
| Fail open | Fail closed | Availability over strict protection |

### R. Middleware Integration Example (Node.js)

```javascript
async function rateLimitMiddleware(req, res, next) {
  const key = req.user?.id || req.ip;
  const endpoint = req.path;
  const [limited, remaining] = await redis.eval(LUA_SLIDING_WINDOW, 1, key, endpoint, 60, 100, Date.now()/1000, uuid());
  if (limited) {
    return res.status(429).json({ error: 'Rate limit exceeded', retry_after: 60 });
  }
  res.set('X-RateLimit-Remaining', remaining);
  next();
}
```

### S. Fixed Window vs Sliding Window (Example)

**Fixed:**
- Window 1: 0-60s, 100 requests at 59s → allowed
- Window 2: 60-120s, 100 requests at 61s → allowed
- Result: 200 requests in 2 seconds!

**Sliding:**
- At 61s: Count requests in [1s, 61s] = 100 + 100 = 200 → rejected

### T. Redis Key Design for Rate Limiter

```
ratelimit:{identifier}:{endpoint}:{window_seconds}
Example: ratelimit:user_123:/api/search:60

For multi-tier, use separate keys:
ratelimit:user_123:/api/search:1
ratelimit:user_123:/api/search:60
ratelimit:user_123:/api/search:3600
```

### U. Rate Limit Response Best Practices

- Always include `X-RateLimit-*` headers on success
- On 429: Include `Retry-After` (seconds)
- Log 429s for analysis (abuse detection, limit tuning)

### V. Rate Limit Bypass Detection

- Monitor for users hitting limits frequently; may indicate need for higher tier
- Alert on sudden spike in 429s from single IP (possible DDoS or misconfigured client)
- Consider exponential backoff guidance in 429 response for clients

### W. Rate Limiter Deployment Options

1. **Embedded**: Library in application (e.g., guava RateLimiter) — single node only
2. **Middleware**: Express/FastAPI middleware calling Redis — distributed
3. **Sidecar**: Envoy proxy with rate limit service — infrastructure-level
4. **API Gateway**: Kong, AWS API Gateway — centralized, less flexible

### X. Summary

Rate limiter: Sliding window (accurate) or token bucket (allows bursts). Redis + Lua for atomic distributed state. Fail open for availability. Multi-tier limits (s, m, h). 429 with Retry-After.

---
*End of Rate Limiter System Design Document*

This document covers the design of a distributed rate limiter. Key takeaways: sliding window vs fixed window, Redis + Lua for atomicity, and fail-open for availability. The Lua script ensures no race conditions when multiple requests arrive simultaneously for the same user. Without atomicity, two requests could both read count=99 and both be allowed, exceeding the limit. Multi-tier limits (per-second, per-minute, per-hour) provide granular control and prevent both burst and sustained abuse.

**Document Version**: 1.0 | **Last Updated**: 2025-03-10 | **Target**: System Design Interview (Easy)

**Key Interview Questions to Prepare**:
- Sliding window vs fixed window: what's the difference?
- How do you prevent race conditions in Redis?
- Fail open vs fail closed: when to use each?
- How would you implement multi-tier rate limits?
- What happens when Redis goes down?
- How do you scale Redis for rate limiting?
- Token bucket vs sliding window: when to use each?
- Lua script: why atomic?


