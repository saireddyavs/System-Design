# Domain Name System (DNS)

## 1. Concept Overview

### Definition
The Domain Name System (DNS) is a hierarchical, distributed database that translates human-readable domain names (e.g., `www.google.com`) into machine-readable IP addresses (e.g., `142.250.80.46`). It is the phonebook of the internet.

### Purpose
- **Name Resolution**: Map domain names to IP addresses
- **Service Discovery**: Find services within a domain (e.g., mail servers, SIP servers)
- **Load Distribution**: Route users to geographically optimal servers
- **Failover**: Redirect traffic when primary servers fail
- **Abstraction**: Decouple service endpoints from infrastructure (change IPs without changing URLs)

### Problems It Solves
1. **Human Memory**: Humans remember names, not 32/128-bit numbers
2. **Infrastructure Flexibility**: IPs change; names provide stable identifiers
3. **Scalability**: Distributed design handles billions of lookups daily
4. **Geographic Routing**: Same domain can resolve to different IPs based on location
5. **Service Decoupling**: Multiple services under one domain (www, api, mail)

---

## 2. Real-World Motivation

### Google
- **8.8.8.8 / 8.8.4.4**: Public DNS resolvers handling 70B+ queries/day
- **Anycast**: Same IP advertised from 200+ locations globally
- **GeoDNS**: `google.com` resolves to nearest datacenter (e.g., Singapore vs Virginia)
- **Latency**: Sub-10ms resolution for cached entries

### Netflix
- **Open Connect CDN**: Custom DNS routes users to nearest Open Connect Appliance (OCA)
- **Regional Routing**: `netflix.com` → different IPs in US-East, EU-West, APAC
- **A/B Testing**: DNS used for canary deployments (subset of users get new IP)
- **Failover**: Automatic reroute if OCA is overloaded

### Uber
- **Microservice Discovery**: Internal DNS for service-to-service communication
- **Regional Endpoints**: API endpoints vary by region for compliance
- **Real-time Routing**: Driver app resolves to nearest dispatch server

### Amazon
- **Route 53**: Managed DNS with 100% SLA, health checks, latency-based routing
- **S3/CloudFront**: `*.s3.amazonaws.com` resolves to regional endpoints
- **Multi-AZ**: DNS failover between availability zones

### Twitter
- **Traffic Management**: DNS-based load balancing across datacenters
- **DDoS Mitigation**: Route 53 + Shield for attack absorption
- **Propagation**: Careful TTL management during migrations

---

## 3. Architecture Diagrams

### DNS Hierarchy

```
                    +------------------+
                    |   ROOT SERVERS   |
                    |   (13 clusters)  |
                    |  a.root-servers  |
                    |  .net ...        |
                    +--------+---------+
                             |
              +--------------+--------------+
              |              |              |
       +------v------+ +-----v-----+ +-----v-----+
       |  .com TLD   | |  .org TLD | |  .net TLD |
       |  (Verisign) | |           | |           |
       +------+------+ +-----------+ +-----------+
              |
       +------v------+
       |  google.com |
       |  (Authoritative)|
       +------+------+
              |
       +------v------+
       | www.google  |
       | A: 142.250..|
       +-------------+
```

### Step-by-Step Resolution Flow

```
CLIENT                    RESOLVER                 ROOT                 TLD                  AUTHORITATIVE
  |                          |                      |                    |                         |
  |-- www.example.com ------->|                      |                    |                         |
  |                          |-- .com? ------------->|                    |                         |
  |                          |<-- NS: a.gtld-servers|                    |                         |
  |                          |-- example.com? -------------------------->|                         |
  |                          |<-- NS: ns1.example.com --------------------|                         |
  |                          |-- www.example.com? ------------------------------------------------>|
  |                          |<-- A: 93.184.216.34 ------------------------------------------------|
  |<-- 93.184.216.34 --------|                      |                    |                         |
```

### Recursive vs Iterative Resolution

```
RECURSIVE (Client delegates to resolver):
Client --> Resolver --> Root --> TLD --> Auth --> Resolver --> Client
         (Resolver does all the work, returns final answer)

ITERATIVE (Resolver asks each level):
Client --> Resolver --> Root (returns TLD ref)
         Resolver --> TLD (returns Auth ref)
         Resolver --> Auth (returns A record)
         Resolver --> Client
```

### DNS Caching Layers

