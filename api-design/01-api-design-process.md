# API Design Process — Interview Guide

## How to Approach Any API Design Question in an Interview

---

## 1. The 5-Step API Design Framework

When an interviewer asks you to "design the API for X", follow this systematic approach:

```
Step 1: Clarify Requirements & Scope
         │
         ▼
Step 2: Identify Resources (Nouns)
         │
         ▼
Step 3: Define Endpoints (Verbs + URLs)
         │
         ▼
Step 4: Design Request/Response Contracts
         │
         ▼
Step 5: Address Cross-Cutting Concerns
        (Auth, Pagination, Errors, Rate Limiting)
```

---

## 2. Step 1 — Clarify Requirements

### What to Ask the Interviewer

| Category | Questions to Ask |
|----------|------------------|
| **Users** | Who are the API consumers? (mobile, web, third-party, internal services) |
| **Scale** | How many requests/sec? How many resources? |
| **Access patterns** | Read-heavy or write-heavy? Real-time needs? |
| **Auth** | Public API or internal? Who can access what? |
| **Scope** | Which features are in scope for this design? |
| **Constraints** | Latency requirements? Backward compatibility needs? |

### Example: "Design the API for Twitter"

**Clarifying Questions:**
- Are we designing the public API or internal microservice APIs?
- Which features? Tweets, timelines, follows, DMs, search?
- Do we need real-time updates (streaming)?
- What clients? Mobile apps, web app, third-party developers?
- Rate limiting requirements?

**Scoped to:** Public REST API for tweets, timelines, and follows.

---

## 3. Step 2 — Identify Resources

### How to Find Resources

Resources are **nouns** — the "things" your API manages. Think about:

1. **Core resources** — the primary entities
2. **Sub-resources** — entities owned by a parent
3. **Action resources** — when CRUD doesn't fit (use sparingly)

### Rules for Resource Naming

| Rule | ✅ Good | ❌ Bad |
|------|---------|--------|
| Use **nouns**, not verbs | `/users` | `/getUsers` |
| Use **plural** nouns | `/orders` | `/order` |
| Use **lowercase** with hyphens | `/order-items` | `/orderItems`, `/OrderItems` |
| Use **hierarchical** nesting | `/users/123/orders` | `/getUserOrders?userId=123` |
| Max **2 levels** of nesting | `/users/123/orders` | `/users/123/orders/456/items/789/reviews` |
| Use IDs, not names | `/users/123` | `/users/alice` (unless username is the ID) |

### Example: Twitter API Resources

```
Core Resources:
├── /users                    → User accounts
├── /tweets                   → Tweet content
├── /timelines                → User timelines
└── /direct-messages          → DMs

Sub-Resources:
├── /users/{id}/followers     → Users who follow this user
├── /users/{id}/following     → Users this user follows
├── /users/{id}/tweets        → Tweets by this user
├── /users/{id}/likes         → Tweets this user liked
├── /tweets/{id}/replies      → Replies to a tweet
├── /tweets/{id}/retweets     → Retweets of a tweet
└── /tweets/{id}/likes        → Users who liked this tweet

Action Resources (when CRUD doesn't fit):
├── /users/{id}/follow        → POST to follow a user
├── /users/{id}/unfollow      → POST to unfollow
├── /tweets/{id}/retweet      → POST to retweet
└── /search/tweets            → Search tweets
```

### Example: E-Commerce API Resources

```
Core Resources:
├── /users
├── /products
├── /orders
├── /categories
└── /carts

Sub-Resources:
├── /users/{id}/addresses
├── /users/{id}/orders
├── /users/{id}/wishlist
├── /products/{id}/reviews
├── /products/{id}/images
├── /orders/{id}/items
├── /orders/{id}/payments
├── /carts/{id}/items
└── /categories/{id}/products

Action Resources:
├── /orders/{id}/cancel       → POST to cancel
├── /orders/{id}/refund       → POST to refund
├── /carts/{id}/checkout      → POST to checkout
└── /search/products          → Search products
```

### Example: Chat Application API Resources

```
Core Resources:
├── /users
├── /conversations
└── /messages

Sub-Resources:
├── /conversations/{id}/members
├── /conversations/{id}/messages
├── /messages/{id}/reactions
└── /messages/{id}/attachments

Action Resources:
├── /conversations/{id}/join    → POST to join
├── /conversations/{id}/leave   → POST to leave
├── /messages/{id}/read         → POST to mark as read
└── /users/{id}/typing          → POST typing indicator
```

