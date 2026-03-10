# Module 4: Networking

---

## 1. HTTP/2 Multiplexing

### Definition
Sending multiple requests and responses simultaneously over a single TCP connection by splitting data into interleaved binary frames.

### Problem It Solves
HTTP/1.1 processes one request at a time per connection (Head-of-Line blocking). Browsers open 6 parallel connections per domain as a workaround.

### How It Works
```
HTTP/1.1:          HTTP/2:
┌────────────┐     ┌────────────┐
│ Request 1  │     │ Frame:Req1 │
│ Response 1 │     │ Frame:Req2 │  ← interleaved
│ Request 2  │     │ Frame:Req1 │
│ Response 2 │     │ Frame:Res1 │
│ Request 3  │     │ Frame:Req3 │
│ Response 3 │     │ Frame:Res2 │
└────────────┘     │ Frame:Res3 │
  Sequential        └────────────┘
  (blocking)         Multiplexed
```

### Key Features
- **Binary framing**: More efficient than text-based HTTP/1.1
- **Stream prioritization**: Important resources first
- **Header compression (HPACK)**: Reduces header overhead by ~90%
- **Server Push**: Server proactively sends resources

### Remaining Problem
TCP-level HOL blocking: if a TCP packet is lost, ALL streams wait for retransmission (kernel enforces ordering). This is what QUIC/HTTP3 solves.

### Summary
HTTP/2 multiplexes multiple request/response streams over one TCP connection using binary frames. Eliminates application-level HOL blocking but not TCP-level.

---

## 2. QUIC / HTTP/3

### Definition
A transport protocol built on UDP that provides reliable, multiplexed, encrypted connections without TCP's head-of-line blocking.

### Problem It Solves
TCP packet loss blocks ALL streams (kernel-level HOL). TCP handshake + TLS handshake = 2-3 RTT before data flows.

### How It Works
```
TCP + TLS 1.2:                    QUIC:
  SYN ──→                          ClientHello + Key ──→
  ←── SYN-ACK                      ←── ServerHello + Data
  ACK ──→                          (1-RTT setup, 0-RTT on reconnect)
  ClientHello ──→
  ←── ServerHello
  Finished ──→
  (3 RTT before data)

QUIC Stream Independence:
  Stream A: [pkt1] ✓  [pkt2] ✗  [pkt3] ✓
  Stream B: [pkt1] ✓  [pkt2] ✓  [pkt3] ✓
                         ↑
  TCP: ALL streams wait for pkt2 retransmit
  QUIC: Only Stream A waits. Stream B continues.
```

### Key Features
- **0-RTT reconnection**: Cached keys enable instant data on reconnect
- **Per-stream flow control**: Lost packet blocks only its stream
- **Connection migration**: Switch from WiFi to LTE without reconnecting (connection ID, not IP-based)
- **Built-in encryption**: TLS 1.3 mandatory

### Impact
YouTube: 30% reduction in buffering on bad networks. Google: used for ~30% of all internet traffic.

### Summary
QUIC replaces TCP with UDP-based transport, eliminating HOL blocking across streams, enabling 0-RTT connections, and allowing connection migration. It's the foundation of HTTP/3.

---

## 3. gRPC

### Definition
A high-performance RPC framework using Protocol Buffers for serialization and HTTP/2 for transport.

### Problem It Solves
REST + JSON is human-readable but slow to parse, has no type safety, and requires manual client code generation.

### How It Works
```
1. Define service in .proto file:
   service UserService {
     rpc GetUser(UserRequest) returns (User);
     rpc ListUsers(Empty) returns (stream User);  // server streaming
   }

2. Generate client/server code (any language)
3. Binary serialization (10x smaller than JSON)
4. HTTP/2 transport (multiplexed, bidirectional)
```

### Streaming Modes
```
Unary:              Client ──req──→ Server ──res──→ Client
Server Streaming:   Client ──req──→ Server ══res══→ (multiple)
Client Streaming:   Client ══req══→ Server ──res──→ Client
Bidirectional:      Client ══req══⇄══res══ Server
```

### Comparison: gRPC vs REST

