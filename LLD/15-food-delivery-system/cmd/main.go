package main

import (
	"context"
	"fmt"
	"food-delivery-system/internal/models"
	"food-delivery-system/internal/repositories"
	"food-delivery-system/internal/services"
	"food-delivery-system/internal/strategies"
)

func main() {
	// Initialize repositories
	restaurantRepo := repositories.NewInMemoryRestaurantRepo()
	orderRepo := repositories.NewInMemoryOrderRepo()
	agentRepo := repositories.NewInMemoryAgentRepo()
	customerRepo := repositories.NewInMemoryCustomerRepo()
	ratingRepo := repositories.NewInMemoryRatingRepo()

	// Initialize strategies
	deliveryStrategy := strategies.NewNearestAgentStrategy()
	pricingStrategy := strategies.NewDefaultPricingStrategy()

	// Initialize payment processor
	paymentProcessor := services.NewInMemoryPaymentProcessor()

	// Initialize observer manager
	observerManager := services.NewOrderObserverManager()
	observerManager.Subscribe(&services.LoggingOrderObserver{})

	// Initialize services
	restaurantService := services.NewRestaurantService(restaurantRepo)
	customerService := services.NewCustomerService(customerRepo)
	deliveryService := services.NewDeliveryService(agentRepo, deliveryStrategy, 5.0)
	paymentService := services.NewPaymentService(paymentProcessor)
	orderService := services.NewOrderService(
		orderRepo, restaurantRepo, customerRepo,
		deliveryService, paymentService, pricingStrategy, observerManager,
	)
	searchService := services.NewSearchService(restaurantRepo)
	ratingService := services.NewRatingService(ratingRepo, restaurantRepo, agentRepo)

	// Seed data
	seedData(restaurantService, customerService, deliveryService)

	// Demo: Search restaurants
	fmt.Println("=== Search: Italian restaurants ===")
	results, _ := searchService.SearchByCuisine("Italian")
	for _, r := range results {
		fmt.Printf("  - %s (Rating: %.1f)\n", r.Name, r.Rating)
	}

	// Demo: Place order
	fmt.Println("\n=== Place Order ===")
	ctx := context.Background()
	order, err := orderService.PlaceOrder(ctx, "C1", "R1", []models.OrderItem{
		{MenuItemID: "M1", Name: "Margherita Pizza", Quantity: 2, Price: 299},
		{MenuItemID: "M2", Name: "Garlic Bread", Quantity: 1, Price: 99},
	}, models.Location{Lat: 12.97, Lng: 77.59})
	if err != nil {
		fmt.Printf("Order failed: %v\n", err)
		return
	}
	fmt.Printf("Order placed: %s, Total: Rs.%.2f, Status: %s\n", order.ID, order.Total, order.Status)

	// Demo: Order lifecycle
	fmt.Println("\n=== Order Lifecycle ===")
	orderService.UpdateOrderStatus(order.ID, models.OrderStatusPreparing)
	orderService.UpdateOrderStatus(order.ID, models.OrderStatusPickedUp)
	orderService.UpdateOrderStatus(order.ID, models.OrderStatusDelivered)
	tracked, _ := orderService.GetOrderTracking(order.ID)
	fmt.Printf("Final status: %s\n", tracked.Status)

	// Demo: Rating
	_ = ratingService.RateRestaurant(order.ID, "C1", "R1", 4.5, "Great food!")
	_ = ratingService.RateAgent(order.ID, "C1", order.AgentID, 5.0, "Fast delivery!")

	fmt.Println("\n=== Food Delivery System Demo Complete ===")
}

func seedData(restaurantService *services.RestaurantService, customerService *services.CustomerService, deliveryService *services.DeliveryService) {
	// Restaurant
	restaurantService.RegisterRestaurant("R1", "Pizza Paradise", []string{"Italian", "Pizza"}, models.Location{Lat: 12.96, Lng: 77.58}, 100)
	restaurantService.AddMenuItem("R1", models.NewMenuItem("M1", "R1", "Margherita Pizza", "Classic tomato and mozzarella", "Pizza", 299))
	restaurantService.AddMenuItem("R1", models.NewMenuItem("M2", "R1", "Garlic Bread", "Crispy bread with garlic butter", "Starter", 99))

	restaurantService.RegisterRestaurant("R2", "Spice Garden", []string{"Indian", "North Indian"}, models.Location{Lat: 12.98, Lng: 77.60}, 150)

	// Customer
	customerService.RegisterCustomer("C1", "John Doe", "john@example.com", "9876543210", models.Location{Lat: 12.97, Lng: 77.59})

	// Delivery agents (near R1)
	deliveryService.RegisterAgent("A1", "Agent One", "1111111111", models.Location{Lat: 12.955, Lng: 77.575})
	deliveryService.RegisterAgent("A2", "Agent Two", "2222222222", models.Location{Lat: 12.965, Lng: 77.585})
}
