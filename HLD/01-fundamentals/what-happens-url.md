# What Happens When You Type a URL

## 1. Concept Overview

### Definition
When you type a URL (e.g., `https://www.google.com`) and press Enter, a complex chain of events occurs across the browser, operating system, network, and servers. This journey encompasses **DNS resolution**, **TCP connection**, **TLS handshake**, **HTTP request/response**, and **browser rendering**. Understanding this end-to-end flow is fundamental to system design and a classic interview question.

### Purpose
- **Comprehensive understanding**: Map the full request lifecycle
- **Performance optimization**: Identify bottlenecks at each stage
- **Debugging**: Trace where failures occur
- **Interview readiness**: Structured answer for "What happens when you type a URL?"

### Problems It Solves
- **Visibility**: Understand what happens behind the scenes
- **Latency breakdown**: Know where time is spent (DNS, TCP, TLS, etc.)
- **Troubleshooting**: Isolate DNS vs network vs server issues

---

## 2. Real-World Motivation

### Why This Matters
- **Google**: Billions of searches; every ms of latency matters
- **Netflix**: Video streaming; DNS to nearest CDN, TCP tuning, HTTP/2
- **Amazon**: 100ms delay = 1% sales drop (research)
- **Every web company**: URL → page load is the critical path

### Interview Context
- **Most asked**: "What happens when you type google.com and press Enter?"
- **Tests**: Understanding of DNS, TCP, HTTP, browser internals
- **Depth**: Can you go deep on any step? (DNS resolution, TLS 1.3, etc.)

---

## 3. Architecture Diagrams

