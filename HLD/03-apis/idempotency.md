# Idempotency

## 1. Concept Overview

### Definition
Idempotency means that performing the same operation multiple times produces the same result as performing it once. For APIs: sending the same request (with same idempotency key) multiple times has the same effect as sending it once.

### Purpose
- **Safe retries**: Clients can retry without creating duplicates (charges, orders, records)
- **Network reliability**: Requests can be duplicated (timeouts, retries); idempotency ensures correctness
- **Exactly-once semantics**: Achieve "process once" despite at-least-once delivery

### Problems It Solves
- **Duplicate charges**: User double-clicks "Pay"; two charges without idempotency
- **Duplicate orders**: Retry after timeout creates second order
- **Duplicate messages**: Message queue redelivery; consumer must handle idempotently

---

## 2. Real-World Motivation

### Stripe
- **Idempotency keys**: Required for payment-related operations. Client sends `Idempotency-Key: unique-key` header. Same key within 24h returns cached response. Used for payments, subscriptions, refunds.

### AWS S3
- **Conditional writes**: `If-Match` (ETag) for optimistic concurrency; prevents overwrite if object changed. PutObject is idempotent with same key (last write wins).

### Kafka
- **Exactly-once semantics**: Idempotent producer (deduplication by producer ID + sequence); transactional commits. Consumer idempotency via deterministic processing + external store.

### Uber
- **Ride creation**: Idempotency key prevents duplicate ride requests when user retries or network fails.

---

## 3. Architecture Diagrams

### Idempotency Key Flow

```
Client                          Server
  │                                │
  │  POST /charges                 │
  │  Idempotency-Key: key-123      │
  │  Body: {amount: 100}           │
  │──────────────────────────────▶│
  │                                │  Check: key-123 in store?
  │                                │  No → Process, store response
  │                                │
  │◀──201 {id: ch_1}──────────────│
  │                                │
  │  (Retry - same request)        │
  │  POST /charges                 │
  │  Idempotency-Key: key-123      │
  │──────────────────────────────▶│
  │                                │  Check: key-123 in store?
  │                                │  Yes → Return cached response
  │◀──201 {id: ch_1}──────────────│  (same response, no new charge)
```

### Database-Level Idempotency

```
INSERT INTO orders (idempotency_key, user_id, total, ...)
VALUES ('key-123', 1, 100, ...)
ON CONFLICT (idempotency_key) DO UPDATE SET updated_at = NOW()
RETURNING *;

-- Or: unique constraint on idempotency_key
-- Second insert fails → catch, return existing row
```

### Message Processing Idempotency

```
Message Queue ──▶ Consumer
                     │
                     ▼
              ┌──────────────┐
              │ Check: msg_id│
              │ in processed?│
              └──────┬───────┘
                     │
         Yes ────────┼──────── No
         │           │           │
         ▼           │           ▼
    Skip (ack)       │      Process
                     │           │
                     │           ▼
                     │      Store msg_id
                     │      in processed
                     │           │
                     │           ▼
                     │      Ack
```

---

## 4. Core Mechanics

### Idempotent HTTP Methods

| Method | Idempotent? | Notes |
|--------|-------------|-------|
| GET | Yes | Same request → same response |
| PUT | Yes | Same body → same state |
| DELETE | Yes | Delete same resource multiple times = same result |
| POST | No | Creates new resource each time |
| PATCH | Depends | Can be idempotent if semantics defined |

### Idempotency Key Implementation

1. **Client**: Generates unique key (UUID) per logical operation; sends in header
2. **Server**: Before processing, check if key exists in store (Redis, DB)
3. **If exists**: Return stored response (same status, body)
4. **If not**: Process request, store response with key, return
5. **TTL**: Expire keys after 24-72h (Stripe: 24h)

### Storage

- **Redis**: Fast; key → serialized response; TTL
- **Database**: Persistent; can survive restarts; unique constraint

---

## 5. Numbers

| Metric | Typical Value |
|--------|---------------|
| Key TTL | 24-72 hours |
| Storage size | ~1-10KB per key (response) |
| Lookup latency | <1ms (Redis) |

---

## 6. Tradeoffs

### Idempotency Key Scope

| Scope | Pros | Cons |
|-------|------|------|
| Per-request | Fine-grained | Client must generate per operation |
| Per-session | Simpler | Fewer operations covered |

### Storage Tradeoffs

| Store | Speed | Persistence | Cost |
|-------|-------|-------------|------|
| Redis | Fast | Volatile | Low |
| DB | Slower | Durable | Medium |

