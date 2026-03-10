package models

import "time"

// Book represents a book entity in the bookstore.
// Core entity for catalog management.
type Book struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	ISBN      string    `json:"isbn"`
	Price     float64   `json:"price"`
	Genre     string    `json:"genre"`
	Stock     int       `json:"stock"`
	CreatedAt time.Time `json:"created_at"`
}

// BookQuery represents search/filter criteria for books.
// Used by Builder pattern for complex queries.
type BookQuery struct {
	Title     string
	Author    string
	Genre     string
	MinPrice  float64
	MaxPrice  float64
	InStock   bool
}
