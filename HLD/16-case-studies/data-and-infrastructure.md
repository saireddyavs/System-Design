# Data & Infrastructure Case Studies

---

## 1. How Amazon S3 Achieves 99.999999999% Durability

### The Problem
Object storage must never lose data. At exabyte scale, even tiny failure rates mean massive data loss. Disk failures, node failures, and datacenter outages are inevitable.

### The Architecture/Solution
**Erasure Coding**: Data split into K data shards + M parity shards (Reed-Solomon). Reconstruct from any K; tolerate M failures. Storage overhead ~1.5x vs. 3x replication. **Multi-AZ**: Objects across ≥3 Availability Zones. **Continuous Verification**: Background checksums; auto-repair. **Versioning**: Optional; protects against overwrites.

### Key Numbers
| Metric | Value |
|--------|-------|
| Durability | 99.999999999% (11 nines) |
| Objects | 350+ trillion |
| AZs | Minimum 3 per region |

### Architecture Diagram (ASCII)
```
    Object → Erasure encode → Shards (D1..D12, P1..P3) → AZ-1, AZ-2, AZ-3
```

### Key Takeaways
- **Erasure coding** = durability with less storage than full replication
- **Multi-AZ** = survive datacenter failure
- **11 nines** = design for "never lose"

### Interview Tip
"S3 achieves 11 nines via erasure coding + multi-AZ. Any K of N shards reconstruct the object. Erasure coding is more storage-efficient than full replication."

---

## 2. How Amazon S3 Works (Object Storage)

### The Problem
Traditional file systems have limitations at scale. Need flat, scalable storage for billions of objects.

### The Architecture/Solution
**Object Model**: Flat namespace; bucket + key. Objects = data + metadata. REST API. **Eventual → Strong Consistency**: Since 2020, all operations strongly consistent. **Components**: Index (key→shard), storage nodes.

### Key Numbers
| Metric | Value |
|--------|-------|
| Max object | 5 TB |
| Consistency | Strong (2020+) |

### Key Takeaways
- **Flat namespace** scales better than hierarchical FS
- **Strong consistency** since 2020

### Interview Tip
"S3 is object storage: flat namespace, bucket + key, REST API. Strong consistency since 2020."

---

## 3. How Amazon S3 Achieves Strong Consistency

### The Problem
Distributed storage: how to guarantee read-after-write returns new value?

### The Architecture/Solution
**Quorum + Witness**: Write not acked until quorum confirms. Witness nodes participate in consensus. **Check-and-Set**: Version fields prevent stale overwrites.

### Key Takeaways
- **Quorum** ensures writes visible before ack
- **Witness** nodes for consensus

### Interview Tip
"S3 strong consistency: quorum-based with witness nodes. Writes not acked until quorum confirms."

---

## 4. How Amazon Lambda Works

### The Problem
Serverless: scale to zero and millions. Isolate workloads; minimize cold start.

### The Architecture/Solution
**Firecracker**: Lightweight microVM; ~125ms boot. **Cold Start**: First request = create VM + load runtime. Java: 1–5s; Node: 100–500ms. **SnapStart**: Snapshot initialized env; restore ~200ms for Java. **Warm**: Reuse env; <10ms.

### Key Numbers
| Metric | Value |
|--------|-------|
| Firecracker boot | ~125ms |
| Cold (Java) | 1–5s (SnapStart ~200ms) |
| Warm | <10ms |

### Key Takeaways
- **Firecracker** = fast, secure isolation
- **SnapStart** for Java cold start
- **Provisioned concurrency** eliminates cold for critical paths

### Interview Tip
"Lambda uses Firecracker microVMs. Cold start = VM + runtime + init. SnapStart helps Java. Use provisioned concurrency for low latency."

---

## 5. How Cloudflare Supports 55M RPS with Only 15 Postgres Clusters

### The Problem
55M HTTP RPS; many need DB lookups. Postgres has connection limits (~100–500/instance).

### The Architecture/Solution
**PgBouncer**: Multiplex thousands of app connections over hundreds of DB connections. **Thundering Herd**: Limit per tenant; throttle by RTT. **Bare Metal**: No virtualization. **Stolon**: HA, failover.

### Key Numbers
| Metric | Value |
|--------|-------|
| RPS | 55 million |
| Clusters | 15 |
| Pooling | PgBouncer |

### Key Takeaways
- **Connection pooling** essential for high RPS
- **PgBouncer** multiplexes
- **Per-tenant limits** prevent noisy neighbors

