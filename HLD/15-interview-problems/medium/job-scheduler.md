# Design a Distributed Job Scheduler

## 1. Problem Statement & Requirements

### Problem Statement
Design a distributed job scheduler that schedules one-time and recurring jobs, executes them reliably (at-least-once), supports priorities, retries with backoff, job dependencies (DAG), and scales horizontally to handle millions of jobs per day.

### Functional Requirements
- **One-time jobs**: Schedule for specific time (e.g., "run at 3pm tomorrow")
- **Recurring jobs**: Cron expressions (e.g., "every day at midnight")
- **Priority**: High/medium/low; higher priority runs first
- **Retry with backoff**: Exponential backoff on failure
- **Job dependencies (DAG)**: Job B runs only after Job A completes
- **Job metadata**: Payload, tags, timeout
- **Status tracking**: PENDING, RUNNING, SUCCEEDED, FAILED, CANCELLED

### Non-Functional Requirements
- **Scale**: Millions of jobs per day
- **Reliability**: At-least-once execution (may run twice on failure; job must be idempotent)
- **Latency**: Schedule-to-run delay < 1 minute for due jobs
- **Availability**: 99.9% (scheduler and workers)
- **Horizontal scaling**: Add workers to increase throughput

### Out of Scope
- Exactly-once execution (complex; at-least-once is common)
- Job result storage (beyond status)
- Job versioning
- Multi-tenant isolation (assume single tenant or simple isolation)

---

## 2. Back-of-Envelope Estimation

### Scale Assumptions
- **Jobs/day**: 10M (one-time + recurring)
- **Peak rate**: 2x average = 20M/24h ≈ 230 jobs/sec
- **Avg job duration**: 30 seconds
- **Concurrent jobs**: 230 × 30 ≈ 7,000 workers needed at peak
- **Recurring vs one-time**: 70% recurring, 30% one-time

### QPS Calculation
| Operation | Daily Volume | Peak QPS |
|-----------|--------------|----------|
| Job creation | 10M | ~120 |
| Job execution (start) | 10M | ~120 |
| Job execution (complete) | 10M | ~120 |
| Schedule check (scheduler) | Every 10s, 1000 due jobs/batch | ~100 |
| Status queries | 50M | ~600 |

### Storage (1 year)
- **Jobs**: 10M × 365 × 500B ≈ 1.8 TB
- **Job runs (history)**: 10M × 365 × 200B ≈ 730 GB
- **Indexes**: 2x data ≈ 5 TB total

### Bandwidth
- **Internal**: Job payloads, status updates; 10M × 2KB ≈ 20 GB/day
- **Worker↔Scheduler**: Heartbeats, job claims; negligible

### Cache
- **Due jobs**: Redis sorted set (score=run_at); hot jobs
- **Worker registry**: Redis (active workers, last heartbeat)
- **Lock**: Redis for distributed lock (prevent duplicate execution)

---

## 3. API Design

### REST Endpoints

```
# Jobs
POST   /api/v1/jobs
Body: {
  "type": "one_time" | "recurring",
  "schedule": "2025-03-11T15:00:00Z" | "0 0 * * *",  // cron for recurring
  "handler": "email_sender",
  "payload": { "to": "...", "subject": "..." },
  "priority": 1-10,           // 10 = highest
  "timeout_seconds": 300,
  "retry_policy": {
    "max_retries": 3,
    "backoff": "exponential",
    "initial_delay_seconds": 60
  },
  "dependencies": ["job_id_1", "job_id_2"]   // DAG: run after these
}
Response: { "job_id": "job_xxx", "status": "PENDING" }

GET    /api/v1/jobs/:id
Response: { "job_id", "status", "created_at", "next_run_at", "last_run_at" }

PUT    /api/v1/jobs/:id/cancel
Response: { "job_id", "status": "CANCELLED" }

GET    /api/v1/jobs
Query: status, type, limit, cursor
Response: { "jobs": [...], "next_cursor": "..." }

# Job Runs (Execution History)
GET    /api/v1/jobs/:id/runs
Query: limit=20
Response: { "runs": [{ "run_id", "status", "started_at", "finished_at", "error" }] }

GET    /api/v1/runs/:run_id
Response: { "run_id", "job_id", "status", "started_at", "finished_at", "output", "error" }

# DAG Jobs
POST   /api/v1/jobs/dag
Body: {
  "jobs": [
    { "id": "job_a", "handler": "...", "payload": {...} },
    { "id": "job_b", "handler": "...", "dependencies": ["job_a"] },
    { "id": "job_c", "handler": "...", "dependencies": ["job_a"] }
  ],
  "schedule": "0 0 * * *"
}
Response: { "dag_id": "dag_xxx", "job_ids": ["job_a", "job_b", "job_c"] }
```

