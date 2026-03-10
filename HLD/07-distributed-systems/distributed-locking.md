# Distributed Locking in Distributed Systems

> Staff+ Engineer Level — FAANG Interview Deep Dive

---

## 1. Concept Overview

### Definition

**Distributed locking** provides **mutual exclusion** across multiple nodes in a distributed system. Only one node (or process) can hold a lock at a time, enabling exclusive access to a shared resource (e.g., a job, a file, a database row).

### Purpose

- **Mutual exclusion**: Prevent concurrent modification of shared state
- **Job deduplication**: Ensure only one instance runs a scheduled job
- **Resource serialization**: Single writer to a resource
- **Critical sections**: Protect non-distributed code from concurrent execution

### Problems Solved

| Problem | Solution |
|---------|----------|
| Duplicate job execution | Distributed lock before job |
| Concurrent writes to same key | Lock per key |
| Thundering herd | Fair queuing (e.g., ZooKeeper sequence) |
| Stale lock holder | Fencing tokens |

---

## 2. Real-World Motivation

### Google Chubby

- Lock service for GFS, Bigtable
- Lease-based locks; fencing tokens
- Used for leader election, configuration

### Amazon DynamoDB

- Conditional writes for optimistic locking
- No traditional locks; application-level coordination

### Netflix

- Distributed locks for batch job coordination
- Prevents duplicate processing of same data

### Uber

- Locks for driver assignment, ride state transitions
- Ensures single driver per ride

### Stripe

- Locks for payment processing idempotency
- Prevents double-charging

---

## 3. Architecture Diagrams

### Redis SET NX EX Lock

```
    Client A                    Redis                    Client B
        |                         |                          |
        | SET lock:resource NX EX 30 "uuid-A"                |
        |------------------------>|                          |
        |<------- OK --------------|                         |
        |  (lock acquired)          |                         |
        |                          | SET lock:resource NX EX 30 "uuid-B"
        |                          |<-------------------------|
        |                          |------- nil -------------->
        |  (B blocked)              |  (key exists)            |
        |                          |                         |
        | DEL lock:resource        |                         |
        |------------------------>|                          |
        |  (lock released)         | SET lock:resource NX EX 30 "uuid-B"
        |                          |<-------------------------|
        |                          |------- OK --------------->
        |                          |  (B acquired)            |
```

### Redlock — Multi-Instance

```
    Client                    Redis R1    Redis R2    Redis R3    Redis R4    Redis R3
       |                         |           |           |           |           |
       | SET NX EX (with random value) to all
       |------------------------>|---------->|---------->|---------->|---------->|
       |<-------- OK ------------|---------->|---------->|---------->|---------->|
       |  (3/5 acquired - quorum)            |           |           |           |
       |  Lock held                           |           |           |           |
       |  On release: DEL on all instances that had lock
```

### ZooKeeper Lock (Ephemeral Sequential)

```
    /locks/my_lock
        /lock_0000000001  (Client A - holder)
        /lock_0000000002  (Client B - waiting, watches 0000000001)
        /lock_0000000003  (Client C - waiting, watches 0000000002)
    
    When A releases (session ends or deletes): 0000000001 gone
    B gets watch notification, checks: am I lowest? Yes -> B acquires
    C watches 0000000002 (B)
```

### Fencing Token Flow

```
    Lock Server              Client A (old)         Storage           Client B (new)
         |                         |                    |                    |
         |  grant lock, token=33   |                    |                    |
         |------------------------>|                    |                    |
         |                         |  write (token=33)  |                    |
         |                         |------------------->|  OK                 |
         |  grant lock, token=34   |                    |                    |
         |--------------------------------------------------------->|
         |                         |                    |  write (token=34)  |
         |                         |                    |<-------------------|
         |                         |  write (token=33)  |  OK                 |
         |                         |------------------->|  REJECT (33 < 34)  |
         |                         |                    |                    |
    (Old holder's write rejected - split-brain prevented)
```

---

## 4. Core Mechanics

### Redis SET NX EX

- **SET key value NX EX seconds**: Set if not exists, with expiry
- **Value**: Unique identifier (UUID) — used to verify ownership before delete
- **Release**: Lua script: delete only if value matches (atomic)
- **Problem**: Single Redis failure = lock lost or stuck

### Redlock Algorithm (Antirez)

1. Get current time
2. Acquire lock on N instances sequentially (SET NX EX with random value)
3. Lock acquired if: got it on majority, total time < TTL
4. If acquired: effective TTL = TTL - elapsed
5. If not: unlock all instances
6. Release: delete on all instances

**Martin Kleppmann's Critique**:
- Assumes synchronous model; real systems have pauses (GC, network)
- Clock skew can cause premature expiry
- No fencing → stale lock holder can corrupt data
- Suggests: use lock for efficiency, but always use fencing for correctness

