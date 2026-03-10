# Authentication & Authorization

## Staff+ Engineer Deep Dive | FAANG Interview Preparation

---

## 1. Concept Overview

**Authentication** answers "Who are you?" — verifying the identity of a user, service, or system. **Authorization** answers "What can you do?" — determining what actions an authenticated entity is permitted to perform on which resources.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    AUTHENTICATION vs AUTHORIZATION                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   AUTHENTICATION (Identity)          AUTHORIZATION (Permissions)             │
│   ─────────────────────────         ───────────────────────────            │
│   • Verify identity                  • Verify access rights                  │
│   • Happens first                    • Happens after auth                    │
│   • Credentials → Identity          • Identity + Resource → Allow/Deny     │
│   • "Are you John?"                 • "Can John delete this file?"           │
│                                                                             │
│   Input: Password, OTP, Biometric   Input: User + Action + Resource         │
│   Output: Session/Token             Output: Allow or Deny                    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Key Principle**: Authentication is binary (verified or not). Authorization is contextual (depends on resource, action, and context).

---

## 2. Real-World Motivation

- **Google OAuth**: Sign in with Google across millions of apps — one identity, delegated access
- **AWS IAM**: Fine-grained permissions for 200+ services — principle of least privilege at scale
- **Auth0/Okta**: Enterprise SSO — one login for 1000+ internal apps
- **Banking Apps**: Biometric + OTP for high-value transactions — multi-factor authentication
- **Kubernetes RBAC**: `kubectl` commands — service accounts vs user permissions

Without proper auth/authz: credential stuffing, session hijacking, privilege escalation, OWASP Top 10 vulnerabilities.

---

## 3. Architecture Diagrams (ASCII)

### OAuth 2.0 Authorization Code Flow

```
┌──────────┐     ┌──────────┐     ┌──────────────┐     ┌──────────┐     ┌──────────┐
│  User    │     │  Client  │     │ Auth Server │     │ Resource │     │  Token   │
│ (Browser)│     │   App    │     │  (Google)   │     │  Server  │     │ Endpoint │
└────┬─────┘     └────┬─────┘     └──────┬──────┘     └────┬─────┘     └────┬─────┘
     │               │                  │                 │                │
     │ 1. Click      │                  │                 │                │
     │ "Login"       │                  │                 │                │
     │──────────────>│                  │                 │                │
     │               │                  │                 │                │
     │ 2. Redirect to Auth Server       │                 │                │
     │<──────────────│                  │                 │                │
     │               │                  │                 │                │
     │ 3. User enters credentials       │                 │                │
     │────────────────────────────────>│                 │                │
     │               │                  │                 │                │
     │ 4. Redirect with auth code       │                 │                │
     │<────────────────────────────────│                 │                │
     │  (callback?code=xyz)              │                 │                │
     │               │                  │                 │                │
     │               │ 5. Exchange code for tokens        │                │
     │               │    (client_secret in body)         │                │
     │               │────────────────>│────────────────>│                │
     │               │                  │                 │                │
     │               │ 6. Access + Refresh tokens         │                │
     │               │<────────────────│<────────────────│                │
     │               │                  │                 │                │
     │               │ 7. API request + Bearer token      │                │
     │               │──────────────────────────────────>│                │
     │               │                  │                 │                │
     │               │ 8. Protected resource              │                │
     │               │<──────────────────────────────────│                │
     │               │                  │                 │                │
```

### JWT Structure

```
┌────────────────────────────────────────────────────────────────────────────────┐
│                         JWT (JSON Web Token) STRUCTURE                          │
├────────────────────────────────────────────────────────────────────────────────┤
│                                                                                │
│   HEADER.PAYLOAD.SIGNATURE                                                      │
│   (Base64URL)     (Base64URL)    (Base64URL)                                    │
│                                                                                │
│   ┌─────────────┐  ┌─────────────────────────────┐  ┌─────────────────────┐   │
│   │  HEADER     │  │  PAYLOAD (Claims)            │  │  SIGNATURE          │   │
│   ├─────────────┤  ├─────────────────────────────┤  ├─────────────────────┤   │
│   │ alg: HS256  │  │ sub: "user-123"              │  │ HMAC-SHA256(        │   │
│   │ typ: JWT    │  │ iat: 1699564800              │  │   base64(header) +  │   │
│   │             │  │ exp: 1699568400              │  │   "." +             │   │
│   │             │  │ aud: "api.example.com"       │  │   base64(payload),  │   │
│   │             │  │ custom: "claim"              │  │   secret            │   │
│   │             │  │                             │  │ )                    │   │
│   └─────────────┘  └─────────────────────────────┘  └─────────────────────┘   │
│                                                                                │
│   Standard Claims: sub, iss, aud, exp, iat, nbf, jti                           │
│   Signature: RS256 (asymmetric) or HS256 (symmetric)                           │
│                                                                                │
└────────────────────────────────────────────────────────────────────────────────┘
```

