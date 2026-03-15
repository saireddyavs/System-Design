# Hotel Management System — Interview Strategy (45 min)

## Time Allocation

| Phase | Time | What to Do |
|-------|------|------------|
| 1. Clarify & Scope | 3 min | Confirm room types, booking flow, refund policy, search criteria |
| 2. Core Models | 7 min | Room, Guest, Booking, Payment — focus on Booking state machine |
| 3. Repository Interfaces | 5 min | RoomRepository, BookingRepository (with GetBookingsForRoomInRange), GuestRepository, PaymentRepository |
| 4. Service Interfaces | 5 min | BookingService (CreateBooking, ConfirmBooking, CheckIn, CheckOut, CancelBooking), RoomService, PricingStrategy |
| 5. Core Service Implementation | 12 min | BookingService.CreateBooking — overlap check, pricing, state validation |
| 6. Handler / main.go Wiring | 5 min | Wire repos, pricing strategy, notification, payment processor |
| 7. Extend & Discuss | 8 min | Overbooking prevention, cancellation refund, pricing chain, Observer for notifications |

## Phase 1: Clarify & Scope (3 min)

**Questions to ask:**
- Room types? (Single, Double, Deluxe, Suite — confirm base prices)
- Booking lifecycle? (Reservation → Payment → Check-in → Check-out; can cancel before check-in)
- Refund policy? (>24h full, <24h 50%, no-show 0%)
- Search: by date range, room type, price range?
- Overbooking: must prevent — how? (Date overlap check)

**Scope in:** Room management, guest registration, booking flow, availability check, payment, cancellation with refund.

**Scope out:** Housekeeping, inventory, multi-hotel, loyalty tiers (mention briefly).

## Phase 2: Core Models (7 min)

**Start with:** `Booking` — ID, GuestID, RoomID, CheckInDate, CheckOutDate, Status (Pending/Confirmed/CheckedIn/CheckedOut/Cancelled), TotalAmount, PaymentStatus. This is the heart of the system.

**Then:** `Room` — ID, Number, Type, Floor, BasePrice, Status (Available/Occupied/Reserved/Maintenance). `Guest` — ID, Name, Email, Phone, LoyaltyPoints. `Payment` — ID, BookingID, Amount, Method, Status.

**Skip for now:** Amenities, ID proof, detailed payment gateway fields.

## Phase 3: Repository Interfaces (5 min)

**Essential:**
- `RoomRepository`: Create, GetByID, Update
- `BookingRepository`: Create, GetByID, Update, **GetBookingsForRoomInRange(roomID, checkIn, checkOut)** — critical for overlap detection
- `GuestRepository`: Create, GetByID, Update
- `PaymentRepository`: Create, GetByBookingID

**Skip:** Search repository; use RoomService to filter rooms + call BookingRepo for availability.

## Phase 4: Service Interfaces (5 min)

**Essential:**
- `BookingService`: CreateBooking, ConfirmBooking, CheckIn, CheckOut, CancelBooking, GetBooking
- `PricingStrategy` interface: `CalculatePrice(ctx *PricingContext) float64` — allows seasonal, weekend, loyalty strategies
- `RoomService`: CreateRoom, GetAvailableRooms (uses SearchCriteria + overlap check)
- `PaymentProcessor`: ProcessPayment, ProcessRefund

**Key abstraction:** PricingStrategy is composable (Base → Seasonal → Weekday → Loyalty).

## Phase 5: Core Service Implementation (12 min)

**Key method:** `BookingService.CreateBooking(guestID, roomID, checkIn, checkOut)` — this is where the core logic lives.

**Algorithm:**
1. Validate checkOut > checkIn
2. Acquire lock (mutex) to prevent race
3. Get room and guest; return error if not found
4. **Overlap check:** `overlapping := bookingRepo.GetBookingsForRoomInRange(roomID, checkIn, checkOut)`; if len(overlapping) > 0, return ErrRoomNotAvailable
5. Overlap formula: `(StartA < EndB) && (EndA > StartB)` — used inside GetBookingsForRoomInRange
6. Calculate nights, build PricingContext, call `pricing.CalculatePrice(ctx)`
7. Create Booking with Status=Pending, persist
8. Return booking

**Concurrency:** Use `sync.RWMutex` on BookingService for CreateBooking and ConfirmBooking to prevent two bookings for same room overlapping.

**Mention:** CancelBooking applies refund policy: hoursUntilCheckIn >= 24 → full; > 0 → 50%; else 0.

## Phase 6: main.go Wiring (5 min)

```go
roomRepo := NewInMemoryRoomRepository()
bookingRepo := NewInMemoryBookingRepository()
guestRepo := NewInMemoryGuestRepository()
paymentRepo := NewInMemoryPaymentRepository()

pricing := NewCompositePricingStrategy()  // Base → Seasonal → Weekday → Loyalty
notification := NewNotificationService()
paymentProcessor := NewMockPaymentProcessor()

bookingSvc := NewBookingService(bookingRepo, roomRepo, guestRepo, paymentRepo,
    paymentSvc, pricing, notification)
```

Dependency injection: services take interfaces, not concrete repos.

## Phase 7: Extend & Discuss (8 min)

**Design patterns to mention:**
- **Strategy:** PricingStrategy — seasonal, weekend, loyalty; easily add holiday premium
- **State:** Booking status transitions; validate in each method
- **Observer:** NotificationService for confirmation, check-in, check-out events
- **Repository:** Swap in-memory for PostgreSQL without changing services

**Extensions:**
- Interval tree for O(log n) overlap queries if many bookings per room
- Distributed lock (Redis) for multi-instance overbooking prevention
- Idempotency keys for payment and booking creation

## Tips

- **Prioritize if low on time:** CreateBooking with overlap check and pricing; skip RoomService search details.
- **Common mistakes:** Forgetting to lock on CreateBooking; wrong overlap formula (use `Before`/`After` correctly); not validating state transitions.
- **What impresses:** Clear overlap formula, composable pricing strategy, mentioning thread safety and overbooking prevention tests.
