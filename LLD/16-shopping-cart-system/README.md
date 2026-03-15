# Shopping Cart System - Low Level Design

A production-quality, interview-ready Go implementation of a Shopping Cart System following Clean Architecture and SOLID principles.

## 1. Problem Description & Requirements

### Problem
Design and implement a complete e-commerce shopping cart system that supports product catalog, user management, cart operations, discount/coupon application, tax calculation, checkout, and order confirmation.

### Functional Requirements
- **Product Management**: Products with categories, pricing, SKU, stock, weight
- **User Management**: Registration, session, address
- **Shopping Cart**: Add, update quantity, remove items; persistence per user
- **Discounts/Coupons**: Percentage, flat amount, BOGO (Buy One Get One)
- **Tax Calculation**: Configurable rate (default 18%), state-specific support
- **Checkout**: Validate cart, apply discount, calculate tax, process payment
- **Order Confirmation**: Order summary, status tracking
- **Stock Validation**: During add-to-cart and checkout
- **Cart Event Observers**: Notify on cart add/update/remove for logging

### Non-Functional Requirements
- Thread-safe (sync.RWMutex)
- SOLID principles
- Clean architecture (layered)
- Unit testable

---

## 2. Core Entities & Relationships

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Product   │     │    User     │     │   Coupon    │
├─────────────┤     ├─────────────┤     ├─────────────┤
│ ID          │     │ ID          │     │ ID          │
│ Name        │     │ Name        │     │ Code        │
│ Price       │     │ Email       │     │ Type        │
│ Stock       │     │ Address     │     │ Value       │
│ CategoryID  │     └──────┬──────┘     │ MinOrder    │
│ SKU, Weight │            │            │ ExpiresAt   │
└──────┬──────┘            │            └─────────────┘
       │                   │
       │    ┌──────────────▼──────────────┐
       │    │           Cart              │
       │    ├─────────────────────────────┤
       │    │ ID, UserID, Status          │
       │    │ Items[] CartItem            │
       └───►│ CreatedAt, UpdatedAt        │
            └──────────────┬───────────────┘
                           │
            ┌──────────────▼──────────────┐
            │        CartItem             │
            ├─────────────────────────────┤
            │ ProductID, ProductName      │
            │ UnitPrice, Quantity         │
            │ Subtotal                    │
            └─────────────────────────────┘
                           │
            ┌──────────────▼──────────────┐
            │          Order              │
            ├─────────────────────────────┤
            │ ID, UserID, Items           │
            │ Subtotal, Discount, Tax     │
            │ Total, Status, PaymentID    │
            │ ShippingAddress             │
            └─────────────────────────────┘
```

### Entity Summary
| Entity    | Key Fields |
|-----------|------------|
| Product   | ID, Name, Price, Stock, CategoryID, SKU, Weight |
| User      | ID, Name, Email, Address |
| Cart      | ID, UserID, Items, Status (Active/Abandoned/CheckedOut) |
| CartItem  | ProductID, ProductName, UnitPrice, Quantity, Subtotal |
| Order     | ID, UserID, Items, Subtotal, Discount, Tax, Total, PaymentID |
| Coupon    | ID, Code, Type, Value, MinOrderAmount, ExpiresAt, MaxUsageLimit |

---

## 3. Checkout Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        CHECKOUT FLOW                                    │
└─────────────────────────────────────────────────────────────────────────┘

  ┌──────────────┐
  │ Start        │
  └──────┬───────┘
         │
         ▼
  ┌──────────────────────┐     ┌─────────────────┐
  │ 1. Validate Cart     │────►│ Empty? → Error  │
  │    (not empty,       │     └─────────────────┘
  │     active status)   │
  └──────┬───────────────┘
         │
         ▼
  ┌──────────────────────┐     ┌─────────────────┐
  │ 2. Stock Validation  │────►│ Insufficient?   │
  │    (all items)       │     │ → Error        │
  └──────┬───────────────┘     └─────────────────┘
         │
         ▼
  ┌──────────────────────┐
  │ 3. Calculate Subtotal│
  └──────┬───────────────┘
         │
         ▼
  ┌──────────────────────┐     ┌─────────────────┐
  │ 4. Apply Coupon      │────►│ Invalid? → Skip │
  │    (validate, calc)  │     └─────────────────┘
  └──────┬───────────────┘
         │
         ▼
  ┌──────────────────────┐
  │ 5. Calculate Tax     │
  │    (subtotal - disc) │
  └──────┬───────────────┘
         │
         ▼
  ┌──────────────────────┐
  │ 6. Process Payment   │
  └──────┬───────────────┘
         │
         ▼
  ┌──────────────────────┐
  │ 7. Decrement Stock   │
  └──────┬───────────────┘
         │
         ▼
  ┌──────────────────────┐
  │ 8. Create Order      │
  └──────┬───────────────┘
         │
         ▼
  ┌──────────────────────┐
  │ 9. Clear Cart        │
  │    (new empty cart)  │
  └──────┬───────────────┘
         │
         ▼
  ┌──────────────────────┐
  │ 10. Return Summary   │
  └──────────────────┬───┘
                     │
                     ▼
              ┌──────────────┐
              │ Order        │
              │ Confirmation │
              └──────────────┘
```

---

## 4. Design Patterns & WHY

| Pattern | Where | Why |
|---------|-------|-----|
| **Strategy** | Discount (%, Flat, BOGO) | Different discount algorithms without changing checkout logic. Open/Closed: add new coupon types without modifying existing code. |
| **Strategy** | Tax (Flat, State-based) | Tax rules vary by jurisdiction. Swap calculator without touching checkout. |
| **Strategy** | Payment (Card, PayPal, Wallet) | Payment gateways differ. Pluggable processors. |
| **Observer** | Cart events | Decouple cart operations from side effects (analytics, notifications). CartService doesn't know who listens. |
| **Factory** | OrderFactory | Encapsulates order creation from cart. Ensures consistent order structure. |
| **Repository** | All data access | Abstracts storage. Swap in-memory for DB without changing services. |
| **Builder** | CheckoutSummaryBuilder | Stepwise construction of complex checkout result. Fluent API. |

