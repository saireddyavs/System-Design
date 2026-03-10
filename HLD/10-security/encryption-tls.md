# Encryption & TLS

## Staff+ Engineer Deep Dive | FAANG Interview Preparation

---

## 1. Concept Overview

**Encryption** transforms plaintext into ciphertext using algorithms and keys, ensuring confidentiality. **TLS (Transport Layer Security)** provides encryption in transit — securing data as it moves between client and server over networks.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    ENCRYPTION LANDSCAPE                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   SYMMETRIC                    ASYMMETRIC                   HASHING         │
│   Same key for encrypt/decrypt  Public/private key pair      One-way         │
│   Fast, bulk data              Key exchange, signatures     Integrity       │
│   AES-256, ChaCha20            RSA, ECC (ECDHE)             SHA-256, bcrypt │
│                                                                             │
│   AT REST                      IN TRANSIT                   END-TO-END      │
│   Database, disk, backups       TLS between client-server    Client-to-client│
│   AES encryption               TLS 1.2, 1.3                 Signal protocol │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Hybrid Approach**: TLS uses asymmetric crypto for key exchange, then symmetric for bulk encryption (best of both).

---

## 2. Real-World Motivation

- **HTTPS**: Every website — TLS encrypts traffic, prevents eavesdropping
- **Database encryption**: AWS RDS, encrypted EBS — protect data at rest
- **Password storage**: bcrypt, Argon2 — never store plaintext
- **WhatsApp/Signal**: E2E encryption — only sender and recipient can read
- **mTLS in Kubernetes**: Service mesh (Istio) — zero-trust service-to-service
- **PCI DSS, HIPAA**: Compliance mandates encryption at rest and in transit

---

## 3. Architecture Diagrams (ASCII)

### TLS 1.2 Full Handshake Sequence

```
┌─────────┐                                              ┌─────────┐
│  Client │                                              │  Server │
└────┬────┘                                              └────┬────┘
     │                                                        │
     │  1. Client Hello                                        │
     │     - Supported TLS version                            │
     │     - Cipher suites (e.g., TLS_ECDHE_RSA_AES_256_GCM)  │
     │     - Random (28 bytes)                                │
     │     - Extensions (SNI, supported groups)              │
     │──────────────────────────────────────────────────────>│
     │                                                        │
     │  2. Server Hello                                        │
     │     - Chosen TLS version, cipher suite                 │
     │     - Random (28 bytes)                                │
     │     - Session ID                                       │
     │<──────────────────────────────────────────────────────│
     │                                                        │
     │  3. Certificate                                        │
     │     - Server's X.509 cert chain                        │
     │     - Public key (RSA or ECC)                          │
     │<──────────────────────────────────────────────────────│
     │                                                        │
     │  4. Server Key Exchange (for DHE/ECDHE)                │
     │     - Ephemeral public key                             │
     │<──────────────────────────────────────────────────────│
     │                                                        │
     │  5. Server Hello Done                                  │
     │<──────────────────────────────────────────────────────│
     │                                                        │
     │  6. Client Key Exchange                                │
     │     - Ephemeral public key (pre-master secret)         │
     │──────────────────────────────────────────────────────>│
     │                                                        │
     │  7. Change Cipher Spec (both sides)                    │
     │<──────────────────────────────────────────────────────>│
     │                                                        │
     │  8. Encrypted Handshake Finished                       │
     │<──────────────────────────────────────────────────────>│
     │                                                        │
     │  9. Application Data (encrypted)                      │
     │<══════════════════════════════════════════════════════>│
     │                                                        │
```

### TLS 1.3 Simplified Handshake (1-RTT)

