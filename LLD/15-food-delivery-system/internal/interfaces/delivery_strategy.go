package interfaces

import (
	"food-delivery-system/internal/models"
)

// DeliveryStrategy defines the contract for delivery agent assignment (Strategy Pattern)
// Different strategies: nearest agent, highest-rated, load-balanced
type DeliveryStrategy interface {
	AssignAgent(restaurantLoc, deliveryAddr models.Location, agents []*models.DeliveryAgent, maxRadiusKm float64) (*models.DeliveryAgent, error)
}
