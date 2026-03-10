package tests

import (
	"testing"
	"time"

	"ride-sharing-service/internal/models"
	"ride-sharing-service/internal/strategies"
)

func TestBaseFareStrategy_Calculate(t *testing.T) {
	calc := strategies.NewBaseFareStrategy(2.0, 1.5, 0.25) // $2 base, $1.5/km, $0.25/min

	pickup := models.Location{Latitude: 37.7749, Longitude: -122.4194}
	dropoff := models.Location{Latitude: 37.7849, Longitude: -122.4094}
	duration := 15 * time.Minute

	fare := calc.Calculate(pickup, dropoff, duration, 1.0)
	distance := models.HaversineDistance(pickup, dropoff)

	expected := 2.0 + (distance*1.5) + (15*0.25)
	if fare < expected-0.01 || fare > expected+0.01 {
		t.Errorf("expected fare ~%.2f, got %.2f", expected, fare)
	}
}

func TestBaseFareStrategy_WithSurgeMultiplier(t *testing.T) {
	calc := strategies.NewBaseFareStrategy(2.0, 1.0, 0.2)
	pickup := models.Location{Latitude: 0, Longitude: 0}
	dropoff := models.Location{Latitude: 0.01, Longitude: 0.01}
	duration := 10 * time.Minute

	baseFare := calc.Calculate(pickup, dropoff, duration, 1.0)
	surgeFare := calc.Calculate(pickup, dropoff, duration, 1.5)

	if surgeFare <= baseFare {
		t.Errorf("surge fare (%.2f) should be > base fare (%.2f)", surgeFare, baseFare)
	}
	expectedSurge := baseFare * 1.5
	if surgeFare < expectedSurge-0.01 || surgeFare > expectedSurge+0.01 {
		t.Errorf("expected surge fare ~%.2f, got %.2f", expectedSurge, surgeFare)
	}
}

func TestHaversineDistance(t *testing.T) {
	// Same point = 0 distance
	loc := models.Location{Latitude: 37.7749, Longitude: -122.4194}
	dist := models.HaversineDistance(loc, loc)
	if dist != 0 {
		t.Errorf("distance to same point should be 0, got %.4f", dist)
	}

	// San Francisco to Oakland ~13km
	sf := models.Location{Latitude: 37.7749, Longitude: -122.4194}
	oakland := models.Location{Latitude: 37.8044, Longitude: -122.2712}
	dist = models.HaversineDistance(sf, oakland)
	if dist < 10 || dist > 20 {
		t.Errorf("SF to Oakland distance should be ~13km, got %.2f", dist)
	}
}
