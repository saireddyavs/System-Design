# Movie Ticket Booking System - Low Level Design

A production-quality, interview-ready Go implementation of a Movie Ticket Booking System following Clean Architecture and SOLID principles.

## 1. Problem Description

Design and implement a system that allows users to:
- Browse movies by title, genre, and city
- View available shows at theatres
- Book tickets with seat selection
- Cancel bookings with refund policy
- Handle concurrent bookings without double-booking

## 2. Requirements

| Requirement | Implementation |
|-------------|----------------|
| Movies (title, genre, duration, rating) | `models.Movie` with Genre enum |
| Theatres with multiple screens | `models.Theatre`, `models.Screen` |
| Show scheduling | `models.Show` (movie + screen + time) |
| Seat map (Regular, Premium, VIP) | `models.Seat` with `SeatCategory` |
| Ticket booking with seat selection | `BookingService.CreateBooking` |
| Cancellation with refund policy | Full refund >24h, 50% otherwise |
| Search by title, genre, city | `SearchService` |
| Concurrent seat booking (no double booking) | Pessimistic locking via `UpdateSeats` |

## 3. Core Entities & Relationships

```
User 1----* Booking
Booking *----1 Show
Show *----1 Movie
Show *----1 Screen
Screen *----1 Theatre
Screen 1----* Seat
```

| Entity | Key Fields |
|--------|------------|
| **Movie** | ID, Title, Genre, Duration, Rating, Language |
| **Theatre** | ID, Name, City, Address |
| **Screen** | ID, TheatreID, Name, TotalCapacity, Seats[] |
| **Seat** | ID, ScreenID, Row, Number, Category (Regular/Premium/VIP) |
| **Show** | ID, MovieID, ScreenID, TheatreID, StartTime, EndTime, SeatStatusMap |
| **Booking** | ID, UserID, ShowID, SeatIDs[], TotalAmount, Status, BookedAt |
| **User** | ID, Name, Email, Phone |

## 4. Design Patterns

### Strategy Pattern (Pricing)
- **Location**: `internal/strategies/pricing_strategy.go`
- **Purpose**: Open/Closed Principle - add new pricing rules without modifying existing code
- **Implementations**: `WeekdayPricingStrategy` (weekend multiplier + seat category), `FlatPricingStrategy`

### Factory Pattern (Booking Creation)
- **Location**: `internal/services/booking_service.go` - `createBooking()`
- **Purpose**: Centralized creation of booking objects with consistent initialization

### Observer Pattern (Notifications)
- **Location**: `internal/services/notification_service.go`
- **Purpose**: Decouple booking/cancellation events from notification delivery
- **Flow**: `BookingService` → `NotificationService.Notify()` → all subscribed handlers

### Repository Pattern
- **Location**: `internal/interfaces/*_repository.go`, `internal/repositories/`
- **Purpose**: Abstract data access; swap in-memory for SQL/NoSQL without changing services

### Builder Pattern (Show Creation)
- **Location**: `internal/services/show_service.go` - `ShowBuilder`
- **Purpose**: Fluent API for constructing complex Show objects with validation

## 5. SOLID Principles Mapping

| Principle | Application |
|-----------|-------------|
| **S**ingle Responsibility | `MovieService` (movies only), `BookingService` (bookings only), `SearchService` (search only) |
| **O**pen/Closed | `PricingStrategy` interface - extend with new strategies without modifying `BookingService` |
| **L**iskov Substitution | Any `PricingStrategy` implementation works in `BookingService` |
| **I**nterface Segregation | `PaymentProcessor`, `NotificationService`, `PricingStrategy` - small, focused interfaces |
| **D**ependency Inversion | Services depend on `interfaces.*`, not concrete `repositories.*` |

## 6. Concurrency Handling

### Problem
Multiple users booking the same seat simultaneously must not both succeed.

### Solution: Pessimistic Locking

1. **Per-Show Mutex**: Each show has its own `sync.Mutex` in `InMemoryShowRepository.showLocks`
2. **Atomic Update**: `ShowRepository.UpdateSeats(showID, updateFn)` acquires the show lock, runs the callback, then releases
3. **Flow**:
   - User A and B both request seat-1
   - A acquires lock → checks availability → marks booked → releases
   - B acquires lock → checks availability → seat-1 is booked → returns `ErrSeatNotAvailable`

### Key Code
```go
// InMemoryShowRepository.UpdateSeats
lock := r.getShowLock(showID)
lock.Lock()
defer lock.Unlock()
// ... atomic read-modify-write of SeatStatusMap
```

### Why Pessimistic over Optimistic?
- Seat inventory is highly contended (popular shows)
- Conflict rate is high → optimistic would cause many retries
- Lock scope is small (single show) → minimal contention across shows

## 7. Interview Explanations

### 3-Minute Pitch
"We built a movie ticket system with clean architecture. Core entities are Movie, Theatre, Screen, Show, Seat, and Booking. We use the Repository pattern for data access, Strategy for pricing (weekday vs weekend, seat category), and Observer for notifications. The critical part is concurrency: we use per-show pessimistic locking so that when two users try to book the same seat, only one succeeds. Refund policy is full refund if cancelled >24h before show, 50% otherwise."

### 10-Minute Deep Dive
1. **Architecture**: Clean separation - models, interfaces, services, repositories. All dependencies flow inward via interfaces.
2. **Concurrency**: `UpdateSeats` encapsulates the lock. The booking service never touches a mutex directly. We use a map of per-show mutexes so different shows can be booked in parallel.
3. **Pricing**: `PricingStrategy` interface with `CalculatePrice(base, category, time)`. Weekday strategy applies 1.25x on weekends and category multipliers (1x Regular, 1.5x Premium, 2x VIP).
4. **Search**: `SearchService` composes MovieRepository, TheatreRepository, ShowRepository to support "movies in city X" and "movies by genre".
5. **Testing**: Unit tests cover booking success, double-booking rejection, cancellation refund, and concurrent same-seat (exactly 1 succeeds) and different-seat (both succeed) scenarios.

## 8. Future Improvements

| Area | Improvement |
|------|-------------|
| **Persistence** | Replace in-memory repos with PostgreSQL (transactions for booking) |
| **Caching** | Redis for hot shows, reduce DB load |
| **Idempotency** | Idempotency keys for payment retries |
| **Time-limited holds** | 10-min seat hold before payment (with TTL/cleanup) |
| **API** | REST/gRPC layer with validation middleware |
| **Observability** | Structured logging, metrics (booking latency, failure rate) |
| **Refund policy** | Configurable rules (e.g., no refund <2h before show) |

## 9. Running the Project

```bash
# Build
go build ./...

# Run demo
go run ./cmd

# Run tests
go test ./tests/... -v
```

## 10. Directory Structure

```
04-movie-ticket-booking/
├── cmd/main.go              # Entry point, DI wiring
├── internal/
│   ├── models/              # Domain entities
│   ├── interfaces/          # Repository & service contracts
│   ├── repositories/        # In-memory implementations
│   ├── services/            # Business logic
│   └── strategies/          # Pricing strategies
├── tests/                   # Unit tests
├── go.mod
└── README.md
```
