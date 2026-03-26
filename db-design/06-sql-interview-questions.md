# SQL Interview Questions — 50+ Problems with Solutions

## Organized by Difficulty: Easy → Medium → Hard

---

## Schema Used Across Questions

```sql
-- Employees & Departments
CREATE TABLE departments (id INT PK, name VARCHAR(100));
CREATE TABLE employees (
    id INT PK, name VARCHAR(100), department_id INT FK, 
    manager_id INT FK(self), salary DECIMAL(10,2), 
    hire_date DATE, status VARCHAR(20)
);

-- E-Commerce
CREATE TABLE users (id INT PK, name VARCHAR(100), email VARCHAR(255), city VARCHAR(100), joined_at TIMESTAMP);
CREATE TABLE products (id INT PK, name VARCHAR(200), category_id INT FK, price DECIMAL(10,2), stock INT);
CREATE TABLE categories (id INT PK, name VARCHAR(100), parent_id INT FK(self));
CREATE TABLE orders (id INT PK, user_id INT FK, total DECIMAL(10,2), status VARCHAR(20), ordered_at TIMESTAMP);
CREATE TABLE order_items (id INT PK, order_id INT FK, product_id INT FK, quantity INT, unit_price DECIMAL(10,2));
CREATE TABLE reviews (id INT PK, user_id INT FK, product_id INT FK, rating INT, comment TEXT, created_at TIMESTAMP);

-- School
CREATE TABLE students (id INT PK, name VARCHAR(100), department VARCHAR(50), enrollment_year INT);
CREATE TABLE courses (id INT PK, name VARCHAR(100), credits INT, department VARCHAR(50));
CREATE TABLE enrollments (student_id INT FK, course_id INT FK, grade CHAR(2), semester VARCHAR(20));
```

---

# EASY (1-15)

---

### Q1: Find all employees earning above $100,000

```sql
SELECT name, salary
FROM employees
WHERE salary > 100000
ORDER BY salary DESC;
```

---

### Q2: Count employees per department

```sql
SELECT d.name AS department, COUNT(e.id) AS employee_count
FROM departments d
LEFT JOIN employees e ON d.id = e.department_id
GROUP BY d.id, d.name
ORDER BY employee_count DESC;

-- LEFT JOIN ensures departments with 0 employees are included
-- COUNT(e.id) not COUNT(*) — COUNT(*) would give 1 for empty departments
```

---

### Q3: Find the highest salary in each department

```sql
SELECT d.name AS department, MAX(e.salary) AS highest_salary
FROM employees e
JOIN departments d ON e.department_id = d.id
GROUP BY d.id, d.name;
```

---

### Q4: List employees who have no manager

```sql
SELECT name, salary
FROM employees
WHERE manager_id IS NULL;
```

---

### Q5: Find duplicate emails

```sql
SELECT email, COUNT(*) AS count
FROM users
GROUP BY email
HAVING COUNT(*) > 1;
```

---

### Q6: Get the 3 most recent orders

```sql
SELECT id, user_id, total, status, ordered_at
FROM orders
ORDER BY ordered_at DESC
LIMIT 3;
```

---

### Q7: Find products that are out of stock

```sql
SELECT name, category_id, price
FROM products
WHERE stock = 0;
```

---

### Q8: Calculate average order value

```sql
SELECT ROUND(AVG(total), 2) AS avg_order_value
FROM orders
WHERE status != 'cancelled';
```

---

### Q9: Find users who have never placed an order

```sql
-- Method 1: LEFT JOIN + IS NULL (anti-join)
SELECT u.name, u.email
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE o.id IS NULL;

-- Method 2: NOT EXISTS
SELECT u.name, u.email
FROM users u
WHERE NOT EXISTS (SELECT 1 FROM orders o WHERE o.user_id = u.id);

-- Method 3: NOT IN (careful with NULLs!)
SELECT name, email
FROM users
WHERE id NOT IN (SELECT user_id FROM orders WHERE user_id IS NOT NULL);
```

---

### Q10: List employees hired in 2023

```sql
SELECT name, hire_date, salary
FROM employees
WHERE hire_date >= '2023-01-01' AND hire_date < '2024-01-01';

-- Or using EXTRACT:
SELECT name, hire_date, salary
FROM employees
WHERE EXTRACT(YEAR FROM hire_date) = 2023;
-- Note: EXTRACT may prevent index usage on hire_date
```

---

### Q11: Find the total revenue per product

```sql
SELECT p.name, SUM(oi.quantity * oi.unit_price) AS total_revenue
FROM products p
JOIN order_items oi ON p.id = oi.product_id
JOIN orders o ON oi.order_id = o.id
WHERE o.status NOT IN ('cancelled', 'refunded')
GROUP BY p.id, p.name
ORDER BY total_revenue DESC;
```

---

### Q12: Find departments with average salary above $100,000

```sql
SELECT d.name AS department, ROUND(AVG(e.salary), 2) AS avg_salary
FROM departments d
JOIN employees e ON d.id = e.department_id
GROUP BY d.id, d.name
HAVING AVG(e.salary) > 100000
ORDER BY avg_salary DESC;
```

---

### Q13: Get unique cities where users are located

```sql
SELECT DISTINCT city
FROM users
WHERE city IS NOT NULL
ORDER BY city;
```

---

### Q14: Update all products in 'Electronics' category to have 10% discount

```sql
UPDATE products
SET price = price * 0.90
WHERE category_id = (SELECT id FROM categories WHERE name = 'Electronics');

-- With RETURNING to see what changed:
UPDATE products
SET price = price * 0.90
WHERE category_id = (SELECT id FROM categories WHERE name = 'Electronics')
RETURNING id, name, price AS new_price;
```

---

### Q15: Delete orders older than 1 year that were cancelled

