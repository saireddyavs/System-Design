# Design ChatGPT

## 1. Problem Statement & Requirements

### Problem Statement
Design an AI chat platform like ChatGPT that enables users to have real-time conversations with a large language model, with token-by-token streaming, conversation history management, context window handling, and optional plugins/tools.

### Functional Requirements
- **Real-time token streaming**: Stream response tokens as they're generated (SSE)
- **Conversation management**: Create, list, rename, delete conversations
- **Context window**: Manage input context (system prompt + history + current query); handle truncation
- **Model serving**: Route to appropriate model (GPT-4, GPT-3.5, etc.); load balancing across GPU clusters
- **Rate limiting & usage metering**: Per-user limits; token counting; billing integration
- **Plugin/tool system**: Function calling; external tool execution (web search, code interpreter)
- **Content moderation**: Safety filters; input/output filtering; policy enforcement

### Non-Functional Requirements
- **Scale**: 100M+ users, billions of tokens/day
- **Latency**: Time to first token (TTFT) < 500ms; tokens/sec throughput
- **Availability**: 99.9% for chat API
- **Cost**: GPU utilization; minimize idle; batch when possible

### Out of Scope
- Model training
- Fine-tuning UI
- Enterprise SSO/SAML (simplified auth)
- Multi-modal (image input) — focus on text

---

## 2. Back-of-Envelope Estimation

### Scale Assumptions
- **Users**: 100M active
- **DAU**: 20M
- **Conversations/user/day**: 5
- **Messages/conversation**: 10
- **Avg tokens/request**: 500 input, 300 output
- **Peak**: 3x average

### QPS Calculation
| Operation | Daily Volume | Peak QPS |
|-----------|--------------|----------|
| Chat completions | 20M × 5 × 10 = 1B | ~35,000 |
| Token streaming | 1B × 300 tokens ≈ 300B tokens | ~10M tokens/s |
| Conversation CRUD | 100M × 5 = 500M | ~6,000 |
| Auth/session | 20M logins | ~700 |

### Storage (1 year)
- **Conversations**: 100M × 50 conv × 20 msg × 500 tokens × 4B ≈ 200 TB (compressed)
- **User metadata**: 100M × 2KB ≈ 200 GB
- **Usage logs**: 1B × 100B ≈ 100 TB

### Bandwidth
- **Input**: 35K QPS × 2KB ≈ 70 MB/s
- **Output (streaming)**: 35K × 300 tokens × 4B ≈ 42 GB/s (peak)
- **SSE**: Chunked; ~50-100 bytes per token event

### GPU
- **Throughput**: 50 tokens/s per A100 (approx)
- **Concurrent requests**: 35K; need ~700 GPUs at 50 tok/s each (simplified)
- **Batching**: Improves utilization; variable batch sizes

---

## 3. API Design

### REST Endpoints

```
# Conversations
POST   /api/v1/conversations
Body: { "title": "..." }
Response: { "conversation_id": "..." }

GET    /api/v1/conversations
Response: { "conversations": [...] }

PATCH  /api/v1/conversations/:id
Body: { "title": "..." }

DELETE /api/v1/conversations/:id

# Chat (Streaming)
POST   /api/v1/chat/completions
Body: {
  "conversation_id": "...",
  "messages": [{ "role": "user", "content": "..." }],
  "model": "gpt-4",
  "stream": true,
  "max_tokens": 1000,
  "temperature": 0.7,
  "tools": [{ "type": "function", "function": {...} }]
}
Response: SSE stream (see below)

# Usage
GET    /api/v1/usage
Response: { "tokens_used": 12345, "limit": 50000 }
```

### SSE (Server-Sent Events) Stream Format

```
data: {"id":"chatcmpl-xxx","choices":[{"delta":{"content":"Hello"},"index":0}]}

data: {"id":"chatcmpl-xxx","choices":[{"delta":{"content":" world"},"index":0}]}

data: {"id":"chatcmpl-xxx","choices":[{"delta":{},"finish_reason":"stop","index":0}]}

data: [DONE]
```

### WebSocket (Alternative)
```
WS /ws/chat
Send: { "conversation_id", "message", "model" }
Receive: { "token": "...", "done": false } ... { "done": true }
```

