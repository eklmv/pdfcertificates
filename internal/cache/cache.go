package cache

type Cache[K comparable, V Cacheable] interface {
	Add(key K, value V)
	Get(key K) (value V, ok bool)
	GetOldest() (value V, ok bool)
	Peek(key K) (value V, ok bool)
	Touch(key K)
	Remove(key K)
	RemoveOldest()
	Purge()
	Contains(key K) bool
	Keys() []K
	Values() []V
	Capacity() uint64
	Len() uint64
	Size() uint64
	Resize(capacity uint64)
}

type Cacheable interface {
	Size() uint64
}
