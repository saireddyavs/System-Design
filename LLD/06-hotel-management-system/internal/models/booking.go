package models

import (
	"sync"
	"time"
)

// BookingStatus represents the state of a booking (State pattern)
type BookingStatus string

const (
	BookingStatusPending    BookingStatus = "Pending"
	BookingStatusConfirmed  BookingStatus = "Confirmed"
	BookingStatusCheckedIn  BookingStatus = "CheckedIn"
	BookingStatusCheckedOut BookingStatus = "CheckedOut"
	BookingStatusCancelled BookingStatus = "Cancelled"
)

// Booking represents a room reservation
type Booking struct {
	ID            string
	GuestID       string
	RoomID        string
	CheckInDate   time.Time
	CheckOutDate  time.Time
	Status        BookingStatus
	TotalAmount   float64
	PaymentStatus PaymentStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
	mu            sync.RWMutex
}

// NewBooking creates a new Booking instance
func NewBooking(id, guestID, roomID string, checkIn, checkOut time.Time, totalAmount float64) *Booking {
	now := time.Now()
	return &Booking{
		ID:            id,
		GuestID:       guestID,
		RoomID:        roomID,
		CheckInDate:   checkIn,
		CheckOutDate:  checkOut,
		Status:        BookingStatusPending,
		TotalAmount:   totalAmount,
		PaymentStatus: PaymentStatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// GetStatus returns booking status (thread-safe)
func (b *Booking) GetStatus() BookingStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Status
}

// SetStatus updates booking status (thread-safe)
func (b *Booking) SetStatus(status BookingStatus) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Status = status
	b.UpdatedAt = time.Now()
}

// GetPaymentStatus returns payment status (thread-safe)
func (b *Booking) GetPaymentStatus() PaymentStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.PaymentStatus
}

// SetPaymentStatus updates payment status (thread-safe)
func (b *Booking) SetPaymentStatus(status PaymentStatus) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.PaymentStatus = status
	b.UpdatedAt = time.Now()
}

// GetTotalAmount returns total amount (thread-safe)
func (b *Booking) GetTotalAmount() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.TotalAmount
}

// GetCheckInDate returns check-in date (thread-safe)
func (b *Booking) GetCheckInDate() time.Time {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.CheckInDate
}

// GetCheckOutDate returns check-out date (thread-safe)
func (b *Booking) GetCheckOutDate() time.Time {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.CheckOutDate
}

// Nights returns number of nights for the booking
func (b *Booking) Nights() int {
	checkIn := b.GetCheckInDate()
	checkOut := b.GetCheckOutDate()
	return int(checkOut.Sub(checkIn).Hours() / 24)
}
