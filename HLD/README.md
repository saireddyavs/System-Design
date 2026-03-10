# System Design Knowledge Base

**A comprehensive reference guide for Staff+ level system design interviews**

Covers all major distributed systems concepts, architectural patterns, and real-world design problems at FAANG-interview depth.

---

## Table of Contents

### 01 — Fundamentals
| File | Topics |
|------|--------|
| [Scalability](01-fundamentals/scalability.md) | Vertical vs horizontal scaling, scaling patterns, real numbers |
| [Availability & Reliability](01-fundamentals/availability-reliability.md) | SLAs, SLOs, SPOF, failover, fault tolerance, redundancy |
| [CAP Theorem & Consistency](01-fundamentals/cap-theorem-consistency.md) | CAP, PACELC, consistency models (strong, eventual, causal) |
| [Consistent Hashing](01-fundamentals/consistent-hashing.md) | Hash rings, virtual nodes, rebalancing, implementations |
| [Back-of-Envelope Estimation](01-fundamentals/back-of-envelope.md) | Latency numbers, throughput calculations, storage estimates |

### 02 — Networking
| File | Topics |
|------|--------|
| [DNS](02-networking/dns.md) | Resolution, record types, caching, GeoDNS, failures |
| [TCP vs UDP](02-networking/tcp-udp.md) | Handshakes, congestion control, when to use each |
| [HTTP Protocols](02-networking/http-protocols.md) | HTTP/1.1, HTTP/2, HTTP/3 (QUIC), TLS, keep-alive |
| [Load Balancing](02-networking/load-balancing.md) | L4 vs L7, algorithms, health checks, global load balancing |
| [CDN](02-networking/cdn.md) | Push vs pull, edge caching, cache invalidation, providers |
| [Proxies](02-networking/proxies.md) | Forward vs reverse proxy, sidecar proxy, service mesh |

### 03 — API Design
| File | Topics |
|------|--------|
| [API Design Principles](03-apis/api-design.md) | Pagination, versioning, error handling, backward compatibility |
| [REST vs GraphQL vs gRPC](03-apis/rest-graphql-grpc.md) | Tradeoffs, when to use each, protocol buffers |
| [API Gateway](03-apis/api-gateway.md) | Routing, auth, rate limiting, request aggregation |
| [WebSockets & Webhooks](03-apis/websockets-webhooks.md) | Long polling, SSE, bidirectional communication |
| [Rate Limiting](03-apis/rate-limiting.md) | Token bucket, sliding window, distributed rate limiting |
| [Idempotency](03-apis/idempotency.md) | Idempotency keys, at-least-once delivery, deduplication |

### 04 — Databases
| File | Topics |
|------|--------|
| [SQL vs NoSQL](04-databases/sql-vs-nosql.md) | Relational vs document vs wide-column vs graph |
| [ACID Transactions](04-databases/acid-transactions.md) | Isolation levels, 2PC, saga pattern |
| [Indexes & Storage Engines](04-databases/indexes-storage-engines.md) | B+ trees, LSM trees, hash indexes, SST files |
| [Sharding](04-databases/sharding.md) | Strategies, hotspots, resharding, cross-shard queries |
| [Replication](04-databases/replication.md) | Leader-follower, multi-leader, leaderless, conflict resolution |
| [Database Types & Selection](04-databases/database-types-selection.md) | Choosing the right database, comparison matrix |
| [Bloom Filters](04-databases/bloom-filters.md) | Probabilistic data structures, false positives, use cases |

### 05 — Caching
| File | Topics |
|------|--------|
| [Caching Strategies](05-caching/caching-strategies.md) | Cache-aside, write-through, write-back, read-through |
| [Cache Eviction](05-caching/cache-eviction.md) | LRU, LFU, TTL, ARC, comparison and pseudocode |
| [Distributed Caching](05-caching/distributed-caching.md) | Redis, Memcached, cache coherence, thundering herd |

### 06 — Messaging & Streaming
| File | Topics |
|------|--------|
| [Message Queues](06-messaging/message-queues.md) | RabbitMQ, SQS, at-least-once/exactly-once delivery |
| [Pub/Sub](06-messaging/pub-sub.md) | Fan-out, topic-based routing, ordering guarantees |
| [Kafka & Event Streaming](06-messaging/kafka-event-streaming.md) | Partitions, consumer groups, log compaction, exactly-once |
| [Change Data Capture](06-messaging/change-data-capture.md) | Debezium, outbox pattern, dual-write problems |

### 07 — Distributed Systems
| File | Topics |
|------|--------|
| [Consistency Models](07-distributed-systems/consistency-models.md) | Linearizability, sequential, causal, eventual consistency |
| [Consensus Algorithms](07-distributed-systems/consensus-algorithms.md) | Paxos, Raft, ZAB, practical applications |
| [Leader Election](07-distributed-systems/leader-election.md) | Bully, ring, ZooKeeper-based, fencing tokens |
| [Distributed Locking](07-distributed-systems/distributed-locking.md) | Redlock, ZooKeeper locks, fencing, correctness |
| [Gossip Protocol](07-distributed-systems/gossip-protocol.md) | Epidemic protocols, failure detection, CRDT |
| [Clocks & Ordering](07-distributed-systems/clocks-ordering.md) | Lamport clocks, vector clocks, hybrid logical clocks |

### 08 — Architectural Patterns
| File | Topics |
|------|--------|
| [Microservices](08-architecture/microservices.md) | Decomposition, communication, data ownership |
| [Event-Driven Architecture](08-architecture/event-driven-architecture.md) | CQRS, event sourcing, saga pattern |
| [Serverless](08-architecture/serverless.md) | FaaS, cold starts, scaling, cost model |
| [Service Mesh](08-architecture/service-mesh.md) | Istio, Envoy, sidecar pattern, mTLS |
| [Circuit Breaker](08-architecture/circuit-breaker.md) | States, bulkhead, timeout patterns |
| [Service Discovery](08-architecture/service-discovery.md) | Client-side, server-side, DNS-based, Consul/etcd |