---

## 4. Data Model / Database Schema

### Database Choice
- **Jobs**: PostgreSQL (ACID, complex queries)
- **Job runs**: PostgreSQL (append-heavy) or Cassandra
- **Due jobs queue**: Redis Sorted Set (score = run_at timestamp)
- **Worker state**: Redis (ephemeral)
- **Distributed lock**: Redis (Redlock or single-node)
- **DAG state**: PostgreSQL

### Schema

**Jobs (PostgreSQL)**
```sql
jobs (
  job_id UUID PRIMARY KEY,
  type VARCHAR(20),              -- one_time, recurring
  schedule VARCHAR(100),         -- ISO timestamp or cron
  handler VARCHAR(100),
  payload JSONB,
  priority INT DEFAULT 5,
  timeout_seconds INT DEFAULT 300,
  max_retries INT DEFAULT 3,
  retry_backoff VARCHAR(20),    -- exponential, linear
  initial_delay_seconds INT DEFAULT 60,
  status VARCHAR(20),            -- PENDING, CANCELLED
  next_run_at TIMESTAMP,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
)

-- For recurring: next_run_at updated after each run
-- For one_time: next_run_at = schedule
```

**Job Dependencies (PostgreSQL)**
```sql
job_dependencies (
  job_id UUID,
  depends_on_job_id UUID,
  PRIMARY KEY (job_id, depends_on_job_id)
)
```

**Job Runs (PostgreSQL)**
```sql
job_runs (
  run_id UUID PRIMARY KEY,
  job_id UUID,
  status VARCHAR(20),            -- PENDING, RUNNING, SUCCEEDED, FAILED
  worker_id VARCHAR(100),
  started_at TIMESTAMP,
  finished_at TIMESTAMP,
  error TEXT,
  attempt INT,
  created_at TIMESTAMP
)

CREATE INDEX idx_runs_job_status ON job_runs(job_id, status);
CREATE INDEX idx_runs_created ON job_runs(created_at);
```

**Dead Letter Queue (PostgreSQL or separate table)**
```sql
failed_jobs (
  run_id UUID PRIMARY KEY,
  job_id UUID,
  error TEXT,
  payload JSONB,
  failed_at TIMESTAMP,
  retry_count INT
)
```

**Redis Structures**
```
# Due jobs (sorted set)
due_jobs: score=run_at_timestamp, member=job_id
ZRANGEBYSCORE due_jobs 0 <current_timestamp> LIMIT 100

# Job lock (prevent duplicate execution)
job_lock:{job_id}: timeout 5 minutes
SET job_lock:{job_id} worker_id NX EX 300

# Worker heartbeat
worker:{worker_id}: last_seen timestamp
SET worker:{worker_id} <timestamp> EX 60
```

---

## 5. High-Level Architecture

### ASCII Architecture Diagram

