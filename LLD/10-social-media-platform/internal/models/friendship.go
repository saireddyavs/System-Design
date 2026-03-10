package models

import (
	"sync"
	"time"
)

// FriendshipStatus represents the state of a friendship request
type FriendshipStatus string

const (
	FriendshipStatusPending  FriendshipStatus = "pending"
	FriendshipStatusAccepted FriendshipStatus = "accepted"
	FriendshipStatusRejected FriendshipStatus = "rejected"
)

// Friendship represents a friend/follow relationship between users.
// S - Single Responsibility: Friendship model only holds relationship data
type Friendship struct {
	ID         string           `json:"id"`
	RequesterID string          `json:"requester_id"`
	ReceiverID string          `json:"receiver_id"`
	Status     FriendshipStatus `json:"status"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
	mu         sync.RWMutex
}

// NewFriendship creates a new Friendship instance (Factory pattern)
func NewFriendship(id, requesterID, receiverID string) *Friendship {
	now := time.Now().UTC()
	return &Friendship{
		ID:          id,
		RequesterID: requesterID,
		ReceiverID:  receiverID,
		Status:      FriendshipStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Accept marks the friendship as accepted
func (f *Friendship) Accept() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Status = FriendshipStatusAccepted
	f.UpdatedAt = time.Now().UTC()
}

// Reject marks the friendship as rejected
func (f *Friendship) Reject() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Status = FriendshipStatusRejected
	f.UpdatedAt = time.Now().UTC()
}
