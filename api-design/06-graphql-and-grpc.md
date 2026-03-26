# GraphQL & gRPC Essentials

## Schema Design, Resolvers, Mutations, Protobuf, Streaming — When to Use Which

---

## 1. GraphQL

### What Is GraphQL?

A **query language** for APIs — clients specify exactly what data they need.

```
REST Problem:                    GraphQL Solution:
─────────────                    ─────────────────
GET /users/1       → all fields  query { user(id:1) { name, email } }
GET /users/1/orders → all orders   → Only name and email returned
GET /orders/5/items → all items    → Single request, single response
3 requests, over-fetched data      1 request, exact data needed
```

### Schema Definition Language (SDL)

```graphql
# ============================================
# TYPES — define the shape of your data
# ============================================

type User {
  id: ID!                       # ! = non-nullable
  name: String!
  email: String!
  bio: String                   # nullable (no !)
  avatar_url: String
  followers_count: Int!
  created_at: DateTime!
  
  # Relationships (resolved lazily)
  orders: [Order!]!             # non-null list of non-null orders
  followers: [User!]!
  address: Address
}

type Order {
  id: ID!
  status: OrderStatus!
  total: Float!
  items: [OrderItem!]!
  user: User!
  created_at: DateTime!
}

type OrderItem {
  id: ID!
  product: Product!
  quantity: Int!
  unit_price: Float!
}

type Product {
  id: ID!
  name: String!
  price: Float!
  category: Category!
  reviews: [Review!]!
  average_rating: Float
}

# Enum
enum OrderStatus {
  PENDING
  CONFIRMED
  SHIPPED
  DELIVERED
  CANCELLED
}

# Input types (for mutations)
input CreateOrderInput {
  items: [OrderItemInput!]!
  shipping_address_id: ID!
  coupon_code: String
}

input OrderItemInput {
  product_id: ID!
  quantity: Int!
}

# ============================================
# QUERIES — read operations
# ============================================

type Query {
  # Get single resource
  user(id: ID!): User
  product(id: ID!): Product
  order(id: ID!): Order
  
  # List resources (with pagination and filters)
  users(first: Int, after: String, filter: UserFilter): UserConnection!
  products(
    first: Int
    after: String
    category: ID
    min_price: Float
    max_price: Float
    sort: ProductSort
  ): ProductConnection!
  
  # Search
  search(query: String!, type: SearchType): SearchResult!
  
  # Current user
  me: User!
}

# Pagination (Relay-style connections)
type UserConnection {
  edges: [UserEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type UserEdge {
  node: User!
  cursor: String!
}

type PageInfo {
  hasNextPage: Boolean!
  hasPreviousPage: Boolean!
  startCursor: String
  endCursor: String
}

# ============================================
# MUTATIONS — write operations
# ============================================

type Mutation {
  # User
  createUser(input: CreateUserInput!): User!
  updateUser(id: ID!, input: UpdateUserInput!): User!
  deleteUser(id: ID!): Boolean!
  
  # Orders
  createOrder(input: CreateOrderInput!): Order!
  cancelOrder(id: ID!, reason: String): Order!
  
  # Auth
  login(email: String!, password: String!): AuthPayload!
  refreshToken(refreshToken: String!): AuthPayload!
}

type AuthPayload {
  access_token: String!
  refresh_token: String!
  user: User!
}

# ============================================
# SUBSCRIPTIONS — real-time
# ============================================

type Subscription {
  orderStatusChanged(orderId: ID!): Order!
  newMessage(conversationId: ID!): Message!
  newNotification: Notification!
}
```

### Queries — Client Examples

```graphql
# Get user with specific fields (no over-fetching)
query GetUser {
  user(id: "123") {
    name
    email
    followers_count
  }
}

# Get user with nested data (no under-fetching)
query GetUserWithOrders {
  user(id: "123") {
    name
    email
    orders(first: 5) {
      edges {
        node {
          id
          status
          total
          items {
            product {
              name
              price
            }
            quantity
          }
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
}

# Variables (parameterized queries)
query GetProduct($productId: ID!) {
  product(id: $productId) {
    name
    price
    average_rating
    reviews(first: 3) {
      edges {
        node {
          rating
          comment
          user { name }
        }
      }
    }
  }
}
# Variables: { "productId": "prod_456" }

# Fragments (reusable field sets)
fragment UserBasic on User {
  id
  name
  avatar_url
}

query GetTimeline {
  me {
    ...UserBasic
    timeline(first: 20) {
      edges {
        node {
          text
          created_at
          author { ...UserBasic }
          likes_count
        }
      }
    }
  }
}
```

### Mutations — Client Examples

