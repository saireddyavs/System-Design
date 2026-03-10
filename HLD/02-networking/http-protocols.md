# HTTP/1.1, HTTP/2, HTTP/3, and TLS

## 1. Concept Overview

### Definition
**HTTP (Hypertext Transfer Protocol)** is the application-layer protocol for transmitting hypermedia documents on the web. **TLS (Transport Layer Security)** provides encryption, authentication, and integrity for HTTP (resulting in HTTPS). The protocol has evolved through HTTP/1.1 (1997), HTTP/2 (2015), and HTTP/3 (2022) to address latency, efficiency, and security.

### Purpose
- **HTTP/1.1**: Text-based, request-response, persistent connections
- **HTTP/2**: Binary framing, multiplexing, header compression—reduce latency
- **HTTP/3**: QUIC-based, UDP transport—eliminate TCP head-of-line blocking, 0-RTT
- **TLS**: Encrypt data in transit, authenticate servers, prevent tampering

### Problems They Solve
1. **Latency**: HTTP/1.1 head-of-line blocking → HTTP/2 multiplexing → HTTP/3 stream independence
2. **Connection overhead**: HTTP/1.1 one request per connection (without pipelining) → HTTP/2 single connection, many streams
3. **Header overhead**: Repeated headers (cookies, user-agent) → HPACK compression
4. **Security**: Plaintext → TLS encryption
5. **Handshake latency**: TLS 1.2 (2 RTT) → TLS 1.3 (1 RTT), QUIC 0-RTT resumption

---

## 2. Real-World Motivation

### Google
- **HTTP/2**: Adopted early, 30%+ latency reduction for multi-resource pages
- **QUIC**: Originated at Google, deployed for YouTube, Gmail, Search
- **TLS 1.3**: Default for all Google properties
- **0-RTT**: Critical for mobile, repeat visits

### Netflix
- **HTTP/2**: For API, metadata—multiplexing reduces connection count
- **Video**: Chunked over HTTP/2 streams, adaptive bitrate
- **TLS**: Full HTTPS for DRM, privacy

### Uber
- **API**: HTTP/2 for mobile app—single connection, many concurrent requests
- **Real-time**: WebSocket over HTTP/1.1 upgrade, or HTTP/2 for streaming
- **TLS**: Certificate pinning for app security

### Amazon
- **CloudFront**: Supports HTTP/1.1, HTTP/2, HTTP/3 (QUIC)
- **ALB**: HTTP/2 to backend, connection pooling
- **TLS**: SNI for multiple domains, ACM for cert management

### Twitter
- **Timeline API**: HTTP/2 multiplexing for parallel requests
- **Media**: HTTP/2 for image/video delivery
- **TLS**: HSTS, certificate transparency

---

## 3. Architecture Diagrams

### HTTP/1.1 vs HTTP/2 vs HTTP/3 Stack

```
HTTP/1.1:
+----------+
|  HTTP    |  Text, one request/response per connection (or pipelining)
+----------+
|  TLS     |  Optional
+----------+
|  TCP     |  One connection per resource (or keep-alive)
+----------+
|  IP      |
+----------+

HTTP/2:
+----------+
|  HTTP/2  |  Binary frames, multiplexed streams, HPACK
+----------+
|  TLS     |  Typically ALPN negotiation
+----------+
|  TCP     |  Single connection, multiple streams
+----------+
|  IP      |
+----------+

HTTP/3:
+----------+
|  HTTP/3  |  Same semantics as HTTP/2, over QUIC
+----------+
|  QUIC    |  TLS 1.3 built-in, streams, 0-RTT
+----------+
|  UDP     |
+----------+
|  IP      |
+----------+
```

### HTTP/1.1 Head-of-Line Blocking

