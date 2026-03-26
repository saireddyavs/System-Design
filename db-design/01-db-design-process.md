# Database Design Process — Interview Guide

## How to Approach Any DB Design Question in an Interview

---

## 1. The 6-Step DB Design Framework

When an interviewer asks you to "design the database for X", follow this systematic approach:

```
Step 1: Clarify Requirements
         │
         ▼
Step 2: Identify Entities
         │
         ▼
Step 3: Define Relationships (ER Diagram)
         │
         ▼
Step 4: Choose Primary Keys & Define Attributes
         │
         ▼
Step 5: Normalize (3NF minimum)
         │
         ▼
Step 6: Create Tables (DDL) & Validate with Queries
```

---

## 2. Step 1 — Clarify Requirements

### What to Ask the Interviewer

Before touching any schema, always clarify:

| Question Type | Example Questions |
|---------------|-------------------|
| **Scale** | How many users? How many transactions/day? |
| **Access Patterns** | What are the most common queries? Read-heavy or write-heavy? |
| **Consistency** | Do we need strong consistency? Financial data? |
| **Features** | What features are in scope? (e.g., search, recommendations, analytics) |
| **Constraints** | Any regulatory requirements (GDPR, PCI-DSS)? |

### Example: "Design a database for an e-commerce platform"

**Clarifying Questions:**
- Do we need to track inventory in real-time?
- Do we support multiple payment methods?
- Is there a seller marketplace, or single vendor?
- Do we need to support product reviews and ratings?
- What's the expected scale (users, orders/day)?

---

## 3. Step 2 — Identify Entities

### How to Find Entities

**Entities = Nouns** in the problem description. Look for:

1. **Core objects** — the main "things" the system manages
2. **Supporting objects** — auxiliary data needed by core objects
3. **Junction/Bridge entities** — represent many-to-many relationships
4. **Metadata entities** — audit trails, logs, configs

### Example: E-Commerce Platform

```
Core Entities:
├── User
├── Product
├── Order
├── Payment
└── Category

Supporting Entities:
├── Address
├── Cart / CartItem
├── ProductImage
├── Review / Rating
└── Seller (if marketplace)

Junction Entities:
├── OrderItem (Order ↔ Product)
├── ProductCategory (Product ↔ Category)
└── Wishlist (User ↔ Product)

Metadata Entities:
├── AuditLog
├── Coupon / Discount
└── ShippingInfo
```

### Example: Social Media Platform

```
Core Entities:
├── User
├── Post
├── Comment
└── Message

Supporting Entities:
├── UserProfile
├── Media (photos, videos)
├── Hashtag
└── Notification

Junction Entities:
├── Friendship (User ↔ User)
├── Like (User ↔ Post)
├── PostHashtag (Post ↔ Hashtag)
└── Follow (User ↔ User)
```

### Example: Hotel Booking System

```
Core Entities:
├── Guest
├── Hotel
├── Room
├── Booking
└── Payment

Supporting Entities:
├── RoomType
├── Amenity
├── Review
└── HotelImage

Junction Entities:
├── RoomAmenity (Room ↔ Amenity)
└── BookingRoom (Booking ↔ Room, for multi-room bookings)
```

---

## 4. Step 3 — Define Relationships (ER Diagram)

### Relationship Types

| Relationship | Symbol | Example | Implementation |
|-------------|--------|---------|----------------|
| **One-to-One (1:1)** | 1 ── 1 | User ↔ UserProfile | FK in either table (or merge) |
| **One-to-Many (1:N)** | 1 ── N | User ↔ Orders | FK in the "many" side |
| **Many-to-Many (M:N)** | M ── N | Student ↔ Course | Junction/bridge table |

### ER Diagram Example: E-Commerce

