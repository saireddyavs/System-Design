# REST API Deep Dive

## HTTP Methods, Status Codes, Headers, Content Negotiation, HATEOAS, and Idempotency

---

## 1. HTTP Methods — Complete Reference

### Method Properties

| Method | Safe? | Idempotent? | Has Body? | Cacheable? | Purpose |
|--------|-------|-------------|-----------|------------|---------|
| `GET` | ✅ Yes | ✅ Yes | ❌ No | ✅ Yes | Retrieve resource(s) |
| `HEAD` | ✅ Yes | ✅ Yes | ❌ No | ✅ Yes | Get headers only (no body) |
| `OPTIONS` | ✅ Yes | ✅ Yes | ❌ No | ❌ No | Discover allowed methods (CORS preflight) |
| `POST` | ❌ No | ❌ No | ✅ Yes | ❌ No | Create resource / trigger action |
| `PUT` | ❌ No | ✅ Yes | ✅ Yes | ❌ No | Replace resource entirely |
| `PATCH` | ❌ No | ❌ No* | ✅ Yes | ❌ No | Partial update |
| `DELETE` | ❌ No | ✅ Yes | Optional | ❌ No | Remove resource |

**Safe** = No side effects (read-only). **Idempotent** = Multiple identical requests = same result as single request.

### When to Use Each Method

#### GET — Retrieve

```
GET /api/v1/users/123
→ 200 OK { "id": "123", "name": "Alice", ... }

GET /api/v1/users?status=active&sort=name:asc&limit=20
→ 200 OK { "data": [...], "meta": { "total": 150 } }

GET /api/v1/users/123/orders?status=completed
→ 200 OK { "data": [...] }
```

- Never use GET for operations with side effects
- Never put sensitive data in URL (logged by servers, proxies)
- Support filtering, sorting, pagination via query parameters

#### POST — Create / Action

```
// Create a new resource
POST /api/v1/users
Body: { "name": "Alice", "email": "alice@example.com" }
→ 201 Created
  Location: /api/v1/users/123
  Body: { "id": "123", "name": "Alice", ... }

// Trigger an action (not creating a resource)
POST /api/v1/orders/456/cancel
→ 200 OK { "id": "456", "status": "cancelled" }

// Search with complex query (when GET query string is too long)
POST /api/v1/search/products
Body: { "query": "wireless headphones", "filters": { ... }, "facets": [...] }
→ 200 OK { "data": [...] }
```

- Return `201 Created` with `Location` header for resource creation
- Return `200 OK` for actions
- Return `202 Accepted` for async operations
- Use `Idempotency-Key` header to make POST idempotent

#### PUT — Full Replace

```
// Replace the ENTIRE resource
PUT /api/v1/users/123
Body: { "name": "Alice Smith", "email": "alice@example.com", "phone": "+1234567890" }
→ 200 OK { "id": "123", "name": "Alice Smith", ... }

// If resource doesn't exist, PUT can create it (upsert)
PUT /api/v1/configs/feature-flags
Body: { "dark_mode": true, "beta_features": false }
→ 201 Created (if new) or 200 OK (if replaced)
```

- Client must send the **complete** resource (any missing fields are set to null/default)
- Idempotent: sending the same PUT twice = same result
- Use for replacing known resources, not for partial updates

#### PATCH — Partial Update

```
// Update only specific fields
PATCH /api/v1/users/123
Body: { "phone": "+1234567890" }  // only update phone
→ 200 OK { "id": "123", "name": "Alice", "phone": "+1234567890", ... }

// JSON Patch (RFC 6902) — more explicit
PATCH /api/v1/users/123
Content-Type: application/json-patch+json
Body: [
  { "op": "replace", "path": "/phone", "value": "+1234567890" },
  { "op": "add", "path": "/preferences/theme", "value": "dark" },
  { "op": "remove", "path": "/old_field" }
]

// JSON Merge Patch (RFC 7396) — simpler
PATCH /api/v1/users/123
Content-Type: application/merge-patch+json
Body: { "phone": "+1234567890", "old_field": null }  // null = remove
```

#### DELETE — Remove

```
// Delete a resource
DELETE /api/v1/users/123
→ 204 No Content  (no body)

// Or return the deleted resource
DELETE /api/v1/users/123
→ 200 OK { "id": "123", "name": "Alice", "deleted_at": "..." }

// Idempotent: deleting already-deleted resource
DELETE /api/v1/users/123
→ 404 Not Found  (or 204 — both are valid)
```

