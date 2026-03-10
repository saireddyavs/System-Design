package interfaces

import "ride-sharing-service/internal/models"

// RideObserver defines the observer interface for ride status updates (Observer pattern)
type RideObserver interface {
	OnRideStatusChanged(ride *models.Ride)
}
