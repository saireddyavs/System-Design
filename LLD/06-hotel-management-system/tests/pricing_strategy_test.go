package tests

import (
	"testing"
	"time"

	"hotel-management-system/internal/interfaces"
	"hotel-management-system/internal/models"
	"hotel-management-system/internal/strategies"
)

func TestPricingStrategy_BasePrice(t *testing.T) {
	base := &strategies.BasePricingStrategy{}
	room := models.NewRoom("R1", "101", models.RoomTypeSingle, 1, 100.0, nil)
	ctx := &interfaces.PricingContext{
		Room:         room,
		Nights:       3,
		CheckInDate:  time.Now(),
		CheckOutDate: time.Now().AddDate(0, 0, 3),
	}
	price := base.CalculatePrice(ctx)
	if price != 300.0 {
		t.Errorf("expected 300, got %.2f", price)
	}
}

func TestPricingStrategy_LoyaltyDiscount(t *testing.T) {
	composite := strategies.NewCompositePricingStrategy()
	room := models.NewRoom("R1", "101", models.RoomTypeDouble, 1, 150.0, nil)
	guest := models.NewGuest("G1", "John", "j@x.com", "1", "P1")
	guest.AddLoyaltyPoints(6000) // Gold tier = 10% discount

	ctx := &interfaces.PricingContext{
		Room:         room,
		Guest:        guest,
		Nights:       2,
		CheckInDate:  time.Now(),
		CheckOutDate: time.Now().AddDate(0, 0, 2),
	}
	price := composite.CalculatePrice(ctx)
	baseExpected := 150.0 * 2
	discountedExpected := baseExpected * 0.9
	if price < discountedExpected*0.95 || price > discountedExpected*1.05 {
		t.Errorf("expected ~%.2f with loyalty discount, got %.2f", discountedExpected, price)
	}
}