```sql
DELETE FROM order_items
WHERE order_id IN (
    SELECT id FROM orders
    WHERE status = 'cancelled' AND ordered_at < CURRENT_DATE - INTERVAL '1 year'
);

DELETE FROM orders
WHERE status = 'cancelled' AND ordered_at < CURRENT_DATE - INTERVAL '1 year';
```

---

# MEDIUM (16-35)

---

### Q16: Find the second highest salary

```sql
-- Method 1: OFFSET
SELECT DISTINCT salary
FROM employees
ORDER BY salary DESC
LIMIT 1 OFFSET 1;

-- Method 2: Subquery
SELECT MAX(salary) AS second_highest
FROM employees
WHERE salary < (SELECT MAX(salary) FROM employees);

-- Method 3: DENSE_RANK (handles ties properly)
SELECT salary FROM (
    SELECT salary, DENSE_RANK() OVER (ORDER BY salary DESC) AS rnk
    FROM employees
) ranked WHERE rnk = 2;

-- Method 4: Nth highest (generic)
-- For Nth highest, change rnk = N
```

---

### Q17: Find the top 2 highest-paid employees per department

```sql
SELECT department_name, name, salary, rnk FROM (
    SELECT 
        d.name AS department_name,
        e.name,
        e.salary,
        DENSE_RANK() OVER (PARTITION BY e.department_id ORDER BY e.salary DESC) AS rnk
    FROM employees e
    JOIN departments d ON e.department_id = d.id
) ranked
WHERE rnk <= 2
ORDER BY department_name, rnk;

-- Use DENSE_RANK for ties included
-- Use ROW_NUMBER for exactly N employees (no ties)
```

---

### Q18: Month-over-month revenue growth

```sql
WITH monthly_revenue AS (
    SELECT 
        DATE_TRUNC('month', ordered_at) AS month,
        SUM(total) AS revenue
    FROM orders
    WHERE status NOT IN ('cancelled', 'refunded')
    GROUP BY DATE_TRUNC('month', ordered_at)
)
SELECT 
    month,
    revenue,
    LAG(revenue) OVER (ORDER BY month) AS prev_month_revenue,
    revenue - LAG(revenue) OVER (ORDER BY month) AS absolute_change,
    ROUND(
        (revenue - LAG(revenue) OVER (ORDER BY month)) * 100.0 
        / LAG(revenue) OVER (ORDER BY month), 2
    ) AS pct_change
FROM monthly_revenue
ORDER BY month;
```

---

### Q19: Find employees earning more than their manager

```sql
SELECT 
    e.name AS employee,
    e.salary AS emp_salary,
    m.name AS manager,
    m.salary AS mgr_salary
FROM employees e
INNER JOIN employees m ON e.manager_id = m.id
WHERE e.salary > m.salary;
```

---

### Q20: Customers who bought ALL products in a category

```sql
-- Relational division: who bought every product in 'Electronics'?
SELECT u.name
FROM users u
WHERE NOT EXISTS (
    -- Products in Electronics that user did NOT buy
    SELECT p.id
    FROM products p
    WHERE p.category_id = (SELECT id FROM categories WHERE name = 'Electronics')
    EXCEPT
    SELECT oi.product_id
    FROM order_items oi
    JOIN orders o ON oi.order_id = o.id
    WHERE o.user_id = u.id
);

-- Alternative approach: COUNT matching = total count
WITH electronics_products AS (
    SELECT id FROM products 
    WHERE category_id = (SELECT id FROM categories WHERE name = 'Electronics')
),
user_electronics AS (
    SELECT DISTINCT o.user_id, oi.product_id
    FROM order_items oi
    JOIN orders o ON oi.order_id = o.id
    WHERE oi.product_id IN (SELECT id FROM electronics_products)
)
SELECT u.name
FROM users u
JOIN user_electronics ue ON u.id = ue.user_id
GROUP BY u.id, u.name
HAVING COUNT(DISTINCT ue.product_id) = (SELECT COUNT(*) FROM electronics_products);
```

---

### Q21: Running total of orders per user

```sql
SELECT 
    u.name,
    o.ordered_at,
    o.total,
    SUM(o.total) OVER (
        PARTITION BY o.user_id 
        ORDER BY o.ordered_at
        ROWS UNBOUNDED PRECEDING
    ) AS running_total,
    COUNT(*) OVER (
        PARTITION BY o.user_id 
        ORDER BY o.ordered_at
        ROWS UNBOUNDED PRECEDING
    ) AS order_number
FROM orders o
JOIN users u ON o.user_id = u.id
WHERE o.status != 'cancelled'
ORDER BY u.name, o.ordered_at;
```

---

### Q22: Find products reviewed but never purchased

```sql
SELECT p.name
FROM products p
WHERE EXISTS (SELECT 1 FROM reviews r WHERE r.product_id = p.id)
  AND NOT EXISTS (
      SELECT 1 FROM order_items oi WHERE oi.product_id = p.id
  );
```

---

### Q23: Pivot: Monthly sales by category

```sql
SELECT 
    c.name AS category,
    SUM(CASE WHEN EXTRACT(MONTH FROM o.ordered_at) = 1 THEN oi.quantity * oi.unit_price ELSE 0 END) AS jan,
    SUM(CASE WHEN EXTRACT(MONTH FROM o.ordered_at) = 2 THEN oi.quantity * oi.unit_price ELSE 0 END) AS feb,
    SUM(CASE WHEN EXTRACT(MONTH FROM o.ordered_at) = 3 THEN oi.quantity * oi.unit_price ELSE 0 END) AS mar,
    SUM(CASE WHEN EXTRACT(MONTH FROM o.ordered_at) = 4 THEN oi.quantity * oi.unit_price ELSE 0 END) AS apr,
    SUM(CASE WHEN EXTRACT(MONTH FROM o.ordered_at) = 5 THEN oi.quantity * oi.unit_price ELSE 0 END) AS may,
    SUM(CASE WHEN EXTRACT(MONTH FROM o.ordered_at) = 6 THEN oi.quantity * oi.unit_price ELSE 0 END) AS jun
FROM categories c
JOIN products p ON c.id = p.category_id
JOIN order_items oi ON p.id = oi.product_id
JOIN orders o ON oi.order_id = o.id
WHERE EXTRACT(YEAR FROM o.ordered_at) = 2024
GROUP BY c.id, c.name
ORDER BY c.name;
```

