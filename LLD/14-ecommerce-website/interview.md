# E-commerce Website — Interview Strategy (45 min)

## Time Allocation

| Phase | Time | What to Do |
|-------|------|------------|
| 1. Clarify & Scope | 3 min | Confirm cart, order flow, payment methods, inventory, coupons; scope out auth, recommendations |
| 2. Core Models | 7 min | Product, Cart, CartItem, Order, OrderItem, Payment, Coupon, User |
| 3. Repository Interfaces | 5 min | ProductRepository (with DecrementStock), CartRepository, OrderRepository, CouponRepository, PaymentRepository |
| 4. Service Interfaces | 5 min | PaymentProcessor (Strategy), DiscountStrategy (Strategy), NotificationService (Observer) |
| 5. Core Service Implementation | 12 min | OrderService.PlaceOrder() — validate stock, apply coupon, decrement inventory, process payment, persist, clear cart; rollback on failure |
| 6. main.go Wiring | 5 min | Repos, payment processors, discount strategies, OrderFactory, observers |
| 7. Extend & Discuss | 8 min | Order lifecycle, coupon types (%, flat, BOGO), inventory race conditions |

## Phase 1: Clarify & Scope (3 min)

**Questions to ask:**
- Cart model? → Per user; ProductID → CartItem (quantity, price)
- Order from cart? → Place order converts cart to order; clears cart on success
- Payment methods? → Credit, Debit, UPI, Wallet — Strategy pattern
- Inventory? → Decrement on order confirmation; validate before placement
- Coupons? → Percentage, flat, BOGO; validate expiry, min order, usage limit
- Order lifecycle? → Placed → Confirmed → Shipped → Delivered; Cancelled/Returned for rollback

**Scope out:** Auth/JWT, product recommendations, cart abandonment, async payment webhooks.

## Phase 2: Core Models (7 min)

**Start with:**
- `Product`: ID, Name, Description, Price, CategoryID, Stock, SKU
- `Cart`: ID, UserID, Items `map[string]CartItem`, UpdatedAt
- `CartItem`: ProductID, Quantity, Price
- `Order`: ID, UserID, Items `[]OrderItem`, TotalAmount, Discount, FinalAmount, Status, ShippingAddress, PaymentID
- `OrderItem`: ProductID, Quantity, Price
- `Payment`: ID, OrderID, Amount, Method, Status, TransactionID
- `Coupon`: ID, Code, Type (Percentage/Flat/BOGO), Value, MinOrderAmount, ExpiresAt, UsageLimit, UsedCount

**Order status:** Placed, Confirmed, Shipped, Delivered, Cancelled, Returned

**Skip for now:** Category hierarchy, Product images/ratings, Address model details.

## Phase 3: Repository Interfaces (5 min)

**Essential:**
- `ProductRepository`: GetByID, Create, Update, **DecrementStock**, **IncrementStock** — atomic for inventory
- `CartRepository`: GetByUserID, Create, Update
- `OrderRepository`: Create, GetByID, GetByUserID, Update, UpdateStatus
- `CouponRepository`: GetByCode, Create, IncrementUsage
- `PaymentRepository`: Create, GetByID, Update

**Key abstraction:** ProductRepository.DecrementStock must be atomic (sync.Mutex or DB transaction); IncrementStock for cancel/return rollback.

## Phase 4: Service Interfaces (5 min)

**Essential:**
- `PaymentProcessor`: ProcessPayment(ctx, amount, method, ref) — Strategy for Credit/Debit/UPI/Wallet
- `DiscountStrategy`: GetType() CouponType, Calculate(coupon, orderAmount, quantity) float64 — Percentage, Flat, BOGO
- `NotificationService`: NotifyOrderStatus(ctx, orderID, userID, status) — Observer for email/SMS

**Key abstraction:** PaymentProcessor and DiscountStrategy are Strategy pattern; add UPI or BOGO without touching OrderService.

## Phase 5: Core Service Implementation (12 min)

**Key method:** `OrderService.PlaceOrder(ctx, input)` — this is where the core logic lives.

