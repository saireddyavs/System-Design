# Error Handling & Versioning

## RFC 7807, Error Taxonomy, Versioning Strategies, Deprecation, and Backward Compatibility

---

## 1. Error Response Design

### Consistent Error Format

Every API error should have the same structure — clients should never have to guess.

```json
// Standard error response
{
  "error": {
    "code": "VALIDATION_ERROR",           // machine-readable code
    "message": "Request validation failed", // human-readable message
    "status": 422,                         // HTTP status code
    "details": [                           // specific validation errors
      {
        "field": "email",
        "code": "INVALID_FORMAT",
        "message": "Must be a valid email address",
        "value": "not-an-email"
      },
      {
        "field": "age",
        "code": "OUT_OF_RANGE",
        "message": "Must be between 13 and 120",
        "value": 5
      }
    ],
    "request_id": "req_abc123",           // for support/debugging
    "documentation_url": "https://docs.example.com/errors/VALIDATION_ERROR"
  }
}
```

### RFC 7807 — Problem Details for HTTP APIs

```json
// RFC 7807 standard format (used by IETF, Microsoft, etc.)
{
  "type": "https://api.example.com/errors/validation-error",
  "title": "Validation Error",
  "status": 422,
  "detail": "The email field is not a valid email address",
  "instance": "/api/v1/users",
  "errors": [
    { "pointer": "/email", "detail": "Invalid email format" }
  ]
}
```

| Field | Required? | Purpose |
|-------|-----------|---------|
| `type` | Yes | URI reference identifying the error type |
| `title` | Yes | Short human-readable summary |
| `status` | Yes | HTTP status code |
| `detail` | No | Human-readable explanation specific to this occurrence |
| `instance` | No | URI reference identifying the specific occurrence |

---

## 2. Error Taxonomy

### Organize errors into categories with consistent codes:

```
Error Codes Taxonomy:
├── Authentication Errors (401)
│   ├── AUTH_TOKEN_MISSING          → No token provided
│   ├── AUTH_TOKEN_INVALID          → Token is malformed or signature invalid
│   ├── AUTH_TOKEN_EXPIRED          → Token has expired
│   └── AUTH_CREDENTIALS_INVALID    → Wrong username/password
│
├── Authorization Errors (403)
│   ├── FORBIDDEN                   → Insufficient permissions
│   ├── RESOURCE_ACCESS_DENIED      → Can't access this specific resource
│   └── ACCOUNT_SUSPENDED           → Account has been suspended
│
├── Validation Errors (400/422)
│   ├── VALIDATION_ERROR            → General validation failure
│   ├── MISSING_REQUIRED_FIELD      → Required field not provided
│   ├── INVALID_FORMAT              → Field format wrong (email, phone, etc.)
│   ├── OUT_OF_RANGE                → Value outside acceptable range
│   ├── INVALID_ENUM_VALUE          → Value not in allowed set
│   └── BODY_TOO_LARGE             → Request body exceeds limit
│
├── Resource Errors (404/409/410)
│   ├── NOT_FOUND                   → Resource doesn't exist
│   ├── ALREADY_EXISTS              → Duplicate (e.g., email already taken)
│   ├── CONFLICT                    → State conflict (e.g., order already shipped)
│   └── GONE                        → Resource permanently deleted
│
├── Rate Limiting (429)
│   └── RATE_LIMITED                → Too many requests
│
└── Server Errors (500/502/503)
    ├── INTERNAL_ERROR              → Unexpected server error
    ├── SERVICE_UNAVAILABLE         → Maintenance or overloaded
    ├── UPSTREAM_ERROR              → Dependency failed
    └── TIMEOUT                     → Request processing timed out
```

### Error Code to HTTP Status Mapping

| Error Code | HTTP Status | When |
|-----------|-------------|------|
| `VALIDATION_ERROR` | 400 or 422 | Invalid input |
| `AUTH_TOKEN_MISSING` | 401 | No authentication |
| `AUTH_TOKEN_EXPIRED` | 401 | Token expired |
| `FORBIDDEN` | 403 | Not authorized |
| `NOT_FOUND` | 404 | Resource doesn't exist |
| `ALREADY_EXISTS` | 409 | Duplicate resource |
| `CONFLICT` | 409 | State conflict |
| `RATE_LIMITED` | 429 | Too many requests |
| `INTERNAL_ERROR` | 500 | Server bug |
| `SERVICE_UNAVAILABLE` | 503 | Temporarily down |

---

## 3. Error Handling Best Practices

### Do's and Don'ts

| ✅ Do | ❌ Don't |
|-------|---------|
| Return specific error codes | Return generic "error occurred" |
| Include field-level validation details | Just say "validation failed" |
| Include `request_id` for debugging | Expect users to reproduce the issue |
| Log full error server-side | Expose stack traces to clients |
| Return documentation links | Leave developers guessing |
| Use correct HTTP status codes | Return 200 with `"success": false` |
| Handle edge cases (empty body, wrong content-type) | Crash on unexpected input |

### Error Response Examples

