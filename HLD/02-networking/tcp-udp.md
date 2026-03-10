# TCP vs UDP

## 1. Concept Overview

### Definition
**TCP (Transmission Control Protocol)** and **UDP (User Datagram Protocol)** are the two primary transport-layer protocols in the Internet Protocol suite. TCP provides reliable, ordered, connection-oriented delivery; UDP provides unreliable, connectionless, minimal-overhead delivery.

### Purpose
- **TCP**: Ensure data arrives correctly, in order, without duplication—critical for applications that cannot tolerate data loss (web, email, file transfer)
- **UDP**: Minimize latency and overhead for applications where timeliness matters more than reliability (gaming, video, VoIP)

### Problems They Solve
1. **Reliability vs Speed**: TCP solves "how do I guarantee delivery?"; UDP solves "how do I minimize delay?"
2. **Multiplexing**: Both allow multiple applications on same host (ports)
3. **Congestion**: TCP actively manages network congestion; UDP does not
4. **Ordering**: TCP guarantees order; UDP does not (application must handle)

---

## 2. Real-World Motivation

### Google
- **Search, Gmail, YouTube (metadata)**: TCP—reliability critical
- **YouTube live streaming**: Initially TCP, moving to QUIC (UDP-based) for lower latency
- **gRPC**: TCP by default, HTTP/2 multiplexing over single connection

### Netflix
- **Video streaming**: TCP (HTTPS) for VOD—reliability over latency
- **Adaptive bitrate**: TCP's congestion control informs bitrate selection
- **Some live**: Exploring QUIC for live events

### Uber
- **API calls**: TCP (HTTPS)—order and reliability for ride data
- **Real-time location**: WebSocket over TCP, or QUIC for lower latency
- **Driver app**: Mix of TCP (critical) and UDP (real-time updates where applicable)

### Amazon
- **E-commerce**: TCP for all transactions
- **S3 transfers**: TCP with multipart upload (parallel TCP streams)
- **GameLift**: UDP for game server communication (low latency)

### Twitter
- **Tweets, API**: TCP
- **Live video (Periscope)**: TCP with adaptive streaming
- **Real-time feed**: WebSocket over TCP

### Gaming (Fortnite, League of Legends)
- **Game state updates**: UDP—30-60 updates/sec, latency critical
- **Chat, purchases**: TCP—reliability required
- **QUIC adoption**: Some games use QUIC for hybrid benefits

---

## 3. Architecture Diagrams

### TCP 3-Way Handshake

```
    CLIENT                                    SERVER
       |                                         |
       |  1. SYN (seq=x)                         |
       |---------------------------------------->|
       |                                         |
       |  2. SYN-ACK (seq=y, ack=x+1)            |
       |<----------------------------------------|
       |                                         |
       |  3. ACK (ack=y+1)                       |
       |---------------------------------------->|
       |                                         |
       |  CONNECTION ESTABLISHED                  |
       |  Data can flow bidirectionally         |
       |                                         |
```

### TCP 4-Way Teardown

```
    CLIENT                                    SERVER
       |                                         |
       |  1. FIN (seq=m)                         |
       |---------------------------------------->|
       |                                         |
       |  2. ACK (ack=m+1)                       |
       |<----------------------------------------|
       |                                         |
       |  3. FIN (seq=n)                         |
       |<----------------------------------------|
       |                                         |
       |  4. ACK (ack=n+1)                       |
       |---------------------------------------->|
       |                                         |
       |  CONNECTION CLOSED                      |
```

### TCP vs UDP Packet Structure

```
TCP HEADER (20-60 bytes):
+--------+--------+--------+--------+--------+--------+
| Source Port (16) | Dest Port (16) | Seq Number (32) |
+--------+--------+--------+--------+--------+--------+
| Ack Number (32)  | Offset|Flags   | Window (16)     |
+--------+--------+--------+--------+--------+--------+
| Checksum (16)   | Urgent Ptr (16) | Options...     |
+--------+--------+--------+--------+--------+--------+

UDP HEADER (8 bytes):
+--------+--------+--------+--------+
| Source Port (16) | Dest Port (16) |
+--------+--------+--------+--------+
| Length (16)      | Checksum (16)  |
+--------+--------+--------+--------+
```