**Interview Tip:** Discuss soft delete vs hard delete:
- **Soft delete**: Set `deleted_at` timestamp, filter in queries
- **Hard delete**: Actually remove from database
- Choose based on audit requirements, data retention policies

---

## 2. HTTP Status Codes — Complete Guide

### Success (2xx)

| Code | Meaning | When to Use |
|------|---------|-------------|
| `200 OK` | Success | GET, PUT, PATCH, DELETE (with body) |
| `201 Created` | Resource created | POST (include `Location` header) |
| `202 Accepted` | Request accepted for processing | Async operations |
| `204 No Content` | Success, no body | DELETE, PUT (when no body needed) |
| `206 Partial Content` | Partial range | Video streaming, large file downloads |

### Redirection (3xx)

| Code | Meaning | When to Use |
|------|---------|-------------|
| `301 Moved Permanently` | Resource URL changed permanently | URL migration |
| `302 Found` | Temporary redirect | OAuth callbacks |
| `304 Not Modified` | Use cached version | ETag/If-None-Match validation |
| `307 Temporary Redirect` | Redirect with same method | Maintains POST (unlike 302) |
| `308 Permanent Redirect` | Permanent redirect with same method | API migration |

### Client Error (4xx)

| Code | Meaning | When to Use |
|------|---------|-------------|
| `400 Bad Request` | Invalid request | Malformed JSON, missing required fields |
| `401 Unauthorized` | Not authenticated | Missing/invalid/expired token |
| `403 Forbidden` | Authenticated but not authorized | Insufficient permissions |
| `404 Not Found` | Resource doesn't exist | Invalid ID, deleted resource |
| `405 Method Not Allowed` | Wrong HTTP method | POST on read-only resource |
| `406 Not Acceptable` | Unsupported Accept header | Client wants XML, server only does JSON |
| `409 Conflict` | State conflict | Duplicate email, race condition |
| `410 Gone` | Permanently deleted | Explicitly removed, not coming back |
| `412 Precondition Failed` | Precondition header failed | `If-Match` ETag mismatch (optimistic lock) |
| `413 Content Too Large` | Body too large | Upload size exceeded |
| `415 Unsupported Media Type` | Wrong Content-Type | Sent XML instead of JSON |
| `422 Unprocessable Entity` | Validation error | Well-formed JSON but invalid data |
| `429 Too Many Requests` | Rate limited | Include `Retry-After` header |

### Server Error (5xx)

| Code | Meaning | When to Use |
|------|---------|-------------|
| `500 Internal Server Error` | Server bug | Unhandled exception |
| `502 Bad Gateway` | Upstream failure | Backend service down |
| `503 Service Unavailable` | Temporarily unavailable | Maintenance, overloaded |
| `504 Gateway Timeout` | Upstream timeout | Backend too slow |

### Interview Tip: Common Mistakes

```
❌ 200 for everything (even errors)
   → Use appropriate 4xx/5xx status codes

❌ 401 when you mean 403
   → 401 = "who are you?" (not authenticated)
   → 403 = "I know who you are, but you can't do this" (not authorized)

❌ 404 when you mean 400
   → 404 = resource not found (valid URL structure, resource doesn't exist)
   → 400 = bad request (malformed input)

❌ 500 for user input errors
   → 500 = server's fault, not the client's
   → Use 400/422 for validation errors
```

---

## 3. HTTP Headers — Essential Knowledge

### Request Headers

| Header | Purpose | Example |
|--------|---------|---------|
| `Authorization` | Authentication token | `Bearer eyJhbGci...` |
| `Content-Type` | Body format | `application/json` |
| `Accept` | Desired response format | `application/json` |
| `Accept-Encoding` | Compression preference | `gzip, deflate, br` |
| `Accept-Language` | Language preference | `en-US, en;q=0.9` |
| `If-None-Match` | Conditional GET (caching) | `"33a64df5"` (ETag value) |
| `If-Match` | Conditional update (optimistic locking) | `"33a64df5"` |
| `If-Modified-Since` | Conditional GET (date-based) | `Mon, 15 Jan 2024 10:30:00 GMT` |
| `Idempotency-Key` | Make POST idempotent | `550e8400-e29b-41d4-a716-446655440000` |
| `X-Request-ID` | Request tracing | `req_abc123` |
| `User-Agent` | Client identification | `MyApp/1.0 (iOS 17.2)` |

### Response Headers

