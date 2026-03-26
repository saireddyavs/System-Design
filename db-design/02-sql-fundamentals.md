# SQL Fundamentals — Complete Reference

## Everything You Need to Know About SQL for Interviews

---

## 1. SQL Command Categories

```
SQL Commands
├── DDL (Data Definition Language)     → Structure
│   ├── CREATE
│   ├── ALTER
│   ├── DROP
│   ├── TRUNCATE
│   └── RENAME
├── DML (Data Manipulation Language)   → Data
│   ├── SELECT
│   ├── INSERT
│   ├── UPDATE
│   ├── DELETE
│   └── MERGE / UPSERT
├── DCL (Data Control Language)        → Permissions
│   ├── GRANT
│   └── REVOKE
├── TCL (Transaction Control Language) → Transactions
│   ├── BEGIN / START TRANSACTION
│   ├── COMMIT
│   ├── ROLLBACK
│   └── SAVEPOINT
└── DQL (Data Query Language)          → Queries
    └── SELECT (with all clauses)
```

---

## 2. DDL — Data Definition Language

### CREATE TABLE

```sql
-- Complete syntax with all constraint types
CREATE TABLE employees (
    -- Column constraints
    id              SERIAL PRIMARY KEY,                    -- auto-increment + PK
    email           VARCHAR(255) NOT NULL UNIQUE,          -- NOT NULL + UNIQUE
    first_name      VARCHAR(100) NOT NULL,
    last_name       VARCHAR(100) NOT NULL,
    department_id   INTEGER REFERENCES departments(id)     -- FOREIGN KEY (inline)
                    ON DELETE SET NULL                      -- referential action
                    ON UPDATE CASCADE,
    salary          DECIMAL(10,2) NOT NULL 
                    CHECK (salary > 0),                    -- CHECK constraint
    hire_date       DATE NOT NULL DEFAULT CURRENT_DATE,    -- DEFAULT value
    status          VARCHAR(20) DEFAULT 'active'
                    CHECK (status IN ('active', 'inactive', 'terminated')),
    manager_id      INTEGER REFERENCES employees(id),      -- self-referencing FK
    
    -- Table-level constraints
    CONSTRAINT uq_emp_email UNIQUE (email),                -- named constraint
    CONSTRAINT chk_salary CHECK (salary BETWEEN 1000 AND 1000000)
);

-- Create table from another table
CREATE TABLE emp_backup AS
SELECT * FROM employees WHERE status = 'active';

-- Create table with composite primary key
CREATE TABLE order_items (
    order_id    INTEGER REFERENCES orders(id),
    product_id  INTEGER REFERENCES products(id),
    quantity    INTEGER NOT NULL CHECK (quantity > 0),
    unit_price  DECIMAL(10,2) NOT NULL,
    PRIMARY KEY (order_id, product_id)
);
```

### All Constraint Types

| Constraint | Purpose | Scope | Example |
|-----------|---------|-------|---------|
| `PRIMARY KEY` | Unique + NOT NULL identifier | Column / Table | `id SERIAL PRIMARY KEY` |
| `FOREIGN KEY` | Referential integrity | Column / Table | `REFERENCES other_table(id)` |
| `UNIQUE` | No duplicate values (NULLs allowed) | Column / Table | `email VARCHAR(255) UNIQUE` |
| `NOT NULL` | No NULL values | Column | `name VARCHAR(100) NOT NULL` |
| `CHECK` | Value must satisfy condition | Column / Table | `CHECK (age >= 18)` |
| `DEFAULT` | Default value if none provided | Column | `DEFAULT CURRENT_TIMESTAMP` |
| `EXCLUDE` | PostgreSQL: exclusion constraint | Table | Prevent overlapping ranges |

### Foreign Key Actions

```sql
-- ON DELETE options:
ON DELETE CASCADE      -- delete child rows when parent deleted
ON DELETE SET NULL     -- set FK to NULL when parent deleted
ON DELETE SET DEFAULT  -- set FK to default when parent deleted
ON DELETE RESTRICT     -- prevent parent deletion if children exist (default)
ON DELETE NO ACTION    -- same as RESTRICT but checked at end of transaction

-- ON UPDATE options (same choices):
ON UPDATE CASCADE      -- update child FK when parent PK changes
ON UPDATE RESTRICT     -- prevent parent PK change if children exist
```

### ALTER TABLE

