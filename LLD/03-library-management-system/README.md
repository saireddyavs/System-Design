# Library Management System - Low Level Design

A production-quality Go implementation of a Library Management System following Clean Architecture and SOLID principles. Designed for 45-50 minute interview implementation scope.

## 1. Problem Description

A library needs a system to manage books, members, lending, reservations, and overdue fines. The system must support:

- **Book Management**: Add, remove, search books with copy tracking
- **Member Management**: Register, update, deactivate members
- **Lending**: Check out and return books with proper tracking
- **Overdue Management**: Calculate and collect fines for late returns
- **Reservations**: Allow members to reserve checked-out books
- **Search**: Find books by title, author, subject, ISBN
- **Notifications**: Due date reminders and reservation availability alerts

## 2. Functional Requirements

| ID | Requirement | Implementation |
|----|-------------|----------------|
| FR1 | Add/remove books with copy tracking | `LibraryService.AddBook`, `RemoveBook` |
| FR2 | Register, update, deactivate members | `LibraryService.RegisterMember`, `UpdateMember`, `DeactivateMember` |
| FR3 | Check out books (max 5 per member) | `LoanService.CheckOut` |
| FR4 | Return books with overdue fine calculation | `LoanService.Return` |
| FR5 | Reserve checked-out books | `ReservationService.Reserve` |
| FR6 | Search by title, author, subject, ISBN | `LibraryService.SearchBooks` |
| FR7 | Due date reminders | `LoanService.SendRemindersForLoansDueWithin` |
| FR8 | Overdue notifications | `FineService.ProcessOverdueLoans` |
| FR9 | Reservation ready notifications | `LoanService.Return` → notifies first in queue |

## 3. Non-Functional Requirements

- **Thread Safety**: All repositories use `sync.RWMutex` for concurrent access
- **Extensibility**: New notification channels and fine strategies via interfaces
- **Testability**: Dependency injection enables unit testing
- **Idiomatic Go**: Interfaces, error handling, no global state

## 4. Core Entities & Relationships

```
┌─────────┐       ┌─────────┐       ┌─────────┐
│  Book   │───1:N─│  Loan   │───N:1─│ Member │
└─────────┘       └─────────┘       └─────────┘
     │                  │                  │
     │                  │                  │
     └──────────┬───────┴──────────────────┘
                │
         ┌──────▼──────┐     ┌─────────┐
         │ Reservation │     │  Fine   │
         └─────────────┘     └─────────┘
```

### Entity Details

| Entity | Key Fields | Relationships |
|--------|------------|---------------|
| **Book** | ID, Title, Author, ISBN, Subject, TotalCopies, AvailableCopies, Status | Has many Loans, Reservations |
| **Member** | ID, Name, Email, MembershipType, BorrowedCount, MaxBorrowedLimit | Has many Loans, Reservations, Fines |
| **Loan** | ID, BookID, MemberID, IssueDate, DueDate, ReturnDate, Status | Belongs to Book, Member; Has one Fine |
| **Reservation** | ID, BookID, MemberID, ReservedAt, Status, NotifiedAt | Belongs to Book, Member |
| **Fine** | ID, LoanID, MemberID, Amount, Status | Belongs to Loan, Member |

## 5. Design Patterns with WHY

### Repository Pattern
**What**: Abstract data access behind interfaces (`BookRepository`, `MemberRepository`, etc.)  
**Why**: Decouples business logic from storage. Swap in-memory for PostgreSQL without changing services. Enables testing with mocks.

### Observer Pattern
**What**: `NotificationManager` broadcasts to multiple `NotificationService` implementations (Console, Email)  
**Why**: Multiple notification channels (SMS, push) can be added without modifying `LoanService` or `FineService`. Loose coupling between event producers and consumers.

### Strategy Pattern
**What**: `PerDayFineCalculator` ($1/day) used by `FineService` and `LoanService`  
**Why**: Fine policies vary (flat rate, tiered, grace period). Calculator is injectable for testing.

### Factory Pattern
**What**: `LoanService.CheckOut` creates `Loan` with computed defaults (DueDate = IssueDate + 14 days)  
**Why**: Centralizes loan creation logic. Ensures consistent initialization.

### State Pattern (Optional)
**What**: `BookStatus` enum (Available, CheckedOut, Reserved, Lost)  
**Why**: Book behavior depends on state. Extensible for future states (UnderRepair, Withdrawn).

## 6. SOLID Principles Mapping

| Principle | Application |
|-----------|-------------|
| **SRP** | `LoanService` = lending + reminders; `FineService` = fines; `ReservationService` = reservations; `LibraryService` = books, members, search |
| **OCP** | New `NotificationService` implementations without modifying existing code. |
| **LSP** | All notifiers (`ConsoleNotifier`, `EmailNotifier`) substitute for `NotificationService` without breaking clients |
| **ISP** | Separate `BookRepository`, `MemberRepository`, `LoanRepository` — clients depend only on what they use |
| **DIP** | Services depend on `interfaces.BookRepository`, not `*InMemoryBookRepo`. High-level modules don't depend on low-level modules. |

## 7. Business Rules Documentation

