# Design a Notification Service

## 1. Problem Statement & Requirements

### Problem Statement
Design a notification service that can deliver 10 billion notifications per day across multiple channels: push notifications (iOS/Android), SMS, email, and in-app notifications. The system must support template management, rate limiting, priority handling, and user preferences.

### Functional Requirements
- **Multi-channel delivery**: Push (APNs, FCM), SMS, email, in-app
- **Template management**: Parameterized templates per channel
- **User preferences**: Opt-in/opt-out per channel and category
- **Rate limiting**: Per user, per channel, global
- **Priority**: Urgent vs normal vs low
- **Delivery tracking**: Sent, delivered, opened, failed
- **Retry**: Exponential backoff for failures
- **Deduplication**: Avoid duplicate notifications

### Non-Functional Requirements
- **Throughput**: 10B notifications/day ≈ 115K QPS
- **Latency**: < 5s from request to delivery attempt
- **Reliability**: At-least-once delivery, no loss
- **Availability**: 99.99%

### Out of Scope
- Rich media in push (images, actions)
- A/B testing templates
- Advanced analytics dashboard
- In-app notification UI components

---

## 2. Back-of-Envelope Estimation

### Scale Assumptions
- **Notifications/day**: 10 billion
- **Channels**: Push 60%, Email 25%, SMS 10%, In-app 5%
- **Peak factor**: 3x average
- **Retry rate**: 10%

### QPS Calculation
| Channel | Daily | QPS (avg) | QPS (peak) |
|---------|-------|-----------|------------|
| Push | 6B | ~70K | ~210K |
| Email | 2.5B | ~29K | ~87K |
| SMS | 1B | ~12K | ~36K |
| In-app | 0.5B | ~6K | ~18K |
| **Total** | **10B** | **~115K** | **~350K** |

### Storage (30 days retention)
- **Notification log**: 10B × 30 × 200 bytes ≈ 60 TB
- **Templates**: 10K × 5KB ≈ 50 MB
- **User preferences**: 1B × 500 bytes ≈ 500 GB

### Bandwidth
- **To third parties**: 10B × 1KB avg ≈ 10 TB/day
- **Internal**: Queue traffic, API calls

### Cache
- **Templates**: Redis, 100% hit
- **User preferences**: Redis, 80% hit
- **Rate limit counters**: Redis

---

## 3. API Design

### REST Endpoints

```
POST   /api/v1/notifications/send
Body: {
  "user_id": "u123",
  "channel": "push|sms|email|in_app",
  "template_id": "welcome_email",
  "parameters": {"name": "John", "link": "..."},
  "priority": "urgent|normal|low",
  "idempotency_key": "optional",
  "metadata": {"source": "order_confirmation"}
}
Response: { "notification_id": "n123", "status": "queued" }

POST   /api/v1/notifications/send-batch
Body: { "notifications": [...] }
Response: { "batch_id": "b123", "results": [...] }

GET    /api/v1/notifications/:id
Response: { "status", "delivered_at", "opened_at", "error" }

GET    /api/v1/users/:id/preferences
PUT    /api/v1/users/:id/preferences
Body: { "push_enabled": true, "email_categories": ["marketing"], "sms_opt_out": false }

GET    /api/v1/templates
POST   /api/v1/templates
PUT    /api/v1/templates/:id
```

---

## 4. Data Model / Database Schema

### Database Choice
- **Notifications**: Cassandra (write-heavy, time-series)
- **Templates**: MySQL/PostgreSQL (relational)
- **User preferences**: Cassandra (key-value by user)
- **Delivery tracking**: Cassandra
- **Rate limits**: Redis

### Schema

**Templates (MySQL)**
```sql
templates (
  template_id VARCHAR(50) PRIMARY KEY,
  channel VARCHAR(20),
  name VARCHAR(100),
  subject VARCHAR(500),      -- for email
  body TEXT,
  variables JSON,           -- ["name", "link", ...]
  created_at TIMESTAMP,
  updated_at TIMESTAMP
)
```

**Notifications (Cassandra)**
```sql
notifications (
  notification_id TIMEUUID,
  user_id BIGINT,
  channel VARCHAR(20),
  template_id VARCHAR(50),
  parameters MAP<TEXT,TEXT>,
  priority INT,
  status VARCHAR(20),        -- queued, sent, delivered, failed
  created_at TIMESTAMP,
  sent_at TIMESTAMP,
  error_message TEXT,
  idempotency_key VARCHAR(100),
  PRIMARY KEY (user_id, notification_id)
) WITH CLUSTERING ORDER BY (notification_id DESC);

notifications_by_id (
  notification_id TIMEUUID PRIMARY KEY,
  user_id BIGINT,
  ...
)
```

