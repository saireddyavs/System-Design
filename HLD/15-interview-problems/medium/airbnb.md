# Design Airbnb

## 1. Problem Statement & Requirements

### Problem Statement
Design a vacation rental marketplace like Airbnb that enables guests to search for accommodations with geo-filtering, hosts to manage listings and calendars, a booking system with availability checks and payments, and a review system for trust and host/guest matching.

### Functional Requirements
- **Search with geo-filtering**: Search by location (map view, radius), dates, guests, price range, amenities
- **Listing management**: Create, edit, delete listings; photos, descriptions, amenities, pricing rules
- **Booking system**: Check availability → reserve (hold) → payment → confirm; prevent double booking
- **Calendar management**: Host sets availability, pricing per date, minimum stay, blocked dates
- **Pricing engine**: Base price, seasonal pricing, cleaning fee, service fee, taxes
- **Payments**: Guest pays; host receives payout; platform fee; refunds, disputes
- **Review system**: Guest reviews host/listing; host reviews guest; ratings, photos
- **Host/guest matching**: Messaging, identity verification, trust signals

### Non-Functional Requirements
- **Scale**: 7M listings, 150M users
- **Latency**: Search < 200ms, booking flow < 2s per step
- **Availability**: 99.99% for booking and payment
- **Consistency**: Strong consistency for availability and double-booking prevention

### Out of Scope
- Host onboarding/verification workflows
- Insurance/guarantee programs
- Experiences (tours, activities)
- Multi-currency conversion (assume USD)
- Regulatory compliance (tax collection per jurisdiction)

---

## 2. Back-of-Envelope Estimation

### Scale Assumptions
- **Listings**: 7M
- **Users**: 150M (guests + hosts)
- **Bookings**: ~500M/year ≈ 1.4M/day
- **Search queries**: 10x bookings ≈ 14M/day
- **Peak**: 3x average ≈ 4.2M searches/day, 4.2M bookings/day

### QPS Calculation
| Operation | Daily Volume | Peak QPS |
|-----------|--------------|----------|
| Search (geo + filters) | 14M | ~500 |
| Listing detail view | 50M | ~1,700 |
| Availability check | 20M | ~700 |
| Reserve (hold) | 2M | ~70 |
| Payment/confirm | 1.4M | ~50 |
| Calendar updates | 5M | ~170 |
| Reviews read | 30M | ~1,000 |
| Reviews write | 1M | ~35 |

### Storage (5 years)
- **Listings**: 7M × 5KB ≈ 35 GB
- **Photos**: 7M × 20 photos × 200KB ≈ 28 TB
- **Bookings**: 2.5B × 1KB ≈ 2.5 TB
- **Calendar/availability**: 7M × 365 × 50B ≈ 128 TB (sparse)
- **Reviews**: 500M × 500B ≈ 250 GB
- **Users**: 150M × 2KB ≈ 300 GB

### Bandwidth
- **Search**: 500 QPS × 50KB ≈ 25 MB/s
- **Listing images**: CDN-heavy; origin ~100 MB/s
- **API**: ~500 MB/s peak

### Cache
- **Search results**: Elasticsearch + Redis; 7M listings indexed
- **Listing detail**: Redis; 7M × 5KB ≈ 35 GB
- **Availability**: Redis per listing-date; hot listings ~10% = 700K × 365 × 8B ≈ 2 TB (with TTL)

---

## 3. API Design

### REST Endpoints

