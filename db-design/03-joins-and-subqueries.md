# Joins & Subqueries — Deep Dive

## Mastering Every JOIN Type and Subquery Pattern

---

## 1. Setup — Sample Tables

We'll use these tables throughout this document:

```sql
-- Departments
CREATE TABLE departments (
    id      SERIAL PRIMARY KEY,
    name    VARCHAR(100) NOT NULL
);
INSERT INTO departments (id, name) VALUES
(1, 'Engineering'), (2, 'Marketing'), (3, 'Sales'), (4, 'HR');

-- Employees
CREATE TABLE employees (
    id              SERIAL PRIMARY KEY,
    name            VARCHAR(100) NOT NULL,
    department_id   INTEGER REFERENCES departments(id),
    manager_id      INTEGER REFERENCES employees(id),
    salary          DECIMAL(10,2) NOT NULL,
    hire_date       DATE NOT NULL
);
INSERT INTO employees (id, name, department_id, manager_id, salary, hire_date) VALUES
(1,  'Alice',   1, NULL,  150000, '2018-01-15'),
(2,  'Bob',     1, 1,     120000, '2019-03-20'),
(3,  'Carol',   1, 1,      95000, '2020-07-10'),
(4,  'Dave',    2, NULL,  130000, '2017-06-01'),
(5,  'Eve',     2, 4,     110000, '2019-09-15'),
(6,  'Frank',   3, NULL,  105000, '2020-01-20'),
(7,  'Grace',   3, 6,      85000, '2021-04-12'),
(8,  'Hank',    NULL, NULL, 90000, '2022-08-01'); -- No department!

-- Projects
CREATE TABLE projects (
    id              SERIAL PRIMARY KEY,
    name            VARCHAR(100) NOT NULL,
    department_id   INTEGER REFERENCES departments(id),
    budget          DECIMAL(12,2)
);
INSERT INTO projects (id, name, department_id, budget) VALUES
(1, 'Project Alpha', 1, 500000),
(2, 'Project Beta',  1, 300000),
(3, 'Project Gamma', 2, 200000),
(4, 'Project Delta', 5, 150000); -- department_id=5 doesn't exist!

-- Employee_Projects (junction table)
CREATE TABLE employee_projects (
    employee_id INTEGER REFERENCES employees(id),
    project_id  INTEGER REFERENCES projects(id),
    role        VARCHAR(50),
    PRIMARY KEY (employee_id, project_id)
);
INSERT INTO employee_projects VALUES
(1, 1, 'Lead'), (2, 1, 'Developer'), (3, 1, 'Developer'),
(2, 2, 'Lead'), (3, 2, 'Developer'),
(4, 3, 'Lead'), (5, 3, 'Developer'),
(1, 2, 'Reviewer');
```

---

## 2. JOIN Types — Visual Guide

```
TABLE A (employees)          TABLE B (departments)
┌────┬───────┬────────┐      ┌────┬─────────────┐
│ id │ name  │ dept_id│      │ id │ name        │
├────┼───────┼────────┤      ├────┼─────────────┤
│ 1  │ Alice │ 1      │      │ 1  │ Engineering │
│ 2  │ Bob   │ 1      │      │ 2  │ Marketing   │
│ 3  │ Carol │ 1      │      │ 3  │ Sales       │
│ 4  │ Dave  │ 2      │      │ 4  │ HR          │
│ 5  │ Eve   │ 2      │      └────┴─────────────┘
│ 6  │ Frank │ 3      │
│ 7  │ Grace │ 3      │
│ 8  │ Hank  │ NULL   │ ← No department
└────┴───────┴────────┘

INNER JOIN:      Only matching rows from both tables
                 (Hank excluded — no dept; HR excluded — no employees)

LEFT JOIN:       All from A + matching from B
                 (Hank included with NULL dept; HR excluded)

RIGHT JOIN:      All from B + matching from A
                 (Hank excluded; HR included with NULL employee)

FULL OUTER JOIN: All from both, NULLs where no match
                 (Hank AND HR both included)

CROSS JOIN:      Every row from A × every row from B
                 (8 × 4 = 32 rows)
```

---

## 3. INNER JOIN

Returns only rows with matching values in **both** tables.

