package repositories

import (
	"fmt"
	"sync"
	"time"

	"splitwise/internal/interfaces"
	"splitwise/internal/models"
)

// InMemoryExpenseRepository implements ExpenseRepository with in-memory storage
type InMemoryExpenseRepository struct {
	mu       sync.RWMutex
	expenses map[string]*models.Expense
	byGroup  map[string][]string
	byUser   map[string][]string
}

// NewInMemoryExpenseRepository creates a new in-memory expense repository
func NewInMemoryExpenseRepository() *InMemoryExpenseRepository {
	return &InMemoryExpenseRepository{
		expenses: make(map[string]*models.Expense),
		byGroup:  make(map[string][]string),
		byUser:   make(map[string][]string),
	}
}

// Ensure InMemoryExpenseRepository implements ExpenseRepository
var _ interfaces.ExpenseRepository = (*InMemoryExpenseRepository)(nil)

func (r *InMemoryExpenseRepository) Create(expense *models.Expense) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.expenses[expense.ID]; exists {
		return fmt.Errorf("expense already exists: %s", expense.ID)
	}
	now := time.Now()
	expense.CreatedAt = now
	expense.UpdatedAt = now
	r.expenses[expense.ID] = copyExpense(expense)

	if expense.GroupID != "" {
		r.byGroup[expense.GroupID] = append(r.byGroup[expense.GroupID], expense.ID)
	}

	// Index by paidBy and all participants
	userIDs := make(map[string]bool)
	userIDs[expense.PaidBy] = true
	for _, s := range expense.Splits {
		userIDs[s.UserID] = true
	}
	for uid := range userIDs {
		r.byUser[uid] = append(r.byUser[uid], expense.ID)
	}

	return nil
}

func (r *InMemoryExpenseRepository) GetByID(id string) (*models.Expense, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.expenses[id]
	if !ok {
		return nil, fmt.Errorf("expense not found: %s", id)
	}
	return copyExpense(e), nil
}

func (r *InMemoryExpenseRepository) GetByGroupID(groupID string) ([]*models.Expense, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids, ok := r.byGroup[groupID]
	if !ok {
		return []*models.Expense{}, nil
	}
	result := make([]*models.Expense, 0, len(ids))
	for _, id := range ids {
		if e, exists := r.expenses[id]; exists {
			result = append(result, copyExpense(e))
		}
	}
	return result, nil
}

func (r *InMemoryExpenseRepository) GetByUserID(userID string) ([]*models.Expense, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids, ok := r.byUser[userID]
	if !ok {
		return []*models.Expense{}, nil
	}
	result := make([]*models.Expense, 0, len(ids))
	for _, id := range ids {
		if e, exists := r.expenses[id]; exists {
			result = append(result, copyExpense(e))
		}
	}
	return result, nil
}

func (r *InMemoryExpenseRepository) GetBetweenUsers(userID1, userID2 string) ([]*models.Expense, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*models.Expense, 0)
	for _, e := range r.expenses {
		has1 := e.PaidBy == userID1
		has2 := e.PaidBy == userID2
		for _, s := range e.Splits {
			if s.UserID == userID1 {
				has1 = true
			}
			if s.UserID == userID2 {
				has2 = true
			}
		}
		if has1 && has2 {
			result = append(result, copyExpense(e))
		}
	}
	return result, nil
}

func (r *InMemoryExpenseRepository) Update(expense *models.Expense) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.expenses[expense.ID]; !exists {
		return fmt.Errorf("expense not found: %s", expense.ID)
	}
	expense.UpdatedAt = time.Now()
	r.expenses[expense.ID] = copyExpense(expense)
	return nil
}

func (r *InMemoryExpenseRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.expenses[id]; !exists {
		return fmt.Errorf("expense not found: %s", id)
	}
	delete(r.expenses, id)
	for gid, ids := range r.byGroup {
		r.byGroup[gid] = filterOut(ids, id)
	}
	for uid, ids := range r.byUser {
		r.byUser[uid] = filterOut(ids, id)
	}
	return nil
}

func copyExpense(e *models.Expense) *models.Expense {
	cpy := *e
	cpy.Splits = make([]models.Split, len(e.Splits))
	copy(cpy.Splits, e.Splits)
	return &cpy
}

func filterOut(ids []string, target string) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		if id != target {
			result = append(result, id)
		}
	}
	return result
}
