# High-Throughput System Case Studies

---

## 1. How PayPal Supports 1B Transactions/Day with 8 VMs

### The Problem
PayPal processes billions of transactions daily. Traditional approach: 1000+ VMs with low throughput each. Need to consolidate, reduce cost, and handle bursty workloads with back-pressure.

### The Architecture/Solution
**Actor Model (Akka)**: Each transaction as an actor; async, non-blocking. **Akka Streams**: Back-pressure propagates upstream; no overload. **Kafka**: Decouple producers/consumers; absorb bursts. **squbs**: PayPal's framework (Akka + Akka HTTP) for production deployments. **JVM Tuning**: Optimize for throughput; reduce GC pauses.

### Key Numbers
| Metric | Value |
|--------|-------|
| Transactions/day | 1 billion |
| VMs | 8 (vs. 1000+ before) |
| Framework | squbs (Akka) |
| Back-pressure | Akka Streams + Kafka |

### Architecture Diagram (ASCII)
```
    Clients → API → Akka Actors → Kafka → Workers
                    (back-pressure)
```

### Key Takeaways
- **Actor model** enables high concurrency with minimal threads
- **Back-pressure** prevents overload; Kafka buffers bursts
- **Reactive** = async + back-pressure + error composition

### Interview Tip
"PayPal uses Akka for 1B transactions on 8 VMs. Actor model + back-pressure (Akka Streams) + Kafka for decoupling. Key: don't let slow consumers block producers; propagate back-pressure."

---

## 2. How Shopify Handles Flash Sales at 32M Requests/Minute

### The Problem
Flash sales (e.g., Kylie Cosmetics) generate millions of requests in seconds. Can crash DB shards. Need to prevent overselling and protect backend.

### The Architecture/Solution
**Redis + Lua**: Atomic check-and-decrement for inventory; no race conditions. **Nginx + Lua (OpenResty)**: Traffic management at edge; PID controllers for fair queuing. **Signed cookies**: Stateless load balancers; timestamp for fairness. **Pod autoscaler**: Scale app tier for sustained load.

### Key Numbers
| Metric | Value |
|--------|-------|
| Peak RPS | 32M requests/min |
| Lua | Atomic Redis ops |
| Edge | Nginx + OpenResty |

### Architecture Diagram (ASCII)
```
    Clients → Nginx+Lua (queue) → App → Redis Lua (inventory) → DB
```

### Key Takeaways
- **Lua in Redis** = atomic; no round-trips
- **Edge queuing** protects backend
- **Fair queuing** via signed cookies

### Interview Tip
"For flash sales, use Redis Lua scripts for atomic inventory decrement. Lua runs atomically in Redis—no race conditions. Queue at edge (Nginx+Lua) to protect backend."

---

## 3. Virtual Waiting Room Architecture at SeatGeek

### The Problem
Ticket drops (e.g., Taylor Swift) attract millions of users. All hit at once; servers crash. Need fair access without overselling.

### The Architecture/Solution
**Queue-fair**: Assign position in virtual queue; serve in order. **Token bucket**: Rate limit entry into queue. **Hold page**: Users wait; poll for turn. **Backend**: Only serves users who "won" the queue.

### Key Numbers
| Metric | Value |
|--------|-------|
| Pattern | Virtual waiting room |
| Fairness | Queue position |
| Rate limit | Token bucket |

### Architecture Diagram (ASCII)
```
    Users → Token Bucket → Queue (position) → Hold Page → Backend (when turn)
```

### Key Takeaways
- **Queue-fair** prevents thundering herd
- **Token bucket** limits entry rate
- **Hold page** reduces active connections

### Interview Tip
"For ticket drops, use a virtual waiting room: token bucket at entry, queue with position, hold page. Only users whose turn it is hit the backend. SeatGeek, Ticketmaster use this."

---

## 4. How Razorpay Handles Flash Sales at 1500 RPS