```
# Search & Discovery
GET    /api/v1/search
Query: lat, lng, check_in, check_out, guests, min_price, max_price, amenities[], room_type
Response: { "listings": [...], "total": N }  # HTTP streaming for large results

GET    /api/v1/listings/:listing_id
Response: { listing details, host, amenities, photos, reviews_summary }

GET    /api/v1/listings/:listing_id/availability
Query: start_date, end_date
Response: { "available_dates": [...], "pricing": {...} }

# Listing Management (Host)
POST   /api/v1/listings
PUT    /api/v1/listings/:listing_id
DELETE /api/v1/listings/:listing_id
PUT    /api/v1/listings/:listing_id/calendar
Body: { "dates": [{ "date": "...", "available": true, "price": 150 }] }

# Booking Flow
POST   /api/v1/bookings/check
Body: { listing_id, check_in, check_out, guests }
Response: { "available": true, "total_price": 450, "breakdown": {...} }

POST   /api/v1/bookings/reserve
Body: { listing_id, check_in, check_out, guests }
Response: { "reservation_id", "expires_at" }  # Hold for 10 min

POST   /api/v1/bookings/:reservation_id/confirm
Body: { "payment_method_id": "..." }
Response: { "booking_id", "status": "CONFIRMED" }

POST   /api/v1/bookings/:reservation_id/cancel

# Payments
POST   /api/v1/payments
Body: { reservation_id, payment_method_id }
Response: { "payment_id", "status" }

# Reviews
GET    /api/v1/listings/:listing_id/reviews
POST   /api/v1/reviews
Body: { booking_id, rating, comment, photos[] }

# Messaging
GET    /api/v1/conversations
POST   /api/v1/conversations
POST   /api/v1/conversations/:id/messages
```

### WebSocket (Optional)
```
WS /ws/conversations/:id   # Real-time messaging
```

### HTTP Streaming (Airbnb Adoption - $84M Savings)
- **Use case**: Search returns 1000s of listings; streaming allows client to render incrementally
- **Implementation**: `Transfer-Encoding: chunked`; server sends JSON array chunks as they're ready
- **Benefit**: Reduces time-to-first-result, lowers server memory, improves perceived latency

---

## 4. Data Model / Database Schema

### Database Choice
- **Listings, Users**: PostgreSQL (ACID, complex queries)
- **Search**: Elasticsearch (full-text, geo)
- **Geo queries**: PostGIS extension in PostgreSQL
- **Availability/Calendar**: PostgreSQL + Redis (hot data)
- **Bookings**: PostgreSQL (strong consistency)
- **Reviews**: PostgreSQL
- **Messaging**: Cassandra (write-heavy, partition by conversation)

### Schema

**Listings (PostgreSQL + PostGIS)**
```sql
listings (
  listing_id BIGSERIAL PRIMARY KEY,
  host_id BIGINT REFERENCES users(user_id),
  title VARCHAR(200),
  description TEXT,
  property_type VARCHAR(50),
  room_type VARCHAR(50),
  accommodates INT,
  bedrooms INT,
  bathrooms DECIMAL(3,1),
  amenities JSONB,
  location GEOGRAPHY(POINT, 4326),  -- PostGIS
  address VARCHAR(500),
  city VARCHAR(100),
  country VARCHAR(100),
  base_price DECIMAL(10,2),
  cleaning_fee DECIMAL(10,2),
  min_nights INT,
  max_nights INT,
  status VARCHAR(20),
  created_at TIMESTAMP,
  updated_at TIMESTAMP
)

CREATE INDEX idx_listings_location ON listings USING GIST(location);
CREATE INDEX idx_listings_host ON listings(host_id);
```

**Calendar / Availability**
```sql
calendar_availability (
  listing_id BIGINT,
  date DATE,
  available BOOLEAN,
  price DECIMAL(10,2),
  min_nights INT,
  PRIMARY KEY (listing_id, date)
)
```

**Bookings**
```sql
bookings (
  booking_id BIGSERIAL PRIMARY KEY,
  listing_id BIGINT,
  guest_id BIGINT,
  check_in DATE,
  check_out DATE,
  guests INT,
  status VARCHAR(20),  -- PENDING, RESERVED, CONFIRMED, CANCELLED, COMPLETED
  total_price DECIMAL(10,2),
  reservation_expires_at TIMESTAMP,  -- For hold
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  UNIQUE (listing_id, check_in, check_out)  -- Partial; need application-level lock
)

reservations_hold (
  reservation_id UUID PRIMARY KEY,
  listing_id BIGINT,
  check_in DATE,
  check_out DATE,
  guest_id BIGINT,
  expires_at TIMESTAMP,
  created_at TIMESTAMP
)
```

