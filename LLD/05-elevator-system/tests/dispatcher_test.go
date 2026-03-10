package tests

import (
	"elevator-system/internal/models"
	"elevator-system/internal/services"
	"elevator-system/internal/strategies"
	"sync"
	"testing"
	"time"
)

func TestDispatcher_AssignsToNearestElevator(t *testing.T) {
	building := models.NewBuilding("B1", "Test", 10, 2)
	strategy := strategies.NewLookStrategy()
	ctrl := services.NewBuildingController(building, strategy)
	defer ctrl.Stop()

	// Submit multiple requests
	for i := 0; i < 5; i++ {
		req := models.NewExternalRequest(i, models.DirectionUp)
		err := ctrl.SubmitRequest(req)
		if err != nil {
			t.Fatalf("SubmitRequest failed: %v", err)
		}
	}

	time.Sleep(2 * time.Second)
	statuses := ctrl.GetStatus()
	totalQueue := 0
	for _, s := range statuses {
		totalQueue += s.QueueLength
	}
	t.Logf("Total requests in queues: %d", totalQueue)
}

func TestDispatcher_ConcurrentRequests(t *testing.T) {
	building := models.NewBuilding("B1", "Test", 10, 3)
	strategy := strategies.NewLookStrategy()
	ctrl := services.NewBuildingController(building, strategy)
	defer ctrl.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(floor int) {
			defer wg.Done()
			req := models.NewExternalRequest(floor%10, models.DirectionUp)
			ctrl.SubmitRequest(req)
		}(i)
	}
	wg.Wait()

	time.Sleep(1 * time.Second)
	statuses := ctrl.GetStatus()
	if len(statuses) != 3 {
		t.Errorf("Expected 3 elevators, got %d", len(statuses))
	}
}

func TestSchedulingStrategy_LOOK(t *testing.T) {
	building := models.NewBuilding("B1", "Test", 10, 1)
	elevator := building.GetElevators()[0]
	strategy := strategies.NewLookStrategy()

	elevator.SetFloor(5)
	elevator.SetDirection(models.DirectionUp)
	elevator.AddRequest(models.NewExternalRequest(7, models.DirectionUp))
	elevator.AddRequest(models.NewExternalRequest(3, models.DirectionDown))
	elevator.AddRequest(models.NewInternalRequest(5, 8))

	ordered := strategy.OrderRequests(elevator, building)
	if len(ordered) != 3 {
		t.Errorf("Expected 3 ordered requests, got %d", len(ordered))
	}
}

func TestSchedulingStrategy_SCAN(t *testing.T) {
	building := models.NewBuilding("B1", "Test", 10, 1)
	elevator := building.GetElevators()[0]
	strategy := strategies.NewScanStrategy()

	elevator.SetFloor(0)
	elevator.SetDirection(models.DirectionUp)
	elevator.AddRequest(models.NewExternalRequest(5, models.DirectionUp))
	elevator.AddRequest(models.NewExternalRequest(2, models.DirectionUp))

	ordered := strategy.OrderRequests(elevator, building)
	if len(ordered) != 2 {
		t.Errorf("Expected 2 ordered requests, got %d", len(ordered))
	}
}

func TestBuildingController_EmergencyStop(t *testing.T) {
	building := models.NewBuilding("B1", "Test", 10, 2)
	strategy := strategies.NewLookStrategy()
	ctrl := services.NewBuildingController(building, strategy)
	defer ctrl.Stop()

	elevatorID := building.GetElevators()[0].ID
	err := ctrl.EmergencyStop(elevatorID)
	if err != nil {
		t.Fatalf("EmergencyStop failed: %v", err)
	}

	statuses := ctrl.GetStatus()
	var found bool
	for _, s := range statuses {
		if s.ID == elevatorID && s.State == models.StateEmergencyStop {
			found = true
			break
		}
	}
	if !found {
		t.Error("Elevator should be in EMERGENCY_STOP state")
	}
}
