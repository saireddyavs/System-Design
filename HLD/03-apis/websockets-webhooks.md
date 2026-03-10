# WebSockets, Webhooks, Long Polling, SSE

## 1. Concept Overview

### Definition
These are mechanisms for real-time or event-driven communication between clients and servers:

- **Short Polling**: Client repeatedly polls server at fixed intervals
- **Long Polling**: Server holds request until data available, then responds
- **SSE (Server-Sent Events)**: Server pushes one-directional events over HTTP
- **WebSockets**: Full-duplex, persistent TCP connection; bidirectional
- **Webhooks**: Server-to-server HTTP callbacks when events occur

### Purpose
- **Real-time updates**: Chat, notifications, live dashboards
- **Efficient push**: Avoid client polling; server initiates when data ready
- **Event notification**: External systems notified of changes (webhooks)

### Problems Each Solves
- **Polling inefficiency**: Wasted requests when no new data
- **Latency**: Polling adds delay; push is immediate
- **Integration**: Webhooks let external systems react to events without polling

---

## 2. Real-World Motivation

### Slack
- **WebSockets**: Real-time message delivery, typing indicators, presence. Single WebSocket connection per client; reconnection with exponential backoff.

### GitHub
- **Webhooks**: Notify external systems on push, PR, issue events. Signature verification (HMAC), retry with exponential backoff.

### Uber
- **WebSockets / Long Polling**: Driver location updates to riders; real-time ETA. High frequency updates (every few seconds).

### Twitter
- **Streaming API**: Long-lived HTTP connection for real-time tweets (similar to SSE). Used for firehose, filtered streams.

### Netflix
- **WebSockets**: Real-time notifications (new content, account changes). Used in internal tools.

---

## 3. Architecture Diagrams

### Short Polling Flow

```
Client                    Server
  │                         │
  │──GET /updates?since=X──▶│
  │                         │ (check for new data)
  │◀──200 {data: []}────────│ (often empty)
  │                         │
  │     (wait 2-5 sec)      │
  │──GET /updates?since=X──▶│
  │◀──200 {data: [new]}────│
  │                         │
```

### Long Polling Flow

```
Client                    Server
  │                         │
  │──GET /updates?since=X──▶│
  │                         │ (hold connection, wait for data)
  │                         │ ...
  │                         │ (data arrives)
  │◀──200 {data: [new]}────│
  │                         │
  │──GET /updates?since=Y──▶│ (immediate next request)
  │                         │ (hold again)
```

### SSE (Server-Sent Events) Flow

```
Client                    Server
  │                         │
  │──GET /stream───────────▶│
  │  Accept: text/event-stream
  │                         │
  │◀──event: update─────────│
  │   data: {"msg": "hi"}   │
  │                         │
  │◀──event: update─────────│
  │   data: {"msg": "bye"}  │
  │                         │ (connection stays open)
```

### WebSocket Handshake and Flow

```
Client                    Server
  │                         │
  │──GET /ws HTTP/1.1──────▶│
  │  Upgrade: websocket     │
  │  Connection: Upgrade    │
  │  Sec-WebSocket-Key: x   │
  │                         │
  │◀──101 Switching Protocols
  │   Upgrade: websocket    │
  │   Sec-WebSocket-Accept  │
  │                         │
  │═════════════════════════│ (full-duplex, frames)
  │◀──Frame: {data}────────│
  │──Frame: {data}────────▶│
  │◀──Ping─────────────────│
  │──Pong──────────────────▶│ (heartbeat)
```

### Webhook Flow

```
Your Server              External Service (e.g., GitHub)
     │                              │
     │  (event occurs: push, PR)     │
     │                              │
     │◀──POST /webhook──────────────│
     │   X-Hub-Signature: sha256=... │
     │   Body: {event, payload}     │
     │                              │
     │──200 OK─────────────────────▶│
     │  (or 5xx → retry)            │
```

---

## 4. Core Mechanics

### Short Polling
- Client sends request every N seconds
- Server returns current state (often empty)
- Simple but wasteful; latency = poll interval

### Long Polling
- Client sends request; server holds until data or timeout
- On response, client immediately sends next request
- Reduces empty responses; server must support long-held connections

### SSE
- HTTP connection kept open; server sends `data:` lines
- One-directional (server → client)
- Auto-reconnect via `Last-Event-ID`
- Text-based; easy to implement

### WebSockets
- Upgrade from HTTP to WebSocket protocol (RFC 6455)
- Full-duplex over single TCP connection
- Binary or text frames
- Ping/Pong for keepalive

### Webhooks
- Server makes HTTP POST to configured URL when event occurs
- Client must be reachable (public URL or tunnel)
- Retries on failure (e.g., 3 retries, exponential backoff)
- Signature (e.g., HMAC) for verification

