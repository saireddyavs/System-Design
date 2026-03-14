# Actor Model

> Staff+ Engineer Level — FAANG Interview Deep Dive

---

## 1. Concept Overview

### Definition

The **Actor Model** is a concurrency paradigm where computation is performed by **actors** — lightweight, independent units that communicate exclusively through **asynchronous message passing**. Each actor has a mailbox (message queue), processes messages one at a time, and can create new actors or send messages to other actors.

### Purpose

- **Concurrency without shared state**: No locks, no race conditions
- **Location transparency**: Actors can be local or remote; same programming model
- **Fault isolation**: One actor's failure doesn't crash others
- **Scalability**: Natural fit for distributed systems

### Problems Solved

| Problem | Solution |
|---------|----------|
| Shared memory races | No shared state; message passing only |
| Deadlocks | No blocking; async messages |
| Thread overhead | Lightweight actors (millions per machine) |
| Distributed coordination | Same model for local and remote actors |

---

## 2. Real-World Motivation

### WhatsApp (Erlang)

- **2M connections per server** using Erlang actors
- Each connection = lightweight process (actor)
- Fault tolerance via supervision trees
- Hot code swapping for zero-downtime deployments

### Microsoft Orleans (Halo, Xbox)

- **Virtual actors** (grains): identity-based, no explicit creation
- Used for Halo game backend, Xbox Live
- Automatic activation/deactivation; scales to millions of grains

### Akka (JVM)

- Used by **LinkedIn, Lightbend, Verizon**
- Actor model for Scala/Java
- Akka Cluster for distributed actors
- Akka Streams for backpressure

### Discord

- Erlang/Elixir for real-time messaging
- Millions of concurrent connections per node

---

## 3. Architecture Diagrams

### Actor Hierarchy and Supervision Tree

```
                    ┌─────────────────────┐
                    │   Root Supervisor   │
                    │   (top-level)       │
                    └──────────┬──────────┘
                               │
        ┌──────────────────────┼──────────────────────┐
        │                      │                      │
        ▼                      ▼                      ▼
┌───────────────┐      ┌───────────────┐      ┌───────────────┐
│ Connection    │      │ Auth          │      │ Message       │
│ Supervisor    │      │ Supervisor    │      │ Supervisor    │
└───────┬───────┘      └───────┬───────┘      └───────┬───────┘
        │                      │                      │
   ┌────┴────┐            ┌────┴────┐            ┌────┴────┐
   ▼    ▼    ▼            ▼    ▼    ▼            ▼    ▼    ▼
 [A1] [A2] [A3]         [A1] [A2]              [A1] [A2] [A3]
 (workers)               (workers)              (workers)
```

### Message Passing Flow

```
    Actor A                    Actor B
    ┌─────────┐                ┌─────────┐
    │ State   │                │ State   │
    │ Mailbox │                │ Mailbox │
    └────┬────┘                └────┬────┘
         │                          │
         │  send(B, {get_user, 123}) │
         │─────────────────────────>│
         │                          │
         │  (async, non-blocking)    │  process message
         │                          │  send(A, {user, data})
         │<─────────────────────────│
         │                          │
    process reply                   │
```

### Actor Mailbox

```
    ┌─────────────────────────────────────────┐
    │              Actor Mailbox               │
    │  ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐ │
    │  │Msg 1│ │Msg 2│ │Msg 3│ │Msg 4│ │Msg 5│ │  FIFO queue
    │  └──┬──┘ └─────┘ └─────┘ └─────┘ └─────┘ │
    │     │                                     │
    │     ▼  Actor processes one at a time      │
    │  ┌─────────────────────────────────────┐ │
    │  │  receive: pattern match on message  │ │
    │  │  handle: update state, send reply   │ │
    │  └─────────────────────────────────────┘ │
    └─────────────────────────────────────────┘
```

### Supervision Tree — Restart Strategy

```
    Supervisor (OneForOne or AllForOne)
         │
         │  Child crashes
         │  ────────────────────────────────
         │  OneForOne: Restart only crashed child
         │  AllForOne: Restart all siblings
         │
         ▼
    ┌─────────┐  ┌─────────┐  ┌─────────┐
    │ Child 1 │  │ Child 2 │  │ Child 3 │
    │ (crash) │  │ (ok)    │  │ (ok)    │
    └─────────┘  └─────────┘  └─────────┘
         │
         │  Restart
         ▼
    ┌─────────┐
    │ Child 1 │  (new instance)
    │ (fresh) │
    └─────────┘
```

