# Parking Lot System - Low Level Design

A production-quality, interview-ready implementation of a multi-level parking lot system in Go. Built with clean architecture, SOLID principles, and common design patterns.

---

## 1. Problem Description

Design and implement a parking lot system that:

- Manages a **multi-level** parking facility with different spot sizes
- Supports **multiple vehicle types** (Motorcycle, Car, Bus/Truck)
- Handles **park** and **unpark** operations
- Tracks **available spaces** in real-time
- Generates **parking tickets** on entry
- Calculates **parking fees** based on duration
- Operates safely under **concurrent access**

---

## 2. Functional Requirements

| Requirement | Description |
|-------------|-------------|
| Park Vehicle | Find suitable spot, assign vehicle, issue ticket |
| Unpark Vehicle | Locate by ticket ID or license plate, free spot, return vehicle |
| Spot Compatibility | Small spots: Motorcycle only; Medium: Motorcycle, Car; Large: All |
| Fee Calculation | Compute fee based on vehicle type and duration |
| Availability | Query available spots per vehicle type |
| Ticket Management | Unique ticket per park, lookup by ID or license |

## 3. Non-Functional Requirements

| Requirement | Implementation |
|-------------|----------------|
| Thread Safety | `sync.RWMutex` on ParkingLot, ParkingLevel, ParkingSpot, ParkingService |
| Extensibility | Strategy pattern for spot allocation and fee calculation |
| Testability | Dependency injection via interfaces |
| Performance | RWMutex for read-heavy operations (availability checks) |

---

## 4. Core Entities & Relationships

```
┌─────────────────┐     has many    ┌──────────────────┐     has many    ┌───────────────┐
│   ParkingLot    │────────────────▶│  ParkingLevel    │────────────────▶│ ParkingSpot   │
│   (Singleton)   │                 │                  │                 │               │
└─────────────────┘                 └──────────────────┘                 └───────┬───────┘
        │                                    │                                    │
        │                                    │                                    │ holds
        │                                    │                                    ▼
        │                                    │                            ┌───────────────┐
        │                                    │                            │   Vehicle     │
        │                                    │                            │ (Motorcycle,  │
        │                                    │                            │  Car, Bus)    │
        │                                    │                            └───────────────┘
        │                                    │
        │ uses                               │ generates
        ▼                                    ▼
┌─────────────────┐                 ┌──────────────────┐
│ ParkingService  │                 │     Ticket       │
│ (park + fees)   │                 │                  │
└─────────────────┘                 └──────────────────┘
```

### Entity Summary

| Entity | Responsibility |
|--------|----------------|
| **Vehicle** | Interface; types define required spot size |
| **ParkingSpot** | Size, occupancy, park/unpark |
| **ParkingLevel** | Collection of spots, availability queries |
| **ParkingLot** | Singleton; manages levels |
| **Ticket** | Links vehicle, spot, entry time |
| **ParkingService** | Orchestrates park/unpark, ticket management, fee calculation |

---

## 5. Design Patterns Used

### 5.1 Singleton
**Where:** `ParkingLot`  
**Why:** Single global instance for the entire parking facility.  
**How:** `sync.Once` for thread-safe lazy initialization.

```go
lot := models.GetInstance()
```

### 5.2 Strategy
**Where:** Spot allocation (`ParkingStrategy`), Fee calculation (`HourlyFeeStrategy`)  
**Why:** Swap algorithms without changing clients.  
**How:** `ParkingStrategy` interface; `HourlyFeeStrategy` for fees (injected into ParkingService).

```go
ps := services.NewParkingService(lot, strategies.NewNearestSpotStrategy(), strategies.NewHourlyFeeStrategy())
```

### 5.3 Factory
**Where:** `NewVehicle(vehicleType, licensePlate)`  
**Why:** Encapsulate vehicle creation; add new types without changing callers.  
**How:** Function returning `Vehicle` interface based on type.

```go
car := models.NewVehicle(models.VehicleTypeCar, "CAR-001")
```

### 5.4 Dependency Injection
**Where:** `ParkingService`  
**Why:** Testability and flexibility.  
**How:** Constructor accepts `ParkingStrategy` interface and `HourlyFeeStrategy`.

---

## 6. SOLID Principles Mapping

| Principle | Application |
|-----------|-------------|
| **SRP** | `ParkingSpot` manages spot state; `ParkingService` orchestrates park/unpark and fee calculation |
| **OCP** | New vehicle types via `NewVehicle`; new strategies via `ParkingStrategy` interface |
| **LSP** | All vehicles implement `Vehicle`; all strategies implement their interfaces |
| **ISP** | `ParkingStrategy` has only `FindSpot` |
| **DIP** | ParkingService depends on `ParkingStrategy` interface, not concrete types |

---

## 7. Directory Structure

```
01-parking-lot-system/
├── cmd/
│   └── main.go              # Demo entry point
├── internal/
│   ├── models/
│   │   ├── vehicle.go       # Vehicle interface, types, factory
│   │   ├── parking_spot.go # Spot model
│   │   ├── parking_level.go # Level model
│   │   ├── parking_lot.go  # Singleton lot
│   │   └── ticket.go       # Ticket model
│   ├── interfaces/
│   │   └── parking_strategy.go
│   ├── services/
│   │   └── parking_service.go
│   └── strategies/
│       ├── nearest_spot_strategy.go
│       └── hourly_fee_strategy.go
├── tests/
│   ├── parking_service_test.go
│   ├── vehicle_test.go
│   └── nearest_spot_strategy_test.go
├── go.mod
└── README.md
```

