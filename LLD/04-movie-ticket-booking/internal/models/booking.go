package models

import "time"

// BookingStatus represents the state of a booking
type BookingStatus string

const (
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusCancelled BookingStatus = "cancelled"
)

// Booking represents a ticket booking
type Booking struct {
	ID          string         `json:"id"`
	UserID      string         `json:"user_id"`
	ShowID      string         `json:"show_id"`
	SeatIDs     []string       `json:"seat_ids"`
	TotalAmount float64        `json:"total_amount"`
	Status      BookingStatus  `json:"status"`
	BookedAt    time.Time      `json:"booked_at"`
	CancelledAt *time.Time     `json:"cancelled_at,omitempty"`
	RefundAmount float64       `json:"refund_amount,omitempty"`
}