---

## 7. Variants / Implementations

### Stripe-Style
- Header: `Idempotency-Key: <key>`
- Store: Response + status
- TTL: 24h

### Database Upsert
- Unique constraint on business key (e.g., `order_id` + `user_id`)
- `INSERT ... ON CONFLICT DO NOTHING` or `DO UPDATE`

### Message Deduplication
- Store processed message IDs

---

## 8. Scaling Strategies

- **Redis cluster**: Shard by key hash
- **DB**: Index on idempotency_key

---

## 9. Failure Scenarios

| Failure | Impact | Mitigation |
|---------|--------|------------|
| Key collision (different ops) | Wrong response returned | Client must use unique key per operation |
| Store failure | Can't check | Fail open (process) or fail closed (reject) |
| Key expired | Duplicate processed | Longer TTL; or client retries with same key |

---

## 10. Performance Considerations

- **Lookup**: Redis GET is O(1); add before every mutating request
- **Storage**: Limit response size stored; or store reference only

---

## 11. Use Cases

| Use Case | Implementation |
|----------|-----------------|
| Payment | Idempotency key (Stripe) |
| Order creation | Idempotency key or unique constraint |
| Message processing | Deduplication store |
| File upload | Resumable upload with offset |

---

## 12. Comparison Tables

### Exactly-Once vs At-Least-Once + Idempotency

| Approach | Guarantee | Complexity |
|----------|------------|------------|
| Exactly-once | Process once | High (distributed transactions) |
| At-least-once + Idempotency | Process once (effect) | Lower; consumer must be idempotent |

---

## 13. Code or Pseudocode

### Idempotency Key Middleware

```python
def idempotency_middleware(request, handler):
    if request.method not in ['POST', 'PATCH', 'PUT']:
        return handler(request)
    
    key = request.headers.get('Idempotency-Key')
    if not key:
        return handler(request)
    
    cache_key = f"idempotency:{key}"
    cached = redis.get(cache_key)
    if cached:
        return Response.from_cache(cached)
    
    response = handler(request)
    if response.status_code in [200, 201]:
        redis.setex(cache_key, 86400, response.serialize())
    
    return response
```

### Database Upsert (Unique Constraint)

```sql
INSERT INTO orders (idempotency_key, user_id, total, status)
VALUES ($1, $2, $3, 'pending')
ON CONFLICT (idempotency_key) DO UPDATE SET updated_at = NOW()
RETURNING *;
```

### Message Deduplication

```python
def process_message(msg):
    if redis.sadd("processed", msg.id) == 0:
        return  # Already processed
    try:
        do_work(msg)
    except:
        redis.srem("processed", msg.id)
        raise
```

---

## 14. Interview Discussion

### How to Explain
"Idempotency means repeating an operation has the same effect as doing it once. For APIs, we use idempotency keys: client sends a unique key; server stores key→response. Retries with same key return cached response. Essential for payments, orders."

### Follow-up Questions
- "How do you handle idempotency key expiration?"
- "What if two requests with same key arrive concurrently?"
- "Design idempotency for a message queue consumer."

---

## Appendix A: Concurrent Request Handling

When two requests with same idempotency key arrive simultaneously:

```
Request A (key-123) ──┐
                      ├──▶ Both check Redis: key not found
Request B (key-123) ──┘
                      │
                      ├──▶ Use distributed lock (Redis SETNX) or
                      │    database unique constraint
                      │
                      └──▶ First to acquire lock processes;
                           second waits, then gets cached result
```

**Solution**: Redis `SET key NX EX` (set if not exists) as lock. First request acquires, processes, stores response. Second request waits or retries, then finds cached response.

---

## Appendix B: Idempotency Key Best Practices

| Practice | Reason |
|----------|--------|
| Client-generated | Server can't know "same" across retries |
| UUID or random | Uniqueness per logical operation |
| Scope to operation | Don't reuse across different operations |
| Include in logs | Debugging duplicate issues |

---

## Appendix C: Database Idempotency Patterns

**Unique constraint**:
```sql
CREATE UNIQUE INDEX ON orders (idempotency_key);
-- INSERT fails on duplicate → catch, SELECT existing
```

**Upsert**:
```sql
INSERT INTO orders (idempotency_key, ...) VALUES (...)
ON CONFLICT (idempotency_key) DO UPDATE SET updated_at = NOW()
RETURNING *;
-- Returns existing row on conflict
```

**Optimistic locking** (for updates):
```sql
UPDATE orders SET total = 100 WHERE id = 1 AND version = 5;
-- If version changed, 0 rows updated → retry or conflict
```

