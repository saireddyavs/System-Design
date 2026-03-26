# Aggregations & Window Functions — Deep Dive

## GROUP BY, HAVING, and the Complete Window Functions Toolkit

---

## 1. Aggregate Functions — Complete Reference

### Basic Aggregate Functions

```sql
-- COUNT
SELECT COUNT(*)          FROM employees;  -- count ALL rows (including NULLs)
SELECT COUNT(manager_id) FROM employees;  -- count non-NULL values only
SELECT COUNT(DISTINCT department_id) FROM employees;  -- count unique values

-- SUM / AVG
SELECT SUM(salary)       FROM employees;  -- total salary
SELECT AVG(salary)       FROM employees;  -- average salary
SELECT SUM(DISTINCT salary) FROM employees;  -- sum of unique salaries

-- MIN / MAX
SELECT MIN(salary)       FROM employees;  -- lowest salary
SELECT MAX(salary)       FROM employees;  -- highest salary
SELECT MIN(hire_date)    FROM employees;  -- earliest hire date

-- STRING_AGG (PostgreSQL) / GROUP_CONCAT (MySQL)
SELECT STRING_AGG(name, ', ' ORDER BY name) FROM employees;  
-- 'Alice, Bob, Carol, Dave, Eve, Frank, Grace, Hank'

-- ARRAY_AGG (PostgreSQL)
SELECT ARRAY_AGG(name ORDER BY salary DESC) FROM employees;
-- '{Alice,Dave,Bob,Eve,Frank,Carol,Hank,Grace}'

-- BOOL_AND / BOOL_OR (PostgreSQL)
SELECT BOOL_AND(is_active) FROM employees;  -- true only if ALL are active
SELECT BOOL_OR(is_active)  FROM employees;  -- true if ANY is active

-- Statistical aggregates
SELECT 
    STDDEV(salary)      AS std_deviation,
    VARIANCE(salary)    AS variance,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY salary) AS median
FROM employees;
```

---

## 2. GROUP BY

### Basics

```sql
-- Count employees per department
SELECT department_id, COUNT(*) AS emp_count
FROM employees
GROUP BY department_id;

-- Note: every column in SELECT must either be:
--   1. In GROUP BY, OR
--   2. Inside an aggregate function
-- Otherwise → error!

-- ❌ WRONG:
SELECT department_id, name, COUNT(*)  -- name is not grouped or aggregated!
FROM employees GROUP BY department_id;

-- ✅ CORRECT:
SELECT department_id, COUNT(*) AS emp_count, AVG(salary) AS avg_salary
FROM employees GROUP BY department_id;
```

### GROUP BY with Multiple Columns

```sql
-- Count per department AND status
SELECT department_id, status, COUNT(*) AS cnt
FROM employees
GROUP BY department_id, status
ORDER BY department_id, status;

-- Result:
-- department_id | status     | cnt
-- 1             | active     | 2
-- 1             | inactive   | 1
-- 2             | active     | 2
-- 3             | active     | 1
-- 3             | inactive   | 1
```

### GROUP BY with Expressions

```sql
-- Group by year of hire
SELECT EXTRACT(YEAR FROM hire_date) AS hire_year, COUNT(*) AS hired
FROM employees
GROUP BY EXTRACT(YEAR FROM hire_date)
ORDER BY hire_year;

-- Group by salary band
SELECT 
    CASE 
        WHEN salary >= 120000 THEN 'Senior'
        WHEN salary >= 90000 THEN 'Mid'
        ELSE 'Junior'
    END AS salary_band,
    COUNT(*) AS emp_count,
    AVG(salary) AS avg_salary
FROM employees
GROUP BY 
    CASE 
        WHEN salary >= 120000 THEN 'Senior'
        WHEN salary >= 90000 THEN 'Mid'
        ELSE 'Junior'
    END;

-- Group by month (common in reports)
SELECT 
    DATE_TRUNC('month', ordered_at) AS month,
    COUNT(*) AS order_count,
    SUM(total) AS revenue
FROM orders
GROUP BY DATE_TRUNC('month', ordered_at)
ORDER BY month;
```

