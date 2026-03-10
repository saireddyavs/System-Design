# Back-of-Envelope Estimation: Staff+ Engineer Deep Dive

## 1. Concept Overview

### Definition
**Back-of-envelope estimation** is the practice of making quick, approximate calculations using round numbers and simplified assumptions to reason about system scale, capacity, and feasibility. It uses "numbers every engineer should know" as mental anchors.

### Purpose
- **Feasibility check**: "Can we store 1 year of logs?" before building
- **Capacity planning**: "How many servers for 1M QPS?"
- **Cost estimation**: "What will this cost at scale?"
- **Interview performance**: Demonstrate quantitative thinking in system design
- **Architecture validation**: Sanity-check designs with numbers

### Why It Exists
- **Speed**: Full analysis takes hours; back-of-envelope takes minutes
- **Communication**: Stakeholders need ballpark numbers, not precision
- **Decision making**: "Order of magnitude" often sufficient (10x vs 100x)
- **Interview standard**: FAANG interviews expect this skill

### Problems It Solves
1. **Over-engineering**: Building for 100M users when you have 10K
2. **Under-provisioning**: Not realizing 1M QPS needs 100+ servers
3. **Wrong assumptions**: Thinking SSD is same speed as RAM
4. **Unclear requirements**: Quantifying "high scale" or "low latency"

---

## 2. Real-World Motivation

### Google SRE Book
- **Latency numbers**: Widely cited table (L1 cache 0.5ns to cross-continent 150ms)
- **Power of 2**: Essential for quick math
- **Use**: Capacity planning, incident response, design review

### Amazon
- **6-pager**: Design docs include capacity and cost estimates
- **Working backwards**: Start with customer need, derive requirements
- **Numbers**: "Prime Day will need 10x normal capacity"

### Netflix
- **Streaming**: 15% of internet; bandwidth calculations critical
- **Open Connect**: 17K servers; storage and bandwidth per server

### Meta/Facebook
- **Scale**: 3B users; storage for photos, videos, messages
- **Real-time**: Chat, feed; latency and throughput requirements

### Stripe
- **Payments**: Millions TPS; consistency and latency
- **Compliance**: Audit trail storage, retention

---

## 3. Architecture Diagrams

### Latency Hierarchy (Visual)

```
LATENCY PYRAMID (each level ~100-1000x slower)
==============================================

                    L1 cache 0.5 ns     █
                   L2 cache 7 ns        ██
                  L3 cache 20 ns        ███
                 Main memory 100 ns     ████
                SSD 100 μs              ████████████████████████████████
               HDD 10 ms                ████████████████████████████████████████████
              Same DC 500 μs             ████████████████████████████████████████████████████
             Cross-continent 150 ms      ████████████████████████████████████████████████████████████████
```

### Request Latency Breakdown

```
TYPICAL WEB REQUEST (100ms target)
==================================
┌─────────────────────────────────────────────────────────────┐
│ DNS lookup            │ 1-5 ms     │ (cache: 0)             │
│ TCP connection        │ 1-50 ms    │ (keep-alive: 0)        │
│ TLS handshake         │ 10-50 ms   │ (resume: 1ms)          │
│ Request send           │ <1 ms      │                        │
│ Server processing      │ 10-100 ms  │ (DB, cache, compute)   │
│ Response transfer      │ 1-50 ms    │ (depends on size)      │
└─────────────────────────────────────────────────────────────┘
Total: 25-250 ms typical
```

### Storage Estimation Flow

```
STORAGE ESTIMATION FRAMEWORK
============================
    Users/Events
         │
         ▼
    Data per user/event (bytes)
         │
         ▼
    Retention period
         │
         ▼
    Replication factor
         │
         ▼
    Total storage = Users × Size × Retention × Replication
```

---

## 4. Core Mechanics