---

### Q24: Find gaps in sequential order IDs

```sql
WITH order_gaps AS (
    SELECT 
        id,
        LEAD(id) OVER (ORDER BY id) AS next_id
    FROM orders
)
SELECT 
    id AS gap_after,
    next_id AS gap_before,
    next_id - id - 1 AS missing_count
FROM order_gaps
WHERE next_id - id > 1
ORDER BY id;
```

---

### Q25: Self-join: Find employees in the same department earning within 10% of each other

```sql
SELECT 
    a.name AS employee_1,
    b.name AS employee_2,
    a.salary AS salary_1,
    b.salary AS salary_2,
    d.name AS department
FROM employees a
JOIN employees b ON a.department_id = b.department_id AND a.id < b.id
JOIN departments d ON a.department_id = d.id
WHERE ABS(a.salary - b.salary) <= 0.10 * GREATEST(a.salary, b.salary);
```

---

### Q26: Calculate moving average of daily revenue (7-day)

```sql
WITH daily_revenue AS (
    SELECT 
        DATE(ordered_at) AS order_date,
        SUM(total) AS daily_total
    FROM orders
    WHERE status = 'completed'
    GROUP BY DATE(ordered_at)
)
SELECT 
    order_date,
    daily_total,
    ROUND(AVG(daily_total) OVER (
        ORDER BY order_date
        ROWS BETWEEN 6 PRECEDING AND CURRENT ROW
    ), 2) AS moving_avg_7day
FROM daily_revenue
ORDER BY order_date;
```

---

### Q27: Find the most popular product pair (frequently bought together)

```sql
SELECT 
    p1.name AS product_1,
    p2.name AS product_2,
    COUNT(*) AS times_bought_together
FROM order_items oi1
JOIN order_items oi2 ON oi1.order_id = oi2.order_id AND oi1.product_id < oi2.product_id
JOIN products p1 ON oi1.product_id = p1.id
JOIN products p2 ON oi2.product_id = p2.id
GROUP BY p1.name, p2.name
ORDER BY times_bought_together DESC
LIMIT 10;
```

---

### Q28: Find users who placed orders in 3 consecutive months

```sql
WITH user_months AS (
    SELECT DISTINCT 
        user_id,
        DATE_TRUNC('month', ordered_at)::DATE AS order_month
    FROM orders
    WHERE status != 'cancelled'
),
with_gaps AS (
    SELECT 
        user_id, order_month,
        order_month - (ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY order_month) * INTERVAL '1 month')::DATE AS grp
    FROM user_months
)
SELECT DISTINCT u.name, wg.user_id
FROM with_gaps wg
JOIN users u ON wg.user_id = u.id
GROUP BY wg.user_id, wg.grp, u.name
HAVING COUNT(*) >= 3;
```

---

### Q29: Cumulative percentage of revenue by product

```sql
WITH product_revenue AS (
    SELECT 
        p.name,
        SUM(oi.quantity * oi.unit_price) AS revenue
    FROM products p
    JOIN order_items oi ON p.id = oi.product_id
    GROUP BY p.id, p.name
)
SELECT 
    name,
    revenue,
    ROUND(revenue * 100.0 / SUM(revenue) OVER (), 2) AS pct_of_total,
    ROUND(SUM(revenue) OVER (ORDER BY revenue DESC ROWS UNBOUNDED PRECEDING) * 100.0 
        / SUM(revenue) OVER (), 2) AS cumulative_pct
FROM product_revenue
ORDER BY revenue DESC;

-- Useful for Pareto analysis (80/20 rule)
```

---

### Q30: Average time between consecutive orders per user

```sql
WITH order_gaps AS (
    SELECT 
        user_id,
        ordered_at,
        LAG(ordered_at) OVER (PARTITION BY user_id ORDER BY ordered_at) AS prev_order,
        ordered_at - LAG(ordered_at) OVER (PARTITION BY user_id ORDER BY ordered_at) AS gap
    FROM orders
    WHERE status != 'cancelled'
)
SELECT 
    u.name,
    COUNT(*) - 1 AS order_gaps_count,
    AVG(gap) AS avg_time_between_orders,
    MIN(gap) AS shortest_gap,
    MAX(gap) AS longest_gap
FROM order_gaps og
JOIN users u ON og.user_id = u.id
WHERE og.prev_order IS NOT NULL
GROUP BY u.id, u.name
HAVING COUNT(*) > 1
ORDER BY avg_time_between_orders;
```

---

### Q31: Students who scored above average in ALL their courses

