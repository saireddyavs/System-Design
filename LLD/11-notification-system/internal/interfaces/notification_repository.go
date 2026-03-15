package interfaces

import (
	"context"

	"notification-system/internal/models"
)

// NotificationRepository defines persistence operations for notifications
type NotificationRepository interface {
	Create(ctx context.Context, notification *models.Notification) error
	Update(ctx context.Context, notification *models.Notification) error
}