```sql
-- Add column
ALTER TABLE employees ADD COLUMN phone VARCHAR(20);

-- Drop column
ALTER TABLE employees DROP COLUMN phone;

-- Rename column
ALTER TABLE employees RENAME COLUMN first_name TO fname;

-- Change data type
ALTER TABLE employees ALTER COLUMN salary TYPE NUMERIC(12,2);

-- Add/Drop constraint
ALTER TABLE employees ADD CONSTRAINT chk_salary CHECK (salary > 0);
ALTER TABLE employees DROP CONSTRAINT chk_salary;

-- Add NOT NULL
ALTER TABLE employees ALTER COLUMN email SET NOT NULL;

-- Remove NOT NULL
ALTER TABLE employees ALTER COLUMN email DROP NOT NULL;

-- Set default
ALTER TABLE employees ALTER COLUMN status SET DEFAULT 'active';

-- Add foreign key
ALTER TABLE employees ADD CONSTRAINT fk_dept 
    FOREIGN KEY (department_id) REFERENCES departments(id);

-- Rename table
ALTER TABLE employees RENAME TO staff;
```

### DROP & TRUNCATE

```sql
-- DROP: removes table and all data permanently
DROP TABLE employees;            -- error if table doesn't exist
DROP TABLE IF EXISTS employees;  -- safe drop
DROP TABLE employees CASCADE;    -- also drops dependent objects (views, FKs)

-- TRUNCATE: removes all rows but keeps table structure
TRUNCATE TABLE employees;                    -- much faster than DELETE (no WAL)
TRUNCATE TABLE employees RESTART IDENTITY;   -- reset auto-increment
TRUNCATE TABLE employees CASCADE;            -- also truncate tables with FKs
```

**Key Difference: DROP vs TRUNCATE vs DELETE**

| Feature | DROP | TRUNCATE | DELETE |
|---------|------|----------|--------|
| Removes table? | Yes | No | No |
| Removes data? | Yes | Yes | Yes (selective) |
| WHERE clause? | N/A | No | Yes |
| Rollback? | No | DB-dependent | Yes |
| Triggers fired? | No | No | Yes |
| Speed | Fast | Fast | Slow (row-by-row) |
| Space reclaimed | Immediate | Immediate | After VACUUM |
| Resets auto-increment? | N/A | Yes (with RESTART) | No |

---

## 3. DML — Data Manipulation Language

### INSERT

```sql
-- Single row insert
INSERT INTO employees (first_name, last_name, email, salary, department_id)
VALUES ('Alice', 'Smith', 'alice@company.com', 85000, 1);

-- Multi-row insert
INSERT INTO employees (first_name, last_name, email, salary, department_id)
VALUES 
    ('Bob', 'Jones', 'bob@company.com', 92000, 2),
    ('Carol', 'Davis', 'carol@company.com', 78000, 1),
    ('Dave', 'Wilson', 'dave@company.com', 105000, 3);

-- Insert from SELECT
INSERT INTO emp_archive (first_name, last_name, email, terminated_at)
SELECT first_name, last_name, email, CURRENT_TIMESTAMP
FROM employees
WHERE status = 'terminated';

-- UPSERT (PostgreSQL: INSERT ... ON CONFLICT)
INSERT INTO products (sku, name, price, stock)
VALUES ('SKU-001', 'Widget', 9.99, 100)
ON CONFLICT (sku) DO UPDATE SET
    price = EXCLUDED.price,
    stock = products.stock + EXCLUDED.stock;

-- UPSERT (MySQL: INSERT ... ON DUPLICATE KEY UPDATE)
INSERT INTO products (sku, name, price, stock)
VALUES ('SKU-001', 'Widget', 9.99, 100)
ON DUPLICATE KEY UPDATE
    price = VALUES(price),
    stock = stock + VALUES(stock);

-- INSERT with RETURNING (PostgreSQL)
INSERT INTO orders (user_id, total)
VALUES (42, 199.99)
RETURNING id, created_at;
```

### UPDATE

```sql
-- Basic update
UPDATE employees SET salary = 90000 WHERE id = 1;

-- Multi-column update
UPDATE employees 
SET salary = salary * 1.10,
    status = 'promoted',
    updated_at = CURRENT_TIMESTAMP
WHERE department_id = 3 AND hire_date < '2020-01-01';

-- Update with subquery
UPDATE products 
SET price = price * 0.9               -- 10% discount
WHERE category_id IN (
    SELECT id FROM categories WHERE name = 'Clearance'
);

-- Update with JOIN (PostgreSQL)
UPDATE order_items oi
SET unit_price = p.price
FROM products p
WHERE oi.product_id = p.id AND oi.unit_price != p.price;

-- Update with JOIN (MySQL)
UPDATE order_items oi
JOIN products p ON oi.product_id = p.id
SET oi.unit_price = p.price
WHERE oi.unit_price != p.price;

-- Update with RETURNING (PostgreSQL)
UPDATE employees SET salary = salary * 1.05
WHERE department_id = 1
RETURNING id, first_name, salary;

-- Conditional update with CASE
UPDATE employees 
SET salary = CASE
    WHEN department_id = 1 THEN salary * 1.10
    WHEN department_id = 2 THEN salary * 1.08
    ELSE salary * 1.05
END
WHERE status = 'active';
```