**User Preferences (Cassandra)**
```sql
user_preferences (
  user_id BIGINT PRIMARY KEY,
  push_enabled BOOLEAN,
  email_enabled BOOLEAN,
  sms_enabled BOOLEAN,
  email_categories SET<TEXT>,
  quiet_hours_start INT,     -- 22 = 10 PM
  quiet_hours_end INT,       -- 8 = 8 AM
  updated_at TIMESTAMP
)
```

**Rate Limits (Redis)**
```
rate_limit:user:{user_id}:push -> counter (incr, expire 1h)
rate_limit:user:{user_id}:sms -> counter
rate_limit:global:sms -> counter
```

---

## 5. High-Level Architecture

```
                                    ┌─────────────────────────────────────────────────────────────┐
                                    │                    CLIENT APPLICATIONS                      │
                                    │              (Order Service, Auth, Marketing, etc.)          │
                                    └─────────────────────────────────────────────────────────────┘
                                                              │
                                                              ▼
┌─────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                          API GATEWAY / LOAD BALANCER                                          │
└─────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
                                                              │
                                                              ▼
┌─────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                    NOTIFICATION ORCHESTRATOR SERVICE                                          │
│  • Validate request  • Check user preferences  • Rate limit  • Deduplicate  • Enrich template  • Route       │
└─────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
                                                              │
                                                              ▼
┌─────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                              MESSAGE QUEUES (Kafka) - Per Channel, Per Priority                               │
│  push-urgent | push-normal | push-low | email-urgent | email-normal | sms-urgent | sms-normal | inapp        │
└─────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
                                                              │
        ┌─────────────────────┬─────────────────────┬─────────────────────┬─────────────────────┐
        ▼                     ▼                     ▼                     ▼                     ▼
┌───────────────┐     ┌───────────────┐     ┌───────────────┐     ┌───────────────┐     ┌───────────────┐
│ Push Worker   │     │ Email Worker  │     │  SMS Worker   │     │ In-App Worker │     │  Template     │
│ ───────────── │     │ ───────────── │     │ ───────────── │     │ ───────────── │     │  Service      │
│ • FCM         │     │ • SendGrid    │     │ • Twilio      │     │ • WebSocket   │     │ ─────────────  │
│ • APNs        │     │ • SES         │     │ • SNS         │     │ • Store DB    │     │ • Render      │
│ • Retry       │     │ • Retry       │     │ • Retry       │     │               │     │ • Variables   │
└───────┬───────┘     └───────┬───────┘     └───────┬───────┘     └───────┬───────┘     └───────────────┘
        │                     │                     │                     │
        ▼                     ▼                     ▼                     ▼
┌───────────────┐     ┌───────────────┐     ┌───────────────┐     ┌───────────────┐
│ APNs / FCM    │     │ SendGrid/SES  │     │ Twilio/SNS    │     │ Cassandra     │
│ (Apple/Google)│     │ (Email)       │     │ (SMS)         │     │ (In-app)      │
└───────────────┘     └───────────────┘     └───────────────┘     └───────────────┘
```

### Notification Flow (ASCII)

```
    CLIENT                    ORCHESTRATOR                 QUEUE                 WORKER              THIRD PARTY
       │                            │                        │                      │                      │
       │  Send notification          │                        │                      │                      │
       │────────────────────────────>│                        │                      │                      │
       │                            │  Check preferences      │                      │                      │
       │                            │  (Redis/DB)            │                      │                      │
       │                            │  Check rate limit      │                      │                      │
       │                            │  Dedupe (idempotency)  │                      │                      │
       │                            │  Render template       │                      │                      │
       │                            │  Store in Cassandra    │                      │                      │
       │                            │  Publish to queue      │                      │                      │
       │                            │───────────────────────>│                      │                      │
       │  Return notification_id    │                        │  Consume             │                      │
       │<────────────────────────────│                        │<─────────────────────│                      │
       │                            │                        │                      │  Call APNs/FCM/etc  │
       │                            │                        │                      │─────────────────────>│
       │                            │                        │                      │  Update status       │
       │                            │                        │                      │  (delivered/failed)  │
       │                            │                        │                      │  Retry if failed     │
       │                            │                        │                      │  (exponential backoff)│
```

---