```sql
WITH course_averages AS (
    SELECT course_id, AVG(
        CASE grade
            WHEN 'A+' THEN 4.3 WHEN 'A' THEN 4.0 WHEN 'A-' THEN 3.7
            WHEN 'B+' THEN 3.3 WHEN 'B' THEN 3.0 WHEN 'B-' THEN 2.7
            WHEN 'C+' THEN 2.3 WHEN 'C' THEN 2.0 WHEN 'C-' THEN 1.7
            WHEN 'D' THEN 1.0 WHEN 'F' THEN 0.0
        END
    ) AS avg_gpa
    FROM enrollments
    GROUP BY course_id
),
student_results AS (
    SELECT 
        e.student_id, e.course_id,
        CASE e.grade
            WHEN 'A+' THEN 4.3 WHEN 'A' THEN 4.0 WHEN 'A-' THEN 3.7
            WHEN 'B+' THEN 3.3 WHEN 'B' THEN 3.0 WHEN 'B-' THEN 2.7
            WHEN 'C+' THEN 2.3 WHEN 'C' THEN 2.0 WHEN 'C-' THEN 1.7
            WHEN 'D' THEN 1.0 WHEN 'F' THEN 0.0
        END AS student_gpa,
        ca.avg_gpa
    FROM enrollments e
    JOIN course_averages ca ON e.course_id = ca.course_id
)
SELECT s.name
FROM students s
WHERE NOT EXISTS (
    SELECT 1 FROM student_results sr
    WHERE sr.student_id = s.id AND sr.student_gpa <= sr.avg_gpa
);
```

---

### Q32: Delete duplicate rows, keeping the first (by id)

```sql
-- Find duplicates
SELECT email, COUNT(*), ARRAY_AGG(id ORDER BY id) AS ids
FROM users
GROUP BY email
HAVING COUNT(*) > 1;

-- Delete duplicates (keep lowest id)
DELETE FROM users
WHERE id NOT IN (
    SELECT MIN(id) FROM users GROUP BY email
);

-- PostgreSQL efficient approach with ctid
DELETE FROM users a
USING users b
WHERE a.email = b.email AND a.id > b.id;

-- Using window function
WITH ranked AS (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY email ORDER BY id) AS rn
    FROM users
)
DELETE FROM users WHERE id IN (SELECT id FROM ranked WHERE rn > 1);
```

---

### Q33: Recursive: Find all subcategories of a given category

```sql
WITH RECURSIVE category_tree AS (
    SELECT id, name, parent_id, 1 AS depth, name::TEXT AS path
    FROM categories
    WHERE name = 'Electronics'  -- root category
    
    UNION ALL
    
    SELECT c.id, c.name, c.parent_id, ct.depth + 1, ct.path || ' > ' || c.name
    FROM categories c
    JOIN category_tree ct ON c.parent_id = ct.id
)
SELECT depth, path FROM category_tree ORDER BY path;

-- Result:
-- 1 | Electronics
-- 2 | Electronics > Computers
-- 3 | Electronics > Computers > Laptops
-- 3 | Electronics > Computers > Desktops
-- 2 | Electronics > Phones
-- 3 | Electronics > Phones > Smartphones
```

---

### Q34: Find users whose total spending increased every month

```sql
WITH monthly_spending AS (
    SELECT 
        user_id,
        DATE_TRUNC('month', ordered_at) AS month,
        SUM(total) AS monthly_total
    FROM orders
    WHERE status = 'completed'
    GROUP BY user_id, DATE_TRUNC('month', ordered_at)
),
with_prev AS (
    SELECT *,
        LAG(monthly_total) OVER (PARTITION BY user_id ORDER BY month) AS prev_total
    FROM monthly_spending
),
decreases AS (
    SELECT user_id
    FROM with_prev
    WHERE prev_total IS NOT NULL AND monthly_total <= prev_total
)
SELECT u.name
FROM users u
WHERE u.id NOT IN (SELECT user_id FROM decreases)
  AND EXISTS (
      SELECT 1 FROM monthly_spending ms WHERE ms.user_id = u.id
      HAVING COUNT(*) >= 3  -- at least 3 months of data
  );
```

---

### Q35: Median salary per department

```sql
-- PostgreSQL: using PERCENTILE_CONT
SELECT 
    d.name AS department,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY e.salary) AS median_salary
FROM employees e
JOIN departments d ON e.department_id = d.id
GROUP BY d.id, d.name;

-- Generic (all databases): using ROW_NUMBER
WITH ranked AS (
    SELECT 
        department_id, salary,
        ROW_NUMBER() OVER (PARTITION BY department_id ORDER BY salary) AS rn,
        COUNT(*) OVER (PARTITION BY department_id) AS cnt
    FROM employees
)
SELECT 
    d.name AS department,
    AVG(r.salary) AS median_salary
FROM ranked r
JOIN departments d ON r.department_id = d.id
WHERE r.rn IN (FLOOR((r.cnt + 1) / 2.0), CEIL((r.cnt + 1) / 2.0))
GROUP BY d.id, d.name;
```

---

# HARD (36-55)

---

### Q36: Cohort Retention Analysis

```sql
-- What percentage of users who signed up in month X are still active in month X+1, X+2, etc.?
WITH user_cohort AS (
    SELECT 
        user_id,
        DATE_TRUNC('month', MIN(ordered_at)) AS cohort_month
    FROM orders
    GROUP BY user_id
),
user_activity AS (
    SELECT DISTINCT
        uc.cohort_month,
        DATE_TRUNC('month', o.ordered_at) AS activity_month,
        o.user_id
    FROM orders o
    JOIN user_cohort uc ON o.user_id = uc.user_id
),
cohort_sizes AS (
    SELECT cohort_month, COUNT(DISTINCT user_id) AS cohort_size
    FROM user_cohort GROUP BY cohort_month
)
SELECT 
    ua.cohort_month,
    EXTRACT(MONTH FROM AGE(ua.activity_month, ua.cohort_month))::INT AS month_number,
    COUNT(DISTINCT ua.user_id) AS active_users,
    cs.cohort_size,
    ROUND(COUNT(DISTINCT ua.user_id) * 100.0 / cs.cohort_size, 2) AS retention_pct
FROM user_activity ua
JOIN cohort_sizes cs ON ua.cohort_month = cs.cohort_month
GROUP BY ua.cohort_month, ua.activity_month, cs.cohort_size
ORDER BY ua.cohort_month, month_number;
```

