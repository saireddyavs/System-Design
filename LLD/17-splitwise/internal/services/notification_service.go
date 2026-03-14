package services

import (
	"fmt"
	"sync"

	"splitwise/internal/interfaces"
	"splitwise/internal/models"
)

// NotificationService implements ExpenseObserver for notifications
type NotificationService struct {
	mu           sync.RWMutex
	expenseLog   []string
	settlementLog []string
}

// NewNotificationService creates a new NotificationService
func NewNotificationService() *NotificationService {
	return &NotificationService{
		expenseLog:   make([]string, 0),
		settlementLog: make([]string, 0),
	}
}

// Ensure NotificationService implements both observer interfaces
var _ interfaces.ExpenseObserver = (*NotificationService)(nil)
var _ interfaces.SettlementObserver = (*NotificationService)(nil)

// OnExpenseAdded is called when a new expense is added
func (n *NotificationService) OnExpenseAdded(expense *models.Expense) {
	n.mu.Lock()
	defer n.mu.Unlock()
	msg := fmt.Sprintf("New expense: %s - $%.2f paid by %s", expense.Description, expense.Amount, expense.PaidBy)
	n.expenseLog = append(n.expenseLog, msg)
}

// OnSettlement is called when a settlement is recorded
func (n *NotificationService) OnSettlement(transaction *models.Transaction) {
	n.mu.Lock()
	defer n.mu.Unlock()
	msg := fmt.Sprintf("Settlement: %s paid $%.2f to %s", transaction.FromUserID, transaction.Amount, transaction.ToUserID)
	n.settlementLog = append(n.settlementLog, msg)
}

// GetExpenseLog returns the expense notification log
func (n *NotificationService) GetExpenseLog() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	result := make([]string, len(n.expenseLog))
	copy(result, n.expenseLog)
	return result
}

// GetSettlementLog returns the settlement notification log
func (n *NotificationService) GetSettlementLog() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	result := make([]string, len(n.settlementLog))
	copy(result, n.settlementLog)
	return result
}
