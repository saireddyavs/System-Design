package middleware

import (
	"context"
	"sync"
	"time"

	"notification-system/internal/interfaces"
	"notification-system/internal/models"
)

// SlidingWindowRateLimiter implements per-user per-channel rate limiting
// Rate limit: max 10 notifications per user per hour per channel
// Critical notifications bypass rate limits
type SlidingWindowRateLimiter struct {
	limit    int
	window   time.Duration
	records  map[string][]time.Time // key: "userID:channel"
	recordsMu sync.RWMutex
}

// NewSlidingWindowRateLimiter creates a rate limiter with configurable limits
func NewSlidingWindowRateLimiter(limit int, window time.Duration) *SlidingWindowRateLimiter {
	rl := &SlidingWindowRateLimiter{
		limit:   limit,
		window:  window,
		records: make(map[string][]time.Time),
	}
	go rl.cleanupLoop()
	return rl
}

// DefaultRateLimiter returns limiter with 10/hour default
func DefaultRateLimiter() *SlidingWindowRateLimiter {
	return NewSlidingWindowRateLimiter(10, time.Hour)
}

func (r *SlidingWindowRateLimiter) key(userID string, channel models.Channel) string {
	return userID + ":" + string(channel)
}

// Allow checks if the user can receive a notification
func (r *SlidingWindowRateLimiter) Allow(ctx context.Context, userID string, channel models.Channel, priority models.Priority) bool {
	if priority == models.PriorityCritical {
		return true
	}

	key := r.key(userID, channel)
	r.recordsMu.Lock()
	defer r.recordsMu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	// Prune old records
	timestamps := r.records[key]
	var valid []time.Time
	for _, t := range timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	return len(valid) < r.limit
}

// Record records a sent notification
func (r *SlidingWindowRateLimiter) Record(ctx context.Context, userID string, channel models.Channel) error {
	key := r.key(userID, channel)
	r.recordsMu.Lock()
	defer r.recordsMu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	timestamps := r.records[key]
	var valid []time.Time
	for _, t := range timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	valid = append(valid, now)
	r.records[key] = valid
	return nil
}

func (r *SlidingWindowRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		r.recordsMu.Lock()
		now := time.Now()
		cutoff := now.Add(-r.window * 2) // Keep extra buffer
		for key, timestamps := range r.records {
			var valid []time.Time
			for _, t := range timestamps {
				if t.After(cutoff) {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(r.records, key)
			} else {
				r.records[key] = valid
			}
		}
		r.recordsMu.Unlock()
	}
}

// Ensure SlidingWindowRateLimiter implements RateLimiter
var _ interfaces.RateLimiter = (*SlidingWindowRateLimiter)(nil)
