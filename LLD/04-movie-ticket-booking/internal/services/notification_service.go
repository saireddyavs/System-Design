package services

import (
	"movie-ticket-booking/internal/interfaces"
	"sync"
)

// ObserverNotificationService implements Observer pattern for notifications
type ObserverNotificationService struct {
	handlers []func(*interfaces.NotificationPayload)
	mu       sync.RWMutex
}

// NewObserverNotificationService creates notification service with observer support
func NewObserverNotificationService() *ObserverNotificationService {
	return &ObserverNotificationService{
		handlers: make([]func(*interfaces.NotificationPayload), 0),
	}
}

// Notify notifies all subscribers (Observer pattern)
func (s *ObserverNotificationService) Notify(payload *interfaces.NotificationPayload) error {
	s.mu.RLock()
	handlers := make([]func(*interfaces.NotificationPayload), len(s.handlers))
	copy(handlers, s.handlers)
	s.mu.RUnlock()

	for _, h := range handlers {
		h(payload)
	}
	return nil
}

// Subscribe adds a notification handler
func (s *ObserverNotificationService) Subscribe(handler func(*interfaces.NotificationPayload)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers = append(s.handlers, handler)
}
