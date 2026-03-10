package models

import (
	"sync"
	"time"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeLike          NotificationType = "like"
	NotificationTypeComment      NotificationType = "comment"
	NotificationTypeFriendRequest NotificationType = "friend_request"
	NotificationTypeFriendAccepted NotificationType = "friend_accepted"
)

// Notification represents a user notification.
// S - Single Responsibility: Notification model only holds notification data
type Notification struct {
	ID           string           `json:"id"`
	UserID       string           `json:"user_id"`
	Type         NotificationType `json:"type"`
	SourceUserID string           `json:"source_user_id"`
	TargetID     string           `json:"target_id"` // PostID, CommentID, or FriendshipID
	Read         bool             `json:"read"`
	CreatedAt    time.Time        `json:"created_at"`
	mu           sync.RWMutex
}

// NewNotification creates a new Notification instance (Factory pattern)
func NewNotification(id, userID string, notifType NotificationType, sourceUserID, targetID string) *Notification {
	return &Notification{
		ID:           id,
		UserID:       userID,
		Type:         notifType,
		SourceUserID: sourceUserID,
		TargetID:     targetID,
		Read:         false,
		CreatedAt:    time.Now().UTC(),
	}
}

// MarkAsRead marks the notification as read
func (n *Notification) MarkAsRead() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Read = true
}
