# Notification System — Interview Strategy (45 min)

## Time Allocation

| Phase | Time | What to Do |
|-------|------|------------|
| 1. Clarify & Scope | 3 min | Ask about channels, rate limits, retries, templates; scope out persistence, worker pool |
| 2. Core Models | 7 min | Notification, User (with Preferences map), Template; enums for Channel, Priority, Status |
| 3. Repository Interfaces | 5 min | NotificationRepository, UserRepository, TemplateRepository — Create, GetByID, Update |
| 4. Service Interfaces | 5 min | NotificationSender (Send, Channel), RateLimiter (Allow, Record), TemplateEngine (Render) |
| 5. Core Service Implementation | 12 min | NotificationService.processNotification — validate → rate limit → sender lookup → send |
| 6. main.go Wiring | 5 min | Wire senders map, RetrySender decorator, SlidingWindowRateLimiter, template service |
| 7. Extend & Discuss | 8 min | Add observer, batch send, discuss token bucket vs sliding window |

## Phase 1: Clarify & Scope (3 min)

**Questions to ask:**
- Which channels? (Email, SMS, Push — confirm these three for LLD)
- Rate limit? (e.g., 10 per user per hour per channel)
- Retries? (Max 3 with exponential backoff)
- Templates? (Yes — `{{variable}}` placeholders)
- Critical bypass? (Critical priority skips rate limit)

**Scope IN:** Multi-channel send, user preferences, templates, rate limiting, retry, single + batch send.

**Scope OUT:** Persistence (DB), worker pool / priority queue, real SMTP/Twilio integration.

## Phase 2: Core Models (7 min)

**Write FIRST:**
1. **Notification** — ID, UserID, Channel, Type, Priority, Subject, Content, Status, CreatedAt, RetryCount, Error
2. **User** — ID, Name, Email, Phone, DeviceToken, Preferences `map[Channel]bool`
3. **Template** — ID, Name, Channel, Subject, Body (with `{{var}}` placeholders)

**Enums:** Channel (Email, SMS, Push), Priority (Low, Medium, High, Critical), Status (Pending, Sent, Failed, Retrying).

**Skip:** TemplateRepository details; keep Template as a simple struct.

## Phase 3: Repository Interfaces (5 min)

```go
type NotificationRepository interface {
    Create(ctx, *Notification) error
    Update(ctx, *Notification) error
}
type UserRepository interface {
    GetByID(ctx, userID string) (*User, error)
}
type TemplateRepository interface {
    GetByID(ctx, id string) (*Template, error)
}
```

In-memory impl: `map[string]*Entity` with `sync.RWMutex`. Don't implement GetAll or Search.

## Phase 4: Service Interfaces (5 min)

```go
type NotificationSender interface {
    Send(ctx, *Notification, *User) error
    Channel() Channel
}
type RateLimiter interface {
    Allow(ctx, userID string, channel Channel, priority Priority) bool
    Record(ctx, userID string, channel Channel) error
}
type TemplateEngine interface {
    Render(template *Template, variables map[string]string) (subject, body string, err error)
}
```

Mention: EmailSender, SMSSender, PushSender implement NotificationSender. RetrySender wraps any sender.

## Phase 5: Core Service Implementation (12 min)

**THE most important method:** `NotificationService.processNotification(ctx, notification)`

1. Get user from UserRepo → fail if not found
2. Check `user.IsChannelEnabled(notification.Channel)` → fail if disabled
3. Lookup sender: `sender, ok := s.senders[notification.Channel]` → fail if not found
4. Rate limit: if not Critical, `s.rateLimiter.Allow(...)` → fail if exceeded
5. Call `sender.Send(ctx, notification, user)`
6. On success: mark Sent, `rateLimiter.Record`, notify observers
7. On failure: increment RetryCount; if CanRetry, mark Retrying; else mark Failed

**Why this method:** It's the heart of the pipeline. Get this right and the rest (createNotification, SendBatch) is straightforward.

**Template rendering:** If request has TemplateID, call TemplateService.Render before creating notification. Use regex `\{\{(\w+)\}\}` and ReplaceAllStringFunc.

## Phase 6: main.go Wiring (5 min)

1. Create repos (in-memory)
2. Create senders: `map[Channel]NotificationSender{ Email: RetrySender(EmailSender), ... }`
3. Create SlidingWindowRateLimiter(10, time.Hour)
4. Create TemplateService(repo, DefaultTemplateEngine)
5. Create NotificationService(repo, userRepo, senders, templateSvc, rateLimiter)
6. Demo: `SendNotification(ctx, SendRequestBuilder(...).Build())`

## Phase 7: Extend & Discuss (8 min)

- **Observer:** Add `Subscribe(NotificationObserver)`; on status change, iterate observers and call `OnNotificationEvent`. Use case: logging, metrics.
- **Batch:** `SendBatch` — spawn goroutine per request, WaitGroup, collect errors.
- **Rate limiter tradeoff:** Sliding window vs token bucket — sliding window more accurate; token bucket simpler. For 10/hour, sliding window prunes timestamps outside window.
- **Retry:** RetrySender wraps inner sender; exponential backoff: `backoff *= 2` each attempt.

## Tips

- Start with `processNotification` flow on paper before coding.
- The channel dispatch map (`senders[channel]`) is the key DS — O(1) lookup.
- Critical bypass: check `!notification.IsCritical()` before rate limit.
- Keep TemplateEngine simple: one regex, one ReplaceAllStringFunc.
- If time runs out, skip worker pool and batch; focus on single send + rate limit + retry.