**Flow:**
1. Get cart; fail if empty
2. **Validate stock** for all items: `product.Stock >= item.Quantity` for each cart item
3. Compute totalAmount (sum of item.Price × item.Quantity)
4. **Apply coupon** (if code provided): `couponService.ApplyCoupon(code, totalAmount, totalQty)` → discount, coupon; validates expiry, min amount, usage limit; uses DiscountStrategy.Calculate
5. Create Payment record (Pending)
6. Create Order via OrderFactory (items, totals, discount, paymentID)
7. **Decrement stock** for each item: `productRepo.DecrementStock(productID, qty)` — **must succeed**; atomic
8. **Process payment:** `paymentService.ProcessPayment(payment)` — **rollback stock** on failure
9. Persist payment; **rollback stock** if payment repo Create fails
10. Increment coupon usage if applied
11. Persist order, UpdateStatus(Confirmed)
12. Clear cart
13. Notify observers: `orderObserver.NotifyOrderStatus(Confirmed)`

**CancelOrder(orderID):**
- Only if status is Placed or Confirmed
- IncrementStock for each order item
- Update status to Cancelled, notify

**Concurrency:** DecrementStock uses `sync.Mutex` (or DB row lock); concurrent orders must not oversell.

## Phase 6: main.go Wiring (5 min)

```go
productRepo := repositories.NewInMemoryProductRepo()
cartRepo := repositories.NewInMemoryCartRepo()
orderRepo := repositories.NewInMemoryOrderRepo()
paymentRepo := repositories.NewInMemoryPaymentRepo()
couponRepo := repositories.NewInMemoryCouponRepo()

paymentProcessors := []interfaces.PaymentProcessor{
    strategies.NewCreditCardProcessor(),
    strategies.NewDebitCardProcessor(),
    strategies.NewUPIProcessor(),
    strategies.NewWalletProcessor(),
}
paymentService := services.NewPaymentService(paymentProcessors)

discountStrategies := []interfaces.DiscountStrategy{
    strategies.NewPercentageDiscountStrategy(),
    strategies.NewFlatDiscountStrategy(),
    strategies.NewBOGODiscountStrategy(),
}
couponService := services.NewCouponService(couponRepo, discountStrategies)

orderFactory := factory.NewOrderFactory(idGen)
orderObserver := observer.NewOrderStatusObserver()
orderObserver.Subscribe(services.NewLoggingNotificationService())

orderService := services.NewOrderService(
    orderRepo, cartRepo, productRepo, paymentRepo, couponRepo,
    orderFactory, paymentService, couponService, orderObserver, idGen,
)
cartService := services.NewCartService(cartRepo, productRepo, idGen)
```

Show: Payment processors and discount strategies injected; OrderFactory for order creation; observer for notifications.

## Phase 7: Extend & Discuss (8 min)

**Design patterns to mention:**
- **Strategy**: PaymentProcessor (Credit/Debit/UPI/Wallet), DiscountStrategy (Percentage/Flat/BOGO)
- **Observer**: OrderStatusObserver — email, SMS, analytics without touching OrderService
- **Factory**: OrderFactory — complex order construction from cart
- **Repository**: DecrementStock/IncrementStock for inventory; abstract persistence

**Extensions:**
- Order lifecycle: Placed → Confirmed (payment + stock) → Shipped → Delivered; Cancelled/Returned restore stock
- Coupon types: Percentage (10% off), Flat ($20 off), BOGO (buy one get one)
- Inventory race: Mutex in DecrementStock; mention optimistic locking (version field) or Redis lock for distributed

## Tips

- **Prioritize if low on time:** PlaceOrder flow (validate stock → coupon → decrement → pay → persist → clear cart + rollback), CartService.AddItem, CouponService.ApplyCoupon. Skip Observer, ProductBuilder.
- **Common mistakes:** Decrementing stock before payment (risk of payment failure with depleted stock); not rolling back stock on payment failure; forgetting coupon validation (expiry, min amount, usage limit).
- **What impresses:** Explicit rollback on each failure step; DiscountStrategy for coupon types; PaymentProcessor Strategy; atomic DecrementStock; clear order lifecycle.