| | gRPC | REST |
|-|------|------|
| Serialization | Protobuf (binary) | JSON (text) |
| Transport | HTTP/2 | HTTP/1.1 or 2 |
| Type safety | Strong (.proto schema) | Weak (OpenAPI optional) |
| Streaming | Native bidirectional | SSE only (server→client) |
| Browser support | Limited (grpc-web) | Native |
| Speed | ~10x faster serialization | Slower but universal |

### When to Use
- **gRPC**: Microservice-to-microservice communication (internal)
- **REST**: Public APIs, browser clients, simplicity

### Real Systems
Google (internal), Netflix, Lyft, Uber, Envoy proxy, Kubernetes API (moving to gRPC)

### Summary
gRPC uses Protobuf binary serialization over HTTP/2 for fast, type-safe, streaming-capable RPC. Ideal for internal microservice communication.

---

## 4. Anycast Routing

### Definition
Advertising the same IP address from multiple locations globally. BGP routes each user to the nearest instance automatically.

### How It Works
```
Cloudflare DNS: 1.1.1.1 announced from 300+ cities

User in Tokyo ──BGP──→ 1.1.1.1 (Tokyo POP)
User in London ──BGP──→ 1.1.1.1 (London POP)
User in NYC ──BGP──→ 1.1.1.1 (NYC POP)

DDoS attack from 100Gbps botnet:
  Traffic splits across 300 POPs
  Each POP absorbs ~333Mbps (manageable)
```

### Unicast vs Anycast
```
Unicast: 1 IP → 1 server    (normal)
Anycast: 1 IP → N servers   (nearest wins via BGP)
```

### Use Cases
- **DDoS mitigation**: Attack distributed across all POPs
- **CDN**: Serve content from nearest edge
- **DNS**: Root DNS servers use anycast

### Limitation
TCP connections can break if routing changes mid-session (packets go to different server). Solved by QUIC (connection ID, not IP-based).

### Summary
Anycast advertises one IP from multiple global locations. BGP routes users to the nearest one. Primary use: DDoS absorption and latency reduction.

---

## 5. SSE vs WebSockets

### Comparison
```
┌── Server-Sent Events (SSE) ──┐   ┌── WebSockets ──────────┐
│ Direction: Server → Client    │   │ Direction: Bidirectional│
│ Protocol:  HTTP/1.1+          │   │ Protocol: ws:// (TCP)   │
│ Reconnect: Automatic          │   │ Reconnect: Manual       │
│ Data: Text only               │   │ Data: Text + Binary     │
│ Overhead: Low                 │   │ Overhead: Lowest        │
└───────────────────────────────┘   └─────────────────────────┘
```

### When to Use What

| Use Case | Choice | Why |
|----------|--------|-----|
| Live sports scores | SSE | Server pushes, client reads |
| Chat application | WebSocket | Both sides send messages |
| Stock ticker | SSE | Server-only updates |
| Online gaming | WebSocket | Real-time bidirectional |
| Twitter timeline | SSE/Long-poll | Mostly reading |
| Collaborative editing | WebSocket | Both sides edit |

### Summary
SSE is simpler for server-to-client streaming (auto-reconnect, works with HTTP). WebSockets are needed for bidirectional real-time communication.

---

## 6. Zero-Copy Networking

### Definition
Transferring data from disk to network without copying through user-space memory, using kernel-level `sendfile()`.

### Problem It Solves
Traditional file transfer involves 4 copies and 4 context switches:

```
Traditional (4 copies):
  Disk ──→ Kernel Read Buffer ──→ User Buffer ──→ Socket Buffer ──→ NIC
           copy 1                 copy 2          copy 3           copy 4

Zero-Copy (sendfile):
  Disk ──→ Kernel Read Buffer ──────────────────→ NIC
           copy 1 (DMA)                           copy 2 (DMA)
           (CPU not involved)
```

### Impact
```
Kafka's secret weapon:
  - Consumers read sequential log files
  - sendfile() transfers log data directly to network
  - Result: 60% less CPU, saturates 10Gbps NIC
```

### System Call
```c
// Linux
sendfile(out_fd, in_fd, offset, count);

// Transfers 'count' bytes from in_fd to out_fd
// entirely in kernel space
```

