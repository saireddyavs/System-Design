# Scaling Case Studies

---

## 1. How to Scale an App to 10 Million Users on AWS

### The Problem
A typical web app starts on a single server. As users grow to millions, the single server becomes a bottleneck: CPU, memory, disk I/O, and network bandwidth all saturate. Database connections exhaust, and a single point of failure risks total outage.

### The Architecture/Solution

**Tier 1: Load Balancer**
- AWS Application Load Balancer (ALB) or Network Load Balancer (NLB) distributes traffic across multiple app servers
- Health checks remove unhealthy instances from rotation
- SSL termination at LB reduces backend CPU load

**Tier 2: Auto-Scaling Compute**
- EC2 Auto Scaling Groups scale app servers horizontally based on CPU, request count, or custom metrics
- Stateless app design: any request can hit any server; session stored in Redis/cookie
- Multiple Availability Zones for high availability

**Tier 3: Database**
- RDS with Multi-AZ for automatic failover
- Read replicas for read-heavy workloads (e.g., 1 write + 3 read replicas)
- Connection pooling (PgBouncer/RDS Proxy) to manage connection limits

**Tier 4: Caching**
- ElastiCache (Redis/Memcached) for session, hot data, query results
- Cache-aside pattern: check cache first, miss → DB → populate cache

**Tier 5: CDN**
- CloudFront for static assets (JS, CSS, images) and optionally dynamic content
- Edge locations reduce latency and offload origin

### Key Numbers
| Component | Typical Config |
|-----------|----------------|
| App servers | 10–50 EC2 instances (auto-scaled) |
| RDS | db.r5.xlarge+ with read replicas |
| ElastiCache | Redis cluster, 3+ nodes |
| CloudFront | 200+ edge locations globally |

### Architecture Diagram (ASCII)

```
                    ┌─────────────────┐
                    │   CloudFront     │
                    │   (CDN)          │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │  ALB (LB)       │
                    └────────┬────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
    ┌────▼────┐         ┌────▼────┐         ┌────▼────┐
    │  EC2    │         │  EC2    │         │  EC2    │
    │  App    │         │  App    │         │  App    │
    └────┬────┘         └────┬────┘         └────┬────┘
         │                   │                   │
         └───────────────────┼───────────────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
         ┌────▼────┐    ┌────▼────┐    ┌────▼────┐
         │ElastiCache│   │   RDS   │    │   S3    │
         │ (Redis)  │   │ Primary │    │ (files) │
         └─────────┘    └────┬────┘    └─────────┘
                            │
                       ┌────▼────┐
                       │   RDS   │
                       │ Replica │
                       └─────────┘
```

### Key Takeaways
- **Stateless apps** enable horizontal scaling; store session in Redis
- **Read replicas** offload read traffic; use for reporting, search
- **Caching** reduces DB load by 80–90% for read-heavy apps
- **CDN** cuts latency and bandwidth costs for static content

### Interview Tip
"When scaling to millions, I'd add a load balancer, auto-scale app servers, add read replicas for the DB, introduce Redis for caching, and put CloudFront in front for static assets. The key is making the app stateless first."

---

## 2. How to Scale an App to 100 Million Users on GCP

### The Problem
At 100M users, single-region databases and simple replication models break down. Cross-region latency, consistency, and operational complexity become critical. You need globally distributed, strongly consistent storage and event-driven architectures.

### The Architecture/Solution

**Cloud Spanner**
- Globally distributed, strongly consistent SQL database
- No manual sharding; automatic replication across regions
- TrueTime-based timestamps for external consistency

**Pub/Sub**
- Decouple producers and consumers; handle millions of messages/sec
- At-least-once delivery; use idempotent consumers
- Fan-out to multiple subscribers (analytics, notifications, search indexing)

**GKE (Kubernetes)**
- Container orchestration with auto-scaling (HPA, VPA, cluster autoscaler)
- Multi-region deployments with GKE Multi-Cluster
- Service mesh (Istio) for traffic management and observability