---

## 3. HAVING

Filters **groups** (aggregated results). WHERE filters **rows** before grouping.

```sql
-- WHERE vs HAVING
-- WHERE:  filters individual rows BEFORE grouping
-- HAVING: filters groups AFTER grouping

-- Departments with more than 2 employees
SELECT department_id, COUNT(*) AS emp_count
FROM employees
GROUP BY department_id
HAVING COUNT(*) > 2;

-- ❌ WRONG: Can't use aggregate in WHERE
SELECT department_id, COUNT(*) FROM employees
WHERE COUNT(*) > 2  -- ERROR!
GROUP BY department_id;

-- Combined WHERE + HAVING
-- "Active departments with more than 2 active employees earning > 80K"
SELECT department_id, COUNT(*) AS count, AVG(salary) AS avg_sal
FROM employees
WHERE status = 'active' AND salary > 80000    -- filter rows first
GROUP BY department_id
HAVING COUNT(*) > 2;                           -- then filter groups

-- HAVING with subquery
SELECT department_id, AVG(salary) AS dept_avg
FROM employees
GROUP BY department_id
HAVING AVG(salary) > (SELECT AVG(salary) FROM employees);
-- Departments with above-average salary
```

---

## 4. GROUPING SETS, ROLLUP, CUBE

Advanced grouping for **multi-dimensional aggregation** (reporting/analytics).

### GROUPING SETS

```sql
-- Standard GROUPING SETS — specify exactly which groups you want
SELECT department_id, status, COUNT(*) AS cnt, SUM(salary) AS total_sal
FROM employees
GROUP BY GROUPING SETS (
    (department_id, status),   -- group by both
    (department_id),           -- group by department only
    (status),                  -- group by status only
    ()                         -- grand total (no grouping)
);

-- This is equivalent to 4 separate GROUP BY queries UNIONed together,
-- but MUCH more efficient (single table scan)!
```

### ROLLUP

```sql
-- ROLLUP: hierarchical subtotals (right to left)
SELECT department_id, status, COUNT(*) AS cnt, SUM(salary) AS total_sal
FROM employees
GROUP BY ROLLUP(department_id, status);

-- ROLLUP(A, B) generates:
-- (A, B)    → detail
-- (A)       → subtotal per A
-- ()        → grand total
-- But NOT (B) alone

-- Example output:
-- dept | status   | cnt | total_sal
-- 1    | active   | 2   | 215000     ← detail
-- 1    | inactive | 1   | 95000      ← detail
-- 1    | NULL     | 3   | 310000     ← subtotal for dept 1
-- 2    | active   | 2   | 240000
-- 2    | NULL     | 2   | 240000     ← subtotal for dept 2
-- NULL | NULL     | 8   | 880000     ← GRAND TOTAL

-- Use GROUPING() to distinguish real NULLs from rollup NULLs
SELECT 
    CASE WHEN GROUPING(department_id) = 1 THEN 'ALL' ELSE department_id::TEXT END AS dept,
    CASE WHEN GROUPING(status) = 1 THEN 'ALL' ELSE status END AS status,
    COUNT(*), SUM(salary)
FROM employees
GROUP BY ROLLUP(department_id, status);
```

### CUBE

```sql
-- CUBE: all possible combinations
SELECT department_id, status, COUNT(*) AS cnt, SUM(salary) AS total_sal
FROM employees
GROUP BY CUBE(department_id, status);

-- CUBE(A, B) generates:
-- (A, B)    → detail
-- (A)       → subtotal per A
-- (B)       → subtotal per B  ← ROLLUP doesn't include this
-- ()        → grand total
```

---

## 5. Window Functions — The Complete Guide

### What Are Window Functions?