---

## 4. Step 3 — Define Endpoints

### HTTP Method Selection

| Action | HTTP Method | URL Pattern | Idempotent? | Example |
|--------|-------------|-------------|-------------|---------|
| List all | `GET` | `/resources` | Yes | `GET /users` |
| Get one | `GET` | `/resources/{id}` | Yes | `GET /users/123` |
| Create | `POST` | `/resources` | No* | `POST /users` |
| Full update | `PUT` | `/resources/{id}` | Yes | `PUT /users/123` |
| Partial update | `PATCH` | `/resources/{id}` | No* | `PATCH /users/123` |
| Delete | `DELETE` | `/resources/{id}` | Yes | `DELETE /users/123` |
| Action | `POST` | `/resources/{id}/action` | Depends | `POST /orders/123/cancel` |
| Search | `GET` | `/resources?query=...` | Yes | `GET /products?q=phone` |

*Can be made idempotent with idempotency keys.

### Full Twitter API Endpoint Design

```
USERS
──────────────────────────────────────────────────────────
GET     /users/{id}                 → Get user profile
PATCH   /users/{id}                 → Update user profile
DELETE  /users/{id}                 → Delete user account
GET     /users/{id}/followers       → List followers (paginated)
GET     /users/{id}/following       → List following (paginated)
POST    /users/{id}/follow          → Follow user
DELETE  /users/{id}/follow          → Unfollow user
GET     /users/{id}/tweets          → List user's tweets (paginated)
GET     /users/{id}/likes           → List user's liked tweets

TWEETS
──────────────────────────────────────────────────────────
POST    /tweets                     → Create tweet
GET     /tweets/{id}                → Get tweet
DELETE  /tweets/{id}                → Delete tweet
GET     /tweets/{id}/replies        → List replies (paginated)
POST    /tweets/{id}/retweet        → Retweet
DELETE  /tweets/{id}/retweet        → Undo retweet
POST    /tweets/{id}/like           → Like tweet
DELETE  /tweets/{id}/like           → Unlike tweet
POST    /tweets/{id}/bookmark       → Bookmark tweet

TIMELINE
──────────────────────────────────────────────────────────
GET     /timeline/home              → Home timeline (paginated, cursor)
GET     /timeline/mentions          → Mentions timeline

SEARCH
──────────────────────────────────────────────────────────
GET     /search/tweets?q=...        → Search tweets
GET     /search/users?q=...         → Search users

AUTH
──────────────────────────────────────────────────────────
POST    /auth/register              → Register new user
POST    /auth/login                 → Login (get tokens)
POST    /auth/refresh               → Refresh access token
POST    /auth/logout                → Logout (invalidate token)
```

---

## 5. Step 4 — Design Request/Response Contracts

### Consistent Response Envelope

```json
// Success response
{
  "data": { ... },              // The actual resource or array
  "meta": {                     // Metadata
    "request_id": "req_abc123",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}

// Success response (list with pagination)
{
  "data": [ ... ],
  "meta": {
    "total_count": 1250,
    "page_size": 20,
    "next_cursor": "eyJpZCI6MTIzfQ=="
  }
}

// Error response
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": [
      { "field": "email", "message": "Invalid email format" },
      { "field": "age", "message": "Must be at least 13" }
    ]
  },
  "meta": {
    "request_id": "req_abc123",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

### Concrete Example: Tweet Endpoints

#### POST /tweets — Create Tweet

```
Request:
  Method: POST
  URL: /api/v1/tweets
  Headers:
    Authorization: Bearer <token>
    Content-Type: application/json
    Idempotency-Key: <uuid>
  Body:
    {
      "text": "Hello, world! #firsttweet",
      "reply_to_id": null,
      "media_ids": ["media_abc123"],
      "poll": {
        "options": ["Yes", "No"],
        "duration_minutes": 1440
      }
    }

Response (201 Created):
  Headers:
    Location: /api/v1/tweets/tweet_789
    X-RateLimit-Remaining: 298
  Body:
    {
      "data": {
        "id": "tweet_789",
        "text": "Hello, world! #firsttweet",
        "author": {
          "id": "user_123",
          "username": "alice",
          "display_name": "Alice"
        },
        "created_at": "2024-01-15T10:30:00Z",
        "metrics": {
          "likes": 0,
          "retweets": 0,
          "replies": 0
        },
        "media": [
          { "id": "media_abc123", "type": "image", "url": "https://..." }
        ]
      }
    }
