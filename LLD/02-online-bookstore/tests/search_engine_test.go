package tests

import (
	"testing"
	"time"

	"online-bookstore/internal/interfaces"
	"online-bookstore/internal/models"
	"online-bookstore/internal/repositories"
)

func setupSearchTest(t *testing.T) interfaces.SearchEngine {
	bookRepo := repositories.NewInMemoryBookRepository()
	books := []*models.Book{
		{ID: "1", Title: "The Go Programming Language", Author: "Alan Donovan", ISBN: "978-1", Price: 49.99, Genre: "Programming", Stock: 10, CreatedAt: time.Now()},
		{ID: "2", Title: "Clean Code", Author: "Robert Martin", ISBN: "978-2", Price: 39.99, Genre: "Programming", Stock: 5, CreatedAt: time.Now()},
		{ID: "3", Title: "Design Patterns", Author: "Gang of Four", ISBN: "978-3", Price: 54.99, Genre: "Software Engineering", Stock: 8, CreatedAt: time.Now()},
	}
	for _, b := range books {
		_ = bookRepo.Create(b)
	}

	return repositories.NewInMemorySearchEngine(bookRepo)
}

func TestSearchEngine_SearchByTitle(t *testing.T) {
	searchEngine := setupSearchTest(t)

	results, err := searchEngine.SearchByTitle("Go")
	if err != nil {
		t.Fatalf("SearchByTitle failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'Go', got %d", len(results))
	}
	if results[0].Title != "The Go Programming Language" {
		t.Errorf("expected 'The Go Programming Language', got %s", results[0].Title)
	}
}

func TestSearchEngine_SearchByTitle_CaseInsensitive(t *testing.T) {
	searchEngine := setupSearchTest(t)

	results, err := searchEngine.SearchByTitle("CLEAN")
	if err != nil {
		t.Fatalf("SearchByTitle failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'CLEAN' (case-insensitive), got %d", len(results))
	}
}

func TestSearchEngine_SearchByAuthor(t *testing.T) {
	searchEngine := setupSearchTest(t)

	results, err := searchEngine.SearchByAuthor("Martin")
	if err != nil {
		t.Fatalf("SearchByAuthor failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'Martin', got %d", len(results))
	}
}

func TestSearchEngine_SearchByGenre(t *testing.T) {
	searchEngine := setupSearchTest(t)

	results, err := searchEngine.SearchByGenre("Programming")
	if err != nil {
		t.Fatalf("SearchByGenre failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results for 'Programming', got %d", len(results))
	}
}

func TestSearchEngine_Search_AllFields(t *testing.T) {
	searchEngine := setupSearchTest(t)

	results, err := searchEngine.Search("Pattern", interfaces.SearchTypeAll)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'Pattern' in all fields, got %d", len(results))
	}
}
