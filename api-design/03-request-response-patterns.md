# Request & Response Patterns

## Pagination, Filtering, Sorting, Partial Responses, Bulk Ops, and Async APIs

---

## 1. Pagination

### Why Pagination Matters

Without pagination, a `GET /users` on a table with 10M rows would:
- Time out
- Consume massive memory
- Crash the client
- Kill your database

### Strategy Comparison

| Strategy | Pros | Cons | Best For |
|----------|------|------|----------|
| **Offset-based** | Simple, supports random page access | Slow for deep pages O(offset), inconsistent with concurrent inserts | Admin panels, small datasets |
| **Cursor-based** | Consistent, fast O(1), handles real-time changes | No random page access, more complex | Feeds, timelines, large datasets |
| **Keyset** | DB-efficient, works with indexes | Must sort by unique column | Large datasets, database-optimal |
| **Page token** | Opaque, flexible, hides implementation | Server must store state or encode | Public APIs (Google, Facebook style) |

### Offset-Based Pagination

```
Request:
  GET /api/v1/products?page=3&per_page=20

Response:
  {
    "data": [ ... 20 items ... ],
    "pagination": {
      "page": 3,
      "per_page": 20,
      "total_items": 1250,
      "total_pages": 63
    }
  }

SQL: SELECT * FROM products ORDER BY id LIMIT 20 OFFSET 40;
  ⚠️ OFFSET 100000 scans 100,020 rows — very slow for deep pages!
```

**Problems with offset pagination:**
- Page 5000? DB must skip 100,000 rows → O(offset) cost
- If items are inserted/deleted between page fetches, items can be skipped or duplicated

### Cursor-Based Pagination

```
Request:
  GET /api/v1/timeline?limit=20                      (first page)
  GET /api/v1/timeline?limit=20&cursor=eyJpZCI6MTAwfQ==  (next page)

Response:
  {
    "data": [ ... 20 items ... ],
    "pagination": {
      "next_cursor": "eyJpZCI6ODB9",     // base64 encoded {"id": 80}
      "has_more": true
    }
  }

SQL: SELECT * FROM tweets 
     WHERE id < 100                       -- cursor = last seen id
     ORDER BY id DESC 
     LIMIT 20;
  ✅ Uses index, O(1) regardless of page depth
```

**Cursor encoding:**
```python
# Encode cursor
import base64, json
cursor_data = {"id": 100, "created_at": "2024-01-15T10:30:00Z"}
cursor = base64.b64encode(json.dumps(cursor_data).encode()).decode()
# "eyJpZCI6IDEwMCwgImNyZWF0ZWRfYXQiOiAiMjAyNC0wMS0xNVQxMDozMDowMFoifQ=="

# Decode cursor
cursor_data = json.loads(base64.b64decode(cursor))
```

### Keyset Pagination

```
Request:
  GET /api/v1/products?limit=20&after_id=100

SQL: SELECT * FROM products WHERE id > 100 ORDER BY id LIMIT 20;
  ✅ Uses index, very efficient
  ❌ Only works with unique, sortable column

Request (multi-column sort):
  GET /api/v1/products?limit=20&after_price=29.99&after_id=500

SQL: SELECT * FROM products
     WHERE (price, id) > (29.99, 500)
     ORDER BY price, id
     LIMIT 20;
```

### Link Header Pagination (GitHub style)

```
Response Headers:
  Link: <https://api.example.com/users?page=2>; rel="next",
        <https://api.example.com/users?page=50>; rel="last",
        <https://api.example.com/users?page=1>; rel="first",
        <https://api.example.com/users?page=1>; rel="prev"
```

### Interview Recommendation

```
"For user-facing feeds and timelines → cursor-based pagination
 For admin dashboards with page numbers → offset-based (acceptable for small scale)
 For database-optimal large datasets → keyset pagination"
```

---

## 2. Filtering

### Query Parameter Patterns