### DELETE

```sql
-- Basic delete
DELETE FROM employees WHERE id = 42;

-- Delete with subquery
DELETE FROM order_items
WHERE order_id IN (
    SELECT id FROM orders WHERE status = 'cancelled'
);

-- Delete with JOIN (PostgreSQL: USING)
DELETE FROM order_items oi
USING orders o
WHERE oi.order_id = o.id AND o.status = 'cancelled';

-- Delete with RETURNING
DELETE FROM sessions
WHERE expires_at < CURRENT_TIMESTAMP
RETURNING user_id, session_token;

-- Delete all rows (prefer TRUNCATE for this)
DELETE FROM temp_data;
```

### MERGE (SQL Standard / SQL Server / Oracle)

```sql
-- MERGE: upsert on steroids — combines INSERT, UPDATE, DELETE
MERGE INTO target_table t
USING source_table s
ON t.id = s.id
WHEN MATCHED AND s.status = 'delete' THEN
    DELETE
WHEN MATCHED THEN
    UPDATE SET t.name = s.name, t.price = s.price
WHEN NOT MATCHED THEN
    INSERT (id, name, price) VALUES (s.id, s.name, s.price);
```

---

## 4. SELECT — The Complete Query

### Logical Order of Execution

This is **critical** for interviews — SQL doesn't execute in the order you write it:

```
Written Order:           Execution Order:
─────────────           ────────────────
SELECT                   1. FROM / JOIN
FROM                     2. WHERE
WHERE                    3. GROUP BY
GROUP BY                 4. HAVING
HAVING                   5. SELECT
ORDER BY                 6. DISTINCT
LIMIT/OFFSET             7. ORDER BY
                         8. LIMIT/OFFSET
```

**Why this matters:**
- You CAN'T use column aliases from SELECT in WHERE (WHERE executes before SELECT)
- You CAN use column aliases in ORDER BY (ORDER BY executes after SELECT)
- You CAN'T use aggregate functions in WHERE (use HAVING instead)

### WHERE Clause — Filtering Rows

```sql
-- Comparison operators
SELECT * FROM products WHERE price > 100;
SELECT * FROM products WHERE price >= 100 AND price <= 500;
SELECT * FROM products WHERE price BETWEEN 100 AND 500;  -- inclusive

-- Logical operators
SELECT * FROM employees 
WHERE department_id = 1 AND salary > 80000;

SELECT * FROM employees 
WHERE department_id = 1 OR department_id = 2;

SELECT * FROM employees 
WHERE NOT (status = 'terminated');

-- IN operator
SELECT * FROM employees 
WHERE department_id IN (1, 2, 3);

-- NOT IN (careful with NULLs!)
SELECT * FROM employees 
WHERE department_id NOT IN (4, 5);  -- won't work if any value is NULL!

-- LIKE (pattern matching)
SELECT * FROM employees WHERE email LIKE '%@gmail.com';   -- ends with
SELECT * FROM employees WHERE first_name LIKE 'A%';       -- starts with
SELECT * FROM employees WHERE email LIKE '%admin%';        -- contains
SELECT * FROM employees WHERE first_name LIKE '_ob';       -- 3 chars, ending 'ob'

-- ILIKE (case-insensitive, PostgreSQL)
SELECT * FROM employees WHERE first_name ILIKE 'alice';

-- NULL handling
SELECT * FROM employees WHERE manager_id IS NULL;          -- is null
SELECT * FROM employees WHERE manager_id IS NOT NULL;      -- is not null
-- NEVER use = NULL or != NULL — they ALWAYS return false!

-- SIMILAR TO / REGEXP (PostgreSQL)
SELECT * FROM products WHERE sku ~ '^[A-Z]{3}-\d{3}$';

-- EXISTS
SELECT * FROM users u
WHERE EXISTS (
    SELECT 1 FROM orders o WHERE o.user_id = u.id
);
```

### ORDER BY

```sql
-- Sort ascending (default)
SELECT * FROM employees ORDER BY salary;
SELECT * FROM employees ORDER BY salary ASC;

-- Sort descending
SELECT * FROM employees ORDER BY salary DESC;

-- Multiple sort columns
SELECT * FROM employees ORDER BY department_id ASC, salary DESC;

-- Sort by column position (not recommended but valid)
SELECT first_name, last_name, salary FROM employees ORDER BY 3 DESC;

-- Sort by expression
SELECT * FROM products ORDER BY (price * stock_quantity) DESC;

-- NULLS FIRST / NULLS LAST (PostgreSQL)
SELECT * FROM employees ORDER BY manager_id NULLS LAST;

-- Sort by CASE
SELECT * FROM orders 
ORDER BY CASE status
    WHEN 'pending' THEN 1
    WHEN 'processing' THEN 2
    WHEN 'shipped' THEN 3
    WHEN 'delivered' THEN 4
    ELSE 5
END;
```