### Head-of-Line Blocking (TCP)

```
TCP STREAM (single connection):
Packet 1 [OK] --> Packet 2 [LOST] --> Packet 3 [OK] --> Packet 4 [OK]
                    |
                    v
        Packet 3, 4 BLOCKED until Packet 2 retransmitted
        Application cannot process 3, 4 until 2 arrives
```

### QUIC (UDP-based, multiple streams)

```
QUIC STREAMS (multiplexed over UDP):
Stream 1: [P1][P2][P3]     --> P2 lost, Stream 1 blocks
Stream 2: [P1][P2][P3]     --> All OK, delivered
Stream 3: [P1][P2][P3]     --> All OK, delivered

Streams 2 and 3 proceed independently; only Stream 1 waits for retransmit
```

---

## 4. Core Mechanics

### TCP Fundamentals

#### 3-Way Handshake (Connection Establishment)
1. **Client → Server: SYN** (synchronize): Client sends seq number x, requests connection
2. **Server → Client: SYN-ACK**: Server acknowledges (ack=x+1), sends own seq y
3. **Client → Server: ACK**: Client acknowledges (ack=y+1)
4. **Connection established**: 1 RTT consumed

#### Reliable Delivery
- **Sequence numbers**: Each byte has a sequence number
- **Acknowledgments**: Receiver sends ACK for received data
- **Retransmission**: Sender retransmits if no ACK within timeout (RTO)
- **Duplicate detection**: Receiver discards duplicates via seq numbers

#### Ordering
- In-order delivery guaranteed by sequence numbers
- Receiver buffers out-of-order segments, delivers in order to application

#### Flow Control (Sliding Window)
- **Receiver window (rwnd)**: Receiver advertises how much buffer space available
- **Sender** cannot send more than min(cwnd, rwnd)
- Prevents overwhelming slow receiver

#### Congestion Control
- **Congestion window (cwnd)**: Sender's estimate of network capacity
- **Slow start**: cwnd doubles each RTT until loss or ssthresh
- **Congestion avoidance**: Additive increase (cwnd += 1/cwnd per ACK)
- **Fast retransmit**: 3 duplicate ACKs → retransmit without waiting for RTO
- **Fast recovery**: Reduce cwnd by half, continue

#### Nagle's Algorithm
- **Problem**: Small packets (e.g., 1-byte) waste bandwidth (40-byte TCP + 20-byte IP = 60 bytes overhead)
- **Nagle**: Buffer small writes until ACK of previous packet received, or buffer full
- **Trade-off**: Reduces small packets but adds latency
- **Disable**: TCP_NODELAY for real-time apps (telnet, games)

### UDP Fundamentals

#### Connectionless
- No handshake, no teardown
- Each datagram independent
- No state at transport layer

#### No Guarantees
- No delivery guarantee
- No order guarantee
- No duplicate detection
- No flow control
- No congestion control

#### Low Overhead
- 8-byte header vs 20-60 bytes TCP
- No retransmission logic
- No connection state

#### Checksum (Optional)
- UDP checksum can be 0 (disabled)
- If non-zero, validates header + payload
- Weaker than TCP (no integrity guarantee in practice)

---

## 5. Numbers

### Header Sizes
| Protocol | Min Header | Max Header | Typical |
|----------|------------|------------|---------|
| **TCP** | 20 bytes | 60 bytes | 20 bytes (no options) |
| **UDP** | 8 bytes | 8 bytes | 8 bytes |
| **IP** | 20 bytes | 60 bytes | 20 bytes |

