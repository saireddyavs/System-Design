package interfaces

import "food-delivery-system/internal/models"

// RatingRepository defines the contract for rating data access
type RatingRepository interface {
	Create(rating *models.Rating) error
	GetByTargetID(targetID string, ratingType models.RatingType) ([]*models.Rating, error)
	GetAverageRating(targetID string, ratingType models.RatingType) (float64, error)
}
