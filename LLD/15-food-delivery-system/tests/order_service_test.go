package tests

import (
	"context"
	"food-delivery-system/internal/models"
	"food-delivery-system/internal/repositories"
	"food-delivery-system/internal/services"
	"food-delivery-system/internal/strategies"
	"testing"
)

func setupOrderService(t *testing.T) (*services.OrderService, *services.RestaurantService, *services.CustomerService, *services.DeliveryService) {
	restaurantRepo := repositories.NewInMemoryRestaurantRepo()
	orderRepo := repositories.NewInMemoryOrderRepo()
	agentRepo := repositories.NewInMemoryAgentRepo()
	customerRepo := repositories.NewInMemoryCustomerRepo()

	deliveryStrategy := strategies.NewNearestAgentStrategy()
	pricingStrategy := strategies.NewDefaultPricingStrategy()
	paymentProcessor := services.NewInMemoryPaymentProcessor()
	observerManager := services.NewOrderObserverManager()

	deliveryService := services.NewDeliveryService(agentRepo, deliveryStrategy, 5.0)
	paymentService := services.NewPaymentService(paymentProcessor)
	orderService := services.NewOrderService(
		orderRepo, restaurantRepo, customerRepo,
		deliveryService, paymentService, pricingStrategy, observerManager,
	)
	restaurantService := services.NewRestaurantService(restaurantRepo)
	customerService := services.NewCustomerService(customerRepo)

	// Seed
	restaurantService.RegisterRestaurant("R1", "Test Restaurant", []string{"Italian"}, models.Location{Lat: 12.96, Lng: 77.58}, 100)
	restaurantService.AddMenuItem("R1", models.NewMenuItem("M1", "R1", "Pizza", "Good pizza", "Main", 200))
	restaurantService.AddMenuItem("R1", models.NewMenuItem("M2", "R1", "Bread", "Garlic bread", "Starter", 50))

	customerService.RegisterCustomer("C1", "Customer", "c@test.com", "123", models.Location{Lat: 12.97, Lng: 77.59})

	agent := models.NewDeliveryAgent("A1", "Agent", "999", models.Location{Lat: 12.955, Lng: 77.575})
	agentRepo.Create(agent)

	return orderService, restaurantService, customerService, deliveryService
}

func TestPlaceOrder_Success(t *testing.T) {
	orderService, _, _, _ := setupOrderService(t)
	ctx := context.Background()

	order, err := orderService.PlaceOrder(ctx, "C1", "R1", []models.OrderItem{
		{MenuItemID: "M1", Quantity: 2, Price: 200},
		{MenuItemID: "M2", Quantity: 1, Price: 50},
	}, models.Location{Lat: 12.97, Lng: 77.59})

	if err != nil {
		t.Fatalf("PlaceOrder failed: %v", err)
	}
	if order == nil {
		t.Fatal("Expected order, got nil")
	}
	if order.Status != models.OrderStatusConfirmed {
		t.Errorf("Expected status confirmed, got %s", order.Status)
	}
	if order.Total <= 0 {
		t.Errorf("Expected positive total, got %.2f", order.Total)
	}
	if order.AgentID == "" {
		t.Error("Expected agent assigned")
	}
}

func TestPlaceOrder_MinOrderNotMet(t *testing.T) {
	orderService, _, _, _ := setupOrderService(t)
	ctx := context.Background()

	_, err := orderService.PlaceOrder(ctx, "C1", "R1", []models.OrderItem{
		{MenuItemID: "M2", Quantity: 1, Price: 50},
	}, models.Location{Lat: 12.97, Lng: 77.59})

	if err != services.ErrMinOrderNotMet {
		t.Errorf("Expected ErrMinOrderNotMet, got %v", err)
	}
}

func TestPlaceOrder_RestaurantClosed(t *testing.T) {
	orderService, restaurantService, _, _ := setupOrderService(t)
	restaurantService.UpdateRestaurantStatus("R1", false)
	ctx := context.Background()

	_, err := orderService.PlaceOrder(ctx, "C1", "R1", []models.OrderItem{
		{MenuItemID: "M1", Quantity: 2, Price: 200},
	}, models.Location{Lat: 12.97, Lng: 77.59})

	if err != services.ErrRestaurantClosed {
		t.Errorf("Expected ErrRestaurantClosed, got %v", err)
	}
}

func TestCancelOrder_BeforePreparation(t *testing.T) {
	orderService, _, _, _ := setupOrderService(t)
	ctx := context.Background()

	order, _ := orderService.PlaceOrder(ctx, "C1", "R1", []models.OrderItem{
		{MenuItemID: "M1", Quantity: 1, Price: 200},
	}, models.Location{Lat: 12.97, Lng: 77.59})

	err := orderService.CancelOrder(order.ID)
	if err != nil {
		t.Fatalf("CancelOrder failed: %v", err)
	}

	updated, _ := orderService.GetOrder(order.ID)
	if updated.Status != models.OrderStatusCancelled {
		t.Errorf("Expected cancelled, got %s", updated.Status)
	}
}

func TestOrderStatusTransition(t *testing.T) {
	orderService, _, _, _ := setupOrderService(t)
	ctx := context.Background()

	order, _ := orderService.PlaceOrder(ctx, "C1", "R1", []models.OrderItem{
		{MenuItemID: "M1", Quantity: 1, Price: 200},
	}, models.Location{Lat: 12.97, Lng: 77.59})

	transitions := []models.OrderStatus{
		models.OrderStatusPreparing,
		models.OrderStatusPickedUp,
		models.OrderStatusDelivered,
	}

	for _, status := range transitions {
		err := orderService.UpdateOrderStatus(order.ID, status)
		if err != nil {
			t.Fatalf("UpdateOrderStatus to %s failed: %v", status, err)
		}
	}

	tracked, _ := orderService.GetOrderTracking(order.ID)
	if tracked.Status != models.OrderStatusDelivered {
		t.Errorf("Expected delivered, got %s", tracked.Status)
	}
}
