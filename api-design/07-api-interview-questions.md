# API Interview Questions — 40+ Problems with Full API Designs

## From Simple CRUD to Complex Real-World Systems

---

## How to Use This Guide

For each question, I provide:
1. **Scope** — What to clarify with the interviewer
2. **Resources** — Identified nouns
3. **Endpoints** — Complete endpoint design
4. **Key Contracts** — Request/response examples for critical endpoints
5. **Discussion Points** — What the interviewer is looking for

---

# EASY (1-10) — Basic CRUD API Design

---

### Q1: Design a REST API for a todo list application

**Resources**: `users`, `lists`, `tasks`

```
POST    /api/v1/auth/register              → Register
POST    /api/v1/auth/login                 → Login

GET     /api/v1/lists                      → Get user's lists
POST    /api/v1/lists                      → Create list
GET     /api/v1/lists/{id}                 → Get list
PATCH   /api/v1/lists/{id}                 → Update list
DELETE  /api/v1/lists/{id}                 → Delete list

GET     /api/v1/lists/{id}/tasks           → Get tasks in list
POST    /api/v1/lists/{id}/tasks           → Create task
PATCH   /api/v1/lists/{id}/tasks/{tid}     → Update task (mark complete, edit)
DELETE  /api/v1/lists/{id}/tasks/{tid}     → Delete task
POST    /api/v1/lists/{id}/tasks/reorder   → Reorder tasks
```

**Key Contract — Create Task:**
```json
POST /api/v1/lists/list_1/tasks
{
  "title": "Buy groceries",
  "description": "Milk, eggs, bread",
  "due_date": "2024-01-20T18:00:00Z",
  "priority": "high"
}

→ 201 Created
{
  "data": {
    "id": "task_abc",
    "title": "Buy groceries",
    "description": "Milk, eggs, bread",
    "completed": false,
    "due_date": "2024-01-20T18:00:00Z",
    "priority": "high",
    "position": 3,
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

---

### Q2: Design a REST API for a URL shortener

**Resources**: `urls`, `analytics`

```
POST    /api/v1/urls                          → Create short URL
GET     /api/v1/urls/{short_code}             → Get URL details
DELETE  /api/v1/urls/{short_code}             → Delete short URL
PATCH   /api/v1/urls/{short_code}             → Update (change destination)
GET     /api/v1/urls/{short_code}/analytics   → Get click analytics
GET     /{short_code}                         → REDIRECT to original URL (302)
```

**Key Contract — Shorten URL:**
```json
POST /api/v1/urls
{
  "url": "https://www.example.com/very/long/path?with=params",
  "custom_code": "my-link",       // optional custom short code
  "expires_at": "2024-12-31T23:59:59Z"  // optional expiry
}

→ 201 Created
{
  "data": {
    "short_code": "my-link",
    "short_url": "https://sho.rt/my-link",
    "original_url": "https://www.example.com/very/long/path?with=params",
    "expires_at": "2024-12-31T23:59:59Z",
    "created_at": "2024-01-15T10:30:00Z",
    "click_count": 0
  }
}
```

---

### Q3: Design a REST API for a blog platform

**Resources**: `users`, `posts`, `comments`, `categories`, `tags`

```
POST    /api/v1/auth/register
POST    /api/v1/auth/login

GET     /api/v1/posts                      → List posts (paginated, filterable)
POST    /api/v1/posts                      → Create post
GET     /api/v1/posts/{id}                 → Get post
PUT     /api/v1/posts/{id}                 → Update post
DELETE  /api/v1/posts/{id}                 → Delete post
POST    /api/v1/posts/{id}/publish         → Publish draft

GET     /api/v1/posts/{id}/comments        → List comments
POST    /api/v1/posts/{id}/comments        → Add comment
DELETE  /api/v1/posts/{id}/comments/{cid}  → Delete comment

POST    /api/v1/posts/{id}/like            → Like post
DELETE  /api/v1/posts/{id}/like            → Unlike post

GET     /api/v1/categories                 → List categories
GET     /api/v1/tags                       → List tags
GET     /api/v1/tags/{tag}/posts           → Posts by tag

