# Ride-Sharing Service — Interview Strategy (45 min)

## Time Allocation

| Phase | Time | What to Do |
|-------|------|------------|
| 1. Clarify & Scope | 3 min | Confirm driver/rider flow, matching criteria, fare formula, surge, cancellation |
| 2. Core Models | 7 min | Driver, Rider, Ride, Location, Payment, Rating |
| 3. Repository Interfaces | 5 min | DriverRepository, RiderRepository, RideRepository (with GetActiveRides, CountActiveRequests) |
| 4. Service Interfaces | 5 min | MatchingStrategy, FareCalculator, RideService, MatchingService |
| 5. Core Service Implementation | 12 min | RideService.RequestRide + MatchingStrategy.FindDriver (Haversine, filter, sort) |
| 6. Handler / main.go Wiring | 5 min | Wire repos, matching strategy, fare calculator, payment processor, notifier |
| 7. Extend & Discuss | 8 min | Haversine formula, surge multiplier, state machine, Observer for notifications |

## Phase 1: Clarify & Scope (3 min)

**Questions to ask:**
- Driver: register, go online/offline, update location?
- Rider: register, request ride with pickup/dropoff?
- Matching: nearest driver? Within radius? Minimum rating?
- Fare: base + distance + time? Surge when demand high?
- Cancellation: penalty if ride already started?
- Payment: after ride completion?

**Scope in:** Driver/rider registration, ride request with nearest-driver matching, fare calculation with surge, ride lifecycle (Requested → DriverAssigned → InProgress → Completed), payment, rating.

**Scope out:** Real-time tracking UI, multiple ride types, driver earnings dashboard.

## Phase 2: Core Models (7 min)

**Start with:** `Ride` — ID, RiderID, DriverID, Pickup, Dropoff, Status (Requested/DriverAssigned/InProgress/Completed/Cancelled), Fare, Distance, Duration, StartTime, EndTime. Status drives the state machine.

**Then:** `Driver` — ID, Name, Phone, Vehicle, Location (lat/lon), Status (Available/OnRide/Offline/Deactivated), Rating. `Rider` — ID, Name, Phone, Email, Location, Rating. `Location` — Latitude, Longitude.

**Skip for now:** Vehicle details, driver documents.

## Phase 3: Repository Interfaces (5 min)

**Essential:**
- `DriverRepository`: Create, GetByID, Update, GetAll (or GetAvailable — for matching)
- `RiderRepository`: Create, GetByID, Update
- `RideRepository`: Create, GetByID, Update, GetActiveRidesByRider, **CountActiveRequests** (for surge)
- `PaymentRepository`: Create

**Skip:** RatingRepository initially; mention for post-ride rating.

## Phase 4: Service Interfaces (5 min)

**Essential:**
- `MatchingStrategy`: `FindDriver(riderLocation, drivers) (*Driver, error)` — swap NearestDriver vs HighestRated
- `FareCalculator`: `Calculate(pickup, dropoff, duration, surgeMultiplier) float64`
- `RideService`: RequestRide, StartRide, CompleteRide, CancelRide
- `MatchingService`: FindNearestDriver(pickup) — uses DriverRepo + MatchingStrategy
- `PaymentProcessor`: ProcessPayment

**Key abstraction:** Strategy pattern for matching and fare — Open/Closed.

## Phase 5: Core Service Implementation (12 min)

**Key method:** `RideService.RequestRide(riderID, pickup, dropoff)` + `NearestDriverStrategy.FindDriver()` — this is where the core logic lives.

**RequestRide flow:**
1. Check rider has no active ride
2. Update rider location to pickup
3. Call `matchingService.FindNearestDriver(pickup)`
4. Create Ride, assign driver, persist
5. Set driver status to OnRide
6. Notify observers

**FindDriver (NearestDriverStrategy) algorithm:**
1. Get all drivers from repo (or available only)
2. Filter: `IsAvailable()`, rating >= 3.0
3. For each driver: `dist := HaversineDistance(riderLocation, driver.Location)`
4. Filter: dist <= MaxSearchRadiusKm (e.g. 50km)
5. Sort by distance ascending
6. Return first, or ErrNoDriverAvailable

**Haversine formula:** `a = sin²(Δlat/2) + cos(lat1)*cos(lat2)*sin²(Δlon/2)`; `c = 2*atan2(√a, √(1-a))`; `d = R*c` (R=6371 km).

**CompleteRide:** Calculate duration, distance (Haversine), get surge multiplier from CountActiveRequests, call FareCalculator.Calculate(..., surgeMultiplier), process payment, free driver.

**Concurrency:** Repositories use sync.RWMutex; multiple riders can request simultaneously.

## Phase 6: main.go Wiring (5 min)

```go
driverRepo := NewInMemoryDriverRepository()
riderRepo := NewInMemoryRiderRepository()
rideRepo := NewInMemoryRideRepository()
paymentRepo := NewInMemoryPaymentRepository()

matchingStrategy := NewNearestDriverStrategy(50)
fareCalculator := NewBaseFareStrategy(2.0, 1.5, 0.25)
paymentProcessor := NewInMemoryPaymentProcessor(paymentRepo)
notifier := NewRideNotifier()

matchingService := NewMatchingService(driverRepo, matchingStrategy, 50)
rideService := NewRideService(rideRepo, driverRepo, riderRepo,
    matchingService, fareCalculator, paymentProcessor, ratingRepo, notifier)
```

## Phase 7: Extend & Discuss (8 min)

**Design patterns to mention:**
- **Strategy:** MatchingStrategy (nearest vs highest-rated), FareCalculator (base + surge)
- **State:** Ride status; validate transitions (can't complete without start)
- **Observer:** RideNotifier for status changes (push, analytics)
- **Repository:** Swap for PostGIS/Redis GEO for production

**Extensions:**
- GeoHash or quadtree for O(log n) "drivers near point" instead of linear scan
- Redis GEO for distributed geo queries
- Surge: time-of-day rules, ML-based dynamic pricing

## Tips

- **Prioritize if low on time:** RequestRide + FindDriver with Haversine; skip CompleteRide payment details.
- **Common mistakes:** Wrong Haversine formula; not filtering by radius or rating; forgetting to update driver status.
- **What impresses:** Correct Haversine, surge multiplier logic, Strategy pattern for matching/fare, thread-safe repos.