### Connection Overhead
| Metric | TCP | UDP |
|--------|-----|-----|
| **Handshake RTT** | 1 RTT | 0 |
| **Teardown** | 1-2 RTT (FIN/ACK) | 0 |
| **Per-packet overhead** | 40 bytes (TCP+IP) | 28 bytes (UDP+IP) |

### Latency Impact
| Scenario | TCP | UDP |
|----------|-----|-----|
| **First byte** | 1 RTT (handshake) + 1 RTT (data) = 2 RTT | 1 RTT |
| **Subsequent** | 1 RTT (pipelined) | 1 RTT |
| **Connection reuse** | ~0 (already connected) | N/A |

### Throughput
- **TCP**: Can saturate link (e.g., 10 Gbps) with proper tuning
- **UDP**: No built-in rate limit—can cause congestion collapse if misused
- **Typical TCP throughput**: 1-10 Gbps per connection (depends on RTT, loss)

### Real-World RTT
| Scenario | RTT |
|----------|-----|
| **Same datacenter** | 0.1-1 ms |
| **Same region** | 10-50 ms |
| **Cross-country (US)** | 40-80 ms |
| **Transatlantic** | 80-150 ms |
| **Transpacific** | 150-250 ms |

---

## 6. Tradeoffs

### TCP vs UDP Summary

| Aspect | TCP | UDP |
|--------|-----|-----|
| **Connection** | Connection-oriented | Connectionless |
| **Reliability** | Guaranteed | Best-effort |
| **Ordering** | Guaranteed | None |
| **Overhead** | High (header, handshake) | Low |
| **Latency (first byte)** | 2 RTT | 1 RTT |
| **Congestion control** | Built-in | None (app responsibility) |
| **Flow control** | Built-in | None |
| **Use case** | Web, email, file transfer | Video, gaming, DNS, VoIP |