Window functions perform calculations **across a set of rows related to the current row**, without collapsing rows like GROUP BY does.

```
GROUP BY:      Collapses rows → 1 row per group
Window Funcs:  Keeps all rows → adds computed column

Example:
┌──────┬────────┬────────┬─────────────┬──────────┐
│ name │ dept   │ salary │ dept_avg    │ rank     │
├──────┼────────┼────────┼─────────────┼──────────┤
│ Alice│ Eng    │ 150000 │ 121666.67   │ 1        │
│ Bob  │ Eng    │ 120000 │ 121666.67   │ 2        │
│ Carol│ Eng    │ 95000  │ 121666.67   │ 3        │
│ Dave │ Mktg   │ 130000 │ 120000.00   │ 1        │
│ Eve  │ Mktg   │ 110000 │ 120000.00   │ 2        │
└──────┴────────┴────────┴─────────────┴──────────┘
  dept_avg = AVG(salary) OVER (PARTITION BY dept)
  rank = RANK() OVER (PARTITION BY dept ORDER BY salary DESC)
```

### Window Function Syntax

```sql
function_name(args) OVER (
    [PARTITION BY column(s)]     -- define the window/group
    [ORDER BY column(s)]         -- define sort within window
    [frame_clause]               -- define rows within partition
)
```

---

## 6. Ranking Functions

### ROW_NUMBER, RANK, DENSE_RANK

```sql
SELECT 
    name, department_id, salary,
    ROW_NUMBER() OVER (ORDER BY salary DESC) AS row_num,
    RANK()       OVER (ORDER BY salary DESC) AS rank,
    DENSE_RANK() OVER (ORDER BY salary DESC) AS dense_rank
FROM employees;

-- Difference illustrated:
-- name  | salary | ROW_NUMBER | RANK | DENSE_RANK
-- Alice | 150000 | 1          | 1    | 1
-- Dave  | 130000 | 2          | 2    | 2
-- Bob   | 120000 | 3          | 3    | 3
-- Eve   | 110000 | 4          | 4    | 4       ← so far, all same
-- Frank | 105000 | 5          | 5    | 5
-- Carol | 95000  | 6          | 6    | 6
-- Hank  | 90000  | 7          | 7    | 7
-- Grace | 85000  | 8          | 8    | 8

-- But with TIES (if Carol and Hank both had 95000):
-- name  | salary | ROW_NUMBER | RANK | DENSE_RANK
-- ...
-- Carol | 95000  | 6          | 6    | 6
-- Hank  | 95000  | 7          | 6    | 6     ← RANK ties, DENSE_RANK ties
-- Grace | 85000  | 8          | 8    | 7     ← RANK skips 7, DENSE_RANK doesn't
```

**Key Differences:**
| Function | Ties | Gaps |
|----------|------|------|
| `ROW_NUMBER` | Breaks ties arbitrarily (unique) | No gaps |
| `RANK` | Same rank for ties | Gaps after ties (1,2,2,4) |
| `DENSE_RANK` | Same rank for ties | No gaps (1,2,2,3) |

### NTILE

Divides rows into N equal(ish) groups.

```sql
-- Divide employees into 4 salary quartiles
SELECT 
    name, salary,
    NTILE(4) OVER (ORDER BY salary DESC) AS quartile
FROM employees;
-- Q1 = top 25%, Q2 = 25-50%, Q3 = 50-75%, Q4 = bottom 25%
```

### Ranking Per Group — Top N Per Category

```sql
-- Top 2 highest-paid per department
SELECT * FROM (
    SELECT 
        name, department_id, salary,
        ROW_NUMBER() OVER (PARTITION BY department_id ORDER BY salary DESC) AS rn
    FROM employees
) ranked
WHERE rn <= 2;

-- Using DENSE_RANK if you want ties included:
SELECT * FROM (
    SELECT 
        name, department_id, salary,
        DENSE_RANK() OVER (PARTITION BY department_id ORDER BY salary DESC) AS dr
    FROM employees
) ranked
WHERE dr <= 2;
```

