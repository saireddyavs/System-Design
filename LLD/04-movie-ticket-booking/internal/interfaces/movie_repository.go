package interfaces

import "movie-ticket-booking/internal/models"

// MovieRepository defines operations for movie data access (Repository Pattern - DIP)
type MovieRepository interface {
	Create(movie *models.Movie) error
	GetByID(id string) (*models.Movie, error)
	GetAll() ([]*models.Movie, error)
	SearchByTitle(title string) ([]*models.Movie, error)
	SearchByGenre(genre models.Genre) ([]*models.Movie, error)
	SearchByLanguage(language string) ([]*models.Movie, error)
}

// TheatreRepository defines operations for theatre data access
type TheatreRepository interface {
	Create(theatre *models.Theatre) error
	GetByID(id string) (*models.Theatre, error)
	GetByCity(city string) ([]*models.Theatre, error)
}

// ScreenRepository defines operations for screen data access
type ScreenRepository interface {
	Create(screen *models.Screen) error
	GetByID(id string) (*models.Screen, error)
}

// ShowRepository defines operations for show data access
type ShowRepository interface {
	Create(show *models.Show) error
	GetByID(id string) (*models.Show, error)
	GetByMovieID(movieID string) ([]*models.Show, error)
	GetByTheatreID(theatreID string) ([]*models.Show, error)
	// UpdateSeats atomically updates show seats with pessimistic locking (prevents double booking)
	UpdateSeats(showID string, updateFn func(*models.Show) error) error
}
