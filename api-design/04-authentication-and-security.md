# Authentication & Security

## OAuth 2.0, JWT, API Keys, RBAC, CORS, Rate Limiting, and Input Validation

---

## 1. Authentication Methods Comparison

| Method | Security | Complexity | Best For |
|--------|----------|------------|----------|
| **API Key** | Low-Medium | Simple | Server-to-server, public data |
| **Basic Auth** | Low | Simple | Internal tools, development |
| **Bearer Token (JWT)** | High | Medium | User-facing APIs, SPAs |
| **OAuth 2.0** | High | Complex | Third-party access, SSO |
| **mTLS** | Very High | Complex | Service-to-service (zero trust) |
| **HMAC Signature** | High | Medium | Webhooks, payment APIs |

---

## 2. API Key Authentication

### How It Works

```
Request:
  GET /api/v1/weather?city=London
  Headers:
    X-API-Key: sk_live_abc123def456
```

### Implementation

```python
# Server-side API key validation
def authenticate_api_key(request):
    api_key = request.headers.get('X-API-Key')
    if not api_key:
        return error(401, "API key required")
    
    # Look up key (store hash, not plaintext!)
    key_hash = hash_key(api_key)
    key_record = db.query("SELECT * FROM api_keys WHERE key_hash = %s", key_hash)
    
    if not key_record:
        return error(401, "Invalid API key")
    if key_record.revoked:
        return error(401, "API key has been revoked")
    if key_record.expires_at and key_record.expires_at < now():
        return error(401, "API key expired")
    
    # Rate limit by key
    if is_rate_limited(key_record.id):
        return error(429, "Rate limit exceeded")
    
    return key_record  # authenticated
```

### API Key Best Practices

| Practice | Why |
|----------|-----|
| Prefix keys (`sk_live_`, `pk_test_`) | Identify key type at a glance |
| Store hashed, not plaintext | If DB is breached, keys aren't exposed |
| Support key rotation | Let users create new key before revoking old |
| Different keys for test/production | Prevent accidental production calls |
| Scope keys to specific permissions | Principle of least privilege |
| Set expiration dates | Don't have keys that last forever |

---

## 3. JWT (JSON Web Token)

### Structure

```
Header.Payload.Signature

eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.
eyJzdWIiOiJ1c2VyXzEyMyIsIm5hbWUiOiJBbGljZSIsInJvbGUiOiJhZG1pbiIsImlhdCI6MTcwNTMxMjYwMCwiZXhwIjoxNzA1MzE2MjAwfQ.
<signature>
```

```json
// Header (algorithm + type)
{ "alg": "RS256", "typ": "JWT" }

// Payload (claims)
{
  "sub": "user_123",          // subject (user ID)
  "name": "Alice",
  "email": "alice@example.com",
  "role": "admin",
  "permissions": ["read", "write", "delete"],
  "iat": 1705312600,          // issued at
  "exp": 1705316200,          // expires at (1 hour)
  "iss": "auth.example.com",  // issuer
  "aud": "api.example.com"    // audience
}

// Signature
RSASHA256(base64UrlEncode(header) + "." + base64UrlEncode(payload), privateKey)
```

### JWT Auth Flow

```
┌────────┐                    ┌────────────┐                 ┌────────┐
│ Client │                    │ Auth Server│                 │  API   │
└───┬────┘                    └─────┬──────┘                 └────┬───┘
    │                               │                             │
    │ POST /auth/login              │                             │
    │ {email, password}             │                             │
    │──────────────────────────────▶│                             │
    │                               │ Validate credentials       │
    │                               │ Generate JWT               │
    │ 200 OK                        │                             │
    │ {access_token, refresh_token} │                             │
    │◀──────────────────────────────│                             │
    │                               │                             │
    │ GET /api/v1/users/me                                       │
    │ Authorization: Bearer <JWT>                                │
    │────────────────────────────────────────────────────────────▶│
    │                                                            │ Verify JWT
    │                                                            │ (check signature,
    │ 200 OK {user data}                                         │  expiry, claims)
    │◀────────────────────────────────────────────────────────────│
    │                               │                             │
    │ (token expired)               │                             │
    │ POST /auth/refresh            │                             │
    │ {refresh_token}               │                             │
    │──────────────────────────────▶│                             │
    │ 200 OK {new access_token}     │                             │
    │◀──────────────────────────────│                             │
```

### Access Token vs Refresh Token

| Property | Access Token | Refresh Token |
|----------|-------------|---------------|
| **Purpose** | Authorize API requests | Get new access tokens |
| **Lifetime** | Short (15 min - 1 hour) | Long (7 days - 30 days) |
| **Storage (web)** | Memory (JS variable) | HttpOnly cookie |
| **Storage (mobile)** | Secure keychain | Secure keychain |
| **Sent to** | API server | Auth server only |
| **Revocable?** | Not easily (stateless) | Yes (server tracks) |

