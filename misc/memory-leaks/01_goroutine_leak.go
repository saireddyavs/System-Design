package main

import (
	"fmt"
	"runtime"
	"time"
)

// =============================================================================
// GOROUTINE LEAK — the most common memory leak in Go
//
// Goroutines that are blocked forever (on channel ops, mutexes, I/O, etc.)
// are never garbage-collected. Each goroutine costs ~2-8 KB of stack that
// grows as needed, so thousands of leaked goroutines can consume significant
// memory.
// =============================================================================

// --- Leak: blocked channel receive with no sender ---
func leakyGoroutineBlockedReceive() {
	for i := 0; i < 100_000; i++ {
		ch := make(chan int)
		go func() {
			// This goroutine blocks forever because nobody sends on ch.
			val := <-ch
			_ = val
		}()
		// ch goes out of scope, but the goroutine is still alive waiting on it.
	}
}

// --- Leak: forgotten context / missing cancellation ---
func leakyGoroutineNoCancel() {
	for i := 0; i < 100_000; i++ {
		done := make(chan struct{})
		go func() {
			// Simulates a worker that should stop when done is closed,
			// but the caller never closes it.
			<-done
			fmt.Println("worker finished") // never reached
		}()
		// forgot: close(done) or use context.WithCancel
	}
}

// --- Fix: use a buffered channel or context ---
func fixedGoroutineWithContext() {
	for i := 0; i < 100_000; i++ {
		done := make(chan struct{})
		go func() {
			select {
			case <-done:
				return
			case <-time.After(1 * time.Second):
				return
			}
		}()
		close(done) // signal goroutine to exit
	}
}

func DemoGoroutineLeak() {
	fmt.Println("=== Goroutine Leak Demo ===")
	fmt.Printf("Before: %d goroutines\n", runtime.NumGoroutine())

	leakyGoroutineBlockedReceive()
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	fmt.Printf("After leak: %d goroutines (these are leaked!)\n", runtime.NumGoroutine())
}
