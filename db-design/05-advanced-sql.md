# Advanced SQL — Deep Dive

## Recursive CTEs, Pivoting, JSON, Full-Text Search, and Query Optimization

---

## 1. Recursive CTEs

### Syntax

```sql
WITH RECURSIVE cte_name AS (
    -- Base case (non-recursive term)
    SELECT ...
    
    UNION ALL  -- or UNION (removes duplicates between iterations)
    
    -- Recursive case (references cte_name)
    SELECT ...
    FROM cte_name
    JOIN ...
)
SELECT * FROM cte_name;
```

### Example 1: Employee Hierarchy (Org Chart)

```sql
-- Find the complete org tree starting from the CEO
WITH RECURSIVE org_tree AS (
    -- Base: CEO (no manager)
    SELECT id, name, manager_id, 1 AS level, 
           name::TEXT AS path
    FROM employees
    WHERE manager_id IS NULL
    
    UNION ALL
    
    -- Recursive: find direct reports of current level
    SELECT e.id, e.name, e.manager_id, ot.level + 1,
           ot.path || ' → ' || e.name
    FROM employees e
    INNER JOIN org_tree ot ON e.manager_id = ot.id
)
SELECT level, REPEAT('  ', level - 1) || name AS org_chart, path
FROM org_tree
ORDER BY path;

-- Result:
-- level | org_chart         | path
-- 1     | Alice             | Alice
-- 2     |   Bob             | Alice → Bob
-- 2     |   Carol           | Alice → Carol
-- 1     | Dave              | Dave
-- 2     |   Eve             | Dave → Eve
-- 1     | Frank             | Frank
-- 2     |   Grace           | Frank → Grace
```

### Example 2: Category Tree (Nested Categories)

```sql
-- Build full category path
WITH RECURSIVE category_tree AS (
    -- Base: root categories (no parent)
    SELECT id, name, parent_id, name::TEXT AS full_path, 1 AS depth
    FROM categories
    WHERE parent_id IS NULL
    
    UNION ALL
    
    -- Recursive: child categories
    SELECT c.id, c.name, c.parent_id, ct.full_path || ' > ' || c.name, ct.depth + 1
    FROM categories c
    INNER JOIN category_tree ct ON c.parent_id = ct.id
)
SELECT depth, full_path FROM category_tree ORDER BY full_path;

-- Result:
-- 1 | Electronics
-- 2 | Electronics > Computers
-- 3 | Electronics > Computers > Laptops
-- 3 | Electronics > Computers > Desktops
-- 2 | Electronics > Phones
-- 1 | Clothing
-- 2 | Clothing > Men
-- 2 | Clothing > Women
```

### Example 3: Generate Number Series

```sql
-- Generate numbers 1 to 100
WITH RECURSIVE nums AS (
    SELECT 1 AS n
    UNION ALL
    SELECT n + 1 FROM nums WHERE n < 100
)
SELECT n FROM nums;

-- Generate dates for a year
WITH RECURSIVE dates AS (
    SELECT '2024-01-01'::DATE AS d
    UNION ALL
    SELECT d + 1 FROM dates WHERE d < '2024-12-31'
)
SELECT d, EXTRACT(DOW FROM d) AS day_of_week, TO_CHAR(d, 'Day') AS day_name
FROM dates;
```

### Example 4: Bill of Materials (BOM) Explosion

```sql
-- Product has components; components may have sub-components
CREATE TABLE bom (
    parent_part_id  INTEGER,
    child_part_id   INTEGER,
    quantity        INTEGER,
    PRIMARY KEY (parent_part_id, child_part_id)
);

-- Explode BOM for a product
WITH RECURSIVE bom_tree AS (
    SELECT child_part_id AS part_id, quantity, 1 AS level
    FROM bom
    WHERE parent_part_id = 100  -- product #100
    
    UNION ALL
    
    SELECT b.child_part_id, b.quantity * bt.quantity, bt.level + 1
    FROM bom b
    INNER JOIN bom_tree bt ON b.parent_part_id = bt.part_id
)
SELECT part_id, SUM(quantity) AS total_needed, MAX(level) AS max_depth
FROM bom_tree
GROUP BY part_id
ORDER BY part_id;
```

### Example 5: Shortest Path in Graph

