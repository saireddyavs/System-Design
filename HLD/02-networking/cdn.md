# Content Delivery Network (CDN)

## 1. Concept Overview

### Definition
A Content Delivery Network (CDN) is a geographically distributed network of proxy servers and data centers that cache content close to end users. The goal is to serve content with high availability and high performance by reducing latency, bandwidth, and load on the origin server.

### Purpose
- **Reduce Latency**: Serve content from edge servers near users (e.g., 20ms vs 200ms)
- **Reduce Origin Load**: Cache at edge, origin only serves cache misses
- **Improve Availability**: Distributed infrastructure, DDoS absorption
- **Bandwidth Savings**: Reduce transit costs, offload from origin
- **Global Scale**: Deliver to users worldwide with consistent performance

### Problems It Solves
1. **Latency**: Users far from origin experience slow load times
2. **Origin Overload**: Single server cannot handle global traffic
3. **Bandwidth Costs**: Serving from origin is expensive at scale
4. **Single Point of Failure**: Origin down = service down
5. **DDoS Vulnerability**: Origin exposed to attacks

---

## 2. Real-World Motivation

### Netflix
- **Open Connect**: Custom CDN, not third-party
- **OCA (Open Connect Appliance)**: Hardware placed inside ISP networks
- **~15% of internet traffic**: At peak hours
- **Regional caching**: Popular content cached at edge; long-tail from origin
- **Tiered cache**: Edge (ISP) -> Regional -> Origin

### YouTube (Google)
- **Google Global Cache (GGC)**: Servers in ISP networks
- **Adaptive bitrate**: CDN serves multiple quality segments
- **Massive scale**: Billions of videos, petabytes of data

### Amazon
- **CloudFront**: AWS CDN, 400+ PoPs
- **S3 integration**: Origin for static assets
- **Lambda@Edge**: Run code at edge (redirect, auth, modify headers)
- **Shield**: DDoS protection integrated

### Cloudflare
- **310+ cities**: Global network
- **Free tier**: CDN, DDoS, SSL
- **Argo Smart Routing**: Optimized path selection
- **Workers**: Edge compute (like Lambda@Edge)

### Akamai
- **350K+ servers**: 130+ countries
- **~30% of web traffic**: One of largest
- **Enterprise focus**: Media, gaming, finance

---

## 3. Architecture Diagrams

### CDN Topology

```
                    +------------------+
                    |   DNS (GeoDNS)   |
                    |   User -> Edge   |
                    +--------+---------+
                             |
         +-------------------+-------------------+
         |                   |                   |
         v                   v                   v
+----------------+  +----------------+  +----------------+
|  Edge PoP 1    |  |  Edge PoP 2    |  |  Edge PoP 3    |
|  (US-East)    |  |  (EU-West)     |  |  (APAC)        |
|  +---------+  |  |  +---------+  |  |  +---------+  |
|  | Cache   |  |  |  | Cache   |  |  |  | Cache   |  |
|  +---------+  |  |  +---------+  |  |  +---------+  |
+-------+-------+  +-------+-------+  +-------+--------+
        |                   |                   |
        |    Cache Miss     |    Cache Miss     |
        |                   |                   |
        +-------------------+-------------------+
                            |
                            v
                    +------------------+
                    |  Shield / Mid     |
                    |  (Regional Cache) |
                    +--------+---------+
                             |
                             | Cache Miss
                             v
                    +------------------+
                    |  ORIGIN SERVER    |
                    |  (Your servers)   |
                    +------------------+
```

### Push vs Pull CDN

```
PULL CDN (On-Demand):
User requests /image.jpg
    -> Edge: Cache miss
    -> Edge fetches from Origin
    -> Origin returns
    -> Edge caches
    -> Edge returns to User

PUSH CDN (Pre-provisioned):
You upload /image.jpg to CDN
    -> CDN distributes to all edges
    -> User requests /image.jpg
    -> Edge: Cache hit (already there)
    -> Edge returns to User
```

### Cache Hierarchy

```
+--------+     +--------+     +--------+
| Edge   |     | Edge   |     | Edge   |
| (L1)   |     | (L1)   |     | (L1)   |
+---+----+     +---+----+     +---+----+
    |              |              |
    |    Miss      |    Miss      |
    +--------------+--------------+
                   |
                   v
            +--------+--------+
            |  Shield / Mid   |
            |  (L2)           |
            +--------+--------+
                   |
                   | Miss
                   v
            +--------+--------+
            |  Origin         |
            +-----------------+
```