```
+------------------+
|  Browser Cache   |  TTL: minutes to hours
+--------+---------+
         |
+--------v---------+
|  OS Cache        |  TTL: hours (e.g., Windows DNS Client, systemd-resolved)
+--------+---------+
         |
+--------v---------+
|  Stub Resolver   |  Local machine
+--------+---------+
         |
+--------v---------+
|  ISP/Recursive   |  TTL: seconds to days (largest cache)
|  Resolver        |
+--------+---------+
         |
+--------v---------+
|  Authoritative   |  Source of truth
|  Nameserver      |
+------------------+
```

---

## 4. Core Mechanics

### DNS Message Format (RFC 1035)
- **Header**: 12 bytes (ID, flags, QDCOUNT, ANCOUNT, NSCOUNT, ARCOUNT)
- **Question**: QNAME (domain), QTYPE, QCLASS
- **Answer/Authority/Additional**: Resource records

### Resolution Process (Detailed)
1. **Stub Resolver** receives query from application
2. **Check Local Cache**: Browser → OS → No cache
3. **Send to Recursive Resolver**: Typically ISP or 8.8.8.8
4. **Recursive Resolver** (if uncached):
   - Query root servers (hardcoded 13 IPs)
   - Get TLD nameserver referral
   - Query TLD for authoritative NS
   - Query authoritative for final record
5. **Cache Result** at each layer based on TTL
6. **Return to Client**

### Record Types Deep Dive

| Type | Purpose | Example |
|------|---------|---------|
| **A** | IPv4 address | `www.example.com → 93.184.216.34` |
| **AAAA** | IPv6 address | `www.example.com → 2606:2800:220:1:248:1893:25c8:1946` |
| **CNAME** | Canonical name (alias) | `www → example.com` |
| **MX** | Mail exchange | `example.com → mail.example.com (priority 10)` |
| **NS** | Nameserver delegation | `example.com → ns1.example.com` |
| **TXT** | Text records | SPF, DKIM, verification |
| **SRV** | Service discovery | `_http._tcp.example.com → host:port` |
| **SOA** | Start of authority | Zone metadata, serial, refresh |
| **PTR** | Reverse lookup | `34.216.184.93.in-addr.arpa → www.example.com` |

### CNAME Flattening
- CNAME adds extra lookup (alias → canonical → A)
- **CNAME flattening** (CloudFlare, Route 53): Return A record directly at alias to reduce latency
- Critical for root domain (e.g., `example.com`) which cannot have CNAME per RFC

---

## 5. Numbers

### Latency
| Scenario | Latency | Notes |
|----------|---------|-------|
| **Cached (browser)** | 0-1 ms | Instant |
| **Cached (OS)** | 0-2 ms | Negligible |
| **Cached (ISP)** | 5-20 ms | Depends on resolver proximity |
| **Uncached (full resolution)** | 50-200 ms | 4-6 round trips |
| **Cold (no cache anywhere)** | 100-500 ms | Worst case |

### Throughput
- **Root servers**: ~100K QPS each (anycast distributes)
- **Google Public DNS**: 70B+ queries/day (~800K QPS average)
- **CloudFlare 1.1.1.1**: 20B+ queries/day
- **Single authoritative server**: 100K-1M QPS with proper setup

### TTL Guidelines
| Use Case | TTL | Rationale |
|----------|-----|-----------|
| **Static content** | 86400 (24h) | Rarely changes |
| **Dynamic/failover** | 60-300 | Quick propagation |
| **Migration** | 5-60 | Minimize stale data |
| **CDN edge** | 3600-86400 | Balance freshness vs load |

### Propagation
- **Typical propagation**: 24-48 hours globally (due to TTL)
- **With low TTL (60s)**: Most resolvers updated in 5-10 minutes
- **Root/TLD changes**: Can take up to 48h (longer TTLs at top)

---

## 6. Tradeoffs

### Recursive vs Iterative

| Aspect | Recursive | Iterative |
|--------|-----------|-----------|
| **Client load** | Minimal | Higher (if client implements) |
| **Resolver load** | High | Lower |
| **Caching** | Centralized at resolver | Distributed |
| **Typical use** | End users | Rare (mostly recursive) |

### Short vs Long TTL

| TTL | Pros | Cons |
|-----|------|------|
| **Short (60s)** | Fast failover, quick propagation | More authoritative load, higher latency |
| **Long (24h)** | Reduced load, better caching | Slow failover, stale data risk |

### GeoDNS Strategies

| Strategy | Use Case | Complexity |
|----------|----------|------------|
| **Latency-based** | Route to nearest DC | Medium |
| **GeoIP-based** | Compliance, localization | Low |
| **Weighted** | Gradual migration | Medium |
| **Failover** | HA, disaster recovery | Low |

---

## 7. Variants / Implementations

