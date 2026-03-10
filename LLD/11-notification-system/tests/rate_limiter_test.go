package tests

import (
	"context"
	"testing"
	"time"

	"notification-system/internal/middleware"
	"notification-system/internal/models"
)

func TestRateLimiter_AllowsWithinLimit(t *testing.T) {
	ctx := context.Background()
	rl := middleware.NewSlidingWindowRateLimiter(5, time.Hour)

	for i := 0; i < 5; i++ {
		if !rl.Allow(ctx, "user1", models.ChannelEmail, models.PriorityLow) {
			t.Errorf("Allow #%d should succeed", i+1)
		}
		_ = rl.Record(ctx, "user1", models.ChannelEmail)
	}
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	ctx := context.Background()
	rl := middleware.NewSlidingWindowRateLimiter(3, time.Hour)

	for i := 0; i < 3; i++ {
		_ = rl.Record(ctx, "user1", models.ChannelEmail)
	}

	if rl.Allow(ctx, "user1", models.ChannelEmail, models.PriorityLow) {
		t.Error("Should be rate limited after 3 records")
	}
}

func TestRateLimiter_CriticalBypasses(t *testing.T) {
	ctx := context.Background()
	rl := middleware.NewSlidingWindowRateLimiter(1, time.Hour)

	_ = rl.Record(ctx, "user1", models.ChannelEmail)

	// Critical should always be allowed
	if !rl.Allow(ctx, "user1", models.ChannelEmail, models.PriorityCritical) {
		t.Error("Critical priority should bypass rate limit")
	}
}

func TestRateLimiter_PerUserPerChannel(t *testing.T) {
	ctx := context.Background()
	rl := middleware.NewSlidingWindowRateLimiter(2, time.Hour)

	_ = rl.Record(ctx, "user1", models.ChannelEmail)
	_ = rl.Record(ctx, "user1", models.ChannelEmail)

	// user1 email: at limit
	if rl.Allow(ctx, "user1", models.ChannelEmail, models.PriorityLow) {
		t.Error("user1 email should be limited")
	}

	// user2 email: ok
	if !rl.Allow(ctx, "user2", models.ChannelEmail, models.PriorityLow) {
		t.Error("user2 email should be allowed")
	}

	// user1 SMS: ok (different channel)
	if !rl.Allow(ctx, "user1", models.ChannelSMS, models.PriorityLow) {
		t.Error("user1 SMS should be allowed")
	}
}
