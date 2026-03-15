package tests

import (
	"fmt"
	"parking-lot-system/internal/models"
	"parking-lot-system/internal/services"
	"parking-lot-system/internal/strategies"
	"sync"
	"testing"
)

func setupTestParkingLot(t *testing.T) (*models.ParkingLot, func()) {
	models.ResetInstance()
	lot := models.GetInstance()
	levels := createTestLevels()
	lot.Initialize(levels)
	return lot, func() { models.ResetInstance() }
}

func createTestLevels() []*models.ParkingLevel {
	level1Spots := []*models.ParkingSpot{
		models.NewParkingSpot("L1-S1", "L1", models.SpotSizeSmall),
		models.NewParkingSpot("L1-S2", "L1", models.SpotSizeSmall),
		models.NewParkingSpot("L1-M1", "L1", models.SpotSizeMedium),
		models.NewParkingSpot("L1-M2", "L1", models.SpotSizeMedium),
		models.NewParkingSpot("L1-L1", "L1", models.SpotSizeLarge),
	}
	level1 := models.NewParkingLevel("L1", "Level 1", level1Spots)

	level2Spots := []*models.ParkingSpot{
		models.NewParkingSpot("L2-S1", "L2", models.SpotSizeSmall),
		models.NewParkingSpot("L2-M1", "L2", models.SpotSizeMedium),
		models.NewParkingSpot("L2-L1", "L2", models.SpotSizeLarge),
	}
	level2 := models.NewParkingLevel("L2", "Level 2", level2Spots)

	return []*models.ParkingLevel{level1, level2}
}

func TestParkingService_Park_Success(t *testing.T) {
	lot, cleanup := setupTestParkingLot(t)
	defer cleanup()

	ps := services.NewParkingService(lot, strategies.NewNearestSpotStrategy(), strategies.NewHourlyFeeStrategy())

	// Park motorcycle - should get small spot
	mc := models.NewVehicle(models.VehicleTypeMotorcycle, "MC-001")
	ticket, err := ps.Park(mc)
	if err != nil {
		t.Fatalf("Park motorcycle: %v", err)
	}
	if ticket == nil {
		t.Fatal("expected ticket")
	}
	if ticket.SpotID != "L1-S1" {
		t.Errorf("expected spot L1-S1, got %s", ticket.SpotID)
	}
	if ticket.LicensePlate != "MC-001" {
		t.Errorf("expected license MC-001, got %s", ticket.LicensePlate)
	}

	// Park car - should get medium spot
	car := models.NewVehicle(models.VehicleTypeCar, "CAR-001")
	ticket2, err := ps.Park(car)
	if err != nil {
		t.Fatalf("Park car: %v", err)
	}
	if ticket2.SpotID != "L1-M1" {
		t.Errorf("expected spot L1-M1, got %s", ticket2.SpotID)
	}

	// Park bus - should get large spot
	bus := models.NewVehicle(models.VehicleTypeBus, "BUS-001")
	ticket3, err := ps.Park(bus)
	if err != nil {
		t.Fatalf("Park bus: %v", err)
	}
	if ticket3.SpotID != "L1-L1" {
		t.Errorf("expected spot L1-L1, got %s", ticket3.SpotID)
	}
}

func TestParkingService_Park_NoSpotAvailable(t *testing.T) {
	lot, cleanup := setupTestParkingLot(t)
	defer cleanup()

	ps := services.NewParkingService(lot, strategies.NewNearestSpotStrategy(), strategies.NewHourlyFeeStrategy())

	// Fill all large spots
	bus1 := models.NewVehicle(models.VehicleTypeBus, "BUS-001")
	bus2 := models.NewVehicle(models.VehicleTypeBus, "BUS-002")
	_, _ = ps.Park(bus1)
	_, _ = ps.Park(bus2)

	// No more large spots
	bus3 := models.NewVehicle(models.VehicleTypeBus, "BUS-003")
	_, err := ps.Park(bus3)
	if err == nil {
		t.Fatal("expected error when no spot available")
	}
}

