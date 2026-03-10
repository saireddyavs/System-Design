# Design Uber / Ride-Sharing System

## 1. Problem Statement & Requirements

### Problem Statement
Design a ride-sharing platform that connects riders with nearby drivers in real-time, handles trip lifecycle, calculates fares, processes payments, and provides real-time tracking with ETA.

### Functional Requirements
- **Rider**: Request ride, view nearby drivers, real-time tracking, see ETA, view fare estimate, pay, view trip history
- **Driver**: Accept/decline ride requests, navigate to pickup, start/end trip, receive earnings, view trip history
- **Matching**: Match rider with nearby available drivers based on proximity, ETA, rating
- **Fare Calculation**: Distance-based, time-based, surge pricing during high demand
- **Payment**: Cash, card, wallet; post-trip or scheduled
- **ETA Service**: Accurate arrival time for driver to rider and rider to destination

### Non-Functional Requirements
- **Latency**: Match rider to driver in < 3 seconds
- **Availability**: 99.99% uptime
- **Consistency**: Strong consistency for trip state; eventual consistency acceptable for location
- **Scale**: 20M rides/day (~230 QPS average, 10x peak = 2,300 QPS)
- **Real-time**: Driver location updates every 4 seconds; sub-second propagation to rider app

### Out of Scope
- Driver onboarding/verification
- In-app messaging/chat
- Ride scheduling (future rides)
- Multi-stop trips
- Driver earnings dashboard (simplified)

---

## 2. Back-of-Envelope Estimation

### Assumptions
- 20M rides/day
- 2M active drivers
- 50M active riders
- Average trip: 15 min, 5 miles
- Location updates: every 4 sec per driver

### QPS Estimates
| Component | Calculation | QPS |
|-----------|-------------|-----|
| Ride requests | 20M / 86400 | ~230 |
| Peak (10x) | 230 × 10 | ~2,300 |
| Location updates | 2M drivers × 0.25/sec | 500,000 |
| Trip state updates | 20M × 4 states / 86400 | ~900 |
| Fare lookups | 20M / 86400 | ~230 |
| ETA requests | 20M × 3 / 86400 | ~700 |

### Storage (1 year)
| Data | Size per record | Records | Total |
|------|-----------------|---------|-------|
| Trips | 1 KB | 7.3B | 7.3 TB |
| Location history | 50 B | 15.7T | 785 TB |
| Driver locations (hot) | 100 B | 2M | 200 MB |
| Users | 500 B | 52M | 26 GB |

### Bandwidth
- Location updates: 500K × 100 B = 50 MB/s outbound
- Real-time tracking: 20M active trips × 0.25 updates = 5M updates/s (simplified)

### Cache
- Active driver locations: 2M × 100 B ≈ 200 MB (Redis)
- Surge multipliers: ~10K zones × 8 B ≈ 80 KB
- ETA cache: 100M entries × 50 B ≈ 5 GB (with TTL)

---

## 3. API Design

### REST Endpoints

```
POST   /api/v1/rides/request          # Request a ride
GET    /api/v1/rides/{rideId}         # Get ride details
PATCH  /api/v1/rides/{rideId}/cancel  # Cancel ride
GET    /api/v1/rides/{rideId}/eta     # Get ETA
GET    /api/v1/rides/{rideId}/fare    # Get fare estimate
POST   /api/v1/rides/{rideId}/pay     # Process payment

POST   /api/v1/drivers/{driverId}/availability  # Set available/busy
POST   /api/v1/drivers/{driverId}/rides/{rideId}/accept   # Accept ride
POST   /api/v1/drivers/{driverId}/rides/{rideId}/decline  # Decline ride
POST   /api/v1/drivers/{driverId}/rides/{rideId}/start    # Start trip
POST   /api/v1/drivers/{driverId}/rides/{rideId}/complete # Complete trip

GET    /api/v1/users/{userId}/trips   # Trip history
GET    /api/v1/nearby/drivers         # Get nearby drivers (rider)
```

### WebSocket Endpoints

```
WS /ws/location                    # Driver pushes location (every 4 sec)
WS /ws/rides/{rideId}              # Rider receives driver location, ETA updates
WS /ws/drivers/{driverId}/requests # Driver receives ride requests
```

