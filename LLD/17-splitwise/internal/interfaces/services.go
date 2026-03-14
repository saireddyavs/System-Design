package interfaces

import (
	"splitwise/internal/models"
)

// ExpenseObserver is notified when expenses are added (Observer Pattern)
type ExpenseObserver interface {
	OnExpenseAdded(expense *models.Expense)
}

// SettlementObserver is notified when settlements are recorded (Observer Pattern)
type SettlementObserver interface {
	OnSettlement(transaction *models.Transaction)
}

// UserService defines user operations
type UserService interface {
	CreateUser(name, email, phone string) (*models.User, error)
	GetUser(id string) (*models.User, error)
}

// GroupService defines group operations
type GroupService interface {
	CreateGroup(name, description, createdBy string, memberIDs []string) (*models.Group, error)
	GetGroup(id string) (*models.Group, error)
	AddMember(groupID, userID string) error
	RemoveMember(groupID, userID string) error
}

// ExpenseService defines expense operations
type ExpenseService interface {
	AddExpense(description string, amount float64, paidBy string, splitType models.SplitType, participantIDs []string, splitParams map[string]float64, groupID string) (*models.Expense, error)
	GetExpensesByGroup(groupID string) ([]*models.Expense, error)
	GetExpensesByUser(userID string) ([]*models.Expense, error)
	GetExpensesBetweenUsers(userID1, userID2 string) ([]*models.Expense, error)
	RegisterObserver(observer ExpenseObserver)
}

// BalanceService defines balance and settlement operations
type BalanceService interface {
	GetBalancesForUser(userID string) ([]*models.Balance, error)
	GetBalancesForGroup(groupID string) ([]*models.Balance, error)
	SimplifyDebts(groupID string) ([]*models.Transaction, error)
	Settle(fromUserID, toUserID string, amount float64, groupID string) (*models.Transaction, error)
}
