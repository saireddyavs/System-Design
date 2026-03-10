package repositories

import (
	"errors"
	"library-management-system/internal/models"
	"library-management-system/internal/interfaces"
	"sync"
)

var ErrMemberNotFound = errors.New("member not found")

// InMemoryMemberRepo implements MemberRepository with thread-safe in-memory storage
type InMemoryMemberRepo struct {
	members map[string]*models.Member
	mu      sync.RWMutex
}

// NewInMemoryMemberRepo creates a new in-memory member repository
func NewInMemoryMemberRepo() interfaces.MemberRepository {
	return &InMemoryMemberRepo{
		members: make(map[string]*models.Member),
	}
}

func (r *InMemoryMemberRepo) Create(member *models.Member) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.members[member.ID] = member
	return nil
}

func (r *InMemoryMemberRepo) GetByID(id string) (*models.Member, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	member, ok := r.members[id]
	if !ok {
		return nil, ErrMemberNotFound
	}
	return member, nil
}

func (r *InMemoryMemberRepo) Update(member *models.Member) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.members[member.ID]; !ok {
		return ErrMemberNotFound
	}
	r.members[member.ID] = member
	return nil
}

func (r *InMemoryMemberRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.members, id)
	return nil
}

func (r *InMemoryMemberRepo) ListAll() ([]*models.Member, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	members := make([]*models.Member, 0, len(r.members))
	for _, m := range r.members {
		members = append(members, m)
	}
	return members, nil
}

func (r *InMemoryMemberRepo) GetByEmail(email string) (*models.Member, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, m := range r.members {
		if m.Email == email {
			return m, nil
		}
	}
	return nil, ErrMemberNotFound
}
