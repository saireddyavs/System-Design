# Splitwise - Low Level Design

A production-quality, interview-ready Go implementation of a Splitwise-like expense tracking system following Clean Architecture and SOLID principles.

## 1. Problem Description & Requirements

### Problem
Design and implement an expense tracking system that allows users to split expenses among friends, track who owes whom, and simplify settlements.

### Functional Requirements
- **User Management**: Create users with name, email, phone
- **Group Management**: Create groups, add/remove members
- **Expense Management**: Add expenses with different split types
- **Split Types**:
  - **EQUAL**: Split equally among all participants
  - **EXACT**: Specify exact amounts for each participant
  - **PERCENTAGE**: Split by percentage (must sum to 100%)
  - **SHARE**: Split by shares (e.g., 2:3:5)
- **Balance Tracking**: Track who owes whom and how much
- **Debt Simplification**: Minimize number of transactions to settle
- **Expense History**: View expenses per group or between users
- **Settlement**: Record payments to settle debts

### Non-Functional Requirements
- Thread-safe (sync.RWMutex)
- SOLID principles
- Clean architecture (layered)
- Unit testable

---

## 2. Core Entities & Relationships

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│    User     │     │    Group    │     │   Expense   │
├─────────────┤     ├─────────────┤     ├─────────────┤
│ ID          │     │ ID          │     │ ID          │
│ Name        │     │ Name        │     │ Description │
│ Email       │     │ MemberIDs[] │     │ Amount      │
│ Phone       │     │ CreatedBy   │     │ PaidBy      │
└──────┬──────┘     └──────┬──────┘     │ SplitType   │
       │                   │            │ Splits[]    │
       │    ┌──────────────┴─────────────┤ GroupID     │
       │    │                           └──────┬──────┘
       │    │                                  │
       │    │    ┌─────────────────────────────┘
       │    │    │
       ▼    ▼    ▼
┌─────────────────────────────────────────────────────┐
│                    Balance                           │
├─────────────────────────────────────────────────────┤
│ DebtorID (owes)                                      │
│ CreditorID (owed by)                                 │
│ Amount                                               │
│ GroupID                                              │
└─────────────────────────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────────────────┐
│                 Transaction (Settlement)             │
├─────────────────────────────────────────────────────┤
│ FromUserID, ToUserID, Amount, GroupID                │
└─────────────────────────────────────────────────────┘
```

### Entity Summary
| Entity     | Key Fields |
|------------|------------|
| User       | ID, Name, Email, Phone |
| Group      | ID, Name, MemberIDs, CreatedBy |
| Expense    | ID, Description, Amount, PaidBy, SplitType, Splits, GroupID |
| Split      | UserID, Amount, Percentage, Share, ExactAmount |
| Balance    | DebtorID, CreditorID, Amount, GroupID |
| Transaction| FromUserID, ToUserID, Amount, GroupID |

---

## 3. Split Calculation Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     SPLIT CALCULATION FLOW                               │
└─────────────────────────────────────────────────────────────────────────┘

  ┌──────────────────┐
  │ AddExpense()     │
  │ (description,    │
  │  amount, paidBy, │
  │  splitType,     │
  │  participants,  │
  │  splitParams)   │
  └────────┬─────────┘
           │
           ▼
  ┌──────────────────┐     ┌─────────────────┐
  │ GetStrategy()    │────►│ Strategy        │
  │ (SplitType)      │     │ Registry        │
  └────────┬─────────┘     └─────────────────┘
           │
           ▼
  ┌──────────────────┐
  │ CalculateSplits()│
  │ (Strategy)       │
  └────────┬─────────┘
           │
     ┌─────┴─────┬─────────────┬─────────────┐
     ▼           ▼             ▼             ▼
  ┌──────┐   ┌──────┐   ┌──────────┐   ┌──────┐
  │EQUAL │   │EXACT │   │PERCENTAGE│   │SHARE │
  │amount/n│  │params│   │params%   │   │params│
  └──┬───┘   └──┬───┘   └────┬─────┘   └──┬───┘
     │           │            │            │
     └───────────┴────────────┴────────────┘
           │
           ▼
  ┌──────────────────┐
  │ Splits[]         │
  │ (UserID, Amount) │
  └────────┬─────────┘
           │
           ▼
  ┌──────────────────┐
  │ Update Balances  │
  │ (debtor→creditor)│
  └──────────────────┘
```

---

## 4. Debt Simplification Algorithm

```
┌─────────────────────────────────────────────────────────────────────────┐
│                  DEBT SIMPLIFICATION (Net Balance)                       │
└─────────────────────────────────────────────────────────────────────────┘

  Input: Balances (A owes B $50, B owes C $30, A owes C $20)

  Step 1: Calculate Net Balance per User
  ┌─────────────────────────────────────────────┐
  │  Net = Sum(received) - Sum(owed)            │
  │  A: -50-20 = -70 (debtor)                   │
  │  B: +50-30 = +20 (creditor)                 │
  │  C: +30+20 = +50 (creditor)                 │
  └─────────────────────────────────────────────┘

  Step 2: Separate Creditors (positive) and Debtors (negative)
  ┌─────────────────────────────────────────────┐
  │  Creditors: [(B, 20), (C, 50)]              │
  │  Debtors:   [(A, 70)]                        │
  └─────────────────────────────────────────────┘

  Step 3: Greedy Matching (largest creditor ↔ largest debtor)
  ┌─────────────────────────────────────────────┐
  │  Match A(70) with C(50): A pays C $50        │
  │  Match A(20) with B(20): A pays B $20        │
  │  Result: 2 transactions (minimal)            │
  └─────────────────────────────────────────────┘

  Output: Suggested settlements to minimize transactions

  Note: Simplified transactions are "net" payments. To execute, either:
  - Settle direct balances one-by-one, or
  - Implement balance transfer (A pays B on behalf of C)
```

