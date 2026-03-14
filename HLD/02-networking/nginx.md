# How Nginx Works

## 1. Concept Overview

### Definition
**Nginx** is a high-performance, event-driven web server, reverse proxy, load balancer, and HTTP cache. It uses an asynchronous, non-blocking architecture that enables it to handle millions of concurrent connections on a single server with minimal resource consumption. Nginx excels at serving static content, proxying to backend servers, and terminating SSL.

### Purpose
- **Reverse proxy**: Forward requests to backend servers (app servers, APIs)
- **Load balancing**: Distribute traffic across multiple backends
- **SSL termination**: Handle HTTPS; decrypt and forward to backends
- **Static file serving**: Serve HTML, CSS, JS, images efficiently
- **C10K problem**: Handle 10,000+ concurrent connections on one machine

### Problems It Solves
- **C10K problem**: Traditional thread-per-connection model doesn't scale
- **Single point of entry**: Centralize routing, SSL, load balancing
- **Static content**: Offload from application servers
- **High concurrency**: Millions of connections with low memory footprint

---

## 2. Real-World Motivation

### Netflix
- **Open Connect**: Nginx-based edge appliances for video delivery
- **Load balancing**: Route to origin, cache layers
- **Scale**: 15% of global internet traffic at peak

### Cloudflare
- **Edge proxy**: Nginx (and custom) at 300+ data centers
- **DDoS protection**: Absorb attacks at edge
- **SSL termination**: Millions of TLS handshakes per second

### WordPress.com
- **Reverse proxy**: Nginx in front of PHP-FPM
- **Static serving**: CDN-like static file delivery
- **Load balancing**: Across thousands of backend servers

### Airbnb, Dropbox, Pinterest
- **API gateway**: Nginx as reverse proxy to microservices
- **SSL termination**: Centralized certificate management
- **Rate limiting**: Limit requests per IP

---

## 3. Architecture Diagrams

