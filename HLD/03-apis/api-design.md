# API Design Principles

## 1. Concept Overview

### Definition
API design encompasses the architectural decisions, conventions, and patterns used to create Application Programming Interfaces that enable software systems to communicate. Well-designed APIs are intuitive, consistent, scalable, and maintainable.

### Purpose
- **Interoperability**: Enable different systems (web, mobile, third-party) to integrate seamlessly
- **Developer Experience**: Reduce integration time and cognitive load for API consumers
- **Evolution**: Support backward-compatible changes without breaking existing clients
- **Security**: Enforce authentication, authorization, and data validation at the boundary

### Problems It Solves
- **Inconsistency**: Ad-hoc naming and structure lead to confusion and bugs
- **Over/Under-fetching**: Clients get too much or too little data, wasting bandwidth or requiring multiple round-trips
- **Versioning Chaos**: Breaking changes without clear migration paths
- **Error Ambiguity**: Generic errors that don't help clients recover
- **Performance**: Inefficient patterns (N+1, no pagination) that don't scale

---

## 2. Real-World Motivation

### Google
- **YouTube Data API v3**: RESTful design with `part` parameter for field selection (partial responses), quota-based rate limiting, OAuth 2.0
- **Google Cloud APIs**: Consistent resource naming (`projects/{project}/locations/{location}/...`), standard error format

### Netflix
- **Falcor**: Developed to solve over-fetching; evolved to GraphQL Federation for microservices
- **API Gateway**: Zuul handles routing, rate limiting, and request transformation at scale

### Uber
- **REST APIs**: Resource-oriented design for trips, riders, drivers; idempotency keys for ride creation
- **gRPC internally**: High-performance service-to-service communication

### Amazon
- **Product Advertising API**: Pagination via `ItemPage` parameter, batch operations for efficiency
- **AWS APIs**: Action-based naming (e.g., `DescribeInstances`), extensive use of idempotency

### Twitter
- **Twitter API v2**: Cursor-based pagination, field expansion, rate limit headers (`x-rate-limit-remaining`)
- **Tweet payload**: Sparse fields with `tweet.fields`, `user.fields` for partial responses

---

## 3. Architecture Diagrams

### RESTful Resource Hierarchy

```
                    ┌─────────────────────────────────────┐
                    │           API Root /api/v1           │
                    └─────────────────────────────────────┘
                                        │
            ┌───────────────────────────┼───────────────────────────┐
            │                           │                           │
            ▼                           ▼                           ▼
    ┌───────────────┐           ┌───────────────┐           ┌───────────────┐
    │   /users      │           │   /orders      │           │   /products   │
    │   Collection  │           │   Collection   │           │   Collection  │
    └───────────────┘           └───────────────┘           └───────────────┘
            │                           │                           │
            ▼                           ▼                           ▼
    ┌───────────────┐           ┌───────────────┐           ┌───────────────┐
    │ /users/{id}   │           │ /orders/{id}  │           │ /products/{id}│
    │   Resource    │           │   Resource    │           │   Resource    │
    └───────────────┘           └───────────────┘           └───────────────┘
            │                           │
            ▼                           ▼
    ┌───────────────┐           ┌───────────────┐
    │/users/{id}/   │           │/orders/{id}/  │
    │  orders       │           │  items        │
    │  Sub-resource │           │  Sub-resource │
    └───────────────┘           └───────────────┘
```

### Request Flow with Cross-Cutting Concerns

```
┌─────────┐     ┌──────────────┐     ┌─────────────┐     ┌──────────────┐
│ Client  │────▶│ API Gateway  │────▶│ Rate Limit  │────▶│ Auth/Token   │
└─────────┘     │ (Routing)     │     │ Middleware  │     │ Validation   │
               └──────────────┘     └─────────────┘     └──────────────┘
                                                                    │
                                                                    ▼
               ┌──────────────┐     ┌─────────────┐     ┌──────────────┐
               │ Response     │◀────│ Business    │◀────│ Idempotency  │
               │ Transform    │     │ Logic       │     │ Check        │
               └──────────────┘     └─────────────┘     └──────────────┘
```

### Pagination Strategies Comparison

```
OFFSET-BASED:                    CURSOR-BASED:
┌────────────────────┐           ┌────────────────────┐
│ Page 1: skip=0     │           │ cursor=abc123      │
│ Page 2: skip=20    │           │ next_cursor=def456  │
│ Page 3: skip=40    │           │ (stable, no drift)  │
│ (drift on inserts) │           └────────────────────┘
└────────────────────┘

KEYSET PAGINATION:
┌────────────────────┐
│ WHERE id > last_id  │
│ ORDER BY id         │
│ LIMIT 20            │
│ (efficient index)   │
└────────────────────┘
```

