# Module 10: Database Isolation & Failure Modes

---

## 1. MVCC (Multi-Version Concurrency Control)

### Definition
A concurrency control method where each write creates a new version of the data instead of overwriting. Readers see a consistent snapshot without blocking writers.

### How It Works
```
Transaction T1 (started at time 100):
  Sees: Row version created at time ≤ 100

Transaction T2 (started at time 105):
  Updates row → creates NEW version at time 105
  Old version still exists for T1

T1 reads row → gets version from time 100 (snapshot)
T2 reads row → gets version from time 105 (latest)

"Readers don't block writers. Writers don't block readers."
```

### Visual
```
Row "user:1" versions:

  Version 1 (time=90):  {name: "Alice", age: 25}  ← T1 sees this
  Version 2 (time=105): {name: "Alice", age: 26}  ← T2 sees this

  T1 started at time=100, so it sees Version 1
  T2 started at time=110, so it sees Version 2
```

### Implementation (PostgreSQL)
```
Each row has:
  xmin: Transaction ID that created this version
  xmax: Transaction ID that deleted/updated this version (0 if alive)

Visibility rule:
  Row visible to transaction T if:
    xmin < T AND (xmax == 0 OR xmax > T)
```

### Tradeoffs

| Pros | Cons |
|------|------|
| High concurrency (no read locks) | Storage bloat (old versions) |
| Consistent snapshots | Needs VACUUM to clean dead tuples |
| No read-write contention | Write-write conflicts still possible |

### Real Systems
PostgreSQL, MySQL (InnoDB), Oracle, CockroachDB, SQLite (WAL mode)

---

## 2. Snapshot Isolation

### Definition
Each transaction sees a consistent snapshot of the database as it existed at the start of the transaction. Changes made by other concurrent transactions are invisible.

### Guarantee
```
Transaction T sees ALL committed data as of T's start time.
T does NOT see:
  - Uncommitted changes from other transactions
  - Changes committed AFTER T started
```

### Write Skew Anomaly
```
Snapshot isolation allows "write skew" — a subtle bug:

  Constraint: at least 1 doctor on call
  Doctor A on call: true    Doctor B on call: true

  T1 reads: both on call → sets A = off call (allowed, B still on)
  T2 reads: both on call → sets B = off call (allowed, A still on)

  Both commit. Result: NEITHER doctor on call!
  Constraint violated because each saw the OTHER as on call.
```

### Fix: Serializable Isolation
Serializable detects write skew and aborts one transaction.

---

## 3. Isolation Levels

### Anomalies at Each Level

```
┌─────────────────┬────────────┬──────────────┬────────────┬─────────────┐
│ Isolation Level │ Dirty Read │ Non-Repeatable│ Phantom    │ Write Skew  │
│                 │            │ Read          │ Read       │             │
├─────────────────┼────────────┼──────────────┼────────────┼─────────────┤
│ Read Uncommitted│ ✗ Possible │ ✗ Possible   │ ✗ Possible │ ✗ Possible  │
│ Read Committed  │ ✓ Prevented│ ✗ Possible   │ ✗ Possible │ ✗ Possible  │
│ Repeatable Read │ ✓ Prevented│ ✓ Prevented  │ ✗ Possible │ ✗ Possible  │
│ Snapshot        │ ✓ Prevented│ ✓ Prevented  │ ✓ Prevented│ ✗ Possible  │
│ Serializable    │ ✓ Prevented│ ✓ Prevented  │ ✓ Prevented│ ✓ Prevented │
└─────────────────┴────────────┴──────────────┴────────────┴─────────────┘
```

### Anomaly Definitions
```
DIRTY READ:
  T1 writes (uncommitted). T2 reads T1's uncommitted data.
  T1 rolls back. T2 has "dirty" data.

NON-REPEATABLE READ:
  T1 reads row. T2 modifies row and commits.
  T1 reads same row again → different value.

PHANTOM READ:
  T1 queries "all orders > $100" → 5 rows.
  T2 inserts new order > $100.
  T1 re-queries → 6 rows. New row "appeared" (phantom).

WRITE SKEW:
  Two transactions read overlapping data, make decisions based on
  it, and write to different rows. Result violates constraint.
```

---

## 4. Read Committed

### Definition
Each statement in a transaction sees only data committed before that statement began. Different statements in the same transaction may see different committed data.

### Behavior
```
T1:  BEGIN
T1:  SELECT balance → $100
                              T2: UPDATE balance = $50; COMMIT
T1:  SELECT balance → $50  ← sees T2's committed change!
T1:  COMMIT

Within T1, the two reads return different values.
```

### Implementation
- Each query gets its own snapshot (not transaction-level)
- Write locks held until commit (prevents dirty writes)

### Default for: PostgreSQL, Oracle, SQL Server

---

## 5. Repeatable Read

### Definition
Once a transaction reads a row, subsequent reads in the same transaction return the same value, even if other transactions modify it.

### Behavior
```
T1:  BEGIN
T1:  SELECT balance → $100
                              T2: UPDATE balance = $50; COMMIT
T1:  SELECT balance → $100  ← still sees $100 (snapshot from T1 start)
T1:  COMMIT
```