```
Connection 1: [Req1] -----> [Resp1] -----> [Req2] -----> [Resp2]
                              |
                    Must wait for Resp1 before Req2
                    (with pipelining: send Req2 early, but still wait for order)

Multiple connections (browser limit ~6 per domain):
Conn1: [Req1] --> [Resp1]
Conn2: [Req2] --> [Resp2]   } Parallel, but connection overhead
Conn3: [Req3] --> [Resp3]
```

### HTTP/2 Multiplexing

```
Single TCP connection:
+------------------------------------------------------------------+
| Stream 1: [HEADERS][DATA][DATA]                                    |
| Stream 2: [HEADERS][DATA]                                          |
| Stream 3: [HEADERS][DATA][DATA][DATA]                              |
| Stream 4: [HEADERS]                                                |
+------------------------------------------------------------------+
Frames interleaved; streams independent
BUT: One lost TCP packet blocks ALL streams (TCP HOL blocking)
```

### HTTP/3 (QUIC) Stream Independence

```
Single QUIC connection (UDP):
+------------------------------------------------------------------+
| Stream 1: [frame][frame][LOST][frame]  <- Stream 1 blocks          |
| Stream 2: [frame][frame][frame]        <- Delivered                |
| Stream 3: [frame][frame][frame]        <- Delivered                |
+------------------------------------------------------------------+
Loss in Stream 1 does NOT block Streams 2, 3 (QUIC streams independent)
```

### TLS 1.2 Handshake (2 RTT)

```
CLIENT                                              SERVER
   |                                                    |
   |  1. ClientHello (supported ciphers, extensions)   |
   |--------------------------------------------------->|
   |                                                    |
   |  2. ServerHello (chosen cipher), Certificate,      |
   |     ServerKeyExchange, CertificateRequest,         |
   |     ServerHelloDone                                |
   |<---------------------------------------------------|
   |                                                    |
   |  3. Certificate, ClientKeyExchange,                |
   |     CertificateVerify, [ChangeCipherSpec],         |
   |     Finished                                       |
   |--------------------------------------------------->|
   |                                                    |
   |  4. [ChangeCipherSpec], Finished                   |
   |<---------------------------------------------------|
   |                                                    |
   |  Encrypted Application Data                        |
   |<==================================================>|
   |                                                    |
   |  Total: 2 RTT before first application data       |
```

### TLS 1.3 Handshake (1 RTT)

```
CLIENT                                              SERVER
   |                                                    |
   |  1. ClientHello (key_share, supported groups)      |
   |--------------------------------------------------->|
   |                                                    |
   |  2. ServerHello, EncryptedExtensions, Certificate,  |
   |     CertificateVerify, Finished                    |
   |<---------------------------------------------------|
   |                                                    |
   |  3. Finished                                       |
   |--------------------------------------------------->|
   |                                                    |
   |  Encrypted Application Data                        |
   |<==================================================>|
   |                                                    |
   |  Total: 1 RTT before first application data        |
```

### TLS 1.3 0-RTT (Resumption)

```
CLIENT (has session ticket)                            SERVER
   |                                                    |
   |  1. ClientHello (early_data, key_share),            |
   |     Application Data (0-RTT)                        |
   |--------------------------------------------------->|
   |                                                    |
   |  2. ServerHello, EncryptedExtensions, Certificate,  |
   |     CertificateVerify, Finished,                   |
   |     Application Data (response)                     |
   |<---------------------------------------------------|
   |                                                    |
   |  Total: 0 RTT for first byte (if resumption)       |
   |  Note: 0-RTT data is not forward-secret; replay risk|
```

---

## 4. Core Mechanics

### HTTP/1.1
- **Text-based**: Human-readable request/response
- **Request format**: Method, URI, Version, Headers, Body
- **Persistent connections**: Keep-Alive header, reuse TCP connection
- **Pipelining**: Send multiple requests without waiting—rarely used (head-of-line blocking, poor server support)
- **Chunked transfer**: Transfer-Encoding: chunked for streaming

