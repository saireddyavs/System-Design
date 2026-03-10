# ATM System - Low Level Design

A production-quality, interview-ready ATM system implementation in Go following Clean Architecture and SOLID principles.

## Problem Description

Design and implement an ATM (Automated Teller Machine) system that supports:
- Card-based authentication
- Multiple transaction types
- Cash dispensing with denomination management
- Transaction logging and history
- Concurrent user access

## Requirements

### Functional Requirements
- **Authentication**: Card number + PIN validation
- **Operations**: Balance inquiry, cash withdrawal, cash deposit, PIN change, mini statement
- **Cash Dispenser**: Manages denominations (₹100, ₹500, ₹1000, ₹2000)
- **Account Types**: Checking and Savings
- **Transaction Logging**: All transactions logged with history

### Non-Functional Requirements
- **Concurrency**: Thread-safe for multiple ATM access
- **Reliability**: Transaction validation, rollback on failure
- **Extensibility**: Easy to add new operations and validators

### Business Rules
| Rule | Value |
|------|-------|
| 3 wrong PIN attempts | Card blocked |
| Daily withdrawal limit | ₹50,000 |
| Minimum withdrawal | ₹100 (multiple of 100) |
| Transaction timeout | 2 minutes |
| ATM out of service | Cash below ₹5,000 |

---

## Core Entities & Relationships

```
┌─────────────┐      1:1      ┌─────────────┐
│    Card     │───────────────│   Account   │
│─────────────│               │─────────────│
│ CardNumber  │               │ AccountNum  │
│ AccountID   │               │ Balance     │
│ ExpiryDate  │               │ PINHash     │
│ IsBlocked   │               │ DailyWithdr │
└─────────────┘               └──────┬──────┘
                                     │ 1:N
                                     ▼
                              ┌─────────────┐
                              │ Transaction │
                              │─────────────│
                              │ Type        │
                              │ Amount      │
                              │ Status      │
                              └─────────────┘

┌─────────────┐
│    ATM      │
│─────────────│
│ CashInventory│  map[Denomination]int
│ State       │
│ Location    │
└─────────────┘
```

### Entity Details

**Account**
- ID, AccountNumber, HolderName, Type (Checking/Savings)
- Balance, PINHash, DailyWithdrawn, LastWithdrawalDate

**Card**
- ID, CardNumber, AccountID, ExpiryDate, IsBlocked

**Transaction**
- ID, AccountID, Type, Amount, BalanceBefore/After, Status, Timestamp

**ATM**
- ID, Location, CashAvailable (denomination map), State, TotalCash

---

## State Machine Diagram

```
                    ┌──────────────────┐
                    │       IDLE       │
                    │  (No card)       │
                    └────────┬─────────┘
                             │ insertCard()
                             ▼
                    ┌──────────────────┐
                    │  CARD_INSERTED   │
                    │  (Awaiting PIN)  │
                    └────────┬─────────┘
                             │ authenticate() success
                             ▼
                    ┌──────────────────┐
         ┌──────────│  AUTHENTICATED    │◄──────────┐
         │          │  (Ready for ops)  │           │
         │          └────────┬─────────┘            │
         │                   │ executeCommand()     │
         │                   ▼                      │
         │          ┌──────────────────┐           │
         │          │ TRANSACTION_IN_   │           │
         │          │    PROGRESS       │───────────┘
         │          └────────┬─────────┘  (complete)
         │                   │
         │                   │ cash < threshold
         │                   ▼
         │          ┌──────────────────┐
         └─────────►│  OUT_OF_SERVICE   │
            eject   │  (Maintenance)    │
                    └──────────────────┘
```

**State Transitions:**
- `IDLE` → `CARD_INSERTED`: User inserts valid card
- `CARD_INSERTED` → `AUTHENTICATED`: Successful PIN verification
- `AUTHENTICATED` → `TRANSACTION_IN_PROGRESS`: Command execution starts
- `TRANSACTION_IN_PROGRESS` → `AUTHENTICATED`: Command completes
- `AUTHENTICATED`/`CARD_INSERTED` → `IDLE`: User ejects card
- Any → `OUT_OF_SERVICE`: Cash below threshold

---

## Cash Dispensing Algorithm

**Greedy Approach**: Use largest denomination first to minimize note count.

```
Algorithm: CalculateDispense(amount, inventory)
1. If amount not multiple of 100 → return failure
2. Sort denominations descending: [2000, 1000, 500, 100]
3. For each denomination d:
   a. needed = min(amount/d, inventory[d])
   b. result[d] = needed
   c. amount -= needed * d
   d. If amount == 0 → return result
4. If amount > 0 → return failure (cannot dispense exact)
```

**Example**: Withdraw ₹2,500
- 1 × ₹2000 = 2000, remaining 500
- 1 × ₹500 = 500, remaining 0
- **Result**: {2000: 1, 500: 1}

---

## Design Patterns

### 1. State Pattern
**Where**: `ATMService`, `models.ATMState`
**Why**: ATM behavior changes based on current state (Idle, CardInserted, Authenticated, etc.). State pattern encapsulates state-specific behavior and makes transitions explicit.

