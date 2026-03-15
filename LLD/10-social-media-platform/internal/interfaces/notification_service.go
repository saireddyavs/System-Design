package interfaces

import (
	"social-media-platform/internal/models"
)

// NotificationRepository defines the contract for notification storage.
type NotificationRepository interface {
	Create(notification *models.Notification) error
	GetByUserID(userID string, limit, offset int) ([]*models.Notification, error)
}

// NotificationObserver defines the observer interface (Observer pattern).
// Components can subscribe to be notified of events
type NotificationObserver interface {
	Notify(notification *models.Notification)
}

// NotificationPublisher allows components to publish notification events.
type NotificationPublisher interface {
	Publish(notification *models.Notification)
}
