package main

import (
	"fmt"
	"runtime"
	"time"
)

// =============================================================================
// TICKER / TIMER LEAK
//
// time.NewTicker and time.NewTimer allocate resources in the Go runtime.
// If you don't call Stop(), the ticker/timer goroutine lives forever.
// time.After() inside a loop creates a new timer on EVERY iteration —
// these pile up if the loop is hot.
// =============================================================================

// --- Leak: time.After in a hot loop ---
func leakyTimeAfterInLoop(quit chan struct{}) {
	ch := make(chan int, 1)

	go func() {
		for i := 0; ; i++ {
			ch <- i
			time.Sleep(1 * time.Millisecond)
		}
	}()

	for {
		select {
		case <-ch:
			// Each iteration creates a new timer that won't be GC'd until it fires.
			// In a hot loop, thousands of pending timers accumulate.
		case <-time.After(1 * time.Minute):
			return
		case <-quit:
			return
		}
	}
}

// --- Fix: reuse a single timer ---
func fixedTimerReuse(quit chan struct{}) {
	ch := make(chan int, 1)

	go func() {
		for i := 0; ; i++ {
			ch <- i
			time.Sleep(1 * time.Millisecond)
		}
	}()

	timer := time.NewTimer(1 * time.Minute)
	defer timer.Stop()

	for {
		select {
		case <-ch:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(1 * time.Minute)
		case <-timer.C:
			return
		case <-quit:
			return
		}
	}
}

// --- Leak: ticker never stopped ---
func leakyTicker() *time.Ticker {
	ticker := time.NewTicker(100 * time.Millisecond)
	go func() {
		for range ticker.C {
			// do work
		}
	}()
	// forgot ticker.Stop() — the ticker goroutine runs forever
	return ticker
}

// --- Fix: always defer Stop ---
func fixedTicker(done chan struct{}) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// do work
		case <-done:
			return
		}
	}
}

func DemoTickerLeak() {
	fmt.Println("\n=== Ticker/Timer Leak Demo ===")

	before := runtime.NumGoroutine()

	for i := 0; i < 100; i++ {
		leakyTicker()
	}

	time.Sleep(200 * time.Millisecond)
	fmt.Printf("Goroutines before: %d, after creating 100 leaked tickers: %d\n",
		before, runtime.NumGoroutine())
}
