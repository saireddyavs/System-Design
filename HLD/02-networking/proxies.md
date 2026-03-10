# Forward Proxy vs Reverse Proxy

## 1. Concept Overview

### Definition
A **proxy** is an intermediary server that sits between clients and other servers, forwarding requests and responses. A **forward proxy** acts on behalf of clients (hiding client identity, enabling access control). A **reverse proxy** acts on behalf of servers (hiding server identity, load balancing, SSL termination).

### Purpose
- **Forward Proxy**: Client-side—anonymity, bypass restrictions, caching, content filtering
- **Reverse Proxy**: Server-side—load balancing, SSL termination, caching, security, single entry point

### Problems They Solve
1. **Forward**: Access control, caching, anonymity, bypass geo-blocks
2. **Reverse**: Scalability, security (hide origin), SSL offload, routing, rate limiting

---

## 2. Real-World Motivation

### Google
- **Corporate forward proxy**: Employees browse through proxy for security, logging
- **Reverse proxy**: Front-end to GFE (Google Front End) for load balancing, TLS

### Netflix
- **Zuul**: Reverse proxy, API gateway—routing, auth, rate limiting
- **No forward proxy** for end users (CDN is different)

### Uber
- **API Gateway**: Reverse proxy for all API traffic
- **Corporate proxy**: Forward proxy for internal tools

### Amazon
- **ALB**: Reverse proxy (L7 load balancer)
- **CloudFront**: Reverse proxy + CDN
- **API Gateway**: Reverse proxy for serverless

### Cloudflare
- **Reverse proxy**: Sits in front of origin, DDoS, WAF, CDN
- **DNS, Workers**: Part of reverse proxy stack

### Corporate Networks
- **Squid, Blue Coat**: Forward proxy for employee internet access
- **Content filtering**: Block malicious sites
- **Caching**: Reduce bandwidth

---

## 3. Architecture Diagrams

### Forward Proxy (Client-Side)

```
    CLIENT                    FORWARD PROXY                 INTERNET
       |                            |                            |
       |  Request (to proxy)        |                            |
       |-------------------------->|                            |
       |                            |  Request (to origin)       |
       |                            |--------------------------->|
       |                            |                            |
       |                            |  Response                  |
       |                            |<---------------------------|
       |  Response                  |                            |
       |<---------------------------|                            |
       |                            |                            |

Client configures: "Use proxy at proxy.company.com:8080"
Client sends ALL requests to proxy; proxy fetches on behalf of client
Origin sees: Request from PROXY, not from client (client hidden)
```

### Reverse Proxy (Server-Side)

```
    CLIENT                    REVERSE PROXY                 ORIGIN SERVERS
       |                            |                            |
       |  Request (to example.com)   |                            |
       |-------------------------->|                            |
       |                            |  Request (to backend)      |
       |                            |--------------------------->|
       |                            |                            |
       |                            |  Response                  |
       |                            |<---------------------------|
       |  Response                  |                            |
       |<---------------------------|                            |
       |                            |                            |

Client sends to example.com (which IS the reverse proxy)
Client does NOT know about backend servers (servers hidden)
Proxy chooses backend, forwards request, returns response
```

### Side-by-Side Comparison

```
FORWARD PROXY:                          REVERSE PROXY:
Client -> [Proxy] -> Internet           Client -> [Proxy] -> Backends
         ^                                       ^
         |                                       |
    Client knows proxy                     Client doesn't know
    Server doesn't know client             Server knows proxy
    Use case: Hide client                  Use case: Hide servers
```

### Sidecar Proxy Pattern (Service Mesh)

```
+------------------+     +------------------+
|   Service A      |     |   Service B      |
|  +------------+   |     |  +------------+   |
|  |   App      |   |     |  |   App      |   |
|  +------+-----+   |     |  +------+-----+   |
|         |         |     |         |         |
|  +------v-----+   |     |  +------v-----+   |
|  |  Sidecar   |   |     |  |  Sidecar   |   |
|  |  (Envoy)   |<--+---->|  |  (Envoy)   |   |
|  +------------+   |     |  +------------+   |
+------------------+     +------------------+

Each service has a sidecar proxy
All traffic goes through sidecar: LB, retry, auth, metrics
```

### API Gateway as Reverse Proxy