**Cloud CDN**
- Edge caching with Google's global network
- Integrates with Cloud Storage and Load Balancing

### Key Numbers
| Component | Scale |
|-----------|-------|
| Cloud Spanner | 99.999% SLA, global replication |
| Pub/Sub | 1M+ messages/sec per topic |
| GKE | Thousands of pods per cluster |
| Cloud CDN | 200+ edge locations |

### Architecture Diagram (ASCII)

```
                    ┌─────────────────┐
                    │  Cloud CDN      │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │  Cloud LB       │
                    └────────┬────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
    ┌────▼────┐         ┌────▼────┐         ┌────▼────┐
    │  GKE    │         │  GKE    │         │  GKE    │
    │  Pod    │         │  Pod    │         │  Pod    │
    └────┬────┘         └────┬────┘         └────┬────┘
         │                   │                   │
         └───────────────────┼───────────────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
         ┌────▼────┐    ┌────▼────┐    ┌────▼────┐
         │ Pub/Sub │    │ Spanner │    │ Cloud   │
         │         │    │(Global) │    │ Storage │
         └────┬────┘    └─────────┘    └─────────┘
              │
         ┌────▼────┐
         │ Workers │
         │(GKE)    │
         └─────────┘
```

### Key Takeaways
- **Cloud Spanner** removes sharding complexity for global scale
- **Pub/Sub** enables event-driven, loosely coupled services
- **GKE** provides portable, scalable container orchestration
- **Multi-region** is required for 100M+ users; plan for it early

### Interview Tip
"For 100M users, I'd use a globally distributed database like Spanner, Pub/Sub for async processing, Kubernetes for compute, and CDN for edge delivery. The shift from single-region to multi-region is the key architectural change."

---

## 3. How Khan Academy Scaled to 30 Million Users

### The Problem
Khan Academy needed to scale from millions to tens of millions of learners, especially during COVID-19 when traffic spiked 2.5x in weeks. A small engineering team couldn't manage traditional infrastructure (servers, DBs, scaling).

### The Architecture/Solution

**Google App Engine**
- Fully managed PaaS; no server provisioning or scaling configuration
- Auto-scales from zero to thousands of instances
- Pay per use; no idle capacity costs

**Google Cloud Datastore**
- NoSQL document store with automatic sharding
- Strong consistency for single-entity reads; eventual consistency for queries
- No connection pooling limits; scales with traffic

**Memcache**
- Caches user sessions, course progress, and frequently accessed data
- Reduces Datastore read load significantly

**Fastly CDN**
- Caches videos and static assets at the edge
- Critical for video-heavy educational content

### Key Numbers
| Metric | Value |
|--------|-------|
| Users (COVID peak) | 30M learners |
| Traffic spike | 2.5x in 2 weeks |
| Monthly users (pre-COVID) | 6M |
| Team size | Small; focused on product |

### Architecture Diagram (ASCII)

```
                    ┌─────────────────┐
                    │   Fastly CDN     │
                    │   (Videos/Static)│
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │  App Engine      │
                    │  (Auto-scale)    │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
         ┌────▼────┐    ┌────▼────┐    ┌────▼────┐
         │Memcache │    │Datastore│    │  GCS    │
         │(Cache)  │    │(NoSQL)  │    │(Blob)   │
         └─────────┘    └─────────┘    └─────────┘
```

### Key Takeaways
- **PaaS** (App Engine) lets small teams focus on product, not infra
- **Managed NoSQL** (Datastore) scales without manual sharding
- **CDN** is essential for video-heavy apps
- **Avoid premature optimization**; start simple, scale when needed

### Interview Tip
"Khan Academy chose App Engine and Datastore to avoid operational overhead. For a small team scaling rapidly, managed services beat self-managed infrastructure. The trade-off is vendor lock-in and less control."

---

## 4. How Dropbox Scaled to 100K Users in a Year After Launch

