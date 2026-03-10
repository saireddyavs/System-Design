# Notification System - Low Level Design

A production-quality, interview-ready notification system implementation in Go supporting multiple channels, templates, priority levels, rate limiting, and reliable delivery with retries.

## 1. Problem Description & Requirements

### Problem
Design a notification system that can deliver messages through multiple channels (Email, SMS, Push) while respecting user preferences, supporting templates, handling priorities, and ensuring reliable delivery.

### Requirements
| Requirement | Description |
|-------------|-------------|
| **Multi-channel** | Email, SMS, Push notification support |
| **User preferences** | Per-channel enable/disable per user |
| **Templates** | Variable substitution with `{{variable}}` placeholders |
| **Priority levels** | Low, Medium, High, Critical |
| **Reliable delivery** | Retry mechanism with exponential backoff (max 3 retries) |
| **Rate limiting** | Max 10 notifications per user per hour per channel |
| **History & tracking** | Notification status (Pending, Sent, Delivered, Failed, Retrying) |
| **Batch notifications** | Send multiple notifications concurrently |

### Business Rules
- **Rate limit**: Max 10 notifications per user per hour per channel
- **Retry**: Max 3 retries with exponential backoff
- **Critical bypass**: Critical priority notifications bypass rate limits
- **User preferences**: Respect channel enable/disable settings

---

## 2. Core Entities & Relationships

```
┌─────────────┐       ┌──────────────────┐       ┌─────────────┐
│    User     │──────<│   Notification   │>──────│  Template   │
└─────────────┘       └──────────────────┘       └─────────────┘
      │                         │
      │                         │ uses
      │                         ▼
      │                ┌──────────────────┐
      └───────────────│ NotificationSender│ (Email/SMS/Push)
                      └──────────────────┘
```

### Entity Details

**Notification**
- `ID`, `UserID`, `Channel`, `Type`, `Priority`, `Subject`, `Content`
- `Status` (Pending | Sent | Delivered | Failed | Retrying)
- `CreatedAt`, `SentAt`, `RetryCount`, `Error`

**User**
- `ID`, `Name`, `Email`, `Phone`, `DeviceToken`
- `Preferences`: map[Channel]bool (channel → enabled)

**Template**
- `ID`, `Name`, `Channel`, `Subject`, `Body`
- Body supports `{{variable}}` placeholders

---

## 3. Processing Pipeline

```
Request → Create Notification → Validate User/Channel → Rate Limit Check
    → Template Render (if template) → Send (via Strategy) → Update Status
    → Record (rate limiter) → Notify Observers
```

1. **Create**: Build notification from request (Builder), persist (Repository)
2. **Validate**: Check user exists, channel enabled in preferences
3. **Rate Limit**: Allow only if within limits (Critical bypasses)
4. **Render**: Substitute template variables if template ID provided
5. **Send**: Delegate to channel-specific sender (Strategy)
6. **Retry**: On failure, retry with exponential backoff (Decorator)
7. **Update**: Mark Sent/Failed, record for rate limiting
8. **Observe**: Publish events to observers

---

## 4. Design Patterns with WHY

| Pattern | Where | Why |
|---------|-------|-----|
| **Strategy** | EmailSender, SMSSender, PushSender implement `NotificationSender` | Swap delivery algorithms (channels) at runtime without changing client code. Adding a new channel = new strategy. |
| **Observer** | `NotificationObserver`, `Subscribe()`, `notifyObservers()` | Decouple notification lifecycle events from consumers. Analytics, logging, metrics can subscribe without modifying core logic. |
| **Template Method** | `processNotification()` in NotificationService | Defines the skeleton (validate → rate limit → send → update) while allowing steps to vary. Subclasses could override specific steps. |
| **Decorator** | `RetrySender`, `RateLimitedSender` wrap base senders | Add retry/rate-limiting behavior without modifying EmailSender, SMSSender. Composable, single-responsibility. |
| **Factory** | `createNotification()` | Centralizes notification creation logic. Ensures consistent ID generation, default values. |
| **Builder** | `SendRequestBuilder` | Fluent API for constructing complex SendRequest. Handles optional template vs raw content, variables. |
| **Chain of Responsibility** | Processing pipeline (validate → rate limit → send) | Each step can pass or reject. Could be extended with more handlers (e.g., deduplication, filtering). |

---

## 5. SOLID Principles Mapping