### LIMIT, OFFSET, FETCH

```sql
-- LIMIT and OFFSET (most databases)
SELECT * FROM products ORDER BY created_at DESC LIMIT 10;            -- first 10
SELECT * FROM products ORDER BY created_at DESC LIMIT 10 OFFSET 20;  -- page 3

-- FETCH (SQL standard, supported by PostgreSQL, Oracle, SQL Server)
SELECT * FROM products ORDER BY created_at DESC
FETCH FIRST 10 ROWS ONLY;

SELECT * FROM products ORDER BY created_at DESC
OFFSET 20 ROWS FETCH NEXT 10 ROWS ONLY;

-- Interview tip: OFFSET-based pagination is inefficient for large datasets
-- Better: keyset/cursor-based pagination
SELECT * FROM products 
WHERE created_at < '2024-01-15 10:30:00'  -- cursor from last page
ORDER BY created_at DESC 
LIMIT 10;
```

### DISTINCT

```sql
-- Remove duplicate rows
SELECT DISTINCT department_id FROM employees;

-- DISTINCT on multiple columns (unique combinations)
SELECT DISTINCT department_id, status FROM employees;

-- DISTINCT ON (PostgreSQL) — first row per group
SELECT DISTINCT ON (department_id) department_id, first_name, salary
FROM employees
ORDER BY department_id, salary DESC;
-- Returns the highest-paid employee per department
```

---

## 5. Data Types Reference

### Numeric Types

| Type | Size | Range | Use |
|------|------|-------|-----|
| `SMALLINT` | 2 bytes | -32K to 32K | Status codes, small counters |
| `INTEGER` / `INT` | 4 bytes | -2.1B to 2.1B | General-purpose IDs, counts |
| `BIGINT` | 8 bytes | -9.2×10¹⁸ to 9.2×10¹⁸ | Large IDs, big data counts |
| `SERIAL` | 4 bytes | Auto-increment INT | Primary keys |
| `BIGSERIAL` | 8 bytes | Auto-increment BIGINT | Large table PKs |
| `DECIMAL(p,s)` / `NUMERIC` | Variable | Exact precision | Money, financial |
| `REAL` / `FLOAT4` | 4 bytes | 6 decimal precision | Scientific (approx) |
| `DOUBLE PRECISION` / `FLOAT8` | 8 bytes | 15 decimal precision | Scientific (approx) |

### String Types

| Type | Description | Use |
|------|-------------|-----|
| `CHAR(n)` | Fixed-length, padded | Fixed-format codes (country: `CHAR(2)`) |
| `VARCHAR(n)` | Variable-length, max n | Most string data |
| `TEXT` | Unlimited length | Large text, descriptions |

### Date/Time Types

| Type | Description | Example |
|------|-------------|---------|
| `DATE` | Date only | `'2024-01-15'` |
| `TIME` | Time only | `'14:30:00'` |
| `TIMESTAMP` | Date + time | `'2024-01-15 14:30:00'` |
| `TIMESTAMP WITH TIME ZONE` | Date + time + TZ | `'2024-01-15 14:30:00+05:30'` |
| `INTERVAL` | Duration | `INTERVAL '2 hours 30 minutes'` |

### Other Types

| Type | Description | Use |
|------|-------------|-----|
| `BOOLEAN` | true/false | Flags |
| `UUID` | Universally unique ID | Distributed primary keys |
| `JSONB` | Binary JSON (PostgreSQL) | Semi-structured data |
| `ARRAY` | Array of values (PostgreSQL) | Tags, categories |
| `INET` | IP address (PostgreSQL) | Network data |
| `ENUM` | Predefined values | Status fields |

---

## 6. String Functions