```

#### GET /timeline/home — Home Timeline

```
Request:
  Method: GET
  URL: /api/v1/timeline/home?limit=20&cursor=eyJpZCI6MTIzfQ==
  Headers:
    Authorization: Bearer <token>

Response (200 OK):
  Body:
    {
      "data": [
        {
          "id": "tweet_800",
          "text": "Great weather today!",
          "author": {
            "id": "user_456",
            "username": "bob",
            "display_name": "Bob"
          },
          "created_at": "2024-01-15T11:00:00Z",
          "metrics": { "likes": 42, "retweets": 5, "replies": 3 },
          "is_liked_by_me": true,
          "is_retweeted_by_me": false
        }
        // ... more tweets
      ],
      "meta": {
        "next_cursor": "eyJpZCI6Nzk5fQ==",
        "has_more": true
      }
    }
```

#### GET /users/{id} — Get User Profile

```
Request:
  Method: GET
  URL: /api/v1/users/user_456
  Headers:
    Authorization: Bearer <token>

Response (200 OK):
  Body:
    {
      "data": {
        "id": "user_456",
        "username": "bob",
        "display_name": "Bob Smith",
        "bio": "Software engineer at Acme Corp",
        "avatar_url": "https://cdn.example.com/avatars/bob.jpg",
        "location": "San Francisco, CA",
        "website": "https://bob.dev",
        "created_at": "2020-03-15T00:00:00Z",
        "metrics": {
          "followers_count": 1250,
          "following_count": 380,
          "tweet_count": 5420
        },
        "is_followed_by_me": true,
        "is_following_me": false,
        "is_blocked": false
      }
    }
```

---

## 6. Step 5 — Cross-Cutting Concerns

Address these in every API design interview:

| Concern | What to Discuss |
|---------|-----------------|
| **Authentication** | OAuth 2.0 / JWT for users; API keys for services |
| **Authorization** | RBAC — who can access what endpoints? |
| **Pagination** | Cursor-based for feeds/timelines; offset-based for admin pages |
| **Rate Limiting** | Per-user, per-endpoint; return `429` with `Retry-After` |
| **Caching** | ETags for individual resources; `Cache-Control` for lists |
| **Versioning** | URI-based (`/v1/`) for simplicity; discuss migration |
| **Idempotency** | `Idempotency-Key` header for POST/PATCH |
| **Error Handling** | Consistent error format, meaningful messages |
| **Compression** | Accept-Encoding: gzip for large payloads |
| **Timeouts** | Client and server-side timeout policies |
| **Documentation** | OpenAPI/Swagger spec |
| **Monitoring** | Request/response logging, latency tracking, error rates |

---

## 7. Common API Design Patterns

### Pattern 1: CRUD API (Resource-Oriented)

```
GET    /resources         → List
GET    /resources/{id}    → Read
POST   /resources         → Create
PUT    /resources/{id}    → Replace
PATCH  /resources/{id}    → Update
DELETE /resources/{id}    → Delete
```

### Pattern 2: Sub-Resource API

```
GET    /users/{id}/orders         → List user's orders
POST   /users/{id}/orders         → Create order for user
GET    /users/{id}/orders/{oid}   → Get specific order
```

### Pattern 3: Action API (RPC-style)

When CRUD doesn't fit — use POST with an action verb:

```
POST   /orders/{id}/cancel        → Cancel an order
POST   /orders/{id}/refund        → Issue refund
POST   /payments/{id}/capture     → Capture payment
POST   /users/{id}/verify-email   → Send verification email
POST   /reports/generate          → Trigger report generation
```

### Pattern 4: Search/Filter API

```
GET    /products?q=phone&category=electronics&min_price=100&max_price=500&sort=price:asc&limit=20
```

### Pattern 5: Batch/Bulk API

```json
POST /batch
{
  "operations": [
    { "method": "POST", "url": "/users", "body": { "name": "Alice" } },
    { "method": "POST", "url": "/users", "body": { "name": "Bob" } },
    { "method": "DELETE", "url": "/users/old_user_123" }
  ]
}

// Response
{
  "results": [
    { "status": 201, "body": { "id": "user_1" } },
    { "status": 201, "body": { "id": "user_2" } },
    { "status": 204, "body": null }
  ]
}
```

### Pattern 6: Long-Running Operations (Async)

```
POST /reports/generate
→ 202 Accepted
  {
    "data": {
      "operation_id": "op_abc123",
      "status": "processing",
      "status_url": "/operations/op_abc123",
      "estimated_completion": "2024-01-15T10:35:00Z"
    }
  }

