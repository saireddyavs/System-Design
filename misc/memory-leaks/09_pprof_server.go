package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // side-effect import: registers /debug/pprof/* handlers
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

// =============================================================================
// PPROF SERVER — exposes profiling endpoints for live analysis
//
// Importing "net/http/pprof" registers handlers on DefaultServeMux:
//   /debug/pprof/              — index page
//   /debug/pprof/heap          — heap profile
//   /debug/pprof/goroutine     — goroutine dump
//   /debug/pprof/allocs        — allocation profile
//   /debug/pprof/profile       — CPU profile (takes ?seconds=N)
//   /debug/pprof/trace         — execution trace
//   /debug/pprof/block         — blocking profile
//   /debug/pprof/mutex         — mutex contention profile
// =============================================================================

func StartPprofServer() {
	// Enable block and mutex profiling (disabled by default).
	runtime.SetBlockProfileRate(1)
	runtime.SetMutexProfileFraction(1)

	go func() {
		fmt.Println("\npprof server listening on http://localhost:6060/debug/pprof/")
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}

// WriteCPUProfile captures a CPU profile to a file.
func WriteCPUProfile(filename string, duration time.Duration) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal("could not create CPU profile:", err)
	}
	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile:", err)
	}

	time.Sleep(duration)
	pprof.StopCPUProfile()
	fmt.Printf("CPU profile written to %s\n", filename)
}

// WriteHeapProfile captures a heap profile snapshot to a file.
func WriteHeapProfile(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal("could not create heap profile:", err)
	}
	defer f.Close()

	runtime.GC()
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("could not write heap profile:", err)
	}
	fmt.Printf("Heap profile written to %s\n", filename)
}
