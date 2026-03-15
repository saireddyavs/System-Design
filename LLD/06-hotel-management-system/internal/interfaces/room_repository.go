package interfaces

import (
	"hotel-management-system/internal/models"
	"time"
)

// RoomRepository defines data access for rooms (Repository pattern)
// S - Single Responsibility: Only room persistence
// D - Dependency Inversion: Services depend on this interface, not concrete impl
type RoomRepository interface {
	Create(room *models.Room) error
	GetByID(id string) (*models.Room, error)
	Update(room *models.Room) error

	// Availability queries
	GetAvailableRooms(checkIn, checkOut time.Time, roomType *models.RoomType) ([]*models.Room, error)
}
