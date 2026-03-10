package services

import (
	"chat-application/internal/apperrors"
	"chat-application/internal/interfaces"
	"chat-application/internal/models"
	"errors"
)

var (
	ErrMessageNotFound = errors.New("message not found")
	ErrNotInRoom       = errors.New("user is not a member of this room")
)

// MessageService handles messaging with real-time delivery (Observer via broker)
type MessageService struct {
	msgRepo    interfaces.MessageRepository
	roomRepo   interfaces.ChatRoomRepository
	broker     interfaces.MessageBroker
	factory    *MessageFactory
}

// NewMessageService creates a new message service
func NewMessageService(
	msgRepo interfaces.MessageRepository,
	roomRepo interfaces.ChatRoomRepository,
	broker interfaces.MessageBroker,
	factory *MessageFactory,
) *MessageService {
	if factory == nil {
		factory = NewMessageFactory()
	}
	return &MessageService{
		msgRepo:  msgRepo,
		roomRepo: roomRepo,
		broker:   broker,
		factory:  factory,
	}
}

// SendMessage sends a message to a room and delivers to online members
func (s *MessageService) SendMessage(senderID, roomID, content string, msgType models.MessageType) (*models.Message, error) {
	room, err := s.roomRepo.GetByID(roomID)
	if err != nil {
		return nil, apperrors.ErrRoomNotFound
	}
	if !room.HasMember(senderID) {
		return nil, ErrNotInRoom
	}

	message := s.factory.Create(senderID, roomID, content, msgType)
	if msgType == "" {
		message.Type = models.MessageTypeText
	}

	if err := s.msgRepo.Create(message); err != nil {
		return nil, err
	}

	// Get recipient IDs (all members except sender)
	recipientIDs := make([]string, 0)
	for _, m := range room.Members {
		if m.UserID != senderID {
			recipientIDs = append(recipientIDs, m.UserID)
		}
	}

	// Real-time delivery via broker (Observer pattern)
	if len(recipientIDs) > 0 {
		_ = s.broker.Publish(message, recipientIDs)
	}

	return message, nil
}

// GetMessageHistory returns paginated message history for a room
func (s *MessageService) GetMessageHistory(roomID, userID string, limit, offset int) ([]*models.Message, error) {
	room, err := s.roomRepo.GetByID(roomID)
	if err != nil {
		return nil, apperrors.ErrRoomNotFound
	}
	if !room.HasMember(userID) {
		return nil, ErrNotInRoom
	}

	if limit <= 0 {
		limit = 50
	}
	return s.msgRepo.GetByRoomID(roomID, limit, offset)
}

// MarkAsRead marks a message as read by user
func (s *MessageService) MarkAsRead(messageID, userID string) error {
	msg, err := s.msgRepo.GetByID(messageID)
	if err != nil {
		return ErrMessageNotFound
	}
	room, err := s.roomRepo.GetByID(msg.RoomID)
	if err != nil {
		return apperrors.ErrRoomNotFound
	}
	if !room.HasMember(userID) {
		return ErrNotInRoom
	}
	return s.msgRepo.MarkAsRead(messageID, userID)
}

// Subscribe returns the message channel for real-time delivery (Observer)
func (s *MessageService) Subscribe(userID string) <-chan *models.Message {
	return s.broker.Subscribe(userID)
}

// Unsubscribe removes user from real-time delivery
func (s *MessageService) Unsubscribe(userID string) {
	s.broker.Unsubscribe(userID)
}

// GetQueuedMessages returns messages queued for offline delivery
func (s *MessageService) GetQueuedMessages(userID string) []*models.Message {
	return s.broker.GetQueuedMessages(userID)
}

// ClearQueuedMessages clears queue after delivering to user
func (s *MessageService) ClearQueuedMessages(userID string) {
	s.broker.ClearQueue(userID)
}
