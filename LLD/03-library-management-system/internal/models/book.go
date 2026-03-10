package models

import "time"

// BookStatus represents the current state of a book (State pattern)
type BookStatus string

const (
	BookStatusAvailable  BookStatus = "Available"
	BookStatusCheckedOut BookStatus = "CheckedOut"
	BookStatusReserved   BookStatus = "Reserved"
	BookStatusLost       BookStatus = "Lost"
)

// Book represents a library book with copy tracking
type Book struct {
	ID               string     `json:"id"`
	Title            string     `json:"title"`
	Author           string     `json:"author"`
	ISBN             string     `json:"isbn"`
	Subject          string     `json:"subject"`
	TotalCopies      int        `json:"total_copies"`
	AvailableCopies  int        `json:"available_copies"`
	Status           BookStatus `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// HasAvailableCopies returns true if at least one copy is available for checkout
func (b *Book) HasAvailableCopies() bool {
	return b.AvailableCopies > 0
}

// CanReserve returns true if book is checked out (can be reserved)
func (b *Book) CanReserve() bool {
	return b.Status == BookStatusCheckedOut || (b.Status == BookStatusAvailable && b.AvailableCopies == 0)
}
