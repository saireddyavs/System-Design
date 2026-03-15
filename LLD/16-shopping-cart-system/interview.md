# Shopping Cart System — Interview Strategy (45 min)

## Time Allocation

| Phase | Time | What to Do |
|-------|------|------------|
| 1. Clarify & Scope | 3 min | Confirm: cart per user, checkout flow, discount types (%, flat, BOGO), tax, payment |
| 2. Core Models | 7 min | Product, User, Cart, CartItem, Order, Coupon; CartItem.RecalculateSubtotal |
| 3. Repository Interfaces | 5 min | CartRepository, ProductRepository, OrderRepository, CouponRepository, UserRepository |
| 4. Service Interfaces | 5 min | DiscountStrategy, TaxCalculator, PaymentProcessor; CheckoutSummaryBuilder |
| 5. Core Service Implementation | 12 min | CheckoutService.Checkout (validate → subtotal → coupon → tax → payment → order) |
| 6. main.go Wiring | 5 min | Wire repos, discount registry, tax calc, payment registry; demo add-to-cart → checkout |
| 7. Extend & Discuss | 8 min | Discount chain, coupon validation, stock decrement; idempotency, inventory service |

## Phase 1: Clarify & Scope (3 min)

**Questions to ask:**
- One cart per user or multiple?
- Which discount types? (Percentage, flat, BOGO?)
- Tax: flat rate or state-based?
- Stock validation: at add-to-cart or checkout only?

**Scope in:** Add/update/remove cart items, checkout with coupon, tax, payment, order creation, cart clear.

**Scope out:** Shipping cost, refunds, wishlist, product recommendations.

## Phase 2: Core Models (7 min)

**Write first:** `Cart` (ID, UserID, Items []CartItem, Status), `CartItem` (ProductID, Quantity, UnitPrice, Subtotal with `RecalculateSubtotal()`), `Product` (ID, Name, Price, Stock), `Coupon` (Code, Type, Value, MinOrderAmount, ExpiresAt).

**Essential fields:**
- `Cart`: Items slice; `Subtotal()` = sum of item subtotals
- `Coupon`: Type (percentage/flat/bogo), Value, MinOrderAmount, ExpiresAt
- `Order`: Items, Subtotal, Discount, Tax, Total, PaymentID

**Skip:** Cart abandonment, addresses (single address field ok), order history pagination.

## Phase 3: Repository Interfaces (5 min)

```go
type CartRepository interface {
    GetByUserID(userID string) (*Cart, error)
    Save(cart *Cart) error
}
type ProductRepository interface { GetByID(id string) (*Product, error); DecrementStock(id string, qty int) error }
type CouponRepository interface { GetByCode(code string) (*Coupon, error); Update(c *Coupon) error }
type OrderRepository interface { Create(o *Order) error; GetByUserID(userID string) ([]*Order, error) }
```

Cart: one active cart per user; on checkout, create new empty cart.

## Phase 4: Service Interfaces (5 min)

```go
type DiscountStrategy interface {
    Supports(couponType CouponType) bool
    Calculate(ctx *DiscountContext) float64
}
type TaxCalculator interface { Calculate(subtotal, discount float64) float64 }
type PaymentProcessor interface { Process(payment *Payment) error; GetMethod() string }
type CheckoutSummaryBuilder interface { ... }  // Optional: fluent builder for summary
```

DiscountContext: Coupon, Items, Subtotal. Registry returns strategy by CouponType.

## Phase 5: Core Service Implementation (12 min)

**THE most important method: `CheckoutService.Checkout(userID, couponCode, paymentMethod)`**

1. Get cart; validate not empty, status Active
2. Validate stock for all items
3. Subtotal = cart.Subtotal()
4. CouponService.ValidateAndGetDiscount(couponCode, subtotal, items) — expiry, min order, usage; registry.GetStrategy(coupon.Type).Calculate(ctx)
5. Tax = taxCalculator.Calculate(subtotal - discount)
6. Total = subtotal - discount + tax
7. Process payment
8. Decrement stock for each item
9. Create order from cart
10. Clear cart (new empty cart for user)
11. Return CheckoutSummary (OrderID, Subtotal, Discount, Tax, Total, PaymentID)

**Why:** End-to-end flow; touches validation, discount strategy, tax, payment, stock, order. Demonstrates Strategy + Repository.

**Discount strategies:** Percentage = subtotal * (Value/100); Flat = min(Value, subtotal); BOGO = free cheapest item per pair.

## Phase 6: main.go Wiring (5 min)

- Repos: Cart, Product, Order, Coupon, User
- DiscountStrategyRegistry with Percentage, Flat, BOGO
- FlatTaxCalculator(0.18)
- PaymentProcessorRegistry with CreditCard
- CheckoutService(cartRepo, orderRepo, productRepo, couponRepo, userRepo, cartService, couponService, orderService, taxCalc, paymentRegistry)
- Demo: AddItem x2, Checkout with "SAVE10", print summary

## Phase 7: Extend & Discuss (8 min)

- **Coupon validation:** ExpiresAt, MinOrderAmount, MaxUsageLimit, CurrentUsage
- **BOGO:** For every 2 items, discount = price of cheapest; handle mixed cart
- **Stock race:** Lock during checkout; or optimistic with version check
- **Idempotency:** Checkout idempotency key to prevent double charge on retry

## Tips

- Cart items: linear scan for ProductID match on add/update; O(n) acceptable for typical cart size
- HashMap for cart lookup: userCarts[userID] = cartID; carts[cartID] = cart
- If time short, implement only Percentage discount; mention Flat/BOGO as extensions