### ZooKeeper Lock

- **Ephemeral sequential** znodes under `/lock`
- **Acquire**: Create znode; if lowest sequence, you have lock
- **Else**: Watch the znode just before yours
- **Release**: Delete your znode (or session end)
- **No herd effect**: Each watcher only watches one node

### etcd Lock

- **Lease + key**: Create key with lease; lease TTL
- **Campaign**: Compare-and-swap; winner holds
- **Keepalive**: Renew lease while holding
- **Release**: Delete key or let lease expire

### Database Locks

- **SELECT FOR UPDATE**: Row-level lock in transaction
- **Advisory locks**: PostgreSQL `pg_try_advisory_lock(id)`
- **Pros**: ACID, durable
- **Cons**: DB bottleneck, connection holding

### Fencing Tokens

- **Monotonic**: Each lock grant gets higher token
- **Storage**: Rejects operations with token < last seen
- **Prevents**: Stale lock holder (e.g., paused process) from writing

---

## 5. Numbers

| Implementation | Acquire Latency | Hold Cost | Fault Tolerance |
|----------------|-----------------|-----------|------------------|
| Redis single | ~1ms | Low | None (SPOF) |
| Redlock (5 nodes) | ~5-15ms | Low | 2 node failures |
| ZooKeeper | ~5-20ms | Session | N/2 - 1 |
| etcd | ~10-30ms | Lease | N/2 - 1 |
| PostgreSQL advisory | ~1-5ms | Connection | DB HA |

### Redlock Configuration

- **N**: Typically 5 (tolerate 2 failures)
- **TTL**: 10-30 seconds typical
- **Clock drift**: Must be << TTL

---

## 6. Tradeoffs

### Redlock vs. ZooKeeper

| Aspect | Redlock | ZooKeeper |
|--------|---------|-----------|
| Correctness | Debated (Kleppmann) | Strong (consensus) |
| Dependencies | Redis cluster | ZooKeeper |
| Latency | Lower | Higher |
| Fencing | No built-in | Can add |
| Use case | Caching, non-critical | Critical sections |

### Lock vs. Lock-Free

| Locks | CRDTs / Optimistic |
|-------|---------------------|
| Blocking | Non-blocking |
| Simple mental model | Complex merge logic |
| Single writer | Multi-writer |
| Conflict avoidance | Conflict resolution |

---

## 7. Variants / Implementations

### Redisson (Java)

- Implements Redlock
- Watchdog: auto-extends lock while held
- Reentrant locks

### Apache Curator (ZooKeeper)

- `InterProcessMutex`: distributed lock
- `InterProcessSemaphore`: distributed semaphore

### etcd concurrency

- `concurrency.NewSession` + `concurrency.NewMutex`
- Lease-based; automatic renewal

### Chubby

- Lock with delay (for fencing)
- Sequence numbers for ordering

---

## 8. Scaling Strategies

1. **Shard locks**: Different resources → different lock keys
2. **Lock granularity**: Coarse (few contentions) vs. fine (more parallelism)
3. **Short hold time**: Minimize critical section
4. **Lock-free alternatives**: CRDTs, version vectors, optimistic concurrency

---

## 9. Failure Scenarios

| Scenario | Redis | Redlock | ZooKeeper |
|----------|-------|---------|-----------|
| Single node failure | Lock lost | Survives (quorum) | Survives |
| Network partition | Stuck | Possible dual holders | Minority blocked |
| Client pause (GC) | Lock expires, another acquires | Same | Same — fencing needed |
| Clock skew | Premature expiry | Redlock vulnerable | Less impact |

### Martin Kleppmann vs. Antirez Debate

**Kleppmann**:
- Redlock assumes bounded pause; GC can pause 100ms+
- No fencing → stale holder can corrupt
- Recommendation: Use ZooKeeper/etcd for correctness; or always pair with fencing

**Antirez**:
- Redlock is probabilistic; good for efficiency, not for critical correctness
- Agrees fencing is needed for strict correctness
- Redlock suitable when you can tolerate edge cases (e.g., cache)

---

## 10. Performance Considerations

- **Latency**: ZooKeeper/etcd > Redis
- **Throughput**: Locks serialize; minimize hold time
- **Herd effect**: ZooKeeper avoids with sequential watch
- **Connection pooling**: DB locks hold connections

---

## 11. Use Cases

| Use Case | Lock Type | Why |
|----------|-----------|-----|
| Scheduled job | Distributed lock | One instance runs |
| Database migration | Lock | Single migrator |
| Resource allocation | Lock per resource | Exclusive access |
| Cache stampede | Lock | One recomputes |
| Payment idempotency | Lock or version | No double charge |

---

## 12. Comparison Tables