## 6. Detailed Component Design

### 6.1 Notification Types and Channels
| Channel | Use Case | Third-Party | Latency |
|---------|----------|--------------|---------|
| Push | App alerts, engagement | APNs, FCM | < 1s |
| Email | Transactional, marketing | SendGrid, SES | 1-30s |
| SMS | OTP, critical alerts | Twilio, SNS | 1-5s |
| In-app | Bell icon, feed | Internal | < 100ms |

### 6.2 Notification Orchestrator
**Responsibilities**:
1. **Validate**: user_id, template_id, channel, parameters
2. **User preferences**: Fetch from Redis/DB, check if channel enabled
3. **Rate limiting**: Check Redis counters (per user, per channel)
4. **Deduplication**: idempotency_key → return existing if present
5. **Template rendering**: Fetch template, substitute variables
6. **Persist**: Write to Cassandra (status=queued)
7. **Route**: Publish to appropriate Kafka topic

### 6.3 Message Queue (Per Channel)
- **Kafka topics**: `push-urgent`, `push-normal`, `push-low`, `email-urgent`, etc.
- **Partitioning**: By user_id (ordering per user)
- **Consumer groups**: One per channel, scale by partition count
- **Priority**: Urgent consumed first (separate topics)

### 6.4 Third-Party Integration
**Push (APNs/FCM)**:
- **APNs**: HTTP/2 API, certificate-based auth
- **FCM**: HTTP v1 API, OAuth
- **Batch**: Send up to 500 per batch (FCM)
- **Response**: Parse success/failure, handle invalid tokens

**Email (SendGrid/SES)**:
- **API**: REST, batch send (1000/request)
- **Templates**: Use dynamic templates or render server-side
- **Bounce handling**: Webhook for bounce, update user preference

**SMS (Twilio/SNS)**:
- **API**: REST, 1 message per request typically
- **Rate**: Carrier limits, cost considerations
- **Opt-out**: Handle STOP keyword

### 6.5 Template Engine
- **Storage**: MySQL, versioned
- **Variables**: `{{name}}`, `{{link}}` placeholders
- **Rendering**: Simple string replace or Jinja2
- **Validation**: Ensure all required variables provided
- **Caching**: Redis, invalidate on update

### 6.6 Rate Limiting
- **Per user**: 100 push/hour, 10 SMS/day, 50 email/day
- **Per channel**: Redis `INCR` with `EXPIRE`
- **Global**: SMS has regulatory limits
- **Implementation**: Check before enqueue, reject if over

### 6.7 Priority Queues
- **Urgent**: OTP, security alerts - dedicated topic, more consumers
- **Normal**: Order confirmation, general
- **Low**: Marketing, can be delayed
- **Quiet hours**: For non-urgent, delay until window

### 6.8 Deduplication
- **Idempotency key**: Client provides, store in Cassandra
- **Lookup**: Before processing, check if key exists
- **Window**: 24 hours
- **Return**: Existing notification_id if duplicate

### 6.9 User Preference Service
- **Storage**: Cassandra by user_id
- **Cache**: Redis, 1 hour TTL
- **Categories**: marketing, transactional, security
- **Quiet hours**: Don't send push 10PM-8AM (configurable)
- **Default**: Opt-in for marketing, opt-out for transactional

### 6.10 Delivery Tracking
- **Statuses**: queued → sent → delivered/failed
- **Webhooks**: APNs/FCM delivery receipts, email open tracking
- **Store**: Update Cassandra, emit to analytics
- **Retry**: On failure, exponential backoff (1s, 2s, 4s, ... max 5 retries)

### 6.11 Retry with Exponential Backoff
- **Transient failures**: Network, 5xx
- **Backoff**: 1s, 2s, 4s, 8s, 16s
- **Dead letter**: After 5 retries, move to DLQ for manual review
- **Circuit breaker**: If provider down, fail fast, don't pile up

---

## 7. Scaling

### Horizontal Scaling
- **Orchestrator**: Stateless, scale with load
- **Workers**: Scale Kafka consumers
- **Kafka partitions**: 100+ for push channel

### Caching
- **Templates**: Redis, all templates
- **Preferences**: Redis, hot users
- **Rate limits**: Redis (required)

### Sharding
- **Notifications**: By user_id
- **Kafka**: Partition by user_id

---

## 8. Failure Handling

### Component Failures
- **Orchestrator down**: Queue at API gateway, or client retry
- **Worker crash**: Kafka rebalance, at-least-once
- **Third-party down**: Retry queue, circuit breaker
- **Cassandra**: RF=3, quorum

