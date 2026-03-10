package main

import (
	"fmt"
	"log"
	"time"

	"hotel-management-system/internal/factory"
	"hotel-management-system/internal/interfaces"
	"hotel-management-system/internal/models"
	"hotel-management-system/internal/observer"
	"hotel-management-system/internal/payment"
	"hotel-management-system/internal/repositories"
	"hotel-management-system/internal/services"
	"hotel-management-system/internal/strategies"
)

func main() {
	// Initialize repositories
	roomRepo := repositories.NewInMemoryRoomRepository()
	bookingRepo := repositories.NewInMemoryBookingRepository()
	guestRepo := repositories.NewInMemoryGuestRepository()
	paymentRepo := repositories.NewInMemoryPaymentRepository()

	// Initialize factory and strategies
	roomFactory := factory.NewRoomFactory()
	pricingStrategy := strategies.NewCompositePricingStrategy()
	notificationSvc := observer.NewNotificationService()
	paymentProcessor := payment.NewMockPaymentProcessor()

	// Subscribe to notifications (Observer pattern)
	notificationSvc.Subscribe(func(p interfaces.NotificationPayload) {
		guestName := ""
		if p.Guest != nil {
			guestName = p.Guest.Name
		}
		log.Printf("[NOTIFICATION] Event: %s, Booking: %s, Guest: %s",
			p.Event, p.Booking.ID, guestName)
	})

	// Initialize services
	roomSvc := services.NewRoomService(roomRepo, bookingRepo, roomFactory)
	guestSvc := services.NewGuestService(guestRepo)
	paymentSvc := services.NewPaymentService(paymentRepo, paymentProcessor)
	bookingSvc := services.NewBookingService(
		bookingRepo, roomRepo, guestRepo, paymentRepo,
		paymentSvc, pricingStrategy, notificationSvc,
	)

	// Seed data: Create rooms
	for i := 1; i <= 3; i++ {
		roomType := []models.RoomType{
			models.RoomTypeSingle,
			models.RoomTypeDouble,
			models.RoomTypeDeluxe,
		}[i-1]
		_, err := roomSvc.CreateRoom(
			fmt.Sprintf("R%d", i),
			fmt.Sprintf("10%d", i),
			roomType,
			1,
		)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Register guest
	guest, err := guestSvc.RegisterGuest("G1", "John Doe", "john@example.com", "+1234567890", "PASS123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Registered guest: %s\n", guest.Name)

	// Search available rooms
	checkIn := time.Now().AddDate(0, 0, 2)
	checkOut := checkIn.AddDate(0, 0, 3)
	rooms, err := roomSvc.GetAvailableRooms(services.SearchCriteria{
		CheckIn:  checkIn,
		CheckOut:  checkOut,
		RoomType:  ptr(models.RoomTypeDouble),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d available rooms\n", len(rooms))

	if len(rooms) > 0 {
		// Create booking
		booking, err := bookingSvc.CreateBooking(guest.ID, rooms[0].ID, checkIn, checkOut)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Created booking %s, Total: $%.2f\n", booking.ID, booking.GetTotalAmount())

		// Confirm with payment
		pmt := models.NewPayment("P1", booking.ID, booking.GetTotalAmount(), models.PaymentMethodCard)
		if err := bookingSvc.ConfirmBooking(booking.ID, pmt); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Booking confirmed")

		// Simulate check-in (would be on check-in date)
		// For demo, we'll just show the flow - in real scenario check-in happens on the date
		fmt.Println("Booking flow: Pending -> Confirmed (payment) -> CheckIn -> CheckOut")
	}

	fmt.Println("\nHotel Management System demo completed successfully.")
}

func ptr[T any](v T) *T { return &v }