| Header | Purpose | Example |
|--------|---------|---------|
| `Location` | URL of created resource | `/api/v1/users/123` |
| `ETag` | Resource version (hash) | `"33a64df5"` |
| `Last-Modified` | Last modification time | `Mon, 15 Jan 2024 10:30:00 GMT` |
| `Cache-Control` | Caching directives | `max-age=3600, public` |
| `Content-Type` | Body format | `application/json; charset=utf-8` |
| `Content-Encoding` | Compression used | `gzip` |
| `X-RateLimit-Limit` | Rate limit max | `100` |
| `X-RateLimit-Remaining` | Requests remaining | `98` |
| `X-RateLimit-Reset` | When limit resets | `1705312800` (Unix timestamp) |
| `Retry-After` | When to retry (rate limited) | `60` (seconds) |
| `X-Request-ID` | Request tracing (echo back) | `req_abc123` |
| `Link` | Pagination links | `<...?cursor=abc>; rel="next"` |

---

## 4. Content Negotiation

```
Client sends:
  Accept: application/json        → wants JSON
  Accept-Encoding: gzip           → wants compressed
  Accept-Language: en-US           → wants English

Server responds:
  Content-Type: application/json; charset=utf-8
  Content-Encoding: gzip
  Content-Language: en-US
  Vary: Accept, Accept-Encoding    → tells cache to vary by these headers
```

### Media Type Versioning (alternative to URL versioning)

```
// Custom media type for versioning
Accept: application/vnd.myapp.v2+json

// GitHub's approach:
Accept: application/vnd.github.v3+json
```

---

## 5. Caching

### ETag-Based Caching (Conditional Requests)

```
Flow:
┌────────┐                    ┌────────┐
│ Client │                    │ Server │
└────┬───┘                    └────┬───┘
     │ GET /users/123              │
     │────────────────────────────▶│
     │                             │  Compute ETag
     │ 200 OK                      │
     │ ETag: "abc123"              │
     │◀────────────────────────────│
     │                             │
     │ ... later ...               │
     │ GET /users/123              │
     │ If-None-Match: "abc123"     │
     │────────────────────────────▶│
     │                             │  ETag matches → not modified
     │ 304 Not Modified            │
     │ (no body — save bandwidth)  │
     │◀────────────────────────────│
```

### Cache-Control Directives

```
Cache-Control: public, max-age=3600           → CDN + browser cache for 1 hour
Cache-Control: private, max-age=300           → browser only, 5 minutes
Cache-Control: no-cache                       → always revalidate with server
Cache-Control: no-store                       → never cache (sensitive data)
Cache-Control: max-age=86400, stale-while-revalidate=3600
  → Use cache for 24h, then serve stale for 1h while revalidating
```

### Optimistic Locking with ETag

```
// Read resource with ETag
GET /api/v1/products/42
→ 200 OK
  ETag: "version-5"
  { "id": 42, "name": "Widget", "price": 9.99, "stock": 50 }

// Update with If-Match (prevent lost update)
PUT /api/v1/products/42
If-Match: "version-5"
{ "name": "Widget Pro", "price": 14.99, "stock": 50 }

→ 200 OK (if ETag still matches — no one else modified it)
→ 412 Precondition Failed (if someone else modified between read and write)
   { "error": { "code": "CONFLICT", "message": "Resource was modified. Fetch latest and retry." } }
```

---

## 6. Idempotency

### Why It Matters

```
Problem: Network is unreliable. Client sends POST, server processes it...
but the response is lost. Client retries → duplicate resource created!

POST /orders  (Idempotency-Key: "key-123")
→ Server crashes after creating order
→ Client retries with same Idempotency-Key
→ Server sees key-123 already processed, returns original response
→ No duplicate order! ✅
```

### Implementation

```python
# Server-side idempotency middleware

def idempotency_middleware(request):
    if request.method not in ('POST', 'PATCH'):
        return process_normally(request)
    
    key = request.headers.get('Idempotency-Key')
    if not key:
        return error(400, "Idempotency-Key header required for POST/PATCH")
    
    # Check if we've seen this key before
    cached = redis.get(f"idempotency:{key}")
    if cached:
        return cached_response(cached)  # Return stored response
    
    # Process request
    response = process_request(request)
    
    # Store response for future replays (TTL: 24 hours)
    redis.setex(f"idempotency:{key}", 86400, serialize(response))
    
    return response
```

### Method Idempotency Reference

| Method | Naturally Idempotent? | How to Ensure |
|--------|----------------------|---------------|
| GET | ✅ Yes | Read-only, no side effects |
| PUT | ✅ Yes | Replaces entire resource — same result each time |
| DELETE | ✅ Yes | Second delete returns 404 (same end state) |
| POST | ❌ No | Use `Idempotency-Key` header |
| PATCH | ❌ Depends | Use `Idempotency-Key` or design as idempotent |