### Adaptive Bitrate Streaming (Video CDN)

```
                    +------------------+
                    |   Video Player    |
                    +--------+---------+
                             |
         +-------------------+-------------------+
         |                   |                   |
         v                   v                   v
+----------------+  +----------------+  +----------------+
| 720p segment 1 |  | 480p segment 1 |  | 360p segment 1 |
| (2 Mbps)       |  | (1 Mbps)       |  | (500 Kbps)     |
+----------------+  +----------------+  +----------------+
         |                   |                   |
         +-------------------+-------------------+
                             |
                    Player selects based on
                    bandwidth, buffer, device
```

---

## 4. Core Mechanics

### Request Flow
1. **User** requests `https://cdn.example.com/image.jpg`
2. **DNS** resolves to nearest edge PoP (GeoDNS)
3. **Edge** checks cache:
   - **Hit**: Return immediately
   - **Miss**: Fetch from parent (shield) or origin
4. **Cache store**: Store response with TTL
5. **Return**: Send to user

### Cache Key
- **Default**: URL (host + path + query string)
- **Vary**: Can include headers (e.g., Accept-Encoding)
- **Custom**: Some CDNs allow custom cache keys (e.g., ignore query params)

### Cache Invalidation
- **TTL**: Expire after time (e.g., 1 hour, 24 hours)
- **Purge**: Invalidate specific URL or pattern
- **Versioned URLs**: `image.jpg?v=123` or `image.abc123.jpg`—change URL to bust cache

### Cache-Control Headers
- **max-age**: Seconds to cache
- **s-maxage**: Seconds for shared (CDN) cache
- **stale-while-revalidate**: Serve stale while fetching fresh
- **no-cache**: Revalidate before use
- **no-store**: Don't cache

---

## 5. Numbers

### Latency
| Scenario | Without CDN | With CDN |
|----------|-------------|----------|
| **US user, US origin** | 20-50 ms | 10-30 ms |
| **EU user, US origin** | 100-150 ms | 20-40 ms |
| **APAC user, US origin** | 200-300 ms | 30-60 ms |

### CDN Scale
| Provider | PoPs | Servers | Countries |
|----------|------|---------|-----------|
| **Akamai** | 4,100+ | 350,000+ | 130+ |
| **Cloudflare** | 310+ | 200+ | 100+ |
| **CloudFront** | 400+ | - | 90+ |
| **Fastly** | 70+ | - | - |

### Cache Hit Ratios
- **Static assets**: 90-99% (JS, CSS, images)
- **HTML**: 50-80% (personalized, dynamic)
- **API**: 0-10% (typically not cached)

### Bandwidth Savings
- **Typical**: 80-95% of requests served from edge
- **Origin load**: 5-20% of total traffic

---

## 6. Tradeoffs

### Push vs Pull

| Aspect | Push CDN | Pull CDN |
|--------|----------|----------|
| **Population** | You upload | On first request |
| **Use case** | Known content, pre-release | Dynamic, unpredictable |
| **Storage** | Pay for storage at edge | Pay per request (origin) |
| **Freshness** | You control | TTL-based |
| **Examples** | Netflix pre-cache | CloudFront, Cloudflare |

### TTL Tradeoffs

| TTL | Pros | Cons |
|-----|------|------|
| **Short (60s)** | Fresh content | More origin load, higher latency |
| **Long (24h)** | Low origin load | Stale content risk |
| **Versioned URLs** | Long TTL, instant freshness | Requires URL change |

### Single vs Multi-CDN

| Strategy | Pros | Cons |
|----------|------|------|
| **Single** | Simpler, volume discounts | Vendor lock-in, single point of failure |
| **Multi** | Resilience, best-of-breed | Complexity, cost |

---

## 7. Variants / Implementations

### Managed CDNs
- **CloudFront**: AWS, S3/ELB origin, Lambda@Edge
- **Cloudflare**: Free tier, Workers, DDoS
- **Fastly**: Real-time purge, VCL
- **Akamai**: Enterprise, media
- **Bunny CDN**: Low cost

### Custom CDNs
- **Netflix Open Connect**: Custom hardware, ISP partnerships
- **YouTube GGC**: Google infrastructure
- **Facebook**: Built for photos, video