### Full Journey Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     WHAT HAPPENS WHEN YOU TYPE A URL                       │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   1. URL PARSED        2. DNS RESOLUTION    3. TCP CONNECTION            │
│   ┌─────────────┐      ┌─────────────┐      ┌─────────────┐             │
│   │ https://    │      │ www.google  │      │ 3-way       │             │
│   │ www.google  │ ───► │ .com → IP   │ ───► │ handshake   │             │
│   │ .com/search │      │ 142.250.80  │      │ SYN,ACK,SYN │             │
│   └─────────────┘      └─────────────┘      └─────────────┘             │
│         │                      │                      │                  │
│         │                      │                      │                  │
│   4. TLS HANDSHAKE      5. HTTP REQUEST      6. SERVER PROCESSING         │
│   ┌─────────────┐      ┌─────────────┐      ┌─────────────┐             │
│   │ ClientHello │      │ GET /search │      │ Parse req   │             │
│   │ ServerHello │ ───► │ Host, etc.  │ ───► │ Query DB   │             │
│   │ Cert, Keys  │      │             │      │ Render      │             │
│   └─────────────┘      └─────────────┘      └─────────────┘             │
│         │                      │                      │                  │
│         │                      │                      │                  │
│   7. HTTP RESPONSE      8. BROWSER RENDERING                              │
│   ┌─────────────┐      ┌─────────────┐                                  │
│   │ 200 OK      │      │  Parse HTML │                                   │
│   │ HTML, CSS   │ ───► │  DOM, CSSOM │                                   │
│   │ JS          │      │  Render tree │                                   │
│   └─────────────┘      │  Layout,Paint│                                  │
│                        └─────────────┘                                   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### DNS Resolution (Recursive)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     DNS RECURSIVE RESOLUTION                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   BROWSER          OS/STUB         RECURSIVE RESOLVER      AUTHORITATIVE │
│   ┌─────┐          ┌─────┐         ┌─────────────────┐      ┌──────────┐  │
│   │     │ 1.       │     │ 2.      │                 │ 3.   │          │  │
│   │     │─────────►│     │────────►│  ISP / 8.8.8.8  │─────►│  Root    │  │
│   │     │  www.    │     │  Query  │                 │      │  .com    │  │
│   │     │  google  │     │         │ 4. Query .com  ◄─┼──────│  google  │  │
│   │     │  .com    │     │         │ 5. Query google │      │  NS      │  │
│   │     │          │     │         │ 6. Get A record │      │          │  │
│   │     │◄─────────│     │◄────────│ 7. Return IP    │      │          │  │
│   │     │ 8. IP    │     │         │                 │      │          │  │
│   └─────┘          └─────┘         └─────────────────┘      └──────────┘  │
│                                                                          │
│   Caching: Browser → OS → Resolver (each layer caches by TTL)             │
│   Typical: 0-200ms (cached: 0ms; uncached: 50-200ms)                     │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### TCP 3-Way Handshake

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     TCP 3-WAY HANDSHAKE                                  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   CLIENT                                    SERVER                       │
│   ┌─────┐                                    ┌─────┐                    │
│   │     │  SYN (seq=x)                        │     │                    │
│   │     │───────────────────────────────────►│     │                    │
│   │     │                                    │     │                    │
│   │     │  SYN-ACK (seq=y, ack=x+1)          │     │                    │
│   │     │◄───────────────────────────────────│     │                    │
│   │     │                                    │     │                    │
│   │     │  ACK (ack=y+1)                     │     │                    │
│   │     │───────────────────────────────────►│     │                    │
│   │     │                                    │     │                    │
│   │     │  Connection ESTABLISHED            │     │                    │
│   └─────┘                                    └─────┘                    │
│                                                                          │
│   RTT: 1 round trip (typically 20-100ms depending on distance)           │
│   HTTPS: Another RTT for TLS handshake (or 0 with TLS 1.3 resumption)    │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### TLS 1.3 Handshake

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     TLS 1.3 HANDSHAKE (1-RTT)                            │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   CLIENT                                    SERVER                       │
│   ┌─────┐                                    ┌─────┐                    │
│   │     │  ClientHello                       │     │                    │
│   │     │  - Supported groups                │     │                    │
│   │     │  - Key share (optional)           │     │                    │
│   │     │──────────────────────────────────►│     │                    │
│   │     │                                    │     │                    │
│   │     │  ServerHello                       │     │                    │
│   │     │  - Selected group                 │     │                    │
│   │     │  - Key share                      │     │                    │
│   │     │  EncryptedExtensions              │     │                    │
│   │     │  Certificate                      │     │                    │
│   │     │  CertificateVerify                │     │                    │
│   │     │  Finished                         │     │                    │
│   │     │◄──────────────────────────────────│     │                    │
│   │     │                                    │     │                    │
│   │     │  Finished                          │     │                    │
│   │     │──────────────────────────────────►│     │                    │
│   │     │                                    │     │                    │
│   │     │  Encrypted application data       │     │                    │
│   └─────┘                                    └─────┘                    │
│                                                                          │
│   TLS 1.3: 1 RTT (vs TLS 1.2: 2 RTTs)                                    │
│   Resumption: 0 RTT (session ticket)                                     │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Browser Rendering Pipeline

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     BROWSER RENDERING PIPELINE                            │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   HTML                    CSS                     JavaScript               │
│   ┌─────┐                 ┌─────┐                 ┌─────┐                 │
│   │     │                 │     │                 │     │                 │
│   │     │                 │     │                 │     │                 │
│   └──┬──┘                 └──┬──┘                 └──┬──┘                 │
│      │                       │                       │                    │
│      ▼                       ▼                       ▼                    │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │  DOM (Document Object Model)     CSSOM (CSS Object Model)       │   │
│   │  Tree of HTML elements             Tree of styles                │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│      │                       │                                            │
│      └───────────┬───────────┘                                            │
│                  ▼                                                         │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │  RENDER TREE: DOM + CSSOM (only visible elements)                │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│                  │                                                         │
│                  ▼                                                         │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │  LAYOUT: Calculate size, position of each element                 │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│                  │                                                         │
│                  ▼                                                         │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │  PAINT: Convert to pixels (layers)                                │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│                  │                                                         │
│                  ▼                                                         │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │  COMPOSITE: GPU layers → final image                             │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Core Mechanics

### Step-by-Step (Detailed)

1. **URL Parsing**
   - Browser parses URL: scheme (https), host (www.google.com), path (/search)
   - Check HSTS list (preload for HTTPS)
   - Check for port (default 443 for HTTPS)

2. **DNS Resolution**
   - Check browser cache → OS cache → stub resolver
   - Recursive resolver: Root → TLD → Authoritative
   - Returns A/AAAA record (IP address)
   - Caching at each layer (TTL)