---

### Q37: Sessionization — group user events into sessions (30-min gap)

```sql
WITH events_tagged AS (
    SELECT 
        user_id, event_time, event_type,
        CASE 
            WHEN event_time - LAG(event_time) OVER (PARTITION BY user_id ORDER BY event_time) 
                 > INTERVAL '30 minutes' 
            THEN 1 ELSE 0 
        END AS is_new_session
    FROM user_events
),
sessions AS (
    SELECT *,
        SUM(is_new_session) OVER (PARTITION BY user_id ORDER BY event_time) AS session_id
    FROM events_tagged
)
SELECT 
    user_id, session_id,
    MIN(event_time) AS session_start,
    MAX(event_time) AS session_end,
    MAX(event_time) - MIN(event_time) AS session_duration,
    COUNT(*) AS events_in_session,
    ARRAY_AGG(event_type ORDER BY event_time) AS event_sequence
FROM sessions
GROUP BY user_id, session_id
ORDER BY user_id, session_start;
```

---

### Q38: Find longest streak of increasing daily revenue

```sql
WITH daily_rev AS (
    SELECT 
        DATE(ordered_at) AS d,
        SUM(total) AS revenue
    FROM orders WHERE status = 'completed'
    GROUP BY DATE(ordered_at)
),
compared AS (
    SELECT 
        d, revenue,
        CASE WHEN revenue > LAG(revenue) OVER (ORDER BY d) THEN 0 ELSE 1 END AS reset
    FROM daily_rev
),
grouped AS (
    SELECT *, SUM(reset) OVER (ORDER BY d) AS grp
    FROM compared
)
SELECT 
    grp,
    MIN(d) AS streak_start,
    MAX(d) AS streak_end,
    COUNT(*) AS streak_length,
    MIN(revenue) AS min_revenue,
    MAX(revenue) AS max_revenue
FROM grouped
GROUP BY grp
ORDER BY streak_length DESC
LIMIT 5;
```

---

### Q39: Funnel Analysis (conversion rates between steps)

```sql
WITH funnel AS (
    SELECT
        COUNT(DISTINCT CASE WHEN event_type = 'page_view' THEN user_id END) AS viewed,
        COUNT(DISTINCT CASE WHEN event_type = 'add_to_cart' THEN user_id END) AS added_to_cart,
        COUNT(DISTINCT CASE WHEN event_type = 'checkout_start' THEN user_id END) AS started_checkout,
        COUNT(DISTINCT CASE WHEN event_type = 'purchase' THEN user_id END) AS purchased
    FROM user_events
    WHERE event_time >= CURRENT_DATE - INTERVAL '30 days'
)
SELECT 
    'Page View → Add to Cart' AS step,
    viewed AS users_at_step,
    added_to_cart AS users_at_next,
    ROUND(added_to_cart * 100.0 / NULLIF(viewed, 0), 2) AS conversion_pct
FROM funnel
UNION ALL
SELECT 
    'Add to Cart → Checkout',
    added_to_cart, started_checkout,
    ROUND(started_checkout * 100.0 / NULLIF(added_to_cart, 0), 2)
FROM funnel
UNION ALL
SELECT 
    'Checkout → Purchase',
    started_checkout, purchased,
    ROUND(purchased * 100.0 / NULLIF(started_checkout, 0), 2)
FROM funnel
UNION ALL
SELECT 
    'Overall: View → Purchase',
    viewed, purchased,
    ROUND(purchased * 100.0 / NULLIF(viewed, 0), 2)
FROM funnel;
```

---

### Q40: Find mutual friends (social network)

```sql
-- Table: friendships (user_id, friend_id) — bidirectional
-- Find mutual friends of user 1 and user 2
WITH friends_of_1 AS (
    SELECT friend_id FROM friendships WHERE user_id = 1
    UNION
    SELECT user_id FROM friendships WHERE friend_id = 1
),
friends_of_2 AS (
    SELECT friend_id FROM friendships WHERE user_id = 2
    UNION
    SELECT user_id FROM friendships WHERE friend_id = 2
)
SELECT u.name AS mutual_friend
FROM friends_of_1 f1
JOIN friends_of_2 f2 ON f1.friend_id = f2.friend_id
JOIN users u ON f1.friend_id = u.id
WHERE f1.friend_id NOT IN (1, 2);

-- Friend recommendations (friends of friends, not already friends)
WITH my_friends AS (
    SELECT friend_id FROM friendships WHERE user_id = 1
    UNION
    SELECT user_id FROM friendships WHERE friend_id = 1
),
friends_of_friends AS (
    SELECT 
        CASE WHEN f.user_id = mf.friend_id THEN f.friend_id ELSE f.user_id END AS fof_id
    FROM friendships f
    JOIN my_friends mf ON f.user_id = mf.friend_id OR f.friend_id = mf.friend_id
    WHERE f.user_id != 1 AND f.friend_id != 1
)
SELECT u.name, COUNT(*) AS mutual_connections
FROM friends_of_friends fof
JOIN users u ON fof.fof_id = u.id
WHERE fof.fof_id NOT IN (SELECT friend_id FROM my_friends)
  AND fof.fof_id != 1
GROUP BY u.id, u.name
ORDER BY mutual_connections DESC
LIMIT 10;
```

---

### Q41: Deadlock detection — find circular references

