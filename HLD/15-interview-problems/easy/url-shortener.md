# System Design: URL Shortener (TinyURL)

## 1. Problem Statement & Requirements

### Functional Requirements

- **Shorten URLs**: Convert long URLs into short, shareable links (e.g., `https://tiny.com/abc123`)
- **Redirect**: When users visit a short URL, redirect them to the original long URL
- **Custom Aliases**: Allow users to specify custom short codes (e.g., `tiny.com/mysite`)
- **Expiry**: Support optional expiration for short URLs (some links expire after N days)
- **Analytics**: Track click counts, referrers, geographic distribution, and timestamps
- **User Accounts** (optional): Link URLs to user accounts for management

### Non-Functional Requirements

- **Scale**: 100M URLs created per day; 10:1 read/write ratio → 1B reads/day
- **Latency**: Redirect should complete in < 100ms (p99)
- **Availability**: 99.99% uptime (redirect is critical for user experience)
- **Consistency**: Eventual consistency acceptable for analytics; strong consistency for redirect
- **Uniqueness**: Short codes must be globally unique; collision prevention is critical

### Out of Scope

- User authentication/authorization (assume basic)
- Advanced analytics dashboards (basic counts only)
- URL preview/metadata
- Malware/phishing detection
- Bulk import/export

---

## 2. Back-of-Envelope Estimation

### Traffic Estimates

| Metric | Value | Calculation |
|--------|-------|-------------|
| Writes per day | 100M | Given |
| Reads per day | 1B | 10:1 read/write ratio |
| Writes per second | ~1,200 | 100M / 86400 |
| Reads per second | ~12,000 | 1B / 86400 |
| Peak QPS (3x) | ~36,000 reads | Assume 3x average for burst |

### Storage Estimates

- **Assumptions**: 
  - Short code: 7 chars (base62) = ~7 bytes
  - Long URL: avg 500 bytes
  - Metadata: 200 bytes (created_at, user_id, expiry, etc.)
  - Analytics: ~50 bytes per click (optional, separate storage)

- **Per URL**: ~700 bytes
- **100M URLs/day × 5 years retention**: ~100M × 365 × 5 × 700 bytes ≈ 128 TB (raw)
- **With indexes**: ~200 TB
- **Analytics**: 1B clicks/day × 50 bytes × 365 days ≈ 18 TB/year

### Bandwidth Estimates

- **Read**: 12,000 QPS × 500 bytes (redirect response) ≈ 6 MB/s ≈ 518 GB/day
- **Write**: 1,200 QPS × 700 bytes ≈ 0.8 MB/s ≈ 70 GB/day
- **Total**: ~600 GB/day

### Cache Estimates

- **Hot URLs**: ~20% of URLs get 80% of traffic (Pareto)
- **Cache hit ratio target**: 80%
- **Hot URLs**: ~20M × 700 bytes ≈ 14 GB
- **Redis memory**: 20-30 GB for URL cache + overhead

---

## 3. API Design

### REST API Endpoints

#### Create Short URL

```
POST /api/v1/shorten
```

**Request:**
```json
{
  "long_url": "https://www.example.com/very/long/path?query=params",
  "custom_alias": "mysite",           // optional
  "expires_in_days": 30,              // optional, null = never expires
  "user_id": "user_123"               // optional
}
```

**Response (201 Created):**
```json
{
  "short_url": "https://tiny.com/abc123",
  "short_code": "abc123",
  "long_url": "https://www.example.com/very/long/path?query=params",
  "expires_at": "2025-04-10T00:00:00Z",
  "created_at": "2025-03-10T12:00:00Z"
}
```

**Error Response (409 Conflict):**
```json
{
  "error": "ALIAS_ALREADY_EXISTS",
  "message": "Custom alias 'mysite' is already taken"
}
```

#### Get Original URL (Internal)

```
GET /api/v1/internal/resolve/{short_code}
```

**Response (200 OK):**
```json
{
  "long_url": "https://www.example.com/very/long/path",
  "expires_at": "2025-04-10T00:00:00Z",
  "is_active": true
}
```

**Response (404 Not Found):**
```json
{
  "error": "URL_NOT_FOUND",
  "message": "Short code does not exist or has expired"
}
```

