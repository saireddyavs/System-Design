package models

import "sync"

// RoomType represents the category of a room
type RoomType string

const (
	RoomTypeSingle RoomType = "Single"
	RoomTypeDouble RoomType = "Double"
	RoomTypeDeluxe RoomType = "Deluxe"
	RoomTypeSuite  RoomType = "Suite"
)

// RoomStatus represents the current availability status
type RoomStatus string

const (
	RoomStatusAvailable RoomStatus = "Available"
	RoomStatusOccupied  RoomStatus = "Occupied"
	RoomStatusReserved  RoomStatus = "Reserved"
)

// Room represents a hotel room entity
type Room struct {
	ID            string
	Number        string
	Type          RoomType
	Floor         int
	BasePrice     float64 // Price per night (base)
	Status        RoomStatus
	Amenities     []string
	mu            sync.RWMutex
}

// NewRoom creates a new Room instance
func NewRoom(id, number string, roomType RoomType, floor int, basePrice float64, amenities []string) *Room {
	return &Room{
		ID:        id,
		Number:    number,
		Type:      roomType,
		Floor:     floor,
		BasePrice: basePrice,
		Status:    RoomStatusAvailable,
		Amenities: amenities,
	}
}

// GetStatus returns the current room status (thread-safe)
func (r *Room) GetStatus() RoomStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Status
}

// SetStatus updates the room status (thread-safe)
func (r *Room) SetStatus(status RoomStatus) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Status = status
}

// GetBasePrice returns the base price (thread-safe)
func (r *Room) GetBasePrice() float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.BasePrice
}

// IsAvailable checks if room can be booked
func (r *Room) IsAvailable() bool {
	return r.GetStatus() == RoomStatusAvailable
}

// RoomTypePrices holds default base prices per room type
var RoomTypePrices = map[RoomType]float64{
	RoomTypeSingle: 100.0,
	RoomTypeDouble: 150.0,
	RoomTypeDeluxe: 250.0,
	RoomTypeSuite:  400.0,
}
