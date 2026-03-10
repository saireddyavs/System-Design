package tests

import (
	"sync"
	"testing"

	"ride-sharing-service/internal/models"
	"ride-sharing-service/internal/repositories"
	"ride-sharing-service/internal/services"
	"ride-sharing-service/internal/strategies"
)

func setupRideService(t *testing.T) *services.RideService {
	driverRepo := repositories.NewInMemoryDriverRepository()
	riderRepo := repositories.NewInMemoryRiderRepository()
	rideRepo := repositories.NewInMemoryRideRepository()
	paymentRepo := repositories.NewInMemoryPaymentRepository()
	ratingRepo := repositories.NewInMemoryRatingRepository()

	driverService := services.NewDriverService(driverRepo)
	riderService := services.NewRiderService(riderRepo)
	paymentProcessor := services.NewInMemoryPaymentProcessor(paymentRepo)
	notifier := services.NewRideNotifier()

	matchingStrategy := strategies.NewNearestDriverStrategy(50)
	matchingService := services.NewMatchingService(driverRepo, matchingStrategy, 50)
	fareCalculator := strategies.NewBaseFareStrategy(2.0, 1.5, 0.25)

	// Register driver and rider
	_, err := driverService.RegisterDriver("D1", "Driver", "123", models.Vehicle{Model: "Car", Number: "X", Type: "sedan"})
	if err != nil {
		t.Fatalf("register driver: %v", err)
	}
	driverService.UpdateLocation("D1", models.Location{Latitude: 37.7750, Longitude: -122.4195})
	driverService.GoOnline("D1")

	_, err = riderService.RegisterRider("R1", "Rider", "456", "r@test.com")
	if err != nil {
		t.Fatalf("register rider: %v", err)
	}
	riderService.UpdateLocation("R1", models.Location{Latitude: 37.7750, Longitude: -122.4195})

	return services.NewRideService(
		rideRepo, driverRepo, riderRepo,
		matchingService, fareCalculator, paymentProcessor, ratingRepo, notifier,
	)
}

func TestRideService_RequestRide(t *testing.T) {
	rideService := setupRideService(t)
	pickup := models.Location{Latitude: 37.7750, Longitude: -122.4195}
	dropoff := models.Location{Latitude: 37.7849, Longitude: -122.4094}

	ride, err := rideService.RequestRide("R1", pickup, dropoff)
	if err != nil {
		t.Fatalf("request ride: %v", err)
	}
	if ride.RiderID != "R1" {
		t.Errorf("expected rider R1, got %s", ride.RiderID)
	}
	if ride.DriverID != "D1" {
		t.Errorf("expected driver D1, got %s", ride.DriverID)
	}
	if ride.GetStatus() != models.RideStatusDriverAssigned {
		t.Errorf("expected status driver_assigned, got %s", ride.GetStatus())
	}
}

func TestRideService_CompleteRide(t *testing.T) {
	rideService := setupRideService(t)
	pickup := models.Location{Latitude: 37.7750, Longitude: -122.4195}
	dropoff := models.Location{Latitude: 37.7849, Longitude: -122.4094}

	ride, _ := rideService.RequestRide("R1", pickup, dropoff)
	_ = rideService.StartRide(ride.ID)

	completed, err := rideService.CompleteRide(ride.ID)
	if err != nil {
		t.Fatalf("complete ride: %v", err)
	}
	if completed.GetStatus() != models.RideStatusCompleted {
		t.Errorf("expected status completed, got %s", completed.GetStatus())
	}
	if completed.Fare <= 0 {
		t.Errorf("expected positive fare, got %.2f", completed.Fare)
	}
}

func TestRideService_CancelBeforeStart(t *testing.T) {
	rideService := setupRideService(t)
	pickup := models.Location{Latitude: 37.7750, Longitude: -122.4195}
	dropoff := models.Location{Latitude: 37.7849, Longitude: -122.4094}

	ride, _ := rideService.RequestRide("R1", pickup, dropoff)
	penalty, err := rideService.CancelRide(ride.ID)
	if err != nil {
		t.Fatalf("cancel ride: %v", err)
	}
	if penalty != 0 {
		t.Errorf("expected no penalty when cancelling before start (driver_assigned), got %.2f", penalty)
	}

	r, _ := rideService.GetRide(ride.ID)
	if r.GetStatus() != models.RideStatusCancelled {
		t.Errorf("expected status cancelled, got %s", r.GetStatus())
	}
}

