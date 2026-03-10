# System Design: Pastebin

## 1. Problem Statement & Requirements

### Functional Requirements

- **Create Paste**: Users can create a new paste with raw text content
- **Read Paste**: Users can view paste content by unique ID/URL
- **Delete Paste**: Paste owners can delete their pastes (if authenticated)
- **Syntax Highlighting**: Support syntax highlighting for various languages (Python, JavaScript, etc.)
- **Expiry**: Pastes can expire after N minutes, N hours, N days, or never
- **Visibility**: Public (anyone with link) vs Private (only owner)
- **Custom URLs**: Optional custom paste IDs (e.g., `pastebin.com/mypaste`)
- **Raw vs Rendered**: Serve raw text or HTML-rendered (with highlighting)

### Non-Functional Requirements

- **Scale**: 5M pastes created per day; read-heavy (10:1 or higher read/write)
- **Latency**: Read latency < 100ms (p99); Create < 200ms
- **Availability**: 99.9% uptime
- **Storage**: Support large pastes (up to 1MB per paste)
- **Durability**: Paste content must not be lost

### Out of Scope

- User authentication (assume optional; anonymous pastes allowed)
- Collaborative editing
- Version history
- Search across pastes
- Paste embedding in iframes

---

## 2. Back-of-Envelope Estimation

### Traffic Estimates

| Metric | Value | Calculation |
|--------|-------|-------------|
| Writes per day | 5M | Given |
| Reads per day | 50M | 10:1 read/write |
| Writes per second | ~60 | 5M / 86400 |
| Reads per second | ~600 | 50M / 86400 |
| Peak QPS | ~2,000 reads | 3x average |

### Storage Estimates

- **Assumptions**:
  - Avg paste size: 5 KB (metadata + content)
  - Max paste: 1 MB
  - Metadata: 500 bytes (id, created_at, expiry, language, visibility)
  - 5M pastes/day × 365 days × 5 KB ≈ 9 TB/year

- **Content vs Metadata**:
  - Content: Store in object storage (S3); ~8 TB/year
  - Metadata: Store in DB; ~200 GB/year

### Bandwidth Estimates

- **Read**: 600 QPS × 5 KB ≈ 3 MB/s ≈ 260 GB/day
- **Write**: 60 QPS × 5 KB ≈ 0.3 MB/s ≈ 26 GB/day
- **CDN**: 80% of reads from CDN → 208 GB/day from CDN, 52 GB/day from origin

### Cache Estimates

- **Hot pastes**: Top 1% get 50% of reads
- **Cache**: 50K pastes × 5 KB ≈ 250 MB
- **Redis**: 1-2 GB for metadata + content cache

---

## 3. API Design

### REST API Endpoints

#### Create Paste

```
POST /api/v1/pastes
```

**Request:**
```json
{
  "content": "def hello():\n    print('Hello, World!')",
  "language": "python",
  "expiry": "1d",                    // "10m", "1h", "1d", "1w", "never"
  "visibility": "public",             // "public" | "private"
  "custom_id": "mypaste"              // optional
}
```

**Response (201 Created):**
```json
{
  "id": "abc123",
  "url": "https://pastebin.com/abc123",
  "created_at": "2025-03-10T12:00:00Z",
  "expires_at": "2025-03-11T12:00:00Z",
  "language": "python",
  "visibility": "public"
}
```

#### Get Paste (Raw)

```
GET /api/v1/pastes/{id}/raw
```

**Response (200 OK):**
```
Content-Type: text/plain; charset=utf-8

def hello():
    print('Hello, World!')
```

#### Get Paste (Rendered with Syntax Highlighting)

```
GET /api/v1/pastes/{id}
```

**Response (200 OK):**
```html
Content-Type: text/html

<!DOCTYPE html>
<html>
<head><link rel="stylesheet" href="highlight.css"></head>
<body><pre><code class="language-python">...</code></pre></body>
</html>
```

#### Get Paste Metadata

```
GET /api/v1/pastes/{id}/meta
```

**Response (200 OK):**
```json
{
  "id": "abc123",
  "created_at": "2025-03-10T12:00:00Z",
  "expires_at": "2025-03-11T12:00:00Z",
  "language": "python",
  "visibility": "public",
  "size_bytes": 42
}
```

#### Delete Paste

```
DELETE /api/v1/pastes/{id}
```

**Headers:** `Authorization: Bearer <token>` (if private)

