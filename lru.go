package lru

import (
	"sync"
	"time"
)

type LRUCache[K comparable, V any] struct {
	mu       sync.Mutex
	Cache    map[K]*Node[K, V]
	Head     *Node[K, V]
	Tail     *Node[K, V]
	Size     int
	Capacity int
	TTL      time.Duration
	done     chan struct{}
}

type Node[K comparable, V any] struct {
	Key       K
	Value     V
	Prev      *Node[K, V]
	Next      *Node[K, V]
	CreatedAt time.Time
}

func createNode[K comparable, V any](key K, value V) *Node[K, V] {
	var node = Node[K, V]{
		Key:       key,
		Value:     value,
		CreatedAt: time.Now(),
	}
	return &node
}

func (c *LRUCache[K, V]) startCleanup() {
	timer := time.NewTicker(c.TTL / 2)
	go func() {
		for {
			select {
			case <-timer.C:
				c.mu.Lock()
				for _, node := range c.Cache {
					if c.IsNodeExpired((node)) {
						delete(c.Cache, node.Key)
						c.remove(node)
					}
				}
				c.mu.Unlock()
			case <-c.done:
				timer.Stop()
				return
			}
		}
	}()
}

func (c *LRUCache[K, V]) addToFront(node *Node[K, V]) {
	node.Next = c.Head.Next
	node.Prev = c.Head
	c.Head.Next.Prev = node
	c.Head.Next = node
	c.Size++
}

func (c *LRUCache[K, V]) remove(node *Node[K, V]) {
	node.Prev.Next = node.Next
	node.Next.Prev = node.Prev
	node.Next = nil
	node.Prev = nil
	c.Size--
}

func NewLRUCache[K comparable, V any](capacity int, ttl time.Duration) *LRUCache[K, V] {
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
		TTL:      ttl,
		done:     make(chan struct{}),
	}

	if ttl > 0 {
		lruCache.startCleanup()
	}

	return &lruCache
}

func (c *LRUCache[K, V]) Close() {
	close(c.done)
}

func (c *LRUCache[K, V]) IsNodeExpired(node *Node[K, V]) bool {
	return time.Now().After(node.CreatedAt.Add(c.TTL))
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, ok := c.Cache[key]
	if !ok {
		var zero V
		return zero, false
	}

	if c.TTL > 0 && c.IsNodeExpired(node) {
		c.remove(node)
		delete(c.Cache, node.Key)
		var zero V
		return zero, false
	}

	c.remove(node)
	c.addToFront(node)
	return node.Value, true
}

func (c *LRUCache[K, V]) Put(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, ok := c.Cache[key]
	if ok {
		node.Value = value
		node.CreatedAt = time.Now()
		c.remove(node)
		c.addToFront(node)
	} else if c.Size < c.Capacity {
		var newNode = createNode(key, value)
		c.addToFront(newNode)
		c.Cache[key] = newNode
	} else {
		var newNode = createNode(key, value)
		delete(c.Cache, c.Tail.Prev.Key)
		c.remove(c.Tail.Prev)
		c.addToFront(newNode)
		c.Cache[key] = newNode
	}
}
