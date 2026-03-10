package interfaces

import "time"

// NotificationType represents the kind of notification
type NotificationType string

const (
	NotificationDueDateReminder NotificationType = "DueDateReminder"
	NotificationOverdue        NotificationType = "Overdue"
	NotificationReservationReady NotificationType = "ReservationReady"
)

// NotificationPayload contains data for notification (Observer pattern)
type NotificationPayload struct {
	Type      NotificationType
	MemberID  string
	MemberEmail string
	BookTitle string
	BookID    string
	DueDate   time.Time
	Message   string
}

// NotificationService defines notification channels (Observer pattern - OCP)
// New channels can be added without modifying existing code
type NotificationService interface {
	Notify(payload NotificationPayload) error
}
