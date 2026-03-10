package services

import (
	"movie-ticket-booking/internal/interfaces"
	"movie-ticket-booking/internal/models"
)

// MovieService handles movie operations (Single Responsibility)
type MovieService struct {
	movieRepo interfaces.MovieRepository
}

// NewMovieService creates a new movie service
func NewMovieService(movieRepo interfaces.MovieRepository) *MovieService {
	return &MovieService{movieRepo: movieRepo}
}

// CreateMovie adds a new movie
func (s *MovieService) CreateMovie(movie *models.Movie) error {
	return s.movieRepo.Create(movie)
}

// GetMovie retrieves movie by ID
func (s *MovieService) GetMovie(id string) (*models.Movie, error) {
	return s.movieRepo.GetByID(id)
}

// SearchByTitle searches movies by title
func (s *MovieService) SearchByTitle(title string) ([]*models.Movie, error) {
	return s.movieRepo.SearchByTitle(title)
}

// SearchByGenre searches movies by genre
func (s *MovieService) SearchByGenre(genre models.Genre) ([]*models.Movie, error) {
	return s.movieRepo.SearchByGenre(genre)
}

// SearchByLanguage searches movies by language
func (s *MovieService) SearchByLanguage(language string) ([]*models.Movie, error) {
	return s.movieRepo.SearchByLanguage(language)
}
