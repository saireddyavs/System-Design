# Module 14: Security & Identity

---

## 1. OAuth 2.0 & OIDC

### Definition
**OAuth 2.0**: Authorization framework for delegating access without sharing passwords. "Let App X access my photos, but not my email."
**OIDC (OpenID Connect)**: Identity layer on top of OAuth. "Who am I?" + "What can I access?"

### OAuth 2.0 Flow (Authorization Code)
```
┌────────┐                           ┌──────────────┐
│  User  │                           │ Auth Server  │
│(Browser)│                           │(Google/Okta) │
└───┬────┘                           └──────┬───────┘
    │  1. Click "Login with Google"         │
    │──────────────────────────────────────→│
    │  2. Login + Consent screen            │
    │←──────────────────────────────────────│
    │  3. Redirect with auth_code           │
    │──→ ┌─────────┐                        │
    │    │  Your   │  4. Exchange code      │
    │    │  App    │   for tokens           │
    │    │(Backend)│───────────────────────→│
    │    │         │←──────────────────────│
    │    │         │  5. Access Token +     │
    │    │         │     Refresh Token      │
    │    └─────────┘                        │
```

### Token Types
```
Access Token:  Short-lived (15min), used to call APIs
Refresh Token: Long-lived (days), used to get new access tokens
ID Token:      JWT with user identity claims (OIDC only)
```

### OAuth 2.0 vs OIDC

| | OAuth 2.0 | OIDC |
|-|-----------|------|
| Purpose | Authorization (what can I do?) | Authentication (who am I?) |
| Token | Access Token | ID Token (JWT) |
| Use case | API access delegation | Login / SSO |

### Real Systems
Google Login, Facebook Login, GitHub OAuth, Auth0, Okta

---

## 2. JWT (JSON Web Tokens)

### Definition
A self-contained, signed token that encodes user claims. The server can verify it without a database lookup — just verify the signature.

### Structure
```
Header.Payload.Signature

eyJhbGciOiJSUzI1NiJ9.     ← Header (algorithm)
eyJ1c2VyIjoiYWxpY2UifQ.   ← Payload (claims)
SflKxwRJSMeKKF2QT4fw...   ← Signature (RSA/HMAC)

Decoded Payload:
{
  "sub": "user_123",
  "name": "Alice",
  "role": "admin",
  "exp": 1709251200,      ← expires at
  "iat": 1709247600       ← issued at
}
```

### Verification
```
Server receives JWT:
  1. Decode header → get algorithm (RS256)
  2. Verify signature using public key
  3. Check exp (not expired?)
  4. Check iss (trusted issuer?)
  5. Trust claims without DB lookup!
```

### Tradeoffs

| Pros | Cons |
|------|------|
| Stateless (no session store) | Cannot be revoked until expiry |
| Scalable (no shared state) | Payload is readable (not encrypted by default) |
| Works across microservices | Token size grows with claims |

### Security Pitfalls
```
1. Don't store sensitive data in JWT (base64 ≠ encrypted)
2. Always verify signature (don't accept alg=none)
3. Short expiry + refresh tokens for revocation
4. Use RS256 (asymmetric) not HS256 (symmetric) in distributed systems
```

---

## 3. End-to-End Encryption (Double Ratchet)

### Definition
An encryption protocol where keys change with every message, providing forward secrecy and future secrecy.

### How It Works (Signal Protocol)
```
Alice and Bob exchange initial key material (X3DH key agreement).

For each message:
  1. Derive a new message key from the "ratchet state"
  2. Encrypt message with this key
  3. Advance the ratchet (old key is deleted)
  4. Periodically exchange new Diffie-Hellman keys (DH ratchet)
```

### Forward Secrecy
```
Key K1 → encrypt msg1 → delete K1
Key K2 → encrypt msg2 → delete K2

Attacker steals K3: Can only read msg3.
Cannot derive K1 or K2 → past messages are SAFE.
```

### Real Systems
Signal, WhatsApp, Matrix/Element, iMessage (partial)

---

## 4. TLS 1.3

### Definition
The latest version of Transport Layer Security, reducing handshake from 2 RTT to 1 RTT (0 RTT on resumption).

### TLS 1.2 vs 1.3 Handshake
```
TLS 1.2 (2 round trips):
  Client → Server: ClientHello
  Server → Client: ServerHello, Certificate, KeyExchange
  Client → Server: KeyExchange, ChangeCipherSpec, Finished
  Server → Client: ChangeCipherSpec, Finished
  (2 RTT before encrypted data)

TLS 1.3 (1 round trip):
  Client → Server: ClientHello + KeyShare (guesses algorithm)
  Server → Client: ServerHello + KeyShare + Encrypted Extensions + Finished
  Client → Server: Finished + Application Data
  (1 RTT, encrypted data with first flight)

TLS 1.3 0-RTT (resumption):
  Client → Server: ClientHello + Early Data (encrypted with cached key)
  (0 RTT for resumed connections!)
```

### Key Improvements
- Removed insecure algorithms (RSA key exchange, SHA-1, RC4)
- Faster handshake (1-RTT vs 2-RTT)
- Encrypted more of the handshake (SNI still visible, ECH coming)
- Simpler protocol (fewer options = fewer bugs)

### Real Systems
Cloudflare, all modern browsers, NGINX, Let's Encrypt

---

## 5. Mutual TLS (mTLS)

### Definition
Both client and server present certificates and verify each other's identity. Standard TLS only authenticates the server.

