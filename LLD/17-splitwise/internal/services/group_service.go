package services

import (
	"fmt"
	"sync"
	"time"

	"splitwise/internal/interfaces"
	"splitwise/internal/models"
)

// GroupService implements group management
type GroupService struct {
	groupRepo interfaces.GroupRepository
	userRepo  interfaces.UserRepository
	mu        sync.RWMutex
	idSeq     int
}

// NewGroupService creates a new GroupService
func NewGroupService(groupRepo interfaces.GroupRepository, userRepo interfaces.UserRepository) *GroupService {
	return &GroupService{
		groupRepo: groupRepo,
		userRepo:  userRepo,
		idSeq:     1,
	}
}

// Ensure GroupService implements interfaces.GroupService
var _ interfaces.GroupService = (*GroupService)(nil)

// CreateGroup creates a new group
func (s *GroupService) CreateGroup(name, description, createdBy string, memberIDs []string) (*models.Group, error) {
	// Validate creator is in members
	hasCreator := false
	for _, id := range memberIDs {
		if id == createdBy {
			hasCreator = true
			break
		}
	}
	if !hasCreator {
		memberIDs = append([]string{createdBy}, memberIDs...)
	}

	s.mu.Lock()
	id := fmt.Sprintf("group%d", s.idSeq)
	s.idSeq++
	s.mu.Unlock()

	group := &models.Group{
		ID:          id,
		Name:        name,
		Description: description,
		MemberIDs:   memberIDs,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := s.groupRepo.Create(group); err != nil {
		return nil, err
	}
	return group, nil
}

// GetGroup retrieves a group by ID
func (s *GroupService) GetGroup(id string) (*models.Group, error) {
	return s.groupRepo.GetByID(id)
}

// AddMember adds a member to the group
func (s *GroupService) AddMember(groupID, userID string) error {
	group, err := s.groupRepo.GetByID(groupID)
	if err != nil {
		return err
	}
	for _, mid := range group.MemberIDs {
		if mid == userID {
			return fmt.Errorf("user %s already in group", userID)
		}
	}
	group.MemberIDs = append(group.MemberIDs, userID)
	return s.groupRepo.Update(group)
}

// RemoveMember removes a member from the group
func (s *GroupService) RemoveMember(groupID, userID string) error {
	group, err := s.groupRepo.GetByID(groupID)
	if err != nil {
		return err
	}
	newMembers := make([]string, 0, len(group.MemberIDs))
	for _, mid := range group.MemberIDs {
		if mid != userID {
			newMembers = append(newMembers, mid)
		}
	}
	if len(newMembers) == len(group.MemberIDs) {
		return fmt.Errorf("user %s not in group", userID)
	}
	group.MemberIDs = newMembers
	return s.groupRepo.Update(group)
}
