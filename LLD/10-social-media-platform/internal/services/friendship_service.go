package services

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"social-media-platform/internal/interfaces"
	"social-media-platform/internal/models"
)

var (
	ErrInvalidFriendshipRequest = errors.New("invalid friendship request")
	ErrCannotFriendSelf         = errors.New("cannot send friend request to self")
)

// FriendshipService handles friend/follow operations
type FriendshipService struct {
	userRepo       interfaces.UserRepository
	friendshipRepo interfaces.FriendshipRepository
	notifService   *NotificationService
	publisher      interfaces.NotificationPublisher
	nextID         atomic.Uint64
	mu             sync.Mutex
}

// NewFriendshipService creates a new friendship service
func NewFriendshipService(
	userRepo interfaces.UserRepository,
	friendshipRepo interfaces.FriendshipRepository,
	notifService *NotificationService,
	publisher interfaces.NotificationPublisher,
) *FriendshipService {
	return &FriendshipService{
		userRepo:       userRepo,
		friendshipRepo: friendshipRepo,
		notifService:   notifService,
		publisher:      publisher,
	}
}

// SendFriendRequest sends a friend request from requester to receiver
func (s *FriendshipService) SendFriendRequest(requesterID, receiverID string) (*models.Friendship, error) {
	if requesterID == "" || receiverID == "" {
		return nil, ErrInvalidFriendshipRequest
	}
	if requesterID == receiverID {
		return nil, ErrCannotFriendSelf
	}

	// Verify both users exist
	if _, err := s.userRepo.GetByID(requesterID); err != nil {
		return nil, err
	}
	if _, err := s.userRepo.GetByID(receiverID); err != nil {
		return nil, err
	}

	// Check if friendship already exists
	existing, err := s.friendshipRepo.GetFriendship(requesterID, receiverID)
	if err == nil {
		if existing.Status == models.FriendshipStatusAccepted {
			return nil, errors.New("already friends")
		}
		if existing.Status == models.FriendshipStatusPending {
			return nil, errors.New("friend request already pending")
		}
	}

	s.mu.Lock()
	id := fmt.Sprintf("friendship-%d", s.nextID.Add(1))
	s.mu.Unlock()

	friendship := models.NewFriendship(id, requesterID, receiverID)
	if err := s.friendshipRepo.Create(friendship); err != nil {
		return nil, err
	}

	// Observer: Publish notification for friend request
	notif := models.NewNotification(
		s.notifService.GenerateID(),
		receiverID,
		models.NotificationTypeFriendRequest,
		requesterID,
		friendship.ID,
	)
	_ = s.notifService.CreateAndPublish(notif)

	return friendship, nil
}

// AcceptFriendRequest accepts a pending friend request
func (s *FriendshipService) AcceptFriendRequest(receiverID, friendshipID string) error {
	friendship, err := s.friendshipRepo.GetByID(friendshipID)
	if err != nil {
		return err
	}
	if friendship.ReceiverID != receiverID {
		return errors.New("unauthorized to accept this request")
	}
	if friendship.Status != models.FriendshipStatusPending {
		return errors.New("friendship is not pending")
	}

	friendship.Accept()
	if err := s.friendshipRepo.Update(friendship); err != nil {
		return err
	}

	// Observer: Notify requester that request was accepted
	notif := models.NewNotification(
		s.notifService.GenerateID(),
		friendship.RequesterID,
		models.NotificationTypeFriendAccepted,
		receiverID,
		friendship.ID,
	)
	_ = s.notifService.CreateAndPublish(notif)

	return nil
}

// RejectFriendRequest rejects a pending friend request
func (s *FriendshipService) RejectFriendRequest(receiverID, friendshipID string) error {
	friendship, err := s.friendshipRepo.GetByID(friendshipID)
	if err != nil {
		return err
	}
	if friendship.ReceiverID != receiverID {
		return errors.New("unauthorized to reject this request")
	}
	if friendship.Status != models.FriendshipStatusPending {
		return errors.New("friendship is not pending")
	}

	friendship.Reject()
	return s.friendshipRepo.Update(friendship)
}

// Unfollow removes a friendship (unfollow)
func (s *FriendshipService) Unfollow(requesterID, receiverID string) error {
	return s.friendshipRepo.Delete(requesterID, receiverID)
}

// GetPendingRequests returns pending friend requests for a user
func (s *FriendshipService) GetPendingRequests(userID string) ([]*models.Friendship, error) {
	return s.friendshipRepo.GetPendingForUser(userID)
}

// GetFriends returns accepted friends for a user
func (s *FriendshipService) GetFriends(userID string) ([]string, error) {
	return s.friendshipRepo.GetAcceptedFriends(userID)
}

// GetFollowers returns users who follow the given user
func (s *FriendshipService) GetFollowers(userID string) ([]string, error) {
	return s.friendshipRepo.GetFollowers(userID)
}

// GetFollowing returns users who the given user follows
func (s *FriendshipService) GetFollowing(userID string) ([]string, error) {
	return s.friendshipRepo.GetFollowing(userID)
}
