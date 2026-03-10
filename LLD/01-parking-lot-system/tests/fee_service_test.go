package tests

import (
	"parking-lot-system/internal/models"
	"parking-lot-system/internal/services"
	"parking-lot-system/internal/strategies"
	"testing"
	"time"
)

func TestFeeService_CalculateFee_Motorcycle(t *testing.T) {
	calc := strategies.NewHourlyFeeStrategy()
	fs := services.NewFeeService(calc)

	mc := models.NewVehicle(models.VehicleTypeMotorcycle, "MC-001")
	ticket := models.NewTicket("T1", mc, "S1", "L1")

	// 1 hour = $10 = 1000 cents
	fee := fs.CalculateFee(ticket, time.Hour)
	if fee != 1000 {
		t.Errorf("expected 1000 cents for 1hr motorcycle, got %d", fee)
	}

	// 2 hours = $20 = 2000 cents
	fee = fs.CalculateFee(ticket, 2*time.Hour)
	if fee != 2000 {
		t.Errorf("expected 2000 cents for 2hr motorcycle, got %d", fee)
	}
}

func TestFeeService_CalculateFee_Car(t *testing.T) {
	calc := strategies.NewHourlyFeeStrategy()
	fs := services.NewFeeService(calc)

	car := models.NewVehicle(models.VehicleTypeCar, "CAR-001")
	ticket := models.NewTicket("T1", car, "M1", "L1")

	// 1 hour = $20 = 2000 cents
	fee := fs.CalculateFee(ticket, time.Hour)
	if fee != 2000 {
		t.Errorf("expected 2000 cents for 1hr car, got %d", fee)
	}

	// 3 hours = $60 = 6000 cents
	fee = fs.CalculateFee(ticket, 3*time.Hour)
	if fee != 6000 {
		t.Errorf("expected 6000 cents for 3hr car, got %d", fee)
	}
}

func TestFeeService_CalculateFee_Bus(t *testing.T) {
	calc := strategies.NewHourlyFeeStrategy()
	fs := services.NewFeeService(calc)

	bus := models.NewVehicle(models.VehicleTypeBus, "BUS-001")
	ticket := models.NewTicket("T1", bus, "L1", "L1")

	// 1 hour = $50 = 5000 cents
	fee := fs.CalculateFee(ticket, time.Hour)
	if fee != 5000 {
		t.Errorf("expected 5000 cents for 1hr bus, got %d", fee)
	}
}

func TestFeeService_CalculateFee_MinimumOneHour(t *testing.T) {
	calc := strategies.NewHourlyFeeStrategy()
	fs := services.NewFeeService(calc)

	car := models.NewVehicle(models.VehicleTypeCar, "CAR-001")
	ticket := models.NewTicket("T1", car, "M1", "L1")

	// 30 min should charge for 1 hour
	fee := fs.CalculateFee(ticket, 30*time.Minute)
	if fee != 2000 {
		t.Errorf("expected 2000 cents (1hr min) for 30min car, got %d", fee)
	}

	// 5 min should charge for 1 hour
	fee = fs.CalculateFee(ticket, 5*time.Minute)
	if fee != 2000 {
		t.Errorf("expected 2000 cents (1hr min) for 5min car, got %d", fee)
	}
}

func TestFeeService_CalculateFee_PartialHours(t *testing.T) {
	calc := strategies.NewHourlyFeeStrategy()
	fs := services.NewFeeService(calc)

	car := models.NewVehicle(models.VehicleTypeCar, "CAR-001")
	ticket := models.NewTicket("T1", car, "M1", "L1")

	// 2.5 hours = 3 hours charged
	fee := fs.CalculateFee(ticket, 2*time.Hour+30*time.Minute)
	if fee != 6000 {
		t.Errorf("expected 6000 cents for 2.5hr car (3hr charged), got %d", fee)
	}
}

func TestFeeService_CalculateFeeForVehicle(t *testing.T) {
	calc := strategies.NewHourlyFeeStrategy()
	fs := services.NewFeeService(calc)

	car := models.NewVehicle(models.VehicleTypeCar, "CAR-001")
	fee := fs.CalculateFeeForVehicle(car, time.Hour)
	if fee != 2000 {
		t.Errorf("expected 2000 cents, got %d", fee)
	}
}

func TestFeeService_CustomRates(t *testing.T) {
	rates := map[models.VehicleType]int64{
		models.VehicleTypeMotorcycle: 500,
		models.VehicleTypeCar:       1500,
		models.VehicleTypeBus:       3000,
	}
	calc := strategies.NewHourlyFeeStrategyWithRates(rates)
	fs := services.NewFeeService(calc)

	car := models.NewVehicle(models.VehicleTypeCar, "CAR-001")
	ticket := models.NewTicket("T1", car, "M1", "L1")

	fee := fs.CalculateFee(ticket, time.Hour)
	if fee != 1500 {
		t.Errorf("expected 1500 cents with custom rate, got %d", fee)
	}
}
