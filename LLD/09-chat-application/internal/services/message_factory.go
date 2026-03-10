package services

import (
	"chat-application/internal/models"
	"github.com/google/uuid"
	"time"
)

// MessageFactory creates messages (Factory pattern - encapsulates creation logic)
type MessageFactory struct{}

// NewMessageFactory creates a new message factory
func NewMessageFactory() *MessageFactory {
	return &MessageFactory{}
}

// Create creates a new message with default values
func (f *MessageFactory) Create(senderID, roomID, content string, msgType models.MessageType) *models.Message {
	now := time.Now()
	return &models.Message{
		ID:        uuid.New().String(),
		SenderID:  senderID,
		RoomID:    roomID,
		Content:   content,
		Type:      msgType,
		Status:    models.MessageStatusSent,
		Timestamp: now,
		ReadBy:    []string{},
	}
}
