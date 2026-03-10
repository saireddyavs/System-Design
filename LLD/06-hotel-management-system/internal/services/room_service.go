package services

import (
	"hotel-management-system/internal/factory"
	"hotel-management-system/internal/interfaces"
	"hotel-management-system/internal/models"
	"time"
)


// RoomService handles room operations
type RoomService struct {
	roomRepo    interfaces.RoomRepository
	bookingRepo interfaces.BookingRepository
	factory     *factory.RoomFactory
}

// NewRoomService creates a new room service
func NewRoomService(roomRepo interfaces.RoomRepository, bookingRepo interfaces.BookingRepository, factory *factory.RoomFactory) *RoomService {
	return &RoomService{
		roomRepo:    roomRepo,
		bookingRepo: bookingRepo,
		factory:     factory,
	}
}

// SearchCriteria for room search
type SearchCriteria struct {
	RoomType   *models.RoomType
	CheckIn    time.Time
	CheckOut   time.Time
	MinPrice   *float64
	MaxPrice   *float64
}

// CreateRoom creates a room using factory
func (s *RoomService) CreateRoom(id, number string, roomType models.RoomType, floor int) (*models.Room, error) {
	room := s.factory.CreateRoom(id, number, roomType, floor)
	if err := s.roomRepo.Create(room); err != nil {
		return nil, err
	}
	return room, nil
}

// GetRoom returns room by ID
func (s *RoomService) GetRoom(id string) (*models.Room, error) {
	return s.roomRepo.GetByID(id)
}

// GetAvailableRooms returns rooms available for the date range (no overlapping bookings)
func (s *RoomService) GetAvailableRooms(criteria SearchCriteria) ([]*models.Room, error) {
	if criteria.CheckOut.Before(criteria.CheckIn) || criteria.CheckOut.Equal(criteria.CheckIn) {
		return nil, ErrInvalidDateRange
	}

	candidates, err := s.roomRepo.GetAvailableRooms(criteria.CheckIn, criteria.CheckOut, criteria.RoomType)
	if err != nil {
		return nil, err
	}

	var available []*models.Room
	for _, room := range candidates {
		overlapping, err := s.bookingRepo.GetBookingsForRoomInRange(room.ID, criteria.CheckIn, criteria.CheckOut)
		if err != nil {
			continue
		}
		if len(overlapping) > 0 {
			continue
		}

		if criteria.MinPrice != nil || criteria.MaxPrice != nil {
			price := room.GetBasePrice()
			if criteria.MinPrice != nil && price < *criteria.MinPrice {
				continue
			}
			if criteria.MaxPrice != nil && price > *criteria.MaxPrice {
				continue
			}
		}
		available = append(available, room)
	}
	return available, nil
}
