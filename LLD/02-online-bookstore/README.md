# Online Bookstore - Low Level Design (LLD)

A production-quality, interview-ready implementation of an Online Bookstore in Go. Follows clean architecture and SOLID principles with design patterns for extensibility.

## 1. Problem Description

### Functional Requirements
- **Book Management**: CRUD for books with title, author, ISBN, price, genre, stock
- **User Management**: Registration and authentication
- **Search**: Search books by title, author, genre (efficient, case-insensitive)
- **Shopping Cart**: Add, remove, update quantities; validate stock
- **Order Placement**: Place orders with payment processing
- **Inventory Management**: Stock tracking, restocking, low-stock notifications
- **Order History**: Track and retrieve user order history

### Non-Functional Requirements
- **Thread Safety**: Concurrent access via `sync.RWMutex`
- **Extensibility**: New payment methods, search backends without modifying core logic
- **Testability**: Dependency injection for unit testing
- **Maintainability**: Clear separation of concerns, idiomatic Go

---

## 2. Core Entities & Relationships

```
┌─────────┐     has      ┌─────────┐     contains     ┌──────────┐
│  User   │─────────────▶│  Cart   │────────────────▶│ CartItem │
└─────────┘              └─────────┘   (BookID→Qty)  └──────────┘
     │                         │                            │
     │ places                  │                            │ references
     ▼                         │                            ▼
┌─────────┐                    │                     ┌─────────┐
│  Order  │◀────────────────────┘                     │  Book   │
└─────────┘   (from cart items)                      └─────────┘
     │
     │ has
     ▼
┌─────────┐
│ Payment │
└─────────┘
```

| Entity | Key Fields | Relationships |
|--------|------------|---------------|
| **Book** | ID, Title, Author, ISBN, Price, Genre, Stock | Referenced by Cart, Order |
| **User** | ID, Name, Email, Password, Address | Owns Cart, places Orders |
| **Cart** | ID, UserID, Items (map) | Belongs to User, contains Books |
| **Order** | ID, UserID, Items, TotalAmount, Status, PaymentMethod | Belongs to User, has Payment |
| **Payment** | ID, OrderID, Amount, Method, Status, TransactionID | Linked to Order |

---

## 3. Design Patterns & WHY

### Repository Pattern
**Where**: `interfaces/book_repository.go`, `repositories/in_memory_book_repo.go`  
**Why**: Abstracts data access. Swap in-memory for PostgreSQL/MySQL without changing services. Enables testing with mocks. **DIP** – services depend on interfaces.

### Strategy Pattern
**Where**: `interfaces/payment_processor.go`, `strategies/payment_strategies.go`  
**Why**: Multiple payment methods (Credit Card, etc.) without `switch`/`if-else`. **OCP** – add new methods by implementing `PaymentProcessor` and registering. No changes to `OrderService`.

### Observer Pattern
**Where**: `interfaces/inventory_observer.go`, `services/inventory_service.go`, `services/low_stock_observer.go`  
**Why**: Decouples inventory from notification logic. Low-stock alerts can go to log, email, Slack, etc. Add observers without modifying `InventoryService`.

### Factory Pattern
**Where**: `strategies/order_factory.go`  
**Why**: Centralizes order creation with validation, ID generation, and item aggregation. Keeps `OrderService` focused on orchestration.

---

## 4. SOLID Principles Mapping

| Principle | Application |
|-----------|-------------|
| **SRP** | `BookService` manages books only; `CartService` manages carts; `OrderService` orchestrates orders. Each service has one reason to change. |
| **OCP** | New payment methods via `PaymentProcessor`; new search backends via `SearchEngine`. Extend without modifying existing code. |
| **LSP** | `InMemoryBookRepository` substitutes `BookRepository`; `CreditCardProcessor` substitutes `PaymentProcessor`. Implementations are interchangeable. |
| **ISP** | `BookRepository`, `OrderRepository`, `SearchEngine`, `PaymentProcessor` are small, focused interfaces. No client depends on methods it doesn't use. |
| **DIP** | Services depend on `interfaces.BookRepository`, not `repositories.InMemoryBookRepository`. High-level modules don't depend on low-level modules. |

---

## 5. Search Algorithm Explanation

**Implementation**: Case-insensitive substring matching via `strings.Contains(strings.ToLower(s), strings.ToLower(substr))`.

**Complexity**:
- Per book: O(n + m) where n = field length, m = query length
- Total: O(B × (n + m)) for B books

**Efficiency**:
- In-memory: Linear scan over all books. Suitable for < 10K books.
- Production: Replace `InMemorySearchEngine` with Elasticsearch/Meilisearch for:
  - Inverted index: O(log N) lookup
  - Full-text search, fuzzy matching, relevance scoring