---

## 7. Value Functions — LAG, LEAD, FIRST_VALUE, LAST_VALUE, NTH_VALUE

### LAG and LEAD

```sql
-- LAG: access the PREVIOUS row's value
-- LEAD: access the NEXT row's value

SELECT 
    name, department_id, salary,
    LAG(salary)  OVER (ORDER BY salary DESC) AS prev_higher_salary,
    LEAD(salary) OVER (ORDER BY salary DESC) AS next_lower_salary,
    salary - LAG(salary) OVER (ORDER BY salary DESC) AS diff_from_prev
FROM employees;

-- Result:
-- name  | salary | prev_higher | next_lower | diff
-- Alice | 150000 | NULL        | 130000     | NULL
-- Dave  | 130000 | 150000      | 120000     | -20000
-- Bob   | 120000 | 130000      | 110000     | -10000
-- Eve   | 110000 | 120000      | 105000     | -10000
-- ...

-- LAG/LEAD with offset and default
LAG(salary, 2, 0) OVER (...)  -- 2 rows back, default 0 if no row

-- Practical: Month-over-month revenue change
SELECT 
    month,
    revenue,
    LAG(revenue) OVER (ORDER BY month) AS prev_month,
    revenue - LAG(revenue) OVER (ORDER BY month) AS mom_change,
    ROUND(
        (revenue - LAG(revenue) OVER (ORDER BY month)) * 100.0 
        / LAG(revenue) OVER (ORDER BY month), 2
    ) AS mom_pct_change
FROM monthly_revenue;
```

### FIRST_VALUE, LAST_VALUE, NTH_VALUE

```sql
SELECT 
    name, department_id, salary,
    FIRST_VALUE(name) OVER (
        PARTITION BY department_id ORDER BY salary DESC
    ) AS highest_paid_in_dept,
    LAST_VALUE(name) OVER (
        PARTITION BY department_id ORDER BY salary DESC
        ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING
    ) AS lowest_paid_in_dept,
    NTH_VALUE(name, 2) OVER (
        PARTITION BY department_id ORDER BY salary DESC
        ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING
    ) AS second_highest_in_dept
FROM employees;

-- ⚠️ IMPORTANT: LAST_VALUE and NTH_VALUE need explicit frame!
-- Default frame is RANGE BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
-- This means LAST_VALUE returns current row by default (confusing!)
-- Always use: ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING
```

---

## 8. Aggregate Window Functions

Any aggregate function can be used as a window function with OVER.

```sql
SELECT 
    name, department_id, salary,
    -- Aggregates over window (no GROUP BY needed, all rows preserved)
    SUM(salary)   OVER (PARTITION BY department_id) AS dept_total,
    AVG(salary)   OVER (PARTITION BY department_id) AS dept_avg,
    COUNT(*)      OVER (PARTITION BY department_id) AS dept_count,
    MIN(salary)   OVER (PARTITION BY department_id) AS dept_min,
    MAX(salary)   OVER (PARTITION BY department_id) AS dept_max,
    -- Compare individual to group
    salary - AVG(salary) OVER (PARTITION BY department_id) AS vs_dept_avg,
    ROUND(salary * 100.0 / SUM(salary) OVER (PARTITION BY department_id), 2) AS pct_of_dept
FROM employees;

-- Result:
-- name  | dept | salary | dept_total | dept_avg   | vs_dept_avg | pct_of_dept
-- Alice | 1    | 150000 | 365000     | 121666.67  | 28333.33    | 41.10
-- Bob   | 1    | 120000 | 365000     | 121666.67  | -1666.67    | 32.88
-- Carol | 1    | 95000  | 365000     | 121666.67  | -26666.67   | 26.03
```

---

## 9. Running Totals & Moving Averages (Frame Clause)

### Frame Clause Syntax