```
┌─────────┐                                    ┌─────────┐
│  Client │                                    │  Server │
└────┬────┘                                    └────┬────┘
     │                                               │
     │  Client Hello                                 │
     │  + Key Share (client's ephemeral public key)  │
     │──────────────────────────────────────────────>│
     │                                               │
     │  Server Hello                                 │
     │  + Key Share (server's ephemeral public key)   │
     │  + Certificate                                │
     │  + Certificate Verify                         │
     │  + Finished                                   │
     │<──────────────────────────────────────────────│
     │                                               │
     │  Finished                                     │
     │──────────────────────────────────────────────>│
     │                                               │
     │  Application Data (encrypted)                  │
     │<══════════════════════════════════════════════>│
     │                                               │
     │  TLS 1.3: 1 RTT (vs 2 RTT in TLS 1.2)         │
     │  Removed: RSA key exchange, static DH        │
     │  Mandatory: Ephemeral keys (Perfect Forward   │
     │             Secrecy)                          │
     │                                               │
```

### Encryption at Rest Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    ENCRYPTION AT REST                                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   Application                    KMS (Key Management Service)                │
│   ┌─────────────┐               ┌─────────────────────────────┐            │
│   │ Data Key    │               │ Master Key (never leaves)    │            │
│   │ (DEK)       │<───Generate───│ Customer Master Key (CMK)    │            │
│   │ Encrypts    │               │                              │            │
│   │ actual data │               │ Envelope Encryption:         │            │
│   └──────┬──────┘               │ DEK encrypted by CMK         │            │
│          │                      └─────────────────────────────┘            │
│          │                                                                  │
│          v                                                                  │
│   ┌─────────────────────────────────────────────────────────────┐          │
│   │ Database / Disk / S3                                        │          │
│   │ ┌─────────────┐  ┌─────────────────────────────────────┐   │          │
│   │ │ Encrypted   │  │ Encrypted DEK (stored with data)     │   │          │
│   │ │ Data        │  │ Decrypt with CMK to get DEK           │   │          │
│   │ └─────────────┘  └─────────────────────────────────────┘   │          │
│   └─────────────────────────────────────────────────────────────┘          │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### mTLS (Mutual TLS) Flow

```
┌─────────────┐                    ┌─────────────┐
│   Client    │                    │   Server    │
│ (has cert)  │                    │ (has cert)  │
└──────┬──────┘                    └──────┬──────┘
       │                                   │
       │  Client Hello                      │
       │──────────────────────────────────>│
       │                                   │
       │  Server Hello + Server Certificate │
       │  + Certificate Request (asks for   │
       │    client cert)                    │
       │<──────────────────────────────────│
       │                                   │
       │  Client Certificate                │
       │  + Client Key Exchange             │
       │  + Certificate Verify (proves     │
       │    client has private key)         │
       │──────────────────────────────────>│
       │                                   │
       │  Both sides verify each other's   │
       │  certificate against trusted CA   │
       │                                   │
       │  Encrypted Application Data        │
       │<══════════════════════════════════>│
       │                                   │
```

---

## 4. Core Mechanics

### Symmetric Encryption (AES-256)

- **Same key** for encrypt and decrypt
- **Block cipher**: 128-bit blocks, 256-bit key
- **Modes**: GCM (authenticated), CBC (legacy), CTR
- **Use**: Bulk data encryption, TLS record layer

### Asymmetric Encryption (RSA, ECC)

- **Key pair**: Public key (encrypt/verify), Private key (decrypt/sign)
- **RSA-2048**: ~112-bit security, slower
- **ECC (P-256)**: Same security, smaller keys, faster
- **Use**: Key exchange, digital signatures

### Hashing (SHA-256, bcrypt)

- **One-way**: Cannot reverse
- **Deterministic**: Same input → same output
- **Password hashing**: bcrypt, Argon2, scrypt — include salt, slow by design
- **Integrity**: SHA-256 for checksums, HMAC for authenticated hashing

### TLS Cipher Suite

Format: `TLS_[Key Exchange]_[Auth]_WITH_[Bulk Cipher]_[Mode]`

Example: `TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384`
- **ECDHE**: Ephemeral Elliptic Curve Diffie-Hellman (PFS)
- **RSA**: Server authentication
- **AES_256_GCM**: Bulk encryption
- **SHA384**: PRF (Pseudo-Random Function)

---

## 5. Numbers

