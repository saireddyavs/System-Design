package models

import "time"

// User represents a chat application user
type User struct {
	ID           string     `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Status       UserStatus `json:"status"`
	LastSeen     time.Time  `json:"last_seen"`
	CreatedAt    time.Time  `json:"created_at"`
}

// IsOnline returns true if user is currently online
func (u *User) IsOnline() bool {
	return u.Status == StatusOnline
}
