# Module 11: Garbage Collection & Memory Management

---

## 1. Garbage Collection (Overview)

### Definition
Automatic memory management where the runtime reclaims memory occupied by objects no longer reachable by the program.

### Why It Matters for System Design
```
GC pauses can cause:
  - Latency spikes (p99 latency)
  - Missed heartbeats → node declared dead
  - Distributed lock expiry → split brain
  - Kafka consumer timeouts → rebalance

Discord switched Go → Rust partly because of GC pauses.
Instagram disabled Python's cyclic GC to save memory.
```

### GC Strategies Overview
```
┌────────────────────────────────────────────────┐
│ Strategy           │ How it works              │
├────────────────────┼───────────────────────────┤
│ Reference Counting │ Count refs, free at zero  │
│ Mark-and-Sweep     │ Trace reachable, free rest│
│ Generational       │ Young/Old gen, scan young │
│ Stop-the-World     │ Pause all threads to GC   │
│ Incremental        │ Small GC steps interleaved│
│ Concurrent         │ GC runs alongside app     │
└────────────────────┴───────────────────────────┘
```

---

## 2. Reference Counting

### Definition
Each object tracks how many references point to it. When the count drops to zero, the object is immediately freed.

### How It Works
```
a = new Object()      // refcount = 1
b = a                  // refcount = 2
b = null               // refcount = 1
a = null               // refcount = 0 → FREE immediately
```

### Visual
```
  a ──→ [Object refcount=2] ←── b
  a = null → refcount = 1
  b = null → refcount = 0 → FREED
```

### The Cycle Problem
```
  A.next = B     (B refcount = 1)
  B.next = A     (A refcount = 1)
  
  Drop all external refs → both still refcount = 1
  MEMORY LEAK — neither ever freed!

  ┌───┐     ┌───┐
  │ A │ ──→ │ B │
  │   │ ←── │   │
  └───┘     └───┘
  refcount=1  refcount=1  ← leaked!
```

### Solutions to Cycles
- **Weak references**: Don't increment refcount (Python, Swift)
- **Cycle detector**: Periodic scan for unreachable cycles (Python)
- **Ownership rules**: Compiler prevents cycles (Rust)

### Tradeoffs

| Pros | Cons |
|------|------|
| Immediate reclamation | Cannot handle cycles |
| Predictable latency | Overhead on every ref assign |
| Simple mental model | Atomic refcount = expensive in MT |

### Real Systems
Python (primary GC + cycle detector), Swift/Objective-C (ARC), Rust (Rc/Arc), COM (Windows)

---

## 3. Mark-and-Sweep

### Definition
A GC algorithm with two phases: MARK all objects reachable from roots, then SWEEP (free) all unmarked objects.

### How It Works
```
Phase 1 - MARK:
  Start from GC roots (stack variables, globals, registers)
  Traverse all references, mark each visited object

Phase 2 - SWEEP:
  Scan entire heap
  Free any object NOT marked
  Clear all marks for next cycle

Roots: [stack] [globals] [registers]
         │         │
         ▼         ▼
        [A] ───→ [B] ───→ [C]    ← reachable (marked)
        
        [D] ───→ [E]              ← unreachable (swept/freed)
```

### Visual
```
Before GC:
  Heap: [A✓][B✓][C✓][D][E][F✓]
               └─marked─┘    └─from root

After Mark:  A✓ B✓ C✓ D✗ E✗ F✓
After Sweep: A  B  C  [free] [free] F
```

### Tradeoffs

| Pros | Cons |
|------|------|
| Handles cycles correctly | Stop-the-world pause during GC |
| No per-reference overhead | Heap fragmentation |
| Complete reclamation | Pause proportional to heap size |

### Real Systems
Early Java GC, Lua, most language runtimes as a base algorithm

---

## 4. Generational GC

### Definition
Divides the heap into generations (Young, Old). Most objects die young, so GC focuses on the young generation (frequent, fast) and rarely scans the old generation.

### The Generational Hypothesis
```
"Most objects die young."

Observation: 90%+ of objects become garbage shortly after creation.
Strategy:   Collect young gen frequently (cheap).
            Promote survivors to old gen.
            Collect old gen rarely (expensive but infrequent).
```

### How It Works
```
┌──────── Young Generation ─────────┐
│ Eden     │ Survivor 0 │ Survivor 1│
│ (new     │ (survived   │ (survived │
│  objects)│  1 GC)      │  2+ GCs)  │
└──────────┴─────────────┴──────────┘
       │ objects that survive N GCs
       ▼
┌──────── Old Generation ───────────┐
│ Long-lived objects                 │
│ (collected infrequently)           │
└────────────────────────────────────┘

Minor GC: Collect Young Gen (~5ms)
Major GC: Collect Old Gen (~100ms+)
Full GC:  Collect everything (STOP THE WORLD)
```

### Java Heap Layout
```
New objects → Eden → Minor GC → Survivor → (promoted after N GCs) → Old

Eden: ~80% of young gen. Cleared frequently.
Survivor: Objects that survived 1+ minor GC.
Old: Objects that survived ~15 minor GCs.
```

### Write Barrier (Card Table)
```
Problem: Old object references Young object. 
         When collecting Young, how to find these references?
         Can't scan entire Old gen every time.

Solution: "Card Table" — bitmap marking dirty old-gen regions.
         When old object is modified, mark its card as dirty.
         Minor GC only scans dirty cards, not entire old gen.
```