```sql
-- Find shortest path between two nodes in a graph
CREATE TABLE edges (from_node INT, to_node INT, weight INT);

WITH RECURSIVE shortest_path AS (
    SELECT to_node, weight, ARRAY[from_node, to_node] AS path
    FROM edges
    WHERE from_node = 1  -- start node
    
    UNION ALL
    
    SELECT e.to_node, sp.weight + e.weight, sp.path || e.to_node
    FROM edges e
    INNER JOIN shortest_path sp ON e.from_node = sp.to_node
    WHERE NOT e.to_node = ANY(sp.path)  -- prevent cycles
)
SELECT to_node, MIN(weight) AS shortest_distance, 
       (ARRAY_AGG(path ORDER BY weight))[1] AS shortest_path
FROM shortest_path
GROUP BY to_node
ORDER BY shortest_distance;
```

### Infinite Loop Protection

```sql
-- Always add a safety limit!
WITH RECURSIVE cte AS (
    SELECT 1 AS n
    UNION ALL
    SELECT n + 1 FROM cte WHERE n < 1000  -- explicit limit
)
SELECT * FROM cte;

-- Or use PostgreSQL's max recursion depth
SET max_recursive_iterations = 100;
```

---

## 2. Pivoting & Unpivoting

### Pivot — Rows to Columns

```sql
-- Given monthly sales data:
-- | region | month    | revenue |
-- | East   | January  | 10000   |
-- | East   | February | 12000   |
-- | West   | January  | 15000   |
-- | West   | February | 11000   |

-- Pivot using CASE + GROUP BY (universal approach):
SELECT 
    region,
    SUM(CASE WHEN month = 'January'  THEN revenue ELSE 0 END) AS jan,
    SUM(CASE WHEN month = 'February' THEN revenue ELSE 0 END) AS feb,
    SUM(CASE WHEN month = 'March'    THEN revenue ELSE 0 END) AS mar,
    SUM(CASE WHEN month = 'April'    THEN revenue ELSE 0 END) AS apr
FROM monthly_sales
GROUP BY region;

-- Result:
-- | region | jan   | feb   | mar   | apr  |
-- | East   | 10000 | 12000 | 8000  | 9000 |
-- | West   | 15000 | 11000 | 13000 | 7000 |

-- Oracle / SQL Server: PIVOT syntax
SELECT * FROM monthly_sales
PIVOT (
    SUM(revenue) FOR month IN ('January' AS jan, 'February' AS feb, 'March' AS mar)
) p;

-- Dynamic pivot with crosstab (PostgreSQL extension)
-- Requires: CREATE EXTENSION tablefunc;
SELECT * FROM crosstab(
    'SELECT region, month, revenue FROM monthly_sales ORDER BY 1, 2',
    'SELECT DISTINCT month FROM monthly_sales ORDER BY 1'
) AS ct(region TEXT, jan NUMERIC, feb NUMERIC, mar NUMERIC);
```

### Unpivot — Columns to Rows

```sql
-- Given pivoted data:
-- | region | q1_revenue | q2_revenue | q3_revenue | q4_revenue |
-- | East   | 30000      | 35000      | 28000      | 40000      |

-- Unpivot using UNION ALL (universal):
SELECT region, 'Q1' AS quarter, q1_revenue AS revenue FROM quarterly_sales
UNION ALL
SELECT region, 'Q2', q2_revenue FROM quarterly_sales
UNION ALL
SELECT region, 'Q3', q3_revenue FROM quarterly_sales
UNION ALL
SELECT region, 'Q4', q4_revenue FROM quarterly_sales;

-- PostgreSQL: UNNEST
SELECT region, quarter, revenue
FROM quarterly_sales,
LATERAL UNNEST(
    ARRAY['Q1', 'Q2', 'Q3', 'Q4'],
    ARRAY[q1_revenue, q2_revenue, q3_revenue, q4_revenue]
) AS t(quarter, revenue);

-- SQL Server: UNPIVOT
SELECT region, quarter, revenue
FROM quarterly_sales
UNPIVOT (revenue FOR quarter IN (q1_revenue, q2_revenue, q3_revenue, q4_revenue)) u;
```

---

## 3. JSON in SQL (PostgreSQL)

### JSON Data Types

