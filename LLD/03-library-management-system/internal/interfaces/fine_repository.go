package interfaces

import "library-management-system/internal/models"

// FineRepository defines data access for fines
type FineRepository interface {
	Create(fine *models.Fine) error
	GetByID(id string) (*models.Fine, error)
	Update(fine *models.Fine) error
	GetByLoanID(loanID string) (*models.Fine, error)
	GetPendingByMemberID(memberID string) ([]*models.Fine, error)
	ListAll() ([]*models.Fine, error)
}
