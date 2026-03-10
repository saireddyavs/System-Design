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

// GetByID retrieves a notification by ID
func (r *InMemoryNotificationRepository) GetByID(ctx context.Context, id string) (*models.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	n, ok := r.notifications[id]
	if !ok {
		return nil, ErrNotFound
	}
	copy := *n
	return &copy, nil
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

// ListByUserID returns notifications for a user
func (r *InMemoryNotificationRepository) ListByUserID(ctx context.Context, userID string, limit int) ([]*models.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.Notification
	for _, n := range r.notifications {
		if n.UserID == userID {
			copy := *n
			result = append(result, &copy)
		}
	}
	// Simple limit - in production would sort by CreatedAt desc
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

// ListByStatus returns notifications with given status
func (r *InMemoryNotificationRepository) ListByStatus(ctx context.Context, status models.Status, limit int) ([]*models.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.Notification
	for _, n := range r.notifications {
		if n.Status == status {
			copy := *n
			result = append(result, &copy)
		}
	}
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

var ErrNotFound = &NotFoundError{}

type NotFoundError struct{}

func (e *NotFoundError) Error() string {
	return "not found"
}

var _ interfaces.NotificationRepository = (*InMemoryNotificationRepository)(nil)