### Request/Response Examples

**Request Ride**
```json
POST /api/v1/rides/request
{
  "rider_id": "r_123",
  "pickup": {"lat": 37.7749, "lng": -122.4194},
  "dropoff": {"lat": 37.7849, "lng": -122.4094},
  "vehicle_type": "economy"
}

Response 201:
{
  "ride_id": "ride_456",
  "status": "MATCHED",
  "driver": {
    "id": "d_789",
    "name": "John",
    "rating": 4.8,
    "eta_minutes": 5,
    "location": {"lat": 37.77, "lng": -122.42}
  },
  "fare_estimate": {"min": 12.50, "max": 18.00},
  "surge_multiplier": 1.2
}
```

---

## 4. Data Model / Database Schema

### Primary Database: PostgreSQL (trips, users) + Cassandra (location history)

### Tables

**trips**
| Column | Type | Description |
|--------|------|--------------|
| id | UUID PK | Trip ID |
| rider_id | UUID FK | Rider |
| driver_id | UUID FK | Driver |
| status | ENUM | REQUESTED, MATCHED, EN_ROUTE, IN_PROGRESS, COMPLETED, CANCELLED |
| pickup_lat, pickup_lng | DECIMAL | Pickup coordinates |
| dropoff_lat, dropoff_lng | DECIMAL | Dropoff coordinates |
| fare_amount | DECIMAL | Final fare |
| surge_multiplier | DECIMAL | Surge at request time |
| created_at | TIMESTAMP | |
| completed_at | TIMESTAMP | |
| distance_km | DECIMAL | |
| duration_minutes | INT | |

**drivers**
| Column | Type |
|--------|------|
| id | UUID PK |
| name | VARCHAR |
| rating | DECIMAL |
| status | ENUM (AVAILABLE, BUSY, OFFLINE) |
| current_lat, current_lng | DECIMAL |
| current_ride_id | UUID FK |
| updated_at | TIMESTAMP |

**driver_locations** (time-series, Cassandra)
| Column | Type |
|--------|------|
| driver_id | UUID PK |
| timestamp | TIMESTAMP PK |
| lat | DOUBLE |
| lng | DOUBLE |
| heading | FLOAT |

### DB Choice Rationale
- **PostgreSQL**: ACID for trips, payments; complex queries for analytics
- **Cassandra**: High write throughput for location; time-series; tunable consistency
- **Redis**: Driver locations (hot data), surge multipliers, session state

---

## 5. High-Level Architecture

```
                                    ┌─────────────────────────────────────────────────────────────┐
                                    │                      LOAD BALANCER                           │
                                    └─────────────────────────────────────────────────────────────┘
                                                              │
                    ┌─────────────────────────────────────────┼─────────────────────────────────────────┐
                    │                                         │                                         │
                    ▼                                         ▼                                         ▼
         ┌──────────────────┐                      ┌──────────────────┐                      ┌──────────────────┐
         │   API Gateway    │                      │   API Gateway     │                      │   API Gateway    │
         │   (Riders)       │                      │   (Drivers)      │                      │   (WebSocket)    │
         └────────┬─────────┘                      └────────┬─────────┘                      └────────┬─────────┘
                  │                                          │                                          │
                  └──────────────────────────────────────────┼──────────────────────────────────────────┘
                                                             │
         ┌──────────────────────────────────────────────────┼──────────────────────────────────────────────────┐
         │                                                  │                                                   │
         ▼                                                  ▼                                                   ▼
┌─────────────────┐                            ┌─────────────────┐                            ┌─────────────────┐
│  Trip Service   │◄───────────────────────────►│ Matching Service│                            │ Location Service│
│  (State Machine)│                            │ (Driver Match)  │                            │ (Geohash/Redis) │
└────────┬────────┘                            └────────┬────────┘                            └────────┬────────┘
         │                                               │                                               │
         │                                               │                                               │
         ▼                                               ▼                                               ▼
┌─────────────────┐                            ┌─────────────────┐                            ┌─────────────────┐
│ Fare Service    │                            │ ETA Service     │                            │ Surge Service   │
│ (Distance/Time) │                            │ (Graph + A*)    │                            │ (Supply/Demand) │
└────────┬────────┘                            └────────┬────────┘                            └────────┬────────┘
         │                                               │                                               │
         └───────────────────────────────────────────────┼───────────────────────────────────────────────┘
                                                         │
         ┌───────────────────────────────────────────────┼───────────────────────────────────────────────┐
         │                                               │                                               │
         ▼                                               ▼                                               ▼
┌─────────────────┐                            ┌─────────────────┐                            ┌─────────────────┐
│ Payment Service │                            │  Message Queue  │                            │  Redis Cluster  │
│ (Stripe/PayPal) │                            │  (Kafka)        │                            │  (Locations)    │
└────────┬────────┘                            └────────┬────────┘                            └────────┬────────┘
         │                                               │                                               │
         └───────────────────────────────────────────────┼───────────────────────────────────────────────┘
                                                         │
         ┌───────────────────────────────────────────────┼───────────────────────────────────────────────┐
         │                                               │                                               │
         ▼                                               ▼                                               ▼
┌─────────────────┐                            ┌─────────────────┐                            ┌─────────────────┐
│   PostgreSQL    │                            │    Cassandra    │                            │   Elasticsearch │
│   (Trips)       │                            │ (Location Hist) │                            │   (Trip Search) │
└─────────────────┘                            └─────────────────┘                            └─────────────────┘
```

