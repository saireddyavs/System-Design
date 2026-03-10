package interfaces

import "chat-application/internal/models"

// MessageBroker defines real-time message delivery (Observer pattern)
// Subscribers receive messages via channels
type MessageBroker interface {
	// Subscribe adds a subscriber for a user - returns channel for receiving messages
	Subscribe(userID string) <-chan *models.Message
	// Unsubscribe removes a subscriber
	Unsubscribe(userID string)
	// Publish delivers message to recipient(s) - uses Strategy for direct vs broadcast
	Publish(message *models.Message, recipientIDs []string) error
	// QueueForOffline queues message for offline users
	QueueForOffline(userID string, message *models.Message)
	// GetQueuedMessages returns queued messages when user comes online
	GetQueuedMessages(userID string) []*models.Message
	// ClearQueue removes queued messages after delivery
	ClearQueue(userID string)
}
