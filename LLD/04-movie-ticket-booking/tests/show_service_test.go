package tests

import (
	"movie-ticket-booking/internal/models"
	"movie-ticket-booking/internal/repositories"
	"movie-ticket-booking/internal/services"
	"testing"
	"time"
)

func TestShowBuilder_Build(t *testing.T) {
	screenRepo := repositories.NewInMemoryScreenRepository()
	movieRepo := repositories.NewInMemoryMovieRepository()
	showRepo := repositories.NewInMemoryShowRepository()

	screen := &models.Screen{
		ID: "s1", TheatreID: "t1", Name: "Screen 1", TotalCapacity: 2,
		Seats: []models.Seat{
			{ID: "seat-1", ScreenID: "s1", Row: "A", Number: 1, Category: models.SeatCategoryRegular},
			{ID: "seat-2", ScreenID: "s1", Row: "A", Number: 2, Category: models.SeatCategoryPremium},
		},
	}
	screenRepo.Create(screen)

	svc := services.NewShowService(showRepo, screenRepo, movieRepo)

	start := time.Now().Add(24 * time.Hour)
	builder := services.NewShowBuilder().
		SetID("show-1").
		SetMovieID("m1").
		SetScreenID("s1").
		SetTheatreID("t1").
		SetStartTime(start).
		SetDuration(120).
		SetBasePrice(150)

	show, err := svc.CreateShow(builder)
	if err != nil {
		t.Fatalf("CreateShow failed: %v", err)
	}
	if show.ID != "show-1" {
		t.Errorf("expected show ID show-1, got %s", show.ID)
	}
	if show.BasePrice != 150 {
		t.Errorf("expected base price 150, got %.2f", show.BasePrice)
	}
	if len(show.SeatStatusMap) != 2 {
		t.Errorf("expected 2 seats in status map, got %d", len(show.SeatStatusMap))
	}
	for _, status := range show.SeatStatusMap {
		if status != models.SeatStatusAvailable {
			t.Errorf("expected all seats available, got %s", status)
		}
	}
}

func TestGetShow(t *testing.T) {
	showRepo := repositories.NewInMemoryShowRepository()
	screenRepo := repositories.NewInMemoryScreenRepository()
	movieRepo := repositories.NewInMemoryMovieRepository()

	show := &models.Show{
		ID: "show-1", MovieID: "m1", ScreenID: "s1", TheatreID: "t1",
		StartTime: time.Now(), EndTime: time.Now().Add(2 * time.Hour),
		SeatStatusMap: map[string]models.SeatStatus{},
		BasePrice:     100,
	}
	showRepo.Create(show)

	svc := services.NewShowService(showRepo, screenRepo, movieRepo)

	got, err := svc.GetShow("show-1")
	if err != nil {
		t.Fatalf("GetShow failed: %v", err)
	}
	if got.ID != "show-1" {
		t.Errorf("expected show-1, got %s", got.ID)
	}
}
