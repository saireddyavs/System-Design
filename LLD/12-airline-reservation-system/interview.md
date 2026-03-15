# Airline Reservation System — Interview Strategy (45 min)

## Time Allocation

| Phase | Time | What to Do |
|-------|------|------------|
| 1. Clarify & Scope | 3 min | Confirm flight search, seat selection (auto vs manual), booking flow, cancellation/refund; scope out multi-leg, loyalty |
| 2. Core Models | 7 min | Flight, Seat, Passenger, Booking; SeatClass, SeatStatus enums |
| 3. Repository Interfaces | 5 min | FlightRepository, BookingRepository, PassengerRepository |
| 4. Service Interfaces | 5 min | SeatAssignmentStrategy, PricingStrategy, PaymentProcessor, BookingObserver |
| 5. Core Service Implementation | 12 min | BookingService.CreateBooking() — validate, assign seats, price, pay, persist, mark booked; rollback on failure |
| 6. main.go Wiring | 5 min | Repos, strategies (Auto/Window seat, ClassMultiplier pricing), BookingNotifier with observers |
| 7. Extend & Discuss | 8 min | Double-booking prevention, refund policy, FlightBuilder, concurrency |

## Phase 1: Clarify & Scope (3 min)

**Questions to ask:**
- Seat assignment? → Auto (first available) or manual (user picks seats)
- Seat classes? → Economy, Business, First; different prices and baggage
- Pricing model? → Base price × class multiplier (Economy 1x, Business 2.5x, First 5x)
- Cancellation refund? → Tiered: >48h 100%, 24–48h 75%, <24h 25%
- Thread safety? → Yes; no double-booking
- Flight creation? → Admin adds flights with seats; use Builder for complex setup

**Scope out:** Multi-leg journeys, loyalty points, seat maps UI, real-time availability push.

## Phase 2: Core Models (7 min)

**Start with:**
- `Flight`: ID, FlightNumber, Origin, Destination (IATA), DepartureTime, ArrivalTime, Aircraft, Seats `[]*Seat`, Status, BasePrice
- `Seat`: ID, FlightID, SeatNumber, Row, Column (A–F), Class, Status (Available/Booked/Blocked), Price
- `Passenger`: ID, Name, Email, Phone, PassportNumber, DateOfBirth
- `Booking`: ID, PassengerID, FlightID, SeatIDs, TotalAmount, Status, BookingRef

**Enums:**
- `SeatClass`: Economy, Business, First
- `SeatStatus`: Available, Booked, Blocked
- `BookingStatus`: Confirmed, Cancelled

**Skip for now:** FlightBuilder details, baggage allowance per class, BookingObserver impl.

## Phase 3: Repository Interfaces (5 min)

**Essential:**
- `FlightRepository`: Create, GetByID, Update
- `BookingRepository`: Create, GetByID, GetByBookingRef, Update
- `PassengerRepository`: Create, GetByID
- `PricingStrategy`: CalculatePrice(basePrice, seats) float64
- `PaymentProcessor`: ProcessPayment(amount, currency, ref), ProcessRefund(txnID, amount)

**Key abstraction:** SeatAssignmentStrategy and PricingStrategy are swappable; PaymentProcessor mocks payment for LLD.

## Phase 4: Service Interfaces (5 min)

**Essential:**
- `SeatAssignmentStrategy`: AssignSeats(availableSeats, count, preferredClass) []string — Auto vs Window preference
- `PricingStrategy`: CalculatePrice(basePrice, seats) float64 — ClassMultiplier (Economy 1x, Business 2.5x, First 5x)
- `PaymentProcessor`: ProcessPayment, ProcessRefund — mock for LLD
- `BookingObserver`: OnBookingCreated(booking), OnBookingCancelled(booking) — decouple notifications

**Key abstraction:** SeatAssignmentStrategy lets you add WindowPreference or FamilyTogether without changing SeatService.

## Phase 5: Core Service Implementation (12 min)

**Key method:** `BookingService.CreateBooking(passengerID, flightID, seatCount, preferredClass)` — this is where the core logic lives.