```
┌──────────┐        ┌───────────┐        ┌──────────┐
│   USER   │1──────N│   ORDER   │1──────N│ ORDER_   │
│          │        │           │        │   ITEM   │
│ id (PK)  │        │ id (PK)   │        │ id (PK)  │
│ name     │        │ user_id   │◀──FK   │ order_id │◀──FK
│ email    │        │ status    │        │ product_id│◀──FK
│ password │        │ total     │        │ quantity │
│ created  │        │ created   │        │ price    │
└──────────┘        └───────────┘        └──────────┘
      │                    │
      │1                   │1
      │                    │
      N                    │
┌──────────┐               │
│ ADDRESS  │               │
│ id (PK)  │               N
│ user_id  │◀──FK    ┌──────────┐
│ street   │         │ PAYMENT  │
│ city     │         │ id (PK)  │
│ zip      │         │ order_id │◀──FK
│ country  │         │ method   │
└──────────┘         │ amount   │
                     │ status   │
┌──────────┐         └──────────┘
│ PRODUCT  │
│ id (PK)  │1──────N  ┌──────────┐
│ name     │          │ REVIEW   │
│ price    │          │ id (PK)  │
│ stock    │          │ user_id  │◀──FK
│ seller_id│          │ prod_id  │◀──FK
│ category │          │ rating   │
└──────────┘          │ comment  │
      │               └──────────┘
      M
      │
      N
┌──────────────┐
│ PRODUCT_     │
│  CATEGORY    │
│ product_id   │◀──FK (composite PK)
│ category_id  │◀──FK
└──────────────┘
```

### Cardinality Rules

| Pattern | When to Use |
|---------|-------------|
| **1:1** | User → Profile; split for performance or access pattern separation |
| **1:N** | User → Orders; parent → children |
| **M:N** | Students → Courses; always needs junction table |
| **Self-referencing** | Employee → Manager (same table); Categories → SubCategories |

### Self-Referencing Relationship Example

```sql
-- Employee-Manager hierarchy
CREATE TABLE employees (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    manager_id  INTEGER REFERENCES employees(id),
    department  VARCHAR(50),
    salary      DECIMAL(10,2)
);

-- Category hierarchy
CREATE TABLE categories (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    parent_id   INTEGER REFERENCES categories(id)
);
```

---

## 5. Step 4 — Choose Primary Keys & Define Attributes

### Primary Key Selection

| Strategy | Pros | Cons | Use When |
|----------|------|------|----------|
| **Auto-increment (SERIAL)** | Simple, sequential | Single DB bottleneck, predictable | Single database, internal IDs |
| **UUID (v4)** | No collision, distributed-safe | 128-bit, no ordering, index bloat | Distributed systems, external IDs |
| **UUID (v7)** | Time-ordered UUID | Newer, less support | Need distributed + time ordering |
| **Composite Key** | Natural, no extra column | Complex joins, harder references | Junction tables, time-series |
| **Natural Key** | Meaningful | May change, encoding issues | ISO codes, stable identifiers |

### Attribute Best Practices

```sql
-- ✅ GOOD: Well-typed, constrained attributes
CREATE TABLE users (
    id              SERIAL PRIMARY KEY,
    email           VARCHAR(255) NOT NULL UNIQUE,
    username        VARCHAR(50) NOT NULL UNIQUE,
    password_hash   CHAR(60) NOT NULL,              -- bcrypt always 60 chars
    full_name       VARCHAR(200) NOT NULL,
    phone           VARCHAR(20),                     -- nullable, not everyone has
    status          VARCHAR(20) NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active', 'inactive', 'banned', 'deleted')),
    email_verified  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ❌ BAD: No constraints, wrong types
CREATE TABLE users (
    id      INT,
    email   TEXT,        -- no unique, no not null
    name    TEXT,
    phone   INT,         -- phone has leading zeros, plus signs
    status  TEXT,        -- no check constraint
    created TEXT         -- storing timestamp as text
);
```

### Common Column Types Guide

