package repositories

import (
	"errors"
	"food-delivery-system/internal/models"
	"sync"
)

var ErrCustomerNotFound = errors.New("customer not found")

// InMemoryCustomerRepo implements CustomerRepository with thread-safe in-memory storage
type InMemoryCustomerRepo struct {
	customers map[string]*models.Customer
	byEmail  map[string]string
	mu       sync.RWMutex
}

// NewInMemoryCustomerRepo creates a new in-memory customer repository
func NewInMemoryCustomerRepo() *InMemoryCustomerRepo {
	return &InMemoryCustomerRepo{
		customers: make(map[string]*models.Customer),
		byEmail:   make(map[string]string),
	}
}

func (r *InMemoryCustomerRepo) Create(customer *models.Customer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.customers[customer.ID]; exists {
		return errors.New("customer already exists")
	}
	r.customers[customer.ID] = customer
	r.byEmail[customer.Email] = customer.ID
	return nil
}

func (r *InMemoryCustomerRepo) GetByID(id string) (*models.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	customer, exists := r.customers[id]
	if !exists {
		return nil, ErrCustomerNotFound
	}
	return customer, nil
}

func (r *InMemoryCustomerRepo) GetByEmail(email string) (*models.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, exists := r.byEmail[email]
	if !exists {
		return nil, ErrCustomerNotFound
	}
	return r.customers[id], nil
}

func (r *InMemoryCustomerRepo) Update(customer *models.Customer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.customers[customer.ID]; !exists {
		return ErrCustomerNotFound
	}
	r.customers[customer.ID] = customer
	r.byEmail[customer.Email] = customer.ID
	return nil
}