### When to Use TCP
- Data must not be lost (financial, file transfer)
- Order matters (streaming a file, database replication)
- Application doesn't need to implement reliability
- Bulk transfer (TCP's congestion control helps)

### When to Use UDP
- Latency critical (gaming, VoIP)
- Loss acceptable (video frame drop better than delay)
- Small messages (DNS, NTP)
- Multicast/broadcast (UDP supports; TCP does not)
- Real-time (stock ticker, sensor data)

### QUIC: Best of Both?
- **UDP-based**: No middlebox issues (NAT, firewalls pass UDP)
- **Reliable**: Built-in retransmission, ordering per stream
- **Multiplexed**: Multiple streams, no head-of-line blocking across streams
- **0-RTT**: Resumption for repeat connections
- **Encrypted**: TLS 1.3 built-in

---

## 7. Variants / Implementations

### TCP Variants
- **TCP Reno**: Classic congestion control
- **TCP Cubic**: Default in Linux, optimized for high-BDP
- **TCP BBR**: Google's loss-based alternative, better for high-latency
- **TCP Fast Open (TFO)**: Send data in SYN, 0-RTT for repeat connections
- **MPTCP**: Multipath TCP, use multiple paths

### UDP-Based Protocols
- **DNS**: Query/response, small, loss = retry
- **NTP**: Time sync, one packet
- **QUIC**: HTTP/3, reliable over UDP
- **DTLS**: TLS over UDP (WebRTC)
- **RTP**: Real-time transport (VoIP, video)
- **Custom game protocols**: Often UDP with custom reliability

### Protocol Stack Comparison

```
HTTP/1.1, HTTP/2:     HTTP/3 (QUIC):
+----------+          +----------+
|   HTTP   |          |   HTTP   |
+----------+          +----------+
|   TLS    |          |   TLS    |
+----------+          +----------+
|   TCP    |          |   QUIC   |
+----------+          +----------+
|   IP     |          |   UDP    |
+----------+          +----------+
                      |   IP     |
                      +----------+
```

---

## 8. Scaling Strategies

### TCP Scaling
1. **Connection pooling**: Reuse connections (HTTP keep-alive, gRPC)
2. **Multiple connections**: Parallelize (HTTP/1.1 browser limit 6/domain)
3. **Tuning**: Increase buffer sizes (tcp_rmem, tcp_wmem), enable window scaling
4. **BBR**: Use BBR for high-BDP paths
5. **TFO**: Reduce connection establishment latency

### UDP Scaling
1. **No connection limit**: No state, scale naturally
2. **Application-level reliability**: Implement selective retransmit
3. **Rate limiting**: Prevent congestion collapse
4. **ECN**: Explicit Congestion Notification (if supported)

### QUIC Scaling
- **Connection migration**: IP change doesn't break connection (mobile)
- **Multiplexing**: Single connection, multiple streams
- **0-RTT**: Faster repeat visits

---

## 9. Failure Scenarios

### TCP Failures

#### Connection Reset (RST)
- **Cause**: Server closed connection, port closed, firewall
- **Mitigation**: Connection pooling with health checks, retry with backoff

#### Timeout
- **Cause**: Network partition, severe congestion, NAT table expiry
- **Mitigation**: Shorter timeouts, keep-alive, connection refresh

#### Silently Dropped
- **Cause**: Middlebox (NAT, firewall) drops packets without RST
- **Mitigation**: Application-level timeouts, health checks

### UDP Failures

#### Packet Loss
- **Cause**: Congestion, buffer overflow, wireless
- **Mitigation**: Application retransmit, FEC (forward error correction), accept loss

#### No Feedback
- **Cause**: UDP has no ACK—sender doesn't know if packet arrived
- **Mitigation**: Application-level ACKs if needed (e.g., QUIC)

### Real Production Incidents

#### AWS TCP RST Issue (2018)
- **What**: Some EC2 instances sent spurious RST packets
- **Impact**: Connection failures, retries
- **Lesson**: TCP behavior can be affected by kernel, NIC drivers

#### QUIC Adoption Challenges
- **What**: Some networks block or throttle UDP
- **Mitigation**: Fallback to TCP (HTTP/2) when QUIC fails
- **Reality**: ~2-5% of networks have UDP issues

---

## 10. Performance Considerations

### TCP Tuning (Linux)
```bash
# Increase buffer sizes
net.ipv4.tcp_rmem = 4096 87380 16777216
net.ipv4.tcp_wmem = 4096 65536 16777216

# Enable window scaling
net.ipv4.tcp_window_scaling = 1

# Enable BBR (if available)
net.core.default_qdisc = fq
net.ipv4.tcp_congestion_control = bbr
```

### Reducing TCP Latency
1. **TFO**: TCP Fast Open
2. **Connection reuse**: Keep-alive, connection pooling
3. **Reduce RTT**: Geographic proximity, anycast
4. **Reduce retransmits**: Better network, ECN

### UDP Best Practices
1. **Rate limit**: Don't flood the network
2. **Packet size**: Keep below MTU (1500 - headers = ~1472)
3. **Checksum**: Enable for integrity
4. **Application ACK**: If reliability needed

---

## 11. Use Cases

| Use Case | Protocol | Rationale |
|----------|----------|-----------|
| **Web (HTTP)** | TCP | Reliability, order |
| **HTTPS** | TCP | Same + TLS |
| **HTTP/3** | QUIC (UDP) | Lower latency, multiplexing |
| **DNS** | UDP | Small, one-shot, retry on loss |
| **Video streaming** | TCP (VOD) / UDP (live) | VOD: reliability; Live: latency |
| **VoIP** | UDP/RTP | Latency over reliability |
| **Gaming** | UDP | Real-time, loss acceptable |
| **File transfer** | TCP | Reliability critical |
| **Email** | TCP | Reliability |
| **NTP** | UDP | One packet, retry trivial |
| **gRPC** | TCP | Reliability, streaming |

---

## 12. Comparison Tables

### Protocol Selection Matrix

| Requirement | TCP | UDP | QUIC |
|-------------|-----|-----|------|
| **Reliability** | ✓ | ✗ | ✓ |
| **Ordering** | ✓ | ✗ | ✓ (per stream) |
| **Low latency** | ✗ | ✓ | ✓ |
| **Multiplexing** | ✗ (app-level) | ✗ | ✓ |
| **0-RTT** | ✗ (TFO partial) | N/A | ✓ |
| **NAT traversal** | Harder | Easier | Easier |
| **Middlebox support** | Universal | Good | Growing |

### Congestion Control Algorithms

| Algorithm | Best For | Behavior |
|-----------|----------|----------|
| **Reno** | General | Loss-based, additive increase |
| **Cubic** | High-BDP | Window growth cubic function |
| **BBR** | High latency, lossy | Bandwidth-delay product based |
| **Vegas** | Low latency | Delay-based |

---

## 13. Code or Pseudocode

### TCP Client (Pseudocode)

```python
# TCP: Connection-oriented, reliable
sock = socket(AF_INET, SOCK_STREAM)
sock.connect((host, port))  # 3-way handshake (1 RTT)

# Send data - TCP handles retransmission, ordering
sock.send(b"Hello")
response = sock.recv(1024)

sock.close()  # 4-way teardown
```

### UDP Client (Pseudocode)

```python
# UDP: Connectionless, no guarantees
sock = socket(AF_INET, SOCK_DGRAM)

# No connect required - send immediately
sock.sendto(b"Hello", (host, port))

# May block if no response - application must handle timeout
response, addr = sock.recvfrom(1024)

# No teardown - just close
sock.close()
```

### Application-Level Reliability (UDP)

```python
def reliable_udp_send(sock, data, addr, max_retries=3):
    seq = next_sequence()
    packet = (seq, data)
    
    for attempt in range(max_retries):
        sock.sendto(serialize(packet), addr)
        
        # Wait for ACK with timeout
        if wait_for_ack(sock, seq, timeout=0.1):
            return True
        
        # Retransmit
        continue
    
    return False  # Give up after max_retries
```

### QUIC Connection (Conceptual)

```python
# QUIC: UDP-based, but provides reliability
# First connection: 1 RTT (similar to TCP+TLS 1.3)
# Resumption: 0 RTT (send data immediately)

quic_conn = quic.connect(host, port)

# 0-RTT if resuming (data may be lost, app must handle)
quic_conn.send(b"GET / HTTP/1.1\r\n...", early_data=True)

# Streams are independent - no head-of-line blocking
stream1 = quic_conn.open_stream()
stream2 = quic_conn.open_stream()
stream1.send(b"data1")
stream2.send(b"data2")  # Not blocked by stream1
```

---

## 14. Interview Discussion

### Key Points to Cover
1. **TCP**: Reliable, ordered, connection-oriented, flow/congestion control
2. **UDP**: Unreliable, connectionless, low overhead
3. **Handshake**: TCP adds 1 RTT; UDP has none
4. **Head-of-line blocking**: TCP blocks entire connection on one lost packet
5. **QUIC**: UDP-based, reliable, multiplexed, 0-RTT
6. **When to use each**: Reliability vs latency tradeoff

### Common Follow-ups
- **"Why does HTTP use TCP?"** → Reliability, order, well-understood
- **"Why does DNS use UDP?"** → Small query/response, retry is cheap, low latency
- **"What is head-of-line blocking?"** → One lost packet blocks subsequent in TCP stream
- **"How does QUIC solve it?"** → Multiple streams, loss in one doesn't block others
- **"Explain TCP congestion control"** → Slow start, congestion avoidance, fast retransmit

### Red Flags to Avoid
- Saying "UDP is faster" without context (first byte: yes; bulk transfer: TCP can be faster)
- Ignoring that QUIC is UDP-based
- Not understanding 3-way handshake
- Confusing flow control (receiver) vs congestion control (network)
