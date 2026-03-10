package services

import (
	"errors"
	"food-delivery-system/internal/interfaces"
	"food-delivery-system/internal/models"
)

var ErrNoAgentAvailable = errors.New("no delivery agent available within radius")

// DeliveryService handles delivery agent assignment
type DeliveryService struct {
	agentRepo  interfaces.AgentRepository
	strategy   interfaces.DeliveryStrategy
	maxRadiusKm float64
}

// NewDeliveryService creates a new delivery service
func NewDeliveryService(agentRepo interfaces.AgentRepository, strategy interfaces.DeliveryStrategy, maxRadiusKm float64) *DeliveryService {
	if maxRadiusKm <= 0 {
		maxRadiusKm = MaxDeliveryRadiusKm
	}
	return &DeliveryService{
		agentRepo:   agentRepo,
		strategy:   strategy,
		maxRadiusKm: maxRadiusKm,
	}
}

// AssignAgent assigns the best available agent using the configured strategy
func (s *DeliveryService) AssignAgent(restaurantLoc, deliveryAddr models.Location) (*models.DeliveryAgent, error) {
	agents, err := s.agentRepo.GetAvailableAgents()
	if err != nil {
		return nil, err
	}

	agent, err := s.strategy.AssignAgent(restaurantLoc, deliveryAddr, agents, s.maxRadiusKm)
	if err != nil {
		return nil, ErrNoAgentAvailable
	}

	agent.SetStatus(models.AgentStatusOnDelivery)
	if err := s.agentRepo.Update(agent); err != nil {
		agent.SetStatus(models.AgentStatusAvailable)
		return nil, err
	}

	return agent, nil
}

// MarkAgentAvailable marks an agent as available (e.g., after delivery)
func (s *DeliveryService) MarkAgentAvailable(agentID string) error {
	agent, err := s.agentRepo.GetByID(agentID)
	if err != nil {
		return err
	}
	agent.SetStatus(models.AgentStatusAvailable)
	return s.agentRepo.Update(agent)
}

// RegisterAgent adds a new delivery agent
func (s *DeliveryService) RegisterAgent(id, name, phone string, loc models.Location) (*models.DeliveryAgent, error) {
	agent := models.NewDeliveryAgent(id, name, phone, loc)
	if err := s.agentRepo.Create(agent); err != nil {
		return nil, err
	}
	return agent, nil
}

// UpdateAgentLocation updates agent's current location
func (s *DeliveryService) UpdateAgentLocation(agentID string, loc models.Location) error {
	return s.agentRepo.UpdateLocation(agentID, loc)
}
