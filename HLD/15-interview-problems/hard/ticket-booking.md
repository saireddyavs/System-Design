# System Design: Ticket Booking System (BookMyShow / Ticketmaster)

## 1. Problem Statement & Requirements

### Problem Statement
Design a ticket booking platform that allows users to browse events (movies, concerts, sports), view seat maps, select seats, and complete bookings with payment. The system must handle high concurrency during flash sales (e.g., millions of users trying to book simultaneously for a popular concert) and prevent double booking.

### Functional Requirements
- **Browse Events**: List events by category, date, venue, location
- **View Seat Map**: Interactive seat map showing availability (available, sold, held)
- **Select Seats**: Choose seats; see real-time availability
- **Book Tickets**: Reserve seats temporarily → pay → confirm or release
- **Payment**: Integrate with payment gateways (card, UPI, wallet)
- **Concurrent Booking Prevention**: No two users get same seat
- **Waiting Queue**: For high-demand events, queue users before allowing booking
- **Notifications**: Booking confirmation (email, SMS, push)

### Non-Functional Requirements
- **Scale**: 10M+ DAU; flash sales with millions of concurrent booking attempts
- **Consistency**: Strong consistency for seat inventory (no overbooking)
- **Latency**: < 200ms for seat availability; < 2s for booking confirmation
- **Availability**: 99.9% during normal load; graceful degradation during flash sales

### Security Requirements
- **Authentication**: JWT; session management
- **Payment**: PCI compliance via gateway; never store raw card data
- **Fraud**: Rate limit; CAPTCHA on high-risk actions; 3D Secure
- **Audit**: Log all booking/payment events for dispute resolution

### Out of Scope
- Recommendation engine
- Loyalty/rewards program
- Refund/cancellation flow (simplified)
- Multi-venue event coordination

---

## 2. Back-of-Envelope Estimation

### Capacity Planning Summary

Ticket booking has two modes: steady state (browse, occasional booking) and flash sale (spike to millions of concurrent users). Design for both.

### Assumptions
- 10M DAU
- 5% book tickets daily = 500K bookings/day
- Flash sale: 5M users try to book in 10 min for 50K seats
- Average 3 seats per booking
- 100K events active at any time

### QPS Estimates
| Operation | Normal | Flash Sale (peak) |
|-----------|--------|-------------------|
| Browse events | 50,000 | 200,000 |
| Seat map / availability | 20,000 | 500,000 |
| Seat selection / hold | 5,000 | 100,000 |
| Payment / confirm | 2,000 | 20,000 |

### Storage Estimates
- **Events**: 100K × 2KB ≈ 200 MB
- **Venues/Seat maps**: 10K venues × 50KB ≈ 500 MB
- **Bookings**: 500K/day × 365 × 1KB ≈ 180 GB/year
- **Inventory state**: 100K events × 5K seats × 4 bytes ≈ 2 GB (in-memory)

### Bandwidth
- **Seat map**: 50KB per load × 500K QPS = 25 GB/s (flash sale)
- **API**: ~10 GB/s peak

### Cache
- **Event catalog**: 90% cache hit; Redis
- **Seat availability count**: Per show; cache with short TTL (5-10s)
- **Seat map**: Cache static layout; invalidate on booking

### Flash Sale Spike

- Normal: 5K hold QPS. Flash sale: 100K hold QPS. 20× spike.
- Mitigation: Queue absorbs load; only 1-2K actually reach hold endpoint per second.
- Pre-scale: Add 5× capacity before sale; scale down after.

---

## 3. API Design

### REST Endpoints

```
# Catalog
GET    /api/v1/events                    # List events (filters: city, date, category)
GET    /api/v1/events/{id}               # Event details
GET    /api/v1/events/{id}/shows         # Shows (date/time) for event
GET    /api/v1/venues/{id}               # Venue details
GET    /api/v1/shows/{id}/seat-map       # Seat map layout + availability

# Inventory & Booking
GET    /api/v1/shows/{id}/availability   # Available seat count (fast)
POST   /api/v1/shows/{id}/hold           # Hold seats (returns hold_id, TTL)
POST   /api/v1/holds/{hold_id}/confirm   # Confirm with payment
POST   /api/v1/holds/{hold_id}/release   # Release hold
GET    /api/v1/bookings/{id}             # Booking details

# Queue (Flash Sales)
POST   /api/v1/shows/{id}/queue/join    # Join waiting queue
GET    /api/v1/shows/{id}/queue/status  # Position in queue
GET    /api/v1/shows/{id}/queue/entry   # Get token when allowed to book (long poll)

# User
GET    /api/v1/users/me/bookings        # My bookings
POST   /api/v1/auth/login               # Login
```

