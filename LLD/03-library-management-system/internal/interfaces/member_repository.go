package interfaces

import "library-management-system/internal/models"

// MemberRepository defines data access for members (Repository pattern)
type MemberRepository interface {
	Create(member *models.Member) error
	GetByID(id string) (*models.Member, error)
	Update(member *models.Member) error
	Delete(id string) error
	ListAll() ([]*models.Member, error)
	GetByEmail(email string) (*models.Member, error)
}