---

## 4. Core Mechanics

### RESTful Resource Naming Conventions

| Principle | Good | Bad |
|-----------|------|-----|
| Use nouns, not verbs | `GET /users` | `GET /getUsers` |
| Plural for collections | `GET /orders` | `GET /order` |
| Hierarchical for relationships | `GET /users/123/orders` | `GET /userOrders?userId=123` |
| Lowercase, hyphen-separated | `GET /order-items` | `GET /orderItems` |
| Avoid deep nesting (>2 levels) | `GET /users/123/orders` | `GET /accounts/1/users/123/orders/456` |

### HTTP Methods Mapping

| Method | Semantics | Idempotent | Safe | Request Body |
|--------|-----------|------------|------|--------------|
| GET | Retrieve resource(s) | Yes | Yes | No |
| POST | Create resource / Action | No | No | Yes |
| PUT | Replace resource (full) | Yes | No | Yes |
| PATCH | Partial update | No* | No | Yes |
| DELETE | Remove resource | Yes | No | Optional |

*PATCH idempotency depends on implementation; RFC 5789 recommends idempotent PATCH.

### Status Code Ranges

| Range | Meaning | Examples |
|-------|---------|----------|
| 2xx | Success | 200 OK, 201 Created, 204 No Content |
| 3xx | Redirection | 301 Moved Permanently, 304 Not Modified |
| 4xx | Client Error | 400 Bad Request, 401 Unauthorized, 404 Not Found, 429 Too Many Requests |
| 5xx | Server Error | 500 Internal Server Error, 503 Service Unavailable |

### Pagination Mechanics

**Offset-based** (simple, but problematic at scale):
```
GET /products?offset=100&limit=20
```
- Problem: Data can shift between requests (insertions/deletions cause duplicates or skips)
- Use when: Small datasets, admin UIs

**Cursor-based** (recommended for large datasets):
```
GET /products?cursor=eyJpZCI6MTAwfQ&limit=20
Response: { "data": [...], "next_cursor": "eyJpZCI6MTIwfQ", "has_more": true }
```
- Cursor is opaque (often base64-encoded last item ID or composite key)
- Stable across concurrent modifications

**Keyset/Seek** (database-efficient):
```
GET /products?after_id=100&limit=20
-- SQL: SELECT * FROM products WHERE id > 100 ORDER BY id LIMIT 20
```
- Uses indexed columns; O(1) instead of O(offset) for database

---

## 5. Numbers

| Metric | Typical Value | Notes |
|--------|---------------|-------|
| API response time (p50) | 50-200ms | Depends on backend complexity |
| API response time (p99) | 200-1000ms | Tail latencies matter for UX |
| Payload size (REST) | 1-100KB typical | Use field selection to reduce |
| Pagination default | 20-50 items | Balance UX and load |
| Max pagination limit | 100-1000 | Prevent abuse |
| Version deprecation | 6-12 months | Industry standard |
| Rate limit (per user) | 100-10,000 req/min | Tiered by plan |

---

## 6. Tradeoffs

### Versioning Strategies

| Strategy | Pros | Cons |
|----------|------|------|
| URI (`/v1/users`) | Explicit, cacheable, easy to route | URL proliferation |
| Header (`Accept: application/vnd.api+json;version=1`) | Clean URLs | Less visible, caching complexity |
| Query param (`?version=1`) | Simple | Often ignored, cache key issues |

### Pagination Tradeoffs

| Approach | Consistency | Performance | Complexity |
|----------|-------------|-------------|------------|
| Offset | Poor (drift) | Degrades with offset | Low |
| Cursor | Good | Stable | Medium |
| Keyset | Good | Best (indexed) | Medium |

### Error Handling Tradeoffs

| Approach | Machine-readable | Human-readable | Standard |
|----------|------------------|----------------|----------|
| Custom JSON | Yes | Varies | No |
| RFC 7807 Problem Details | Yes | Yes | Yes |
| HTTP status only | Partial | No | Yes |

---

## 7. Variants / Implementations

### Versioning Implementations

**URI Versioning** (most common):
```
https://api.example.com/v1/users
https://api.example.com/v2/users  # New version
```

**Header Versioning**:
```
GET /users
Accept: application/vnd.example.v1+json
```

**Query Parameter**:
```
GET /users?api_version=1
```