3. **TCP Connection**
   - 3-way handshake: SYN, SYN-ACK, ACK
   - 1 RTT
   - Connection established; ready for TLS

4. **TLS Handshake**
   - ClientHello: Cipher suites, extensions
   - ServerHello: Selected cipher, certificate
   - Key exchange (ECDHE)
   - TLS 1.3: 1 RTT; TLS 1.2: 2 RTTs
   - Encrypted application data

5. **HTTP Request**
   - GET /search?q=... HTTP/1.1 or HTTP/2
   - Headers: Host, User-Agent, Accept, Cookie
   - HTTP/2: Multiplexing, header compression
   - Request sent over TLS

6. **Server Processing**
   - Load balancer → Web server → App server
   - Parse request; route to handler
   - Query DB, cache; generate response
   - Return HTML, CSS, JS

7. **HTTP Response**
   - 200 OK; Content-Type: text/html
   - Body: HTML document
   - Headers: Cache-Control, Set-Cookie

8. **Browser Rendering**
   - Parse HTML → DOM
   - Parse CSS → CSSOM
   - Execute JS (may block parsing)
   - Render tree → Layout → Paint → Composite
   - Display page

9. **Additional Resources**
   - HTML references CSS, JS, images
   - Each triggers new requests (or reuse connection with HTTP/2)
   - DNS may be cached; TCP/TLS may reuse connection

---

## 5. Numbers

| Step | Latency (Typical) | Cached |
|------|-------------------|--------|
| **DNS** | 20-200 ms | 0 ms |
| **TCP** | 20-100 ms (1 RTT) | 0 (reuse) |
| **TLS** | 20-100 ms (1 RTT) | 0 (resumption) |
| **HTTP request** | 50-500 ms | 10-100 ms |
| **Server processing** | 10-500 ms | - |
| **Rendering** | 50-500 ms | - |
| **Total (first load)** | 200-2000 ms | - |
| **Total (cached)** | 50-200 ms | - |

### HTTP/2 Multiplexing
- **HTTP/1.1**: 6 connections per origin; head-of-line blocking
- **HTTP/2**: 1 connection; multiple streams; no HOL blocking
- **Benefit**: Parallel requests over single connection

---

## 6. Tradeoffs

### DNS: Recursive vs Iterative
- **Recursive**: Client delegates to resolver; resolver does all work
- **Iterative**: Resolver asks each level; returns referral
- **Typical**: Recursive (client uses resolver)

### HTTP/1.1 vs HTTP/2
- **HTTP/1.1**: Simple; one request per connection (or pipelining)
- **HTTP/2**: Multiplexing; header compression; server push
- **HTTP/3**: QUIC; UDP; 0-RTT handshake

### Caching Tradeoffs
- **Aggressive cache**: Fast repeat visits; stale data risk
- **No cache**: Fresh data; slower
- **TTL**: Balance freshness vs performance

---

## 7. Variants / Implementations

### DNS
- **DoH** (DNS over HTTPS): Encrypt DNS; port 443
- **DoT** (DNS over TLS): Encrypt DNS; port 853
- **EDNS**: Client subnet for geo-routing

### TLS
- **TLS 1.2**: 2 RTT handshake
- **TLS 1.3**: 1 RTT; 0-RTT resumption
- **OCSP stapling**: Certificate validation in handshake

### HTTP
- **HTTP/1.1**: Persistent connections, chunked encoding
- **HTTP/2**: Binary, multiplexing, HPACK
- **HTTP/3**: QUIC, UDP-based

---

## 8. Scaling Strategies

- **DNS**: CDN; anycast; geo-routing
- **TCP**: Connection pooling; keep-alive
- **TLS**: Session resumption; TLS 1.3
- **Server**: Load balancer; CDN; caching
- **Rendering**: Code splitting; lazy load; critical CSS

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|------------|
| **DNS failure** | No resolution | Fallback resolver; retry |
| **TCP timeout** | No connection | Retry; check firewall |
| **TLS error** | Certificate invalid | Fix cert; renew |
| **HTTP 5xx** | Server error | Retry; circuit breaker |
| **Slow server** | Slow page load | CDN; caching; optimize |

---

## 10. Performance Considerations

