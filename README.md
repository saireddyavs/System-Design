# System Design & Engineering Concepts

Staff+ interview preparation — HLD, LLD, and deep engineering fundamentals.

---

## High-Level Design (HLD)

> [Full HLD README](HLD/README.md)

| Section | Topics |
|---------|--------|
| [Fundamentals](HLD/01-fundamentals/) | Scalability, CAP, Consistent Hashing, Estimation |
| [Networking](HLD/02-networking/) | DNS, TCP/UDP, HTTP, Load Balancing, CDN, Proxies |
| [APIs](HLD/03-apis/) | REST/gRPC/GraphQL, Rate Limiting, Idempotency, Webhooks |
| [Databases](HLD/04-databases/) | SQL vs NoSQL, ACID, Sharding, Replication, Bloom Filters |
| [Caching](HLD/05-caching/) | Strategies, Eviction, Distributed Caching, Redis |
| [Messaging](HLD/06-messaging/) | Kafka, Pub/Sub, Message Queues, CDC |
| [Distributed Systems](HLD/07-distributed-systems/) | Consensus, Gossip, Clocks, Locking, Leader Election |
| [Architecture](HLD/08-architecture/) | Microservices, Event-Driven, Circuit Breaker, Service Mesh |
| [Reliability](HLD/09-reliability/) | Fault Tolerance, Retries, Disaster Recovery |
| [Security](HLD/10-security/) | Auth, Encryption, API Security, Secrets |
| [Observability](HLD/11-observability/) | Logging, Metrics, Tracing |
| [Storage & Data](HLD/12-storage-and-data/) | OLTP vs OLAP, Data Lakes, Storage Types |
| [Tradeoffs](HLD/13-tradeoffs/) | System Design Decision Framework |
| [Estimation](HLD/14-numbers-and-estimation/) | Latency Numbers, Scale Reference |
| [Interview Problems](HLD/15-interview-problems/) | [Easy](HLD/15-interview-problems/easy/) &#124; [Medium](HLD/15-interview-problems/medium/) &#124; [Hard](HLD/15-interview-problems/hard/) |
| [Case Studies](HLD/16-case-studies/) | Company Architectures, Scaling Stories |

---

## Low-Level Design (LLD)

17 Go projects — each demonstrates SOLID, design patterns, and concurrency.

| # | Problem | Patterns | Links |
|---|---------|----------|-------|
| 01 | Parking Lot | Strategy, DIP | [README](LLD/01-parking-lot-system/README.md) &#124; [Interview](LLD/01-parking-lot-system/interview.md) |
| 02 | Online Bookstore | Strategy, Observer, Factory | [README](LLD/02-online-bookstore/README.md) &#124; [Interview](LLD/02-online-bookstore/interview.md) |
| 03 | Library Management | Observer, Strategy, Facade | [README](LLD/03-library-management-system/README.md) &#124; [Interview](LLD/03-library-management-system/interview.md) |
| 04 | Movie Ticket Booking | Strategy, Observer, Builder | [README](LLD/04-movie-ticket-booking/README.md) &#124; [Interview](LLD/04-movie-ticket-booking/interview.md) |
| 05 | Elevator System | Strategy, State | [README](LLD/05-elevator-system/README.md) &#124; [Interview](LLD/05-elevator-system/interview.md) |
| 06 | Hotel Management | Strategy, Observer, Factory | [README](LLD/06-hotel-management-system/README.md) &#124; [Interview](LLD/06-hotel-management-system/interview.md) |
| 07 | Ride Sharing (Uber) | Strategy, Observer | [README](LLD/07-ride-sharing-service/README.md) &#124; [Interview](LLD/07-ride-sharing-service/interview.md) |
| 08 | File Storage (Dropbox) | Composite, Observer | [README](LLD/08-file-storage-system/README.md) &#124; [Interview](LLD/08-file-storage-system/interview.md) |
| 09 | Chat Application | Pub/Sub, Strategy, Factory | [README](LLD/09-chat-application/README.md) &#124; [Interview](LLD/09-chat-application/interview.md) |
| 10 | Social Media | Strategy, Observer | [README](LLD/10-social-media-platform/README.md) &#124; [Interview](LLD/10-social-media-platform/interview.md) |
| 11 | Notification System | Strategy, Decorator, Chain | [README](LLD/11-notification-system/README.md) &#124; [Interview](LLD/11-notification-system/interview.md) |
| 12 | Airline Reservation | Strategy, Observer, Builder | [README](LLD/12-airline-reservation-system/README.md) &#124; [Interview](LLD/12-airline-reservation-system/interview.md) |
| 13 | ATM System | Command, State, CoR | [README](LLD/13-atm-system/README.md) &#124; [Interview](LLD/13-atm-system/interview.md) |
| 14 | E-Commerce | Strategy, Observer, Factory | [README](LLD/14-ecommerce-website/README.md) &#124; [Interview](LLD/14-ecommerce-website/interview.md) |
| 15 | Food Delivery | Strategy, Observer | [README](LLD/15-food-delivery-system/README.md) &#124; [Interview](LLD/15-food-delivery-system/interview.md) |
| 16 | Shopping Cart | Strategy, Observer | [README](LLD/16-shopping-cart-system/README.md) &#124; [Interview](LLD/16-shopping-cart-system/interview.md) |
| 17 | Splitwise | Strategy, Builder | [README](LLD/17-splitwise/README.md) &#124; [Interview](LLD/17-splitwise/interview.md) |