```sql
-- Basic INNER JOIN
SELECT e.name AS employee, d.name AS department
FROM employees e
INNER JOIN departments d ON e.department_id = d.id;

-- Result (7 rows — Hank excluded):
-- Alice    | Engineering
-- Bob      | Engineering
-- Carol    | Engineering
-- Dave     | Marketing
-- Eve      | Marketing
-- Frank    | Sales
-- Grace    | Sales

-- Multi-table INNER JOIN
SELECT e.name AS employee, d.name AS department, p.name AS project, ep.role
FROM employees e
INNER JOIN departments d ON e.department_id = d.id
INNER JOIN employee_projects ep ON e.id = ep.employee_id
INNER JOIN projects p ON ep.project_id = p.id;

-- INNER JOIN with additional conditions
SELECT e.name, d.name AS department
FROM employees e
INNER JOIN departments d ON e.department_id = d.id AND e.salary > 100000;
-- Note: condition in ON vs WHERE — for INNER JOIN, same result
-- For OUTER JOINs, it matters! (see LEFT JOIN section)
```

---

## 4. LEFT JOIN (LEFT OUTER JOIN)

Returns ALL rows from the **left** table + matching from right (NULLs if no match).

```sql
-- Basic LEFT JOIN
SELECT e.name AS employee, d.name AS department
FROM employees e
LEFT JOIN departments d ON e.department_id = d.id;

-- Result (8 rows — including Hank):
-- Alice    | Engineering
-- Bob      | Engineering
-- Carol    | Engineering
-- Dave     | Marketing
-- Eve      | Marketing
-- Frank    | Sales
-- Grace    | Sales
-- Hank     | NULL          ← included with NULL department

-- LEFT JOIN to find unmatched rows (anti-join pattern)
SELECT e.name AS employee
FROM employees e
LEFT JOIN departments d ON e.department_id = d.id
WHERE d.id IS NULL;
-- Result: Hank (employees without a department)

-- ⚠️ CRITICAL: WHERE vs ON in LEFT JOIN
-- Condition in ON: filters the right table BEFORE joining
-- Condition in WHERE: filters the result AFTER joining

-- Example: "All employees, with their department IF salary > 100000"
SELECT e.name, e.salary, d.name AS department
FROM employees e
LEFT JOIN departments d ON e.department_id = d.id AND e.salary > 100000;
-- All 8 employees returned; only those with salary > 100K have dept filled

-- vs. "Only employees with salary > 100000, with their department"
SELECT e.name, e.salary, d.name AS department
FROM employees e
LEFT JOIN departments d ON e.department_id = d.id
WHERE e.salary > 100000;
-- Only 4 employees returned (Alice, Bob, Dave, Eve)
```

---

## 5. RIGHT JOIN (RIGHT OUTER JOIN)

Returns ALL rows from the **right** table + matching from left.

```sql
-- Basic RIGHT JOIN
SELECT e.name AS employee, d.name AS department
FROM employees e
RIGHT JOIN departments d ON e.department_id = d.id;

-- Result (8 rows — including HR with no employees):
-- Alice    | Engineering
-- Bob      | Engineering
-- Carol    | Engineering
-- Dave     | Marketing
-- Eve      | Marketing
-- Frank    | Sales
-- Grace    | Sales
-- NULL     | HR            ← included, no employee in HR

-- Departments without employees
SELECT d.name AS department
FROM employees e
RIGHT JOIN departments d ON e.department_id = d.id
WHERE e.id IS NULL;
-- Result: HR

-- Note: RIGHT JOIN can always be rewritten as LEFT JOIN by swapping tables
-- Most developers prefer LEFT JOIN for readability
```

---

## 6. FULL OUTER JOIN

Returns ALL rows from **both** tables, with NULLs where there's no match.

