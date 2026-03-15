package models

import "time"

// LoanStatus represents the state of a book loan
type LoanStatus string

const (
	LoanStatusActive   LoanStatus = "Active"
	LoanStatusReturned LoanStatus = "Returned"
	LoanStatusOverdue  LoanStatus = "Overdue"
)

// Loan represents a book lending transaction
type Loan struct {
	ID        string     `json:"id"`
	BookID    string     `json:"book_id"`
	MemberID  string     `json:"member_id"`
	IssueDate time.Time  `json:"issue_date"`
	DueDate   time.Time  `json:"due_date"`
	ReturnDate *time.Time `json:"return_date,omitempty"`
	Status    LoanStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
}

// IsOverdue returns true if loan is past due and not returned
func (l *Loan) IsOverdue() bool {
	return l.Status == LoanStatusActive && time.Now().After(l.DueDate)
}