### HTTP/2
- **Binary framing**: Frames (HEADERS, DATA, SETTINGS, etc.) with stream ID
- **Streams**: Logical channels within one connection; multiplexed
- **HPACK**: Header compression—static table (common headers) + dynamic table (request-specific)
- **Server push**: Server can push resources before client requests (declining use)
- **Flow control**: Per-stream flow control (similar to TCP)
- **Priority**: Stream dependency tree for prioritization

### HTTP/3
- **Same semantics as HTTP/2**: Frames, streams, HPACK
- **QUIC transport**: UDP-based, built-in TLS 1.3
- **Stream independence**: Loss in one stream doesn't block others
- **Connection migration**: Change IP (e.g., WiFi to cellular) without breaking connection
- **0-RTT**: Resumption sends data in first packet

### TLS 1.2 vs 1.3
- **1.2**: Supports legacy ciphers (RC4, 3DES, etc.), 2 RTT handshake
- **1.3**: Removed weak ciphers, 1 RTT handshake, 0-RTT resumption
- **Key exchange**: 1.3 uses (EC)DHE only; 1.2 allowed static RSA
- **Session resumption**: 1.2: Session ID or ticket; 1.3: Ticket only

---

## 5. Numbers

### Latency (RTT = 50ms)

| Scenario | HTTP/1.1 | HTTP/2 | HTTP/3 |
|----------|----------|--------|--------|
| **TLS handshake** | 2 RTT (100ms) | 1 RTT (50ms) | 1 RTT (50ms) |
| **First request** | 1 RTT (50ms) | 1 RTT (50ms) | 1 RTT (50ms) |
| **Total (new connection)** | 3 RTT (150ms) | 2 RTT (100ms) | 2 RTT (100ms) |
| **Resumption (0-RTT)** | N/A | N/A | 0 RTT (0ms) |
| **100 resources** | 100 connections or pipelining | 1 connection, multiplexed | 1 connection |

### Header Overhead
| Protocol | Uncompressed | Compressed (HPACK) |
|----------|--------------|---------------------|
| **HTTP/1.1** | ~500-800 bytes/request | N/A |
| **HTTP/2** | Same | ~50-100 bytes (typical) |
| **Savings** | - | 80-90% |

### Throughput
- **HTTP/1.1**: Limited by connection count (6/domain), head-of-line blocking
- **HTTP/2**: Single connection can saturate link; TCP HOL still applies
- **HTTP/3**: Better under loss (stream independence); similar in ideal conditions

### Adoption (2024)
- **HTTP/2**: ~50-60% of websites
- **HTTP/3**: ~25-30% (growing)
- **TLS 1.3**: ~90%+ of HTTPS connections

---

## 6. Tradeoffs

### HTTP/1.1 vs HTTP/2 vs HTTP/3

| Aspect | HTTP/1.1 | HTTP/2 | HTTP/3 |
|--------|----------|--------|--------|
| **Transport** | TCP | TCP | QUIC (UDP) |
| **Multiplexing** | No (or pipelining) | Yes | Yes |
| **Header compression** | No | HPACK | HPACK |
| **HOL blocking** | Application | TCP (all streams) | Per-stream only |
| **Server push** | No | Yes (declining) | Deprecated |
| **0-RTT** | No | No | Yes |
| **Deployment** | Universal | Wide | Growing |

### TLS 1.2 vs 1.3

| Aspect | TLS 1.2 | TLS 1.3 |
|--------|---------|---------|
| **Handshake RTT** | 2 | 1 |
| **0-RTT** | No | Yes (resumption) |
| **Ciphers** | Legacy allowed | Modern only |
| **Forward secrecy** | With DHE | Always (DHE required) |
| **Compatibility** | Older clients | Newer only |

### Connection Pooling
- **HTTP/1.1**: Pool = multiple TCP connections (e.g., 10-100)
- **HTTP/2**: Pool = fewer connections (1-2 per origin), many streams
- **HTTP/3**: Similar to HTTP/2, plus connection migration

---

## 7. Variants / Implementations