### The Problem
Payment gateway during flash sales: 1500+ RPS. Must not double-charge; must not lose payments. Idempotency and rate limiting critical.

### The Architecture/Solution
**Idempotent APIs**: Client sends idempotency key; duplicate requests return same result. **Rate limiting**: Per-merchant, per-endpoint. **Queue**: Buffer spikes; process at sustainable rate. **Idempotency store**: Redis/DB; key → response cache.

### Key Numbers
| Metric | Value |
|--------|-------|
| RPS | 1500+ |
| Idempotency | Key-based |
| Rate limit | Per merchant |

### Key Takeaways
- **Idempotency** prevents double charge
- **Rate limiting** protects backend
- **Queue** for spike absorption

### Interview Tip
"For payment APIs, idempotency is mandatory. Client sends idempotency key; server caches response. Duplicate request = return cached. Razorpay, Stripe use this."

---

## 5. How Tinder Scaled to 1.6B Swipes/Day

### The Problem
Tinder: swipe left/right. Billions of swipes daily. Need low latency, geo-based matching (nearby users), and scale.

### The Architecture/Solution
**Geosharding**: Shard by geographic region; users in same region on same shard. **Kubernetes**: Auto-scaling; container orchestration. **Caching**: Hot profiles in Redis. **Async**: Swipe write async; match logic async.

### Key Numbers
| Metric | Value |
|--------|-------|
| Swipes/day | 1.6 billion |
| Sharding | Geo-based |
| Orchestration | Kubernetes |

### Key Takeaways
- **Geosharding** fits location-based apps
- **Kubernetes** for scaling
- **Async** for non-blocking UX

### Interview Tip
"For location-based apps like Tinder, geosharding: users in same region on same shard. Reduces cross-shard queries for 'nearby'."

---

## 6. How McDonald's Food Delivery Platform Handles 20K Orders/Second

### The Problem
McDonald's delivery: 20K orders/sec peak. Order placement, kitchen display, driver assignment, real-time tracking. Need reliability and low latency.

### The Architecture/Solution
**Event-driven**: Order events to Kafka; services consume. **Microservices**: Order, kitchen, dispatch, payment. **CQRS**: Write-optimized; read from materialized views. **Caching**: Menu, store info cached.

### Key Numbers
| Metric | Value |
|--------|-------|
| Orders/sec | 20K |
| Pattern | Event-driven |
| Messaging | Kafka |

### Key Takeaways
- **Event-driven** decouples services
- **CQRS** separates write/read paths
- **Kafka** for order event stream

### Interview Tip
"For high-order throughput, event-driven + Kafka. Order service publishes; kitchen, dispatch, payment consume. CQRS for read scaling."

---

## 7. How Uber Computes ETA at 500K RPS

### The Problem
Uber needs ETA for every trip phase: eyeball, dispatch, pickup, on-trip. 500K ETA requests/sec. Routing is expensive; can't run Dijkstra per request.

### The Architecture/Solution
**Graph partitioning**: Precompute paths within partitions; 500K ops → 700 for Bay Area. **ML (DeepETA)**: Routing engine gives base ETA; ML predicts residual error. **Caching**: Same OD pair → cache. **Traffic data**: Real-time; fed into model.

### Key Numbers
| Metric | Value |
|--------|-------|
| ETA RPS | 500K |
| Requests/trip | ~1000 |
| Approach | Precomputed + ML |

### Architecture Diagram (ASCII)
```
    Request → Graph Partition Lookup → Base ETA → ML (DeepETA) → Final ETA
```

### Key Takeaways
- **Precompute** paths; don't run Dijkstra live
- **ML** for residual; routing for structure
- **Partitioning** reduces complexity

### Interview Tip
"Uber ETA: precompute paths within graph partitions (not full Dijkstra). ML (DeepETA) corrects residual. 500K RPS requires precomputation + caching."

---

## 8. How Uber Finds Nearby Drivers at 1M RPS