### Session vs Token Architecture

```
SESSION-BASED                          TOKEN-BASED (JWT)
──────────────                         ─────────────────

┌────────┐    Cookie     ┌────────┐    ┌────────┐    Bearer Token   ┌────────┐
│ Client │──────────────>│ Server │    │ Client │─────────────────>│ Server │
└────────┘  (session_id) └───┬────┘    └────────┘  (self-contained) └───┬────┘
                             │                                          │
                             │ Lookup session in                         │ Verify signature
                             │ Redis/DB                                  │ (stateless)
                             │                                          │
                             v                                          v
                      ┌────────────┐                            ┌────────────┐
                      │ Session    │                            │ No lookup   │
                      │ Store      │                            │ needed     │
                      └────────────┘                            └────────────┘
```

---

## 4. Core Mechanics

### Authentication Methods

| Method | How It Works | Use Case |
|--------|--------------|----------|
| **Session-based** | Server stores session in Redis/DB, client gets session ID in cookie | Traditional web apps |
| **JWT** | Stateless token with claims, signed by server | APIs, microservices |
| **OAuth 2.0** | Delegated authorization, third-party identity | "Sign in with Google" |
| **OIDC** | OAuth + identity layer (id_token with user info) | SSO, enterprise |
| **API Keys** | Static string in header | Server-to-server, internal APIs |
| **mTLS** | Client certificate for mutual authentication | Service mesh, zero-trust |

### Token Lifecycle: Access + Refresh

```
Access Token (short-lived: 15 min)     Refresh Token (long-lived: 7 days)
├── Used for API calls                 ├── Stored securely (httpOnly cookie)
├── Stored in memory only              ├── Used only to get new access token
├── If stolen: limited window          ├── If stolen: can get new access tokens
└── Rotation: refresh before expiry    └── Rotation: issue new refresh on use
```

### OAuth 2.0 Grant Types

- **Authorization Code**: Web apps, most secure, PKCE for SPAs
- **Client Credentials**: Machine-to-machine, no user context
- **Implicit** (deprecated): SPA-only, token in URL fragment — avoid
- **Resource Owner Password**: Legacy migration only — avoid
- **Refresh Token**: Obtain new access token

### Authorization Models

**RBAC (Role-Based Access Control)**
```
User → Role → Permissions
admin → AdminRole → [read:*, write:*, delete:*]
editor → EditorRole → [read:*, write:*]
viewer → ViewerRole → [read:*]
```

**ABAC (Attribute-Based Access Control)**
```
Policy: ALLOW if
  user.department == resource.owner_department AND
  resource.classification <= user.clearance AND
  request.time in business_hours
```

**ACL (Access Control List)**
```
Resource: /documents/secret.pdf
  user:alice → read, write
  user:bob → read
  group:legal → read, write, delete
```

---

## 5. Numbers

| Metric | Value | Context |
|--------|-------|---------|
| JWT size | 200-500 bytes | Typical payload |
| Session cookie | 32-64 bytes | Session ID only |
| OAuth flow latency | 200-500ms | Redirect + token exchange |
| JWT verification | <1ms | Local signature check |
| Session lookup | 1-5ms | Redis round-trip |
| bcrypt cost factor | 10-12 | ~100ms per hash |
| Token expiry (access) | 15 min - 1 hr | Balance security vs UX |
| Token expiry (refresh) | 7-30 days | Revocable |
| OAuth PKCE code_verifier | 43-128 chars | S256: 43 chars min |

---

## 6. Tradeoffs (Comparison Tables)

### Session vs JWT

