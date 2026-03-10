package senders

import (
	"context"
	"fmt"

	"notification-system/internal/models"
)

// EmailSender sends notifications via email
// Strategy Pattern: Implements NotificationSender for email channel
type EmailSender struct {
	// In production: SMTP client, API client, etc.
}

// NewEmailSender creates a new email sender
func NewEmailSender() *EmailSender {
	return &EmailSender{}
}

// Send delivers the notification via email
func (e *EmailSender) Send(ctx context.Context, notification *models.Notification, user *models.User) error {
	// Simulate email delivery - in production: use SMTP, SendGrid, etc.
	if user.Email == "" {
		return fmt.Errorf("user has no email address")
	}
	// In production: send actual email
	// log.Printf("Email sent to %s: %s - %s", user.Email, notification.Subject, notification.Content)
	return nil
}

// Channel returns the email channel
func (e *EmailSender) Channel() models.Channel {
	return models.ChannelEmail
}
