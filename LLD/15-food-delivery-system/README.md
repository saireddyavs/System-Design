# Food Delivery System - Low Level Design

A production-quality, interview-ready LLD implementation of a food delivery platform (like Zomato/DoorDash) in Go, following clean architecture and SOLID principles.

## Problem Description

Design a food delivery system that enables:
- **Restaurants** to manage menus and accept orders
- **Customers** to search restaurants, place orders, and track delivery
- **Delivery agents** to receive assignments and complete deliveries
- **Real-time order tracking** through the order lifecycle
- **Payment processing** and **rating system**

## Requirements

| Feature | Description |
|---------|-------------|
| Restaurant Management | Register, menus, open/closed status, min order |
| Customer Registration | Profile, addresses, order history |
| Order Lifecycle | Placed → Confirmed → Preparing → Picked Up → Delivered |
| Delivery Assignment | Nearest available agent within radius (Haversine) |
| Search | By cuisine, name, location |
| Payment | Process payment (simulated) |
| Ratings | Rate restaurants and delivery agents |
| Order cancellation | Only before preparation starts |

## Core Entities & Relationships

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Restaurant │────<│  MenuItem   │     │   Customer  │
└──────┬──────┘     └─────────────┘     └──────┬──────┘
       │                                       │
       │              ┌─────────────┐     ┌────┴────┐
       └─────────────>│    Order    │<────┘         │
                      └──────┬──────┘               │
                             │                      │
                      ┌──────┴──────┐     ┌────────┴────────┐
                      │    Payment   │     │ DeliveryAgent   │
                      └─────────────┘     └─────────────────┘
                             │
                      ┌──────┴──────┐
                      │   Rating    │
                      └─────────────┘
```

### Entity Details

| Entity | Key Fields |
|--------|------------|
| **Restaurant** | ID, Name, Cuisines, Location, Rating, IsOpen, Menu, MinOrder |
| **MenuItem** | ID, RestaurantID, Name, Price, Category, IsAvailable |
| **Customer** | ID, Name, Email, Phone, Location, Addresses |
| **DeliveryAgent** | ID, Name, Phone, Location, Status (Available/OnDelivery/Offline), Rating |
| **Order** | ID, CustomerID, RestaurantID, AgentID, Items, SubTotal, DeliveryFee, Total, Status |
| **Location** | Lat, Lng (with Haversine distance) |

## Delivery Assignment Algorithm

1. **Fetch** all agents with status `Available`
2. **Filter** agents within `maxRadiusKm` (default 5km) from restaurant using Haversine distance
3. **Select** using strategy (default: nearest agent)
4. **Assign** and mark agent as `OnDelivery`
5. **Handle** no agents: return error, order cannot be placed

### Haversine Formula

```go
func (l Location) Distance(other Location) float64 {
    // Converts lat/lng to radians, computes great-circle distance
    // Returns distance in kilometers
}
```

## Order State Machine

```
     ┌─────────┐
     │ PLACED  │
     └────┬────┘
          │ confirm
          ▼
     ┌─────────────┐     cancel
     │ CONFIRMED   │◄─────────────┐
     └──────┬──────┘              │
            │ start preparing     │
            ▼                     │
     ┌─────────────┐     cancel   │
     │ PREPARING   │──────────────┘
     └──────┬──────┘
            │ agent picks up
            ▼
     ┌─────────────┐
     │ PICKED_UP   │
     └──────┬──────┘
            │ delivered
            ▼
     ┌─────────────┐
     │ DELIVERED   │
     └─────────────┘

     CANCELLED (terminal, from PLACED or CONFIRMED only)