```sql
-- JSON: stores as text, validates format
-- JSONB: binary format, faster queries, indexable
-- Always prefer JSONB unless you need key ordering

CREATE TABLE events (
    id      SERIAL PRIMARY KEY,
    type    VARCHAR(50),
    payload JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO events (type, payload) VALUES
('user_signup', '{"user_id": 1, "name": "Alice", "email": "alice@example.com", "tags": ["premium", "early_adopter"]}'),
('purchase', '{"user_id": 1, "product": {"id": 42, "name": "Widget"}, "amount": 29.99}'),
('user_signup', '{"user_id": 2, "name": "Bob", "email": "bob@example.com", "tags": ["basic"]}');
```

### Accessing JSON Data

```sql
-- -> returns JSON element
-- ->> returns TEXT element
-- #> path returns JSON
-- #>> path returns TEXT

-- Access top-level key
SELECT payload -> 'name' AS name_json FROM events;          -- "Alice" (JSON)
SELECT payload ->> 'name' AS name_text FROM events;          -- Alice (text)

-- Access nested key
SELECT payload -> 'product' -> 'name' FROM events;           -- "Widget" (JSON)
SELECT payload -> 'product' ->> 'name' FROM events;          -- Widget (text)
SELECT payload #>> '{product, name}' FROM events;             -- Widget (path)

-- Access array element
SELECT payload -> 'tags' -> 0 FROM events;                    -- "premium"
SELECT payload -> 'tags' ->> 0 FROM events;                   -- premium

-- Filter by JSON value
SELECT * FROM events WHERE payload ->> 'name' = 'Alice';
SELECT * FROM events WHERE (payload ->> 'amount')::DECIMAL > 20;

-- Check if key exists
SELECT * FROM events WHERE payload ? 'amount';                -- has 'amount' key
SELECT * FROM events WHERE payload ?| ARRAY['amount', 'qty']; -- has any key
SELECT * FROM events WHERE payload ?& ARRAY['name', 'email']; -- has all keys

-- Containment
SELECT * FROM events WHERE payload @> '{"user_id": 1}';      -- contains
SELECT * FROM events WHERE '{"user_id": 1}' <@ payload;      -- is contained by

-- Array containment
SELECT * FROM events WHERE payload -> 'tags' ? 'premium';    -- array contains 'premium'
```

### Modifying JSON

```sql
-- Set a key
UPDATE events SET payload = payload || '{"verified": true}'
WHERE payload ->> 'name' = 'Alice';

-- Remove a key
UPDATE events SET payload = payload - 'email'
WHERE payload ->> 'name' = 'Bob';

-- Set nested key
UPDATE events SET payload = jsonb_set(payload, '{product, price}', '29.99')
WHERE type = 'purchase';

-- Build JSON
SELECT jsonb_build_object('name', first_name, 'salary', salary) FROM employees;
-- {"name": "Alice", "salary": 150000}

-- Aggregate to JSON array
SELECT jsonb_agg(jsonb_build_object('id', id, 'name', name)) FROM employees;
-- [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}, ...]

-- Expand JSON to rows
SELECT id, key, value
FROM events, jsonb_each(payload)
WHERE type = 'user_signup';

-- Expand JSON array to rows
SELECT id, tag
FROM events, jsonb_array_elements_text(payload -> 'tags') AS tag
WHERE type = 'user_signup';
```

### Indexing JSON

```sql
-- GIN index on entire JSONB column (supports @>, ?, ?|, ?&)
CREATE INDEX idx_events_payload ON events USING GIN(payload);

-- GIN index on specific path
CREATE INDEX idx_events_user_id ON events USING GIN((payload -> 'user_id'));

-- B-tree index on specific extracted value
CREATE INDEX idx_events_name ON events((payload ->> 'name'));

-- Use for: WHERE payload @> '{"user_id": 1}'
-- Use for: WHERE payload ->> 'name' = 'Alice'
```

---

## 4. Full-Text Search (PostgreSQL)

### Basic Full-Text Search