---

## 5. Numbers

| Approach | Latency | Connections/Client | Server Load | Use Case |
|----------|---------|-------------------|-------------|----------|
| Short Poll | 0.5-5s | 1 per poll | High (many empty) | Simple, low-freq |
| Long Poll | ~0 (event-driven) | 1 held | Medium | Moderate real-time |
| SSE | ~0 | 1 per stream | Low | Server→client push |
| WebSocket | ~0 | 1 persistent | Low | Bidirectional |
| Webhook | ~0 (event-driven) | 0 (server-initiated) | Low | Server-to-server |

---

## 6. Tradeoffs

### Comparison Table

| Aspect | Short Poll | Long Poll | SSE | WebSocket | Webhook |
|--------|------------|-----------|-----|-----------|---------|
| Direction | Client→Server | Client→Server | Server→Client | Both | Server→Server |
| Connection | Per request | Held | Held | Persistent | Per event |
| Protocol | HTTP | HTTP | HTTP | WS (HTTP upgrade) | HTTP |
| Browser support | Full | Full | Full (except IE) | Full | N/A (server) |
| Complexity | Low | Medium | Low | Medium | Medium |

---

## 7. Variants / Implementations

### SSE Variants
- **EventSource API**: Browser native; auto-reconnect
- **Polyfill**: For older browsers

### WebSocket Variants
- **Socket.IO**: Fallback to long polling if WS unavailable
- **SockJS**: Similar fallback strategy

### Webhook Variants
- **Signed payloads**: HMAC-SHA256 in header
- **Retry policies**: Exponential backoff, dead letter queue

---

## 8. Scaling Strategies

| Approach | Strategy |
|----------|----------|
| WebSocket | Sticky sessions, or Redis pub/sub for multi-instance |
| SSE | Sticky sessions; or use message queue to fan out |
| Webhooks | Async queue (e.g., SQS); worker processes send; retry logic |

---

## 9. Failure Scenarios

| Failure | Impact | Mitigation |
|---------|--------|------------|
| WebSocket disconnect | Missed messages | Reconnect, resync, or message queue |
| Webhook endpoint down | Events lost | Retries, dead letter queue |
| Long poll timeout | Client reconnects | Idempotent handlers; client retry |
| SSE connection drop | Missed events | Last-Event-ID for replay |

---

## 10. Performance Considerations

- **WebSocket**: Connection pooling, heartbeat to detect dead connections
- **Webhooks**: Async processing, batch events where possible
- **SSE**: Efficient for one-way; no overhead of WebSocket handshake for simple push

---

## 11. Use Cases

| Use Case | Best Fit |
|----------|----------|
| Chat | WebSocket |
| Live dashboard | SSE or WebSocket |
| Notifications | SSE, WebSocket, or push (FCM) |
| CI/CD events | Webhook |
| Payment confirmation | Webhook |
| Stock ticker | SSE or WebSocket |

---

## 12. Comparison Tables

### When to Use Each

| Scenario | Recommendation |
|----------|----------------|
| Server needs to push to client | SSE (simple) or WebSocket (bidirectional) |
| Client needs to send frequently | WebSocket |
| Notify external system of event | Webhook |
| Simple, infrequent updates | Short polling |
| Can't use WebSocket (proxy, etc.) | Long polling or SSE |

---

## 13. Code or Pseudocode

### WebSocket Server (Pseudocode)

```python
async def handle_websocket(ws):
    await ws.accept()
    try:
        while True:
            data = await ws.receive_text()
            # Process and maybe broadcast
            await ws.send_text(json.dumps({"echo": data}))
    except ConnectionClosed:
        pass
```

### Webhook Signature Verification

```python
def verify_webhook(payload: bytes, signature: str, secret: str) -> bool:
    expected = "sha256=" + hmac.new(
        secret.encode(), payload, hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(expected, signature)
```

### SSE Response

```python
def sse_stream():
    def generate():
        while True:
            data = get_next_event()  # Block until event
            yield f"data: {json.dumps(data)}\n\n"
    
    return Response(
        generate(),
        mimetype="text/event-stream",
        headers={"Cache-Control": "no-cache", "X-Accel-Buffering": "no"}
    )
```

---

## 14. Interview Discussion

### How to Explain
"Short polling is simple but wasteful. Long polling reduces empty responses. SSE is for server-to-client push over HTTP. WebSockets give full-duplex over one connection. Webhooks are server-to-server callbacks for events."

### Follow-up Questions
- "How would you scale WebSockets across multiple servers?"
- "Design a webhook system with retries. How do you avoid duplicates?"
- "When would you choose SSE over WebSocket?"
- "How does WebSocket handle reconnection and message ordering?"

