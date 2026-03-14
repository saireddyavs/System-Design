# Remote Procedure Call (RPC)

> Staff+ Engineer Level — FAANG Interview Deep Dive

---

## 1. Concept Overview

### Definition

**Remote Procedure Call (RPC)** is a mechanism that allows a program to execute a procedure (function) on a remote machine as if it were a local call. The caller invokes a function; the RPC framework handles marshaling, transport, and unmarshaling transparently.

### Purpose

- **Abstraction**: Hide network complexity; call remote code like local
- **Performance**: Binary protocols; efficient serialization
- **Type safety**: IDL defines contracts; generated stubs
- **Inter-service communication**: Microservices, distributed systems

### Problems Solved

| Problem | Solution |
|---------|----------|
| Network complexity | Stubs hide marshaling, transport |
| Serialization | IDL + codegen (Protobuf, etc.) |
| Service discovery | Built into framework or separate |
| Cross-language | IDL generates clients for many languages |

---

## 2. Real-World Motivation

### LinkedIn

- **Adopted Protocol Buffers** — reduced latency by **60%** vs JSON
- gRPC/Thrift for internal services
- High-throughput, low-latency requirements

### Google

- **gRPC** — open-source RPC framework
- **Protocol Buffers** — IDL and serialization
- Used internally (Stubby); gRPC is public evolution

### Netflix

- **gRPC** for inter-service communication
- **Thrift** historically; migrating to gRPC

### Uber

- **gRPC** for real-time, high-throughput services
- Protocol Buffers for schema evolution

### Microsoft

- **gRPC** in .NET for microservices
- **WCF** (legacy) — Windows Communication Foundation

---

## 3. Architecture Diagrams

### RPC Flow — Client to Server

```
    Client                          Server
    ┌─────────────┐                 ┌─────────────┐
    │ Application │                 │ Application │
    │  get_user() │                 │  get_user() │
    └──────┬──────┘                 └──────┬──────┘
           │                               │
           │  (local call)                 │  (local call)
           ▼                               ▲
    ┌─────────────┐                 ┌─────────────┐
    │ Client Stub │                 │ Server Stub │
    │ Marshal     │                 │ Unmarshal   │
    └──────┬──────┘                 └──────┬──────┘
           │                               │
           │  Serialized request            │  Deserialize
           │  (Protobuf, etc.)             │
           ▼                               │
    ┌─────────────┐                 ┌──────┴──────┐
    │  Transport  │  ────network───> │  Transport  │
    │  (HTTP/2,   │                 │  (HTTP/2)   │
    │   TCP)      │  <───network──── │              │
    └─────────────┘                 └─────────────┘
```

### Stub Generation from IDL

```
    ┌─────────────────────────────────────────┐
    │  IDL (Interface Definition Language)    │
    │  service UserService {                  │
    │    rpc GetUser(GetUserReq) returns User;│
    │  }                                      │
    │  message User { ... }                   │
    └──────────────────┬────────────────────┘
                        │
                        │  protoc / thrift compiler
                        ▼
    ┌─────────────────────────────────────────────────────┐
    │  Generated Code                                      │
    │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │
    │  │ Client Stub │  │ Server Stub │  │ Data types  │  │
    │  │ (Java)      │  │ (Java)      │  │ (User, etc) │  │
    │  └─────────────┘  └─────────────┘  └─────────────┘  │
    │  Same for Go, Python, C++, etc.                     │
    └─────────────────────────────────────────────────────┘
```

### gRPC Request/Response Flow

```
    Client                    gRPC Server
       │                           │
       │  HTTP/2 stream            │
       │  POST /UserService/GetUser│
       │  Body: Protobuf binary    │
       │─────────────────────────>│
       │                           │
       │  Headers:                  │
       │  Content-Type:             │
       │  application/grpc         │
       │                           │
       │  Response: Protobuf        │
       │<─────────────────────────│
       │                           │
```

### RPC vs REST — Call Model

