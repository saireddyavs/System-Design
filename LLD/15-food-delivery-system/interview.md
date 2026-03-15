# Food Delivery System — Interview Strategy (45 min)

## Time Allocation

| Phase | Time | What to Do |
|-------|------|------------|
| 1. Clarify & Scope | 3 min | Confirm: restaurant/customer/agent flows, order lifecycle, delivery assignment rule, pricing (distance + surge) |
| 2. Core Models | 7 min | Restaurant, Customer, Order, DeliveryAgent, Location, MenuItem; OrderStatus enum; canTransition map |
| 3. Repository Interfaces | 5 min | RestaurantRepository, OrderRepository, AgentRepository, CustomerRepository; key methods |
| 4. Service Interfaces | 5 min | DeliveryStrategy, PricingStrategy; OrderObserver; PaymentProcessor |
| 5. Core Service Implementation | 12 min | OrderService.PlaceOrder (validate → assign agent → calculate total → create order) |
| 6. main.go Wiring | 5 min | Wire repos, strategies, services; minimal demo: search → place order → status updates |
| 7. Extend & Discuss | 8 min | Haversine, nearest-agent linear scan, order state machine; scaling, persistence |

## Phase 1: Clarify & Scope (3 min)

**Questions to ask:**
- Single restaurant per order or multi-vendor?
- How is delivery agent assigned? (Nearest available? Load-balanced?)
- What drives pricing? (Distance, surge, tips?)
- Order cancellation rules? (Before preparing only?)

**Scope in:** Place order, assign agent, calculate total, order lifecycle (Placed → Delivered), search restaurants by cuisine/name/location.

**Scope out:** Payment gateway integration, real-time tracking UI, multi-restaurant orders, tips.

## Phase 2: Core Models (7 min)

**Write first:** `Location` (Lat, Lng + `Distance(other)` — stub Haversine), `Order` (ID, CustomerID, RestaurantID, AgentID, Items, SubTotal, DeliveryFee, Total, Status), `OrderStatus` enum.

**Essential fields:**
- `Order`: Status, Items (MenuItemID, Quantity, Price), DeliveryAddress (Location)
- `DeliveryAgent`: ID, Location, Status (Available/OnDelivery)
- `Restaurant`: ID, Name, Cuisines, Location, IsOpen, Menu
- `Order.canTransition(to)` — map of valid transitions

**Skip:** Rating, Payment details, MenuItem full model (ID + Price enough).

## Phase 3: Repository Interfaces (5 min)

```go
type RestaurantRepository interface {
    GetByID(id string) (*Restaurant, error)
    SearchByCuisine(cuisine string) ([]*Restaurant, error)
    SearchByLocation(loc Location, radiusKm float64) ([]*Restaurant, error)
}
type OrderRepository interface { Create(o *Order) error; GetByID(id string) (*Order, error); Update(o *Order) error }
type AgentRepository interface { GetAvailableAgents() ([]*DeliveryAgent, error); Update(a *DeliveryAgent) error }
```

Keep interfaces small; in-memory impl = map + mutex.

## Phase 4: Service Interfaces (5 min)

```go
type DeliveryStrategy interface {
    AssignAgent(restaurantLoc, deliveryAddr Location, agents []*DeliveryAgent, maxRadiusKm float64) (*DeliveryAgent, error)
}
type PricingStrategy interface {
    CalculateDeliveryFee(restaurantLoc, deliveryAddr Location) float64
    CalculateSurgeFee(orderTime time.Time) float64
    CalculateTotal(subTotal, deliveryFee, surgeFee float64) float64
}
type OrderObserver interface { OnOrderStatusChanged(order *Order, oldStatus, newStatus OrderStatus) }
```

## Phase 5: Core Service Implementation (12 min)

**THE most important method: `OrderService.PlaceOrder`**

1. Validate restaurant open, menu items available, min order met
2. Calculate subtotal from items
3. Call `DeliveryService.AssignAgent(restaurantLoc, deliveryAddr)` — uses strategy (linear scan + Haversine)
4. Call `PricingStrategy.CalculateDeliveryFee`, `CalculateSurgeFee`, `CalculateTotal`
5. Create Order with Status=Placed, persist
6. Confirm order (Status=Confirmed), notify observers
7. Return order

**Why:** This ties together validation, delivery assignment, pricing, and state. Demonstrates Strategy + Repository + Observer.

**NearestAgentStrategy.AssignAgent:** Loop agents, skip non-Available; compute `restaurantLoc.Distance(agent.Location)`; track min; return nearest if within radius.

## Phase 6: main.go Wiring (5 min)

- Create in-memory repos
- `NearestAgentStrategy`, `DefaultPricingStrategy`
- `OrderService` with orderRepo, restaurantRepo, customerRepo, deliveryService, paymentService, pricingStrategy, observerManager
- Demo: seed 1 restaurant + 1 customer + 2 agents; `PlaceOrder`; `UpdateOrderStatus` through lifecycle

## Phase 7: Extend & Discuss (8 min)

- **Haversine:** `a = sin²(Δlat/2) + cos(lat1)*cos(lat2)*sin²(Δlon/2)`; `d = R * 2 * atan2(√a, √(1-a))`
- **Alternatives:** k-d tree for many agents; PostGIS for restaurant search
- **Order state machine:** `canTransition(from, to)` map; prevent invalid flows
- **Scaling:** Shard by region; event sourcing for order audit

## Tips

- Start with Order and Location; everything else supports PlaceOrder
- Stub Haversine as `math.Sqrt((lat2-lat1)² + (lng2-lng1)²)` if time-constrained; mention real formula
- Keep PlaceOrder synchronous; mention async payment/notification as extension
