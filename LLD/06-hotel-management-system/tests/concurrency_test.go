package tests

import (
	"sync"
	"testing"
	"time"

	"hotel-management-system/internal/models"
	"hotel-management-system/internal/services"
)

func TestConcurrentBookingCreation(t *testing.T) {
	bookingSvc, _, _ := setupBookingTest(t)

	checkIn := time.Now().AddDate(0, 0, 2)
	checkOut := checkIn.AddDate(0, 0, 3)

	var wg sync.WaitGroup
	results := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := bookingSvc.CreateBooking("G1", "R1", checkIn, checkOut)
			results <- err
		}()
	}

	wg.Wait()
	close(results)

	successCount := 0
	for err := range results {
		if err == nil {
			successCount++
		}
	}

	// Only one should succeed due to overbooking prevention
	if successCount != 1 {
		t.Errorf("expected 1 successful booking, got %d", successCount)
	}
}

func TestConcurrentRoomSearch(t *testing.T) {
	roomSvc := setupRoomTestSimple(t)
	roomSvc.CreateRoom("R1", "101", models.RoomTypeSingle, 1)
	roomSvc.CreateRoom("R2", "102", models.RoomTypeDouble, 1)

	checkIn := time.Now().AddDate(0, 0, 2)
	checkOut := checkIn.AddDate(0, 0, 3)
	criteria := services.SearchCriteria{CheckIn: checkIn, CheckOut: checkOut}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = roomSvc.GetAvailableRooms(criteria)
		}()
	}
	wg.Wait()
	// No panic = thread-safe
}