GET     /api/v1/users/{id}/posts           → User's posts
GET     /api/v1/search/posts?q=...         → Search posts
GET     /api/v1/posts/feed                 → Personalized feed
```

**Discussion Points:**
- Pagination: cursor-based for feed, offset for admin dashboard
- Draft vs published states (status field)
- Rich text content (Markdown or HTML in body)
- SEO: slugs vs IDs in URLs (`/posts/my-first-post` vs `/posts/post_123`)

---

### Q4: Design a REST API for a contact management system

```
GET     /api/v1/contacts                             → List contacts
POST    /api/v1/contacts                             → Create contact
GET     /api/v1/contacts/{id}                        → Get contact
PATCH   /api/v1/contacts/{id}                        → Update contact
DELETE  /api/v1/contacts/{id}                        → Delete contact
POST    /api/v1/contacts/import                      → Bulk import (CSV)
GET     /api/v1/contacts/export?format=csv           → Export contacts

GET     /api/v1/contacts/{id}/notes                  → Get notes for contact
POST    /api/v1/contacts/{id}/notes                  → Add note

GET     /api/v1/groups                               → List groups
POST    /api/v1/groups                               → Create group
POST    /api/v1/groups/{id}/contacts                 → Add contact to group
DELETE  /api/v1/groups/{id}/contacts/{cid}           → Remove from group

GET     /api/v1/contacts?q=alice&tag=vip&sort=-updated_at&limit=20
```

---

### Q5: Design a REST API for a file storage service (like Dropbox)

```
GET     /api/v1/files                                → List files/folders
POST    /api/v1/files/upload                         → Upload file
POST    /api/v1/files/upload/initiate                → Initiate large upload
PUT     /api/v1/files/upload/{upload_id}/chunks/{n}  → Upload chunk
POST    /api/v1/files/upload/{upload_id}/complete     → Complete upload
GET     /api/v1/files/{id}                           → Get file metadata
GET     /api/v1/files/{id}/download                  → Download file
DELETE  /api/v1/files/{id}                           → Delete file
PATCH   /api/v1/files/{id}                           → Rename / move file
POST    /api/v1/files/{id}/copy                      → Copy file

POST    /api/v1/folders                              → Create folder
GET     /api/v1/folders/{id}/contents                → List folder contents

POST    /api/v1/files/{id}/share                     → Create share link
GET     /api/v1/shares/{token}                       → Access shared file
DELETE  /api/v1/files/{id}/share                     → Revoke share

GET     /api/v1/files/{id}/versions                  → List file versions
GET     /api/v1/files/{id}/versions/{vid}/download   → Download specific version
```

---

# MEDIUM (11-25) — Complex API Design

---

### Q6: Design a REST API for an e-commerce platform

**Scope**: Product catalog, cart, orders, payments, reviews

```
PRODUCTS
────────────────────────────────────
GET     /api/v1/products                           → List (filter, sort, paginate)
GET     /api/v1/products/{id}                      → Get product
GET     /api/v1/products/{id}/reviews              → Get reviews
POST    /api/v1/products/{id}/reviews              → Write review
GET     /api/v1/categories                         → List categories
GET     /api/v1/categories/{id}/products           → Products in category
GET     /api/v1/search/products?q=...              → Search

CART
────────────────────────────────────
GET     /api/v1/cart                               → Get current user's cart
POST    /api/v1/cart/items                         → Add item
PATCH   /api/v1/cart/items/{item_id}               → Update quantity
DELETE  /api/v1/cart/items/{item_id}               → Remove item
DELETE  /api/v1/cart                               → Clear cart
POST    /api/v1/cart/checkout                      → Start checkout

ORDERS
────────────────────────────────────
GET     /api/v1/orders                             → List user's orders
POST    /api/v1/orders                             → Place order
GET     /api/v1/orders/{id}                        → Get order details
POST    /api/v1/orders/{id}/cancel                 → Cancel order
GET     /api/v1/orders/{id}/tracking               → Track shipment
POST    /api/v1/orders/{id}/return                 → Request return

PAYMENTS
────────────────────────────────────
POST    /api/v1/payment-intents                    → Create payment intent
POST    /api/v1/payment-intents/{id}/confirm       → Confirm payment
GET     /api/v1/payment-methods                    → List saved payment methods
POST    /api/v1/payment-methods                    → Add payment method

ADDRESSES
────────────────────────────────────
GET     /api/v1/addresses                          → List user's addresses
POST    /api/v1/addresses                          → Add address
PATCH   /api/v1/addresses/{id}                     → Update address
DELETE  /api/v1/addresses/{id}                     → Delete address
```

**Key Contract — Checkout Flow:**
```json
// Step 1: Checkout → creates order
POST /api/v1/cart/checkout
{
  "shipping_address_id": "addr_123",
  "payment_method_id": "pm_456",
  "coupon_code": "SAVE20"
}

