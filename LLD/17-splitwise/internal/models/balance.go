package models

import "time"

// Balance represents what one user owes another
// DebtorID owes CreditorID the Amount
type Balance struct {
	DebtorID   string    `json:"debtor_id"`
	CreditorID string    `json:"creditor_id"`
	Amount     float64   `json:"amount"`
	GroupID    string    `json:"group_id"` // Empty for non-group balances
	UpdatedAt  time.Time `json:"updated_at"`
}

// Transaction represents a settlement payment
type Transaction struct {
	ID          string    `json:"id"`
	FromUserID  string    `json:"from_user_id"`  // Payer (debtor)
	ToUserID    string    `json:"to_user_id"`   // Payee (creditor)
	Amount      float64   `json:"amount"`
	GroupID     string    `json:"group_id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