#### Redirect (Public)

```
GET /{short_code}
```

**Response (301 Moved Permanently or 302 Found):**
```
Location: https://www.example.com/very/long/path
```

#### Get Analytics

```
GET /api/v1/analytics/{short_code}
```

**Response (200 OK):**
```json
{
  "short_code": "abc123",
  "long_url": "https://www.example.com/...",
  "total_clicks": 15420,
  "clicks_by_day": {"2025-03-09": 1200, "2025-03-10": 800},
  "top_referrers": ["google.com", "twitter.com"],
  "top_countries": ["US", "IN", "UK"]
}
```

#### Delete Short URL

```
DELETE /api/v1/shorten/{short_code}
```

**Response (204 No Content)**

---

## 4. Data Model / Database Schema

### Database Choice

**Primary**: **Cassandra** or **DynamoDB** (NoSQL)

**Justification**:
- High write throughput (100M/day)
- Simple key-value access pattern (short_code → long_url)
- Horizontal scaling via partitioning
- No complex joins needed
- Optional: **PostgreSQL** for smaller scale (simpler, ACID for custom aliases)

### Schema (Cassandra)

```sql
-- URLs table (partition by short_code)
CREATE TABLE urls (
    short_code VARCHAR PRIMARY KEY,
    long_url TEXT,
    user_id VARCHAR,
    created_at TIMESTAMP,
    expires_at TIMESTAMP,
    expires_in_days INT,
    is_custom_alias BOOLEAN,
    metadata MAP<VARCHAR, TEXT>
);

-- Custom alias lookup (for uniqueness check)
CREATE TABLE custom_aliases (
    alias VARCHAR PRIMARY KEY,
    short_code VARCHAR,
    created_at TIMESTAMP
);

-- User's URLs (for listing user's shortened URLs)
CREATE TABLE urls_by_user (
    user_id VARCHAR,
    created_at TIMESTAMP,
    short_code VARCHAR,
    long_url TEXT,
    PRIMARY KEY (user_id, created_at)
) WITH CLUSTERING ORDER BY (created_at DESC);

-- Analytics (optional, separate table, time-series)
CREATE TABLE click_analytics (
    short_code VARCHAR,
    date DATE,
    hour INT,
    timestamp TIMESTAMP,
    referrer VARCHAR,
    country VARCHAR,
    user_agent VARCHAR,
    PRIMARY KEY ((short_code, date), hour, timestamp)
);
```

### Alternative: PostgreSQL Schema

```sql
CREATE TABLE urls (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(10) UNIQUE NOT NULL,
    long_url TEXT NOT NULL,
    user_id VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    is_custom_alias BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_urls_short_code ON urls(short_code);
CREATE INDEX idx_urls_user_id ON urls(user_id);
CREATE INDEX idx_urls_expires_at ON urls(expires_at) WHERE expires_at IS NOT NULL;
```

---

## 5. High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              CLIENT REQUEST                                       │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           CDN (CloudFront / Cloudflare)                           │
│                    Cache 301/302 redirects for popular URLs                        │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         LOAD BALANCER (Application LB)                             │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
                    ┌───────────────────┼───────────────────┐
                    ▼                   ▼                   ▼
┌───────────────────────┐   ┌───────────────────────┐   ┌───────────────────────┐
│   API Server 1        │   │   API Server 2        │   │   API Server N        │
│   - Create Short URL  │   │   - Redirect Handler   │   │   - Analytics         │
└───────────────────────┘   └───────────────────────┘   └───────────────────────┘
                    │                   │                   │
                    └───────────────────┼───────────────────┘
                                        │
                    ┌───────────────────┼───────────────────┐
                    ▼                   ▼                   ▼
┌───────────────────────┐   ┌───────────────────────┐   ┌───────────────────────┐
│   Redis Cache         │   │   Key Generation      │   │   Cassandra/DB        │
│   (short_code→URL)    │   │   Service (KGS)        │   │   (Primary Storage)    │
└───────────────────────┘   └───────────────────────┘   └───────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│   Analytics Service (Async) → Kafka → Click Processor → Time-Series DB/Data Lake  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 6. Detailed Component Design

### 6.1 Short Code Generation

