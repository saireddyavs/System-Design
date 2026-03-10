# REST vs GraphQL vs gRPC

## 1. Concept Overview

### Definition
**REST** (Representational State Transfer) is an architectural style using HTTP methods and resource URIs. **GraphQL** is a query language and runtime that lets clients request exactly the data they need. **gRPC** is a high-performance RPC framework using Protocol Buffers and HTTP/2.

### Purpose
- **REST**: Simple, cacheable, HTTP-native resource access
- **GraphQL**: Flexible queries, single endpoint, client-driven data shape
- **gRPC**: Fast, typed, streaming-capable service-to-service communication

### Problems Each Solves
- **REST**: Over-fetching (getting full resource when only need subset), under-fetching (multiple round-trips for related data)
- **GraphQL**: N+1 queries, rigid endpoints, versioning complexity
- **gRPC**: JSON overhead, lack of streaming, weak typing in polyglot environments

---

## 2. Real-World Motivation

### GitHub
- **GraphQL API**: Primary API for complex queries (repos, issues, PRs with nested data). Single request fetches exactly what the client needs.
- **REST API**: Still used for simple CRUD, webhooks, OAuth

### Google
- **gRPC internally**: All microservices use gRPC; Protocol Buffers for schema, HTTP/2 for multiplexing. Stubby (internal RPC) evolved to gRPC.

### Netflix
- **GraphQL Federation**: Multiple GraphQL schemas composed into one; each domain (e.g., user, content) owns its schema. Enables unified API for UIs.

### Uber
- **gRPC for microservices**: Driver location, trip matching, pricing—all use gRPC for low latency and high throughput.
- **REST for external**: Partner and public APIs use REST for simplicity.

### Amazon
- **REST APIs**: AWS services expose REST; SDKs handle signing and retries.
- **Internal**: Various RPC mechanisms for service-to-service.

### Twitter
- **REST + GraphQL**: REST for public API; GraphQL for internal Twitter clients (mobile, web) for flexible data fetching.

---

## 3. Architecture Diagrams

### REST Architecture

```
┌─────────────┐                    ┌─────────────────────────────────────┐
│   Client    │                    │              REST API               │
│             │  GET /users/1      │  ┌─────────────────────────────┐   │
│             │──────────────────▶│  │  User Service                │   │
│             │                    │  │  GET /users/1 → User DB      │   │
│             │                    │  └─────────────────────────────┘   │
│             │  GET /users/1/orders│  ┌─────────────────────────────┐   │
│             │──────────────────▶│  │  Order Service               │   │
│             │                    │  │  GET /users/1/orders → Order DB│   │
│             │  GET /orders/5     │  └─────────────────────────────┘   │
│             │──────────────────▶│  (Multiple round-trips)             │
└─────────────┘                    └─────────────────────────────────────┘
```

### GraphQL Architecture

```
┌─────────────┐                    ┌─────────────────────────────────────┐
│   Client    │                    │           GraphQL Server            │
│             │  Single Query:     │  ┌─────────────────────────────┐   │
│             │  { user(id:1) {    │  │  Resolver: user              │   │
│             │    name, orders {  │  │  → User Service              │   │
│             │      total, items  │  │  Resolver: user.orders        │   │
│             │    }               │  │  → Order Service (DataLoader) │   │
│             │  }                 │  │  (Batched, single round-trip) │   │
│             │  }                 │  └─────────────────────────────┘   │
│             │──────────────────▶│  ┌─────────────────────────────┐   │
│             │                    │  │  Schema: type User, Order     │   │
│             │  Response: exact   │  │  Introspection: schema query  │   │
│             │  shape requested   │  └─────────────────────────────┘   │
└─────────────┘                    └─────────────────────────────────────┘
```

### gRPC Architecture

```
┌─────────────┐     HTTP/2           ┌─────────────────────────────────────┐
│   Client    │     Multiplexed      │  gRPC Server                         │
│  (Stub)     │     Single TCP       │  ┌─────────────────────────────┐   │
│             │     Connection       │  │  Service: UserService       │   │
│             │◀───────────────────▶│  │  rpc GetUser(id) → User      │   │
│             │     Binary Proto     │  │  rpc GetUsers(stream id)     │   │
│             │     Streaming:       │  │  rpc WatchLocation(stream)  │   │
│             │     Unary/Bi/Stream  │  └─────────────────────────────┘   │
└─────────────┘                    └─────────────────────────────────────┘
```

### REST Over-fetching vs Under-fetching

```
OVER-FETCHING (REST):                    UNDER-FETCHING (REST):
┌────────────────────────────────┐      ┌────────────────────────────────┐
│ GET /users/1                    │      │ GET /users/1       → User       │
│ Returns: id, name, email,       │      │ GET /users/1/orders → Orders    │
│   bio, avatar, settings, ...    │      │ GET /orders/5/items → Items     │
│ Client only needs: name, email  │      │ 3 round-trips for 1 screen      │
└────────────────────────────────┘      └────────────────────────────────┘
```

---

## 4. Core Mechanics