→ 201 Created
{
  "data": {
    "order_id": "ord_789",
    "payment_intent_id": "pi_abc",
    "status": "pending_payment",
    "subtotal": 149.98,
    "discount": -30.00,
    "shipping": 5.99,
    "tax": 12.60,
    "total": 138.57,
    "payment_url": "/api/v1/payment-intents/pi_abc/confirm"
  }
}

// Step 2: Confirm payment
POST /api/v1/payment-intents/pi_abc/confirm
{
  "payment_method_id": "pm_456"
}

→ 200 OK
{
  "data": {
    "order_id": "ord_789",
    "status": "confirmed",
    "estimated_delivery": "2024-01-22"
  }
}
```

---

### Q7: Design a REST API for a social media platform (Instagram-like)

```
USERS & PROFILES
────────────────────────────────────
GET     /api/v1/users/{id}                         → Get profile
PATCH   /api/v1/users/{id}                         → Update profile
POST    /api/v1/users/{id}/follow                  → Follow
DELETE  /api/v1/users/{id}/follow                  → Unfollow
GET     /api/v1/users/{id}/followers               → List followers
GET     /api/v1/users/{id}/following               → List following
POST    /api/v1/users/{id}/block                   → Block user
DELETE  /api/v1/users/{id}/block                   → Unblock user

POSTS
────────────────────────────────────
GET     /api/v1/feed                               → Home feed (cursor-paginated)
POST    /api/v1/posts                              → Create post (with media)
GET     /api/v1/posts/{id}                         → Get post
DELETE  /api/v1/posts/{id}                         → Delete post
POST    /api/v1/posts/{id}/like                    → Like
DELETE  /api/v1/posts/{id}/like                    → Unlike
GET     /api/v1/posts/{id}/likes                   → List who liked
GET     /api/v1/posts/{id}/comments                → List comments
POST    /api/v1/posts/{id}/comments                → Add comment
DELETE  /api/v1/posts/{id}/comments/{cid}          → Delete comment
POST    /api/v1/posts/{id}/save                    → Bookmark
DELETE  /api/v1/posts/{id}/save                    → Remove bookmark

STORIES
────────────────────────────────────
GET     /api/v1/stories                            → List stories from followed users
POST    /api/v1/stories                            → Create story
GET     /api/v1/users/{id}/stories                 → Get user's stories
DELETE  /api/v1/stories/{id}                       → Delete story

MEDIA
────────────────────────────────────
POST    /api/v1/media/upload                       → Upload media (pre-signed URL)

SEARCH & EXPLORE
────────────────────────────────────
GET     /api/v1/search?q=...&type=user|post|tag    → Search
GET     /api/v1/explore                            → Explore/discover feed
GET     /api/v1/hashtags/{tag}/posts               → Posts by hashtag

NOTIFICATIONS
────────────────────────────────────
GET     /api/v1/notifications                      → List notifications
PATCH   /api/v1/notifications/{id}/read            → Mark as read
POST    /api/v1/notifications/read-all             → Mark all as read
```

**Discussion Points:**
- Feed pagination must be cursor-based (fan-out-on-read vs fan-out-on-write)
- Stories auto-expire after 24 hours (TTL)
- Media upload via pre-signed URLs (don't stream through API server)
- Rate limit post creation aggressively (spam prevention)

---

### Q8: Design a REST API for a ride-sharing service (Uber)

```
RIDERS
────────────────────────────────────
GET     /api/v1/riders/me                          → Get rider profile
PATCH   /api/v1/riders/me                          → Update profile
GET     /api/v1/riders/me/payment-methods           → List payment methods
POST    /api/v1/riders/me/payment-methods           → Add payment method

RIDES
────────────────────────────────────
POST    /api/v1/rides/estimate                     → Get fare estimate
POST    /api/v1/rides                              → Request ride
GET     /api/v1/rides/{id}                         → Get ride details
POST    /api/v1/rides/{id}/cancel                  → Cancel ride
GET     /api/v1/rides/{id}/track                   → Live tracking (SSE/polling)
POST    /api/v1/rides/{id}/rate                    → Rate the ride
GET     /api/v1/rides/history                      → Ride history