**Options:**

1. **Base62 Encoding of Auto-Increment ID**
   - Pros: Short, predictable length, no collisions
   - Cons: Sequential IDs reveal growth; need distributed ID generator

2. **Hash + Truncation (MD5/SHA)**
   - Input: long_url + timestamp
   - Output: First 7 chars of base62(MD5(input))
   - Pros: No central service
   - Cons: Collision risk (birthday paradox); need collision handling

3. **Key Generation Service (KGS)**
   - Pre-generate keys in batches (e.g., 1M keys)
   - Store in Redis/DB; API servers fetch keys on demand
   - Pros: No collisions, fast, no computation
   - Cons: Single point of failure (mitigate with standby KGS)

**Recommended: KGS + Base62**

```python
# Base62 encoding (0-9, a-z, A-Z = 62 chars)
CHARS = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

def base62_encode(num):
    if num == 0: return "0"
    result = []
    while num:
        result.append(CHARS[num % 62])
        num //= 62
    return ''.join(reversed(result))

# 7 chars = 62^7 ≈ 3.5 trillion unique codes
```

### 6.2 Redirect Handler

**Flow:**
1. User requests `GET /abc123`
2. Check Redis cache first (hot path)
3. If miss, query DB
4. If found, return 301 or 302
5. If analytics enabled, async fire event to Kafka

**301 vs 302:**
- **301 (Permanent)**: Browser caches redirect; reduces server load; SEO passes to original URL
- **302 (Temporary)**: No caching; every request hits server; analytics accurate; can change destination later

**Recommendation**: Use **302** for analytics; **301** for static, never-changing URLs.

### 6.3 Key Generation Service (KGS)

- Pre-generates 1M keys, stores in Redis list
- API servers call `LPOP` to get key
- When pool < 100K, KGS generates 1M more
- KGS runs in primary + standby; uses ZooKeeper/etcd for leader election

### 6.4 Custom Alias Handling

- Check `custom_aliases` table for uniqueness
- Use conditional write (CAS) to prevent race
- If collision, return 409 to user

### 6.5 Analytics Pipeline

- Async: Fire-and-forget event to Kafka on redirect
- Consumer: Aggregate by short_code, date, hour
- Store in Time-Series DB (InfluxDB, TimescaleDB) or Data Lake
- Batch aggregation for daily/weekly reports

### 6.6 Expiry Handling

- **Lazy deletion**: On redirect, check `expires_at`; if expired, return 404 and optionally delete
- **Background job**: Cron job deletes expired URLs daily (e.g., `DELETE FROM urls WHERE expires_at < NOW()`)

---

## 7. Scaling

### Horizontal Scaling

- **API Servers**: Stateless; scale via load balancer
- **KGS**: Leader-follower; scale followers for key generation throughput
- **Database**: Cassandra shards by short_code (partition key)

### Sharding Strategy

- **Cassandra**: Partition by `short_code`; natural distribution
- **PostgreSQL**: Shard by `short_code` hash (e.g., `hash(short_code) % 10`)

### Caching Strategy

- **Redis**: Cache hit = 80%+; TTL = 24h for non-expiring; shorter for expiring
- **Cache keys**: `url:{short_code}` → `{long_url, expires_at}`
- **Cache invalidation**: On delete (remove key); on update (rare)

### CDN Usage

- Cache 301/302 at edge for top 1% of URLs
- Cache-Control: `max-age=86400` for 301
- CDN reduces origin load by ~20-30% for popular URLs

### Rate Limiting

- Per user: 100 creates/minute
- Per IP: 1000 redirects/minute (prevent abuse)

---

## 8. Failure Handling

| Component | Failure | Mitigation |
|-----------|---------|------------|
| Redis | Down | Fallback to DB; slower but works; repopulate cache on recovery |
| KGS | Down | Standby KGS; fallback to hash-based generation |
| Cassandra | Node down | Replication (RF=3); hinted handoff |
| API Server | Down | Load balancer removes; no single point |
| Kafka | Down | Buffer events in memory; retry; or drop analytics (non-critical) |

### Redundancy

- Redis: Cluster mode (3+ nodes)
- Cassandra: RF=3, multi-AZ
- KGS: Active-standby with failover

