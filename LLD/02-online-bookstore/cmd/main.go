// Package main demonstrates the Online Bookstore LLD implementation.
// Wire-up with dependency injection following clean architecture.
package main

import (
	"fmt"
	"log"

	"online-bookstore/internal/interfaces"
	"online-bookstore/internal/repositories"
	"online-bookstore/internal/services"
	"online-bookstore/internal/strategies"
)

func main() {
	// Repositories (DIP: inject interfaces)
	var bookRepo interfaces.BookRepository = repositories.NewInMemoryBookRepository()
	var orderRepo interfaces.OrderRepository = repositories.NewInMemoryOrderRepository()
	var userRepo interfaces.UserRepository = repositories.NewInMemoryUserRepository()
	var cartRepo interfaces.CartRepository = repositories.NewInMemoryCartRepository()

	// Search engine (depends on book repo)
	searchEngine := repositories.NewInMemorySearchEngine(bookRepo)

	// Strategies
	orderFactory := strategies.NewOrderFactory()
	paymentRegistry := strategies.NewPaymentProcessorRegistry()

	// Services
	bookService := services.NewBookService(bookRepo)
	authService := services.NewAuthService(userRepo)
	cartService := services.NewCartService(cartRepo, bookRepo)
	orderService := services.NewOrderService(orderRepo, cartRepo, bookRepo, orderFactory, paymentRegistry)
	searchService := services.NewSearchService(searchEngine)
	inventoryService := services.NewInventoryService(bookRepo, 5)

	// Observer: Register low-stock notifier
	inventoryService.RegisterObserver(services.NewLowStockObserver())

	// Demo flow
	runDemo(bookService, authService, cartService, orderService, searchService, inventoryService)
}

func runDemo(
	bookSvc *services.BookService,
	authSvc *services.AuthService,
	cartSvc *services.CartService,
	orderSvc *services.OrderService,
	searchSvc *services.SearchService,
	invSvc *services.InventoryService,
) {
	fmt.Println("=== Online Bookstore Demo ===")

	// 1. Add books
	b1, _ := bookSvc.CreateBook("The Go Programming Language", "Alan Donovan", "978-0134190440", 49.99, "Programming", 10)
	b2, _ := bookSvc.CreateBook("Clean Code", "Robert Martin", "978-0132350884", 39.99, "Programming", 3)
	b3, _ := bookSvc.CreateBook("Design Patterns", "Gang of Four", "978-0201633610", 54.99, "Software Engineering", 15)
	fmt.Printf("Added books: %s, %s, %s\n\n", b1.Title, b2.Title, b3.Title)

	// 2. Register user
	user, err := authSvc.Register("John Doe", "john@example.com", "secret123", "123 Main St")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Registered user: %s (%s)\n\n", user.Name, user.Email)

	// 3. Create cart and add items
	_, _ = cartSvc.CreateCart(user.ID)
	_ = cartSvc.AddToCart(user.ID, b1.ID, 2)
	_ = cartSvc.AddToCart(user.ID, b2.ID, 1)
	cart, _ := cartSvc.GetCart(user.ID)
	fmt.Printf("Cart created with %d item types\n\n", len(cart.Items))

	// 4. Search books
	results, _ := searchSvc.SearchByTitle("Go")
	fmt.Printf("Search 'Go': found %d book(s)\n", len(results))
	for _, b := range results {
		fmt.Printf("  - %s by %s ($%.2f)\n", b.Title, b.Author, b.Price)
	}
	fmt.Println()

	// 5. Place order
	order, err := orderSvc.PlaceOrder(user.ID, "credit_card")
	if err != nil {
		log.Printf("Order failed: %v", err)
	} else {
		fmt.Printf("Order placed: ID=%s, Total=$%.2f, Status=%s\n\n", order.ID, order.TotalAmount, order.Status)
	}

	// 6. Order history
	history, _ := orderSvc.GetOrderHistory(user.ID)
	fmt.Printf("Order history: %d order(s)\n", len(history))

	// 7. Check low stock (Clean Code has 3, threshold is 5)
	invSvc.CheckLowStock()
	fmt.Println("\n=== Demo Complete ===")
}
