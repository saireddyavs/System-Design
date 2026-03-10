package interfaces

import "ride-sharing-service/internal/models"

// RatingRepository defines data access operations for ratings
type RatingRepository interface {
	Create(rating *models.Rating) error
	GetRatingsForUser(userID string) ([]*models.Rating, error)
	GetAverageRating(userID string) (float64, int, error)
}
