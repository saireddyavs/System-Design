package repositories

import (
	"chat-application/internal/interfaces"
	"chat-application/internal/models"
	"sync"
)

// InMemoryMessageRepository implements MessageRepository with in-memory storage
type InMemoryMessageRepository struct {
	messages map[string]*models.Message
	byRoom   map[string][]string // roomID -> message IDs (ordered)
	mu       sync.RWMutex
}

// NewInMemoryMessageRepository creates a new in-memory message repository
func NewInMemoryMessageRepository() interfaces.MessageRepository {
	return &InMemoryMessageRepository{
		messages: make(map[string]*models.Message),
		byRoom:   make(map[string][]string),
	}
}

func (r *InMemoryMessageRepository) Create(message *models.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.messages[message.ID]; exists {
		return ErrAlreadyExists
	}

	msgCopy := *message
	if msgCopy.ReadBy == nil {
		msgCopy.ReadBy = []string{}
	}
	r.messages[message.ID] = &msgCopy
	r.byRoom[message.RoomID] = append(r.byRoom[message.RoomID], message.ID)
	return nil
}

func (r *InMemoryMessageRepository) GetByID(id string) (*models.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	msg, exists := r.messages[id]
	if !exists {
		return nil, ErrNotFound
	}
	msgCopy := *msg
	readByCopy := make([]string, len(msg.ReadBy))
	copy(readByCopy, msg.ReadBy)
	msgCopy.ReadBy = readByCopy
	return &msgCopy, nil
}

func (r *InMemoryMessageRepository) GetByRoomID(roomID string, limit, offset int) ([]*models.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	msgIDs, exists := r.byRoom[roomID]
	if !exists {
		return []*models.Message{}, nil
	}

	// Messages are in chronological order (append order)
	total := len(msgIDs)
	if offset >= total {
		return []*models.Message{}, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	var messages []*models.Message
	for i := offset; i < end; i++ {
		msgID := msgIDs[i]
		if msg, ok := r.messages[msgID]; ok {
			msgCopy := *msg
			readByCopy := make([]string, len(msg.ReadBy))
			copy(readByCopy, msg.ReadBy)
			msgCopy.ReadBy = readByCopy
			messages = append(messages, &msgCopy)
		}
	}
	return messages, nil
}

func (r *InMemoryMessageRepository) Update(message *models.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.messages[message.ID]; !exists {
		return ErrNotFound
	}

	msgCopy := *message
	r.messages[message.ID] = &msgCopy
	return nil
}

func (r *InMemoryMessageRepository) UpdateStatus(messageID string, status models.MessageStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	msg, exists := r.messages[messageID]
	if !exists {
		return ErrNotFound
	}
	msg.Status = status
	return nil
}

func (r *InMemoryMessageRepository) MarkAsRead(messageID string, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	msg, exists := r.messages[messageID]
	if !exists {
		return ErrNotFound
	}
	if !msg.IsReadBy(userID) {
		msg.ReadBy = append(msg.ReadBy, userID)
	}
	msg.Status = models.MessageStatusRead
	return nil
}

// GetMessageCount returns total messages for testing
func (r *InMemoryMessageRepository) GetMessageCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.messages)
}
