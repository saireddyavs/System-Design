package services

import (
	"fmt"
	"food-delivery-system/internal/interfaces"
	"food-delivery-system/internal/models"
	"time"
)

// RatingService handles ratings for restaurants and delivery agents
type RatingService struct {
	ratingRepo     interfaces.RatingRepository
	restaurantRepo interfaces.RestaurantRepository
	agentRepo      interfaces.AgentRepository
}

// NewRatingService creates a new rating service
func NewRatingService(
	ratingRepo interfaces.RatingRepository,
	restaurantRepo interfaces.RestaurantRepository,
	agentRepo interfaces.AgentRepository,
) *RatingService {
	return &RatingService{
		ratingRepo:     ratingRepo,
		restaurantRepo: restaurantRepo,
		agentRepo:      agentRepo,
	}
}

// RateRestaurant adds a restaurant rating
func (s *RatingService) RateRestaurant(orderID, customerID, restaurantID string, score float64, comment string) error {
	rating := models.NewRating(fmt.Sprintf("RAT-R-%d", time.Now().UnixNano()), orderID, customerID, models.RatingTypeRestaurant, restaurantID, score, comment)
	if err := s.ratingRepo.Create(rating); err != nil {
		return err
	}
	avg, _ := s.ratingRepo.GetAverageRating(restaurantID, models.RatingTypeRestaurant)
	restaurant, _ := s.restaurantRepo.GetByID(restaurantID)
	if restaurant != nil {
		restaurant.UpdateRating(avg)
		s.restaurantRepo.Update(restaurant)
	}
	return nil
}

// RateAgent adds a delivery agent rating
func (s *RatingService) RateAgent(orderID, customerID, agentID string, score float64, comment string) error {
	rating := models.NewRating(fmt.Sprintf("RAT-A-%d", time.Now().UnixNano()), orderID, customerID, models.RatingTypeAgent, agentID, score, comment)
	if err := s.ratingRepo.Create(rating); err != nil {
		return err
	}
	avg, _ := s.ratingRepo.GetAverageRating(agentID, models.RatingTypeAgent)
	agent, _ := s.agentRepo.GetByID(agentID)
	if agent != nil {
		agent.UpdateRating(avg)
		s.agentRepo.Update(agent)
	}
	return nil
}
