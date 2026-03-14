package services

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"splitwise/internal/interfaces"
	"splitwise/internal/models"
)

// BalanceService implements balance tracking and settlement
type BalanceService struct {
	balanceRepo     interfaces.BalanceRepository
	transactionRepo interfaces.TransactionRepository
	observers       []interfaces.SettlementObserver
	mu              sync.RWMutex
	txSeq           int
}

// NewBalanceService creates a new BalanceService
func NewBalanceService(
	balanceRepo interfaces.BalanceRepository,
	transactionRepo interfaces.TransactionRepository,
) *BalanceService {
	return &BalanceService{
		balanceRepo:     balanceRepo,
		transactionRepo: transactionRepo,
		observers:       make([]interfaces.SettlementObserver, 0),
		txSeq:           1,
	}
}

// RegisterObserver adds a settlement observer
func (s *BalanceService) RegisterObserver(observer interfaces.SettlementObserver) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, observer)
}

func (s *BalanceService) notifySettlement(tx *models.Transaction) {
	s.mu.RLock()
	obs := make([]interfaces.SettlementObserver, len(s.observers))
	copy(obs, s.observers)
	s.mu.RUnlock()
	for _, o := range obs {
		o.OnSettlement(tx)
	}
}

// Ensure BalanceService implements interfaces.BalanceService
var _ interfaces.BalanceService = (*BalanceService)(nil)

// GetBalancesForUser returns all balances for a user (both as debtor and creditor)
func (s *BalanceService) GetBalancesForUser(userID string) ([]*models.Balance, error) {
	return s.balanceRepo.GetAllForUser(userID)
}

// GetBalancesForGroup returns all balances within a group
func (s *BalanceService) GetBalancesForGroup(groupID string) ([]*models.Balance, error) {
	return s.balanceRepo.GetAllForGroup(groupID)
}

// SimplifyDebts minimizes transactions using net-balance greedy algorithm
// Returns suggested settlement transactions (does not execute them)
func (s *BalanceService) SimplifyDebts(groupID string) ([]*models.Transaction, error) {
	balances, err := s.balanceRepo.GetAllForGroup(groupID)
	if err != nil {
		return nil, err
	}

	// Calculate net balance per user
	// Positive = owed money (creditor), Negative = owes money (debtor)
	netBalance := make(map[string]float64)
	for _, b := range balances {
		netBalance[b.DebtorID] -= b.Amount
		netBalance[b.CreditorID] += b.Amount
	}

	// Separate creditors (positive) and debtors (negative)
	type balanceEntry struct {
		userID string
		amount float64
	}
	var creditors []balanceEntry
	var debtors []balanceEntry
	for userID, amt := range netBalance {
		if math.Abs(amt) < 0.01 {
			continue
		}
		if amt > 0 {
			creditors = append(creditors, balanceEntry{userID, amt})
		} else {
			debtors = append(debtors, balanceEntry{userID, -amt})
		}
	}

	// Sort: creditors descending, debtors descending
	sort.Slice(creditors, func(i, j int) bool { return creditors[i].amount > creditors[j].amount })
	sort.Slice(debtors, func(i, j int) bool { return debtors[i].amount > debtors[j].amount })

	// Greedy matching: match largest creditor with largest debtor
	suggested := make([]*models.Transaction, 0)
	i, j := 0, 0
	for i < len(creditors) && j < len(debtors) {
		credAmt := creditors[i].amount
		debtAmt := debtors[j].amount
		amount := math.Min(credAmt, debtAmt)
		if amount < 0.01 {
			break
		}
		suggested = append(suggested, &models.Transaction{
			FromUserID:  debtors[j].userID,
			ToUserID:    creditors[i].userID,
			Amount:      math.Round(amount*100) / 100,
			GroupID:     groupID,
			Description: "Simplified settlement",
		})
		creditors[i].amount -= amount
		debtors[j].amount -= amount
		if creditors[i].amount < 0.01 {
			i++
		}
		if debtors[j].amount < 0.01 {
			j++
		}
	}

	return suggested, nil
}

// Settle records a payment from fromUserID to toUserID
func (s *BalanceService) Settle(fromUserID, toUserID string, amount float64, groupID string) (*models.Transaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}
	if fromUserID == toUserID {
		return nil, fmt.Errorf("cannot settle with self")
	}

	// Get current balance - fromUserID (debtor) owes toUserID (creditor)
	balance, err := s.balanceRepo.Get(fromUserID, toUserID, groupID)
	if err != nil || balance == nil || balance.Amount < amount {
		return nil, fmt.Errorf("insufficient balance: %s does not owe %s enough", fromUserID, toUserID)
	}

	// Reduce balance
	if err := s.balanceRepo.AddBalance(fromUserID, toUserID, groupID, -amount); err != nil {
		return nil, err
	}

	s.mu.Lock()
	txID := fmt.Sprintf("tx%d", s.txSeq)
	s.txSeq++
	s.mu.Unlock()

	tx := &models.Transaction{
		ID:          txID,
		FromUserID:  fromUserID,
		ToUserID:    toUserID,
		Amount:      amount,
		GroupID:     groupID,
		Description: "Settlement",
		CreatedAt:   time.Now(),
	}
	if err := s.transactionRepo.Create(tx); err != nil {
		return nil, err
	}

	s.notifySettlement(tx)
	return tx, nil
}
