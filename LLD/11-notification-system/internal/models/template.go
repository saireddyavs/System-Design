package models

// Template represents a notification template with variable placeholders
type Template struct {
	ID      string
	Name    string
	Channel Channel
	Subject string // For email; empty for SMS/Push
	Body    string // Supports {{variable}} placeholders
}

// NewTemplate creates a new notification template
func NewTemplate(id, name string, channel Channel, subject, body string) *Template {
	return &Template{
		ID:      id,
		Name:    name,
		Channel: channel,
		Subject: subject,
		Body:    body,
	}
}