GET /operations/op_abc123
→ 200 OK
  {
    "data": {
      "operation_id": "op_abc123",
      "status": "completed",
      "result_url": "/reports/rpt_xyz789",
      "completed_at": "2024-01-15T10:33:00Z"
    }
  }
```

### Pattern 7: Webhook API

```json
// Register webhook
POST /webhooks
{
  "url": "https://myapp.com/webhook",
  "events": ["order.created", "order.shipped", "payment.completed"],
  "secret": "whsec_..."
}

// Webhook payload sent to your URL
POST https://myapp.com/webhook
Headers:
  X-Webhook-Signature: sha256=...
  X-Webhook-Event: order.created
  X-Webhook-Delivery: del_abc123
Body:
{
  "id": "evt_123",
  "type": "order.created",
  "created_at": "2024-01-15T10:30:00Z",
  "data": {
    "order_id": "ord_456",
    "user_id": "user_789",
    "total": 99.99
  }
}
```

---

## 8. Resource Modeling for Common Scenarios

### Scenario 1: Ride-Sharing (Uber)

```
/riders/{id}                    → Rider profile
/drivers/{id}                   → Driver profile
/drivers/{id}/location          → Driver's current location (GET/PUT)
/rides                          → POST to request ride
/rides/{id}                     → GET ride details
/rides/{id}/cancel              → POST to cancel
/rides/{id}/rate                → POST to rate
/rides/{id}/route               → GET ride route
/rides/estimate                 → POST to get fare estimate
/payments/{id}                  → Payment info
```

### Scenario 2: Food Delivery (DoorDash)

```
/restaurants                    → List/search restaurants
/restaurants/{id}               → Restaurant details
/restaurants/{id}/menu          → Menu items
/orders                         → POST to place order
/orders/{id}                    → GET order details
/orders/{id}/track              → GET live tracking
/orders/{id}/cancel             → POST to cancel
/deliveries/{id}                → Delivery status
/deliveries/{id}/location       → Live driver location
```

### Scenario 3: Payment Gateway (Stripe)

```
/customers                      → Customer management
/payment-methods                → Cards, bank accounts
/payment-intents                → Initiate payment
/payment-intents/{id}/confirm   → Confirm payment
/payment-intents/{id}/capture   → Capture authorized payment
/payment-intents/{id}/cancel    → Cancel payment
/refunds                        → Issue refund
/charges                        → List charges
/webhooks                       → Webhook subscriptions
/events                         → Event log
```

---

## 9. Interview Tips

### Do's
- ✅ Start with requirements clarification
- ✅ Identify resources (nouns) before endpoints
- ✅ Use standard HTTP methods and status codes
- ✅ Design request/response contracts with examples
- ✅ Address pagination, auth, error handling upfront
- ✅ Discuss trade-offs (REST vs GraphQL, cursor vs offset)
- ✅ Mention real-world APIs for reference (Stripe, Twitter, GitHub)

### Don'ts
- ❌ Use verbs in URLs (`/getUsers`, `/createOrder`)
- ❌ Design endpoints without clarifying scope
- ❌ Forget about pagination for list endpoints
- ❌ Skip error handling
- ❌ Ignore security/authentication
- ❌ Over-nest URLs beyond 2 levels

### Time Management (45-min API design interview)

| Phase | Time | Activity |
|-------|------|----------|
| Clarify | 5 min | Scope, users, features, constraints |
| Resources | 5 min | List resources, define relationships |
| Endpoints | 15 min | Design URLs, methods, key contracts |
| Contracts | 10 min | Request/response JSON for 2-3 key endpoints |
| Cross-cutting | 10 min | Auth, pagination, errors, rate limiting, versioning |

---

## 10. API Design Principles (Quick Reference)

| Principle | Description |
|-----------|-------------|
| **Consistency** | Same patterns across all endpoints (naming, errors, pagination) |
| **Predictability** | If you know one endpoint, you can guess the others |
| **Simplicity** | Easy to understand without reading documentation |
| **Backward compatibility** | Don't break existing clients when evolving |
| **Least surprise** | API behaves as developers expect |
| **Documentation-driven** | Write the API spec before implementation |
| **Consumer-first** | Design for the API consumer, not the backend |