### Estimation Principles
1. **Round numbers**: Use 10, 100, 1000, 1M - not 37 or 1,847
2. **Power of 2**: 2^10=1K, 2^20=1M, 2^30=1G - for bytes, memory
3. **Order of magnitude**: 10x precision often enough
4. **State assumptions**: "Assume 1KB per request" - document it
5. **Sanity check**: Compare to known systems (Twitter, YouTube)

### Calculation Framework
1. **Clarify**: What are we estimating? (QPS, storage, bandwidth, cost)
2. **Decompose**: Break into factors (users × actions × size)
3. **Estimate each**: Use known numbers, round
4. **Multiply/Add**: Combine factors
5. **Validate**: Does it make sense? Compare to references

### Unit Conversions
- 1 day = 86,400 seconds ≈ 10^5 sec
- 1 month ≈ 2.6 × 10^6 sec ≈ 30 days
- 1 year ≈ 3.15 × 10^7 sec ≈ π × 10^7
- 1 Gbps = 10^9 bits/sec = 125 MB/s
- 1 TB = 10^12 bytes = 1000 GB

---

## 5. Numbers

### Latency Numbers Every Engineer Should Know

| Operation | Time | Human Scale |
|-----------|------|-------------|
| L1 cache reference | 0.5 ns | |
| L2 cache reference | 7 ns | |
| L3 cache reference | 20 ns | |
| Main memory reference | 100 ns | |
| SSD random read | 100 μs (0.1 ms) | |
| SSD sequential read | 200 MB/s | |
| HDD seek | 10 ms | |
| HDD sequential read | 100 MB/s | |
| Round trip same datacenter | 500 μs (0.5 ms) | |
| Round trip cross-datacenter (same region) | 1-5 ms | |
| Round trip cross-continent | 150 ms | |
| Packet round trip (same continent) | 50-100 ms | |

### Power of 2 Reference Table

| 2^n | Value | Approx | Use |
|-----|-------|--------|-----|
| 2^10 | 1,024 | 10^3 | 1 KB |
| 2^20 | 1,048,576 | 10^6 | 1 MB |
| 2^30 | 1,073,741,824 | 10^9 | 1 GB |
| 2^40 | 1.1 × 10^12 | 10^12 | 1 TB |
| 2^50 | 1.1 × 10^15 | 10^15 | 1 PB |
| 2^60 | 1.2 × 10^18 | 10^18 | 1 EB |
| 2^32 | 4.3 × 10^9 | 4B | IPv4, 32-bit |
| 2^64 | 1.8 × 10^19 | 10^19 | 64-bit |

### Throughput Numbers (Typical)

| System | Read QPS | Write QPS | Notes |
|--------|----------|-----------|-------|
| Redis (single) | 100K-500K | 80K-100K | In-memory |
| Memcached | 100K-1M | 100K-1M | In-memory |
| PostgreSQL | 10K-50K | 5K-20K | Single node |
| MySQL | 50K-100K | 10K-50K | Single node |
| Cassandra | 100K+ | 100K+ | Per node |
| DynamoDB | 3K RCU/partition | 1K WCU/partition | Per partition |
| Kafka | 100K-1M msg/s | Per broker | Batching |
| Elasticsearch | 10K-50K | 5K-20K | Depends on query |
| S3 | 3,500 PUT/s | 5,500 GET/s | Per prefix |

### Single Server Capacity (Rough)

| Metric | Small | Medium | Large |
|--------|-------|--------|-------|
| CPU | 4 core | 16 core | 64 core |
| RAM | 8 GB | 64 GB | 256 GB |
| Disk | 100 GB SSD | 1 TB SSD | 4 TB NVMe |
| Network | 1 Gbps | 10 Gbps | 25 Gbps |
| Requests/sec (API) | 500-2K | 5K-20K | 20K-100K |

---

## 6. Tradeoffs

### Precision vs Speed

| Approach | Time | Accuracy | Use Case |
|----------|------|----------|----------|
| Back-of-envelope | Minutes | 2-10x | Interview, feasibility |
| Spreadsheet model | Hours | 1.5-2x | Planning |
| Load test | Days | 1.2x | Pre-launch |
| Production metrics | Continuous | Exact | Operations |