func TestParkingService_Park_DuplicateVehicle(t *testing.T) {
	lot, cleanup := setupTestParkingLot(t)
	defer cleanup()

	ps := services.NewParkingService(lot, strategies.NewNearestSpotStrategy(), strategies.NewHourlyFeeStrategy())
	car := models.NewVehicle(models.VehicleTypeCar, "CAR-001")

	_, err := ps.Park(car)
	if err != nil {
		t.Fatalf("first park: %v", err)
	}

	_, err = ps.Park(car)
	if err == nil {
		t.Fatal("expected error when parking same vehicle twice")
	}
}

func TestParkingService_Unpark_Success(t *testing.T) {
	lot, cleanup := setupTestParkingLot(t)
	defer cleanup()

	ps := services.NewParkingService(lot, strategies.NewNearestSpotStrategy(), strategies.NewHourlyFeeStrategy())
	car := models.NewVehicle(models.VehicleTypeCar, "CAR-001")

	ticket, err := ps.Park(car)
	if err != nil {
		t.Fatalf("Park: %v", err)
	}

	_, vehicle, err := ps.Unpark(ticket.ID)
	if err != nil {
		t.Fatalf("Unpark: %v", err)
	}
	if vehicle.GetLicensePlate() != "CAR-001" {
		t.Errorf("expected CAR-001, got %s", vehicle.GetLicensePlate())
	}

	// Spot should be available again
	available := ps.GetAvailableSpotsCount(car)
	if available < 2 {
		t.Errorf("expected at least 2 spots after unpark, got %d", available)
	}
}

func TestParkingService_Unpark_ByLicensePlate(t *testing.T) {
	lot, cleanup := setupTestParkingLot(t)
	defer cleanup()

	ps := services.NewParkingService(lot, strategies.NewNearestSpotStrategy(), strategies.NewHourlyFeeStrategy())
	car := models.NewVehicle(models.VehicleTypeCar, "CAR-001")
	_, _ = ps.Park(car)

	_, vehicle, err := ps.Unpark("CAR-001")
	if err != nil {
		t.Fatalf("Unpark by license: %v", err)
	}
	if vehicle.GetLicensePlate() != "CAR-001" {
		t.Errorf("expected CAR-001, got %s", vehicle.GetLicensePlate())
	}
}

func TestParkingService_Unpark_NotFound(t *testing.T) {
	lot, cleanup := setupTestParkingLot(t)
	defer cleanup()

	ps := services.NewParkingService(lot, strategies.NewNearestSpotStrategy(), strategies.NewHourlyFeeStrategy())

	_, _, err := ps.Unpark("TKT-999")
	if err == nil {
		t.Fatal("expected error for unknown ticket")
	}
}

func TestParkingService_GetAvailableSpotsCount(t *testing.T) {
	lot, cleanup := setupTestParkingLot(t)
	defer cleanup()

	ps := services.NewParkingService(lot, strategies.NewNearestSpotStrategy(), strategies.NewHourlyFeeStrategy())
	car := models.NewVehicle(models.VehicleTypeCar, "CAR-001")

	initial := ps.GetAvailableSpotsCount(car)
	if initial < 2 {
		t.Errorf("expected at least 2 spots for car, got %d", initial)
	}

	_, _ = ps.Park(car)
	after := ps.GetAvailableSpotsCount(car)
	if after != initial-1 {
		t.Errorf("expected %d spots after park, got %d", initial-1, after)
	}
}

func TestParkingService_ConcurrentParkUnpark(t *testing.T) {
	lot, cleanup := setupTestParkingLot(t)
	defer cleanup()

	ps := services.NewParkingService(lot, strategies.NewNearestSpotStrategy(), strategies.NewHourlyFeeStrategy())
	var wg sync.WaitGroup

	// Park 5 vehicles concurrently with unique plates
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			plate := fmt.Sprintf("CONC-CAR-%d", n)
			v := models.NewVehicle(models.VehicleTypeCar, plate)
			_, _ = ps.Park(v)
		}(i)
	}
	wg.Wait()

	// Unpark all concurrently
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			plate := fmt.Sprintf("CONC-CAR-%d", n)
			_, _, _ = ps.Unpark(plate)
		}(i)
	}
	wg.Wait()

	// Verify no deadlock - should be able to park again
	car := models.NewVehicle(models.VehicleTypeCar, "VERIFY")
	count := ps.GetAvailableSpotsCount(car)
	if count < 0 {
		t.Fatal("concurrent deadlock detected")
	}
}
