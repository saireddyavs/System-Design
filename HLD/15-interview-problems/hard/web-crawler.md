# Design a Distributed Web Crawler

## 1. Problem Statement & Requirements

### Problem Statement
Design a distributed web crawler that can crawl billions of web pages per day, extract content and links, handle duplicates, respect robots.txt, maintain politeness (rate limiting per domain), and support dynamic JavaScript-rendered pages.

### Functional Requirements
- **URL discovery**: Extract URLs from crawled pages; add to crawl queue
- **Content extraction**: Extract text, metadata, links from HTML
- **Deduplication**: Avoid crawling same URL multiple times
- **robots.txt**: Respect crawl delays and disallowed paths per domain
- **Politeness**: Rate limit requests per domain (e.g., 1 req/sec)
- **Dynamic pages**: Support JavaScript-rendered content (headless browser)
- **Prioritization**: Crawl important/fresh pages first
- **Scheduling**: Recrawl based on page change frequency

### Non-Functional Requirements
- **Throughput**: 1B pages/day ≈ 11,500 pages/sec average; 100K+ peak
- **Latency**: Not critical for crawl; but URL processing should be fast
- **Storage**: Petabytes for raw HTML, extracted content, URL index
- **Availability**: 99.9% (crawler can tolerate some downtime)

### Out of Scope
- Search index building (assume crawler feeds a separate indexer)
- Ad fraud detection
- Malware scanning (simplified)
- Distributed storage for raw HTML (assume object store)

---

## 2. Back-of-Envelope Estimation

### Assumptions
- 1B pages/day
- Average page: 100 KB HTML, 50 links
- 100M unique domains
- 10B unique URLs in frontier
- 50% duplicate rate (URL dedup)

### QPS Estimates
| Component | Calculation | QPS |
|-----------|-------------|-----|
| Page fetches | 1B / 86400 | ~11,600 |
| Peak (10x) | 11600 × 10 | ~116,000 |
| URL extractions | 1B × 50 | 50B links/day → ~580K/sec |
| URL dedup checks | 1B + 50B | ~590K/sec |
| DNS lookups | 116K / 100 (cache) | ~1,200 |

### Storage (1 year)
| Data | Size/record | Records | Total |
|------|-------------|---------|-------|
| Raw HTML | 100 KB | 365B | 36.5 PB |
| Extracted content | 20 KB | 365B | 7.3 PB |
| URL index | 100 B | 500B | 50 TB |
| robots.txt cache | 10 KB | 100M | 1 TB |
| Bloom filter (URLs) | - | 500B | ~60 GB (1% FP) |

### Bandwidth
- Fetch: 116K × 100 KB = 11.6 GB/s
- Outbound: Dominated by HTTP requests

### Cache
- DNS cache: 100M domains × 200 B ≈ 20 GB
- robots.txt: 100M × 10 KB ≈ 1 TB (distributed)
- URL seen (Bloom): ~60 GB

---

## 3. API Design

### Internal APIs (Service-to-Service)

```
POST   /internal/urls/add              # Add URLs to frontier (batch)
GET    /internal/urls/next             # Get next URL to crawl (blocking)
POST   /internal/pages/store           # Store crawled page
GET    /internal/robots/{domain}       # Get robots.txt (cached)
GET    /internal/dns/resolve           # Resolve hostname (cached)
```

### Control Plane (Admin)

```
POST   /admin/crawl/start              # Start crawl with seed URLs
POST   /admin/crawl/stop                # Stop crawl
GET    /admin/stats                    # Crawl statistics
POST   /admin/domains/block             # Block domain
POST   /admin/domains/rate              # Set rate limit per domain
```

### Message Queue (Kafka)
- **Topic: urls_to_crawl**: URL, priority, depth
- **Topic: crawled_pages**: URL, HTML, metadata, links
- **Topic: urls_extracted**: Extracted URLs for dedup and frontier

---

## 4. Data Model / Database Schema

### URL Frontier

**url_frontier** (priority queue)
| Column | Type |
|--------|------|
| url_hash | VARCHAR PK |
| url | TEXT |
| domain | VARCHAR |
| priority | INT |
| depth | INT |
| next_fetch_at | TIMESTAMP |
| created_at | TIMESTAMP |

**Partitioning**: By domain hash (for politeness)

### URL Deduplication

**Option 1: Bloom filter**
- Probabilistic; 1% false positive rate
- 500B URLs → ~60 GB
- Cannot remove (use counting Bloom or separate "crawled" set)

