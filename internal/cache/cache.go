package cache

import (
	"sync"
	"time"
)

type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
}

// Single cache item
type cacheItem struct {
	value      interface{}
	expiration time.Time
}

// InMemoryCache implements the Cache interface with thread-safe
type InMemoryCache struct {
	mu       sync.RWMutex
	items    map[string]cacheItem
	ttl      time.Duration
	stopChan chan struct{}
}

// NewInMemoryCache creates a new in-memory cache with the specified TTL
func NewInMemoryCache(ttl time.Duration) *InMemoryCache {
	cache := &InMemoryCache{
		items:    make(map[string]cacheItem),
		ttl:      ttl,
		stopChan: make(chan struct{}),
	}

	go cache.cleanup()

	return cache
}

// Get retrieves a value from the cache
// Returns the value and true if the item is found otherwise it will return false and nil
func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// check if item has expired
	if time.Now().After(item.expiration) {
		return nil, false
	}

	return item.value, true
}

// Stores a value in cache with given ttl
func (c *InMemoryCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = cacheItem{
		value:      value,
		expiration: time.Now().Add(c.ttl),
	}
}

// removes all expired items from cache
func (c *InMemoryCache) removeExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if now.After(item.expiration) {
			delete(c.items, key)
		}
	}
}

// periodically removes expired items from cache
func (c *InMemoryCache) cleanup() {
	ticker := time.NewTicker(c.ttl / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.removeExpired()
		case <-c.stopChan:
			return
		}
	}
}

// stops the cleanup goroutine
func (c *InMemoryCache) Stop() {
	close(c.stopChan)
}

// returns the current number of items in cache
func (c *InMemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}
