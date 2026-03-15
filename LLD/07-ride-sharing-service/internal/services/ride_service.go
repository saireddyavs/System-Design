package services

import (
	"errors"
	"fmt"
	"ride-sharing-service/internal/interfaces"
	"ride-sharing-service/internal/models"
	"time"
)

const (
	SurgeThreshold      = 10  // Active requests to trigger surge
	SurgeMultiplier     = 1.5 // 50% surge when threshold exceeded
	CancellationPenalty = 0.5 // 50% of base fare if cancelled in progress
)

var (
	ErrRiderHasActiveRide   = errors.New("rider already has an active ride")
	ErrDriverHasActiveRide  = errors.New("driver already has an active ride")
	ErrInvalidRideState    = errors.New("invalid ride state for this operation")
	ErrRideNotFound        = errors.New("ride not found")
)

// RideService handles ride lifecycle and orchestration
type RideService struct {
	rideRepo         interfaces.RideRepository
	driverRepo       interfaces.DriverRepository
	riderRepo        interfaces.RiderRepository
	matchingService  *MatchingService
	fareCalculator   interfaces.FareCalculator
	paymentProcessor interfaces.PaymentProcessor
	ratingRepo       interfaces.RatingRepository
	notifier         *RideNotifier
}

// NewRideService creates a new ride service
func NewRideService(
	rideRepo interfaces.RideRepository,
	driverRepo interfaces.DriverRepository,
	riderRepo interfaces.RiderRepository,
	matchingService *MatchingService,
	fareCalculator interfaces.FareCalculator,
	paymentProcessor interfaces.PaymentProcessor,
	ratingRepo interfaces.RatingRepository,
	notifier *RideNotifier,
) *RideService {
	return &RideService{
		rideRepo:         rideRepo,
		driverRepo:       driverRepo,
		riderRepo:        riderRepo,
		matchingService:  matchingService,
		fareCalculator:   fareCalculator,
		paymentProcessor: paymentProcessor,
		ratingRepo:       ratingRepo,
		notifier:         notifier,
	}
}

// RequestRide creates a ride request and matches with nearest driver
func (s *RideService) RequestRide(riderID string, pickup, dropoff models.Location) (*models.Ride, error) {
	// Check rider doesn't have active ride
	activeRides, _ := s.rideRepo.GetActiveRidesByRider(riderID)
	if len(activeRides) > 0 {
		return nil, ErrRiderHasActiveRide
	}

	// Update rider location to pickup
	rider, err := s.riderRepo.GetByID(riderID)
	if err != nil {
		return nil, err
	}
	rider.UpdateLocation(pickup)
	_ = s.riderRepo.Update(rider)

	// Find driver using matching strategy
	driver, err := s.matchingService.FindNearestDriver(pickup)
	if err != nil {
		return nil, err
	}

	// Create ride
	rideID := generateID("ride")
	ride := models.NewRide(rideID, riderID, pickup, dropoff)
	ride.AssignDriver(driver.ID)

	if err := s.rideRepo.Create(ride); err != nil {
		return nil, err
	}

	// Update driver status
	driver.SetStatus(models.DriverStatusOnRide)
	_ = s.driverRepo.Update(driver)

	s.notifier.Notify(ride)
	return ride, nil
}

// StartRide begins the ride
func (s *RideService) StartRide(rideID string) error {
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return err
	}
	if ride.GetStatus() != models.RideStatusDriverAssigned {
		return ErrInvalidRideState
	}
	ride.Start()
	_ = s.rideRepo.Update(ride)
	s.notifier.Notify(ride)
	return nil
}

