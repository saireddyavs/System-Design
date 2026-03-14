# Must-Know HTTP Headers

> Staff+ Engineer Level — FAANG Interview Deep Dive

---

## 1. Concept Overview

### Definition

**HTTP Headers** are key-value pairs sent with HTTP requests and responses. They convey metadata about the request/response, control caching, enable CORS, support authentication, and influence security behavior.

### Purpose

- **Request context**: What the client wants (Accept, Content-Type)
- **Authentication**: Credentials (Authorization, Cookie)
- **Caching**: Control cache behavior (Cache-Control, ETag)
- **Security**: Mitigate attacks (CSP, HSTS, X-Frame-Options)
- **CORS**: Cross-origin resource sharing

### Problems Solved

| Problem | Headers |
|---------|---------|
| Content negotiation | Accept, Content-Type |
| Auth | Authorization, Cookie, Set-Cookie |
| Caching | Cache-Control, ETag, If-None-Match |
| Security | CSP, HSTS, X-Frame-Options |
| Cross-origin | Access-Control-* |

---

## 2. Real-World Motivation

### CDNs (CloudFront, Cloudflare)

- **Cache-Control**: TTL, stale-while-revalidate
- **ETag, If-None-Match**: Conditional requests; 304 Not Modified
- **X-Cache**: Hit/Miss from edge

### API Gateways (Kong, AWS)

- **Authorization**: Bearer token, API key
- **X-Request-ID**: Request tracing
- **Rate limit headers**: X-RateLimit-Remaining, Retry-After

### Security (OWASP)

- **CSP**: Restrict script sources
- **HSTS**: Force HTTPS
- **X-Frame-Options**: Clickjacking protection

---

## 3. Architecture Diagrams

### Request Headers Flow

```
    Client Request
         │
         │  GET /api/users/123
         │  ┌─────────────────────────────────────────┐
         │  │ Accept: application/json                │
         │  │ Authorization: Bearer <token>            │
         │  │ Content-Type: application/json           │
         │  │ Cache-Control: no-cache                  │
         │  │ User-Agent: MyApp/1.0                    │
         │  │ If-None-Match: "abc123"                  │
         │  └─────────────────────────────────────────┘
         ▼
    Server / API Gateway
```

### Response Headers Flow

```
    Server Response
         │
         │  200 OK
         │  ┌─────────────────────────────────────────┐
         │  │ Content-Type: application/json           │
         │  │ Cache-Control: max-age=3600              │
         │  │ ETag: "abc123"                           │
         │  │ Set-Cookie: session=xyz; HttpOnly        │
         │  │ X-Request-ID: req-123                    │
         │  │ Access-Control-Allow-Origin: *           │
         │  └─────────────────────────────────────────┘
         ▼
    Client / Browser
```

### Caching Flow with ETag

```
    Client                    Server
       │                         │
       │  GET /resource          │
       │  If-None-Match: "v1"   │
       │───────────────────────>│
       │                         │
       │  304 Not Modified       │
       │  ETag: "v1"             │  (unchanged; use cache)
       │<───────────────────────│
       │                         │
    Use cached copy              │
```

### CORS Preflight Flow

```
    Browser                    Server
       │                         │
       │  OPTIONS /api/data     │
       │  Origin: https://app.com
       │  Access-Control-Request-Method: POST
       │───────────────────────>│
       │                         │
       │  200 OK                 │
       │  Access-Control-Allow-Origin: https://app.com
       │  Access-Control-Allow-Methods: GET, POST
       │  Access-Control-Max-Age: 86400
       │<───────────────────────│
       │                         │
       │  POST /api/data         │
       │  (actual request)       │
       │───────────────────────>│
```

---

## 4. Core Mechanics

### Request Headers

| Header | Purpose | Example |
|--------|---------|---------|
| **Accept** | Response format | `application/json`, `text/html` |
| **Authorization** | Credentials | `Bearer <token>`, `Basic base64` |
| **Content-Type** | Request body type | `application/json` |
| **Cache-Control** | Cache directives | `no-cache`, `max-age=0` |
| **User-Agent** | Client identifier | `Mozilla/5.0...` |
| **If-None-Match** | Conditional GET (ETag) | `"abc123"` |
| **If-Modified-Since** | Conditional GET (date) | `Wed, 21 Oct 2024 07:28:00 GMT` |
| **Origin** | CORS; request origin | `https://app.example.com` |