### How It Works
```
Regular TLS:
  Client ──→ Server presents cert ──→ Client verifies server
  (Only server is authenticated)

Mutual TLS:
  Client ──→ Server presents cert ──→ Client verifies server
  Server ──→ Client presents cert ──→ Server verifies client
  (Both sides authenticated!)
```

### Use Case: Zero Trust
```
Traditional (perimeter security):
  Firewall protects internal network
  Internal services trust each other ← WRONG (one breach = all compromised)

Zero Trust (mTLS):
  Every service has a certificate
  Service A → Service B: both verify certificates
  No implicit trust, even inside the network
```

### Real Systems
Istio (service mesh), Google BeyondCorp, Uber (internal services), Kubernetes (API server)

---

## 6. DDoS Mitigation

### Techniques
```
┌─── LAYER 3/4 (Volumetric) ──────────────────────┐
│ Attack: Flood bandwidth (100+ Gbps)              │
│ Defense:                                          │
│   - Anycast: Split traffic across 300 POPs       │
│   - BGP blackhole: Drop traffic at ISP level     │
│   - SYN cookies: Validate TCP without state      │
└──────────────────────────────────────────────────┘

┌─── LAYER 7 (Application) ───────────────────────┐
│ Attack: HTTP flood (legitimate-looking requests)  │
│ Defense:                                          │
│   - Rate limiting per IP/token                    │
│   - CAPTCHA challenges                            │
│   - WAF (Web Application Firewall) rules          │
│   - Bot detection (JS challenge, fingerprinting)  │
└──────────────────────────────────────────────────┘

┌─── SCRUBBING ───────────────────────────────────┐
│ All traffic routed through scrubbing center       │
│ Malicious traffic dropped                         │
│ Clean traffic forwarded to origin                 │
└──────────────────────────────────────────────────┘
```

### Real Systems
Cloudflare, Akamai, AWS Shield, Google Cloud Armor

---

## 7. Macaroons

### Definition
A token-based authorization credential that allows holders to add restrictions (caveats) without contacting the issuer.

### How It Works
```
Server issues macaroon:
  MAC = HMAC(secret, "user=alice, service=storage")

Alice adds caveat (locally, no server contact):
  MAC' = HMAC(MAC, "time < 2026-12-31")

Alice adds another caveat:
  MAC'' = HMAC(MAC', "ip = 10.0.0.1")

Server verifies by chaining HMACs from root secret.

Key: Caveats can only be ADDED, never removed.
     Token becomes MORE restrictive over time.
```

### Macaroons vs JWT

| | JWT | Macaroons |
|-|-----|-----------|
| Restrictions | Set at issuance | Can be added by holder |
| Delegation | Share entire token | Share with extra caveats |
| Revocation | Cannot (until expiry) | Third-party caveats enable it |
| Verification | Signature check | HMAC chain |

### Real Systems
Google (internal auth), Bitcoin Lightning Network

---

## 8. Password Hashing (Salt & Pepper)

### Why Not Plain Hash?
```
MD5("password123") = 482c811da5d5b4bc...
Attacker has precomputed "rainbow table" of all common passwords.
Lookup: 482c811... → "password123"   CRACKED instantly!
```

### Salt
```
Random string per user, stored alongside hash.

hash("password123" + "a1b2c3") = 7f3d...  (unique per user)

Even if two users have same password, hashes differ.
Rainbow tables are useless (can't precompute all salts).
```

### Pepper
```
Secret key stored separately (HSM or env variable).
NOT in the database.

hash("password123" + salt + pepper)

Even if entire DB is stolen, attacker doesn't have pepper.
```

### Algorithms (2026)
```
┌──────────┬──────────┬──────────────────────────────┐
│ Algorithm│ Speed    │ Notes                         │
├──────────┼──────────┼──────────────────────────────┤
│ MD5      │ Fast     │ NEVER use (broken)            │
│ SHA-256  │ Fast     │ NOT for passwords (too fast)  │
│ bcrypt   │ Slow ✓   │ Good, widely used             │
│ scrypt   │ Slow ✓   │ Memory-hard (resists GPU)     │
│ Argon2id │ Slow ✓   │ BEST: memory + time hard      │
└──────────┴──────────┴──────────────────────────────┘

"Slow" is a FEATURE — makes brute force impractical.
Argon2id with 64MB memory, 3 iterations = ~1 second to verify.
Attacker must spend 1 second PER GUESS.
```

---

## 9. Padding Oracle Attack

### Definition
An attack exploiting error messages that reveal whether decryption padding is valid, allowing byte-by-byte decryption without the key.

### How It Works
```
CBC-mode encrypted ciphertext:
  Attacker modifies a byte in ciphertext
  Sends to server for decryption

Server responds:
  "Padding error" → padding was wrong
  "MAC error"     → padding was correct, but content wrong
  "Success"       → decrypted correctly

The difference between "padding error" and "MAC error"
leaks 1 bit of information. Repeat 256 × N times
to decrypt N bytes.
```

### Prevention
```
1. ALWAYS return generic error: "Decryption failed"
2. Use authenticated encryption (AES-GCM) instead of CBC
3. Verify MAC BEFORE checking padding (Encrypt-then-MAC)
4. Use TLS 1.3 (eliminates vulnerable cipher suites)
```

### Famous Vulnerability
POODLE attack (2014): Exploited SSL 3.0's CBC padding. Forced browsers to stop supporting SSL 3.0 entirely.

### Summary
Padding oracle attacks exploit timing/error differences in decryption. Prevent by using authenticated encryption (AES-GCM) and returning generic errors.