```graphql
# Create an order
mutation CreateOrder($input: CreateOrderInput!) {
  createOrder(input: $input) {
    id
    status
    total
    items {
      product { name }
      quantity
      unit_price
    }
  }
}
# Variables:
# {
#   "input": {
#     "items": [
#       { "product_id": "prod_1", "quantity": 2 },
#       { "product_id": "prod_2", "quantity": 1 }
#     ],
#     "shipping_address_id": "addr_123"
#   }
# }

# Cancel order
mutation CancelOrder {
  cancelOrder(id: "ord_456", reason: "Changed my mind") {
    id
    status
  }
}
```

### Resolvers (Server Implementation)

```javascript
// Resolvers tell GraphQL how to fetch each field

const resolvers = {
  Query: {
    user: async (_, { id }, context) => {
      return await context.db.users.findById(id);
    },
    
    products: async (_, { first, after, category, sort }, context) => {
      const products = await context.db.products.find({
        category,
        limit: first,
        cursor: after,
        sort,
      });
      return toConnection(products);
    },
    
    me: async (_, __, context) => {
      if (!context.user) throw new AuthenticationError('Not authenticated');
      return context.user;
    }
  },
  
  // Field-level resolvers (lazy loading)
  User: {
    orders: async (user, { first, after }, context) => {
      return await context.db.orders.findByUser(user.id, { first, after });
    },
    
    followers_count: async (user, _, context) => {
      return await context.db.follows.count({ following_id: user.id });
    }
  },
  
  Mutation: {
    createOrder: async (_, { input }, context) => {
      if (!context.user) throw new AuthenticationError('Not authenticated');
      return await context.orderService.create(context.user.id, input);
    }
  },
  
  Subscription: {
    orderStatusChanged: {
      subscribe: (_, { orderId }, context) => {
        return context.pubsub.asyncIterator(`ORDER_STATUS_${orderId}`);
      }
    }
  }
};
```

### N+1 Problem & DataLoader

```javascript
// ❌ N+1 Problem: 
// Query: { users { orders { ... } } }
// 1 query for users + N queries for each user's orders

// ✅ Solution: DataLoader (batches + caches)
const DataLoader = require('dataloader');

const orderLoader = new DataLoader(async (userIds) => {
  // Single batched query instead of N queries
  const orders = await db.query(
    'SELECT * FROM orders WHERE user_id = ANY($1)',
    [userIds]
  );
  
  // Return orders grouped by user_id, in same order as userIds
  return userIds.map(id => orders.filter(o => o.user_id === id));
});

// In resolver:
User: {
  orders: (user) => orderLoader.load(user.id)  // batched automatically
}
```

### GraphQL Security Concerns

| Threat | Mitigation |
|--------|------------|
| **Deep nesting** (query bomb) | Max query depth limit (e.g., 10 levels) |
| **Wide queries** (request all fields) | Query complexity analysis, max cost limit |
| **Batch queries** (100 queries in one request) | Limit operations per request |
| **Introspection in production** | Disable `__schema` and `__type` queries |
| **Injection** | Use parameterized variables, validate inputs |

```javascript
// Example: Complexity limit
const { createComplexityRule } = require('graphql-query-complexity');

const complexityRule = createComplexityRule({
  maximumComplexity: 1000,
  estimators: [
    // Each field costs 1, lists cost limit * child cost
    fieldExtensionsEstimator(),
    simpleEstimator({ defaultComplexity: 1 })
  ],
  onComplete: (complexity) => {
    console.log(`Query complexity: ${complexity}`);
  }
});
```

---

## 2. gRPC

### What Is gRPC?

A high-performance RPC framework using Protocol Buffers (protobuf) for serialization and HTTP/2 for transport.

```
REST:                              gRPC:
─────                              ─────
JSON (text, ~2x larger)            Protobuf (binary, compact)
HTTP/1.1 (one request per conn)    HTTP/2 (multiplexed, streaming)
Any language                       Code-generated stubs (type-safe)
Human-readable                     Machine-optimized
```

### Protocol Buffers (Protobuf)

