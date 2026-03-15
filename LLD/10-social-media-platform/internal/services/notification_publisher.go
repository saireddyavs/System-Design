package services

import (
	"sync"

	"social-media-platform/internal/interfaces"
	"social-media-platform/internal/models"
)

// NotificationPublisherImpl implements the Observer pattern for notifications.
// When events occur (like, comment, friend request), observers are notified.
type NotificationPublisherImpl struct {
	observers []interfaces.NotificationObserver
	mu        sync.RWMutex
}

// NewNotificationPublisher creates a new notification publisher
func NewNotificationPublisher() *NotificationPublisherImpl {
	return &NotificationPublisherImpl{
		observers: make([]interfaces.NotificationObserver, 0),
	}
}

func (p *NotificationPublisherImpl) Publish(notification *models.Notification) {
	p.mu.RLock()
	observers := make([]interfaces.NotificationObserver, len(p.observers))
	copy(observers, p.observers)
	p.mu.RUnlock()

	for _, obs := range observers {
		obs.Notify(notification)
	}
}