### REST
- **Resource-based**: URLs represent resources; HTTP methods define actions
- **Stateless**: No server-side session; each request self-contained
- **Over-fetching**: GET returns full resource; client may ignore fields
- **Under-fetching**: Related data requires additional requests (e.g., `/users/1/orders`)

### GraphQL
- **Schema-first**: Schema defines types, queries, mutations

```graphql
type User {
  id: ID!
  name: String!
  orders: [Order!]!
}
type Query {
  user(id: ID!): User
}
```

- **Resolvers**: Each field has a resolver function; can hit different backends
- **N+1 problem**: Resolver for `user.orders` called per user → N+1 queries

**DataLoader** (batching):
```javascript
// Without DataLoader: N+1 queries for 100 users
// With DataLoader: 1 batch query for all orders
const orderLoader = new DataLoader(async (userIds) => {
  const orders = await db.orders.findByUserIds(userIds);
  return userIds.map(id => orders.filter(o => o.userId === id));
});
```

- **Introspection**: `__schema` query returns schema; enables tooling (GraphiQL, codegen)
- **Subscriptions**: WebSocket-based for real-time updates

### gRPC
- **Protocol Buffers**: Schema in `.proto`; compiled to language-specific code

```protobuf
message User {
  string id = 1;
  string name = 2;
}
service UserService {
  rpc GetUser(GetUserRequest) returns (User);
  rpc ListUsers(stream ListRequest) returns (stream User);  // Bidirectional
}
```

- **HTTP/2**: Multiplexing, header compression, binary framing
- **Streaming**: Unary (request-response), server stream, client stream, bidirectional

---

## 5. Numbers

| Metric | REST | GraphQL | gRPC |
|--------|------|---------|------|
| Payload size (vs JSON) | Baseline | Similar or smaller | 30-50% smaller (binary) |
| Latency (ms) | 50-200 | 50-200 (similar) | 10-50 (lower overhead) |
| Throughput (req/s) | 1-10K | 1-10K | 10-100K |

---

## 6. Tradeoffs

### Paradigm Comparison

| Aspect | REST | GraphQL | gRPC |
|--------|------|---------|------|
| Paradigm | Resource | Query | RPC |
| Transport | HTTP/1.1 or HTTP/2 | HTTP/1.1 or HTTP/2 | HTTP/2 |
| Serialization | JSON | JSON | Protocol Buffers |
| Streaming | Limited (chunked) | Subscriptions (WS) | Native (unary, stream) |
| Caching | HTTP cache (GET) | Complex (query-based) | Application-level |
| Browser support | Native | Native | Limited (gRPC-Web) |

---

## 7. Variants / Implementations

### REST Variants
- **JSON:API**: Standard for JSON APIs with relationships, sparse fieldsets
- **Hypermedia (HATEOAS)**: Links in responses for discoverability

### GraphQL Variants
- **Apollo**: Client, server, federation
- **GraphQL Federation**: Multiple schemas stitched (Netflix, Apollo Federation)
- **Schema Stitching**: Combine schemas at gateway

### gRPC Variants
- **gRPC-Web**: Proxy (Envoy) for browser; uses HTTP/1.1 or HTTP/2
- **gRPC-Gateway**: REST to gRPC translation (generate REST API from proto)

---

## 8. Scaling Strategies

| Approach | REST | GraphQL | gRPC |
|----------|------|---------|------|
| Horizontal | Stateless, easy | Stateless, resolvers can scale | Stateless, easy |
| Caching | CDN, HTTP cache | Query cache, persisted queries | Application cache |
| Batching | Batch endpoints | DataLoader | Native streaming |

---

## 9. Failure Scenarios

| Failure | REST | GraphQL | gRPC |
|---------|------|---------|------|
| Expensive query | Client controls URL | Client can request huge graph → limit depth, complexity | Server controls |
| N+1 | Multiple endpoints | DataLoader batching | N/A (single RPC) |
| Schema evolution | Versioning | Additive changes, deprecation | Proto backward compatibility |

---

## 10. Performance Considerations

- **REST**: Use field selection, pagination, compression
- **GraphQL**: Persisted queries, query complexity limits, DataLoader
- **gRPC**: Connection reuse, keep-alive, streaming for large payloads

---

## 11. Use Cases

| Use Case | Best Fit |
|----------|----------|
| Public API, third-party | REST |
| Mobile app, complex UI | GraphQL |
| Microservices (internal) | gRPC |
| Real-time (subscriptions) | GraphQL or WebSocket |
| High-throughput streaming | gRPC |

---

## 12. Comparison Tables

### Full Comparison

| Column | REST | GraphQL | gRPC |
|--------|------|---------|------|
| Paradigm | Resource-based | Query language | RPC |
| Transport | HTTP/1.1, HTTP/2 | HTTP, WebSocket | HTTP/2 |
| Serialization | JSON | JSON | Protocol Buffers |
| Streaming | Chunked transfer | Subscriptions | Unary, server, client, bidirectional |
| Browser support | Full | Full | gRPC-Web (proxy) |
| Tooling | OpenAPI, Postman | GraphiQL, Apollo | Codegen, grpcurl |
| Use case | Public APIs, CRUD | Flexible UIs, aggregations | Internal microservices |