// CompleteRide ends the ride, calculates fare, processes payment
func (s *RideService) CompleteRide(rideID string) (*models.Ride, error) {
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return nil, err
	}
	if ride.GetStatus() != models.RideStatusInProgress {
		return nil, ErrInvalidRideState
	}

	// Calculate duration and distance
	startTime := ride.StartTime
	if startTime == nil {
		startTime = &ride.RequestedAt
	}
	duration := time.Since(*startTime)
	distance := models.HaversineDistance(ride.Pickup, ride.Dropoff)

	// Get surge multiplier
	surgeMultiplier := s.getSurgeMultiplier()

	// Calculate fare
	fare := s.fareCalculator.Calculate(ride.Pickup, ride.Dropoff, duration, surgeMultiplier)

	// Complete ride
	ride.Complete(fare, distance, duration)
	_ = s.rideRepo.Update(ride)

	// Process payment
	_, _ = s.paymentProcessor.ProcessPayment(rideID, fare, models.PaymentMethodCard)

	// Free driver
	driver, _ := s.driverRepo.GetByID(ride.DriverID)
	if driver != nil {
		driver.SetStatus(models.DriverStatusAvailable)
		_ = s.driverRepo.Update(driver)
	}

	s.notifier.Notify(ride)
	return ride, nil
}

// CancelRide cancels a ride (with penalty if in progress)
func (s *RideService) CancelRide(rideID string) (penalty float64, err error) {
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return 0, err
	}
	status := ride.GetStatus()
	if status == models.RideStatusCompleted || status == models.RideStatusCancelled {
		return 0, ErrInvalidRideState
	}

	penalty = 0
	if ride.GetStatus() == models.RideStatusInProgress {
		// Cancellation penalty only when ride has actually started (driver picked up rider)
		baseFare := s.fareCalculator.Calculate(ride.Pickup, ride.Dropoff, 0, 1.0)
		penalty = baseFare * CancellationPenalty
		// Process penalty payment
		if penalty > 0 {
			_, _ = s.paymentProcessor.ProcessPayment(rideID, penalty, models.PaymentMethodCard)
		}
	}

	ride.Cancel()
	_ = s.rideRepo.Update(ride)

	// Free driver if was assigned
	if ride.DriverID != "" {
		driver, _ := s.driverRepo.GetByID(ride.DriverID)
		if driver != nil {
			driver.SetStatus(models.DriverStatusAvailable)
			_ = s.driverRepo.Update(driver)
		}
	}

	s.notifier.Notify(ride)
	return penalty, nil
}

// AddRating adds a rating for driver or rider after ride completion
func (s *RideService) AddRating(rideID, fromUserID, toUserID string, score float64, comment string) error {
	ride, err := s.rideRepo.GetByID(rideID)
	if err != nil {
		return err
	}
	if ride.GetStatus() != models.RideStatusCompleted {
		return ErrInvalidRideState
	}
	if score < 1 || score > 5 {
		return errors.New("rating must be between 1 and 5")
	}

	rating := &models.Rating{
		ID:         generateID("rating"),
		RideID:     rideID,
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Score:      score,
		Comment:    comment,
		CreatedAt:  time.Now(),
	}
	if err := s.ratingRepo.Create(rating); err != nil {
		return err
	}

	// Update user's average rating (totalRatings is count after adding this rating)
	_, totalRatings, _ := s.ratingRepo.GetAverageRating(toUserID)
	totalRatings++

	// Check if toUserID is driver - update rating and deactivate if below 3.0
	driver, _ := s.driverRepo.GetByID(toUserID)
	if driver != nil {
		driver.UpdateRating(score, totalRatings)
		if driver.GetRating() < 3.0 {
			driver.SetStatus(models.DriverStatusDeactivated)
		}
		_ = s.driverRepo.Update(driver)
	} else {
		rider, _ := s.riderRepo.GetByID(toUserID)
		if rider != nil {
			rider.UpdateRating(score, totalRatings)
			_ = s.riderRepo.Update(rider)
		}
	}

	return nil
}

// GetRide retrieves ride by ID
func (s *RideService) GetRide(rideID string) (*models.Ride, error) {
	return s.rideRepo.GetByID(rideID)
}

func (s *RideService) getSurgeMultiplier() float64 {
	count, err := s.rideRepo.CountActiveRequests()
	if err != nil || count < SurgeThreshold {
		return 1.0
	}
	return SurgeMultiplier
}

func generateID(prefix string) string {
	return prefix + "-" + fmt.Sprintf("%d", time.Now().UnixNano())
}
