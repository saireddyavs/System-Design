package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// =============================================================================
// HTTP RESPONSE BODY LEAK
//
// http.Get / http.Client.Do return a response with a Body that MUST be fully
// read and closed. If you skip either step:
//   - Not closing: the underlying TCP connection is never returned to the pool,
//     and the file descriptor leaks.
//   - Not draining: the connection can't be reused (HTTP keep-alive won't work).
// =============================================================================

// --- Leak: response body not closed ---
func leakyHTTPRequest() {
	resp, err := http.Get("https://example.com")
	if err != nil {
		fmt.Println("request failed:", err)
		return
	}

	// BUG: body is never closed. The underlying TCP connection leaks.
	_ = resp
}

// --- Leak: body closed but not drained ---
func partiallyLeakyHTTPRequest() {
	resp, err := http.Get("https://example.com")
	if err != nil {
		fmt.Println("request failed:", err)
		return
	}
	defer resp.Body.Close() // closed, but not drained — connection won't be reused
}

// --- Fix: drain and close the body ---
func fixedHTTPRequest() {
	resp, err := http.Get("https://example.com")
	if err != nil {
		fmt.Println("request failed:", err)
		return
	}
	defer func() {
		io.Copy(io.Discard, resp.Body) // drain remaining bytes
		resp.Body.Close()
	}()

	// Process the response...
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Response length:", len(body))
}

// DemoHTTPResponseLeak shows the pattern without making real HTTP calls.
func DemoHTTPResponseLeak() {
	fmt.Println("\n=== HTTP Response Body Leak Demo ===")

	// Simulated demonstration — real leak requires a live server.
	// The patterns above show the three variants:
	//   leakyHTTPRequest()          — body never closed (FD leak)
	//   partiallyLeakyHTTPRequest() — body closed but not drained (connection not reused)
	//   fixedHTTPRequest()          — body drained AND closed (correct)

	fmt.Println("  Pattern 1 (leaky):    resp.Body never closed → FD leak")
	fmt.Println("  Pattern 2 (partial):  resp.Body.Close() without drain → connection not reused")
	fmt.Println("  Pattern 3 (fixed):    io.Copy(io.Discard, resp.Body) + Close() → correct")

	_ = strings.NewReader // suppress unused import
}
