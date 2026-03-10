package interfaces

import "chat-application/internal/models"

// MessageRepository defines data access for messages (SRP, DIP)
type MessageRepository interface {
	Create(message *models.Message) error
	GetByID(id string) (*models.Message, error)
	GetByRoomID(roomID string, limit, offset int) ([]*models.Message, error)
	Update(message *models.Message) error
	UpdateStatus(messageID string, status models.MessageStatus) error
	MarkAsRead(messageID string, userID string) error
}
