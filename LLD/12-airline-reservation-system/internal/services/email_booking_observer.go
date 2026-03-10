package services

import (
	"airline-reservation-system/internal/interfaces"
	"airline-reservation-system/internal/models"
	"fmt"
)

// EmailBookingObserver sends email notifications for booking events
type EmailBookingObserver struct {
	// In production, would have email client
}

// NewEmailBookingObserver creates a new email observer
func NewEmailBookingObserver() interfaces.BookingObserver {
	return &EmailBookingObserver{}
}

func (e *EmailBookingObserver) OnBookingCreated(booking *models.Booking) {
	// In production: send confirmation email
	fmt.Printf("[Email] Booking %s confirmed. Ref: %s, Amount: $%.2f\n",
		booking.ID, booking.BookingRef, booking.TotalAmount)
}

func (e *EmailBookingObserver) OnBookingCancelled(booking *models.Booking) {
	// In production: send cancellation email
	fmt.Printf("[Email] Booking %s cancelled. Ref: %s\n", booking.ID, booking.BookingRef)
}