### The Problem
Match riders to nearby drivers. 1M RPS for "find drivers near me." Geospatial queries at scale.

### The Architecture/Solution
**S2 cells (Google)**: Hierarchical grid; encode lat/lon to cell ID. **Geohash**: Alternative; divide world into grid. **Index**: Drivers in cell; query = get cells in radius, merge. **Caching**: Hot areas cached.

### Key Numbers
| Metric | Value |
|--------|-------|
| RPS | 1M |
| Index | S2/Geohash |
| Query | Range on cell IDs |

### Key Takeaways
- **S2/Geohash** for geospatial indexing
- **Cell-based** = efficient range queries
- **Cache** hot areas

### Interview Tip
"For 'find nearby', use S2 cells or Geohash. Encode location to cell; index drivers by cell. Query = cells in radius. Uber, Lyft use this."

---

## 9. How Uber Payment System Handles 30M Transactions/Day

### The Problem
Uber: rides, Eats, freight. 30M payment transactions/day. Must be reliable, idempotent, and support retries.

### The Architecture/Solution
**Idempotency**: Key per transaction. **Saga pattern**: Distributed tx across ride, payment, notification; compensate on failure. **Eventual consistency**: Payment status async. **Retries**: Exponential backoff; idempotent.

### Key Numbers
| Metric | Value |
|--------|-------|
| Transactions/day | 30M |
| Pattern | Saga |
| Idempotency | Yes |

### Key Takeaways
- **Saga** for distributed transactions
- **Idempotency** for retries
- **Compensation** on failure

### Interview Tip
"For distributed payments, use Saga: each step has a compensation. Payment fails → compensate ride. Idempotency for retries."

---

## 10. How Giphy Delivers 10B GIFs/Day to 1B Users

### The Problem
Giphy serves 10B GIFs/day. Search, trending, random. Need low latency and global reach.

### The Architecture/Solution
**CDN**: 99%+ from edge; origin for cache miss. **Caching**: Multi-layer; edge, origin. **Search**: Elasticsearch; index metadata. **Storage**: S3/GCS for GIFs; CDN in front.

### Key Numbers
| Metric | Value |
|--------|-------|
| GIFs/day | 10B |
| Users | 1B |
| CDN | Primary delivery |

### Key Takeaways
- **CDN** absorbs almost all traffic
- **Search** on metadata; not blob
- **Cache** at every layer

### Interview Tip
"Giphy: CDN for delivery, Elasticsearch for search. 10B/day = CDN is mandatory. Origin only for cache miss."

---

## 11. How Facebook Supports 1B Users via Software Load Balancer (Katran)

### The Problem
Facebook's scale: hardware LBs are expensive and inflexible. Need software-based LB that handles 1B users.

### The Architecture/Solution
**Katran**: Open-source L4 LB. **XDP/eBPF**: Kernel bypass; process packets in kernel before full stack. **High throughput**: Millions of connections. **Consistent hashing**: Minimal connection churn on backend change.

### Key Numbers
| Metric | Value |
|--------|-------|
| Users | 1B |
| Tech | XDP, eBPF |
| LB | Katran (software) |

### Key Takeaways
- **eBPF/XDP** = kernel-level packet processing
- **Software LB** = flexibility, cost
- **Katran** = Facebook's solution

### Interview Tip
"Facebook's Katran uses XDP/eBPF for software L4 load balancing. Kernel bypass = high throughput. Alternative to hardware LB."

---

## 12. How WhatsApp Supports 50B Messages/Day with 32 Engineers

### The Problem
50B messages/day. Minimal team. Need simplicity, reliability, low cost.

### The Architecture/Solution
**Erlang**: Lightweight processes; 2M connections/server. **Simplicity**: No unnecessary features. **FreeBSD**: Tuned kernel; high connection limits. **Minimalist protocol**: Small binary format.

