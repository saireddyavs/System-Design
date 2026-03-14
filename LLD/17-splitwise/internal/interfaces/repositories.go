package interfaces

import (
	"splitwise/internal/models"
)

// UserRepository defines user data access
type UserRepository interface {
	Create(user *models.User) error
	GetByID(id string) (*models.User, error)
	GetByIDs(ids []string) ([]*models.User, error)
	Update(user *models.User) error
	Delete(id string) error
}

// GroupRepository defines group data access
type GroupRepository interface {
	Create(group *models.Group) error
	GetByID(id string) (*models.Group, error)
	GetByUserID(userID string) ([]*models.Group, error)
	Update(group *models.Group) error
	Delete(id string) error
}

// ExpenseRepository defines expense data access
type ExpenseRepository interface {
	Create(expense *models.Expense) error
	GetByID(id string) (*models.Expense, error)
	GetByGroupID(groupID string) ([]*models.Expense, error)
	GetByUserID(userID string) ([]*models.Expense, error)
	GetBetweenUsers(userID1, userID2 string) ([]*models.Expense, error)
	Update(expense *models.Expense) error
	Delete(id string) error
}

// BalanceRepository defines balance data access
type BalanceRepository interface {
	Upsert(balance *models.Balance) error
	AddBalance(debtorID, creditorID, groupID string, amount float64) error
	Get(debtorID, creditorID, groupID string) (*models.Balance, error)
	GetAllForUser(userID string) ([]*models.Balance, error)
	GetAllForGroup(groupID string) ([]*models.Balance, error)
	Delete(debtorID, creditorID, groupID string) error
}

// TransactionRepository defines settlement transaction data access
type TransactionRepository interface {
	Create(tx *models.Transaction) error
	GetByID(id string) (*models.Transaction, error)
	GetByUserID(userID string) ([]*models.Transaction, error)
	GetByGroupID(groupID string) ([]*models.Transaction, error)
}
