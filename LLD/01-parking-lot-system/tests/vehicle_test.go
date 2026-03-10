package tests

import (
	"parking-lot-system/internal/models"
	"testing"
)

func TestVehicleFactory_Motorcycle(t *testing.T) {
	v := models.NewVehicle(models.VehicleTypeMotorcycle, "MC-001")
	if v.GetType() != models.VehicleTypeMotorcycle {
		t.Errorf("expected Motorcycle, got %s", v.GetType().String())
	}
	if v.GetLicensePlate() != "MC-001" {
		t.Errorf("expected MC-001, got %s", v.GetLicensePlate())
	}
	if v.GetRequiredSpotSize() != models.SpotSizeSmall {
		t.Errorf("expected Small spot, got %s", v.GetRequiredSpotSize().String())
	}
}

func TestVehicleFactory_Car(t *testing.T) {
	v := models.NewVehicle(models.VehicleTypeCar, "CAR-001")
	if v.GetRequiredSpotSize() != models.SpotSizeMedium {
		t.Errorf("expected Medium spot, got %s", v.GetRequiredSpotSize().String())
	}
}

func TestVehicleFactory_Bus(t *testing.T) {
	v := models.NewVehicle(models.VehicleTypeBus, "BUS-001")
	if v.GetRequiredSpotSize() != models.SpotSizeLarge {
		t.Errorf("expected Large spot, got %s", v.GetRequiredSpotSize().String())
	}
}

func TestParkingSpot_CanFit(t *testing.T) {
	spot := models.NewParkingSpot("S1", "L1", models.SpotSizeMedium)

	// Small vehicle fits in medium spot
	mc := models.NewVehicle(models.VehicleTypeMotorcycle, "MC-001")
	if !spot.CanFit(mc) {
		t.Error("motorcycle should fit in medium spot")
	}

	// Car fits in medium spot
	car := models.NewVehicle(models.VehicleTypeCar, "CAR-001")
	if !spot.CanFit(car) {
		t.Error("car should fit in medium spot")
	}

	// Bus does not fit in medium spot
	bus := models.NewVehicle(models.VehicleTypeBus, "BUS-001")
	if spot.CanFit(bus) {
		t.Error("bus should not fit in medium spot")
	}
}

func TestParkingSpot_ParkUnpark(t *testing.T) {
	spot := models.NewParkingSpot("S1", "L1", models.SpotSizeSmall)
	mc := models.NewVehicle(models.VehicleTypeMotorcycle, "MC-001")

	if !spot.IsAvailable() {
		t.Error("spot should be available initially")
	}

	if !spot.Park(mc) {
		t.Error("park should succeed")
	}
	if spot.IsAvailable() {
		t.Error("spot should be occupied")
	}

	vehicle, duration := spot.Unpark()
	if vehicle == nil {
		t.Fatal("expected vehicle from unpark")
	}
	if vehicle.GetLicensePlate() != "MC-001" {
		t.Errorf("expected MC-001, got %s", vehicle.GetLicensePlate())
	}
	if duration < 0 {
		t.Error("duration should be non-negative")
	}

	// Double unpark returns nil
	vehicle2, _ := spot.Unpark()
	if vehicle2 != nil {
		t.Error("second unpark should return nil")
	}
}
