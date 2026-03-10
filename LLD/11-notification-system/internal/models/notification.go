package models

import "time"

// Notification represents a notification to be sent
type Notification struct {
	ID         string
	UserID     string
	Channel    Channel
	Type       NotificationType
	Priority   Priority
	Subject    string
	Content    string
	Status     Status
	CreatedAt  time.Time
	SentAt     *time.Time
	RetryCount int
	Error      string // Last error if failed
}

// IsCritical returns true if notification has critical priority
func (n *Notification) IsCritical() bool {
	return n.Priority == PriorityCritical
}

// CanRetry returns true if notification can be retried
func (n *Notification) CanRetry(maxRetries int) bool {
	return n.RetryCount < maxRetries
}