### Video CDN
- **HLS/DASH**: Adaptive bitrate segments
- **Chunked transfer**: Segment-based caching
- **Live vs VOD**: Live has different caching (sliding window)

---

## 8. Scaling Strategies

### Cache Warming
- **Pre-populate**: Request popular content before users do
- **Crawl**: Crawl sitemap, pre-fetch
- **Predict**: Pre-populate based on trending

### Multi-CDN
- **Failover**: Primary CDN down -> switch to secondary
- **Load balancing**: Split traffic (e.g., 70/30)
- **Geo-optimization**: Different CDN per region

### Edge Compute
- **Lambda@Edge, Cloudflare Workers**: Run code at edge
- **Use cases**: A/B testing, redirects, auth, personalization
- **Latency**: Sub-ms for simple logic

### Dynamic Content
- **Cache API responses**: Short TTL, vary by user
- **Edge-side includes (ESI)**: Cache fragments, assemble at edge
- **Stale-while-revalidate**: Serve stale, update in background

---

## 9. Failure Scenarios

### CDN Outage
- **Fastly (2021)**: Outage took down many sites
- **Mitigation**: Multi-CDN, origin fallback
- **Netflix**: Open Connect isolated from third-party CDN

### Origin Overload
- **Scenario**: CDN cache miss storm (e.g., all edges miss)
- **Mitigation**: Cache warming, longer TTL, origin scaling

### Stale Content
- **Scenario**: Deploy new version, CDN serves old
- **Mitigation**: Versioned URLs, purge, short TTL for HTML

### DDoS
- **CDN as shield**: Absorb attack at edge
- **Cloudflare, Akamai**: DDoS protection
- **Mitigation**: Rate limiting, challenge (CAPTCHA)

### Real Incidents
- **Cloudflare (2019)**: Configuration error caused global outage
- **Fastly (2021)**: Single config change caused 1-hour outage
- **Lesson**: Multi-CDN, failover to origin

---

## 10. Performance Considerations

### Optimization Techniques
1. **HTTP/2, HTTP/3**: CDN supports; enable for multiplexing
2. **Brotli/Gzip**: Compression at edge
3. **Image optimization**: Resize, WebP at edge
4. **Minification**: JS/CSS at edge (or build time)

### Cache Headers
```
Cache-Control: public, max-age=31536000, immutable  # Static assets
Cache-Control: public, max-age=3600, stale-while-revalidate=86400  # Semi-static
Cache-Control: no-cache  # Dynamic (revalidate always)
```

### Monitoring
- **Cache hit ratio**: Target 90%+ for static
- **Origin load**: Should be low
- **Latency**: P50, P95, P99 by region
- **Error rate**: 4xx, 5xx

---

## 11. Use Cases

| Use Case | CDN Strategy | TTL |
|----------|--------------|-----|
| **Static assets** | Long TTL, versioned URLs | 1 year |
| **Images** | Long TTL, image optimization | 1 week - 1 year |
| **HTML** | Short TTL or no-cache | 0-60s |
| **API** | Rarely cache | 0 |
| **Video VOD** | Segment caching, long TTL | 24h+ |
| **Video live** | Sliding window | Seconds |
| **Software downloads** | Long TTL | 24h+ |

---

## 12. Comparison Tables

### CDN Providers

| Provider | Strengths | Weaknesses | Best For |
|----------|-----------|------------|----------|
| **CloudFront** | AWS integration | Less features | AWS workloads |
| **Cloudflare** | Free, DDoS, Workers | Less control | General use |
| **Fastly** | Real-time purge, VCL | Cost | Developers |
| **Akamai** | Scale, media | Cost | Enterprise |
| **Bunny** | Low cost | Smaller network | Budget |

### Cache Invalidation Methods

| Method | Speed | Granularity | Use Case |
|--------|-------|-------------|----------|
| **TTL** | Automatic | Per-object | Default |
| **Purge** | Manual, instant | URL, pattern | Emergency |
| **Versioned URL** | Instant | Per-deploy | Best practice |

---

## 13. Code or Pseudocode

### Cache Logic (Edge)