```
// Simple equality
GET /products?status=active&category=electronics

// Comparison operators
GET /products?min_price=100&max_price=500
GET /products?price[gte]=100&price[lte]=500       // bracket notation
GET /products?price=gte:100,lte:500               // colon notation

// Multiple values (OR)
GET /products?status=active,pending                 // comma-separated
GET /products?status[]=active&status[]=pending      // repeated param

// Date ranges
GET /orders?created_after=2024-01-01&created_before=2024-02-01
GET /orders?created_at[gte]=2024-01-01T00:00:00Z

// Text search
GET /products?q=wireless+headphones
GET /products?search=wireless+headphones

// Nested field filter
GET /orders?user.city=San+Francisco
GET /products?metadata.color=red
```

### Filter Design Best Practices

| Pattern | ✅ Good | ❌ Bad |
|---------|---------|--------|
| User-friendly names | `?status=active` | `?s=1` |
| Standard operators | `?min_price=100` | `?price_gt=100` (inconsistent) |
| Multiple values | `?status=active,pending` | Separate endpoints per status |
| Date ranges | ISO 8601 format | Custom date formats |
| Default filtering | Active items only | Return everything including deleted |

---

## 3. Sorting

```
// Single field sort
GET /products?sort=price              // ascending (default)
GET /products?sort=-price             // descending (prefix with -)
GET /products?sort=price:asc          // explicit direction

// Multi-field sort
GET /products?sort=-created_at,name   // newest first, then alphabetical
GET /products?sort=price:asc,name:asc

// Sort + Filter + Pagination (combined)
GET /products?category=electronics&sort=-rating&limit=20&cursor=abc123
```

### Sorting Best Practice

```
Default sort: Always have a default sort order (e.g., created_at DESC)
Allowed fields: Whitelist sortable fields (don't expose all DB columns)
Index alignment: Ensure sort fields are indexed in the database

// Document allowed sort fields in API docs:
// Sortable fields: name, price, created_at, rating, popularity
```

---

## 4. Partial Responses (Field Selection)

Reduce payload size by requesting only needed fields.

### Sparse Fieldsets

```
// Google style: fields parameter
GET /users/123?fields=name,email,avatar_url

Response:
{
  "data": {
    "name": "Alice",
    "email": "alice@example.com",
    "avatar_url": "https://..."
    // Other fields omitted
  }
}

// Nested field selection
GET /orders/456?fields=id,status,items(name,quantity)

// Twitter API v2 style: typed field selection
GET /tweets?ids=123,456&tweet.fields=text,created_at&user.fields=name,username

// JSON:API style
GET /articles?fields[articles]=title,body&fields[authors]=name
```

### Expanding Relationships

```
// Include related resources (avoid N+1)
GET /orders/456?expand=user,items.product

Response:
{
  "data": {
    "id": "ord_456",
    "user": {                          // expanded, not just user_id
      "id": "user_123",
      "name": "Alice"
    },
    "items": [
      {
        "quantity": 2,
        "product": {                   // nested expansion
          "id": "prod_789",
          "name": "Widget"
        }
      }
    ]
  }
}

// Without expand:
GET /orders/456
{
  "data": {
    "id": "ord_456",
    "user_id": "user_123",             // just the ID
    "items": [
      { "quantity": 2, "product_id": "prod_789" }
    ]
  }
}
```

---

## 5. Bulk/Batch Operations

### Pattern 1: Batch create

```json
// POST /api/v1/users/batch
{
  "users": [
    { "name": "Alice", "email": "alice@example.com" },
    { "name": "Bob", "email": "bob@example.com" },
    { "name": "Carol", "email": "carol@example.com" }
  ]
}

// Response: 207 Multi-Status
{
  "results": [
    { "index": 0, "status": 201, "data": { "id": "user_1" } },
    { "index": 1, "status": 201, "data": { "id": "user_2" } },
    { "index": 2, "status": 409, "error": { "code": "DUPLICATE_EMAIL", "message": "Email already exists" } }
  ],
  "summary": { "total": 3, "succeeded": 2, "failed": 1 }
}
```

### Pattern 2: Multi-operation batch (like Google API)

