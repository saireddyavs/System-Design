package services

import (
	"elevator-system/internal/interfaces"
	"elevator-system/internal/models"
	"sync"
)

// BuildingController implements ElevatorController - the main API for the elevator system.
// Facade Pattern: Simplifies the complex subsystem behind a simple interface.
// SOLID-DIP: Depends on interfaces, not concrete implementations.
type BuildingController struct {
	dispatcher *DispatcherService
	building   *models.Building
	mu         sync.RWMutex
}

// NewBuildingController creates the complete elevator system.
// Dependency Injection: Strategy and building are injected.
func NewBuildingController(building *models.Building, strategy interfaces.SchedulingStrategy) *BuildingController {
	dispatcher := NewDispatcherService(building, strategy)

	// Create and register elevator services
	for _, elevator := range building.GetElevators() {
		svc := NewElevatorService(elevator, building, strategy)
		dispatcher.RegisterElevatorService(svc)
		svc.Start()
	}
	dispatcher.Start()

	return &BuildingController{
		dispatcher: dispatcher,
		building:   building,
	}
}

// SubmitRequest adds a request to the system.
func (c *BuildingController) SubmitRequest(req *models.Request) error {
	return c.dispatcher.SubmitRequest(req)
}

// GetStatus returns current status of all elevators.
func (c *BuildingController) GetStatus() []models.ElevatorStatus {
	return c.dispatcher.GetStatus()
}

// EmergencyStop stops a specific elevator.
func (c *BuildingController) EmergencyStop(elevatorID string) error {
	return c.dispatcher.EmergencyStop(elevatorID)
}

// ResumeElevator resumes an elevator after emergency/maintenance.
func (c *BuildingController) ResumeElevator(elevatorID string) error {
	return c.dispatcher.ResumeElevator(elevatorID)
}

// SetPeakMode enables/disables peak period handling.
func (c *BuildingController) SetPeakMode(enabled bool) {
	c.dispatcher.SetPeakMode(enabled)
}

// Stop shuts down the controller and all elevators.
func (c *BuildingController) Stop() {
	c.dispatcher.Stop()
}
