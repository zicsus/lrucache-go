# easylru-go

A small, generic LRU cache for Go with optional TTL expiration.

## Install

```sh
go get github.com/zicsus/easylru-go
```

Requires Go generics.

## Usage

```go
package main

import (
	"fmt"
	"time"

	"github.com/zicsus/easylru-go"
)

func main() {
	// capacity = 100, ttl = 5 minutes (use 0 for no TTL)
	cache := easylru.New[string, int](100, 5*time.Minute)
	defer cache.Close()

	cache.Put("answer", 42)

	if v, ok := cache.Get("answer"); ok {
		fmt.Println(v) // 42
	}
}
```

### No TTL

Pass `0` as the TTL to disable expiration. No background cleanup goroutine will run.

```go
cache := easylru.New[string, string](1000, 0)
```

### Eviction

When the cache is full, the least recently used entry is evicted on `Put`:

```go
cache := easylru.New[string, int](2, 0)
cache.Put("a", 1)
cache.Put("b", 2)
cache.Get("a")    // "a" is now most recent
cache.Put("c", 3) // evicts "b"
```

## API

```go
func New[K comparable, V any](capacity int, ttl time.Duration) *LRUCache[K, V]
```
Creates a cache. If `ttl > 0`, a background goroutine periodically removes expired entries.

```go
func (c *LRUCache[K, V]) Put(key K, value V)
```
Inserts or updates an entry and marks it most recently used. Evicts the LRU entry when full.

```go
func (c *LRUCache[K, V]) Get(key K) (V, bool)
```
Returns the value and marks it most recently used. Returns the zero value and `false` if the key is missing or expired.

```go
func (c *LRUCache[K, V]) Peek(key K) (V, bool)
```
Returns the value without updating recency or checking expiration.

```go
func (c *LRUCache[K, V]) Size() int
```
Returns the current number of entries.

```go
func (c *LRUCache[K, V]) Close()
```
Stops the background TTL cleanup goroutine. Call this when you're done with a TTL-enabled cache.

## License

MIT
