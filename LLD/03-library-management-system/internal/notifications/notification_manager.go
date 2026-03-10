package notifications

import (
	"library-management-system/internal/interfaces"
	"sync"
)

// Ensure NotificationManager implements NotifyBroadcaster
var _ interfaces.NotifyBroadcaster = (*NotificationManager)(nil)

// NotificationManager manages multiple notification channels (Observer pattern)
// Notifiers can be registered/removed without modifying existing code (OCP)
type NotificationManager struct {
	notifiers []interfaces.NotificationService
	mu        sync.RWMutex
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager() *NotificationManager {
	return &NotificationManager{
		notifiers: make([]interfaces.NotificationService, 0),
	}
}

// Register adds a notification channel
func (n *NotificationManager) Register(notifier interfaces.NotificationService) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.notifiers = append(n.notifiers, notifier)
}

// NotifyAll sends notification to all registered channels
func (n *NotificationManager) NotifyAll(payload interfaces.NotificationPayload) {
	n.mu.RLock()
	notifiers := make([]interfaces.NotificationService, len(n.notifiers))
	copy(notifiers, n.notifiers)
	n.mu.RUnlock()

	for _, notifier := range notifiers {
		_ = notifier.Notify(payload)
	}
}
