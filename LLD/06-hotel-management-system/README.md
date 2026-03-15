# Hotel Management System - Low Level Design

A production-quality, interview-ready Go implementation of a Hotel Management System following Clean Architecture and SOLID principles.

## Problem Description

Design a system to manage hotel operations including:
- **Room Management**: Multiple room types (Single, Double, Deluxe, Suite) with different base prices
- **Guest Management**: Registration, check-in, check-out
- **Booking Lifecycle**: Reservation → Confirmation → Check-in → Check-out
- **Availability**: Prevent overbooking with date-range overlap checks
- **Payments**: Process payments and refunds
- **Cancellation**: Refund policy based on timing (>24h: full, <24h: 50%, no-show: 0%)
- **Search**: Find available rooms by type, date range, price range

## Requirements

| Requirement | Implementation |
|-------------|----------------|
| Room types with prices | Single ($100), Double ($150), Deluxe ($250), Suite ($400) |
| Guest management | Register, loyalty points, tier-based discounts |
| Booking flow | Pending → Confirmed (payment) → CheckedIn → CheckedOut |
| Overbooking prevention | `sync.RWMutex` + overlap check in `GetBookingsForRoomInRange` |
| Payment processing | `PaymentProcessor` interface, mock implementation |
| Refund policy | Configurable via `RefundFullHours`, `RefundPartialPct` |
| Search | `SearchCriteria` with type, dates, min/max price |

## Core Entities & Relationships

```
┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐
│  Room   │────<│ Booking │>────│  Guest  │     │ Payment │
└─────────┘     └────┬────┘     └─────────┘     └────┬────┘
     │               │                                  │
     │               └──────────────────────────────────┘
     │
     ├── RoomType: Single | Double | Deluxe | Suite
     ├── RoomStatus: Available | Occupied | Reserved | Maintenance
     └── BasePrice, Amenities

Booking States (State Machine):
  Pending → Confirmed → CheckedIn → CheckedOut
       ↘ Cancelled
```

### Entity Details

| Entity | Key Fields |
|--------|-------------|
| **Room** | ID, Number, Type, Floor, BasePrice, Status, Amenities |
| **Guest** | ID, Name, Email, Phone, IDProof, LoyaltyPoints |
| **Booking** | ID, GuestID, RoomID, CheckIn/Out, Status, TotalAmount, PaymentStatus |
| **Payment** | ID, BookingID, Amount, Method, Status, TransactionID, PaidAt |

## Design Patterns

### 1. Strategy Pattern – Pricing
**Location**: `internal/strategies/pricing_strategy.go`, `internal/interfaces/pricing_strategy.go`

**Why**: Price calculation varies by season, weekday/weekend, and loyalty tier. Strategy allows adding new rules (e.g., holiday premium) without changing existing code.

**Implementation**: `PricingStrategy` interface with `CalculatePrice(ctx)`. Chain: Base → Seasonal → Weekday/Weekend → Loyalty Discount.

### 2. State Pattern – Booking Status
**Location**: `models/booking.go`, `services/booking_service.go`

**Why**: Booking has strict lifecycle (Pending → Confirmed → CheckedIn → CheckedOut). Invalid transitions (e.g., CheckOut without CheckIn) must be rejected.

**Implementation**: `BookingStatus` enum; transitions validated in `BookingService` methods.

### 3. Factory Pattern – Room Creation
**Location**: `internal/factory/room_factory.go`

**Why**: Room creation depends on type (amenities, base price). Factory centralizes this logic.

**Implementation**: `RoomFactory.CreateRoom(id, number, type, floor)` returns configured `Room`.

### 4. Observer Pattern – Notifications
**Location**: `internal/observer/notification_service.go`, `internal/interfaces/notification_service.go`

**Why**: Multiple subscribers (email, SMS, analytics) need to react to booking events without coupling.

**Implementation**: `NotificationService.Subscribe(handler)`; `Notify(payload)` broadcasts to all handlers.

### 5. Repository Pattern – Data Access
**Location**: `internal/interfaces/*_repository.go`, `internal/repositories/*.go`

**Why**: Decouples business logic from storage. Swap in-memory for PostgreSQL without changing services.

**Implementation**: Interfaces `RoomRepository`, `BookingRepository`, etc.; in-memory implementations with `sync.RWMutex`.

## Data Structures & Algorithms

| DS/Algorithm | Where Used | Why | Alternatives/Tradeoffs |
|--------------|------------|-----|------------------------|
| **HashMap** | All repositories (`InMemoryRoomRepository`, `InMemoryBookingRepository`, `InMemoryGuestRepository`, `InMemoryPaymentRepository`) | O(1) lookup by ID; fast Create/GetByID/Update | B-tree for range queries; skip-list for ordered access |
| **Date range overlap detection** | `BookingRepository.GetBookingsForRoomInRange`, `BookingService.CreateBooking` | Prevent overbooking: two ranges overlap iff `(StartA < EndB) && (EndA > StartB)` | Interval tree for O(log n) queries over many bookings; current linear scan is O(n) |
| **Composite pricing strategy** | `strategies/pricing_strategy.go`: Base → Seasonal → Weekday/Weekend → Loyalty | Chain of multipliers: seasonal (peak/off-peak), weekend premium, loyalty discount | Rule engine; lookup table for fixed combinations |
| **sync.RWMutex** | All repositories, `BookingService` | Thread-safe concurrent access; RLock for reads, Lock for writes | Distributed lock (Redis) for multi-instance; optimistic locking |