### Recovery

- Cache miss → DB hit → repopulate cache
- KGS key pool exhaustion → Alert; scale KGS; temporary hash fallback

---

## 9. Monitoring & Observability

### Key Metrics

| Metric | Target | Alert |
|--------|--------|-------|
| Redirect latency | p99 < 100ms | Critical |
| Create latency | p99 < 200ms | Warning |
| Cache hit rate | > 80% | Warning |
| Error rate | < 0.1% | Critical |
| KGS pool size | > 100K | Warning |
| DB connection pool | < 80% | Warning |

### Dashboards

- QPS (reads/writes)
- Latency percentiles (p50, p95, p99)
- Cache hit rate
- Error rate by endpoint
- Top URLs by traffic

### Alerting

- Redirect latency > 200ms p99
- Error rate > 1%
- KGS pool exhaustion
- Database connection errors

---

## 10. Interview Tips

### Common Follow-Up Questions

1. **How do you handle collisions?** Hash + retry with salt; or KGS for collision-free
2. **Why 301 vs 302?** 301 for SEO/cache; 302 for analytics/flexibility
3. **How to scale to 10B URLs?** Sharding, KGS, caching, CDN
4. **How to prevent abuse?** Rate limiting, CAPTCHA for create, blocklist for malicious URLs
5. **How to generate 7-char codes?** Base62 of 43 bits gives 62^7 ≈ 3.5T

### What Interviewers Look For

- Clear requirements gathering
- Back-of-envelope math
- Trade-off discussion (301 vs 302, SQL vs NoSQL)
- Scalability thinking (caching, sharding)
- Failure handling
- API design completeness

### Common Mistakes

- Skipping collision handling
- Choosing SQL without justification for scale
- Ignoring analytics (often asked)
- Not discussing caching
- Overcomplicating (e.g., microservices for MVP)

---

## Appendix: Additional Design Considerations

### A. Hash-Based Collision Handling

When using MD5/SHA + truncation:

```python
def generate_short_code(long_url, max_retries=5):
    for i in range(max_retries):
        salt = str(uuid.uuid4()) if i > 0 else ""
        hash_input = f"{long_url}{salt}{time.time()}"
        hash_hex = hashlib.md5(hash_input.encode()).hexdigest()
        # Convert to base62 and take first 7 chars
        num = int(hash_hex[:12], 16)
        code = base62_encode(num)[:7]
        if not db.exists(code):
            return code
    raise CollisionError("Could not generate unique code")
```

### B. Rate Limiting Implementation

- Use Redis INCR with TTL for sliding window
- Key: `ratelimit:create:{user_id}:{minute}`
- Limit: 100/minute per user

### C. Security Considerations

- **Blocklist**: Maintain list of blocked domains (phishing, malware)
- **Input validation**: Max URL length 2048 chars; sanitize custom aliases
- **CORS**: Restrict API to allowed origins

### D. Database Choice Deep Dive

**Why NoSQL (Cassandra/DynamoDB) over PostgreSQL?**

| Factor | PostgreSQL | Cassandra/DynamoDB |
|--------|------------|-------------------|
| Write throughput | Limited by single node | Horizontal scale |
| Partitioning | Manual sharding | Built-in |
| Schema | Rigid | Flexible |
| Joins | Supported | Not needed (key-value) |
| Consistency | Strong | Tunable |

At 100M writes/day (~1.2K/sec), PostgreSQL can handle it with a single node. But for growth to 1B/day, NoSQL scales horizontally. **Recommendation**: Start with PostgreSQL for simplicity; migrate to Cassandra when scaling.

### E. KGS Implementation Details

```
KGS Process:
1. Generate 1M unique keys (base62 encode of counter)
2. Store in Redis list: RPUSH key_pool key1 key2 ... key1000000
3. API servers: LPOP key_pool → get key
4. When pool size < 100K: Trigger KGS to generate 1M more
5. KGS uses distributed lock to prevent multiple generators
```

### F. Analytics Storage Options

- **Time-series DB** (InfluxDB, TimescaleDB): Optimized for metrics
- **Data Lake** (S3 + Athena): Cheap, query on demand
- **ClickHouse**: Columnar; fast aggregations
- **Redis HyperLogLog**: Approximate unique visitors (memory efficient)

