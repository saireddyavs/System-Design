# API Design — Interview Preparation

Comprehensive, hands-on guide for API design interviews — from REST endpoint design and HTTP semantics to authentication, error handling, rate limiting, and writing real API contracts.

---

## Contents

| # | Topic | Description | Link |
|---|-------|-------------|------|
| 01 | API Design Process | How to approach any API design question: resource modeling, URL design, method selection | [Guide](01-api-design-process.md) |
| 02 | REST API Deep Dive | HTTP methods, status codes, headers, content negotiation, HATEOAS, idempotency | [Guide](02-rest-api-deep-dive.md) |
| 03 | Request & Response Patterns | Pagination, filtering, sorting, partial responses, bulk operations, async APIs | [Guide](03-request-response-patterns.md) |
| 04 | Authentication & Security | OAuth 2.0, JWT, API keys, RBAC, CORS, rate limiting, input validation | [Guide](04-authentication-and-security.md) |
| 05 | Error Handling & Versioning | RFC 7807, error taxonomy, versioning strategies, deprecation, backward compatibility | [Guide](05-error-handling-and-versioning.md) |
| 06 | GraphQL & gRPC Essentials | Schema design, resolvers, mutations, Protobuf, streaming, when to use which | [Guide](06-graphql-and-grpc.md) |
| 07 | API Interview Questions | 40+ interview questions with complete API designs across real-world scenarios | [Guide](07-api-interview-questions.md) |

---

## How to Use This Guide

1. **Start with [01-api-design-process](01-api-design-process.md)** — learn the systematic approach
2. **Master REST** in [02](02-rest-api-deep-dive.md) and [03](03-request-response-patterns.md)
3. **Understand security** in [04](04-authentication-and-security.md)
4. **Handle edge cases** with [05](05-error-handling-and-versioning.md)
5. **Know alternatives** in [06](06-graphql-and-grpc.md)
6. **Practice** with [07](07-api-interview-questions.md) — 40+ real interview scenarios

---

## Quick Reference: API Design Checklist

```
API Design Checklist (use in interviews)
├── Resource Modeling
│   ├── Identify resources (nouns, not verbs)
│   ├── Define relationships (1:1, 1:N, M:N)
│   └── Design URL hierarchy (max 2 levels deep)
├── Endpoints
│   ├── Use correct HTTP methods (GET, POST, PUT, PATCH, DELETE)
│   ├── Use proper status codes (201 Created, 404 Not Found, etc.)
│   ├── Make POST/PUT idempotent where possible
│   └── Support filtering, sorting, pagination
├── Request/Response
│   ├── Consistent JSON envelope
│   ├── Pagination metadata (cursor-based for large datasets)
│   ├── Partial responses (fields parameter)
│   └── Proper Content-Type headers
├── Error Handling
│   ├── Consistent error format (RFC 7807)
│   ├── Meaningful error messages
│   ├── Validation error details
│   └── Rate limit headers
├── Security
│   ├── Authentication (OAuth 2.0 / JWT / API Key)
│   ├── Authorization (RBAC / ABAC)
│   ├── Input validation & sanitization
│   ├── HTTPS everywhere
│   └── CORS policy
├── Versioning
│   ├── Strategy (URI / Header / Query param)
│   ├── Deprecation policy
│   └── Backward compatibility
└── Non-Functional
    ├── Rate limiting
    ├── Caching (ETags, Cache-Control)
    ├── Compression (gzip)
    ├── Request timeouts
    └── Documentation (OpenAPI/Swagger)
```