```json
// POST /api/v1/batch
{
  "operations": [
    { "id": "op1", "method": "POST", "url": "/users", "body": { "name": "Alice" } },
    { "id": "op2", "method": "PATCH", "url": "/users/123", "body": { "status": "active" } },
    { "id": "op3", "method": "DELETE", "url": "/users/456" }
  ]
}

// Response
{
  "results": [
    { "id": "op1", "status": 201, "body": { "id": "user_789" } },
    { "id": "op2", "status": 200, "body": { "id": "123", "status": "active" } },
    { "id": "op3", "status": 204, "body": null }
  ]
}
```

### Pattern 3: Bulk update

```json
// PATCH /api/v1/products/batch
{
  "updates": [
    { "id": "prod_1", "price": 29.99 },
    { "id": "prod_2", "price": 49.99 },
    { "id": "prod_3", "stock": 0, "status": "out_of_stock" }
  ]
}
```

### Pattern 4: Bulk delete

```
// DELETE /api/v1/users/batch
{
  "ids": ["user_1", "user_2", "user_3"]
}

// Or with query param
DELETE /api/v1/users?ids=user_1,user_2,user_3
```

### Batch Design Considerations

| Consideration | Recommendation |
|---------------|----------------|
| Max batch size | 100-1000 items per request |
| Atomicity | Usually non-atomic (partial success OK) |
| Status per item | Return individual status for each item |
| Error handling | Continue processing even if one item fails |
| Idempotency | Support idempotency key per batch operation |

---

## 6. Async APIs / Long-Running Operations

### Pattern: Polling

```
Step 1: Submit long-running job
  POST /api/v1/reports/generate
  Body: { "type": "annual_revenue", "year": 2024 }
  
  → 202 Accepted
  {
    "data": {
      "job_id": "job_abc123",
      "status": "queued",
      "status_url": "/api/v1/jobs/job_abc123",
      "estimated_duration_seconds": 120
    }
  }

Step 2: Poll for completion
  GET /api/v1/jobs/job_abc123
  
  → 200 OK
  { "status": "processing", "progress": 45 }      // still working

  → 200 OK
  {
    "status": "completed",
    "result": {
      "download_url": "/api/v1/reports/rpt_xyz789/download",
      "expires_at": "2024-01-16T10:30:00Z"
    }
  }

  → 200 OK
  { "status": "failed", "error": { "code": "TIMEOUT", "message": "..." } }
```

### Pattern: Webhooks + Polling

```
Step 1: Submit job with callback URL
  POST /api/v1/reports/generate
  Body: { 
    "type": "annual_revenue",
    "callback_url": "https://myapp.com/webhooks/report-ready"
  }
  → 202 Accepted { "job_id": "job_abc123" }

Step 2: Server calls callback when done
  POST https://myapp.com/webhooks/report-ready
  Body: {
    "event": "report.completed",
    "job_id": "job_abc123",
    "download_url": "..."
  }
```

### Pattern: Server-Sent Events (SSE)

```
GET /api/v1/jobs/job_abc123/stream
Accept: text/event-stream

→ Response (streaming):
  data: {"status": "processing", "progress": 10}

  data: {"status": "processing", "progress": 45}

  data: {"status": "processing", "progress": 88}

  data: {"status": "completed", "result_url": "/reports/rpt_xyz789"}
```

---

## 7. File Upload Patterns

### Pattern 1: Multipart Upload

```
POST /api/v1/users/123/avatar
Content-Type: multipart/form-data; boundary=----FormBoundary

------FormBoundary
Content-Disposition: form-data; name="file"; filename="avatar.jpg"
Content-Type: image/jpeg

<binary data>
------FormBoundary--

→ 200 OK
{
  "data": {
    "url": "https://cdn.example.com/avatars/123/avatar.jpg",
    "size": 245760,
    "content_type": "image/jpeg"
  }
}
```

### Pattern 2: Pre-Signed URL (Recommended for Large Files)

```
Step 1: Request upload URL
  POST /api/v1/uploads
  Body: { "filename": "document.pdf", "content_type": "application/pdf", "size": 5242880 }
  
  → 200 OK
  {
    "upload_url": "https://s3.amazonaws.com/bucket/..?X-Amz-Signature=...",
    "upload_id": "upload_abc123",
    "expires_at": "2024-01-15T11:00:00Z"
  }

Step 2: Client uploads directly to cloud storage
  PUT https://s3.amazonaws.com/bucket/...
  Body: <binary file data>
  → 200 OK

Step 3: Confirm upload
  POST /api/v1/uploads/upload_abc123/confirm
  → 200 OK { "file_url": "https://cdn.example.com/files/document.pdf" }
```

