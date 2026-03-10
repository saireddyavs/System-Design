# Secrets Management

## Staff+ Engineer Deep Dive | FAANG Interview Preparation

---

## 1. Concept Overview

**Secrets** are sensitive credentials that grant access to systems, data, or services. **Secrets management** is the practice of securely storing, accessing, rotating, and auditing these credentials throughout their lifecycle.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    WHAT ARE SECRETS?                                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   CREDENTIALS              KEYS & CERTS              TOKENS                 │
│   ───────────              ─────────────             ──────                 │
│   • Database passwords     • TLS certificates       • API keys             │
│   • SSH keys               • Encryption keys         • OAuth tokens         │
│   • Cloud IAM credentials  • Signing keys            • JWT signing keys     │
│   • Third-party API creds  • PGP keys                 • Webhook secrets      │
│                                                                             │
│   LIFECYCLE: Create → Store → Distribute → Use → Rotate → Revoke            │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Core Principle**: Secrets should never appear in source code, logs, or environment variables in plaintext. They require dedicated, audited storage with access controls.

---

## 2. Real-World Motivation

- **GitHub 2022 breach**: Private keys in code → attacker accessed npm packages
- **Uber 2016**: AWS credentials in GitHub → 57M user records exposed
- **Equifax 2017**: Default credentials on database → 147M records
- **Why not .env?**: Committed to git, visible in process list, no rotation, no audit
- **Compliance**: PCI DSS, HIPAA, SOC 2 require secrets management controls

---

## 3. Architecture Diagrams (ASCII)