### JWT Implementation

```python
import jwt
from datetime import datetime, timedelta

# Generate tokens
def create_tokens(user):
    access_payload = {
        "sub": user.id,
        "name": user.name,
        "role": user.role,
        "type": "access",
        "iat": datetime.utcnow(),
        "exp": datetime.utcnow() + timedelta(minutes=15),
        "iss": "auth.example.com"
    }
    refresh_payload = {
        "sub": user.id,
        "type": "refresh",
        "iat": datetime.utcnow(),
        "exp": datetime.utcnow() + timedelta(days=7),
        "jti": str(uuid4()),  # unique ID for revocation
    }
    
    access_token = jwt.encode(access_payload, PRIVATE_KEY, algorithm="RS256")
    refresh_token = jwt.encode(refresh_payload, PRIVATE_KEY, algorithm="RS256")
    
    # Store refresh token hash for revocation
    db.store_refresh_token(refresh_payload["jti"], user.id, refresh_payload["exp"])
    
    return {"access_token": access_token, "refresh_token": refresh_token}

# Verify access token
def verify_token(token):
    try:
        payload = jwt.decode(token, PUBLIC_KEY, algorithms=["RS256"],
                           audience="api.example.com", issuer="auth.example.com")
        if payload["type"] != "access":
            raise ValueError("Not an access token")
        return payload
    except jwt.ExpiredSignatureError:
        raise AuthError(401, "Token expired")
    except jwt.InvalidTokenError:
        raise AuthError(401, "Invalid token")
```

### JWT vs Session-Based Auth

| Aspect | JWT | Session |
|--------|-----|---------|
| **State** | Stateless (token has all info) | Stateful (server stores session) |
| **Scalability** | Easy (no shared state) | Needs sticky sessions or shared store |
| **Revocation** | Hard (must wait for expiry) | Easy (delete session) |
| **Size** | Larger (payload in token) | Small (just session ID) |
| **Cross-domain** | Works (Bearer header) | Cookie issues across domains |
| **Mobile** | Easy | Cookie management is harder |

---

## 4. OAuth 2.0

### Grant Types

| Grant Type | Use Case | Flow |
|-----------|----------|------|
| **Authorization Code** | Web apps with backend | Most secure, redirect-based |
| **Authorization Code + PKCE** | SPAs, mobile apps | Like above but for public clients |
| **Client Credentials** | Machine-to-machine | Service accounts, no user involved |
| **Device Code** | TVs, IoT, CLI tools | User authenticates on separate device |
| ~~Implicit~~ | ~~SPAs~~ | ❌ Deprecated — use PKCE instead |
| ~~Password~~ | ~~Trusted apps~~ | ❌ Deprecated — don't use |

### Authorization Code Flow (Web App)

```
┌────────┐     ┌────────────────┐     ┌──────────────┐
│ Browser│     │  Your Backend  │     │ Auth Provider │
│ (User) │     │  (Confidential)│     │ (Google, etc)│
└───┬────┘     └───────┬────────┘     └──────┬───────┘
    │                  │                      │
    │ 1. Click "Login with Google"            │
    │─────────────────▶│                      │
    │                  │                      │
    │ 2. Redirect to Auth Provider            │
    │◀─────────────────│                      │
    │ Location: https://auth.google.com/authorize?
    │   client_id=xxx&redirect_uri=xxx&scope=email+profile
    │   &response_type=code&state=random_csrf
    │                  │                      │
    │ 3. User logs in and grants permission   │
    │─────────────────────────────────────────▶│
    │                  │                      │
    │ 4. Redirect back with authorization code│
    │◀─────────────────────────────────────────│
    │ Location: https://yourapp.com/callback?code=AUTH_CODE&state=random_csrf
    │                  │                      │
    │ 5. Forward code  │                      │
    │─────────────────▶│                      │
    │                  │ 6. Exchange code for tokens
    │                  │──────────────────────▶│
    │                  │ POST /oauth/token     │
    │                  │ {code, client_secret} │
    │                  │                      │
    │                  │ 7. Return tokens      │
    │                  │◀──────────────────────│
    │                  │ {access_token,       │
    │                  │  refresh_token,      │
    │                  │  id_token}           │
    │                  │                      │
    │ 8. Set session   │                      │
    │◀─────────────────│                      │
```

### OAuth 2.0 Scopes

```
// Request specific permissions
GET /authorize?scope=read:users+write:orders+read:products

// Common scope patterns:
read:users        → Read user data
write:users       → Modify user data
admin:users       → Full user management
openid            → OpenID Connect (get ID token)
profile           → User profile info
email             → User email
```