### Redundancy
- **Multi-AZ**: All components
- **Kafka**: Replication 3
- **Multiple providers**: Fallback SendGrid → SES for email

### Degradation
- **Provider rate limit**: Queue, slow down
- **DB slow**: Cache more, async writes

---

## 9. Monitoring & Observability

### Key Metrics
- **Throughput**: Notifications/sec per channel
- **Latency**: Request to delivery (p50, p99)
- **Success rate**: Delivered / Sent
- **Retry rate**: % requiring retry
- **Queue depth**: Kafka lag per topic
- **Rate limit hits**: Rejections per user

### Alerts
- Queue lag > 100K
- Success rate < 95%
- p99 latency > 30s
- Third-party error rate > 5%

### Tracing
- Trace ID from request through queue to delivery
- Correlate with provider response

---

## 10. Interview Tips

### Follow-up Questions
1. **How to handle 10x spike?** Auto-scale workers, queue absorbs
2. **User gets 1000 emails?** Rate limit, burst limit, dedupe
3. **Template change mid-flight?** Use version, in-flight use old
4. **SMS cost optimization?** Batch, use push/email when possible
5. **Delivery receipt for push?** APNs/FCM async callbacks

### Common Mistakes
- **Single queue**: Can't prioritize
- **No rate limiting**: Abuse, cost explosion
- **No deduplication**: Duplicate notifications
- **Synchronous third-party call**: Blocks, use queue
- **Ignore preferences**: Legal issues (GDPR, CAN-SPAM)

### Key Points
1. **Queue per channel** for isolation and scaling
2. **Priority** via separate topics
3. **Rate limit** before enqueue
4. **Idempotency** for exactly-once semantics
5. **Retry + DLQ** for reliability
6. **User preferences** are critical (legal, UX)

---

## Appendix A: Extended Design Details

### A.1 Template Variable Validation
```python
# Required variables per template
template_vars = {"welcome_email": ["name", "login_link"]}
def validate(request):
    for var in template_vars[request.template_id]:
        if var not in request.parameters:
            raise ValidationError(f"Missing {var}")
```

### A.2 Rate Limit Algorithm (Sliding Window)
```
Redis: INCR rate_limit:user:123:push
       EXPIRE rate_limit:user:123:push 3600  # 1 hour
Check: GET rate_limit:user:123:push
If > 100: reject
```

### A.3 Quiet Hours Logic
```
if priority != "urgent":
    user_tz = get_user_timezone(user_id)
    current_hour = now_in_tz(user_tz).hour
    if quiet_start <= current_hour < quiet_end:
        delay_until(quiet_end)  # Schedule for later
```

### A.4 Circuit Breaker States
```
CLOSED (normal) -> OPEN (after 5 failures in 30s)
OPEN -> HALF_OPEN (after 60s cooldown)
HALF_OPEN -> CLOSED (on success) or OPEN (on failure)
```

### A.5 Batch Send Optimization
```
Email: SendGrid allows 1000/request - batch by template
SMS: No batch, 1/request - use connection pooling
Push: FCM allows 500/request - batch by priority
```

### A.6 Idempotency Key Storage
```
Cassandra: idempotency_keys (key, notification_id, created_at) TTL 24h
Lookup: SELECT notification_id WHERE key = ?
If exists: return existing notification_id
If not: proceed, insert key -> notification_id
```

### A.7 Sample Template
```
Subject: Welcome {{name}}!
Body: Hi {{name}}, click here to get started: {{link}}
Variables: [name, link]
Channel: email
```

### A.8 DLQ Processing
- **Manual review**: Dashboard for failed notifications
- **Retry**: Fix template/params, re-queue
- **Skip**: Mark as permanently failed
- **Alert**: Notify if DLQ depth > 10K

---

## Appendix B: Channel-Specific Details

### B.1 APNs Token Management
- **Invalid token**: FCM/APNs returns error, remove from DB
- **Token refresh**: Client sends new token on app launch
- **Batch**: Group by token for efficiency (500/batch for FCM)

### B.2 Email Bounce Handling
- **Webhook**: SendGrid/SES sends bounce event
- **Action**: Mark email as invalid, update user preference
- **Suppression list**: Don't retry bounced addresses
- **Compliance**: Honor unsubscribe immediately

