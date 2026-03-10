package services

import (
	"movie-ticket-booking/internal/interfaces"
	"movie-ticket-booking/internal/models"
	"time"
)

// ShowService handles show scheduling (Builder pattern for show creation)
type ShowService struct {
	showRepo   interfaces.ShowRepository
	screenRepo interfaces.ScreenRepository
	movieRepo  interfaces.MovieRepository
}

// NewShowService creates a new show service
func NewShowService(
	showRepo interfaces.ShowRepository,
	screenRepo interfaces.ScreenRepository,
	movieRepo interfaces.MovieRepository,
) *ShowService {
	return &ShowService{
		showRepo:   showRepo,
		screenRepo: screenRepo,
		movieRepo:  movieRepo,
	}
}

// ShowBuilder builds a show (Builder Pattern)
type ShowBuilder struct {
	show *models.Show
}

// NewShowBuilder creates a new show builder
func NewShowBuilder() *ShowBuilder {
	return &ShowBuilder{
		show: &models.Show{
			SeatStatusMap: make(map[string]models.SeatStatus),
		},
	}
}

// SetID sets show ID
func (b *ShowBuilder) SetID(id string) *ShowBuilder {
	b.show.ID = id
	return b
}

// SetMovieID sets movie ID
func (b *ShowBuilder) SetMovieID(movieID string) *ShowBuilder {
	b.show.MovieID = movieID
	return b
}

// SetScreenID sets screen ID
func (b *ShowBuilder) SetScreenID(screenID string) *ShowBuilder {
	b.show.ScreenID = screenID
	return b
}

// SetTheatreID sets theatre ID
func (b *ShowBuilder) SetTheatreID(theatreID string) *ShowBuilder {
	b.show.TheatreID = theatreID
	return b
}

// SetStartTime sets start time
func (b *ShowBuilder) SetStartTime(t time.Time) *ShowBuilder {
	b.show.StartTime = t
	return b
}

// SetDuration sets end time based on movie duration
func (b *ShowBuilder) SetDuration(minutes int) *ShowBuilder {
	b.show.EndTime = b.show.StartTime.Add(time.Duration(minutes) * time.Minute)
	return b
}

// SetBasePrice sets base price
func (b *ShowBuilder) SetBasePrice(price float64) *ShowBuilder {
	b.show.BasePrice = price
	return b
}

// Build initializes seat status map from screen and returns the show
func (b *ShowBuilder) Build(screen *models.Screen) *models.Show {
	for _, seat := range screen.Seats {
		b.show.SeatStatusMap[seat.ID] = models.SeatStatusAvailable
	}
	return b.show
}

// CreateShow creates a show using builder
func (s *ShowService) CreateShow(builder *ShowBuilder) (*models.Show, error) {
	screen, err := s.screenRepo.GetByID(builder.show.ScreenID)
	if err != nil {
		return nil, err
	}
	show := builder.Build(screen)
	if err := s.showRepo.Create(show); err != nil {
		return nil, err
	}
	return show, nil
}

// GetShow retrieves show by ID
func (s *ShowService) GetShow(id string) (*models.Show, error) {
	return s.showRepo.GetByID(id)
}

// GetShowsByMovie retrieves shows for a movie
func (s *ShowService) GetShowsByMovie(movieID string) ([]*models.Show, error) {
	return s.showRepo.GetByMovieID(movieID)
}

// GetShowsByTheatre retrieves shows for a theatre
func (s *ShowService) GetShowsByTheatre(theatreID string) ([]*models.Show, error) {
	return s.showRepo.GetByTheatreID(theatreID)
}
