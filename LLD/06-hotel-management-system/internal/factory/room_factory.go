package factory

import (
	"hotel-management-system/internal/models"
)

// RoomFactory creates rooms by type (Factory pattern)
// Encapsulates room creation logic - O (Open/Closed): new room types without modifying factory
type RoomFactory struct {
	amenitiesByType map[models.RoomType][]string
}

// NewRoomFactory creates a new room factory
func NewRoomFactory() *RoomFactory {
	return &RoomFactory{
		amenitiesByType: map[models.RoomType][]string{
			models.RoomTypeSingle: {"TV", "WiFi", "AC"},
			models.RoomTypeDouble: {"TV", "WiFi", "AC", "Mini-bar"},
			models.RoomTypeDeluxe: {"TV", "WiFi", "AC", "Mini-bar", "Bathtub", "Sea View"},
			models.RoomTypeSuite:  {"TV", "WiFi", "AC", "Mini-bar", "Bathtub", "Sea View", "Living Room", "Kitchen"},
		},
	}
}

// CreateRoom creates a room with type-specific configuration
func (f *RoomFactory) CreateRoom(id, number string, roomType models.RoomType, floor int) *models.Room {
	basePrice := models.RoomTypePrices[roomType]
	amenities := f.amenitiesByType[roomType]
	if amenities == nil {
		amenities = []string{"TV", "WiFi"}
	}
	return models.NewRoom(id, number, roomType, floor, basePrice, amenities)
}
