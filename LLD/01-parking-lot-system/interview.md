# Parking Lot System — Interview Strategy (45 min)

## Time Allocation

| Phase | Time | What to Do |
|-------|------|------------|
| 1. Clarify & Scope | 3 min | Confirm multi-level, spot sizes, vehicle types, park/unpark, tickets, fees; scope out reservation, API |
| 2. Core Models | 7 min | Vehicle (interface), ParkingSpot, ParkingLevel, ParkingLot (Singleton), Ticket |
| 3. Repository Interfaces | 5 min | Skip—no separate repos; ParkingService holds tickets map; ParkingLot holds levels |
| 4. Service Interfaces | 5 min | ParkingStrategy interface (FindSpot); HourlyFeeStrategy for fees |
| 5. Core Service Implementation | 12 min | ParkingService.Park() and Unpark()—spot finding, ticket creation, license lookup |
| 6. Handler / main.go Wiring | 5 min | GetInstance(), NewParkingService with NearestSpotStrategy, demo park/unpark |
| 7. Extend & Discuss | 8 min | Concurrency (RWMutex), Strategy for allocation, fee calculation, Factory for vehicles |

## Phase 1: Clarify & Scope (3 min)

**Questions to ask:**
- Single level or multi-level? → Multi-level
- Spot sizes? → Small (MC only), Medium (MC+Car), Large (all)
- Vehicle types? → Motorcycle, Car, Bus
- Unpark by ticket ID only, or also license plate? → Both
- Fee calculation? → By duration and vehicle type
- Thread safety required? → Yes

**Scope out:** Reservation, REST API, persistence, real-time displays.

## Phase 2: Core Models (7 min)

**Start with:**
- `Vehicle` interface: `GetType()`, `GetLicensePlate()`, `CanFit(SpotSize)`
- `ParkingSpot`: ID, Size (Small/Medium/Large), Occupied bool, Parked Vehicle
- `ParkingLevel`: ID, Spots []*ParkingSpot, GetAvailableSpots(vehicle), GetSpot(id)

**Then:**
- `ParkingLot`: Singleton, Levels []*ParkingLevel, levelID map for O(1) lookup
- `Ticket`: ID, LicensePlate, SpotID, LevelID, EntryTime

**Skip for now:** Fee strategy details, different allocation strategies.

## Phase 3: Repository Interfaces (5 min)

**Essential:** None—ParkingService holds `tickets map[string]*Ticket`; ParkingLot holds levels. No separate repository layer for LLD scope.

**If interviewer asks:** "We could add TicketRepository for persistence; for now tickets live in ParkingService."

## Phase 4: Service Interfaces (5 min)

**Essential:**
- `ParkingStrategy` interface: `FindSpot(vehicle, spots) *ParkingSpot`
- `HourlyFeeStrategy` (or interface): `Calculate(vehicle, duration) int64`

**Key abstraction:** Strategy lets you swap nearest-first vs farthest-first without changing ParkingService.

## Phase 5: Core Service Implementation (12 min)

**Key method:** `ParkingService.Park(vehicle)` — this is where the core logic lives.

**Flow:**
1. Check vehicle not already parked (linear scan over tickets by license—or maintain license→ticket map)
2. For each level: `GetAvailableSpots(vehicle)` → filter spots that can fit
3. Call `strategy.FindSpot(vehicle, spots)` — NearestSpotStrategy returns first available
4. `spot.Park(vehicle)` — mark occupied
5. Create Ticket, store in `tickets[ticketID]`, return ticket

**Unpark(ticketIDOrLicense):**
1. Lookup ticket: by ID (map) or license (linear scan)
2. Get level, get spot, `spot.Unpark()`, delete from tickets, return vehicle

**Concurrency:** Use `sync.RWMutex` on tickets map; RLock for availability checks, Lock for park/unpark.

## Phase 6: main.go Wiring (5 min)

```go
lot := models.GetInstance()
lot.Initialize(levels)  // levels with spots
strategy := strategies.NewNearestSpotStrategy()
feeStrategy := strategies.NewHourlyFeeStrategy()
ps := services.NewParkingService(lot, strategy, feeStrategy)

ticket, _ := ps.Park(car)
_, vehicle, _ := ps.Unpark(ticket.ID)
fee := ps.CalculateFee(ticket, 0)
```

Show dependency injection: strategy and fee strategy passed in.

## Phase 7: Extend & Discuss (8 min)

**Design patterns to mention:**
- Singleton (ParkingLot)
- Strategy (spot allocation, fee calculation)
- Factory (NewVehicle)
- DIP (ParkingService depends on interfaces)

**Extensions:**
- Observer for availability displays
- Different fee strategies (peak/off-peak)
- Reservation with time windows
- Persistence (TicketRepository)

## Tips

- **Prioritize if low on time:** Park/Unpark flow, ticket lookup, spot compatibility. Skip fee calculation details.
- **Common mistakes:** Forgetting thread safety; not handling license-plate unpark; spot compatibility (Small=MC only).
- **What impresses:** Clear spot compatibility matrix, RWMutex for read-heavy availability, Strategy for both allocation and fees.
