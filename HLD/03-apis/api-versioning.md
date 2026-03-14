# API Versioning

> Staff+ Engineer Level — FAANG Interview Deep Dive

---

## 1. Concept Overview

### Definition

**API Versioning** is the practice of managing multiple versions of an API simultaneously to support backward compatibility while allowing evolution. Clients can target specific versions, and providers can deprecate old versions gracefully.

### Purpose

- **Backward compatibility**: Old clients continue working when API evolves
- **Controlled evolution**: Introduce breaking changes in new versions
- **Deprecation path**: Sunset old versions with advance notice
- **Client choice**: Clients opt into new features/versions

### Problems Solved

| Problem | Solution |
|---------|----------|
| Breaking changes | New version; old version remains |
| Client migration | Overlap period; gradual migration |
| Multiple client types | Each uses appropriate version |
| Deprecation | Version lifecycle; sunset policy |

---

## 2. Real-World Motivation

### Stripe

- **Header versioning**: `Stripe-Version: 2023-10-16` (date-based)
- Lock clients to specific API version
- New versions released regularly; old versions supported for years

### Twitter (X) API

- **URL versioning**: `/1.1/`, `/2/` (e.g., `api.twitter.com/2/tweets`)
- v1.1 legacy; v2 modern
- Clear migration path documented

### GitHub

- **Accept header**: `Accept: application/vnd.github.v3+json`
- **URL**: `/api/v3/` for REST
- GraphQL has versioning via schema

### Google APIs

- **URL versioning**: `v1`, `v2` in path (e.g., `youtube/v3`)
- Version in discovery document

### AWS

- **URL versioning**: `ec2.amazonaws.com` with API version in request
- **Header**: `x-amz-api-version` for some services

---

## 3. Architecture Diagrams

### URL Versioning — Version in Path

```
    Client Request
         │
         │  GET /v1/users/123
         │  GET /v2/users/123
         ▼
    ┌─────────────────────────────────────────┐
    │           API Gateway / Router          │
    │  Path: /v1/* → Service v1               │
    │  Path: /v2/* → Service v2               │
    └─────────────────────────────────────────┘
         │                    │
         ▼                    ▼
    ┌─────────┐          ┌─────────┐
    │  v1     │          │  v2     │
    │ Service │          │ Service │
    └─────────┘          └─────────┘
```

### Header Versioning — Version in Request Header

```
    Client Request
         │
         │  GET /users/123
         │  Stripe-Version: 2023-10-16
         │  X-API-Version: 2
         ▼
    ┌─────────────────────────────────────────┐
    │           API Gateway                    │
    │  Route by header → appropriate backend  │
    └─────────────────────────────────────────┘
         │
         ▼
    Backend selects handler based on version
```

### Query Parameter Versioning

```
    GET /users/123?version=2
    GET /users/123?api_version=v2
         │
         ▼
    Router/Gateway parses query param
    Routes to v2 handler
```

### Content Negotiation (Accept Header)

```
    GET /users/123
    Accept: application/vnd.myapi.v2+json
         │
         ▼
    Server returns representation for v2
    Content-Type: application/vnd.myapi.v2+json
```

### Version Routing Flow

```
                    ┌──────────────────┐
                    │  Incoming Request │
                    └────────┬─────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
              ▼              ▼              ▼
        Check path    Check header   Check query
        /v1/, /v2/    X-API-Version  ?version=
              │              │              │
              └──────────────┼──────────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │  Version Resolver │
                    │  Default: latest │
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │  Route to Handler │
                    │  v1 / v2 / v3    │
                    └──────────────────┘
```

---

## 4. Core Mechanics

### Versioning Strategies

| Strategy | Example | Pros | Cons |
|----------|---------|------|------|
| **URL path** | `/v1/users` | Simple, cacheable, visible | URL pollution |
| **Header** | `X-API-Version: 2` | Clean URLs | Less discoverable |
| **Query param** | `?version=2` | Easy to add | Caching issues, optional |
| **Accept header** | `Accept: application/vnd.api.v2+json` | RESTful, content negotiation | Complex |
| **Custom header** | `Stripe-Version: 2023-10-16` | Flexible, date-based | Client must send |

### Semantic Versioning for APIs

- **MAJOR**: Breaking changes (new version)
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible
- Example: `v2.1.3` — major=2, minor=1, patch=3

### Backward Compatibility Rules

- **Additive**: New fields, new endpoints — OK
- **Breaking**: Remove fields, change types, rename — new version
- **Optional params**: Adding optional params — OK
- **Required params**: Adding required params — breaking

---

## 5. Numbers

| Metric | Typical Value |
|--------|---------------|
| Stripe version support | Years (old versions maintained) |
| Deprecation notice | 6-12 months common |
| GitHub API v3 | Supported alongside v4 (GraphQL) |
| Twitter v1.1 → v2 | Multi-year migration |

---

## 6. Tradeoffs

### URL vs Header Versioning

| Aspect | URL | Header |
|--------|-----|--------|
| **Visibility** | Clear in URL | Hidden |
| **Caching** | Per-URL (good) | Same URL, different versions (careful) |
| **Client default** | Must specify in path | May forget header |
| **Tooling** | Easy (curl, browser) | Need to set header |

