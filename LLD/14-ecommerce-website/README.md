# E-commerce Website - Low Level Design (LLD)

A production-quality, interview-ready Go implementation of an e-commerce platform following **Clean Architecture** and **SOLID** principles.

---

## 1. Problem Description & Requirements

### Problem
Design and implement the backend for an e-commerce website that supports product catalog, user management, shopping cart, order placement, payment processing, inventory management, and discount system.

### Functional Requirements
| Feature | Description |
|---------|-------------|
| **Product Catalog** | Categories, pricing, inventory, SKU, images, ratings |
| **User Management** | Registration, login, profile, multiple addresses |
| **Shopping Cart** | Add, remove, update quantity, clear |
| **Order Placement** | Place order from cart, track status |
| **Payment Processing** | Multiple methods: Credit Card, Debit Card, UPI, Wallet |
| **Inventory Management** | Stock tracking, decrement on order (via OrderService) |
| **Order History** | List orders for a user |
| **Coupon/Discount** | Percentage, flat, BOGO (buy-one-get-one) |

### Business Rules
- Stock must be validated before order placement
- Stock decremented on order confirmation
- Coupon validation: expiry, min order amount, usage limit
- Order lifecycle: Placed → Confirmed → Shipped → Delivered (or Cancelled/Returned)
- Cart abandoned after 24 hours (future feature)

---

## 2. Core Entities & Relationships

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Category  │────<│   Product   │     │    User     │
└─────────────┘     └──────┬──────┘     └──────┬──────┘
                          │                    │
                          │                    │
                    ┌─────▼─────┐        ┌─────▼─────┐
                    │ CartItem  │        │  Address  │
                    └─────┬─────┘        └───────────┘
                          │
                    ┌─────▼─────┐
                    │   Cart    │
                    └─────┬─────┘
                          │
                    ┌─────▼─────┐     ┌─────────────┐
                    │   Order   │────>│   Payment   │
                    └─────┬─────┘     └─────────────┘
                          │
                    ┌─────▼─────┐
                    │  Coupon   │
                    └───────────┘
```

### Entity Details

| Entity | Key Fields |
|--------|------------|
| **Product** | ID, Name, Description, Price, CategoryID, Stock, SKU, Images, Rating |
| **Category** | ID, Name, ParentID, Description |
| **User** | ID, Name, Email, Password, Addresses[], Phone |
| **Cart** | ID, UserID, Items{ProductID→CartItem}, UpdatedAt |
| **CartItem** | ProductID, Quantity, Price |
| **Order** | ID, UserID, Items[], TotalAmount, Discount, FinalAmount, Status, ShippingAddress, PaymentID |
| **Payment** | ID, OrderID, Amount, Method, Status, TransactionID |
| **Coupon** | ID, Code, Type, Value, MinOrderAmount, ExpiresAt, UsageLimit, UsedCount |

---

## 3. Order Lifecycle

```
    ┌─────────┐     ┌───────────┐     ┌─────────┐     ┌──────────┐
    │ Placed  │────>│ Confirmed  │────>│ Shipped │────>│ Delivered│
    └─────────┘     └───────────┘     └─────────┘     └──────────┘
         │                 │                │
         │                 │                │
         ▼                 ▼                ▼
    ┌──────────┐     ┌──────────┐     ┌──────────┐
    │Cancelled │     │ Returned │     │ Returned │
    └──────────┘     └──────────┘     └──────────┘
