# API Security Best Practices

> Staff+ Engineer Level — FAANG Interview Deep Dive

---

## 1. Concept Overview

### Definition

**API Security** encompasses the practices, mechanisms, and controls used to protect APIs from unauthorized access, abuse, and attacks. It includes authentication, authorization, rate limiting, input validation, and protection against common vulnerabilities.

### Purpose

- **Authentication**: Verify who is making the request
- **Authorization**: Verify what they are allowed to do
- **Protection**: Prevent injection, abuse, data exposure
- **Compliance**: Meet security standards (OWASP, PCI-DSS)

### Problems Solved

| Problem | Solution |
|---------|----------|
| Unauthorized access | Authentication (OAuth, JWT, API keys) |
| Abuse / DoS | Rate limiting, throttling |
| Injection attacks | Input validation, parameterized queries |
| Data leakage | Least privilege, response filtering |
| Cross-site attacks | CORS, CSRF tokens |

---

## 2. Real-World Motivation

### OWASP API Security Top 10 (2023)

1. Broken Object Level Authorization (BOLA)
2. Broken Authentication
3. Broken Object Property Level Authorization
4. Unrestricted Resource Consumption
5. Broken Function Level Authorization
6. Unrestricted Access to Sensitive Business Flows
7. Server Side Request Forgery (SSRF)
8. Security Misconfiguration
9. Improper Inventory Management
10. Unsafe Consumption of APIs

### Industry Practices

- **Stripe**: API keys, OAuth for apps, webhooks with signatures
- **Twilio**: API key + secret, request signing
- **AWS**: IAM, signature v4, API keys
- **GitHub**: Tokens, OAuth apps, fine-grained permissions

---

## 3. Architecture Diagrams

### OAuth 2.0 Authorization Code Flow

```
    User          Client App       Auth Server       API Server
      │                │                 │                 │
      │  1. Login      │                 │                 │
      │──────────────>│                 │                 │
      │                │  2. Redirect to auth               │
      │                │────────────────>│                 │
      │                │                 │  3. User consents│
      │                │<────────────────│                 │
      │                │  4. Auth code    │                 │
      │                │<────────────────│                 │
      │                │  5. Exchange code for token        │
      │                │────────────────>│                 │
      │                │  6. Access + Refresh token         │
      │                │<────────────────│                 │
      │                │  7. API call + Bearer token        │
      │                │──────────────────────────────────>│
      │                │  8. Response    │                 │
      │                │<──────────────────────────────────│
```

### JWT Token Validation Flow

```
    Client                    API Gateway / Service
       │                               │
       │  GET /api/users               │
       │  Authorization: Bearer <JWT>  │
       │─────────────────────────────>│
       │                               │
       │                       ┌───────┴───────┐
       │                       │ 1. Extract JWT │
       │                       │ 2. Verify sig  │
       │                       │    (public key)│
       │                       │ 3. Check exp   │
       │                       │ 4. Extract     │
       │                       │    user/claims │
       │                       └───────┬───────┘
       │                               │
       │                       Valid?  │
       │                       ├─Yes──> Forward to backend
       │                       └─No──> 401 Unauthorized
       │                               │
```

### Rate Limiting Architecture

```
    Request
       │
       ▼
    ┌─────────────────┐
    │  Rate Limiter   │
    │  (per key: IP,  │
    │   user, API key)│
    └────────┬────────┘
             │
             │  Check: Redis / in-memory counter
             │  Limit: 100 req/min per key
             │
       ┌─────┴─────┐
       │           │
    Under limit  Over limit
       │           │
       ▼           ▼
    Forward    429 Too Many Requests
    to API     Retry-After: 60
```

### CORS Preflight

```
    Browser                Server
       │                      │
       │  OPTIONS /api/data   │
       │  Origin: https://app.example.com
       │  Access-Control-Request-Method: POST
       │  Access-Control-Request-Headers: X-Custom
       │─────────────────────>│
       │                      │
       │  200 OK              │
       │  Access-Control-Allow-Origin: https://app.example.com
       │  Access-Control-Allow-Methods: GET, POST
       │  Access-Control-Allow-Headers: X-Custom
       │<─────────────────────│
       │                      │
       │  POST /api/data      │
       │  (actual request)    │
       │─────────────────────>│
```

