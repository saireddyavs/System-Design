package tests

import (
	"testing"
	"time"

	"hotel-management-system/internal/factory"
	"hotel-management-system/internal/models"
	"hotel-management-system/internal/repositories"
	"hotel-management-system/internal/services"
)

func setupRoomTestSimple(t *testing.T) *services.RoomService {
	roomRepo := repositories.NewInMemoryRoomRepository()
	bookingRepo := repositories.NewInMemoryBookingRepository()
	roomFactory := factory.NewRoomFactory()
	return services.NewRoomService(roomRepo, bookingRepo, roomFactory)
}

func TestRoomService_SearchAvailableRooms(t *testing.T) {
	roomSvc := setupRoomTestSimple(t)

	roomSvc.CreateRoom("R1", "101", models.RoomTypeSingle, 1)
	roomSvc.CreateRoom("R2", "102", models.RoomTypeDouble, 1)
	roomSvc.CreateRoom("R3", "103", models.RoomTypeDouble, 1)

	checkIn := time.Now().AddDate(0, 0, 2)
	checkOut := checkIn.AddDate(0, 0, 3)

	rooms, err := roomSvc.GetAvailableRooms(services.SearchCriteria{
		CheckIn:  checkIn,
		CheckOut: checkOut,
	})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(rooms) != 3 {
		t.Errorf("expected 3 rooms, got %d", len(rooms))
	}

	// Filter by type
	doubleType := models.RoomTypeDouble
	rooms, err = roomSvc.GetAvailableRooms(services.SearchCriteria{
		CheckIn:   checkIn,
		CheckOut:  checkOut,
		RoomType:  &doubleType,
	})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(rooms) != 2 {
		t.Errorf("expected 2 double rooms, got %d", len(rooms))
	}
}

func TestRoomService_InvalidDateRange(t *testing.T) {
	roomSvc := setupRoomTestSimple(t)

	checkIn := time.Now().AddDate(0, 0, 5)
	checkOut := time.Now().AddDate(0, 0, 2) // Before check-in

	_, err := roomSvc.GetAvailableRooms(services.SearchCriteria{
		CheckIn:  checkIn,
		CheckOut: checkOut,
	})
	if err != services.ErrInvalidDateRange {
		t.Errorf("expected ErrInvalidDateRange, got %v", err)
	}
}

func TestRoomService_PriceRangeFilter(t *testing.T) {
	roomSvc := setupRoomTestSimple(t)

	roomSvc.CreateRoom("R1", "101", models.RoomTypeSingle, 1)
	roomSvc.CreateRoom("R2", "102", models.RoomTypeSuite, 1)

	checkIn := time.Now().AddDate(0, 0, 2)
	checkOut := checkIn.AddDate(0, 0, 3)

	minPrice := 100.0
	maxPrice := 200.0
	rooms, err := roomSvc.GetAvailableRooms(services.SearchCriteria{
		CheckIn:  checkIn,
		CheckOut: checkOut,
		MinPrice: &minPrice,
		MaxPrice: &maxPrice,
	})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	// Single is 100, Suite is 400 - only Single should match
	if len(rooms) != 1 {
		t.Errorf("expected 1 room in price range, got %d", len(rooms))
	}
	if len(rooms) > 0 && rooms[0].Type != models.RoomTypeSingle {
		t.Errorf("expected Single room, got %s", rooms[0].Type)
	}
}