### The Problem
Dropbox needed to sync files across devices efficiently. Naive approach: upload entire file on every change. At 100K users with large files, bandwidth and storage costs would explode. They needed to transfer only what changed.

### The Architecture/Solution

**Block-Based Sync (Delta Sync)**
- Files split into 4 MB chunks; each chunk hashed (SHA-256)
- Only modified chunks transferred; unchanged chunks reused via hash matching
- Similar to rsync; server stores blocklists, not full file copies

**Amazon S3**
- Object storage for file chunks; durability and scalability
- Chunks are content-addressed (hash = key)

**Metadata Architecture**
- **Client**: Local SQLite for file metadata, chunk hashes, sync state
- **Server**: Server File Journal stores namespace + blocklist per file
- **Cursor (Journal ID)**: Clients track position in journal for incremental sync

**Namespace + Path**
- Every file uniquely identified by namespace + relative path
- Enables multi-device sync and conflict resolution

### Key Numbers
| Metric | Value |
|--------|-------|
| Chunk size | 4 MB |
| Early users | 100K in first year |
| Engineers | ~9 at 100K users |
| Hash | SHA-256 per chunk |

### Architecture Diagram (ASCII)

```
    ┌─────────────┐                    ┌─────────────┐
    │  Client A   │                    │  Client B   │
    │  (Device 1) │                    │  (Device 2) │
    └──────┬──────┘                    └──────┬──────┘
           │                                  │
           │ 1. Compute chunk hashes           │
           │ 2. Send only new/changed chunks   │
           │ 3. Receive blocklist + chunks    │
           │                                  │
           └──────────────┬───────────────────┘
                          │
                 ┌────────▼────────┐
                 │  Dropbox API    │
                 │  (Metadata)     │
                 └────────┬────────┘
                          │
              ┌───────────┼───────────┐
              │           │           │
         ┌────▼────┐ ┌────▼────┐ ┌────▼────┐
         │Metadata │ │   S3    │ │ Block   │
         │  DB     │ │(Chunks) │ │ Store   │
         └─────────┘ └─────────┘ └─────────┘
```

### Key Takeaways
- **Delta sync** (block-based) reduces bandwidth by 90%+ for typical edits
- **Content-addressed storage** (hash = identity) enables deduplication
- **Metadata separate from content**; sync logic is in metadata, not blob transfer
- **Start simple**; Dropbox later rewrote sync engine (Nucleus) for petabyte scale

### Interview Tip
"For a file sync service, I'd use block-based delta sync: chunk files, hash each chunk, transfer only changed chunks. Store chunks in S3, metadata in a DB. This is how Dropbox and Google Drive work."

---

## 5. Tech Stack Evolution at Levels.fyi

### The Problem
Levels.fyi started as a side project to collect salary data. It needed to scale from zero to millions of users without over-engineering early. The challenge: evolve the stack as usage grew while avoiding rewrites.

### The Architecture/Solution

**Version 0: Google Forms + Sheets**
- Data collection via Forms; storage in Sheets
- Zero infra; validated demand before building

**Version 1: Lambda + S3**
- Node.js Lambda reads from Sheets; stores JSON in S3
- Static site; serverless for low cost

**Version 2: NestJS + EC2**
- Full backend with NestJS; static site generator
- API Gateway + EC2 for lower latency

**Version 3: PostgreSQL + Microservices**
- RDS PostgreSQL as source of truth
- 5 microservices; service mesh
- Multi-region replication
- Metabase for analytics
- **Kept Sheets** for internal data review (10M cell limit workaround)

**Search at Scale**
- PostgreSQL with materialized views and indexes
- 10M+ search queries/month; p99 < 20ms

### Key Numbers
| Metric | Value |
|--------|-------|
| Monthly users | 2.5M unique |
| Search queries | 10M+/month |
| Search p99 | < 20ms |
| Sheets limit hit | 10M cells |

### Architecture Diagram (ASCII)

