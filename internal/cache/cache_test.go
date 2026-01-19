package cache

import (
	"testing"
	"time"
)

func TestInMemoryCache_SetGet(t *testing.T) {
	cache := NewInMemoryCache(1 * time.Minute)
	defer cache.Stop()

	// test Set and Get
	cache.Set("key1", "value1")

	value, found := cache.Get("key1")
	if !found {
		t.Error("Expected to find key1 in cache")
	}
	if value != "value1" {
		t.Errorf("Expected value1 but go %v", value)
	}
}

func TestInMemoryCache_GetNonExistent(t *testing.T) {
	cache := NewInMemoryCache(1 * time.Minute)
	defer cache.Stop()

	value, found := cache.Get("nonexistent")
	if found {
		t.Error("Expected not to find nonexistent key")
	}
	if value != nil {
		t.Errorf("Expected nil value, got %v", value)
	}
}

func TestInMemoryCache_Expiration(t *testing.T) {
	cache := NewInMemoryCache(50 * time.Millisecond)
	defer cache.Stop()

	cache.Set("key1", "value1")

	_, found := cache.Get("key1")
	if !found {
		t.Error("Expected to find key1 immediately after setting")
	}

	time.Sleep(100 * time.Millisecond)

	_, found = cache.Get("key1")
	if found {
		t.Error("Expected key1 to be expired")
	}
}

func TestInMemoryCache_Overwrite(t *testing.T) {
	cache := NewInMemoryCache(1 * time.Minute)
	defer cache.Stop()

	cache.Set("key1", "value1")
	cache.Set("key1", "value2")

	value, found := cache.Get("key1")
	if !found {
		t.Error("Expected to find key1")
	}
	if value != "value2" {
		t.Errorf("Expected value2, got %v", value)
	}
}

func TestInMemoryCache_Size(t *testing.T) {
	cache := NewInMemoryCache(1 * time.Minute)
	defer cache.Stop()

	if cache.Size() != 0 {
		t.Errorf("Expected cache to be empty, got size %d", cache.Size())
	}

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	if cache.Size() != 2 {
		t.Errorf("Expected size to be 2, we got %d", cache.Size())
	}
}

func TestInMemoryCache_DifferentTypes(t *testing.T) {
	cache := NewInMemoryCache(1 * time.Minute)
	defer cache.Stop()

	// testing with different datatypes
	cache.Set("string", "hello")
	cache.Set("int", 42)
	cache.Set("struct", struct{ Name string }{"test"})

	if v, _ := cache.Get("string"); v != "hello" {
		t.Errorf("Expected string value, we got %v", v)
	}
	if v, _ := cache.Get("int"); v != 42 {
		t.Errorf("Expected int value, we got %v", v)
	}
	if v, _ := cache.Get("struct"); v != struct{ Name string }{"test"} {
		t.Errorf("Expected struct value, we got %v", v)
	}
}
