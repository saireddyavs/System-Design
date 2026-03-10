package interfaces

import (
	"library-management-system/internal/models"
	"time"
)

// LoanRepository defines data access for loans (Repository pattern)
type LoanRepository interface {
	Create(loan *models.Loan) error
	GetByID(id string) (*models.Loan, error)
	Update(loan *models.Loan) error
	GetActiveByBookID(bookID string) (*models.Loan, error)
	GetActiveByMemberID(memberID string) ([]*models.Loan, error)
	GetOverdueLoans() ([]*models.Loan, error)
	GetLoansDueBefore(date time.Time) ([]*models.Loan, error)
}