```
                    +------------------+
                    |   API Gateway    |
                    |  (Reverse Proxy) |
                    +--------+---------+
                             |
         +-------------------+-------------------+
         |                   |                   |
         v                   v                   v
+----------------+  +----------------+  +----------------+
|  User Service  |  |  Order Service |  |  Payment Svc   |
+----------------+  +----------------+  +----------------+

Gateway: Auth, rate limit, routing, SSL termination
```

---

## 4. Core Mechanics

### Forward Proxy
- **Client configuration**: Proxy URL (host:port) in browser or system
- **Request flow**: Client -> Proxy -> Origin
- **Protocol**: HTTP CONNECT for HTTPS (tunnel), or HTTP for HTTP
- **Caching**: Proxy can cache responses (e.g., Squid)
- **Filtering**: Block by URL, category, content

### Reverse Proxy
- **Client configuration**: None—client connects to proxy as if it's the server
- **Request flow**: Client -> Proxy -> Backend(s)
- **Features**: Load balancing, SSL termination, caching, compression, routing
- **Backend selection**: Round robin, least connections, etc.

### HTTPS with Proxies
- **Forward proxy + HTTPS**: CONNECT method establishes tunnel; proxy doesn't decrypt (usually)
- **Reverse proxy + HTTPS**: Proxy terminates TLS, optionally re-encrypts to backend

---

## 5. Numbers

### Forward Proxy
- **Squid**: 10K+ concurrent connections per instance
- **Caching**: Can reduce bandwidth 30-50% for repeat requests
- **Latency**: Adds 1-10ms (cache hit) or full RTT (cache miss)

### Reverse Proxy
- **Nginx**: 50K+ concurrent, 10K+ RPS
- **HAProxy**: Millions of connections
- **Envoy**: 100K+ connections
- **Latency overhead**: 0.5-2ms typically

### API Gateway
- **AWS API Gateway**: 10K RPS default (can increase)
- **Kong**: 10K+ RPS per node
- **Latency**: 1-5ms overhead

---

## 6. Tradeoffs

### Forward vs Reverse Proxy

| Aspect | Forward Proxy | Reverse Proxy |
|--------|---------------|---------------|
| **Position** | Client-side | Server-side |
| **Client awareness** | Client configures proxy | Client unaware |
| **Hides** | Client identity | Server identity |
| **Use case** | Corporate, VPN, bypass | Load balancing, gateway |
| **Examples** | Squid, corporate proxy | Nginx, ALB, Zuul |

### When to Use Forward Proxy
- Corporate network (security, logging)
- Bypass geo-restrictions
- Anonymity (Tor is forward proxy chain)
- Caching in LAN (school, office)

### When to Use Reverse Proxy
- Load balancing
- SSL termination
- API gateway
- Caching (CDN-like)
- Rate limiting, WAF

---

## 7. Variants / Implementations

### Forward Proxy Software
- **Squid**: Open source, caching, ACL
- **Blue Coat (Broadcom)**: Enterprise
- **Shadowsocks**: Bypass restrictions
- **Tor**: Onion routing (chain of proxies)

### Reverse Proxy Software
- **Nginx**: High performance, L7
- **HAProxy**: L4/L7, mature
- **Envoy**: Service mesh, dynamic config
- **Caddy**: Automatic HTTPS
- **Traefik**: Kubernetes-native

### API Gateways (Reverse Proxy + More)
- **Kong**: Plugins, rate limit, auth
- **AWS API Gateway**: Serverless, Lambda
- **Zuul (Netflix)**: Routing, filtering
- **Apigee**: Enterprise API management

### Sidecar (Service Mesh)
- **Envoy**: Istio, Consul Connect
- **Linkerd**: Lightweight
- **Proxies per pod**: Transparent to app

---

## 8. Scaling Strategies

### Forward Proxy
- **Multiple proxies**: Load balance clients across proxies
- **Caching**: Reduce upstream load
- **Hierarchy**: Parent/child caches (Squid)

### Reverse Proxy
- **Horizontal scaling**: Multiple proxy instances behind DNS/LB
- **Connection pooling**: Reuse connections to backends
- **Caching**: Reduce backend load

### API Gateway
- **Auto scaling**: Scale with traffic
- **Caching**: Cache responses
- **Backend scaling**: Gateway scales independently