### Error Response Formats

**RFC 7807 Problem Details**:
```json
{
  "type": "https://api.example.com/errors/validation",
  "title": "Validation Error",
  "status": 400,
  "detail": "Email format is invalid",
  "instance": "/users",
  "errors": [
    { "field": "email", "message": "Must be valid email format" }
  ]
}
```

### Partial Response (Field Selection)

**Google-style `fields` parameter**:
```
GET /users/123?fields=id,name,email
```

**GraphQL-style (REST) `expand`**:
```
GET /orders/456?expand=items,shipping_address
```

---

## 8. Scaling Strategies

1. **Horizontal scaling**: Stateless API design enables adding instances
2. **Caching**: ETags, `Cache-Control`, `If-None-Match` for conditional requests
3. **Pagination**: Prevent large payloads; use cursor-based for consistency
4. **Field selection**: Reduce payload size and backend load
5. **Bulk operations**: `POST /batch` for multiple operations in one request
6. **Async for long operations**: Return 202 Accepted with `Location` header for status polling

---

## 9. Failure Scenarios

### Production Failures and Mitigations

| Failure | Impact | Mitigation |
|---------|--------|------------|
| Breaking change deployed | All clients break | Versioning, deprecation periods, feature flags |
| Missing idempotency | Duplicate charges/orders | Idempotency keys for POST/PATCH |
| Poor error messages | Clients can't recover | RFC 7807, actionable error codes |
| No rate limiting | DoS, resource exhaustion | Token bucket, tiered limits |
| Over-fetching | Slow responses, bandwidth | Field selection, sparse fieldsets |
| Pagination drift | Duplicate/missing items | Cursor-based pagination |

---

## 10. Performance Considerations

- **Compression**: Enable gzip/brotli for responses >1KB
- **Connection pooling**: HTTP/2 or keep-alive for multiple requests
- **Caching**: `ETag` + `If-None-Match` for conditional GET; `Cache-Control` headers
- **Database**: Avoid N+1; use batch loading, keyset pagination
- **Timeouts**: Set reasonable client and server timeouts (e.g., 30s)

---

## 11. Use Cases

| System | API Style | Key Design Choices |
|--------|-----------|-------------------|
| Stripe | REST | Idempotency keys, webhooks, expandable objects |
| GitHub | REST + GraphQL | GraphQL for complex queries; REST for simple CRUD |
| Slack | REST + WebSocket | REST for actions; WebSocket for real-time |
| Twilio | REST | Resource-oriented, webhooks for async |
| Shopify | REST | Versioned, rate limited, webhooks |

---

## 12. Comparison Tables

### Good vs Bad API Design

| Aspect | Good | Bad |
|--------|------|-----|
| Resource naming | `GET /users/123/orders` | `GET /getUserOrders?userId=123` |
| Status codes | 201 for create, 204 for delete | 200 for everything |
| Errors | Structured with field-level details | "Error" string |
| Pagination | Cursor with `next_cursor` | Offset only, no total |
| Versioning | `/v1/` in path | No versioning, breaking changes |
| Idempotency | `Idempotency-Key` header for POST | No support, duplicate risk |

### HATEOAS (Hypermedia as the Engine of Application State)

| Benefit | Description |
|---------|-------------|
| Discoverability | Links in response guide client to next actions |
| Decoupling | Server controls URLs; client follows links |
| Evolution | New links can be added without breaking clients |

Example:
```json
{
  "id": "123",
  "status": "pending",
  "_links": {
    "self": { "href": "/orders/123" },
    "cancel": { "href": "/orders/123/cancel", "method": "POST" },
    "pay": { "href": "/orders/123/pay", "method": "POST" }
  }
}
```

---

## 13. Code or Pseudocode

### Idempotency Key Middleware

```python
def idempotency_middleware(request, handler):
    if request.method not in ['POST', 'PATCH', 'PUT']:
        return handler(request)
    
    key = request.headers.get('Idempotency-Key')
    if not key:
        return handler(request)  # Or require for certain endpoints
    
    cache_key = f"idempotency:{key}"
    cached = redis.get(cache_key)
    if cached:
        return Response.from_cache(cached)  # Return stored response
    
    response = handler(request)
    if response.status_code in [200, 201]:
        redis.setex(cache_key, 86400, response.serialize())  # 24h TTL
    
    return response
```

### Cursor-Based Pagination