---

## Engineering Concepts

> [Curriculum Index](engineering_concepts/00_curriculum_index.md) — 100-day deep dive

| # | Module | Link |
|---|--------|------|
| 01 | Scale & Data Structures | [Notes](engineering_concepts/01_scale_data_structures.md) |
| 02 | Database Engineering | [Notes](engineering_concepts/02_database_engineering.md) |
| 03 | Consensus & Coordination | [Notes](engineering_concepts/03_consensus_coordination.md) |
| 04 | Networking | [Notes](engineering_concepts/04_networking.md) |
| 05 | Architecture Patterns | [Notes](engineering_concepts/05_architecture_patterns.md) |
| 06 | Rate Limiting & Load Balancing | [Notes](engineering_concepts/06_rate_limiting_lb.md) |
| 07 | Caching | [Notes](engineering_concepts/07_caching.md) |
| 08 | Messaging & Streams | [Notes](engineering_concepts/08_messaging_streams.md) |
| 09 | Concurrency & Lock-Free | [Notes](engineering_concepts/09_concurrency.md) |
| 10 | DB Isolation & Failures | [Notes](engineering_concepts/10_isolation_failures.md) |
| 11 | GC & Memory | [Notes](engineering_concepts/11_gc_memory.md) |
| 12 | OS Primitives | [Notes](engineering_concepts/12_os_primitives.md) |
| 13 | Storage & Compression | [Notes](engineering_concepts/13_storage_compression.md) |
| 14 | Security & Identity | [Notes](engineering_concepts/14_security.md) |
| 15 | Reliability & Ops | [Notes](engineering_concepts/15_reliability_ops.md) |
| 16 | Case Studies | [Notes](engineering_concepts/16_case_studies.md) |
| 17 | AI Engineering | [Notes](engineering_concepts/17_ai_engineering.md) |
| 18 | White Papers | [Notes](engineering_concepts/18_white_papers.md) |

---

## Database Design (Interview Prep)

> [Full DB Design README](db-design/README.md)

| # | Topic | Description |
|---|-------|-------------|
| 01 | [DB Design Process](db-design/01-db-design-process.md) | Entity identification, ER diagrams, normalization (1NF→BCNF), full schema examples |
| 02 | [SQL Fundamentals](db-design/02-sql-fundamentals.md) | DDL/DML, data types, constraints, SELECT execution order, indexes, transactions |
| 03 | [Joins & Subqueries](db-design/03-joins-and-subqueries.md) | All JOIN types, correlated subqueries, anti-join/semi-join, LATERAL |
| 04 | [Aggregations & Windows](db-design/04-aggregations-and-window-functions.md) | GROUP BY, HAVING, ROLLUP/CUBE, ROW_NUMBER, LAG/LEAD, running totals |
| 05 | [Advanced SQL](db-design/05-advanced-sql.md) | Recursive CTEs, pivoting, JSON/JSONB, full-text search, query optimization |
| 06 | [SQL Interview Questions](db-design/06-sql-interview-questions.md) | 55 problems (Easy/Medium/Hard) with complete solutions |
| 07 | [NoSQL Essentials](db-design/07-nosql-essentials.md) | MongoDB, Redis, Cassandra, DynamoDB patterns & examples |

---

## API Design (Interview Prep)

> [Full API Design README](api-design/README.md)

| # | Topic | Description |
|---|-------|-------------|
| 01 | [API Design Process](api-design/01-api-design-process.md) | 5-step framework, resource modeling, endpoint design, request/response contracts |
| 02 | [REST API Deep Dive](api-design/02-rest-api-deep-dive.md) | HTTP methods, status codes, headers, caching, idempotency, HATEOAS |
| 03 | [Request & Response Patterns](api-design/03-request-response-patterns.md) | Pagination, filtering, sorting, bulk ops, async APIs, file uploads |
| 04 | [Authentication & Security](api-design/04-authentication-and-security.md) | OAuth 2.0, JWT, API keys, RBAC, CORS, rate limiting |
| 05 | [Error Handling & Versioning](api-design/05-error-handling-and-versioning.md) | RFC 7807, error taxonomy, versioning strategies, deprecation |
| 06 | [GraphQL & gRPC](api-design/06-graphql-and-grpc.md) | Schema design, resolvers, Protobuf, streaming, when to use which |
| 07 | [API Interview Questions](api-design/07-api-interview-questions.md) | 40+ problems with complete API designs |

---

## Misc

| Topic | Link |
|-------|------|
| Go Memory Leaks & pprof | [Guide](misc/memory-leaks/PPROF_GUIDE.md) |