---

## 4. Core Mechanics

### Actor Properties

| Property | Description |
|----------|-------------|
| **Encapsulation** | State is private; only accessible via messages |
| **Single-threaded** | Processes one message at a time (per actor) |
| **Asynchronous** | Send is non-blocking; no wait for reply |
| **Address** | Each actor has unique address (mailbox) |

### Message Passing

- **Fire-and-forget**: `actor ! message` — no reply expected
- **Request-reply**: Send message, include sender ref; receiver sends reply to sender
- **Ordering**: Messages from A to B are delivered in order (per-sender FIFO)

### Supervision

- **Supervisor**: Watches child actors; decides restart strategy
- **OneForOne**: Restart only the failed child
- **AllForOne**: Restart all children when one fails
- **Restart limit**: Max restarts per time window; escalate if exceeded

### Erlang OTP Behaviours

- **gen_server**: Client-server pattern
- **gen_supervisor**: Supervision tree
- **gen_statem**: Finite state machine
- **gen_event**: Event handling

---

## 5. Numbers

| Metric | Typical Value |
|--------|---------------|
| Erlang process (actor) memory | ~2-3 KB per process |
| Actors per machine (Erlang) | 1-2 million |
| Message latency (local) | Microseconds |
| Message latency (remote) | Milliseconds (network) |
| WhatsApp connections/server | ~2 million |
| Orleans grains | Millions per cluster |

---

## 6. Tradeoffs

### Actor Model vs Threads

| Aspect | Actor Model | Threads |
|--------|-------------|---------|
| State | Isolated per actor | Shared (needs locks) |
| Concurrency | Message-driven | Preemptive/cooperative |
| Failure | Isolated; supervisor | Can crash entire process |
| Scalability | Millions of actors | Thousands of threads |
| Debugging | Message traces | Race conditions |

### Actor Model vs CSP (Go Channels)

| Aspect | Actor Model | CSP (Go) |
|--------|-------------|----------|
| Abstraction | Identity (actor ref) | Channel (anonymous) |
| Communication | Send to actor | Send to channel |
| State | In actor | In goroutine |
| Focus | Who receives | What is sent |
| Example | Erlang, Akka | Go, Clojure core.async |

### Actor Model vs Shared Memory

| Aspect | Actor Model | Shared Memory |
|--------|-------------|---------------|
| Synchronization | None (messages) | Locks, semaphores |
| Deadlock | Impossible (no blocking) | Possible |
| Distribution | Natural fit | Difficult |
| Performance | Message overhead | Direct access |

---

## 7. Variants / Implementations

### Erlang/OTP

- **Origin**: Ericsson, 1986; for telecom
- **Lightweight processes**: Scheduler in VM
- **Let it crash**: Fail fast; supervisor restarts
- **Hot code swap**: Upgrade without downtime

### Akka (JVM)

- **Scala/Java**: Typed and classic APIs
- **Cluster**: Distributed actors; sharding
- **Persistence**: Event sourcing for actors
- **Streams**: Backpressure, reactive

### Microsoft Orleans

- **Virtual actors (grains)**: Identity = activation
- **No explicit create**: Runtime activates on first message
- **Automatic deactivation**: Idle grains reclaimed
- **.NET**: C# ecosystem

### Akka vs Orleans

| Aspect | Akka | Orleans |
|--------|------|---------|
| Model | Explicit actor refs | Virtual (identity-based) |
| Creation | Explicit spawn | Implicit on first message |
| Location | Explicit placement | Runtime decides |
| Language | Scala, Java | C# |

---

## 8. Scaling Strategies

1. **Horizontal**: Distribute actors across nodes (Akka Cluster, Erlang distribution)
2. **Sharding**: Partition actors by key (e.g., user_id) across nodes
3. **Pool/Router**: Route messages to pool of workers
4. **Location transparency**: Same code for local/remote; runtime handles

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|------------|
| Actor crash | Only that actor affected | Supervisor restarts |
| Mailbox overflow | Backpressure or drop | Bounded mailbox, circuit breaker |
| Network partition | Remote messages fail | Timeout, retry, dead letters |
| Supervisor crash | Children orphaned | Hierarchical supervision |
| Message loss | Lost request | Idempotency, ack protocol |