---

## 8. Data Structures & Algorithms

| DS/Algorithm | Where Used | Why | Alternatives/Tradeoffs |
|-------------|------------|-----|------------------------|
| **HashMap** | `ParkingService.tickets` (map[string]*Ticket), `ParkingLot.levelID` | O(1) ticket lookup by ID; O(1) level lookup by ID; Unpark supports license plate via linear scan over tickets | License→ticket could use separate map for O(1) license lookup; tradeoff: extra memory vs faster unpark |
| **Linear scan** | `NearestSpotStrategy.FindSpot()` | Iterates spots in order to find first available; simple, predictable (nearest-to-entrance) | Could use min-heap by distance; for small spot counts, linear is sufficient |
| **sync.RWMutex** | ParkingLot, ParkingLevel, ParkingSpot, ParkingService | Read-heavy (availability checks) use RLock; park/unpark use Lock; prevents race conditions | Coarse-grained; fine-grained per-spot locks could improve parallelism but add complexity |
| **Singleton** | `ParkingLot.GetInstance()` | Single global lot; `sync.Once` for thread-safe lazy init | Alternative: dependency injection of lot instance; Singleton simplifies demo, DI better for tests |

---

## 9. How to Run

### Prerequisites
- Go 1.21+

### Build
```bash
go build ./cmd/...
```

### Run Demo
```bash
go run ./cmd/main.go
```

### Run Tests
```bash
go test ./tests/... -v
```

### Run All Tests (including coverage)
```bash
go test ./tests/... -cover
```

---

## 10. Concurrency Considerations

| Component | Mechanism | Rationale |
|-----------|-----------|-----------|
| ParkingLot | `sync.RWMutex` | Multiple readers (availability), single writer (init) |
| ParkingLevel | `sync.RWMutex` | Protects spot map |
| ParkingSpot | `sync.RWMutex` | Per-spot lock for park/unpark |
| ParkingService | `sync.RWMutex` | Protects ticket map and park/unpark |
| Singleton | `sync.Once` | Thread-safe lazy init |

**Lock Ordering:** Services → Lot → Level → Spot (top-down to avoid deadlock)

**Read vs Write:** Availability checks use `RLock`; park/unpark use `Lock`.

---

## 11. Future Improvements

1. **Observer Pattern** – Notify displays/APIs when availability changes
2. **Persistence** – Store tickets and state in DB
3. **REST API** – HTTP endpoints for park/unpark/status
4. **Reservation** – Pre-book spots for a time window
5. **Tiered Fees** – Peak/off-peak, first-hour discount
6. **Metrics** – Prometheus metrics for occupancy, revenue
7. **Config** – Load levels/spots from config file

---

## 12. Interview Explanation

### 3-Minute Version

> "I designed a multi-level parking lot with spot sizes (Small, Medium, Large) and vehicle types (Motorcycle, Car, Bus). Spots can hold vehicles of their size or smaller.
>
> I used a **Singleton** for the ParkingLot, **Strategy** for spot allocation (e.g., nearest-first) and fee calculation (HourlyFeeStrategy), and a **Factory** for creating vehicles. ParkingService handles park/unpark and fee calculation.
>
> The system is **thread-safe** with RWMutex. Park finds a spot via the strategy, assigns the vehicle, and returns a ticket. Unpark accepts ticket ID or license plate, frees the spot, and returns the vehicle. Fees are calculated by ParkingService.CalculateFee using the injected HourlyFeeStrategy."

### 10-Minute Version

> "**Problem:** Multi-level parking lot with different spot sizes and vehicle types. Need park, unpark, tickets, and fee calculation. Must be thread-safe.
>
> **Design:**
> - **Models:** Vehicle (interface with Motorcycle, Car, Bus), ParkingSpot (size + occupancy), ParkingLevel (collection of spots), ParkingLot (singleton), Ticket.
> - **Spot compatibility:** Small=MC only, Medium=MC+Car, Large=all. A spot holds vehicles of its size or smaller.
>
> **Patterns:**
> - **Singleton** for ParkingLot with sync.Once.
> - **Strategy** for ParkingStrategy (FindSpot) and HourlyFeeStrategy (Calculate). ParkingService handles both park/unpark and fee calculation.
> - **Factory** for NewVehicle(type, plate).
> - **DIP:** ParkingService takes ParkingStrategy interface and HourlyFeeStrategy.
>
> **SOLID:** SRP per component, OCP for new vehicles/strategies, LSP for Vehicle implementations, ISP for small interfaces, DIP for services.
>
> **Concurrency:** RWMutex on lot, level, spot, and service. Read locks for availability, write locks for park/unpark.
>
> **Flow:** Park → strategy finds spot → spot.Park() → create ticket. Unpark → lookup ticket → spot.Unpark() → return vehicle. ParkingService.CalculateFee(ticket) uses duration and vehicle type."