### Assumption Sensitivity
- **High impact**: User count, retention, data size - get these right
- **Medium**: Replication factor, compression ratio
- **Low**: Exact byte sizes - round to 1KB, 1MB

---

## 7. Variants / Implementations

### Estimation Types

1. **Capacity**: "How many servers for X QPS?"
2. **Storage**: "How much for 1 year of data?"
3. **Bandwidth**: "What's the network requirement?"
4. **Cost**: "What's the monthly cloud bill?"
5. **Latency**: "Can we hit 100ms p99?"

### Industry Benchmarks (Reference)

| Metric | Small | Medium | Large |
|--------|-------|--------|-------|
| DAU/MAU ratio | 0.1-0.2 | 0.2-0.4 | 0.4-0.6 |
| Requests per DAU | 10-50 | 50-200 | 200-1000 |
| Peak/avg ratio | 2-3x | 3-5x | 5-10x |
| Cache hit rate | 80% | 90% | 95%+ |

---

## 8. Scaling Strategies (in Estimation)

### Scaling Factors
- **Linear**: Double users → double servers (stateless)
- **Sub-linear**: Caching reduces backend load (90% hit → 10% to DB)
- **Super-linear**: Hot spots, thundering herd (need over-provisioning)

### Rule of Thumb
- **Database**: Often bottleneck; 1 DB ~ 10K-50K writes/sec
- **Cache**: 1 Redis ~ 100K QPS
- **App server**: 1 server ~ 1K-10K RPS (depends on logic)
- **Bandwidth**: 1 Gbps ≈ 125 MB/s ≈ 10^6 small requests/s (1KB each)

---

## 9. Failure Scenarios (Estimation Impact)

### Common Estimation Errors
- **Off by 10x**: Forgot replication, retention, or peak factor
- **Wrong unit**: Bytes vs bits, seconds vs milliseconds
- **Ignored factor**: Compression, deduplication, caching
- **Wrong baseline**: Used dev numbers for production

### Sanity Checks
- Compare to known systems (Twitter, YouTube public data)
- Cross-check: storage = QPS × size × retention
- Reasonableness: 1B users × 1KB/day = 1 TB/day (plausible)

---

## 10. Performance Considerations

### Latency Budget Allocation
- **Total budget**: e.g., 100ms p99
- **Breakdown**: 20ms DB, 10ms cache, 5ms network, 65ms app logic
- **Margin**: Leave 20-30% for variance

### Bottleneck Identification
- **CPU**: High QPS, compute-heavy
- **Memory**: Large working set, caching
- **Disk I/O**: Database, logs
- **Network**: Large payloads, cross-region

---

## 11. Use Cases (Worked Examples)

### Example 1: Twitter Storage per Day

**Assumptions:**
- 500M DAU (daily active users)
- 10 tweets per user per day = 5B tweets/day
- 280 chars per tweet (max) + metadata ≈ 500 bytes
- Media: 20% of tweets have images, 1MB average = 1B × 1MB = 1 PB (images)
- Text: 5B × 500 bytes = 2.5 TB

**Calculation:**
- Text: 5 × 10^9 × 500 = 2.5 × 10^12 bytes = 2.5 TB/day
- Images: 1 × 10^9 × 1 × 10^6 = 1 × 10^15 bytes = 1 PB/day (if 20% have 1MB image)
- Refined: 20% × 5B = 1B with media; 1B × 200KB avg = 200 TB (more realistic)
- **Total: ~200-500 TB/day** (text + compressed images + indexes)

### Example 2: YouTube Bandwidth

**Assumptions:**
- 2B users, 1B watch videos/month
- Average video: 10 min, 5 Mbps bitrate (1080p)
- Data per view: 10 min × 60 sec × 5 Mbps / 8 = 375 MB per view