```
ROWS | RANGE BETWEEN frame_start AND frame_end

frame_start / frame_end options:
  UNBOUNDED PRECEDING   — from the partition start
  N PRECEDING           — N rows before current
  CURRENT ROW           — current row
  N FOLLOWING           — N rows after current
  UNBOUNDED FOLLOWING   — to the partition end
```

### Running Total

```sql
-- Running total of salary within department
SELECT 
    name, department_id, salary,
    SUM(salary) OVER (
        PARTITION BY department_id 
        ORDER BY hire_date
        ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
    ) AS running_total
FROM employees;

-- Cumulative percentage
SELECT 
    name, salary,
    SUM(salary) OVER (ORDER BY salary DESC ROWS UNBOUNDED PRECEDING) AS cumulative,
    ROUND(
        SUM(salary) OVER (ORDER BY salary DESC ROWS UNBOUNDED PRECEDING) * 100.0 
        / SUM(salary) OVER (), 2
    ) AS cumulative_pct
FROM employees;
```

### Moving Average

```sql
-- 3-day moving average of sales
SELECT 
    sale_date, 
    amount,
    AVG(amount) OVER (
        ORDER BY sale_date 
        ROWS BETWEEN 2 PRECEDING AND CURRENT ROW
    ) AS moving_avg_3day,
    AVG(amount) OVER (
        ORDER BY sale_date 
        ROWS BETWEEN 6 PRECEDING AND CURRENT ROW
    ) AS moving_avg_7day
FROM daily_sales;

-- Moving sum (last 30 days)
SELECT
    order_date,
    daily_revenue,
    SUM(daily_revenue) OVER (
        ORDER BY order_date
        RANGE BETWEEN INTERVAL '30 days' PRECEDING AND CURRENT ROW
    ) AS rolling_30day_revenue
FROM daily_revenue;
```

### ROWS vs RANGE

```sql
-- ROWS: counts physical rows (exact number of rows)
-- RANGE: counts logical range of values (based on ORDER BY value)

-- With ROWS BETWEEN 2 PRECEDING AND CURRENT ROW:
--   Always includes exactly 3 rows (or fewer at start)

-- With RANGE BETWEEN 2 PRECEDING AND CURRENT ROW:
--   Includes all rows where ORDER BY value is within 2 of current
--   If ORDER BY is salary: rows where salary is within 2 of current salary
--   Can include more or fewer than 3 rows!

-- Example: If sorted by date, RANGE BETWEEN INTERVAL '7 days' PRECEDING AND CURRENT ROW
-- includes ALL rows within last 7 days (could be 5 rows or 100 rows)
```

---

## 10. Window Function Interview Questions

### Q1: Running rank of sales reps by cumulative revenue

```sql
SELECT 
    rep_name,
    sale_date,
    amount,
    SUM(amount) OVER (PARTITION BY rep_name ORDER BY sale_date) AS cumulative_revenue,
    RANK() OVER (ORDER BY SUM(amount) OVER (PARTITION BY rep_name) DESC) AS overall_rank
FROM sales;
```

### Q2: Year-over-year growth rate

```sql
WITH yearly AS (
    SELECT 
        EXTRACT(YEAR FROM ordered_at) AS year,
        SUM(total) AS revenue
    FROM orders
    GROUP BY EXTRACT(YEAR FROM ordered_at)
)
SELECT 
    year, 
    revenue,
    LAG(revenue) OVER (ORDER BY year) AS prev_year,
    ROUND(
        (revenue - LAG(revenue) OVER (ORDER BY year)) * 100.0 
        / LAG(revenue) OVER (ORDER BY year), 2
    ) AS yoy_growth_pct
FROM yearly;
```

### Q3: Find consecutive day streaks (gaps and islands)

