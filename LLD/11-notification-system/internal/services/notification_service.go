package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"notification-system/internal/interfaces"
	"notification-system/internal/models"
)

const (
	MaxRetries       = 3
	RateLimitPerHour = 10
)

// NotificationEvent represents an event in the notification lifecycle
// Observer Pattern: Events for notification state changes
type NotificationEvent struct {
	NotificationID string
	Status         models.Status
	Timestamp      time.Time
}

// NotificationObserver observes notification events
type NotificationObserver interface {
	OnNotificationEvent(event NotificationEvent)
}

// NotificationService orchestrates notification delivery
// Template Method: ProcessNotification defines the pipeline
type NotificationService struct {
	repo           interfaces.NotificationRepository
	userRepo       interfaces.UserRepository
	senders        map[models.Channel]interfaces.NotificationSender
	templateSvc    *TemplateService
	preferenceSvc  *PreferenceService
	rateLimiter    interfaces.RateLimiter
	observers      []NotificationObserver
	observersMu    sync.RWMutex
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	repo interfaces.NotificationRepository,
	userRepo interfaces.UserRepository,
	senders map[models.Channel]interfaces.NotificationSender,
	templateSvc *TemplateService,
	preferenceSvc *PreferenceService,
	rateLimiter interfaces.RateLimiter,
) *NotificationService {
	return &NotificationService{
		repo:          repo,
		userRepo:      userRepo,
		senders:       senders,
		templateSvc:   templateSvc,
		preferenceSvc: preferenceSvc,
		rateLimiter:   rateLimiter,
		observers:     make([]NotificationObserver, 0),
	}
}

// Subscribe adds an observer for notification events
func (s *NotificationService) Subscribe(observer NotificationObserver) {
	s.observersMu.Lock()
	defer s.observersMu.Unlock()
	s.observers = append(s.observers, observer)
}

// notifyObservers publishes events to all observers
func (s *NotificationService) notifyObservers(event NotificationEvent) {
	s.observersMu.RLock()
	observers := make([]NotificationObserver, len(s.observers))
	copy(observers, s.observers)
	s.observersMu.RUnlock()

	for _, obs := range observers {
		obs.OnNotificationEvent(event)
	}
}

// SendNotification sends a single notification
func (s *NotificationService) SendNotification(ctx context.Context, req *SendRequest) error {
	notification, err := s.createNotification(ctx, req)
	if err != nil {
		return err
	}
	return s.processNotification(ctx, notification)
}

// SendBatch sends multiple notifications concurrently
func (s *NotificationService) SendBatch(ctx context.Context, requests []*SendRequest) []error {
	var wg sync.WaitGroup
	errors := make([]error, len(requests))
	var mu sync.Mutex

	for i, req := range requests {
		wg.Add(1)
		go func(idx int, r *SendRequest) {
			defer wg.Done()
			if err := s.SendNotification(ctx, r); err != nil {
				mu.Lock()
				errors[idx] = err
				mu.Unlock()
			}
		}(i, req)
	}
	wg.Wait()
	return errors
}

// SendRequest is the Builder pattern for constructing send requests
type SendRequest struct {
	UserID     string
	Channel    models.Channel
	Type       models.NotificationType
	Priority   models.Priority
	TemplateID string
	Variables  map[string]string
	Subject    string
	Content    string
}

// SendRequestBuilder implements Builder pattern for SendRequest
type SendRequestBuilder struct {
	req *SendRequest
}

// NewSendRequestBuilder creates a new builder
func NewSendRequestBuilder(userID string) *SendRequestBuilder {
	return &SendRequestBuilder{
		req: &SendRequest{
			UserID:    userID,
			Variables: make(map[string]string),
		},
	}
}

func (b *SendRequestBuilder) WithChannel(ch models.Channel) *SendRequestBuilder {
	b.req.Channel = ch
	return b
}

func (b *SendRequestBuilder) WithType(t models.NotificationType) *SendRequestBuilder {
	b.req.Type = t
	return b
}