---

## 5. Data Structures & Algorithms

| DS/Algorithm | Where Used | Why | Alternatives/Tradeoffs |
|--------------|------------|-----|------------------------|
| **HashMap (map)** | Cart repo (`carts`, `userCarts`), product lookup | O(1) cart/user/product lookup by ID; cart items by ProductID | Slice for small carts; B-tree for ordered iteration |
| **Discount chain** | `DiscountStrategyRegistry` → Percentage, Flat, BOGO | Each coupon type uses its strategy; registry selects by `CouponType` | Chain of Responsibility; composite discounts |
| **Tax calculation strategies** | `FlatTaxCalculator`, `StateTaxCalculator` | Tax = (subtotal - discount) × rate; state-specific rates in map | Rule engine; external tax service |
| **Coupon validation** | `CouponService.ValidateAndGetDiscount()` | Check expiry, min order, usage limit; compute discount via strategy | Caching for hot coupons; idempotency for usage |
| **Checkout summary builder** | `CheckoutService.Checkout()` | Build summary stepwise: subtotal → discount → tax → total | DTO; immutable result object |

---

## 6. SOLID Principles Mapping

| Principle | Implementation |
|-----------|----------------|
| **S - Single Responsibility** | CartService: cart ops only. CheckoutService: checkout flow. CouponService: coupon validation. Each class has one reason to change. |
| **O - Open/Closed** | Strategy pattern: add PercentageDiscountStrategy, FlatDiscountStrategy without modifying CheckoutService. Extend via new strategies. |
| **L - Liskov Substitution** | All DiscountStrategy implementations are interchangeable. All PaymentProcessors can replace each other. |
| **I - Interface Segregation** | ProductRepository, CartRepository, OrderRepository are small, focused. No fat interfaces. |
| **D - Dependency Inversion** | Services depend on interfaces (ProductRepository, TaxCalculator), not concrete repos. Injected via constructors. |

---

## 7. Interview Explanations

### 3-Minute Pitch
> "I designed a shopping cart system with clean architecture. Core entities are Product, User, Cart, CartItem, Order, and Coupon. The checkout flow validates cart and stock, applies discounts via Strategy pattern, calculates tax, processes payment, decrements stock, creates order, and clears cart. I used Strategy for discounts (percentage, flat, BOGO), tax, and payment so we can add new types without changing checkout logic. Observer pattern notifies on cart events. Repositories abstract data access for testability. All repositories use sync.RWMutex for thread safety."

### 10-Minute Deep Dive
> "**Architecture**: Layered—models, interfaces, services, repositories, strategies. Services orchestrate; repositories persist; strategies encapsulate algorithms.
>
> **Checkout Flow**: Ten steps: validate cart, validate stock, subtotal, apply coupon (with expiry/min-order/usage checks), tax, payment, decrement stock, create order, clear cart, return summary. Each step can fail independently; we return early on validation errors.
>
> **Discount Strategy**: Coupon has Type (percentage/flat/BOGO). Registry returns the right strategy. Each strategy implements Calculate(ctx). BOGO gives free item per pair; percentage and flat are straightforward.
>
> **Thread Safety**: Repositories use RWMutex. Read-heavy workloads benefit from RLock. Checkout uses Mutex for the full flow to avoid race between stock check and decrement.
>
> **Cart Persistence**: One active cart per user. userCarts map links userID to cartID. On checkout, we create a new empty cart and update the mapping. Old cart stays with status CheckedOut for history.
>
> **Testing**: Unit tests for cart add/update/remove, checkout success/empty/stock validation, coupon validation for all types. Concurrent add test verifies thread safety."

---

## 8. Future Improvements

- **Persistence**: Replace in-memory repos with PostgreSQL/Redis
- **API Layer**: HTTP handlers (Gin/Echo) with REST endpoints
- **Authentication**: JWT/session-based auth, middleware
- **Event Sourcing**: Publish cart/order events to Kafka for analytics
- **Idempotency**: Idempotency keys for checkout to prevent double charges
- **Inventory Service**: Separate service for stock, with eventual consistency
- **Shipping**: Shipping cost calculation, address validation
- **Refunds**: Order cancellation, partial refunds
- **Caching**: Redis for product catalog, cart
- **Rate Limiting**: Per-user checkout rate limits

---

## 9. Running the Project

```bash
# Build
go build ./...

# Run demo
go run ./cmd

# Run tests
go test ./tests/... -v
```

### Sample Output
```
=== Shopping Cart System Demo ===

1. Adding items to cart...
   Cart subtotal: $2029.97 (3 items)

2. Checkout with coupon SAVE10...
   Order ID: ord_...
   Subtotal: $2029.97
   Discount: $203.00
   Tax: $328.86
   Total: $2155.83
   Payment ID: cc_...

3. Cart after checkout: 0 items (empty)
4. Order history: 1 order(s)
=== Demo Complete ===
```

---

## 10. Directory Structure

```
16-shopping-cart-system/
├── cmd/main.go                 # Entry point, wiring, demo
├── internal/
│   ├── models/                 # Domain entities
│   ├── interfaces/             # Repository & strategy interfaces
│   ├── services/               # Business logic
│   ├── repositories/           # In-memory implementations
│   └── strategies/             # Discount, tax, payment strategies
├── tests/                      # Unit tests
├── go.mod
└── README.md
```
