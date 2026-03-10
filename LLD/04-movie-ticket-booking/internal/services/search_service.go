package services

import (
	"movie-ticket-booking/internal/interfaces"
	"movie-ticket-booking/internal/models"
)

// SearchCriteria represents search filters
type SearchCriteria struct {
	Title    string
	Genre    models.Genre
	City     string
	Language string
}

// SearchResult combines movie with available shows in a city
type SearchResult struct {
	Movie   *models.Movie
	Shows   []*models.Show
	Theatre *models.Theatre
}

// SearchService handles cross-entity search (Facade pattern)
type SearchService struct {
	movieRepo   interfaces.MovieRepository
	theatreRepo interfaces.TheatreRepository
	showRepo    interfaces.ShowRepository
}

// NewSearchService creates a new search service
func NewSearchService(
	movieRepo interfaces.MovieRepository,
	theatreRepo interfaces.TheatreRepository,
	showRepo interfaces.ShowRepository,
) *SearchService {
	return &SearchService{
		movieRepo:   movieRepo,
		theatreRepo: theatreRepo,
		showRepo:    showRepo,
	}
}

// SearchMoviesByTitle finds movies matching title
func (s *SearchService) SearchMoviesByTitle(title string) ([]*models.Movie, error) {
	return s.movieRepo.SearchByTitle(title)
}

// SearchMoviesByGenre finds movies by genre
func (s *SearchService) SearchMoviesByGenre(genre models.Genre) ([]*models.Movie, error) {
	return s.movieRepo.SearchByGenre(genre)
}

// SearchByCity finds movies with shows in a city
func (s *SearchService) SearchByCity(city string) ([]*SearchResult, error) {
	theatres, err := s.theatreRepo.GetByCity(city)
	if err != nil {
		return nil, err
	}

	moviesSeen := make(map[string]bool)
	var results []*SearchResult

	for _, theatre := range theatres {
		shows, err := s.showRepo.GetByTheatreID(theatre.ID)
		if err != nil {
			continue
		}
		for _, show := range shows {
			if moviesSeen[show.MovieID] {
				continue
			}
			movie, err := s.movieRepo.GetByID(show.MovieID)
			if err != nil {
				continue
			}
			moviesSeen[show.MovieID] = true
			allShows, _ := s.showRepo.GetByTheatreID(theatre.ID)
			var movieShows []*models.Show
			for _, sh := range allShows {
				if sh.MovieID == show.MovieID {
					movieShows = append(movieShows, sh)
				}
			}
			results = append(results, &SearchResult{
				Movie:   movie,
				Shows:   movieShows,
				Theatre: theatre,
			})
		}
	}
	return results, nil
}

// Search combines title, genre, and city filters
func (s *SearchService) Search(criteria *SearchCriteria) ([]*SearchResult, error) {
	var movies []*models.Movie
	var err error

	if criteria.Title != "" {
		movies, err = s.movieRepo.SearchByTitle(criteria.Title)
	} else if criteria.Genre != "" {
		movies, err = s.movieRepo.SearchByGenre(criteria.Genre)
	} else if criteria.Language != "" {
		movies, err = s.movieRepo.SearchByLanguage(criteria.Language)
	} else {
		movies, err = s.movieRepo.GetAll()
	}
	if err != nil {
		return nil, err
	}

	if criteria.City == "" {
		var results []*SearchResult
		for _, m := range movies {
			results = append(results, &SearchResult{Movie: m})
		}
		return results, nil
	}

	return s.filterByCity(movies, criteria.City)
}

func (s *SearchService) filterByCity(movies []*models.Movie, city string) ([]*SearchResult, error) {
	theatres, err := s.theatreRepo.GetByCity(city)
	if err != nil {
		return nil, err
	}

	theatreIDs := make(map[string]bool)
	for _, t := range theatres {
		theatreIDs[t.ID] = true
	}

	var results []*SearchResult
	for _, movie := range movies {
		shows, _ := s.showRepo.GetByMovieID(movie.ID)
		for _, show := range shows {
			if theatreIDs[show.TheatreID] {
				theatre, _ := s.theatreRepo.GetByID(show.TheatreID)
				theatreShows, _ := s.showRepo.GetByTheatreID(show.TheatreID)
				var movieShows []*models.Show
				for _, sh := range theatreShows {
					if sh.MovieID == movie.ID {
						movieShows = append(movieShows, sh)
					}
				}
				results = append(results, &SearchResult{
					Movie:   movie,
					Shows:   movieShows,
					Theatre: theatre,
				})
				break
			}
		}
	}
	return results, nil
}