```
    RPC:                          REST:
    userService.GetUser(123)      GET /users/123
         │                             │
         │  Single round-trip           │  Single round-trip
         │  Binary payload              │  JSON/XML
         │  Strongly typed              │  Schema in docs
         ▼                             ▼
    User object                    JSON → parse
```

---

## 4. Core Mechanics

### Components

| Component | Role |
|-----------|------|
| **IDL** | Define service and messages (`.proto`, `.thrift`) |
| **Stubs** | Generated client/server code |
| **Marshaling** | Serialize request/response (Protobuf, Thrift binary) |
| **Transport** | HTTP/2 (gRPC), TCP (Thrift), etc. |

### Protocol Buffers (Protobuf)

- **Schema**: `.proto` files
- **Binary format**: Compact, fast
- **Backward compatible**: Add optional fields; don't remove
- **Codegen**: `protoc` generates for 10+ languages

### gRPC Features

- **HTTP/2**: Multiplexing, streaming
- **Streaming**: Unary, server stream, client stream, bidirectional
- **Load balancing**: Client-side; use with service mesh
- **Metadata**: Custom headers (auth, tracing)

### Thrift (Apache)

- **IDL**: `.thrift` files
- **Transport**: TCP, HTTP
- **Serialization**: Binary, JSON, Compact
- **Used by**: Facebook, Evernote, Pinterest

---

## 5. Numbers

| Metric | Typical Value |
|--------|---------------|
| Protobuf vs JSON size | 3-10x smaller |
| Protobuf vs JSON parse | 5-10x faster |
| LinkedIn latency reduction | 60% with Protobuf |
| gRPC overhead | ~1ms (vs REST ~2-5ms) |
| HTTP/2 multiplexing | Many streams over 1 connection |

---

## 6. Tradeoffs

### RPC vs REST

| Aspect | RPC | REST |
|--------|-----|------|
| **Model** | Procedure call | Resource + HTTP verbs |
| **Payload** | Binary (Protobuf) | JSON, XML |
| **Contract** | IDL, codegen | OpenAPI, docs |
| **Caching** | Harder | HTTP caching |
| **Browser** | Limited (gRPC-web) | Native |
| **Performance** | Lower latency | Higher (JSON) |

### RPC vs GraphQL

| Aspect | RPC | GraphQL |
|--------|-----|---------|
| **Model** | Procedure | Query language |
| **Overfetching** | N/A (fixed schema) | Client specifies fields |
| **Batching** | Manual | Built-in |
| **Streaming** | gRPC streams | Subscriptions |
| **Use case** | Service-to-service | Client-to-API |

### gRPC vs Thrift vs Cap'n Proto

| Aspect | gRPC | Thrift | Cap'n Proto |
|--------|------|--------|-------------|
| **Transport** | HTTP/2 | TCP, HTTP | Custom |
| **IDL** | Protobuf | Thrift | Cap'n Proto |
| **Zero-copy** | No | No | Yes |
| **Streaming** | Yes | Limited | Yes |
| **Ecosystem** | Large | Mature | Smaller |

---

## 7. Variants / Implementations

### gRPC

- **Google**: Open-source, Protobuf, HTTP/2
- **Languages**: 10+ official
- **gRPC-Web**: Browser support via proxy

### Thrift

- **Apache**: Multi-language, multiple transports
- **Facebook**: Origin; used widely

### Cap'n Proto

- **Kenton Varda**: Zero-copy, schema evolution
- **Faster** than Protobuf in some cases

### JSON-RPC

- **Simple**: JSON over HTTP
- **No IDL**: Ad-hoc; less type safety

---

## 8. Scaling Strategies

1. **Load balancing**: Client-side (pick from list) or server-side (proxy)
2. **Connection pooling**: Reuse HTTP/2 connections
3. **Streaming**: Reduce round-trips for bulk data
4. **Service mesh**: mTLS, retries, observability (Istio, Linkerd)

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|------------|
| Service down | Call fails | Retries, circuit breaker |
| Timeout | Hung call | Set deadlines |
| Schema mismatch | Deserialize error | Version IDL; compatibility |
| Network partition | Unreachable | Timeout, fail fast |

