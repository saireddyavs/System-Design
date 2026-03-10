# Ride-Sharing Service - Low Level Design

A production-quality, interview-ready implementation of a ride-sharing service (like Uber/Lyft) in Go, following clean architecture and SOLID principles.

## Problem Description

Design a ride-sharing platform that connects riders with drivers in real-time. The system must handle driver/rider registration, ride matching, fare calculation, payment processing, and ratings.

## Requirements

| Feature | Description |
|---------|-------------|
| **Driver Management** | Registration, availability toggle, location updates |
| **Rider Management** | Registration, ride requests |
| **Ride Matching** | Match nearest available driver to rider |
| **Real-time Tracking** | Ride status updates (Observer pattern) |
| **Fare Calculation** | Based on distance and time, with surge pricing |
| **Cancellation** | Handle cancellation with penalty when in progress |
| **Payment** | Process payments for completed rides |
| **Rating System** | Rate drivers and riders after ride completion |

## Core Entities & Relationships

```
┌─────────┐     requests      ┌──────┐     assigns      ┌────────┐
│  Rider  │ ────────────────► │ Ride │ ◄────────────── │ Driver │
└────┬────┘                    └──┬───┘                 └────────┘
     │                            │
     │ rates                      │ generates
     ▼                            ▼
┌────────┐                   ┌─────────┐
│ Rating │                   │ Payment │
└────────┘                   └─────────┘
```

### Entity Details

1. **Driver**: ID, Name, Phone, Vehicle, Location, Status (Available/OnRide/Offline/Deactivated), Rating
2. **Rider**: ID, Name, Phone, Email, Location, Rating
3. **Ride**: ID, RiderID, DriverID, Pickup, Dropoff, Status, Fare, Distance, Duration, StartTime, EndTime
4. **Location**: Latitude, Longitude (with Haversine distance calculation)
5. **Payment**: ID, RideID, Amount, Method (Card/Wallet/Cash), Status
6. **Rating**: ID, RideID, FromUserID, ToUserID, Score (1-5), Comment

## Matching Algorithm - Nearest Driver

The **Haversine formula** calculates great-circle distance between two points on Earth:

```
a = sin²(Δlat/2) + cos(lat1) × cos(lat2) × sin²(Δlon/2)
c = 2 × atan2(√a, √(1−a))
d = R × c  (R = 6371 km)
```

**Algorithm Steps:**
1. Get all available drivers within `MaxSearchRadiusKm` (default 50km)
2. Filter drivers with rating ≥ 3.0 (business rule: low-rated drivers excluded)
3. Calculate Haversine distance from rider location to each driver
4. Sort by distance ascending
5. Return nearest driver, or `ErrNoDriverAvailable` if none in radius

**Alternative Strategy**: `HighestRatedStrategy` - same filtering, but sorts by rating (desc) then distance (asc).

## Fare Calculation Strategy

**Base Formula**: `Fare = BaseFare + (Distance × PerKmRate) + (Duration × PerMinuteRate)`

**Surge Pricing**: When active requests exceed threshold (default 10), apply multiplier (default 1.5x).

```go
effectiveFare = baseFare × surgeMultiplier
```

## Design Patterns

| Pattern | Where | Why |
|---------|-------|-----|
| **Strategy** | `MatchingStrategy`, `FareCalculator` | Swap matching algorithms (nearest vs highest-rated) and fare formulas (base vs surge) without changing client code. Open/Closed Principle. |
| **Observer** | `RideNotifier`, `RideObserver` | Notify multiple subscribers (push notifications, analytics, logging) when ride status changes. Loose coupling. |
| **State** | `RideStatus` (Requested → DriverAssigned → InProgress → Completed/Cancelled) | Encapsulate state-specific behavior. Invalid transitions are prevented. |
| **Repository** | `DriverRepository`, `RideRepository`, etc. | Abstract data access. Swap in-memory for PostgreSQL without changing business logic. |

## SOLID Principles Mapping

