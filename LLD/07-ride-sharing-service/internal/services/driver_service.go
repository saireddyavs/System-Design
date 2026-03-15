package services

import (
	"errors"
	"ride-sharing-service/internal/interfaces"
	"ride-sharing-service/internal/models"
)

var (
	ErrDriverAlreadyExists = errors.New("driver already exists")
)

// DriverService handles driver business logic
type DriverService struct {
	driverRepo interfaces.DriverRepository
}

// NewDriverService creates a new driver service
func NewDriverService(driverRepo interfaces.DriverRepository) *DriverService {
	return &DriverService{driverRepo: driverRepo}
}

// RegisterDriver registers a new driver
func (s *DriverService) RegisterDriver(id, name, phone string, vehicle models.Vehicle) (*models.Driver, error) {
	existing, _ := s.driverRepo.GetByID(id)
	if existing != nil {
		return nil, ErrDriverAlreadyExists
	}
	driver := models.NewDriver(id, name, phone, vehicle)
	if err := s.driverRepo.Create(driver); err != nil {
		return nil, err
	}
	return driver, nil
}

// SetAvailability toggles driver availability
func (s *DriverService) SetAvailability(driverID string, available bool) error {
	driver, err := s.driverRepo.GetByID(driverID)
	if err != nil {
		return err
	}
	if available {
		driver.SetStatus(models.DriverStatusAvailable)
	} else {
		driver.SetStatus(models.DriverStatusOffline)
	}
	return s.driverRepo.Update(driver)
}

// UpdateLocation updates driver's location
func (s *DriverService) UpdateLocation(driverID string, location models.Location) error {
	driver, err := s.driverRepo.GetByID(driverID)
	if err != nil {
		return err
	}
	driver.UpdateLocation(location)
	return s.driverRepo.Update(driver)
}

// GoOnline sets driver status to available
func (s *DriverService) GoOnline(driverID string) error {
	return s.SetAvailability(driverID, true)
}

