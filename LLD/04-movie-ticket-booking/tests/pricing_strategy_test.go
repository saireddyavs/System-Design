package tests

import (
	"movie-ticket-booking/internal/models"
	"movie-ticket-booking/internal/strategies"
	"testing"
	"time"
)

func TestWeekdayPricingStrategy_RegularSeat(t *testing.T) {
	ps := strategies.NewWeekdayPricingStrategy()

	// Tuesday (weekday)
	tuesday := time.Date(2025, 3, 11, 14, 0, 0, 0, time.UTC)
	price := ps.CalculatePrice(100, models.SeatCategoryRegular, tuesday)
	if price != 100 {
		t.Errorf("expected 100 for weekday regular, got %.2f", price)
	}
}

func TestWeekdayPricingStrategy_WeekendPremium(t *testing.T) {
	ps := strategies.NewWeekdayPricingStrategy()

	// Saturday
	saturday := time.Date(2025, 3, 15, 14, 0, 0, 0, time.UTC)
	price := ps.CalculatePrice(100, models.SeatCategoryPremium, saturday)
	expected := 100 * 1.25 * 1.5 // weekend * premium
	if price != expected {
		t.Errorf("expected %.2f for weekend premium, got %.2f", expected, price)
	}
}

func TestWeekdayPricingStrategy_VIP(t *testing.T) {
	ps := strategies.NewWeekdayPricingStrategy()

	price := ps.CalculatePrice(100, models.SeatCategoryVIP, time.Now())
	if price != 200 {
		t.Errorf("expected 200 for VIP, got %.2f", price)
	}
}
