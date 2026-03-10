package models

import "sync"

const DefaultStorageQuota = 1 * 1024 * 1024 * 1024 // 1GB in bytes

// User represents a user in the file storage system with storage quota tracking.
type User struct {
	mu           sync.RWMutex
	ID           string
	Name         string
	Email        string
	StorageQuota int64  // in bytes
	UsedStorage  int64  // in bytes
}

// NewUser creates a new user with default storage quota.
func NewUser(id, name, email string) *User {
	return &User{
		ID:           id,
		Name:         name,
		Email:        email,
		StorageQuota: DefaultStorageQuota,
		UsedStorage:  0,
	}
}

// GetAvailableStorage returns remaining storage quota.
func (u *User) GetAvailableStorage() int64 {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.StorageQuota - u.UsedStorage
}

// AddUsedStorage atomically adds to used storage.
func (u *User) AddUsedStorage(amount int64) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.UsedStorage += amount
}

// RemoveUsedStorage atomically removes from used storage.
func (u *User) RemoveUsedStorage(amount int64) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.UsedStorage-amount < 0 {
		u.UsedStorage = 0
	} else {
		u.UsedStorage -= amount
	}
}
