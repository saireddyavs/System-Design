# Module 9: Concurrency & Lock-Free Programming

---

## 1. Thread Pools

### Definition
A collection of pre-created threads that pick up tasks from a queue, avoiding the overhead of creating/destroying threads per request.

### Why Not Thread-Per-Request?
```
Thread-Per-Request:
  10,000 requests → 10,000 threads
  Each thread: ~1MB stack → 10GB RAM
  OS scheduling: ~microseconds per context switch × 10,000 = catastrophic

Thread Pool (size=100):
  10,000 requests → Queue → 100 threads process them
  Fixed memory, controlled concurrency
```

### Architecture
```
  ┌─────────────────┐     ┌─────────────┐
  │  Task Queue     │     │ Thread Pool │
  │ [task5][task4]  │ ──→ │ [T1] [T2]  │
  │ [task3][task2]  │     │ [T3] [T4]  │
  │ [task1]         │     │            │
  └─────────────────┘     └─────────────┘
  Tasks wait in queue      Fixed number of threads
                           pick tasks and execute
```

### Sizing
```
CPU-bound tasks:  pool_size = num_cores
I/O-bound tasks:  pool_size = num_cores × (1 + wait_time/compute_time)
  Example: 8 cores, 90% I/O wait → 8 × (1 + 9) = 80 threads
```

### Real Systems
Java (ExecutorService), Go (goroutine scheduler), Tomcat, NGINX (worker processes), Node.js (libuv thread pool for file I/O)

---

## 2. Work Stealing

### Definition
Each thread has its own task queue. When a thread's queue is empty, it "steals" tasks from another thread's queue — enabling dynamic load balancing.

### How It Works
```
Thread 1 Queue: [T1a][T1b][T1c][T1d]  ← busy
Thread 2 Queue: [T2a]                   ← almost done
Thread 3 Queue: []                      ← idle, STEALS from Thread 1

After stealing:
Thread 1 Queue: [T1a][T1b][T1c]
Thread 2 Queue: [T2a]
Thread 3 Queue: [T1d]  ← stolen from Thread 1's tail
```

### Key Detail
Owner pops from HEAD (LIFO — cache-warm tasks). Thief steals from TAIL (FIFO — older, likely larger tasks). This minimizes contention.

### Real Systems
Java Fork/Join Framework, Go scheduler, Tokio (Rust), .NET ThreadPool, Intel TBB

### Summary
Work stealing dynamically balances load across threads by letting idle threads steal from busy ones. Owner: LIFO (cache-warm). Thief: FIFO (large tasks).

---

## 3. Event Loop

### Definition
A single-threaded loop that waits for events (I/O completion, timers) and dispatches callbacks. Handles thousands of concurrent connections with one thread.

### How It Works
```
while (true) {
    events = poll(registered_fds, timeout)  // epoll/kqueue
    for event in events:
        callback = lookup(event.fd)
        callback(event.data)
}
```

### Visual
```
       ┌──────────────────────────┐
       │      Event Loop          │
       │                          │
  ────→│  1. Poll for events      │
  │    │  2. Execute callbacks    │
  │    │  3. Process timers       │
  │    │  4. Back to step 1       │
  │    └──────────────────────────┘
  │         │
  │         ├── Callback: handle HTTP request
  │         ├── Callback: DB response arrived
  │         └── Callback: timer fired
  │
  └── Non-blocking I/O: register interest, get notified
```

### Why Single Thread Works
```
Thread-per-connection:   1 thread × 10,000 connections = 10,000 threads
Event loop:              1 thread × 10,000 connections = 1 thread

The thread never blocks on I/O. It registers interest and gets
notified when data is ready. Between events, it handles other
connections.
```

### Warning: Don't Block the Event Loop
```
// BAD: blocks the loop, no other connections served
const result = heavyComputation()  // 5 seconds of CPU

// GOOD: offload to worker thread
workerPool.run(heavyComputation).then(callback)
```

### Real Systems
Node.js (libuv), Nginx, Redis (single-threaded), Python asyncio, Java NIO