```sql
-- Basic FULL OUTER JOIN
SELECT e.name AS employee, d.name AS department
FROM employees e
FULL OUTER JOIN departments d ON e.department_id = d.id;

-- Result (9 rows):
-- Alice    | Engineering
-- Bob      | Engineering
-- Carol    | Engineering
-- Dave     | Marketing
-- Eve      | Marketing
-- Frank    | Sales
-- Grace    | Sales
-- Hank     | NULL          ← employee without department
-- NULL     | HR            ← department without employee

-- Find ALL unmatched rows (from either side)
SELECT 
    COALESCE(e.name, 'NO EMPLOYEE') AS employee,
    COALESCE(d.name, 'NO DEPARTMENT') AS department
FROM employees e
FULL OUTER JOIN departments d ON e.department_id = d.id
WHERE e.id IS NULL OR d.id IS NULL;
-- Result: Hank | NO DEPARTMENT
--         NO EMPLOYEE | HR

-- MySQL doesn't support FULL OUTER JOIN — use UNION:
SELECT e.name, d.name
FROM employees e LEFT JOIN departments d ON e.department_id = d.id
UNION
SELECT e.name, d.name
FROM employees e RIGHT JOIN departments d ON e.department_id = d.id;
```

---

## 7. CROSS JOIN

Returns the **Cartesian product** — every row from A paired with every row from B.

```sql
-- Explicit CROSS JOIN
SELECT e.name, d.name
FROM employees e
CROSS JOIN departments d;
-- If employees has 8 rows and departments has 4 rows → 32 result rows

-- Implicit CROSS JOIN (comma syntax)
SELECT e.name, d.name
FROM employees e, departments d;
-- Same result as above

-- Practical use case: Generate all possible combinations
-- Example: Generate a report with all months × all products
SELECT m.month_name, p.name AS product
FROM (
    SELECT generate_series(1, 12) AS month_num,
           TO_CHAR(TO_DATE(generate_series(1,12)::TEXT, 'MM'), 'Month') AS month_name
) m
CROSS JOIN products p;

-- Practical use case: Compare every employee to every other
SELECT a.name AS emp1, b.name AS emp2, 
       ABS(a.salary - b.salary) AS salary_diff
FROM employees a
CROSS JOIN employees b
WHERE a.id < b.id  -- avoid pairing with self and duplicates
ORDER BY salary_diff DESC;
```

---

## 8. SELF JOIN

A table joined to **itself** — for hierarchical or comparative queries.

```sql
-- Employee-Manager relationship
SELECT 
    e.name AS employee, 
    m.name AS manager
FROM employees e
LEFT JOIN employees m ON e.manager_id = m.id;

-- Result:
-- Alice    | NULL    (CEO, no manager)
-- Bob      | Alice
-- Carol    | Alice
-- Dave     | NULL    (dept head, no manager)
-- Eve      | Dave
-- Frank    | NULL
-- Grace    | Frank
-- Hank     | NULL

-- Employees who earn more than their manager
SELECT e.name AS employee, e.salary, m.name AS manager, m.salary AS manager_salary
FROM employees e
INNER JOIN employees m ON e.manager_id = m.id
WHERE e.salary > m.salary;

-- Employees in the same department
SELECT a.name AS emp1, b.name AS emp2, d.name AS department
FROM employees a
INNER JOIN employees b ON a.department_id = b.department_id AND a.id < b.id
INNER JOIN departments d ON a.department_id = d.id;
-- a.id < b.id prevents duplicate pairs (Alice-Bob = Bob-Alice)
```

---

## 9. NATURAL JOIN

Automatically joins on columns with the **same name** in both tables. 

```sql
-- NATURAL JOIN (USE WITH CAUTION)
SELECT e.name, department_id
FROM employees e
NATURAL JOIN departments d;
-- Joins on ALL columns with same name
-- ⚠️ Dangerous: if both tables have 'id', it joins on id too!

-- This is equivalent to:
SELECT e.name, e.department_id
FROM employees e
INNER JOIN departments d ON e.department_id = d.department_id AND e.id = d.id;

-- 🚫 Interview tip: NEVER recommend NATURAL JOIN in interviews
-- It's fragile — adding a same-named column to either table changes behavior
-- Always use explicit ON conditions
```

---

## 10. USING Clause

A shorthand when join columns have the **same name**.

```sql
-- USING clause
SELECT e.name, department_id, d.name AS dept_name
FROM employees e
JOIN departments d USING (department_id);
-- Cleaner than ON when column names match

-- Note: with USING, the shared column doesn't need table prefix
-- "department_id" instead of "e.department_id" or "d.department_id"

-- Multiple columns: USING (col1, col2)
```

