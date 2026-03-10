package models

import "time"

// FineStatus represents payment state
type FineStatus string

const (
	FineStatusPending FineStatus = "Pending"
	FineStatusPaid    FineStatus = "Paid"
	FineStatusWaived  FineStatus = "Waived"
)

// Fine represents an overdue fine for a loan
type Fine struct {
	ID        string     `json:"id"`
	LoanID    string     `json:"loan_id"`
	MemberID  string     `json:"member_id"`
	Amount    float64    `json:"amount"`
	Status    FineStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	PaidAt    *time.Time `json:"paid_at,omitempty"`
}