```sql
-- Find circular dependencies in a directed graph (e.g., task dependencies)
WITH RECURSIVE paths AS (
    SELECT from_task AS start, to_task AS current, ARRAY[from_task, to_task] AS path, FALSE AS has_cycle
    FROM task_dependencies
    
    UNION ALL
    
    SELECT p.start, td.to_task, p.path || td.to_task,
           td.to_task = ANY(p.path) AS has_cycle
    FROM paths p
    JOIN task_dependencies td ON p.current = td.from_task
    WHERE NOT p.has_cycle  -- stop when cycle detected
)
SELECT DISTINCT path
FROM paths
WHERE has_cycle;
```

---

### Q42: Inventory management — products that will run out in N days

```sql
WITH daily_sales AS (
    SELECT 
        oi.product_id,
        DATE(o.ordered_at) AS sale_date,
        SUM(oi.quantity) AS daily_qty
    FROM order_items oi
    JOIN orders o ON oi.order_id = o.id
    WHERE o.ordered_at >= CURRENT_DATE - INTERVAL '30 days'
    GROUP BY oi.product_id, DATE(o.ordered_at)
),
avg_daily_sales AS (
    SELECT 
        product_id,
        AVG(daily_qty) AS avg_daily_sales
    FROM daily_sales
    GROUP BY product_id
)
SELECT 
    p.name,
    p.stock AS current_stock,
    ROUND(ads.avg_daily_sales, 2) AS avg_daily_sales,
    CASE 
        WHEN ads.avg_daily_sales > 0 
        THEN ROUND(p.stock / ads.avg_daily_sales, 1)
        ELSE NULL 
    END AS days_until_stockout
FROM products p
JOIN avg_daily_sales ads ON p.id = ads.product_id
WHERE p.stock > 0
ORDER BY days_until_stockout ASC NULLS LAST
LIMIT 20;
```

---

### Q43: Recursive hierarchy with aggregation — total team salary under each manager

```sql
WITH RECURSIVE team_tree AS (
    SELECT id, name, manager_id, salary, id AS root_manager_id
    FROM employees
    WHERE manager_id IS NOT NULL
    
    UNION ALL
    
    SELECT e.id, e.name, e.manager_id, e.salary, tt.root_manager_id
    FROM employees e
    JOIN team_tree tt ON e.manager_id = tt.id
    WHERE e.id != tt.root_manager_id  -- prevent infinite loop
)
SELECT 
    m.name AS manager,
    COUNT(*) AS team_size,
    SUM(tt.salary) AS total_team_salary,
    ROUND(AVG(tt.salary), 2) AS avg_team_salary
FROM team_tree tt
JOIN employees m ON tt.root_manager_id = m.id
GROUP BY m.id, m.name
ORDER BY total_team_salary DESC;
```

---

### Q44: Temporal: Find overlapping bookings

```sql
-- Find all pairs of overlapping room bookings
SELECT 
    a.id AS booking_1, a.guest_name AS guest_1, a.start_date AS start_1, a.end_date AS end_1,
    b.id AS booking_2, b.guest_name AS guest_2, b.start_date AS start_2, b.end_date AS end_2
FROM room_bookings a
JOIN room_bookings b ON a.room_id = b.room_id 
    AND a.id < b.id  -- avoid duplicate pairs and self-join
    AND a.start_date < b.end_date 
    AND a.end_date > b.start_date;  -- overlap condition

-- General overlap formula:
-- Two ranges [A_start, A_end] and [B_start, B_end] overlap when:
-- A_start < B_end AND A_end > B_start
```

---

### Q45: Implement a bank transfer with proper transaction handling

```sql
-- Transfer $500 from account 1 to account 2
BEGIN;
    -- Lock rows in consistent order to prevent deadlocks
    SELECT * FROM accounts WHERE id IN (1, 2) ORDER BY id FOR UPDATE;
    
    -- Verify sufficient balance
    DO $$
    DECLARE
        v_balance DECIMAL(10,2);
    BEGIN
        SELECT balance INTO v_balance FROM accounts WHERE id = 1;
        IF v_balance < 500 THEN
            RAISE EXCEPTION 'Insufficient funds: balance=%, required=500', v_balance;
        END IF;
    END $$;
    
    -- Perform transfer
    UPDATE accounts SET balance = balance - 500 WHERE id = 1;
    UPDATE accounts SET balance = balance + 500 WHERE id = 2;
    
    -- Log transaction
    INSERT INTO transactions (from_account, to_account, amount, type, created_at)
    VALUES (1, 2, 500, 'transfer', CURRENT_TIMESTAMP);
    
COMMIT;
```

---

### Q46: Moving window: identify anomalous days (revenue > 2 std dev from 30-day average)

```sql
WITH daily_stats AS (
    SELECT 
        DATE(ordered_at) AS d,
        SUM(total) AS revenue
    FROM orders
    WHERE status = 'completed'
    GROUP BY DATE(ordered_at)
),
with_windows AS (
    SELECT 
        d, revenue,
        AVG(revenue) OVER (ORDER BY d ROWS BETWEEN 30 PRECEDING AND 1 PRECEDING) AS avg_30d,
        STDDEV(revenue) OVER (ORDER BY d ROWS BETWEEN 30 PRECEDING AND 1 PRECEDING) AS std_30d
    FROM daily_stats
)
SELECT 
    d, revenue, 
    ROUND(avg_30d, 2) AS avg_30d,
    ROUND(std_30d, 2) AS std_30d,
    ROUND((revenue - avg_30d) / NULLIF(std_30d, 0), 2) AS z_score,
    CASE 
        WHEN revenue > avg_30d + 2 * std_30d THEN 'SPIKE'
        WHEN revenue < avg_30d - 2 * std_30d THEN 'DROP'
        ELSE 'NORMAL'
    END AS anomaly
FROM with_windows
WHERE ABS(revenue - avg_30d) > 2 * NULLIF(std_30d, 0)
ORDER BY d;
```

---

