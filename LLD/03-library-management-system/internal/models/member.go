package models

import "time"

// MembershipType defines member tier
type MembershipType string

const (
	MembershipStandard MembershipType = "Standard"
	MembershipPremium  MembershipType = "Premium"
)

// Member represents a library member
type Member struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Email          string         `json:"email"`
	MembershipType MembershipType `json:"membership_type"`
	JoinDate       time.Time      `json:"join_date"`
	IsActive       bool           `json:"is_active"`
	BorrowedCount  int            `json:"borrowed_count"` // Current number of books borrowed

	// MaxBorrowedLimit is configurable per membership (default 5)
	MaxBorrowedLimit int `json:"max_borrowed_limit"`
}

// CanBorrow returns true if member can borrow more books
func (m *Member) CanBorrow() bool {
	return m.IsActive && m.BorrowedCount < m.MaxBorrowedLimit
}

// DefaultMaxBorrowed returns default limit based on membership type
func DefaultMaxBorrowed(mt MembershipType) int {
	if mt == MembershipPremium {
		return 10
	}
	return 5
}
