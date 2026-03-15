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
