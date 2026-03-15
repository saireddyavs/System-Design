package repositories

import (
	"context"
	"sync"

	"notification-system/internal/interfaces"
	"notification-system/internal/models"
)

// InMemoryNotificationRepository is a thread-safe in-memory implementation
type InMemoryNotificationRepository struct {
	notifications map[string]*models.Notification
	mu           sync.RWMutex
}

// NewInMemoryNotificationRepository creates a new in-memory repo
func NewInMemoryNotificationRepository() *InMemoryNotificationRepository {
	return &InMemoryNotificationRepository{
		notifications: make(map[string]*models.Notification),
	}
}

// Create stores a new notification
func (r *InMemoryNotificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	n := *notification
	r.notifications[n.ID] = &n
	return nil
}

// Update updates an existing notification
func (r *InMemoryNotificationRepository) Update(ctx context.Context, notification *models.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.notifications[notification.ID]; !ok {
		return ErrNotFound
	}
	n := *notification
	r.notifications[n.ID] = &n
	return nil
}

var ErrNotFound = &NotFoundError{}

type NotFoundError struct{}

func (e *NotFoundError) Error() string {
	return "not found"
}

var _ interfaces.NotificationRepository = (*InMemoryNotificationRepository)(nil)