func TestRideService_CancelInProgress_AppliesPenalty(t *testing.T) {
	rideService := setupRideService(t)
	pickup := models.Location{Latitude: 37.7750, Longitude: -122.4195}
	dropoff := models.Location{Latitude: 37.7849, Longitude: -122.4094}

	ride, _ := rideService.RequestRide("R1", pickup, dropoff)
	_ = rideService.StartRide(ride.ID) // Ride now in progress

	penalty, err := rideService.CancelRide(ride.ID)
	if err != nil {
		t.Fatalf("cancel ride: %v", err)
	}
	if penalty <= 0 {
		t.Errorf("expected penalty when cancelling in progress, got %.2f", penalty)
	}
}

func TestRideService_AddRating(t *testing.T) {
	rideService := setupRideService(t)
	pickup := models.Location{Latitude: 37.7750, Longitude: -122.4195}
	dropoff := models.Location{Latitude: 37.7849, Longitude: -122.4094}

	ride, _ := rideService.RequestRide("R1", pickup, dropoff)
	_ = rideService.StartRide(ride.ID)
	_, _ = rideService.CompleteRide(ride.ID)

	err := rideService.AddRating(ride.ID, "R1", "D1", 5.0, "Great!")
	if err != nil {
		t.Fatalf("add rating: %v", err)
	}
}

func TestRideService_ConcurrentRequests(t *testing.T) {
	driverRepo := repositories.NewInMemoryDriverRepository()
	riderRepo := repositories.NewInMemoryRiderRepository()
	rideRepo := repositories.NewInMemoryRideRepository()
	paymentRepo := repositories.NewInMemoryPaymentRepository()
	ratingRepo := repositories.NewInMemoryRatingRepository()

	driverService := services.NewDriverService(driverRepo)
	riderService := services.NewRiderService(riderRepo)

	// Register multiple drivers
	for i := 1; i <= 5; i++ {
		id := string(rune('0' + i))
		_, _ = driverService.RegisterDriver("D"+id, "Driver"+id, id, models.Vehicle{})
		driverService.UpdateLocation("D"+id, models.Location{
			Latitude:  37.7750 + float64(i)*0.001,
			Longitude: -122.4195 + float64(i)*0.001,
		})
		driverService.GoOnline("D" + id)
	}

	// Register multiple riders
	for i := 1; i <= 3; i++ {
		id := string(rune('0' + i))
		_, _ = riderService.RegisterRider("R"+id, "Rider"+id, id, "r"+id+"@test.com")
		riderService.UpdateLocation("R"+id, models.Location{
			Latitude:  37.7750 + float64(i)*0.002,
			Longitude: -122.4195 + float64(i)*0.002,
		})
	}

	matchingStrategy := strategies.NewNearestDriverStrategy(50)
	matchingService := services.NewMatchingService(driverRepo, matchingStrategy, 50)
	fareCalculator := strategies.NewBaseFareStrategy(2.0, 1.5, 0.25)
	paymentProcessor := services.NewInMemoryPaymentProcessor(paymentRepo)
	notifier := services.NewRideNotifier()

	rideService := services.NewRideService(
		rideRepo, driverRepo, riderRepo,
		matchingService, fareCalculator, paymentProcessor, ratingRepo, notifier,
	)

	var wg sync.WaitGroup
	errors := make(chan error, 10)
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(riderNum int) {
			defer wg.Done()
			riderID := "R" + string(rune('0'+riderNum))
			pickup := models.Location{
				Latitude:  37.7750 + float64(riderNum)*0.002,
				Longitude: -122.4195 + float64(riderNum)*0.002,
			}
			dropoff := models.Location{
				Latitude:  pickup.Latitude + 0.01,
				Longitude: pickup.Longitude + 0.01,
			}
			_, err := rideService.RequestRide(riderID, pickup, dropoff)
			if err != nil {
				errors <- err
			}
		}(i)
	}
	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent request error: %v", err)
	}
}
