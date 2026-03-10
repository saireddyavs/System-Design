package interfaces

import (
	"context"

	"notification-system/internal/models"
)

// NotificationRepository defines persistence operations for notifications
type NotificationRepository interface {
	Create(ctx context.Context, notification *models.Notification) error
	GetByID(ctx context.Context, id string) (*models.Notification, error)
	Update(ctx context.Context, notification *models.Notification) error
	ListByUserID(ctx context.Context, userID string, limit int) ([]*models.Notification, error)
	ListByStatus(ctx context.Context, status models.Status, limit int) ([]*models.Notification, error)
}