| Principle | Implementation |
|-----------|----------------|
| **S - Single Responsibility** | `TemplateService` (templates only), `PreferenceService` (preferences only), `NotificationService` (orchestration). Senders do one thing: send. |
| **O - Open/Closed** | New channels: add new `NotificationSender` implementation. New observers: implement `NotificationObserver`. No changes to existing code. |
| **L - Liskov Substitution** | Any `NotificationSender` can replace another. `RetrySender` wraps any sender and remains substitutable. |
| **I - Interface Segregation** | Small interfaces: `NotificationSender` (Send, Channel), `RateLimiter` (Allow, Record), `TemplateEngine` (Render). Clients depend only on what they need. |
| **D - Dependency Inversion** | Services depend on `interfaces.NotificationRepository`, `interfaces.NotificationSender`, etc. High-level modules don't depend on low-level (e.g., InMemoryNotificationRepository). |

---

## 6. Concurrency Model

### Worker Pool
- **Priority Queue**: `container/heap` for processing high-priority notifications first
- **Workers**: Configurable N goroutines poll the queue
- **Job**: `SendRequest` + `Priority`; higher priority processed first

### Thread Safety
- **Repositories**: `sync.RWMutex` on in-memory maps
- **Rate Limiter**: `sync.RWMutex` on per-user-per-channel records
- **User Preferences**: `sync.RWMutex` on `User.Preferences`
- **Observers**: `sync.RWMutex` on observer list; copy before iterating

### Async Delivery
- `SendBatch()`: Spawns goroutines per request, `sync.WaitGroup` for completion
- Worker pool: Jobs submitted via `Submit()`, processed asynchronously by workers

---

## 7. Interview Explanations

### 3-Minute Summary
"We built a notification system with Email, SMS, and Push. Each channel is a Strategy implementing `NotificationSender`. We use a Builder for requests and a Template Engine for `{{variable}}` substitution. User preferences and rate limits (10/hour/channel) are enforced; Critical bypasses limits. Retries use exponential backoff via a Decorator. Observers get lifecycle events. A worker pool with a priority queue processes notifications asynchronously. All components depend on interfaces for testability and extensibility."

### 10-Minute Deep Dive
1. **Architecture**: Clean architecture—models, interfaces, services, senders, repositories. Dependency injection in `main.go`.
2. **Strategy**: `EmailSender`, `SMSSender`, `PushSender` implement `NotificationSender`. Service looks up sender by channel. Adding Slack = new struct implementing the interface.
3. **Decorator**: `RetrySender` wraps any sender, adds 3 retries with exponential backoff. `RateLimitedSender` adds rate checks. Composable.
4. **Observer**: `Subscribe(observer)`. On status change, we iterate observers and call `OnNotificationEvent`. Used for logging, metrics.
5. **Concurrency**: Worker pool with `heap`-based priority queue. Batch send uses goroutines + WaitGroup. All shared state protected by mutexes.
6. **SOLID**: Each service has one job. Interfaces are small. Repositories are swappable. New behavior via new implementations, not edits.

---

## 8. Future Improvements

| Area | Improvement |
|------|-------------|
| **Persistence** | Replace in-memory repos with PostgreSQL/Redis |
| **Queue** | Use RabbitMQ/Kafka for durable, distributed job processing |
| **Real senders** | Integrate SendGrid (email), Twilio (SMS), FCM (push) |
| **Metrics** | Prometheus counters for sent/failed/rate-limited |
| **Config** | Rate limits, retry counts from config/env |
| **Deduplication** | Hash content+user+channel to avoid duplicate sends |
| **Dead letter** | Move failed-after-retries to DLQ for manual review |
| **Idempotency** | Idempotency keys for exactly-once delivery |

---

## Project Structure

```
11-notification-system/
├── cmd/main.go                 # Entry point, DI, worker pool demo
├── internal/
│   ├── models/                 # Notification, User, Template, enums
│   ├── interfaces/             # Sender, Repository, TemplateEngine, RateLimiter
│   ├── services/               # Notification, Template, Preference services
│   ├── senders/                # Email, SMS, Push (Strategy)
│   ├── middleware/             # Rate limiter, Retry sender (Decorator)
│   └── repositories/           # In-memory implementations
├── tests/                      # Unit tests
├── go.mod
└── README.md
```

## Run

```bash
go run ./cmd
go test ./...
```