---

## 11. Anti-Join and Semi-Join Patterns

These aren't SQL syntax but **logical patterns** frequently asked in interviews.

### Semi-Join: "Rows in A that have a match in B"

```sql
-- Method 1: EXISTS (preferred — stops at first match)
SELECT e.name, e.salary
FROM employees e
WHERE EXISTS (
    SELECT 1 FROM employee_projects ep WHERE ep.employee_id = e.id
);

-- Method 2: IN
SELECT e.name, e.salary
FROM employees e
WHERE e.id IN (SELECT employee_id FROM employee_projects);

-- Method 3: JOIN with DISTINCT
SELECT DISTINCT e.name, e.salary
FROM employees e
JOIN employee_projects ep ON e.id = ep.employee_id;
```

### Anti-Join: "Rows in A that have NO match in B"

```sql
-- Method 1: NOT EXISTS (preferred — most reliable with NULLs)
SELECT e.name, e.salary
FROM employees e
WHERE NOT EXISTS (
    SELECT 1 FROM employee_projects ep WHERE ep.employee_id = e.id
);

-- Method 2: LEFT JOIN + IS NULL
SELECT e.name, e.salary
FROM employees e
LEFT JOIN employee_projects ep ON e.id = ep.employee_id
WHERE ep.employee_id IS NULL;

-- Method 3: NOT IN (⚠️ CAREFUL WITH NULLs!)
SELECT e.name, e.salary
FROM employees e
WHERE e.id NOT IN (SELECT employee_id FROM employee_projects);
-- ⚠️ If subquery returns ANY NULL, entire NOT IN returns NULL (no rows!)
-- Fix: add WHERE employee_id IS NOT NULL to subquery

-- Performance comparison:
-- EXISTS:              Stops at first match ✅
-- LEFT JOIN + IS NULL: Full join, then filter ⚠️
-- NOT IN:              Converts to OR list, NULL-unsafe ❌
```

---

## 12. Subqueries — Complete Guide

### Types of Subqueries

| Type | Returns | Used In | Example |
|------|---------|---------|---------|
| **Scalar** | Single value | SELECT, WHERE, HAVING | `(SELECT MAX(salary) FROM employees)` |
| **Row** | Single row, multiple columns | WHERE | `WHERE (dept, salary) = (SELECT ...)` |
| **Table** | Multiple rows and columns | FROM, JOIN | `FROM (SELECT ... ) AS sub` |
| **Correlated** | References outer query | WHERE, SELECT | `WHERE salary > (SELECT AVG... WHERE dept=e.dept)` |

### Scalar Subquery

```sql
-- In SELECT clause
SELECT 
    name, 
    salary,
    salary - (SELECT AVG(salary) FROM employees) AS vs_avg
FROM employees;

-- In WHERE clause
SELECT name, salary
FROM employees
WHERE salary > (SELECT AVG(salary) FROM employees);

-- In HAVING clause
SELECT department_id, AVG(salary) AS avg_salary
FROM employees
GROUP BY department_id
HAVING AVG(salary) > (SELECT AVG(salary) FROM employees);
```

### Row Subquery

```sql
-- Find employee with highest salary (row comparison)
SELECT name, salary, department_id
FROM employees
WHERE (department_id, salary) = (
    SELECT department_id, MAX(salary)
    FROM employees
    GROUP BY department_id
    ORDER BY MAX(salary) DESC
    LIMIT 1
);
```

### Table Subquery (Derived Table)

```sql
-- Subquery in FROM — must be aliased
SELECT dept_stats.department_name, dept_stats.avg_salary, dept_stats.emp_count
FROM (
    SELECT 
        d.name AS department_name,
        AVG(e.salary) AS avg_salary,
        COUNT(*) AS emp_count
    FROM employees e
    JOIN departments d ON e.department_id = d.id
    GROUP BY d.name
) AS dept_stats
WHERE dept_stats.emp_count > 2;

-- Inline view for Top-N per group
SELECT * FROM (
    SELECT 
        name, department_id, salary,
        ROW_NUMBER() OVER (PARTITION BY department_id ORDER BY salary DESC) AS rn
    FROM employees
) ranked
WHERE rn <= 2;
-- Top 2 highest-paid employees per department
```