**Response (204 No Content)**

**Response (404 Not Found):** Paste doesn't exist or already expired

---

## 4. Data Model / Database Schema

### Storage Strategy

**Content**: **Object Storage (S3, GCS)**
- Pastes can be large (up to 1MB)
- Cheap, durable, scalable
- Key: `pastes/{paste_id}` or `pastes/{shard_id}/{paste_id}`

**Metadata**: **PostgreSQL** or **DynamoDB**
- Need to lookup by ID, check expiry, visibility
- Relational for user's pastes listing (if auth)

### Schema (PostgreSQL)

```sql
CREATE TABLE pastes (
    id VARCHAR(20) PRIMARY KEY,
    content_key VARCHAR(255) NOT NULL,      -- S3 key or path
    user_id VARCHAR(50),                     -- nullable for anonymous
    language VARCHAR(50) DEFAULT 'plaintext',
    visibility VARCHAR(20) DEFAULT 'public',
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    size_bytes INT,
    is_custom_id BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_pastes_expires_at ON pastes(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_pastes_user_id ON pastes(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_pastes_created_at ON pastes(created_at DESC);
```

### Object Storage Layout

```
bucket: pastebin-content
  pastes/
    a/
      abc123          # content for paste abc123
    b/
      b7x9k2
    ...
```

Sharding by first char or hash prefix for distribution.

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
│              Cache popular pastes (raw + rendered), static assets                  │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              LOAD BALANCER                                        │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
                    ┌───────────────────┼───────────────────┐
                    ▼                   ▼                   ▼
┌───────────────────────┐   ┌───────────────────────┐   ┌───────────────────────┐
│   API Server 1        │   │   API Server 2        │   │   API Server N         │
│   - Create/Read/Delete│   │   - Syntax highlighting│   │   - Metadata          │
└───────────────────────┘   └───────────────────────┘   └───────────────────────┘
                    │                   │                   │
                    └───────────────────┼───────────────────┘
                                        │
        ┌───────────────────────────────┼───────────────────────────────┐
        ▼                               ▼                               ▼
┌───────────────────┐       ┌───────────────────────┐       ┌───────────────────────┐
│   Redis Cache     │       │   PostgreSQL          │       │   S3 / Object Store   │
│   (metadata +     │       │   (metadata only)     │       │   (paste content)     │
│   hot content)    │       │                       │       │                       │
└───────────────────┘       └───────────────────────┘       └───────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│   Cleanup Service (Cron) → Delete expired pastes from DB + S3                     │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 6. Detailed Component Design

### 6.1 Paste ID Generation

**Options:**
1. **UUID**: 36 chars, too long for URLs
2. **Base62 of random number**: 8 chars = 62^8 ≈ 218T combinations
3. **Base62 of timestamp + random**: Sortable, unique

**Recommended:**
```python
import random
import string
from datetime import datetime

CHARS = string.ascii_lowercase + string.digits  # 36 chars for shorter IDs

def generate_paste_id(length=8):
    return ''.join(random.choices(CHARS, k=length))
```

For 8 chars with 36 chars: 36^8 ≈ 2.8T combinations. Collision check before insert.

### 6.2 Create Paste Flow

1. Validate content (size ≤ 1MB, charset)
2. Generate ID (or use custom_id if provided)
3. Check custom_id uniqueness if applicable
4. Upload content to S3: `pastes/{id}`
5. Insert metadata into PostgreSQL
6. Return URL to user

**Atomicity**: If S3 upload fails, don't insert DB. If DB insert fails after S3 upload, orphaned object; cleanup job can remove.

### 6.3 Read Paste Flow

1. **Check cache** (Redis): key = `paste:{id}`
2. **Cache miss**: Fetch metadata from DB
3. Check expiry: if `expires_at < now`, return 404
4. Check visibility: if private, verify auth
5. Fetch content from S3 (or cache)
6. If rendered request: Apply syntax highlighting (client-side or server-side)
7. Cache result in Redis (TTL = min(1h, time_to_expiry))

### 6.4 Syntax Highlighting

**Options:**
- **Server-side**: Pygments, Highlight.js (Node) — more control, consistent
- **Client-side**: Highlight.js in browser — offload work, faster first byte

**Recommendation**: Serve raw content + include Highlight.js in HTML template. Client does highlighting. Reduces server load.

For raw API: Just return content. For HTML: Return template with content in `<pre><code>`, Highlight.js runs in browser.