| Data | Use | Don't Use |
|------|-----|-----------|
| Money/currency | `DECIMAL(10,2)` or `NUMERIC` | `FLOAT`, `DOUBLE` (precision loss) |
| Timestamps | `TIMESTAMP WITH TIME ZONE` | `VARCHAR`, `DATE` for datetimes |
| Boolean flags | `BOOLEAN` | `INT`, `CHAR(1)` |
| Email | `VARCHAR(255)` with CHECK | `TEXT` without constraint |
| Phone | `VARCHAR(20)` | `INTEGER` (leading zeros, +) |
| Status/enum | `VARCHAR` with CHECK or ENUM | `INT` (magic numbers) |
| Large text | `TEXT` | `VARCHAR(10000)` |
| IP address | `INET` (PostgreSQL) | `VARCHAR` |
| JSON data | `JSONB` (PostgreSQL) | `TEXT` |

---

## 6. Step 5 — Normalization

### Why Normalize?

Normalization eliminates **data redundancy** and **update anomalies**:

| Anomaly | Problem | Example |
|---------|---------|---------|
| **Insert anomaly** | Can't insert data without other data | Can't add a department without an employee |
| **Update anomaly** | Update in one place, stale elsewhere | Change dept name → must update all rows |
| **Delete anomaly** | Deleting data loses other data | Delete last employee → lose department info |

### Normal Forms Explained with Examples

#### 1NF — First Normal Form
**Rule**: All values must be atomic (no repeating groups, no arrays in cells)

```
❌ VIOLATES 1NF:
┌────────┬───────────────────────┐
│ student│ courses               │
├────────┼───────────────────────┤
│ Alice  │ Math, Physics, Chem   │  ← multiple values in one cell
│ Bob    │ Math, English         │
└────────┴───────────────────────┘

✅ IN 1NF:
┌────────┬─────────┐
│ student│ course  │
├────────┼─────────┤
│ Alice  │ Math    │
│ Alice  │ Physics │
│ Alice  │ Chem    │
│ Bob    │ Math    │
│ Bob    │ English │
└────────┴─────────┘
```

#### 2NF — Second Normal Form
**Rule**: 1NF + no partial dependencies (every non-key column must depend on the **entire** primary key)

Only applies when you have a **composite primary key**.

```
❌ VIOLATES 2NF (composite PK: student_id + course_id):
┌────────────┬───────────┬──────────────┬───────────────┐
│ student_id │ course_id │ student_name │ course_name   │
├────────────┼───────────┼──────────────┼───────────────┤
│ 1          │ 101       │ Alice        │ Mathematics   │
│ 1          │ 102       │ Alice        │ Physics       │
│ 2          │ 101       │ Bob          │ Mathematics   │
└────────────┴───────────┴──────────────┴───────────────┘
  student_name depends ONLY on student_id (partial dependency)
  course_name  depends ONLY on course_id  (partial dependency)

✅ IN 2NF — split into three tables:
Students: (student_id PK, student_name)
Courses:  (course_id PK, course_name)
Enrollments: (student_id FK, course_id FK) — composite PK
```