---

## 7. HATEOAS — Hypermedia as the Engine of Application State

### Concept

The API tells the client what actions are available via links in the response.

```json
// GET /api/v1/orders/456
{
  "data": {
    "id": "ord_456",
    "status": "confirmed",
    "total": 99.99,
    "items": [ ... ]
  },
  "links": {
    "self":    { "href": "/api/v1/orders/456", "method": "GET" },
    "cancel":  { "href": "/api/v1/orders/456/cancel", "method": "POST" },
    "pay":     { "href": "/api/v1/orders/456/pay", "method": "POST" },
    "items":   { "href": "/api/v1/orders/456/items", "method": "GET" }
  }
}

// After payment, links change:
{
  "data": {
    "id": "ord_456",
    "status": "paid"
  },
  "links": {
    "self":     { "href": "/api/v1/orders/456", "method": "GET" },
    "refund":   { "href": "/api/v1/orders/456/refund", "method": "POST" },
    "track":    { "href": "/api/v1/orders/456/tracking", "method": "GET" }
    // "cancel" and "pay" are GONE — no longer valid actions
  }
}
```

### Interview Tip: HATEOAS Reality

- **Theory**: Beautiful, self-documenting APIs
- **Practice**: Most APIs don't implement full HATEOAS
- **Practical middle ground**: Include pagination links and key action URLs
- **Who uses it well**: GitHub API (partial), PayPal API

---

## 8. REST Maturity Model (Richardson Maturity)

```
Level 3: HATEOAS (Hypermedia Controls)
  ↑  API responses include links to possible actions
  │
Level 2: HTTP Methods + Status Codes
  ↑  Proper use of GET/POST/PUT/DELETE + 200/201/404/409/etc.
  │
Level 1: Resources
  ↑  Resource-oriented URLs (/users/123/orders)
  │
Level 0: The Swamp of POX
     Single URL, POST everything, custom status in body
     POST /api  { "action": "getUser", "userId": 123 }
```

Most production APIs are at **Level 2**. Level 3 is aspirational but rarely fully implemented.

---

## 9. PUT vs PATCH vs POST — When to Use Which

| Scenario | Method | Why |
|----------|--------|-----|
| Create new user | POST /users | Don't know ID yet |
| Create with known ID | PUT /users/123 | Client provides ID (upsert) |
| Update entire profile | PUT /users/123 | Replace all fields |
| Update just the phone | PATCH /users/123 | Change one field |
| Change password | POST /users/123/change-password | Action, not resource update |
| Upload avatar | PUT /users/123/avatar | Replace entire avatar |
| Add item to cart | POST /carts/123/items | Create sub-resource |
| Change item quantity | PATCH /carts/123/items/456 | Partial update |
| Remove item from cart | DELETE /carts/123/items/456 | Remove sub-resource |

### Interview Discussion: PUT vs PATCH

```
PUT:   "Here is the complete new version of this resource"
PATCH: "Here are the changes to apply to this resource"

PUT   → Must send full object → missing fields become null/default
PATCH → Send only changed fields → other fields preserved

PUT   → Always idempotent (same full object = same result)
PATCH → Not always idempotent ("increment counter by 1" is not idempotent)
```

---

## 10. REST API Design Anti-Patterns

### Anti-Pattern 1: Verbs in URLs

```
❌ GET  /getUsers
❌ POST /createUser
❌ POST /deleteUser/123

✅ GET    /users
✅ POST   /users
✅ DELETE  /users/123
```

### Anti-Pattern 2: Ignoring Status Codes

```
❌ 200 OK { "success": false, "error": "User not found" }
✅ 404 Not Found { "error": { "code": "NOT_FOUND", "message": "User not found" } }
```

### Anti-Pattern 3: Nested Too Deep

```
❌ GET /companies/1/departments/2/teams/3/members/4/projects/5/tasks
✅ GET /tasks/5  (direct access with ID)
✅ GET /teams/3/members  (max 2 levels)
```

### Anti-Pattern 4: Exposing Internal Details

```
❌ GET /users/123 → { "password_hash": "...", "db_id": 42, "_internal_flag": true }
✅ GET /users/123 → { "id": "user_123", "name": "Alice", "email": "..." }
```

### Anti-Pattern 5: Breaking Changes Without Versioning

```
❌ Renaming "username" to "handle" in response (breaks all clients)
✅ Add "handle" alongside "username", deprecate "username" with notice
✅ Or: /v2/ introduces new field names
```