### 6.5 Content Storage: Why Object Storage?

- **Database**: BLOBs in DB work but don't scale well for large objects; DB becomes bottleneck
- **Object Storage**: Designed for large blobs; cheap ($0.023/GB/month S3); 99.999999999% durability
- **Hybrid**: Metadata in DB (fast lookup), content in S3 (cheap storage)

### 6.6 Cleanup Service for Expired Pastes

- **Cron**: Run every hour
- **Query**: `SELECT id, content_key FROM pastes WHERE expires_at < NOW() LIMIT 10000`
- **Delete**: Remove from S3, then from DB (or vice versa; DB first avoids orphan metadata)
- **Batch**: Process in batches to avoid long transactions

---

## 7. Scaling

### Horizontal Scaling

- **API Servers**: Stateless; scale based on QPS
- **Database**: Read replicas for metadata reads
- **S3**: Naturally distributed; no scaling concern

### Sharding Strategy

- **Metadata DB**: Shard by `id` hash if needed (e.g., `hash(id) % 10`)
- **S3**: Prefix sharding `pastes/{hash_prefix}/{id}` for even distribution

### Caching Strategy

- **Redis**: Cache metadata + content for hot pastes
- **Key**: `paste:{id}` → JSON or raw content
- **TTL**: 1 hour or time until expiry
- **Eviction**: LRU when memory full

### CDN Usage

- Cache GET responses for public pastes
- Cache-Control: `public, max-age=3600` (1 hour)
- Invalidate on delete (optional; or short TTL)
- CDN for static assets: Highlight.js CSS/JS

### Read-Heavy Optimization

- Majority of traffic is reads
- Aggressive caching at CDN + Redis
- DB read replicas for metadata
- S3 has built-in caching; consider CloudFront in front of S3

---

## 8. Failure Handling

| Component | Failure | Mitigation |
|-----------|---------|------------|
| Redis | Down | Fallback to DB + S3; slower; repopulate on recovery |
| PostgreSQL | Primary down | Failover to replica; async replication |
| S3 | Unavailable | Retry with exponential backoff; S3 has 99.99% SLA |
| API Server | Down | LB health check; remove from pool |
| Cleanup job | Fails | Idempotent; retry next run; orphaned objects acceptable temporarily |

### Redundancy

- Redis: Cluster or Sentinel
- PostgreSQL: Primary + 2 replicas
- S3: Multi-AZ by default

### Recovery

- Cache miss → DB + S3 → repopulate cache
- Orphaned S3 objects: Cleanup job scans DB, deletes S3 objects not in DB (or vice versa)

---

## 9. Monitoring & Observability

### Key Metrics

| Metric | Target | Alert |
|--------|--------|-------|
| Read latency | p99 < 100ms | Critical |
| Create latency | p99 < 200ms | Warning |
| Cache hit rate | > 70% | Warning |
| Error rate | < 0.1% | Critical |
| S3 upload failure | < 0.01% | Critical |
| Expired pastes cleanup | Run successfully | Warning if failed |

### Dashboards

- QPS (create, read, delete)
- Latency percentiles
- Cache hit rate
- Storage usage (S3, DB)
- Cleanup job status

### Alerting

- Read latency > 200ms
- Create failure rate > 1%
- S3 errors
- DB connection pool exhaustion

---

## 10. Interview Tips

### Common Follow-Up Questions

1. **Why separate content and metadata?** Content is large; DB isn't ideal for blobs. S3 is cheap and durable.
2. **How to handle very popular pastes?** CDN + Redis; cache at edge
3. **How to implement syntax highlighting?** Client-side (Highlight.js) vs server-side (Pygments); trade-off: server load vs first-byte time
4. **How to delete expired pastes?** Background job; batch delete from S3 and DB
5. **What if S3 upload succeeds but DB insert fails?** Orphaned object; cleanup job; or two-phase: reserve ID in DB first, then upload, then mark complete

### What Interviewers Look For

- Separation of metadata and content storage
- Appropriate storage choices (S3 for blobs)
- Caching for read-heavy workload
- Cleanup strategy for expiry
- API design (raw vs rendered)

### Common Mistakes

- Storing large content in database
- Ignoring expiry/cleanup
- Not considering read-heavy nature
- Overcomplicating syntax highlighting

---

## Appendix: Additional Design Considerations

### A. Paste Size Limits

