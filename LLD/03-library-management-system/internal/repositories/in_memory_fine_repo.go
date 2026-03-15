package repositories

import (
	"errors"
	"library-management-system/internal/models"
	"library-management-system/internal/interfaces"
	"sync"
)

var ErrFineNotFound = errors.New("fine not found")

// InMemoryFineRepo implements FineRepository with thread-safe in-memory storage
type InMemoryFineRepo struct {
	fines map[string]*models.Fine
	mu    sync.RWMutex
}

// NewInMemoryFineRepo creates a new in-memory fine repository
func NewInMemoryFineRepo() interfaces.FineRepository {
	return &InMemoryFineRepo{
		fines: make(map[string]*models.Fine),
	}
}

func (r *InMemoryFineRepo) Create(fine *models.Fine) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fines[fine.ID] = fine
	return nil
}

func (r *InMemoryFineRepo) GetByID(id string) (*models.Fine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fine, ok := r.fines[id]
	if !ok {
		return nil, ErrFineNotFound
	}
	return fine, nil
}

func (r *InMemoryFineRepo) Update(fine *models.Fine) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.fines[fine.ID]; !ok {
		return ErrFineNotFound
	}
	r.fines[fine.ID] = fine
	return nil
}

func (r *InMemoryFineRepo) GetByLoanID(loanID string) (*models.Fine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, f := range r.fines {
		if f.LoanID == loanID {
			return f, nil
		}
	}
	return nil, ErrFineNotFound
}

func (r *InMemoryFineRepo) GetPendingByMemberID(memberID string) ([]*models.Fine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Fine
	for _, f := range r.fines {
		if f.MemberID == memberID && f.Status == models.FineStatusPending {
			result = append(result, f)
		}
	}
	return result, nil
}
