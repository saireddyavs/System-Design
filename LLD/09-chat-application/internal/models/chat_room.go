package models

import "time"

// RoomMember represents a member in a chat room
type RoomMember struct {
	UserID    string     `json:"user_id"`
	Role      MemberRole `json:"role"`
	JoinedAt  time.Time  `json:"joined_at"`
}

// ChatRoom represents a chat room (one-on-one or group)
type ChatRoom struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Type      ChatRoomType  `json:"type"`
	Members   []RoomMember  `json:"members"`
	CreatedBy string        `json:"created_by"`
	CreatedAt time.Time     `json:"created_at"`
}

const MaxGroupMembers = 100

// IsGroup returns true if this is a group chat
func (r *ChatRoom) IsGroup() bool {
	return r.Type == ChatRoomTypeGroup
}

// HasMember checks if user is a member
func (r *ChatRoom) HasMember(userID string) bool {
	for _, m := range r.Members {
		if m.UserID == userID {
			return true
		}
	}
	return false
}

// CanManageMembers checks if user can add/remove members (creator or admin)
func (r *ChatRoom) CanManageMembers(userID string) bool {
	for _, m := range r.Members {
		if m.UserID == userID && (m.Role == RoleCreator || m.Role == RoleAdmin) {
			return true
		}
	}
	return false
}