---

## 4. Data Model / Database Schema

### Database Choice
- **Conversations, messages**: Cassandra or DynamoDB (write-heavy, partition by user)
- **User metadata**: PostgreSQL or DynamoDB
- **Usage/billing**: ClickHouse or TimescaleDB (analytics)
- **Rate limit state**: Redis
- **Model routing**: Config store (etcd/Consul)

### Schema

**Conversations (Cassandra)**
```sql
conversations (
  user_id UUID,
  conversation_id UUID,
  title VARCHAR(200),
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  PRIMARY KEY (user_id, conversation_id)
) WITH CLUSTERING ORDER BY (conversation_id DESC);
```

**Messages (Cassandra)**
```sql
messages (
  conversation_id UUID,
  message_id UUID,
  role VARCHAR(20),       -- user, assistant, system
  content TEXT,
  token_count INT,
  created_at TIMESTAMP,
  PRIMARY KEY (conversation_id, message_id)
) WITH CLUSTERING ORDER BY (message_id ASC);
```

**Users (PostgreSQL)**
```sql
users (
  user_id UUID PRIMARY KEY,
  email VARCHAR(255),
  plan VARCHAR(20),       -- free, plus, team
  created_at TIMESTAMP
)
```

**Usage (ClickHouse/TimescaleDB)**
```sql
token_usage (
  user_id UUID,
  timestamp TIMESTAMP,
  model VARCHAR(50),
  input_tokens INT,
  output_tokens INT,
  conversation_id UUID
)
```

**Rate Limits (Redis)**
```
Key: ratelimit:{user_id}:{window}
Value: token_count
TTL: window (e.g., 3600 for hourly)
```

---

## 5. High-Level Architecture

### ASCII Architecture Diagram

```
                                    ┌─────────────────────────────────────────────────────────────┐
                                    │                    CLIENTS (Web, Mobile, API)                │
                                    └───────────────────────────┬─────────────────────────────────┘
                                                                  │
                                                                  │ HTTPS / SSE
                                                                  ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                              API GATEWAY                                                       │
│                                    Auth, Rate Limit, Request Routing                                          │
└───────────────────────────┬──────────────────────────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┬───────────────────┬───────────────────┐
        ▼                   ▼                   ▼                   ▼                   ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│ Conversation  │   │ Chat          │   │ Usage         │   │ Moderation     │   │ Plugin/Tool   │
│ Service       │   │ Orchestrator  │   │ Service       │   │ Service       │   │ Service       │
│               │   │               │   │               │   │               │   │               │
│ CRUD          │   │ Context       │   │ Metering      │   │ Input filter   │   │ Function      │
│ History       │   │ assembly      │   │ Billing       │   │ Output filter  │   │ calling       │
└───────┬───────┘   └───────┬───────┘   └───────┬───────┘   └───────┬───────┘   └───────┬───────┘
        │                   │                   │                   │                   │
        │                   │                   │                   │                   │
        ▼                   ▼                   ▼                   ▼                   ▼
┌───────────────┐   ┌───────────────────────────────────────────────────────────────────────────────┐
│ Cassandra     │   │                    MODEL SERVING LAYER                                         │
│ (Conv, Msgs)  │   │                                                                               │
└───────────────┘   │  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐   │
                    │  │ Load        │───▶│ GPU Cluster │    │ GPU Cluster │    │ GPU Cluster │   │
                    │  │ Balancer    │    │ (GPT-4)     │    │ (GPT-3.5)   │    │ (Region 2)  │   │
                    │  │             │    │             │    │             │    │             │   │
                    │  │ - Model     │    │ vLLM /      │    │ vLLM /      │    │ ...         │   │
                    │  │   routing   │    │ TensorRT    │    │ TensorRT    │    │             │   │
                    │  │ - Queue     │    │             │    │             │    │             │   │
                    │  └─────────────┘    └──────┬──────┘    └──────┬──────┘    └──────┬──────┘   │
                    │                            │                   │                   │         │
                    │                            └───────────────────┴───────────────────┘         │
                    │                                        │                                      │
                    │                                        ▼                                      │
                    │                            Token-by-token stream (SSE)                        │
                    └───────────────────────────────────────────────────────────────────────────────┘
```

