# Elevator System - Low Level Design

A production-quality, interview-ready elevator control system implementation in Go. Implements SCAN/LOOK scheduling, handles concurrent requests, and demonstrates SOLID principles with clean architecture.

## Problem Description

Design an elevator control system for a building with:
- **N floors** (0 to N-1)
- **M elevators** operating concurrently
- **External requests**: Floor buttons (up/down) pressed by passengers waiting
- **Internal requests**: Destination floor selected inside elevator
- **Scheduling**: SCAN/LOOK algorithm for efficient request handling
- **Edge cases**: Door open/close, overweight, emergency stop, peak periods

## Requirements

| Requirement | Implementation |
|-------------|----------------|
| N floors, M elevators | `Building` struct with configurable floors and elevator count |
| External requests | `NewExternalRequest(sourceFloor, direction)` |
| Internal requests | `NewInternalRequest(sourceFloor, destFloor)` |
| SCAN/LOOK algorithm | `ScanStrategy`, `LookStrategy` (Strategy pattern) |
| Concurrent requests | Goroutines + channels + mutex |
| Status display | `GetStatus()` returns all elevator states |
| Door open/close | `StateDoorOpen` + configurable duration |
| Overweight | `OverweightThreshold` (90% capacity), `CanAcceptPassenger()` |
| Emergency stop | `StateEmergencyStop`, `EmergencyStop()` |
| Peak period | `SetPeakMode(enabled)` |

## Core Entities & Relationships

```
┌─────────────┐     contains      ┌─────────────┐
│  Building   │──────────────────│  Elevator   │
│ - ID       │                   │ - ID       │
│ - Floors   │                   │ - Floor    │
│ - Elevators│                   │ - State    │
└─────────────┘                   │ - Queue    │
       │                          └─────┬─────┘
       │                                │
       │ dispatches                     │ processes
       ▼                                ▼
┌─────────────┐                   ┌─────────────┐
│ Dispatcher  │                   │  Request    │
│ - Strategy  │◄──────────────────│ - Type      │
│ - Channels  │    assigns       │ - Source    │
└─────────────┘                   │ - Dest      │
                                 └─────────────┘
```

### Entity Details

**Building**: ID, Name, TotalFloors, Elevators[]  
**Elevator**: ID, CurrentFloor, Direction, State, Capacity, CurrentLoad, RequestQueue  
**Request**: ID, Type (External/Internal), SourceFloor, DestFloor, Direction, Timestamp  
**Direction**: Up, Down, Idle  
**ElevatorState**: Idle, MovingUp, MovingDown, DoorOpen, Maintenance, EmergencyStop  

## Scheduling Algorithm: SCAN vs LOOK

### SCAN (Elevator Algorithm)

```
Direction: UP
Floors:    0  1  2  3  4  5  6  7  8  9  10
           │  │  │  │  │  │  │  │  │  │  │
Request at:       ●        ●        ●
Current:    *

Path: * → 2 → 2 → 5 → 5 → 8 → 8 → 10 (end) → REVERSE → 9 → 8 → ...
       ↑ pickup  ↑ pickup  ↑ pickup  ↑ goes to end
```

- Continues in current direction until **end of building**
- Reverses at top/bottom floor
- Serves all requests along the way

### LOOK Algorithm

```
Direction: UP
Floors:    0  1  2  3  4  5  6  7  8  9  10
Request at:       ●        ●        ●
Current:    *

Path: * → 2 → 5 → 8 → REVERSE (no more requests ahead) → ...
       ↑ stops at highest request, doesn't go to floor 10
```