### Lock Implementation Summary

| Implementation | Consistency | Fencing | Complexity | Best For |
|----------------|-------------|---------|------------|----------|
| Redis NX EX | Single node | No | Low | Non-critical |
| Redlock | Quorum | No | Medium | Caching |
| ZooKeeper | Strong | Add-on | Medium | Coordination |
| etcd | Strong | Add-on | Medium | K8s ecosystem |
| DB advisory | Strong | N/A | Low | DB-centric |

### When to Use What

| Scenario | Recommendation |
|----------|----------------|
| Cache stampede | Redis lock (best-effort OK) |
| Financial transaction | ZooKeeper/etcd + fencing |
| Job deduplication | Any; ensure idempotency as backup |
| Multi-datacenter | Consider CRDTs instead |

---

## 13. Code or Pseudocode

### Redis Lock (Acquire + Release)

```python
import uuid
import time

def acquire_lock(redis, resource, ttl=10):
    identifier = str(uuid.uuid4())
    end = time.time() + 3  # 3 second wait
    while time.time() < end:
        if redis.set(resource, identifier, nx=True, ex=ttl):
            return identifier
        time.sleep(0.001)
    return None

def release_lock(redis, resource, identifier):
    script = """
    if redis.call("get", KEYS[1]) == ARGV[1] then
        return redis.call("del", KEYS[1])
    else
        return 0
    end
    """
    redis.eval(script, 1, resource, identifier)
```

### Redlock (Simplified)

```python
def redlock_acquire(redis_instances, resource, ttl):
    identifier = str(uuid.uuid4())
    start = time.time()
    acquired = 0
    for r in redis_instances:
        if r.set(resource, identifier, nx=True, ex=ttl):
            acquired += 1
    elapsed = time.time() - start
    if acquired >= len(redis_instances) // 2 + 1 and elapsed < ttl * 0.01:
        return identifier
    # Unlock all
    for r in redis_instances:
        r.delete(resource)
    return None
```

### ZooKeeper Lock (Curator-style)

```python
def acquire_zk_lock(zk, lock_path="/locks/resource"):
    path = zk.create(lock_path + "/lock_", ephemeral=True, sequential=True)
    while True:
        children = sorted(zk.get_children(lock_path))
        if path.endswith(children[0]):
            return path  # We have the lock
        watch_node = lock_path + "/" + children[children.index(path.split("/")[-1]) - 1]
        event = zk.exists(watch_node, watch=True)
        event.wait()  # Block until predecessor gone
```

### Fencing Token (Pseudocode)

```python
# Lock server
token_counter = 0
def grant_lock(client_id):
    token_counter += 1
    return (lock_handle, token_counter)

# Storage
last_token = 0
def write(data, token):
    if token <= last_token:
        raise StaleTokenError("Reject: stale lock holder")
    last_token = token
    do_write(data)
```

---

## 15. Deep Dive: Martin Kleppmann's Redlock Critique

**Blog**: "Is Redis Redlock safe?" (2016)

**Arguments**:
1. **Timing assumptions**: Redlock assumes bounded pause. GC can pause JVM 100ms+; network can delay. During pause, lock may expire; another acquires; both think they hold.
2. **No fencing**: Even with correct timing, if process pauses after "releasing" (del) but before actually doing so, another acquires. Old process resumes, does DEL, deletes new holder's lock. Fencing prevents this.
3. **Conclusion**: Use Redlock for efficiency (e.g., cache stampede) where occasional duplicate is OK. For correctness (e.g., financial), use lock + fencing. Or use ZooKeeper/etcd.

**Antirez response**: Agrees fencing is needed for strict correctness. Redlock is a "best-effort" lock. For critical sections, use something else.

---

## 16. Deep Dive: ZooKeeper Herd Effect Avoidance

**Herd effect**: When lock held by A, 100 waiters all watch the lock. When A releases, all 100 get notified, all try to acquire. Thundering herd.

**Solution**: Sequential nodes. Each waiter creates `/lock/lock_0000000005`. Only the node with smallest number holds. Others watch the node *just before* theirs: `/lock/lock_0000000004`. When 4 goes away, 5 checks: am I smallest? If yes, acquire. If no, watch the new predecessor. Each notification wakes exactly one waiter.

---

## 17. Deep Dive: Lock Granularity

**Coarse**: One lock for entire resource (e.g., "scheduler"). Simple, but serializes everything.

**Fine**: Lock per resource (e.g., per job ID, per user). More parallelism, higher overhead.

**Hierarchical**: Lock parent, then children. E.g., lock table, then row.

**Choosing**: Balance contention vs. overhead. If 1000 jobs, lock per job. If 10 jobs, maybe one lock.

---

## 18. Lock-Free Alternatives: When to Use