**Calculation:**
- 1B views/month × 375 MB = 375 PB/month
- Per day: 375/30 ≈ 12.5 PB/day
- Per second (peak, 5x avg): 12.5 × 10^15 / (24 × 3600) × 5 ≈ 700 TB/s peak? 
- More realistic: 15% of internet = ~50 Tbps global; YouTube share ~50 Tbps × 0.15 ≈ 7.5 Tbps
- **Peak bandwidth: ~5-10 Tbps** (order of magnitude)

### Example 3: Uber Rides QPS

**Assumptions:**
- 25M rides per day globally
- Peak hour = 10% of daily (2.5M rides)
- Peak hour = 3600 seconds
- Requests: ride request, ETA, matching, etc. - ~10 requests per ride

**Calculation:**
- Rides per second (peak): 2.5 × 10^6 / 3600 ≈ 700 rides/sec
- Total requests (10x): 7,000 QPS
- With 5x peak/avg: 35,000 QPS peak
- **Estimate: 10K-50K QPS** for ride-related APIs

### Example 4: WhatsApp Messages per Second

**Assumptions:**
- 2B users, 100B messages per day
- Messages per second: 100 × 10^9 / 86400 ≈ 1.15 × 10^6 msg/sec
- Peak (3x): ~3.5 × 10^6 msg/sec

**Calculation:**
- **~1-3 million messages per second** (average to peak)

### Example 5: Netflix Streaming QPS

**Assumptions:**
- 250M subscribers
- 10% concurrent during peak = 25M streams
- Each stream: 1 initial request + keep-alive every 10 sec
- Requests: 25M / 10 = 2.5M QPS (just keep-alive)
- Initial + metadata: add 20% = 3M QPS
- **Estimate: 2-5M QPS** for streaming API (video data is CDN, not API)

### Example 6: Server Count for 1M QPS

**Assumptions:**
- 1M QPS total
- Each server: 5K RPS (moderate logic, DB calls)
- Cache hit rate: 90% (reduces DB load)
- Effective backend: 100K QPS to DB
- DB: 20K QPS per instance → 5 DB instances (with replicas)
- App servers: 1M / 5K = 200 servers
- **Estimate: 200-300 app servers, 5-10 DB instances** (with replicas)

---

## 12. Comparison Tables

### Latency Comparison (Same Operation)

| Storage Type | Read (random) | Read (sequential) | Write |
|--------------|---------------|---------------------|-------|
| L1 cache | 0.5 ns | - | - |
| L2 cache | 7 ns | - | - |
| RAM | 100 ns | 100 GB/s | 100 GB/s |
| SSD | 100 μs | 500 MB/s | 400 MB/s |
| HDD | 10 ms | 100 MB/s | 100 MB/s |
| Network (DC) | 500 μs RTT | 10 Gbps | 10 Gbps |

### Data Size Reference

| Item | Size | Notes |
|------|------|-------|
| UUID | 16 bytes | 36 chars as string |
| IPv4 | 4 bytes | |
| IPv6 | 16 bytes | |
| Tweet (text) | 280 bytes | Max 280 chars |
| Tweet (with metadata) | 500 bytes - 1 KB | |
| Email | 10-100 KB | |
| Web page | 100 KB - 2 MB | |
| Image (thumbnail) | 10-50 KB | |
| Image (full) | 100 KB - 5 MB | |
| 1 min video (720p) | 50-100 MB | |
| 1 min video (1080p) | 100-200 MB | |
| 1 min audio (MP3) | 1 MB | |

### Time Reference

| Duration | Seconds | Use |
|----------|---------|-----|
| 1 second | 1 | |
| 1 minute | 60 | |
| 1 hour | 3,600 | |
| 1 day | 86,400 ≈ 10^5 | |
| 1 week | 604,800 | |
| 1 month | ~2.6 × 10^6 | 30 days |
| 1 year | ~3.15 × 10^7 | π × 10^7 |

---

## 13. Code or Pseudocode

### Estimation Helper Functions