```

## Design Patterns

| Pattern | Where | Why |
|---------|-------|-----|
| **Strategy** | `DeliveryStrategy`, `PricingStrategy` | Different strategies for agent assignment (nearest agent) and pricing (base + delivery + surge) without changing core logic |
| **Observer** | `OrderObserver`, `OrderObserverManager` | Notify multiple consumers (logging, push, email) when order status changes |
| **State** | Order status transitions | `canTransition()` enforces valid state transitions; prevents invalid flows |
| **Factory** | `OrderService.PlaceOrder` | Creates order with validated items, calculated amounts, assigned agent |
| **Repository** | All `*Repository` interfaces | Abstracts data access; swap in-memory for DB without changing services |

## Data Structures & Algorithms

| DS/Algorithm | Where Used | Why | Alternatives/Tradeoffs |
|--------------|------------|-----|------------------------|
| **Haversine formula** | `Location.Distance()` | Great-circle distance between restaurant, customer, and delivery agent for fee and assignment | Euclidean approximation (faster, less accurate); PostGIS/Redis Geo for scale |
| **Linear scan** | `NearestAgentStrategy.AssignAgent()` | Iterate available agents, compute distance, pick nearest within radius | k-d tree / spatial index for large agent counts; heap for top-k |
| **Order state machine** | `Order.canTransition()`, `Order.SetStatus()` | Enforce valid lifecycle: Placed → Confirmed → Preparing → PickedUp → Delivered | State pattern with explicit states; event sourcing for audit |
| **Restaurant search with filters** | `SearchService.SearchRestaurants()` | Intersect results from name, cuisine, location; filter open only | Inverted index for text; geospatial index for location |
| **Pricing (distance + surge)** | `DefaultPricingStrategy` | Delivery fee = f(distance); surge = f(time of day) | Dynamic surge from demand; zone-based pricing |

## SOLID Principles Mapping

| Principle | Implementation |
|-----------|-----------------|
| **S**ingle Responsibility | Each service has one concern: `OrderService` (orders), `DeliveryService` (agents), `SearchService` (search) |
| **O**pen/Closed | New delivery strategies (e.g. `LoadBalancedStrategy`) without modifying `DeliveryService` |
| **L**iskov Substitution | Any `DeliveryStrategy` implementation works; any `PricingStrategy` works |
| **I**nterface Segregation | Small interfaces: `RestaurantRepository`, `OrderRepository`, `PaymentProcessor` |
| **D**ependency Inversion | Services depend on interfaces (repositories, strategies), not concrete implementations |

## Business Rules

- Restaurant must be **open** to accept orders
- Menu items must be **available**
- Order must meet **minimum order amount** per restaurant
- Delivery fee based on **distance** (Haversine)
- **Surge pricing** (20%) during peak hours (12-2 PM, 7-9 PM)
- Order **cancellation** only before preparation starts

## Directory Structure

```
15-food-delivery-system/
├── cmd/main.go                 # Entry point, demo
├── internal/
│   ├── models/                 # Domain entities
│   ├── interfaces/             # Repository & strategy contracts
│   ├── services/               # Business logic
│   ├── repositories/           # In-memory implementations
│   └── strategies/             # Delivery & pricing strategies
├── tests/                      # Unit tests
├── go.mod
└── README.md
```

## Running the Application

```bash
go run ./cmd/main.go
```

## Running Tests

```bash
go test ./tests/... -v
```

## Interview Explanations

### 3-Minute Summary

"We built a food delivery system with clean architecture. Core entities: Restaurant, Customer, Order, DeliveryAgent. Order flows through states: Placed → Confirmed → Preparing → Picked Up → Delivered. We use the Strategy pattern for delivery assignment—we pick the nearest available agent within 5km using Haversine distance. Pricing uses another strategy: base + distance-based delivery fee + surge during peak hours. The Observer pattern notifies consumers on status changes. Repositories abstract data access; we can swap in-memory for PostgreSQL. All services depend on interfaces, following SOLID."

### 10-Minute Deep Dive

1. **Architecture**: Clean architecture with models, interfaces, services, repositories. Dependencies point inward.

2. **Order Lifecycle**: State machine with `canTransition()`—only valid transitions allowed. Cancellation only before preparing.

3. **Delivery Assignment**: Strategy pattern. `NearestAgentStrategy` finds agents within radius, picks closest. Haversine for distance.

4. **Pricing**: `PricingStrategy` computes delivery fee from distance, surge from time. Easy to add new strategies.

5. **Concurrency**: Repositories use `sync.RWMutex` for thread-safe access. Multiple concurrent orders handled.

6. **Testing**: Unit tests for order placement, cancellation, delivery assignment, search, concurrent orders.

7. **Extensibility**: Add new strategies without changing services. Add observers for push/email. Swap repositories for DB.

## Future Improvements

1. **Persistence**: Replace in-memory repos with PostgreSQL/Redis
2. **Real-time**: WebSocket for live order tracking
3. **Load-balanced strategy**: Distribute orders across agents
4. **Idempotency**: Idempotency keys for payment/order APIs
5. **Event sourcing**: Store order events for audit
6. **Caching**: Cache restaurant search results
7. **Circuit breaker**: For payment gateway failures
8. **Saga pattern**: Distributed transactions for order + payment
