# Module 12: OS Primitives

---

## 1. Virtual Memory

### Definition
An abstraction that gives each process the illusion of having its own large, contiguous address space, while the OS maps it to physical RAM (and disk) transparently.

### How It Works
```
Process A sees:     Physical RAM:         Disk (Swap):
┌────────────┐     ┌────────────┐        ┌────────────┐
│ 0x0000     │ ──→ │ Frame 42   │        │            │
│ 0x1000     │ ──→ │ Frame 7    │        │            │
│ 0x2000     │ ──→ │            │ ──→    │ Page on    │
│ 0x3000     │ ──→ │ Frame 100  │        │ disk       │
└────────────┘     └────────────┘        └────────────┘
  Virtual           Physical              Swap
  Pages             Frames

Page Table: virtual page → physical frame (or disk)
```

### Key Concepts
- **Page**: Fixed-size block (typically 4KB)
- **Page table**: Maps virtual → physical addresses
- **TLB**: CPU cache for page table (fast lookup)
- **Page fault**: Access to page not in RAM → load from disk

### Why It Matters for System Design
```
1. Process isolation: A can't access B's memory
2. Memory overcommit: Promise more RAM than available
3. mmap: Map files into address space
4. Copy-on-Write: Fork without copying
5. Swap: Use disk when RAM is full (DON'T for databases!)
```

### Real System Impact
```
Elasticsearch: Disable swap! (mlockall)
  If ES pages are swapped out, queries go from 1ms → 1000ms

Kubernetes: Set memory limits to avoid OOM killer
Redis: fork() for BGSAVE uses CoW (virtual memory magic)
```

---

## 2. Paging

### Definition
Dividing virtual and physical memory into fixed-size pages/frames. The OS manages a page table mapping virtual pages to physical frames.

### Page Table Structure
```
Virtual Address: [Page Number | Offset]
  Page Number → lookup in Page Table → Frame Number
  Physical Address = Frame Number + Offset

Example (4KB pages):
  Virtual: 0x12345678
  Page: 0x12345 (top 20 bits)
  Offset: 0x678 (bottom 12 bits)
  Page Table[0x12345] = Frame 42
  Physical: 42 × 4096 + 0x678
```

### Multi-Level Page Tables
```
Why? A flat page table for 64-bit addresses would be enormous.

4-level page table (x86-64):
  PML4 → PDPT → PD → PT → Physical Frame

Only allocate table entries for address ranges actually used.
Most of virtual address space is unmapped → saves memory.
```

### TLB (Translation Lookaside Buffer)
```
CPU cache for page table lookups.
TLB hit: 1 cycle (~0.5ns)
TLB miss: 10-100 cycles (walk page table in memory)

Context switch flushes TLB → expensive!
This is why fewer threads (event loop) perform better.
```

### Huge Pages
```
Standard page: 4KB  → many page table entries, TLB pressure
Huge page:     2MB  → 512x fewer entries
1GB page:      1GB  → for databases with large working sets

Use case: Databases (Oracle, Postgres), JVM (-XX:+UseLargePages)
```

---

## 3. Segmentation

### Definition
Dividing a process's memory into variable-length segments (code, data, stack, heap) each with its own base address and bounds.

### Segment vs Page

| | Segmentation | Paging |
|-|---|---|
| Unit size | Variable | Fixed (4KB) |
| Fragmentation | External (gaps) | Internal (wasted space in page) |
| Modern use | Mostly replaced by paging | Primary memory management |
| Protection | Per-segment (R/W/X) | Per-page |

### Modern Reality
x86-64 uses paging exclusively for memory management. Segmentation is vestigial (flat segments covering entire address space). Protection bits are on page table entries.

---

## 4. Copy-on-Write (CoW)

### Definition
When a process forks, the OS doesn't copy memory. Both processes share the same physical pages. Only when one WRITES does the OS copy that specific page.

### How It Works
```
Before fork():
  Process A: Page 1 [data] → Frame 100
             Page 2 [data] → Frame 200

After fork():  (no copy!)
  Process A: Page 1 → Frame 100 (read-only)
             Page 2 → Frame 200 (read-only)
  Process B: Page 1 → Frame 100 (shared, read-only)
             Page 2 → Frame 200 (shared, read-only)

Process B writes to Page 1:
  1. Page fault (read-only violation)
  2. OS copies Frame 100 → Frame 300
  3. Process B: Page 1 → Frame 300 (now read-write)
  Process A: Page 1 → Frame 100 (unchanged)
```