| Metric | Value | Context |
|--------|-------|---------|
| AES-256 throughput | 1-10+ GB/s | Hardware AES-NI acceleration |
| RSA-2048 sign/verify | 1-10 ms | Per operation |
| RSA-2048 encrypt | ~0.1 ms | Encrypt with public key |
| ECDHE P-256 | ~1 ms | Key exchange |
| TLS handshake (1.2) | 50-200 ms | 2 RTT + crypto |
| TLS handshake (1.3) | 25-100 ms | 1 RTT |
| bcrypt cost 12 | ~250 ms | Per password hash |
| SHA-256 | ~500 MB/s | Hashing throughput |
| Certificate validity | 90 days (Let's Encrypt) | Was 398 days max |
| Key size equivalence | RSA-2048 ≈ ECC-224 | Security level |

---

## 6. Tradeoffs (Comparison Tables)

### Symmetric vs Asymmetric

| Aspect | Symmetric (AES) | Asymmetric (RSA/ECC) |
|--------|-----------------|----------------------|
| **Speed** | Very fast (GB/s) | Slow (ms per op) |
| **Key size** | 256 bits | 2048+ bits (RSA), 256 bits (ECC) |
| **Key distribution** | Must share secretly | Public key can be shared |
| **Use case** | Bulk encryption | Key exchange, signatures |
| **Key management** | N keys for N pairs | 1 key pair per entity |

### TLS 1.2 vs TLS 1.3

| Aspect | TLS 1.2 | TLS 1.3 |
|--------|---------|---------|
| **Handshake RTT** | 2 RTT | 1 RTT |
| **Key exchange** | RSA, DHE, ECDHE | ECDHE only (PFS mandatory) |
| **Cipher suites** | 37+ | 5 (all AEAD) |
| **0-RTT** | No | Yes (optional, replay risk) |
| **Downgrade** | Possible | Encrypted Client Hello |
| **Latency** | Higher | ~50% lower |

### Encryption at Rest Options

| Approach | Pros | Cons |
|----------|------|------|
| **Application-level** | Full control, E2E | Key management burden |
| **Database TDE** | Transparent | Keys in DB layer |
| **Disk encryption** | OS-level, broad | Key in memory |
| **Cloud KMS** | Managed, audit | Vendor lock-in |

---

## 7. Variants/Implementations

### Perfect Forward Secrecy (PFS)

- **Ephemeral keys**: Each session uses unique key pair
- **Compromise**: Past sessions remain secure even if long-term key leaked
- **TLS 1.3**: PFS mandatory (only ECDHE)
- **TLS 1.2**: Use ECDHE cipher suites, not RSA key exchange

### Envelope Encryption

```
1. Generate Data Encryption Key (DEK) — random, per-object
2. Encrypt data with DEK (fast, symmetric)
3. Encrypt DEK with Master Key (in KMS)
4. Store: encrypted_data + encrypted_DEK
5. Decrypt: KMS decrypts DEK → use DEK to decrypt data
```

**Why**: Master key never leaves KMS; DEKs can be cached; rotate DEK without re-encrypting all data.

### Certificate Validation Chain

```
Leaf Certificate (server.example.com)
    ↓ signed by
Intermediate CA (Let's Encrypt R3)
    ↓ signed by
Root CA (ISRG Root X1)
    ↓ in trust store (OS/browser)
```

**Validation**: Verify signature chain, check expiry, check hostname (CN/SAN), check revocation (CRL/OCSP).

### End-to-End Encryption (Signal Protocol)

- **Double Ratchet**: Combines Diffie-Hellman ratchet + symmetric key ratchet
- **Forward secrecy**: Compromise of long-term key doesn't reveal past messages
- **Post-compromise security**: New keys after compromise
- **Used by**: Signal, WhatsApp, Messenger (optional)

---

## 8. Scaling Strategies

- **TLS termination**: Offload at load balancer (AWS ALB, nginx) — centralize cert management
- **Session resumption**: TLS session tickets or PSK — skip full handshake for returning clients
- **Certificate caching**: Store certs at edge (CDN) — reduce origin load
- **KMS**: Use regional endpoints, cache DEKs in application
- **Hardware**: AES-NI, cryptographic accelerators — 10x throughput

---

## 9. Failure Scenarios

| Failure | Impact | Mitigation |
|---------|--------|------------|
| **Certificate expiry** | TLS fails, site unreachable | Auto-renewal (certbot, ACM) |
| **Private key leak** | Attacker can impersonate | Revoke cert, reissue |
| **Weak cipher** | Downgrade attack | TLS 1.3, strict cipher config |
| **KMS down** | Cannot decrypt data | Multi-region, caching |
| **CRL/OCSP unavailable** | Revocation check fails | Soft-fail, OCSP stapling |
| **Clock skew** | Cert validation fails | NTP, sync time |
| **Heartbleed** | Private key leak | Patch, rotate certs |

---

## 10. Performance Considerations

- **TLS overhead**: ~1-2% CPU for bulk transfer; handshake dominates for short connections
- **Session resumption**: Saves full handshake — use for API with many short requests
- **OCSP stapling**: Server includes OCSP response — avoids client-side OCSP fetch
- **False start**: (TLS 1.2) Send app data before Finished — reduces latency, subtle security tradeoff
- **0-RTT (TLS 1.3)**: Replay risk — use only for idempotent operations

---

## 11. Use Cases

| Use Case | Solution |
|----------|----------|
| HTTPS website | TLS 1.3, Let's Encrypt |
| API security | TLS + JWT/OAuth |
| Database at rest | TDE, AWS RDS encryption |
| Password storage | bcrypt/Argon2, salt |
| Service mesh | mTLS (Istio, Linkerd) |
| File encryption | AES-256-GCM, envelope |
| E2E messaging | Signal protocol |
| Code signing | RSA/ECC signatures |

---

## 12. Comparison Tables

### Cipher Suite Security Levels

| Suite | Key Exchange | Bulk Cipher | Security |
|-------|--------------|-------------|----------|
| TLS_ECDHE_RSA_AES_256_GCM | ECDHE (PFS) | AES-256-GCM | Strong |
| TLS_ECDHE_RSA_AES_128_GCM | ECDHE (PFS) | AES-128-GCM | Strong |
| TLS_RSA_AES_256_CBC | RSA (no PFS) | AES-256-CBC | Weak (no PFS) |
| TLS_RSA_3DES_EDE_CBC | RSA | 3DES | Deprecated |

### KMS Comparison

| Service | Envelope Encryption | Key Rotation | HSM | Integration |
|---------|---------------------|--------------|-----|-------------|
| AWS KMS | ✓ | Automatic | ✓ | Native AWS |
| GCP KMS | ✓ | ✓ | ✓ | Native GCP |
| Azure Key Vault | ✓ | ✓ | ✓ | Native Azure |
| HashiCorp Vault | ✓ | ✓ | ✓ | Multi-cloud |

### Password Hashing Algorithms

| Algorithm | Salt | Adaptive | Use Case |
|-----------|------|----------|----------|
| bcrypt | ✓ | Cost factor | Passwords |
| Argon2 | ✓ | Memory + time | Passwords (preferred) |
| scrypt | ✓ | Memory + time | Passwords |
| PBKDF2 | ✓ | Iterations | Legacy |
| MD5/SHA (plain) | ✗ | ✗ | Never for passwords |

---

## 13. Code/Pseudocode

### TLS Handshake (Conceptual)

```python
# Client side - simplified
def tls_connect(host, port):
    # 1. TCP connect
    sock = connect(host, port)
    
    # 2. Client Hello
    client_hello = build_client_hello(
        random=os.urandom(32),
        cipher_suites=[TLS_ECDHE_RSA_AES_256_GCM, ...],
        extensions=[sni(host), supported_groups]
    )
    send(sock, client_hello)
    
    # 3. Receive Server Hello, Certificate, Key Exchange
    server_hello = recv(sock)
    server_cert = recv(sock)
    server_key_exchange = recv(sock)
    
    # 4. Verify certificate
    verify_certificate_chain(server_cert, host)
    check_revocation(server_cert)  # OCSP
    
    # 5. Compute shared secret (ECDHE)
    shared_secret = ecdhe_client(our_private_key, server_public_key)
    master_secret = prf(shared_secret, "master secret", client_random + server_random)
    
    # 6. Derive keys
    key_block = prf(master_secret, "key expansion", server_random + client_random)
    client_write_key, server_write_key, client_iv, server_iv = split(key_block)
    
    # 7. Change cipher spec, Finished
    send_encrypted(sock, finished_message, client_write_key)
    verify_encrypted(sock, server_finished, server_write_key)
    
    return TLSConnection(sock, client_write_key, server_write_key)
```

### Envelope Encryption (Pseudocode)

```python
def encrypt_with_envelope(plaintext: bytes, kms_client) -> tuple[bytes, bytes]:
    # 1. Generate DEK
    dek = os.urandom(32)  # 256-bit AES key
    
    # 2. Encrypt data with DEK
    cipher = AES.new(dek, AES.MODE_GCM)
    ciphertext, tag = cipher.encrypt_and_digest(plaintext)
    
    # 3. Encrypt DEK with KMS
    encrypted_dek = kms_client.encrypt(
        KeyId='alias/my-key',
        Plaintext=dek
    )
    
    return ciphertext + tag, encrypted_dek

def decrypt_with_envelope(encrypted_data: bytes, encrypted_dek: bytes, kms_client) -> bytes:
    # 1. Decrypt DEK with KMS
    dek = kms_client.decrypt(CiphertextBlob=encrypted_dek)['Plaintext']
    
    # 2. Decrypt data with DEK
    ciphertext, tag = encrypted_data[:-16], encrypted_data[-16:]
    cipher = AES.new(dek, AES.MODE_GCM, nonce=...)
    return cipher.decrypt_and_verify(ciphertext, tag)
```

### Password Hashing (bcrypt)

```python
import bcrypt

def hash_password(password: str) -> str:
    salt = bcrypt.gensalt(rounds=12)  # Cost factor 12
    return bcrypt.hashpw(password.encode(), salt).decode()

def verify_password(password: str, hashed: str) -> bool:
    return bcrypt.checkpw(password.encode(), hashed.encode())
```

---

## 14. Interview Discussion

### Key Points to Articulate

1. **Symmetric vs Asymmetric**: "Symmetric is fast for bulk data — AES does GB/s. Asymmetric is for key exchange and signatures — we use it to establish a shared secret, then switch to symmetric for the actual encryption."

2. **TLS Handshake**: "Client sends supported ciphers; server picks one and sends cert. They do key exchange (ECDHE for PFS). Both derive session keys. The handshake is 1-2 RTT; TLS 1.3 reduced it to 1 RTT."

3. **Perfect Forward Secrecy**: "With ephemeral Diffie-Hellman, each session has unique keys. If the server's long-term private key is compromised later, past sessions can't be decrypted."

4. **Envelope Encryption**: "We encrypt data with a random DEK, then encrypt the DEK with the master key in KMS. The master key never leaves KMS. We can rotate DEKs without touching the master key."

5. **Certificate Validation**: "Verify the chain to a trusted root, check hostname in CN or SAN, verify not expired, and check revocation via OCSP or CRL."

### Common Follow-ups

- **"Why not use RSA for everything?"** — RSA is 1000x slower than AES. We use it only for key exchange and signatures; bulk encryption uses symmetric.

- **"What's the risk of TLS 1.3 0-RTT?"** — Replay: attacker could replay a 0-RTT request. Use only for idempotent operations (GET, or ensure server deduplicates).

- **"How does mTLS differ from regular TLS?"** — Both sides present certificates. Server requests client cert; client sends it and proves possession of private key. Used for service-to-service in zero-trust.

- **"How would you rotate encryption keys?"** — Envelope encryption: generate new DEK, re-encrypt data, encrypt new DEK with same or new master key. Master key rotation: create new key, re-encrypt all DEKs.