---

## 4. Core Mechanics

### Authentication Methods

| Method | Use Case | Pros | Cons |
|--------|----------|------|------|
| **API Key** | Server-to-server, scripts | Simple | Key leakage risk |
| **OAuth 2.0** | Third-party apps, user delegation | Standard, flexible | Complex |
| **JWT** | Stateless auth, microservices | No server session | Revocation hard |
| **mTLS** | Service-to-service | Strong, no token | Cert management |

### Rate Limiting Strategies

| Strategy | Description | Example |
|----------|-------------|---------|
| **Fixed window** | Count in fixed time window | 100 req/min |
| **Sliding window** | Rolling window | More accurate |
| **Token bucket** | Tokens refill at rate | Burst + sustained |
| **Leaky bucket** | Process at fixed rate | Smooth traffic |

### Input Validation

- **Allowlist**: Only allow known-good characters
- **Schema validation**: JSON schema, OpenAPI
- **Parameterized queries**: Prevent SQL injection
- **Size limits**: Max body, max field length
- **Type checking**: Enforce expected types

### Security Headers

| Header | Purpose |
|--------|---------|
| **CSP** | Content Security Policy — restrict script sources |
| **X-Frame-Options** | Prevent clickjacking |
| **HSTS** | Force HTTPS |
| **X-Content-Type-Options** | Prevent MIME sniffing |
| **X-XSS-Protection** | Legacy XSS filter |

---

## 5. Numbers

| Metric | Typical Value |
|--------|---------------|
| JWT expiry | 15 min - 1 hour (access), 7-30 days (refresh) |
| Rate limit (API) | 100-10K req/min per key |
| API key entropy | 32+ bytes (256 bits) |
| bcrypt cost | 10-12 |
| Token size (JWT) | 200-500 bytes |

---

## 6. Tradeoffs

### API Key vs OAuth vs JWT

| Aspect | API Key | OAuth | JWT |
|--------|---------|-------|-----|
| **Complexity** | Low | High | Medium |
| **User context** | No | Yes | Yes |
| **Revocation** | Revoke key | Revoke token | Hard (until exp) |
| **Stateless** | Yes | Depends | Yes |

### Rate Limit Granularity

| Level | Example | Pros | Cons |
|-------|---------|------|------|
| **Global** | 10K req/s total | Simple | Unfair distribution |
| **Per IP** | 100 req/min per IP | Fair | VPN/proxy bypass |
| **Per user** | 1000 req/min per user | Fair | Need auth first |
| **Per endpoint** | Different limits | Flexible | Complex |

---

## 7. Variants / Implementations

### OAuth 2.0 Flows

- **Authorization Code**: Web apps (with PKCE for SPAs)
- **Client Credentials**: Server-to-server
- **Implicit**: Deprecated
- **Resource Owner Password**: Legacy; avoid

### JWT Best Practices

- Short expiry for access tokens
- Use refresh tokens for long sessions
- Store sensitive data server-side; JWT for identity only
- RS256 (asymmetric) for multi-party validation

### API Key Storage

- **Client**: Environment variable, secrets manager
- **Server**: Hashed (like passwords); never log
- **Rotation**: Support multiple keys during rotation

---

## 8. Scaling Strategies

1. **Distributed rate limiting**: Redis, consistent hashing
2. **Edge rate limiting**: CDN, API gateway (Kong, AWS)
3. **JWT validation**: Cache public keys; validate at edge
4. **OAuth**: Central auth service; scale independently

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|------------|
| Key/token leak | Unauthorized access | Rotate, revoke; short expiry |
| Rate limit bypass | DoS, abuse | Multiple layers; per-user limits |
| SQL injection | Data breach | Parameterized queries; ORM |
| XSS | Session hijack | Escape output; CSP |
| CORS misconfig | Data theft | Whitelist origins; no wildcard |

---

## 10. Performance Considerations