DRIVERS
────────────────────────────────────
PATCH   /api/v1/drivers/me/status                  → Go online/offline
PUT     /api/v1/drivers/me/location                → Update location (frequent)
GET     /api/v1/drivers/me/rides                   → Current/upcoming rides
POST    /api/v1/drivers/rides/{id}/accept           → Accept ride request
POST    /api/v1/drivers/rides/{id}/arrive           → Arrived at pickup
POST    /api/v1/drivers/rides/{id}/start            → Start ride
POST    /api/v1/drivers/rides/{id}/complete         → Complete ride
```

**Key Contract — Request Ride:**
```json
POST /api/v1/rides
{
  "pickup": { "lat": 37.7749, "lng": -122.4194, "address": "123 Market St" },
  "dropoff": { "lat": 37.7849, "lng": -122.4094, "address": "456 Mission St" },
  "ride_type": "uberx",
  "payment_method_id": "pm_123"
}

→ 201 Created
{
  "data": {
    "id": "ride_abc",
    "status": "matching",
    "estimated_arrival": "3 min",
    "estimated_fare": { "min": 12.50, "max": 16.00, "currency": "USD" },
    "tracking_url": "/api/v1/rides/ride_abc/track"
  }
}
```

---

### Q9: Design a REST API for a payment gateway (Stripe-like)

```
CUSTOMERS
────────────────────────────────────
POST    /api/v1/customers                          → Create customer
GET     /api/v1/customers/{id}                     → Get customer
PATCH   /api/v1/customers/{id}                     → Update customer
DELETE  /api/v1/customers/{id}                     → Delete customer

PAYMENT METHODS
────────────────────────────────────
POST    /api/v1/payment-methods                    → Create (tokenize card)
GET     /api/v1/payment-methods/{id}               → Get payment method
POST    /api/v1/payment-methods/{id}/attach         → Attach to customer
POST    /api/v1/payment-methods/{id}/detach         → Detach from customer

PAYMENT INTENTS
────────────────────────────────────
POST    /api/v1/payment-intents                    → Create intent
GET     /api/v1/payment-intents/{id}               → Get intent
POST    /api/v1/payment-intents/{id}/confirm       → Confirm payment
POST    /api/v1/payment-intents/{id}/capture       → Capture authorized amount
POST    /api/v1/payment-intents/{id}/cancel        → Cancel intent

REFUNDS
────────────────────────────────────
POST    /api/v1/refunds                            → Create refund
GET     /api/v1/refunds/{id}                       → Get refund

WEBHOOKS
────────────────────────────────────
POST    /api/v1/webhooks                           → Register webhook
GET     /api/v1/webhooks                           → List webhooks
DELETE  /api/v1/webhooks/{id}                      → Deregister webhook

EVENTS
────────────────────────────────────
GET     /api/v1/events                             → List events (audit log)
GET     /api/v1/events/{id}                        → Get event details
```

**Discussion Points:**
- Idempotency keys are CRITICAL (payment duplication = catastrophic)
- Payment Intents pattern (two-step: authorize → capture)
- PCI compliance: never store raw card numbers; use tokenization
- Webhook signatures for security (HMAC-SHA256)
- Strong consistency required (no eventual consistency for payments)

---

### Q10: Design a REST API for a chat application (Slack-like)

```
WORKSPACES
────────────────────────────────────
POST    /api/v1/workspaces                         → Create workspace
GET     /api/v1/workspaces/{id}                    → Get workspace
POST    /api/v1/workspaces/{id}/invite             → Invite user

CHANNELS
────────────────────────────────────
GET     /api/v1/channels                           → List channels
POST    /api/v1/channels                           → Create channel
GET     /api/v1/channels/{id}                      → Get channel info
PATCH   /api/v1/channels/{id}                      → Update channel
POST    /api/v1/channels/{id}/join                 → Join channel
POST    /api/v1/channels/{id}/leave                → Leave channel
GET     /api/v1/channels/{id}/members              → List members

MESSAGES
────────────────────────────────────
GET     /api/v1/channels/{id}/messages             → Get messages (cursor-paginated)
POST    /api/v1/channels/{id}/messages             → Send message
PATCH   /api/v1/messages/{id}                      → Edit message
DELETE  /api/v1/messages/{id}                      → Delete message
POST    /api/v1/messages/{id}/reactions             → Add reaction
DELETE  /api/v1/messages/{id}/reactions/{emoji}     → Remove reaction
GET     /api/v1/messages/{id}/thread               → Get thread replies
POST    /api/v1/messages/{id}/thread               → Reply in thread

