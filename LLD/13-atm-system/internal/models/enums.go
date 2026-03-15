package models

// AccountType represents the type of bank account
type AccountType string

const (
	AccountTypeChecking AccountType = "CHECKING"
	AccountTypeSavings  AccountType = "SAVINGS"
)

// TransactionType represents the type of ATM transaction
type TransactionType string

const (
	TransactionTypeWithdrawal     TransactionType = "WITHDRAWAL"
	TransactionTypeDeposit        TransactionType = "DEPOSIT"
	TransactionTypeBalanceInquiry TransactionType = "BALANCE_INQUIRY"
	TransactionTypePINChange      TransactionType = "PIN_CHANGE"
	TransactionTypeMiniStatement  TransactionType = "MINI_STATEMENT"
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "PENDING"
	TransactionStatusCompleted TransactionStatus = "COMPLETED"
)

// ATMState represents the current state of the ATM
type ATMState string

const (
	ATMStateIdle                 ATMState = "IDLE"
	ATMStateCardInserted         ATMState = "CARD_INSERTED"
	ATMStateAuthenticated        ATMState = "AUTHENTICATED"
	ATMStateTransactionInProgress ATMState = "TRANSACTION_IN_PROGRESS"
	ATMStateOutOfService         ATMState = "OUT_OF_SERVICE"
)

// Denomination represents available cash denominations
type Denomination int

const (
	Denomination100  Denomination = 100
	Denomination500  Denomination = 500
	Denomination1000 Denomination = 1000
	Denomination2000 Denomination = 2000
)

// Business constants
const (
	DailyWithdrawalLimit    = 50000
	MinWithdrawalAmount    = 100
	MaxPINAttempts         = 3
	TransactionTimeoutSecs = 120
	CashThreshold          = 5000 // ATM goes out of service below this
)