---

## 6. Detailed Component Design

### 6.1 Token-by-Token Streaming Architecture

1. **Request**: Client sends messages; `stream: true`
2. **Context assembly**: Orchestrator fetches conversation history; truncates to fit context window (e.g., 128K tokens)
3. **Moderation**: Input filter checks for policy violations
4. **Model routing**: Select GPU cluster by model (gpt-4 vs gpt-3.5)
5. **Inference**: Model generates tokens; each token pushed to response stream immediately
6. **Output moderation**: Stream tokens through output filter (async or buffered)
7. **SSE**: Each token sent as `data: {...}\n\n`
8. **Persistence**: After stream ends, save assistant message to Cassandra
9. **Usage**: Count tokens; update Redis/ClickHouse

### 6.2 Model Serving Infrastructure

**GPU Clusters**
- vLLM, TensorRT-LLM, or custom serving
- Continuous batching: Add new requests to batch as others complete
- KV cache: Cached for prefix (system prompt, history) to avoid recompute

**Load Balancing**
- Round-robin or least-loaded across GPU workers
- Model-specific pools: gpt-4 pool, gpt-3.5 pool
- Queue: If all GPUs busy, queue with timeout; return 503 if overloaded

**Scaling**
- Horizontal: Add GPU nodes
- Spot instances for cost savings (with fallback)

### 6.3 Context Window Management

- **Limit**: 128K tokens (model-dependent)
- **Composition**: System prompt (2K) + history (N messages) + current query
- **Truncation**: Drop oldest messages; or summarize old messages
- **Sliding window**: Keep last N tokens
- **Summary**: Periodically summarize long conversations; store summary; use as context

### 6.4 Rate Limiting & Usage Metering

- **Redis**: `INCR` per user per window; `EXPIRE`
- **Tiers**: Free (50 req/day), Plus (higher), Team (custom)
- **Token counting**: Tiktoken or similar; count input + output
- **Billing**: Usage events → billing service; monthly aggregation

### 6.5 Plugin / Tool System (Function Calling)

- **Schema**: Tools defined with name, description, parameters (JSON Schema)
- **Flow**: Model returns `tool_calls` in response; orchestrator executes; inject result; continue generation
- **Execution**: Sandboxed; timeout; rate limit per tool
- **Examples**: Web search, code interpreter, database query

### 6.6 Content Moderation & Safety Filters

- **Input**: Before inference; block or flag
- **Output**: Per-token or per-chunk; block if policy violation
- **Model**: Fine-tuned classifier or rule-based
- **Fallback**: Refuse to answer; generic response

---

## 7. Scaling

### Horizontal Scaling
- **API**: Stateless; scale behind load balancer
- **Model serving**: Add GPU nodes; load balancer distributes
- **Cassandra**: Add nodes; partition by user_id

### Caching
- **Context cache**: KV cache for repeated prefixes (same system prompt)
- **Model weights**: Loaded on GPU; shared across requests
- **User metadata**: Redis cache

### Queue
- **Request queue**: If GPUs saturated; FIFO or priority by tier
- **Async**: For non-streaming; return job ID; poll for result

### Batching
- **Continuous batching**: vLLM-style; dynamic batch size
- **Improves**: GPU utilization from ~30% to 70%+

---

## 8. Failure Handling

### Model Serving Failure
- Health check; remove unhealthy workers from pool
- Retry on different worker
- Fallback to smaller model (gpt-4 → gpt-3.5) if configured

### Stream Interruption
- Client reconnects; resend last token offset; server resumes (complex)
- Simpler: Client shows partial response; user can retry

### Cassandra Unavailable
- Circuit breaker; return cached conversation (if any)
- Degrade: Allow chat without history

### Rate Limit Redis Down
- Fail open: Allow requests (avoid blocking all users)
- Or fail closed: Reject (safer for cost)

---

## 9. Monitoring & Observability

### Key Metrics
- **TTFT**: Time to first token (p50, p99)
- **Tokens/sec**: Throughput per request
- **GPU utilization**: Per cluster
- **Queue depth**: Pending requests
- **Error rate**: Per model, per endpoint
- **Usage**: Tokens per user, cost