#### 3NF — Third Normal Form
**Rule**: 2NF + no transitive dependencies (non-key column shouldn't depend on another non-key column)

```
❌ VIOLATES 3NF:
┌─────────────┬──────────┬───────────────┬──────────────────┐
│ employee_id │ dept_id  │ dept_name     │ dept_location    │
├─────────────┼──────────┼───────────────┼──────────────────┤
│ 1           │ 10       │ Engineering   │ Building A       │
│ 2           │ 10       │ Engineering   │ Building A       │
│ 3           │ 20       │ Marketing     │ Building B       │
└─────────────┴──────────┴───────────────┴──────────────────┘
  dept_name and dept_location depend on dept_id,
  NOT directly on employee_id → transitive dependency

✅ IN 3NF:
Employees:   (employee_id PK, dept_id FK)
Departments: (dept_id PK, dept_name, dept_location)
```

#### BCNF — Boyce-Codd Normal Form
**Rule**: For every functional dependency X → Y, X must be a superkey (candidate key or superset of one)

```
❌ VIOLATES BCNF:
Table: StudentCourseInstructor
PK: (student, course)
Functional dependencies:
  student, course → instructor     ✅ (LHS is superkey)
  instructor → course              ❌ (instructor is NOT a superkey)

✅ IN BCNF — decompose:
Table1: InstructorCourse (instructor PK, course)
Table2: StudentInstructor (student, instructor) — composite PK
```

### When to Stop Normalizing

| Level | When to Use |
|-------|------------|
| **1NF** | Always — baseline for any relational design |
| **2NF** | Always — avoid obvious duplication |
| **3NF** | Default for OLTP systems — sweet spot for most apps |
| **BCNF** | When asked specifically, or data integrity is paramount |
| **Denormalize** | High-read, OLAP, reporting, caching tables |

---

## 7. Step 6 — Create Tables & Validate

### Full Schema Example: E-Commerce

```sql
-- ============================================
-- E-COMMERCE DATABASE SCHEMA
-- ============================================

-- Users
CREATE TABLE users (
    id              SERIAL PRIMARY KEY,
    email           VARCHAR(255) NOT NULL UNIQUE,
    username        VARCHAR(50) NOT NULL UNIQUE,
    password_hash   CHAR(60) NOT NULL,
    full_name       VARCHAR(200) NOT NULL,
    phone           VARCHAR(20),
    status          VARCHAR(20) NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active', 'inactive', 'banned')),
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Addresses (1:N with users)
CREATE TABLE addresses (
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    label       VARCHAR(50) DEFAULT 'home',    -- home, work, other
    street      VARCHAR(255) NOT NULL,
    city        VARCHAR(100) NOT NULL,
    state       VARCHAR(100),
    zip_code    VARCHAR(20) NOT NULL,
    country     VARCHAR(100) NOT NULL DEFAULT 'US',
    is_default  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Categories (self-referencing for hierarchy)
CREATE TABLE categories (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    parent_id   INTEGER REFERENCES categories(id),
    slug        VARCHAR(100) NOT NULL UNIQUE,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Sellers
CREATE TABLE sellers (
    id              SERIAL PRIMARY KEY,
    user_id         INTEGER NOT NULL REFERENCES users(id),
    business_name   VARCHAR(200) NOT NULL,
    rating          DECIMAL(2,1) DEFAULT 0.0,
    total_sales     INTEGER DEFAULT 0,
    verified        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Products
CREATE TABLE products (
    id              SERIAL PRIMARY KEY,
    seller_id       INTEGER NOT NULL REFERENCES sellers(id),
    name            VARCHAR(300) NOT NULL,
    description     TEXT,
    price           DECIMAL(10,2) NOT NULL CHECK (price > 0),
    compare_price   DECIMAL(10,2),           -- original price for discounts
    stock_quantity  INTEGER NOT NULL DEFAULT 0 CHECK (stock_quantity >= 0),
    sku             VARCHAR(100) UNIQUE,
    weight_grams    INTEGER,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Product-Category (M:N junction)
CREATE TABLE product_categories (
    product_id  INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, category_id)
);

-- Product Images
CREATE TABLE product_images (
    id          SERIAL PRIMARY KEY,
    product_id  INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    url         VARCHAR(500) NOT NULL,
    alt_text    VARCHAR(200),
    sort_order  INTEGER DEFAULT 0,
    is_primary  BOOLEAN DEFAULT FALSE
);

-- Orders
CREATE TABLE orders (
    id                  SERIAL PRIMARY KEY,
    user_id             INTEGER NOT NULL REFERENCES users(id),
    shipping_address_id INTEGER REFERENCES addresses(id),
    status              VARCHAR(30) NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending', 'confirmed', 'processing',
                               'shipped', 'delivered', 'cancelled', 'refunded')),
    subtotal            DECIMAL(10,2) NOT NULL,
    tax                 DECIMAL(10,2) NOT NULL DEFAULT 0,
    shipping_cost       DECIMAL(10,2) NOT NULL DEFAULT 0,
    discount            DECIMAL(10,2) NOT NULL DEFAULT 0,
    total               DECIMAL(10,2) NOT NULL,
    notes               TEXT,
    ordered_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    delivered_at        TIMESTAMP
);

-- Order Items (junction: Order ↔ Product)
CREATE TABLE order_items (
    id          SERIAL PRIMARY KEY,
    order_id    INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id  INTEGER NOT NULL REFERENCES products(id),
    quantity    INTEGER NOT NULL CHECK (quantity > 0),
    unit_price  DECIMAL(10,2) NOT NULL,       -- snapshot price at time of order
    total_price DECIMAL(10,2) NOT NULL,
    UNIQUE (order_id, product_id)              -- prevent duplicate items per order
);

-- Payments
CREATE TABLE payments (
    id              SERIAL PRIMARY KEY,
    order_id        INTEGER NOT NULL REFERENCES orders(id),
    method          VARCHAR(30) NOT NULL
                    CHECK (method IN ('credit_card', 'debit_card', 'upi',
                           'net_banking', 'wallet', 'cod')),
    status          VARCHAR(20) NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'completed', 'failed', 'refunded')),
    amount          DECIMAL(10,2) NOT NULL,
    transaction_id  VARCHAR(200),
    paid_at         TIMESTAMP,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Reviews
CREATE TABLE reviews (
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id),
    product_id  INTEGER NOT NULL REFERENCES products(id),
    rating      SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    title       VARCHAR(200),
    body        TEXT,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, product_id)               -- one review per user per product
);

-- Cart (per user)
CREATE TABLE carts (
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER NOT NULL UNIQUE REFERENCES users(id),
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Cart Items
CREATE TABLE cart_items (
    id          SERIAL PRIMARY KEY,
    cart_id     INTEGER NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    product_id  INTEGER NOT NULL REFERENCES products(id),
    quantity    INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0),
    added_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (cart_id, product_id)
);

-- Coupons
CREATE TABLE coupons (
    id              SERIAL PRIMARY KEY,
    code            VARCHAR(50) NOT NULL UNIQUE,
    discount_type   VARCHAR(20) NOT NULL CHECK (discount_type IN ('percentage', 'fixed')),
    discount_value  DECIMAL(10,2) NOT NULL,
    min_order_value DECIMAL(10,2) DEFAULT 0,
    max_uses        INTEGER,
    current_uses    INTEGER DEFAULT 0,
    valid_from      TIMESTAMP NOT NULL,
    valid_until     TIMESTAMP NOT NULL,
    is_active       BOOLEAN DEFAULT TRUE
);

-- ============================================
-- INDEXES for common query patterns
-- ============================================
CREATE INDEX idx_products_seller ON products(seller_id);
CREATE INDEX idx_products_price ON products(price);
CREATE INDEX idx_products_active ON products(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_orders_user ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_order_items_order ON order_items(order_id);
CREATE INDEX idx_reviews_product ON reviews(product_id);
CREATE INDEX idx_reviews_user ON reviews(user_id);
CREATE INDEX idx_addresses_user ON addresses(user_id);
```

### Validate with Sample Queries

```sql
-- "Get all orders for a user with their items"
SELECT o.id, o.status, o.total, oi.product_id, p.name, oi.quantity, oi.unit_price
FROM orders o
JOIN order_items oi ON o.id = oi.order_id
JOIN products p ON oi.product_id = p.id
WHERE o.user_id = 42
ORDER BY o.ordered_at DESC;

-- "Top 10 products by revenue"
SELECT p.id, p.name, SUM(oi.total_price) AS revenue, SUM(oi.quantity) AS units_sold
FROM products p
JOIN order_items oi ON p.id = oi.product_id
JOIN orders o ON oi.order_id = o.id
WHERE o.status NOT IN ('cancelled', 'refunded')
GROUP BY p.id, p.name
ORDER BY revenue DESC
LIMIT 10;

-- "Average rating per product with review count"
SELECT p.id, p.name, 
       COALESCE(AVG(r.rating), 0) AS avg_rating,
       COUNT(r.id) AS review_count
FROM products p
LEFT JOIN reviews r ON p.id = r.product_id
GROUP BY p.id, p.name
HAVING COUNT(r.id) >= 5
ORDER BY avg_rating DESC;
```

---

## 8. Complete Worked Example: Library Management System

### Step 1: Requirements
- Track books, authors, members, and borrowing
- A book can have multiple authors; an author can write multiple books
- Members borrow books; track due dates and returns
- Track fines for late returns
- Books have copies; each copy can be independently borrowed

### Step 2: Entities
```
Core: Book, Author, Member, Borrow
Supporting: BookCopy, Fine, Publisher
Junction: BookAuthor (Book ↔ Author)
```

### Step 3: Relationships
```
Author M───N Book         → BookAuthor junction
Book   1───N BookCopy     → FK in BookCopy
Member 1───N Borrow       → FK in Borrow
BookCopy 1──N Borrow      → FK in Borrow
Borrow 1──1  Fine         → FK in Fine (optional)
Publisher 1─N Book        → FK in Book
```

### Step 4-6: Schema

```sql
CREATE TABLE publishers (
    id      SERIAL PRIMARY KEY,
    name    VARCHAR(200) NOT NULL,
    country VARCHAR(100)
);

CREATE TABLE authors (
    id      SERIAL PRIMARY KEY,
    name    VARCHAR(200) NOT NULL,
    bio     TEXT
);

CREATE TABLE books (
    id              SERIAL PRIMARY KEY,
    title           VARCHAR(500) NOT NULL,
    isbn            VARCHAR(13) UNIQUE,
    publisher_id    INTEGER REFERENCES publishers(id),
    published_year  SMALLINT,
    genre           VARCHAR(50),
    total_copies    INTEGER NOT NULL DEFAULT 0,
    available_copies INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE book_authors (
    book_id     INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    author_id   INTEGER NOT NULL REFERENCES authors(id) ON DELETE CASCADE,
    PRIMARY KEY (book_id, author_id)
);

CREATE TABLE book_copies (
    id          SERIAL PRIMARY KEY,
    book_id     INTEGER NOT NULL REFERENCES books(id),
    copy_number INTEGER NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'available'
                CHECK (status IN ('available', 'borrowed', 'lost', 'damaged', 'reserved')),
    UNIQUE (book_id, copy_number)
);

CREATE TABLE members (
    id              SERIAL PRIMARY KEY,
    name            VARCHAR(200) NOT NULL,
    email           VARCHAR(255) NOT NULL UNIQUE,
    phone           VARCHAR(20),
    membership_type VARCHAR(20) NOT NULL DEFAULT 'basic'
                    CHECK (membership_type IN ('basic', 'premium', 'student')),
    joined_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE borrows (
    id              SERIAL PRIMARY KEY,
    member_id       INTEGER NOT NULL REFERENCES members(id),
    book_copy_id    INTEGER NOT NULL REFERENCES book_copies(id),
    borrowed_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    due_at          TIMESTAMP NOT NULL,
    returned_at     TIMESTAMP,
    status          VARCHAR(20) NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active', 'returned', 'overdue', 'lost'))
);

CREATE TABLE fines (
    id          SERIAL PRIMARY KEY,
    borrow_id   INTEGER NOT NULL UNIQUE REFERENCES borrows(id),
    amount      DECIMAL(8,2) NOT NULL,
    reason      VARCHAR(100),
    paid        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    paid_at     TIMESTAMP
);
```

### Validation Queries

```sql
-- "Which books are currently borrowed by a member?"
SELECT b.title, bc.copy_number, br.borrowed_at, br.due_at
FROM borrows br
JOIN book_copies bc ON br.book_copy_id = bc.id
JOIN books b ON bc.book_id = b.id
WHERE br.member_id = 1 AND br.status = 'active';

-- "Books with all their authors"
SELECT b.title, STRING_AGG(a.name, ', ') AS authors
FROM books b
JOIN book_authors ba ON b.id = ba.book_id
JOIN authors a ON ba.author_id = a.id
GROUP BY b.title;

-- "Members with unpaid fines"
SELECT m.name, m.email, SUM(f.amount) AS total_unpaid
FROM members m
JOIN borrows br ON m.id = br.member_id
JOIN fines f ON br.id = f.borrow_id
WHERE f.paid = FALSE
GROUP BY m.id, m.name, m.email
ORDER BY total_unpaid DESC;

-- "Most borrowed books this month"
SELECT b.title, COUNT(*) AS borrow_count
FROM borrows br
JOIN book_copies bc ON br.book_copy_id = bc.id
JOIN books b ON bc.book_id = b.id
WHERE br.borrowed_at >= DATE_TRUNC('month', CURRENT_DATE)
GROUP BY b.id, b.title
ORDER BY borrow_count DESC
LIMIT 10;
```

---

## 9. Common DB Design Interview Scenarios

### Scenario 1: "Design a database for Twitter"
**Key entities**: User, Tweet, Follow, Like, Retweet, Hashtag, DirectMessage
**Key challenges**: 
- Fan-out on write vs fan-out on read for timeline
- Follow is a self-referencing M:N relationship
- Tweet can be a reply (self-referencing)

### Scenario 2: "Design a database for Uber"
**Key entities**: Rider, Driver, Trip, Vehicle, Payment, Rating, Location
**Key challenges**:
- Real-time location updates (consider time-series or geospatial)
- Trip state machine (requested → matched → in-progress → completed)
- Pricing model (surge pricing, estimates)

### Scenario 3: "Design a database for Netflix"
**Key entities**: User, Profile, Content, Genre, Watchlist, ViewingHistory, Subscription, Plan
**Key challenges**:
- Multiple profiles per account
- Content ↔ Genre is M:N
- Viewing history is massive (analytics use case → consider OLAP)
- Recommendations (separate system, but need viewing data)

### Scenario 4: "Design a database for a Chat Application"
**Key entities**: User, Conversation, Message, ConversationMember, Attachment
**Key challenges**:
- Group chats (Conversation ↔ User is M:N)
- Message ordering (use timestamp + sequence)
- Read receipts (per user per message)
- Message search (consider full-text index)

---

## 10. Interview Tips

### Do's
- ✅ Start with requirements clarification
- ✅ Draw ER diagram before writing SQL
- ✅ Explain your normalization choices
- ✅ Add appropriate constraints (NOT NULL, CHECK, UNIQUE, FK)
- ✅ Create indexes for your expected query patterns
- ✅ Validate schema with 2-3 sample queries
- ✅ Mention trade-offs (normalized vs denormalized)

### Don'ts
- ❌ Jump straight into CREATE TABLE
- ❌ Forget about NULL handling
- ❌ Use INT for everything
- ❌ Skip foreign key constraints
- ❌ Design without considering query patterns
- ❌ Over-normalize (4NF, 5NF) unless specifically asked
- ❌ Forget about indexes

### Time Management (45-min DB design interview)
| Phase | Time | Activity |
|-------|------|----------|
| Clarify | 5 min | Ask questions, scope the problem |
| Entities | 5 min | List entities, define relationships |
| ER Diagram | 10 min | Draw on whiteboard, discuss cardinality |
| Schema | 15 min | Write DDL, explain constraints |
| Queries | 10 min | Write 2-3 validation queries, discuss indexes |