### Key Numbers
| Metric | Value |
|--------|-------|
| Messages/day | 50B |
| Engineers | 32 |
| Stack | Erlang, FreeBSD |

### Key Takeaways
- **Erlang** = connection density
- **Simplicity** = fewer bugs, less code
- **Right tool** for the job

### Interview Tip
"WhatsApp: Erlang for 2M connections/server, 32 engineers for 50B messages/day. Simplicity and Erlang's concurrency model are the keys."

---

## 13. 11 Reasons YouTube Supports 100M Video Views/Day with 9 Engineers

### The Problem
100M views/day. Small team. Need to leverage existing infra (Google) and focus on product.

### The Architecture/Solution
**Google infra**: Colossus (storage), Borg (orchestration), Bigtable. **CDN**: Google's global network. **Minimal custom**: Use managed services. **Focus**: Product, not reinventing infra.

### Key Numbers
| Metric | Value |
|--------|-------|
| Views/day | 100M |
| Engineers | 9 |
| Infra | Google (Colossus, Borg) |

### Key Takeaways
- **Leverage platform**; don't build everything
- **Managed services** reduce ops
- **Small team** = focus on product

### Interview Tip
"YouTube scales with 9 engineers by using Google's infra: Colossus, Borg, Bigtable. Key: leverage platform, don't reinvent."

---

## 14. 5 Reasons Zoom Supports 300M Video Calls/Day

### The Problem
300M video calls/day. Real-time, low latency. Need to scale media routing.

### The Architecture/Solution
**SFU (Selective Forwarding Unit)**: Server receives streams, forwards to participants; no mixing. **WebRTC**: Browser standard. **MCU fallback**: For low-end devices; mix streams. **Edge**: Media servers at edge; reduce latency.

### Key Numbers
| Metric | Value |
|--------|-------|
| Calls/day | 300M |
| Architecture | SFU |
| Protocol | WebRTC |

### Key Takeaways
- **SFU** = scalable; no decode/encode at server
- **WebRTC** = standard
- **Edge** for latency

### Interview Tip
"Zoom uses SFU: server forwards streams, doesn't mix. Lower server load than MCU. WebRTC for browser support."

---

## 15. How Zapier Automates Billions of Tasks

### The Problem
Zapier: connect apps (e.g., Gmail → Slack). Billions of tasks. Need reliability, retries, and scale.

### The Architecture/Solution
**Event-driven**: Trigger → Task → Action. **Worker pools**: Scale workers for each app. **Retries**: Exponential backoff; dead letter. **Idempotency**: Task ID; no duplicate execution.

### Key Numbers
| Metric | Value |
|--------|-------|
| Tasks | Billions |
| Pattern | Event-driven |
| Workers | Per integration |

### Key Takeaways
- **Workers** per app/integration
- **Retries** + dead letter
- **Idempotency** for reliability

### Interview Tip
"Zapier: event-driven, worker pools per integration. Retries with backoff. Idempotency for task execution."

---

## 16. How LinkedIn Scaled to 930M Users

### The Problem
LinkedIn: profiles, feed, connections, jobs. 930M users. Need to scale data layer and real-time features.

### The Architecture/Solution
**Espresso**: NoSQL document store; low latency. **Kafka**: Event streaming; activity, notifications. **Samza**: Stream processing; feed, recommendations. **Graph**: Connections; custom storage.

### Key Numbers
| Metric | Value |
|--------|-------|
| Users | 930M |
| Storage | Espresso |
| Streaming | Kafka, Samza |

### Key Takeaways
- **Espresso** = document store at scale
- **Kafka** = event backbone
- **Samza** = stream processing

### Interview Tip
"LinkedIn: Espresso for storage, Kafka for events, Samza for stream processing. Event-driven architecture at scale."

---

## 17. LinkedIn Adopted Protocol Buffers to Reduce Latency by 60%

### The Problem
JSON serialization/deserialization was a bottleneck. Need faster, smaller payloads.

