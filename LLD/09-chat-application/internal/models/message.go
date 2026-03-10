package models

import "time"

// Message represents a chat message
type Message struct {
	ID         string        `json:"id"`
	SenderID   string        `json:"sender_id"`
	RoomID     string        `json:"room_id"`
	Content    string        `json:"content"`
	Type       MessageType   `json:"type"`
	Status     MessageStatus `json:"status"`
	Timestamp  time.Time     `json:"timestamp"`
	ReadBy     []string      `json:"read_by"` // User IDs who have read the message
}

// IsReadBy checks if a user has read the message
func (m *Message) IsReadBy(userID string) bool {
	for _, id := range m.ReadBy {
		if id == userID {
			return true
		}
	}
	return false
}
