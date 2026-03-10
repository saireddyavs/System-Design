# Scale Numbers for Real Systems

> **Staff+ Engineer Level** — Production scale numbers for major tech systems. Essential for FAANG system design interviews and realistic architecture discussions.

---

## Table of Contents

1. [Concept Overview](#1-concept-overview)
2. [Twitter / X](#2-twitter--x)
3. [YouTube](#3-youtube)
4. [Uber](#4-uber)
5. [WhatsApp](#5-whatsapp)
6. [Netflix](#6-netflix)
7. [Instagram](#7-instagram)
8. [Google Search](#8-google-search)
9. [Facebook / Meta](#9-facebook--meta)
10. [Storage & Bandwidth Calculations](#10-storage--bandwidth-calculations)
11. [Interview Discussion](#11-interview-discussion)

---

## 1. Concept Overview

Real production numbers help with:

- **Sanity checks** — Is your estimate in the right ballpark?
- **Interview credibility** — "Twitter handles 300K read QPS" shows depth
- **Architecture decisions** — Scale informs sharding, caching, CDN needs
- **Capacity planning** — Model growth from known benchmarks

**Note:** Numbers are approximate and from public disclosures, engineering blogs, and industry estimates. They change over time.

---

## 2. Twitter / X

### 2.1 Key Metrics

| Metric | Value | Notes |
|--------|-------|-------|
| **DAU** | ~250-350M | Varies by year |
| **Tweets per day** | ~500M | Peak: higher |
| **Tweets per second (avg)** | ~6,000 | 500M / 86,400 |
| **Tweets per second (peak)** | ~15,000-20,000 | 2-3x average |
| **Read QPS (timeline)** | ~300,000 | 10x write (read-heavy) |
| **Read QPS (peak)** | ~1M+ | |
| **Data growth** | ~600 TB/year | Tweets + media + logs |
| **Storage (total)** | Multi-PB | Historical tweets, media |

### 2.2 Traffic Breakdown

| Operation | QPS (approx) | Daily Volume |
|-----------|--------------|--------------|
| Tweet creation | 6,000 avg | 500M |
| Timeline read | 300,000 | 26B |
| Like/follow | 50,000+ | 4B+ |
| Search | 10,000+ | 1B+ |

### 2.3 Storage Calculation (Example)

```
Per tweet (text only): ~300 bytes (tweet + metadata)
500M tweets/day × 300 bytes = 150 GB/day text
With media, indexes, replication: ~2-5x = 300-750 GB/day
Annual: ~100-270 TB (text), ~600 TB+ (total)
```

### 2.4 Bandwidth

```
Write: 6K tweets/sec × 1 KB avg = 6 MB/s write
Read: 300K reads/sec × 10 KB avg (timeline) = 3 GB/s read
CDN for media: 10x+ of API traffic
```

### 2.5 Architecture Implications

- **Fan-out:** Pre-compute timeline (write fan-out) vs read-time merge
- **Cache:** Redis/Memcached for hot timelines; 80/20 rule
- **Sharding:** By user_id for tweets, timeline
- **Search:** Elasticsearch/similar; real-time index

---

## 3. YouTube

### 3.1 Key Metrics

| Metric | Value | Notes |
|--------|-------|-------|
| **MAU** | ~2.5B+ | |
| **Video uploads** | 500 hours/minute | 720,000 hours/day |
| **Watch time** | 1B+ hours/day | |
| **Storage** | ~1 EB (exabyte) | Videos, multiple encodings |
| **CDN traffic** | 15-20% of global internet | Peak |
| **Videos (total)** | 5B+ | |

### 3.2 Traffic Breakdown

| Operation | Rate | Daily Volume |
|-----------|------|--------------|
| Upload | 500 hours/min = 30K hours/hour | 720K hours/day |
| Upload (bytes) | ~100-500 GB/min | ~150-700 TB/day (raw) |
| Stream requests | Millions of QPS | 1B+ hours watched |
| Thumbnails | 10x video requests | |

### 3.3 Storage Calculation

```
500 hours/min = 30,000 hours/hour = 720,000 hours/day
Assume 1 GB/hour average (1080p): 720 TB/day raw upload
With encodings (multiple resolutions): 5-10 versions = 3.6-7.2 PB/day
With replication (3x): 10-20 PB/day
Annual growth: ~1-5 EB (historical accumulation)
Total storage: ~1 EB (exabyte)
```

### 3.4 Bandwidth (CDN)

```
1B hours watched/day = 41.7M hours/hour
Assume 5 Mbps avg bitrate: 41.7M × 5 = 208 Tbps peak (simplified)
Actual: 15-20% of internet = 10-20 Tbps+ egress
```

### 3.5 Architecture Implications

- **Transcoding:** Async pipeline; multiple resolutions (144p to 8K)
- **CDN:** Edge caching; 90%+ cache hit for popular
- **Storage:** Cold storage for old videos; hot for trending
- **Metadata:** Separate from video blob storage

---

## 4. Uber

### 4.1 Key Metrics

| Metric | Value | Notes |
|--------|-------|-------|
| **Rides per day** | ~20M | Global |
| **Driver location updates** | Every 4 seconds | Per active driver |
| **Concurrent connections** | ~1M+ | WebSocket/long poll |
| **Active drivers** | ~5M+ | At peak |
| **Cities** | 10,000+ | |
| **QPS (total)** | 100K-1M+ | Reads dominate |

### 4.2 Traffic Breakdown

| Operation | QPS | Notes |
|-----------|-----|-------|
| Location updates | 1.25M+ | 5M drivers × 1/4 sec |
| Ride requests | 500-1000 | 20M / 86,400 × 2 |
| ETA/distance | 10K-50K | Per request |
| Map tile requests | 100K-500K | |
| Driver-rider matching | 1K-5K | |

### 4.3 Storage (Location History)

```
1.25M updates/sec × 100 bytes = 125 MB/s
Per day: 125 × 86,400 = 10.8 TB/day (raw)
With retention (7 days hot): ~75 TB
With indexes, replication: 2-3x
```

### 4.4 Bandwidth

```
Location: 125 MB/s write
Map tiles: 100K req × 50 KB = 5 GB/s read
Total: 10+ GB/s
```

### 4.5 Architecture Implications

- **Geospatial:** Redis Geo, S2, or custom for driver location
- **Real-time:** WebSocket for driver app; push for rider
- **Matching:** Geohash-based; find drivers within radius
- **Map tiles:** CDN; pre-rendered at zoom levels

---

## 5. WhatsApp

### 5.1 Key Metrics

| Metric | Value | Notes |
|--------|-------|-------|
| **Users** | 2B+ | |
| **Messages per day** | 100B+ | |
| **Messages per second (avg)** | ~1.2M | 100B / 86,400 |
| **Messages per second (peak)** | 3.5M+ | |
| **Groups** | 50M+ | |
| **Engineers (legend)** | 50 | At acquisition by Facebook |

### 5.2 Traffic Breakdown

| Operation | QPS | Daily |
|-----------|-----|-------|
| Message send | 1.2M avg, 3.5M peak | 100B |
| Message delivery | 2-3x send (multi-device, groups) | 200-300B |
| Read receipts | 2x messages | |
| Media | 10-20% of messages | |

### 5.3 Storage Calculation

```
100B messages/day × 500 bytes avg = 50 TB/day
With media (10%): +20 TB = 70 TB/day
With indexes, replication (3x): 200+ TB/day
Annual: ~70 PB
```

### 5.4 Bandwidth

```
1.2M msg/sec × 1 KB = 1.2 GB/s write
Delivery (fan-out): 3x = 3.6 GB/s
Read receipts, presence: +1 GB/s
Total: 5+ GB/s
```

### 5.5 Architecture Implications

- **Ephemeral:** Messages may not persist long (depending on type)
- **Fan-out:** Group message to N members = N deliveries
- **Ordering:** Sequence numbers per chat
- **E2E encryption:** Server doesn't see content; metadata only

---

## 6. Netflix

### 6.1 Key Metrics

| Metric | Value | Notes |
|--------|-------|-------|
| **Subscribers** | 250M+ | |
| **Internet traffic share** | 15-17% | Downstream |
| **Concurrent streams** | 10-20% of subs at peak | 25-50M |
| **Content** | 100K+ video files per title | Encodings, languages |
| **Catalog** | 15K+ titles | |
| **Peak traffic** | 15-20% of global internet | |

### 6.2 Traffic Breakdown

| Operation | Rate | Notes |
|-----------|------|-------|
| Stream requests | 25-50M concurrent | Peak |
| Bitrate | 5-25 Mbps | 4K = 25 Mbps |
| Aggregate bandwidth | 100+ Tbps | Peak |
| Recommendation requests | 100K+ QPS | |

### 6.3 Storage

```
100K files × 15K titles = 1.5B files (approx)
Average movie: 2 GB (1080p) × 10 encodings = 20 GB
15K titles × 20 GB = 300 TB (simplified)
With replication, regional: Multi-PB
Total: 100+ PB
```

### 6.4 Bandwidth (CDN)

```
50M streams × 5 Mbps = 250 Tbps (theoretical max)
Actual: 15% of internet ≈ 10-20 Tbps
```

### 6.5 Architecture Implications

- **Open Connect:** Netflix CDN appliances in ISP datacenters
- **Pre-positioning:** Popular content cached at edge
- **Encoding:** Per-title optimization; multiple resolutions
- **Recommendation:** Separate service; batch + real-time

---

## 7. Instagram

### 7.1 Key Metrics

| Metric | Value | Notes |
|--------|-------|-------|
| **MAU** | 2B+ | |
| **DAU** | 500M+ | |
| **Photos uploaded/day** | 100M+ | |
| **Likes/day** | 1.4B+ | |
| **Stories** | 500M+ daily active | |
| **Reels** | Billions of plays/day | |

### 7.2 Traffic Breakdown

| Operation | QPS | Daily |
|-----------|-----|-------|
| Photo upload | 1,200+ | 100M |
| Like | 16,000+ | 1.4B |
| Feed read | 100K-500K | |
| Story view | 50K-200K | |
| Reels | 100K+ | |

### 7.3 Storage

```
100M photos × 3 MB = 300 TB/day (photos)
Stories (24h): 500M × 5 MB = 2.5 PB/day (ephemeral, lower retention)
With thumbnails, indexes: 2x
Annual: 100+ PB
```

### 7.4 Bandwidth

```
Upload: 1.2K × 3 MB = 3.6 GB/s
Feed: 300K × 100 KB = 30 GB/s
CDN for media: 50+ GB/s
```

### 7.5 Architecture Implications

- **Feed:** Fan-out on write; pre-compute for followers
- **Explore:** ML pipeline; batch + real-time
- **Media:** S3/Blob; CDN for delivery
- **Stories:** 24h TTL; separate pipeline

---

## 8. Google Search

### 8.1 Key Metrics

| Metric | Value | Notes |
|--------|-------|-------|
| **Queries per day** | 8.5B+ | |
| **Queries per second** | ~100,000 | Avg |
| **Peak QPS** | 200K-500K | |
| **Index size** | Trillions of pages | |
| **Index storage** | 100+ PB | Compressed |
| **Crawl** | Billions of pages/day | |

### 8.2 Traffic Breakdown

| Operation | QPS | Notes |
|-----------|-----|-------|
| Search query | 100K | |
| Autocomplete | 500K+ | Per keystroke |
| Image search | 20K+ | |
| Ads | 100K+ | Per search |

### 8.3 Storage

```
Index: Trillions of URLs × 1 KB = PB scale
Document store: 100+ PB
Logs, signals: 10+ PB
Total: 100+ PB
```

### 8.4 Architecture Implications

- **Index:** Distributed; sharded by document
- **Query:** Fan-out to shards; merge results
- **Caching:** Query result cache; 20-30% hit rate
- **Latency:** p99 < 100 ms target

---

## 9. Facebook / Meta

### 9.1 Key Metrics

| Metric | Value | Notes |
|--------|-------|-------|
| **MAU** | 3B+ | Facebook + Instagram + WhatsApp |
| **Photos uploaded** | 10M+/sec (peak) | |
| **Messenger messages** | 240B/day | |
| **Video views** | Billions/day | |
| **Storage** | 300+ PB | Photos, videos |
| **Hadoop cluster** | 100+ PB | Analytics |

### 9.2 Traffic Breakdown

| Operation | QPS | Daily |
|-----------|-----|-------|
| Photo upload | 10M peak | 2B+ |
| Messenger | 3M+ | 240B |
| Feed read | 1M+ | |
| Like/reaction | 500K+ | |
| Video | 100K+ | |

### 9.3 Storage

```
Photos: 2B × 2 MB = 4 PB/day (raw)
With replication (3x): 12 PB/day
Videos: 1B × 50 MB = 50 PB/day (raw)
Cold storage for old: Haystack, Blob storage
Total: 300+ PB
```

### 9.4 Architecture Implications

- **Haystack:** Custom object store for photos
- **TAO:** Graph store (Users, Objects, Associations)
- **Feed:** Multi-stage ranking; real-time + batch
- **Messenger:** Similar to WhatsApp; E2E for secret convos

---

## 9.5 Additional Systems (Quick Reference)

### Amazon (E-commerce)

| Metric | Value |
|--------|-------|
| Products | 350M+ |
| Orders/day | 10M+ |
| Prime members | 200M+ |
| QPS (peak) | 1M+ |
| Storage | 100+ PB |

### LinkedIn

| Metric | Value |
|--------|-------|
| Members | 900M+ |
| Jobs | 30M+ |
| Feed updates | 100K+ QPS |
| Storage | 50+ PB |

### Airbnb

| Metric | Value |
|--------|-------|
| Listings | 7M+ |
| Bookings | 1M+ per day |
| Search QPS | 10K-50K |
| Storage | 10+ PB |

---

## 10. Storage & Bandwidth Calculations

### 10.1 Storage Estimation Formula

```
Daily storage = (operations/day) × (bytes per operation) × (replication factor)
Annual storage = Daily × 365 × (1 + growth rate)
```

### 10.2 Bandwidth Estimation Formula

```
Read bandwidth = QPS × (bytes per response)
Write bandwidth = QPS × (bytes per request)
Total = Read + Write + (CDN/media multiplier)
```

### 10.3 QPS from DAU

```
QPS = DAU × (requests per user per day) / 86,400
Peak QPS = Avg QPS × (2 to 3)  // typical peak multiplier
```

### 10.4 Comparison Table (Summary)

| System | Write QPS | Read QPS | Storage (est.) | Bandwidth |
|--------|-----------|----------|----------------|------------|
| Twitter | 6K | 300K | 600 TB/yr | 3+ GB/s |
| YouTube | 1K (upload) | 1M+ (stream) | 1 EB | 10+ Tbps |
| Uber | 1.25M (location) | 500K | 75 TB (7d) | 10 GB/s |
| WhatsApp | 3.5M peak | 3.5M | 70 PB/yr | 5 GB/s |
| Netflix | Low | 50M concurrent | 100 PB | 20 Tbps |
| Instagram | 1.2K (photo) | 500K | 100 PB/yr | 50 GB/s |
| Google | 100K (crawl) | 100K | 100 PB | - |
| Facebook | 10M (photo) | 1M+ | 300 PB | 100+ GB/s |

---

## 11. Interview Discussion

### 11.1 How to Use These Numbers

1. **Sanity check:** "For a Twitter-scale timeline, we'd need ~300K read QPS—that's 300 Redis instances at 1K QPS each, or 30 at 10K."
2. **Sharding:** "100B messages/day = 1.2M QPS—shard by (sender_id, receiver_id) or message_id."
3. **Cache:** "300K read QPS × 10 KB = 3 GB/s—need 20+ Gbps bandwidth for cache misses."
4. **Storage:** "500M tweets × 300 bytes = 150 GB/day—manageable with 10 DB nodes."

### 11.2 Common Mistakes

- **Overestimating:** Start with DAU, not MAU; not all users active at once
- **Ignoring peak:** Use 2-3x average for peak QPS
- **Forgetting replication:** Storage × 3 for typical replication
- **Media vs metadata:** Video/photo dominates storage; metadata is small

### 11.3 Key Takeaways

| System | Key Number to Remember |
|--------|------------------------|
| Twitter | 6K write, 300K read QPS |
| YouTube | 500 hours/min upload, 1 EB storage |
| Uber | 1.25M location updates/sec |
| WhatsApp | 100B messages/day, 50 engineers |
| Netflix | 15% of internet traffic |
| Instagram | 100M photos/day |
| Google | 100K search QPS, trillions of pages |
| Facebook | 10M photos/sec, 300 PB |

### 11.4 Scaling Milestones

| Scale | QPS | Storage | Typical Architecture |
|-------|-----|---------|----------------------|
| Startup | 100-1K | 1-10 TB | Single DB, cache |
| Growth | 1K-10K | 10-100 TB | Read replicas, sharding |
| Scale | 10K-100K | 100 TB-1 PB | Distributed DB, CDN |
| Hyperscale | 100K-1M+ | 1+ PB | Multi-region, custom |

### 11.5 Data Growth Projections

When estimating for a new system, model growth:

```
Year 1: 1x (baseline)
Year 2: 2-3x (typical growth)
Year 3: 5-10x
Year 5: 20-50x

Example: 1 TB Year 1 → 10 TB Year 3 → 50 TB Year 5
```

---

*Document: Scale Reference — 450+ lines | FAANG Interview Essential*
