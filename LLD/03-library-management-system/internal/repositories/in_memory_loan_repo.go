package repositories

import (
	"errors"
	"library-management-system/internal/models"
	"library-management-system/internal/interfaces"
	"sync"
	"time"
)

var ErrLoanNotFound = errors.New("loan not found")

// InMemoryLoanRepo implements LoanRepository with thread-safe in-memory storage
type InMemoryLoanRepo struct {
	loans map[string]*models.Loan
	mu    sync.RWMutex
}

// NewInMemoryLoanRepo creates a new in-memory loan repository
func NewInMemoryLoanRepo() interfaces.LoanRepository {
	return &InMemoryLoanRepo{
		loans: make(map[string]*models.Loan),
	}
}

func (r *InMemoryLoanRepo) Create(loan *models.Loan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.loans[loan.ID] = loan
	return nil
}

func (r *InMemoryLoanRepo) GetByID(id string) (*models.Loan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	loan, ok := r.loans[id]
	if !ok {
		return nil, ErrLoanNotFound
	}
	return loan, nil
}

func (r *InMemoryLoanRepo) Update(loan *models.Loan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.loans[loan.ID]; !ok {
		return ErrLoanNotFound
	}
	r.loans[loan.ID] = loan
	return nil
}

func (r *InMemoryLoanRepo) GetActiveByBookID(bookID string) (*models.Loan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, l := range r.loans {
		if l.BookID == bookID && l.Status == models.LoanStatusActive {
			return l, nil
		}
	}
	return nil, ErrLoanNotFound
}

func (r *InMemoryLoanRepo) GetOverdueLoans() ([]*models.Loan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	now := time.Now()
	var result []*models.Loan
	for _, l := range r.loans {
		if l.Status == models.LoanStatusActive && now.After(l.DueDate) {
			result = append(result, l)
		}
	}
	return result, nil
}

func (r *InMemoryLoanRepo) GetLoansDueBefore(date time.Time) ([]*models.Loan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Loan
	for _, l := range r.loans {
		if l.Status == models.LoanStatusActive && !l.DueDate.After(date) {
			result = append(result, l)
		}
	}
	return result, nil
}
