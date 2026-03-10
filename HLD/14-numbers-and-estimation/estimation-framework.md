# Back-of-Envelope Estimation Framework

> **Staff+ Engineer Level** — Step-by-step methodology for system design estimation. Essential for FAANG interviews and real-world capacity planning.

---

## Table of Contents

1. [Concept Overview](#1-concept-overview)
2. [Step 1: Clarify Requirements](#2-step-1-clarify-requirements)
3. [Step 2: Estimate Traffic](#3-step-2-estimate-traffic)
4. [Step 3: Estimate Storage](#4-step-3-estimate-storage)
5. [Step 4: Estimate Bandwidth](#5-step-4-estimate-bandwidth)
6. [Step 5: Estimate Memory/Cache](#6-step-5-estimate-memorycache)
7. [Formulas & Conversion Shortcuts](#7-formulas--conversion-shortcuts)
8. [Rules of Thumb](#8-rules-of-thumb)
9. [Sanity Check Tips](#9-sanity-check-tips)
10. [Worked Example 1: Twitter Timeline](#10-worked-example-1-twitter-timeline)
11. [Worked Example 2: YouTube](#11-worked-example-2-youtube)
12. [Worked Example 3: URL Shortener](#12-worked-example-3-url-shortener)
13. [Worked Example 4: Chat System](#13-worked-example-4-chat-system)
14. [Worked Example 5: Uber](#14-worked-example-5-uber)
15. [Interview Discussion](#15-interview-discussion)

---

## 1. Concept Overview

Back-of-envelope estimation is the art of quickly approximating system requirements using simple math and known constants. It's used for:

- **System design interviews** — Demonstrate capacity planning
- **Architecture decisions** — Size databases, caches, CDN
- **Cost estimation** — Budget for infrastructure
- **Feasibility** — "Can we build this with 10 servers?"

**Principles:**
- Round numbers (use 100M, not 97.3M)
- Powers of 2 and 10
- State assumptions explicitly
- Sanity check against known systems

---

## 2. Step 1: Clarify Requirements

### 2.1 Key Questions to Ask

| Question | Why It Matters |
|----------|----------------|
| **DAU (Daily Active Users)?** | Base for traffic |
| **Peak / average ratio?** | 2-3x typical |
| **Read vs write ratio?** | 10:1, 100:1 common |
| **Storage retention?** | 1 year? Forever? |
| **Latency SLA?** | p99 < 100ms? |
| **Geographic distribution?** | Single region vs global |

### 2.2 Assumptions Template

```
Assume:
- DAU: X million
- Peak: 3x average
- Read:Write = 100:1
- Retention: 5 years
- Data per operation: Y bytes
```

### 2.3 Common DAU References

| DAU | Typical Scale |
|-----|---------------|
| 1M | Mid-size startup |
| 10M | Growing product |
| 100M | Large (Twitter scale) |
| 1B | Facebook, WhatsApp scale |

---

## 3. Step 2: Estimate Traffic

### 3.1 Formula

```
Daily requests = DAU × (requests per user per day)
QPS (avg) = Daily requests / 86,400
QPS (peak) = QPS (avg) × Peak multiplier (2-3)
```

### 3.2 Requests per User (Typical)

| Product Type | Reads/User/Day | Writes/User/Day |
|--------------|----------------|-----------------|
| Social feed | 50-200 | 5-20 |
| Chat | 100-500 | 50-200 |
| E-commerce | 20-50 | 2-10 |
| Search | 5-20 | 0.1-1 |
| Video | 10-50 | 0.1-1 |
| URL shortener | 10-50 | 0.1-1 |

### 3.3 Example Calculation

```
DAU = 100M
Reads/user/day = 100
Writes/user/day = 10
Read/Write = 10:1

Daily reads = 100M × 100 = 10B
Daily writes = 100M × 10 = 1B

Read QPS = 10B / 86,400 ≈ 115,000
Write QPS = 1B / 86,400 ≈ 12,000

Peak (3x): Read 350K, Write 36K
```

---

## 4. Step 3: Estimate Storage

### 4.1 Formula

```
Storage per day = (Operations per day) × (Bytes per operation)
Total storage = Daily storage × Retention days × Replication factor
```

### 4.2 Bytes per Operation (Typical)

| Data Type | Size | Notes |
|-----------|------|-------|
| Tweet | 300 bytes | Text + metadata |
| User profile | 1 KB | |
| Photo metadata | 500 bytes | |
| Photo (full) | 3-5 MB | |
| Video (1 min) | 50-100 MB | 1080p |
| Short URL mapping | 100 bytes | |
| Chat message | 200-500 bytes | |
| Location update | 50-100 bytes | |

### 4.3 Replication Factor

| System | Typical |
|--------|---------|
| Database | 3 |
| Object storage | 3 |
| Cache | 2 (optional) |
| CDN | 1 (no replication, distributed) |

### 4.4 Example Calculation

```
Daily writes = 1B
Bytes per write = 500 bytes
Retention = 5 years = 1,825 days
Replication = 3

Daily storage = 1B × 500 = 500 GB
Total = 500 GB × 1,825 × 3 = 2.7 PB
```

---

## 5. Step 4: Estimate Bandwidth

### 5.1 Formula

```
Read bandwidth = Read QPS × Bytes per response
Write bandwidth = Write QPS × Bytes per request
Total = Read + Write
```

### 5.2 Bytes per Request/Response (Typical)

| Operation | Request | Response |
|-----------|---------|----------|
| API (JSON) | 500 B | 2-10 KB |
| Timeline | 200 B | 10-50 KB |
| Search | 100 B | 5-20 KB |
| Video stream | 100 B | 5-25 Mbps |
| Photo upload | 3 MB | 100 B |
| Chat message | 500 B | 500 B |

### 5.3 Example

```
Read QPS = 100K
Avg response = 10 KB

Read bandwidth = 100K × 10 KB = 1 GB/s = 8 Gbps
```

---

## 6. Step 5: Estimate Memory/Cache

### 6.1 80/20 Rule

```
20% of data drives 80% of traffic
Cache hit rate target: 80%+
```

### 6.2 Formula

```
Cache size = Daily active data × 20% × Replication
Or: Cache size = (QPS × Avg object size) × (TTL in seconds) × 20%
```

### 6.3 Example

```
Daily reads = 10B
Avg object = 5 KB
Hot data = 20%

Cache size = 10B × 5 KB × 0.2 = 10 TB (for hot set)
```

---

## 7. Formulas & Conversion Shortcuts

### 7.1 Core Formulas

```
QPS = DAU × (requests per user per day) / 86,400
Peak QPS = Avg QPS × 2 to 3

Storage = Ops/day × Bytes/op × Retention × Replication

Bandwidth (bps) = QPS × Bytes × 8

1 day = 86,400 seconds
1 year = 31,536,000 seconds ≈ 3.15 × 10^7
```

### 7.2 Conversion Shortcuts

| Conversion | Shortcut |
|------------|----------|
| QPS → Daily | × 86,400 |
| Daily → QPS | ÷ 86,400 |
| 1K QPS | 86.4M/day |
| 10K QPS | 864M/day |
| 100K QPS | 8.64B/day |
| 1 GB/s | 8 Gbps |
| 1 MB/s | 8 Mbps |

### 7.3 Power of 2 (Quick)

```
2^10 = 1K
2^20 = 1M
2^30 = 1G
2^40 = 1T
```

---

## 8. Rules of Thumb

| Rule | Value |
|------|-------|
| Peak / average | 2-3x |
| Read / write | 10:1 to 100:1 common |
| Cache hit rate | 80% target |
| Cache size | 20% of hot data |
| Replication | 3x |
| Single server QPS | 1K-10K (app), 100K (cache) |
| Single server storage | 1-10 TB |
| Single DB write | 1K-10K TPS |


---

## 9. Sanity Check Tips

### 9.1 Cross-Reference

| Your Estimate | Compare To |
|---------------|------------|
| Twitter | 6K write, 300K read QPS |
| WhatsApp | 100B msg/day |
| YouTube | 500 hours/min upload |
| 100K QPS | 8.64B requests/day |

### 9.2 Red Flags

- **QPS too high:** DAU × 1000 reads/day = 1.16M QPS for 100M DAU—possible but verify
- **Storage too low:** 1B tweets × 300 bytes = 300 GB—not 30 GB
- **Bandwidth:** 100K QPS × 10 KB = 1 GB/s—can you serve that?

### 9.3 Order of Magnitude

```
Is it 1K, 10K, 100K, or 1M QPS?
Is it 1 TB, 10 TB, 100 TB, or 1 PB?
Is it 1 Gbps, 10 Gbps, or 100 Gbps?
```

---

## 10. Worked Example 1: Twitter Timeline

### 10.1 Requirements

- DAU: 200M
- Tweets per user per day: 5
- Timeline reads per user per day: 100
- Read:Write = 20:1

### 10.2 Step 1: Traffic

```
Daily tweets = 200M × 5 = 1B
Daily timeline reads = 200M × 100 = 20B

Write QPS = 1B / 86,400 ≈ 12,000 (round to 10K)
Read QPS = 20B / 86,400 ≈ 230,000 (round to 250K)

Peak (2.5x): Write 25K, Read 600K
```

### 10.3 Step 2: Storage

```
Tweet size: 300 bytes (text + metadata)
Retention: 7 years
Replication: 3

Daily storage = 1B × 300 = 300 GB
Annual = 300 × 365 = 110 TB
7 years = 770 TB
With replication: 770 × 3 = 2.3 PB
```

### 10.4 Step 3: Bandwidth

```
Write: 25K × 1 KB = 25 MB/s
Read: 600K × 20 KB (timeline) = 12 GB/s

Total read: 12 GB/s = 96 Gbps
```

### 10.5 Step 4: Cache

```
Hot data: 20% of daily reads
Daily read volume = 20B × 20 KB = 400 TB

Cache 20% = 80 TB
But cache stores objects, not raw reads: 
  Unique timelines × 20 KB × 20% ≈ 10-20 TB
```

### 10.6 Summary

| Metric | Estimate |
|--------|----------|
| Write QPS | 10K avg, 25K peak |
| Read QPS | 250K avg, 600K peak |
| Storage (7 yr) | 2.3 PB |
| Bandwidth | 96 Gbps read |
| Cache | 10-20 TB |

---

## 11. Worked Example 2: YouTube

### 11.1 Requirements

- 500 hours video uploaded per minute
- 1B hours watched per day
- Avg video: 10 min, 500 MB (1080p)
- CDN for delivery

### 11.2 Step 1: Upload Traffic

```
500 hours/min = 30,000 hours/hour = 720,000 hours/day

Upload bytes/day = 720,000 hours × (500 MB / 10 min) × 60 min/hour
                 = 720,000 × 50 MB × 60
                 = 2.16 PB/day (simplified)

More realistic: 500 MB per 10 min = 50 MB/hour of video
720,000 hours × 50 MB = 36 TB/hour = 864 TB/day raw
With multiple encodings (5x): 4.3 PB/day
```

### 11.3 Step 2: Storage

```
Daily upload: 864 TB (raw)
With encodings (5x): 4.3 PB/day
Replication (3x): 13 PB/day
Annual retention: 4.7 EB
Historical (10 years): 47 EB
```

### 11.4 Step 3: CDN Bandwidth

```
1B hours watched/day
Avg bitrate: 5 Mbps

Peak hour factor: 10% of daily in 1 hour
Peak hours = 100M
100M hours × 5 Mbps = 500 Tbps (theoretical)

More realistic: 50M concurrent × 5 Mbps = 250 Tbps
Or: 15% of internet = 10-20 Tbps
```

### 11.5 Summary

| Metric | Estimate |
|--------|----------|
| Upload | 500 hours/min, 864 TB/day raw |
| Storage | 4.7 EB/year |
| CDN | 10-20 Tbps peak |
| Encodings | 5-10 per video |

---

## 12. Worked Example 3: URL Shortener

### 12.1 Requirements

- DAU: 50M
- Shortens per user per day: 2
- Clicks per short URL per day: 10
- Retention: 5 years

### 12.2 Step 1: Traffic

```
Daily shortens = 50M × 2 = 100M
Daily clicks = 100M × 10 = 1B (each short URL clicked 10x)

Write QPS = 100M / 86,400 ≈ 1,200
Read QPS = 1B / 86,400 ≈ 12,000

Peak (3x): Write 3.6K, Read 36K
```

### 12.3 Step 2: Key Space

```
Short URL: 7 chars, alphanumeric (62 options)
62^7 = 3.5 trillion (enough for 100M/day for 35,000 days)

Or 6 chars: 62^6 = 56 billion (enough for 56 days at 1B/day)
```

### 12.4 Step 3: Storage

```
Mapping: short_url (7 bytes) + long_url (100 bytes) + metadata (50 bytes) ≈ 150 bytes

Daily = 100M × 150 = 15 GB
5 years = 15 × 365 × 5 = 27 TB
With replication: 81 TB
```

### 12.5 Step 4: Bandwidth

```
Read: 36K × 200 bytes = 7.2 MB/s (negligible)
Write: 3.6K × 500 bytes = 1.8 MB/s
```

### 12.6 Summary

| Metric | Estimate |
|--------|----------|
| Write QPS | 1.2K avg, 3.6K peak |
| Read QPS | 12K avg, 36K peak |
| Key space | 62^7 = 3.5T (7 chars) |
| Storage (5 yr) | 81 TB |
| Bandwidth | 10 MB/s |

---

## 13. Worked Example 4: Chat System

### 13.1 Requirements

- DAU: 100M
- Messages per user per day: 50
- 1:1 and group (avg 10 people per group message)
- Retention: 2 years

### 13.2 Step 1: Traffic

```
Daily messages = 100M × 50 = 5B

Assume 1:1 = 60%, Group = 40%
1:1 deliveries = 2 (sender + receiver)
Group deliveries = 10 (avg)

1:1: 3B × 2 = 6B deliveries
Group: 2B × 10 = 20B deliveries
Total deliveries = 26B

Write QPS (ingest) = 5B / 86,400 ≈ 60,000
Read QPS (delivery) = 26B / 86,400 ≈ 300,000

Peak (2x): Write 120K, Read 600K
```

### 13.3 Step 2: WebSocket Connections

```
Concurrent users: 10% of DAU at peak = 10M
Connections per user: 1 (mobile) or 2 (mobile + web)

Peak connections = 10M × 1.5 = 15M
```

### 13.4 Step 3: Storage

```
Message size: 100 bytes (metadata) + 200 bytes (text) = 300 bytes

Daily = 5B × 300 = 1.5 TB
2 years = 1.5 × 730 = 1,095 TB ≈ 1 PB
With replication: 3 PB
```

### 13.5 Step 4: Bandwidth

```
Write: 120K × 300 bytes = 36 MB/s
Read: 600K × 300 bytes = 180 MB/s

Total: 216 MB/s ≈ 1.7 Gbps
```

### 13.6 Summary

| Metric | Estimate |
|--------|----------|
| Write QPS | 60K avg, 120K peak |
| Read QPS | 300K avg, 600K peak |
| WebSocket connections | 15M peak |
| Storage (2 yr) | 3 PB |
| Bandwidth | 1.7 Gbps |

---

## 14. Worked Example 5: Uber

### 14.1 Requirements

- 20M rides/day
- 5M active drivers
- Location update every 4 seconds per driver
- Map tile: 50 KB avg
- Retention: 7 days for location

### 14.2 Step 1: Location Updates

```
Updates per driver per second = 1/4 = 0.25
Total updates = 5M × 0.25 = 1.25M QPS

Peak (1.5x): 2M QPS
```

### 14.3 Step 2: Ride Requests

```
20M rides / 86,400 = 230 QPS (avg)
Peak (3x): 700 QPS
```

### 14.4 Step 3: Map Tile Requests

```
Assume 10 tile requests per ride request + 5 per driver per min
Ride: 700 × 10 = 7K
Driver: 5M × 5/60 = 400K

Total map tile: ~400K QPS
```

### 14.5 Step 4: Storage (Location)

```
Location update: 100 bytes (lat, lng, driver_id, timestamp)
1.25M × 100 = 125 MB/s
Daily = 125 × 86,400 = 10.8 TB
7 days retention: 75 TB
With replication: 225 TB
```

### 14.6 Step 5: Bandwidth

```
Location write: 1.25M × 100 = 125 MB/s
Map tile read: 400K × 50 KB = 20 GB/s

Total: 20 GB/s ≈ 160 Gbps
```

### 14.7 Summary

| Metric | Estimate |
|--------|----------|
| Location updates | 1.25M QPS |
| Ride requests | 230 QPS avg, 700 peak |
| Map tile requests | 400K QPS |
| Location storage (7d) | 225 TB |
| Bandwidth | 160 Gbps |

---

## 15. Interview Discussion

### 15.1 Framework to Present

1. **Clarify:** "Let me assume DAU of 100M, 100 reads/user/day, 10 writes/user/day."
2. **Traffic:** "That's 10B reads, 1B writes per day. 115K read QPS, 12K write QPS. Peak 3x."
3. **Storage:** "1B × 500 bytes = 500 GB/day. 5 years × 3 replication = 2.7 PB."
4. **Bandwidth:** "350K × 10 KB = 3.5 GB/s read."
5. **Cache:** "20% hot = 10 TB cache for 80% hit rate."

### 15.2 Common Interview Questions

**Q: How many servers for 100K QPS?**  
A: "If each server handles 1K QPS (app): 100 servers. With cache: 2K QPS: 50 app servers. Cache layer: 100K / 50K = 2 Redis clusters."

**Q: How much storage for 1B users?**  
A: "Depends on data per user. 1 KB profile: 1 TB. 100 KB (with activity): 100 TB. With replication: 300 TB."

**Q: Bandwidth for 1M concurrent video streams?**  
A: "1M × 5 Mbps = 5 Tbps. That's 15% of global internet—need massive CDN."

### 15.3 Key Takeaways

| Step | Formula | Example |
|------|---------|---------|
| Traffic | DAU × req/user/day / 86400 | 100M × 100 / 86400 = 115K |
| Storage | Ops × bytes × retention × replication | 1B × 500 × 1825 × 3 = 2.7 PB |
| Bandwidth | QPS × bytes × 8 | 100K × 10K × 8 = 8 Gbps |
| Cache | 20% of hot data | 10B × 5K × 0.2 = 10 TB |

---

*Document: Estimation Framework — 550+ lines | FAANG Interview Essential*