**Flow:**
1. Validate passenger exists
2. Get flight; reject if cancelled
3. **Auto-assign seats:** `seatService.AutoAssignSeats(flightID, count, preferredClass)` → returns seatIDs (uses strategy)
4. Get seat objects, calculate total: `pricingStrategy.CalculatePrice(flight.BasePrice, seats)`
5. **Process payment:** `paymentProcessor.ProcessPayment(totalAmount, "USD", "booking")` — fail fast if payment fails
6. Create booking via BookingFactory (ID, BookingRef, status)
7. **Mark seats booked:** `seatService.MarkSeatsBooked(flightID, seatIDs)` — atomic; prevents double-booking
8. Persist booking: `bookingRepo.Create(booking)` — **rollback seats** if Create fails
9. Notify observers: `notifier.NotifyBookingCreated(booking)`

**CreateBookingWithSeats** (manual): Same flow but `ManualAssignSeats` validates each seatID is available before proceeding.

**CancelBooking(bookingID):**
1. Get booking, reject if already cancelled
2. Get flight, compute hours until departure
3. Refund % = >48h 100%, 24–48h 75%, <24h 25%
4. Process refund, ReleaseSeats, Update status, NotifyBookingCancelled

**Concurrency:** Repositories use `sync.RWMutex`; MarkSeatsBooked and ReleaseSeats must be atomic. In production: distributed lock (Redis) for multi-instance.

## Phase 6: main.go Wiring (5 min)

```go
flightRepo := repositories.NewInMemoryFlightRepository()
bookingRepo := repositories.NewInMemoryBookingRepository()
passengerRepo := repositories.NewInMemoryPassengerRepository()

seatStrategy := strategies.NewAutoAssignFirstAvailable()
pricingStrategy := strategies.NewClassMultiplierPricing()
paymentProcessor := strategies.NewMockPaymentProcessor()

flightService := services.NewFlightService(flightRepo, bookingRepo)
seatService := services.NewSeatService(flightRepo, seatStrategy)
searchService := services.NewSearchService(flightRepo)
bookingFactory := services.NewBookingFactory()
notifier := services.NewBookingNotifier()
notifier.Subscribe(services.NewEmailBookingObserver())

bookingService := services.NewBookingService(
    bookingRepo, flightRepo, passengerRepo,
    seatService, pricingStrategy, paymentProcessor,
    bookingFactory, notifier,
)

// Create flight via Builder
flight, _ := models.NewFlightBuilder().
    ID("FL-001").FlightNumber("AA100").Route("JFK", "LAX").
    Schedule(departure, arrival).BasePrice(150).
    AddSeatSection(5, []string{"A","B","C"}, models.SeatClassEconomy).
    AddSeatSection(2, []string{"A","B"}, models.SeatClassBusiness).
    Build()
```

Show: Strategies injected; observers subscribed; Builder for flight+seats.

## Phase 7: Extend & Discuss (8 min)

**Design patterns to mention:**
- **Strategy**: SeatAssignmentStrategy (Auto vs Window), PricingStrategy (class multiplier)
- **Factory**: BookingFactory — ID/ref generation, encapsulates creation
- **Observer**: BookingNotifier — email, SMS, analytics without touching BookingService
- **Builder**: FlightBuilder — fluent API for flight + seat sections

**Extensions:**
- Double-booking prevention: Seat status + atomic MarkSeatsBooked; distributed lock for multi-instance
- Refund policy: Tiered by hours; mention idempotency for payment retries
- FlightBuilder: AddSeatSection(rows, columns, class) — avoids telescoping constructors

## Tips

- **Prioritize if low on time:** CreateBooking flow (validate → assign → pay → persist → mark booked + rollback), SeatAssignmentStrategy, refund policy. Skip FlightBuilder, Observer.
- **Common mistakes:** Marking seats booked before payment (race); not rolling back seats on booking create failure; forgetting to release seats on cancel.
- **What impresses:** Explicit rollback on failure; SeatAssignmentStrategy for extensibility; refund policy as a clear function; RWMutex for concurrent access; mentioning distributed lock for scale.