```json
// 400 Bad Request — malformed JSON
{
  "error": {
    "code": "INVALID_JSON",
    "message": "Request body is not valid JSON",
    "status": 400
  }
}

// 401 Unauthorized — expired token
{
  "error": {
    "code": "AUTH_TOKEN_EXPIRED",
    "message": "Your authentication token has expired. Please refresh your token.",
    "status": 401
  }
}

// 403 Forbidden — insufficient permissions
{
  "error": {
    "code": "FORBIDDEN",
    "message": "You don't have permission to delete this resource. Required role: admin.",
    "status": 403
  }
}

// 404 Not Found
{
  "error": {
    "code": "NOT_FOUND",
    "message": "User with ID 'user_999' not found",
    "status": 404
  }
}

// 409 Conflict — duplicate
{
  "error": {
    "code": "ALREADY_EXISTS",
    "message": "A user with email 'alice@example.com' already exists",
    "status": 409,
    "details": {
      "field": "email",
      "value": "alice@example.com"
    }
  }
}

// 422 Validation Error — multiple field errors
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "status": 422,
    "details": [
      { "field": "name", "code": "REQUIRED", "message": "Name is required" },
      { "field": "email", "code": "INVALID_FORMAT", "message": "Invalid email" },
      { "field": "price", "code": "OUT_OF_RANGE", "message": "Must be > 0" }
    ]
  }
}

// 429 Rate Limited
{
  "error": {
    "code": "RATE_LIMITED",
    "message": "Rate limit exceeded. Try again in 30 seconds.",
    "status": 429,
    "retry_after": 30
  }
}

// 500 Internal Error — never expose implementation details!
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "An unexpected error occurred. Please try again later.",
    "status": 500,
    "request_id": "req_abc123"
  }
}
// ❌ NEVER: { "error": "NullPointerException at UserService.java:42" }
```

---

## 4. Error Handling Implementation

```python
# Centralized error handler

class APIError(Exception):
    def __init__(self, code, message, status, details=None):
        self.code = code
        self.message = message
        self.status = status
        self.details = details

class NotFoundError(APIError):
    def __init__(self, resource_type, resource_id):
        super().__init__(
            code="NOT_FOUND",
            message=f"{resource_type} with ID '{resource_id}' not found",
            status=404
        )

class ValidationError(APIError):
    def __init__(self, field_errors):
        super().__init__(
            code="VALIDATION_ERROR",
            message="Request validation failed",
            status=422,
            details=field_errors
        )

class ConflictError(APIError):
    def __init__(self, message, field=None, value=None):
        super().__init__(
            code="ALREADY_EXISTS",
            message=message,
            status=409,
            details={"field": field, "value": value} if field else None
        )

# Error handler middleware
@app.errorhandler(APIError)
def handle_api_error(error):
    response = {
        "error": {
            "code": error.code,
            "message": error.message,
            "status": error.status,
            "request_id": g.request_id
        }
    }
    if error.details:
        response["error"]["details"] = error.details
    
    # Log server-side (with stack trace for 500s)
    if error.status >= 500:
        logger.error(f"Server error: {error.code}", exc_info=True,
                    extra={"request_id": g.request_id})
    
    return jsonify(response), error.status

# Usage in endpoint
@app.route('/api/v1/users/<user_id>')
def get_user(user_id):
    user = db.get_user(user_id)
    if not user:
        raise NotFoundError("User", user_id)
    return jsonify({"data": user.to_dict()})
```

---

## 5. API Versioning

### Versioning Strategies

| Strategy | Example | Pros | Cons |
|----------|---------|------|------|
| **URI path** | `/v1/users`, `/v2/users` | Simple, clear, cacheable | URL clutter, breaks HATEOAS |
| **Query param** | `/users?version=2` | Easy to add | Easy to forget, not standard |
| **Header** | `Accept: application/vnd.api.v2+json` | Clean URLs | Hidden, harder to test |
| **Content negotiation** | `Accept: application/vnd.github.v3+json` | RESTful | Complex |

### Recommendation: URI Path Versioning

```
/api/v1/users        → Version 1 (current stable)
/api/v2/users        → Version 2 (new features)

Why? 
- Most popular (Stripe, Twilio, GitHub, Twitter all use it)
- Obvious, no ambiguity
- Easy to route in load balancer/API gateway
- Easy to test in browser
```

### When to Version

| Change Type | Version Bump? | Example |
|-------------|--------------|---------|
| **Adding** a new endpoint | ❌ No | `POST /users/export` |
| **Adding** a new optional field | ❌ No | Adding `phone` to user response |
| **Adding** a new query parameter | ❌ No | `?include_inactive=true` |
| **Removing** a field | ✅ Yes (or deprecate) | Removing `username` from response |
| **Renaming** a field | ✅ Yes | `username` → `handle` |
| **Changing** field type | ✅ Yes | `id: number` → `id: string` |
| **Changing** URL structure | ✅ Yes | `/users/{id}/posts` → `/posts?user_id={id}` |
| **Changing** error format | ✅ Yes | Different error response structure |
| **Changing** behavior | ✅ Yes | POST now returns 201 instead of 200 |

---

## 6. Backward Compatibility

### Rules for Non-Breaking Changes

