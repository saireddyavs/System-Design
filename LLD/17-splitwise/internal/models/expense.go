package models

import "time"

// Split represents one participant's share in an expense
type Split struct {
	UserID     string  `json:"user_id"`
	Amount     float64 `json:"amount"`      // Amount this user owes
	Percentage float64 `json:"percentage"`  // Used for PERCENTAGE split
	Share      int     `json:"share"`      // Used for SHARE split (e.g., 2 in 2:3:5)
	ExactAmount float64 `json:"exact_amount"` // Used for EXACT split
}

// Expense represents a shared expense
type Expense struct {
	ID            string        `json:"id"`
	Description   string        `json:"description"`
	Amount        float64       `json:"amount"`
	PaidBy        string        `json:"paid_by"`        // User who paid
	SplitType     SplitType     `json:"split_type"`
	Splits        []Split       `json:"splits"`
	GroupID       string        `json:"group_id"`       // Empty if not group expense
	Status        ExpenseStatus `json:"status"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}