---

## Appendix D: Exactly-Once Semantics in Distributed Systems

| Layer | Mechanism |
|-------|-----------|
| Producer | Idempotent producer (Kafka: producer ID + seq) |
| Transport | At-most-once, at-least-once, or exactly-once |
| Consumer | Idempotent processing + deduplication store |

**Kafka exactly-once**: Idempotent producer + transactional consumer (read-process-write in single transaction). Consumer commits offset only after successful processing.

---

## Appendix E: Stripe Idempotency Key Semantics

- **Scope**: Per API key, per endpoint
- **TTL**: 24 hours
- **Same key, different params**: Returns error (key already used for different request)
- **Same key, same params**: Returns cached response
- **Key format**: Any string; recommend UUID

---

## Appendix F: Idempotency for Partial Failures

Bulk operation: 10 items, item 5 fails. Retry with same key:
- **Option A**: Re-process all 10; items 1-4 idempotent (no-op), 5 retried, 6-10 idempotent
- **Option B**: Store per-item results; on retry, skip 1-4, retry 5, skip 6-10
- **Option B** requires storing partial state keyed by idempotency key + item index

---

## Appendix G: Idempotency Key Storage Schema

```sql
CREATE TABLE idempotency_keys (
    key VARCHAR(255) PRIMARY KEY,
    response_status INT,
    response_body JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP
);

CREATE INDEX idx_expires ON idempotency_keys(expires_at);
-- Cleanup job deletes expired keys
```

---

## Appendix H: Idempotency for Updates (PATCH)

PATCH can be idempotent if semantics are defined:
- "Set field X to value Y" → idempotent (same result if repeated)
- "Increment field X by 1" → NOT idempotent

Use idempotency key for PATCH when the operation has side effects (e.g., sending email).

---

## Appendix I: Idempotency in Event Sourcing

In event sourcing, idempotency = deduplication by event ID:
- Each event has unique ID
- Before applying: check if event ID already in store
- If yes: skip (already applied)
- If no: apply, store event ID

---

## Appendix J: Idempotency Key Header Alternatives

| Header | Used By |
|--------|---------|
| Idempotency-Key | Stripe, standard |
| X-Idempotency-Key | Some implementations |
| X-Request-Id | Sometimes used (less standard) |

Recommend: `Idempotency-Key` per Stripe convention.

---

## Appendix K: Testing Idempotency

1. **Same key, same request**: Expect cached response, no duplicate side effects
2. **Same key, different request**: Expect error (key already used)
3. **Different key, same request**: Expect new processing
4. **Concurrent same key**: One processes, one gets cached (or both get same result)

---

## Appendix L: Idempotency for DELETE

DELETE is idempotent by HTTP spec: deleting same resource multiple times = same result (resource gone).
- First DELETE: 204 or 200
- Subsequent: 404 (resource not found) or 204 (already gone)
- Both are "success" from idempotency perspective

---

## Appendix M: Idempotency Key Generation (Client)

```javascript
// Good: UUID per operation
const idempotencyKey = crypto.randomUUID();

// Good: Deterministic from operation
const idempotencyKey = `order-${userId}-${cartId}-${timestamp}`;

// Bad: Reusing across operations
const idempotencyKey = "my-key";  // Will collide
```

---

## Appendix N: Idempotency Scope by Endpoint

| Endpoint | Idempotency Key Required? |
|----------|---------------------------|
| POST /charges | Yes (payment) |
| POST /orders | Yes (order creation) |
| POST /refunds | Yes (financial) |
| GET /users | No (read-only) |
| PUT /users/123 | No (idempotent by method) |
| PATCH /users/123 | Optional (if side effects) |

---

## Appendix O: Idempotency and Distributed Transactions

In saga pattern, each step should be idempotent:
- Step 1: Create order (idempotency key)
- Step 2: Reserve inventory (idempotent: same order_id = no-op)
- Step 3: Charge payment (idempotency key)
- Compensating actions also idempotent

---

## Appendix P: Idempotency Key Storage Size Considerations

- Store only essential response (status + body); avoid storing large payloads
- Set TTL to reclaim space (24-72h typical)
- Consider separate storage for high-volume: Redis for hot keys, archive to cold storage if needed for audit

---

## Appendix Q: Idempotency in REST vs gRPC

| Protocol | Idempotency |
|----------|-------------|
| REST POST | Requires Idempotency-Key header (custom) |
| gRPC | Method-level; CreateUser is not idempotent by default; add custom key in request message |
