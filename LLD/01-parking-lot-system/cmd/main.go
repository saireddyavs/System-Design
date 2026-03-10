// Package main demonstrates the Parking Lot System.
package main

import (
	"fmt"
	"log"
	"parking-lot-system/internal/models"
	"parking-lot-system/internal/services"
	"parking-lot-system/internal/strategies"
	"time"
)

func main() {
	// Initialize parking lot with levels and spots
	lot := setupParkingLot()
	lot.Initialize(createLevels())

	// Create services with strategies (DIP: inject dependencies)
	parkingStrategy := strategies.NewNearestSpotStrategy()
	feeCalculator := strategies.NewHourlyFeeStrategy()

	parkingService := services.NewParkingService(lot, parkingStrategy)
	feeService := services.NewFeeService(feeCalculator)

	// Demo: Park vehicles
	fmt.Println("=== Parking Lot System Demo ===")

	// Park a motorcycle
	motorcycle := models.NewVehicle(models.VehicleTypeMotorcycle, "MC-001")
	ticket1, err := parkingService.Park(motorcycle)
	if err != nil {
		log.Fatalf("Park motorcycle: %v", err)
	}
	fmt.Printf("Parked motorcycle: %s, Ticket: %s, Spot: %s\n",
		motorcycle.GetLicensePlate(), ticket1.ID, ticket1.SpotID)

	// Park a car
	car := models.NewVehicle(models.VehicleTypeCar, "CAR-001")
	ticket2, err := parkingService.Park(car)
	if err != nil {
		log.Fatalf("Park car: %v", err)
	}
	fmt.Printf("Parked car: %s, Ticket: %s, Spot: %s\n",
		car.GetLicensePlate(), ticket2.ID, ticket2.SpotID)

	// Park another car
	car2 := models.NewVehicle(models.VehicleTypeCar, "CAR-002")
	ticket3, err := parkingService.Park(car2)
	if err != nil {
		log.Fatalf("Park car2: %v", err)
	}
	fmt.Printf("Parked car: %s, Ticket: %s, Spot: %s\n",
		car2.GetLicensePlate(), ticket3.ID, ticket3.SpotID)

	// Check available spots
	availableForCar := parkingService.GetAvailableSpotsCount(car)
	fmt.Printf("\nAvailable spots for car: %d\n", availableForCar)

	// Simulate time passing for fee calculation
	time.Sleep(100 * time.Millisecond)

	// Unpark and calculate fee
	ticket, vehicle, err := parkingService.Unpark(ticket2.ID)
	if err != nil {
		log.Fatalf("Unpark: %v", err)
	}
	fee := feeService.CalculateFee(ticket, 0)
	fmt.Printf("\nUnparked vehicle: %s, Fee: %d cents ($%.2f)\n",
		vehicle.GetLicensePlate(), fee, float64(fee)/100)

	// Unpark by license plate
	finalTicket, finalVehicle, err := parkingService.Unpark("MC-001")
	if err != nil {
		log.Fatalf("Unpark by license: %v", err)
	}
	finalFee := feeService.CalculateFee(finalTicket, 0)
	fmt.Printf("Unparked by license: %s, Fee: %d cents ($%.2f)\n",
		finalVehicle.GetLicensePlate(), finalFee, float64(finalFee)/100)

	// Unpark last vehicle
	_, _, _ = parkingService.Unpark(ticket3.ID)

	fmt.Println("\n=== Demo Complete ===")
}

func setupParkingLot() *models.ParkingLot {
	return models.GetInstance()
}

func createLevels() []*models.ParkingLevel {
	// Level 1: 2 small, 2 medium, 1 large
	level1Spots := []*models.ParkingSpot{
		models.NewParkingSpot("L1-S1", "L1", models.SpotSizeSmall),
		models.NewParkingSpot("L1-S2", "L1", models.SpotSizeSmall),
		models.NewParkingSpot("L1-M1", "L1", models.SpotSizeMedium),
		models.NewParkingSpot("L1-M2", "L1", models.SpotSizeMedium),
		models.NewParkingSpot("L1-L1", "L1", models.SpotSizeLarge),
	}
	level1 := models.NewParkingLevel("L1", "Level 1", level1Spots)

	// Level 2: 1 small, 2 medium, 2 large
	level2Spots := []*models.ParkingSpot{
		models.NewParkingSpot("L2-S1", "L2", models.SpotSizeSmall),
		models.NewParkingSpot("L2-M1", "L2", models.SpotSizeMedium),
		models.NewParkingSpot("L2-M2", "L2", models.SpotSizeMedium),
		models.NewParkingSpot("L2-L1", "L2", models.SpotSizeLarge),
		models.NewParkingSpot("L2-L2", "L2", models.SpotSizeLarge),
	}
	level2 := models.NewParkingLevel("L2", "Level 2", level2Spots)

	return []*models.ParkingLevel{level1, level2}
}
