package models

import (
	"sync"
	"time"
)

// User represents a user profile in the social media platform.
// S - Single Responsibility: User model only holds user data
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	Bio          string    `json:"bio"`
	ProfilePicURL string   `json:"profile_pic_url"`
	CreatedAt    time.Time `json:"created_at"`
	IsActive     bool      `json:"is_active"`
	mu           sync.RWMutex
}

// NewUser creates a new User instance (Factory pattern)
func NewUser(id, username, email, bio, profilePicURL string) *User {
	return &User{
		ID:            id,
		Username:      username,
		Email:         email,
		Bio:           bio,
		ProfilePicURL: profilePicURL,
		CreatedAt:     time.Now().UTC(),
		IsActive:      true,
	}
}

// UpdateProfile updates user profile fields (Facade: encapsulates profile update logic)
func (u *User) UpdateProfile(bio, profilePicURL string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if bio != "" {
		u.Bio = bio
	}
	if profilePicURL != "" {
		u.ProfilePicURL = profilePicURL
	}
}

// Deactivate marks the user as inactive
func (u *User) Deactivate() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.IsActive = false
}