### Correlated Subquery

A subquery that **references a column from the outer query**. Executes once per outer row.

```sql
-- Employees earning above their department average
SELECT e.name, e.salary, e.department_id
FROM employees e
WHERE e.salary > (
    SELECT AVG(e2.salary)
    FROM employees e2
    WHERE e2.department_id = e.department_id  -- correlation!
);

-- Department with highest average salary
SELECT d.name, 
    (SELECT AVG(e.salary) FROM employees e WHERE e.department_id = d.id) AS avg_salary
FROM departments d
ORDER BY avg_salary DESC NULLS LAST
LIMIT 1;

-- Exists with correlation
SELECT d.name
FROM departments d
WHERE EXISTS (
    SELECT 1 FROM employees e 
    WHERE e.department_id = d.id AND e.salary > 100000
);
-- Departments that have at least one employee earning > 100K
```

### ALL, ANY, SOME

```sql
-- ALL: compare to every value in subquery
-- "Employees who earn more than ALL employees in Marketing"
SELECT name, salary
FROM employees
WHERE salary > ALL (
    SELECT salary FROM employees WHERE department_id = 2
);

-- ANY / SOME: compare to at least one value
-- "Employees who earn more than ANY employee in Marketing"
SELECT name, salary
FROM employees
WHERE salary > ANY (
    SELECT salary FROM employees WHERE department_id = 2
);
-- Equivalent to: salary > (SELECT MIN(salary) ... WHERE dept=2)

-- > ALL (list)  ≡  > MAX(list)
-- > ANY (list)  ≡  > MIN(list)
-- < ALL (list)  ≡  < MIN(list)
-- < ANY (list)  ≡  < MAX(list)
-- = ANY (list)  ≡  IN (list)
```

---

## 13. LATERAL JOIN (PostgreSQL)

A subquery in FROM that can reference earlier tables in the FROM clause.

```sql
-- "Top 3 highest-paid employees per department"
SELECT d.name AS department, top_emp.name, top_emp.salary
FROM departments d
CROSS JOIN LATERAL (
    SELECT e.name, e.salary
    FROM employees e
    WHERE e.department_id = d.id
    ORDER BY e.salary DESC
    LIMIT 3
) AS top_emp;

-- Equivalent to correlated subquery in FROM (not normally allowed!)
-- Without LATERAL, subqueries in FROM can't reference other FROM items

-- Use case: Unpack arrays
SELECT u.name, t.tag
FROM users u
CROSS JOIN LATERAL UNNEST(u.tags) AS t(tag);
```

---

## 14. Common Join Interview Questions

### Q1: Find employees who are NOT assigned to any project

```sql
-- Anti-join pattern
SELECT e.name
FROM employees e
LEFT JOIN employee_projects ep ON e.id = ep.employee_id
WHERE ep.employee_id IS NULL;

-- Or with NOT EXISTS
SELECT e.name
FROM employees e
WHERE NOT EXISTS (
    SELECT 1 FROM employee_projects ep WHERE ep.employee_id = e.id
);
```

### Q2: Find departments with their employee count (including departments with 0 employees)

```sql
SELECT d.name, COUNT(e.id) AS employee_count
FROM departments d
LEFT JOIN employees e ON d.id = e.department_id
GROUP BY d.name
ORDER BY employee_count DESC;
-- Must use LEFT JOIN (not INNER) to include HR dept with 0 employees
-- Must use COUNT(e.id) not COUNT(*) — COUNT(*) would count 1 for HR
```

### Q3: Find the second highest salary

```sql
-- Method 1: Subquery
SELECT MAX(salary) AS second_highest
FROM employees
WHERE salary < (SELECT MAX(salary) FROM employees);

-- Method 2: OFFSET
SELECT DISTINCT salary
FROM employees
ORDER BY salary DESC
LIMIT 1 OFFSET 1;

-- Method 3: Window function
SELECT salary
FROM (
    SELECT salary, DENSE_RANK() OVER (ORDER BY salary DESC) as rnk
    FROM employees
) ranked
WHERE rnk = 2;
```

### Q4: Find employees who work on ALL projects in their department