### Key Request/Response Examples

**Hold seats**:
```json
POST /api/v1/shows/show_123/hold
{
  "seat_ids": ["A1", "A2", "A3"],
  "user_id": "user_456"
}
Response: {
  "hold_id": "hold_789",
  "expires_at": "2024-03-10T14:35:00Z",
  "seats": [{"id": "A1", "row": "A", "number": 1}, ...]
}
```

**Confirm booking**:
```json
POST /api/v1/holds/hold_789/confirm
{
  "payment_id": "pay_xyz",
  "payment_method": "card"
}
Response: {
  "booking_id": "book_101",
  "status": "confirmed",
  "tickets": [...]
}
```

---

## 4. Data Model / Database Schema

### Database Choice
- **Catalog**: PostgreSQL — events, venues, shows, seat maps (relational)
- **Inventory**: Redis + PostgreSQL — Redis for fast lock/availability; PostgreSQL for durability
- **Bookings**: PostgreSQL — transactional
- **Queue**: Redis (sorted set or list) — waiting queue

### Schema

**events**
| Column | Type | Description |
|--------|------|-------------|
| event_id | UUID PK | |
| title | VARCHAR | |
| category | ENUM | movie, concert, sports |
| description | TEXT | |
| image_url | VARCHAR | |
| duration_minutes | INT | |
| created_at | TIMESTAMP | |

**venues**
| Column | Type | Description |
|--------|------|-------------|
| venue_id | UUID PK | |
| name | VARCHAR | |
| city | VARCHAR | |
| address | TEXT | |
| capacity | INT | |
| created_at | TIMESTAMP | |

**shows**
| Column | Type | Description |
|--------|------|-------------|
| show_id | UUID PK | |
| event_id | UUID FK | |
| venue_id | UUID FK | |
| start_time | TIMESTAMP | |
| end_time | TIMESTAMP | |
| total_seats | INT | |
| created_at | TIMESTAMP | |

**seats**
| Column | Type | Description |
|--------|------|-------------|
| seat_id | UUID PK | |
| show_id | UUID FK | |
| row_label | VARCHAR | A, B, C |
| seat_number | INT | |
| section | VARCHAR | VIP, Standard |
| x, y | FLOAT | Position on seat map |
| version | INT | For optimistic locking |

**holds**
| Column | Type | Description |
|--------|------|-------------|
| hold_id | UUID PK | |
| show_id | UUID FK | |
| user_id | UUID FK | |
| seat_ids | JSONB | ["A1", "A2"] |
| expires_at | TIMESTAMP | |
| status | ENUM | active, confirmed, expired |
| created_at | TIMESTAMP | |

**bookings**
| Column | Type | Description |
|--------|------|-------------|
| booking_id | UUID PK | |
| hold_id | UUID FK | |
| user_id | UUID FK | |
| show_id | UUID FK | |
| seat_ids | JSONB | |
| amount | DECIMAL | |
| payment_id | VARCHAR | |
| status | ENUM | confirmed, cancelled, refunded |
| created_at | TIMESTAMP | |

### Indexes

- `shows(event_id, start_time)` — List shows for event
- `seats(show_id)` — Seat map for show
- `holds(show_id, status)` — Active holds per show
- `bookings(user_id, created_at)` — User's booking history
- `bookings(show_id)` — Bookings per show (for inventory reconciliation)

### Redis Keys (Inventory)

- `hold:{show_id}:{seat_id}` — Set to user_id or hold_id; TTL 300s
- `available_count:{show_id}` — Cache of available seats; invalidate on hold/confirm
- `queue:{show_id}` — Sorted set for waiting queue

---

## 5. High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              CLIENTS (Web, Mobile)                                │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│  CDN (static: seat map assets, images)  │  API GATEWAY / LOAD BALANCER           │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
          ┌─────────────────────────────┼─────────────────────────────┐
          ▼                             ▼                             ▼