```python
def power_of_2(n):
    """2^n for common values"""
    return {10: 1024, 20: 1024**2, 30: 1024**3, 40: 1024**4}[n]

def bytes_to_human(b):
    """Convert bytes to human readable"""
    for unit in ['B', 'KB', 'MB', 'GB', 'TB', 'PB']:
        if b < 1024:
            return f"{b:.1f} {unit}"
        b /= 1024
    return f"{b:.1f} EB"

def qps_to_servers(qps, rps_per_server=5000):
    """Estimate server count for QPS"""
    return (qps + rps_per_server - 1) // rps_per_server

def storage_estimate(users, data_per_user, retention_days, replication=3):
    """Estimate total storage"""
    return users * data_per_user * retention_days * replication
```

### Worked Example (Python)

```python
def estimate_twitter_storage():
    """Twitter daily storage estimate"""
    tweets_per_day = 500_000_000  # 500M DAU × 10 tweets
    bytes_per_tweet = 500  # text + metadata
    text_storage = tweets_per_day * bytes_per_tweet  # 250 GB
    media_tweets = tweets_per_day * 0.2  # 20% with media
    media_size = media_tweets * 200_000  # 200 KB avg
    media_storage = media_size  # 20 TB
    total = text_storage + media_storage
    return total  # ~20 TB/day (conservative)
```

---

## 14. Interview Discussion

### How to Approach Estimation Questions
1. **Clarify**: "Are we estimating for Twitter scale or a startup?"
2. **State assumptions**: "Assume 500M DAU, 10 tweets per user per day"
3. **Decompose**: "Storage = tweets × size × retention"
4. **Calculate**: Use round numbers, show work
5. **Sanity check**: "That's 20 TB/day - YouTube does 100+ PB, so plausible"

### Key Numbers to Memorize
- **Latency**: L1 0.5ns, RAM 100ns, SSD 100μs, HDD 10ms, DC 500μs, cross-continent 150ms
- **Power of 2**: 2^10=1K, 2^20=1M, 2^30=1G, 2^40=1T
- **Time**: 1 day = 10^5 sec, 1 year = π×10^7 sec
- **Throughput**: Redis 100K QPS, PostgreSQL 10K writes, 1 server 5K RPS

### Estimation Framework (Verbal)
1. "Let's estimate X. I'll assume A, B, C."
2. "The formula is: [decompose]"
3. "So we have: [calculation with round numbers]"
4. "That gives us approximately [result]"
5. "As a sanity check, [compare to known system]"

### Common Estimation Questions
- "Estimate Twitter's storage per day"
- "How many servers for 1M QPS?"
- "Estimate YouTube's bandwidth"
- "How much does it cost to store 1 PB?"
- "Estimate WhatsApp messages per second"
- "Design a URL shortener - how many URLs can you store?"

### Follow-Up Questions
- "What if we have 10x more users?"
- "Where's the bottleneck?"
- "How would you reduce the storage?"
- "What's the cost of this design?"

### Common Mistakes
- Wrong units (bytes vs bits, sec vs ms)
- Forgetting replication, retention, or peak factor
- Unrealistic assumptions (100% cache hit)
- No sanity check against known systems
- Over-precision (47.3 servers → say "~50")

---

## 15. Additional Worked Examples

### Example 7: URL Shortener Capacity

**Assumptions:**
- 6-character alphanumeric short URL (a-z, A-Z, 0-9) = 62^6 = 56.8 billion combinations
- 7 characters = 62^7 = 3.5 trillion
- 100M URLs created per year
- 10-year retention

**Calculation:**
- 62^6 ≈ 57 billion unique URLs
- At 100M/year, lasts 570 years
- Storage: 100M × (6 bytes short + 200 bytes long URL + 50 bytes metadata) ≈ 25 GB/year
- **Conclusion: 6 chars sufficient; ~25 GB/year storage**

### Example 8: Cost to Store 1 PB

**Assumptions:**
- AWS S3: $0.023/GB/month (standard)
- 1 PB = 1000 TB = 1,000,000 GB