```

**Flow:**
1. **Placed** – Order created from cart, payment initiated
2. **Confirmed** – Payment successful, stock decremented, cart cleared
3. **Shipped** – Order dispatched
4. **Delivered** – Order received
5. **Cancelled** – Order cancelled (Placed/Confirmed only), stock restored
6. **Returned** – Items returned, stock restored

---

## 4. Design Patterns with WHY

| Pattern | Where | Why |
|--------|-------|-----|
| **Strategy** | Payment (CreditCard, DebitCard, UPI, Wallet) | Different payment gateways have different APIs. Strategy allows adding new methods without modifying OrderService. **Open/Closed Principle**. |
| **Strategy** | Discount (Percentage, Flat, BOGO) | Discount calculation logic varies by coupon type. Strategy encapsulates each algorithm. Easy to add new types (e.g., Category-specific). |
| **Observer** | Inventory low stock, Order status | Multiple subscribers (email, SMS, analytics) need to react to events. Observer decouples event source from handlers. **Single Responsibility**. |
| **Factory** | Order creation from cart | Order construction is complex (items, totals, discount). Factory centralizes creation logic and ensures valid orders. |
| **Repository** | All data access | Abstracts storage. Swap in-memory for PostgreSQL without changing services. **Dependency Inversion**. |
| **Builder** | Product creation | Product has many optional fields. Builder provides fluent API and enforces required fields. |

---

## 5. Data Structures & Algorithms

| DS/Algorithm | Where Used | Why | Alternatives/Tradeoffs |
|--------------|------------|-----|------------------------|
| **HashMap repositories** | `InMemoryProductRepo.products map[string]*Product`, Cart, Order, etc. | O(1) lookup by ID; standard in-memory storage for LLD | Slice + linear search: O(n); B-tree: for ordered range queries |
| **Order total with discounts** | `OrderService.PlaceOrder` — cart items loop, `CouponService.ApplyCoupon`, `DiscountStrategy.Calculate` | Sum item prices; apply coupon via strategy; final = total − discount | Single formula: less flexible; multiple discount stacking: more complex |
| **Inventory decrement (sync.Mutex)** | `InMemoryProductRepo.DecrementStock` — `r.mu.Lock()` | Atomic stock update; prevent oversell under concurrent orders | Optimistic locking: version field; DB transaction: for persistence |
| **Payment strategy dispatch** | `PaymentService.processors map[PaymentMethod]PaymentProcessor` | O(1) lookup by method; route to correct processor (Credit, Debit, UPI, Wallet) | Switch: O(n) branches; Strategy interface: same pattern |

---

## 6. SOLID Principles Mapping

| Principle | Implementation |
|----------|----------------|
| **S - Single Responsibility** | `CartService` only handles cart. `OrderService` orchestrates but delegates to repositories, payment, coupon. |
| **O - Open/Closed** | New payment methods: add `PaymentProcessor` impl, register in `PaymentService`. New discount types: add `DiscountStrategy`. No changes to existing code. |
| **L - Liskov Substitution** | Any `ProductRepository` impl (in-memory, PostgreSQL) can replace another. Same for `PaymentProcessor`, `DiscountStrategy`. |
| **I - Interface Segregation** | Small, focused interfaces: `ProductRepository`, `CartRepository`, `PaymentProcessor`, `NotificationService`. No fat interfaces. |
| **D - Dependency Inversion** | Services depend on `interfaces.ProductRepository`, not `*InMemoryProductRepo`. High-level modules don't depend on low-level. |

---

## 7. Interview Explanations

### 3-Minute Pitch
> "I've designed an e-commerce backend in Go using Clean Architecture. The core flow is: User adds items to cart → Places order with optional coupon → We validate stock, apply discount via Strategy, process payment via Strategy, decrement inventory, and notify via Observer. Repositories abstract storage, so we can swap in-memory for DB. Payment and discount use Strategy pattern so we can add UPI or BOGO coupons without touching existing code. All repositories are thread-safe with RWMutex. The order lifecycle goes Placed → Confirmed → Shipped → Delivered, with Cancelled/Returned for rollback."

### 10-Minute Deep Dive
1. **Architecture**: `cmd/` for entry, `internal/` for domain. Models are pure structs. Interfaces define contracts. Services contain business logic. Repositories implement persistence.
2. **Order Flow**: `PlaceOrder` validates stock, applies coupon (Strategy), creates payment, decrements stock (with rollback on failure), processes payment (Strategy), persists order, clears cart, notifies (Observer).
3. **Thread Safety**: All in-memory repos use `sync.RWMutex`. `DecrementStock` is atomic. Concurrent tests verify no race conditions.
4. **Extensibility**: New payment method = implement `PaymentProcessor`, register. New discount = implement `DiscountStrategy`, register. New notification channel = implement `NotificationService`, subscribe to Observer.
5. **Testing**: Unit tests for cart, order, inventory, coupon, and concurrent stock updates. Mocks not needed—in-memory repos suffice.

---

## 8. Future Improvements

| Area | Improvement |
|------|-------------|
| **Persistence** | Replace in-memory repos with PostgreSQL (GORM/Sqlx) |
| **Authentication** | JWT-based auth, password hashing (bcrypt) |
| **API Layer** | HTTP handlers (Chi/Gin), request validation |
| **Caching** | Redis for product catalog, cart |
| **Event-Driven** | Kafka/RabbitMQ for order events, async notifications |
| **Cart Abandonment** | Cron job to clear carts older than 24h |
| **Idempotency** | Idempotency keys for payment/order to prevent duplicates |
| **Distributed Lock** | Redis lock for stock decrement in multi-instance deployment |

---

## 9. Project Structure

```
14-ecommerce-website/
├── cmd/main.go                 # Application entry, dependency wiring
├── internal/
│   ├── models/                 # Domain entities
│   ├── interfaces/             # Repository & strategy contracts
│   ├── services/               # Business logic
│   ├── repositories/           # In-memory implementations
│   ├── strategies/             # Payment & discount strategies
│   ├── observer/               # Inventory & order observers
│   └── factory/                # Order factory
├── tests/                      # Unit tests
├── go.mod
└── README.md
```

---

## 10. Running the Project

```bash
# Build
go build ./...

# Run
go run ./cmd/main.go

# Tests
go test ./tests/... -v
```

---

## 11. Key Design Decisions

1. **In-memory storage** – Simplifies demo; production would use DB
2. **Synchronous payment** – Real system would use async webhooks
3. **No auth middleware** – Focus on domain logic; auth is a cross-cutting concern
4. **Observer for notifications** – Logging implementation; can add email/SMS
5. **Builder for Product** – Demonstrates creational pattern; models are simple enough to construct directly too
