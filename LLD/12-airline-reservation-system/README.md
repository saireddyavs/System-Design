# Airline Reservation System - Low Level Design

A production-quality, interview-ready Go implementation of an Airline Reservation System following Clean Architecture and SOLID principles.

## 1. Problem Description & Requirements

### Problem
Design a system that allows airlines to manage flights, seats, passengers, and bookings with support for search, seat assignment, and cancellation with refunds.

### Functional Requirements
- **Flight Management**: Add, update, cancel flights
- **Seat Management**: Economy, Business, First class with availability tracking
- **Passenger Management**: Create and manage passenger profiles
- **Booking**: Create bookings with automatic or manual seat assignment
- **Search**: Search flights by route, date, and class
- **Cancellation**: Cancel bookings with tiered refund policy
- **Scheduling**: Departure/arrival times, flight status
- **Baggage**: Allowance per class (Economy 23kg, Business 32kg, First 40kg)

### Non-Functional Requirements
- Thread-safe for concurrent bookings
- No double-booking of seats
- Flight capacity management

---

## 2. Core Entities & Relationships

```
┌─────────┐     has many    ┌──────┐
│ Flight  │────────────────▶│ Seat │
└─────────┘                 └──────┘
     │                           │
     │ 1:N                        │ N:1
     ▼                           ▼
┌─────────┐                 ┌──────────┐
│ Booking │◀────────────────│ Passenger│
└─────────┘                 └──────────┘
```

| Entity | Key Fields | Relationships |
|--------|------------|---------------|
| **Flight** | ID, FlightNumber, Origin, Destination (IATA codes), DepartureTime, ArrivalTime, Aircraft, Seats, Status | Has many Seats, has many Bookings |
| **Seat** | ID, FlightID, SeatNumber, Row, Column, Class, Status, Price | Belongs to Flight, referenced by Booking |
| **Passenger** | ID, Name, Email, Phone, PassportNumber, DateOfBirth | Has many Bookings |
| **Booking** | ID, PassengerID, FlightID, SeatIDs, TotalAmount, Status, BookingRef | Belongs to Passenger & Flight |

---

## 3. Seat Assignment Algorithm

Two strategies (Strategy Pattern):

1. **AutoAssignFirstAvailable**: Assigns first N available seats in order. O(n) scan.

2. **WindowPreferenceAssignment**: Partitions available seats into window (A, F) and others. Assigns window seats first, then fills with remaining.

**Flow**:
```
GetAvailableSeats(flightID, class) → Filter by status=Available, optional class
    ↓
Strategy.AssignSeats(seats, count, preferredClass) → Returns seat IDs
    ↓
MarkSeatsBooked(flightID, seatIDs)
```

---

## 4. Pricing Strategy

**Base Formula**: `Price = BasePrice × ClassMultiplier`

| Class | Multiplier | Baggage |
|-------|------------|---------|
| Economy | 1.0x | 23 kg |
| Business | 2.5x | 32 kg |
| First | 5.0x | 40 kg |

---

## 5. Refund Policy

| Hours Before Departure | Refund |
|------------------------|--------|
| > 48 hours | 100% |
| 24–48 hours | 75% |
| < 24 hours | 25% |

---

## 6. Design Patterns with WHY

| Pattern | Where | Why |
|---------|-------|-----|
| **Strategy** | Seat assignment (Auto/Window) | Different algorithms for seat selection; swappable at runtime without changing client code. |
| **Strategy** | Pricing (ClassMultiplier) | Class-based pricing; easy to add new strategies (e.g., seasonal, loyalty). |
| **Factory** | BookingFactory | Centralizes booking creation with ID/ref generation; encapsulates complex object construction. |
| **Observer** | BookingNotifier + EmailBookingObserver | Decouples booking events from notifications; add SMS, push, analytics without modifying core logic. |
| **Repository** | FlightRepository, BookingRepository | Abstracts data access; swap in-memory for PostgreSQL without changing services. |
| **Builder** | FlightBuilder | Complex flight+seats construction; fluent API for readability; avoids telescoping constructors. |

---

## 7. SOLID Principles Mapping

| Principle | Implementation |
|-----------|----------------|
| **S - Single Responsibility** | FlightService (flights), BookingService (bookings), SearchService (search), SeatService (seats). Each service has one reason to change. |
| **O - Open/Closed** | New seat strategies (e.g., FamilyTogether) or pricing strategies can be added without modifying existing code. |
| **L - Liskov Substitution** | Any SeatAssignmentStrategy or PricingStrategy can replace another; clients depend on interfaces. |
| **I - Interface Segregation** | FlightRepository, BookingRepository, PaymentProcessor are small, focused interfaces. No fat interfaces. |
| **D - Dependency Inversion** | Services depend on interfaces (FlightRepository, not InMemoryFlightRepository); high-level modules don't depend on low-level. |

---

## 8. Interview Explanations

### 3-Minute Summary
"We built an airline reservation system with clean architecture. Core entities are Flight, Seat, Passenger, and Booking. We use the Strategy pattern for seat assignment (auto, window, aisle preference) and pricing (class multiplier, demand-based). The Repository pattern abstracts data access for easy testing and DB swap. A Builder creates flights with seats; a Factory creates bookings. Observers notify on booking events. Cancellation uses a tiered refund policy. All repositories are thread-safe with RWMutex. SOLID is applied throughout—services depend on interfaces, and new strategies can be added without changing existing code."

### 10-Minute Deep Dive
1. **Architecture**: Clean architecture with models, interfaces, services, repositories, strategies. No framework lock-in.
2. **Concurrency**: In-memory repos use sync.RWMutex. Booking flow: validate → assign seats → process payment → create booking → mark seats. Rollback on failure.
3. **Double-booking prevention**: Seat status (Available/Booked/Blocked). ManualAssignSeats validates; MarkSeatsBooked is called only after successful payment and booking creation.
4. **Extensibility**: New seat strategy? Implement SeatAssignmentStrategy. New pricing? Implement PricingStrategy. New notification channel? Implement BookingObserver.
5. **Testing**: Unit tests for booking (create, cancel, double-book prevention, concurrent), search (route, class, cancelled exclusion), seat (auto/manual assign, strategies), pricing, baggage.

---

## 9. Future Improvements

- **Database**: Replace in-memory repos with PostgreSQL; add transactions for booking flow.
- **Distributed locking**: Redis-based locks for seat assignment in multi-instance deployment.
- **Event sourcing**: Store booking events for audit and replay.
- **Idempotency**: Idempotency keys for payment and booking to handle retries.
- **Caching**: Cache flight search results with TTL.
- **API layer**: REST/gRPC with validation, rate limiting.
- **Saga pattern**: For cross-service booking (flights + hotels).

---

## 10. Running the Project

```bash
# Build
go build -o airline ./cmd/main.go

# Run
./airline

# Tests
go test ./tests/... -v
```

---

## Directory Structure

```
12-airline-reservation-system/
├── cmd/main.go              # Entry point, wire-up, demo
├── internal/
│   ├── models/              # Entities
│   ├── interfaces/          # Contracts (repositories, strategies)
│   ├── services/            # Business logic
│   ├── repositories/        # In-memory implementations
│   └── strategies/          # Seat assignment, pricing, payment mock
├── tests/                   # Unit tests
├── go.mod
└── README.md
```