DIRECT MESSAGES
────────────────────────────────────
GET     /api/v1/dm                                 → List DM conversations
POST    /api/v1/dm                                 → Start DM conversation
GET     /api/v1/dm/{id}/messages                   → Get DM messages
POST    /api/v1/dm/{id}/messages                   → Send DM

REAL-TIME (WebSocket)
────────────────────────────────────
WS      /api/v1/ws                                 → WebSocket connection
Events: message.new, message.updated, message.deleted,
        typing.start, typing.stop,
        presence.online, presence.offline,
        channel.member_joined, channel.member_left

SEARCH
────────────────────────────────────
GET     /api/v1/search?q=...&in=channel_123        → Search messages
```

---

# HARD (11-20) — System Design Level API Design

---

### Q11: Design a REST API for a notification system

```
NOTIFICATIONS (consumer-facing)
────────────────────────────────────
GET     /api/v1/notifications                           → List (paginated)
GET     /api/v1/notifications/unread-count              → Get unread count
PATCH   /api/v1/notifications/{id}/read                 → Mark as read
POST    /api/v1/notifications/mark-all-read             → Mark all as read
DELETE  /api/v1/notifications/{id}                      → Dismiss

PREFERENCES (consumer-facing)
────────────────────────────────────
GET     /api/v1/notification-preferences                → Get preferences
PATCH   /api/v1/notification-preferences                → Update preferences
{
  "email": { "marketing": false, "order_updates": true, "security": true },
  "push": { "marketing": true, "order_updates": true, "security": true },
  "sms": { "marketing": false, "order_updates": false, "security": true }
}

SEND (internal service API)
────────────────────────────────────
POST    /internal/v1/notifications/send
{
  "user_ids": ["user_123", "user_456"],
  "type": "order_shipped",
  "channels": ["push", "email"],         // or null = use preferences
  "data": {
    "order_id": "ord_789",
    "tracking_url": "https://..."
  },
  "template": "order_shipped_v2",
  "priority": "high",                     // high = immediate, low = batched
  "idempotency_key": "ord_789_shipped"
}
```

**Discussion Points:**
- Fan-out: one event → multiple users × multiple channels
- Priority queue (security alerts > marketing)
- Deduplication (idempotency key prevents duplicate sends)
- Notification templates (server-side rendering for email/push)
- Real-time delivery via WebSocket + push notification fallback

---

### Q12: Design a REST API for a rate-limited API gateway

```
API KEYS
────────────────────────────────────
POST    /admin/v1/api-keys                         → Create API key
GET     /admin/v1/api-keys                         → List API keys
GET     /admin/v1/api-keys/{id}                    → Get key details
PATCH   /admin/v1/api-keys/{id}                    → Update (permissions, limits)
DELETE  /admin/v1/api-keys/{id}                    → Revoke key

RATE LIMIT POLICIES
────────────────────────────────────
POST    /admin/v1/rate-limits                      → Create policy
GET     /admin/v1/rate-limits                      → List policies
PATCH   /admin/v1/rate-limits/{id}                 → Update policy
DELETE  /admin/v1/rate-limits/{id}                 → Delete policy

{
  "name": "free_tier",
  "rules": [
    { "endpoint": "*", "limit": 100, "window": "1m" },
    { "endpoint": "POST /orders", "limit": 10, "window": "1m" },
    { "endpoint": "GET /search", "limit": 30, "window": "1m" }
  ]
}

ANALYTICS
────────────────────────────────────
GET     /admin/v1/analytics/usage                  → API usage stats
GET     /admin/v1/analytics/usage?api_key=...      → Per-key usage
GET     /admin/v1/analytics/errors                 → Error rates
GET     /admin/v1/analytics/latency                → Latency percentiles
```

---

### Q13: Design APIs for a multi-tenant SaaS platform

**Key pattern:** Tenant isolation in URL or header

```
Option A: Tenant in URL path
  /api/v1/tenants/{tenant_id}/users
  /api/v1/tenants/{tenant_id}/projects

Option B: Tenant in subdomain
  acme.api.example.com/v1/users
  bigcorp.api.example.com/v1/users

