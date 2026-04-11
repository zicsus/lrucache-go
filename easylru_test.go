package easylru

import (
	"testing"
)

func TestGetAndPut(t *testing.T) {
	cache := New[string, int](10, 0)
	cache.Put("key1", 1)
	value, ok := cache.Get("key1")
	if !ok {
		t.Errorf("key1 should be in the cache but got %v", ok)
	}
	if value != 1 {
		t.Errorf("key1 should be 1 but got %v", value)
	}
}

func TestKeyThatDoesNotExist(t *testing.T) {
	cache := New[string, int](10, 0)
	value, ok := cache.Get("key1")
	if ok {
		t.Errorf("key1 should not be in the cache but got %v", ok)
	}
	if value != 0 {
		t.Errorf("key1 should be 0 but got %v", value)
	}
}

func TestEviction(t *testing.T) {
	cache := New[string, int](2, 0)
	cache.Put("key1", 1)
	cache.Put("key2", 2)
	cache.Get("key1")
	cache.Put("key3", 3)

	value, ok := cache.Get("key2")
	if ok {
		t.Errorf("key2 should not be in the cache but got %v", ok)
	}
	if value != 0 {
		t.Errorf("key2 should be 0 but got %v", value)
	}
}