### HashiCorp Vault Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                    VAULT-BASED SECRETS MANAGEMENT                                    │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                     │
│   ┌─────────────┐     ┌─────────────┐     ┌─────────────────────────────────────┐  │
│   │  Application│     │ Vault Agent │     │         HashiCorp Vault              │  │
│   │  (Pod/VM)   │     │  (Sidecar)  │     │                                     │  │
│   └──────┬──────┘     └──────┬──────┘     │  ┌─────────┐  ┌─────────────────┐  │  │
│          │                   │            │  │  Auth   │  │  Secret Engines  │  │  │
│          │ 1. Request        │            │  │  Engine │  │  • KV (static)   │  │  │
│          │    secrets       │            │  │  • K8s  │  │  • Database      │  │  │
│          │────────────────>│            │  │  • AppRole│  │  • AWS (dynamic) │  │  │
│          │                   │            │  │  • JWT  │  │  • PKI           │  │  │
│          │                   │ 2. Auth    │  └────┬────┘  └────────┬────────┘  │  │
│          │                   │ (K8s SA)   │       │                │           │  │
│          │                   │───────────>│       │                │           │  │
│          │                   │            │       v                v           │  │
│          │                   │ 3. Secrets │  ┌─────────────────────────────────┐ │  │
│          │                   │<───────────│  │  Lease Manager (TTL, renewal)   │ │  │
│          │                   │            │  └─────────────────────────────────┘ │  │
│          │ 4. Secrets        │            │                                     │  │
│          │    (file/env)     │            └─────────────────────────────────────┘  │
│          │<──────────────────│                         │                          │
│          │                   │                         │                          │
│          v                   v                         v                          │
│   ┌─────────────────────────────────────────────────────────────────────────┐    │
│   │  Secrets written to: /vault/secrets/*.json or env vars                   │    │
│   │  Auto-renewal: Vault Agent renews before lease expiry                   │    │
│   └─────────────────────────────────────────────────────────────────────────┘    │
│                                                                                     │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

### Dynamic Secrets Flow (Database)

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Application │     │    Vault     │     │   Database   │     │  Vault DB    │
│              │     │              │     │   (MySQL)     │     │  Plugin      │
└──────┬───────┘     └──────┬───────┘     └──────┬───────┘     └──────┬───────┘
       │                   │                   │                    │
       │ 1. Get DB creds    │                   │                    │
       │──────────────────>│                   │                    │
       │                   │                   │                    │
       │                   │ 2. Create user     │                    │
       │                   │    (if needed)    │                    │
       │                   │──────────────────>│                    │
       │                   │                   │ 3. CREATE USER     │
       │                   │                   │    (Vault plugin)  │
       │                   │                   │<───────────────────│
       │                   │                   │                    │
       │                   │ 4. Return creds    │                    │
       │                   │    + lease (1h)    │                    │
       │                   │<───────────────────│                    │
       │                   │                   │                    │
       │ 5. username,       │                   │                    │
       │    password       │                   │                    │
       │<──────────────────│                   │                    │
       │                   │                   │                    │
       │ 6. Connect to DB  │                   │                    │
       │──────────────────────────────────────>│                    │
       │                   │                   │                    │
       │                   │ 7. Before lease    │ 8. DROP USER       │
       │                   │    expiry: renew  │    (auto-cleanup)  │
       │                   │    or let expire  │<───────────────────│
       │                   │                   │                    │
```

### Envelope Encryption with KMS

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    ENVELOPE ENCRYPTION                                           │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│   Application                    AWS KMS / Vault Transit                        │
│   ┌─────────────────┐           ┌─────────────────────────────┐                │
│   │                 │           │  Customer Master Key (CMK)  │                │
│   │  Data (plain)   │           │  Never leaves HSM           │                │
│   │        │        │           └──────────────┬──────────────┘                │
│   │        v        │                          │                                │
│   │  Generate DEK   │                          │                                │
│   │  (Data Enc Key) │                          │                                │
│   │        │        │                          │                                │
│   │        v        │           Encrypt DEK    │                                │
│   │  Encrypt data   │<──────────with CMK───────┘                                │
│   │  with DEK       │                                                           │
│   │        │        │                                                           │
│   └────────┼────────┘                                                           │
│            v                                                                     │
│   Store: encrypted_data + encrypted_DEK                                         │
│                                                                                 │
│   Decrypt: KMS decrypts DEK → Application decrypts data with DEK                │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Core Mechanics

### Why Not Hardcode or .env?

| Approach | Problems |
|----------|----------|
| **Hardcoded** | In git history forever, visible to anyone with repo access |
| **.env file** | Often committed, no rotation, no audit, process list exposure |
| **Environment variables** | Visible in `/proc`, no rotation, scattered across configs |
| **Config files** | Same as .env, often in version control |

### Vault Secret Engines

| Engine | Type | Use Case |
|--------|------|----------|
| **KV (Key-Value)** | Static | API keys, config secrets |
| **Database** | Dynamic | Short-lived DB credentials |
| **AWS** | Dynamic | Temporary IAM credentials |
| **PKI** | Dynamic | TLS certificates |
| **Transit** | Encryption | Encrypt/decrypt without storing keys |
| **SSH** | Dynamic | One-time SSH keys |

### Dynamic Secrets Benefits

- **Short-lived**: Credentials expire (e.g., 1 hour)
- **Just-in-time**: Created when requested
- **Auto-revocation**: Deleted when lease expires
- **Reduced blast radius**: Compromised cred has limited validity

---

## 5. Numbers

| Metric | Value | Context |
|--------|-------|---------|
| Vault lease (DB) | 1-24 hours | Typical dynamic secret |
| Vault lease (AWS) | 15 min - 1 hr | IAM credentials |
| Secret rotation | 90 days | Compliance (PCI DSS) |
| KMS API latency | 1-10 ms | Per encrypt/decrypt |
| Vault unseal | ~1 sec | Per key share |
| Kubernetes secret | Base64 (not encrypted) | Must use external encryption |
| SOPS overhead | ~100 ms | Encrypt/decrypt file |

---

## 6. Tradeoffs (Comparison Tables)

### Secrets Management Tools

| Tool | Dynamic Secrets | Rotation | K8s Native | Multi-Cloud | Cost |
|------|-----------------|----------|------------|-------------|------|
| **HashiCorp Vault** | ✓ | ✓ | CSI, Agent | ✓ | OSS/Enterprise |
| **AWS Secrets Manager** | Limited | ✓ | External | AWS only | Per secret |
| **AWS SSM Parameter Store** | No | ✓ | External | AWS only | Free/Paid |
| **GCP Secret Manager** | No | ✓ | Workload Identity | GCP only | Per version |
| **Azure Key Vault** | No | ✓ | CSI | Azure only | Per operation |
| **SOPS** | No | Manual | GitOps | ✓ | Free |

### Storage Comparison

| Approach | Encryption | Audit | Rotation | Distribution |
|----------|------------|-------|----------|--------------|
| **Vault** | ✓ | ✓ | ✓ | API, Agent, CSI |
| **Secrets Manager** | ✓ | ✓ | ✓ | SDK, Lambda |
| **K8s Secrets** | At rest (etcd) | Limited | Manual | Volume/env |
| **SOPS** | ✓ (file-level) | Git history | Manual | Git |
| **.env** | ✗ | ✗ | ✗ | Copy |

---

## 7. Variants/Implementations

### Vault Agent Injector (Kubernetes)

```yaml
# Pod annotation triggers injection
annotations:
  vault.hashicorp.com/agent-inject: "true"
  vault.hashicorp.com/role: "myapp"
  vault.hashicorp.com/agent-inject-secret-db: "secret/data/myapp/db"
  vault.hashicorp.com/agent-inject-template-db: |
    {{- with secret "secret/data/myapp/db" -}}
    export DB_PASSWORD="{{ .Data.data.password }}"
    {{- end -}}
```

- **Init container**: Authenticates, fetches secrets, writes to shared volume
- **App container**: Reads secrets from `/vault/secrets/`
- **No code changes**: Secrets appear as files

### Vault CSI Driver

- **Mounts secrets as volumes**: No init container
- **On-demand**: Fetches when pod starts
- **Auto-renewal**: Volume content updated before expiry
- **Use case**: Apps that read from file, not env

### SOPS (Mozilla)

```
# Encrypt file with KMS key
sops --encrypt --kms arn:aws:kms:us-east-1:123:key/abc config.yaml > config.enc.yaml

# File structure: encrypted values, plaintext keys (for readability)
# Decrypt at deploy time (CI/CD has KMS access)
sops --decrypt config.enc.yaml
```

- **Git-friendly**: Encrypted files in repo
- **Partial encryption**: Only values encrypted, keys visible
- **Key management**: KMS, PGP, age

### Kubernetes Secrets (Reality Check)

```yaml
# Base64 is NOT encryption - it's encoding
apiVersion: v1
kind: Secret
data:
  password: c3VwZXJzZWNyZXQ=  # "supersecret" in base64
```

- **etcd encryption**: Can enable encryption at rest for etcd
- **RBAC**: Control who can read secrets
- **Still base64**: Anyone with `kubectl get secret` sees it
- **Best practice**: Use external secrets operator or CSI to sync from Vault

---

## 8. Scaling Strategies

- **Vault cluster**: 3-5 nodes, Raft consensus, performance replication
- **Caching**: Vault Agent caches secrets, reduces API calls
- **Lease tuning**: Longer leases = fewer renewals, shorter = more secure
- **Read replicas**: Vault performance replicas for read scaling
- **Namespace isolation**: Vault namespaces for multi-tenant

---

## 9. Failure Scenarios

| Failure | Impact | Mitigation |
|---------|--------|------------|
| **Vault unavailable** | Apps can't get secrets | Caching (Agent), fallback to cached |
| **Vault sealed** | No secrets access | Unseal with threshold of key shares |
| **KMS down** | Can't decrypt | Multi-region KMS, caching |
| **Lease expiry** | App loses DB access | Auto-renewal (Agent), health checks |
| **Secret leak** | Credential compromise | Immediate rotation, short leases |
| **Wrong secret version** | App misconfiguration | Version pinning, canary deploys |

---

## 10. Performance Considerations

- **Vault Agent**: Reduces per-request Vault calls — fetch once, cache
- **Batch requests**: Vault supports batch read for multiple secrets
- **Lease duration**: Balance security (short) vs renewal load (long)
- **KMS**: Cache DEKs in application; only call KMS for new data
- **Connection pooling**: Reuse Vault client connections

---

## 11. Use Cases

| Use Case | Solution |
|----------|----------|
| **Database credentials** | Vault Database engine (dynamic) |
| **AWS access** | Vault AWS engine or IAM roles |
| **TLS certs** | Vault PKI, cert-manager + Vault |
| **API keys** | Vault KV, Secrets Manager |
| **Encryption keys** | KMS, Vault Transit |
| **GitOps secrets** | SOPS, Sealed Secrets |
| **Kubernetes** | CSI driver, External Secrets Operator |
| **Legacy apps** | Sidecar injector, env from file |

---

## 12. Comparison Tables

### When to Use Which Tool

| Scenario | Recommended |
|----------|-------------|
| **AWS-native app** | Secrets Manager or SSM |
| **Multi-cloud** | HashiCorp Vault |
| **Kubernetes** | Vault + CSI or ESO |
| **GitOps** | SOPS or Sealed Secrets |
| **Encryption keys** | KMS (cloud) or Vault Transit |
| **Dynamic DB creds** | Vault Database engine |

### Secret Rotation Strategies

| Strategy | How | Use Case |
|----------|-----|----------|
| **Time-based** | Rotate every N days | Compliance (PCI 90 days) |
| **On-demand** | Manual or event-triggered | Incident response |
| **Versioned** | New version, gradual migration | Zero-downtime |
| **Dynamic** | Never store — generate on demand | Vault dynamic secrets |

---

## 13. Code/Pseudocode

### Vault Client (Pseudocode)

```python
import hvac

# 1. Authenticate (e.g., Kubernetes auth)
client = hvac.Client(url='https://vault:8200')
client.auth_kubernetes(role='myapp', jwt=open('/var/run/secrets/kubernetes.io/serviceaccount/token').read())

# 2. Read static secret
secret = client.secrets.kv.v2.read_secret_version(path='myapp/db')
db_password = secret['data']['data']['password']

# 3. Read dynamic secret (database)
secret = client.secrets.database.read_credentials(name='myapp-mysql')
username = secret['data']['username']
password = secret['data']['password']
lease_id = secret['lease_id']  # Renew before expiry

# 4. Renew lease
client.sys.renew_lease(lease_id)
```

### Vault Agent Config (HCL)

```hcl
vault {
  address = "https://vault:8200"
}

auto_auth {
  method "kubernetes" {
    config = {
      role = "myapp"
    }
  }

  sink "file" {
    config = {
      path = "/vault/token"
    }
  }
}

template_config {
  exit_on_retry_failure = true
}

template {
  destination = "/vault/secrets/db.env"
  contents    = <<EOT
{{- with secret "database/creds/myapp" }}
DB_USERNAME="{{ .Data.username }}"
DB_PASSWORD="{{ .Data.password }}"
{{ end }}
EOT
}
```

### AWS Secrets Manager Rotation (Lambda)

```python
# Rotation Lambda - invoked by Secrets Manager
def lambda_handler(event, context):
    if event['Step'] == 'createSecret':
        # Generate new password
        new_secret = generate_password()
        put_secret_value(SecretId=event['SecretId'], SecretString=new_secret)
    elif event['Step'] == 'setSecret':
        # Update DB with new password
        update_db_password(event['SecretId'], event['SecretString'])
    elif event['Step'] == 'testSecret':
        # Verify new creds work
        test_db_connection(event['SecretString'])
    elif event['Step'] == 'finishSecret':
        # Mark rotation complete
        pass
```

### SOPS Encrypt/Decrypt

```bash
# Encrypt with AWS KMS
sops -e -k "arn:aws:kms:us-east-1:123456789:key/abc-123" secrets.yaml > secrets.enc.yaml

# Decrypt (requires KMS access)
sops -d secrets.enc.yaml

# Edit in place (decrypts, edits, re-encrypts)
sops secrets.enc.yaml
```

---

## 14. Interview Discussion

### Key Points to Articulate

1. **Why not .env?** — "Environment variables are visible in process listings, often end up in logs, and provide no rotation or audit trail. For production, we need a dedicated secrets store with access control and encryption."

2. **Dynamic vs Static Secrets** — "Dynamic secrets are generated on demand and expire. A compromised credential has a 1-hour window vs forever. Vault's Database engine creates a DB user when you request it and drops it when the lease expires."

3. **Envelope Encryption** — "We never send bulk data to KMS. We generate a DEK locally, encrypt data with it, then encrypt only the DEK with the master key. The master key stays in KMS; we can cache DEKs for performance."

4. **Vault Architecture** — "Vault has secret engines (KV, Database, AWS), auth methods (Kubernetes, AppRole), and a lease manager. The Agent runs as a sidecar, authenticates with K8s service account, fetches secrets, and writes them to a shared volume for the app."

5. **Kubernetes Secrets** — "Base64 is encoding, not encryption. Anyone with kubectl access can decode. We use External Secrets Operator or Vault CSI to sync from a real secrets store, so secrets never live in Git or etcd long-term."

### Common Follow-ups

- **"How would you implement secret rotation with zero downtime?"** — Versioned secrets: create new version, update app to support both, deploy, then remove old. Or use dynamic secrets that auto-rotate.

- **"What if Vault goes down?"** — Vault Agent caches secrets and tokens. Apps can run on cached secrets for the lease duration. For critical systems, run Vault cluster with multiple nodes.

- **"How do you get secrets to a new pod?"** — Init container or CSI: before app starts, Vault Agent fetches secrets and writes to volume. App reads from file. No secrets in image or env at build time.

- **"SOPS vs Vault?"** — SOPS for GitOps: secrets in Git, encrypted. Good for config that needs to be in repo. Vault for runtime: dynamic secrets, audit, rotation. Use both: SOPS for bootstrap, Vault for runtime.