**Option 2: Distributed key-value**
- Key: URL hash (e.g., SHA-256)
- Value: timestamp
- Redis/Cassandra

### Crawl State

**crawled_urls**
| Column | Type |
|--------|------|
| url_hash | VARCHAR PK |
| url | TEXT |
| last_crawled | TIMESTAMP |
| next_crawl_at | TIMESTAMP |
| change_frequency | ENUM |
| http_status | INT |

### robots.txt Cache

**robots_cache**
| Column | Type |
|--------|------|
| domain | VARCHAR PK |
| rules | TEXT |
| crawl_delay | INT |
| fetched_at | TIMESTAMP |

### DB Choice
- **Redis**: URL frontier (sorted sets by next_fetch_at), Bloom filter, DNS cache
- **Cassandra**: Crawled URLs, URL seen (high write throughput)
- **PostgreSQL**: Crawl metadata, admin config
- **S3/GCS**: Raw HTML, extracted content

---

## 5. High-Level Architecture

```
                                    ┌─────────────────────────────────────────────────────────────┐
                                    │                    SEED URL INJECTION                        │
                                    │              (Manual, Sitemaps, Discovered)                   │
                                    └─────────────────────────────────────────────────────────────┘
                                                              │
                                                              ▼
                                    ┌─────────────────────────────────────────────────────────────┐
                                    │                    URL NORMALIZER                            │
                                    │         (Canonicalize, remove fragments, sort params)       │
                                    └─────────────────────────────────────────────────────────────┘
                                                              │
                                                              ▼
                                    ┌─────────────────────────────────────────────────────────────┐
                                    │                    URL DEDUPLICATOR                         │
                                    │              (Bloom filter / KV check)                      │
                                    └─────────────────────────────────────────────────────────────┘
                                                              │
                                                              ▼
┌─────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                          URL FRONTIER (Priority Queue)                                           │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐               │
│  │ Domain A    │ │ Domain B    │ │ Domain C    │ │ Domain D    │ │ ...         │ │ Domain N    │               │
│  │ Queue       │ │ Queue       │ │ Queue       │ │ Queue       │ │             │ │ Queue       │               │
│  │ (FIFO +     │ │ (FIFO +     │ │ (FIFO +     │ │ (FIFO +     │ │             │ │ (FIFO +     │               │
│  │  politeness)│ │  politeness)│ │  politeness)│ │  politeness)│ │             │ │  politeness)│               │
│  └──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └──────┬──────┘               │
└─────────┼──────────────┼──────────────┼──────────────┼──────────────┼──────────────┼───────────────────────┘
          │              │              │              │              │              │
          └──────────────┴──────────────┴──────────────┴──────────────┴──────────────┘
                                          │
                                          ▼
                                    ┌─────────────────────────────────────────────────────────────┐
                                    │                    CRAWL SCHEDULER                          │
                                    │     (Select URL from frontier respecting politeness)        │
                                    └─────────────────────────────────────────────────────────────┘
                                          │
          ┌───────────────────────────────┼───────────────────────────────┐
          │                               │                               │
          ▼                               ▼                               ▼
┌─────────────────┐             ┌─────────────────┐             ┌─────────────────┐
│  DNS Resolver   │             │  robots.txt     │             │  Fetcher Pool    │
│  (Cached)       │             │  Fetcher        │             │  (HTTP Clients)  │
└────────┬────────┘             └────────┬────────┘             └────────┬────────┘
         │                               │                               │
         └──────────────────────────────┼───────────────────────────────┘
                                        │
                                        ▼
                              ┌─────────────────┐
                              │  HTTP Fetch     │
                              │  (with redirect,│
                              │   timeout)      │
                              └────────┬────────┘
                                        │
                                        ▼
                              ┌─────────────────┐
                              │  Content Parser  │
                              │  (HTML → links, │
                              │   text, metadata)│
                              └────────┬────────┘
                                        │
          ┌─────────────────────────────┼─────────────────────────────┐
          │                             │                             │
          ▼                             ▼                             ▼
┌─────────────────┐           ┌─────────────────┐           ┌─────────────────┐
│  Content        │           │  Link Extractor │           │  Content        │
│  Dedup          │           │  → URL Frontier │           │  Store (S3)     │
│  (SimHash)      │           │  (normalize,    │           │  (raw HTML)     │
│                  │           │   dedup)       │           │                 │
└─────────────────┘           └─────────────────┘           └─────────────────┘
```