- **1 MB** typical limit (like GitHub Gist)
- Validate before upload; reject with 413 Payload Too Large
- Consider streaming upload for large pastes

### B. Rate Limiting

- Anonymous: 10 pastes/hour
- Authenticated: 100 pastes/hour
- Per IP: 60 pastes/hour (prevent abuse)

### C. Content Security

- **XSS**: Sanitize or escape content when rendering HTML
- **Malicious content**: Consider virus scanning for executable-looking content
- **Abuse**: Rate limit; blocklist for spam

### D. Custom ID Collision

- Use unique constraint in DB
- On conflict: Return 409; suggest alternative
- Consider reserved IDs (e.g., "api", "help")

### E. Two-Phase Create for Consistency

To avoid orphaned S3 objects when DB insert fails:

1. **Phase 1**: Insert metadata with `status='pending'`, get ID
2. **Phase 2**: Upload to S3
3. **Phase 3**: Update metadata to `status='active'`
4. If Phase 2 or 3 fails: Mark as `status='failed'`; cleanup job deletes

### F. Highlight.js Integration Example

```html
<!-- Served from CDN -->
<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.0.0/highlight.min.js"></script>
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.0.0/styles/github.min.css">

<pre><code class="language-python" id="paste-content">{{ content }}</code></pre>
<script>hljs.highlightElement(document.getElementById('paste-content'));</script>
```

### G. S3 Presigned URLs (Alternative)

For very large pastes, use presigned URLs:
- Client uploads directly to S3
- API returns presigned URL; client uploads
- Reduces server load; API only handles metadata

### H. Metadata vs Content Storage Rationale

Storing content in a database leads to:
- **Bloated tables**: BLOBs slow down backups, replication, and queries
- **Connection pool exhaustion**: Large result sets hold connections longer
- **Memory pressure**: DB buffers large objects
- **Cost**: DB storage is 10-50x more expensive than object storage

Object storage (S3) is designed for:
- **Durability**: 11 nines
- **Cost**: ~$0.023/GB/month
- **Throughput**: No single bottleneck
- **Lifecycle policies**: Auto-archive to Glacier for old pastes

### I. Expiry Implementation Options

1. **Lazy deletion**: Check on read; delete if expired
2. **Scheduled job**: Cron every hour; batch delete
3. **TTL in Redis**: If caching, set TTL = time_to_expiry
4. **S3 lifecycle**: Rule to delete objects with specific prefix after N days (if using date-based keys)

### J. Language Detection for Syntax Highlighting

If user doesn't specify language:
- **Detect from content**: Heuristics (e.g., `def ` → Python, `function` → JavaScript)
- **Default**: plaintext
- **Libraries**: Linguist (GitHub), Pygments language detection

### K. Sample Create Paste Flow (Pseudocode)

```python
def create_paste(content, language, expiry, visibility):
    # 1. Validate
    if len(content) > 1_000_000:
        raise PayloadTooLarge()

    # 2. Generate ID
    paste_id = generate_id()  # or custom_id

    # 3. Upload to S3
    s3_key = f"pastes/{paste_id[0]}/{paste_id}"
    s3.put_object(Bucket=BUCKET, Key=s3_key, Body=content)

    # 4. Insert metadata
    expires_at = compute_expiry(expiry) if expiry else None
    db.insert(paste_id, s3_key, language, visibility, expires_at)

    return {"id": paste_id, "url": f"https://pastebin.com/{paste_id}"}
```

### L. Read-Heavy Optimization Checklist

- [ ] CDN for static assets (Highlight.js, CSS)
- [ ] Redis for metadata + hot content
- [ ] DB read replicas
- [ ] Connection pooling
- [ ] Cache-Control headers (1h for public pastes)

### M. Content Types and Encoding

- **Storage**: UTF-8; reject invalid UTF-8
- **Content-Type**: `text/plain; charset=utf-8` for raw
- **Size**: 1 MB typical; configurable per tier
- **Binary**: Not recommended; use base64 if needed (increases size 33%)

### N. Paste Deletion Flow

1. Verify ownership (if private) or use delete token
2. Delete from S3 (or mark for deletion)
3. Delete metadata from DB
4. Invalidate cache keys
5. Return 204 No Content

### O. Complete Interview Walkthrough (45 min)