**Users**
```sql
users (
  user_id BIGSERIAL PRIMARY KEY,
  email VARCHAR(255) UNIQUE,
  name VARCHAR(100),
  avatar_url VARCHAR(500),
  verified BOOLEAN,
  created_at TIMESTAMP
)
```

**Reviews**
```sql
reviews (
  review_id BIGSERIAL PRIMARY KEY,
  booking_id BIGINT,
  listing_id BIGINT,
  reviewer_id BIGINT,
  reviewee_id BIGINT,
  rating INT,
  comment TEXT,
  created_at TIMESTAMP
)
```

**Payments**
```sql
payments (
  payment_id BIGSERIAL PRIMARY KEY,
  booking_id BIGINT,
  amount DECIMAL(10,2),
  currency VARCHAR(3),
  status VARCHAR(20),
  stripe_payment_id VARCHAR(100),
  created_at TIMESTAMP
)
```

---

## 5. High-Level Architecture

### ASCII Architecture Diagram

```
                                    ┌─────────────────────────────────────────────────────────────┐
                                    │                    CLIENTS (Web, Mobile)                     │
                                    └───────────────────────────┬─────────────────────────────────┘
                                                                  │
                                                                  │ HTTPS / HTTP Streaming
                                                                  ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                              API GATEWAY (Kong/AWS ALB)                                        │
│                                         Rate limiting, Auth, Routing                                           │
└───────────────────────────┬──────────────────────────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┬───────────────────┬───────────────────┐
        ▼                   ▼                   ▼                   ▼                   ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│ Search        │   │ Listing       │   │ Booking       │   │ Payment       │   │ Review        │
│ Service       │   │ Service       │   │ Service       │   │ Service       │   │ Service       │
│               │   │               │   │               │   │               │   │               │
│ - Geo filter  │   │ - CRUD        │   │ - Availability│   │ - Stripe      │   │ - CRUD        │
│ - Filters     │   │ - Calendar    │   │ - Reserve     │   │ - Payouts     │   │ - Ratings     │
│ - Streaming   │   │ - Pricing     │   │ - Confirm     │   │               │   │               │
└───────┬───────┘   └───────┬───────┘   └───────┬───────┘   └───────┬───────┘   └───────┬───────┘
        │                   │                   │                   │                   │
        │                   │                   │                   │                   │
        ▼                   ▼                   ▼                   ▼                   ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│ Elasticsearch │   │ PostgreSQL    │   │ PostgreSQL    │   │ Stripe API    │   │ PostgreSQL    │
│ PostGIS       │   │ (Listings)    │   │ (Bookings)    │   │               │   │ (Reviews)     │
│ (Search)      │   │ Redis (Cache) │   │ Redis (Lock)  │   │               │   │               │
└───────────────┘   └───────────────┘   └───────────────┘   └───────────────┘   └───────────────┘
        │                   │                   │
        │                   │                   │  Distributed Lock (Redlock/ZooKeeper)
        │                   │                   ▼
        │                   │           ┌───────────────┐
        │                   │           │ Redis Cluster │
        │                   │           │ (Locks, Cache)│
        │                   │           └───────────────┘
        │                   │
        ▼                   ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                    CDN (CloudFront / Cloudflare)                                               │
│                              Listing photos, static assets                                                     │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 6. Detailed Component Design

### 6.1 Search Service (Elasticsearch + PostGIS)

**Elasticsearch**
- Index: listings with title, description, amenities, city, country
- Geo-distance query: `geo_distance` filter for lat/lng + radius
- Filters: price range, room type, amenities (bool query)
- Sort: relevance, price, distance

**PostGIS (PostgreSQL)**
- For complex geo: polygons (neighborhoods), `ST_DWithin` for radius
- Hybrid: Elasticsearch for full-text + filters; PostGIS for advanced geo

**HTTP Streaming**
- Query Elasticsearch with scroll or search_after
- Stream chunks: `[{...}, {...}]\n` as NDJSON or chunked JSON array
- Client renders incrementally; reduces TTFB

### 6.2 Booking Service & Double Booking Prevention

**Flow**: Check → Reserve (hold) → Payment → Confirm

1. **Check**: Query `calendar_availability` + `bookings` for conflicts
2. **Reserve**: Acquire distributed lock on `(listing_id, check_in, check_out)`; insert into `reservations_hold`; release lock after insert
3. **Payment**: Charge via Stripe
4. **Confirm**: Insert into `bookings`; delete from `reservations_hold`; update `calendar_availability`

**Distributed Locking (Double Booking Prevention)**
- **Key**: `lock:listing:{id}:{check_in}:{check_out}`
- **Implementation**: Redlock (Redis) or ZooKeeper
- **TTL**: 30 seconds (prevent deadlock)
- **Retry**: Exponential backoff if lock fails
- **Pessimistic**: Lock before any write; optimistic alternative: version/conditional update

### 6.3 Pricing Engine

- **Base price**: From listing
- **Date-specific**: Override from `calendar_availability`
- **Cleaning fee**: One-time
- **Service fee**: Platform % (e.g., 14%)
- **Taxes**: Per jurisdiction (simplified: flat %)
- **Formula**: `(nightly × nights) + cleaning + service + tax`

### 6.4 Calendar Management

- Host updates `calendar_availability` in bulk
- Block dates: `available = false`
- Dynamic pricing: `price` per date
- Min/max nights: Per date or listing-level
- Sync: Event to update Elasticsearch if listing attributes change

### 6.5 Review System

- Guest reviews after checkout (within 14 days)
- Host can respond
- Ratings: 1-5; aggregate stored on listing for display
- Moderation: Flagged content queue

### 6.6 Host/Guest Matching

- Messaging: Store in Cassandra (partition by conversation_id)
- Trust: Verified ID, reviews, response rate
- Search ranking: Boost highly-rated, verified hosts

---

## 7. Scaling

### Sharding
- **PostgreSQL**: Shard listings by `listing_id` hash; bookings by `listing_id` or `guest_id`
- **Elasticsearch**: Shards by listing_id
- **Cassandra**: Partition by conversation_id

### Caching
- **Listing detail**: Redis, key `listing:{id}`, TTL 1h
- **Availability**: Redis, key `avail:{listing_id}:{date}`, TTL 5m
- **Search**: Elasticsearch is the cache; warm indices for popular regions

### CDN
- All listing photos served via CDN
- Cache-Control: long TTL for immutable URLs

### Database
- Read replicas for search, listing reads
- Primary for bookings, payments (strong consistency)

---

## 8. Failure Handling

### Double Booking
- Distributed lock + idempotent confirm (idempotency key)
- If payment fails after reserve: release hold, allow others to book
- If confirm fails: retry with idempotency; eventual consistency check

### Payment Failures
- Retry with exponential backoff
- Webhook from Stripe for async confirmation
- Refund flow: separate service

### Search Downtime
- Fallback: Query PostgreSQL with PostGIS (slower, less features)
- Cache: Serve stale results with disclaimer

### Availability Cache Inconsistency
- Short TTL; on booking confirm, invalidate cache for dates
- Write-through: Update cache on calendar change

---

## 9. Monitoring & Observability

### Key Metrics
- **Search**: Latency p50/p99, result count, streaming chunk latency
- **Booking**: Reserve success rate, confirm success rate, lock contention
- **Payment**: Success rate, Stripe error rate
- **Availability**: Cache hit ratio, DB query latency

### Alerts
- Double booking detected (audit log)
- Payment failure rate > 1%
- Search latency p99 > 500ms
- Lock timeout rate > 5%

### Tracing
- Trace ID across: Search → Listing → Reserve → Payment → Confirm
- Correlation for debugging failed bookings

---

## 10. Interview Tips

### Follow-up Questions
- "How would you prevent double booking if two users reserve the same dates simultaneously?"
- "How does HTTP streaming save money? Walk through the flow."
- "How would you implement 'similar listings' recommendation?"
- "How do you handle timezone issues for check-in/check-out?"
- "How would you add instant booking (no host approval)?"

### Common Mistakes
- **Ignoring double booking**: Must discuss distributed locking or serialization
- **No geo**: Elasticsearch/PostGIS is core for search
- **Booking as single step**: Reserve → Payment → Confirm is critical
- **Ignoring pricing**: Cleaning fee, service fee, taxes matter

### Key Points to Emphasize
- **Search**: Elasticsearch + PostGIS; geo + filters; HTTP streaming for large result sets
- **Booking flow**: Check → Reserve (lock) → Payment → Confirm
- **Double booking**: Distributed lock on (listing, dates)
- **Pricing engine**: Base + date override + fees + tax
- **Scale**: 7M listings, 150M users

---

## Appendix: Extended Design Details & Walkthrough Scenarios

### A. HTTP Streaming Deep Dive (Airbnb $84M Savings)

**Problem**: Search returns 1000s of listings; building full response blocks server, increases memory, delays first byte.

**Solution**: Stream results as they're found.
- Server: `Transfer-Encoding: chunked`; send JSON objects or array chunks
- Client: Parse incrementally; render cards as they arrive
- **Savings**: Lower server CPU/memory; faster TTFB; better UX; reduced infra cost

**Implementation**:
```
HTTP/1.1 200 OK
Transfer-Encoding: chunked
Content-Type: application/x-ndjson