### Component Descriptions
- **URL Frontier**: Priority queue partitioned by domain; politeness per domain
- **Crawl Scheduler**: Picks next URL respecting rate limits
- **DNS Resolver**: Cached; avoid bottleneck
- **Fetcher**: HTTP client pool; redirects, timeouts
- **Content Parser**: Extract links, text, metadata
- **URL Deduplication**: Bloom filter or KV
- **Content Deduplication**: SimHash/MinHash for near-duplicates

---

## 6. Detailed Component Design

### 6.1 URL Frontier

**Structure**: Per-domain queues + priority

- **Politeness**: One queue per domain; FIFO within domain
- **Rate limit**: 1 request per second per domain (configurable)
- **Priority**: Higher priority URLs (e.g., homepage, sitemap) first
- **Scheduling**: `next_fetch_at` = last_fetch + crawl_delay

**Implementation**:
- **Redis**: Sorted set per domain: score = next_fetch_at, value = url
- **Kafka**: Topic per domain (or partition by domain hash)
- **Custom**: In-memory queues per domain; persistent to disk

**Selection algorithm**:
1. Get domains with `next_fetch_at <= now`
2. Pick domain (round-robin or by priority)
3. Pop URL from domain queue
4. Update `next_fetch_at` for domain
5. Return URL to fetcher

### 6.2 URL Normalization

