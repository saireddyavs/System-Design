package senders

import (
	"context"
	"fmt"

	"notification-system/internal/models"
)

// PushSender sends push notifications to mobile devices
// Strategy Pattern: Implements NotificationSender for push channel
type PushSender struct {
	// In production: FCM, APNs, OneSignal, etc.
}

// NewPushSender creates a new push notification sender
func NewPushSender() *PushSender {
	return &PushSender{}
}

// Send delivers the push notification to the user's device
func (p *PushSender) Send(ctx context.Context, notification *models.Notification, user *models.User) error {
	if user.DeviceToken == "" {
		return fmt.Errorf("user has no device token")
	}
	// In production: send via FCM/APNs
	// log.Printf("Push sent to device %s: %s", user.DeviceToken, notification.Content)
	return nil
}

// Channel returns the push channel
func (p *PushSender) Channel() models.Channel {
	return models.ChannelPush
}
