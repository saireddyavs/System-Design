package tests

import (
	"testing"

	"ride-sharing-service/internal/models"
	"ride-sharing-service/internal/strategies"
)

func TestNearestDriverStrategy_FindsNearest(t *testing.T) {
	strategy := strategies.NewNearestDriverStrategy(50)

	riderLoc := models.Location{Latitude: 37.7750, Longitude: -122.4195}
	drivers := []*models.Driver{
		models.NewDriver("D1", "Far", "1", models.Vehicle{}),
		models.NewDriver("D2", "Near", "2", models.Vehicle{}),
		models.NewDriver("D3", "Mid", "3", models.Vehicle{}),
	}
	drivers[0].UpdateLocation(models.Location{Latitude: 37.8000, Longitude: -122.4000})
	drivers[1].UpdateLocation(models.Location{Latitude: 37.7751, Longitude: -122.4196}) // Closest
	drivers[2].UpdateLocation(models.Location{Latitude: 37.7800, Longitude: -122.4150})

	for _, d := range drivers {
		d.SetStatus(models.DriverStatusAvailable)
	}

	driver, err := strategy.FindDriver(riderLoc, drivers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if driver.ID != "D2" {
		t.Errorf("expected nearest driver D2, got %s", driver.ID)
	}
}

func TestNearestDriverStrategy_NoDriverAvailable(t *testing.T) {
	strategy := strategies.NewNearestDriverStrategy(50)
	riderLoc := models.Location{Latitude: 37.7750, Longitude: -122.4195}

	_, err := strategy.FindDriver(riderLoc, nil)
	if err != strategies.ErrNoDriverAvailable {
		t.Errorf("expected ErrNoDriverAvailable, got %v", err)
	}

	_, err = strategy.FindDriver(riderLoc, []*models.Driver{})
	if err != strategies.ErrNoDriverAvailable {
		t.Errorf("expected ErrNoDriverAvailable for empty list, got %v", err)
	}
}

func TestNearestDriverStrategy_ExcludesLowRatedDrivers(t *testing.T) {
	strategy := strategies.NewNearestDriverStrategy(50)
	riderLoc := models.Location{Latitude: 37.7750, Longitude: -122.4195}

	lowRated := models.NewDriver("D1", "Low", "1", models.Vehicle{})
	lowRated.UpdateLocation(riderLoc)
	lowRated.SetStatus(models.DriverStatusAvailable)
	lowRated.Rating = 2.5

	goodDriver := models.NewDriver("D2", "Good", "2", models.Vehicle{})
	goodDriver.UpdateLocation(models.Location{Latitude: 37.7751, Longitude: -122.4196})
	goodDriver.SetStatus(models.DriverStatusAvailable)

	drivers := []*models.Driver{lowRated, goodDriver}
	driver, err := strategy.FindDriver(riderLoc, drivers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if driver.ID != "D2" {
		t.Errorf("should skip low-rated driver, got %s", driver.ID)
	}
}

func TestHighestRatedStrategy_PrefersHigherRating(t *testing.T) {
	strategy := strategies.NewHighestRatedStrategy(50)
	riderLoc := models.Location{Latitude: 37.7750, Longitude: -122.4195}

	d1 := models.NewDriver("D1", "4.5", "1", models.Vehicle{})
	d1.UpdateLocation(riderLoc)
	d1.SetStatus(models.DriverStatusAvailable)
	d1.Rating = 4.5

	d2 := models.NewDriver("D2", "5.0", "2", models.Vehicle{})
	d2.UpdateLocation(models.Location{Latitude: 37.7752, Longitude: -122.4197})
	d2.SetStatus(models.DriverStatusAvailable)
	d2.Rating = 5.0

	drivers := []*models.Driver{d1, d2}
	driver, err := strategy.FindDriver(riderLoc, drivers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if driver.ID != "D2" {
		t.Errorf("expected highest rated D2, got %s", driver.ID)
	}
}
