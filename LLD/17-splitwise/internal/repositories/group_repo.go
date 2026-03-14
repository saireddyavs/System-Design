package repositories

import (
	"fmt"
	"sync"
	"time"

	"splitwise/internal/interfaces"
	"splitwise/internal/models"
)

// InMemoryGroupRepository implements GroupRepository with in-memory storage
type InMemoryGroupRepository struct {
	mu     sync.RWMutex
	groups map[string]*models.Group
}

// NewInMemoryGroupRepository creates a new in-memory group repository
func NewInMemoryGroupRepository() *InMemoryGroupRepository {
	return &InMemoryGroupRepository{
		groups: make(map[string]*models.Group),
	}
}

// Ensure InMemoryGroupRepository implements GroupRepository
var _ interfaces.GroupRepository = (*InMemoryGroupRepository)(nil)

func (r *InMemoryGroupRepository) Create(group *models.Group) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.groups[group.ID]; exists {
		return fmt.Errorf("group already exists: %s", group.ID)
	}
	now := time.Now()
	group.CreatedAt = now
	group.UpdatedAt = now
	r.groups[group.ID] = copyGroup(group)
	return nil
}

func (r *InMemoryGroupRepository) GetByID(id string) (*models.Group, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	g, ok := r.groups[id]
	if !ok {
		return nil, fmt.Errorf("group not found: %s", id)
	}
	return copyGroup(g), nil
}

func (r *InMemoryGroupRepository) GetByUserID(userID string) ([]*models.Group, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*models.Group, 0)
	for _, g := range r.groups {
		for _, mid := range g.MemberIDs {
			if mid == userID {
				result = append(result, copyGroup(g))
				break
			}
		}
	}
	return result, nil
}

func (r *InMemoryGroupRepository) Update(group *models.Group) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.groups[group.ID]; !exists {
		return fmt.Errorf("group not found: %s", group.ID)
	}
	group.UpdatedAt = time.Now()
	r.groups[group.ID] = copyGroup(group)
	return nil
}

func (r *InMemoryGroupRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.groups[id]; !exists {
		return fmt.Errorf("group not found: %s", id)
	}
	delete(r.groups, id)
	return nil
}

func copyGroup(g *models.Group) *models.Group {
	cpy := *g
	cpy.MemberIDs = make([]string, len(g.MemberIDs))
	copy(cpy.MemberIDs, g.MemberIDs)
	return &cpy
}
