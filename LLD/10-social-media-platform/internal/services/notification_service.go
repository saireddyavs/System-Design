package services

import (
	"fmt"
	"sync/atomic"

	"social-media-platform/internal/interfaces"
	"social-media-platform/internal/models"
)

// NotificationService handles notification creation and storage.
// Implements NotificationObserver to receive events (Observer pattern)
type NotificationService struct {
	repo      interfaces.NotificationRepository
	publisher interfaces.NotificationPublisher
	nextID    atomic.Uint64
}

// NewNotificationService creates a new notification service
func NewNotificationService(repo interfaces.NotificationRepository, publisher interfaces.NotificationPublisher) *NotificationService {
	return &NotificationService{
		repo:      repo,
		publisher: publisher,
	}
}

// Notify implements NotificationObserver - persists notifications when published
func (s *NotificationService) Notify(notification *models.Notification) {
	_ = s.repo.Create(notification)
}

// CreateAndPublish creates a notification and publishes to all observers
func (s *NotificationService) CreateAndPublish(notification *models.Notification) error {
	if err := s.repo.Create(notification); err != nil {
		return err
	}
	s.publisher.Publish(notification)
	return nil
}

// GenerateID generates a unique notification ID
func (s *NotificationService) GenerateID() string {
	return fmt.Sprintf("notif-%d", s.nextID.Add(1))
}

// GetUserNotifications retrieves notifications for a user with pagination
func (s *NotificationService) GetUserNotifications(userID string, limit, offset int) ([]*models.Notification, error) {
	return s.repo.GetByUserID(userID, limit, offset)
}
