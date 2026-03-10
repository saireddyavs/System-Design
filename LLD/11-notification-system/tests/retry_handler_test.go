package tests

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"notification-system/internal/middleware"
	"notification-system/internal/models"
	"notification-system/internal/senders"
)

func TestRetrySender_SucceedsOnRetry(t *testing.T) {
	ctx := context.Background()
	attempts := int32(0)

	failingSender := &mockSender{
		sendFunc: func(ctx context.Context, n *models.Notification, u *models.User) error {
			if atomic.AddInt32(&attempts, 1) < 2 {
				return errors.New("transient failure")
			}
			return nil
		},
		channel: models.ChannelEmail,
	}

	wrapped := middleware.NewRetrySender(failingSender, 3, 10*time.Millisecond)

	user := models.NewUser("u1", "Test", "test@example.com", "+1", "token")
	notif := &models.Notification{UserID: "u1", Channel: models.ChannelEmail}

	err := wrapped.Send(ctx, notif, user)
	if err != nil {
		t.Fatalf("Expected success after retry: %v", err)
	}
	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetrySender_FailsAfterMaxRetries(t *testing.T) {
	ctx := context.Background()

	failingSender := &mockSender{
		sendFunc: func(ctx context.Context, n *models.Notification, u *models.User) error {
			return errors.New("always fails")
		},
		channel: models.ChannelEmail,
	}

	wrapped := middleware.NewRetrySender(failingSender, 2, 5*time.Millisecond)

	user := models.NewUser("u1", "Test", "test@example.com", "+1", "token")
	notif := &models.Notification{UserID: "u1", Channel: models.ChannelEmail}

	err := wrapped.Send(ctx, notif, user)
	if err == nil {
		t.Fatal("Expected error after max retries")
	}
}

func TestRetrySender_SucceedsFirstTry(t *testing.T) {
	ctx := context.Background()
	realSender := senders.NewEmailSender()
	wrapped := middleware.NewRetrySender(realSender, 3, 10*time.Millisecond)

	user := models.NewUser("u1", "Test", "test@example.com", "+1", "token")
	notif := &models.Notification{UserID: "u1", Channel: models.ChannelEmail}

	err := wrapped.Send(ctx, notif, user)
	if err != nil {
		t.Fatalf("Expected success: %v", err)
	}
}

type mockSender struct {
	sendFunc func(context.Context, *models.Notification, *models.User) error
	channel  models.Channel
}

func (m *mockSender) Send(ctx context.Context, n *models.Notification, u *models.User) error {
	return m.sendFunc(ctx, n, u)
}

func (m *mockSender) Channel() models.Channel {
	return m.channel
}
