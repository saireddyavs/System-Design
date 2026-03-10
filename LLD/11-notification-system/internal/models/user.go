package models

import "sync"

// User represents a user who can receive notifications
type User struct {
	ID           string
	Name         string
	Email        string
	Phone        string
	DeviceToken  string
	Preferences  map[Channel]bool // channel -> enabled
	preferencesMu sync.RWMutex
}

// NewUser creates a new user with default preferences (all channels enabled)
func NewUser(id, name, email, phone, deviceToken string) *User {
	return &User{
		ID:          id,
		Name:        name,
		Email:       email,
		Phone:       phone,
		DeviceToken: deviceToken,
		Preferences: map[Channel]bool{
			ChannelEmail: true,
			ChannelSMS:   true,
			ChannelPush:  true,
		},
	}
}

// IsChannelEnabled returns whether the user has enabled the given channel
func (u *User) IsChannelEnabled(channel Channel) bool {
	u.preferencesMu.RLock()
	defer u.preferencesMu.RUnlock()
	enabled, ok := u.Preferences[channel]
	return ok && enabled
}

// SetChannelPreference updates the preference for a channel
func (u *User) SetChannelPreference(channel Channel, enabled bool) {
	u.preferencesMu.Lock()
	defer u.preferencesMu.Unlock()
	u.Preferences[channel] = enabled
}
