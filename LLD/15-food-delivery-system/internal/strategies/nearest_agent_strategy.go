package strategies

import (
	"errors"
	"food-delivery-system/internal/interfaces"
	"food-delivery-system/internal/models"
	"math"
)

var ErrNoAvailableAgents = errors.New("no available agents within radius")

// NearestAgentStrategy assigns the delivery agent closest to the restaurant (Strategy Pattern)
type NearestAgentStrategy struct{}

// NewNearestAgentStrategy creates a new nearest agent strategy
func NewNearestAgentStrategy() interfaces.DeliveryStrategy {
	return &NearestAgentStrategy{}
}

// AssignAgent finds the nearest available agent within maxRadiusKm
func (s *NearestAgentStrategy) AssignAgent(restaurantLoc, deliveryAddr models.Location, agents []*models.DeliveryAgent, maxRadiusKm float64) (*models.DeliveryAgent, error) {
	if len(agents) == 0 {
		return nil, ErrNoAvailableAgents
	}

	var nearest *models.DeliveryAgent
	minDist := math.MaxFloat64

	for _, agent := range agents {
		if agent.Status != models.AgentStatusAvailable {
			continue
		}
		dist := restaurantLoc.Distance(agent.Location)
		if dist <= maxRadiusKm && dist < minDist {
			minDist = dist
			nearest = agent
		}
	}

	if nearest == nil {
		return nil, ErrNoAvailableAgents
	}
	return nearest, nil
}
