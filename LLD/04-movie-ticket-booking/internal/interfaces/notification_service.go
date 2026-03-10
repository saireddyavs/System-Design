package interfaces

import "movie-ticket-booking/internal/models"

// NotificationEvent represents events that trigger notifications (Observer Pattern)
type NotificationEvent string

const (
	EventBookingConfirmed NotificationEvent = "booking_confirmed"
	EventBookingCancelled NotificationEvent = "booking_cancelled"
)

// NotificationPayload contains data for notification
type NotificationPayload struct {
	Event   NotificationEvent
	Booking *models.Booking
	User    *models.User
	Message string
}

// NotificationService defines notification operations (Observer Pattern)
type NotificationService interface {
	Notify(payload *NotificationPayload) error
	Subscribe(handler func(payload *NotificationPayload))
}