```sql
-- tsvector: processed document
-- tsquery: processed query
-- @@: match operator

-- Simple search
SELECT * FROM products
WHERE to_tsvector('english', name || ' ' || description) @@ to_tsquery('english', 'wireless & bluetooth');

-- Ranking results
SELECT 
    name, description,
    ts_rank(to_tsvector('english', name || ' ' || description), 
            to_tsquery('english', 'wireless & bluetooth')) AS rank
FROM products
WHERE to_tsvector('english', name || ' ' || description) @@ to_tsquery('english', 'wireless & bluetooth')
ORDER BY rank DESC;

-- Index for full-text search
ALTER TABLE products ADD COLUMN search_vector TSVECTOR;
UPDATE products SET search_vector = to_tsvector('english', name || ' ' || COALESCE(description, ''));
CREATE INDEX idx_products_search ON products USING GIN(search_vector);

-- Query with index
SELECT * FROM products WHERE search_vector @@ to_tsquery('english', 'wireless');

-- Auto-update search_vector with trigger
CREATE TRIGGER trg_update_search_vector
BEFORE INSERT OR UPDATE ON products
FOR EACH ROW EXECUTE FUNCTION
tsvector_update_trigger(search_vector, 'pg_catalog.english', name, description);
```

### tsquery Operators

```sql
-- AND: both terms
to_tsquery('english', 'coffee & organic')

-- OR: either term
to_tsquery('english', 'coffee | tea')

-- NOT: exclude term
to_tsquery('english', 'coffee & !decaf')

-- Phrase search (adjacent words)
to_tsquery('english', 'dark <-> roast')    -- "dark roast"
to_tsquery('english', 'dark <2> roast')    -- "dark" within 2 words of "roast"

-- Prefix matching
to_tsquery('english', 'wire:*')            -- wireless, wired, wire, etc.

-- plainto_tsquery: simpler input
plainto_tsquery('english', 'dark roast coffee')  -- dark & roast & coffee
-- websearch_to_tsquery: Google-like input
websearch_to_tsquery('english', '"dark roast" coffee -decaf')
```

---

## 5. Query Optimization

### EXPLAIN ANALYZE Deep Dive

```sql
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT e.name, d.name
FROM employees e
JOIN departments d ON e.department_id = d.id
WHERE e.salary > 100000;

-- Output breakdown:
-- Hash Join  (cost=1.09..2.31 rows=3 width=32) (actual time=0.025..0.030 rows=4 loops=1)
--   Hash Cond: (e.department_id = d.id)
--   ->  Seq Scan on employees e  (cost=0.00..1.10 rows=3 width=20) (actual time=0.008..0.010 rows=4 loops=1)
--         Filter: (salary > 100000)
--         Rows Removed by Filter: 4
--   ->  Hash  (cost=1.04..1.04 rows=4 width=20) (actual time=0.006..0.006 rows=4 loops=1)
--         Buckets: 1024  Batches: 1  Memory Usage: 9kB
--         ->  Seq Scan on departments d  (cost=0.00..1.04 rows=4 width=20)
-- Planning Time: 0.150 ms
-- Execution Time: 0.060 ms
```

### What to Look For

| Symbol | Meaning | Action |
|--------|---------|--------|
| `Seq Scan` | Full table scan | Consider adding index |
| `Index Scan` | Using index | Good ✅ |
| `Index Only Scan` | All data from index | Best ✅ |
| `Bitmap Index Scan` | Multiple index entries | OK, combining indexes |
| `Nested Loop` | O(n×m) | OK for small tables; bad for large |
| `Hash Join` | Build hash + probe | Good for medium-large tables |
| `Merge Join` | Both sorted, merge | Good for pre-sorted data |
| `Sort` | Explicit sort | Check if index can avoid |
| `Rows Removed by Filter` | Rows read but discarded | Index might eliminate these |

### Common Optimization Patterns

#### 1. Index-Based Optimizations

```sql
-- ❌ SLOW: Function on indexed column defeats index
SELECT * FROM users WHERE LOWER(email) = 'alice@example.com';
-- ✅ FIX: Expression index
CREATE INDEX idx_users_lower_email ON users(LOWER(email));

-- ❌ SLOW: Leading wildcard can't use B-tree index
SELECT * FROM products WHERE name LIKE '%Widget%';
-- ✅ FIX: Use full-text search or trigram index
CREATE INDEX idx_products_name_trgm ON products USING GIN(name gin_trgm_ops);
-- Requires: CREATE EXTENSION pg_trgm;

-- ❌ SLOW: OR conditions on different columns
SELECT * FROM products WHERE category_id = 5 OR seller_id = 10;
-- ✅ FIX: Rewrite as UNION
SELECT * FROM products WHERE category_id = 5
UNION
SELECT * FROM products WHERE seller_id = 10;
```

