package tests

import (
	"food-delivery-system/internal/models"
	"food-delivery-system/internal/repositories"
	"food-delivery-system/internal/services"
	"food-delivery-system/internal/strategies"
	"testing"
)

func TestNearestAgentStrategy_AssignsNearest(t *testing.T) {
	agentRepo := repositories.NewInMemoryAgentRepo()
	strategy := strategies.NewNearestAgentStrategy()

	// Agent A1: 1km from restaurant
	// Agent A2: 3km from restaurant
	// Agent A3: 4km from restaurant (within 5km)
	restaurantLoc := models.Location{Lat: 12.96, Lng: 77.58}
	a1 := models.NewDeliveryAgent("A1", "Agent1", "1", models.Location{Lat: 12.969, Lng: 77.58}) // ~1km N
	a2 := models.NewDeliveryAgent("A2", "Agent2", "2", models.Location{Lat: 12.99, Lng: 77.58})  // ~3.3km N
	a3 := models.NewDeliveryAgent("A3", "Agent3", "3", models.Location{Lat: 13.0, Lng: 77.58}) // ~4.4km N

	agentRepo.Create(a1)
	agentRepo.Create(a2)
	agentRepo.Create(a3)

	deliveryAddr := models.Location{Lat: 12.97, Lng: 77.59}
	deliveryService := services.NewDeliveryService(agentRepo, strategy, 5.0)

	agent, err := deliveryService.AssignAgent(restaurantLoc, deliveryAddr)
	if err != nil {
		t.Fatalf("AssignAgent failed: %v", err)
	}
	if agent.ID != "A1" {
		t.Errorf("Expected nearest agent A1, got %s", agent.ID)
	}
}

func TestNearestAgentStrategy_NoAgentsAvailable(t *testing.T) {
	agentRepo := repositories.NewInMemoryAgentRepo()
	strategy := strategies.NewNearestAgentStrategy()
	deliveryService := services.NewDeliveryService(agentRepo, strategy, 5.0)

	restaurantLoc := models.Location{Lat: 12.96, Lng: 77.58}
	deliveryAddr := models.Location{Lat: 12.97, Lng: 77.59}

	_, err := deliveryService.AssignAgent(restaurantLoc, deliveryAddr)
	if err != services.ErrNoAgentAvailable {
		t.Errorf("Expected ErrNoAgentAvailable, got %v", err)
	}
}

func TestNearestAgentStrategy_AgentsOutsideRadius(t *testing.T) {
	agentRepo := repositories.NewInMemoryAgentRepo()
	strategy := strategies.NewNearestAgentStrategy()

	// Agent 10km away
	restaurantLoc := models.Location{Lat: 12.96, Lng: 77.58}
	a1 := models.NewDeliveryAgent("A1", "Agent1", "1", models.Location{Lat: 13.05, Lng: 77.58}) // ~10km
	agentRepo.Create(a1)

	deliveryAddr := models.Location{Lat: 12.97, Lng: 77.59}
	deliveryService := services.NewDeliveryService(agentRepo, strategy, 5.0)

	_, err := deliveryService.AssignAgent(restaurantLoc, deliveryAddr)
	if err != services.ErrNoAgentAvailable {
		t.Errorf("Expected ErrNoAgentAvailable when agents outside radius, got %v", err)
	}
}

func TestHaversineDistance(t *testing.T) {
	loc1 := models.Location{Lat: 12.96, Lng: 77.58}
	loc2 := models.Location{Lat: 12.97, Lng: 77.59}

	dist := loc1.Distance(loc2)
	if dist <= 0 || dist > 20 {
		t.Errorf("Expected reasonable distance (~1-2km), got %.2f km", dist)
	}
}
