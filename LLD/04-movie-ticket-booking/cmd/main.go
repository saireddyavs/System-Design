package main

import (
	"fmt"
	"log"
	"movie-ticket-booking/internal/interfaces"
	"movie-ticket-booking/internal/models"
	"movie-ticket-booking/internal/repositories"
	"movie-ticket-booking/internal/services"
	"movie-ticket-booking/internal/strategies"
	"time"
)

func main() {
	// Dependency Injection - all dependencies wired explicitly (DIP)
	movieRepo := repositories.NewInMemoryMovieRepository()
	theatreRepo := repositories.NewInMemoryTheatreRepository()
	screenRepo := repositories.NewInMemoryScreenRepository()
	showRepo := repositories.NewInMemoryShowRepository()
	bookingRepo := repositories.NewInMemoryBookingRepository()
	userRepo := repositories.NewInMemoryUserRepository()

	payment := services.NewMockPaymentProcessor()
	notification := services.NewObserverNotificationService()

	// Subscribe to notifications (Observer pattern)
	notification.Subscribe(func(p *interfaces.NotificationPayload) {
		log.Printf("[Notification] %s: %s", p.Event, p.Message)
	})

	// Strategy pattern: Choose pricing strategy
	pricingStrategy := strategies.NewWeekdayPricingStrategy()

	_ = services.NewMovieService(movieRepo)
	_ = services.NewShowService(showRepo, screenRepo, movieRepo)
	bookingService := services.NewBookingService(
		bookingRepo, showRepo, screenRepo, userRepo,
		payment, notification, pricingStrategy,
	)
	searchService := services.NewSearchService(movieRepo, theatreRepo, showRepo)

	// Seed data
	seedData(movieRepo, theatreRepo, screenRepo, showRepo, userRepo)

	// Demo: Search movies
	fmt.Println("=== Search: Action movies in Bangalore ===")
	results, _ := searchService.Search(&services.SearchCriteria{Genre: models.GenreAction, City: "Bangalore"})
	for _, r := range results {
		fmt.Printf("  %s - %d shows at %s\n", r.Movie.Title, len(r.Shows), r.Theatre.Name)
	}

	// Demo: Create booking
	fmt.Println("\n=== Create Booking ===")
	booking, err := bookingService.CreateBooking(&services.CreateBookingRequest{
		UserID:  "user-1",
		ShowID:  "show-1",
		SeatIDs: []string{"seat-1", "seat-2"},
	})
	if err != nil {
		log.Fatalf("Booking failed: %v", err)
	}
	fmt.Printf("  Booking ID: %s, Amount: %.2f\n", booking.ID, booking.TotalAmount)

	// Demo: Try double booking same seat (should fail)
	fmt.Println("\n=== Double Booking Attempt (should fail) ===")
	_, err = bookingService.CreateBooking(&services.CreateBookingRequest{
		UserID:  "user-2",
		ShowID:  "show-1",
		SeatIDs: []string{"seat-1"},
	})
	if err != nil {
		fmt.Printf("  Correctly rejected: %v\n", err)
	}

	// Demo: Cancel booking
	fmt.Println("\n=== Cancel Booking ===")
	cancelled, _ := bookingService.CancelBooking(booking.ID)
	fmt.Printf("  Cancelled. Refund: %.2f\n", cancelled.RefundAmount)

	fmt.Println("\nDone.")
}

func seedData(
	movieRepo interfaces.MovieRepository,
	theatreRepo interfaces.TheatreRepository,
	screenRepo interfaces.ScreenRepository,
	showRepo interfaces.ShowRepository,
	userRepo interfaces.UserRepository,
) {
	// Users
	userRepo.Create(&models.User{ID: "user-1", Name: "Alice", Email: "alice@test.com", Phone: "1234567890"})
	userRepo.Create(&models.User{ID: "user-2", Name: "Bob", Email: "bob@test.com", Phone: "0987654321"})

	// Movies
	movieRepo.Create(&models.Movie{ID: "movie-1", Title: "Inception", Genre: models.GenreSciFi, Duration: 148, Rating: "U/A", Language: "English"})
	movieRepo.Create(&models.Movie{ID: "movie-2", Title: "The Dark Knight", Genre: models.GenreAction, Duration: 152, Rating: "U/A", Language: "English"})

	// Theatre & Screen
	theatreRepo.Create(&models.Theatre{ID: "th-1", Name: "PVR Forum", City: "Bangalore", Address: "Koramangala"})
	screenRepo.Create(&models.Screen{
		ID: "screen-1", TheatreID: "th-1", Name: "Screen 1", TotalCapacity: 4,
		Seats: []models.Seat{
			{ID: "seat-1", ScreenID: "screen-1", Row: "A", Number: 1, Category: models.SeatCategoryRegular},
			{ID: "seat-2", ScreenID: "screen-1", Row: "A", Number: 2, Category: models.SeatCategoryRegular},
			{ID: "seat-3", ScreenID: "screen-1", Row: "B", Number: 1, Category: models.SeatCategoryPremium},
			{ID: "seat-4", ScreenID: "screen-1", Row: "B", Number: 2, Category: models.SeatCategoryVIP},
		},
	})

	// Show
	start := time.Now().Add(48 * time.Hour) // 2 days from now
	show := services.NewShowBuilder().
		SetID("show-1").
		SetMovieID("movie-2").
		SetScreenID("screen-1").
		SetTheatreID("th-1").
		SetStartTime(start).
		SetDuration(152).
		SetBasePrice(200).
		Build(&models.Screen{
			Seats: []models.Seat{
				{ID: "seat-1"}, {ID: "seat-2"}, {ID: "seat-3"}, {ID: "seat-4"},
			},
		})
	showRepo.Create(show)
}
