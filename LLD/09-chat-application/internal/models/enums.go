package models

// UserStatus represents online/offline status
type UserStatus string

const (
	StatusOnline  UserStatus = "online"
	StatusOffline UserStatus = "offline"
)

// MessageType represents the type of message content
type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeImage MessageType = "image"
	MessageTypeFile  MessageType = "file"
)

// MessageStatus represents delivery/read status
type MessageStatus string

const (
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
)

// ChatRoomType represents one-on-one vs group chat
type ChatRoomType string

const (
	ChatRoomTypeOneOnOne ChatRoomType = "one_on_one"
	ChatRoomTypeGroup    ChatRoomType = "group"
)

// MemberRole represents user role in a group chat
type MemberRole string

const (
	RoleMember MemberRole = "member"
	RoleAdmin  MemberRole = "admin"
	RoleCreator MemberRole = "creator"
)
