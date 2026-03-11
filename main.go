package main

import (
	"fmt"
	"sync"
)

type LRUCache[K comparable, V any] struct {
	mu       sync.Mutex
	Cache    map[K]*Node[K, V]
	Head     *Node[K, V]
	Tail     *Node[K, V]
	Size     int
	Capacity int
}

type Node[K comparable, V any] struct {
	Key   K
	Value V
	Prev  *Node[K, V]
	Next  *Node[K, V]
}

func main() {
	cache := NewLRUCache[int, int](2)
	cache.Put(1, 10)
	cache.Put(2, 20)
	fmt.Println(cache.Get(1))
	cache.Put(3, 30)
	fmt.Println(cache.Get(2))
	fmt.Println(cache.Get(3))

	cache.Put(3, 99)
	fmt.Println(cache.Get(3))

	fmt.Println(cache.Get(1))
	cache.Put(4, 40)
	fmt.Println(cache.Get(1))
	fmt.Println(cache.Get(3))
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	var head = Node[K, V]{}
	var tail = Node[K, V]{}

	head.Next = &tail
	tail.Prev = &head

	var lruCache = LRUCache[K, V]{
		Cache:    map[K]*Node[K, V]{},
		Head:     &head,
		Tail:     &tail,
		Size:     0,
		Capacity: capacity,
	}

	return &lruCache
}

func (c *LRUCache[K, V]) AddToFront(node *Node[K, V]) {
	node.Next = c.Head.Next
	node.Prev = c.Head
	c.Head.Next.Prev = node
	c.Head.Next = node
}

func (c *LRUCache[K, V]) Remove(node *Node[K, V]) {
	node.Prev.Next = node.Next
	node.Next.Prev = node.Prev
	node.Next = nil
	node.Prev = nil
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, ok := c.Cache[key]
	if !ok {
		var zero V
		return zero, false
	}

	c.Remove(node)
	c.AddToFront(node)
	return node.Value, true
}

func (c *LRUCache[K, V]) Put(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, ok := c.Cache[key]
	if ok {
		node.Value = value
		c.Remove(node)
		c.AddToFront(node)
	} else if c.Size < c.Capacity {
		var newNode = Node[K, V]{
			Key:   key,
			Value: value,
		}
		c.AddToFront(&newNode)
		c.Cache[key] = &newNode
		c.Size++
	} else {
		var newNode = Node[K, V]{
			Key:   key,
			Value: value,
		}

		delete(c.Cache, c.Tail.Prev.Key)
		c.Remove(c.Tail.Prev)
		c.AddToFront(&newNode)
		c.Cache[key] = &newNode
	}
}
