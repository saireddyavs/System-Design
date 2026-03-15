# Movie Ticket Booking System — Interview Strategy (45 min)

## Time Allocation

| Phase | Time | What to Do |
|-------|------|------------|
| 1. Clarify & Scope | 3 min | Confirm movies, theatres, shows, seats, booking, cancellation; scope out payment gateway, hold |
| 2. Core Models | 7 min | Movie, Theatre, Screen, Seat, Show, Booking |
| 3. Repository Interfaces | 5 min | MovieRepository, TheatreRepository, ShowRepository, BookingRepository |
| 4. Service Interfaces | 5 min | PricingStrategy, PaymentProcessor, NotificationService |
| 5. Core Service Implementation | 12 min | BookingService.CreateBooking() — seat selection, atomic update, per-show lock |
| 6. Handler / main.go Wiring | 5 min | Wire repos, ShowRepository with UpdateSeats, PricingStrategy, BookingService |
| 7. Extend & Discuss | 8 min | Per-show mutex for concurrency, pricing strategy, refund policy |

## Phase 1: Clarify & Scope (3 min)

**Questions to ask:**
- Seat selection? → User picks seats; must be available
- Concurrent booking? → No double-booking; pessimistic locking
- Refund policy? → Full refund >24h before show, 50% otherwise
- Seat categories? → Regular, Premium, VIP (different prices)
- Search? → By movie, genre, city

**Scope out:** Seat hold (10 min), payment gateway integration, loyalty points.

## Phase 2: Core Models (7 min)

**Start with:**
- `Movie`: ID, Title, Genre, Duration, Rating, Language
- `Theatre`: ID, Name, City, Address
- `Screen`: ID, TheatreID, Name, Seats []Seat
- `Seat`: ID, ScreenID, Row, Number, Category (Regular/Premium/VIP)

**Then:**
- `Show`: ID, MovieID, ScreenID, TheatreID, StartTime, EndTime, SeatStatusMap map[string]SeatStatus
- `Booking`: ID, UserID, ShowID, SeatIDs, TotalAmount, Status, BookedAt

**Skip for now:** User model details; focus on Show and Booking.

## Phase 3: Repository Interfaces (5 min)

**Essential:**
- `MovieRepository`: Create, GetByID, GetAll (or search)
- `TheatreRepository`: Create, GetByID, GetByCity
- `ShowRepository`: Create, GetByID, GetByMovieID, GetByTheatreID, **UpdateSeats(showID, updateFn)**
- `BookingRepository`: Create, GetByID, GetByUserID

**Critical:** `UpdateSeats` must support atomic read-modify-write for seat status. Signature: `UpdateSeats(showID string, updateFn func(*Show) error) error`

## Phase 4: Service Interfaces (5 min)

**Essential:**
- `PricingStrategy`: CalculatePrice(show, seatIDs, time) — base × category multiplier × weekend multiplier
- `PaymentProcessor`: ProcessPayment(booking) (simplified)
- `NotificationService`: NotifyBookingCreated, NotifyCancelled

**Key abstraction:** PricingStrategy allows weekday vs weekend, seat category multipliers without changing BookingService.

## Phase 5: Core Service Implementation (12 min)

**Key method:** `BookingService.CreateBooking(userID, showID, seatIDs)` — this is where the core logic lives.

**Flow:**
1. Get show; validate seatIDs exist and are in this show's screen
2. Call `showRepo.UpdateSeats(showID, func(show *Show) error { ... })`:
   - Inside callback: for each seatID, check SeatStatusMap[seatID] == Available
   - If any not available, return ErrSeatNotAvailable
   - Mark all seatIDs as Booked in SeatStatusMap
   - Return nil
3. Calculate price via PricingStrategy (base × category × weekend)
4. Create Booking, process payment (simplified)
5. Notify via NotificationService
6. Return booking

**Concurrency:** `UpdateSeats` acquires per-show mutex (`showLocks[showID]`). Only one booking per show at a time. Different shows can book in parallel.

**Cancellation:** Check time to show; full refund if >24h, 50% otherwise. Mark seats Available in SeatStatusMap (same UpdateSeats pattern).

## Phase 6: main.go Wiring (5 min)

```go
movieRepo := repositories.NewInMemoryMovieRepository()
theatreRepo := repositories.NewInMemoryTheatreRepository()
showRepo := repositories.NewInMemoryShowRepository()  // has showLocks map
bookingRepo := repositories.NewInMemoryBookingRepository()

pricingStrategy := strategies.NewWeekdayPricingStrategy()
bookingSvc := services.NewBookingService(bookingRepo, showRepo, pricingStrategy, ...)
searchSvc := services.NewSearchService(movieRepo, theatreRepo, showRepo)
```

**Critical:** ShowRepository must have `getShowLock(showID)` returning per-show mutex; UpdateSeats uses it.

## Phase 7: Extend & Discuss (8 min)

**Design patterns to mention:**
- Repository (data access)
- Strategy (pricing—weekday, seat category)
- Observer (notifications)
- Pessimistic locking (per-show mutex)

**Extensions:**
- Seat hold with TTL
- Optimistic locking (version field) — discuss tradeoff: high conflict → pessimistic better
- Different refund policies
- Idempotency for payment retries

## Tips

- **Prioritize if low on time:** CreateBooking with UpdateSeats and per-show lock. Skip pricing formula details.
- **Common mistakes:** No locking (double-booking); global lock (blocks all shows); forgetting to mark seats in callback.
- **What impresses:** Per-show mutex (parallelism across shows), UpdateSeats callback pattern (atomic), clear refund policy.