```sql
-- Division operation — "who works on all?"
SELECT e.name
FROM employees e
WHERE NOT EXISTS (
    -- Projects in employee's department
    SELECT p.id FROM projects p
    WHERE p.department_id = e.department_id
    EXCEPT
    -- Projects employee is assigned to
    SELECT ep.project_id FROM employee_projects ep
    WHERE ep.employee_id = e.id
);
-- If there are NO unassigned projects, the employee works on all of them
```

### Q5: Employees with their manager's name and manager's salary (3-level join)

```sql
SELECT 
    e.name AS employee,
    e.salary AS emp_salary,
    m.name AS manager,
    m.salary AS mgr_salary,
    mm.name AS manager_of_manager
FROM employees e
LEFT JOIN employees m ON e.manager_id = m.id
LEFT JOIN employees mm ON m.manager_id = mm.id;
```

### Q6: Find duplicate emails

```sql
-- Duplicates
SELECT email, COUNT(*) AS count
FROM users
GROUP BY email
HAVING COUNT(*) > 1;

-- Delete duplicates, keep lowest id
DELETE FROM users
WHERE id NOT IN (
    SELECT MIN(id)
    FROM users
    GROUP BY email
);

-- PostgreSQL: DELETE with self-join
DELETE FROM users a
USING users b
WHERE a.email = b.email AND a.id > b.id;
```

### Q7: Pivot — count employees by department and status

```sql
SELECT 
    d.name AS department,
    COUNT(CASE WHEN e.status = 'active' THEN 1 END) AS active,
    COUNT(CASE WHEN e.status = 'inactive' THEN 1 END) AS inactive,
    COUNT(CASE WHEN e.status = 'terminated' THEN 1 END) AS terminated
FROM departments d
LEFT JOIN employees e ON d.id = e.department_id
GROUP BY d.name;
```

### Q8: Running total of orders per user

```sql
SELECT 
    user_id,
    order_date,
    total,
    SUM(total) OVER (
        PARTITION BY user_id 
        ORDER BY order_date 
        ROWS UNBOUNDED PRECEDING
    ) AS running_total
FROM orders;
```

### Q9: Find users who placed orders in EVERY month of 2024

```sql
SELECT u.name
FROM users u
JOIN orders o ON u.id = o.user_id
WHERE o.ordered_at >= '2024-01-01' AND o.ordered_at < '2025-01-01'
GROUP BY u.id, u.name
HAVING COUNT(DISTINCT EXTRACT(MONTH FROM o.ordered_at)) = 12;
```

### Q10: Recursive self-join — get full management chain

```sql
-- Recursive CTE for hierarchy
WITH RECURSIVE management_chain AS (
    -- Base case: start with a specific employee
    SELECT id, name, manager_id, 1 AS level
    FROM employees
    WHERE id = 3  -- Carol
    
    UNION ALL
    
    -- Recursive case: find manager of current level
    SELECT e.id, e.name, e.manager_id, mc.level + 1
    FROM employees e
    INNER JOIN management_chain mc ON e.id = mc.manager_id
)
SELECT * FROM management_chain;

-- Result:
-- 3 | Carol | 1    | 1  (Carol reports to Alice)
-- 1 | Alice | NULL | 2  (Alice is top-level)
```

---

## 15. JOIN Performance Tips

| Tip | Why |
|-----|-----|
| Index foreign key columns | JOINs use these for lookups |
| Use EXPLAIN ANALYZE | See which join strategy the optimizer picks |
| Prefer EXISTS over IN for large subqueries | EXISTS stops at first match |
| Filter early (WHERE before JOIN if possible) | Reduce rows before joining |
| Avoid SELECT * with JOINs | Only select columns you need |
| Use INNER JOIN over OUTER when possible | Less data, simpler plan |
| Consider materialized views for complex joins | Pre-compute expensive joins |

### Join Algorithms

```
NESTED LOOP JOIN:
  For each row in A:
    For each row in B:
      If A.key = B.key → output
  O(n × m) — good for small tables or indexed inner table

HASH JOIN:
  Build hash table from smaller table
  Probe with larger table
  O(n + m) — good for large unsorted tables

MERGE JOIN (Sort-Merge):
  Sort both tables on join key
  Merge like merge sort
  O(n log n + m log m) — good if tables already sorted or indexed
```