---

## Appendix: Deep Dive

### WebSocket Frame Types

| Opcode | Type | Description |
|--------|------|-------------|
| 0x0 | Continuation | Part of fragmented message |
| 0x1 | Text | UTF-8 text |
| 0x2 | Binary | Binary data |
| 0x8 | Close | Connection close |
| 0x9 | Ping | Heartbeat request |
| 0xA | Pong | Heartbeat response |

### Scaling WebSockets with Redis Pub/Sub

```
Client A ──▶ Server 1 ──┐
                        ├──▶ Redis Pub/Sub ──▶ Broadcast to all servers
Client B ──▶ Server 2 ──┘
```

Each server subscribes to channels; when Server 1 receives a message, it publishes to Redis; all servers receive and forward to their connected clients.

### Webhook Retry Strategy (GitHub-style)

- Retry up to 3-5 times
- Exponential backoff: 5m, 10m, 20m, 40m
- Include `X-GitHub-Delivery` (unique ID) for idempotency

---

## Appendix B: WebSocket Reconnection Strategy

1. **Exponential backoff**: 1s, 2s, 4s, 8s, max 30s
2. **Jitter**: Add random delay to prevent thundering herd
3. **Resync**: On reconnect, client may request missed messages (e.g., since last_seq)
4. **Heartbeat**: Ping every 30s; if no pong, close and reconnect

---

## Appendix C: SSE vs WebSocket Decision Matrix

| Requirement | SSE | WebSocket |
|-------------|-----|-----------|
| Server → Client only | ✓ | ✓ |
| Client → Server frequently | ✗ | ✓ |
| HTTP proxies/firewalls | ✓ (standard HTTP) | May block |
| Binary data | ✗ (text only) | ✓ |
| Auto-reconnect | ✓ (built-in) | Manual |

---

## Appendix D: Webhook Security Best Practices

1. **HTTPS only**: Never send webhooks over HTTP
2. **Signature verification**: HMAC-SHA256 of payload with shared secret
3. **Idempotency**: Use delivery ID to deduplicate
4. **Time limit**: Reject if `X-Webhook-Timestamp` too old (replay attack)

---

## Appendix E: Long Polling Implementation

```python
def long_poll(endpoint, timeout=30):
    start = time.time()
    while time.time() - start < timeout:
        events = get_events_since(last_id)
        if events:
            return events
        time.sleep(0.5)  # Avoid tight loop
    return []
```

Server holds connection by not responding until data available or timeout.

---

## Appendix F: WebSocket Connection Lifecycle

1. **Connect**: HTTP upgrade handshake
2. **Authenticate**: Send auth message over WS (or pass token in query)
3. **Subscribe**: Client subscribes to channels (e.g., user:123, room:456)
4. **Receive**: Server pushes messages
5. **Heartbeat**: Ping/pong every 30s
6. **Reconnect**: On disconnect, exponential backoff, resync state

---

## Appendix G: Short Polling Implementation

```python
def short_poll(client):
    while True:
        response = requests.get(f"/api/updates?since={client.last_sync}")
        for event in response.json()["events"]:
            client.handle(event)
        client.last_sync = response.json()["timestamp"]
        time.sleep(5)  # Poll every 5 seconds
```

Inefficient: Many requests return empty when no updates.

---

## Appendix H: SSE Event Format

```
event: message
data: {"user": "alice", "text": "hello"}
id: 42

event: heartbeat
data: {}

```

- `event`: Event type (optional)
- `data`: Payload (required)
- `id`: For reconnection; client sends `Last-Event-ID: 42`

---

## Appendix I: Webhook Payload Signing (HMAC)

```python
def sign_payload(payload: bytes, secret: str) -> str:
    return "sha256=" + hmac.new(
        secret.encode(), payload, hashlib.sha256
    ).hexdigest()

def verify(payload: bytes, signature: str, secret: str) -> bool:
    expected = sign_payload(payload, secret)
    return hmac.compare_digest(expected, signature)
```

Always use `compare_digest` to prevent timing attacks.

---

## Appendix J: Comparison Summary Table

| Feature | Short Poll | Long Poll | SSE | WebSocket | Webhook |
|---------|------------|-----------|-----|-----------|---------|
| Direction | C→S | C→S | S→C | Both | S→S |
| Latency | Poll interval | Event-driven | Event-driven | Event-driven | Event-driven |
| Server load | High | Medium | Low | Low | Low |
| Implementation | Easiest | Medium | Easy | Medium | Medium |
| Use case | Simple | Moderate real-time | Notifications | Chat, gaming | Integrations |