---

## 10. Performance Considerations

- **Message serialization**: Remote actors need serialization (Protobuf, etc.)
- **Mailbox size**: Unbounded can cause OOM; use bounded or drop
- **Actor placement**: Colocate frequently communicating actors
- **Batching**: Batch messages to reduce overhead

---

## 11. Use Cases

| Use Case | Why Actors |
|----------|------------|
| Chat/messaging | One actor per connection; natural fit |
| Game backends | One actor per player/session |
| IoT | One actor per device |
| Real-time systems | Low latency, fault isolation |
| Event sourcing | Actor = aggregate; messages = events |

---

## 12. Comparison Tables

### Concurrency Models

| Model | Unit | Communication | Shared State |
|-------|------|---------------|--------------|
| **Actors** | Actor | Messages | No |
| **Threads** | Thread | Shared memory | Yes (locks) |
| **CSP (Go)** | Goroutine | Channels | No (via channels) |
| **Async/Await** | Task | Promises | No (single-threaded) |

### Actor Implementations

| Implementation | Language | Virtual Actors | Distribution |
|----------------|----------|----------------|--------------|
| Erlang | Erlang | No | Built-in |
| Akka | Scala, Java | No | Akka Cluster |
| Orleans | C# | Yes (grains) | Built-in |
| Akka Typed | Scala, Java | No | Akka Cluster |

---

## 13. Code / Pseudocode

### Erlang Actor (gen_server)

```erlang
-module(user_actor).
-behaviour(gen_server).

init([]) ->
    {ok, #{}}.

handle_call({get, UserId}, _From, State) ->
    User = maps:get(UserId, State, not_found),
    {reply, User, State};

handle_call({put, UserId, UserData}, _From, State) ->
    NewState = maps:put(UserId, UserData, State),
    {reply, ok, NewState}.

% Usage: gen_server:call(Pid, {get, 123})
```

### Akka Actor (Scala)

```scala
class UserActor extends Actor {
  var users: Map[String, User] = Map.empty

  def receive: Receive = {
    case GetUser(id) =>
      sender() ! users.get(id)
    case PutUser(id, user) =>
      users = users + (id -> user)
      sender() ! Ok
  }
}

// Usage: userActor ! GetUser("123")
```

### Supervision Tree (Pseudocode)

```python
# Supervisor creates and monitors children
class ConnectionSupervisor(Supervisor):
    def start_children(self):
        for i in range(100):
            child = spawn(ConnectionActor, i)
            self.watch(child)  # Restart on crash

    def handle_failure(self, child, reason):
        if self.restart_count < 5:
            self.restart(child)  # OneForOne
        else:
            self.escalate()  # Give up, let parent handle
```

### Message Passing (Pseudocode)

```python
class Actor:
    def __init__(self):
        self.mailbox = Queue()
        self.state = {}

    def send(self, target, message):
        target.mailbox.put((self, message))  # Include sender for reply

    def run(self):
        while True:
            sender, message = self.mailbox.get()
            response = self.handle(sender, message)
            if response and sender:
                sender.mailbox.put((self, response))
```

---

## 14. Interview Discussion

### Key Points

1. **Actors = encapsulation + message passing** — No shared state; no locks
2. **Supervision trees** — "Let it crash"; supervisor restarts failed actors
3. **Location transparency** — Same model for local and distributed
4. **WhatsApp, Orleans** — Real-world scale (2M connections/server)

### Common Questions

- **"Actor model vs threads?"** — Actors: no shared state, message passing, fault isolation. Threads: shared memory, locks, can deadlock.
- **"Actor model vs CSP (Go channels)?"** — Actors: identity-based (send to actor). CSP: channel-based (send to channel). Actors focus on who; CSP on what.
- **"How does Erlang achieve 2M connections per server?"** — Lightweight processes (~2KB each), async I/O, scheduler in VM, no blocking.
- **"What is a supervision tree?"** — Hierarchical structure; parent watches children, restarts on crash. OneForOne vs AllForOne.

### Red Flags

- Using actors for CPU-bound tasks (threads/pools better)
- Blocking inside actor (defeats purpose)
- Ignoring message ordering guarantees
