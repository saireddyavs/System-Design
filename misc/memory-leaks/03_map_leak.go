package main

import (
	"fmt"
	"runtime"
)

// =============================================================================
// MAP LEAK — maps never shrink in Go
//
// Go's map implementation never releases its internal bucket storage, even
// after deleting all keys. A map that once held 1 million entries still
// occupies the memory for 1 million buckets after all entries are deleted.
// The only way to reclaim that memory is to create a new map.
// =============================================================================

// --- Leak: adding and removing entries doesn't free map memory ---
func leakyMap() {
	m := make(map[int][128]byte)

	for i := 0; i < 500_000; i++ {
		m[i] = [128]byte{}
	}

	var stats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&stats)
	fmt.Printf("  Map with 500k entries: HeapAlloc = %d MB\n", stats.HeapAlloc/1024/1024)

	for i := 0; i < 500_000; i++ {
		delete(m, i)
	}

	runtime.GC()
	runtime.ReadMemStats(&stats)
	fmt.Printf("  Map after deleting all entries: HeapAlloc = %d MB (buckets NOT freed)\n",
		stats.HeapAlloc/1024/1024)

	_ = m // map is still referenced
}

// --- Fix: replace the map with a fresh one ---
func fixedMap() {
	m := make(map[int][128]byte)

	for i := 0; i < 500_000; i++ {
		m[i] = [128]byte{}
	}

	// Instead of deleting keys, replace the entire map.
	m = make(map[int][128]byte)

	runtime.GC()
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	fmt.Printf("  After replacing map: HeapAlloc = %d MB (memory reclaimed)\n",
		stats.HeapAlloc/1024/1024)

	_ = m
}

func DemoMapLeak() {
	fmt.Println("\n=== Map Leak Demo ===")
	leakyMap()
	fixedMap()
}