```protobuf
// user.proto — Schema definition

syntax = "proto3";

package ecommerce;

import "google/protobuf/timestamp.proto";

// ============================================
// MESSAGES — data structures
// ============================================

message User {
  string id = 1;                    // field number (not value!)
  string name = 2;
  string email = 3;
  string bio = 4;
  UserStatus status = 5;
  google.protobuf.Timestamp created_at = 6;
  
  // Nested message
  Address address = 7;
  
  // Repeated (list)
  repeated string tags = 8;
}

message Address {
  string street = 1;
  string city = 2;
  string state = 3;
  string zip = 4;
  string country = 5;
}

enum UserStatus {
  UNKNOWN = 0;           // proto3 requires 0 as default
  ACTIVE = 1;
  INACTIVE = 2;
  BANNED = 3;
}

// Request/Response messages
message GetUserRequest {
  string id = 1;
}

message ListUsersRequest {
  int32 page_size = 1;
  string page_token = 2;
  string filter = 3;      // e.g., "status=ACTIVE"
}

message ListUsersResponse {
  repeated User users = 1;
  string next_page_token = 2;
  int32 total_count = 3;
}

message CreateUserRequest {
  string name = 1;
  string email = 2;
  string bio = 3;
}

// ============================================
// SERVICE — RPC definitions
// ============================================

service UserService {
  // Unary RPC (simple request-response)
  rpc GetUser(GetUserRequest) returns (User);
  rpc CreateUser(CreateUserRequest) returns (User);
  rpc UpdateUser(UpdateUserRequest) returns (User);
  rpc DeleteUser(DeleteUserRequest) returns (google.protobuf.Empty);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
  
  // Server streaming (server sends multiple responses)
  rpc WatchUser(GetUserRequest) returns (stream User);
  
  // Client streaming (client sends multiple requests)
  rpc UploadUserPhotos(stream UploadPhotoRequest) returns (UploadPhotoResponse);
  
  // Bidirectional streaming
  rpc Chat(stream ChatMessage) returns (stream ChatMessage);
}
```

### gRPC Streaming Types

```
1. UNARY (most common):
   Client ──request──▶ Server
   Client ◀──response── Server

2. SERVER STREAMING:
   Client ──request──▶ Server
   Client ◀──response 1── Server
   Client ◀──response 2── Server
   Client ◀──response 3── Server
   Use: Live feeds, notifications, event streams

3. CLIENT STREAMING:
   Client ──request 1──▶ Server
   Client ──request 2──▶ Server
   Client ──request 3──▶ Server
   Client ◀──response── Server
   Use: File upload, batch sends, telemetry

4. BIDIRECTIONAL STREAMING:
   Client ──request──▶ Server
   Client ◀──response── Server
   Client ──request──▶ Server
   Client ◀──response── Server
   Use: Chat, real-time collaboration, gaming
```

### gRPC Server Implementation (Go)

```go
// server.go

type userServer struct {
    pb.UnimplementedUserServiceServer
    db *database.DB
}

func (s *userServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    user, err := s.db.GetUser(ctx, req.Id)
    if err != nil {
        return nil, status.Errorf(codes.NotFound, "user %s not found", req.Id)
    }
    return user.ToProto(), nil
}

func (s *userServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
    users, nextToken, total, err := s.db.ListUsers(ctx, req.PageSize, req.PageToken, req.Filter)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
    }
    return &pb.ListUsersResponse{
        Users:         toProtoUsers(users),
        NextPageToken: nextToken,
        TotalCount:    total,
    }, nil
}

// Server streaming
func (s *userServer) WatchUser(req *pb.GetUserRequest, stream pb.UserService_WatchUserServer) error {
    for {
        user, err := s.db.GetUser(stream.Context(), req.Id)
        if err != nil {
            return status.Errorf(codes.NotFound, "user not found")
        }
        if err := stream.Send(user.ToProto()); err != nil {
            return err
        }
        time.Sleep(5 * time.Second)  // poll interval
    }
}

// Start server
func main() {
    lis, _ := net.Listen("tcp", ":50051")
    s := grpc.NewServer(
        grpc.UnaryInterceptor(authInterceptor),  // middleware
        grpc.StreamInterceptor(streamAuthInterceptor),
    )
    pb.RegisterUserServiceServer(s, &userServer{db: database.New()})
    s.Serve(lis)
}
```

### gRPC Client (Go)

```go
// client.go

conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
defer conn.Close()

client := pb.NewUserServiceClient(conn)

// Unary call
user, err := client.GetUser(ctx, &pb.GetUserRequest{Id: "user_123"})
if err != nil {
    st, _ := status.FromError(err)
    if st.Code() == codes.NotFound {
        fmt.Println("User not found")
    }
}

// Server streaming
stream, err := client.WatchUser(ctx, &pb.GetUserRequest{Id: "user_123"})
for {
    user, err := stream.Recv()
    if err == io.EOF {
        break
    }
    fmt.Printf("User update: %v\n", user)
}
```

### gRPC Error Codes

| Code | HTTP Equivalent | When to Use |
|------|----------------|-------------|
| `OK` | 200 | Success |
| `CANCELLED` | 499 | Client cancelled |
| `INVALID_ARGUMENT` | 400 | Bad request |
| `NOT_FOUND` | 404 | Resource not found |
| `ALREADY_EXISTS` | 409 | Duplicate |
| `PERMISSION_DENIED` | 403 | Forbidden |
| `UNAUTHENTICATED` | 401 | Not authenticated |
| `RESOURCE_EXHAUSTED` | 429 | Rate limited |
| `FAILED_PRECONDITION` | 400 | State conflict |
| `ABORTED` | 409 | Concurrency conflict |
| `UNIMPLEMENTED` | 501 | Not implemented |
| `INTERNAL` | 500 | Server error |
| `UNAVAILABLE` | 503 | Service unavailable |
| `DEADLINE_EXCEEDED` | 504 | Timeout |

