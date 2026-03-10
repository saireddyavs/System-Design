package services

import (
	"errors"
	"food-delivery-system/internal/interfaces"
	"food-delivery-system/internal/models"
)

var ErrCustomerNotFound = errors.New("customer not found")

// CustomerService handles customer business logic
type CustomerService struct {
	repo interfaces.CustomerRepository
}

// NewCustomerService creates a new customer service
func NewCustomerService(repo interfaces.CustomerRepository) *CustomerService {
	return &CustomerService{repo: repo}
}

// RegisterCustomer creates a new customer
func (s *CustomerService) RegisterCustomer(id, name, email, phone string, loc models.Location) (*models.Customer, error) {
	customer := models.NewCustomer(id, name, email, phone, loc)
	if err := s.repo.Create(customer); err != nil {
		return nil, err
	}
	return customer, nil
}

// GetCustomer retrieves a customer by ID
func (s *CustomerService) GetCustomer(id string) (*models.Customer, error) {
	return s.repo.GetByID(id)
}

// AddAddress adds a delivery address to a customer
func (s *CustomerService) AddAddress(customerID string, addr models.Address) error {
	customer, err := s.repo.GetByID(customerID)
	if err != nil {
		return err
	}
	customer.AddAddress(addr)
	return s.repo.Update(customer)
}
