package notifications

import (
	"fmt"
	"library-management-system/internal/interfaces"
)

// EmailNotifier implements NotificationService for email (simulated in-memory for demo)
type EmailNotifier struct{}

// NewEmailNotifier creates a new email notifier
func NewEmailNotifier() *EmailNotifier {
	return &EmailNotifier{}
}

// Notify simulates sending email - in production would use SMTP/SendGrid
func (e *EmailNotifier) Notify(payload interfaces.NotificationPayload) error {
	fmt.Printf("[EMAIL] %s -> %s: %s\n", payload.Type, payload.MemberEmail, payload.Message)
	return nil
}