Before dedup and frontier:
- Lowercase scheme and host
- Remove default port (80, 443)
- Remove fragment (#)
- Sort query parameters
- Remove trailing slash (or keep consistent)
- Decode percent-encoding (carefully)
- Resolve relative URLs to absolute

Example:
```
https://Example.com:443/path/?b=2&a=1#section
→ https://example.com/path/?a=1&b=2
```

### 6.3 URL Deduplication

**Bloom filter**:
- 500B URLs, 1% FP → ~60 GB (10 bits per element)
- Add URL hash on first see
- Check before add to frontier: if present, skip
- False positive: may skip new URL (acceptable for crawl)
- Cannot remove: use "crawled" set for recrawl

**Alternative: Cassandra/Redis**:
- Key: SHA-256(url)
- Add if not exists (idempotent)
- TTL for recrawl

### 6.4 DNS Resolver

**Problem**: DNS can be bottleneck (slow, rate-limited)

**Solution**:
- Local DNS cache (in-memory)
- TTL from DNS response (typically 300–3600 sec)
- Prefetch: resolve before fetch
- Batch resolution for multiple hosts

**Cache**: 100M domains × 200 B ≈ 20 GB

### 6.5 Fetcher

**HTTP client**:
- Connection pooling (keep-alive)
- Timeout: 10–30 sec
- Follow redirects (max 5)
- User-Agent: Identify crawler
- Respect robots.txt (check before fetch)

**Politeness**:
- Wait `crawl_delay` from robots.txt
- Default: 1 sec between requests per domain
- Back off on 429 (Too Many Requests)

**Handling**:
- 4xx/5xx: Log; optionally retry later
- Timeout: Re-queue with backoff
- Redirect: Add new URL to frontier; may dedup

### 6.6 Content Parser

**Extract**:
- Links: `<a href>`, `<link>`, `<script src>`, `<img src>`
- Text: Strip HTML; extract body text
- Metadata: title, meta description, Open Graph
- Structure: headings, paragraphs (for indexing)

**Output**:
- List of URLs (normalized)
- Plain text (for indexer)
- Metadata (for storage)

### 6.7 Content Deduplication (Near-Duplicate Detection)

**SimHash**:
- Compute SimHash of document (64-bit fingerprint)
- Similar documents have similar hashes (small Hamming distance)
- Compare: if Hamming distance < 3, consider duplicate
- Store: LSH (Locality-Sensitive Hashing) for fast lookup

**MinHash**:
- Shingle document (n-grams)
- MinHash signatures
- Jaccard similarity for near-duplicate detection

**Use**: Skip storing/indexing near-duplicates

### 6.8 robots.txt Compliance

**Fetch**: Before first request to domain, fetch `https://domain/robots.txt`
**Parse**: Extract Disallow, Allow, Crawl-delay
**Cache**: 24 hours (or Cache-Control)
**Apply**: Before each fetch, check if path is allowed; wait crawl_delay

**Example**:
```
User-agent: *
Disallow: /admin/
Disallow: /search?
Crawl-delay: 2
```

### 6.9 Crawl Scheduling (Recrawl Frequency)

**Change frequency**:
- News: Recrawl every hour
- Blog: Daily
- Static: Weekly/monthly

**Strategy**:
- Track last-modified or hash
- If changed frequently → recrawl more often
- If static → recrawl less
- Store `next_crawl_at` per URL

### 6.10 Trap Detection (Infinite URL Spaces)

**Problem**: Calendar, session IDs, pagination create infinite URLs

**Detection**:
- Depth limit: Don't follow beyond depth 10
- URL pattern: Limit same pattern (e.g., `/page/1`, `/page/2`, ...) to N
- Content similarity: If page very similar to seen, may be template
- Canonical URL: Use `<link rel="canonical">` to dedup

### 6.11 Dynamic (JavaScript-Rendered) Pages

**Problem**: SPA content not in raw HTML

**Solution**:
- **Headless browser**: Puppeteer, Playwright, Selenium
- **Cost**: 10–100x slower than HTTP fetch
- **Strategy**: Use for subset (e.g., known SPA domains)
- **Hybrid**: HTTP first; if content sparse, try headless

### 6.12 Distributed Architecture

**Partitioning**:
- URL frontier: Partition by domain hash
- Each worker: Own set of domains
- Avoids lock contention; natural politeness

**Scaling**:
- Add fetcher workers
- Each worker: Get URLs from frontier (partition)
- Parallel fetch; respect per-domain rate
- Publish to Kafka for parsing

**Fault tolerance**:
- At-least-once: Re-queue on failure
- Idempotent: Content store overwrites
- Checkpoint: Frontier state to durable storage

---

## 7. Scaling

### Sharding
- **URL frontier**: By domain hash (e.g., 1000 partitions)
- **Fetchers**: Each handles subset of partitions
- **Content store**: By URL hash

### Caching
- **DNS**: In-memory per worker; shared Redis
- **robots.txt**: Redis; 24h TTL
- **URL seen**: Bloom filter (distributed; e.g., partitioned by hash)

### Rate Limiting
- Token bucket per domain
- Redis: `INCR` with TTL for rate limit
- Or: Queue with delay per domain

### Message Queue
- **Kafka**: Decouple fetcher from parser
- **Kafka**: URLs to crawl (from seeds, extracted links)
- Retention: 7 days for replay

---

## 8. Failure Handling

### Fetcher Crash
- URLs in flight: Re-queue (at-least-once)
- Frontier: Durable; no loss

### Parser Crash
- Kafka: Replay from offset
- Idempotent parsing

### DNS Failure
- Retry with backoff
- Skip URL after N failures
- Log for manual review

### robots.txt Unavailable
- Default: Allow all, 1 sec delay
- Retry later

### Storage Full
- Compress HTML (gzip)
- Tier: Hot (recent) vs cold (S3 Glacier)
- Delete old content based on policy

---

## 9. Monitoring & Observability

### Key Metrics
| Metric | Target |
|--------|--------|
| Pages crawled/sec | 11,600+ |
| Fetch success rate | > 95% |
| URL dedup hit rate | > 50% |
| Frontier size | Monitor growth |
| Fetch latency p99 | < 10 sec |
| robots.txt cache hit | > 99% |

### Logging
- Fetch failures (4xx, 5xx, timeout)
- Blocked by robots.txt
- Trap detection (URL pattern)
- Duplicate content (SimHash)

### Dashboards
- Crawl rate over time
- Top domains by volume
- Error rate by domain
- Frontier depth distribution

---

## 10. Interview Tips

### Follow-up Questions
1. How would you handle CAPTCHA-protected pages?
2. How do you prioritize which URLs to crawl first?
3. How would you crawl the "dark web" or non-HTTP content?
4. How do you handle sitemaps?
5. How would you implement incremental crawl (only changed pages)?

### Common Mistakes
- Ignoring politeness (hammering servers)
- Not handling robots.txt
- Single-threaded design (bottleneck)
- No URL normalization (duplicate crawls)
- Ignoring trap detection

### What to Emphasize
- URL frontier design (per-domain queue, politeness)
- Deduplication (Bloom filter, URL normalization)
- Content dedup (SimHash)
- robots.txt compliance
- Distributed partitioning by domain
- Crawl scheduling and recrawl frequency

---

## Appendix A: Bloom Filter Sizing

### Formula
- n = expected elements (500B)
- p = false positive rate (0.01)
- m = bits needed = -n * ln(p) / (ln 2)² ≈ 4.8 billion bits ≈ 600 MB per 10B elements
- k = hash functions = m/n * ln 2 ≈ 7

### For 500B URLs
- m ≈ 480 billion bits ≈ 60 GB
- With 1% FP: 1% of new URLs incorrectly skipped
- Trade-off: Larger filter = lower FP, more memory

### Counting Bloom Filter
- Allows deletion (decrement counters)
- 4x memory of standard Bloom
- Used when URLs can "expire" (recrawl)

---

## Appendix B: SimHash Algorithm

### Steps
1. Tokenize document (words or n-grams)
2. Hash each token to 64-bit integer
3. For each bit position i: sum +1 if token hash has bit i set, else -1
4. Result: 64-bit vector where bit i = sign of sum
5. Similar docs → similar hashes (small Hamming distance)

### Comparison
- Hamming distance < 3 → likely duplicate
- LSH (Locality-Sensitive Hashing): Partition 64 bits into bands; same band → candidate pair
- Reduces comparisons from O(n²) to O(n)

---

## Appendix C: robots.txt Parsing

### Rules
- User-agent: Match crawler name (or *)
- Disallow: Path prefix to block (e.g., /admin/)
- Allow: Override Disallow (e.g., Allow /admin/public/)
- Crawl-delay: Seconds between requests (non-standard, some bots support)
- Sitemap: URL to sitemap (optional)

### Precedence
- Longest matching path wins
- Allow can override Disallow for subpath
- Default: Allow all if no match

### Example Parse
```
User-agent: *
Disallow: /private/
Allow: /private/public/
Crawl-delay: 2
```
→ Block /private/* except /private/public/*; wait 2 sec between requests

---

## Appendix D: Sitemap Handling

### Format
- XML: `<url><loc>...</loc><lastmod>...</lastmod><changefreq>...</changefreq></url>`
- Sitemap index: Links to multiple sitemaps
- Gzip: sitemap.xml.gz

### Integration
- Fetch sitemap from domain root or robots.txt
- Parse URLs; add to frontier with priority (sitemap URLs often important)
- Use lastmod, changefreq for recrawl scheduling
- Respect 50,000 URL limit per sitemap

---

## Appendix E: Trap Detection Strategies

### URL Pattern Limits
- `/page/1`, `/page/2`, ...: Limit to 100 pages per pattern
- `/date/2024/01/01`, ...: Limit date range
- Session IDs: Block URLs with long random strings

### Content-Based
- SimHash: If new page matches existing, skip
- Boilerplate ratio: If 90% of page is same as others, may be template

### Structural
- Depth limit: Max 10 hops from seed
- Domain limit: Max 10,000 pages per domain (configurable)

---

## Appendix F: Fetcher Pool Design

### Connection Pool
- Per-host connection pool (HTTP/1.1 keep-alive)
- Max 2–5 connections per host (politeness)
- Total pool: 10,000 connections across 2,000 hosts

### Timeout Strategy
- Connect: 5 sec
- Read: 30 sec
- Total: 60 sec max per request
- On timeout: Re-queue with exponential backoff (1 min, 5 min, 30 min)

### Retry Logic
- 5xx: Retry 3 times with backoff
- 429: Retry after Retry-After header
- 4xx (except 429): No retry; log
- Network error: Retry 5 times

---

## Appendix G: Content Extraction Pipeline

### HTML Parsing
- Use library: BeautifulSoup, Jsoup, or html5lib
- Parse to DOM
- Extract: title, meta description, meta keywords, Open Graph
- Extract links: a[href], link[href], script[src], img[src]
- Resolve relative URLs to absolute

### Text Extraction
- Remove script, style tags
- Get text from body
- Normalize whitespace
- Optional: Extract headings (h1-h6) for structure

### Output Schema
```json
{
  "url": "https://example.com/page",
  "title": "Page Title",
  "description": "Meta description",
  "links": ["https://...", "..."],
  "text": "Plain text content...",
  "crawled_at": "2024-01-01T00:00:00Z"
}
```

---

## Appendix H: Distributed Frontier Partitioning

### Partition by Domain Hash
- hash(domain) % N = partition_id
- Each fetcher worker owns K partitions
- Worker polls its partitions for next URL
- Ensures one worker per domain (no duplicate fetches)

### Work Stealing (Optional)
- If worker idle (all its domains waiting for politeness)
- Steal from busy worker's queue
- Complex: must respect per-domain rate

---

## Appendix I: Headless Browser Strategy

### When to Use
- Known SPA domains (e.g., React, Vue apps)
- When HTTP fetch returns minimal content
- For specific high-value domains

### Optimization
- Pool of browser instances (e.g., 100)
- Reuse pages (don't create new for each URL)
- Disable images, fonts (faster)
- Set viewport to minimal
- Timeout: 15 sec

### Cost
- ~10 sec per page vs ~0.5 sec for HTTP
- 20x slower; use sparingly
- Consider hybrid: HTTP first, headless on low-content detection
