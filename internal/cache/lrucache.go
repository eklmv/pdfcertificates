package cache

type LRUCache[K comparable, V Cacheable] struct {
	capacity   uint64
	used       uint64
	head       *node[K, V]
	tail       *node[K, V]
	items      map[K]*node[K, V]
	onEviction func(key K, value V)
}

type node[K comparable, V Cacheable] struct {
	next  *node[K, V]
	prev  *node[K, V]
	key   K
	value V
}

func NewLRUCache[K comparable, V Cacheable](capacity uint64) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		capacity: capacity,
		items:    make(map[K]*node[K, V]),
	}
}

func NewLRUCacheWithEviction[K comparable, V Cacheable](capacity uint64, evictionCallback func(key K, value V)) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		capacity:   capacity,
		items:      make(map[K]*node[K, V]),
		onEviction: evictionCallback,
	}
}

func (c *LRUCache[K, V]) addToHead(item *node[K, V]) {
	item.prev = c.head
	item.next = nil
	if c.head != nil {
		c.head.next = item
	}
	if c.tail == nil {
		c.tail = item
	}
	c.head = item
}

func (c *LRUCache[K, V]) removeFromList(item *node[K, V]) {
	if item == c.tail {
		c.tail = item.next
	}
	if item == c.head {
		c.head = item.prev
	}
	if item.prev != nil {
		item.prev.next = item.next
	}
	if item.next != nil {
		item.next.prev = item.prev
	}
	item.prev = nil
	item.next = nil
}

func (c *LRUCache[K, V]) Add(key K, value V) {
	if c.Contains(key) {
		c.Remove(key)
	}
	n := node[K, V]{
		key:   key,
		value: value,
	}
	c.items[key] = &n
	c.addToHead(&n)
	c.used += value.Size()
	for c.capacity != 0 && c.used > c.capacity {
		c.RemoveOldest()
	}
}

func (c *LRUCache[K, V]) Get(key K) (value V, ok bool) {
	if n, ok := c.items[key]; ok {
		c.removeFromList(n)
		c.addToHead(n)
		return n.value, true
	}
	return
}

func (c *LRUCache[K, V]) GetOldest() (value V, ok bool) {
	if c.tail == nil {
		return
	}
	if n, ok := c.items[c.tail.key]; ok {
		c.removeFromList(n)
		c.addToHead(n)
		return n.value, true
	}
	return
}

func (c *LRUCache[K, V]) Peek(key K) (value V, ok bool) {
	if n, ok := c.items[key]; ok {
		return n.value, true
	}
	return
}

func (c *LRUCache[K, V]) Touch(key K) {
	if n, ok := c.items[key]; ok {
		c.removeFromList(n)
		c.addToHead(n)
	}
}

func (c *LRUCache[K, V]) Remove(key K) {
	if n, ok := c.items[key]; ok {
		delete(c.items, key)
		c.removeFromList(n)
		c.used -= n.value.Size()
		if c.onEviction != nil {
			c.onEviction(n.key, n.value)
		}
	}
}

func (c *LRUCache[K, V]) RemoveOldest() {
	if c.tail != nil {
		c.Remove(c.tail.key)
	}
}

func (c *LRUCache[K, V]) Purge() {
	for _, k := range c.Keys() {
		c.Remove(k)
	}
}

func (c *LRUCache[K, V]) Contains(key K) bool {
	_, ok := c.items[key]
	return ok
}

func (c *LRUCache[K, V]) Keys() []K {
	keys := make([]K, 0, c.Len())
	for n := c.tail; n != nil; n = n.next {
		keys = append(keys, n.key)
	}
	return keys
}

func (c *LRUCache[K, V]) Values() []V {
	values := make([]V, 0, c.Len())
	for n := c.tail; n != nil; n = n.next {
		values = append(values, n.value)
	}
	return values
}

func (c *LRUCache[K, V]) Capacity() uint64 {
	return c.capacity
}

func (c *LRUCache[K, V]) Len() uint64 {
	return uint64(len(c.items))
}

func (c *LRUCache[K, V]) Size() uint64 {
	return c.used
}

func (c *LRUCache[K, V]) Resize(capacity uint64) {
	old := c.capacity
	c.capacity = capacity
	if c.capacity >= old {
		return
	}
	for c.capacity != 0 && c.used > c.capacity {
		c.RemoveOldest()
	}
}
