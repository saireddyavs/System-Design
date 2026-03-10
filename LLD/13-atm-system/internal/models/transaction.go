package models

import "time"

// Transaction represents an ATM transaction
type Transaction struct {
	ID            string
	AccountID     string
	Type          TransactionType
	Amount        float64
	BalanceBefore float64
	BalanceAfter  float64
	Status        TransactionStatus
	Description   string
	Timestamp     time.Time
	Metadata      map[string]string
}

// NewTransaction creates a new transaction record
func NewTransaction(id, accountID string, txType TransactionType, amount float64, balanceBefore, balanceAfter float64) *Transaction {
	return &Transaction{
		ID:            id,
		AccountID:     accountID,
		Type:          txType,
		Amount:        amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		Status:        TransactionStatusPending,
		Timestamp:     time.Now(),
		Metadata:      make(map[string]string),
	}
}
