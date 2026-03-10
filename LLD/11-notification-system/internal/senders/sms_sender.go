package senders

import (
	"context"
	"fmt"

	"notification-system/internal/models"
)

// SMSSender sends notifications via SMS
// Strategy Pattern: Implements NotificationSender for SMS channel
type SMSSender struct {
	// In production: Twilio, AWS SNS, etc.
}

// NewSMSSender creates a new SMS sender
func NewSMSSender() *SMSSender {
	return &SMSSender{}
}

// Send delivers the notification via SMS
func (s *SMSSender) Send(ctx context.Context, notification *models.Notification, user *models.User) error {
	if user.Phone == "" {
		return fmt.Errorf("user has no phone number")
	}
	// In production: send actual SMS via provider API
	// log.Printf("SMS sent to %s: %s", user.Phone, notification.Content)
	return nil
}

// Channel returns the SMS channel
func (s *SMSSender) Channel() models.Channel {
	return models.ChannelSMS
}