### Q47: Dynamic ranking: calculate percentile rank

```sql
SELECT 
    name, department_id, salary,
    PERCENT_RANK() OVER (ORDER BY salary) AS pct_rank_overall,
    PERCENT_RANK() OVER (PARTITION BY department_id ORDER BY salary) AS pct_rank_dept,
    CUME_DIST() OVER (ORDER BY salary) AS cume_dist,
    NTILE(10) OVER (ORDER BY salary) AS decile
FROM employees
ORDER BY salary;

-- PERCENT_RANK: (rank - 1) / (total rows - 1), ranges 0 to 1
-- CUME_DIST: rank / total rows, ranges > 0 to 1
-- NTILE(10): divide into 10 groups (deciles)
```

---

### Q48: RFM Analysis (Recency, Frequency, Monetary)

```sql
WITH rfm_raw AS (
    SELECT 
        user_id,
        CURRENT_DATE - MAX(DATE(ordered_at)) AS recency_days,
        COUNT(*) AS frequency,
        SUM(total) AS monetary
    FROM orders
    WHERE status = 'completed'
    GROUP BY user_id
),
rfm_scored AS (
    SELECT *,
        NTILE(5) OVER (ORDER BY recency_days DESC) AS r_score,    -- lower recency = better
        NTILE(5) OVER (ORDER BY frequency ASC) AS f_score,        -- higher frequency = better
        NTILE(5) OVER (ORDER BY monetary ASC) AS m_score          -- higher monetary = better
    FROM rfm_raw
)
SELECT 
    u.name, u.email,
    r.recency_days, r.frequency, r.monetary,
    r.r_score, r.f_score, r.m_score,
    r.r_score + r.f_score + r.m_score AS rfm_total,
    CASE 
        WHEN r.r_score >= 4 AND r.f_score >= 4 AND r.m_score >= 4 THEN 'Champions'
        WHEN r.r_score >= 4 AND r.f_score >= 3 THEN 'Loyal'
        WHEN r.r_score >= 3 AND r.f_score <= 2 AND r.m_score >= 3 THEN 'Big Spender at Risk'
        WHEN r.r_score <= 2 AND r.f_score >= 3 THEN 'Needs Attention'
        WHEN r.r_score <= 2 AND r.f_score <= 2 THEN 'Lost'
        ELSE 'Other'
    END AS segment
FROM rfm_scored r
JOIN users u ON r.user_id = u.id
ORDER BY rfm_total DESC;
```

---

### Q49: Generate a complete report even for dates with no data

```sql
-- Show daily revenue, including days with 0 revenue
WITH date_series AS (
    SELECT generate_series(
        (SELECT MIN(DATE(ordered_at)) FROM orders),
        CURRENT_DATE,
        '1 day'::INTERVAL
    )::DATE AS d
),
daily_revenue AS (
    SELECT DATE(ordered_at) AS d, SUM(total) AS revenue, COUNT(*) AS orders
    FROM orders WHERE status = 'completed'
    GROUP BY DATE(ordered_at)
)
SELECT 
    ds.d AS date,
    COALESCE(dr.revenue, 0) AS revenue,
    COALESCE(dr.orders, 0) AS order_count,
    TO_CHAR(ds.d, 'Day') AS day_name,
    CASE WHEN EXTRACT(DOW FROM ds.d) IN (0, 6) THEN 'Weekend' ELSE 'Weekday' END AS day_type
FROM date_series ds
LEFT JOIN daily_revenue dr ON ds.d = dr.d
ORDER BY ds.d;
```

---

### Q50: Implement SCD Type 2 (Slowly Changing Dimension)

```sql
-- Track changes to employee salary over time
CREATE TABLE employee_salary_history (
    id              SERIAL PRIMARY KEY,
    employee_id     INTEGER NOT NULL,
    salary          DECIMAL(10,2) NOT NULL,
    effective_from  DATE NOT NULL,
    effective_to    DATE DEFAULT '9999-12-31',
    is_current      BOOLEAN DEFAULT TRUE
);

-- When salary changes: close old record, insert new
-- Close old record
UPDATE employee_salary_history
SET effective_to = CURRENT_DATE - 1, is_current = FALSE
WHERE employee_id = 1 AND is_current = TRUE;

-- Insert new record
INSERT INTO employee_salary_history (employee_id, salary, effective_from)
VALUES (1, 160000, CURRENT_DATE);

-- Query: "What was employee 1's salary on 2023-06-15?"
SELECT salary
FROM employee_salary_history
WHERE employee_id = 1 
  AND '2023-06-15' BETWEEN effective_from AND effective_to;

-- Query: All salary changes for employee
SELECT salary, effective_from, effective_to,
    LAG(salary) OVER (ORDER BY effective_from) AS prev_salary,
    salary - LAG(salary) OVER (ORDER BY effective_from) AS change
FROM employee_salary_history
WHERE employee_id = 1
ORDER BY effective_from;
```

---

### Q51: Island detection — find consecutive date ranges

```sql
-- Given a table of dates when a user was active, find continuous date ranges
WITH user_activity AS (
    SELECT DISTINCT user_id, DATE(activity_time) AS active_date
    FROM user_events
    WHERE user_id = 42
),
with_row_num AS (
    SELECT 
        active_date,
        active_date - (ROW_NUMBER() OVER (ORDER BY active_date))::INT AS island_group
    FROM user_activity
)
SELECT 
    MIN(active_date) AS range_start,
    MAX(active_date) AS range_end,
    MAX(active_date) - MIN(active_date) + 1 AS range_days
FROM with_row_num
GROUP BY island_group
ORDER BY range_start;

-- Example result:
-- range_start | range_end  | range_days
-- 2024-01-05  | 2024-01-09 | 5          (5-day streak)
-- 2024-01-15  | 2024-01-15 | 1          (single day)
-- 2024-01-20  | 2024-01-28 | 9          (9-day streak)
```