#### 2. Query Rewriting

```sql
-- ❌ SLOW: Correlated subquery in SELECT (executes once per row)
SELECT name, (SELECT COUNT(*) FROM orders o WHERE o.user_id = u.id) AS order_count
FROM users u;
-- ✅ FIX: Use JOIN + GROUP BY
SELECT u.name, COUNT(o.id) AS order_count
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
GROUP BY u.id, u.name;

-- ❌ SLOW: NOT IN with NULLs
SELECT * FROM employees WHERE id NOT IN (SELECT manager_id FROM employees);
-- ✅ FIX: Use NOT EXISTS
SELECT * FROM employees e
WHERE NOT EXISTS (SELECT 1 FROM employees e2 WHERE e2.manager_id = e.id);

-- ❌ SLOW: SELECT *
SELECT * FROM orders JOIN order_items ON orders.id = order_items.order_id;
-- ✅ FIX: Select only needed columns
SELECT o.id, o.total, oi.product_id, oi.quantity FROM orders o JOIN order_items oi ON o.id = oi.order_id;

-- ❌ SLOW: OFFSET for deep pagination
SELECT * FROM products ORDER BY id LIMIT 20 OFFSET 10000;  -- scans 10020 rows!
-- ✅ FIX: Keyset pagination
SELECT * FROM products WHERE id > 10000 ORDER BY id LIMIT 20;
```

#### 3. N+1 Query Problem

```sql
-- ❌ N+1 Problem (in application code):
-- Query 1: SELECT * FROM users LIMIT 100
-- For each user:
--   Query 2-101: SELECT * FROM orders WHERE user_id = ?

-- ✅ FIX: Single query with JOIN
SELECT u.*, o.id AS order_id, o.total
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
ORDER BY u.id;

-- Or batch with IN:
SELECT * FROM orders WHERE user_id IN (1, 2, 3, ..., 100);
```

---

## 6. Temporary Tables & CTEs

```sql
-- Temporary table (session-scoped)
CREATE TEMP TABLE temp_high_earners AS
SELECT * FROM employees WHERE salary > 100000;

-- Use it
SELECT * FROM temp_high_earners;

-- Temp table with explicit structure
CREATE TEMP TABLE temp_results (
    id INTEGER,
    metric DECIMAL(10,2),
    label VARCHAR(50)
);

-- Unlogged table (no WAL, faster writes, lost on crash)
CREATE UNLOGGED TABLE staging_data (
    id SERIAL PRIMARY KEY,
    raw_data JSONB
);

-- CTE with INSERT/UPDATE/DELETE (writable CTEs, PostgreSQL)
WITH moved_rows AS (
    DELETE FROM active_orders
    WHERE status = 'completed'
    RETURNING *
)
INSERT INTO order_archive
SELECT * FROM moved_rows;
```

---

## 7. String Aggregation & Array Operations

### String Aggregation

```sql
-- Concatenate values from multiple rows
-- PostgreSQL:
SELECT department_id, STRING_AGG(name, ', ' ORDER BY name) AS employees
FROM employees
GROUP BY department_id;
-- Result: 1 | 'Alice, Bob, Carol'
--         2 | 'Dave, Eve'

-- MySQL:
SELECT department_id, GROUP_CONCAT(name ORDER BY name SEPARATOR ', ') AS employees
FROM employees
GROUP BY department_id;
```

### Array Operations (PostgreSQL)