### When to Use Each

| Scenario | Recommendation |
|----------|----------------|
| Public API for partners | REST |
| Mobile app with many screens | GraphQL |
| Service-to-service (same org) | gRPC |
| Real-time dashboards | GraphQL subscriptions or WebSocket |
| Streaming large datasets | gRPC streaming |

---

## 13. Code or Pseudocode

### REST Example

```http
GET /users/123
Host: api.example.com
Accept: application/json

Response:
{
  "id": "123",
  "name": "Alice",
  "email": "alice@example.com"
}
```

### GraphQL Example

```graphql
query GetUserWithOrders($userId: ID!) {
  user(id: $userId) {
    name
    email
    orders(first: 5) {
      id
      total
      items {
        name
        quantity
      }
    }
  }
}
```

### gRPC Example

```protobuf
syntax = "proto3";
package user;

service UserService {
  rpc GetUser(GetUserRequest) returns (User);
  rpc StreamUsers(StreamUsersRequest) returns (stream User);
}

message GetUserRequest {
  string id = 1;
}

message User {
  string id = 1;
  string name = 2;
}
```

```go
// Server
func (s *server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
  return s.db.GetUser(req.Id)
}
```

---

## 14. Interview Discussion

### How to Explain
"REST is resource-oriented and HTTP-native; simple but can over/under-fetch. GraphQL gives clients control over the response shape; great for UIs but requires careful design (N+1, complexity). gRPC is for high-performance internal services; binary, typed, streaming-capable."

### Follow-up Questions
- "How would you prevent a client from requesting a 1000-level deep GraphQL query?"
- "How does DataLoader solve the N+1 problem?"
- "When would you use gRPC over REST for a public API?"
- "Design a system that uses both REST and GraphQL. How do they coexist?"

---

## Appendix: Deep Dive Topics

### GraphQL N+1 Problem in Detail

When a query requests `users { orders { items } }`, the resolver for `orders` is invoked once per user. Without batching:

```
Resolver: users → [User1, User2, User3]
Resolver: User1.orders → Query DB (1)
Resolver: User2.orders → Query DB (2)
Resolver: User3.orders → Query DB (3)
Total: 4 queries
```

With DataLoader, orders are batched:

```
Resolver: users → [User1, User2, User3]
DataLoader: collect userIds [1,2,3]
Resolver: batch load orders for [1,2,3] → 1 query
Total: 2 queries
```

### gRPC Streaming Types

| Type | Request | Response | Use Case |
|------|---------|----------|----------|
| Unary | Single | Single | Standard RPC |
| Server stream | Single | Stream | Server pushes (e.g., log tail) |
| Client stream | Stream | Single | Client uploads (e.g., batch) |
| Bidirectional | Stream | Stream | Chat, real-time sync |

### REST Caching with ETag

```
GET /users/123
Response: 200 OK, ETag: "abc123"

GET /users/123
If-None-Match: "abc123"
Response: 304 Not Modified (no body)
```

### GraphQL Persisted Queries

Instead of sending full query string, client sends hash; server has mapping. Reduces payload, enables allowlisting.

---

## Appendix B: GraphQL Complexity Analysis

Prevent expensive queries by assigning cost to fields:

```graphql
type Query {
  user(id: ID!): User    # cost: 1
}
type User {
  name: String           # cost: 1
  friends: [User!]!     # cost: 10 (multiplier)
}
```

Query `user { friends { friends { name } } }` → depth 3, list multiplier → reject if > max complexity.

---

## Appendix C: gRPC Error Handling

gRPC uses status codes (similar to HTTP):

| Code | Meaning |
|------|---------|
| OK | Success |
| INVALID_ARGUMENT | Bad request |
| NOT_FOUND | Resource not found |
| ALREADY_EXISTS | Duplicate |
| RESOURCE_EXHAUSTED | Rate limited |
| UNAVAILABLE | Service down |

---

## Appendix D: REST vs GraphQL Caching

**REST**: GET requests cacheable by URL. CDN, browser cache work well.

**GraphQL**: Single endpoint, POST for queries (often). Cache key = hash(query + variables). More complex; persisted queries help.

---

## Appendix E: Real-World Architecture Examples

### GitHub's Dual API
- **REST**: Simple CRUD, webhooks, OAuth, file uploads
- **GraphQL**: Complex queries (repo with issues, PRs, reviews in one request), real-time features

### Netflix GraphQL Federation
- Each domain team owns a subgraph (e.g., user, content, playback)
- Gateway composes into unified schema
- Enables independent deployment

### Google's gRPC Ecosystem
- All internal services use gRPC
- Stubby (internal) evolved to gRPC for open source
- Envoy for gRPC-Web at edge (browser support)

---

## Appendix F: GraphQL Subscriptions Architecture

```
Client ◀──WebSocket──▶ GraphQL Server ◀──Pub/Sub (Redis)──▶ Backend Events
```

Server subscribes to backend events; pushes to connected clients via WebSocket.