### Managed DNS Services
- **AWS Route 53**: Health checks, latency routing, 100% SLA
- **CloudFlare DNS**: Free tier, DDoS protection, DNSSEC
- **Google Cloud DNS**: Integration with GCP
- **Azure DNS**: Private DNS zones, Azure integration
- **NS1**: Filter chains, advanced traffic management

### DNS-over-HTTPS (DoH)
- Encrypts DNS over HTTPS (port 443)
- Hides DNS queries from ISP
- Used by Firefox, Chrome (optional)
- Resolvers: Cloudflare 1.1.1.1, Google 8.8.8.8

### DNS-over-TLS (DoT)
- Encrypts DNS over TLS (port 853)
- Similar to DoH, different port
- Used by Android 9+

### DNSSEC
- Cryptographic signing of DNS records
- Prevents cache poisoning, spoofing
- Adds RRSIG, NSEC records
- Validation adds ~1-5ms

---

## 8. Scaling Strategies

### Authoritative DNS Scaling
1. **Anycast**: Same IP from multiple locations, BGP routes to nearest
2. **Sharding**: Split zones across nameservers
3. **Read replicas**: Multiple servers for same zone
4. **Rate limiting**: Prevent abuse

### Caching Strategy
- **Aggressive TTL** for static records
- **Negative caching**: Cache NXDOMAIN (typically 60-300s)
- **Prefetching**: Browser prefetches DNS for links
- **Connection coalescing**: Reuse connections (HTTP/2)

### Netflix-Style Geo Routing
1. User requests `netflix.com`
2. Resolver (often ISP) is geo-aware
3. Returns IP of nearest OCA
4. Or: Use EDNS Client Subnet (ECS) to pass client subnet to authoritative
5. Authoritative returns geo-optimized IP

---

## 9. Failure Scenarios

### Dyn DDoS (October 2016)
- **What**: Mirai botnet DDoS'd Dyn (major DNS provider)
- **Impact**: Twitter, GitHub, Netflix, Reddit, PayPal down for hours
- **Root cause**: 100K+ IoT devices, 1.2 Tbps traffic
- **Mitigation**: Multi-provider DNS, anycast, DDoS mitigation (Cloudflare, Akamai)

### Route 53 S3 Bucket Naming (2017)
- **What**: S3 bucket naming conflict with Route 53
- **Impact**: Brief AWS outage
- **Lesson**: DNS is critical path; single points of failure cascade

### DNS Cache Poisoning (Kaminsky, 2008)
- **What**: Attacker injects forged DNS responses
- **Mitigation**: Randomize source port, DNSSEC, use multiple resolvers

### Mitigation Strategies
| Failure | Mitigation |
|---------|------------|
| **Authoritative down** | Multiple NS records, anycast |
| **DDoS** | Anycast, DDoS scrubbing, multi-CDN |
| **Misconfiguration** | Change management, validation, rollback |
| **Propagation delay** | Lower TTL before change, monitor |
| **Cache poisoning** | DNSSEC, secure resolvers |

---

## 10. Performance Considerations

### Optimization Techniques
1. **CNAME flattening**: Eliminate extra lookup
2. **Response rate limiting**: Prevent amplification attacks
3. **TCP fallback**: For large responses (>512 bytes)
4. **EDNS0**: Larger UDP payload (4096 bytes)
5. **Prefetching**: Browser DNS prefetch, preconnect

### Connection Reuse
- HTTP/2 connection coalescing: Same IP for multiple hostnames
- Reduces DNS lookups for subresources
- Requires careful certificate (SAN) setup

### Monitoring
- **Resolution time**: P50, P95, P99
- **Failure rate**: NXDOMAIN, timeout, SERVFAIL
- **Cache hit ratio**: At resolver level
- **Propagation**: After changes, verify globally

---

## 11. Use Cases

| Use Case | DNS Feature | Example |
|----------|-------------|---------|
| **Web hosting** | A/AAAA records | `www.example.com → 1.2.3.4` |
| **CDN** | CNAME to CDN | `cdn.example.com → d1234.cloudfront.net` |
| **Email** | MX records | `example.com → mail.example.com` |
| **Load balancing** | Multiple A records | Round-robin by default |
| **Geo routing** | GeoDNS | Netflix, CloudFront |
| **Failover** | Health checks + DNS | Route 53 failover |
| **Service discovery** | Internal DNS | Kubernetes CoreDNS |
| **Subdomain delegation** | NS records | `blog.example.com → Blogger` |

---

## 12. Comparison Tables

### DNS Providers