- **DNS**: Prefetch; preconnect; cache
- **TCP**: Keep-alive; connection reuse
- **TLS**: 1.3; resumption; OCSP stapling
- **HTTP**: HTTP/2; compression; caching
- **Rendering**: Critical path; defer JS; async; lazy load

---

## 11. Use Cases

| Use Case | Optimization |
|----------|---------------|
| **First visit** | DNS prefetch; CDN; minimize round trips |
| **Repeat visit** | Cache DNS, TCP, TLS; cache assets |
| **Mobile** | Reduce RTT; fewer requests; compress |
| **E-commerce** | Fast TTFB; critical rendering path |
| **Video** | CDN; adaptive bitrate; range requests |

---

## 12. Comparison Tables

### Latency Breakdown (Typical)

| Step | First Load | Cached |
|------|------------|--------|
| DNS | 50-200 ms | 0 |
| TCP | 30-80 ms | 0 |
| TLS | 30-80 ms | 0 |
| Request | 50-300 ms | 20-100 ms |
| **Total (before render)** | **160-660 ms** | **20-100 ms** |

### HTTP Versions

| Version | Multiplexing | Compression | RTT |
|---------|--------------|-------------|-----|
| HTTP/1.1 | No | Optional | 1 per request |
| HTTP/2 | Yes | HPACK | 1 |
| HTTP/3 | Yes | QPACK | 0 (QUIC) |

---

## 13. Code or Pseudocode

### DNS Resolution (Pseudocode)

```python
def resolve(domain: str) -> str:
    # 1. Check cache
    if domain in cache and not expired(cache[domain]):
        return cache[domain].ip
    
    # 2. Query recursive resolver
    ip = recursive_resolver.query(domain)
    
    # 3. Cache result
    cache[domain] = (ip, ttl)
    
    return ip
```

### TCP Handshake (Conceptual)

```python
# Client
send(SYN, seq=x)
recv(SYN-ACK, seq=y, ack=x+1)
send(ACK, ack=y+1)
# Connection established
```

### HTTP Request (Simplified)

```http
GET /search?q=system+design HTTP/1.1
Host: www.google.com
User-Agent: Mozilla/5.0 ...
Accept: text/html
Cookie: session=abc123
```

---

## 14. Interview Discussion

### Key Points to Articulate
1. **DNS**: Resolve hostname to IP; recursive; caching
2. **TCP**: 3-way handshake; 1 RTT
3. **TLS**: Encrypt; 1 RTT (TLS 1.3)
4. **HTTP**: Request/response; HTTP/2 multiplexing
5. **Rendering**: DOM, CSSOM, render tree, layout, paint
6. **Caching**: At each layer; repeat visits faster

### Ideal Answer Structure (2-3 minutes)

1. **URL Parsing**: Browser parses scheme, host, path
2. **DNS**: Check cache; recursive resolution (Root→TLD→Auth); get IP
3. **TCP**: 3-way handshake; establish connection
4. **TLS**: Handshake; encrypt channel
5. **HTTP**: Send GET request; server processes; returns HTML
6. **Rendering**: Parse HTML→DOM; CSS→CSSOM; render tree; layout; paint; composite
7. **Optional**: Additional resources (CSS, JS, images); repeat for each (or HTTP/2 multiplex)

### Common Follow-ups
- **"How does DNS work?"** → Recursive vs iterative; caching; record types
- **"What is the 3-way handshake?"** → SYN, SYN-ACK, ACK; why
- **"TLS 1.3 vs 1.2?"** → 1 RTT vs 2; 0-RTT resumption
- **"What is HTTP/2 multiplexing?"** → Multiple streams over one connection
- **"Critical rendering path?"** → DOM, CSSOM, render tree, layout, paint; block JS
- **"How can you optimize?"** → DNS prefetch, HTTP/2, CDN, caching, minify

### Red Flags to Avoid
- Skipping DNS; going straight to TCP
- Not mentioning TLS for HTTPS
- Ignoring browser rendering
- Vague on any step

### Deep Dive Topics (if asked)
- **DNS**: Recursive vs iterative; anycast; DNSSEC
- **TCP**: Congestion control; slow start
- **TLS**: Certificate chain; ECDHE; session resumption
- **HTTP/2**: Frames; streams; priorities
- **Rendering**: Reflow; repaint; compositing