### Real Systems
Kafka, Nginx, Apache, Java NIO (FileChannel.transferTo)

### Summary
Zero-copy uses `sendfile()` to transfer disk data to network without user-space copies. Reduces CPU usage by 60% — the reason Kafka can saturate 10Gbps networks.

---

## 7. Adaptive Bitrate Streaming (ABR)

### Definition
Dynamically switching video quality based on network conditions by encoding video in multiple quality levels and letting the client choose per-chunk.

### How It Works
```
Server stores:
  video_chunk_001_240p.mp4   (100KB)
  video_chunk_001_480p.mp4   (300KB)
  video_chunk_001_1080p.mp4  (1MB)
  video_chunk_001_4K.mp4     (3MB)

Client algorithm (every 4 seconds):
  1. Measure current throughput
  2. Check buffer level
  3. Pick highest quality that won't cause buffering
  4. Download next chunk at that quality
```

### Algorithms
- **BBA (Buffer-Based)**: Choose quality based on buffer fullness
- **MPC (Model Predictive Control)**: Predict future bandwidth, optimize QoE
- **BOLA**: Uses Lyapunov optimization theory

### Netflix's Per-Title Encoding
Each title gets custom quality ladders. A cartoon needs fewer bits than an action movie at the same visual quality → 20% bandwidth savings.

### Summary
ABR encodes video in chunks at multiple qualities. The client dynamically selects the best quality per chunk based on bandwidth and buffer level, preventing buffering.

---

## 8. Kernel Bypass (DPDK / io_uring)

### Definition
Techniques to avoid the kernel network stack overhead, processing packets directly in user space.

### Problem It Solves
Each packet through the kernel = system call + context switch + buffer copies. At 10M+ packets/sec, the kernel becomes the bottleneck.

### Approaches
```
┌─── DPDK ─────────────────────────┐
│ Poll-mode driver in user space    │
│ NIC → User memory (no kernel)     │
│ CPU core dedicated to polling     │
│ Used by: Cloudflare, F5, telecom  │
└───────────────────────────────────┘

┌─── io_uring ─────────────────────┐
│ Shared ring buffer between        │
│   user space and kernel           │
│ Batch submit I/O, batch complete  │
│ Less radical than DPDK            │
│ Used by: modern Linux apps        │
└───────────────────────────────────┘

┌─── eBPF / XDP ───────────────────┐
│ Run code IN the kernel safely     │
│ Process packets before they       │
│   enter the network stack         │
│ Used by: Cloudflare, Cilium       │
└───────────────────────────────────┘
```

### Summary
Kernel bypass (DPDK) processes packets in user space at line rate. io_uring reduces syscall overhead with shared ring buffers. eBPF/XDP runs custom logic in kernel for maximum speed.

---

## 9. Memory-Mapped I/O (mmap)

### Definition
Mapping a file directly into the process's virtual address space, allowing file access via memory reads/writes instead of `read()`/`write()` syscalls.

### How It Works
```
Traditional I/O:
  read(fd, buffer, size)   ← syscall, copy to user buffer

Memory-mapped I/O:
  ptr = mmap(fd)           ← file mapped into address space
  data = ptr[offset]       ← just a memory access (page fault loads data)
```

### Visual
```
Virtual Memory:
  ┌──────────────┐
  │ Code segment  │
  │ Stack         │
  │ Heap          │
  │ mmap region ──┼──→ File on disk
  │  [page1]      │    (loaded on demand via page faults)
  │  [page2]      │
  └──────────────┘
```

### Tradeoffs

| Pros | Cons |
|------|------|
| No explicit read/write syscalls | Page faults are unpredictable |
| OS manages caching (page cache) | Can't control eviction |
| Shared memory between processes | Complex error handling (SIGBUS) |
| Great for read-heavy workloads | mmap thundering herd possible |

### Real Systems
MongoDB (original storage engine), SQLite, Kafka (index files), Elasticsearch (Lucene)

### Warning
Modern databases (Postgres, MySQL) prefer direct I/O because they can manage their own cache more efficiently than the OS page cache.

### Summary
mmap maps files into virtual memory, enabling file access via memory operations. Great for read-heavy workloads but gives up control over caching to the OS.
