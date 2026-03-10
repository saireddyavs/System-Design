package interfaces

import "hotel-management-system/internal/models"

// BookingEvent represents events for observer pattern
type BookingEvent string

const (
	EventBookingConfirmed  BookingEvent = "booking_confirmed"
	EventBookingCancelled  BookingEvent = "booking_cancelled"
	EventCheckIn          BookingEvent = "check_in"
	EventCheckOut         BookingEvent = "check_out"
	EventPaymentReceived  BookingEvent = "payment_received"
)

// NotificationPayload carries event data
type NotificationPayload struct {
	Event   BookingEvent
	Booking *models.Booking
	Guest   *models.Guest
	Room    *models.Room
}

// NotificationService defines observer for booking events
// O - Open/Closed: New notification channels can be added without modifying core
type NotificationService interface {
	Notify(payload NotificationPayload) error
	Subscribe(handler func(NotificationPayload)) // For observer pattern
}