### Component Descriptions
- **API Gateway**: Auth, rate limiting, routing
- **Trip Service**: Orchestrates trip lifecycle, state machine
- **Matching Service**: Finds and ranks nearby drivers, dispatches
- **Location Service**: Stores/queries driver locations (geohash, Redis)
- **ETA Service**: Road network graph, Dijkstra/A*, traffic
- **Fare Service**: Distance, time, surge
- **Surge Service**: Supply/demand, zone-based multipliers
- **Payment Service**: Payment gateway integration

---

## 6. Detailed Component Design

### 6.1 Location Service

**Storage Strategy**: Geohashing (e.g., H3 hexagonal grid or Geohash)

- **Geohash**: Encode (lat, lng) → string. Prefix match = same region.
- **H3**: Uber's hexagonal grid; uniform cell size; 15 resolution levels.
- **Storage**: Redis sorted sets or spatial index.
  - Key: `drivers:geohash:{geohash}` or `drivers:h3:{h3_index}`
  - Value: `driver_id`, `lat`, `lng`, `heading`, `timestamp`
  - TTL: 30 seconds (stale = offline)

**Driver Location Update Flow**:
1. Driver app sends location every 4 seconds via WebSocket/HTTP
2. Location Service computes geohash/H3 for new position
3. Remove driver from previous cell, add to new cell (if changed)
4. Update Redis: `HSET driver:{id} lat lng updated_at`
5. Optionally write to Cassandra for history/analytics (async)

**Query Nearby Drivers**:
1. Compute geohash of pickup point
2. Get drivers in same cell + neighboring cells (expand radius)
3. Filter by max distance (e.g., 5 km), availability
4. Return sorted by distance or ETA

### 6.2 Matching Service

**Algorithm**:
1. Receive ride request with pickup location
2. Call Location Service: get nearby available drivers
3. Call ETA Service: get ETA from each driver to pickup
4. Rank by: ETA (primary), rating, acceptance rate
5. Dispatch to top driver(s) in parallel (first accept wins)
6. Timeout: 15 seconds; if no accept, expand radius and retry

**Dispatch**:
- Send request to top 3–5 drivers concurrently
- First acceptance wins; others get cancellation
- Track acceptance rate for future ranking

### 6.3 Trip Service — State Machine

```
REQUESTED ──(match)──► MATCHED ──(driver_arrives)──► EN_ROUTE ──(start)──► IN_PROGRESS ──(complete)──► COMPLETED
    │                      │                              │                      │
    └──(cancel)────────────┴──(cancel)────────────────────┴──(cancel)─────────────┘
                                    │
                                    ▼
                               CANCELLED
```

