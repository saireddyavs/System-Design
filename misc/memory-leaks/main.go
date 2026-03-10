package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║          Go Memory Leak Patterns & pprof Profiling          ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")

	StartPprofServer()

	DemoGoroutineLeak()
	DemoSliceLeak()
	DemoMapLeak()
	DemoChannelLeak()
	DemoGlobalCacheLeak()
	DemoTickerLeak()
	DemoHTTPResponseLeak()
	DemoStringLeak()

	WriteHeapProfile("heap.prof")

	fmt.Println("\n════════════════════════════════════════════════════════════════")
	fmt.Println("All demos complete. pprof server still running on :6060.")
	fmt.Println("Open http://localhost:6060/debug/pprof/ in your browser.")
	fmt.Println("")
	fmt.Println("Try these commands in another terminal:")
	fmt.Println("  go tool pprof http://localhost:6060/debug/pprof/heap")
	fmt.Println("  go tool pprof http://localhost:6060/debug/pprof/goroutine")
	fmt.Println("  go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap")
	fmt.Println("════════════════════════════════════════════════════════════════")

	if len(os.Args) > 1 && os.Args[1] == "--exit" {
		return
	}

	// Keep the process alive so pprof stays accessible.
	fmt.Println("\nPress Ctrl+C to exit...")
	for {
		time.Sleep(time.Hour)
	}
}