```python
def list_products(cursor=None, limit=20):
    query = Product.query.order_by(Product.id)
    if cursor:
        last_id = decode_cursor(cursor)
        query = query.filter(Product.id > last_id)
    
    items = query.limit(limit + 1).all()
    has_more = len(items) > limit
    if has_more:
        items = items[:limit]
    
    next_cursor = encode_cursor(items[-1].id) if has_more else None
    return {"data": items, "next_cursor": next_cursor, "has_more": has_more}
```

### ETag Caching

```python
def get_user(user_id):
    user = db.get_user(user_id)
    etag = hashlib.md5(json.dumps(user, sort_keys=True).encode()).hexdigest()
    
    if request.headers.get('If-None-Match') == etag:
        return Response(status=304)  # Not Modified
    
    return Response(json=user, headers={'ETag': etag})
```

### Bulk Operations Design

```python
# POST /orders/batch
# Body: { "operations": [ {"type": "create", "data": {...}}, ... ] }
# Response: { "results": [ {"id": "ord_1", "status": 201}, {"id": null, "status": 400, "error": "..."} ] }

def batch_create_orders(operations):
    results = []
    for op in operations:
        try:
            order = create_order(op["data"])
            results.append({"id": order.id, "status": 201})
        except ValidationError as e:
            results.append({"id": None, "status": 400, "error": str(e)})
    return results
```

### OpenAPI/Swagger Documentation Structure

```yaml
openapi: 3.0.0
paths:
  /users/{id}:
    get:
      summary: Get user by ID
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Success
        '404':
          description: Not found
```

---

## 14. Interview Discussion

### How to Explain
Start with the problem: "APIs are contracts between systems. Poor design leads to integration pain, breaking changes, and performance issues." Then cover: resource naming (nouns, hierarchy), HTTP semantics (methods, status codes), pagination (cursor vs offset), versioning (URI vs header), and error handling (RFC 7807). Emphasize consistency and backward compatibility.

### Follow-up Questions
- "How would you design pagination for a feed with 100M items?"
- "What's the difference between PUT and PATCH? When is PATCH not idempotent?"
- "How do you handle API versioning when you need to change a response schema?"
- "Design an API for bulk creating 10,000 orders. What are the tradeoffs?"
- "How would you implement partial response (field selection) in a REST API?"

---

## Appendix A: HTTP Status Code Deep Dive

| Code | Meaning | When to Use |
|------|---------|-------------|
| 200 OK | Success, body contains representation | GET, PUT, PATCH success |
| 201 Created | Resource created, Location header | POST success |
| 204 No Content | Success, no body | DELETE success |
| 400 Bad Request | Malformed request | Validation errors |
| 401 Unauthorized | Not authenticated | Missing/invalid token |
| 403 Forbidden | Authenticated but not authorized | Insufficient permissions |
| 404 Not Found | Resource doesn't exist | Invalid ID |
| 409 Conflict | State conflict | Duplicate, version mismatch |
| 422 Unprocessable Entity | Semantic validation failed | Business rule violation |
| 429 Too Many Requests | Rate limited | Retry-After header |
| 500 Internal Server Error | Unexpected server error | Log and alert |
| 503 Service Unavailable | Temporary overload | Retry later |

---

## Appendix B: API Versioning Migration Strategy

1. **Phase 1**: Deploy v2 alongside v1; both active
2. **Phase 2**: Deprecation headers; `Sunset` header with date
3. **Phase 3**: Client migration; support both for 6-12 months
4. **Phase 4**: Disable v1; return 410 Gone for v1 requests

---

## Appendix C: Request/Response Design Patterns

### Request Design
- **Consistent casing**: camelCase for JSON (JavaScript convention) or snake_case (Python)
- **Nested vs flat**: Prefer flat for simple; nested for hierarchies
- **Optional vs required**: Document clearly; validate server-side

### Response Design
- **Envelope**: `{ "data": {...}, "meta": {...}, "errors": [...] }` for consistency
- **Pagination metadata**: `page`, `per_page`, `total`, `next_cursor`
- **Error envelope**: `{ "error": { "code": "...", "message": "...", "details": [...] } }`

---

## Appendix D: HATEOAS Example Extended

```json
{
  "id": "ord_123",
  "status": "pending",
  "total": 99.99,
  "_links": {
    "self": { "href": "/orders/ord_123", "method": "GET" },
    "cancel": { "href": "/orders/ord_123/cancel", "method": "POST" },
    "pay": { "href": "/orders/ord_123/pay", "method": "POST" },
    "items": { "href": "/orders/ord_123/items", "method": "GET" }
  }
}
```

Client discovers actions from links; no hardcoded URLs.