### Response Headers

| Header | Purpose | Example |
|--------|---------|---------|
| **Content-Type** | Response format | `application/json` |
| **Cache-Control** | Cache directives | `max-age=3600, public` |
| **ETag** | Resource version hash | `"abc123"` |
| **Last-Modified** | Resource change time | `Wed, 21 Oct 2024 07:28:00 GMT` |
| **Set-Cookie** | Set cookie | `session=xyz; HttpOnly; Secure` |
| **X-Request-ID** | Request tracing | `req-abc-123` |
| **Location** | Redirect target | `https://new-url.com` |

### Security Headers

| Header | Purpose | Example |
|--------|---------|---------|
| **Content-Security-Policy** | Restrict script/source | `default-src 'self'` |
| **X-Frame-Options** | Clickjacking | `DENY`, `SAMEORIGIN` |
| **Strict-Transport-Security** | Force HTTPS | `max-age=31536000` |
| **X-Content-Type-Options** | Prevent MIME sniffing | `nosniff` |
| **X-XSS-Protection** | Legacy XSS filter | `1; mode=block` |

### CORS Headers

| Header | Direction | Purpose |
|--------|-----------|---------|
| **Access-Control-Allow-Origin** | Response | Allowed origins |
| **Access-Control-Allow-Methods** | Response | Allowed methods |
| **Access-Control-Allow-Headers** | Response | Allowed headers |
| **Access-Control-Max-Age** | Response | Preflight cache |
| **Access-Control-Request-Method** | Request (preflight) | Method being used |
| **Access-Control-Request-Headers** | Request (preflight) | Headers being sent |

### Caching Headers

| Header | Purpose |
|--------|---------|
| **Cache-Control** | `max-age`, `no-cache`, `no-store`, `stale-while-revalidate` |
| **ETag** | Opaque version identifier |
| **If-None-Match** | Conditional GET; 304 if ETag matches |
| **If-Modified-Since** | Conditional GET; 304 if not modified |
| **Expires** | Legacy; absolute expiry time |

---

## 5. Numbers

| Header / Value | Typical |
|---------------|---------|
| Cache-Control max-age | 60 - 86400 (1 min - 24 hr) |
| HSTS max-age | 31536000 (1 year) |
| CORS preflight cache | 86400 (24 hr) |
| ETag length | 32-64 chars (hash) |

---

## 6. Tradeoffs

### Cache-Control Directives

| Directive | Effect |
|-----------|--------|
| **no-store** | Don't cache at all |
| **no-cache** | Revalidate before use |
| **max-age=N** | Fresh for N seconds |
| **stale-while-revalidate** | Serve stale; revalidate in background |
| **private** | Only browser cache |
| **public** | CDN can cache |

### CORS: * vs Specific Origin

| Value | Use Case | Risk |
|-------|----------|------|
| `*` | Public API, no credentials | Any origin can call |
| `https://app.com` | Specific app | Safer; credentials allowed |
| `null` | Same-origin only | Restrictive |

---

## 7. Variants / Implementations

### Cache-Control Combinations

- **Static assets**: `max-age=31536000, immutable`
- **API response**: `no-cache` or `max-age=60`
- **Private data**: `private, no-store`
- **CDN**: `public, max-age=3600, stale-while-revalidate=86400`

### Content-Type for APIs

- **JSON**: `application/json`
- **JSON with charset**: `application/json; charset=utf-8`
- **Custom**: `application/vnd.api+json`

### Security Header Presets

- **Strict**: CSP, HSTS, X-Frame-Options DENY, X-Content-Type-Options
- **Moderate**: HSTS, X-Frame-Options SAMEORIGIN
- **Minimal**: X-Content-Type-Options nosniff

---

## 8. Scaling Strategies

1. **CDN caching**: Cache-Control, ETag at edge
2. **Request ID**: X-Request-ID for distributed tracing
3. **Compression**: Content-Encoding: gzip, br
4. **Connection**: Keep-Alive for connection reuse

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|------------|
| Missing CORS | Browser blocks | Set Access-Control-Allow-Origin |
| Wrong Content-Type | Parse error | Set correctly |
| Cache-Control no-store on static | Poor performance | Use max-age for static |
| HSTS too short | Less protection | max-age >= 1 year |