Option C: Tenant in header (most common for SaaS)
  X-Tenant-ID: tenant_acme
  GET /api/v1/users

TENANT MANAGEMENT (super admin)
────────────────────────────────────
POST    /admin/v1/tenants                          → Create tenant
GET     /admin/v1/tenants                          → List tenants
PATCH   /admin/v1/tenants/{id}                     → Update tenant
POST    /admin/v1/tenants/{id}/suspend             → Suspend tenant

TENANT-SCOPED (auto-filtered by tenant)
────────────────────────────────────
GET     /api/v1/users                              → List users (in current tenant)
POST    /api/v1/users                              → Create user (in current tenant)
GET     /api/v1/projects                           → List projects (in current tenant)
```

---

### Q14: Design a webhook delivery system API

```
WEBHOOK ENDPOINTS (consumer registers)
────────────────────────────────────
POST    /api/v1/webhooks
{
  "url": "https://myapp.com/webhooks",
  "events": ["order.*", "payment.completed"],
  "secret": "whsec_..."
}

GET     /api/v1/webhooks                           → List registered webhooks
PATCH   /api/v1/webhooks/{id}                      → Update (URL, events)
DELETE  /api/v1/webhooks/{id}                      → Delete webhook

WEBHOOK DELIVERIES (observability)
────────────────────────────────────
GET     /api/v1/webhooks/{id}/deliveries           → List delivery attempts
GET     /api/v1/webhooks/{id}/deliveries/{did}     → Get delivery details
POST    /api/v1/webhooks/{id}/deliveries/{did}/retry → Retry failed delivery
POST    /api/v1/webhooks/{id}/test                 → Send test event

EVENTS
────────────────────────────────────
GET     /api/v1/events                             → List events
GET     /api/v1/events/{id}                        → Get event details
```

**Key Design:**
```json
// Delivery record
{
  "id": "del_123",
  "webhook_id": "wh_456",
  "event_type": "order.created",
  "status": "failed",
  "attempts": [
    {
      "attempted_at": "2024-01-15T10:30:00Z",
      "response_status": 500,
      "response_body": "Internal Server Error",
      "duration_ms": 2500
    },
    {
      "attempted_at": "2024-01-15T10:31:00Z",
      "response_status": 200,
      "response_body": "OK",
      "duration_ms": 150
    }
  ],
  "next_retry_at": null,
  "max_retries": 5,
  "retry_count": 1
}

// Retry policy: exponential backoff
// Attempt 1: immediate
// Attempt 2: 1 minute
// Attempt 3: 5 minutes
// Attempt 4: 30 minutes
// Attempt 5: 2 hours
// Then: marked as failed, alert webhook owner
```

---

### Q15: Design a REST API for a content moderation system

```
CONTENT REVIEW
────────────────────────────────────
POST    /api/v1/moderation/submit
{
  "content_type": "text|image|video",
  "content_id": "post_123",
  "content_url": "https://cdn.example.com/images/abc.jpg",
  "text_content": "...",
  "context": { "user_id": "user_456", "source": "post" }
}

→ 202 Accepted (async processing)
{
  "review_id": "rev_789",
  "status": "pending",
  "status_url": "/api/v1/moderation/reviews/rev_789"
}

GET     /api/v1/moderation/reviews/{id}            → Get review result
GET     /api/v1/moderation/reviews?status=pending   → List pending reviews

HUMAN REVIEW (moderator dashboard)
────────────────────────────────────
GET     /api/v1/moderation/queue                   → Get moderation queue
POST    /api/v1/moderation/reviews/{id}/approve    → Approve content
POST    /api/v1/moderation/reviews/{id}/reject     → Reject content
POST    /api/v1/moderation/reviews/{id}/escalate   → Escalate to senior mod

APPEALS
────────────────────────────────────
POST    /api/v1/moderation/appeals                 → User appeals decision
GET     /api/v1/moderation/appeals/{id}            → Get appeal status

POLICIES
────────────────────────────────────
GET     /api/v1/moderation/policies                → List moderation policies
POST    /api/v1/moderation/policies                → Create policy rule
```

---

# SCENARIO-BASED QUESTIONS (21-40)

---

### Q16: "How do you design an API that needs to support both mobile and web clients efficiently?"

**Answer:**
```
1. Use GraphQL for BFF (Backend for Frontend) layer
   - Mobile: request minimal fields (save bandwidth)
   - Web: request more fields (richer UI)

