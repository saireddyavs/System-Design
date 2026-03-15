package repositories

import (
	"sort"
	"sync"

	"social-media-platform/internal/models"
)

// InMemoryNotificationRepository implements NotificationRepository.
type InMemoryNotificationRepository struct {
	notifications map[string]*models.Notification
	byUserID      map[string][]string
	mu            sync.RWMutex
}

// NewInMemoryNotificationRepository creates a new in-memory notification repository
func NewInMemoryNotificationRepository() *InMemoryNotificationRepository {
	return &InMemoryNotificationRepository{
		notifications: make(map[string]*models.Notification),
		byUserID:      make(map[string][]string),
	}
}

func (r *InMemoryNotificationRepository) Create(notification *models.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.notifications[notification.ID] = notification
	r.byUserID[notification.UserID] = append(r.byUserID[notification.UserID], notification.ID)
	return nil
}

func (r *InMemoryNotificationRepository) GetByUserID(userID string, limit, offset int) ([]*models.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids, exists := r.byUserID[userID]
	if !exists {
		return []*models.Notification{}, nil
	}

	var notifications []*models.Notification
	for _, id := range ids {
		if n, ok := r.notifications[id]; ok {
			notifications = append(notifications, n)
		}
	}
	sort.Slice(notifications, func(i, j int) bool {
		return notifications[i].CreatedAt.After(notifications[j].CreatedAt)
	})

	start := offset
	if start > len(notifications) {
		return []*models.Notification{}, nil
	}
	end := start + limit
	if end > len(notifications) {
		end = len(notifications)
	}
	return notifications[start:end], nil
}