### B.3 SMS Compliance
- **Opt-out**: "STOP" keyword -> immediate unsubscribe
- **Opt-in**: Must have explicit consent (TCPA)
- **Rate**: 1 msg/sec per number (carrier limit)
- **Cost**: ~$0.01-0.05 per SMS, optimize with push/email

### B.4 In-App Notification Delivery
- **WebSocket**: Push to connected clients
- **Polling fallback**: GET /notifications/unread (for offline clients)
- **Badge count**: Separate endpoint, cached

---

## Appendix C: Walkthrough Scenarios

### C.1 Scenario: Order Confirmation Email
1. Order Service creates order, calls POST /api/v1/notifications/send
2. Body: {user_id, channel: "email", template_id: "order_confirmation", parameters: {order_id, total}}
3. Orchestrator: Validates, checks user preferences (email enabled)
4. Rate limit: User at 45/50 emails today -> OK
5. Idempotency: No key provided, proceed
6. Renders template: "Your order #12345 for $99.99 is confirmed"
7. Writes to Cassandra (status=queued)
8. Publishes to Kafka topic email-normal
9. Returns notification_id to Order Service
10. Email worker consumes, calls SendGrid API
11. SendGrid accepts, returns 202
12. Worker updates status=sent
13. Later: SendGrid webhook reports delivered
14. Update status=delivered
15. Total: < 30 seconds

### C.2 Scenario: OTP SMS (Urgent)
1. Auth Service: User requests login, needs OTP
2. POST /notifications/send {channel: "sms", template_id: "otp", parameters: {code: "123456"}, priority: "urgent"}
3. Orchestrator: Validates, checks SMS enabled
4. Rate limit: 3/10 SMS today -> OK
5. Renders: "Your code is 123456. Valid for 10 minutes."
6. Publishes to Kafka topic sms-urgent (high priority)
7. SMS worker consumes immediately (urgent topic has dedicated consumers)
8. Calls Twilio API
9. Twilio delivers to carrier
10. Total: < 5 seconds

### C.3 Scenario: Duplicate Request (Idempotency)
1. Order Service sends same request twice (network retry)
2. First request: Processes, returns notification_id=n123
3. Second request: Same idempotency_key="order_456_confirm"
4. Orchestrator: Lookup key in Cassandra
5. Found: notification_id=n123
6. Return n123 without creating duplicate
7. User receives exactly one email

### C.4 Scenario: User Opted Out of Marketing
1. Marketing Service sends campaign email
2. Orchestrator fetches user preferences
3. user_preferences.email_categories = ["transactional"]
4. Request has category "marketing"
5. Reject: 403 "User has opted out of marketing emails"
6. No notification created, no queue
7. Return error to Marketing Service

---

## Appendix D: Technology Choices Rationale

### D.1 Why Kafka over SQS?
- **Ordering**: Per-partition ordering per user
- **Throughput**: 115K QPS, Kafka handles
- **Retention**: Replay for debugging/recovery
- **Multiple consumers**: Same topic, different consumer groups
- **Priority**: Separate topics for urgent vs normal

### D.2 Why Cassandra for Notifications?
- **Write-heavy**: 115K notifications/sec
- **Time-series**: Partition by user_id, cluster by time
- **Scale**: Add nodes, no downtime
- **Flexible**: Add status columns, idempotency_key
- **TTL**: Optional for auto-cleanup

### D.3 Why Separate Workers Per Channel?
- **Isolation**: Push outage doesn't affect email
- **Scale**: Scale push workers (70K QPS) independently
- **Specialization**: APNs vs SendGrid different expertise
- **Priority**: Urgent topics get more consumers
- **Deploy**: Deploy push worker without touching email

### D.4 Why Template Engine Server-Side?
- **Security**: Don't expose template logic to client
- **Consistency**: Same rendering for all channels
- **Validation**: Check variables before send
- **Caching**: Cache compiled templates

### D.5 Interview Discussion Points
- **Multi-channel**: Each channel has different latency, cost, reliability
- **User preferences**: Legal requirement (GDPR, TCPA, CAN-SPAM)
- **Rate limiting**: Protect from abuse, control cost (SMS expensive)

---

## Quick Reference Summary

| Component | Technology | Key Reason |
|-----------|------------|------------|
| Orchestrator | Stateless service | Validate, route |
| Queue | Kafka (per channel) | Scale, priority |
| Workers | Per channel | APNs, SendGrid, Twilio |
| Templates | MySQL + Redis | Versioned, cached |
| Preferences | Cassandra + Redis | Per user |
| Rate limit | Redis | Counters, TTL |
