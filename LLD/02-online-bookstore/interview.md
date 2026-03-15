# Online Bookstore — Interview Strategy (45 min)

## Time Allocation

| Phase | Time | What to Do |
|-------|------|------------|
| 1. Clarify & Scope | 3 min | Confirm books, users, cart, orders, search, payment; scope out reviews, recommendations |
| 2. Core Models | 7 min | Book, User, Cart (Items map), Order, Payment |
| 3. Repository Interfaces | 5 min | BookRepository, UserRepository, OrderRepository, CartRepository |
| 4. Service Interfaces | 5 min | SearchEngine, PaymentProcessor, PaymentProcessorRegistry |
| 5. Core Service Implementation | 12 min | OrderService.PlaceOrder() — cart→order, payment, stock decrement |
| 6. Handler / main.go Wiring | 5 min | Wire repos, SearchEngine, PaymentRegistry, OrderService, CartService |
| 7. Extend & Discuss | 8 min | Observer for low-stock, Strategy for payment, SearchEngine swap |

## Phase 1: Clarify & Scope (3 min)

**Questions to ask:**
- Search by what? → Title, author, genre (case-insensitive)
- Payment methods? → At least one (e.g., Credit Card); extensible
- Cart per user? → Yes, one active cart
- Stock validation? → On add-to-cart and checkout
- Low-stock alerts? → Observer pattern (optional)

**Scope out:** Reviews, recommendations, wishlist, admin panel.

## Phase 2: Core Models (7 min)

**Start with:**
- `Book`: ID, Title, Author, ISBN, Price, Genre, Stock
- `User`: ID, Name, Email, Password (hash)
- `Cart`: ID, UserID, Items map[BookID]int (quantity)

**Then:**
- `Order`: ID, UserID, Items, TotalAmount, Status, PaymentMethod
- `Payment`: ID, OrderID, Amount, Method, Status, TransactionID

**Skip for now:** Address, order history pagination, coupon.

## Phase 3: Repository Interfaces (5 min)

**Essential:**
- `BookRepository`: Create, GetByID, GetByISBN, Update, UpdateStock, GetAll
- `UserRepository`: Create, GetByID, GetByEmail
- `CartRepository`: GetByUserID, Create, Update
- `OrderRepository`: Create, GetByID, GetByUserID

**Skip:** PaymentRepository (can live in OrderService or separate).

## Phase 4: Service Interfaces (5 min)

**Essential:**
- `SearchEngine`: Search(query, searchType) — Title, Author, Genre, All
- `PaymentProcessor`: Process(payment), GetMethodName()
- `PaymentProcessorRegistry`: Register(processor), GetProcessor(method)

**Key abstraction:** Registry lets you add new payment methods without touching OrderService.

## Phase 5: Core Service Implementation (12 min)

**Key method:** `OrderService.PlaceOrder(userID, cartID, paymentMethod)` — this is where the core logic lives.

**Flow:**
1. Get cart by userID; validate not empty
2. For each cart item: validate stock >= quantity
3. Calculate total from cart items
4. Get processor from registry: `registry.GetProcessor("credit_card")`
5. Create Payment, call `processor.Process(payment)`
6. Decrement stock for each item (rollback on failure)
7. Create Order from cart items
8. Clear cart (or mark checked out)
9. Return order

**Search:** `SearchEngine.Search(query, type)` — linear scan over books, `strings.Contains(strings.ToLower(field), strings.ToLower(query))` for case-insensitive match.

**Concurrency:** RWMutex on all repos; Lock for Create/Update, RLock for Get.

## Phase 6: main.go Wiring (5 min)

```go
bookRepo := repositories.NewInMemoryBookRepository()
userRepo := repositories.NewInMemoryUserRepository()
cartRepo := repositories.NewInMemoryCartRepository()
orderRepo := repositories.NewInMemoryOrderRepository()

searchEngine := repositories.NewInMemorySearchEngine(bookRepo)
registry := strategies.NewPaymentProcessorRegistry()
registry.Register(strategies.NewCreditCardProcessor())

orderSvc := services.NewOrderService(orderRepo, cartRepo, bookRepo, registry, ...)
cartSvc := services.NewCartService(cartRepo, bookRepo, ...)
bookSvc := services.NewBookService(bookRepo, searchEngine)
```

## Phase 7: Extend & Discuss (8 min)

**Design patterns to mention:**
- Repository (data access)
- Strategy + Registry (payment methods)
- Observer (low-stock notifications)
- Factory (OrderFactory for order creation)

**Extensions:**
- Elasticsearch for SearchEngine
- Multiple payment methods (PayPal, UPI)
- Coupon/discount at checkout
- Idempotency for payment retries

## Tips

- **Prioritize if low on time:** PlaceOrder flow, cart→order, payment processing. Skip search implementation details.
- **Common mistakes:** Not validating stock before payment; not clearing cart; fat OrderService.
- **What impresses:** Registry for payment methods (OCP), SearchEngine interface for swap, Observer for low-stock.