### Alerts
- TTFT p99 > 2s
- Error rate > 1%
- GPU utilization < 20% (inefficient) or > 95% (saturation)
- Queue depth > 1000

### Tracing
- Trace: Request → Orchestrator → Model → Stream
- Token-level latency optional

---

## 10. Interview Tips

### Follow-up Questions
- "How do you minimize time to first token?"
- "How would you implement conversation summarization for long chats?"
- "How does function calling work end-to-end?"
- "How do you handle a GPU node failure mid-generation?"
- "How would you add support for 100 different fine-tuned models?"

### Common Mistakes
- **Ignoring streaming**: Core UX; must design for SSE/token flow
- **No context management**: 128K limit is real; truncation strategy needed
- **Single GPU**: Need cluster, load balancing, queue
- **No rate limiting**: Cost explosion

### Key Points to Emphasize
- **Streaming**: SSE; token-by-token; reduces perceived latency
- **Model serving**: GPU clusters; continuous batching; load balancing
- **Context window**: Truncation, summarization, sliding window
- **Scale**: 100M users, billions of tokens/day
- **Plugins**: Function calling; tool execution; inject results

---

## Appendix: Extended Design Details & Walkthrough Scenarios

### A. SSE vs WebSocket for Streaming

| Aspect | SSE | WebSocket |
|--------|-----|-----------|
| Direction | Server → Client only | Bidirectional |
| Protocol | HTTP | WS (upgrade) |
| Reconnect | Built-in (EventSource) | Manual |
| Complexity | Simpler | More complex |
| Use case | Token streaming | Chat with back-and-forth |

**ChatGPT uses SSE**: Client sends HTTP POST; server streams tokens. No need for bidirectional.

### B. Context Window Truncation Strategies

1. **Drop oldest**: Remove oldest user/assistant pairs until under limit
2. **Summarize**: Run summarization model on old messages; replace with summary
3. **Sliding window**: Keep last N tokens
4. **Importance**: Score messages; keep most "important" (e.g., recent + high relevance)

### C. Continuous Batching (vLLM)

- **Static batching**: Batch of 8; wait for all to finish → poor utilization
- **Continuous**: As each request completes, remove from batch; add new request
- **Result**: GPU always busy; higher throughput

### D. Function Calling Flow

1. User: "What's the weather in SF?"
2. Model returns: `tool_calls: [{ "name": "get_weather", "args": { "city": "San Francisco" } }]`
3. Orchestrator calls `get_weather("San Francisco")` → "72°F, sunny"
4. Orchestrator injects: `{ "role": "tool", "content": "72°F, sunny" }`
5. Model continues: "The weather in San Francisco is 72°F and sunny."
6. Stream this to client

### E. KV Cache Reuse

- **Prefix**: System prompt + conversation history often same across turns
- **KV cache**: Store key-value cache for this prefix
- **New turn**: Only compute for new user message + assistant response
- **Saves**: Significant compute for long conversations

### F. Rate Limit Tiers

| Tier | Requests/day | Tokens/min |
|------|--------------|------------|
| Free | 50 | 1000 |
| Plus | 500 | 5000 |
| Team | Custom | Custom |

### G. Moderation Pipeline

1. **Input**: Regex + ML classifier for PII, harmful content
2. **Output**: Per-chunk check; if violation, truncate and append "[Content filtered]"
3. **Log**: Flagged content for review
4. **Appeal**: User can appeal; human review

### H. Conversation Persistence Flow

1. User sends message → save to `messages` (user role)
2. Model streams response
3. On stream end: Save assistant message to `messages`
4. Update `conversations.updated_at`
5. Update usage in Redis + async to ClickHouse

### I. Multi-Region Model Serving

- **Challenge**: GPU clusters expensive; not in every region
- **Approach**: Centralized GPU region; API in multiple regions
- **Latency**: API region → GPU region adds ~50-100ms
- **Mitigation**: Edge for auth, routing; inference in GPU region

### J. Cost Optimization

- **Spot instances**: 70% cheaper; handle preemption
- **Quantization**: INT8/INT4 for inference; 2x speedup
- **Smaller models**: Route simple queries to GPT-3.5
- **Caching**: Cache common prompts (e.g., "Explain X")
- **Batch**: Non-urgent requests batched