```sql
-- Find users with login streaks
WITH user_logins AS (
    SELECT DISTINCT user_id, login_date::DATE AS login_date
    FROM logins
),
groups AS (
    SELECT 
        user_id, login_date,
        login_date - ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY login_date)::INT * INTERVAL '1 day' AS grp
    FROM user_logins
)
SELECT 
    user_id, 
    MIN(login_date) AS streak_start,
    MAX(login_date) AS streak_end,
    COUNT(*) AS streak_length
FROM groups
GROUP BY user_id, grp
HAVING COUNT(*) >= 3  -- streaks of 3+ days
ORDER BY streak_length DESC;
```

### Q4: Percentage of total per category

```sql
SELECT 
    category,
    product_name,
    revenue,
    ROUND(revenue * 100.0 / SUM(revenue) OVER (PARTITION BY category), 2) AS pct_of_category,
    ROUND(revenue * 100.0 / SUM(revenue) OVER (), 2) AS pct_of_total
FROM product_sales;
```

### Q5: Median salary per department

```sql
-- Using PERCENTILE_CONT
SELECT DISTINCT
    department_id,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY salary) 
        OVER (PARTITION BY department_id) AS median_salary
FROM employees;

-- Using ROW_NUMBER (for DBs without PERCENTILE_CONT)
WITH ranked AS (
    SELECT 
        department_id, salary,
        ROW_NUMBER() OVER (PARTITION BY department_id ORDER BY salary) AS rn,
        COUNT(*) OVER (PARTITION BY department_id) AS cnt
    FROM employees
)
SELECT department_id, AVG(salary) AS median_salary
FROM ranked
WHERE rn IN (CEIL(cnt/2.0), CEIL(cnt/2.0) + (1 - cnt%2))
GROUP BY department_id;
```

### Q6: Sessionization — group events into sessions

```sql
-- Events within 30 minutes are part of the same session
WITH events_with_gap AS (
    SELECT 
        user_id, event_time,
        CASE 
            WHEN event_time - LAG(event_time) OVER (PARTITION BY user_id ORDER BY event_time) 
                 > INTERVAL '30 minutes' 
            THEN 1 
            ELSE 0 
        END AS new_session
    FROM user_events
),
sessions AS (
    SELECT 
        user_id, event_time,
        SUM(new_session) OVER (PARTITION BY user_id ORDER BY event_time) AS session_id
    FROM events_with_gap
)
SELECT 
    user_id, session_id,
    MIN(event_time) AS session_start,
    MAX(event_time) AS session_end,
    COUNT(*) AS event_count,
    MAX(event_time) - MIN(event_time) AS session_duration
FROM sessions
GROUP BY user_id, session_id;
```

### Q7: Detect salary anomalies (> 2 std deviations from department mean)

```sql
SELECT name, department_id, salary, dept_avg, dept_stddev
FROM (
    SELECT 
        name, department_id, salary,
        AVG(salary) OVER (PARTITION BY department_id) AS dept_avg,
        STDDEV(salary) OVER (PARTITION BY department_id) AS dept_stddev
    FROM employees
) stats
WHERE ABS(salary - dept_avg) > 2 * dept_stddev;
```

### Q8: Find the first and last purchase per customer

```sql
SELECT DISTINCT
    customer_id,
    FIRST_VALUE(order_date) OVER w AS first_purchase,
    LAST_VALUE(order_date) OVER w AS last_purchase,
    FIRST_VALUE(total) OVER w AS first_order_amount,
    LAST_VALUE(total) OVER w AS last_order_amount,
    COUNT(*) OVER (PARTITION BY customer_id) AS total_orders
FROM orders
WINDOW w AS (
    PARTITION BY customer_id 
    ORDER BY order_date
    ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING
);
-- Note: WINDOW clause (named window) for reuse — cleaner than repeating OVER
```

### Q9: Cohort retention analysis