---

## 9. Failure Scenarios

### Forward Proxy Failure
- **Impact**: Clients cannot reach internet (if mandatory)
- **Mitigation**: Multiple proxies, failover
- **Bypass**: Direct access if proxy optional

### Reverse Proxy Failure
- **Impact**: Service unreachable
- **Mitigation**: Multiple instances, health checks, DNS failover
- **Single point**: Proxy is SPOF—must be HA

### Misconfiguration
- **Wrong routing**: Traffic to wrong backend
- **SSL issues**: Certificate, termination
- **Mitigation**: Staging tests, gradual rollout

### Real Incidents
- **Cloudflare (2019)**: Config error caused global outage
- **Lesson**: Proxy is critical path; test changes

---

## 10. Performance Considerations

### Forward Proxy
- **Caching**: Reduces latency for cache hits
- **Connection reuse**: Keep-alive to origin
- **Tunnel overhead**: CONNECT adds minimal latency

### Reverse Proxy
- **SSL termination**: CPU-intensive—hardware acceleration
- **Connection pooling**: Essential for backend efficiency
- **Buffering**: Can buffer slow clients (protect backend)

### Optimization
- **HTTP/2**: Multiplexing at proxy
- **Compression**: gzip/Brotli at proxy
- **Caching**: Cache static at proxy

---

## 11. Use Cases

| Use Case | Proxy Type | Example |
|----------|------------|---------|
| **Corporate browsing** | Forward | Squid, Blue Coat |
| **VPN** | Forward | Encrypted tunnel |
| **Bypass geo-block** | Forward | Access region-locked content |
| **Load balancing** | Reverse | Nginx, ALB |
| **API gateway** | Reverse | Kong, API Gateway |
| **SSL termination** | Reverse | Nginx, Cloudflare |
| **CDN** | Reverse | CloudFront, Cloudflare |
| **Service mesh** | Sidecar (both) | Envoy, Istio |
| **Rate limiting** | Reverse | Kong, API Gateway |
| **WAF** | Reverse | Cloudflare, AWS WAF |

---

## 12. Comparison Tables

### Proxy Types Summary

| Type | Location | Hides | Primary Use |
|------|----------|-------|-------------|
| **Forward** | Client-side | Client | Access control, anonymity |
| **Reverse** | Server-side | Servers | LB, gateway, security |
| **Sidecar** | Per-service | Both | Service mesh |

### Nginx vs HAProxy vs Envoy

| Aspect | Nginx | HAProxy | Envoy |
|--------|-------|---------|-------|
| **Forward** | Yes | Limited | Yes |
| **Reverse** | Yes | Yes | Yes |
| **L7** | Yes | Yes | Yes |
| **Dynamic** | Limited | Limited | Yes (xDS) |
| **Service mesh** | No | No | Yes |

---

## 13. Code or Pseudocode

### Forward Proxy (Simplified)

```python
# Client sends: GET http://example.com/page HTTP/1.1
# Via: proxy.company.com

def forward_proxy_handle(client_socket):
    request = parse_request(client_socket)
    # Client wants example.com
    origin_socket = connect(request.host, request.port)
    origin_socket.send(request.raw)
    
    response = origin_socket.recv()
    # Optionally cache
    if request.method == "GET" and is_cacheable(response):
        cache.set(request.url, response)
    
    client_socket.send(response)
```

### Reverse Proxy (Nginx Config)

```nginx
upstream backend {
    least_conn;
    server 10.0.1.1:8080;
    server 10.0.1.2:8080;
    server 10.0.1.3:8080;
}

server {
    listen 443 ssl;
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### API Gateway Routing (Pseudocode)

```python
def api_gateway_route(request):
    # Auth
    if not authenticate(request):
        return 401
    
    # Rate limit
    if rate_limiter.is_exceeded(request.client_id):
        return 429
    
    # Route by path
    if request.path.startswith("/users"):
        backend = "user-service"
    elif request.path.startswith("/orders"):
        backend = "order-service"
    else:
        return 404
    
    # Forward
    return proxy_to(backend, request)