```
✅ SAFE (backward compatible):
  - Add new endpoints
  - Add new optional request fields
  - Add new response fields (clients should ignore unknown fields)
  - Add new query parameters
  - Add new enum values (if client handles "unknown" gracefully)
  - Add new error codes (if client handles "unknown" gracefully)

❌ BREAKING (requires new version):
  - Remove or rename endpoints
  - Remove or rename response fields
  - Change field types (string → number)
  - Make optional fields required
  - Change URL structure
  - Change authentication mechanism
  - Change pagination format
  - Remove enum values
```

### Graceful Evolution Example

```json
// v1 response
{
  "id": "user_123",
  "username": "alice",
  "email": "alice@example.com"
}

// v1 evolution (backward compatible — just add fields)
{
  "id": "user_123",
  "username": "alice",          // keep old field
  "handle": "alice",            // add new field (same value for now)
  "email": "alice@example.com",
  "avatar_url": null            // new optional field
}

// v2 (breaking change — new version)
{
  "id": "user_123",
  "handle": "alice",            // renamed from "username"
  "email": "alice@example.com",
  "avatar_url": null
  // "username" field removed
}
```

---

## 7. Deprecation Policy

### How to Deprecate

```
Step 1: Announce deprecation (N months before removal)
  - Documentation update
  - Deprecation notice in response headers
  - Email API consumers
  - Changelog entry

Step 2: Add deprecation headers
  Response Headers:
    Deprecation: true
    Sunset: Sat, 15 Jun 2025 00:00:00 GMT
    Link: <https://docs.example.com/migration/v2>; rel="deprecation"

Step 3: Log usage of deprecated endpoints
  - Track which API keys still use v1
  - Reach out to heavy users proactively

Step 4: Remove after sunset date
  Response:
    HTTP/1.1 410 Gone
    {
      "error": {
        "code": "API_VERSION_RETIRED",
        "message": "API v1 has been retired. Please migrate to v2.",
        "documentation_url": "https://docs.example.com/migration/v2"
      }
    }
```

### Real-World Deprecation Timelines

| Company | Deprecation Notice | Sunset Period |
|---------|-------------------|---------------|
| Stripe | 2+ years | Very long support |
| GitHub | 1 year notice | 6 months grace |
| Google Cloud | Per API, minimum 1 year | With migration guides |
| Twitter | 90 days notice | Abrupt (controversial!) |

---

## 8. API Changelog

### Maintain a changelog for every API change:

```markdown
# API Changelog

## v2.3.0 (2024-03-15)
### Added
- `GET /users/{id}/preferences` endpoint
- `avatar_url` field to user response
- Support for `sort` parameter on `GET /products`

### Changed
- `GET /orders` now defaults to last 30 days (was all time)

### Deprecated
- `username` field in user response (use `handle` instead, removal: 2025-01-01)
- `GET /users/{id}/posts` (use `GET /posts?user_id={id}`, removal: 2025-06-01)

## v2.2.0 (2024-02-01)
### Added  
- Cursor-based pagination on all list endpoints
- `X-Request-ID` response header

### Fixed
- `GET /products` now correctly applies `min_price` filter

### Removed
- `page` and `per_page` parameters (use `cursor` and `limit`)
```

---

## 9. Health Check & Status Endpoints

```
GET /health               → Basic health check (load balancer)
GET /health/ready          → Readiness check (can accept traffic)
GET /health/live           → Liveness check (is the process alive)
GET /status                → Detailed system status

// Health check response
{
  "status": "healthy",
  "version": "2.3.0",
  "uptime": "72h30m",
  "checks": {
    "database": { "status": "healthy", "latency_ms": 2 },
    "redis": { "status": "healthy", "latency_ms": 1 },
    "external_api": { "status": "degraded", "latency_ms": 450 }
  }
}
```

---

## 10. OpenAPI / Swagger Specification

```yaml
# openapi.yaml — machine-readable API contract
openapi: 3.0.3
info:
  title: E-Commerce API
  version: '2.0'
  description: API for managing products, orders, and users

servers:
  - url: https://api.example.com/v2

paths:
  /users:
    post:
      summary: Create a new user
      tags: [Users]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [name, email]
              properties:
                name:
                  type: string
                  maxLength: 200
                email:
                  type: string
                  format: email
                phone:
                  type: string
      responses:
        '201':
          description: User created
          headers:
            Location:
              schema:
                type: string
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '409':
          description: Email already exists
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        '422':
          description: Validation error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ValidationError'

components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
          example: "user_123"
        name:
          type: string
        email:
          type: string
          format: email
        created_at:
          type: string
          format: date-time
          
    Error:
      type: object
      properties:
        error:
          type: object
          properties:
            code:
              type: string
            message:
              type: string
            status:
              type: integer
              
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

security:
  - bearerAuth: []
```

### Why OpenAPI Matters in Interviews

| Benefit | Description |
|---------|-------------|
| **Contract-first design** | Design API before implementing |
| **Auto-generate docs** | Swagger UI, Redoc |
| **Auto-generate SDKs** | Client libraries in any language |
| **Testing** | Validate requests/responses against spec |
| **Team alignment** | Frontend/backend agree on contract before coding |