### Why It Matters
```
Redis BGSAVE:
  1. fork() → instant (no memory copy)
  2. Child process writes snapshot to disk
  3. Parent continues serving requests
  4. Only modified pages are actually copied

Without CoW: fork() would copy entire Redis dataset (GBs)
With CoW: fork() is O(1), copies only pages parent modifies
```

### Real Systems
Redis (BGSAVE, BGREWRITEAOF), PostgreSQL (WAL archiving), Linux fork(), Docker layers

---

## 5. Context Switching

### Definition
The OS saving the state of the current process/thread and loading the state of another, so the CPU can switch between tasks.

### What Gets Saved/Restored
```
Per context switch:
  1. CPU registers (general purpose, PC, SP)
  2. Stack pointer
  3. Program counter
  4. Floating point state
  5. TLB flush (for process switch, not thread switch)
  6. Page table base register update
```

### Cost
```
Thread switch (same process):  ~1-5 microseconds
  Saves registers, switches stack
  NO TLB flush (same address space)

Process switch:               ~5-30 microseconds
  Same as thread switch PLUS:
  TLB flush → subsequent memory accesses are slow
  Cache pollution → cold cache after switch

Goroutine switch (Go):        ~100 nanoseconds
  User-space only, no kernel involvement
  Tiny stack (2KB vs 1MB for OS thread)
```

### System Design Impact
```
Why event loops (Node.js, Redis) are fast:
  1 thread, 0 context switches for I/O handling
  Thousands of connections, one thread

Why Go scales:
  100,000 goroutines = ~100 OS threads
  User-space scheduling, minimal context switching
```

---

## 6. File Descriptors

### Definition
An integer handle representing an open file, socket, pipe, or device in Unix/Linux. Every I/O operation uses file descriptors.

### Key Limits
```
$ ulimit -n
1024        ← default max open FDs per process (MUST increase for servers)

Production settings:
  Nginx:     65535+ FDs
  Kafka:     100000+ FDs (many open log segment files)
  Database:  Depends on connection count

Set in: /etc/security/limits.conf
  * soft nofile 65535
  * hard nofile 65535
```

### Common Error
```
"Too many open files" — means you hit ulimit

Causes:
  - Connection leak (sockets not closed)
  - File handle leak (files not closed)
  - Too many concurrent connections

Debug: ls -la /proc/<pid>/fd | wc -l
```

---

## 7. False Sharing

### Definition
A performance killer where two threads on different CPU cores modify independent variables that happen to share the same CPU cache line (64 bytes).

### How It Works
```
Cache line (64 bytes):
  [counter_A (8 bytes)] [counter_B (8 bytes)] [padding...]

Core 1: Increments counter_A
Core 2: Increments counter_B

Both are on the SAME cache line.
Core 1 writes → invalidates Core 2's cache line
Core 2 writes → invalidates Core 1's cache line
→ Cache line bounces between cores (100x slower!)
```

### Visual
```
Core 1 Cache: [counter_A=5 | counter_B=3] ← owns line
Core 2: wants to write counter_B
  → Invalidate Core 1's line
  → Transfer line to Core 2
Core 2 Cache: [counter_A=5 | counter_B=4] ← owns line
Core 1: wants to write counter_A
  → Invalidate Core 2's line
  → Transfer AGAIN

This "ping-pong" kills performance!
```

### Solutions
```
1. PADDING: Add padding between variables
   struct { long counter_A; long pad[7]; long counter_B; }
   → Each variable on its own cache line

2. JAVA: @Contended annotation
   @sun.misc.Contended
   volatile long counter;
   → JVM adds 128 bytes of padding

3. ALIGNMENT: Align to cache line boundaries
   __attribute__((aligned(64))) long counter;
```

### Real Systems
LMAX Disruptor (padded sequence counters), Java ConcurrentHashMap, any high-performance concurrent code

---

## 8. NUMA (Non-Uniform Memory Access)

