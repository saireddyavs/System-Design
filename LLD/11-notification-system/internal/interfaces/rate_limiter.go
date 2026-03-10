package interfaces

import (
	"context"

	"notification-system/internal/models"
)

// RateLimiter checks if a notification can be sent within rate limits
// Decorator Pattern: Wraps senders to add rate limiting behavior
type RateLimiter interface {
	// Allow checks if the user can receive a notification on the given channel
	// Returns true if within limits, false if rate limited
	Allow(ctx context.Context, userID string, channel models.Channel, priority models.Priority) bool
	// Record records a sent notification for rate limit tracking
	Record(ctx context.Context, userID string, channel models.Channel) error
}