---

### Q52: Implement a leaderboard with ranking and percentile

```sql
WITH player_scores AS (
    SELECT 
        user_id, 
        SUM(score) AS total_score,
        COUNT(*) AS games_played,
        MAX(score) AS best_score,
        AVG(score) AS avg_score
    FROM game_results
    WHERE played_at >= CURRENT_DATE - INTERVAL '30 days'
    GROUP BY user_id
)
SELECT 
    u.username,
    ps.total_score,
    ps.games_played,
    ps.best_score,
    ROUND(ps.avg_score, 2) AS avg_score,
    RANK() OVER (ORDER BY ps.total_score DESC) AS rank,
    DENSE_RANK() OVER (ORDER BY ps.total_score DESC) AS dense_rank,
    ROUND(PERCENT_RANK() OVER (ORDER BY ps.total_score) * 100, 2) AS percentile,
    CASE 
        WHEN PERCENT_RANK() OVER (ORDER BY ps.total_score DESC) <= 0.01 THEN 'Top 1%'
        WHEN PERCENT_RANK() OVER (ORDER BY ps.total_score DESC) <= 0.05 THEN 'Top 5%'
        WHEN PERCENT_RANK() OVER (ORDER BY ps.total_score DESC) <= 0.10 THEN 'Top 10%'
        WHEN PERCENT_RANK() OVER (ORDER BY ps.total_score DESC) <= 0.25 THEN 'Top 25%'
        ELSE 'Below Top 25%'
    END AS tier
FROM player_scores ps
JOIN users u ON ps.user_id = u.id
ORDER BY rank
LIMIT 100;
```

---

### Q53: Weighted moving average (with exponential Smoothing)

```sql
-- Weighted moving average: more recent days have higher weight
WITH daily_sales AS (
    SELECT DATE(ordered_at) AS d, SUM(total) AS revenue
    FROM orders WHERE status = 'completed'
    GROUP BY DATE(ordered_at)
),
with_weights AS (
    SELECT 
        d, revenue,
        ROW_NUMBER() OVER (ORDER BY d DESC) AS days_ago
    FROM daily_sales
)
SELECT 
    d, revenue,
    -- Simple weighted average (linear weights) over last 7 days
    SUM(revenue * (8 - LEAST(days_ago, 7))) OVER (
        ORDER BY d ROWS BETWEEN 6 PRECEDING AND CURRENT ROW
    ) / NULLIF(SUM(8 - LEAST(days_ago, 7)) OVER (
        ORDER BY d ROWS BETWEEN 6 PRECEDING AND CURRENT ROW
    ), 0) AS weighted_avg_7d
FROM with_weights
ORDER BY d DESC
LIMIT 30;
```

---

### Q54: Multi-currency total with exchange rates

```sql
WITH latest_rates AS (
    -- Get latest exchange rate for each currency
    SELECT DISTINCT ON (currency_code) 
        currency_code, rate_to_usd, rate_date
    FROM exchange_rates
    ORDER BY currency_code, rate_date DESC
)
SELECT 
    u.name,
    SUM(o.total) AS total_original,
    o.currency AS original_currency,
    SUM(o.total * lr.rate_to_usd) AS total_usd,
    STRING_AGG(DISTINCT o.currency, ', ') AS currencies_used
FROM orders o
JOIN users u ON o.user_id = u.id
JOIN latest_rates lr ON o.currency = lr.currency_code
GROUP BY u.id, u.name, o.currency
ORDER BY total_usd DESC;
```

---

### Q55: Full-text search with ranking and highlighting

```sql
-- Search products with ranking and snippet highlighting
SELECT 
    p.id, p.name, p.description, p.price,
    ts_rank(
        to_tsvector('english', p.name || ' ' || COALESCE(p.description, '')),
        websearch_to_tsquery('english', 'wireless bluetooth headphones')
    ) AS relevance,
    ts_headline(
        'english',
        p.name || ' — ' || COALESCE(p.description, ''),
        websearch_to_tsquery('english', 'wireless bluetooth headphones'),
        'StartSel=<b>, StopSel=</b>, MaxFragments=2, FragmentDelimiter=...'
    ) AS highlighted_snippet
FROM products p
WHERE to_tsvector('english', p.name || ' ' || COALESCE(p.description, '')) 
    @@ websearch_to_tsquery('english', 'wireless bluetooth headphones')
ORDER BY relevance DESC
LIMIT 20;
```

---

## Quick Reference: Which SQL Feature for Which Problem

| Problem Type | Key SQL Feature |
|-------------|----------------|
| Find Nth highest/lowest | `DENSE_RANK()` / `OFFSET` |
| Top N per group | `ROW_NUMBER() OVER (PARTITION BY ...)` |
| Running total | `SUM() OVER (ORDER BY ... ROWS UNBOUNDED PRECEDING)` |
| Month-over-month | `LAG()` window function |
| Hierarchies | Recursive CTE |
| Gaps and islands | `ROW_NUMBER()` difference trick |
| Duplicates | `GROUP BY + HAVING COUNT(*) > 1` |
| Exists/Not Exists | Anti-join / Semi-join patterns |
| Pivoting | `CASE WHEN` inside aggregates |
| Time-series gaps | `generate_series` LEFT JOIN |
| Sessionization | `LAG()` + `SUM()` window combo |
| Percentiles | `NTILE()` / `PERCENTILE_CONT()` |
| Deduplication | `ROW_NUMBER()` + delete where rn > 1 |
| Overlapping ranges | Self-join with `A.start < B.end AND A.end > B.start` |
| Consecutive streaks | Gap-and-island pattern |