### Master-Worker Model

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     NGINX MASTER-WORKER MODEL                            │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   MASTER PROCESS (root, privileged)                                      │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │  - Reads configuration                                            │   │
│   │  - Binds to ports (80, 443)                                       │   │
│   │  - Manages worker processes                                       │   │
│   │  - Graceful reload (re-exec, workers reload config)               │   │
│   │  - Does NOT handle connections                                    │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│                                │                                         │
│                    fork()      │      fork()                              │
│                    ┌───────────┴───────────┐                             │
│                    │                       │                             │
│                    ▼                       ▼                             │
│   WORKER 1                    WORKER 2                    WORKER N       │
│   ┌─────────────────┐        ┌─────────────────┐        ┌─────────────┐  │
│   │  Event loop      │        │  Event loop      │        │  Event loop │  │
│   │  - epoll/kqueue  │        │  - epoll/kqueue  │        │  - epoll    │  │
│   │  - Non-blocking  │        │  - Non-blocking  │        │  - Non-block│  │
│   │  - Handles 1000s │        │  - Handles 1000s │        │  - Handles  │  │
│   │    of conns      │        │    of conns      │        │    1000s    │  │
│   └─────────────────┘        └─────────────────┘        └─────────────┘  │
│                                                                          │
│   Workers: 1 per CPU core (typically); run as unprivileged user          │
│   All workers accept() from same socket (SO_REUSEPORT or accept mutex)   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Event Loop (epoll)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     NGINX EVENT LOOP (epoll)                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   TRADITIONAL (Apache, thread-per-connection):                           │
│   ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐     ┌─────┐                            │
│   │Conn1│ │Conn2│ │Conn3│ │Conn4│ ... │Conn │  10K conns = 10K threads   │
│   │thread│ │thread│ │thread│ │thread│     │10K │  = 10GB+ RAM (1MB/thread) │
│   └─────┘ └─────┘ └─────┘ └─────┘     └─────┘  Context switch overhead   │
│                                                                          │
│   NGINX (event-driven, single thread):                                   │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │  Event Loop (one per worker)                                      │   │
│   │                                                                   │   │
│   │  while (true) {                                                   │   │
│   │    events = epoll_wait(...)  // Block until I/O ready             │   │
│   │    for (event in events) {                                         │   │
│   │      if (event.readable)  → read from socket, process              │   │
│   │      if (event.writable)  → write to socket                        │   │
│   │      if (event.new_conn)  → accept(), add to epoll                 │   │
│   │    }                                                               │   │
│   │  }                                                                 │   │
│   │                                                                   │   │
│   │  10K conns = 1 thread, ~10-50 MB RAM (connection structs only)   │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│   Key: Non-blocking I/O; never block on read/write; epoll multiplexes   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### C10K Problem Solved

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     C10K PROBLEM: NGINX SOLUTION                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   Problem (2000s): How to handle 10,000 concurrent connections?          │
│                                                                          │
│   Thread-per-connection (Apache):                                        │
│   - 10K threads × 1 MB stack = 10 GB RAM                                 │
│   - Context switching kills CPU                                          │
│   - Doesn't scale                                                        │
│                                                                          │
│   Nginx approach:                                                        │
│   - Event-driven: epoll (Linux), kqueue (BSD), select/poll               │
│   - One thread handles 1000s of connections                               │
│   - Non-blocking: read/write returns immediately; callback when ready   │
│   - Memory: ~1-2 KB per connection (vs 1 MB per thread)                   │
│                                                                          │
│   Result: 1M+ concurrent connections on single server (with tuning)     │
│   - worker_connections 65535                                               │
│   - multi_accept on                                                       │
│   - epoll                                                                 │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Nginx as Reverse Proxy

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     NGINX REVERSE PROXY                                   │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   CLIENT                    NGINX                         BACKENDS       │
│   ┌─────┐                   ┌─────────────────┐          ┌───────────┐   │
│   │     │  HTTPS (443)       │  - SSL terminate│          │  App 1    │   │
│   │     │───────────────────►│  - Load balance │─────────►│  :8080    │   │
│   │     │                    │  - Reverse proxy│          ├───────────┤   │
│   │     │                    │  - Static files │─────────►│  App 2    │   │
│   │     │                    │  - Cache        │          │  :8080    │   │
│   │     │                    └─────────────────┘          ├───────────┤   │
│   │     │                                                 │  App 3    │   │
│   │     │                                                 │  :8080    │   │
│   └─────┘                                                 └───────────┘   │
│                                                                          │
│   Client sees: https://example.com (Nginx)                                │
│   Backend sees: HTTP from Nginx (or HTTPS if configured)                  │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Core Mechanics

### Event-Driven Architecture
- **epoll** (Linux): Edge-triggered or level-triggered; scalable to 100K+ fds
- **kqueue** (BSD, macOS): Similar to epoll
- **select/poll**: Fallback; less efficient
- **Non-blocking sockets**: read()/write() return EAGAIN if no data; don't block

### Request Processing
1. **Accept**: Worker accepts connection; adds fd to epoll
2. **Read**: Client sends request; epoll signals readable; Nginx reads
3. **Parse**: Parse HTTP request (method, URI, headers)
4. **Process**: Match location block; proxy to backend or serve static
5. **Write**: Response ready; epoll signals writable; Nginx writes
6. **Keep-alive**: Connection stays open for next request (or close)

### Worker Model
- **worker_processes**: Typically = CPU cores
- **worker_connections**: Max connections per worker (e.g., 65535)
- **Total capacity**: worker_processes × worker_connections
- **Accept mutex**: Workers compete to accept; or SO_REUSEPORT (kernel distributes)

### Load Balancing
- **Upstream block**: Define backend servers
- **Methods**: Round-robin, least_conn, ip_hash, hash
- **Health checks**: Passive (max_fails, fail_timeout) or active (nginx-plus)

---