### G. Capacity Planning Summary

| Resource | 100M URLs/day | 1B URLs/day |
|----------|---------------|-------------|
| API servers | 10-20 | 50-100 |
| Redis nodes | 3-5 (cluster) | 10-20 |
| DB nodes | 3 (Cassandra) | 10+ |
| KGS instances | 2 (active-standby) | 2-4 |

### H. Sample Redirect Flow (Pseudocode)

```python
def redirect(short_code):
    # 1. Check cache
    cached = redis.get(f"url:{short_code}")
    if cached:
        track_analytics_async(short_code)
        return redirect(302, cached["long_url"])

    # 2. DB lookup
    row = db.get_by_short_code(short_code)
    if not row or (row.expires_at and row.expires_at < now()):
        return 404

    # 3. Cache and redirect
    redis.setex(f"url:{short_code}", 86400, json.dumps({"long_url": row.long_url}))
    track_analytics_async(short_code)
    return redirect(302, row.long_url)
```

### I. Custom Alias Validation Rules

- **Length**: 4-20 characters
- **Characters**: Alphanumeric, hyphen, underscore only
- **Reserved**: Block `api`, `admin`, `help`, `www`, `mail`, `ftp`
- **Profanity**: Filter blocklist
- **Case**: Store lowercase; accept case-insensitive lookup

### J. Edge Cases to Handle

1. **Empty or invalid URL**: Reject with 400
2. **Duplicate long URL**: Optionally return existing short URL (idempotent)
3. **Custom alias taken**: 409 Conflict; suggest alternatives
4. **Expired URL**: 404 with message "Link has expired"
5. **Redirect loop**: Detect if long_url points to same domain; reject

### K. Complete Interview Walkthrough (45 min)

**0-5 min**: Clarify requirements. Ask: scale? custom aliases? analytics? expiry?
**5-10 min**: Back-of-envelope. 100M writes, 1B reads → QPS, storage, bandwidth.
**10-15 min**: API design. Create, redirect, analytics. Discuss 301 vs 302.
**15-25 min**: Data model. Cassandra vs PostgreSQL. KGS vs hash. Collision handling.
**25-35 min**: Architecture diagram. Components: LB, API, Redis, KGS, DB, analytics.
**35-40 min**: Deep dive. KGS design. Caching. Scaling. Failure handling.
**40-45 min**: Trade-offs. 301 vs 302. SQL vs NoSQL. Fail open/closed for analytics.

### L. Quick Reference Cheat Sheet

| Topic | Key Points |
|-------|------------|
| Scale | 100M writes/day, 1B reads/day, 10:1 ratio |
| Key gen | KGS (best) or hash+truncation with collision handling |
| Encoding | Base62: 7 chars = 62^7 ≈ 3.5T combinations |
| Redirect | 301=cached, SEO; 302=analytics, flexible |
| DB | Cassandra/DynamoDB for scale; PostgreSQL for simplicity |
| Cache | Redis, 80% hit rate, hot URLs |
| Analytics | Async to Kafka, batch to time-series DB |

### M. Further Reading & Real-World Examples

- **TinyURL**: One of the first; simple design
- **bit.ly**: Custom domains, analytics, enterprise
- **Google URL Shortener** (discontinued): Integrated with Google ecosystem
- **AWS S3 Redirect**: Use S3 static website redirect for simple cases

### N. Design Alternatives Considered

| Decision | Alternative | Why Rejected |
|----------|-------------|--------------|
| KGS | Hash + truncation | Collision handling complex; retries |
| Cassandra | PostgreSQL | Scale; but PG fine for <10M/day |
| 302 | 301 | 302 for analytics; 301 for static |
| Sync analytics | Async | Sync adds latency to redirect path |

### O. Base62 Encoding (Full Implementation)

```python
BASE62 = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

def encode(num):
    if num == 0: return BASE62[0]
    result = []
    while num:
        result.append(BASE62[num % 62])
        num //= 62
    return ''.join(reversed(result))

def decode(s):
    num = 0
    for c in s:
        num = num * 62 + BASE62.index(c)
    return num
```
