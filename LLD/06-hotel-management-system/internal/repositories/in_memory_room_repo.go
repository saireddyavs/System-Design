package repositories

import (
	"errors"
	"hotel-management-system/internal/interfaces"
	"hotel-management-system/internal/models"
	"sync"
	"time"
)

var ErrRoomNotFound = errors.New("room not found")

// InMemoryRoomRepository implements RoomRepository with thread-safe in-memory storage
type InMemoryRoomRepository struct {
	rooms map[string]*models.Room
	mu    sync.RWMutex
}

// NewInMemoryRoomRepository creates a new in-memory room repository
func NewInMemoryRoomRepository() interfaces.RoomRepository {
	return &InMemoryRoomRepository{
		rooms: make(map[string]*models.Room),
	}
}

func (r *InMemoryRoomRepository) Create(room *models.Room) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.rooms[room.ID]; exists {
		return errors.New("room already exists")
	}
	r.rooms[room.ID] = room
	return nil
}

func (r *InMemoryRoomRepository) GetByID(id string) (*models.Room, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	room, exists := r.rooms[id]
	if !exists {
		return nil, ErrRoomNotFound
	}
	return room, nil
}

func (r *InMemoryRoomRepository) Update(room *models.Room) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.rooms[room.ID]; !exists {
		return ErrRoomNotFound
	}
	r.rooms[room.ID] = room
	return nil
}

// GetAvailableRooms requires booking repo for overlap check - we use a callback pattern
// For standalone use, we check room status only. Booking overlap is checked in service.
func (r *InMemoryRoomRepository) GetAvailableRooms(checkIn, checkOut time.Time, roomType *models.RoomType) ([]*models.Room, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Room
	for _, room := range r.rooms {
		if room.Status != models.RoomStatusAvailable && room.Status != models.RoomStatusReserved {
			continue
		}
		if roomType != nil && room.Type != *roomType {
			continue
		}
		result = append(result, room)
	}
	return result, nil
}
