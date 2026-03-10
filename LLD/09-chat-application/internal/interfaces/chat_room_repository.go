package interfaces

import "chat-application/internal/models"

// ChatRoomRepository defines data access for chat rooms (SRP, DIP)
type ChatRoomRepository interface {
	Create(room *models.ChatRoom) error
	GetByID(id string) (*models.ChatRoom, error)
	GetByUserID(userID string) ([]*models.ChatRoom, error)
	Update(room *models.ChatRoom) error
	GetOneOnOneRoom(user1ID, user2ID string) (*models.ChatRoom, error)
}
