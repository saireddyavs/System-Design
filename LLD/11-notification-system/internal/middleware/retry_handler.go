package middleware

import (
	"context"
	"time"

	"notification-system/internal/interfaces"
	"notification-system/internal/models"
)

// RetrySender wraps a NotificationSender with retry logic
// Decorator Pattern: Adds retry with exponential backoff
type RetrySender struct {
	inner       interfaces.NotificationSender
	maxRetries  int
	baseBackoff time.Duration
}

// NewRetrySender creates a sender wrapper with retry
func NewRetrySender(inner interfaces.NotificationSender, maxRetries int, baseBackoff time.Duration) *RetrySender {
	return &RetrySender{
		inner:       inner,
		maxRetries:  maxRetries,
		baseBackoff: baseBackoff,
	}
}

// Send attempts delivery with exponential backoff retries
func (r *RetrySender) Send(ctx context.Context, notification *models.Notification, user *models.User) error {
	var lastErr error
	backoff := r.baseBackoff

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				backoff *= 2 // Exponential backoff
			}
		}

		lastErr = r.inner.Send(ctx, notification, user)
		if lastErr == nil {
			return nil
		}
	}

	return lastErr
}

// Channel delegates to inner sender
func (r *RetrySender) Channel() models.Channel {
	return r.inner.Channel()
}

// RateLimitedSender wraps a NotificationSender with rate limiting
// Decorator Pattern: Adds rate limiting before sending
type RateLimitedSender struct {
	inner       interfaces.NotificationSender
	rateLimiter interfaces.RateLimiter
}

// NewRateLimitedSender creates a sender wrapper with rate limiting
func NewRateLimitedSender(inner interfaces.NotificationSender, rateLimiter interfaces.RateLimiter) *RateLimitedSender {
	return &RateLimitedSender{
		inner:       inner,
		rateLimiter: rateLimiter,
	}
}

// Send checks rate limit before delegating (rate limit check is also in service - this is for pipeline)
func (r *RateLimitedSender) Send(ctx context.Context, notification *models.Notification, user *models.User) error {
	if !notification.IsCritical() {
		if !r.rateLimiter.Allow(ctx, notification.UserID, notification.Channel, notification.Priority) {
			return ErrRateLimited
		}
	}
	err := r.inner.Send(ctx, notification, user)
	if err == nil {
		_ = r.rateLimiter.Record(ctx, notification.UserID, notification.Channel)
	}
	return err
}

// Channel delegates to inner sender
func (r *RateLimitedSender) Channel() models.Channel {
	return r.inner.Channel()
}

// ErrRateLimited is returned when rate limit is exceeded
var ErrRateLimited = &RateLimitError{}

type RateLimitError struct{}

func (e *RateLimitError) Error() string {
	return "rate limit exceeded"
}
