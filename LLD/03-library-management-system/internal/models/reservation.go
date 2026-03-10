package models

import "time"

// ReservationStatus represents the state of a reservation
type ReservationStatus string

const (
	ReservationStatusPending   ReservationStatus = "Pending"
	ReservationStatusNotified  ReservationStatus = "Notified"
	ReservationStatusFulfilled ReservationStatus = "Fulfilled"
	ReservationStatusCancelled ReservationStatus = "Cancelled"
	ReservationStatusExpired   ReservationStatus = "Expired"
)

// Reservation represents a book reservation by a member
type Reservation struct {
	ID         string             `json:"id"`
	BookID     string             `json:"book_id"`
	MemberID   string             `json:"member_id"`
	ReservedAt time.Time          `json:"reserved_at"`
	Status     ReservationStatus  `json:"status"`
	NotifiedAt *time.Time         `json:"notified_at,omitempty"`
	ExpiresAt  time.Time          `json:"expires_at"` // Typically 3 days after notification
	CreatedAt  time.Time          `json:"created_at"`
}

// IsActive returns true if reservation is still valid
func (r *Reservation) IsActive() bool {
	return r.Status == ReservationStatusPending || r.Status == ReservationStatusNotified
}