- **JWT validation**: Local (public key); no DB call
- **Rate limit check**: Redis; sub-ms
- **OAuth**: Token endpoint can be bottleneck; cache tokens
- **Input validation**: Fail fast; before heavy processing

---

## 11. Use Cases

| Use Case | Security Approach |
|----------|-------------------|
| Public API | API key, rate limit, OAuth for user context |
| Internal API | mTLS, service account |
| Mobile app | OAuth + PKCE, certificate pinning |
| Partner API | API key, IP allowlist, contract |

---

## 12. Comparison Tables

### Authentication Methods

| Method | Stateless | Revocable | User Context |
|--------|-----------|-----------|--------------|
| **API Key** | Yes | Yes | No |
| **OAuth** | Optional | Yes | Yes |
| **JWT** | Yes | No (until exp) | Yes |
| **Session cookie** | No | Yes | Yes |
| **mTLS** | Yes | Revoke cert | Via cert |

### OWASP Top 10 Mitigations

| Risk | Mitigation |
|------|------------|
| BOLA | Check resource ownership per request |
| Broken Auth | Strong tokens, secure storage |
| Broken Function Auth | RBAC, least privilege |
| Resource Consumption | Rate limiting |
| SSRF | Validate URLs; block internal IPs |
| Misconfiguration | Security headers; disable debug |

---

## 13. Code / Pseudocode

### JWT Validation (Pseudocode)

```python
def validate_jwt(token):
    parts = token.split('.')
    if len(parts) != 3:
        raise InvalidToken()
    header, payload, sig = parts
    # Verify signature with public key
    if not verify_signature(header + '.' + payload, sig, public_key):
        raise InvalidToken()
    claims = json.loads(base64_decode(payload))
    if claims['exp'] < time.time():
        raise TokenExpired()
    return claims
```

### Rate Limiter (Token Bucket)

```python
def rate_limit(key, limit=100, window=60):
    now = time.time()
    bucket = redis.get(f"ratelimit:{key}") or {"tokens": limit, "last": now}
    elapsed = now - bucket["last"]
    bucket["tokens"] = min(limit, bucket["tokens"] + elapsed * (limit / window))
    bucket["last"] = now
    if bucket["tokens"] >= 1:
        bucket["tokens"] -= 1
        redis.setex(f"ratelimit:{key}", window, bucket)
        return True
    return False
```

### Input Validation (SQL Injection Prevention)

```python
# BAD
query = f"SELECT * FROM users WHERE id = {user_id}"

# GOOD
cursor.execute("SELECT * FROM users WHERE id = %s", (user_id,))
```

### CORS Configuration

```python
# Allow specific origins only
ALLOWED_ORIGINS = ["https://app.example.com", "https://admin.example.com"]

def cors_middleware(request, response):
    origin = request.headers.get("Origin")
    if origin in ALLOWED_ORIGINS:
        response["Access-Control-Allow-Origin"] = origin
    response["Access-Control-Allow-Methods"] = "GET, POST, OPTIONS"
    response["Access-Control-Allow-Headers"] = "Authorization, Content-Type"
    return response
```

---

## 14. Interview Discussion

### Key Points

1. **Auth + Authz + Rate limit + Validation** — core pillars
2. **OAuth for delegation**, **JWT for stateless**, **API key for simple**
3. **OWASP API Top 10** — BOLA, broken auth, rate limiting
4. **Defense in depth** — multiple layers

### Common Questions

- **"How do you secure an API?"** — Auth (OAuth/JWT/API key), rate limiting, input validation, HTTPS, CORS, security headers.
- **"API key vs OAuth?"** — API key: simple, server-to-server. OAuth: user delegation, third-party apps.
- **"How does rate limiting work?"** — Token bucket or sliding window; key by IP/user; Redis for distributed.
- **"Prevent SQL injection?"** — Parameterized queries; never concatenate user input.
- **"JWT revocation?"** — Short expiry; refresh token (stored, revocable); or blocklist (Redis) for logout.

### Red Flags

- Storing secrets in code
- No rate limiting
- Wildcard CORS
- Trusting client input without validation