- **REQUESTED**: Ride created, matching in progress
- **MATCHED**: Driver accepted, navigating to pickup
- **EN_ROUTE**: Driver at pickup, rider entering
- **IN_PROGRESS**: Trip started, en route to destination
- **COMPLETED**: Trip ended, fare finalized
- **CANCELLED**: Anytime before IN_PROGRESS (with possible fee)

### 6.4 Fare Calculation

```
fare = base_fare + (distance_km × rate_per_km) + (duration_min × rate_per_min) × surge_multiplier
```

- **Base fare**: Fixed (e.g., $2.50)
- **Distance rate**: Per km
- **Time rate**: Per minute (traffic)
- **Surge**: Multiplier 1.0–3.0+ by zone

### 6.5 ETA Service

**Data**: Road network graph (nodes = intersections, edges = road segments with length, speed limit)

**Algorithm**:
- **Dijkstra** or **A*** for shortest path
- **Weights**: Distance or time (prefer time for ETA)
- **Traffic**: Adjust edge weights from real-time speed data
- **Contraction Hierarchies**: Precompute shortcuts for fast queries (used in production)

**Flow**:
1. Snap origin/destination to nearest graph nodes
2. Run A* with traffic-weighted edges
3. Sum edge times → ETA
4. Cache: (origin_cell, dest_cell) → ETA (5 min TTL)

### 6.6 Real-Time Tracking

- **Driver → Server**: Location every 4 sec via WebSocket or batched HTTP
- **Server → Rider**: Push via WebSocket (same ride_id)
- **Fan-out**: Location Service publishes to Kafka; WebSocket service consumes and pushes to connected riders

### 6.7 Surge Pricing

**Inputs**: Supply (available drivers), demand (pending requests) per zone

**Formula** (simplified):
```
surge = max(1.0, min(3.0, demand / supply × k))
```

- Zones: Geohash/H3 cells
- Update every 1–5 minutes
- Store in Redis: `surge:{zone_id}` → multiplier

### 6.8 Payment Integration

- Post-trip: Trigger payment when status → COMPLETED
- Use Stripe/PayPal: tokenize cards, charge via API
- Idempotency: `payment_id` + `ride_id` to avoid double charge
- Retries with exponential backoff

---

## 7. Scaling

### Sharding
- **Trips**: Shard by `ride_id` (UUID) or `rider_id` (user-based)
- **Location**: By geohash/H3 (each region → own cluster)
- **Cassandra**: Partition by `(driver_id, date)` for location history

### Caching
- **Redis**: Driver locations (hot), surge multipliers, ETA cache
- **CDN**: Static assets (app binaries, images)
- **Application cache**: User sessions, trip metadata

### Read Replicas
- PostgreSQL: Replicas for trip history, analytics
- Cassandra: Multiple replicas per partition

### Message Queue
- **Kafka**: Location events, trip events, payment events
- Enables async processing, replay, multiple consumers

---

## 8. Failure Handling

### Component Failures
- **Location Service down**: Use last known location from DB; degrade gracefully
- **Matching Service down**: Queue requests; process when back
- **ETA Service down**: Fallback to straight-line distance estimate
- **Payment Service down**: Queue payment; retry; notify user

### Redundancy
- Multi-AZ deployment for all services
- Redis Cluster: Replication, automatic failover
- PostgreSQL: Primary + standby
- Cassandra: Replication factor 3

### Data Loss
- Kafka: Retention for replay
- Cassandra: Multiple replicas
- Payments: Idempotent, auditable

---

## 9. Monitoring & Observability

### Key Metrics
| Metric | Target | Alert |
|--------|--------|-------|
| Match latency p99 | < 3 sec | > 5 sec |
| Location update latency | < 100 ms | > 500 ms |
| Trip creation success rate | > 99.9% | < 99% |
| Payment success rate | > 99.5% | < 99% |
| API error rate | < 0.1% | > 1% |

### Logging
- Structured logs (JSON) with trace_id, ride_id, user_id
- Centralized (e.g., ELK, Splunk)

### Tracing
- Distributed tracing (Jaeger, Zipkin) across Trip, Matching, ETA, Payment

### Dashboards
- Rides/min by region
- Active drivers by region
- Surge zones
- Match success rate, latency percentiles

---

## 10. Interview Tips

