# ACID Transactions

## Staff+ Engineer Deep Dive for FAANG Interviews

---

## 1. Concept Overview

### Definition
**ACID** is an acronym describing four properties that guarantee reliable processing of database transactions:
- **Atomicity**: All operations in a transaction succeed or all fail—no partial commits
- **Consistency**: Transaction brings database from one valid state to another; invariants preserved
- **Isolation**: Concurrent transactions don't interfere; each sees a consistent view
- **Durability**: Once committed, data persists even after system failure

### Purpose
ACID transactions ensure data integrity when multiple operations must be treated as a single logical unit. Without ACID, partial failures (e.g., money debited but not credited) would corrupt data.

### Why It Exists
Early database systems had no transaction guarantees. If a system crashed mid-operation, data could be left inconsistent. The ACID properties, formalized in the 1980s, became the foundation for reliable database systems.

### Problems Solved
| Problem | ACID Solution |
|---------|---------------|
| Partial failure | Atomicity: rollback all or commit all |
| Invalid states | Consistency: invariants always hold |
| Concurrent corruption | Isolation: serializable view |
| Data loss after commit | Durability: WAL, replication |

---

## 2. Real-World Motivation

### Google Spanner
- Globally distributed database with external consistency (stronger than serializability)
- Uses TrueTime for ordering across datacenters
- Financial transactions, AdWords—cannot afford double-spend or lost updates

### Stripe
- Payment processing: charge card + update balance + send receipt must be atomic
- Idempotency keys to handle retries safely within transaction semantics

### Uber
- Trip lifecycle: create trip → assign driver → start → complete → charge
- Saga pattern for distributed case; ACID within each service's database

### Amazon
- Inventory: decrement stock + create order must be atomic
- Overselling = lost customer trust; ACID prevents this

### Netflix
- Billing: subscription renewal + payment + entitlement update
- Uses transactional outbox pattern for reliability

### Banking (Any)
- Transfer: debit A, credit B—both or neither
- Classic textbook example of atomicity requirement

---

## 3. Architecture Diagrams

### ACID Transaction Lifecycle
```
┌─────────────────────────────────────────────────────────────────────────┐
│                     ACID TRANSACTION LIFECYCLE                            │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   BEGIN TRANSACTION                                                      │
│         │                                                                │
│         ▼                                                                │
│   ┌─────────────┐     ┌─────────────┐     ┌─────────────┐               │
│   │  Read A     │────▶│  Write B    │────▶│  Read C     │               │
│   └─────────────┘     └─────────────┘     └─────────────┘               │
│         │                    │                    │                     │
│         │    ISOLATION:      │    ATOMICITY:      │                     │
│         │    Invisible to    │    All or nothing  │                     │
│         │    other txns      │                    │                     │
│         │                    │                    │                     │
│         ▼                    ▼                    ▼                     │
│   ┌─────────────────────────────────────────────────────────┐            │
│   │              CONSISTENCY CHECK (invariants)              │            │
│   └─────────────────────────────────────────────────────────┘            │
│         │                    │                                            │
│    COMMIT                 ROLLBACK                                        │
│         │                    │                                            │
│         ▼                    ▼                                            │
│   ┌─────────────┐     ┌─────────────┐                                    │
│   │  WAL flush  │     │  Undo log   │                                    │
│   │  DURABILITY │     │  revert     │                                    │
│   └─────────────┘     └─────────────┘                                    │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Isolation Levels and Anomalies
```
                    Dirty    Non-Repeatable   Phantom
                    Read     Read            Read
                    ─────    ───────────     ─────
Read Uncommitted    ✗ YES    ✗ YES          ✗ YES
Read Committed      ✓ No     ✗ YES          ✗ YES
Repeatable Read     ✓ No     ✓ No           ✗ YES*
Serializable        ✓ No     ✓ No           ✓ No

