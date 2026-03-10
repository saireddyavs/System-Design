package services

import (
	"chat-application/internal/apperrors"
	"chat-application/internal/interfaces"
	"chat-application/internal/models"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotAMember        = errors.New("user is not a member")
	ErrAlreadyMember     = errors.New("user is already a member")
	ErrCannotManage      = errors.New("only creator or admin can add/remove members")
	ErrMaxMembersReached = errors.New("group chat max 100 members reached")
)

// ChatService mediates chat room operations (Mediator pattern - room mediates between members)
type ChatService struct {
	roomRepo   interfaces.ChatRoomRepository
	userRepo   interfaces.UserRepository
}

// NewChatService creates a new chat service
func NewChatService(roomRepo interfaces.ChatRoomRepository, userRepo interfaces.UserRepository) *ChatService {
	return &ChatService{
		roomRepo: roomRepo,
		userRepo: userRepo,
	}
}

// CreateOneOnOneRoom creates or gets existing 1:1 room between two users
func (s *ChatService) CreateOneOnOneRoom(user1ID, user2ID string) (*models.ChatRoom, error) {
	if room, err := s.roomRepo.GetOneOnOneRoom(user1ID, user2ID); err == nil {
		return room, nil
	}

	room := &models.ChatRoom{
		ID:        uuid.New().String(),
		Name:      "",
		Type:      models.ChatRoomTypeOneOnOne,
		CreatedBy: user1ID,
		CreatedAt: time.Now(),
		Members: []models.RoomMember{
			{UserID: user1ID, Role: models.RoleMember, JoinedAt: time.Now()},
			{UserID: user2ID, Role: models.RoleMember, JoinedAt: time.Now()},
		},
	}
	if err := s.roomRepo.Create(room); err != nil {
		return nil, err
	}
	return room, nil
}

// CreateGroupRoom creates a new group chat
func (s *ChatService) CreateGroupRoom(creatorID, name string, memberIDs []string) (*models.ChatRoom, error) {
	totalMembers := len(memberIDs) + 1
	if totalMembers > models.MaxGroupMembers {
		return nil, ErrMaxMembersReached
	}

	members := []models.RoomMember{
		{UserID: creatorID, Role: models.RoleCreator, JoinedAt: time.Now()},
	}
	seen := map[string]bool{creatorID: true}
	for _, id := range memberIDs {
		if !seen[id] {
			seen[id] = true
			members = append(members, models.RoomMember{UserID: id, Role: models.RoleMember, JoinedAt: time.Now()})
		}
	}

	room := &models.ChatRoom{
		ID:        uuid.New().String(),
		Name:      name,
		Type:      models.ChatRoomTypeGroup,
		Members:   members,
		CreatedBy: creatorID,
		CreatedAt: time.Now(),
	}
	if err := s.roomRepo.Create(room); err != nil {
		return nil, err
	}
	return room, nil
}

// AddMember adds a member to group (only creator/admin)
func (s *ChatService) AddMember(roomID, actorID, newMemberID string) error {
	room, err := s.roomRepo.GetByID(roomID)
	if err != nil {
		return apperrors.ErrRoomNotFound
	}
	if !room.IsGroup() {
		return errors.New("cannot add members to one-on-one chat")
	}
	if !room.CanManageMembers(actorID) {
		return ErrCannotManage
	}
	if room.HasMember(newMemberID) {
		return ErrAlreadyMember
	}
	if len(room.Members) >= models.MaxGroupMembers {
		return ErrMaxMembersReached
	}

	room.Members = append(room.Members, models.RoomMember{
		UserID: newMemberID, Role: models.RoleMember, JoinedAt: time.Now(),
	})
	return s.roomRepo.Update(room)
}

// RemoveMember removes a member from group
func (s *ChatService) RemoveMember(roomID, actorID, memberToRemoveID string) error {
	room, err := s.roomRepo.GetByID(roomID)
	if err != nil {
		return apperrors.ErrRoomNotFound
	}
	if !room.IsGroup() {
		return errors.New("cannot remove members from one-on-one chat")
	}
	if !room.CanManageMembers(actorID) {
		return ErrCannotManage
	}
	if !room.HasMember(memberToRemoveID) {
		return ErrNotAMember
	}

	var newMembers []models.RoomMember
	for _, m := range room.Members {
		if m.UserID != memberToRemoveID {
			newMembers = append(newMembers, m)
		}
	}
	room.Members = newMembers
	return s.roomRepo.Update(room)
}

// LeaveRoom allows a member to leave a group
func (s *ChatService) LeaveRoom(roomID, userID string) error {
	room, err := s.roomRepo.GetByID(roomID)
	if err != nil {
		return apperrors.ErrRoomNotFound
	}
	if !room.HasMember(userID) {
		return ErrNotAMember
	}
	if room.Type == models.ChatRoomTypeOneOnOne {
		return errors.New("cannot leave one-on-one chat")
	}

	var newMembers []models.RoomMember
	for _, m := range room.Members {
		if m.UserID != userID {
			newMembers = append(newMembers, m)
		}
	}
	room.Members = newMembers
	return s.roomRepo.Update(room)
}

// GetUserRooms returns all rooms for a user
func (s *ChatService) GetUserRooms(userID string) ([]*models.ChatRoom, error) {
	return s.roomRepo.GetByUserID(userID)
}

// GetRoom returns room by ID
func (s *ChatService) GetRoom(roomID string) (*models.ChatRoom, error) {
	room, err := s.roomRepo.GetByID(roomID)
	if err != nil {
		return nil, apperrors.ErrRoomNotFound
	}
	return room, nil
}
