package notifications

import (
	"fmt"
	"library-management-system/internal/interfaces"
)

// ConsoleNotifier implements NotificationService for console output
type ConsoleNotifier struct{}

// NewConsoleNotifier creates a new console notifier
func NewConsoleNotifier() interfaces.NotificationService {
	return &ConsoleNotifier{}
}

// Notify prints notification to console
func (c *ConsoleNotifier) Notify(payload interfaces.NotificationPayload) error {
	fmt.Printf("[%s] To: %s | %s\n", payload.Type, payload.MemberEmail, payload.Message)
	return nil
}