## 5. Numbers

| Metric | Value |
|--------|-------|
| Connections per worker | 10K-65K (configurable) |
| Memory per connection | ~1-2 KB (idle) |
| 1M connections | ~2-4 GB RAM (Nginx) vs 1 TB (thread-per-conn) |
| Latency overhead | <1 ms (proxy) |
| Throughput | 100K+ req/sec (static); 50K+ (proxy) |
| C10K | Solved; C1M achievable |

### Apache vs Nginx (Thread-per-conn vs Event-driven)

| Metric | Apache (prefork) | Nginx |
|--------|------------------|-------|
| 10K connections | 10K processes/threads | 1-8 workers |
| Memory (10K conn) | 5-10 GB | 50-100 MB |
| Context switches | High | Low |
| Static file serving | Good | Excellent |

---

## 6. Tradeoffs

### Nginx vs Apache

| Aspect | Nginx | Apache |
|--------|-------|--------|
| **Model** | Event-driven | Thread/process per connection |
| **Concurrency** | High (C10K+) | Lower (thread limit) |
| **Static files** | Excellent | Good |
| **Dynamic (PHP)** | Via FastCGI | mod_php (in-process) |
| **Config** | Declarative | .htaccess, mod_rewrite |
| **Modules** | Fewer | Many (mod_*) |

### Nginx vs HAProxy

| Aspect | Nginx | HAProxy |
|--------|-------|---------|
| **Primary use** | Web server, proxy | Load balancer |
| **Layer 7** | Yes | Yes |
| **Layer 4** | Limited | Excellent (TCP) |
| **SSL** | Yes | Yes |
| **Static** | Yes | No |
| **Config** | nginx.conf | haproxy.cfg |

---

## 7. Variants / Implementations

### Nginx Variants
- **Nginx OSS**: Open source
- **Nginx Plus**: Commercial; active health checks, API, support
- **OpenResty**: Nginx + Lua; dynamic routing, custom logic
- **Kong**: API gateway built on OpenResty

### Use Cases
- **Web server**: Serve static files
- **Reverse proxy**: Forward to app servers
- **Load balancer**: Distribute across backends
- **SSL termination**: HTTPS at edge
- **Caching**: Proxy cache for backends
- **API gateway**: Rate limiting, auth (with modules)

---

## 8. Scaling Strategies

- **Vertical**: More worker_connections; tune OS (file descriptors)
- **Horizontal**: Multiple Nginx instances behind DNS/LB
- **Caching**: Proxy cache to reduce backend load
- **Connection pooling**: Keep-alive to backends
- **SO_REUSEPORT**: Kernel load balance across workers

---

## 9. Failure Scenarios