```sql
-- Create arrays
SELECT ARRAY[1, 2, 3];
SELECT ARRAY_AGG(id) FROM employees WHERE department_id = 1;  -- {1, 2, 3}

-- Array functions
SELECT ARRAY_LENGTH(ARRAY[1, 2, 3], 1);  -- 3
SELECT UNNEST(ARRAY['a', 'b', 'c']);       -- expands to 3 rows
SELECT ARRAY_APPEND(ARRAY[1, 2], 3);       -- {1, 2, 3}
SELECT ARRAY_REMOVE(ARRAY[1, 2, 3], 2);    -- {1, 3}
SELECT ARRAY_CAT(ARRAY[1, 2], ARRAY[3, 4]); -- {1, 2, 3, 4}

-- Array containment
SELECT ARRAY[1, 2] <@ ARRAY[1, 2, 3];     -- true (is contained by)
SELECT ARRAY[1, 2, 3] @> ARRAY[1, 2];     -- true (contains)
SELECT ARRAY[1, 2] && ARRAY[2, 3];         -- true (overlap/intersection)

-- Use case: Tags
CREATE TABLE articles (
    id SERIAL PRIMARY KEY,
    title TEXT,
    tags TEXT[]
);
INSERT INTO articles (title, tags) VALUES ('SQL Guide', ARRAY['sql', 'database', 'tutorial']);

SELECT * FROM articles WHERE tags @> ARRAY['sql'];            -- has 'sql' tag
SELECT * FROM articles WHERE 'sql' = ANY(tags);               -- same
CREATE INDEX idx_articles_tags ON articles USING GIN(tags);    -- index for array queries
```

---

## 8. Common Table Expression Patterns

### CTE for Data Transformation Pipeline

```sql
-- Step-by-step transformation (like pipes)
WITH 
-- Step 1: Get raw data
raw_orders AS (
    SELECT user_id, product_id, quantity, unit_price, order_date
    FROM order_items oi
    JOIN orders o ON oi.order_id = o.id
    WHERE o.status = 'completed'
),
-- Step 2: Calculate totals per user per product
user_product_totals AS (
    SELECT 
        user_id, product_id,
        SUM(quantity) AS total_qty,
        SUM(quantity * unit_price) AS total_spent
    FROM raw_orders
    GROUP BY user_id, product_id
),
-- Step 3: Rank products per user
ranked AS (
    SELECT *,
        ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY total_spent DESC) AS product_rank
    FROM user_product_totals
)
-- Step 4: Get top 3 products per user
SELECT u.name, p.name AS product, r.total_qty, r.total_spent
FROM ranked r
JOIN users u ON r.user_id = u.id
JOIN products p ON r.product_id = p.id
WHERE r.product_rank <= 3
ORDER BY u.name, r.product_rank;
```

### CTE for Deduplication

```sql
-- Keep only the latest record per user (dedup by email)
WITH ranked AS (
    SELECT *,
        ROW_NUMBER() OVER (PARTITION BY email ORDER BY created_at DESC) AS rn
    FROM user_signups
)
DELETE FROM user_signups
WHERE id IN (SELECT id FROM ranked WHERE rn > 1);
```

---

## 9. Conditional Expressions Beyond CASE

```sql
-- COALESCE: first non-NULL
SELECT COALESCE(nickname, first_name, 'Anonymous') AS display_name FROM users;

-- NULLIF: return NULL if equal (avoid division by zero)
SELECT revenue / NULLIF(cost, 0) AS roi FROM financials;

-- GREATEST / LEAST
SELECT GREATEST(score1, score2, score3) AS best_score FROM exams;
SELECT LEAST(price, max_price) AS effective_price FROM products;

-- CAST / ::
SELECT '42'::INTEGER + 1;  -- 43
SELECT CAST('2024-01-15' AS DATE);
```

---

## 10. Advanced Patterns

### Gap and Island Detection

```sql
-- Find gaps in sequential IDs
WITH id_range AS (
    SELECT 
        id,
        LEAD(id) OVER (ORDER BY id) AS next_id
    FROM orders
)
SELECT id + 1 AS gap_start, next_id - 1 AS gap_end
FROM id_range
WHERE next_id - id > 1;

-- Find islands of consecutive dates
WITH date_groups AS (
    SELECT 
        user_id, login_date,
        login_date - (ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY login_date))::INT AS island_group
    FROM user_logins
)
SELECT 
    user_id,
    MIN(login_date) AS island_start,
    MAX(login_date) AS island_end,
    COUNT(*) AS consecutive_days
FROM date_groups
GROUP BY user_id, island_group
ORDER BY user_id, island_start;
```

### Recursive Fibonacci

```sql
WITH RECURSIVE fib AS (
    SELECT 0 AS n, 0::BIGINT AS fib_n, 1::BIGINT AS fib_next
    UNION ALL
    SELECT n + 1, fib_next, fib_n + fib_next
    FROM fib
    WHERE n < 20
)
SELECT n, fib_n FROM fib;
```