---

## 5. Design Patterns & WHY

| Pattern | Where | Why |
|---------|-------|-----|
| **Strategy** | Split strategies (Equal, Exact, Percentage, Share) | Different split algorithms can be added without modifying existing code. Open/Closed principle. |
| **Observer** | ExpenseObserver, SettlementObserver | Decouple notification logic from core business logic. Notifications (email, push) can be added without changing expense/balance services. |
| **Factory** | SplitStrategyRegistry | Centralized creation of strategy by type. Encapsulates strategy selection. |
| **Repository** | UserRepo, GroupRepo, ExpenseRepo, BalanceRepo | Abstract data access. Swap in-memory for SQL/NoSQL without changing services. |
| **Builder** | ExpenseBuilder | Construct complex Expense objects with many optional fields. Fluent API for readability. |

---

## 6. SOLID Principles Mapping

| Principle | Application |
|-----------|-------------|
| **Single Responsibility** | UserService (users only), GroupService (groups only), ExpenseService (expenses only), BalanceService (balances/settlements only) |
| **Open/Closed** | New split types via new strategy classes; no changes to ExpenseService |
| **Liskov Substitution** | All SplitStrategy implementations interchangeable; any strategy can replace another |
| **Interface Segregation** | ExpenseObserver vs SettlementObserver (separate; not one fat interface); small repository interfaces |
| **Dependency Inversion** | Services depend on interfaces (UserRepository, GroupRepository, etc.), not concrete implementations |

---

## 7. Interview Explanations

### 3-Minute Pitch
"Splitwise tracks shared expenses among friends. Users create groups, add expenses with different split types (equal, exact, percentage, share), and the system updates balances. When a user pays, we record who owes whom. The debt simplification algorithm uses net balance: compute each user's net (credits minus debits), separate creditors and debtors, then greedily match largest creditor with largest debtor to minimize transactions. We use Strategy for split types, Observer for notifications, Repository for data access, and Builder for complex expense construction."

### 10-Minute Deep Dive
**Q: How do you add a new split type?**
A: Implement the SplitStrategy interface (CalculateSplits, Supports), register it in SplitStrategyRegistry. No changes to ExpenseService—Open/Closed.

**Q: How does debt simplification work?**
A: Net balance approach. For each user, sum what they're owed minus what they owe. Positive = creditor, negative = debtor. Sort both by amount descending. Greedily match: largest debtor pays largest creditor until one is exhausted. Repeat. This minimizes transactions (optimal for the simplified graph).

**Q: How do you handle concurrent updates?**
A: Repositories use sync.RWMutex. Read locks for Get/List, write locks for Create/Update/Delete. AddBalance is atomic within the lock.

**Q: How would you scale this?**
A: Shard by groupID; each group's balances are independent. Use event sourcing for audit trail. Cache net balances per user. Async notifications via message queue.

---

## 8. Future Improvements

- **Currency support**: Multi-currency with conversion
- **Recurring expenses**: Monthly rent, subscriptions
- **Categories**: Food, travel, utilities
- **Audit log**: Full history of all changes
- **Persistence**: PostgreSQL/MongoDB repositories
- **Notifications**: Email, push when someone adds expense or settles
- **Optimization**: Cash flow minimization (not just transaction count)

---

## 9. Running Instructions

```bash
# Build
go build -o splitwise ./cmd

# Run demo
./splitwise   # or: go run ./cmd

# Run tests
go test ./... -v
```

---

## 10. Directory Structure

```
17-splitwise/
├── cmd/
│   └── main.go              # Entry point with demo
├── internal/
│   ├── models/
│   │   ├── user.go          # User entity
│   │   ├── group.go         # Group entity
│   │   ├── expense.go       # Expense, Split entities
│   │   ├── expense_builder.go # Builder pattern
│   │   ├── balance.go       # Balance, Transaction entities
│   │   └── enums.go         # SplitType, ExpenseStatus
│   ├── interfaces/
│   │   ├── repositories.go # Repository interfaces
│   │   ├── services.go      # Service interfaces
│   │   └── strategies.go    # Split strategy interface
│   ├── services/
│   │   ├── user_service.go
│   │   ├── group_service.go
│   │   ├── expense_service.go
│   │   ├── balance_service.go
│   │   └── notification_service.go
│   ├── repositories/
│   │   ├── user_repo.go
│   │   ├── group_repo.go
│   │   ├── expense_repo.go
│   │   ├── balance_repo.go
│   │   └── transaction_repo.go
│   └── strategies/
│       ├── equal_split.go
│       ├── exact_split.go
│       ├── percentage_split.go
│       ├── share_split.go
│       └── registry.go
├── tests/
│   ├── expense_test.go
│   ├── balance_test.go
│   ├── split_strategy_test.go
│   └── integration_test.go
├── go.mod
└── README.md
```
