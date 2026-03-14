package services

import (
	"errors"
	"fmt"
	"sync"

	"splitwise/internal/interfaces"
	"splitwise/internal/models"
	"splitwise/internal/strategies"
)

// ExpenseService implements expense management
type ExpenseService struct {
	expenseRepo  interfaces.ExpenseRepository
	balanceRepo  interfaces.BalanceRepository
	groupRepo    interfaces.GroupRepository
	userRepo     interfaces.UserRepository
	registry     *strategies.SplitStrategyRegistry
	observers    []interfaces.ExpenseObserver
	mu           sync.RWMutex
}

// NewExpenseService creates a new ExpenseService
func NewExpenseService(
	expenseRepo interfaces.ExpenseRepository,
	balanceRepo interfaces.BalanceRepository,
	groupRepo interfaces.GroupRepository,
	userRepo interfaces.UserRepository,
	registry *strategies.SplitStrategyRegistry,
) *ExpenseService {
	return &ExpenseService{
		expenseRepo: expenseRepo,
		balanceRepo: balanceRepo,
		groupRepo:   groupRepo,
		userRepo:    userRepo,
		registry:    registry,
		observers:   make([]interfaces.ExpenseObserver, 0),
	}
}

// Ensure ExpenseService implements interfaces.ExpenseService
var _ interfaces.ExpenseService = (*ExpenseService)(nil)

// RegisterObserver adds an observer for expense events (Observer Pattern)
func (s *ExpenseService) RegisterObserver(observer interfaces.ExpenseObserver) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, observer)
}

func (s *ExpenseService) notifyExpenseAdded(expense *models.Expense) {
	s.mu.RLock()
	obs := make([]interfaces.ExpenseObserver, len(s.observers))
	copy(obs, s.observers)
	s.mu.RUnlock()
	for _, o := range obs {
		o.OnExpenseAdded(expense)
	}
}

// AddExpense adds a new expense
// participantIDs: users involved in the split (including payer for EQUAL)
// splitParams: for EXACT map[userID]=amount, for PERCENTAGE map[userID]=pct, for SHARE map[userID]=share
func (s *ExpenseService) AddExpense(
	description string,
	amount float64,
	paidBy string,
	splitType models.SplitType,
	participantIDs []string,
	splitParams map[string]float64,
	groupID string,
) (*models.Expense, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	if paidBy == "" {
		return nil, errors.New("paidBy is required")
	}
	if len(participantIDs) == 0 {
		return nil, errors.New("at least one participant required")
	}

	// Validate group if provided
	if groupID != "" {
		group, err := s.groupRepo.GetByID(groupID)
		if err != nil {
			return nil, err
		}
		memberSet := make(map[string]bool)
		for _, mid := range group.MemberIDs {
			memberSet[mid] = true
		}
		for _, pid := range participantIDs {
			if !memberSet[pid] {
				return nil, fmt.Errorf("participant %s not in group", pid)
			}
		}
		if !memberSet[paidBy] {
			return nil, fmt.Errorf("payer %s not in group", paidBy)
		}
	}

	// Get strategy and calculate splits
	strategy := s.registry.GetStrategy(splitType)
	if strategy == nil {
		return nil, fmt.Errorf("unsupported split type: %s", splitType)
	}

	splits, err := strategy.CalculateSplits(amount, paidBy, participantIDs, splitParams)
	if err != nil {
		return nil, err
	}

	// Build expense (Builder Pattern)
	expense := models.NewExpenseBuilder().
		WithDescription(description).
		WithAmount(amount).
		WithPaidBy(paidBy).
		WithSplitType(splitType).
		WithSplits(splits).
		WithGroupID(groupID).
		Build()

	if err := s.expenseRepo.Create(expense); err != nil {
		return nil, err
	}

	// Update balances: each participant (except payer) owes the payer
	for _, split := range splits {
		if split.UserID != paidBy && split.Amount > 0 {
			if err := s.balanceRepo.AddBalance(split.UserID, paidBy, groupID, split.Amount); err != nil {
				return nil, err
			}
		}
	}

	s.notifyExpenseAdded(expense)
	return expense, nil
}

// GetExpensesByGroup returns all expenses for a group
func (s *ExpenseService) GetExpensesByGroup(groupID string) ([]*models.Expense, error) {
	return s.expenseRepo.GetByGroupID(groupID)
}

// GetExpensesByUser returns all expenses involving a user
func (s *ExpenseService) GetExpensesByUser(userID string) ([]*models.Expense, error) {
	return s.expenseRepo.GetByUserID(userID)
}

// GetExpensesBetweenUsers returns expenses between two users
func (s *ExpenseService) GetExpensesBetweenUsers(userID1, userID2 string) ([]*models.Expense, error) {
	return s.expenseRepo.GetBetweenUsers(userID1, userID2)
}