- Reverses when **no more requests** in current direction
- More efficient than SCAN (doesn't go to end unnecessarily)

### Example Flow

1. **External request** at floor 3 (UP): Dispatcher assigns to nearest elevator moving toward floor 3
2. **Elevator arrives** at floor 3: Opens door, passenger boards (simulated 70kg)
3. **Internal request** at floor 8: Added to same elevator's queue
4. **LOOK** orders queue: Visit 3 (pickup) → 8 (dropoff)
5. **Elevator** moves up, stops at 8, door opens, passenger exits

## Design Patterns

| Pattern | Where | Why |
|---------|-------|-----|
| **Strategy** | `SchedulingStrategy` interface, `ScanStrategy`, `LookStrategy` | Interchangeable scheduling algorithms without modifying dispatcher. OCP: new strategies without changing existing code. |
| **State** | `ElevatorState` (Idle, MovingUp, MovingDown, DoorOpen, Maintenance, EmergencyStop) | Each state encapsulates behavior. State transitions are explicit. |
| **Command** | `Request` struct | Encapsulates request as object. Can be queued, logged, undone. |
| **Singleton** | `GetDispatcher()` | Single dispatcher per building. Ensures consistent request routing. |
| **Facade** | `BuildingController` | Simplifies complex subsystem (dispatcher + elevators) behind simple API. |

## SOLID Principles

| Principle | Implementation |
|-----------|----------------|
| **S**ingle Responsibility | `Elevator` = state, `Dispatcher` = routing, `Strategy` = ordering |
| **O**pen/Closed | New scheduling strategy = new type implementing interface, no changes to dispatcher |
| **L**iskov Substitution | Any `SchedulingStrategy` implementation works in `ElevatorService` |
| **I**nterface Segregation | `SchedulingStrategy` (order), `ElevatorController` (API) |
| **D**ependency Inversion | `ElevatorService` depends on `SchedulingStrategy` interface, not concrete SCAN/LOOK |

## Concurrency Model

```
                    ┌─────────────────┐
                    │   Request API    │
                    └────────┬─────────┘
                             │
                    ┌────────▼─────────┐
                    │   Dispatcher     │
                    │  (requestCh)     │
                    └────────┬─────────┘
                             │
                    ┌────────▼─────────┐
                    │   assignRequest  │
                    │  selectElevator  │
                    └────────┬─────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
   ┌────▼────┐          ┌────▼────┐          ┌────▼────┐
   │Elevator 1│          │Elevator 2│          │Elevator 3│
   │(goroutine)│          │(goroutine)│          │(goroutine)│
   │requestCh  │          │requestCh  │          │requestCh  │
   └──────────┘          └──────────┘          └──────────┘
```

- **Channels**: `requestCh` for dispatcher → elevator communication
- **Mutex**: `sync.RWMutex` on Elevator, Building, Dispatcher for shared state
- **Goroutines**: One per elevator (run loop), one for dispatcher
- **Worker pool**: Dispatcher processes requests sequentially, assigns to elevator workers

## Directory Structure

```
05-elevator-system/
├── cmd/main.go              # Demo entry point
├── internal/
│   ├── models/              # Domain entities
│   │   ├── elevator.go
│   │   ├── building.go
│   │   ├── request.go
│   │   └── enums.go
│   ├── interfaces/          # Contract definitions
│   │   ├── scheduling_strategy.go
│   │   └── elevator_controller.go
│   ├── services/            # Business logic
│   │   ├── elevator_service.go
│   │   ├── dispatcher_service.go
│   │   └── building_controller.go
│   └── strategies/          # Scheduling algorithms
│       ├── scan_strategy.go
│       └── look_strategy.go
├── tests/
│   ├── elevator_service_test.go
│   └── dispatcher_test.go
├── go.mod
└── README.md
```

## Usage

```go
// Create building: 10 floors, 2 elevators
building := models.NewBuilding("B1", "Main Tower", 10, 2)
strategy := strategies.NewLookStrategy()

// Create controller
ctrl := services.NewBuildingController(building, strategy)
defer ctrl.Stop()

// External request (floor button)
req := models.NewExternalRequest(3, models.DirectionUp)
ctrl.SubmitRequest(req)

// Internal request (destination floor)
req2 := models.NewInternalRequest(3, 8)
ctrl.SubmitRequest(req2)

// Status
for _, s := range ctrl.GetStatus() {
    fmt.Printf("Elevator %s: Floor %d, %s\n", s.ID, s.CurrentFloor, s.State)
}

// Emergency
ctrl.EmergencyStop("B1-E0")
ctrl.ResumeElevator("B1-E0")
```

## Run

```bash
go run ./cmd/main.go
go test ./tests/... -v
```

## Interview Explanations

### 3-Minute Summary

"We have a building with N floors and M elevators. Each elevator runs as a goroutine with its own request queue. The dispatcher receives external requests (floor buttons) and assigns them to the nearest suitable elevator using a scoring function: prefer elevators moving toward the request.

Inside each elevator, we use the LOOK algorithm: continue in current direction, serve requests along the way, reverse when no more requests ahead. Requests are ordered by the strategy's floor sequence.

We use the Strategy pattern for SCAN/LOOK, State pattern for elevator states (Idle, MovingUp, DoorOpen, EmergencyStop), and channels for communication. All shared state is protected by mutexes for thread safety."

### 10-Minute Deep Dive

**1. Request Flow**: External request → Dispatcher → selectElevator (nearest, same direction, not overweight) → elevator.requestCh → elevator queue → strategy.OrderRequests → processNextStep loop.

**2. SCAN vs LOOK**: SCAN goes to end of building before reversing; LOOK reverses when no more requests ahead. LOOK is more efficient. Both produce an ordered floor sequence via `ScanStrategy` and `LookStrategy`; we map requests to that sequence.

**3. Concurrency**: Dispatcher has one goroutine reading from requestCh. Each elevator has one goroutine with a ticker (100ms) calling processNextStep. Mutex protects shared state. No shared mutable state between elevators.

**4. Edge Cases**: Overweight: CanAcceptPassenger checks before adding load. Emergency: SetState(EmergencyStop) stops elevator loop. Door: StateDoorOpen for configurable duration.

**5. Peak Mode**: SetPeakMode enables future optimizations (e.g., pre-position elevators, different assignment strategy).

## Future Improvements

1. **Pre-positioning**: In peak mode, idle elevators move to high-traffic floors
2. **Zone-based**: Assign elevators to floor zones (e.g., E1: 0-3, E2: 4-7, E3: 8-10)
3. **Predictive**: ML-based demand prediction
4. **Express elevators**: Skip floors for certain elevators
5. **Metrics**: Prometheus metrics for wait time, queue length
6. **Persistence**: Request queue persistence across restarts
7. **gRPC/HTTP API**: REST or gRPC for external control