---

## 4. Reactor Pattern

### Definition
An event-driven architecture where a single dispatcher (reactor) demultiplexes I/O events and dispatches them to registered handlers.

### Architecture
```
  ┌─────────────────────────────────────────┐
  │              Reactor                    │
  │                                         │
  │  ┌──────────────┐                      │
  │  │ Demultiplexer│ (epoll/select/kqueue)│
  │  │ Wait for I/O │                      │
  │  └──────┬───────┘                      │
  │         │ events                        │
  │         ▼                               │
  │  ┌──────────────┐                      │
  │  │  Dispatcher  │                      │
  │  │ Route event  │                      │
  │  │ to handler   │                      │
  │  └──┬───┬───┬───┘                     │
  │     │   │   │                           │
  │     ▼   ▼   ▼                           │
  │  [H1] [H2] [H3]  ← Event Handlers     │
  └─────────────────────────────────────────┘
```

### Single Reactor vs Multi-Reactor
```
Single Reactor:  1 thread handles accept + read/write (Redis)
Multi-Reactor:   Main reactor accepts, sub-reactors handle I/O (Netty)

  Main Reactor (accept) → Sub-Reactor 1 (handle connections 1-1000)
                        → Sub-Reactor 2 (handle connections 1001-2000)
```

### Real Systems
Netty (Java), libuv (Node.js), Twisted (Python), Redis

---

## 5. Proactor Pattern

### Definition
Like Reactor, but the OS performs the I/O operation asynchronously and notifies the handler when it's COMPLETE (not when it's ready).

### Reactor vs Proactor
```
Reactor:
  1. OS notifies: "Socket is READY to read"
  2. Handler reads data (application does I/O)

Proactor:
  1. Application requests: "Read 1024 bytes from socket"
  2. OS performs the read in background
  3. OS notifies: "Read COMPLETE, here's the data"
  (Application never calls read() — OS did it)
```

### Comparison

| | Reactor | Proactor |
|-|---------|----------|
| I/O performed by | Application (non-blocking) | OS/kernel (async) |
| Notification | "Ready to read" | "Read complete" |
| Complexity | Simpler | More complex |
| OS support | Universal (epoll, kqueue) | Limited (Windows IOCP, io_uring) |
| Example | Nginx, Redis | Windows IOCP, Boost.Asio |

### Real Systems
Windows IOCP, io_uring (Linux 5.1+), Boost.Asio (C++)

---

## 6. Lock-Free Programming

### Definition
Concurrent data structures and algorithms that use atomic operations (CAS) instead of locks, guaranteeing that at least one thread makes progress.

### Lock vs Lock-Free vs Wait-Free
```
Lock-based:   Threads wait for lock → all blocked if holder stalls
Lock-free:    At least ONE thread always makes progress
Wait-free:    EVERY thread makes progress (strongest guarantee)
```

### Why Lock-Free?
```
Problems with locks:
  1. Priority inversion (low-priority thread holds lock)
  2. Deadlock (two threads wait for each other)
  3. Convoying (all threads queue behind slow lock holder)
  4. Not composable (combining two locked operations is hard)

Lock-free avoids all of these.
```

### CAS Loop Pattern
```
do {
    old_value = read(location)
    new_value = compute(old_value)
} while (!CAS(location, old_value, new_value))
// If someone changed location between read and CAS, retry
```

### Real Systems
Java ConcurrentHashMap, Go sync/atomic, LMAX Disruptor, Linux kernel (RCU)

---

## 7. CAS (Compare-And-Swap)

### Definition
An atomic CPU instruction: "If the value at address X is still V, set it to V'. Otherwise, do nothing." Returns whether it succeeded.

### How It Works
```
CAS(address, expected, new_value):
  ATOMICALLY:
    if *address == expected:
        *address = new_value
        return true
    else:
        return false  // someone else changed it

This is a SINGLE atomic CPU instruction (CMPXCHG on x86).
```

