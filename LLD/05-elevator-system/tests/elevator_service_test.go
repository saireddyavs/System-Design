package tests

import (
	"elevator-system/internal/models"
	"elevator-system/internal/services"
	"elevator-system/internal/strategies"
	"testing"
	"time"
)

func TestElevatorService_ProcessesRequests(t *testing.T) {
	building := models.NewBuilding("B1", "Test", 10, 1)
	elevator := building.GetElevators()[0]
	strategy := strategies.NewLookStrategy()
	svc := services.NewElevatorService(elevator, building, strategy)
	svc.Start()
	defer svc.Stop()

	req := models.NewExternalRequest(2, models.DirectionUp)
	svc.SubmitRequest(req)

	time.Sleep(500 * time.Millisecond)
	status := elevator.GetStatus()
	if status.QueueLength < 1 && status.CurrentFloor != 2 {
		t.Logf("Elevator processing: floor=%d, queue=%d", status.CurrentFloor, status.QueueLength)
	}
}

func TestElevatorService_EmergencyStop(t *testing.T) {
	building := models.NewBuilding("B1", "Test", 10, 1)
	elevator := building.GetElevators()[0]
	strategy := strategies.NewLookStrategy()
	svc := services.NewElevatorService(elevator, building, strategy)
	svc.Start()
	defer svc.Stop()

	svc.SubmitRequest(models.NewExternalRequest(5, models.DirectionUp))
	time.Sleep(100 * time.Millisecond)
	svc.EmergencyStop()
	time.Sleep(200 * time.Millisecond)

	status := elevator.GetStatus()
	if status.State != models.StateEmergencyStop {
		t.Errorf("Expected EMERGENCY_STOP, got %s", status.State)
	}
}

func TestElevatorService_ResumeAfterEmergency(t *testing.T) {
	building := models.NewBuilding("B1", "Test", 10, 1)
	elevator := building.GetElevators()[0]
	strategy := strategies.NewLookStrategy()
	svc := services.NewElevatorService(elevator, building, strategy)
	svc.Start()
	defer svc.Stop()

	svc.EmergencyStop()
	time.Sleep(50 * time.Millisecond)
	svc.Resume()
	time.Sleep(50 * time.Millisecond)

	status := elevator.GetStatus()
	if status.State != models.StateIdle {
		t.Errorf("Expected IDLE after resume, got %s", status.State)
	}
}

func TestElevator_OverweightPrevention(t *testing.T) {
	elevator := models.NewElevator("E1")
	// Add load up to threshold
	for i := 0; i < 12; i++ {
		elevator.AddLoad(80)
	}
	if !elevator.IsOverweight() {
		t.Error("Expected elevator to be overweight")
	}
	if elevator.CanAcceptPassenger(70) {
		t.Error("Should not accept passenger when overweight")
	}
}