* PostgreSQL RR prevents phantoms via predicate locking; MySQL uses gap locks
```

### Two-Phase Commit (2PC)
```
┌─────────────────────────────────────────────────────────────────────────┐
│                        TWO-PHASE COMMIT                                  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   COORDINATOR          PARTICIPANT A         PARTICIPANT B               │
│        │                     │                     │                    │
│        │  PREPARE            │                     │                    │
│        │────────────────────▶│                     │                    │
│        │                     │  PREPARE             │                    │
│        │──────────────────────────────────────────▶│                    │
│        │                     │                     │                    │
│        │  VOTE-COMMIT        │  VOTE-COMMIT        │                    │
│        │◀────────────────────│                     │                    │
│        │                     │                     │                    │
│        │                     │  VOTE-COMMIT        │                    │
│        │◀──────────────────────────────────────────│                    │
│        │                     │                     │                    │
│        │  COMMIT             │  COMMIT             │                    │
│        │────────────────────▶│                     │                    │
│        │                     │─────────────────────▶│                    │
│        │                     │                     │                    │
│   All voted YES → COMMIT     Blocking: if coordinator fails,              │
│   Any voted NO  → ABORT      participants wait (blocked)                 │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Saga Pattern (Choreography)
```
┌─────────────────────────────────────────────────────────────────────────┐
│                    SAGA - CHOREOGRAPHY                                    │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   Order Service    Payment Service    Inventory Service    Shipping       │
│        │                  │                    │                │        │
│        │ Create Order     │                    │                │         │
│        │─────────────────▶│ Charge             │                │         │
│        │                  │──────────────────▶│ Reserve         │         │
│        │                  │                   │────────────────▶│ Ship   │
│        │                  │                   │                │         │
│        │                  │                   │                │         │
│   FAILURE: Compensating transactions (reverse order)                      │
│        │                  │                   │                │         │
│        │                  │                   │  Release       │         │
│        │                  │  Refund           │◀───────────────│ Cancel  │
│        │  Cancel Order    │◀──────────────────│                │         │
│        │◀─────────────────│                   │                │         │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Saga Pattern (Orchestration)
```
┌─────────────────────────────────────────────────────────────────────────┐
│                    SAGA - ORCHESTRATION                                   │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│                    ORCHESTRATOR (Central Coordinator)                     │
│                              │                                            │
│         ┌────────────────────┼────────────────────┐                       │
│         │                    │                    │                       │
│         ▼                    ▼                    ▼                       │
│   ┌──────────┐        ┌──────────┐        ┌──────────┐                  │
│   │  Order   │        │ Payment  │        │Inventory │                  │
│   │  Service │        │ Service  │        │ Service  │                  │
│   └──────────┘        └──────────┘        └──────────┘                  │
│         │                    │                    │                       │
│         │ 1. Create          │ 2. Charge          │ 3. Reserve           │
│         │◀───────────────────│◀───────────────────│                       │
│         │                    │                    │                       │
│   Orchestrator tracks state; on failure, executes compensating txns      │
│   in reverse order                                                        │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Core Mechanics

### Atomicity Implementation
- **Write-Ahead Log (WAL)**: All changes logged before applying to data pages
- **Undo Log**: For rollback; store before-images
- **Commit Protocol**: Mark transaction committed in log; fsync log to disk
- **Recovery**: Replay committed transactions; undo uncommitted

### Consistency Implementation
- **Constraints**: Primary key, foreign key, check, unique
- **Triggers**: Application logic at DB level
- **Application**: Business rules enforced in transaction
- **Invariants**: e.g., sum(credits) = sum(debits)

### Isolation Implementation (MVCC)
- **Multi-Version Concurrency Control**: Each transaction sees snapshot
- **Version chains**: Each row has multiple versions; visible based on transaction ID
- **Visibility rules**: Row visible if created before snapshot and not deleted
- **Garbage collection**: Vacuum removes old versions

### Durability Implementation
- **WAL**: Log records persisted before commit acknowledged
- **fsync**: Force OS to flush buffers to disk
- **Replication**: Async/sync replica for additional durability

### Isolation Levels Deep Dive
- **Read Uncommitted**: Read latest version; no isolation (rarely used)
- **Read Committed**: See committed data at statement start; new snapshot per statement
- **Repeatable Read**: Snapshot at transaction start; same rows throughout
- **Serializable**: Full isolation; detect conflicts, abort if necessary

---

## 5. Numbers

| Metric | Value |
|--------|-------|
| 2PC latency | +2-3 round trips (prepare + commit) |
| 2PC blocking | Coordinator failure = participants blocked until timeout |
| Saga compensation | N services = up to N compensating transactions |
| MVCC overhead | ~20-30% storage for version chains |
| Lock wait timeout | Typically 50ms-30s |
| Deadlock detection | O(n²) in number of transactions |

### Scale
- **Spanner**: 2PC across continents; <10ms p99 for single-region
- **PostgreSQL**: 1000s TPS per instance with proper tuning
- **MySQL**: Similar; InnoDB optimized for short transactions