---

## 10. Performance Considerations

- **Binary serialization**: Protobuf faster than JSON
- **HTTP/2**: Multiplexing; single connection for many calls
- **Connection reuse**: Avoid per-request TCP handshake
- **Streaming**: For large payloads; avoid loading all in memory

---

## 11. Use Cases

| Use Case | Why RPC |
|----------|---------|
| Microservices | Low latency, type safety |
| Real-time | Streaming (gRPC) |
| Internal APIs | Performance over REST |
| Mobile backends | Efficiency (binary) |

---

## 12. Comparison Tables

### RPC vs REST vs GraphQL

| Aspect | RPC | REST | GraphQL |
|--------|-----|------|---------|
| **Paradigm** | Procedure | Resource | Query |
| **Payload** | Binary | JSON | JSON |
| **Latency** | Low | Medium | Medium |
| **Caching** | Hard | Easy | Complex |
| **Browser** | gRPC-web | Native | Native |
| **Service-to-service** | Best | Good | Less common |

### IDL Comparison

| Feature | Protobuf | Thrift | OpenAPI |
|---------|----------|--------|---------|
| **Format** | Binary | Binary/JSON | JSON (REST) |
| **Schema evolution** | Optional fields | Similar | Versioning |
| **Codegen** | Yes | Yes | Yes |
| **Streaming** | Yes | Limited | No |

---

## 13. Code / Pseudocode

### Protobuf IDL

```protobuf
syntax = "proto3";

service UserService {
  rpc GetUser(GetUserRequest) returns (User);
  rpc ListUsers(ListUsersRequest) returns (stream User);
}

message GetUserRequest {
  string user_id = 1;
}

message User {
  string id = 1;
  string name = 2;
  string email = 3;
}
```

### gRPC Client (Python)

```python
import grpc
from user_pb2 import GetUserRequest
from user_pb2_grpc import UserServiceStub

channel = grpc.insecure_channel('localhost:50051')
stub = UserServiceStub(channel)

request = GetUserRequest(user_id='123')
response = stub.GetUser(request)
print(response.name)
```

### gRPC Server (Python)

```python
from concurrent import futures
import grpc
from user_pb2_grpc import add_UserServiceServicer_to_server, UserServiceServicer

class UserServiceImpl(UserServiceServicer):
    def GetUser(self, request, context):
        user = fetch_user(request.user_id)
        return User(id=user.id, name=user.name, email=user.email)

server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
add_UserServiceServicer_to_server(UserServiceImpl(), server)
server.add_insecure_port('[::]:50051')
server.start()
```

### Marshaling (Conceptual)

```python
def marshal_request(method, args):
    # Serialize to Protobuf binary
    request = build_protobuf_message(method, args)
    return request.SerializeToString()

def unmarshal_response(data, response_type):
    response = response_type()
    response.ParseFromString(data)
    return response
```

---

## 14. Interview Discussion

### Key Points

1. **RPC = remote call looks local** — stubs, marshaling, transport
2. **gRPC + Protobuf** — HTTP/2, binary, streaming
3. **LinkedIn: 60% latency reduction** with Protobuf
4. **RPC vs REST**: RPC for performance, REST for broad compatibility

### Common Questions

- **"How does RPC work?"** — Client calls stub; stub marshals request, sends over network. Server stub unmarshals, calls real function, marshals response, sends back.
- **"RPC vs REST?"** — RPC: procedure call, binary, lower latency. REST: resource-based, JSON, caching, browser-friendly.
- **"Why Protobuf over JSON?"** — Smaller payload, faster parse, schema evolution, type safety.
- **"gRPC streaming?"** — Unary, server stream, client stream, bidirectional. Good for real-time, bulk data.

### Red Flags

- Using RPC for public browser APIs (REST/GraphQL better)
- Ignoring schema evolution (breaking changes)
- No timeouts/deadlines (hung calls)
