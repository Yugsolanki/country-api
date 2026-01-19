package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestInMemoryCache_ConcurrentAccess(t *testing.T) {
	cache := NewInMemoryCache(1 * time.Minute)
	defer cache.Stop()

	var wg sync.WaitGroup
	numGoroutines := 100
	numOperations := 100

	// concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				cache.Set(key, j)
			}
		}(i)
	}

	// concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				cache.Get(key)
			}
		}(i)
	}

	wg.Wait()
}

func TestInMemoryCache_ConcurrentReadWrite(t *testing.T) {
	cache := NewInMemoryCache(1 * time.Minute)
	defer cache.Stop()

	var wg sync.WaitGroup
	numGoroutines := 100
	numOperations := 100
	key := "shared-key"

	// start multiple writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				cache.Set(key, val)
			}
		}(i)
	}

	// start multiple readers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				cache.Get(key)
			}
		}()
	}

	wg.Wait()
}

func TestInMemoryCache_ConcurrentWithExpiration(t *testing.T) {
	cache := NewInMemoryCache(10 * time.Millisecond)
	defer cache.Stop()

	var wg sync.WaitGroup
	numOperations := 100

	// continuously set values
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numOperations; i++ {
			cache.Set("key", i)
			time.Sleep(1 * time.Millisecond)
		}
	}()

	// continuously get values
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numOperations; i++ {
			cache.Get("key")
			time.Sleep(1 * time.Millisecond)
		}
	}()

	wg.Wait()
}
