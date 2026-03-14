# 100-Day Engineering Concepts Curriculum

> **Purpose:** Bridge academic theory → real-world hyper-scale architecture.
> **Format:** Each concept: Definition → Problem → How It Works → Diagram → Tradeoffs → Interview Tips

---

## Module Index

| # | File | Concepts | Days |
|---|------|----------|------|
| 1 | [Scale & Data Structures](01_scale_data_structures.md) | OT, CRDTs, Gossip, Consistent Hashing, Rendezvous Hashing, Vector/Lamport Clocks, Bloom Filters, HyperLogLog, Count-Min Sketch, Merkle Trees, LSM Trees | 1–10 |
| 2 | [Database Engineering](02_database_engineering.md) | Sharding, Snowflake/UUID/ULID, WAL, Checkpointing, B/B+ Trees, SSTables, Inverted Index, Geohashing, H3, Replication, Columnar/Row Storage | 11–20 |
| 3 | [Consensus & Coordination](03_consensus_coordination.md) | CAP Theorem, Paxos, Raft, 2PC, 3PC, Leader Election, Quorum, Eventual/Strong Consistency, ZooKeeper, etcd, Distributed Locks, Lease-Based Locks | 21–28 |
| 4 | [Networking](04_networking.md) | HTTP/2, QUIC/HTTP3, gRPC, Anycast, SSE vs WebSockets, Zero-Copy, ABR, Kernel Bypass, Memory Mapping | 29–35 |
| 5 | [Architecture Patterns](05_architecture_patterns.md) | Idempotency, Event Sourcing, CQRS, Saga, Bulkhead, Circuit Breaker, Backpressure, Strangler Fig, Sidecar, Service Mesh, API Gateway, Thundering Herd, Distributed Tracing | 36–45 |
| 6 | [Rate Limiting & Load Balancing](06_rate_limiting_lb.md) | Token Bucket, Leaky Bucket, Sliding Window, Load Balancing, Consistent LB, Sticky Sessions, Service Discovery | 46–50 |
| 7 | [Caching](07_caching.md) | Cache-Aside, Write-Through, Write-Back, LRU, LFU, Clock, Redis Eviction, Cache Invalidation | 51–55 |
| 8 | [Messaging & Streams](08_messaging_streams.md) | Kafka, Exactly-Once, At-Least/Most-Once, Pub/Sub, DLQ, CDC, Lambda/Kappa, Actor Model, Disruptor | 56–62 |
| 9 | [Concurrency & Lock-Free](09_concurrency.md) | Thread Pools, Work Stealing, Event Loop, Reactor, Proactor, Lock-Free, CAS, ABA Problem, Hazard Pointers, Epoch-Based Reclamation | 63–68 |
| 10 | [Database Isolation & Failures](10_isolation_failures.md) | MVCC, Snapshot/Serializable/RC/RR Isolation, Deadlock, Starvation, Priority Inversion | 69–73 |
| 11 | [GC & Memory](11_gc_memory.md) | Reference Counting, Mark-and-Sweep, Generational GC, Stop-the-World, Incremental, Concurrent GC | 74–78 |
| 12 | [OS Primitives](12_os_primitives.md) | Virtual Memory, Paging, Segmentation, CoW, Context Switching, File Descriptors, False Sharing, NUMA, SIMD, Direct I/O | 79–85 |
| 13 | [Storage & Compression](13_storage_compression.md) | Delta Encoding, RLE, Erasure Coding, Zstd, Vectorized Execution, Roaring Bitmaps, Trie, Cuckoo Hashing | 86–90 |
| 14 | [Security & Identity](14_security.md) | OAuth/OIDC, JWT, E2E Encryption, TLS 1.3, mTLS, DDoS Mitigation, Macaroons, Password Hashing, Padding Oracle | 91–95 |
| 15 | [Reliability & Ops](15_reliability_ops.md) | Chaos Engineering, Blue/Green, Canary, C10K, CDN, Health Checks, Graceful Degradation, Shuffle Sharding | 96–98 |
| 16 | [Case Studies](16_case_studies.md) | Amazon Prime, Facebook Haystack, Twitter Fanout, WhatsApp 2M Connections, Discord GC, Dynamo, Spanner, MapReduce→Spark→Flink, CAP Proof | 99–100 |
| 17 | [AI Engineering](17_ai_engineering.md) | AI Agents, AI Coding Workflow, Context Engineering, ChatGPT Apps, LLM Concepts, MCP, Reinforcement Learning, Fine-tuning vs RAG vs Prompt vs Context | 101–108 |
| 18 | [White Papers](18_white_papers.md) | Amazon Dynamo, Google Spanner, Meta XFaaS, How to Read Systems Papers | 109–113 |

---

## Quick Reference: Concept → Company

| Concept | Famous Company |
|---------|---------------|
| Operational Transformation | Google Docs |
| CRDTs | Figma, Apple Notes |
| Gossip Protocol | Amazon DynamoDB, Uber |
| Consistent Hashing | Discord, Akamai |
| Vector Clocks | Amazon Dynamo |
| Bloom Filters | Google BigTable |
| LSM Trees | Facebook RocksDB |
| HyperLogLog | Reddit, Redis |
| Sharding | Instagram |
| Snowflake IDs | Twitter, Discord |
| WAL | Postgres, Kafka |
| MVCC | Postgres |
| Raft | etcd, Kubernetes |
| Paxos | Google Chubby |
| Zero-Copy | Kafka (LinkedIn) |
| Circuit Breaker | Netflix Hystrix |
| Event Sourcing | LMAX Exchange |
| Kafka | LinkedIn |
| Actor Model | WhatsApp (Erlang) |
| CDN | Cloudflare, Akamai |
| AI Agents (ReAct) | LangChain, AutoGPT, CrewAI |
| MCP | Anthropic, OpenAI (ChatGPT Apps) |
| RLHF | ChatGPT, Claude |
