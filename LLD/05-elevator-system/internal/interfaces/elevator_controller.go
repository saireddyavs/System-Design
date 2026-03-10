package interfaces

import "elevator-system/internal/models"

// ElevatorController defines the contract for elevator system control.
// SOLID-ISP: Focused interface for controller operations.
type ElevatorController interface {
	// SubmitRequest adds a request to the system.
	SubmitRequest(req *models.Request) error

	// GetStatus returns current status of all elevators.
	GetStatus() []models.ElevatorStatus

	// EmergencyStop stops a specific elevator.
	EmergencyStop(elevatorID string) error

	// ResumeElevator resumes an elevator after emergency/maintenance.
	ResumeElevator(elevatorID string) error

	// SetPeakMode enables/disables peak period handling.
	SetPeakMode(enabled bool)
}