┌─────────────────┐           ┌─────────────────┐           ┌─────────────────┐
│ Catalog Service │           │ Inventory Svc   │           │ Booking Service  │
│ (events, venues,│           │ (seat hold,     │           │ (confirm,        │
│  shows, seat    │           │  release,       │           │  payment)        │
│  map layout)    │           │  availability)  │           │                  │
└────────┬────────┘           └────────┬────────┘           └────────┬────────┘
         │                             │                             │
         ▼                             ▼                             ▼
┌─────────────────┐           ┌─────────────────┐           ┌─────────────────┐
│ PostgreSQL      │           │ Redis           │           │ PostgreSQL      │
│ (read replica   │           │ (locks,         │           │ (bookings)      │
│  for reads)     │           │  availability)  │           │                  │
└─────────────────┘           └─────────────────┘           └─────────────────┘
         │                             │                             │
         │                     ┌───────┴───────┐                     │
         │                     ▼               ▼                     │
         │             ┌──────────────┐ ┌──────────────┐             │
         │             │ Queue Service│ │ Payment      │             │
         │             │ (flash sale)│ │ Gateway      │             │
         │             └──────────────┘ └──────────────┘             │
         │                     │               │                     │
         └─────────────────────┴───────────────┴─────────────────────┘
                                        │
                                        ▼
                              ┌─────────────────┐
                              │ Notification    │
                              │ (email, SMS)    │
                              └─────────────────┘

                    BOOKING FLOW (ASCII)
┌──────────┐   1. Browse    ┌──────────┐   2. View map   ┌──────────┐
│  User   │ ──────────────► │ Catalog  │ ──────────────► │ Seat Map │
└──────────┘                └──────────┘                └────┬─────┘
                                                             │ 3. Select seats
                                                             ▼
┌──────────┐   4. Hold      ┌──────────┐   5. Pay       ┌──────────┐
│ Booking │ ◄────────────── │Inventory │ ◄────────────── │ Payment  │
│ Confirm │ ──────────────► │ (Redis   │ ──────────────► │ Gateway  │
└──────────┘   7. Confirm   │  lock)   │   6. Callback  └──────────┘
                            └──────────┘
