package interfaces

import (
	"context"

	"notification-system/internal/models"
)

// NotificationSender defines the strategy interface for sending notifications
// Strategy Pattern: Each channel (Email, SMS, Push) implements this interface
type NotificationSender interface {
	// Send delivers the notification through the channel
	Send(ctx context.Context, notification *models.Notification, user *models.User) error
	// Channel returns the channel this sender handles
	Channel() models.Channel
}