---

## 3. REST vs GraphQL vs gRPC — Decision Matrix

| Criterion | REST | GraphQL | gRPC |
|-----------|------|---------|------|
| **Client type** | Any (web, mobile, 3rd party) | Mobile, complex UIs | Internal services |
| **Data fetching** | Fixed per endpoint | Client specifies exactly | Fixed per RPC |
| **Over/under fetching** | Common problem | Solved | Solved (by design) |
| **Caching** | Easy (HTTP caching) | Complex (POST-based) | Custom |
| **Learning curve** | Low | Medium | Medium-High |
| **Tooling** | Mature, ubiquitous | Growing (Apollo, Relay) | Code generation |
| **Performance** | Good | Good (with DataLoader) | Best (protobuf + HTTP/2) |
| **Streaming** | SSE / WebSocket | Subscriptions | Native streaming |
| **Browser support** | Native | Native (via HTTP) | Needs gRPC-web proxy |
| **Schema/Contract** | OpenAPI (optional) | SDL (required) | Protobuf (required) |
| **File upload** | Easy (multipart) | Complex | Streaming |

### When to Use What

```
PUBLIC API (third-party developers):
  → REST (universal, simple, well-understood)
  → Examples: Stripe, Twilio, GitHub

MOBILE APP / COMPLEX FRONTEND:
  → GraphQL (avoid over-fetching, single endpoint)
  → Examples: Facebook, Shopify, GitHub (both!)

INTERNAL MICROSERVICES:
  → gRPC (fast, type-safe, streaming)
  → Examples: Uber, Netflix, Google

REAL-TIME:
  → GraphQL Subscriptions (web clients)
  → gRPC Streaming (service-to-service)
  → WebSocket (general bidirectional)

SIMPLE CRUD:
  → REST (don't overcomplicate it)
```

### Polyglot API (Real World)

```
┌──────────────────────────────────────────────────────────┐
│                     API Gateway                           │
│  ┌─────────┐  ┌──────────────┐  ┌────────────────────┐  │
│  │  REST    │  │   GraphQL    │  │ gRPC-web Proxy     │  │
│  │ /api/v1  │  │ /graphql     │  │ (for web clients)  │  │
│  └────┬─────┘  └──────┬───────┘  └────────┬───────────┘  │
│       │               │                    │              │
└───────┼───────────────┼────────────────────┼──────────────┘
        │               │                    │
        ▼               ▼                    ▼
   ┌─────────┐    ┌──────────┐         ┌──────────┐
   │ User    │◀──▶│ Order    │◀──gRPC──▶│ Payment  │
   │ Service │    │ Service  │         │ Service  │
   └─────────┘    └──────────┘         └──────────┘
        ▲               ▲
        │    gRPC        │    gRPC
        ▼               ▼
   ┌─────────┐    ┌──────────┐
   │ Auth    │    │ Search   │
   │ Service │    │ Service  │
   └─────────┘    └──────────┘

External clients → REST or GraphQL
Internal services → gRPC
```

---

## 4. Interview Discussion Points

### GraphQL Questions

| Question | Answer |
|----------|--------|
| "When would you choose GraphQL over REST?" | When clients need flexible queries, mobile apps with bandwidth constraints, or when you have multiple client types needing different data shapes |
| "How do you handle the N+1 problem?" | DataLoader for batching and caching per request |
| "How do you secure a GraphQL API?" | Query depth limits, complexity analysis, rate limiting, authentication via context |
| "How do you do pagination?" | Relay-style cursor connections (edges, nodes, pageInfo) |
| "How do you version a GraphQL API?" | Usually don't — add new fields, deprecate old ones with `@deprecated` |

### gRPC Questions

| Question | Answer |
|----------|--------|
| "Why gRPC over REST for microservices?" | Binary protobuf (10x smaller), HTTP/2 multiplexing, bidirectional streaming, code generation for type safety |
| "How do you handle backward compatibility in protobuf?" | Never reuse field numbers, only add new fields, don't change types |
| "How does gRPC work in browsers?" | grpc-web proxy translates HTTP/2 to HTTP/1.1 |
| "How do you do load balancing with gRPC?" | Client-side (built-in), proxy-based (Envoy), or service mesh (Istio) |
| "How do you handle errors?" | Use standard gRPC status codes + google.rpc.Status for detailed errors |