## SOLID Principles Mapping

| Principle | Application |
|-----------|-------------|
| **S**ingle Responsibility | Each service handles one domain (RoomService, BookingService, GuestService, PaymentService) |
| **O**pen/Closed | New pricing strategies, room types, notification channels without modifying core |
| **L**iskov Substitution | All `PricingStrategy` implementations interchangeable; all repositories swappable |
| **I**nterface Segregation | Focused interfaces: `PaymentProcessor`, `NotificationService`, `PricingStrategy` |
| **D**ependency Inversion | Services depend on `interfaces.RoomRepository`, not `repositories.InMemoryRoomRepository` |

## Booking State Machine

```
                    ┌──────────────┐
                    │   Pending    │
                    └──────┬───────┘
                           │ ConfirmBooking (payment)
                           ▼
                    ┌──────────────┐
                    │  Confirmed   │
                    └──────┬───────┘
              ┌────────────┼────────────┐
              │ CheckIn    │            │ CancelBooking
              ▼            │            ▼
       ┌──────────────┐    │     ┌──────────────┐
       │  CheckedIn   │    │     │  Cancelled   │
       └──────┬───────┘    │     └──────────────┘
              │ CheckOut   │
              ▼            │
       ┌──────────────┐    │
       │ CheckedOut   │    │
       └──────────────┘    │
              ▲            │
              └────────────┘
         (Cancel from Pending)
```

**Valid Transitions**:
- Pending → Confirmed (payment required)
- Pending → Cancelled
- Confirmed → CheckedIn
- Confirmed → Cancelled
- CheckedIn → CheckedOut

## Interview Explanations

### 3-Minute Summary

> "I designed a Hotel Management System in Go with Clean Architecture. Core entities are Room, Guest, Booking, and Payment. The booking lifecycle uses a state machine: Pending → Confirmed → CheckedOut, with payment required before confirmation. I used the Strategy pattern for pricing—seasonal, weekday/weekend, and loyalty discounts—and the Repository pattern for data access. Overbooking is prevented with date-range overlap checks and mutex-based locking. Cancellation follows a refund policy: full refund if >24h before check-in, 50% if <24h, and no refund for no-shows. The Observer pattern handles notifications for events like confirmation and check-out."

### 10-Minute Deep Dive

1. **Architecture**: Clean Architecture with `models`, `interfaces`, `services`, `repositories`, `strategies`, `factory`, and `observer` packages. Services orchestrate; repositories abstract storage.

2. **Overbooking Prevention**: `BookingService.CreateBooking` acquires a lock, checks `GetBookingsForRoomInRange` for overlaps, and rejects if any exist. Concurrency tests verify only one of 10 simultaneous bookings for the same room succeeds.

3. **Pricing**: Composite strategy: Base (room × nights) → Seasonal (peak/off-peak multiplier) → Weekday/Weekend (weekend premium) → Loyalty (tier-based discount). Each strategy is independently testable.

4. **State Machine**: `BookingStatus` enum; each transition method validates current state. Invalid transitions return `ErrInvalidStateTransition`.

5. **Thread Safety**: Repositories use `sync.RWMutex`; models use per-entity mutex for field updates. `BookingService` uses a service-level mutex for create/confirm to prevent races.

6. **Testing**: Unit tests cover booking flow, cancellation refund, overbooking prevention, invalid transitions, pricing, room search, and concurrent access.

## Future Improvements

| Area | Improvement |
|------|-------------|
| **Persistence** | Replace in-memory repos with PostgreSQL; add migrations |
| **API** | REST/gRPC layer with validation, rate limiting |
| **Idempotency** | Idempotency keys for payment and booking creation |
| **Events** | Event sourcing or message queue (Kafka) for audit trail |
| **Caching** | Redis for availability queries |
| **Distributed Lock** | Redis/ZooKeeper for multi-instance overbooking prevention |
| **Observability** | Structured logging, metrics, tracing |

## Running the Project

```bash
# Build
go build ./cmd/...

# Run
go run ./cmd/main.go

# Test
go test ./tests/... -v
```

## Directory Structure

```
06-hotel-management-system/
├── cmd/main.go                 # Entry point
├── internal/
│   ├── models/                 # Entities
│   ├── interfaces/             # Abstractions
│   ├── services/               # Business logic
│   ├── repositories/           # In-memory implementations
│   ├── strategies/             # Pricing strategies
│   ├── factory/                # Room factory
│   ├── observer/               # Notification service
│   └── payment/                # Mock payment processor
├── tests/                      # Unit tests
├── go.mod
└── README.md
```