### Pattern 3: Chunked Upload (for very large files)

```
Step 1: Initiate
  POST /api/v1/uploads/initiate
  Body: { "filename": "video.mp4", "total_size": 1073741824, "chunk_size": 10485760 }
  → 200 OK { "upload_id": "up_123", "total_chunks": 103 }

Step 2: Upload chunks
  PUT /api/v1/uploads/up_123/chunks/1
  Content-Range: bytes 0-10485759/1073741824
  Body: <chunk data>
  → 200 OK { "chunk": 1, "received_bytes": 10485760 }

Step 3: Complete
  POST /api/v1/uploads/up_123/complete
  → 200 OK { "file_url": "..." }
```

---

## 8. Response Envelope Patterns

### Pattern 1: Minimal (Stripe-style)

```json
// Success — resource directly at top level
{ "id": "cus_123", "name": "Alice", "email": "alice@example.com" }

// List — array with metadata
{
  "data": [ ... ],
  "has_more": true,
  "url": "/v1/customers"
}

// Error
{
  "error": {
    "type": "invalid_request_error",
    "message": "Invalid email address",
    "param": "email"
  }
}
```

### Pattern 2: Wrapped (Consistent envelope)

```json
// Always wrapped in data/error
{
  "data": { "id": "123", "name": "Alice" },
  "meta": { "request_id": "req_abc", "timestamp": "..." }
}

{
  "error": { "code": "NOT_FOUND", "message": "..." },
  "meta": { "request_id": "req_abc" }
}
```

### Pattern 3: JSON:API

```json
{
  "data": {
    "type": "users",
    "id": "123",
    "attributes": { "name": "Alice", "email": "alice@example.com" },
    "relationships": {
      "orders": { "data": [{ "type": "orders", "id": "456" }] }
    }
  },
  "included": [
    { "type": "orders", "id": "456", "attributes": { "total": 99.99 } }
  ]
}
```

### Which Pattern to Choose?

| Pattern | When to Use |
|---------|-------------|
| **Minimal** | Public APIs, developer-friendly (Stripe, Twilio) |
| **Wrapped** | Internal APIs, when you need metadata on every response |
| **JSON:API** | When you need standardized relationships, sparse fieldsets |

---

## 9. Real-Time Patterns

| Pattern | Protocol | Direction | Use Case |
|---------|----------|-----------|----------|
| **Polling** | HTTP | Client → Server | Simple status checks |
| **Long Polling** | HTTP | Server holds | Chat, notifications |
| **SSE** | HTTP | Server → Client | Live feeds, dashboards |
| **WebSocket** | WS | Bidirectional | Chat, gaming, collaboration |
| **Webhooks** | HTTP | Server → Server | Event notifications |

### Comparison

```
Polling:       Client asks every N seconds → wasteful if no updates
Long Polling:  Server holds request until data available → better but complex
SSE:           Server pushes events over HTTP → simple, one-directional
WebSocket:     Full duplex, persistent connection → most flexible, most complex
Webhooks:      Server POSTs to your URL on event → async, decoupled
```

---

## 10. API Response Best Practices

| Practice | Description |
|----------|-------------|
| Use **camelCase** or **snake_case** consistently | Pick one and stick with it across all endpoints |
| Include **timestamps** in ISO 8601 | `"2024-01-15T10:30:00Z"` with timezone |
| Use **string IDs** in responses | `"id": "user_123"` not `"id": 123` (avoid int overflow in JS) |
| Include **self link** | `"url": "/api/v1/users/123"` for easy resource lookup |
| Return **created resource** on POST | Don't just return `201` — return the full resource |
| Use **null** for missing optional fields | Don't omit them entirely (makes parsing harder) |
| Include **request_id** | For debugging and support tickets |
| Don't expose **internal IDs** | Use UUIDs or prefixed IDs (`user_abc123`) |
