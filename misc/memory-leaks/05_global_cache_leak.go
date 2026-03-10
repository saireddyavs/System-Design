package main

import (
	"fmt"
	"runtime"
	"sync"
)

// =============================================================================
// GLOBAL / CACHE LEAK — unbounded caches that grow forever
//
// Global maps or caches that accumulate entries without eviction are a classic
// memory leak. The GC cannot free values reachable from global variables.
// =============================================================================

var globalCache sync.Map // simulates a global in-memory cache

// --- Leak: entries added to global cache, never evicted ---
func leakyGlobalCache() {
	for i := 0; i < 100_000; i++ {
		key := fmt.Sprintf("session-%d", i)
		value := make([]byte, 1024) // 1 KB per entry = ~100 MB total
		globalCache.Store(key, value)
	}
}

// --- Fix: use an LRU cache or TTL-based eviction ---
// In production, use a library like github.com/hashicorp/golang-lru.
// Below is a simplified bounded cache.
type BoundedCache struct {
	mu      sync.Mutex
	data    map[string][]byte
	order   []string
	maxSize int
}

func NewBoundedCache(maxSize int) *BoundedCache {
	return &BoundedCache{
		data:    make(map[string][]byte),
		order:   make([]string, 0, maxSize),
		maxSize: maxSize,
	}
}

func (c *BoundedCache) Set(key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.data[key]; !exists {
		if len(c.order) >= c.maxSize {
			oldest := c.order[0]
			c.order = c.order[1:]
			delete(c.data, oldest)
		}
		c.order = append(c.order, key)
	}
	c.data[key] = value
}

func DemoGlobalCacheLeak() {
	fmt.Println("\n=== Global Cache Leak Demo ===")

	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	fmt.Printf("Before: HeapAlloc = %d MB\n", m.HeapAlloc/1024/1024)

	leakyGlobalCache()

	runtime.GC()
	runtime.ReadMemStats(&m)
	fmt.Printf("After 100k cache entries (never evicted): HeapAlloc = %d MB\n",
		m.HeapAlloc/1024/1024)
}