```
                                    ┌─────────────────────────────────────────────────────────────┐
                                    │                    API CLIENTS                              │
                                    │  (Create jobs, query status, cancel)                          │
                                    └───────────────────────────┬─────────────────────────────────┘
                                                                  │
                                                                  │ HTTPS
                                                                  ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                         API GATEWAY                                                           │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
                                                                  │
                                                                  ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                         JOB SCHEDULER SERVICE                                                 │
│                                                                                                               │
│  ┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐    │
│  │                    SCHEDULER LEADER (Leader Election via etcd/ZooKeeper)                              │    │
│  │  - Polls for due jobs every 10s                                                                       │    │
│  │  - Pushes due jobs to queue                                                                          │    │
│  │  - Computes next_run_at for recurring jobs                                                          │    │
│  │  - Handles DAG: topological sort, enqueue when deps satisfied                                         │    │
│  └─────────────────────────────────────────────────────────────────────────────────────────────────────┘    │
│                                                                                                               │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐                                         │
│  │  Job API        │    │  Queue Service   │    │  Run Tracker    │                                         │
│  │  (CRUD jobs)    │    │  (Enqueue due    │    │  (Status updates)│                                         │
│  │                 │    │   jobs)          │    │                 │                                         │
│  └────────┬────────┘    └────────┬────────┘    └────────┬────────┘                                         │
│           │                      │                      │                                                    │
└───────────┼──────────────────────┼──────────────────────┼───────────────────────────────────────────────────┘
            │                      │                      │
            ▼                      ▼                      ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                         DATA LAYER                                                            │
│                                                                                                               │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐                                                      │
│  │PostgreSQL│  │  Redis   │  │  etcd    │  │  Kafka   │                                                      │
│  │ (Jobs,   │  │ (Queue,  │  │ (Leader  │  │ (DLQ,    │                                                      │
│  │  Runs)   │  │  Lock)   │  │  Election)│  │  Events) │                                                      │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘                                                      │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
                                                                  │
                                                                  │ Poll / Consume
                                                                  ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                         WORKER POOL                                                           │
│                                                                                                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                        │
│  │  Worker 1   │  │  Worker 2   │  │  Worker 3   │  │  Worker N   │  │  ...        │                        │
│  │  - Claim job│  │  - Claim job │  │  - Claim job │  │  - Claim job │  │             │                        │
│  │  - Execute  │  │  - Execute  │  │  - Execute  │  │  - Execute  │  │             │                        │
│  │  - Update   │  │  - Update   │  │  - Update   │  │  - Update   │  │             │                        │
│  │    status   │  │    status   │  │    status   │  │    status   │  │             │                        │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘                        │
│                                                                                                               │
│  Process isolation: Each job runs in container/process; timeout kill                                          │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                    JOB STATE MACHINE                                                           │
│                                                                                                               │
│     ┌─────────┐     claim      ┌─────────┐    success    ┌───────────┐                                       │
│     │ PENDING │───────────────▶│ RUNNING │──────────────▶│ SUCCEEDED │                                       │
│     └─────────┘                └────┬────┘               └───────────┘                                       │
│          │                           │                                                                        │
│          │ cancel                    │ failure                                                                │
│          ▼                           ▼                                                                        │
│     ┌───────────┐               ┌─────────┐    retries    ┌─────────┐                                        │
│     │ CANCELLED │               │ FAILED  │◀──────────────│ RUNNING │                                        │
│     └───────────┘               └────┬────┘   (retry)     └─────────┘                                        │
│                                     │                                                                        │
│                                     │ max retries exceeded                                                   │
│                                     ▼                                                                        │
│                                ┌─────────┐                                                                   │
│                                │   DLQ   │                                                                   │
│                                └─────────┘                                                                   │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

### Component Descriptions
- **Scheduler Leader**: Single leader (election); polls for due jobs; enqueues
- **Job API**: CRUD jobs; validates cron; stores in DB
- **Queue Service**: Priority queue (Redis); workers consume
- **Worker Pool**: Claim job, execute, update status; horizontal scale
- **Run Tracker**: Persists status; triggers retries; DLQ for exhausted retries

---

## 6. Detailed Component Design

### 6.1 Job Queue (Priority Queue)

**Backed by Redis**
- **Sorted Set**: `due_jobs` with score = (priority * -1) for priority (higher first) or run_at for time
- **Alternative**: Separate queues per priority (high, medium, low); workers poll high first
- **List**: LPUSH job_id, BRPOP for workers (simple FIFO)
- **Priority**: Use multiple lists: high_queue, medium_queue, low_queue; worker checks high first

**Alternative: Kafka**
- Topic: `jobs_to_run`
- Partition by job_id for ordering
- Consumer group: workers
- Retry: Dead letter topic after N failures

**Alternative: Database**
- `SELECT * FROM jobs WHERE next_run_at <= NOW() AND status='PENDING' ORDER BY priority DESC LIMIT 100 FOR UPDATE SKIP LOCKED`
- `SKIP LOCKED` prevents two workers from claiming same job
- Update status to RUNNING, set worker_id

### 6.2 Job Execution (Worker Pool)

**Claiming a Job**
- **Redis**: BRPOP from queue; atomic
- **DB**: UPDATE jobs SET status='RUNNING', worker_id=? WHERE job_id=? AND status='PENDING' (conditional update)
- **Distributed lock**: SET lock_key NX EX timeout; only one worker gets lock

**Execution**
- **Process isolation**: Run in subprocess, container, or Lambda
- **Timeout**: Kill after timeout_seconds
- **Heartbeat**: Worker sends heartbeat; if missed, job considered stuck; re-queue

**Completion**
- Update job_runs: status=SUCCEEDED/FAILED, finished_at
- For recurring: Compute next_run_at, update jobs
- For one_time: Mark complete
- Release lock

### 6.3 Scheduling (Cron, Delayed Execution)

**Cron Parsing**
- Use library (e.g., cron-parser) to compute next run from "0 0 * * *"
- Store next_run_at in DB
- Scheduler: SELECT jobs WHERE next_run_at <= NOW()
- After run: next_run_at = cron_parser.next(schedule)

**Delayed Execution (One-Time)**
- schedule = "2025-03-11T15:00:00Z"
- next_run_at = schedule
- Same poll logic

**Polling Interval**
- Scheduler runs every 10-60 seconds
- Trade-off: Lower interval = more DB load; higher = more delay
- For sub-minute precision: Use delay queue (Redis, SQS) with second-level granularity

### 6.4 Distributed Coordination

**Leader Election**
- **etcd/ZooKeeper**: Create ephemeral node; leader holds lease
- **Redis**: SET scheduler_leader worker_id NX EX 30; leader renews before expiry
- **Raft**: Built into etcd
- Only leader runs scheduler loop; followers standby

**Prevent Duplicate Execution**
- **Distributed lock**: Before run, worker acquires lock on job_id
- **Optimistic locking**: UPDATE ... WHERE version=X; if 0 rows, another worker got it
- **Idempotency**: Job handler should be idempotent (at-least-once)

### 6.5 Job State Machine

- **PENDING**: Created, waiting for run_at
- **RUNNING**: Worker claimed, executing
- **SUCCEEDED**: Completed successfully
- **FAILED**: Failed; will retry or go to DLQ
- **CANCELLED**: User cancelled
- **DLQ**: Max retries exceeded; manual intervention

### 6.6 Retry with Backoff

- **Exponential**: delay = initial_delay * 2^attempt
- **Linear**: delay = initial_delay * (attempt + 1)
- **Max retries**: 3 (configurable)
- **Implementation**: On failure, create new job_run with attempt+1, schedule run_at = now + delay
- **Backoff cap**: e.g., max 1 hour

### 6.7 DAG Execution (Job Dependencies)

**Topological Sort**
- Build graph: job -> [dependencies]
- Sort: Jobs with no deps first
- Execution order: A, B, C (B and C depend on A)

**Runtime**
- When job A completes: Check jobs that depend on A
- For each dependent job B: Are all deps of B complete? If yes, enqueue B
- Store: job_dependencies, job_run status per job in DAG
- **DAG run**: Single execution of entire DAG; each job has run_id; track per-job status

**Parallelism**
- Jobs with same deps can run in parallel (e.g., B and C both depend only on A)
- Worker pool naturally parallelizes

**Failure**
- If job A fails: Jobs B, C (depend on A) are CANCELLED or BLOCKED
- Option: Retry A; on success, enqueue B, C
- Option: Fail entire DAG

### 6.8 Dead Letter Queue (DLQ)

- **When**: Max retries exceeded
- **Storage**: failed_jobs table or Kafka topic
- **Alert**: Notify on new DLQ entry
- **Manual**: Dashboard to retry or discard
- **Re-queue**: Admin can move back to pending with reset retry count

---

## 7. Scaling

### Sharding
- **Jobs**: Shard by tenant_id or job_id hash
- **Runs**: Shard by job_id (co-locate with job)
- **Queue**: Partition by priority or tenant

### Horizontal Scaling
- **Workers**: Add more workers; they consume from same queue
- **Scheduler**: Only one leader; but API servers can scale
- **DB**: Read replicas for status queries

### Caching
- **Due jobs**: Redis sorted set; avoid repeated DB scans
- **Job config**: Cache in worker memory (handler, timeout)
- **Worker registry**: Redis for active workers

### Queue
- **Kafka**: High throughput; partition by job_type for parallelism
- **Redis**: Simpler; sufficient for 100s of jobs/sec
- **SQS**: Managed; visibility timeout for at-least-once

---

## 8. Failure Handling

### Component Failures
- **Scheduler leader dies**: New leader elected; resume polling
- **Worker dies mid-job**: Lock expires; job re-queued (at-least-once)
- **DB down**: Scheduler can't enqueue; workers can't update; queue in memory/Redis
- **Redis down**: Queue unavailable; use DB as fallback (slower)

### Redundancy
- **Scheduler**: Leader + standby; fast failover
- **Workers**: Stateless; any can run any job
- **DB**: Primary + replica; failover
- **Redis**: Cluster mode or replica

### Degradation
- **High load**: Queue depth grows; add workers
- **Slow jobs**: Timeout; kill; retry
- **DLQ**: Don't block; move to DLQ; alert

### Recovery
- **Stuck jobs**: Heartbeat timeout; re-queue
- **Orphaned RUNNING**: Cron job finds runs with started_at > 2*timeout ago; mark FAILED, re-queue

---

## 9. Monitoring & Observability

### Key Metrics
- **Queue depth**: Jobs waiting to run
- **Job latency**: Time from run_at to started_at
- **Execution time**: started_at to finished_at (p50, p99)
- **Success rate**: SUCCEEDED / (SUCCEEDED + FAILED)
- **DLQ size**: Jobs in DLQ
- **Worker count**: Active workers
- **Scheduler lag**: How far behind is next_run_at processing

### Alerts
- **Queue depth > 10,000**
- **Job latency > 5 minutes**
- **Success rate < 95%**
- **DLQ has new entries**
- **No leader** (scheduler down)
- **Worker heartbeat stopped** (all workers down)

### Tracing
- **Trace ID**: job_id or run_id across scheduler → queue → worker
- **Logs**: Structured; job_id, run_id, status, duration

### Dashboards
- **Throughput**: Jobs/sec over time
- **Latency distribution**: Histogram
- **Error rate**: By handler, by error type

---

## 10. Interview Tips

### Follow-up Questions
- "How would you achieve exactly-once execution?"
- "How do you handle a job that runs for 24 hours?"
- "How would you schedule 100M jobs for the same time (e.g., New Year campaign)?"
- "How do you prevent a single slow job from blocking the queue?"
- "How would you add job prioritization with preemption (kill low-priority for high)?"

### Common Mistakes
- **No leader election**: Multiple schedulers = duplicate job execution
- **No distributed lock**: Two workers run same job
- **Synchronous execution**: Block scheduler; use queue
- **Ignoring retries**: Failures happen; need backoff
- **DAG complexity**: Start with linear chain; then DAG

### Key Points to Emphasize
- **At-least-once**: Job must be idempotent; design for duplicate runs
- **Leader election**: Single scheduler; prevent duplicate scheduling
- **Distributed lock**: Prevent duplicate execution
- **Priority queue**: Higher priority first
- **DAG**: Topological sort; enqueue when deps satisfied
- **DLQ**: Don't lose failed jobs; manual intervention

---

## Appendix: Deep Dive Topics

### A. Cron Expression Examples
| Expression | Meaning |
|------------|---------|
| `0 0 * * *` | Midnight daily |
| `*/5 * * * *` | Every 5 minutes |
| `0 9 * * 1-5` | 9am weekdays |
| `0 0 1 * *` | 1st of month |
| `0 0 * * 0` | Sunday midnight |

### B. Leader Election (etcd)
```
1. Create lease with TTL (e.g., 30s)
2. Try to create key /scheduler/leader with lease
3. If success: I am leader; renew lease before expiry
4. If fail: Watch key; when deleted, retry
5. On crash: Lease expires; key deleted; others can become leader
```

### C. Job Claim Patterns
- **Database**: `UPDATE jobs SET worker_id=?, status='RUNNING' WHERE job_id=? AND status='PENDING'`; check rows affected
- **Redis**: `LPOP queue` (atomic)
- **Kafka**: Consumer group; each message to one consumer
- **SQS**: ReceiveMessage with visibility timeout; extend if job long

### D. DAG Example
```
A (no deps) → B, C (depend on A) → D (depends on B, C)
Execution: A runs first; when A done, B and C run in parallel; when both done, D runs
```

### E. Exactly-Once Considerations
- **Challenges**: Duplicate delivery, duplicate processing
- **Approaches**: Idempotent handlers, deduplication table, transactional outbox
- **Trade-off**: Complexity vs at-least-once + idempotency (simpler)

### F. Worker Heartbeat and Stuck Job Detection
- **Heartbeat**: Worker sends every 30s while job running
- **Timeout**: If no heartbeat for 2x interval, job considered stuck
- **Action**: Mark run FAILED; re-queue job (new run)
- **Orphan cleanup**: Cron job finds RUNNING with started_at > 2*timeout; re-queue

### G. Priority Queue Implementation (Redis)
- **Option 1**: Multiple lists (high, medium, low); worker BRPOP high, medium, low
- **Option 2**: Sorted set with score = (run_at << 32) | (priority)
- **Option 3**: Single list; worker fetches N, picks highest priority

### H. Cron Next-Run Calculation
- **Library**: cron-parser, node-cron, or similar
- **Input**: "0 0 * * *" (midnight daily)
- **Output**: Next N run times
- **Timezone**: Store user timezone; compute in that TZ

### I. Job Timeout and Kill
- **Subprocess**: Worker spawns child process for job
- **Timer**: Start timer when job starts
- **On timeout**: Send SIGKILL to child; mark run FAILED
- **Cleanup**: Orphan processes; use process group for cleanup

### J. High-Volume Scheduling (100M jobs at same time)
- **Shard by time**: Distribute across workers by job_id hash
- **Pre-warm queue**: Don't enqueue all at once; batch enqueue
- **Rate limit**: Throttle PSP/DB if downstream bottleneck
- **Prioritize**: Process high-priority first; low can wait

### K. Job Result Storage
- **Option 1**: Store in job_runs (output column)
- **Option 2**: Separate results table (run_id, key, value)
- **Option 3**: Object storage (S3) for large outputs

### L. Process Isolation
- **Container**: Each job runs in Docker container; resource limits
- **Subprocess**: Simpler; same OS; less isolation
- **Lambda**: Serverless; cold start; 15 min limit
- **Sandbox**: gVisor, Firecracker for stronger isolation; use for untrusted job code

### M. Metrics Export
- **Prometheus**: Scrape job_scheduler_* metrics from workers