```

---

## 6. Detailed Component Design

### 6.1 Catalog Service

**Responsibilities**:
- Serve event list, filters (city, date, category)
- Event details, venue details
- Show list for event (date/time slots)
- Seat map layout (static: rows, sections, coordinates)

**Seat Map**:
- **Layout**: JSON defining rows, seats, sections, (x,y) for rendering
- **Availability**: Fetched from Inventory Service (available, held, sold)
- **Caching**: Layout cached (rarely changes); availability has short TTL

### 6.2 Inventory Service & Concurrency

**Core challenge**: Prevent double booking when 100K users select same seat.

**Approach 1: Pessimistic Locking (SELECT FOR UPDATE)**
```sql
BEGIN;
SELECT * FROM seats WHERE show_id = ? AND seat_id IN (?) FOR UPDATE;
-- Check if available, then update status
UPDATE seats SET status = 'held', hold_id = ? WHERE ...;
COMMIT;
```
- Blocks concurrent access; can cause lock contention
- Good for moderate concurrency

**Approach 2: Optimistic Locking**
```sql
UPDATE seats SET status = 'held', version = version + 1
WHERE show_id = ? AND seat_id IN (?) AND status = 'available' AND version = ?;
-- If rows_affected < N, conflict; retry or fail
```
- No lock; may need retries
- Good when conflict rate is low

**Approach 3: Distributed Lock (Redis)**
- Lock key: `show:{show_id}:seat:{seat_id}`
- `SET lock_key user_id NX EX 300` (5 min TTL)
- If success, hold seat; else seat taken
- Release on confirm or timeout
- **Best for flash sales**: Fast, no DB lock contention

**Approach 4: Reservation Token**
- User gets "token" to attempt booking (from queue)
- Token grants exclusive access to inventory for 5 min
- Reduces load on inventory service

### 6.3 Hold → Confirm Flow

1. **Hold**: User selects seats; backend tries to lock seats (Redis or DB)
   - On success: Create hold record, TTL 5-10 min
   - On failure: Return "seats no longer available"
2. **Payment**: User redirected to payment gateway
3. **Confirm**: Payment callback hits backend; verify hold valid; create booking; release lock
4. **Release**: If user abandons or timeout, release lock; seats back to available

**Idempotency**: Payment callback may retry; use idempotency key to avoid double booking.

### 6.4 Virtual Waiting Queue (Flash Sales)

**Problem**: 5M users hit booking page; only 50K seats. Server overload, poor UX.

**Solution: Queue-based admission**

1. **Join Queue**: User requests to join queue for show; gets queue token/position
2. **Wait**: User polls or long-polls for "your turn"
3. **Entry**: When slot available, user gets booking token (valid 5 min)
4. **Book**: User with token can attempt hold; others get "queue first"

**Implementation**:
- **Redis Sorted Set**: Score = timestamp; value = user_id. FIFO.
- **Rate limit**: Allow N users per second to enter booking (e.g., 1000/s)
- **Fairness**: First-come-first-served by join time

**Pre-warming**:
- Cache event details, seat map before sale
- Pre-scale servers
- Connection pooling

### 6.5 Payment Integration

- **Redirect flow**: User → Payment page → Callback to backend
- **Webhook**: Payment gateway sends async confirmation
- **Idempotency**: Store payment_id; reject duplicate callbacks
- **Timeout**: Hold expires before payment? Release seats; inform user

### 6.6 Seat Map Rendering

- **Client**: SVG or Canvas; fetch layout (JSON) + availability (API)
- **Availability**: Batch fetch for visible section; or WebSocket for real-time
- **Caching**: Layout in CDN; availability in Redis (5-10s TTL)

### 6.7 Caching Strategy

| Data | Cache | TTL | Invalidation |
|------|-------|-----|---------------|
| Event list | Redis/CDN | 5 min | On event update |
| Event details | Redis | 10 min | On update |
| Seat map layout | CDN | 1 day | Rarely |
| Available count | Redis | 5 s | On hold/confirm |
| Seat availability | Redis | 5 s | On hold/confirm |

### 6.8 Seat Selection UX & Real-Time Updates

**Problem**: User A and User B both see seat A1 as available; both try to book. Need real-time feedback.

**Approaches**:
- **Optimistic UI**: Show selection immediately; on hold failure, revert and show "no longer available"
- **Pessimistic**: Lock seat on "hover" or "click" for 30s; prevents others from selecting (complex, can abuse)
- **Refresh**: Poll availability every 5s; stale but simple
- **WebSocket**: Push availability updates; more complex, better UX
- **Hybrid**: Show count as "X seats left" (cached); on hold, validate against source of truth

**Seat map data**: Return `{ seat_id, row, number, section, status, price }`. Status: available, held, sold. For held, optionally show "held by you" or "held by another" (privacy).

### 6.9 Refund & Cancellation Flow

**Cancellation**:
1. User requests cancel within policy (e.g., 24h before show)
2. Backend: Mark booking cancelled; release seats; trigger refund
3. Payment gateway: Refund API (async)
4. Notify user when refund processed

**Challenges**: Partial cancellation (3 of 5 seats); refund amount (fees?); inventory back to pool atomically.

### 6.10 Notification Service

**Channels**: Email, SMS, push (mobile).

**Events**: Booking confirmation (immediate), reminder (24h before), cancellation/refund, queue entry ("your turn").

**Implementation**: Event-driven; booking service publishes to message queue; notification workers consume and send via providers (SendGrid, Twilio, FCM).

### 6.11 Anti-Abuse & Fraud

**Bot prevention**: CAPTCHA on hold/queue join; rate limit per IP/user; behavioral analysis.

**Fraud**: Use gateway's fraud detection; 3D Secure for cards; store evidence for chargebacks.

**Hold abuse**: Limit holds per user per show (e.g., 2 active); require login for hold.

---

## 7. Scaling

### Sharding

- **Events/Shows**: Shard by show_id or event_id
- **Bookings**: Shard by user_id or show_id
- **Inventory**: Redis cluster; key = show_id + seat_id

### Caching

- **Read-heavy catalog**: Multiple read replicas; Redis for hot data
- **Inventory**: Redis primary; avoid DB for availability checks when possible

### Rate Limiting

- **Queue join**: Limit per IP/user
- **Hold**: Only users with queue token (during flash sale)
- **API**: Global rate limit to protect backend

### CDN

- Static assets: seat map images, event images
- Reduce load on origin

### Database Optimization

- **Indexes**: `shows(venue_id, start_time)`, `holds(show_id, status)`, `bookings(user_id, created_at)`
- **Partitioning**: Bookings table by `created_at` (monthly) for faster range queries
- **Read replicas**: Catalog and booking history on replicas; inventory and hold on primary

### Queue Implementation Details

**Redis Sorted Set**:
```
ZADD queue:show_123 {timestamp} user_456
ZRANGE queue:show_123 0 0  # Get first in line
ZREM queue:show_123 user_456  # Remove when entered
```

**Token bucket**: Allow 1000 users/min to enter booking. `INCR` counter per minute; if < 1000, allow.

---

## 8. Failure Handling

### Component Failures

| Component | Failure | Mitigation |
|-----------|---------|------------|
| Inventory (Redis) | Down | Fallback to DB with locking; slower but works |
| Payment gateway | Timeout | Retry; extend hold if possible; queue for async |
| DB | Primary down | Failover to replica; possible brief inconsistency |
| Queue | Redis down | Degrade to no queue; allow all to attempt (may overload) |

### Double Booking Prevention

- **Idempotency**: Payment callback idempotency key
- **Atomicity**: Hold + confirm in transaction where possible
- **Reconciliation**: Nightly job to detect inconsistencies (hold expired but not released)

### Graceful Degradation (Flash Sale)

- **Queue full**: Reject new joins; show "try again later"
- **Inventory overload**: Return 503; retry with backoff
- **Payment slow**: Extend hold; or release and notify

### Hold Expiry Background Job

- **Cron**: Every 1 min, find holds where `expires_at < NOW()` and `status = 'active'`
- **Action**: Update status to 'expired'; release Redis locks; decrement ref_counts if any
- **Idempotent**: Safe to run multiple times; check status before acting

### Payment Webhook Retry

- Payment gateway retries webhook 3-5 times over 24h on failure
- Backend must be idempotent: `payment_id` unique; reject duplicate
- If confirm fails (DB down): Return 5xx; gateway retries. Ensure hold TTL long enough.

---

## 9. Monitoring & Observability

### Key Metrics

| Metric | Description | Alert |
|--------|-------------|-------|
| hold_success_rate | % of hold requests that succeed | < 95% |
| hold_latency_p99 | Time to hold seats | > 500ms |
| payment_callback_latency | Time to confirm after payment | > 2s |
| double_booking_count | Detected overbooks | > 0 |
| queue_wait_time_p99 | Time in queue before entry | Track |
| cache_hit_rate | Catalog cache | < 90% |

### Logging

- Hold: user_id, show_id, seat_ids, result
- Confirm: booking_id, payment_id, amount
- Queue: join, entry, abandon

### Alerts

- Inventory service errors
- Payment callback failures
- Redis/DB latency spike

### SLOs

- **Hold success rate**: 95% (during normal load); 80% during flash sale
- **Confirm latency**: p99 < 2s from payment to confirmation
- **Double booking**: 0 (critical; alert on any detected)

### Dashboards

- **Operational**: Hold QPS, confirm QPS, queue length, error rates
- **Business**: Bookings per hour, revenue, conversion (hold → confirm)
- **Flash sale**: Queue join rate, entry rate, hold success, time to sellout

---

## 10. Interview Tips

### Follow-up Questions

1. **How do you handle a user holding seats and never paying?** TTL on hold (5-10 min); background job releases expired holds.
2. **What if payment succeeds but confirm fails?** Idempotent retry; or manual reconciliation; refund if needed.
3. **How do you scale the seat map for 50K seats?** Paginate or load visible section; aggregate availability by section for count.
4. **How does the queue ensure fairness?** FIFO by join timestamp; prevent gaming (e.g., rate limit joins per user).
5. **How do you handle partial failures (2 of 3 seats available)?** Atomic hold: all or nothing. Or offer "best available" subset.

### Common Mistakes

1. **No hold mechanism**: Direct book causes race; need temporary reservation.
2. **DB lock for all availability checks**: Too slow; use Redis for hot path.
3. **No queue for flash sales**: Server overload; need admission control.
4. **Ignoring idempotency**: Payment retry can double-book.
5. **Cache invalidation**: Stale availability; short TTL and invalidate on write.

### What to Emphasize

- **Hold → Confirm flow**: Critical for preventing double booking
- **Distributed lock (Redis)**: Fast, scalable for inventory
- **Queue for flash sales**: Protects backend, fair UX
- **Idempotency**: For payment and critical operations
- **Caching with invalidation**: Balance freshness and load

### Sample Discussion Flow

1. **Clarify**: "Flash sale or normal? How many seats?" — Both; flash sale is key differentiator.
2. **Core flow**: Browse → Select seats → Hold → Pay → Confirm. Emphasize hold as buffer.
3. **Concurrency**: "Two users, same seat?" → Lock (Redis or DB); first wins.
4. **Flash sale**: "5M users, 50K seats?" → Queue; rate limit entry to booking.
5. **Payment**: Idempotency; webhook retry; hold timeout.

### Time-Boxed Approach (45 min interview)

- **0-5 min**: Clarify flash sale vs normal, scale, consistency
- **5-15 min**: High-level (catalog, inventory, booking, payment)
- **15-25 min**: Hold flow, concurrency (Redis lock), queue for flash sale
- **25-35 min**: Scaling, caching, failure handling
- **35-45 min**: Double booking prevention, idempotency, follow-ups

### Additional Deep-Dive Topics

**Seat pricing**:
- Different sections (VIP, Standard) have different prices. Store in `seats` or `show_sections` table.
- Dynamic pricing: Adjust by demand. Update price in cache; invalidate on change.

**Partial hold**:
- User wants A1, A2, A3; only A1, A2 available. Options: (1) All or nothing; (2) Offer "best available" (A1, A2 + suggest A4).
- Atomic lock: Lock all or none. Simpler; better UX to suggest alternatives.

**Multi-show events**:
- Concert has multiple dates. Each date = separate show. User picks date → show → seats.
- Reuse venue/seat layout across shows; only availability differs.

**Audit trail**:
- Log all hold/confirm/release events for debugging and fraud analysis. Append-only log or events table.

### Design Alternatives Considered

**Optimistic vs pessimistic locking**: Optimistic (version check) less contention, may need retries. Pessimistic (SELECT FOR UPDATE) blocks. For flash sale, Redis lock is faster than both.

**Queue vs no queue**: No queue: all users hit hold; server overload, poor UX. Queue: fair, protects backend. Queue essential for flash sale.

**Hold on hover vs hold on confirm click**: Hold on hover blocks seat for 30s; prevents others. Can be abused (hover many seats). Hold on confirm simpler; race possible but handled by lock.

**Synchronous vs async payment**: Sync: wait for gateway response before confirm. Async: webhook. Async scales better; need idempotency.

### Phased Rollout

**Phase 1 (MVP)**: Catalog, seat map, hold with DB lock, confirm with sync payment. No queue. Single region.

**Phase 2**: Redis for inventory lock. Async payment webhook. Idempotency. Queue for high-demand events.

**Phase 3**: Flash sale queue. Pre-warming. Rate limiting. CDN. Multi-region.

**Phase 4**: Dynamic pricing. Fraud detection. Refund automation. 10M+ DAU.

### Quick Reference Card

| Concept | Key Point |
|---------|-----------|
| Flow | Browse → Select → Hold → Pay → Confirm |
| Concurrency | Redis lock or DB SELECT FOR UPDATE |
| Hold TTL | 5-10 min; release on timeout |
| Flash sale | Queue → rate limit entry → hold |
| Idempotency | Payment callback; reject duplicate |
| Double book | Lock before hold; atomic confirm |
| Cache | Event 10min; availability 5s |

### Glossary

- **Hold**: Temporary reservation; TTL; released if not confirmed
- **Flash sale**: High-demand event; queue required
- **Idempotency**: Same request twice = same result; critical for payment
- **Optimistic lock**: Version check; retry on conflict
- **Pessimistic lock**: SELECT FOR UPDATE; blocks others