### Follow-up Questions
1. How would you handle a driver going offline mid-request?
2. How do you prevent thundering herd when many riders request at once?
3. How would you implement ride scheduling (future rides)?
4. How do you handle split payments?
5. How would you add multi-stop trips?

### Common Mistakes
- Ignoring location update scale (500K QPS)
- Not considering geospatial indexing (geohash/H3)
- Overcomplicating matching (start with simple distance-based)
- Forgetting idempotency for payments
- Not defining clear trip state transitions

### What to Emphasize
- Location Service design (geohash, Redis, high write throughput)
- Trip state machine (clear, auditable)
- ETA with traffic (graph + A*)
- Surge pricing (supply/demand)
- Real-time architecture (WebSocket, Kafka)

---

## Appendix A: Geohash Deep Dive

### How Geohash Works
- Encode (lat, lng) into a single string
- Each character adds precision (base32)
- Example: "9q8yy" = San Francisco area
- Prefix match: "9q8" contains all "9q8yy", "9q8yz", etc.
- Neighbors: Same-length hashes for adjacent cells

### H3 Hexagonal Grid (Uber)
- Uber developed H3 for more uniform cell sizes
- Hexagons avoid edge cases at poles
- 15 resolution levels (0 = global, 15 = ~1 m²)
- Better for "drivers in radius" queries
- Library: h3-py, h3-js

### Redis Sorted Set for Location
```
ZADD drivers:zone:9q8yy {timestamp} {driver_id}
ZRANGEBYSCORE drivers:zone:9q8yy -inf +inf
# Get all drivers in cell, filter by timestamp (active)
```

---

## Appendix B: Supply-Demand Forecasting

### Purpose
- Predict demand by time/area for driver positioning
- Surge pricing input
- Capacity planning

### Features
- Historical: Same hour, day of week, weather
- Events: Concerts, sports, holidays
- Real-time: Current requests, cancellations

### Model
- Time series (ARIMA, Prophet) or ML (XGBoost, LSTM)
- Output: Expected demand per zone per hour
- Use: Recommend drivers to move to high-demand zones

---

## Appendix C: Matching Algorithm Variants

### Immediate Dispatch
- Send to single best driver
- Fast but lower acceptance rate

### Broadcast (First Accept)
- Send to top 3–5 drivers
- First acceptance wins; others get cancel
- Higher match rate; slight delay

### Auction
- Drivers bid; system picks best
- More complex; used for premium services

### Batching
- Collect requests for 30–60 sec
- Batch match to optimize global efficiency
- Trade-off: Latency vs. efficiency

---

## Appendix D: WebSocket Message Flow

```
Rider App                API Gateway              Location Service           Driver App
    |                         |                          |                        |
    |-- Connect (ride_id) ---->|                          |                        |
    |                         |-- Subscribe ride:123 ---->|                        |
    |                         |                          |                        |
    |                         |                          |<-- Location (lat,lng)-|
    |                         |                          |   (every 4 sec)        |
    |                         |<-- Publish location -----|                        |
    |<-- Push location -------|                          |                        |
    |                         |                          |                        |
```

### Message Format
```json
{"type": "location", "driver_id": "d_789", "lat": 37.77, "lng": -122.42, "heading": 45, "ts": 1234567890}
```

---

## Appendix E: Fare Calculation Examples

### Base Case
- Base: $2.50, $1.50/km, $0.25/min
- Trip: 5 km, 15 min → $2.50 + $7.50 + $3.75 = $13.75

### With Surge
- Surge: 1.5x → $13.75 × 1.5 = $20.63

### Minimum Fare
- Short trip: $5 minimum regardless of distance/time

---

## Appendix F: ETA Cache Key Design

### Cache Key
- `eta:{origin_geohash}:{dest_geohash}:{mode}`
- Geohash at resolution 6 (≈1.2 km) to balance accuracy vs. cache hit rate
- TTL: 5 minutes (traffic changes)

### Invalidation
- Traffic event (accident): Invalidate affected segments
- Time-based: Let TTL handle

---

## Appendix G: Driver Location Update Flow (Detailed)

### Client-Side
1. GPS provides (lat, lng) every 1 sec
2. Throttle: Send to server every 4 sec (or on significant movement)
3. Include: driver_id, lat, lng, heading, timestamp
4. Use WebSocket for low latency, or HTTP POST batch