```
  V0: Forms → Sheets
  V1: Lambda → Sheets/S3 → Static Site
  V2: EC2 + NestJS → Sheets/S3
  V3:
                    ┌─────────────────┐
                    │   CloudFront    │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │  API Gateway    │
                    └────────┬────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
    ┌────▼────┐         ┌────▼────┐         ┌────▼────┐
    │Service 1│         │Service 2│         │Service N│
    └────┬────┘         └────┬────┘         └────┬────┘
         │                   │                   │
         └───────────────────┼───────────────────┘
                             │
                    ┌────────▼────────┐
                    │   PostgreSQL    │
                    │   (RDS)         │
                    └─────────────────┘
```

### Key Takeaways
- **Start with the simplest thing** (Forms + Sheets) to validate
- **Incremental evolution** beats big-bang rewrites
- **Sheets as interface** persisted even after moving to PostgreSQL
- **Avoid premature optimization**; scale the stack when limits hit

### Interview Tip
"Levels.fyi shows that you can start with Google Sheets and evolve to microservices. The key is to defer complexity until you hit real limits. Don't build for 10M users when you have 100."

---

## 6. How Instagram Scaled to 2.5 Billion Users

### The Problem
Instagram grew from a startup to billions of users. Early stack: Django, PostgreSQL, Redis. At scale: database sharding, feed generation, photo storage, and global latency became critical. Celebrity accounts created hot partitions.

### The Architecture/Solution

**Application Layer**
- Python/Django for API and web
- Cython/C for CPU-intensive paths (image processing, feed ranking)

**Database**
- **MySQL** (sharded) for user data, relationships, metadata
- **Cassandra** for feed storage; separate clusters per region (US, Europe) to manage latency and consistency
- Sharding by user_id; celebrity accounts handled with dedicated logic (hybrid fan-out)

**Caching**
- **Memcached** for user profiles, feed fragments, counters
- **Redis** for real-time features, rate limiting, leaderboards
- Multi-layer caching: in-app, Memcached, then DB

**Photo Storage**
- Migrated from S3 to Facebook Everstore (custom object storage)
- CDN in front for global delivery

**Async Processing**
- RabbitMQ + Celery for image processing, notifications, feed fan-out
- Fan-out on write for normal users; pull on read for celebrities (>10K followers)

### Key Numbers
| Metric | Value |
|--------|-------|
| Users | 2.5B+ |
| Photos | Billions |
| Cassandra | Per-region clusters |
| Celebrity threshold | ~10K followers (pull vs push) |

### Architecture Diagram (ASCII)

```
                    ┌─────────────────┐
                    │      CDN        │
                    │  (Photos)       │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │  Load Balancer  │
                    └────────┬────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
    ┌────▼────┐         ┌────▼────┐         ┌────▼────┐
    │ Django  │         │ Django  │         │ Django  │
    │  App    │         │  App    │         │  App    │
    └────┬────┘         └────┬────┘         └────┬────┘
         │                   │                   │
         └───────────────────┼───────────────────┘
                             │
    ┌────────────────────────┼────────────────────────┐
    │                        │                        │
┌───▼────┐  ┌─────────┐  ┌──▼────┐  ┌─────────┐  ┌──▼────┐
│Memcache│  │  Redis  │  │MySQL  │  │Cassandra│  │Everstore│
│        │  │         │  │Sharded│  │(Feeds)  │  │(Photos) │
└────────┘  └─────────┘  └───────┘  └─────────┘  └────────┘
```

### Key Takeaways
- **Hybrid fan-out**: Push for normal users, pull for celebrities
- **Per-region Cassandra** avoids cross-continent latency
- **Sharding + caching** are foundational; add both early
- **Cython/C** for hot paths when Python is too slow

### Interview Tip
"For a feed system, use hybrid fan-out: push to followers' timelines for normal users, but for celebrities with millions of followers, don't fan out—merge their tweets at read time. Instagram and Twitter both use this pattern."