---

## 6. Tradeoffs

### Isolation vs Performance
| Level | Consistency | Performance | Use Case |
|-------|--------------|-------------|----------|
| Read Uncommitted | Lowest | Highest | Rarely used |
| Read Committed | Low | High | Default in many DBs |
| Repeatable Read | Medium | Medium | Avoid non-repeatable reads |
| Serializable | Highest | Lowest | Critical sections |

### 2PC vs Saga
| Aspect | 2PC | Saga |
|--------|-----|------|
| Consistency | Strong | Eventual |
| Blocking | Yes (coordinator failure) | No |
| Complexity | Coordinator logic | Compensation logic |
| Use case | Same DB or tightly coupled | Distributed services |

---

## 7. Variants / Implementations

### Distributed Transaction Protocols
- **2PC**: Blocking; coordinator single point of failure
- **3PC**: Adds pre-commit phase; reduces but doesn't eliminate blocking
- **Paxos/Raft**: For consensus, not general transactions
- **Calvin**: Deterministic ordering; no 2PC for commit
- **Spanner**: 2PC + TrueTime; external consistency

### Saga Variants
- **Choreography**: Event-driven; no central coordinator; harder to reason
- **Orchestration**: Central coordinator; easier to understand; single point of logic

### Database Implementations
| DB | Default Isolation | Serializable Support |
|----|-------------------|----------------------|
| PostgreSQL | Read Committed | Full serializable |
| MySQL | Repeatable Read | Serializable |
| Oracle | Read Committed | Serializable |
| SQL Server | Read Committed | Serializable |

---

## 8. Scaling Strategies

- **Shard by entity**: Avoid cross-shard transactions
- **Saga for cross-service**: Accept eventual consistency
- **Optimistic concurrency**: Version numbers; retry on conflict
- **Reduce transaction scope**: Shorter = less contention
- **Read replicas**: Read-only transactions can use replicas (read committed semantics)

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|-------------|
| Coordinator crash (2PC) | Participants blocked | Timeout + abort; manual intervention |
| Network partition | 2PC blocks | Saga; avoid 2PC |
| Deadlock | One transaction aborted | Deadlock detection; retry |
| Long transaction | Block others, hold locks | Set statement timeout |
| Replica lag | Read stale data | Read from primary for critical reads |

---

## 10. Performance Considerations

- **Keep transactions short**: Release locks quickly
- **Avoid long-running transactions**: Risk of bloat, lock contention
- **Index for lock efficiency**: Reduce rows locked
- **Batch operations**: Fewer round trips
- **Connection pooling**: Reuse connections

---

## 11. Use Cases

| Use Case | ACID Requirement | Approach |
|----------|------------------|----------|
| Bank transfer | Atomicity critical | Single DB transaction |
| E-commerce order | Atomicity | Transaction or Saga |
| Inventory | Consistency | Strong consistency |
| Analytics | Weak | Read committed, replicas |
| Session update | Low | Eventual consistency OK |

---

## 12. Comparison Tables

### Isolation Level Comparison
| Level | Dirty Read | Non-Repeatable | Phantom | Implementation |
|-------|------------|----------------|---------|----------------|
| Read Uncommitted | Possible | Possible | Possible | No locking |
| Read Committed | No | Possible | Possible | Row locks (short) |
| Repeatable Read | No | No | Possible* | Snapshot + row locks |
| Serializable | No | No | No | Full serialization |

### Distributed Transaction Comparison
| Protocol | Consistency | Availability | Blocking | Complexity |
|----------|--------------|--------------|----------|------------|
| 2PC | Strong | No (blocks) | Yes | Medium |
| 3PC | Strong | Better | Reduced | High |
| Saga | Eventual | Yes | No | High (compensation) |
| Calvin | Strong | Yes | No | Very High |

---

## 13. Code or Pseudocode

### Atomicity Example
```sql
BEGIN;
  UPDATE accounts SET balance = balance - 100 WHERE id = 'A';
  UPDATE accounts SET balance = balance + 100 WHERE id = 'B';
  -- If either fails, both rollback
COMMIT;
```

### Isolation Level
```sql
-- Set isolation for transaction
SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;
BEGIN;
  SELECT * FROM accounts WHERE id = 'A';  -- Snapshot
  -- Other transactions' changes invisible
  UPDATE accounts SET balance = 100 WHERE id = 'A';
COMMIT;
```