---

## 10. Performance Considerations

- **Conditional requests**: If-None-Match → 304 saves bandwidth
- **Compression**: Content-Encoding: gzip
- **Connection reuse**: Keep-Alive
- **Preload**: Link header for critical resources

---

## 11. Use Cases

| Use Case | Key Headers |
|----------|-------------|
| **CDN** | Cache-Control, ETag, Vary |
| **API Gateway** | Authorization, X-Request-ID, Rate limit headers |
| **Auth** | Authorization, Set-Cookie |
| **CORS** | Access-Control-* |
| **Security** | CSP, HSTS, X-Frame-Options |

---

## 12. Comparison Tables

### Caching Strategies by Content Type

| Content | Cache-Control | ETag |
|---------|---------------|------|
| **Static JS/CSS** | max-age=31536000, immutable | Optional |
| **Images** | max-age=86400, public | Yes |
| **API (dynamic)** | no-cache or max-age=0 | Yes (for 304) |
| **Private** | private, no-store | No |

### Security Headers Checklist

| Header | Recommended |
|--------|-------------|
| **Content-Security-Policy** | Yes; restrict sources |
| **X-Frame-Options** | SAMEORIGIN or DENY |
| **Strict-Transport-Security** | max-age=31536000 |
| **X-Content-Type-Options** | nosniff |
| **Referrer-Policy** | strict-origin-when-cross-origin |

---

## 13. Code / Pseudocode

### Setting Security Headers (Middleware)

```python
def security_headers_middleware(response):
    response['Strict-Transport-Security'] = 'max-age=31536000; includeSubDomains'
    response['X-Frame-Options'] = 'SAMEORIGIN'
    response['X-Content-Type-Options'] = 'nosniff'
    response['Content-Security-Policy'] = "default-src 'self'; script-src 'self'"
    return response
```

### Conditional GET with ETag

```python
def get_user(user_id):
    user = db.get_user(user_id)
    etag = hashlib.md5(json.dumps(user).encode()).hexdigest()
    if request.headers.get('If-None-Match') == etag:
        return Response(status=304, headers={'ETag': etag})
    return Response(json.dumps(user), headers={
        'ETag': etag,
        'Cache-Control': 'private, max-age=60'
    })
```

### CORS Configuration

```python
ALLOWED_ORIGINS = ['https://app.example.com']

def cors_headers(origin):
    if origin in ALLOWED_ORIGINS:
        return {
            'Access-Control-Allow-Origin': origin,
            'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
            'Access-Control-Allow-Headers': 'Authorization, Content-Type',
            'Access-Control-Max-Age': '86400'
        }
    return {}
```

### Cache-Control by Resource Type

```python
def get_cache_control(path):
    if path.startswith('/static/'):
        return 'public, max-age=31536000, immutable'
    if path.startswith('/api/'):
        return 'private, no-cache'
    return 'public, max-age=3600'
```

---

## 14. Interview Discussion

### Key Points

1. **Request**: Accept, Authorization, Content-Type, If-None-Match
2. **Response**: Cache-Control, ETag, Set-Cookie, X-Request-ID
3. **Security**: CSP, HSTS, X-Frame-Options, X-Content-Type-Options
4. **CORS**: Access-Control-Allow-Origin, preflight (OPTIONS)

### Common Questions

- **"What headers control caching?"** — Cache-Control (max-age, no-cache), ETag, If-None-Match for conditional GET.
- **"How does CORS work?"** — Browser sends Origin; server responds with Access-Control-Allow-Origin. Preflight (OPTIONS) for non-simple requests.
- **"ETag vs Last-Modified?"** — ETag: any change; hash. Last-Modified: timestamp; less precise for sub-second changes.
- **"Security headers?"** — CSP (script sources), HSTS (force HTTPS), X-Frame-Options (clickjacking), X-Content-Type-Options (MIME sniffing).
- **"X-Request-ID?"** — Trace request across services; correlation ID for logging/debugging.

### Red Flags

- `Access-Control-Allow-Origin: *` with credentials
- No Cache-Control on static assets
- Missing security headers (CSP, HSTS)
- Trusting client-supplied Content-Type
