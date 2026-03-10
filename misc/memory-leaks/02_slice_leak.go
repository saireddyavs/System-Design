package main

import (
	"fmt"
	"runtime"
)

// =============================================================================
// SLICE / SUBSLICE LEAK
//
// When you subslice a large slice, the subslice still references the original
// backing array. The GC cannot free the original array as long as ANY subslice
// referencing it is alive — even if you only keep 10 bytes of a 100 MB array.
// =============================================================================

// --- Leak: keeping a small subslice pins the entire backing array ---
func leakySlice() []byte {
	bigSlice := make([]byte, 100*1024*1024) // 100 MB
	for i := range bigSlice {
		bigSlice[i] = byte(i % 256)
	}

	// Only need the first 10 bytes, but the entire 100 MB stays in memory
	// because smallSlice shares the same backing array.
	smallSlice := bigSlice[:10]
	return smallSlice
}

// --- Fix: copy the data you need into a new, independent slice ---
func fixedSlice() []byte {
	bigSlice := make([]byte, 100*1024*1024) // 100 MB
	for i := range bigSlice {
		bigSlice[i] = byte(i % 256)
	}

	smallSlice := make([]byte, 10)
	copy(smallSlice, bigSlice[:10])
	// bigSlice is now eligible for GC because nothing references its backing array.
	return smallSlice
}

func DemoSliceLeak() {
	fmt.Println("\n=== Slice Leak Demo ===")

	var m runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&m)
	fmt.Printf("Before: HeapAlloc = %d MB\n", m.HeapAlloc/1024/1024)

	leaked := leakySlice()
	runtime.GC()
	runtime.ReadMemStats(&m)
	fmt.Printf("After leaky subslice (kept %d bytes, pinning entire array): HeapAlloc = %d MB\n",
		len(leaked), m.HeapAlloc/1024/1024)

	_ = leaked

	fixed := fixedSlice()
	runtime.GC()
	runtime.ReadMemStats(&m)
	fmt.Printf("After fixed copy (kept %d bytes): HeapAlloc = %d MB\n",
		len(fixed), m.HeapAlloc/1024/1024)

	_ = fixed
}