### 09 — Reliability
| File | Topics |
|------|--------|
| [Fault Tolerance & Failover](09-reliability/fault-tolerance-failover.md) | Active-passive, active-active, health checks |
| [Retries & Backoff](09-reliability/retries-backoff.md) | Exponential backoff, jitter, retry budgets, dead letter queues |
| [Disaster Recovery](09-reliability/disaster-recovery.md) | RPO, RTO, multi-region, chaos engineering |

### 10 — Security
| File | Topics |
|------|--------|
| [Authentication & Authorization](10-security/authentication-authorization.md) | OAuth2, JWT, RBAC, ABAC, SSO |
| [Encryption & TLS](10-security/encryption-tls.md) | Symmetric/asymmetric, TLS handshake, mTLS |
| [Secrets Management](10-security/secrets-management.md) | Vault, KMS, rotation, environment variables |

### 11 — Observability
| File | Topics |
|------|--------|
| [Logging, Metrics & Tracing](11-observability/logging-metrics-tracing.md) | ELK, Prometheus, Jaeger, OpenTelemetry, SLIs/SLOs |

### 12 — Storage & Data Systems
| File | Topics |
|------|--------|
| [Storage Types](12-storage-and-data/storage-types.md) | Object, block, file storage, HDFS, S3 |
| [OLTP vs OLAP](12-storage-and-data/oltp-vs-olap.md) | Row vs columnar, data warehousing, ETL |
| [Data Lakes & Streaming](12-storage-and-data/data-lakes-streaming.md) | Lambda/Kappa architecture, batch vs stream |

### 13 — Tradeoffs
| File | Topics |
|------|--------|
| [System Design Tradeoffs](13-tradeoffs/system-design-tradeoffs.md) | All major tradeoff comparisons with tables |

### 14 — Numbers & Estimation
| File | Topics |
|------|--------|
| [Latency Numbers](14-numbers-and-estimation/latency-numbers.md) | Every engineer should know, hardware to network |
| [Scale Reference](14-numbers-and-estimation/scale-reference.md) | Twitter, YouTube, Uber, WhatsApp scale numbers |
| [Estimation Framework](14-numbers-and-estimation/estimation-framework.md) | Step-by-step estimation with worked examples |

### 15 — System Design Interview Problems

#### Easy
| File | Problem |
|------|---------|
| [URL Shortener](15-interview-problems/easy/url-shortener.md) | Design TinyURL |
| [Pastebin](15-interview-problems/easy/pastebin.md) | Design Pastebin |
| [Distributed Cache](15-interview-problems/easy/distributed-cache.md) | Design a distributed caching system |
| [Rate Limiter](15-interview-problems/easy/rate-limiter.md) | Design a rate limiter |
| [Key-Value Store](15-interview-problems/easy/key-value-store.md) | Design a distributed key-value store |

#### Medium
| File | Problem |
|------|---------|
| [Twitter](15-interview-problems/medium/twitter.md) | Design Twitter / X |
| [Instagram](15-interview-problems/medium/instagram.md) | Design Instagram |
| [WhatsApp](15-interview-problems/medium/whatsapp.md) | Design WhatsApp |
| [Notification Service](15-interview-problems/medium/notification-service.md) | Design a notification system |
| [YouTube](15-interview-problems/medium/youtube.md) | Design YouTube |
| [Netflix](15-interview-problems/medium/netflix.md) | Design Netflix |
| [Spotify](15-interview-problems/medium/spotify.md) | Design Spotify |
| [Payment System](15-interview-problems/medium/payment-system.md) | Design a payment system |
| [Job Scheduler](15-interview-problems/medium/job-scheduler.md) | Design a distributed job scheduler |
| [Tinder](15-interview-problems/medium/tinder.md) | Design Tinder |

#### Hard
| File | Problem |
|------|---------|
| [Uber](15-interview-problems/hard/uber.md) | Design Uber / ride-sharing |
| [Google Docs](15-interview-problems/hard/google-docs.md) | Design collaborative editing |
| [Google Maps](15-interview-problems/hard/google-maps.md) | Design Google Maps |
| [Distributed Web Crawler](15-interview-problems/hard/web-crawler.md) | Design a web crawler |
| [Dropbox](15-interview-problems/hard/dropbox.md) | Design Dropbox / file sync |
| [Zoom](15-interview-problems/hard/zoom.md) | Design Zoom / video conferencing |
| [Ticket Booking](15-interview-problems/hard/ticket-booking.md) | Design BookMyShow / event booking |

---

## How to Use This Guide

1. **Start with Fundamentals** — Read sections 01 and 14 first to build your foundation
2. **Deep dive by topic** — Work through sections 02–13 based on weak areas
3. **Practice problems** — Attempt section 15 problems, starting easy and progressing to hard
4. **Review tradeoffs** — Section 13 is critical for interview discussions
5. **Memorize key numbers** — Section 14 gives you the ammunition for estimation questions

## Key References

- [Designing Data-Intensive Applications](https://dataintensive.net/) — Martin Kleppmann
- [System Design Interview](https://www.amazon.com/System-Design-Interview-insiders-Second/dp/B08CMF2CQF) — Alex Xu
- [awesome-system-design-resources](https://github.com/ashishps1/awesome-system-design-resources) — Ashish Pratap Singh
- [The Google SRE Book](https://sre.google/sre-book/table-of-contents/)
