# Latency Numbers Every Engineer Should Know

> **Staff+ Engineer Level** — Comprehensive latency reference for system design, performance tuning, and FAANG interviews. Based on Jeff Dean's "Numbers Everyone Should Know" and extended for modern systems.

---

## Table of Contents

1. [Concept Overview](#1-concept-overview)
2. [Real-World Motivation](#2-real-world-motivation)
3. [CPU & Memory Hierarchy](#3-cpu--memory-hierarchy)
4. [Storage Latencies](#4-storage-latencies)
5. [Network Latencies](#5-network-latencies)
6. [Software & Service Latencies](#6-software--service-latencies)
7. [Power of 2 Reference Table](#7-power-of-2-reference-table)
8. [Time Units Conversion](#8-time-units-conversion)
9. [Availability & Downtime Table](#9-availability--downtime-table)
10. [Common Throughput Numbers](#10-common-throughput-numbers)
11. [Interview Discussion](#11-interview-discussion)

---

## 1. Concept Overview

Understanding latency numbers is critical for:

- **Back-of-envelope estimation** — Quick calculations during system design
- **Performance debugging** — Identifying bottlenecks (is it network? disk? CPU?)
- **Architecture decisions** — Choosing between local cache vs remote DB
- **SLA definition** — Setting realistic p50, p99 latency targets

**Key insight:** There's a ~10^9 difference between L1 cache access (0.5ns) and cross-continent network round-trip (150ms). Every order of magnitude matters.

---

## 2. Real-World Motivation

| Scenario | Why Latency Matters |
|----------|---------------------|
| **User experience** | 100ms feels instant; 1s feels slow; 3s users abandon |
| **Trading systems** | Microseconds = millions in arbitrage |
| **Database design** | 1000x difference between memory and SSD—cache everything hot |
| **Geo-distribution** | 150ms RTT cross-continent—can't hide physics |
| **API design** | N+1 queries × 1ms each = 100ms for 100 items |

---

## 3. CPU & Memory Hierarchy

### 3.1 Reference Card (Approximate)

| Operation | Latency | Human-Scale Analogy | Notes |
|-----------|---------|---------------------|-------|
| **L1 cache reference** | 0.5 ns | 1 second | ~1-2 cycles |
| **L2 cache reference** | 7 ns | 14 seconds | ~20-40 cycles |
| **L3 cache reference** | 20-40 ns | 40-80 seconds | Shared across cores |
| **Main memory reference** | 100 ns | 3 minutes | ~200-300 cycles |
| **Branch mispredict** | 5 ns | 10 seconds | Pipeline flush |
| **Mutex lock/unlock** | 25 ns | 50 seconds | Contention adds |
| **Compress 1KB (Snappy)** | 3,000 ns | ~1.5 hours | |
| **Send 1K bytes over 1 Gbps** | 10,000 ns | ~5.5 hours | 1 μs = 1000 ns |

### 3.2 Memory Hierarchy Diagram (ASCII)

```
                    Latency (ns)     Size
┌─────────────────────────────────────────────────────────┐
│ L1 Cache     │ 0.5 ns  │ 32-64 KB per core              │
├─────────────────────────────────────────────────────────┤
│ L2 Cache     │ 7 ns    │ 256 KB - 1 MB per core         │
├─────────────────────────────────────────────────────────┤
│ L3 Cache     │ 20-40 ns│ 8-32 MB shared                 │
├─────────────────────────────────────────────────────────┤
│ Main Memory  │ 100 ns  │ 16-512 GB                      │
├─────────────────────────────────────────────────────────┤
│ SSD          │ 150 μs  │ TBs                            │
├─────────────────────────────────────────────────────────┤
│ HDD          │ 10 ms   │ TBs                            │
└─────────────────────────────────────────────────────────┘
```

### 3.3 Key Ratios

| Comparison | Ratio | Implication |
|------------|-------|-------------|
| L1 vs Main Memory | 200x | Cache miss is expensive |
| Main Memory vs SSD | 1,500x | In-memory DB vs disk DB |
| SSD vs HDD | 66x | SSD random read vs HDD seek |
| L1 vs HDD | 20,000,000x | Why we cache aggressively |

---

## 4. Storage Latencies

### 4.1 Storage Reference Card

| Operation | Latency | Throughput | Notes |
|-----------|---------|------------|-------|
| **SSD random read** | 150 μs (0.15 ms) | 100K-500K IOPS | 4K block |
| **SSD random write** | 50-500 μs | 50K-200K IOPS | Varies by durability |
| **SSD sequential read** | ~1 ms / MB | 500-3000 MB/s | NVMe faster |
| **SSD sequential write** | ~1 ms / MB | 300-2000 MB/s | |
| **HDD seek** | 10 ms | - | Mechanical |
| **HDD sequential read** | ~30 MB/s | - | ~33 ms / MB |
| **HDD random read** | 10+ ms | ~100 IOPS | Seek dominated |
| **NVMe SSD read** | 100-200 μs | 500K+ IOPS | PCIe |
| **NVMe sequential** | - | 3-7 GB/s | |

### 4.2 Storage Comparison Table

| Storage Type | Random Read | Sequential | Use Case |
|--------------|-------------|------------|----------|
| L1/L2/L3 | 0.5-40 ns | - | CPU cache |
| RAM | 100 ns | 10+ GB/s | In-memory DB |
| NVMe SSD | 100-200 μs | 3-7 GB/s | Hot data |
| SATA SSD | 150 μs | 500 MB/s | Warm data |
| HDD | 10 ms | 100-200 MB/s | Cold/archive |

---

## 5. Network Latencies

### 5.1 Network Reference Card

| Scenario | Latency (RTT) | Bandwidth | Notes |
|----------|---------------|------------|-------|
| **Same datacenter (same rack)** | 50-100 μs | 10-100 Gbps | Sub-millisecond |
| **Same datacenter (cross-rack)** | 200-500 μs | 10-100 Gbps | |
| **Same region (cross-AZ)** | 1-5 ms | 1-10 Gbps | AWS us-east-1a → 1b |
| **Cross-region (same continent)** | 20-50 ms | 1-10 Gbps | US East → US West |
| **Cross-continent** | 100-200 ms | 100 Mbps - 1 Gbps | US → Europe |
| **Trans-Pacific** | 150-200 ms | - | US → Asia |
| **Satellite** | 500-700 ms | - | GEO orbit |

### 5.2 Bandwidth-Limited Latency

| Data Size | 1 Gbps | 10 Gbps | 100 Gbps |
|-----------|--------|---------|----------|
| 1 KB | 8 μs | 0.8 μs | 0.08 μs |
| 1 MB | 8 ms | 0.8 ms | 0.08 ms |
| 10 MB | 80 ms | 8 ms | 0.8 ms |
| 100 MB | 800 ms | 80 ms | 8 ms |

**Rule of thumb:** 1 Gbps ≈ 10 ms to send 1 MB (accounting for overhead, ~8 ms theoretical)

### 5.3 Network Hierarchy (ASCII)

```
Same Rack      Same DC        Cross-Region    Cross-Continent
  50μs           500μs           20ms             150ms
┌──────┐       ┌──────┐       ┌──────┐         ┌──────┐
│  A   │       │  A   │       │  A   │         │  A   │
│  B   │       │  B   │       │      │         │      │
└──────┘       └──────┘       │  B   │         │  B   │
                              └──────┘         └──────┘
```

---

## 6. Software & Service Latencies

### 6.1 Database & Cache

| System | Operation | Latency (p50) | Latency (p99) | Notes |
|--------|-----------|---------------|---------------|-------|
| **Redis** | GET (local) | 0.1 ms | 1 ms | In-memory |
| **Redis** | GET (cross-region) | 5-50 ms | 50-100 ms | Network dominated |
| **Memcached** | GET | 0.1 ms | 1 ms | Similar to Redis |
| **MySQL** | Simple query (indexed) | 0.5-2 ms | 5-20 ms | Depends on data |
| **PostgreSQL** | Simple query | 0.5-2 ms | 5-20 ms | |
| **MongoDB** | Find by _id | 1-5 ms | 10-50 ms | |
| **DynamoDB** | GetItem | 1-5 ms | 10-20 ms | Single-digit ms |
| **Cassandra** | Read (local) | 1-5 ms | 10-50 ms | |
| **Elasticsearch** | Simple search | 5-50 ms | 50-200 ms | |

### 6.2 External Services

| Service | Latency | Notes |
|---------|---------|-------|
| **DNS lookup** | 10-100 ms | Cached: <1 ms; uncached: 20-120 ms |
| **TLS handshake** | 50-300 ms | Full: 1-2 RTT; resumption: 1 RTT |
| **HTTP request (simple)** | 100-500 ms | Depends on server, payload |
| **CDN cache hit** | 10-50 ms | Edge to user |
| **CDN cache miss** | 100-500 ms | Origin fetch |
| **OAuth token validation** | 5-50 ms | JWKS fetch + verify |
| **S3 GET** | 10-50 ms | First byte |
| **S3 PUT** | 50-200 ms | Depends on size |
| **Kafka produce** | 1-10 ms | Ack=1; ack=all adds latency |

### 6.3 End-to-End Request Breakdown (Typical API)

| Component | Latency | Cumulative |
|-----------|---------|------------|
| DNS | 0-50 ms | 50 ms |
| TCP connect | 1 RTT | +50 ms (cross-region) |
| TLS handshake | 1-2 RTT | +100 ms |
| HTTP request | 1 RTT | +50 ms |
| Application | 10-100 ms | +100 ms |
| Database query | 1-10 ms | +10 ms |
| **Total (cross-region)** | | **~300-400 ms** |
| **Total (same DC)** | | **~20-50 ms** |

---

## 7. Power of 2 Reference Table

Essential for capacity planning and estimation.

| Power | Value | Name | Approximate |
|-------|-------|------|-------------|
| 2^10 | 1,024 | 1 K (Kilo) | 10^3 |
| 2^20 | 1,048,576 | 1 M (Mega) | 10^6 |
| 2^30 | 1,073,741,824 | 1 G (Giga) | 10^9 |
| 2^40 | 1,099,511,627,776 | 1 T (Tera) | 10^12 |
| 2^50 | 1,125,899,906,842,624 | 1 P (Peta) | 10^15 |
| 2^60 | ~1.15 × 10^18 | 1 E (Exa) | 10^18 |

### 7.1 Quick Conversions

| From | To | Factor |
|------|-----|--------|
| KB → MB | ÷ 1024 | |
| MB → GB | ÷ 1024 | |
| GB → TB | ÷ 1024 | |
| **Approximate** | KB→MB: ÷1000, MB→GB: ÷1000 | Easier for mental math |

### 7.2 Time in Seconds (Powers of 10)

| Value | Seconds | Human |
|-------|---------|-------|
| 10^0 | 1 s | 1 second |
| 10^1 | 10 s | 10 seconds |
| 10^2 | 100 s | ~1.5 min |
| 10^3 | 1000 s | ~17 min |
| 10^4 | 10,000 s | ~2.8 hours |
| 10^5 | 100,000 s | ~1.2 days |
| 10^6 | 1,000,000 s | ~11.6 days |
| 10^7 | 10^7 s | ~116 days |
| 10^8 | 10^8 s | ~3.2 years |
| 10^9 | 10^9 s | ~31.7 years |

### 7.3 Requests per Time Period

| QPS | Per Minute | Per Hour | Per Day |
|-----|------------|----------|---------|
| 1 | 60 | 3,600 | 86,400 |
| 10 | 600 | 36,000 | 864,000 |
| 100 | 6,000 | 360,000 | 8.64 M |
| 1,000 | 60,000 | 3.6 M | 86.4 M |
| 10,000 | 600,000 | 36 M | 864 M |
| 100,000 | 6 M | 360 M | 8.64 B |

**Rule of thumb:** 1K QPS ≈ 86M requests/day; 10K QPS ≈ 864M/day

---

## 8. Time Units Conversion

| Unit | In Seconds | In Milliseconds | In Microseconds |
|------|------------|-----------------|-----------------|
| 1 second | 1 | 1,000 | 1,000,000 |
| 1 millisecond (ms) | 0.001 | 1 | 1,000 |
| 1 microsecond (μs) | 0.000001 | 0.001 | 1 |
| 1 nanosecond (ns) | 10^-9 | 10^-6 | 0.001 |

### 8.1 Conversion Shortcuts

```
1 second = 1,000 ms = 1,000,000 μs = 1,000,000,000 ns
1 ms = 1,000 μs
1 μs = 1,000 ns

To convert ms → μs: multiply by 1,000
To convert μs → ms: divide by 1,000
```

### 8.2 Human Perception of Latency

| Latency | User Perception |
|---------|-----------------|
| < 100 ms | Instant |
| 100-300 ms | Slight delay, acceptable |
| 300-1000 ms | Noticeable, tolerable |
| 1-3 s | Slow, users notice |
| > 3 s | Unacceptable, users abandon |

---

## 9. Availability & Downtime Table

### 9.1 Availability to Downtime

| Availability | Downtime/Year | Downtime/Month | Downtime/Week |
|--------------|---------------|----------------|---------------|
| 90% | 36.5 days | 3 days | 16.8 hours |
| 99% | 3.65 days | 7.2 hours | 1.68 hours |
| 99.9% (3 nines) | 8.76 hours | 43.2 min | 10.1 min |
| 99.95% | 4.38 hours | 21.6 min | 5.04 min |
| 99.99% (4 nines) | 52.6 min | 4.32 min | 1.01 min |
| 99.999% (5 nines) | 5.26 min | 25.9 sec | 6.05 sec |
| 99.9999% (6 nines) | 31.5 sec | 2.59 sec | 0.605 sec |

### 9.2 Calculation Formula

```
Downtime per year = (1 - availability) × 365.25 × 24 × 60 × 60 seconds

Examples:
99.9% → 0.001 × 31,557,600 ≈ 31,557 sec ≈ 8.76 hours
99.99% → 0.0001 × 31,557,600 ≈ 3,155 sec ≈ 52.6 min
```

### 9.3 Industry Standards

| System Type | Typical SLA |
|-------------|-------------|
| Consumer web | 99.9% |
| Enterprise SaaS | 99.95% |
| Financial, healthcare | 99.99%+ |
| Critical infrastructure | 99.999% |

---

## 10. Common Throughput Numbers

### 10.1 Per-Node Throughput (Approximate)

| System | Throughput | Notes |
|--------|------------|-------|
| **Redis** | 100K-500K ops/sec | Single node, simple ops |
| **Memcached** | 100K-500K ops/sec | Similar to Redis |
| **MySQL** | 1K-10K TPS | Depends on workload |
| **PostgreSQL** | 1K-10K TPS | |
| **Kafka** | 100K-1M msg/sec | Per broker, partitioned |
| **Nginx** | 10K-50K req/sec | Static, keep-alive |
| **gRPC** | 50K-100K req/sec | Single connection |
| **REST API** | 1K-10K req/sec | App dependent |

### 10.2 Network Throughput

| Link | Bandwidth | 1 MB Transfer |
|------|------------|----------------|
| 1 Gbps | 125 MB/s | 8 ms |
| 10 Gbps | 1.25 GB/s | 0.8 ms |
| 100 Gbps | 12.5 GB/s | 0.08 ms |

### 10.3 Storage Throughput

| Storage | Read | Write |
|---------|------|-------|
| NVMe SSD | 3-7 GB/s | 2-5 GB/s |
| SATA SSD | 500 MB/s | 400 MB/s |
| HDD | 100-200 MB/s | 100-200 MB/s |
| S3 (aggregate) | 3,500 PUT/COPY, 5,500 GET/sec per prefix | |

---

## 11. Interview Discussion

### 11.1 Key Numbers to Memorize

1. **L1 cache:** 0.5 ns  
2. **Main memory:** 100 ns  
3. **SSD random read:** 150 μs  
4. **HDD seek:** 10 ms  
5. **Same DC RTT:** 500 μs  
6. **Cross-continent RTT:** 150 ms  
7. **Redis GET:** 0.1 ms  
8. **MySQL simple query:** 0.5-2 ms  

### 11.2 Estimation Shortcuts

- **1 Gbps ≈ 10 ms per MB** (rough)
- **1K QPS ≈ 86M requests/day**
- **99.9% = 8.76 hours downtime/year**
- **2^10 = 1K, 2^20 = 1M, 2^30 = 1G**

### 11.3 Common Interview Questions

**Q: Why is caching so effective?**  
A: L1 (0.5ns) vs SSD (150μs) = 300,000x. Memory vs SSD = 1,500x. Avoiding disk pays off massively.

**Q: Why does cross-region add latency?**  
A: Speed of light: ~150ms RTT US-Europe. Physics limit. Mitigate with edge, replication.

**Q: How many requests per day at 10K QPS?**  
A: 10K × 86,400 = 864M requests/day.

**Q: What's 99.99% downtime?**  
A: 52.6 minutes per year.

---

*Document: Latency Numbers Reference — 450+ lines | FAANG Interview Essential*