| Rule | Value | Enforced In |
|------|-------|-------------|
| Max books per member | 5 (Standard), 10 (Premium) | `Member.CanBorrow()`, `LoanService.CheckOut` |
| Loan period | 14 days | `LoanService.CheckOut` |
| Fine rate | $1 per day overdue | `PerDayFineCalculator` |
| Reservation eligibility | Book must have 0 available copies | `ReservationService.Reserve` |
| No duplicate reservations | One pending reservation per member per book | `ReservationService.Reserve` |
| Deactivation | Only if no active loans | `LibraryService.DeactivateMember` |
| Book removal | Only if all copies available | `LibraryService.RemoveBook` |

## 8. Concurrency Considerations

- **Repositories**: Each uses `sync.RWMutex` for thread-safe map access. Read-heavy workloads use `RLock`; writes use `Lock`.
- **No shared mutable state in services**: Services are stateless; all state in repositories.
- **NotificationManager**: Uses `RWMutex` for notifier list. `NotifyAll` copies slice under read lock before iterating to avoid deadlock if a notifier blocks.
- **Future**: For distributed systems, consider optimistic locking (version field) or distributed locks for loan/reservation conflicts.

## 9. Interview Explanations

### 3-Minute Summary
> "I built a Library Management System in Go using Clean Architecture. The core is separation of concerns: **models** define entities, **interfaces** define contracts, **repositories** handle data (in-memory, thread-safe), and **services** implement business logic. I used the **Repository pattern** for data access, **Observer** for notifications (email, console), and **PerDayFineCalculator** for fine calculation ($1/day). LibraryService handles books, members, and search; LoanService handles lending, returns, and due-date reminders. SOLID is applied throughout: each service has a single responsibility, new notification channels extend without modification (OCP), and services depend on interfaces (DIP). Key business rules: max 5 books per member, 14-day loan period, reservations for checked-out books with FIFO notification when returned."

### 10-Minute Deep Dive
> "Let me walk through the architecture. **Models** (Book, Member, Loan, Reservation, Fine) are pure data. **Interfaces** define the contracts: `BookRepository`, `NotificationService`, `NotifyBroadcaster`. Services receive these via constructor injection.
>
> **LoanService** handles checkout and return. On checkout, it validates book availability and member limit, creates a Loan with 14-day due date, decrements available copies, and increments member's borrowed count. On return, it reverses those, creates a Fine if overdue using the injected `PerDayFineCalculator`, and notifies the first person in the reservation queue via `NotifyBroadcaster`.
>
> **FineService** has a `ProcessOverdueLoans` method typically called by a cron job. It finds overdue loans, creates fines using PerDayFineCalculator, and sends notifications.
>
> **ReservationService** only allows reservations when a book has no available copies. Reservations are FIFO; when a book is returned, `LoanService.Return` notifies the first pending reservation.
>
> For **concurrency**, all repositories use `sync.RWMutex`. The design avoids shared mutable state in services. For production, I'd add a PostgreSQL repository implementation and run ProcessOverdueLoans and SendReminders on a scheduler."

## 10. Future Improvements

1. **Persistence**: Replace in-memory repos with PostgreSQL/MySQL implementations
2. **API Layer**: Add HTTP/gRPC handlers with request validation
3. **Scheduler**: Cron for `ProcessOverdueLoans` and `SendRemindersForLoansDueWithin`
4. **Reservation expiry**: Auto-cancel reservations not picked up within 3 days of notification
5. **Audit log**: Track all state changes for compliance
6. **Caching**: Redis for hot search queries
7. **Rate limiting**: Prevent abuse of notification channels
8. **Metrics**: Prometheus for loan volume, overdue rate, reservation conversion

## 11. Data Structures & Algorithms

| DS/Algorithm | Where Used | Why | Alternatives/Tradeoffs |
|-------------|------------|-----|------------------------|
| **HashMap** | All repositories (books, members, loans, reservations, fines) | O(1) lookup by ID; standard in-memory storage | Production: PostgreSQL with indexes; HashMap for demo/testing |
| **Linear search with multi-field filtering** | `LibraryService.SearchBooks(SearchCriteria)` | Iterates all books; filters by ISBN, Title, Author, Subject (case-insensitive partial match) | Elasticsearch for full-text; for small catalogs, linear scan acceptable |
| **Date arithmetic** | `PerDayFineCalculator.Calculate()`, `LoanService.CheckOut` (DueDate = IssueDate + 14 days) | Overdue = now.After(DueDate); fine = daysOverdue × rate; loan period computation | time.Time.AddDate for date math; consider timezone for distributed systems |
| **sync.RWMutex** | All repositories, NotificationManager | Thread-safe concurrent access; RLock for reads | Same as other projects; essential for concurrent checkout/return |

---

## 12. Running the Project

```bash
# Build
go build -o library ./cmd

# Run demo
go run ./cmd

# Run tests
go test ./tests/... -v
```

## 13. Directory Structure

```
03-library-management-system/
├── cmd/main.go                 # Demo entry point
├── internal/
│   ├── models/                 # Domain entities
│   ├── interfaces/             # Repository & service contracts
│   ├── services/               # Business logic
│   ├── repositories/           # In-memory implementations
│   └── notifications/         # Observer implementations
├── tests/                      # Unit tests
├── go.mod
└── README.md
```