```python
def handle_request(request):
    cache_key = get_cache_key(request.url, request.headers)
    
    # Check edge cache
    cached = edge_cache.get(cache_key)
    if cached and not cached.expired():
        return cached.response
    
    # Check shield
    shield_response = fetch_from_shield(request)
    if shield_response:
        edge_cache.set(cache_key, shield_response, ttl=shield_response.ttl)
        return shield_response
    
    # Fetch from origin
    origin_response = fetch_from_origin(request)
    edge_cache.set(cache_key, origin_response, ttl=origin_response.ttl)
    return origin_response
```

### Versioned URL (Best Practice)

```html
<!-- Instead of -->
<link href="/style.css" rel="stylesheet">

<!-- Use -->
<link href="/style.css?v=abc123" rel="stylesheet">
<!-- Or -->
<link href="/style.abc123.css" rel="stylesheet">
```

### Purge API (Pseudocode)

```python
# CloudFront invalidation
client.create_invalidation(
    DistributionId='E1234',
    InvalidationBatch={
        'Paths': {'Quantity': 2, 'Items': ['/images/*', '/api/static/*']},
        'CallerReference': str(uuid.uuid4())
    }
)
```

### Cache-Control Headers (Origin)

```python
# Static asset
response.headers['Cache-Control'] = 'public, max-age=31536000, immutable'

# HTML (dynamic)
response.headers['Cache-Control'] = 'no-cache, must-revalidate'

# API (no cache)
response.headers['Cache-Control'] = 'no-store'
```

---

## 14. Interview Discussion

### Key Points to Cover
1. **Purpose**: Reduce latency, bandwidth, origin load
2. **How it works**: DNS -> edge -> cache hit/miss -> origin
3. **Push vs pull**: Pre-populate vs on-demand
4. **Cache invalidation**: TTL, purge, versioned URLs
5. **Multi-CDN**: Resilience, failover

### Common Follow-ups
- **"How would you design CDN for Netflix?"** → Custom, ISP partnerships, tiered cache, popular content at edge
- **"How to invalidate cache?"** → Versioned URLs (best), purge (emergency), TTL
- **"What if CDN is down?"** → Multi-CDN, failover to origin
- **"Cache dynamic content?"** → Short TTL, vary by user, ESI
- **"How does video CDN work?"** → Segments, adaptive bitrate, HLS/DASH

### Red Flags to Avoid
- Caching everything (no cache for API, user-specific)
- No cache invalidation strategy
- Forgetting origin can fail (design for CDN miss)
- Ignoring cost (egress, purge costs)

### Advanced Topics

#### Edge Compute (Lambda@Edge, Workers)
- **Run code at edge**: Before request hits origin (viewer request), after origin (viewer response)
- **Use cases**: Redirect, A/B test, auth, modify headers, generate response
- **Latency**: Sub-millisecond for simple logic
- **Limitations**: Cold start, execution time limits, no persistent storage

#### Cache Fragmentation
- **Problem**: Same content cached with different keys (query params, headers)
- **Solution**: Normalize cache key (ignore irrelevant params), use Vary carefully
- **Example**: `?utm_source=twitter` vs `?utm_source=facebook`—same content, different cache entries

#### DDoS Protection at CDN
- **Absorption**: CDN has large capacity; absorbs attack traffic
- **Scrubbing**: Filter malicious traffic before reaching origin
- **Challenge**: CAPTCHA, JS challenge for suspicious requests
- **Rate limiting**: Per-IP, per-ASN limits

#### Netflix Open Connect Deep Dive
- **Partnership model**: Netflix places OCA (appliance) in ISP datacenters for free
- **Benefit to ISP**: Traffic stays on ISP network, reduces transit costs
- **Content**: Popular titles pre-cached; long-tail from regional/origin
- **Scale**: 15,000+ OCAs globally, 15% of internet traffic at peak

#### Cache-Control Deep Dive
- **public**: Can be cached by CDN, browser
- **private**: Browser only (e.g., user-specific)
- **no-cache**: Must revalidate before use (can store, but validate)
- **no-store**: Don't store anywhere
- **immutable**: Content won't change (e.g., versioned asset)—browser can skip revalidation

#### HLS/DASH Segment Caching
- **Segment size**: Typically 2-10 seconds of video
- **Manifest**: M3U8 (HLS) or MPD (DASH) lists segment URLs
- **Caching**: Each segment cached independently; manifest has short TTL
- **Adaptive**: Player selects quality based on bandwidth; CDN serves requested segment