---

## 5. Authorization (RBAC & ABAC)

### Role-Based Access Control (RBAC)

```
Roles:
  admin     → full access to everything
  manager   → read/write own department
  member    → read/write own resources
  viewer    → read-only access

Permission Matrix:
┌──────────────┬───────┬─────────┬────────┬────────┐
│ Endpoint     │ admin │ manager │ member │ viewer │
├──────────────┼───────┼─────────┼────────┼────────┤
│ GET /users   │ ✅    │ ✅ own  │ ✅ own │ ✅ own │
│ POST /users  │ ✅    │ ✅      │ ❌     │ ❌     │
│ PUT /users/* │ ✅    │ ✅ own  │ ✅ own │ ❌     │
│ DELETE /users│ ✅    │ ❌      │ ❌     │ ❌     │
│ GET /reports │ ✅    │ ✅      │ ❌     │ ✅     │
│ GET /settings│ ✅    │ ✅      │ ❌     │ ❌     │
└──────────────┴───────┴─────────┴────────┴────────┘
```

### Authorization Middleware

```python
def authorize(required_permission):
    def middleware(request):
        user = request.authenticated_user  # from auth middleware
        
        if not user:
            return error(401, "Authentication required")
        
        # Check if user's role has the required permission
        if required_permission not in get_permissions(user.role):
            return error(403, "Insufficient permissions")
        
        # Resource-level check (e.g., can only edit own resources)
        if is_resource_specific(request):
            resource = get_resource(request)
            if not can_access(user, resource):
                return error(403, "You don't have access to this resource")
        
        return proceed(request)
    return middleware

# Usage:
@app.route('/api/v1/users', methods=['POST'])
@authorize('users:create')
def create_user(request):
    ...

@app.route('/api/v1/users/<id>', methods=['DELETE'])
@authorize('users:delete')
def delete_user(request, id):
    ...
```

### Attribute-Based Access Control (ABAC)

```python
# ABAC evaluates policies with attributes
# More flexible than RBAC — can consider context

def abac_authorize(user, resource, action, context):
    policies = [
        # Policy 1: Admins can do anything
        lambda u, r, a, c: u.role == 'admin',
        
        # Policy 2: Users can edit their own resources
        lambda u, r, a, c: a in ('read', 'update') and r.owner_id == u.id,
        
        # Policy 3: Managers can read department resources
        lambda u, r, a, c: a == 'read' and u.department == r.department,
        
        # Policy 4: Can only delete during business hours
        lambda u, r, a, c: a == 'delete' and 9 <= c.hour <= 17,
        
        # Policy 5: Sensitive data only from office IP
        lambda u, r, a, c: r.is_sensitive and c.ip in OFFICE_IPS,
    ]
    
    return any(policy(user, resource, action, context) for policy in policies)
```

---

## 6. Rate Limiting

### Strategies

| Strategy | Description | Example |
|----------|-------------|---------|
| **Fixed Window** | Count requests in fixed time window | 100 req/minute, resets on the minute |
| **Sliding Window** | Rolling window from current time | 100 req in last 60 seconds |
| **Token Bucket** | Tokens replenish at fixed rate | 10 tokens/sec, burst up to 100 |
| **Leaky Bucket** | Requests processed at fixed rate | Queue overflow → reject |

### Rate Limit Headers

```
Response Headers:
  X-RateLimit-Limit: 100           # max requests per window
  X-RateLimit-Remaining: 42        # remaining requests
  X-RateLimit-Reset: 1705316200    # when window resets (Unix timestamp)
  Retry-After: 30                  # seconds to wait (only on 429)

Response on rate limit hit:
  HTTP/1.1 429 Too Many Requests
  Retry-After: 30
  {
    "error": {
      "code": "RATE_LIMITED",
      "message": "Rate limit exceeded. Try again in 30 seconds.",
      "limit": 100,
      "remaining": 0,
      "reset_at": "2024-01-15T11:00:00Z"
    }
  }
```

### Token Bucket Implementation

```python
import time, redis

def check_rate_limit(user_id, limit=100, window=60):
    """Sliding window rate limiter with Redis."""
    key = f"rate_limit:{user_id}"
    now = time.time()
    
    pipe = redis.pipeline()
    pipe.zremrangebyscore(key, 0, now - window)     # remove old entries
    pipe.zadd(key, {f"{now}:{uuid4()}": now})        # add current request
    pipe.zcard(key)                                   # count in window
    pipe.expire(key, window)                          # set TTL
    _, _, count, _ = pipe.execute()
    
    remaining = max(0, limit - count)
    
    if count > limit:
        return {
            "allowed": False,
            "remaining": 0,
            "retry_after": window - (now - float(redis.zrange(key, 0, 0, withscores=True)[0][1]))
        }
    
    return {"allowed": True, "remaining": remaining}
```