| Aspect | Session-based | JWT |
|--------|---------------|-----|
| **State** | Stateful (server stores) | Stateless |
| **Scalability** | Needs shared session store | Horizontal scaling trivial |
| **Revocation** | Delete session = instant | Hard (until expiry) |
| **Payload size** | Minimal (ID only) | Larger (claims in token) |
| **Cross-domain** | CORS/cookie complexity | Bearer token works anywhere |
| **Mobile** | Cookie handling awkward | Native support |
| **Security** | Server controls | Client holds token |

### OAuth Grant Types

| Grant | Security | Use Case | User Interaction |
|-------|----------|----------|------------------|
| Authorization Code | High | Web apps | Yes |
| Authorization Code + PKCE | High | SPAs, mobile | Yes |
| Client Credentials | Medium | M2M | No |
| Implicit | Low (deprecated) | — | Yes |
| Refresh Token | — | Token renewal | No |

### Auth Provider Comparison

| Provider | SSO | MFA | OIDC | Enterprise | Pricing |
|----------|-----|-----|------|------------|---------|
| Auth0 | ✓ | ✓ | ✓ | ✓ | Freemium |
| Okta | ✓ | ✓ | ✓ | ✓ | Enterprise |
| AWS Cognito | ✓ | ✓ | ✓ | Limited | Usage-based |
| Firebase Auth | ✓ | ✓ | ✓ | No | Freemium |
| Keycloak (OSS) | ✓ | ✓ | ✓ | ✓ | Free |

---

## 7. Variants/Implementations

### PKCE (Proof Key for Code Exchange)

Prevents authorization code interception in public clients (SPAs, mobile):

```
1. Client generates: code_verifier (random 43-128 chars)
2. code_challenge = BASE64URL(SHA256(code_verifier))
3. Auth request includes: code_challenge, code_challenge_method=S256
4. Token exchange includes: code_verifier
5. Server verifies: SHA256(code_verifier) == stored code_challenge
```

### Token Rotation

- **Refresh token rotation**: Each refresh returns new refresh token, invalidates old
- **Sliding sessions**: Extend session on activity
- **Reuse detection**: Reject refresh token if previous one used again (token theft)

### Policy Engines: OPA (Open Policy Agent)

```rego
# OPA policy example
package authz

default allow = false

allow {
    input.method == "GET"
    input.path == ["users", user_id]
    input.user.id == user_id
}

allow {
    input.method == "GET"
    input.path == ["users"]
    input.user.roles[_] == "admin"
}
```

---

## 8. Scaling Strategies

- **Session store**: Redis Cluster, consistent hashing for session affinity
- **JWT**: No server state — add more API servers
- **OAuth**: Cache JWKS (public keys) at edge, rate limit token endpoint
- **Policy evaluation**: OPA sidecar or centralized — cache policy decisions
- **Auth service**: Stateless, horizontal scaling; database for user store

---

## 9. Failure Scenarios

| Failure | Impact | Mitigation |
|--------|--------|------------|
| Session store down | All users logged out | Redis Sentinel, multi-AZ |
| JWT secret leaked | Forge any token | Rotate secret, short expiry |
| OAuth provider down | Cannot login | Multiple IdPs, cached tokens |
| Token theft | Attacker impersonates | Short expiry, refresh rotation |
| CORS misconfiguration | Token sent to attacker site | Strict origin, SameSite cookies |
| CSRF | Unauthorized actions | CSRF tokens, SameSite=Strict |

---

## 10. Performance Considerations

- **JWT**: Avoid large payloads — use token for identity, fetch details from DB
- **Session**: Use Redis over DB for session store (10x faster)
- **OAuth**: Cache IdP's JWKS (public keys) — avoid fetch per request
- **Policy evaluation**: OPA compiles policies — cache compiled result
- **bcrypt**: Cost factor 10-12 — tune for your hardware

---

## 11. Use Cases

| Use Case | Recommended Approach |
|----------|---------------------|
| Web app (server-rendered) | Session + secure cookie |
| SPA (React, Vue) | OAuth + PKCE, JWT |
| Mobile app | OAuth + PKCE, refresh tokens |
| Microservices | mTLS or JWT with service account |
| Internal API | API key or mTLS |
| Enterprise SSO | SAML or OIDC with IdP |
| Serverless | JWT (stateless) |
| Third-party API access | OAuth client credentials |

---

## 12. Comparison Tables

### Authentication Methods Summary

