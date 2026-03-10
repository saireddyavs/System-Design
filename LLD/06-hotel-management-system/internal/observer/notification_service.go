package observer

import (
	"hotel-management-system/internal/interfaces"
	"sync"
)

// NotificationService implements observer pattern for booking events
type NotificationService struct {
	handlers []func(interfaces.NotificationPayload)
	mu       sync.RWMutex
}

// NewNotificationService creates a new notification service
func NewNotificationService() interfaces.NotificationService {
	return &NotificationService{
		handlers: make([]func(interfaces.NotificationPayload), 0),
	}
}

// Notify broadcasts event to all subscribers (Observer pattern)
func (n *NotificationService) Notify(payload interfaces.NotificationPayload) error {
	n.mu.RLock()
	handlers := make([]func(interfaces.NotificationPayload), len(n.handlers))
	copy(handlers, n.handlers)
	n.mu.RUnlock()

	for _, h := range handlers {
		if h != nil {
			h(payload)
		}
	}
	return nil
}

// Subscribe adds an observer
func (n *NotificationService) Subscribe(handler func(interfaces.NotificationPayload)) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.handlers = append(n.handlers, handler)
}