**Extensibility**: `SearchEngine` interface allows swapping implementations (e.g., Elasticsearch) without changing callers. Use `SearchEngine` directly; no separate SearchService.

---

## 6. Concurrency Considerations

- **sync.RWMutex** on all in-memory repositories: multiple readers OR single writer
- **Read-heavy workloads**: `RLock()` for `GetByID`, `GetAll`, `Search`; `Lock()` for `Create`, `Update`, `Delete`
- **Order placement**: Sequential stock updates; for high concurrency, consider optimistic locking or distributed locks
- **Cart updates**: Single cart per user; mutex prevents lost updates on concurrent `AddToCart`
- **Future**: Connection pooling, prepared statements for DB; rate limiting for payment gateway

---

## 7. Interview Explanation

### 3-Minute Version
> "This is an Online Bookstore with books, users, carts, and orders. I used **Repository** to abstract data access so we can swap in-memory for a database. **Strategy** handles payment methods (Credit Card via registry) so adding new methods doesn't touch order logic. **Observer** notifies when stock is low. **Factory** creates orders from cart items. Services depend on interfaces (**DIP**), and each service has one job (**SRP**). Search uses `SearchEngine` directly with case-insensitive substring matching; for scale, we'd use Elasticsearch. All repos use `RWMutex` for thread safety."

### 10-Minute Version
> "The system has five core entities: Book, User, Cart, Order, Payment. Users register and authenticate. They search books by title, author, or genre using a `SearchEngine` interface—currently in-memory with substring matching, but we can plug in Elasticsearch.
>
> Cart and Order services orchestrate the flow. When placing an order, we use an **OrderFactory** to build the order from cart items, then a **PaymentProcessor** strategy (Credit Card via registry) processes payment. Stock is decremented, cart cleared.
>
> **Design patterns**: Repository for data abstraction, Strategy for payments, Observer for low-stock alerts, Factory for order creation. **SOLID**: SRP in services, OCP for payments/search, LSP for repository implementations, ISP with small interfaces, DIP throughout.
>
> Concurrency: `RWMutex` on all in-memory stores. For production, we'd add DB transactions, idempotency for payments, and consider event sourcing for order history."

---

## 8. Future Improvements

1. **Persistence**: PostgreSQL/MySQL with migrations (e.g., golang-migrate)
2. **API Layer**: REST/GraphQL with chi or gin; JWT authentication
3. **Search**: Elasticsearch/Meilisearch for full-text search
4. **Caching**: Redis for hot books, cart session
5. **Payment**: Stripe/PayPal SDK integration; webhook handling
6. **Events**: Order placed → inventory update, email notification (event-driven)
7. **Testing**: Integration tests, table-driven tests, mocks with gomock
8. **Observability**: Structured logging, metrics (Prometheus), tracing (OpenTelemetry)

---

## Data Structures & Algorithms

| DS/Algorithm | Where Used | Why | Alternatives/Tradeoffs |
|-------------|------------|-----|------------------------|
| **HashMap** | `InMemoryBookRepository.books`, `InMemoryOrderRepository`, `InMemoryCartRepository` | O(1) lookup by ID; repositories use map for CRUD | Alternative: B-tree for range queries; HashMap sufficient for ID-based access |
| **String matching** | `InMemorySearchEngine.Search()`, `containsIgnoreCase()` | Case-insensitive substring search via `strings.Contains(strings.ToLower(s), strings.ToLower(substr))` | Production: Elasticsearch/Meilisearch for O(log N) full-text; in-memory linear scan OK for <10K books |
| **Registry pattern** | `PaymentProcessorRegistry.processors` (map[string]PaymentProcessor) | O(1) payment method lookup; add new methods without modifying OrderService | Strategy + Registry = Open/Closed; alternative: factory with switch |
| **sync.RWMutex** | All in-memory repositories | Read-heavy (GetByID, Search) use RLock; Create/Update/Delete use Lock | Per-entity locking could improve concurrency; RWMutex simpler for LLD scope |

---

## Quick Start

```bash
# Build
go build ./...

# Run demo
go run ./cmd/main.go

# Run tests
go test ./tests/... -v
```

## Directory Structure

```
02-online-bookstore/
├── cmd/main.go              # Wire-up and demo
├── internal/
│   ├── models/              # Book, User, Cart, Order, Payment
│   ├── interfaces/          # Repository, SearchEngine, PaymentProcessor, Observer
│   ├── repositories/        # In-memory implementations
│   ├── services/            # Business logic
│   └── strategies/          # Payment, OrderFactory
├── tests/                   # Unit tests
├── go.mod
└── README.md
```