```sql
-- Length
SELECT LENGTH('Hello World');                -- 11
SELECT CHAR_LENGTH('Hello World');           -- 11

-- Case
SELECT UPPER('hello');                       -- 'HELLO'
SELECT LOWER('HELLO');                       -- 'hello'
SELECT INITCAP('hello world');              -- 'Hello World' (PostgreSQL)

-- Substring
SELECT SUBSTRING('Hello World' FROM 1 FOR 5);  -- 'Hello'
SELECT SUBSTR('Hello World', 7, 5);             -- 'World'
SELECT LEFT('Hello World', 5);                  -- 'Hello'
SELECT RIGHT('Hello World', 5);                 -- 'World'

-- Trim
SELECT TRIM('  Hello  ');                    -- 'Hello'
SELECT LTRIM('  Hello');                     -- 'Hello'
SELECT RTRIM('Hello  ');                     -- 'Hello'
SELECT TRIM(LEADING '0' FROM '000123');      -- '123'

-- Concatenation
SELECT 'Hello' || ' ' || 'World';           -- 'Hello World' (PostgreSQL, Oracle)
SELECT CONCAT('Hello', ' ', 'World');        -- 'Hello World' (all DBs)
SELECT CONCAT_WS(', ', 'Alice', 'Bob');      -- 'Alice, Bob'

-- Search
SELECT POSITION('World' IN 'Hello World');   -- 7
SELECT STRPOS('Hello World', 'World');       -- 7 (PostgreSQL)

-- Replace
SELECT REPLACE('Hello World', 'World', 'SQL');  -- 'Hello SQL'

-- Repeat / Pad
SELECT REPEAT('ab', 3);                     -- 'ababab'
SELECT LPAD('42', 5, '0');                   -- '00042'
SELECT RPAD('Hi', 10, '.');                  -- 'Hi........'

-- Reverse
SELECT REVERSE('Hello');                     -- 'olleH'

-- Split (PostgreSQL)
SELECT STRING_TO_ARRAY('a,b,c', ',');        -- {a,b,c}
SELECT SPLIT_PART('a.b.c', '.', 2);         -- 'b'

-- Aggregate strings
SELECT STRING_AGG(name, ', ' ORDER BY name)  -- 'Alice, Bob, Carol'
FROM employees;

-- Regular expressions (PostgreSQL)
SELECT 'abc123' ~ '\d+';                    -- true
SELECT REGEXP_REPLACE('abc123def', '\d+', 'NUM');  -- 'abcNUMdef'
```

---

## 7. Date/Time Functions

```sql
-- Current date/time
SELECT CURRENT_DATE;                         -- 2024-01-15
SELECT CURRENT_TIME;                         -- 14:30:00+00
SELECT CURRENT_TIMESTAMP;                    -- 2024-01-15 14:30:00+00
SELECT NOW();                                -- same as CURRENT_TIMESTAMP

-- Extract parts
SELECT EXTRACT(YEAR FROM TIMESTAMP '2024-01-15 14:30:00');   -- 2024
SELECT EXTRACT(MONTH FROM CURRENT_DATE);                      -- 1
SELECT EXTRACT(DOW FROM CURRENT_DATE);                        -- 0-6 (Sun=0)
SELECT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP);                 -- Unix timestamp
SELECT DATE_PART('year', CURRENT_DATE);                       -- 2024

-- Date arithmetic
SELECT CURRENT_DATE + INTERVAL '7 days';                -- add 7 days
SELECT CURRENT_DATE - INTERVAL '1 month';               -- subtract 1 month
SELECT CURRENT_TIMESTAMP + INTERVAL '2 hours 30 minutes';
SELECT '2024-12-31'::DATE - '2024-01-01'::DATE;         -- 365 (days between)

-- Truncate (round down)
SELECT DATE_TRUNC('month', CURRENT_TIMESTAMP);   -- first day of month
SELECT DATE_TRUNC('year', CURRENT_TIMESTAMP);    -- first day of year
SELECT DATE_TRUNC('hour', CURRENT_TIMESTAMP);    -- start of current hour

-- Format (PostgreSQL)
SELECT TO_CHAR(CURRENT_TIMESTAMP, 'YYYY-MM-DD HH24:MI:SS');
SELECT TO_CHAR(CURRENT_DATE, 'Day, DD Month YYYY');

-- Parse string to date
SELECT TO_DATE('15-01-2024', 'DD-MM-YYYY');
SELECT TO_TIMESTAMP('2024-01-15 14:30', 'YYYY-MM-DD HH24:MI');

-- Age calculation
SELECT AGE(CURRENT_DATE, '1990-05-20'::DATE);   -- '33 years 7 mons 26 days'

-- Generate series of dates
SELECT generate_series(
    '2024-01-01'::DATE,
    '2024-01-31'::DATE,
    '1 day'::INTERVAL
) AS date;
```

---

## 8. Numeric Functions

```sql
-- Rounding
SELECT ROUND(3.7);          -- 4
SELECT ROUND(3.14159, 2);   -- 3.14
SELECT CEIL(3.2);           -- 4   (round up)
SELECT FLOOR(3.8);          -- 3   (round down)
SELECT TRUNC(3.789, 2);     -- 3.78 (truncate, no rounding)

-- Absolute / Sign
SELECT ABS(-42);             -- 42
SELECT SIGN(-42);            -- -1
SELECT SIGN(0);              -- 0
SELECT SIGN(42);             -- 1

-- Power / Root
SELECT POWER(2, 10);         -- 1024
SELECT SQRT(144);            -- 12
SELECT CBRT(27);             -- 3

-- Modulo
SELECT MOD(17, 5);           -- 2
SELECT 17 % 5;               -- 2

-- Greatest / Least
SELECT GREATEST(10, 20, 30); -- 30
SELECT LEAST(10, 20, 30);    -- 10

-- Random
SELECT RANDOM();                       -- 0.0 to 1.0
SELECT FLOOR(RANDOM() * 100) + 1;      -- random int 1-100

-- Logarithm
SELECT LOG(100);              -- 2
SELECT LN(2.71828);           -- ~1
```