### The Architecture/Solution
**Protocol Buffers**: Binary format; smaller than JSON. **60% latency reduction**: Less CPU, less bandwidth. **Schema evolution**: Backward compatible. **gRPC**: Uses Protobuf; RPC with binary.

### Key Numbers
| Metric | Value |
|--------|-------|
| Latency reduction | 60% |
| Format | Protobuf |
| Use case | RPC, storage |

### Key Takeaways
- **Protobuf** = smaller, faster than JSON
- **Binary** = less parsing
- **Schema** = type safety, evolution

### Interview Tip
"Protobuf reduces latency vs JSON: binary, smaller payload, less CPU. LinkedIn saw 60% improvement. Use for high-throughput RPC."

---

## 18. How Lyft Supports Rides for 21M Users

### The Problem
Lyft: matching, ETA, payments. 21M users. Need reliability and observability.

### The Architecture/Solution
**Service mesh (Envoy)**: Traffic management, retries, circuit breaking. **Microservices**: Matching, ETA, payment. **Observability**: Tracing, metrics. **Multi-region**: Failover.

### Key Numbers
| Metric | Value |
|--------|-------|
| Users | 21M |
| Mesh | Envoy |
| Pattern | Microservices |

### Key Takeaways
- **Envoy** = service mesh
- **Observability** built-in
- **Retries, circuit breaking** at mesh

### Interview Tip
"Lyft uses Envoy service mesh: retries, circuit breaking, tracing. Offload resilience to infrastructure."

---

## 19. How Hashnode Generates Feed at Scale

### The Problem
Hashnode: developer blog platform. Feed = posts from followed users/tags. Fan-out at scale.

### The Architecture/Solution
**Fan-out on write**: Precompute feed when user follows/followed. **Aggregation**: Merge multiple sources (followed users, tags). **Caching**: Feed in Redis. **Pagination**: Cursor-based.

### Key Numbers
| Metric | Value |
|--------|-------|
| Pattern | Fan-out on write |
| Cache | Redis |
| Merge | Aggregation service |

### Key Takeaways
- **Fan-out on write** for precomputed feed
- **Aggregation** for multi-source
- **Cache** for read path

### Interview Tip
"For feed generation, fan-out on write: when user posts, push to followers' feeds. Cache in Redis. Hashnode, Twitter use variants."

---

## 20. How Halo Scaled to 11.6M Users Using Saga Pattern

### The Problem
Halo (gaming): purchases, entitlements, matchmaking. Distributed transactions across services.

### The Architecture/Solution
**Saga pattern**: Choreography or orchestration. Each step has compensation. **Purchase flow**: Payment → Entitlement → Notification. **Compensation**: Payment fails → rollback entitlement. **Idempotency**: Per step.

### Key Numbers
| Metric | Value |
|--------|-------|
| Users | 11.6M |
| Pattern | Saga |
| Use case | Purchases |

### Key Takeaways
- **Saga** for distributed transactions
- **Compensation** on failure
- **No 2PC**; eventual consistency

### Interview Tip
"Saga: sequence of local transactions with compensations. No 2PC. Halo uses it for purchases. Each step reversible."

---

## 21. How Discord Boosts Performance With Code-Splitting

### The Problem
Discord web app: large bundle; slow load. Need faster initial load.

### The Architecture/Solution
**Code-splitting**: Split by route; load on demand. **Lazy loading**: Dynamic import. **Webpack**: Chunk splitting. **Smaller initial bundle**: Faster TTI.

### Key Numbers
| Metric | Value |
|--------|-------|
| Approach | Code-splitting |
| Tool | Webpack |
| Benefit | Faster load |

### Key Takeaways
- **Code-split** by route
- **Lazy load** non-critical
- **Smaller initial** = faster

### Interview Tip
"Discord uses code-splitting: load only what's needed for current route. Webpack chunks. Reduces initial bundle size."

---