```sql
-- First, find each user's cohort (month of first purchase)
WITH user_cohort AS (
    SELECT 
        user_id,
        DATE_TRUNC('month', MIN(ordered_at)) AS cohort_month
    FROM orders
    GROUP BY user_id
),
-- Then, count active users per cohort per month
user_activity AS (
    SELECT
        uc.cohort_month,
        DATE_TRUNC('month', o.ordered_at) AS activity_month,
        COUNT(DISTINCT o.user_id) AS active_users
    FROM orders o
    JOIN user_cohort uc ON o.user_id = uc.user_id
    GROUP BY uc.cohort_month, DATE_TRUNC('month', o.ordered_at)
),
-- Calculate cohort size
cohort_size AS (
    SELECT cohort_month, COUNT(*) AS total_users
    FROM user_cohort
    GROUP BY cohort_month
)
SELECT 
    ua.cohort_month,
    ua.activity_month,
    EXTRACT(MONTH FROM AGE(ua.activity_month, ua.cohort_month)) AS months_since_join,
    ua.active_users,
    cs.total_users AS cohort_size,
    ROUND(ua.active_users * 100.0 / cs.total_users, 2) AS retention_pct
FROM user_activity ua
JOIN cohort_size cs ON ua.cohort_month = cs.cohort_month
ORDER BY ua.cohort_month, ua.activity_month;
```

### Q10: Dense ranking with tiebreaker

```sql
-- Rank products by revenue, with ties broken by units sold (descending)
SELECT 
    product_name,
    revenue,
    units_sold,
    RANK() OVER (ORDER BY revenue DESC) AS revenue_rank,
    ROW_NUMBER() OVER (ORDER BY revenue DESC, units_sold DESC) AS unique_rank
FROM product_sales;
```

---

## 11. Window Function Cheat Sheet

| Function | Category | Description |
|----------|----------|-------------|
| `ROW_NUMBER()` | Ranking | Unique sequential number per partition |
| `RANK()` | Ranking | Rank with gaps on ties |
| `DENSE_RANK()` | Ranking | Rank without gaps on ties |
| `NTILE(n)` | Ranking | Divide into n equal groups |
| `LAG(col, n, default)` | Value | Access previous row's value |
| `LEAD(col, n, default)` | Value | Access next row's value |
| `FIRST_VALUE(col)` | Value | First value in window frame |
| `LAST_VALUE(col)` | Value | Last value in window frame |
| `NTH_VALUE(col, n)` | Value | Nth value in window frame |
| `SUM() OVER(...)` | Aggregate | Running/windowed sum |
| `AVG() OVER(...)` | Aggregate | Running/windowed average |
| `COUNT() OVER(...)` | Aggregate | Running/windowed count |
| `MIN() OVER(...)` | Aggregate | Windowed minimum |
| `MAX() OVER(...)` | Aggregate | Windowed maximum |
| `PERCENT_RANK()` | Distribution | Relative rank (0 to 1) |
| `CUME_DIST()` | Distribution | Cumulative distribution |
| `PERCENTILE_CONT(p)` | Distribution | Interpolated percentile |
| `PERCENTILE_DISC(p)` | Distribution | Discrete percentile |

### Named Windows (WINDOW clause)

```sql
-- Define reusable windows
SELECT 
    name, department_id, salary,
    RANK() OVER dept_salary AS rank_in_dept,
    AVG(salary) OVER dept_salary AS dept_avg,
    SUM(salary) OVER by_dept AS dept_total
FROM employees
WINDOW 
    dept_salary AS (PARTITION BY department_id ORDER BY salary DESC),
    by_dept AS (PARTITION BY department_id);
```

---

## 12. Performance Considerations

| Consideration | Impact |
|--------------|--------|
| Window functions execute after WHERE, GROUP BY, HAVING | Can't be used in WHERE |
| Each OVER clause may require a separate sort | Multiple sorts = slower |
| Use WINDOW clause to reduce redundant sorts | Optimizer can combine |
| ROWS is faster than RANGE | RANGE needs value comparison |
| Avoid window functions on very large result sets | Memory-intensive |
| Consider pre-aggregation with GROUP BY first | Then apply window functions |
