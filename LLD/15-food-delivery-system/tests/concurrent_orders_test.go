package tests

import (
	"context"
	"fmt"
	"food-delivery-system/internal/models"
	"food-delivery-system/internal/repositories"
	"food-delivery-system/internal/services"
	"food-delivery-system/internal/strategies"
	"sync"
	"testing"
)

func setupConcurrentOrderService(t *testing.T) *services.OrderService {
	restaurantRepo := repositories.NewInMemoryRestaurantRepo()
	orderRepo := repositories.NewInMemoryOrderRepo()
	agentRepo := repositories.NewInMemoryAgentRepo()
	customerRepo := repositories.NewInMemoryCustomerRepo()

	restaurantService := services.NewRestaurantService(restaurantRepo)
	customerService := services.NewCustomerService(customerRepo)

	restaurantService.RegisterRestaurant("R1", "Test", []string{"Italian"}, models.Location{Lat: 12.96, Lng: 77.58}, 100)
	restaurantService.AddMenuItem("R1", models.NewMenuItem("M1", "R1", "Pizza", "Pizza", "Main", 200))

	customerService.RegisterCustomer("C1", "C1", "c1@t.com", "1", models.Location{Lat: 12.97, Lng: 77.59})
	customerService.RegisterCustomer("C2", "C2", "c2@t.com", "2", models.Location{Lat: 12.97, Lng: 77.59})

	// Multiple agents for concurrent orders
	for i := 1; i <= 5; i++ {
		agent := models.NewDeliveryAgent(
			fmt.Sprintf("A%d", i),
			"Agent",
			"phone",
			models.Location{Lat: 12.96 + float64(i)*0.001, Lng: 77.58},
		)
		agentRepo.Create(agent)
	}

	deliveryStrategy := strategies.NewNearestAgentStrategy()
	pricingStrategy := strategies.NewDefaultPricingStrategy()
	paymentProcessor := services.NewInMemoryPaymentProcessor()
	observerManager := services.NewOrderObserverManager()

	deliveryService := services.NewDeliveryService(agentRepo, deliveryStrategy, 5.0)
	paymentService := services.NewPaymentService(paymentProcessor)

	return services.NewOrderService(
		orderRepo, restaurantRepo, customerRepo,
		deliveryService, paymentService, pricingStrategy, observerManager,
	)
}

func TestConcurrentOrderPlacement(t *testing.T) {
	orderService := setupConcurrentOrderService(t)
	ctx := context.Background()

	var wg sync.WaitGroup
	orders := make([]*models.Order, 5)
	errors := make([]error, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			customerID := "C1"
			if idx%2 == 1 {
				customerID = "C2"
			}
			order, err := orderService.PlaceOrder(ctx, customerID, "R1", []models.OrderItem{
				{MenuItemID: "M1", Quantity: 1, Price: 200},
			}, models.Location{Lat: 12.97, Lng: 77.59})
			orders[idx] = order
			errors[idx] = err
		}(i)
	}

	wg.Wait()

	successCount := 0
	for i, err := range errors {
		if err == nil && orders[i] != nil {
			successCount++
		}
	}

	if successCount < 1 {
		t.Errorf("Expected at least 1 successful order, got %d", successCount)
	}
}