### Server-Side
1. Validate: driver exists, is available or on trip
2. Update Redis: HSET driver:{id} lat lng heading updated_at
3. Compute new geohash/H3 cell
4. If cell changed: SREM from old cell, SADD to new cell
5. If on trip: Publish to Kafka for rider push
6. Async: Write to Cassandra for history (optional)

### Failure Handling
- Driver offline: No update for 30 sec → remove from available set
- Network loss: Driver buffers; sends batch on reconnect
- Duplicate: Idempotent by timestamp (last write wins)

---

## Appendix H: Trip State Transitions (Complete)

| From | To | Trigger | Side Effects |
|------|-----|---------|--------------|
| REQUESTED | MATCHED | Driver accepts | Notify rider, start ETA updates |
| REQUESTED | CANCELLED | Rider/driver cancels, timeout | Refund if applicable |
| MATCHED | EN_ROUTE | Driver arrives at pickup | Notify rider |
| MATCHED | CANCELLED | Rider/driver cancels | Cancellation fee if driver en route |
| EN_ROUTE | IN_PROGRESS | Driver starts trip | Start fare meter |
| EN_ROUTE | CANCELLED | Rider no-show (timeout) | No-show fee |
| IN_PROGRESS | COMPLETED | Driver ends trip | Calculate fare, trigger payment |
| * | CANCELLED | Anytime before IN_PROGRESS | Varies by policy |

---

## Appendix I: Payment Idempotency

### Problem
- Trip completes; payment triggered
- Retry on failure could double-charge

### Solution
- Idempotency key: `payment:{ride_id}`
- First request: Process payment; store result
- Second request: Return stored result (no new charge)
- Use payment gateway's idempotency (Stripe)

### Flow
```
1. Trip COMPLETED
2. Create payment intent with idempotency_key=ride_123
3. If 200: Done
4. If 5xx: Retry with same idempotency_key
5. Stripe returns cached result
```

---

## Appendix J: Surge Zone Calculation

### Inputs (per zone, per 5 min)
- Available drivers: count
- Pending requests: count
- Historical baseline: demand/supply ratio

### Formula
```
ratio = pending_requests / max(available_drivers, 1)
surge = clamp(1.0 + k * ratio, 1.0, 3.0)
```
- k = tuning constant (e.g., 0.5)
- Clamp to 1.0–3.0 (or higher for extreme events)

### Zone Definition
- Geohash resolution 5 or 6 (≈5 km or 1.2 km)
- Or: H3 resolution 7 or 8
- Overlap: Use center point for zone assignment

---

## Appendix K: Load Balancing for Real-Time

### Sticky Sessions
- WebSocket: Sticky by ride_id or user_id
- Ensures driver location updates and rider receive on same server
- Fallback: Redis Pub/Sub for cross-server fan-out

### Health Checks
- WebSocket: Ping/pong every 30 sec
- Reconnect: Exponential backoff (1, 2, 4, 8 sec)
- Session recovery: Re-fetch ride state on reconnect

---

## Appendix L: Database Indexing for Trips

### Primary Queries
- Get trip by ID: Primary key
- Get trips by rider: INDEX on rider_id, created_at
- Get trips by driver: INDEX on driver_id, created_at
- Get active trip by rider: INDEX on rider_id WHERE status IN (...)
- Get active trip by driver: INDEX on driver_id WHERE status = IN_PROGRESS

### Partitioning
- By created_at (monthly) for archival
- By rider_id hash for sharding

---

## Appendix M: Kafka Topics for Event-Driven Flow

### Topics
- **location_updates**: Driver location (lat, lng, driver_id, ts)
- **trip_events**: Trip state changes (ride_id, status, timestamp)
- **payment_events**: Payment initiated, completed, failed
- **surge_updates**: Zone surge multiplier changes

### Consumers
- Location → WebSocket service (push to riders)
- Trip events → Analytics, notifications
- Payment events → Receipt generation, accounting

### Retention
- Location: 1 hour (real-time only)
- Trip events: 7 days (replay, analytics)
- Payment: 90 days (compliance)
