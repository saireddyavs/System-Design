package models

// Channel represents the notification delivery channel
type Channel string

const (
	ChannelEmail Channel = "email"
	ChannelSMS   Channel = "sms"
	ChannelPush  Channel = "push"
)

// Priority represents notification priority level
type Priority int

const (
	PriorityLow Priority = iota
	PriorityHigh
	PriorityCritical
)

// Status represents notification delivery status
type Status string

const (
	StatusPending  Status = "pending"
	StatusSent     Status = "sent"
	StatusFailed   Status = "failed"
	StatusRetrying Status = "retrying"
)

// NotificationType represents the type/category of notification
type NotificationType string

const (
	TypeAlert NotificationType = "alert"
)
