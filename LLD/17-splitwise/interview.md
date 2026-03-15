# Splitwise — Interview Strategy (45 min)

## Time Allocation

| Phase | Time | What to Do |
|-------|------|------------|
| 1. Clarify & Scope | 3 min | Confirm: groups, split types (equal, exact, %, share), balance tracking, debt simplification |
| 2. Core Models | 7 min | User, Group, Expense, Split, Balance (DebtorID, CreditorID, Amount) |
| 3. Repository Interfaces | 5 min | UserRepository, GroupRepository, ExpenseRepository, BalanceRepository |
| 4. Service Interfaces | 5 min | SplitStrategy, ExpenseBuilder; BalanceService.SimplifyDebts |
| 5. Core Service Implementation | 12 min | ExpenseService.AddExpense (strategy → splits → update balances) + BalanceService.SimplifyDebts |
| 6. main.go Wiring | 5 min | Wire repos, split registry, services; demo: add expense → get balances → simplify |
| 7. Extend & Discuss | 8 min | Net balance, greedy matching; split strategies; settlement flow |

## Phase 1: Clarify & Scope (3 min)

**Questions to ask:**
- Group-based or also individual expenses?
- Which split types? (Equal, Exact, Percentage, Share?)
- Simplify debts to minimize transactions?
- Settlement: record payment and reduce balance?

**Scope in:** Add expense with split, update balances, get balances, simplify debts (suggested settlements).

**Scope out:** Currency conversion, recurring expenses, categories, audit log.

## Phase 2: Core Models (7 min)

**Write first:** `User` (ID, Name, Email), `Group` (ID, Name, MemberIDs), `Expense` (ID, Description, Amount, PaidBy, SplitType, Splits), `Split` (UserID, Amount), `Balance` (DebtorID, CreditorID, Amount, GroupID).

**Essential fields:**
- `Expense`: PaidBy, Splits (who owes how much)
- `Balance`: debtor owes creditor; AddBalance(debtor, creditor, groupID, amount) adds to existing
- `Split`: UserID, Amount (positive = owes)

**Skip:** Transaction history, expense categories, expense status beyond active.

## Phase 3: Repository Interfaces (5 min)

```go
type UserRepository interface { Create(u *User) error; GetByID(id string) (*User, error) }
type GroupRepository interface { Create(g *Group) error; GetByID(id string) (*Group, error) }
type ExpenseRepository interface { Create(e *Expense) error; GetByGroupID(groupID string) ([]*Expense, error) }
type BalanceRepository interface {
    AddBalance(debtor, creditor, groupID string, amount float64) error
    GetAllForGroup(groupID string) ([]*Balance, error)
}
```

BalanceRepo.AddBalance: find or create Balance(debtor, creditor), add amount (can be negative for settlement).

## Phase 4: Service Interfaces (5 min)

```go
type SplitStrategy interface {
    Supports(splitType SplitType) bool
    CalculateSplits(amount float64, paidBy string, participants []string, params map[string]float64) ([]Split, error)
}
// Registry: GetStrategy(splitType) returns strategy
```

SplitParams: for EXACT map[userID]=amount; for PERCENTAGE map[userID]=pct; for SHARE map[userID]=share.

## Phase 5: Core Service Implementation (12 min)

**THE most important methods:**

**1. ExpenseService.AddExpense(description, amount, paidBy, splitType, participantIDs, splitParams, groupID)**

1. Validate amount > 0, paidBy in group, participants in group
2. strategy = registry.GetStrategy(splitType)
3. splits = strategy.CalculateSplits(amount, paidBy, participantIDs, splitParams)
4. Build expense (Builder: WithDescription, WithAmount, WithPaidBy, WithSplits, WithGroupID)
5. Create expense
6. For each split where UserID != paidBy and Amount > 0: balanceRepo.AddBalance(split.UserID, paidBy, groupID, split.Amount)
7. Notify observers

**2. BalanceService.SimplifyDebts(groupID)** — THE algorithm

1. Get all balances for group
2. Net balance per user: net[debtor] -= amount, net[creditor] += amount
3. Separate creditors (net > 0) and debtors (net < 0)
4. Sort both by amount descending
5. Greedy: match largest creditor with largest debtor; amount = min(cred, debt); create Transaction; reduce both; advance when exhausted
6. Return suggested []Transaction

**Why:** AddExpense shows Strategy + Builder + balance update. SimplifyDebts shows graph reduction + greedy optimization.

## Phase 6: main.go Wiring (5 min)

- Repos: User, Group, Expense, Balance, Transaction
- SplitStrategyRegistry with Equal, Exact, Percentage, Share
- ExpenseService, BalanceService
- Demo: create group, add expense (EQUAL), add expense (EXACT), GetBalancesForGroup, SimplifyDebts

## Phase 7: Extend & Discuss (8 min)

- **Net balance:** Converts directed graph to net positions; creditors vs debtors
- **Greedy matching:** Largest-first minimizes transaction count (not always globally optimal, but good heuristic)
- **Equal split rounding:** Last person gets remainder to avoid floating-point drift
- **Settlement:** AddBalance(debtor, creditor, groupID, -amount) reduces balance; create Transaction record

## Tips

- Start with Equal split only; add Exact/Percentage if time allows
- SimplifyDebts returns suggested transactions; does not execute (Settle does)
- ExpenseBuilder: fluent With*().Build() — makes AddExpense readable