### Definition
In multi-socket servers, each CPU has "local" memory (fast) and "remote" memory (slow, across interconnect). NUMA-aware software uses local memory for performance.

### Architecture
```
┌──────────────────┐    Interconnect    ┌──────────────────┐
│    Socket 0      │ ←───────────────→ │    Socket 1      │
│  ┌────┐ ┌────┐  │    (QPI/UPI)      │  ┌────┐ ┌────┐  │
│  │Core│ │Core│  │                    │  │Core│ │Core│  │
│  └────┘ └────┘  │                    │  └────┘ └────┘  │
│  ┌────────────┐  │                    │  ┌────────────┐  │
│  │ Local RAM  │  │                    │  │ Local RAM  │  │
│  │ ~80ns      │  │                    │  │ ~80ns      │  │
│  └────────────┘  │                    │  └────────────┘  │
└──────────────────┘                    └──────────────────┘

Socket 0 accessing Socket 1's RAM: ~150ns (2x slower!)
```

### Impact on System Design
```
Database on 2-socket server:
  NUMA-unaware: Memory randomly allocated → 50% remote access
  NUMA-aware:   Pin process to socket, use local memory → 2x faster

JVM: -XX:+UseNUMA (allocate young gen on local node)
Linux: numactl --cpubind=0 --membind=0 ./my_server
```

### Real Systems
Databases (Oracle, SQL Server NUMA-aware), JVM, Redis (single-threaded, pin to one socket)

---

## 9. SIMD (Single Instruction, Multiple Data)

### Definition
CPU instructions that operate on multiple data elements simultaneously. One instruction processes 4, 8, or 16 values at once.

### How It Works
```
Scalar (normal):
  for i in 0..4:
    C[i] = A[i] + B[i]    ← 4 instructions

SIMD (vectorized):
  C[0:4] = A[0:4] + B[0:4]  ← 1 instruction!

  256-bit SIMD register (AVX2):
  [A0][A1][A2][A3]  +  [B0][B1][B2][B3]  =  [C0][C1][C2][C3]
     one CPU cycle
```

### Instruction Sets
```
SSE:   128-bit (4 × 32-bit floats)
AVX2:  256-bit (8 × 32-bit floats)
AVX-512: 512-bit (16 × 32-bit floats)  ← Intel Xeon
ARM NEON: 128-bit
```

### Real Systems
- **ClickHouse**: Vectorized query execution (SIMD-optimized)
- **DuckDB**: Vectorized analytics engine
- **Postgres**: SIMD for comparison operations
- **JSON parsers**: simdjson (parses JSON at GB/s)

### Impact
Analytical queries (SUM, AVG, filter) run 4-8x faster with SIMD vectorization.

---

## 10. Direct I/O vs Buffered I/O

### Buffered I/O (Default)
```
App → write() → Page Cache (RAM) → eventually → Disk
App ← read()  ← Page Cache (RAM) ← loads from ← Disk

OS manages the cache. Re-reads are fast (from RAM).
Good for general workloads.
```

### Direct I/O
```
App → write() → Disk  (bypasses page cache)
App ← read()  ← Disk  (bypasses page cache)

O_DIRECT flag in open()
```

### Why Databases Use Direct I/O
```
Problem with buffered I/O for databases:
  1. Double caching: DB has its own buffer pool + OS page cache
     → Wastes RAM (same data cached twice)
  2. OS eviction policy (LRU) is dumber than DB's policy
     → DB knows access patterns better
  3. Unpredictable page cache eviction causes latency spikes

Direct I/O lets the database manage its own cache optimally.
```

### Comparison

| | Buffered I/O | Direct I/O |
|-|---|---|
| Cache | OS page cache | Application manages |
| Read perf | Great (cached reads) | App must cache itself |
| Write perf | Good (async writeback) | Slower (immediate disk) |
| Memory usage | Double buffering possible | Efficient |
| Use case | General applications | Databases |
| Examples | Most apps | MySQL, Postgres, RocksDB |

### Summary
Buffered I/O uses the OS page cache (good for general use). Direct I/O bypasses it (better for databases that manage their own cache). Databases prefer direct I/O to avoid double-caching and unpredictable eviction.
