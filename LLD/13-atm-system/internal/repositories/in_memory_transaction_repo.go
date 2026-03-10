package repositories

import (
	"atm-system/internal/models"
	"atm-system/internal/interfaces"
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

var ErrTransactionNotFound = errors.New("transaction not found")

// InMemoryTransactionRepository implements TransactionRepository (Repository Pattern)
type InMemoryTransactionRepository struct {
	transactions map[string]*models.Transaction
	byAccount    map[string][]string // accountID -> transaction IDs (ordered by time)
	mu           sync.RWMutex
}

// NewInMemoryTransactionRepository creates a new in-memory transaction repository
func NewInMemoryTransactionRepository() *InMemoryTransactionRepository {
	return &InMemoryTransactionRepository{
		transactions: make(map[string]*models.Transaction),
		byAccount:    make(map[string][]string),
	}
}

var _ interfaces.TransactionRepository = (*InMemoryTransactionRepository)(nil)

func (r *InMemoryTransactionRepository) Save(ctx context.Context, tx *models.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.transactions[tx.ID] = tx
	r.byAccount[tx.AccountID] = append(r.byAccount[tx.AccountID], tx.ID)
	return nil
}

func (r *InMemoryTransactionRepository) GetByAccountID(ctx context.Context, accountID string, limit int) ([]*models.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids, ok := r.byAccount[accountID]
	if !ok {
		return []*models.Transaction{}, nil
	}
	// Get most recent first
	start := len(ids) - limit
	if start < 0 {
		start = 0
	}
	result := make([]*models.Transaction, 0, limit)
	for i := len(ids) - 1; i >= start && len(result) < limit; i-- {
		if tx, ok := r.transactions[ids[i]]; ok {
			result = append(result, tx)
		}
	}
	return result, nil
}

func (r *InMemoryTransactionRepository) GetByID(ctx context.Context, id string) (*models.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tx, ok := r.transactions[id]
	if !ok {
		return nil, ErrTransactionNotFound
	}
	return tx, nil
}

func (r *InMemoryTransactionRepository) GetByAccountIDAndDateRange(ctx context.Context, accountID string, from, to time.Time) ([]*models.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids, ok := r.byAccount[accountID]
	if !ok {
		return []*models.Transaction{}, nil
	}
	result := make([]*models.Transaction, 0)
	for _, id := range ids {
		tx, ok := r.transactions[id]
		if !ok {
			continue
		}
		if !tx.Timestamp.Before(from) && !tx.Timestamp.After(to) {
			result = append(result, tx)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.After(result[j].Timestamp)
	})
	return result, nil
}