2. If REST: use field selection + expansion
   GET /users/123?fields=name,avatar&expand=recent_orders(limit:3)

3. Response compression (gzip, brotli)

4. Image optimization: return different sizes
   "avatar": {
     "small": "https://cdn.../s.jpg",     // 50x50 for mobile
     "medium": "https://cdn.../m.jpg",     // 200x200 for web
     "large": "https://cdn.../l.jpg"       // 800x800 for profile page
   }

5. Pagination: smaller page sizes for mobile (10 vs 20)

6. Consider separate BFF services if needs diverge significantly
```

---

### Q17: "How do you handle API versioning when you have mobile clients that can't be forced to update?"

**Answer:**
```
1. URI versioning: /v1/, /v2/ — both run simultaneously
2. Long support windows: maintain v1 for 12-18 months after v2 launch
3. Feature flags within versions for gradual rollout
4. Deprecation headers on v1 responses: Sunset: <date>
5. Track v1 usage per app version — reach out to top users
6. In-app update prompts for mobile users on old versions
7. Never remove fields from a version — only add
8. Use API gateway to route versions to different backends
```

---

### Q18: "How do you design idempotent APIs for a payment system?"

**Answer:**
```
1. Idempotency-Key header on all POST requests
2. Server stores {key → response} in Redis/DB with TTL (24-48 hours)
3. On duplicate key: return cached response (don't reprocess)
4. Key should be client-generated UUID
5. For payment specifically:
   - Payment Intent pattern (Stripe): create → confirm → capture
   - Each step is independently idempotent
   - Use database transactions for state changes
   - Log all payment operations for audit
6. Handle race conditions: use distributed lock on idempotency key
```

---

### Q19: "Design the API for a search-as-you-type autocomplete feature"

**Answer:**
```
GET /api/v1/autocomplete?q=wirele&type=product&limit=5

→ 200 OK
{
  "data": {
    "query": "wirele",
    "suggestions": [
      { "text": "wireless headphones", "type": "product", "count": 1250 },
      { "text": "wireless mouse", "type": "product", "count": 890 },
      { "text": "wireless charger", "type": "product", "count": 650 },
      { "text": "wireless keyboard", "type": "product", "count": 420 },
      { "text": "wireless earbuds", "type": "product", "count": 380 }
    ],
    "took_ms": 12
  }
}

Design considerations:
- Debounce on client (200-300ms) to avoid flooding
- Response time < 100ms (use Elasticsearch, prefix trees, or Redis)
- Cache popular prefixes
- Rate limit per user (50 req/min for autocomplete)
- Support multiple types: products, categories, users
```

---

### Q20: "How do you handle partial failures in a bulk API?"

**Answer:**
```json
// Use 207 Multi-Status — individual status per operation
POST /api/v1/orders/batch

// Response: 207 Multi-Status
{
  "results": [
    { "index": 0, "status": 201, "data": { "id": "ord_1" } },
    { "index": 1, "status": 422, "error": { "code": "OUT_OF_STOCK", "message": "..." } },
    { "index": 2, "status": 201, "data": { "id": "ord_3" } },
    { "index": 3, "status": 409, "error": { "code": "DUPLICATE", "message": "..." } }
  ],
  "summary": {
    "total": 4,
    "succeeded": 2,
    "failed": 2
  }
}

// Client decides: retry failed ones, report errors, or rollback all.
// API doesn't roll back succeeds — operations are independent.
```

---

## Quick Reference: API Design Decision Matrix

| Design Question | Recommendation |
|----------------|----------------|
| What protocol? | REST for public, gRPC for internal, GraphQL for complex UIs |
| How to paginate? | Cursor-based for feeds, offset for admin, keyset for DB-heavy |
| How to authenticate? | JWT + refresh tokens for users, API keys for services |
| How to version? | URI path (/v1/) for simplicity |
| How to handle errors? | Consistent JSON format with machine + human readable codes |
| How to handle bulk ops? | 207 Multi-Status with per-item results |
| How to handle async? | 202 Accepted + polling or webhooks |
| How to handle files? | Pre-signed URLs for upload, CDN for download |
| How to rate limit? | Sliding window, per-user + per-endpoint |
| How to search? | Dedicated search endpoint, Elasticsearch backend |
| How to do real-time? | WebSocket for bidirectional, SSE for server-push |
| How to handle auth? | Middleware pattern, separate from business logic |