### Rate Limiting Levels

| Level | Granularity | Example |
|-------|-------------|---------|
| **Global** | Entire API | 10,000 req/min across all users |
| **Per-user** | Individual user | 100 req/min per user |
| **Per-endpoint** | Specific route | POST /orders: 10/min; GET /products: 1000/min |
| **Per-API-key** | Different tiers | Free: 100/hr; Pro: 10,000/hr; Enterprise: unlimited |

---

## 7. CORS (Cross-Origin Resource Sharing)

```
Scenario: Frontend at app.example.com calls API at api.example.com

Preflight Request (browser sends automatically for complex requests):
  OPTIONS /api/v1/users
  Origin: https://app.example.com
  Access-Control-Request-Method: POST
  Access-Control-Request-Headers: Authorization, Content-Type

Server Response:
  Access-Control-Allow-Origin: https://app.example.com
  Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE
  Access-Control-Allow-Headers: Authorization, Content-Type, X-Request-ID
  Access-Control-Allow-Credentials: true
  Access-Control-Max-Age: 86400        # cache preflight for 24 hours
```

### CORS Configuration

```python
# ✅ Specific origins (production)
CORS_ORIGINS = ["https://app.example.com", "https://admin.example.com"]

# ❌ Never in production
Access-Control-Allow-Origin: *          # allows any website
Access-Control-Allow-Credentials: true  # can't use * with credentials

# Programmatic origin check
def get_cors_origin(request_origin):
    allowed = ["https://app.example.com", "https://admin.example.com"]
    if request_origin in allowed:
        return request_origin
    return None  # reject
```

---

## 8. Input Validation & Security

### Validation Layers

```
Layer 1: Schema Validation (JSON Schema / OpenAPI)
  → Required fields present?
  → Correct types?
  → Within length limits?

Layer 2: Business Validation
  → Email format valid?
  → Age >= 13?
  → Price > 0?
  → Referenced resources exist?

Layer 3: Security Validation
  → No SQL injection?
  → No XSS in text fields?
  → File type allowed?
  → File size within limits?
```

### Common Security Threats & Mitigations

| Threat | Mitigation |
|--------|------------|
| **SQL Injection** | Parameterized queries, ORM, input validation |
| **XSS** | Sanitize output, CSP headers, HTML encoding |
| **CSRF** | CSRF tokens, SameSite cookies, check Origin header |
| **Mass Assignment** | Whitelist allowed fields, don't bind request body directly |
| **Broken Auth** | Rate limit login, account lockout, MFA |
| **IDOR** | Always check resource ownership in authorization |
| **Rate Abuse** | Rate limiting, CAPTCHA for anonymous endpoints |
| **Data Exposure** | Don't return sensitive fields (password_hash, internal IDs) |

### IDOR (Insecure Direct Object Reference)

```python
# ❌ VULNERABLE: Trusts user-provided ID without authorization check
@app.route('/api/v1/orders/<order_id>')
def get_order(order_id):
    order = db.get_order(order_id)
    return order  # Any authenticated user can access any order!

# ✅ SECURE: Check ownership
@app.route('/api/v1/orders/<order_id>')
@authenticate
def get_order(order_id, current_user):
    order = db.get_order(order_id)
    if not order:
        return error(404, "Order not found")
    if order.user_id != current_user.id and current_user.role != 'admin':
        return error(403, "Access denied")
    return order
```

---

## 9. API Security Checklist

| Category | Checks |
|----------|--------|
| **Transport** | ✅ HTTPS only, ✅ HSTS header, ✅ TLS 1.2+ |
| **Authentication** | ✅ Strong passwords, ✅ Rate limit login, ✅ MFA support |
| **Authorization** | ✅ RBAC/ABAC, ✅ Resource-level checks, ✅ Principle of least privilege |
| **Input** | ✅ Validate all input, ✅ Parameterized queries, ✅ Max request size |
| **Output** | ✅ Don't expose internal errors, ✅ Don't leak sensitive data |
| **Headers** | ✅ CORS properly configured, ✅ Security headers (CSP, X-Frame-Options) |
| **Rate Limiting** | ✅ Per-user limits, ✅ Per-endpoint limits, ✅ 429 with Retry-After |
| **Logging** | ✅ Log auth events, ✅ Don't log tokens/passwords, ✅ Anomaly detection |
| **Keys** | ✅ Rotate regularly, ✅ Store securely, ✅ Different keys per environment |
