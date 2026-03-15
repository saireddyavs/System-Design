# Elevator System — Interview Strategy (45 min)

## Time Allocation

| Phase | Time | What to Do |
|-------|------|------------|
| 1. Clarify & Scope | 3 min | Confirm N floors, M elevators, external + internal requests, scheduling; scope out peak mode, express |
| 2. Core Models | 7 min | Building, Elevator, Request, Direction, ElevatorState |
| 3. Repository Interfaces | 5 min | Skip—no persistence; focus on ElevatorController, SchedulingStrategy |
| 4. Service Interfaces | 5 min | SchedulingStrategy (OrderRequests), ElevatorController |
| 5. Core Service Implementation | 12 min | LookStrategy.OrderRequests() + ElevatorService.processNextStep(); Dispatcher selectElevator |
| 6. Handler / main.go Wiring | 5 min | Building, LookStrategy, BuildingController, SubmitRequest, GetStatus |
| 7. Extend & Discuss | 8 min | LOOK vs SCAN, state machine, min-distance dispatching, concurrency |

## Phase 1: Clarify & Scope (3 min)

**Questions to ask:**
- Floors? → 0 to N-1
- Elevators? → M elevators, each with own queue
- Request types? → External (floor button, direction) and Internal (destination floor)
- Scheduling algorithm? → LOOK or SCAN
- Edge cases? → Door open/close, overweight, emergency stop

**Scope out:** Peak mode pre-positioning, zone-based assignment, ML prediction.

## Phase 2: Core Models (7 min)

**Start with:**
- `Request`: ID, Type (External/Internal), SourceFloor, DestFloor, Direction
- `Elevator`: ID, CurrentFloor, Direction, State, RequestQueue []*Request, Capacity, CurrentLoad
- `Building`: ID, TotalFloors, Elevators []*Elevator

**Then:**
- `Direction`: Up, Down, Idle
- `ElevatorState`: Idle, MovingUp, MovingDown, DoorOpen, Maintenance, EmergencyStop

**Skip for now:** Door duration, floor travel time (use constants).

## Phase 3: Repository Interfaces (5 min)

**Essential:** None. Use interfaces:
- `SchedulingStrategy`: OrderRequests(elevator, building) []*Request
- `ElevatorController`: SubmitRequest(req), GetStatus(), EmergencyStop(id), Resume(id)

**Skip:** Persistence, request history.

## Phase 4: Service Interfaces (5 min)

**Essential:**
- `SchedulingStrategy`: OrderRequests(elevator, building) — returns requests in visit order
- `ElevatorController`: Facade over Dispatcher + Elevators

**Key abstraction:** Strategy encapsulates LOOK vs SCAN; swap without changing ElevatorService.

## Phase 5: Core Service Implementation (12 min)

**Key method:** `LookStrategy.OrderRequests(elevator, building)` and `DispatcherService.selectElevator(req)` — these are where the core logic lives.

**LOOK algorithm (OrderRequests):**
1. Get current floor, direction from elevator
2. Build set of floors from requests (source + dest)
3. If direction Up: visit floors from current to max requested, then reverse to min, then below
4. If direction Down: visit floors from current down to min, then up to max
5. If Idle: pick direction of nearest request
6. Map requests to floor sequence; sort requests by first stop order
7. Return ordered []*Request

**Dispatcher selectElevator:**
1. For each elevator: skip if EmergencyStop, Maintenance, Overweight
2. Compute score: prefer same direction toward request; idle by distance; penalty for wrong direction
3. Return elevator with lowest score

**ElevatorService.processNextStep (simplified):**
1. Get ordered queue from strategy
2. If queue empty: set Idle
3. Next floor in sequence: move (update CurrentFloor), open door, process pickup/dropoff, remove completed requests
4. Update direction based on queue

**Concurrency:** RWMutex on Elevator (RequestQueue, CurrentFloor); Dispatcher uses mutex for elevatorSvcs map; channels for request submission.

## Phase 6: main.go Wiring (5 min)

```go
building := models.NewBuilding("B1", "Main Tower", 10, 2)
strategy := strategies.NewLookStrategy()
ctrl := services.NewBuildingController(building, strategy)
defer ctrl.Stop()

req := models.NewExternalRequest(3, models.DirectionUp)
ctrl.SubmitRequest(req)
req2 := models.NewInternalRequest(3, 8)
ctrl.SubmitRequest(req2)

for _, s := range ctrl.GetStatus() {
    fmt.Printf("Elevator %s: Floor %d, %s\n", s.ID, s.CurrentFloor, s.State)
}
```

## Phase 7: Extend & Discuss (8 min)

**Design patterns to mention:**
- Strategy (LOOK vs SCAN)
- State (ElevatorState—Idle, MovingUp, MovingDown, DoorOpen, EmergencyStop)
- Singleton (Dispatcher)
- Facade (BuildingController)

**Extensions:**
- SCAN: go to building end before reverse
- Min-distance dispatching: score = distance + penalty for wrong direction
- Overweight: CanAcceptPassenger(weight) before adding load
- Emergency stop: SetState(EmergencyStop), stop processing
- Peak mode: pre-position idle elevators

## Tips

- **Prioritize if low on time:** LOOK floor sequence (Up: current→max→min; Down: current→min→max). Skip full request ordering.
- **Common mistakes:** SCAN vs LOOK confusion (LOOK reverses when no requests ahead); wrong direction handling in dispatcher; not handling Idle direction.
- **What impresses:** Clear LOOK explanation, min-distance scoring for dispatcher, state machine for elevator, goroutine per elevator.