---

## 9. NULL Handling

```sql
-- COALESCE: returns first non-NULL argument
SELECT COALESCE(phone, email, 'No contact') AS contact FROM users;

-- NULLIF: returns NULL if arguments are equal
SELECT NULLIF(stock, 0);  -- returns NULL instead of 0 (avoids division by zero)
SELECT total / NULLIF(count, 0) AS average;

-- CASE with NULL
SELECT 
    CASE 
        WHEN manager_id IS NULL THEN 'CEO'
        ELSE 'Has Manager'
    END AS manager_status
FROM employees;

-- NULL in aggregations
-- NULL values are IGNORED by aggregate functions (COUNT, SUM, AVG, etc.)
-- Exception: COUNT(*) counts all rows including NULLs
SELECT COUNT(*) AS total_rows,          -- counts all rows
       COUNT(phone) AS with_phone,      -- counts non-NULL phones
       COUNT(*) - COUNT(phone) AS without_phone
FROM users;

-- NULL in comparisons
-- ANY comparison with NULL returns NULL (not true, not false)
-- x = NULL   → NULL (not false!)
-- x != NULL  → NULL (not true!)
-- NULL = NULL → NULL (not true!)
-- Use IS NULL / IS NOT NULL instead

-- NULL in boolean logic
-- NULL AND TRUE  → NULL
-- NULL AND FALSE → FALSE
-- NULL OR TRUE   → TRUE
-- NULL OR FALSE  → NULL
```

---

## 10. CASE Expressions

```sql
-- Simple CASE
SELECT first_name, 
    CASE department_id
        WHEN 1 THEN 'Engineering'
        WHEN 2 THEN 'Marketing'
        WHEN 3 THEN 'Sales'
        ELSE 'Other'
    END AS department_name
FROM employees;

-- Searched CASE (more flexible)
SELECT first_name, salary,
    CASE 
        WHEN salary >= 150000 THEN 'Executive'
        WHEN salary >= 100000 THEN 'Senior'
        WHEN salary >= 70000 THEN 'Mid-Level'
        ELSE 'Junior'
    END AS salary_band
FROM employees;

-- CASE in WHERE
SELECT * FROM products
WHERE CASE 
    WHEN category = 'perishable' THEN stock > 100
    ELSE stock > 20
END;

-- CASE in ORDER BY
SELECT * FROM orders
ORDER BY CASE status
    WHEN 'urgent' THEN 1
    WHEN 'pending' THEN 2
    WHEN 'processing' THEN 3
    ELSE 4
END;

-- CASE in aggregation (pivot-style)
SELECT 
    department_id,
    COUNT(CASE WHEN status = 'active' THEN 1 END) AS active_count,
    COUNT(CASE WHEN status = 'inactive' THEN 1 END) AS inactive_count,
    SUM(CASE WHEN status = 'active' THEN salary ELSE 0 END) AS active_salary_total
FROM employees
GROUP BY department_id;

-- CASE with NULL
SELECT 
    CASE 
        WHEN bonus IS NULL THEN 'No Bonus'
        WHEN bonus = 0 THEN 'Zero Bonus'
        ELSE 'Has Bonus: ' || bonus::TEXT
    END AS bonus_status
FROM employees;
```

---

## 11. Type Casting

```sql
-- PostgreSQL CAST
SELECT CAST('42' AS INTEGER);
SELECT CAST('2024-01-15' AS DATE);
SELECT CAST(3.14 AS INTEGER);          -- 3 (truncates)

-- PostgreSQL :: shorthand
SELECT '42'::INTEGER;
SELECT '2024-01-15'::DATE;
SELECT 3.14::INTEGER;

-- CAST in expressions
SELECT CAST(COUNT(*) AS DECIMAL) / CAST(total AS DECIMAL) AS ratio
FROM items;

-- Implicit vs Explicit casting
-- Implicit: DB automatically converts (e.g., INT + DECIMAL → DECIMAL)
-- Explicit: You specify the conversion (required for incompatible types)
```

---

## 12. Indexes — Essential Knowledge

### Types of Indexes

| Index Type | Use Case | Example |
|-----------|----------|---------|
| **B-Tree** (default) | Equality, range, sorting | `WHERE price > 100 ORDER BY price` |
| **Hash** | Equality only | `WHERE email = 'x@y.com'` |
| **GIN** | Full-text, JSONB, arrays | `WHERE tags @> '{sql}'` |
| **GiST** | Geometric, range types | `WHERE location <-> point(x,y)` |
| **BRIN** | Large tables, naturally ordered | Time-series data |