| Scenario | Mitigation |
|----------|------------|
| **Worker crash** | Master restarts; few connections lost |
| **Backend down** | Upstream max_fails; next backend |
| **SSL cert expiry** | Monitor; automate renewal (Let's Encrypt) |
| **DDoS** | Rate limiting; connection limits; upstream protection |
| **Config error** | nginx -t before reload; graceful reload |

---

## 10. Performance Considerations

- **worker_processes**: = CPU cores
- **worker_connections**: 65535 typical
- **multi_accept on**: Accept multiple connections per wake
- **sendfile on**: Kernel sends file directly; no user-space copy
- **tcp_nopush, tcp_nodelay**: Tune TCP
- **Keep-alive**: Reuse connections to client and backend
- **Gzip**: Compress responses
- **Open file cache**: Cache file descriptors for static files

---

## 11. Use Cases

| Use Case | Nginx Role |
|----------|------------|
| **Static file server** | Serve HTML, CSS, JS, images |
| **Reverse proxy** | Forward to app servers |
| **Load balancer** | Distribute across backends |
| **SSL termination** | HTTPS at edge |
| **API gateway** | Rate limit, auth, routing |
| **Caching** | Proxy cache |
| **WebSocket proxy** | Upgrade support |

---

## 12. Comparison Tables

### Web Server / Proxy Comparison

| Feature | Nginx | Apache | HAProxy |
|---------|-------|--------|---------|
| **Event-driven** | Yes | No (default) | Yes |
| **C10K** | Yes | No | Yes |
| **Reverse proxy** | Yes | Yes | Yes |
| **Load balance** | Yes | Limited | Yes |
| **Static files** | Yes | Yes | No |
| **SSL** | Yes | Yes | Yes |
| **Config** | nginx.conf | httpd.conf | haproxy.cfg |

### When to Use Nginx

| Use Nginx | Use Alternative |
|-----------|-----------------|
| High concurrency | Apache (simpler config) |
| Static + proxy | Nginx |
| Pure L4 LB | HAProxy |
| Dynamic routing | OpenResty, Kong |

---

## 13. Code or Pseudocode

### Nginx Configuration Example

```nginx
worker_processes auto;
worker_connections 65535;
events {
    use epoll;
    multi_accept on;
}

http {
    upstream backend {
        least_conn;
        server 10.0.1.1:8080 weight=1;
        server 10.0.1.2:8080 weight=1;
        server 10.0.1.3:8080 backup;
    }

    server {
        listen 443 ssl;
        server_name example.com;
        ssl_certificate /etc/ssl/cert.pem;
        ssl_certificate_key /etc/ssl/key.pem;

        location / {
            proxy_pass http://backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }

        location /static/ {
            alias /var/www/static/;
            expires 1d;
        }
    }
}
```

### Event Loop Pseudocode

```python
# Simplified Nginx-style event loop
def worker_event_loop(listen_socket):
    epoll = select.epoll()
    epoll.register(listen_socket, select.EPOLLIN)
    connections = {}

    while True:
        events = epoll.poll()
        for fd, event in events:
            if fd == listen_socket.fileno():
                conn, addr = listen_socket.accept()
                conn.setblocking(False)
                epoll.register(conn, select.EPOLLIN)
                connections[conn.fileno()] = Connection(conn)
            elif event & select.EPOLLIN:
                conn = connections[fd]
                data = conn.socket.recv(1024)
                if data:
                    conn.process_request(data)
                    epoll.modify(fd, select.EPOLLOUT)
            elif event & select.EPOLLOUT:
                conn = connections[fd]
                conn.send_response()
                epoll.unregister(fd)
                conn.close()
```

---

## 14. Interview Discussion

### Key Points to Articulate
1. **Event-driven**: No thread per connection; epoll/kqueue multiplex
2. **Master-worker**: Master manages; workers handle connections
3. **C10K**: Solved; 1M connections possible with low memory
4. **Use cases**: Reverse proxy, load balancer, SSL termination, static
5. **Compare**: Apache = thread-per-conn; HAProxy = pure LB

### Common Interview Questions
- **"How does Nginx handle 1M connections?"** → Event-driven; epoll; non-blocking I/O; one worker handles 10K+ conns; low memory per conn
- **"What is the C10K problem?"** → Handling 10K concurrent connections; thread-per-conn doesn't scale; Nginx solves with event loop
- **"Nginx vs Apache?"** → Nginx: event-driven, high concurrency. Apache: thread-per-conn, more modules
- **"How does Nginx work as reverse proxy?"** → Accept connection; read request; forward to backend; return response; SSL terminate at Nginx
- **"Master-worker model?"** → Master: config, bind ports, manage workers. Workers: event loop, handle connections

### Red Flags to Avoid
- Saying Nginx uses one thread per connection
- Confusing with Apache model
- Not mentioning epoll/event-driven

### Ideal Answer Structure
1. Define Nginx (event-driven web server, reverse proxy)
2. Explain master-worker and event loop (epoll)
3. C10K: problem and how Nginx solves it
4. Use cases: reverse proxy, load balance, SSL, static
5. Compare with Apache, HAProxy
