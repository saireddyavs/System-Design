package main

import (
	"fmt"
	"runtime"
	"time"
)

// =============================================================================
// CHANNEL LEAK — unbounded producers / forgotten consumers
//
// Channels themselves hold references to the data sent through them.
// A buffered channel that is never drained keeps all buffered values alive.
// An unbuffered channel with a blocked sender keeps the sender goroutine
// (and its stack) alive.
// =============================================================================

// --- Leak: buffered channel filled but never consumed ---
func leakyBufferedChannel() {
	ch := make(chan []byte, 1000)

	go func() {
		for i := 0; i < 1000; i++ {
			data := make([]byte, 1024*1024) // 1 MB per message
			ch <- data
		}
	}()

	// Consumer reads a few then stops — remaining 1 MB messages stay buffered.
	for i := 0; i < 5; i++ {
		<-ch
	}

	// ch still holds ~995 MB of data. If ch is never drained and never GC'd,
	// this memory is leaked.
	_ = ch
}

// --- Leak: sender blocked forever on full channel ---
func leakySenderBlocked() {
	ch := make(chan int) // unbuffered

	for i := 0; i < 1000; i++ {
		go func(val int) {
			ch <- val // blocks forever — no receiver
		}(i)
	}
	// 1000 goroutines permanently blocked
}

// --- Fix: use select with context or timeout ---
func fixedChannelWithTimeout() {
	ch := make(chan int)

	for i := 0; i < 1000; i++ {
		go func(val int) {
			select {
			case ch <- val:
			case <-time.After(1 * time.Second):
				return // give up after timeout
			}
		}(i)
	}

	time.Sleep(2 * time.Second) // let timeouts expire
}

func DemoChannelLeak() {
	fmt.Println("\n=== Channel Leak Demo ===")
	before := runtime.NumGoroutine()

	leakySenderBlocked()
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	fmt.Printf("Goroutines before: %d, after: %d (blocked senders leaked)\n",
		before, runtime.NumGoroutine())
}