**0-5 min**: Clarify: create/read/delete, syntax highlighting, expiry, visibility, scale.
**5-10 min**: Estimates. 5M pastes/day, 10:1 read/write. Storage: content vs metadata.
**10-15 min**: API. Create, get raw, get rendered, delete. Content-Type handling.
**15-25 min**: Data model. Why S3 for content? Why DB for metadata? Schema.
**25-35 min**: Architecture. CDN, API servers, Redis, PostgreSQL, S3. Cleanup job.
**35-40 min**: Scaling. Read-heavy optimizations. Caching. CDN.
**40-45 min**: Trade-offs. Client vs server syntax highlighting. Lazy vs scheduled expiry.

### P. Quick Reference Cheat Sheet

| Topic | Key Points |
|-------|------------|
| Scale | 5M pastes/day, 10:1 read/write |
| Content | S3 (cheap, durable); not DB (bloats, expensive) |
| Metadata | PostgreSQL or DynamoDB |
| ID | 8-char random; collision check |
| Highlighting | Client-side (Highlight.js) vs server (Pygments) |
| Expiry | Lazy on read + scheduled cleanup job |
| Caching | Redis for hot; CDN for popular |

### Q. Further Reading & Real-World Examples

- **Pastebin.com**: Original; syntax highlighting, expiry, API
- **GitHub Gist**: 1MB limit; git-based; markdown
- **Hastebin**: Open source; Node.js
- **PrivateBin**: E2E encryption; self-hosted

### R. Design Alternatives Considered

| Decision | Alternative | Why Rejected |
|----------|-------------|--------------|
| S3 for content | DB BLOB | DB bloats; expensive; slow |
| Client-side highlight | Server Pygments | Server load; client is simpler |
| Lazy expiry | TTL in DB | DB doesn't support; cleanup job needed |
| 8-char ID | UUID | UUID too long for URL |

### S. Sample Read Flow with Caching

```
1. Request: GET /abc123
2. Check Redis: key = paste:abc123
3. If hit: Return cached content (raw or HTML)
4. If miss:
   a. Query DB for metadata (id, content_key, expires_at, language)
   b. If expired: 404
   c. Fetch content from S3: pastes/a/abc123
   d. Cache in Redis (TTL = min(1h, time_to_expiry))
   e. Return content
5. For HTML request: Wrap in template with Highlight.js
```

### T. Cleanup Job Pseudocode

```python
def cleanup_expired_pastes():
    batch = db.query("SELECT id, content_key FROM pastes WHERE expires_at < NOW() LIMIT 1000")
    for paste in batch:
        s3.delete_object(Key=paste.content_key)
        db.delete(paste.id)
```

### U. Paste Size Distribution (Typical)

- 50% of pastes: < 1 KB
- 30%: 1-10 KB
- 15%: 10-100 KB
- 5%: 100 KB - 1 MB
- Design for 1 MB max; optimize for < 10 KB average

### V. Supported Languages for Syntax Highlighting

Common: Python, JavaScript, Java, C++, Go, Ruby, PHP, SQL, HTML, CSS, JSON, XML, Bash, Markdown. Use Highlight.js or Prism.js; 200+ languages supported.

### W. Cost Estimation (Rough)

- S3: $0.023/GB/month; 9 TB/year ≈ $2,500/year
- PostgreSQL: $100-500/month for managed
- Redis: $50-200/month for cache
- CDN: $0.01-0.10/GB; 200 GB/day ≈ $600-6,000/month at scale

### X. Summary

Pastebin: S3 for content (cheap, durable), DB for metadata. Read-heavy → cache + CDN. Client-side syntax highlighting. Expiry via lazy + cleanup job. 8-char random IDs.

---
*End of Pastebin System Design Document*

This document covers the design of a Pastebin-like service. Key takeaways: S3 for content, DB for metadata, read-heavy optimization, and client-side syntax highlighting. The separation of content and metadata storage is a critical design decision that enables scalability and cost efficiency. Object storage (S3) is 10-50x cheaper than database storage for large blobs. The cleanup job for expired pastes should run periodically (e.g., hourly) and process in batches to avoid long-running transactions.

**Document Version**: 1.0 | **Last Updated**: 2025-03-10 | **Target**: System Design Interview (Easy)

**Key Interview Questions to Prepare**:
- Why S3 for content instead of database?
- How would you implement syntax highlighting?
- How do you handle expired paste cleanup?
- What's the read vs write ratio and how does it affect design?
- How would you scale to 50M pastes per day?
- Client-side vs server-side syntax highlighting trade-offs?
- Lazy vs scheduled expiry cleanup?


