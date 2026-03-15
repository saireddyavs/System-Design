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

| # | Problem | Patterns | Link |
|---|---------|----------|------|
| 01 | Parking Lot | Strategy, DIP | [README](LLD/01-parking-lot-system/README.md) |
| 02 | Online Bookstore | Strategy, Observer, Factory | [README](LLD/02-online-bookstore/README.md) |
| 03 | Library Management | Observer, Strategy, Facade | [README](LLD/03-library-management-system/README.md) |
| 04 | Movie Ticket Booking | Strategy, Observer, Builder | [README](LLD/04-movie-ticket-booking/README.md) |
| 05 | Elevator System | Strategy, State | [README](LLD/05-elevator-system/README.md) |
| 06 | Hotel Management | Strategy, Observer, Factory | [README](LLD/06-hotel-management-system/README.md) |
| 07 | Ride Sharing (Uber) | Strategy, Observer | [README](LLD/07-ride-sharing-service/README.md) |
| 08 | File Storage (Dropbox) | Composite, Observer | [README](LLD/08-file-storage-system/README.md) |
| 09 | Chat Application | Pub/Sub, Strategy, Factory | [README](LLD/09-chat-application/README.md) |
| 10 | Social Media | Strategy, Observer | [README](LLD/10-social-media-platform/README.md) |
| 11 | Notification System | Strategy, Decorator, Chain | [README](LLD/11-notification-system/README.md) |
| 12 | Airline Reservation | Strategy, Observer, Builder | [README](LLD/12-airline-reservation-system/README.md) |
| 13 | ATM System | Command, State, CoR | [README](LLD/13-atm-system/README.md) |
| 14 | E-Commerce | Strategy, Observer, Factory | [README](LLD/14-ecommerce-website/README.md) |
| 15 | Food Delivery | Strategy, Observer | [README](LLD/15-food-delivery-system/README.md) |
| 16 | Shopping Cart | Strategy, Observer | [README](LLD/16-shopping-cart-system/README.md) |
| 17 | Splitwise | Strategy, Builder | [README](LLD/17-splitwise/README.md) |

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

## Misc

| Topic | Link |
|-------|------|
| Go Memory Leaks & pprof | [Guide](misc/memory-leaks/PPROF_GUIDE.md) |
