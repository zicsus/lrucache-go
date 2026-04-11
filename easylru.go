package easylru

import (
	"sync"
	"time"
)

type LRUCache[K comparable, V any] struct {
	mu       sync.Mutex
	cache    map[K]*node[K, V]
	head     *node[K, V]
	tail     *node[K, V]
	size     int
	capacity int
	ttl      time.Duration
	done     chan struct{}
}

type node[K comparable, V any] struct {
	key       K
	value     V
	prev      *node[K, V]
	next      *node[K, V]
	createdAt time.Time
}

func createNode[K comparable, V any](key K, value V) *node[K, V] {
	var node = node[K, V]{
		key:       key,
		value:     value,
		createdAt: time.Now(),
	}
	return &node
}

func (c *LRUCache[K, V]) startCleanup() {
	timer := time.NewTicker(c.ttl / 2)
	go func() {
		for {
			select {
			case <-timer.C:
				c.mu.Lock()
				for _, node := range c.cache {
					if c.isNodeExpired((node)) {
						delete(c.cache, node.key)
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

func (c *LRUCache[K, V]) addToFront(node *node[K, V]) {
	node.next = c.head.next
	node.prev = c.head
	c.head.next.prev = node
	c.head.next = node
	c.size++
}

func (c *LRUCache[K, V]) remove(node *node[K, V]) {
	node.prev.next = node.next
	node.next.prev = node.prev
	node.next = nil
	node.prev = nil
	c.size--
}

func New[K comparable, V any](capacity int, ttl time.Duration) *LRUCache[K, V] {
	var head = node[K, V]{}
	var tail = node[K, V]{}

	head.next = &tail
	tail.prev = &head

	var lruCache = LRUCache[K, V]{
		cache:    map[K]*node[K, V]{},
		head:     &head,
		tail:     &tail,
		size:     0,
		capacity: capacity,
		ttl:      ttl,
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

func (c *LRUCache[K, V]) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.size
}

func (c *LRUCache[K, V]) isNodeExpired(node *node[K, V]) bool {
	return time.Now().After(node.createdAt.Add(c.ttl))
}

func (c *LRUCache[K, V]) Put(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, ok := c.cache[key]
	if ok {
		node.value = value
		node.createdAt = time.Now()
		c.remove(node)
		c.addToFront(node)
	} else if c.size < c.capacity {
		var newNode = createNode(key, value)
		c.addToFront(newNode)
		c.cache[key] = newNode
	} else {
		var newNode = createNode(key, value)
		delete(c.cache, c.tail.prev.key)
		c.remove(c.tail.prev)
		c.addToFront(newNode)
		c.cache[key] = newNode
	}
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, ok := c.cache[key]
	if !ok {
		var zero V
		return zero, false
	}

	if c.ttl > 0 && c.isNodeExpired(node) {
		c.remove(node)
		delete(c.cache, node.key)
		var zero V
		return zero, false
	}

	c.remove(node)
	c.addToFront(node)
	return node.value, true
}

func (c *LRUCache[K, V]) Peek(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, ok := c.cache[key]
	if !ok || (c.ttl > 0 && c.isNodeExpired(node)) {
		var zero V
		return zero, false
	}

	return node.value, true
}

func (c *LRUCache[K, V]) Contains(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, ok := c.cache[key]
	if !ok || (c.ttl > 0 && c.isNodeExpired(node)) {
		return false
	}

	return true
}

func (c *LRUCache[K, V]) Delete(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, ok := c.cache[key]
	if !ok {
		return false
	}

	c.remove(node)
	delete(c.cache, node.key)
	return true
}

func (c *LRUCache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[K]*node[K, V])
	c.head.next = c.tail
	c.tail.prev = c.head
	c.size = 0
}
