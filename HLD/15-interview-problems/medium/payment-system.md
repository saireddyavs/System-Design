# Design a Payment System

## 1. Problem Statement & Requirements

### Problem Statement
Design a payment system that processes payments, refunds, supports multiple payment methods (credit card, bank transfer, digital wallets), maintains transaction history, and handles settlement with payment service providers (PSPs).

### Functional Requirements
- **Process payments**: One-time and recurring payments
- **Refunds**: Full and partial refunds
- **Payment methods**: Credit/debit card, bank transfer (ACH/SEPA), digital wallets (PayPal, Apple Pay)
- **Transaction history**: Searchable, filterable history per user/merchant
- **Settlement**: Batch settlement to merchant accounts
- **Webhooks**: Notify merchants of payment status changes

### Non-Functional Requirements
- **Scale**: Millions of transactions per day
- **Consistency**: Strong consistency for balance updates; eventual for reporting
- **Security**: PCI DSS compliant; no raw card data stored
- **Availability**: 99.99% (payments are critical)
- **Idempotency**: Prevent double charges on retries

### Out of Scope
- Card issuing (we're the acquirer/processor)
- Cryptocurrency payments
- Multi-currency FX (assume single currency or simple conversion)
- Dispute/chargeback automation (manual process)

---

## 2. Back-of-Envelope Estimation

### Scale Assumptions
- **Transactions**: 10M/day = ~115 QPS average, 500 QPS peak
- **Refunds**: 2% of transactions = 200K/day
- **Merchants**: 100K
- **Users**: 50M (customers making payments)

### QPS Calculation
| Operation | Daily Volume | Peak QPS |
|-----------|--------------|----------|
| Payment (authorize/capture) | 10M | ~120 |
| Refund | 200K | ~25 |
| Transaction lookup | 50M | ~600 |
| Webhook delivery | 12M | ~140 |
| Settlement (batch) | 100K batches | ~2 |

### Storage (5 years)
- **Transactions**: 10M × 365 × 5 × 500B ≈ 9 TB
- **Ledger entries**: 2 entries per tx (debit/credit) × 2 ≈ 18 TB
- **Merchant data**: 100K × 10KB ≈ 1 GB
- **Payment methods (tokens)**: 50M × 200B ≈ 10 GB

### Bandwidth
- **API**: 10M × 2KB request + 1KB response ≈ 30 GB/day
- **Webhooks**: 12M × 500B ≈ 6 GB/day
- **PSP calls**: 10M × 1KB ≈ 10 GB/day

### Cache
- **Merchant config**: 100K × 2KB ≈ 200 MB
- **Token lookup**: 50M × 100B ≈ 5 GB (hot tokens only)
- **Idempotency keys**: 24h TTL, 10M × 50B ≈ 500 MB

---

## 3. API Design

### REST Endpoints

```
# Payment Methods (Tokenization)
POST   /api/v1/payment_methods
Body: { "type": "card", "card": { "number", "exp_month", "exp_year", "cvc" } }
      OR { "type": "bank_account", "bank_account": { ... } }
Response: { "payment_method_id": "pm_xxx", "last4": "4242" }
Note: Raw card data never stored; tokenized by PSP

# Payments
POST   /api/v1/payments
Body: {
  "amount": 1000,           // cents
  "currency": "usd",
  "payment_method_id": "pm_xxx",
  "merchant_id": "m_xxx",
  "idempotency_key": "unique_key_123",
  "metadata": { "order_id": "ord_123" },
  "capture": true          // true = charge immediately, false = auth only
}
Response: { "payment_id": "pay_xxx", "status": "succeeded" | "pending" | "failed" }

POST   /api/v1/payments/:id/capture
Body: { "amount": 1000 }   // partial capture
Response: { "payment_id", "status": "succeeded" }

POST   /api/v1/payments/:id/cancel
Response: { "payment_id", "status": "canceled" }

# Refunds
POST   /api/v1/refunds
Body: {
  "payment_id": "pay_xxx",
  "amount": 500,           // partial refund
  "reason": "requested_by_customer",
  "idempotency_key": "ref_123"
}
Response: { "refund_id": "ref_xxx", "status": "succeeded" }

# Transaction History
GET    /api/v1/transactions
Query: merchant_id, customer_id, status, from_date, to_date, limit, cursor
Response: { "transactions": [...], "next_cursor": "..." }

GET    /api/v1/transactions/:id
Response: { "transaction_id", "type", "amount", "status", "created_at", ... }

# Ledger (Internal / Merchant Dashboard)
GET    /api/v1/ledger/balance
Query: merchant_id, as_of_date
Response: { "available_balance", "pending_balance" }

GET    /api/v1/ledger/entries
Query: merchant_id, from_date, to_date
Response: { "entries": [...] }

# Webhooks (Merchant registers)
POST   /api/v1/webhooks
Body: { "url": "https://merchant.com/webhook", "events": ["payment.succeeded", "refund.completed"] }
Response: { "webhook_id", "secret" }

# Settlement (Internal / Scheduled)
POST   /api/v1/settlements
Body: { "merchant_id", "period_start", "period_end" }
Response: { "settlement_id", "status": "pending" }
```

---

## 4. Data Model / Database Schema

### Database Choice
- **Transactions**: PostgreSQL (ACID, strong consistency)
- **Ledger**: PostgreSQL (double-entry, critical for correctness)
- **Merchants**: PostgreSQL
- **Idempotency**: Redis (fast lookup) + PostgreSQL (audit)
- **Webhooks**: PostgreSQL + Kafka (async delivery)
- **Audit log**: Append-only store (S3 + Athena or ClickHouse)

### Schema

**Merchants (PostgreSQL)**
```sql
merchants (
  merchant_id UUID PRIMARY KEY,
  name VARCHAR(200),
  psp_account_id VARCHAR(100),   -- Stripe/Adyen account
  default_currency VARCHAR(3),
  webhook_secret VARCHAR(255),
  created_at TIMESTAMP,
  updated_at TIMESTAMP
)
```

**Payment Methods (PostgreSQL)**
```sql
payment_methods (
  payment_method_id UUID PRIMARY KEY,
  customer_id UUID,
  type VARCHAR(20),              -- card, bank_account, wallet
  psp_token VARCHAR(255),        -- PSP's token (we never store raw card)
  last4 VARCHAR(4),
  brand VARCHAR(20),             -- visa, mastercard
  exp_month INT,
  exp_year INT,
  created_at TIMESTAMP
)
```

**Transactions (PostgreSQL)**
```sql
transactions (
  transaction_id UUID PRIMARY KEY,
  idempotency_key VARCHAR(255) UNIQUE,
  type VARCHAR(20),              -- payment, refund, adjustment
  status VARCHAR(20),            -- pending, succeeded, failed
  amount_cents BIGINT,
  currency VARCHAR(3),
  merchant_id UUID,
  customer_id UUID,
  payment_id UUID,              -- for refunds, links to original payment
  psp_reference VARCHAR(255),
  metadata JSONB,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
)
```

**Ledger (Double-Entry) (PostgreSQL)**
```sql
ledger_entries (
  entry_id BIGINT PRIMARY KEY,
  transaction_id UUID,
  account_id UUID,               -- merchant settlement account
  type VARCHAR(20),              -- debit, credit
  amount_cents BIGINT,
  balance_after BIGINT,
  created_at TIMESTAMP
)

CREATE INDEX idx_ledger_account_created ON ledger_entries(account_id, created_at);
```

**Idempotency (Redis + PostgreSQL)**
```sql
idempotency_keys (
  idempotency_key VARCHAR(255) PRIMARY KEY,
  transaction_id UUID,
  response_hash VARCHAR(64),      -- hash of response for replay
  created_at TIMESTAMP,
  expires_at TIMESTAMP
)
```

**Webhooks (PostgreSQL)**
```sql
webhooks (
  webhook_id UUID PRIMARY KEY,
  merchant_id UUID,
  url VARCHAR(500),
  events TEXT[],                 -- ['payment.succeeded', ...]
  secret VARCHAR(255),
  created_at TIMESTAMP
)

webhook_deliveries (
  delivery_id UUID PRIMARY KEY,
  webhook_id UUID,
  event_type VARCHAR(50),
  payload JSONB,
  status VARCHAR(20),            -- pending, delivered, failed
  attempts INT,
  next_retry_at TIMESTAMP,
  created_at TIMESTAMP
)
```

---

## 5. High-Level Architecture

### ASCII Architecture Diagram

```
                                    ┌─────────────────────────────────────────────────────────────┐
                                    │                    MERCHANT / CUSTOMER                       │
                                    │  (E-commerce site, Mobile app)                              │
                                    └───────────────────────────┬─────────────────────────────────┘
                                                                  │
                                                                  │ HTTPS (Tokenization, Payment API)
                                                                  ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                         API GATEWAY                                                          │
│                                    (Rate limit, Auth, TLS)                                                   │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
                                                                  │
                                                                  ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                         PAYMENT SERVICE                                                       │
│                                                                                                               │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐                  │
│  │ Idempotency     │    │ Payment         │    │ Refund           │    │ Ledger          │                  │
│  │ Check (Redis)   │───▶│ Orchestrator    │───▶│ Service         │───▶│ Service         │                  │
│  └─────────────────┘    └────────┬────────┘    └─────────────────┘    └────────┬────────┘                  │
│                                  │                                                      │                    │
│                                  │                                                      │                    │
│                                  ▼                                                      ▼                    │
│                         ┌─────────────────┐                                    ┌─────────────────┐          │
│                         │ PSP Adapter      │                                    │ PostgreSQL       │          │
│                         │ (Stripe, Adyen) │                                    │ (Ledger, Txns)  │          │
│                         └────────┬────────┘                                    └─────────────────┘          │
│                                  │                                                                           │
└──────────────────────────────────┼───────────────────────────────────────────────────────────────────────────┘
                                   │
                                   │ PSP API (Auth, Capture, Refund)
                                   ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                    PAYMENT SERVICE PROVIDER (PSP)                                            │
│                                    (Stripe, Adyen, Braintree)                                                 │
│                                    - Card network (Visa, MC)                                                  │
│                                    - Bank (ACH, SEPA)                                                         │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                    PAYMENT FLOW (Auth → Capture → Settlement)                                 │
│                                                                                                               │
│  Merchant          Payment Service              PSP                    Ledger                                 │
│     │                     │                      │                      │                                    │
│     │  POST /payments     │                      │                      │                                    │
│     │────────────────────▶                      │                      │                                    │
│     │                     │  Check idempotency   │                      │                                    │
│     │                     │─────────────────────│                      │                                    │
│     │                     │  (Redis)             │                      │                                    │
│     │                     │                      │                      │                                    │
│     │                     │  Auth/Capture        │                      │                                    │
│     │                     │─────────────────────▶                      │                                    │
│     │                     │                      │                      │                                    │
│     │                     │  Success             │                      │                                    │
│     │                     │◀─────────────────────                      │                                    │
│     │                     │                      │                      │                                    │
│     │                     │  Debit customer     │                      │                                    │
│     │                     │  Credit merchant    │                      │─────────────────────▶               │
│     │                     │──────────────────────────────────────────────────────────────────               │
│     │                     │                      │                      │                                    │
│     │                     │  Publish event       │                      │                                    │
│     │                     │  (Kafka)             │                      │                                    │
│     │                     │─────────────────────│                      │                                    │
│     │                     │                      │                      │                                    │
│     │  200 OK             │                      │                      │                                    │
│     │◀────────────────────                      │                      │                                    │
│     │                     │                      │                      │                                    │
│     │                     │  Webhook delivery    │                      │                                    │
│     │                     │  (async)            │                      │                                    │
│     │  POST /webhook       │                      │                      │                                    │
│     │◀────────────────────                      │                      │                                    │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

### Component Descriptions
- **API Gateway**: Auth, rate limiting, request validation
- **Idempotency Check**: Redis lookup; return cached response if key exists
- **Payment Orchestrator**: Coordinates auth, capture, ledger, webhook
- **PSP Adapter**: Abstraction over Stripe/Adyen; retry logic
- **Ledger Service**: Double-entry bookkeeping; balance calculation
- **Webhook Service**: Async delivery with retries

---

## 6. Detailed Component Design

### 6.1 Payment Flow (Auth → Capture → Settlement)

**Authorization**
- Reserve funds on customer's card (hold)
- PSP returns auth_id; valid for 7 days (card) or 30 days (some)
- No money moved yet

**Capture**
- **Immediate capture**: Auth + capture in one PSP call (most e-commerce)
- **Delayed capture**: Auth now, capture later (e.g., when ship)
- **Partial capture**: Capture less than auth amount
- **Auto-void**: Release auth if not captured in time

**Settlement**
- PSP batches transactions; settles to merchant bank account (T+1, T+2)
- We maintain internal ledger; settlement is our payout to merchant

### 6.2 Idempotency

**Why**
- Network retries, client bugs can cause duplicate requests
- Double charge is unacceptable

**How**
- Client sends `Idempotency-Key: unique_key_123` header
- Server: Check Redis/DB for key
  - **Exists**: Return cached response (same status, same payment_id)
  - **Not exists**: Process; store (key → response) with TTL (24h)
- **Key scope**: Per merchant or global
- **Storage**: Redis for fast lookup; PostgreSQL for audit

### 6.3 PSP Integration

**Adapter Pattern**
- `PSPAdapter` interface: `authorize()`, `capture()`, `refund()`, `void()`
- Implementations: `StripeAdapter`, `AdyenAdapter`
- Config per merchant: which PSP, credentials

**Retry**
- Transient failures (timeout, 5xx): Exponential backoff, max 3 retries
- Idempotent PSP calls: Use idempotency key on PSP side too
- Non-idempotent: Careful; may need to check status before retry

**Error Handling**
- Card declined: Return to client immediately (no retry)
- PSP down: Retry with backoff; queue for async processing if needed

### 6.4 Ledger System (Double-Entry Bookkeeping)

**Principles**
- Every transaction has equal debits and credits
- Balance = Sum(credits) - Sum(debits) per account

**Example: Payment of $100**
- Debit: Customer (or external) - $100
- Credit: Merchant settlement account + $100
- Ledger entries:
  - (customer, debit, 100)
  - (merchant, credit, 100)

**Refund of $50**
- Debit: Merchant - $50
- Credit: Customer + $50

**Balance**
- `available_balance`: Settled funds
- `pending_balance`: Auth but not yet settled
- Query: `SELECT SUM(amount) FROM ledger_entries WHERE account_id=? AND type='credit'` - debits

### 6.5 Reconciliation

- **Daily**: Match our ledger vs PSP settlement report
- **Discrepancies**: Alert; manual investigation
- **Automated**: Parse PSP CSV/API; compare transaction IDs, amounts

### 6.6 Fraud Detection

- **Rules**: Velocity (too many txns), amount thresholds, country mismatch
- **ML**: Anomaly detection on user behavior
- **3DS**: Redirect to bank for authentication (reduces chargebacks)
- **Async**: Don't block payment; flag for review; can refund later

### 6.7 PCI DSS Compliance

- **Never store**: Full card number, CVV
- **Tokenization**: PSP returns token; we store token only
- **Encryption**: TLS in transit; encrypt tokens at rest
- **Scope reduction**: Use hosted fields (Stripe Elements) so card data never touches our servers

### 6.8 Webhook Notifications

- **Events**: payment.succeeded, payment.failed, refund.completed
- **Delivery**: Async; POST to merchant URL with HMAC signature
- **Retry**: Exponential backoff (1min, 5min, 30min, 2h, 12h)
- **Idempotency**: Merchant should handle duplicate events (same event_id)

### 6.9 Eventual Consistency

- **Ledger**: Strong consistency (single DB, transaction)
- **Reporting**: Eventually consistent (replicated to warehouse)
- **Webhooks**: At-least-once; merchant must idempotent

---

## 7. Scaling

### Sharding
- **Transactions**: Shard by merchant_id (merchant's txns together)
- **Ledger**: Same shard as transactions for atomicity
- **Idempotency**: Shard by key hash

### Caching
- **Merchant config**: Redis, 100K merchants
- **Idempotency**: Redis, 24h TTL
- **Balance**: Cache with short TTL; invalidate on write

### Database
- **Read replicas**: For transaction history queries
- **Connection pooling**: PgBouncer
- **Partitioning**: Transactions by created_at (monthly)

### Async
- **Webhooks**: Kafka → Webhook worker
- **Settlement**: Batch job, not real-time
- **Reconciliation**: Daily batch

---

## 8. Failure Handling

### Component Failures
- **PSP down**: Retry with backoff; queue for later; return "temporarily unavailable"
- **Ledger DB down**: Block new payments (consistency over availability)
- **Redis (idempotency) down**: Fallback to DB; slower but correct

### Redundancy
- **Multi-region**: Active-passive for payment service
- **PSP**: Multi-PSP (failover to Adyen if Stripe down)
- **DB**: Primary + sync replica; failover

### Degradation
- **Webhook delivery**: Retry later; don't block payment
- **Fraud check timeout**: Proceed with warning (configurable)

### Recovery
- **Reconciliation**: Detect and correct discrepancies
- **Idempotency replay**: If response lost, client can retry with same key

---

## 9. Monitoring & Observability

### Key Metrics
- **Payment**: Success rate, latency (p50, p99), PSP latency
- **Refund**: Success rate, latency
- **Ledger**: Balance accuracy, entry count
- **Webhooks**: Delivery success rate, retry count
- **Idempotency**: Cache hit rate

### Alerts
- **Payment success rate < 99%**
- **PSP latency > 5s**
- **Ledger imbalance** (debits ≠ credits)
- **Webhook delivery failure > 5%**

### Tracing
- **Trace ID**: Across payment → PSP → ledger
- **Correlation**: Link refund to original payment

### Audit
- **Immutable log**: All transactions, ledger entries
- **Compliance**: PCI audit trail

---

## 10. Interview Tips

### Follow-up Questions
- "How would you handle a split payment (e.g., $50 card + $50 gift card)?"
- "How do you prevent race conditions when two refunds for same payment?"
- "How would you design a subscription billing system on top?"
- "How do you handle currency conversion?"
- "What if the ledger and PSP get out of sync?"

### Common Mistakes
- **Storing card data**: Never; always tokenize
- **Ignoring idempotency**: Critical for payments
- **Single PSP**: No failover
- **Synchronous webhooks**: Block payment; should be async
- **No ledger**: Hard to reconcile, audit

### Key Points to Emphasize
- **Idempotency**: Every payment/refund endpoint; prevent double charge
- **Double-entry ledger**: Debits = credits; audit trail
- **PSP abstraction**: Adapter pattern; multi-PSP support
- **PCI DSS**: No raw card data; tokenization
- **Webhooks**: Async, retry, idempotent consumer

---

## Appendix: Deep Dive Topics

### A. Payment Flow States
| State | Description |
|-------|-------------|
| created | Payment initiated |
| authorized | Funds held, not captured |
| capture_pending | Capture requested |
| succeeded | Funds transferred |
| failed | Declined or error |
| canceled | Auth released |
| refunded | Partial or full refund |

### B. Idempotency Key Best Practices
- **Format**: UUID or `{resource}_{client_generated_id}`
- **Scope**: Per merchant or global
- **TTL**: 24 hours typical; 7 days for compliance
- **Replay**: Return exact same response (status, IDs)
- **Storage**: Redis for speed; DB for audit

### C. Ledger Account Types
- **Customer**: External; represents money owed/paid
- **Merchant settlement**: Internal; balance to pay out
- **Fees**: Internal; our revenue
- **Refund liability**: Internal; hold for refunds

### D. Webhook Signature Verification
```
HMAC-SHA256(webhook_secret, timestamp + "." + body)
```
- Merchant verifies signature in `X-Webhook-Signature` header
- Replay protection: Include timestamp; reject if > 5 min old
- Idempotency: Merchant stores event_id; skip if seen

### E. Retry Strategy for PSP Calls
- **Retryable**: 5xx, timeout, connection error
- **Non-retryable**: 4xx (except 429), card declined
- **Backoff**: 1s, 2s, 4s, 8s (exponential)
- **Max attempts**: 3-5
- **Idempotency**: Use same idempotency key on retry

### F. Split Payment Design
- **Multiple payment methods**: Card + gift card + wallet
- **Order**: Apply gift card first (no fee), then wallet, then card
- **Atomicity**: All-or-nothing; if one fails, rollback others
- **Ledger**: Separate debit/credit per payment method

### G. Subscription Billing on Top
- **Recurring**: Job scheduler triggers charge at renewal
- **Invoice**: Generate before charge; store invoice_id
- **Retry**: If charge fails, retry 3x over 2 weeks; then suspend
- **Webhook**: subscription.renewed, subscription.payment_failed

### H. Reconciliation Process
- **Daily job**: Fetch PSP settlement report (CSV/API)
- **Match**: transaction_id, amount, status
- **Discrepancies**: Log; alert; manual review
- **Auto-fix**: Rare; usually PSP or our bug

### I. 3D Secure (3DS) Flow
- **Trigger**: High-risk transaction or merchant config
- **Redirect**: User sent to bank's 3DS page (OTP, biometric)
- **Callback**: Bank redirects back with auth result
- **Resume**: Complete auth/capture with 3DS result

### J. Currency Conversion
- **FX rates**: Daily update from provider; cache
- **Display**: Show amount in user currency
- **Settlement**: Charge in merchant currency; convert for display