### HTTP/2 Implementations
- **Nginx**: http2 on; ALPN
- **Apache**: mod_http2
- **Node.js**: http2 module
- **Go**: net/http with HTTP/2
- **Envoy**: HTTP/2 by default

### HTTP/3 Implementations
- **Cloudflare**: Full HTTP/3 support
- **Chrome, Firefox, Edge**: Enabled by default
- **Nginx**: Experimental quic module
- **Caddy**: Built-in HTTP/3
- **AWS CloudFront, ALB**: HTTP/3 support

### TLS Libraries
- **OpenSSL**: 1.1.1+ for TLS 1.3
- **BoringSSL**: Google's fork, QUIC support
- **rustls**: Rust, memory-safe
- **s2n-tls**: AWS, simplicity focus

---

## 8. Scaling Strategies

### Connection Management
1. **Keep-alive**: Reuse connections (HTTP/1.1)
2. **Connection pooling**: Client-side pool per origin
3. **HTTP/2**: Single connection, many streams—reduce connection count
4. **Domain sharding**: Historical (HTTP/1.1)—multiple domains to bypass 6-connection limit; less needed with HTTP/2

### TLS Optimization
1. **Session resumption**: Ticket or session ID
2. **OCSP stapling**: Server provides OCSP response, avoid client lookup
3. **TLS 1.3**: Faster handshake
4. **0-RTT**: For repeat connections (idempotent requests only—replay risk)

### CDN and Edge
- **Edge TLS termination**: TLS at edge, HTTP to origin
- **Certificate management**: ACM, Let's Encrypt, custom
- **SNI**: Multiple domains on same IP

---

## 9. Failure Scenarios

### TLS Failures
- **Certificate expired**: Monitor, auto-renew
- **Certificate chain incomplete**: Include intermediate CAs
- **SNI mismatch**: Ensure certificate covers domain
- **Cipher mismatch**: Client/server don't share cipher—upgrade TLS

### HTTP/2 Issues
- **Connection coalescing**: Same IP for different hostnames—verify certificate SAN
- **Server push**: Poor adoption, can waste bandwidth—many disable
- **TCP HOL**: Under loss, all streams block—motivation for HTTP/3

### HTTP/3 (QUIC) Failures
- **UDP blocked**: ~2-5% of networks block/throttle UDP—fallback to HTTP/2
- **0-RTT replay**: Non-idempotent requests vulnerable—use only for GET
- **Connection migration**: Some NATs change port—connection may drop

### Real Incidents
- **Cloudflare 0-RTT replay (2019)**: Research showed replay risk—recommendations updated
- **QUIC fallback**: Browsers fallback to HTTP/2 when QUIC fails
- **TLS 1.0/1.1 deprecation**: PCI-DSS, browsers removing support

---

## 10. Performance Considerations

### Metrics to Monitor
- **TLS handshake time**: P50, P95
- **Time to first byte (TTFB)**: Affected by handshake, server processing
- **Connection establishment**: Success rate, latency
- **HTTP/3 adoption**: % of requests over QUIC

### Optimization Checklist
- [ ] Enable HTTP/2 (and HTTP/3 if supported)
- [ ] Use TLS 1.3
- [ ] Enable OCSP stapling
- [ ] Session resumption (tickets)
- [ ] Connection pooling
- [ ] Avoid domain sharding (HTTP/2)
- [ ] Consider 0-RTT for idempotent APIs

### HPACK Tuning
- **Dynamic table size**: Larger = better compression, more memory
- **Static table**: 61 common header entries—no tuning

---

## 11. Use Cases

| Use Case | Protocol | Rationale |
|----------|----------|-----------|
| **Legacy systems** | HTTP/1.1 | Compatibility |
| **Modern web** | HTTP/2 | Multiplexing, compression |
| **Mobile, high latency** | HTTP/3 | 0-RTT, connection migration |
| **API (REST, gRPC)** | HTTP/2 | Multiplexing, streaming |
| **Video streaming** | HTTP/2 | Chunked transfer, multiplexing |
| **Real-time** | WebSocket (HTTP/1.1 upgrade) or HTTP/2 | Bidirectional |
| **CDN** | HTTP/2, HTTP/3 | Edge optimization |