### Interview Tip
"Cloudflare: PgBouncer pools connections. 55M RPS ≠ 55M DB connections. Multiplex app→DB."

---

## 6. How Figma Scaled Postgres to 4M Users

### The Problem
Single Postgres hit limits: write contention, query latency, connection saturation.

### The Architecture/Solution
**Phase 1**: Vertical scaling + PgBouncer. **Phase 2**: Vertical partitioning (split by domain: Files DB, Orgs DB). **Phase 3**: Horizontal sharding via DBProxy for largest tables. Stayed on Postgres.

### Key Numbers
| Metric | Value |
|--------|-------|
| Users | 4M |
| Approach | Vertical partition → horizontal shard |
| DBProxy | Custom, 9 months |

### Key Takeaways
- **Vertical partitioning first**; simpler than horizontal
- **Defer horizontal** until single-table limits
- **DBProxy** for shard routing

### Interview Tip
"Figma: vertical partitioning first, then horizontal sharding when tables hit TBs. DBProxy for routing. Avoid horizontal until needed."

---

## 7. This Is How Quora Shards MySQL to Handle 13+ TB

### The Problem
13+ TB, hundreds of thousands QPS. Single DB can't handle.

### The Architecture/Solution
**Vertical Sharding**: Move tables to different servers. **Horizontal Sharding**: Split large tables by key. **Proxy Layer**: Routes queries; ZooKeeper for config. **Avoid Cross-Shard**: Denormalize.

### Key Numbers
| Metric | Value |
|--------|-------|
| Data | 13+ TB |
| Config | ZooKeeper |
| Proxy | Custom/ProxySQL |

### Key Takeaways
- **Vertical then horizontal**
- **Proxy** centralizes routing
- **No cross-shard joins**

### Interview Tip
"Quora: vertical (tables to servers), then horizontal (split by key). Proxy + ZooKeeper. Avoid cross-shard."

---

## 8. Tumblr Shares Database Migration Strategy With 60+ Billion Rows

### The Problem
Migrate 60B rows with zero downtime. Big-bang too risky.

### The Architecture/Solution
**Dual Writes**: Write to both old and new DB. **Shadow Traffic**: Replay to new DB; validate. **Backfill**: Batch copy historical. **Cutover**: Stop old writes; switch reads.

### Key Numbers
| Metric | Value |
|--------|-------|
| Rows | 60+ billion |
| Downtime | Zero |

### Key Takeaways
- **Dual writes** enable zero-downtime migration
- **Shadow traffic** validates
- **Backfill** for historical

### Interview Tip
"For 60B row migration: dual writes, backfill, shadow traffic. Cutover when new ready."

---

## 9. How YouTube Was Able to Support 2.49 Billion Users With MySQL

### The Problem
MySQL doesn't scale horizontally by default. Billions of users.

### The Architecture/Solution
**Vitess**: Open-source MySQL sharding layer. Automatic sharding, connection pooling, query routing. **Sharding**: By user_id or video_id. **VStream**: CDC for analytics.

### Key Numbers
| Metric | Value |
|--------|-------|
| Users | 2.49 billion |
| Solution | Vitess |
| DB | MySQL (sharded) |

### Key Takeaways
- **Vitess** = MySQL at scale
- **Minimal app changes**
- **CNCF project**

### Interview Tip
"YouTube uses Vitess: sharding, pooling, routing. App talks MySQL protocol. Used by Slack, etc."

---

## 10. How Meta Achieves 99.99999999% Cache Consistency

### The Problem
TAO serves billions. Cache inconsistency: DB updated, cache stale. Invalidation races.

### The Architecture/Solution
**Version Fields**: Reject older overwriting newer. **Invalidation Protocol**: Version in invalidation; cache ignores if incoming < cached. **TAO**: Write-through, instant invalidation. **Memcache**: Write-behind, delayed invalidation for less critical. **Monitoring**: Detect, measure, fix.

### Key Numbers
| Metric | Value |
|--------|-------|
| Consistency | 99.99999999% (10 nines) |
| Improvement | From 99.9999% |

### Key Takeaways
- **Version fields** prevent stale overwrites
- **Invalidation ordering** critical
- **Monitoring** essential at scale

### Interview Tip
"Meta: versioned invalidation. If invalidation arrives before fetch, version check prevents stale overwrite. 10 nines requires monitoring."