### Creating Indexes

```sql
-- Basic index
CREATE INDEX idx_employees_email ON employees(email);

-- Unique index (enforces uniqueness)
CREATE UNIQUE INDEX idx_users_email ON users(email);

-- Composite index (multi-column)
CREATE INDEX idx_orders_user_date ON orders(user_id, created_at DESC);

-- Partial index (only index subset of rows)
CREATE INDEX idx_active_products ON products(name) WHERE is_active = TRUE;
-- Only indexes active products → smaller, faster

-- Covering index (INCLUDE: store extra columns in index)
CREATE INDEX idx_orders_user ON orders(user_id) INCLUDE (status, total);
-- Index-only scan: all needed data is in the index

-- Expression index
CREATE INDEX idx_users_lower_email ON users(LOWER(email));
-- Enables: WHERE LOWER(email) = 'alice@example.com'

-- GIN index for JSONB
CREATE INDEX idx_products_metadata ON products USING GIN(metadata);

-- GIN index for full-text search
CREATE INDEX idx_products_search ON products USING GIN(to_tsvector('english', name || ' ' || description));
```

### EXPLAIN ANALYZE

```sql
-- View query plan
EXPLAIN SELECT * FROM employees WHERE email = 'alice@example.com';

-- View query plan with actual execution metrics
EXPLAIN ANALYZE SELECT * FROM employees WHERE email = 'alice@example.com';

-- What to look for:
-- Seq Scan          → full table scan (bad for large tables)
-- Index Scan        → using index (good)
-- Index Only Scan   → even better (covering index)
-- Bitmap Index Scan → combining multiple indexes
-- Hash Join         → joining with hash table
-- Nested Loop       → joining row by row (can be slow)
-- Sort              → sorting in memory or disk
```

### Index Guidelines

| Do | Don't |
|-----|-------|
| Index columns in WHERE, JOIN, ORDER BY | Don't index every column |
| Index FKs (not auto-indexed in PostgreSQL) | Don't index tiny tables |
| Use partial indexes for filtered queries | Don't index columns with low cardinality alone |
| Monitor with `pg_stat_user_indexes` | Don't forget: indexes slow down writes |
| Composite index: most selective column first | Don't duplicate indexes |

---

## 13. Transactions & Isolation Levels

### Transaction Basics

```sql
-- Explicit transaction
BEGIN;
    UPDATE accounts SET balance = balance - 100 WHERE id = 1;
    UPDATE accounts SET balance = balance + 100 WHERE id = 2;
COMMIT;

-- Rollback on error
BEGIN;
    UPDATE accounts SET balance = balance - 100 WHERE id = 1;
    -- something goes wrong
ROLLBACK;

-- Savepoint (partial rollback)
BEGIN;
    INSERT INTO orders (...) VALUES (...);
    SAVEPOINT before_payment;
    INSERT INTO payments (...) VALUES (...);  -- fails
    ROLLBACK TO before_payment;               -- undo only payment
    -- order insert is preserved
COMMIT;
```

### Isolation Levels

| Level | Dirty Read | Non-Repeatable Read | Phantom Read | Performance |
|-------|-----------|--------------------|--------------| ------------|
| **READ UNCOMMITTED** | ✅ Possible | ✅ Possible | ✅ Possible | Fastest |
| **READ COMMITTED** | ❌ Prevented | ✅ Possible | ✅ Possible | Fast |
| **REPEATABLE READ** | ❌ Prevented | ❌ Prevented | ✅ Possible | Moderate |
| **SERIALIZABLE** | ❌ Prevented | ❌ Prevented | ❌ Prevented | Slowest |

```sql
-- Set isolation level
SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;
BEGIN;
    -- queries here run at serializable isolation
COMMIT;

-- PostgreSQL default: READ COMMITTED
-- MySQL InnoDB default: REPEATABLE READ
```

### Concurrency Problems Explained

```
DIRTY READ: Transaction reads uncommitted data from another transaction.
  T1: UPDATE salary = 50000 WHERE id=1  (not committed)
  T2: SELECT salary WHERE id=1 → reads 50000 (dirty!)
  T1: ROLLBACK                 → T2 has wrong data

NON-REPEATABLE READ: Same query returns different data within same transaction.
  T1: SELECT salary WHERE id=1 → 40000
  T2: UPDATE salary = 50000 WHERE id=1; COMMIT
  T1: SELECT salary WHERE id=1 → 50000 (changed!)

PHANTOM READ: New rows appear between reads in same transaction.
  T1: SELECT COUNT(*) WHERE dept=1 → 5
  T2: INSERT INTO employees (...dept=1...); COMMIT
  T1: SELECT COUNT(*) WHERE dept=1 → 6 (phantom row!)
```