### Implementation
Transaction-level snapshot (PostgreSQL) or row-level locks (MySQL).

### Note
MySQL's "Repeatable Read" actually provides snapshot isolation (prevents phantoms too). PostgreSQL's RR is true snapshot isolation.

---

## 6. Serializable Isolation

### Definition
Transactions execute as if they ran one at a time (serial order), even though they actually run concurrently. Strongest isolation level.

### Implementation Approaches
```
1. ACTUAL SERIAL EXECUTION (VoltDB, Redis)
   Single thread processes all transactions.
   Simple but limited to single-core throughput.

2. TWO-PHASE LOCKING (2PL) (MySQL)
   Growing phase: acquire locks, no releases.
   Shrinking phase: release locks, no acquires.
   Causes deadlocks and reduced concurrency.

3. SERIALIZABLE SNAPSHOT ISOLATION (SSI) (PostgreSQL)
   Detect serialization conflicts at commit time.
   Abort one transaction if conflict detected.
   Optimistic: no blocking, just validation at commit.
```

### SSI (PostgreSQL ≥ 9.1)
```
T1: Read X, Write Y
T2: Read Y, Write X

SSI detects: T1 reads X that T2 writes
             T2 reads Y that T1 writes
             → cycle → one must abort

This catches write skew that Snapshot Isolation misses.
```

### Tradeoffs

| Approach | Concurrency | Overhead | Deadlock |
|----------|-------------|----------|----------|
| Actual serial | None (1 at a time) | Lowest | No |
| 2PL | Low (lock contention) | Lock management | Yes |
| SSI | High (optimistic) | Abort + retry cost | No |

---

## 7. Deadlock

### Definition
Two or more transactions permanently block each other, each waiting for a lock the other holds.

### Classic Example
```
T1: LOCK row A            T2: LOCK row B
T1: LOCK row B → WAITS    T2: LOCK row A → WAITS
         ↓                          ↓
    Both wait forever → DEADLOCK

Wait-for graph:
  T1 ──→ T2
   ↑      │
   └──────┘  ← cycle = deadlock
```

### Detection and Resolution
```
1. TIMEOUT: If transaction waits > X ms, abort it
   Simple but may abort unnecessarily

2. WAIT-FOR GRAPH: Build dependency graph, detect cycles
   Abort one transaction in the cycle (victim selection)
   PostgreSQL does this

3. LOCK ORDERING: Always acquire locks in consistent order
   Prevents cycles by design

4. OPTIMISTIC CONCURRENCY: No locks, validate at commit
   Retry on conflict
```

### Real Production Issue
2017: Postgres deadlocks under high contention on foreign key constraint checks. Fix: batch updates and explicit lock ordering.

---

## 8. Starvation

### Definition
A thread or transaction is perpetually unable to acquire a resource because other higher-priority threads continuously preempt it.

### Examples
```
1. READER-WRITER LOCK:
   Continuous stream of readers → writer never gets lock
   Solution: Fair lock (readers wait if writer is queued)

2. DEADLOCK VICTIM:
   Same transaction repeatedly chosen as deadlock victim
   Solution: Randomize or rotate victim selection

3. LOCK CONVOY:
   Thread holds lock during I/O → all other threads queue up
   Solution: Release lock before I/O
```

### Prevention
- Fair locks (FIFO ordering)
- Priority aging (increase priority over time)
- Bounded wait times with retry

---

## 9. Priority Inversion

### Definition
A high-priority task is blocked by a low-priority task that holds a needed resource, while a medium-priority task preempts the low-priority task — effectively making the high-priority task wait for the medium-priority one.

### Famous Example: Mars Pathfinder (1997)
```
High priority:   Bus management task (needs mutex)
Medium priority: Communication task (long-running)
Low priority:    Meteorological task (holds mutex)

  Low takes mutex → Medium preempts Low → High needs mutex
  High BLOCKED by Low, but Low can't run because Medium is running

  High effectively has LOWEST priority!
```

### Visual
```
Priority:  HIGH ──────[BLOCKED]────────────────[RUNS]
           MED  ──────────[RUNNING]────[DONE]──
           LOW  ─[LOCK]──[PREEMPTED]───────────[UNLOCK]
                        ↑
                 Medium preempts Low
                 High waits for Low's lock
                 But Low can't run!
```

### Solutions
```
1. PRIORITY INHERITANCE: Low temporarily gets High's priority
   → Medium can't preempt Low → Low finishes, releases lock

2. PRIORITY CEILING: Mutex has a ceiling priority
   Any thread that locks it gets ceiling priority
   Prevents preemption entirely

3. LOCK-FREE ALGORITHMS: No locks → no inversion possible
```

### Real Systems
RTOS (Real-Time Operating Systems), Mars Pathfinder (fixed with priority inheritance), Linux (PI futexes)

### Summary
Priority inversion occurs when a high-priority task waits for a low-priority task's lock while medium-priority tasks preempt. Fix with priority inheritance or lock-free algorithms.
