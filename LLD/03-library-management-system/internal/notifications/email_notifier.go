package notifications

import (
	"fmt"
	"library-management-system/internal/interfaces"
)

// EmailNotifier implements NotificationService for email (simulated in-memory for demo)
type EmailNotifier struct {
	sentLog []interfaces.NotificationPayload
}

// NewEmailNotifier creates a new email notifier
func NewEmailNotifier() *EmailNotifier {
	return &EmailNotifier{
		sentLog: make([]interfaces.NotificationPayload, 0),
	}
}

// Notify simulates sending email - in production would use SMTP/SendGrid
func (e *EmailNotifier) Notify(payload interfaces.NotificationPayload) error {
	// Simulate: log for testing
	e.sentLog = append(e.sentLog, payload)
	fmt.Printf("[EMAIL] %s -> %s: %s\n", payload.Type, payload.MemberEmail, payload.Message)
	return nil
}

// GetSentLog returns notifications sent (for testing)
func (e *EmailNotifier) GetSentLog() []interfaces.NotificationPayload {
	return e.sentLog
}
