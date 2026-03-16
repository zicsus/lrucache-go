package cmd

import (
	"fmt"

	"github.com/zicsus/lrucache-go/lru"
)

func main() {
	cache := lru.NewLRUCache[string, int](10, 0)
	cache.Put("hello", 42)
	fmt.Println(cache.Get("hello"))
}