| Method | Stateless | Revocable | Cross-Domain | Complexity |
|--------|-----------|-----------|--------------|------------|
| Session | No | Yes | Hard | Low |
| JWT | Yes | No* | Easy | Medium |
| OAuth | Depends | Yes | Easy | High |
| API Key | Yes | Yes | Easy | Low |
| mTLS | Yes | Yes (revoke cert) | Medium | High |

*JWT revocation requires blacklist or short expiry

### Authorization Model Selection

| Model | When to Use | Complexity |
|-------|-------------|------------|
| RBAC | Roles map cleanly to permissions | Low |
| ABAC | Fine-grained, context-dependent | High |
| ACL | Per-resource permissions | Medium |
| Policy Engine | Complex rules, audit requirements | High |

---

## 13. Code/Pseudocode

### JWT Verification (Pseudocode)

```python
def verify_jwt(token: str, jwks: dict) -> Optional[Claims]:
    header_b64, payload_b64, signature_b64 = token.split('.')
    header = base64url_decode(header_b64)
    payload = base64url_decode(payload_b64)
    
    # 1. Check expiry
    if payload['exp'] < now():
        return None
    
    # 2. Get signing key from JWKS
    key_id = header.get('kid')
    public_key = jwks.get_key(key_id)
    
    # 3. Verify signature
    message = f"{header_b64}.{payload_b64}"
    if not verify_signature(message, signature_b64, public_key):
        return None
    
    return payload
```

### OAuth Authorization Code Flow (Pseudocode)

```python
# Step 1: Redirect to auth server
def initiate_login():
    state = random_string(32)
    code_verifier = random_string(43)  # PKCE
    code_challenge = base64url(sha256(code_verifier))
    store_in_session(state=state, code_verifier=code_verifier)
    
    redirect(f"{auth_server}/authorize?"
             f"client_id={client_id}&"
             f"redirect_uri={redirect_uri}&"
             f"response_type=code&"
             f"scope=openid profile&"
             f"state={state}&"
             f"code_challenge={code_challenge}&"
             f"code_challenge_method=S256")

# Step 2: Callback - exchange code for tokens
def handle_callback(code: str, state: str):
    if state != get_from_session('state'):
        raise CSRFError()
    
    code_verifier = get_from_session('code_verifier')
    tokens = post(auth_server + "/token", data={
        "grant_type": "authorization_code",
        "code": code,
        "redirect_uri": redirect_uri,
        "client_id": client_id,
        "client_secret": client_secret,
        "code_verifier": code_verifier
    })
    
    store_tokens(tokens['access_token'], tokens['refresh_token'])
```

### RBAC Check (Pseudocode)

```python
def check_permission(user: User, action: str, resource: str) -> bool:
    for role in user.roles:
        permissions = ROLE_PERMISSIONS[role]
        if f"{action}:{resource}" in permissions or f"{action}:*" in permissions:
            return True
    return False
```

---

## 14. Interview Discussion

### Key Points to Articulate

1. **Auth vs Authz**: "Authentication establishes identity; authorization determines what that identity can do. They're sequential — auth first, then authz on each request."

2. **JWT Tradeoffs**: "JWTs are stateless and scale horizontally, but revocation is hard. Use short expiry (15 min) + refresh tokens. For instant revocation, you need a blacklist — which adds state."

3. **OAuth Flow Choice**: "Authorization code with PKCE for SPAs — the code is exchanged server-side, and PKCE prevents code interception. Never use implicit flow; it's deprecated."

4. **Session vs JWT**: "Sessions give instant revocation and smaller client payload. JWTs enable stateless scaling and work better for APIs and mobile. Choose based on revocation needs and architecture."

5. **RBAC vs ABAC**: "RBAC is simpler — roles map to permissions. ABAC handles complex policies like 'same department' or 'business hours' but needs a policy engine."

### Common Follow-ups

- **"How would you design auth for a microservices architecture?"** — API gateway validates JWT or mTLS; service-to-service uses either. Consider OPA for centralized policy.

- **"What if JWT secret is compromised?"** — Rotate immediately; all tokens invalid. Use asymmetric (RS256) so private key stays server-side.

- **"How does OAuth prevent CSRF?"** — `state` parameter: client generates random state, stores it, includes in auth request; callback validates state matches.

- **"When would you use mTLS?"** — Service mesh, zero-trust networks, high-security internal APIs. Both client and server present certificates.
