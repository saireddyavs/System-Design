# Finding Memory Leaks in Go with pprof

## Table of Contents

1. [Quick Start](#quick-start)
2. [What is pprof?](#what-is-pprof)
3. [Enabling pprof in Your App](#enabling-pprof-in-your-app)
4. [Profile Types](#profile-types)
5. [CLI Commands](#cli-commands)
6. [Web UI](#web-ui)
7. [Hunting Each Leak Type](#hunting-each-leak-type)
8. [Comparing Profiles (Diffing)](#comparing-profiles-diffing)
9. [Continuous Profiling in Production](#continuous-profiling-in-production)
10. [Cheat Sheet](#cheat-sheet)

---

## Quick Start

```bash
# Run the demo program
cd memory-leaks
go run .

# In another terminal — interactive heap analysis
go tool pprof http://localhost:6060/debug/pprof/heap

# Or open the visual web UI directly
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap
```

---

## What is pprof?

`pprof` is Go's built-in profiling tool. It can capture:

| Profile         | What it measures                           |
|----------------|--------------------------------------------|
| **heap**       | Current live objects on the heap            |
| **allocs**     | All allocations since program start         |
| **goroutine**  | Stack traces of all current goroutines      |
| **profile**    | CPU usage (sampled)                         |
| **block**      | Where goroutines block on synchronization   |
| **mutex**      | Mutex contention                            |
| **trace**      | Execution trace (for `go tool trace`)       |

---

## Enabling pprof in Your App

### Option 1: HTTP server (recommended for services)

```go
import (
    "net/http"
    _ "net/http/pprof"  // registers /debug/pprof/* handlers
)

func main() {
    go http.ListenAndServe("localhost:6060", nil)
    // ... your application code ...
}
```

### Option 2: Write profiles to files (for CLI tools / batch jobs)

```go
import (
    "os"
    "runtime"
    "runtime/pprof"
)

func main() {
    // Heap profile
    f, _ := os.Create("heap.prof")
    defer f.Close()
    runtime.GC()
    pprof.WriteHeapProfile(f)

    // CPU profile
    cf, _ := os.Create("cpu.prof")
    defer cf.Close()
    pprof.StartCPUProfile(cf)
    defer pprof.StopCPUProfile()
}
```

### Important: enable block/mutex profiling explicitly

```go
runtime.SetBlockProfileRate(1)     // 1 = capture every block event
runtime.SetMutexProfileFraction(1) // 1 = capture every mutex event
```

---

## Profile Types

### Heap Profile (`/debug/pprof/heap`)

Shows memory currently allocated and still in use.

Two key views controlled by `-sample_index`:
- `inuse_space` (default) — bytes currently allocated (live objects)
- `inuse_objects` — count of live objects
- `alloc_space` — total bytes allocated since start (includes freed)
- `alloc_objects` — total objects allocated since start

```bash
# Live memory (default — best for finding leaks)
go tool pprof http://localhost:6060/debug/pprof/heap

# All allocations since start (best for reducing GC pressure)
go tool pprof -sample_index=alloc_space http://localhost:6060/debug/pprof/heap
```

### Goroutine Profile (`/debug/pprof/goroutine`)

Shows stack traces of all goroutines. Essential for finding goroutine leaks.

```bash
# Text dump — look for large groups of goroutines with the same stack
curl http://localhost:6060/debug/pprof/goroutine?debug=1

# Full stack dump with labels
curl http://localhost:6060/debug/pprof/goroutine?debug=2

# Interactive
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

### CPU Profile (`/debug/pprof/profile`)

```bash
# 30-second CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

---

## CLI Commands

Once inside `go tool pprof`, these are the most useful commands:

```
(pprof) top            # top functions by memory/CPU
(pprof) top -cum       # top functions by cumulative memory (including callees)
(pprof) list funcName  # source-level annotation of a function
(pprof) tree           # call tree
(pprof) web            # open a graph in the browser (needs graphviz)
(pprof) png            # save graph as PNG
(pprof) traces         # show all stack traces
```

### Filtering

```
(pprof) top -nodecount=20                         # show top 20
(pprof) top -focus="mypackage"                    # only functions matching pattern
(pprof) top -ignore="runtime|testing"             # exclude stdlib noise
(pprof) list leakyGoroutineBlockedReceive         # source annotation for specific func
```

---

## Web UI

The web UI is the easiest way to explore profiles visually:

```bash
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap
```

This opens a browser with:

- **Graph** — call graph with box sizes proportional to memory
- **Flame Graph** — interactive flame graph (click to zoom)
- **Top** — sortable table of functions
- **Source** — line-by-line allocation counts
- **Peek** — shows callers and callees of a function

### Reading the Graph

- **Box size** = memory attributed to that function
- **Edge thickness** = memory flowing through that call path
- **Red/large boxes** = likely leak sources
- Flat vs. Cumulative:
  - **Flat**: memory allocated directly by the function
  - **Cum**: memory allocated by the function + everything it calls

---

## Hunting Each Leak Type

### 1. Goroutine Leaks

```bash
# Check goroutine count over time
curl -s http://localhost:6060/debug/pprof/goroutine?debug=1 | head -1
# Output: "goroutine profile: total 100234"  ← way too many!

# Get grouped stack traces
go tool pprof http://localhost:6060/debug/pprof/goroutine
(pprof) top
(pprof) traces
```

**What to look for**: Thousands of goroutines with the same stack trace, typically
blocked on channel receive (`chan receive`), `select`, or `sync.Mutex.Lock`.

### 2. Heap / Slice / Map Leaks

```bash
go tool pprof http://localhost:6060/debug/pprof/heap
(pprof) top
```

**What to look for**: Functions with large `flat` or `cum` values. Follow the
call graph upstream to find what is holding the reference.

```
(pprof) top -cum
(pprof) list leakySlice    # shows exact line allocating memory
```

### 3. Global Cache Leaks

Same as heap, but the top allocators will be cache `Set` / `Store` functions.

```
(pprof) list leakyGlobalCache
```

### 4. Ticker/Timer Leaks

These show up in the goroutine profile as stuck `time.Ticker` or `time.Timer` goroutines:

```bash
go tool pprof http://localhost:6060/debug/pprof/goroutine
(pprof) top
# Look for time.(*Ticker).Stop or runtime.timerproc stacks
```

### 5. HTTP Response Body / FD Leaks

Not directly visible in heap profiles. Look for:
- Growing number of goroutines in `net/http` stack frames
- Increasing file descriptors: `ls /proc/<pid>/fd | wc -l` (Linux)

---

## Comparing Profiles (Diffing)

The most powerful leak-finding technique: take two heap profiles at different
times and diff them. The diff shows what grew.

```bash
# Capture baseline
curl -o base.prof http://localhost:6060/debug/pprof/heap

# Wait, let the leak accumulate...
sleep 60

# Capture current
curl -o current.prof http://localhost:6060/debug/pprof/heap

# Diff: shows only what GREW between snapshots
go tool pprof -base=base.prof current.prof
(pprof) top          # functions that allocated more since baseline
(pprof) list funcName
```

### Visual diff

```bash
go tool pprof -http=:8080 -diff_base=base.prof current.prof
```

Red = grew, Green = shrank. Look for the big red nodes.

---

## Continuous Profiling in Production

### Programmatic snapshots

```go
import "runtime/pprof"

func captureProfilePeriodically() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for t := range ticker.C {
        filename := fmt.Sprintf("heap_%s.prof", t.Format("20060102_150405"))
        f, err := os.Create(filename)
        if err != nil {
            continue
        }
        runtime.GC()
        pprof.WriteHeapProfile(f)
        f.Close()
    }
}
```

### Using `runtime.MemStats` for monitoring

```go
var m runtime.MemStats
runtime.ReadMemStats(&m)

fmt.Printf("HeapAlloc:    %d MB\n", m.HeapAlloc/1024/1024)    // current live heap
fmt.Printf("HeapSys:      %d MB\n", m.HeapSys/1024/1024)      // heap memory obtained from OS
fmt.Printf("HeapObjects:  %d\n", m.HeapObjects)                // live object count
fmt.Printf("NumGoroutine: %d\n", runtime.NumGoroutine())       // goroutine count
fmt.Printf("NumGC:        %d\n", m.NumGC)                      // GC cycle count
```

### Production tools

- **Pyroscope** — continuous profiling platform
- **Grafana Phlare** — integrated with Grafana
- **Google Cloud Profiler** — agent-based, low overhead
- **Datadog Continuous Profiler** — commercial APM with profiling

---

## Cheat Sheet

```bash
# ─── CAPTURE ──────────────────────────────────────────────
# Heap profile (live objects)
go tool pprof http://localhost:6060/debug/pprof/heap

# Allocation profile (all allocs since start)
go tool pprof -sample_index=alloc_space http://localhost:6060/debug/pprof/heap

# Goroutine profile
go tool pprof http://localhost:6060/debug/pprof/goroutine

# CPU profile (30 seconds)
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Block profile
go tool pprof http://localhost:6060/debug/pprof/block

# Mutex profile
go tool pprof http://localhost:6060/debug/pprof/mutex

# ─── ANALYZE (inside pprof) ──────────────────────────────
top                    # top allocators
top -cum               # top by cumulative
list <func>            # source annotation
tree                   # call tree
web                    # graph in browser
traces                 # full stack traces

# ─── WEB UI ──────────────────────────────────────────────
go tool pprof -http=:8080 heap.prof
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap

# ─── DIFF ────────────────────────────────────────────────
go tool pprof -base=old.prof new.prof
go tool pprof -http=:8080 -diff_base=old.prof new.prof

# ─── SAVE TO FILE ────────────────────────────────────────
curl -o heap.prof http://localhost:6060/debug/pprof/heap
curl -o goroutine.prof http://localhost:6060/debug/pprof/goroutine
curl -o cpu.prof http://localhost:6060/debug/pprof/profile?seconds=30
```

---

## Summary of Leak Types and Detection

| Leak Type             | pprof Profile     | What to Look For                                   | Fix                                      |
|-----------------------|-------------------|-----------------------------------------------------|------------------------------------------|
| Goroutine leak        | `goroutine`       | Thousands of goroutines with same stack             | Use context/cancellation, timeouts        |
| Slice/subslice pin    | `heap`            | Large `inuse_space` from subslice creation          | `copy()` data to new slice               |
| Map never shrinks     | `heap`            | Map holds memory after all keys deleted             | Replace with `make(map[...]...)`         |
| Channel buffer leak   | `heap`+`goroutine`| Buffered channels with unconsumed data              | Drain channels, use bounded buffers       |
| Global cache leak     | `heap`            | Growing `inuse_space` in cache Store/Set funcs      | LRU/TTL eviction                         |
| Ticker/Timer leak     | `goroutine`       | Many goroutines in `time.(*Ticker)`                 | Always `defer ticker.Stop()`             |
| HTTP body leak        | `goroutine`       | Goroutines stuck in `net/http` read                 | `io.Copy(io.Discard, resp.Body)` + Close |
| String substring pin  | `heap`            | Large strings retained by small substrings          | `strings.Clone()` or `string([]byte())` |