{"id":1,"title":"..."}
{"id":2,"title":"..."}
...
```

### B. Double Booking Prevention Walkthrough

**Scenario**: User A and User B both try to book Listing X for Jan 10-12.

1. Both call `/bookings/check` → both get `available: true`
2. User A calls `/bookings/reserve` first
3. Booking service acquires lock `lock:listing:X:2025-01-10:2025-01-12`
4. Inserts into `reservations_hold`; releases lock
5. User B calls `/bookings/reserve`
6. Lock acquired; check `reservations_hold` and `bookings` for conflicts
7. Conflict found (User A's hold); return `available: false`
8. User A completes payment; confirm; delete hold; insert booking
9. User B cannot book (availability updated)

### C. Elasticsearch Geo Query Example

```json
{
  "query": {
    "bool": {
      "must": [
        {
          "geo_distance": {
            "distance": "10km",
            "location": { "lat": 37.77, "lon": -122.41 }
          }
        },
        { "range": { "base_price": { "gte": 50, "lte": 200 } } },
        { "term": { "amenities": "wifi" } }
      ]
    }
  },
  "sort": [{ "_geo_distance": { "location": [37.77, -122.41], "order": "asc" } }]
}
```

### D. Redlock Algorithm (Distributed Lock)

1. Get current time
2. Try to acquire lock on N/2+1 Redis nodes (N=5 typically)
3. Set key with value (random), TTL (e.g., 30s)
4. If acquired on majority: lock held; compute elapsed time
5. If elapsed < TTL: success
6. Release: delete key on all nodes (only if value matches)

### E. Pricing Engine Formula

```
nightly_total = sum(price[date] for each night)
cleaning_fee = listing.cleaning_fee
subtotal = nightly_total + cleaning_fee
service_fee = subtotal * 0.14
tax = (subtotal + service_fee) * tax_rate
total = subtotal + service_fee + tax
```

### F. Calendar Bulk Update

- Host selects date range; sets available/price/min_nights
- Batch upsert into `calendar_availability`
- Invalidate Redis cache for affected listing+dates
- Optionally: Event to sync to search index if pricing affects ranking

### G. Review Aggregation

- On new review: Update `listings.review_count`, `listings.avg_rating` (or separate table)
- Use Redis for hot listings: `review_summary:{listing_id}`
- Periodic batch: Recompute from `reviews` if drift

### H. Idempotency for Payment

- Client sends `Idempotency-Key: uuid` on confirm
- Server: Check if key seen; if yes, return cached response
- Store: Redis or DB; TTL 24h
- Prevents duplicate charges on retry

### I. Reserve Hold Expiry

- `reservations_hold` has `expires_at` (e.g., 10 min)
- Background job: Delete expired holds every minute
- On confirm: Delete hold immediately
- Prevents permanent blocks from abandoned carts

### J. Search Result Ranking Factors

- Relevance: Text match score
- Distance: Closer = higher
- Price: Within budget
- Rating: Higher = higher
- Availability: Must have dates
- Recency: Recently updated listings
- Host response rate
