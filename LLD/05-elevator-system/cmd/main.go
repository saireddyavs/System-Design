// Package main demonstrates the elevator system.
package main

import (
	"fmt"
	"elevator-system/internal/models"
	"elevator-system/internal/services"
	"elevator-system/internal/strategies"
	"time"
)

func main() {
	// Create building: 10 floors, 2 elevators
	building := models.NewBuilding("B1", "Main Tower", 10, 2)

	// Use LOOK strategy (more efficient than SCAN)
	strategy := strategies.NewLookStrategy()

	// Create controller (starts elevators and dispatcher)
	ctrl := services.NewBuildingController(building, strategy)
	defer ctrl.Stop()

	fmt.Println("=== Elevator System Demo ===")
	fmt.Printf("Building: %s, Floors: 0-%d, Elevators: %d\n", building.Name, building.TotalFloors, len(building.GetElevators()))
	fmt.Printf("Strategy: %s\n\n", strategy.Name())

	// Submit external requests (floor buttons)
	req1 := models.NewExternalRequest(3, models.DirectionUp)
	req2 := models.NewExternalRequest(7, models.DirectionDown)
	req3 := models.NewExternalRequest(0, models.DirectionUp)

	ctrl.SubmitRequest(req1)
	ctrl.SubmitRequest(req2)
	ctrl.SubmitRequest(req3)

	// Internal request (passenger in elevator at floor 3 going to 8)
	req4 := models.NewInternalRequest(3, 8)
	ctrl.SubmitRequest(req4)

	fmt.Println("Submitted 4 requests. Running for 5 seconds...")

	// Display status periodically
	for i := 0; i < 10; i++ {
		time.Sleep(500 * time.Millisecond)
		printStatus(ctrl.GetStatus())
		fmt.Println("---")
	}

	fmt.Println("Demo complete.")
}

func printStatus(statuses []models.ElevatorStatus) {
	for _, s := range statuses {
		overweight := ""
		if s.IsOverweight {
			overweight = " [OVERWEIGHT]"
		}
		fmt.Printf("Elevator %s: Floor %d, %s, %s, Load %d/%d kg, Queue: %d%s\n",
			s.ID, s.CurrentFloor, s.Direction, s.State, s.CurrentLoad, s.Capacity, s.QueueLength, overweight)
	}
}