### Example: Atomic Counter
```
// Lock-based:              // CAS-based (lock-free):
lock()                      do {
counter++                       old = counter
unlock()                        new = old + 1
                            } while (!CAS(&counter, old, new))
```

### Performance
```
Low contention:  CAS is faster than lock (no kernel syscall)
High contention: CAS wastes CPU (spinning retries)
                 Lock is better (threads sleep)
```

### Summary
CAS is the foundation of lock-free programming. It atomically compares-and-swaps a value. Under low contention it outperforms locks; under high contention it wastes CPU with retries.

---

## 8. ABA Problem

### Definition
A flaw in CAS-based algorithms: the value changes from A→B→A. CAS sees A and thinks nothing changed, but the state may be different.

### Example
```
Thread 1: Reads top of stack = A
Thread 1: Suspended...

Thread 2: Pop A, Pop B, Push A (different A or reallocated memory)

Thread 1: CAS(top, A, new_node) → SUCCEEDS
          But the stack is now corrupted!

  Before: A → B → C
  Thread 2: pop A, pop B → C.  Push new A → A' → C
  Thread 1 thinks: top is still A, CAS succeeds
  But B is gone, and A might point to freed memory!
```

### Solutions
```
1. TAGGED POINTERS: Pair value with a counter
   CAS((A, v1), (A, v1), (B, v2))
   Counter changes even if value returns to A

2. HAZARD POINTERS: Track which nodes threads are reading
   Don't free a node if any thread has it as hazard pointer

3. EPOCH-BASED RECLAMATION: Defer freeing until all readers leave epoch
```

### Summary
ABA occurs when CAS can't distinguish "unchanged" from "changed back." Fix with version counters (tagged pointers), hazard pointers, or epoch-based reclamation.

---

## 9. Hazard Pointers

### Definition
A technique for safe memory reclamation in lock-free data structures. Each thread publishes pointers it's currently reading; those nodes can't be freed.

### How It Works
```
Thread 1 reads node X:
  1. Publish hazard pointer: HP[Thread1] = X
  2. Read X safely
  3. Clear: HP[Thread1] = null

Thread 2 wants to free X:
  1. Check all hazard pointers
  2. HP[Thread1] = X → X is protected, don't free
  3. Add X to "retire list"
  4. Periodically scan retire list against all HPs
  5. Free nodes not in any HP
```

### Complexity
- Per-thread overhead: O(K) hazard pointers (K typically 1-2)
- Reclamation scan: O(N × K) where N = threads

### Real Systems
Facebook Folly (C++), Java lock-free queues, some database internals

---

## 10. Epoch-Based Reclamation (EBR)

### Definition
A memory reclamation scheme where threads enter "epochs" and freed memory is only reclaimed when all threads have left the epoch in which the memory was freed.

### How It Works
```
Global epoch: 2

Thread A: enters epoch 2, reads data
Thread B: frees node X, adds to "epoch 2 retire list"
Thread C: enters epoch 2, reads data

All threads leave epoch 2 → advance to epoch 3
→ Safe to free everything in epoch 1's retire list
  (epoch 2 retire list freed when all leave epoch 3)

Key: Nothing freed until ALL threads have advanced past the epoch.
```

### Visual
```
Epoch 0: [garbage from epoch 0 freed after all threads pass epoch 1]
Epoch 1: [garbage from epoch 1 freed after all threads pass epoch 2]
Epoch 2: ← current epoch (threads operating here)
```

### Compared to Hazard Pointers

| | Hazard Pointers | Epoch-Based |
|-|-----------------|-------------|
| Overhead per access | Publish pointer (fence) | Check/update epoch |
| Reclamation speed | Immediate once safe | Delayed (batch) |
| Stalled thread impact | Only protects its nodes | Blocks ALL reclamation |
| Complexity | Medium | Simpler |

### Real Systems
Crossbeam (Rust), Linux RCU (variant), some lock-free data structures

### Summary
EBR batches memory reclamation by epoch. Free memory is only released when all threads have advanced past the epoch. Simpler than hazard pointers but a stalled thread blocks all reclamation.
