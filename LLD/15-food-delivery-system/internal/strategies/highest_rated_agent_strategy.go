package strategies

import (
	"food-delivery-system/internal/interfaces"
	"food-delivery-system/internal/models"
)

// HighestRatedAgentStrategy assigns the highest-rated available agent within radius
type HighestRatedAgentStrategy struct{}

// NewHighestRatedAgentStrategy creates a new highest-rated agent strategy
func NewHighestRatedAgentStrategy() interfaces.DeliveryStrategy {
	return &HighestRatedAgentStrategy{}
}

// AssignAgent finds the highest-rated available agent within maxRadiusKm
func (s *HighestRatedAgentStrategy) AssignAgent(restaurantLoc, deliveryAddr models.Location, agents []*models.DeliveryAgent, maxRadiusKm float64) (*models.DeliveryAgent, error) {
	if len(agents) == 0 {
		return nil, ErrNoAvailableAgents
	}

	var best *models.DeliveryAgent
	bestRating := -1.0

	for _, agent := range agents {
		if agent.Status != models.AgentStatusAvailable {
			continue
		}
		dist := restaurantLoc.Distance(agent.Location)
		if dist <= maxRadiusKm && agent.Rating > bestRating {
			bestRating = agent.Rating
			best = agent
		}
	}

	if best == nil {
		return nil, ErrNoAvailableAgents
	}
	return best, nil
}
