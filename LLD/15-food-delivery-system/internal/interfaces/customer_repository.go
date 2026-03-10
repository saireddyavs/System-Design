package interfaces

import "food-delivery-system/internal/models"

// CustomerRepository defines the contract for customer data access
type CustomerRepository interface {
	Create(customer *models.Customer) error
	GetByID(id string) (*models.Customer, error)
	GetByEmail(email string) (*models.Customer, error)
	Update(customer *models.Customer) error
}