### Version Granularity

| Approach | Example | When |
|----------|---------|------|
| **Global** | Entire API is v1 or v2 | Simpler; all-or-nothing |
| **Resource-level** | `/v1/users`, `/v2/orders` | Fine-grained; complex |
| **Date-based** | `2023-10-16` (Stripe) | Rolling versions |

---

## 7. Variants / Implementations

### Stripe (Date-based)

- `Stripe-Version: 2023-10-16`
- Each release is a version; clients pin to date
- Changelog per version

### GitHub (Accept + URL)

- REST: `/api/v3/` or `Accept: application/vnd.github.v3+json`
- GraphQL: Single endpoint; schema evolution

### Twitter (URL)

- `/1.1/` and `/2/` in path
- Clear v1.1 vs v2 migration

### Google (URL)

- `youtube/v3`, `drive/v3`
- Per-API versioning

---

## 8. Scaling Strategies

1. **Shared backend**: Single codebase; version routing to handlers
2. **Separate services**: Each version = separate deployment
3. **Adapter layer**: New version adapts to old internal model
4. **Feature flags**: Gradual rollout of new version behavior

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|------------|
| Client omits version | Undefined behavior | Default to latest or reject |
| Deprecated version used | May break suddenly | Sunset policy; warnings |
| Version header wrong | 400 or wrong response | Validate; return clear error |
| Cache key collision | Wrong version cached | Include version in cache key |

---

## 10. Performance Considerations

- **Routing overhead**: Minimal for path/header parsing
- **Code duplication**: Multiple versions can bloat codebase — use adapters
- **Cache keys**: Must include version to avoid serving wrong version

---

## 11. Use Cases

| Use Case | Versioning Approach |
|----------|---------------------|
| Public API (Stripe) | Header, date-based |
| Internal microservices | URL or header |
| Mobile app backends | URL; app ships with version |
| Partner integrations | Contract versioning; long support |

---

## 12. Comparison Tables

### Versioning Methods

| Method | Example | RESTful? | Cacheable? |
|--------|---------|----------|------------|
| **URL path** | `/v1/users` | Debatable | Yes |
| **Query param** | `/users?v=1` | Yes | Careful |
| **Header** | `X-API-Version: 1` | Yes | Same URL |
| **Accept** | `Accept: ...v1+json` | Yes | Same URL |
| **Media type** | `application/vnd.api.v1+json` | Yes | Same URL |

### Company Practices

| Company | Method | Format |
|---------|--------|--------|
| Stripe | Header | Date (2023-10-16) |
| Twitter | URL | Integer (1.1, 2) |
| GitHub | URL + Accept | v3, v4 |
| Google | URL | v1, v2, v3 |
| AWS | URL/Header | Varies by service |

---

## 13. Code / Pseudocode

### URL Version Routing (Express-style)

```javascript
// Route by path prefix
app.use('/v1', v1Router);
app.use('/v2', v2Router);

// Or middleware to extract version
app.use((req, res, next) => {
  const match = req.path.match(/^\/v(\d+)\//);
  req.apiVersion = match ? parseInt(match[1]) : 1;
  next();
});
```

### Header Version Routing

```python
def get_api_version(request):
    return request.headers.get('X-API-Version', '1') or \
           request.headers.get('Stripe-Version') or \
           '1'

def route_request(request):
    version = get_api_version(request)
    handler = handlers.get(version, handlers['1'])
    return handler(request)
```

### Deprecation Response Headers

```python
def add_deprecation_headers(response, version, sunset_date):
    if version == 'v1':
        response['Deprecation'] = 'true'
        response['Sunset'] = sunset_date  # RFC 8594
        response['Link'] = '</api/v2/docs>; rel="alternate"'
    return response
```

### Semantic Version Parsing

```python
def parse_version(version_str):
    # "v2.1.3" or "2.1.3"
    match = re.match(r'v?(\d+)\.(\d+)\.(\d+)', version_str)
    if match:
        return (int(match[1]), int(match[2]), int(match[3]))
    return None

def is_compatible(requested, supported):
    # Same major = compatible
    return requested[0] == supported[0]
```

---

## 14. Interview Discussion

### Key Points

1. **Versioning enables evolution without breaking clients**
2. **URL vs header**: URL is visible and cache-friendly; header keeps URLs clean
3. **Backward compatibility**: Additive changes OK; breaking = new version
4. **Deprecation**: Sunset policy; headers (Deprecation, Sunset)

### Common Questions

- **"How would you version an API?"** — URL path (`/v1/`) is simplest. Header versioning (Stripe-style) for clean URLs. Consider caching and client adoption.
- **"URL vs header versioning?"** — URL: visible, cacheable, easy. Header: clean URLs, but clients must send it; caching needs version in key.
- **"How do you deprecate a version?"** — Announce timeline (6-12 months), add Deprecation/Sunset headers, provide migration guide, monitor usage.
- **"Semantic versioning for APIs?"** — Major = breaking, minor = additive, patch = fix. Not all companies use it (Stripe uses dates).

### Red Flags

- No versioning strategy (breaking changes break all clients)
- Too many versions (maintenance burden)
- No deprecation policy (surprise shutdowns)