**Optimistic concurrency**: Use version numbers. Read (v=1), compute, write if version still 1. If conflict, retry. No lock held. Good for low contention.

**CRDTs**: Conflict-free merge. No lock at all. Good for collaborative editing, counters.

**Event sourcing**: Append-only log. No update conflicts. Replay for state. Good for audit, replay.

**When to use locks**: When you need exclusive access and can't merge (e.g., "run this job exactly once").

---

## 19. Database Advisory Locks: PostgreSQL Example

```sql
-- Session 1
SELECT pg_try_advisory_lock(12345);  -- returns true if acquired
-- do work
SELECT pg_advisory_unlock(12345);

-- Session 2
SELECT pg_try_advisory_lock(12345);  -- blocks or returns false
```

**Pros**: No extra service; ACID; durable. **Cons**: Connection held; DB bottleneck; not distributed across DBs.

**Use case**: Single PostgreSQL; lock for migration, job scheduling.

---

## 20. Interview Walkthrough: Designing Distributed Lock

**Question**: "Design a distributed lock for a job that runs every hour."

**Answer structure**:
1. **Requirements**: One instance runs; failover if holder crashes; no duplicate runs
2. **Mechanism**: ZooKeeper ephemeral sequential or etcd lease
3. **Acquire**: Create znode; if lowest, acquire. Else watch predecessor.
4. **Release**: Delete znode (or session end). Job completes or crashes.
5. **Fencing**: If job does critical storage ops, use fencing tokens
6. **Timeout**: Lock has TTL (Redis) or session (ZK) for crash recovery
7. **Idempotency**: Job should be idempotent as backup (e.g., check "last run" before doing work)

---

## 21. Redlock: Detailed Algorithm

1. Get current time in milliseconds.
2. Try to acquire lock on all N instances sequentially, using same key and random value. Use timeout (e.g., 5ms) per instance.
3. Lock acquired if: got it on at least N/2+1 instances, and total time < lock validity time (TTL).
4. If acquired: effective validity = TTL - elapsed - clock drift margin.
5. If not acquired: unlock ALL instances (even those we didn't get).
6. Release: send Lua script to delete key only if value matches, to all instances.

**Why random value**: Prevents deleting another client's lock.

---

## 22. etcd Concurrency Package

**Session**: Lease-based. Create session with TTL. Session keeps lease alive while held.

**Mutex**: `concurrency.NewMutex(session, "lock-key")`. `Lock()` blocks until acquired. `Unlock()` releases. Uses compare-and-swap on key with lease.

**Election**: `concurrency.Election`. `Campaign()` blocks until elected. `Resign()` releases. Similar to mutex but for leader election.

---

## 23. Lock Timeout and Deadlock Prevention

**Timeout**: Always set. Prevents permanent deadlock if holder crashes without release.

**Refresh**: For long-held locks, refresh before expiry. "Watchdog" in Redisson extends lock while client is alive.

**Deadlock**: A waits for B, B waits for A. Prevention: order locks (always acquire in same order). Or use try-lock with timeout; back off and retry.

---

## 24. Distributed Semaphore

**Semaphore**: Allow K holders (vs. lock = 1). Use case: limit concurrent jobs.

**Implementation**: ZooKeeper — create K ephemeral nodes. Acquire = create (blocks if K exist). Release = delete. etcd: similar with lease.

**Redis**: Use sorted set with timestamp. Add with score=now. Remove oldest when over limit. Or use multiple keys (lock_1, lock_2, ...).

---

## 25. Summary: Lock Implementation Choice

| Scenario | Recommendation |
|----------|----------------|
| Cache stampede prevention | Redis SET NX EX |
| Job deduplication (non-critical) | Redlock or Redis |
| Critical section (financial) | ZooKeeper/etcd + fencing |
| Kubernetes ecosystem | etcd |
| Single PostgreSQL | Advisory lock |
| Multi-writer merge | CRDT, skip lock |

---

## 14. Interview Discussion

### Key Points

1. **Why distributed locks?** — Mutual exclusion across nodes
2. **Redlock debate** — Kleppmann: use fencing; Antirez: probabilistic
3. **ZooKeeper** — Ephemeral sequential; no herd; strong consistency
4. **Fencing** — Always use for critical correctness

### Common Questions

- **"Is Redlock safe?"** — For non-critical: yes. For critical: add fencing or use ZooKeeper
- **"How does ZooKeeper avoid herd effect?"** — Sequential nodes; each watches only predecessor
- **"What is fencing?"** — Monotonic token; storage rejects stale holder
- **"Lock vs. CRDT?"** — Lock: mutual exclusion. CRDT: merge concurrent updates

### Red Flags

- Using Redis lock for financial transactions without fencing
- Ignoring Kleppmann's critique
- No lock timeout (risk of deadlock)