---

## 12. Comparison Tables

### Protocol Evolution Summary

| Feature | HTTP/1.0 | HTTP/1.1 | HTTP/2 | HTTP/3 |
|---------|----------|----------|--------|--------|
| **Connections** | 1 per request | Keep-alive | Multiplexed | Multiplexed |
| **Format** | Text | Text | Binary | Binary |
| **Compression** | Optional (body) | Optional | HPACK | HPACK |
| **Streaming** | No | Chunked | Yes | Yes |
| **Server push** | No | No | Yes | Deprecated |
| **Transport** | TCP | TCP | TCP | QUIC/UDP |

### Cipher Suites (TLS 1.3)

| Cipher | Key Exchange | Authentication |
|--------|--------------|----------------|
| **TLS_AES_256_GCM_SHA384** | (EC)DHE | RSA or ECDSA |
| **TLS_CHACHA20_POLY1305_SHA256** | (EC)DHE | RSA or ECDSA |
| **TLS_AES_128_GCM_SHA256** | (EC)DHE | RSA or ECDSA |

---

## 13. Code or Pseudocode

### HTTP/1.1 Request (Raw)

```http
GET /api/users HTTP/1.1
Host: example.com
User-Agent: Mozilla/5.0
Accept: application/json
Connection: keep-alive

```

### HTTP/2 Frame (Conceptual)

```
+-----------------------------------------------+
|                 Length (24)                   |
+---------------+---------------+---------------+
|   Type (8)    |   Flags (8)   |
+-+-------------+---------------+-------------------------------+
|R|                 Stream Identifier (31)                      |
+=+=============================================================+
|                   Frame Payload (0...)                       ...
+---------------------------------------------------------------+
```

### TLS 1.3 Handshake (Pseudocode)

```python
# Client
client_hello = ClientHello(
    random=random_bytes(32),
    cipher_suites=[TLS_AES_256_GCM_SHA384, ...],
    extensions=[key_share, supported_versions, ...]
)
send(client_hello)

server_hello = receive()
# Verify server cert, compute keys
send(Finished)

# Encrypted application data
send(encrypt(application_data))
```

### HTTP/2 Stream Priority

```python
# Client sends request with priority
# Stream 5 depends on Stream 3, weight 256
headers_frame = HeadersFrame(
    stream_id=5,
    priority=PriorityFrame(depends_on=3, weight=256),
    headers=[(":method", "GET"), (":path", "/style.css")]
)
```

---

## 14. Interview Discussion

### Key Points to Cover
1. **HTTP/1.1**: Text, keep-alive, one request per connection (effectively), head-of-line blocking
2. **HTTP/2**: Binary, multiplexing, HPACK, single connection—but TCP HOL blocks all streams
3. **HTTP/3**: QUIC, UDP, stream independence, 0-RTT
4. **TLS 1.2 vs 1.3**: 2 RTT vs 1 RTT, 0-RTT resumption
5. **When to use each**: Compatibility vs performance

### Common Follow-ups
- **"Why does HTTP/2 use a single connection?"** → Reduce connection overhead, multiplexing
- **"What is head-of-line blocking?"** → One slow/lost item blocks subsequent
- **"How does HTTP/3 fix it?"** → QUIC streams independent; loss in one doesn't block others
- **"What is 0-RTT?"** → Resumption sends data in first packet; replay risk for non-idempotent
- **"Explain HPACK"** → Static + dynamic table, Huffman encoding

### Red Flags to Avoid
- Saying HTTP/2 eliminates all head-of-line blocking (only application-level; TCP remains)
- Ignoring TLS in HTTPS discussion
- Not understanding 0-RTT replay risk
- Confusing HTTP/2 and HTTP/3 transport (TCP vs QUIC)
