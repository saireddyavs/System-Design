package models

// Genre represents movie genre
type Genre string

const (
	GenreAction Genre = "Action"
	GenreSciFi  Genre = "Sci-Fi"
)

// Movie represents a movie entity
type Movie struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Genre    Genre  `json:"genre"`
	Duration int    `json:"duration"` // in minutes
	Rating   string `json:"rating"`   // U, U/A, A
	Language string `json:"language"`
}