### 2. Strategy Pattern
**Where**: `interfaces.CashDispenser`, `GreedyCashDispenser`
**Why**: Cash dispensing algorithm can vary (greedy, dynamic programming for exact change). Strategy allows swapping algorithms without changing client code.

### 3. Chain of Responsibility
**Where**: `TransactionValidator` chain (Amount → Balance → DailyLimit)
**Why**: Validation rules can be added/removed/reordered without modifying core logic. Each validator handles one concern.

### 4. Command Pattern
**Where**: `ATMCommand` interface, `WithdrawalCommand`, `BalanceInquiryCommand`, etc.
**Why**: Encapsulates each operation as an object. Enables logging, undo, queueing, and uniform execution interface.

### 5. Repository Pattern
**Where**: `AccountRepository`, `TransactionRepository`, `CardRepository`
**Why**: Abstracts data access. Business logic doesn't know about storage (in-memory, DB). Easy to swap implementations.

### 6. Singleton Pattern
**Where**: `GetATMInstance()` in `atm_factory.go`
**Why**: Single ATM instance per deployment. Ensures consistent state and resource sharing.

---

## SOLID Principles Mapping

| Principle | Implementation |
|-----------|----------------|
| **S**ingle Responsibility | Each service has one job: `AuthService` (auth only), `TransactionService` (transactions only), `AccountService` (account ops) |
| **O**pen/Closed | `CashDispenser` interface: add new strategies without modifying existing code. `TransactionValidator` chain: add validators without changing chain |
| **L**iskov Substitution | All `CashDispenser` implementations interchangeable. All `TransactionValidator` implementations work in chain |
| **I**nterface Segregation | Small, focused interfaces: `AuthService`, `CashDispenser`, `ReceiptPrinter` - clients depend only on what they need |
| **D**ependency Inversion | `ATMService` depends on `interfaces.AuthService`, not concrete `AuthService`. All services depend on repository interfaces |

---

## Interview Explanations

### 3-Minute Summary

"We've built an ATM system with clean architecture. The core entities are Account, Card, Transaction, and ATM. We use a **State machine** for ATM flow: Idle → CardInserted → Authenticated → TransactionInProgress. **Command pattern** encapsulates each operation (withdraw, deposit, balance) as executable objects. **Chain of Responsibility** validates transactions: amount rules, balance check, daily limit. **Strategy pattern** for cash dispensing - we use a greedy algorithm (largest denomination first). **Repository pattern** abstracts data access. Everything is thread-safe with mutexes for concurrent ATM access. Key business rules: 3 wrong PINs block the card, ₹50K daily limit, ₹100 minimum withdrawal."

### 10-Minute Deep Dive

**Architecture**: We follow clean architecture - models at center, interfaces define contracts, services contain business logic, repositories handle persistence. The `internal/` structure keeps dependencies pointing inward.

**State Machine**: The ATM has 5 states. Each transition is guarded - e.g., you can't withdraw in Idle state. State changes are atomic with mutex protection. OutOfService triggers when cash drops below threshold.

**Cash Dispensing**: Greedy algorithm - always use largest available denomination. We validate that exact amount can be dispensed before deducting from account. On dispense failure, we rollback the account deduction.

**Validation Chain**: AmountValidator checks min/multiple rules. BalanceValidator ensures sufficient funds. DailyLimitValidator checks daily withdrawal. Each passes to next on success. Order matters - we check amount before balance.

**Concurrency**: Account uses RWMutex for balance operations. ATMService uses mutex for state transitions. Repositories use mutex for map access. Our concurrent test runs two ATMs against the same account - balance ends up correct.

**Error Handling**: Failed transactions rollback. Card blocks after 3 wrong PINs. Session expires after 2 minutes. We return structured errors, not panics.

---

## Future Improvements

1. **Database Persistence**: Replace in-memory repos with PostgreSQL/MySQL
2. **PIN Hashing**: Use bcrypt/argon2 for PIN storage (currently plain for demo)
3. **Bank API Integration**: Real account validation via banking APIs
4. **Receipt Persistence**: Store receipts for dispute resolution
5. **Metrics & Observability**: Prometheus metrics, structured logging
6. **Rate Limiting**: Prevent brute-force PIN attempts
7. **Multi-Currency**: Support different denominations per region
8. **Undo/Compensation**: Command pattern enables transaction reversal
9. **Event Sourcing**: Store events for audit trail
10. **Circuit Breaker**: For external bank API calls

---

## Running the Project

```bash
# Build
go build ./...

# Run demo
go run ./cmd/main.go

# Run tests
go test ./tests/... -v
```

---

## Project Structure

```
13-atm-system/
├── cmd/main.go              # Entry point, demo flow
├── internal/
│   ├── models/              # Domain entities
│   ├── interfaces/           # Abstractions (DIP)
│   ├── services/             # Business logic
│   ├── repositories/         # Data access
│   └── hardware/             # Hardware abstractions
├── tests/                    # Unit tests
├── go.mod
└── README.md
```
