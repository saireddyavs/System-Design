package main

import (
	"fmt"
	"ride-sharing-service/internal/models"
	"ride-sharing-service/internal/repositories"
	"ride-sharing-service/internal/services"
	"ride-sharing-service/internal/strategies"
	"time"
)

func main() {
	// Initialize repositories
	driverRepo := repositories.NewInMemoryDriverRepository()
	riderRepo := repositories.NewInMemoryRiderRepository()
	rideRepo := repositories.NewInMemoryRideRepository()
	paymentRepo := repositories.NewInMemoryPaymentRepository()
	ratingRepo := repositories.NewInMemoryRatingRepository()

	// Initialize strategies
	matchingStrategy := strategies.NewNearestDriverStrategy(50)
	fareCalculator := strategies.NewBaseFareStrategy(2.0, 1.5, 0.25) // $2 base, $1.5/km, $0.25/min

	// Initialize services
	driverService := services.NewDriverService(driverRepo)
	riderService := services.NewRiderService(riderRepo)
	paymentProcessor := services.NewInMemoryPaymentProcessor(paymentRepo)
	notifier := services.NewRideNotifier()

	matchingService := services.NewMatchingService(driverRepo, matchingStrategy, 50)
	rideService := services.NewRideService(
		rideRepo, driverRepo, riderRepo,
		matchingService, fareCalculator, paymentProcessor, ratingRepo, notifier,
	)

	// Demo: Register driver and rider
	_, _ = driverService.RegisterDriver("D1", "John Driver", "+1234567890", models.Vehicle{
		Model: "Toyota Camry", Number: "ABC-123", Type: "sedan",
	})
	driverService.UpdateLocation("D1", models.Location{Latitude: 37.7749, Longitude: -122.4194})
	driverService.GoOnline("D1")

	_, _ = riderService.RegisterRider("R1", "Jane Rider", "+0987654321", "jane@example.com")
	riderService.UpdateLocation("R1", models.Location{Latitude: 37.7750, Longitude: -122.4195})

	// Demo: Request ride
	pickup := models.Location{Latitude: 37.7750, Longitude: -122.4195}
	dropoff := models.Location{Latitude: 37.7849, Longitude: -122.4094}

	ride, err := rideService.RequestRide("R1", pickup, dropoff)
	if err != nil {
		fmt.Printf("Request ride error: %v\n", err)
		return
	}
	fmt.Printf("Ride requested: %s, Driver: %s\n", ride.ID, ride.DriverID)

	// Start and complete ride
	_ = rideService.StartRide(ride.ID)
	time.Sleep(100 * time.Millisecond) // Simulate ride duration

	completedRide, _ := rideService.CompleteRide(ride.ID)
	fmt.Printf("Ride completed: Fare=$%.2f, Distance=%.2f km\n", completedRide.Fare, completedRide.Distance)

	// Add ratings
	_ = rideService.AddRating(ride.ID, "R1", "D1", 5.0, "Great driver!")
	fmt.Println("Demo completed successfully.")
}
