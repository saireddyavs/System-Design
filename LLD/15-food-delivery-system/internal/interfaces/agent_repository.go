package interfaces

import (
	"food-delivery-system/internal/models"
)

// AgentRepository defines the contract for delivery agent data access (Repository Pattern)
type AgentRepository interface {
	Create(agent *models.DeliveryAgent) error
	GetByID(id string) (*models.DeliveryAgent, error)
	GetAll() ([]*models.DeliveryAgent, error)
	GetAvailableAgents() ([]*models.DeliveryAgent, error)
	Update(agent *models.DeliveryAgent) error
}
