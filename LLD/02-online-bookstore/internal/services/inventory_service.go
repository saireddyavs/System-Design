package services

import (
	"sync"

	"online-bookstore/internal/interfaces"
	"online-bookstore/internal/models"
)

const DefaultLowStockThreshold = 5

// InventoryService manages stock tracking and restocking (SRP).
// Observer pattern: Notifies observers when stock falls below threshold.
type InventoryService struct {
	bookRepo   interfaces.BookRepository
	observers  []interfaces.InventoryObserver
	threshold  int
	mu         sync.RWMutex
}

func NewInventoryService(bookRepo interfaces.BookRepository, threshold int) *InventoryService {
	if threshold <= 0 {
		threshold = DefaultLowStockThreshold
	}
	return &InventoryService{
		bookRepo:  bookRepo,
		observers: make([]interfaces.InventoryObserver, 0),
		threshold: threshold,
	}
}

func (s *InventoryService) RegisterObserver(observer interfaces.InventoryObserver) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, observer)
}

func (s *InventoryService) CheckLowStock() {
	books, err := s.bookRepo.GetAll()
	if err != nil {
		return
	}
	for _, book := range books {
		s.checkAndNotify(book)
	}
}

func (s *InventoryService) checkAndNotify(book *models.Book) {
	if book.Stock <= s.threshold && book.Stock > 0 {
		s.mu.RLock()
		obs := s.observers
		s.mu.RUnlock()
		for _, o := range obs {
			o.OnLowStock(book, s.threshold)
		}
	}
}
