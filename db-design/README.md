# Database Design — Interview Preparation

Comprehensive guide for database design interviews — from identifying entities and designing schemas to writing complex SQL queries, plus essential NoSQL patterns.

---

## SQL (Primary Focus)

| # | Topic | Description | Link |
|---|-------|-------------|------|
| 01 | DB Design Process | Entity identification, ER diagrams, normalization, schema design methodology | [Guide](01-db-design-process.md) |
| 02 | SQL Fundamentals | Data types, constraints, DDL/DML, all clause types | [Guide](02-sql-fundamentals.md) |
| 03 | Joins & Subqueries | All join types, correlated subqueries, CTEs, derived tables | [Guide](03-joins-and-subqueries.md) |
| 04 | Aggregations & Window Functions | GROUP BY, HAVING, ROLLUP, RANK, LAG/LEAD, running totals | [Guide](04-aggregations-and-window-functions.md) |
| 05 | Advanced SQL | Recursive CTEs, pivoting, JSON, full-text search, query optimization | [Guide](05-advanced-sql.md) |
| 06 | SQL Interview Questions | 50+ interview questions with solutions across all difficulty levels | [Guide](06-sql-interview-questions.md) |

## NoSQL

| # | Topic | Description | Link |
|---|-------|-------------|------|
| 07 | NoSQL Essentials | MongoDB, Redis, Cassandra, DynamoDB — key patterns and examples | [Guide](07-nosql-essentials.md) |

---

## How to Use This Guide

1. **Start with [01-db-design-process](01-db-design-process.md)** — learn how to approach any DB design question
2. **Master SQL fundamentals** in [02](02-sql-fundamentals.md) and [03](03-joins-and-subqueries.md)
3. **Level up** with [04](04-aggregations-and-window-functions.md) and [05](05-advanced-sql.md)
4. **Practice** with [06](06-sql-interview-questions.md) — 50+ real interview questions
5. **Review NoSQL** patterns in [07](07-nosql-essentials.md) for completeness

---

## Quick Reference: SQL Concept Map

```
SQL Concepts
├── DDL (Data Definition Language)
│   ├── CREATE, ALTER, DROP, TRUNCATE
│   └── Constraints: PK, FK, UNIQUE, CHECK, DEFAULT, NOT NULL
├── DML (Data Manipulation Language)
│   ├── INSERT, UPDATE, DELETE, MERGE/UPSERT
│   └── SELECT (the big one)
├── SELECT Mastery
│   ├── WHERE, ORDER BY, LIMIT/OFFSET
│   ├── JOINs: INNER, LEFT, RIGHT, FULL, CROSS, SELF, NATURAL
│   ├── Subqueries: Scalar, Row, Table, Correlated, EXISTS/NOT EXISTS
│   ├── CTEs: WITH, Recursive CTEs
│   ├── Aggregations: COUNT, SUM, AVG, MIN, MAX, GROUP BY, HAVING
│   ├── Window Functions: ROW_NUMBER, RANK, DENSE_RANK, NTILE
│   │   ├── LAG, LEAD, FIRST_VALUE, LAST_VALUE
│   │   ├── SUM/AVG/COUNT OVER (PARTITION BY ... ORDER BY ...)
│   │   └── Frame: ROWS/RANGE BETWEEN
│   ├── Set Operations: UNION, INTERSECT, EXCEPT
│   └── Advanced: PIVOT, ROLLUP, CUBE, GROUPING SETS
├── TCL (Transaction Control)
│   ├── BEGIN, COMMIT, ROLLBACK, SAVEPOINT
│   └── Isolation Levels
├── DCL (Data Control Language)
│   ├── GRANT, REVOKE
│   └── Roles and Permissions
├── Indexing
│   ├── B-Tree, Hash, GIN, GiST
│   ├── Composite, Partial, Covering
│   └── EXPLAIN ANALYZE
└── Schema Design
    ├── Normalization: 1NF → 2NF → 3NF → BCNF
    ├── Denormalization strategies
    └── ER Diagrams
```