func (b *SendRequestBuilder) WithPriority(p models.Priority) *SendRequestBuilder {
	b.req.Priority = p
	return b
}

func (b *SendRequestBuilder) WithTemplate(id string, vars map[string]string) *SendRequestBuilder {
	b.req.TemplateID = id
	b.req.Variables = vars
	return b
}

func (b *SendRequestBuilder) WithContent(subject, content string) *SendRequestBuilder {
	b.req.Subject = subject
	b.req.Content = content
	return b
}

func (b *SendRequestBuilder) Build() *SendRequest {
	return b.req
}

// createNotification creates a notification from request (Factory-like)
func (s *NotificationService) createNotification(ctx context.Context, req *SendRequest) (*models.Notification, error) {
	subject := req.Subject
	content := req.Content

	if req.TemplateID != "" {
		var err error
		subject, content, err = s.templateSvc.RenderTemplate(ctx, req.TemplateID, req.Variables)
		if err != nil {
			return nil, fmt.Errorf("template render: %w", err)
		}
	}

	notification := &models.Notification{
		ID:        generateID(),
		UserID:    req.UserID,
		Channel:   req.Channel,
		Type:      req.Type,
		Priority:  req.Priority,
		Subject:   subject,
		Content:   content,
		Status:    models.StatusPending,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		return nil, err
	}

	s.notifyObservers(NotificationEvent{
		NotificationID: notification.ID,
		Status:         models.StatusPending,
		Timestamp:      time.Now(),
	})

	return notification, nil
}

func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// processNotification is the Template Method - defines the processing pipeline
func (s *NotificationService) processNotification(ctx context.Context, notification *models.Notification) error {
	user, err := s.userRepo.GetByID(ctx, notification.UserID)
	if err != nil {
		return s.markFailed(ctx, notification, err)
	}

	if !user.IsChannelEnabled(notification.Channel) {
		return s.markFailed(ctx, notification, fmt.Errorf("channel %s disabled by user", notification.Channel))
	}

	sender, ok := s.senders[notification.Channel]
	if !ok {
		return s.markFailed(ctx, notification, fmt.Errorf("no sender for channel %s", notification.Channel))
	}

	// Rate limit check (Critical bypasses)
	if !notification.IsCritical() && !s.rateLimiter.Allow(ctx, notification.UserID, notification.Channel, notification.Priority) {
		return fmt.Errorf("rate limit exceeded for user %s on channel %s", notification.UserID, notification.Channel)
	}

	err = sender.Send(ctx, notification, user)
	if err != nil {
		return s.handleSendError(ctx, notification, err)
	}

	return s.markSent(ctx, notification)
}

func (s *NotificationService) markSent(ctx context.Context, n *models.Notification) error {
	now := time.Now()
	n.Status = models.StatusSent
	n.SentAt = &now
	s.repo.Update(ctx, n)
	s.rateLimiter.Record(ctx, n.UserID, n.Channel)
	s.notifyObservers(NotificationEvent{NotificationID: n.ID, Status: models.StatusSent, Timestamp: now})
	return nil
}

func (s *NotificationService) markFailed(ctx context.Context, n *models.Notification, err error) error {
	n.Status = models.StatusFailed
	n.Error = err.Error()
	s.repo.Update(ctx, n)
	s.notifyObservers(NotificationEvent{NotificationID: n.ID, Status: models.StatusFailed, Timestamp: time.Now()})
	return err
}

func (s *NotificationService) handleSendError(ctx context.Context, n *models.Notification, err error) error {
	n.RetryCount++
	n.Error = err.Error()

	if n.CanRetry(MaxRetries) {
		n.Status = models.StatusRetrying
		s.repo.Update(ctx, n)
		s.notifyObservers(NotificationEvent{NotificationID: n.ID, Status: models.StatusRetrying, Timestamp: time.Now()})
		// In production: enqueue for retry with backoff
		return err
	}

	return s.markFailed(ctx, n, err)
}
