package services

import (
	"airline-reservation-system/internal/models"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// BookingFactory creates bookings (Factory Pattern)
type BookingFactory struct{}

// NewBookingFactory creates a new booking factory
func NewBookingFactory() *BookingFactory {
	return &BookingFactory{}
}

// CreateBooking creates a new booking with generated ID and reference
func (f *BookingFactory) CreateBooking(passengerID, flightID string, seatIDs []string, totalAmount float64) (*models.Booking, error) {
	if len(seatIDs) == 0 {
		return nil, fmt.Errorf("at least one seat required")
	}
	if totalAmount <= 0 {
		return nil, fmt.Errorf("invalid amount")
	}

	id := f.generateID()
	bookingRef := f.generateBookingRef()

	return models.NewBooking(id, passengerID, flightID, seatIDs, totalAmount, bookingRef), nil
}

func (f *BookingFactory) generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "BKG-" + hex.EncodeToString(b)
}

func (f *BookingFactory) generateBookingRef() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	rand.Read(b)
	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}
	return string(b)
}
