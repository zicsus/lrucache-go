package easylru

import (
	"sync"
	"time"
)

type LRUCache[K comparable, V any] struct {
	mu       sync.Mutex
	Cache    map[K]*Node[K, V]
	head     *Node[K, V]
	tail     *Node[K, V]
	size     int
	capacity int
	ttl      time.Duration
	done     chan struct{}
}

type Node[K comparable, V any] struct {
	Key       K
	Value     V
	prev      *Node[K, V]
	next      *Node[K, V]
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
	timer := time.NewTicker(c.ttl / 2)
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
	node.next = c.head.next
	node.prev = c.head
	c.head.next.prev = node
	c.head.next = node
	c.size++
}

func (c *LRUCache[K, V]) remove(node *Node[K, V]) {
	node.prev.next = node.next
	node.next.prev = node.prev
	node.next = nil
	node.prev = nil
	c.size--
}

func New[K comparable, V any](capacity int, ttl time.Duration) *LRUCache[K, V] {
	var head = Node[K, V]{}
	var tail = Node[K, V]{}

	head.next = &tail
	tail.prev = &head

	var lruCache = LRUCache[K, V]{
		Cache:    map[K]*Node[K, V]{},
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

func (c *LRUCache[K, V]) IsNodeExpired(node *Node[K, V]) bool {
	return time.Now().After(node.CreatedAt.Add(c.ttl))
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
	} else if c.size < c.capacity {
		var newNode = createNode(key, value)
		c.addToFront(newNode)
		c.Cache[key] = newNode
	} else {
		var newNode = createNode(key, value)
		delete(c.Cache, c.tail.prev.Key)
		c.remove(c.tail.prev)
		c.addToFront(newNode)
		c.Cache[key] = newNode
	}
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, ok := c.Cache[key]
	if !ok {
		var zero V
		return zero, false
	}

	if c.ttl > 0 && c.IsNodeExpired(node) {
		c.remove(node)
		delete(c.Cache, node.Key)
		var zero V
		return zero, false
	}

	c.remove(node)
	c.addToFront(node)
	return node.Value, true
}

func (c *LRUCache[K, V]) Peak(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, ok := c.Cache[key]
	if !ok {
		var zero V
		return zero, false
	}

	return node.Value, true
}
