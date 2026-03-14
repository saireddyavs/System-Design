package repositories

import (
	"fmt"
	"sync"
	"time"

	"splitwise/internal/interfaces"
	"splitwise/internal/models"
)

// InMemoryTransactionRepository implements TransactionRepository with in-memory storage
type InMemoryTransactionRepository struct {
	mu          sync.RWMutex
	transactions map[string]*models.Transaction
	byUser      map[string][]string
	byGroup     map[string][]string
}

// NewInMemoryTransactionRepository creates a new in-memory transaction repository
func NewInMemoryTransactionRepository() *InMemoryTransactionRepository {
	return &InMemoryTransactionRepository{
		transactions: make(map[string]*models.Transaction),
		byUser:      make(map[string][]string),
		byGroup:     make(map[string][]string),
	}
}

// Ensure InMemoryTransactionRepository implements TransactionRepository
var _ interfaces.TransactionRepository = (*InMemoryTransactionRepository)(nil)

func (r *InMemoryTransactionRepository) Create(tx *models.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.transactions[tx.ID]; exists {
		return fmt.Errorf("transaction already exists: %s", tx.ID)
	}
	tx.CreatedAt = time.Now()
	r.transactions[tx.ID] = copyTransaction(tx)
	r.byUser[tx.FromUserID] = append(r.byUser[tx.FromUserID], tx.ID)
	r.byUser[tx.ToUserID] = append(r.byUser[tx.ToUserID], tx.ID)
	if tx.GroupID != "" {
		r.byGroup[tx.GroupID] = append(r.byGroup[tx.GroupID], tx.ID)
	}
	return nil
}

func (r *InMemoryTransactionRepository) GetByID(id string) (*models.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tx, ok := r.transactions[id]
	if !ok {
		return nil, fmt.Errorf("transaction not found: %s", id)
	}
	return copyTransaction(tx), nil
}

func (r *InMemoryTransactionRepository) GetByUserID(userID string) ([]*models.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids, ok := r.byUser[userID]
	if !ok {
		return []*models.Transaction{}, nil
	}
	result := make([]*models.Transaction, 0, len(ids))
	for _, id := range ids {
		if tx, exists := r.transactions[id]; exists {
			result = append(result, copyTransaction(tx))
		}
	}
	return result, nil
}

func (r *InMemoryTransactionRepository) GetByGroupID(groupID string) ([]*models.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids, ok := r.byGroup[groupID]
	if !ok {
		return []*models.Transaction{}, nil
	}
	result := make([]*models.Transaction, 0, len(ids))
	for _, id := range ids {
		if tx, exists := r.transactions[id]; exists {
			result = append(result, copyTransaction(tx))
		}
	}
	return result, nil
}

func copyTransaction(tx *models.Transaction) *models.Transaction {
	cpy := *tx
	return &cpy
}