## 22. Stripe Rate Limiting for Scalable APIs

### The Problem
Stripe: payment API. Must prevent abuse, ensure fair usage. Rate limiting at scale.

### The Architecture/Solution
**Token bucket**: Allow burst; sustain rate. **Leaky bucket**: Alternative; smooth output. **Per-key**: API key, user, endpoint. **Headers**: Return remaining, reset time.

### Key Numbers
| Metric | Value |
|--------|-------|
| Algorithm | Token bucket |
| Scope | Per key, endpoint |
| Headers | X-RateLimit-* |

### Key Takeaways
- **Token bucket** = burst + sustain
- **Per-key** = fairness
- **Headers** = client visibility

### Interview Tip
"Stripe rate limiting: token bucket per API key. Allows burst, enforces sustained rate. Return remaining in headers."

---

## 23. How Stripe Prevents Double Payment Using Idempotent API

### The Problem
Network retries can cause duplicate charges. Must ensure exactly-once semantics.

### The Architecture/Solution
**Idempotency key**: Client sends with request. **Server**: First request with key → process, cache response. **Duplicate**: Same key → return cached, no reprocess. **TTL**: 24 hours typical.

### Key Numbers
| Metric | Value |
|--------|-------|
| Key | Client-provided |
| Storage | Redis/DB |
| TTL | 24h |

### Key Takeaways
- **Idempotency key** = exactly-once
- **Cache response** for duplicates
- **Mandatory** for payments

### Interview Tip
"Stripe idempotency: client sends Idempotency-Key. First request processes; duplicates return cached response. Essential for payments."

---

## 24. 6 Proven Guidelines on Open Sourcing From Tumblr

### The Problem
Tumblr open-sourced tools. How to do it well: licensing, maintenance, community.

### The Architecture/Solution
**Clear license**: MIT, Apache. **Documentation**: README, examples. **Scope**: Single-purpose tools. **Maintenance**: Assign owners. **Community**: Contributing guide. **Quality**: CI, tests.

### Key Takeaways
- **License** early
- **Docs** matter
- **Ownership** for maintenance

### Interview Tip
"Open sourcing: clear license, good docs, single purpose. Tumblr's guidelines: scope small, maintain, document."

---

## 25. Airbnb HTTP Streaming ($84M Savings)

### The Problem
Airbnb: large API responses; slow. Need to improve perceived performance and reduce cost.

### The Architecture/Solution
**HTTP streaming**: Stream response chunks as ready. **Progressive loading**: Send critical data first. **Reduced TTFB**: Don't wait for full response. **$84M savings**: Less compute, faster UX.

### Key Numbers
| Metric | Value |
|--------|-------|
| Savings | $84M |
| Technique | HTTP streaming |
| Benefit | Faster TTFB |

### Key Takeaways
- **Stream** don't buffer
- **Progressive** = better UX
- **Cost** and latency win

### Interview Tip
"Airbnb saved $84M with HTTP streaming: send chunks as ready, don't wait for full response. Reduces TTFB and compute."

---

## 26. How Shopify Handled 30TB/Minute With Modular Monolith

### The Problem
Shopify: high data volume. Microservices added complexity. Need balance: modularity without distributed system overhead.

### The Architecture/Solution
**Modular monolith**: Single deployable; logical modules. **Bounded contexts**: Clear boundaries. **Shared DB**: For now; can split later. **30TB/min**: Handled with good modular design. **Escape hatch**: Extract to microservice when needed.

### Key Numbers
| Metric | Value |
|--------|-------|
| Data | 30TB/min |
| Architecture | Modular monolith |
| Benefit | Simplicity + modularity |

### Key Takeaways
- **Modular monolith** = structure without distribution
- **Extract** when module needs independent scale
- **Not everything** needs microservices

### Interview Tip
"Shopify uses modular monolith: single deployable, logical modules. Avoid microservices until you need independent scaling. 30TB/min shows it scales."