### Saga Compensation (Pseudocode)
```python
def create_order_saga(order_data):
    try:
        order = order_service.create(order_data)
        payment = payment_service.charge(order.total)
        inventory_service.reserve(order.items)
        shipping_service.ship(order)
        return success
    except PaymentError:
        order_service.cancel(order.id)  # Compensate
        raise
    except InventoryError:
        order_service.cancel(order.id)
        payment_service.refund(payment.id)  # Compensate
        raise
```

### 2PC Pseudocode
```python
def two_phase_commit(participants, transaction):
    # Phase 1: Prepare
    for p in participants:
        if not p.prepare(transaction):
            for q in participants[:p]: q.abort()
            return ABORT
    # Phase 2: Commit
    for p in participants:
        p.commit(transaction)
    return COMMIT
```

---

## 14. Interview Discussion

### Key Points
1. **Atomicity ≠ Consistency**: Atomicity is all-or-nothing; consistency is valid state
2. **Isolation is a spectrum**: Trade off consistency for performance
3. **2PC blocks**: Coordinator failure = participants stuck
4. **Saga trades strong consistency for availability**: Compensation logic is complex
5. **MVCC**: Explain snapshot isolation; no read locks

### Common Questions
- **Q**: "What's the difference between consistency and atomicity?"
  - **A**: Atomicity = all or nothing execution. Consistency = database invariants hold (e.g., no negative balance)
- **Q**: "When would you use Saga over 2PC?"
  - **A**: Cross-service, microservices; when you can't afford 2PC blocking; when eventual consistency is acceptable
- **Q**: "What causes phantom reads?"
  - **A**: Transaction A reads rows matching predicate; transaction B inserts new matching row; A reads again, sees new row
- **Q**: "How does MVCC avoid locks?"
  - **A**: Readers never block writers; each sees snapshot; writers create new versions; conflict detection on commit

---

## 15. MVCC Deep Dive

### How MVCC Works
- Each row has multiple versions; each version has creation and deletion transaction IDs (xmin, xmax)
- Transaction sees row if: xmin < snapshot_id AND (xmax is NULL OR xmax > snapshot_id)
- Snapshot ID = oldest active transaction at start
- Writers create new version; old version retained until no transaction needs it
- Vacuum removes dead versions

### Version Chain Example
```
Row versions: [v1: xmin=100, xmax=200] [v2: xmin=200, xmax=NULL]
Txn 150 reads: sees v1 (100<150, 200>150)
Txn 250 reads: sees v2 (200<250, xmax NULL)
```

---

## 16. Lock Types and Granularity

| Lock Type | Scope | Blocks |
|-----------|-------|--------|
| Row lock | Single row | Other writers to same row |
| Page lock | Page (8KB) | Writers to page |
| Table lock | Entire table | Most operations |
| Intent lock | Hierarchy | Prevents incompatible locks |
| Advisory lock | Application | Application-controlled |

### Lock Escalation
- SQL Server: Row locks can escalate to table lock under memory pressure
- PostgreSQL: No escalation; many row locks = many lock structures

---

## 17. Recovery and WAL

### Crash Recovery
1. **Analysis**: Scan WAL; determine which transactions committed
2. **Redo**: Replay committed transactions
3. **Undo**: Roll back uncommitted transactions
4. **Checkpoint**: Periodic flush of dirty pages; reduces recovery time

### WAL Record Structure
- LSN (Log Sequence Number): Unique identifier
- Transaction ID, operation type, before/after images
- Force write to disk (fsync) before commit acknowledged

---

## 18. Anomaly Examples (Detailed)

### Dirty Read
```
T1: UPDATE accounts SET balance=0 WHERE id='A';  (not committed)
T2: SELECT balance FROM accounts WHERE id='A';   → 0 (reads uncommitted)
T1: ROLLBACK;
T2 now has seen data that never existed.
```

### Non-Repeatable Read
```
T1: SELECT balance FROM accounts WHERE id='A';  → 100
T2: UPDATE accounts SET balance=50 WHERE id='A'; COMMIT;
T1: SELECT balance FROM accounts WHERE id='A';  → 50 (different!)
```

### Phantom Read
```
T1: SELECT * FROM orders WHERE status='pending';  → 3 rows
T2: INSERT INTO orders (..., status) VALUES (..., 'pending'); COMMIT;
T1: SELECT * FROM orders WHERE status='pending';  → 4 rows (new row appeared)
```
