# Library Management System — Interview Strategy (45 min)

## Time Allocation

| Phase | Time | What to Do |
|-------|------|------------|
| 1. Clarify & Scope | 3 min | Confirm books, members, checkout, return, reservations, fines; scope out multiple copies, renewals |
| 2. Core Models | 7 min | Book, Member, Loan, Reservation, Fine |
| 3. Repository Interfaces | 5 min | BookRepository, MemberRepository, LoanRepository, ReservationRepository, FineRepository |
| 4. Service Interfaces | 5 min | NotifyBroadcaster, PerDayFineCalculator (or interface) |
| 5. Core Service Implementation | 12 min | LoanService.CheckOut() and Return() — due date, overdue fine, reservation notify |
| 6. Handler / main.go Wiring | 5 min | Wire repos, NotificationManager, FineCalculator, services |
| 7. Extend & Discuss | 8 min | Observer for notifications, fine strategy, reservation FIFO |

## Phase 1: Clarify & Scope (3 min)

**Questions to ask:**
- Max books per member? → 5 (Standard), 10 (Premium)
- Loan period? → 14 days
- Fine rate? → $1/day overdue
- Reservations? → When book checked out; FIFO when returned
- Notifications? → Due reminder, overdue, reservation ready

**Scope out:** Renewals, multiple copies per book (simplify to single copy or totalCopies/availableCopies).

## Phase 2: Core Models (7 min)

**Start with:**
- `Book`: ID, Title, Author, ISBN, Subject, TotalCopies, AvailableCopies, Status
- `Member`: ID, Name, Email, MembershipType, BorrowedCount, MaxBorrowedLimit, IsActive
- `Loan`: ID, BookID, MemberID, IssueDate, DueDate, ReturnDate, Status

**Then:**
- `Reservation`: ID, BookID, MemberID, ReservedAt, Status, NotifiedAt
- `Fine`: ID, LoanID, MemberID, Amount, Status, PaidAt

**Skip for now:** Member deactivation, book removal rules.

## Phase 3: Repository Interfaces (5 min)

**Essential:**
- `BookRepository`: Create, GetByID, GetByISBN, ListAll, Update
- `MemberRepository`: Create, GetByID, GetByEmail, Update
- `LoanRepository`: Create, GetByID, GetByBookID, GetOverdueLoans, Update
- `ReservationRepository`: Create, GetPendingByBookID, Update
- `FineRepository`: Create, GetByID, GetByLoanID, GetPendingByMemberID, Update

**Skip:** Complex queries; keep methods minimal.

## Phase 4: Service Interfaces (5 min)

**Essential:**
- `NotifyBroadcaster`: NotifyAll(payload) — for due, overdue, reservation ready
- `FineCalculator` (or concrete): Calculate(loan, now) — days overdue × rate
- `NotificationService`: ConsoleNotifier, EmailNotifier implement it

**Key abstraction:** Observer pattern—LoanService and FineService don't know notification channels.

## Phase 5: Core Service Implementation (12 min)

**Key method:** `LoanService.CheckOut(bookID, memberID)` and `LoanService.Return(loanID)` — this is where the core logic lives.

**CheckOut flow:**
1. Get book; validate AvailableCopies > 0
2. Get member; validate BorrowedCount < MaxBorrowedLimit
3. Create Loan: IssueDate=now, DueDate=now+14 days
4. Decrement book.AvailableCopies, increment member.BorrowedCount
5. Save loan
6. (Optional) Send due reminder via NotifyBroadcaster

**Return flow:**
1. Get loan; validate not already returned
2. If overdue: create Fine via FineCalculator (daysOverdue × rate), notify via NotifyBroadcaster
3. Increment book.AvailableCopies, decrement member.BorrowedCount
4. Set loan.ReturnDate, Status
5. Get first pending reservation for this book; notify them (reservation ready)
6. Update reservation status

**FineService.ProcessOverdueLoans():** Cron job—get overdue loans, create fines, notify. Date arithmetic: `daysOverdue = (now - DueDate).Hours()/24`.

## Phase 6: main.go Wiring (5 min)

```go
bookRepo := repositories.NewInMemoryBookRepo()
memberRepo := repositories.NewInMemoryMemberRepo()
loanRepo := repositories.NewInMemoryLoanRepo()
reservationRepo := repositories.NewInMemoryReservationRepo()
fineRepo := repositories.NewInMemoryFineRepo()

notifier := notifications.NewNotificationManager()
notifier.Register(notifications.NewConsoleNotifier())
notifier.Register(notifications.NewEmailNotifier())

fineCalc := services.NewPerDayFineCalculator(1.0)  // $1/day
loanSvc := services.NewLoanService(loanRepo, bookRepo, memberRepo, reservationRepo, notifier, fineCalc)
```

## Phase 7: Extend & Discuss (8 min)

**Design patterns to mention:**
- Repository (data access)
- Observer (NotifyBroadcaster, multiple notification channels)
- Strategy (FineCalculator—flat, tiered, grace period)
- Factory (Loan creation with DueDate)

**Extensions:**
- Reservation expiry (3 days to pick up)
- SendRemindersForLoansDueWithin(days)
- Different fine rates by membership
- Scheduler for ProcessOverdueLoans

## Tips

- **Prioritize if low on time:** CheckOut and Return; due date calculation; fine on return. Skip reservation notify details.
- **Common mistakes:** Forgetting to decrement BorrowedCount on return; wrong date arithmetic for fine; not notifying first reservation.
- **What impresses:** Observer for notifications, PerDayFineCalculator injectable, FIFO reservation queue.
