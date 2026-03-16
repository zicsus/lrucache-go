package cmd

import (
	"fmt"

	lru "github.com/zicsus/lrucache-go"
)

func main() {
	cache := lru.NewLRUCache[string, int](10, 0)
	cache.Put("hello", 42)
	fmt.Println(cache.Get("hello"))
}