| Provider | Strengths | Weaknesses | Best For |
|----------|-----------|------------|----------|
| **Route 53** | Health checks, integration | Cost at scale | AWS workloads |
| **CloudFlare** | Free, DDoS, fast | Less control | General use |
| **Google Cloud DNS** | GCP integration | Less features | GCP workloads |
| **NS1** | Filter chains, advanced | Complexity | Enterprise |
| **Akamai** | Scale, performance | Cost | Enterprise |

### Record Type Selection

| Need | Record | Notes |
|------|--------|-------|
| **IPv4** | A | Primary |
| **IPv6** | AAAA | Dual-stack |
| **Alias** | CNAME or ALIAS | CNAME: subdomain only |
| **Mail** | MX | Priority matters |
| **Delegation** | NS | Child zones |
| **Service discovery** | SRV | host:port:priority |

---

## 13. Code or Pseudocode

### Simple DNS Resolver (Pseudocode)

```python
def resolve(domain: str, record_type: str = "A") -> list:
    # 1. Check cache
    cached = cache.get((domain, record_type))
    if cached and not cached.expired():
        return cached.records
    
    # 2. Get root hints (hardcoded)
    nameservers = ROOT_SERVERS
    
    # 3. Iterative resolution
    while True:
        for ns in nameservers:
            response = query(ns, domain, record_type)
            if response.answer:
                cache.set((domain, record_type), response.answer, response.ttl)
                return response.answer
            if response.authority:
                nameservers = resolve_ns(response.authority)
                break
            if response.additional:
                nameservers = extract_ips(response.additional)
        else:
            raise ResolutionError("No answer found")
```

### Health Check + Failover (Route 53 style)

```yaml
# Route 53 Failover Configuration
primary:
  name: api.example.com
  type: A
  value: 1.2.3.4
  health_check: http://1.2.3.4/health
  failover: SECONDARY

secondary:
  name: api.example.com
  type: A
  value: 5.6.7.8
  health_check: http://5.6.7.8/health
  failover: PRIMARY
```

### GeoDNS Response Logic

```python
def get_geo_response(client_ip: str, domain: str) -> str:
    region = geoip_lookup(client_ip)
    
    if region in ["us-east", "us-west"]:
        return "10.0.1.1"  # US datacenter
    elif region in ["eu-west", "eu-central"]:
        return "10.0.2.1"  # EU datacenter
    elif region in ["ap-south", "ap-southeast"]:
        return "10.0.3.1"  # APAC datacenter
    else:
        return get_latency_based_ip(client_ip)  # Fallback
```

---

## 14. Interview Discussion

### Key Points to Cover
1. **Hierarchy**: Root → TLD → Authoritative (explain each level)
2. **Caching**: Where, why, TTL tradeoffs
3. **Recursive vs Iterative**: Who does the work
4. **Record types**: A, AAAA, CNAME, MX, NS—when to use each
5. **GeoDNS**: How companies route globally
6. **Failover**: Health checks + DNS
7. **Security**: DNSSEC, DoH, cache poisoning
8. **Scale**: Anycast, 70B queries/day

### Common Follow-ups
- **"How would you design DNS for a global application?"** → GeoDNS, anycast, multi-provider, health checks
- **"What if DNS is slow?"** → Caching, TTL, prefetching, connection coalescing
- **"How does Netflix route users?"** → Open Connect, GeoDNS, ISP partnerships
- **"Explain DNS propagation"** → TTL, caching layers, 24-48h typical
- **"What is DNSSEC?"** → Signing, chain of trust, prevents poisoning

### Red Flags to Avoid
- Saying "DNS is centralized" (it's distributed)
- Ignoring caching (critical for performance)
- Not mentioning TTL
- Confusing recursive and iterative
- Forgetting CNAME limitations (no CNAME at apex)

### Advanced Topics

#### EDNS (Extension Mechanisms for DNS)
- **EDNS0**: Extends DNS message size beyond 512 bytes
- **Buffer size**: Up to 4096 bytes (common)
- **ECS (Client Subnet)**: Pass client subnet to authoritative for geo-routing
- **Cookie**: Optional client/server cookie for DDoS mitigation

#### DNS over HTTPS (DoH) vs DNS over TLS (DoT)
- **DoH**: Port 443, looks like normal HTTPS—harder to block
- **DoT**: Port 853, dedicated—easier to identify and block
- **Privacy**: Both encrypt DNS; ISP cannot see queries
- **Controversy**: DoH can bypass enterprise DNS policies

#### Anycast Deep Dive
- **Same IP, multiple locations**: BGP advertises from many PoPs
- **Traffic flow**: User routed to nearest PoP by BGP
- **Failure**: If one PoP down, BGP withdraws route; traffic goes to next nearest
- **Used by**: Root servers, Google 8.8.8.8, Cloudflare 1.1.1.1