```

### Sidecar Proxy (Envoy Config Concept)

```yaml
# Envoy sidecar: intercepts outbound traffic
clusters:
  - name: order-service
    lb_policy: ROUND_ROBIN
    hosts:
      - socket_address: { address: order-service, port_value: 8080 }
```

---

## 14. Interview Discussion

### Key Points to Cover
1. **Forward**: Client-side, hides client, corporate/VPN
2. **Reverse**: Server-side, hides servers, load balancing, gateway
3. **Difference**: Who configures, who is hidden
4. **Sidecar**: Per-pod proxy, service mesh
5. **API Gateway**: Reverse proxy + auth, rate limit, routing

### Common Follow-ups
- **"When use forward vs reverse?"** → Forward: client control; Reverse: server scaling
- **"What is sidecar proxy?"** → Proxy next to each service, transparent
- **"How does API gateway work?"** → Reverse proxy + middleware (auth, rate limit)
- **"SSL termination at proxy?"** → Proxy decrypts, forwards to backend (plaintext or re-encrypt)
- **"Single point of failure?"** → Multiple proxy instances, HA

### Red Flags to Avoid
- Confusing forward and reverse (who is hidden)
- Saying "proxy" without specifying type
- Ignoring SSL termination
- Not considering proxy as SPOF

### Advanced Topics

#### Transparent Proxy
- **No client config**: Intercept traffic at network level (e.g., WCCP, inline)
- **Use case**: Corporate networks where users can't bypass
- **Implementation**: Router redirects traffic to proxy

#### Proxy Chaining
- **Forward chain**: Client -> Proxy1 -> Proxy2 -> Origin
- **Use case**: Tor (onion routing), multi-hop anonymity
- **Reverse chain**: Client -> CDN -> Regional proxy -> Origin (cache hierarchy)

#### SSL Inspection (Forward Proxy)
- **Corporate security**: Proxy decrypts HTTPS, inspects, re-encrypts
- **Requires**: CA cert installed on client (man-in-the-middle)
- **Privacy concern**: Employer can see all traffic
- **Implementation**: Squid with SSL bump, Blue Coat

#### Connection Pooling at Reverse Proxy
- **Problem**: Each request = new connection to backend (overhead)
- **Solution**: Pool of persistent connections to backends
- **Benefit**: Reuse connections, reduce latency, reduce backend load
- **Config**: Nginx `keepalive`, HAProxy `server ... maxconn`

#### Rate Limiting at Reverse Proxy
- **Token bucket**: Allow burst, then throttle
- **Sliding window**: Per time window (e.g., 100 req/min)
- **Per key**: IP, user ID, API key
- **Response**: 429 Too Many Requests, Retry-After header

#### Circuit Breaker Pattern (API Gateway)
- **Closed**: Normal operation
- **Open**: Fail fast, don't call backend (after threshold failures)
- **Half-open**: Try one request; if success, close; if fail, open
- **Prevents**: Cascading failure when backend is down

#### X-Forwarded-* Headers
- **X-Forwarded-For**: Original client IP (chain of proxies)
- **X-Forwarded-Proto**: Original scheme (http/https)
- **X-Forwarded-Host**: Original Host header
- **Importance**: Backend needs to know real client for logging, rate limiting, geo

#### System Design: When to Use Each
- **Forward proxy**: Corporate network, VPN, content filtering, bypass restrictions
- **Reverse proxy**: Any web service—load balancing, SSL, caching, single entry point
- **API Gateway**: Microservices—centralized auth, rate limit, routing
- **Sidecar**: Service mesh—transparent LB, retry, observability per service

#### Load Balancer vs Reverse Proxy
- **Overlap**: Many reverse proxies ARE load balancers (Nginx, HAProxy)
- **LB focus**: Distribution algorithm, health checks
- **Reverse proxy focus**: SSL termination, caching, routing
- **In practice**: Terms used interchangeably; ALB is "application load balancer" (reverse proxy)

#### CONNECT Method (HTTPS Tunneling)
- **Forward proxy + HTTPS**: Client sends `CONNECT example.com:443` to proxy
- **Proxy**: Establishes TCP tunnel; does NOT decrypt
- **End-to-end encryption**: Client and origin communicate directly through tunnel
- **Proxy sees**: Only destination host:port, not content
- **SSL inspection**: Corporate proxies can decrypt (with installed CA cert) for security—privacy tradeoff
