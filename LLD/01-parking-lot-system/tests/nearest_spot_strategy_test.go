package tests

import (
	"parking-lot-system/internal/models"
	"parking-lot-system/internal/strategies"
	"testing"
)

func TestNearestSpotStrategy_FindSpot(t *testing.T) {
	strategy := strategies.NewNearestSpotStrategy()

	spots := []*models.ParkingSpot{
		models.NewParkingSpot("S1", "L1", models.SpotSizeSmall),
		models.NewParkingSpot("S2", "L1", models.SpotSizeMedium),
		models.NewParkingSpot("S3", "L1", models.SpotSizeLarge),
	}

	// Motorcycle should get first small spot
	mc := models.NewVehicle(models.VehicleTypeMotorcycle, "MC-001")
	spot := strategy.FindSpot(mc, spots)
	if spot == nil {
		t.Fatal("expected to find spot")
	}
	if spot.ID != "S1" {
		t.Errorf("expected S1 (nearest), got %s", spot.ID)
	}

	// Car should get first medium spot
	car := models.NewVehicle(models.VehicleTypeCar, "CAR-001")
	spot = strategy.FindSpot(car, spots)
	if spot == nil {
		t.Fatal("expected to find spot")
	}
	if spot.ID != "S2" {
		t.Errorf("expected S2, got %s", spot.ID)
	}

	// Bus should get large spot
	bus := models.NewVehicle(models.VehicleTypeBus, "BUS-001")
	spot = strategy.FindSpot(bus, spots)
	if spot == nil {
		t.Fatal("expected to find spot")
	}
	if spot.ID != "S3" {
		t.Errorf("expected S3, got %s", spot.ID)
	}
}

func TestNearestSpotStrategy_NoSpot(t *testing.T) {
	strategy := strategies.NewNearestSpotStrategy()
	spots := []*models.ParkingSpot{
		models.NewParkingSpot("S1", "L1", models.SpotSizeSmall),
	}

	bus := models.NewVehicle(models.VehicleTypeBus, "BUS-001")
	spot := strategy.FindSpot(bus, spots)
	if spot != nil {
		t.Error("expected nil when no suitable spot")
	}
}

func TestNearestSpotStrategy_EmptySpots(t *testing.T) {
	strategy := strategies.NewNearestSpotStrategy()
	spots := []*models.ParkingSpot{}

	car := models.NewVehicle(models.VehicleTypeCar, "CAR-001")
	spot := strategy.FindSpot(car, spots)
	if spot != nil {
		t.Error("expected nil for empty spots")
	}
}
