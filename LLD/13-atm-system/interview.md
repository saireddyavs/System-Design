# ATM System — Interview Strategy (45 min)

## Time Allocation

| Phase | Time | What to Do |
|-------|------|------------|
| 1. Clarify & Scope | 3 min | Ask about operations, denominations, daily limit, PIN attempts; scope out receipt persistence |
| 2. Core Models | 7 min | Account, Card, Transaction, ATM (CashInventory, State) |
| 3. Repository Interfaces | 5 min | AccountRepository, TransactionRepository, CardRepository |
| 4. Service Interfaces | 5 min | ATMCommand, CashDispenser, TransactionValidator, AuthService |
| 5. Core Service Implementation | 12 min | WithdrawalCommand.Execute — validate chain → deduct → dispense (greedy) → update inventory |
| 6. main.go Wiring | 5 min | Wire ATM, auth, cash dispenser, validator chain, ATMService; demo flow |
| 7. Extend & Discuss | 8 min | State transitions, other commands; discuss greedy optimality for Indian denominations |

## Phase 1: Clarify & Scope (3 min)

**Questions to ask:**
- Operations? (Balance, Withdraw, Deposit, PIN change, Mini statement)
- Denominations? (₹100, ₹500, ₹1000, ₹2000)
- Daily limit? (e.g., ₹50,000)
- Min withdrawal? (₹100, multiple of 100)
- PIN attempts? (3 wrong → block card)
- Session timeout? (2 min)

**Scope IN:** Card auth, balance, withdraw, deposit, PIN change, mini statement, cash dispenser, state machine.

**Scope OUT:** Receipt persistence, multi-ATM coordination, bank API integration.

## Phase 2: Core Models (7 min)

**Write FIRST:**
1. **Account** — ID, AccountNumber, HolderName, Balance, PINHash, DailyWithdrawn, LastWithdrawalDate
2. **Card** — ID, CardNumber, AccountID, ExpiryDate, IsBlocked
3. **Transaction** — ID, AccountID, Type, Amount, BalanceBefore, BalanceAfter, Timestamp
4. **ATM** — ID, Location, CashAvailable `map[Denomination]int`, State (Idle, CardInserted, Authenticated, TransactionInProgress, OutOfService)

**Enums:** Denomination (100, 500, 1000, 2000), ATMState, TransactionType.

**Skip:** Receipt model; keep Transaction minimal.

## Phase 3: Repository Interfaces (5 min)

```go
type AccountRepository interface {
    GetByID(ctx, id string) (*Account, error)
    Update(ctx, *Account) error
}
type TransactionRepository interface {
    Create(ctx, *Transaction) error
    GetByAccountID(ctx, accountID string, limit int) ([]*Transaction, error)
}
type CardRepository interface {
    GetByCardNumber(ctx, cardNumber string) (*Card, error)
    Update(ctx, *Card) error
}
```

AuthService can use these internally. In-memory: maps with mutex.

## Phase 4: Service Interfaces (5 min)

```go
type ATMCommand interface {
    Execute(ctx) (*CommandResult, error)
    GetType() TransactionType
}
type CashDispenser interface {
    CanDispense(ctx, amount float64, inventory CashInventory) bool
    Dispense(ctx, amount float64, inventory CashInventory) (*DispenseResult, error)
}
type TransactionValidator interface {
    Validate(ctx, account *Account, amount float64, txType TransactionType) *ValidationResult
    SetNext(validator TransactionValidator)
}
type AuthService interface {
    ValidateCard(ctx, cardNumber string) (*Card, error)
    Authenticate(ctx, cardNumber, pin string) (*AuthResult, error)
}
```

Chain of Responsibility: AmountValidator → BalanceValidator → DailyLimitValidator. Each has SetNext.

## Phase 5: Core Service Implementation (12 min)

**THE most important method:** `WithdrawalCommand.Execute(ctx)`

1. **Chain of Responsibility:** `validator.Validate(ctx, account, amount, Withdrawal)` → fail if !Valid
2. **Can dispense?** `cashDispenser.CanDispense(ctx, amount, atm.GetCashInventory())` → fail if false
3. **Deduct balance** — `account.DeductBalance(amount)`; fail if false
4. **Record withdrawal** for daily limit — `account.RecordWithdrawal(amount)`
5. **Dispense (Greedy):** `cashDispenser.Dispense(ctx, amount, inventory)` → returns map[Denomination]int
6. On dispense failure: **rollback** — AddBalance, RecordWithdrawal(-amount)
7. **Update ATM inventory** — subtract dispensed notes
8. Create transaction record
9. Return success with dispensed breakdown

**Greedy algorithm (CalculateDispense):**
- Order: [2000, 1000, 500, 100]
- For each denom: needed = min(remaining/denom, inventory[denom]); result[denom]=needed; remaining -= needed*denom
- Return (result, true) if remaining==0 else (nil, false)

**Why this method:** Withdraw is the most complex — validation chain, greedy dispense, rollback. Nail this and you've shown Command, Chain, Strategy.

## Phase 6: main.go Wiring (5 min)

1. Seed Account, Card (with PIN), Transaction (optional)
2. Create ATM with CashInventory (e.g., 10×2000, 10×1000, 20×500, 50×100)
3. Build validator chain: Amount → Balance → DailyLimit
4. Create GreedyCashDispenser, AuthService, AccountService, TransactionService
5. Create ATMService(atm, authService, cashDispenser, receiptPrinter, accountRepo, validator, accountSvc, transactionSvc)
6. Demo: InsertCard → Authenticate → GetBalance → Withdraw(2500) → PrintReceipt → EjectCard

## Phase 7: Extend & Discuss (8 min)

- **State machine:** Idle → CardInserted (insertCard) → Authenticated (authenticate) → TransactionInProgress (executeCommand) → Authenticated. EjectCard → Idle. OutOfService when cash < threshold.
- **Other commands:** BalanceInquiry, Deposit, PINChange, MiniStatement — each is a Command with Execute.
- **Greedy optimality:** Indian denominations are "canonical" — greedy gives minimum notes. Non-canonical (e.g., 1, 3, 4) would need DP.
- **Rollback:** Always rollback account deduction if dispense fails — critical for consistency.

## Tips

- Start with WithdrawalCommand — it exercises the most patterns.
- Greedy loop: iterate denominations largest-first; `remaining %= denom` implicitly by subtracting.
- Validation chain order: Amount (min, multiple) → Balance → DailyLimit. Amount first to fail fast.
- State checks in ATMService: GetState() before each operation; reject if wrong state.
- If time runs out, skip Deposit/PINChange; Balance + Withdraw is enough to demonstrate the design.