---

## 14. Views

```sql
-- Create view
CREATE VIEW active_employees AS
SELECT id, first_name, last_name, email, department_id, salary
FROM employees
WHERE status = 'active';

-- Use view like a table
SELECT * FROM active_employees WHERE department_id = 1;

-- Materialized view (PostgreSQL) — cached query result
CREATE MATERIALIZED VIEW monthly_revenue AS
SELECT 
    DATE_TRUNC('month', ordered_at) AS month,
    SUM(total) AS revenue,
    COUNT(*) AS order_count
FROM orders
WHERE status = 'delivered'
GROUP BY DATE_TRUNC('month', ordered_at);

-- Refresh materialized view
REFRESH MATERIALIZED VIEW monthly_revenue;
REFRESH MATERIALIZED VIEW CONCURRENTLY monthly_revenue;  -- no lock

-- Drop view
DROP VIEW active_employees;
DROP MATERIALIZED VIEW monthly_revenue;
```

---

## 15. Stored Procedures & Functions (PostgreSQL)

```sql
-- Function: returns a value
CREATE OR REPLACE FUNCTION calculate_discount(
    price DECIMAL, 
    discount_pct INTEGER
) RETURNS DECIMAL AS $$
BEGIN
    RETURN price * (1 - discount_pct / 100.0);
END;
$$ LANGUAGE plpgsql;

-- Use function
SELECT calculate_discount(100.00, 15);  -- 85.00

-- Function returning table
CREATE OR REPLACE FUNCTION get_department_stats(dept_id INTEGER)
RETURNS TABLE(total_employees BIGINT, avg_salary NUMERIC, max_salary NUMERIC) AS $$
BEGIN
    RETURN QUERY
    SELECT COUNT(*), AVG(salary), MAX(salary)
    FROM employees
    WHERE department_id = dept_id AND status = 'active';
END;
$$ LANGUAGE plpgsql;

-- Use table-returning function
SELECT * FROM get_department_stats(1);

-- Trigger function
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger
CREATE TRIGGER trg_update_timestamp
    BEFORE UPDATE ON employees
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();
```

---

## 16. Set Operations

```sql
-- UNION: combine results, remove duplicates
SELECT city FROM customers
UNION
SELECT city FROM suppliers;

-- UNION ALL: combine results, keep duplicates (faster)
SELECT city FROM customers
UNION ALL
SELECT city FROM suppliers;

-- INTERSECT: rows in both
SELECT email FROM users
INTERSECT
SELECT email FROM newsletter_subscribers;

-- EXCEPT: rows in first but not second (MINUS in Oracle)
SELECT email FROM users
EXCEPT
SELECT email FROM unsubscribed;

-- Rules:
-- 1. Same number of columns
-- 2. Compatible data types
-- 3. Column names come from first query
```

---

## 17. Common Table Expressions (CTEs)

```sql
-- Basic CTE
WITH active_orders AS (
    SELECT user_id, COUNT(*) AS order_count, SUM(total) AS total_spent
    FROM orders
    WHERE status != 'cancelled'
    GROUP BY user_id
)
SELECT u.first_name, u.email, ao.order_count, ao.total_spent
FROM users u
JOIN active_orders ao ON u.id = ao.user_id
WHERE ao.total_spent > 1000;

-- Multiple CTEs
WITH 
monthly_sales AS (
    SELECT 
        DATE_TRUNC('month', ordered_at) AS month,
        SUM(total) AS revenue
    FROM orders
    WHERE status = 'delivered'
    GROUP BY DATE_TRUNC('month', ordered_at)
),
avg_sales AS (
    SELECT AVG(revenue) AS avg_revenue FROM monthly_sales
)
SELECT 
    ms.month, 
    ms.revenue,
    a.avg_revenue,
    CASE WHEN ms.revenue > a.avg_revenue THEN 'Above' ELSE 'Below' END AS vs_avg
FROM monthly_sales ms, avg_sales a
ORDER BY ms.month;

-- CTE with INSERT (writable CTE, PostgreSQL)
WITH deleted_rows AS (
    DELETE FROM sessions
    WHERE expires_at < CURRENT_TIMESTAMP
    RETURNING *
)
INSERT INTO session_archive
SELECT * FROM deleted_rows;
```

### Interview Tip: CTE vs Subquery

| CTE | Subquery |
|-----|----------|
| More readable | Inline, compact |
| Can be referenced multiple times | Must repeat if needed twice |
| Can be recursive | Cannot be recursive |
| Named and documented | Anonymous |
| PostgreSQL: may/may not materialize | Optimizer can inline |