### Generate Calendar Table

```sql
WITH RECURSIVE calendar AS (
    SELECT '2024-01-01'::DATE AS dt
    UNION ALL
    SELECT dt + 1 FROM calendar WHERE dt < '2024-12-31'
)
SELECT 
    dt AS date,
    EXTRACT(YEAR FROM dt) AS year,
    EXTRACT(MONTH FROM dt) AS month,
    EXTRACT(DAY FROM dt) AS day,
    TO_CHAR(dt, 'Day') AS day_name,
    EXTRACT(DOW FROM dt) AS day_of_week,
    CASE WHEN EXTRACT(DOW FROM dt) IN (0, 6) THEN TRUE ELSE FALSE END AS is_weekend,
    EXTRACT(WEEK FROM dt) AS week_number,
    EXTRACT(QUARTER FROM dt) AS quarter
FROM calendar;
```

### Running Difference / Delta

```sql
-- Show price changes over time
SELECT 
    product_id, recorded_at, price,
    price - LAG(price) OVER (PARTITION BY product_id ORDER BY recorded_at) AS price_change,
    CASE 
        WHEN price > LAG(price) OVER (PARTITION BY product_id ORDER BY recorded_at) THEN '📈'
        WHEN price < LAG(price) OVER (PARTITION BY product_id ORDER BY recorded_at) THEN '📉'
        ELSE '➡️'
    END AS direction
FROM price_history;
```

### Materialized Path Pattern for Hierarchies

```sql
-- Alternative to adjacency list and recursive CTE
-- Store full path as string
CREATE TABLE categories_mp (
    id      SERIAL PRIMARY KEY,
    name    VARCHAR(100),
    path    VARCHAR(500)    -- e.g., '/1/5/12/'
);

-- Find all descendants of node 5
SELECT * FROM categories_mp WHERE path LIKE '%/5/%';

-- Find depth
SELECT *, LENGTH(path) - LENGTH(REPLACE(path, '/', '')) - 1 AS depth
FROM categories_mp;

-- Pros: Single query for subtree (no recursion)
-- Cons: Path update requires updating all descendants
```

---

## 11. Database Constraints Advanced

### Exclusion Constraints (PostgreSQL)

```sql
-- Prevent overlapping date ranges (e.g., room bookings)
CREATE EXTENSION btree_gist;

CREATE TABLE room_bookings (
    id          SERIAL PRIMARY KEY,
    room_id     INTEGER NOT NULL,
    start_date  DATE NOT NULL,
    end_date    DATE NOT NULL,
    guest_name  VARCHAR(200),
    EXCLUDE USING gist (
        room_id WITH =,
        daterange(start_date, end_date) WITH &&
    )
);

-- This prevents inserting a booking that overlaps with an existing booking
-- for the same room! The DB enforces this automatically.
```

### Conditional Constraints

```sql
-- Partial unique index: unique email per active user (allow duplicates for deleted)
CREATE UNIQUE INDEX idx_unique_active_email ON users(email) WHERE status = 'active';

-- Check constraints with functions
CREATE TABLE reservations (
    id          SERIAL PRIMARY KEY,
    start_time  TIMESTAMP NOT NULL,
    end_time    TIMESTAMP NOT NULL,
    CHECK (end_time > start_time),
    CHECK (end_time - start_time <= INTERVAL '8 hours')  -- max 8-hour reservation
);
```

---

## 12. Import/Export & Bulk Operations

```sql
-- COPY: Bulk import from CSV (PostgreSQL, much faster than INSERT)
COPY employees(first_name, last_name, email, salary)
FROM '/tmp/employees.csv'
WITH (FORMAT csv, HEADER true, DELIMITER ',');

-- COPY: Export to CSV
COPY (SELECT * FROM employees WHERE status = 'active')
TO '/tmp/active_employees.csv'
WITH (FORMAT csv, HEADER true);

-- \COPY from psql client (doesn't need superuser)
\copy employees FROM 'employees.csv' CSV HEADER

-- Bulk INSERT with generate_series
INSERT INTO test_data (value)
SELECT random() * 1000 FROM generate_series(1, 1000000);
```