### Real Systems
Java (G1, ZGC, Shenandoah), .NET (CLR), V8 (JavaScript), Go (variant)

---

## 5. Stop-the-World GC

### Definition
GC that pauses ALL application threads while it runs. The application is completely frozen during collection.

### Why It's Needed
```
Mutation during marking:
  Thread marks object A (reachable)
  App thread moves reference from A to B
  GC finishes: B is NOT marked → B is freed
  But B was still in use! → USE-AFTER-FREE BUG

Stop-the-world prevents this by freezing all mutators.
```

### Impact
```
  App Thread:  ────[running]────[PAUSED]────[running]────
  GC Thread:   ──────────────[GC work]──────────────────

  Pause duration:
    Young gen GC:  1-10ms (acceptable)
    Full GC:       100ms-10s (catastrophic for latency-sensitive apps)
```

### Production Impact
```
Java Full GC: 5-second pause
  → Kafka consumer misses heartbeat
  → Broker thinks consumer is dead
  → Triggers partition rebalance
  → 30 seconds of no processing
  All from a 5-second GC pause!
```

### Mitigation
- Tune heap size and GC parameters
- Use concurrent GC (G1, ZGC, Shenandoah)
- Reduce allocation rate
- Use off-heap memory (Netty ByteBuf)

---

## 6. Incremental GC

### Definition
GC that breaks work into small increments interleaved with application execution, reducing individual pause times.

### How It Works
```
Stop-the-World:
  App:  ═══════╗          ╔═══════
  GC:          ╚══════════╝
               ^──200ms──^

Incremental:
  App:  ═══╗ ═╗ ═╗ ═╗ ═╗ ═══════
  GC:      ╚═╝ ╚═╝ ╚═╝ ╚═╝
           5ms 5ms 5ms 5ms
  Same total work, but max pause = 5ms
```

### Tri-Color Marking
```
Colors:
  WHITE: Not yet visited (potentially garbage)
  GRAY:  Visited but children not fully scanned
  BLACK: Visited and all children scanned

Start: all WHITE
Process: Move roots to GRAY
         Take GRAY object → scan children (make them GRAY) → make it BLACK
         When no GRAY left → all WHITE are garbage

Can pause between any GRAY processing step!
```

### Write Barrier for Incremental GC
```
Problem: After GC marks A (BLACK), app creates A → C (WHITE).
         C is reachable but WHITE → would be freed!

Solution: Write barrier catches the mutation:
         When BLACK object gets new reference to WHITE:
         Re-mark the WHITE object as GRAY (re-scan it later)
```

---

## 7. Concurrent GC

### Definition
GC that runs simultaneously with application threads, minimizing or eliminating stop-the-world pauses.

### Java's Modern Concurrent GCs

```
┌─── G1 (Garbage First) ────────────────────────────┐
│ Default since Java 9                               │
│ Heap divided into regions (not fixed young/old)    │
│ Collects regions with most garbage first           │
│ Target pause time: 200ms (configurable)            │
│ Short STW pauses for initial mark + final remark   │
└────────────────────────────────────────────────────┘

┌─── ZGC ───────────────────────────────────────────┐
│ Sub-millisecond pauses (< 1ms)                    │
│ Concurrent relocation using colored pointers       │
│ Scales to multi-TB heaps                           │
│ No generational split (until JDK 21)               │
│ Pause time independent of heap size                │
└────────────────────────────────────────────────────┘

┌─── Shenandoah ────────────────────────────────────┐
│ Sub-millisecond pauses                             │
│ Concurrent compaction using Brooks pointers        │
│ Similar goals to ZGC, different implementation     │
│ Available in OpenJDK                               │
└────────────────────────────────────────────────────┘
```

### Comparison

| | G1 | ZGC | Shenandoah |
|-|---|-----|------------|
| Max pause | 200ms target | < 1ms | < 10ms |
| Heap size | Up to ~32GB practical | Multi-TB | Multi-TB |
| Throughput cost | ~5-10% | ~5-15% | ~5-15% |
| Availability | All JDKs | JDK 15+ | OpenJDK |

### GC Pauses Summary Table
```
┌──────────────────┬─────────────────┬─────────────────┐
│ GC Type          │ Typical Pause    │ Best For         │
├──────────────────┼─────────────────┼─────────────────┤
│ Serial           │ 100ms - 10s     │ Small heaps      │
│ Parallel         │ 100ms - 10s     │ Batch processing │
│ G1               │ 50ms - 200ms    │ General purpose  │
│ ZGC              │ < 1ms           │ Low-latency      │
│ Shenandoah       │ < 10ms          │ Low-latency      │
│ No GC (Rust/C++) │ 0ms             │ Real-time        │
└──────────────────┴─────────────────┴─────────────────┘
```

### Interview Tip
"For a latency-sensitive service like a payment gateway, I'd use ZGC (sub-millisecond pauses). For batch processing, Parallel GC maximizes throughput. For general microservices, G1 is the default choice."

### Summary
Concurrent GCs (G1, ZGC, Shenandoah) perform most work alongside application threads. ZGC achieves sub-millisecond pauses even on TB-sized heaps. The tradeoff is ~5-15% throughput reduction vs stop-the-world collectors.