| Principle | Implementation |
|-----------|----------------|
| **S - Single Responsibility** | `DriverService` only handles driver logic; `MatchingService` only handles matching; `RideService` orchestrates. |
| **O - Open/Closed** | New matching strategies (e.g., `EcoFriendlyStrategy`) or fare strategies can be added without modifying existing code. |
| **L - Liskov Substitution** | Any `MatchingStrategy` implementation can replace another; any `FareCalculator` works in `RideService`. |
| **I - Interface Segregation** | Small, focused interfaces: `DriverRepository`, `RideRepository`, `MatchingStrategy`, `FareCalculator`. |
| **D - Dependency Inversion** | Services depend on interfaces (e.g., `interfaces.DriverRepository`), not concrete `InMemoryDriverRepository`. |

## Business Rules

- **Surge pricing**: When active requests > 10, fare × 1.5
- **Cancellation penalty**: 50% of base fare if ride is already in progress (driver picked up rider)
- **Driver deactivation**: Rating < 3.0 → status set to Deactivated
- **Max search radius**: 50 km for driver matching

## Concurrency Considerations

- **Thread-safe repositories**: All in-memory repos use `sync.RWMutex` for concurrent read/write
- **Model-level locking**: `Driver`, `Rider`, `Ride` use `sync.RWMutex` for location/status updates
- **Observer notification**: `RideNotifier` copies observer list under lock before notifying to avoid deadlock
- **Concurrent ride requests**: Multiple riders can request simultaneously; each gets nearest available driver

## Project Structure

```
07-ride-sharing-service/
├── cmd/main.go              # Demo wiring
├── internal/
│   ├── models/             # Domain entities
│   ├── interfaces/         # Repository & strategy interfaces
│   ├── services/           # Business logic
│   ├── repositories/      # In-memory implementations
│   └── strategies/         # Matching & fare strategies
├── tests/                  # Unit tests
├── go.mod
└── README.md
```

## Running the Demo

```bash
go run ./cmd/main.go
```

## Running Tests

```bash
go test ./tests/... -v
```

## Interview Explanations

### 3-Minute Summary

"We built a ride-sharing service with clean architecture. **Drivers** and **riders** register and update locations. When a rider requests a ride, we use a **Strategy** to find the nearest available driver via Haversine distance. The **Ride** moves through states: Requested → DriverAssigned → InProgress → Completed. We use **Repository** pattern for data access and **Observer** for status notifications. Fare is calculated with base + distance + time, with surge pricing when demand is high. Cancellation has a penalty if the ride has started. Drivers with rating below 3.0 get deactivated. All components are thread-safe and depend on interfaces for testability."

### 10-Minute Deep Dive

1. **Architecture**: Clean architecture with models, interfaces, services, and repositories. Dependencies point inward; business logic has no framework dependencies.

2. **Matching**: We inject `MatchingStrategy` into `MatchingService`. Default is `NearestDriverStrategy` using Haversine. We could add `HighestRatedStrategy` or `EcoFriendlyStrategy` without changing the service.

3. **Fare**: `FareCalculator` interface allows different formulas. `BaseFareStrategy` does base + distance + time. Surge multiplier is computed from active request count.

4. **State**: Ride status is an enum. Transitions are validated in the service (e.g., can't complete a ride that hasn't started).

5. **Observer**: `RideNotifier` holds observers. When ride status changes, we notify all. Useful for push notifications, analytics, webhooks.

6. **Concurrency**: Repositories use RWMutex. Models like Driver have internal mutexes for location/status. Tests include concurrent ride requests.

7. **Testing**: Unit tests cover matching (nearest, no driver, low-rated exclusion), fare (base, surge), and ride lifecycle (request, complete, cancel, rating).

## Future Improvements

- **Persistence**: Replace in-memory repos with PostgreSQL/MongoDB
- **Real-time**: WebSocket/SSE for live ride tracking
- **Geospatial**: Use PostGIS or Redis Geo for efficient "drivers near point" queries
- **Event sourcing**: Store ride events for audit and replay
- **Idempotency**: Idempotency keys for ride requests to handle retries
- **Circuit breaker**: For payment processor failures
- **Rate limiting**: Per-user request limits
- **Caching**: Cache driver locations with TTL for high read load