**Calculation:**
- 1,000,000 GB × $0.023 = $23,000/month
- Glacier (archive): $0.004/GB = $4,000/month
- **Range: $4K-$23K/month depending on tier**

### Example 9: Bandwidth for 10M Concurrent Video Streams

**Assumptions:**
- 10M concurrent viewers
- 2 Mbps average bitrate (720p)
- Total: 10M × 2 Mbps = 20,000 Gbps = 20 Tbps

**Calculation:**
- 20 Tbps = 20,000 Gbps
- With CDN (80% cache hit at edge): 20% × 20 Tbps = 4 Tbps origin
- **Origin bandwidth: 4 Tbps; Edge: 20 Tbps total**

### Example 10: Database Connections for 1000 App Servers

**Assumptions:**
- 1000 app servers
- 100 connections per server (pool)
- Total: 100,000 connections

**Calculation:**
- PostgreSQL max_connections default: 100
- Need: 100,000 / 100 = 1000 DB instances (wrong!)
- Solution: Connection pooling (PgBouncer) - 1000 servers × 10 connections = 10,000
- Pooler maintains 10K to DB, multiplexes 100K from app
- **With pooling: 10-20 DB instances sufficient**

---

## 16. Formulas Reference Card

### Storage Formulas
```
Total Storage = Records × Bytes_per_record × Retention_days × Replication_factor
Compressed = Raw × Compression_ratio  (typically 3-10x for text, 2x for images)
```

### Throughput Formulas
```
QPS = Requests_per_user × DAU / 86400
Peak_QPS = Average_QPS × Peak_factor  (typically 2-5x)
Servers = QPS / (RPS_per_server × (1 - Cache_hit_rate))
```

### Bandwidth Formulas
```
Bandwidth_bps = Concurrent_users × Bitrate_per_user
Bandwidth_Bps = Bandwidth_bps / 8
Monthly_data = Bandwidth_Bps × 86400 × 30
```

### Latency Formulas
```
End_to_end = DNS + TCP + TLS + Request + Processing + Response
Processing = Cache_lookup + DB_query + Serialization
DB_query = Disk_seek + Data_transfer  (or Memory_lookup for cache)
```

---

## 17. Cloud Cost Reference (Approximate 2024)

| Resource | AWS | GCP | Azure |
|----------|-----|-----|-------|
| EC2/Compute (large) | $0.10/hr | $0.09/hr | $0.10/hr |
| S3 storage | $0.023/GB/mo | $0.02/GB/mo | $0.018/GB/mo |
| RDS (db.r5.large) | $0.24/hr | - | - |
| DynamoDB (on-demand) | $1.25/million writes | - | - |
| Data transfer out | $0.09/GB | $0.12/GB | $0.087/GB |
| Redis (cache.r5.large) | $0.119/hr | - | - |

### Cost Estimation Example
- 100 servers × $0.10/hr × 720 hr = $7,200/month compute
- 100 TB storage × $23/TB = $2,300/month
- 50 TB egress × $90/TB = $4,500/month
- **Total: ~$14,000/month** for medium-scale system

---

## 18. Quick Reference: Numbers to Memorize

### The "1" Rule
- 1 ns = 1 billionth second
- 1 μs = 1 millionth second (1000 ns)
- 1 ms = 1 thousandth second (1000 μs)
- 1 second = 1000 ms

### Ratios
- 1 ms : 1 μs = 1000 : 1
- 1 μs : 1 ns = 1000 : 1
- L1 to RAM = 200x (0.5ns to 100ns)
- RAM to SSD = 1000x (100ns to 100μs)
- SSD to HDD = 100x (100μs to 10ms)

### Approximations
- π ≈ 3.14 (useful: π × 10^7 ≈ 1 year in seconds)
- e ≈ 2.72
- 2^10 ≈ 1000
- 365 ≈ 400 (for rough yearly estimates)
- 1 day ≈ 10^5 seconds (86,400)
