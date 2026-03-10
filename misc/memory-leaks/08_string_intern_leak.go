package main

import (
	"fmt"
	"runtime"
	"strings"
)

// =============================================================================
// STRING INTERNING / SUBSTRING LEAK
//
// Prior to Go 1.18, substrings shared the parent string's backing array
// (similar to subslice pinning). In modern Go (1.18+) the compiler is
// smarter, but you can still trigger this via unsafe conversions or when
// holding references to large strings.
//
// The general pattern: reading a huge string (e.g., file contents) and
// keeping a tiny substring alive pins the entire original string.
// =============================================================================

// --- Leak: substring of a large string stays referenced ---
func leakySubstring() string {
	bigString := strings.Repeat("A", 100*1024*1024) // 100 MB

	// Extracting a small piece — depending on compiler escape analysis,
	// the large string may still be retained.
	small := bigString[:10]
	return small
}

// --- Fix: force a copy via string conversion ---
func fixedSubstring() string {
	bigString := strings.Repeat("A", 100*1024*1024) // 100 MB

	small := string([]byte(bigString[:10]))
	// Or: small := strings.Clone(bigString[:10])  // Go 1.20+
	return small
}

func DemoStringLeak() {
	fmt.Println("\n=== String/Substring Leak Demo ===")

	var m runtime.MemStats

	_ = leakySubstring()
	runtime.GC()
	runtime.ReadMemStats(&m)
	fmt.Printf("After leaky substring: HeapAlloc = %d MB\n", m.HeapAlloc/1024/1024)

	_ = fixedSubstring()
	runtime.GC()
	runtime.ReadMemStats(&m)
	fmt.Printf("After fixed substring:  HeapAlloc = %d MB\n", m.HeapAlloc/1024/1024)
}
