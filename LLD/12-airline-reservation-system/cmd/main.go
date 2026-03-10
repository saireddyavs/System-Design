package main

import (
	"airline-reservation-system/internal/models"
	"airline-reservation-system/internal/repositories"
	"airline-reservation-system/internal/services"
	"airline-reservation-system/internal/strategies"
	"fmt"
	"log"
	"time"
)

func main() {
	// Initialize repositories
	flightRepo := repositories.NewInMemoryFlightRepository()
	bookingRepo := repositories.NewInMemoryBookingRepository()
	passengerRepo := repositories.NewInMemoryPassengerRepository()

	// Initialize strategies
	seatStrategy := strategies.NewAutoAssignFirstAvailable()
	pricingStrategy := strategies.NewClassMultiplierPricing()
	paymentProcessor := strategies.NewMockPaymentProcessor()

	// Initialize services
	flightService := services.NewFlightService(flightRepo, bookingRepo)
	seatService := services.NewSeatService(flightRepo, seatStrategy)
	searchService := services.NewSearchService(flightRepo)
	bookingFactory := services.NewBookingFactory()
	notifier := services.NewBookingNotifier()

	// Subscribe observers
	notifier.Subscribe(services.NewEmailBookingObserver())

	bookingService := services.NewBookingService(
		bookingRepo, flightRepo, passengerRepo,
		seatService, pricingStrategy, paymentProcessor,
		bookingFactory, notifier,
	)

	// Create a flight using Builder pattern
	departure := time.Now().Add(72 * time.Hour)
	arrival := departure.Add(3 * time.Hour)
	flight, err := models.NewFlightBuilder().
		ID("FL-001").
		FlightNumber("AA100").
		Route("JFK", "LAX").
		Schedule(departure, arrival).
		Aircraft("Boeing 737").
		BasePrice(150.0).
		AddSeatSection(5, []string{"A", "B", "C"}, models.SeatClassEconomy).
		AddSeatSection(2, []string{"A", "B"}, models.SeatClassBusiness).
		Build()
	if err != nil {
		log.Fatal(err)
	}

	if err := flightService.AddFlight(flight); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Flight AA100 added successfully")

	// Create passenger
	passenger := models.NewPassenger("P1", "John Doe", "john@example.com", "+1234567890", "AB123456", time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC))
	if err := passengerRepo.Create(passenger); err != nil {
		log.Fatal(err)
	}

	// Search flights
	results, err := searchService.SearchFlights(services.SearchCriteria{
		Origin:      "JFK",
		Destination: "LAX",
		Date:        departure,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d flight(s)\n", len(results))
	for _, r := range results {
		fmt.Printf("  %s: %d seats, from $%.2f\n", r.Flight.FlightNumber, r.AvailableSeats, r.MinPrice)
	}

	// Create booking
	booking, err := bookingService.CreateBooking("P1", "FL-001", 2, models.SeatClassEconomy)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Booking created: %s, Ref: %s, Amount: $%.2f\n", booking.ID, booking.BookingRef, booking.TotalAmount)

	// Get booking
	retrieved, err := bookingService.GetBooking(booking.BookingRef)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Retrieved booking: %s\n", retrieved.ID)

	// Cancel booking (full refund - >48h)
	refund, err := bookingService.CancelBooking(booking.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Booking cancelled. Refund: $%.2f\n", refund)

	fmt.Println("\nAirline Reservation System demo completed successfully.")
}
