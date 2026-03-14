package repositories

import (
	"fmt"
	"sync"
	"time"

	"splitwise/internal/interfaces"
	"splitwise/internal/models"
)

// balanceKey is used for map lookup
type balanceKey struct {
	DebtorID   string
	CreditorID string
	GroupID    string
}

// InMemoryBalanceRepository implements BalanceRepository with in-memory storage
type InMemoryBalanceRepository struct {
	mu       sync.RWMutex
	balances map[balanceKey]*models.Balance
	byUser   map[string][]balanceKey
	byGroup  map[string][]balanceKey
}

// NewInMemoryBalanceRepository creates a new in-memory balance repository
func NewInMemoryBalanceRepository() *InMemoryBalanceRepository {
	return &InMemoryBalanceRepository{
		balances: make(map[balanceKey]*models.Balance),
		byUser:   make(map[string][]balanceKey),
		byGroup:  make(map[string][]balanceKey),
	}
}

// Ensure InMemoryBalanceRepository implements BalanceRepository
var _ interfaces.BalanceRepository = (*InMemoryBalanceRepository)(nil)

func (r *InMemoryBalanceRepository) Upsert(balance *models.Balance) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := balanceKey{balance.DebtorID, balance.CreditorID, balance.GroupID}
	balance.UpdatedAt = time.Now()

	if _, ok := r.balances[key]; ok {
		balance.UpdatedAt = time.Now()
		r.balances[key] = copyBalance(balance)
		return nil
	}

	r.balances[key] = copyBalance(balance)
	r.addToIndex(key)
	return nil
}

// AddBalance atomically adds amount to the balance (creates if not exists)
func (r *InMemoryBalanceRepository) AddBalance(debtorID, creditorID, groupID string, amount float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := balanceKey{debtorID, creditorID, groupID}
	existing := r.balances[key]
	var newAmt float64
	if existing != nil {
		newAmt = existing.Amount + amount
	} else {
		newAmt = amount
		r.addToIndex(key)
	}
	if newAmt <= 0 {
		delete(r.balances, key)
		r.removeFromIndex(key)
		return nil
	}
	r.balances[key] = &models.Balance{
		DebtorID:   debtorID,
		CreditorID: creditorID,
		GroupID:    groupID,
		Amount:     newAmt,
		UpdatedAt:  time.Now(),
	}
	return nil
}

func (r *InMemoryBalanceRepository) addToIndex(key balanceKey) {
	// By user (debtor or creditor)
	for _, uid := range []string{key.DebtorID, key.CreditorID} {
		found := false
		for _, k := range r.byUser[uid] {
			if k == key {
				found = true
				break
			}
		}
		if !found {
			r.byUser[uid] = append(r.byUser[uid], key)
		}
	}
	// By group
	if key.GroupID != "" {
		found := false
		for _, k := range r.byGroup[key.GroupID] {
			if k == key {
				found = true
				break
			}
		}
		if !found {
			r.byGroup[key.GroupID] = append(r.byGroup[key.GroupID], key)
		}
	}
}

func (r *InMemoryBalanceRepository) removeFromIndex(key balanceKey) {
	for _, uid := range []string{key.DebtorID, key.CreditorID} {
		keys := r.byUser[uid]
		for i, k := range keys {
			if k == key {
				r.byUser[uid] = append(keys[:i], keys[i+1:]...)
				break
			}
		}
	}
	if key.GroupID != "" {
		keys := r.byGroup[key.GroupID]
		for i, k := range keys {
			if k == key {
				r.byGroup[key.GroupID] = append(keys[:i], keys[i+1:]...)
				break
			}
		}
	}
}

func (r *InMemoryBalanceRepository) Get(debtorID, creditorID, groupID string) (*models.Balance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	key := balanceKey{debtorID, creditorID, groupID}
	b, ok := r.balances[key]
	if !ok {
		return nil, fmt.Errorf("balance not found")
	}
	return copyBalance(b), nil
}

func (r *InMemoryBalanceRepository) GetAllForUser(userID string) ([]*models.Balance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	keys, ok := r.byUser[userID]
	if !ok {
		return []*models.Balance{}, nil
	}
	result := make([]*models.Balance, 0, len(keys))
	for _, key := range keys {
		if b, exists := r.balances[key]; exists && b.Amount > 0 {
			result = append(result, copyBalance(b))
		}
	}
	return result, nil
}

func (r *InMemoryBalanceRepository) GetAllForGroup(groupID string) ([]*models.Balance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	keys, ok := r.byGroup[groupID]
	if !ok {
		return []*models.Balance{}, nil
	}
	result := make([]*models.Balance, 0, len(keys))
	for _, key := range keys {
		if b, exists := r.balances[key]; exists && b.Amount > 0 {
			result = append(result, copyBalance(b))
		}
	}
	return result, nil
}

func (r *InMemoryBalanceRepository) Delete(debtorID, creditorID, groupID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := balanceKey{debtorID, creditorID, groupID}
	if _, exists := r.balances[key]; !exists {
		return fmt.Errorf("balance not found")
	}
	delete(r.balances, key)
	r.removeFromIndex(key)
	return nil
}

func copyBalance(b *models.Balance) *models.Balance {
	cpy := *b
	return &cpy
}
