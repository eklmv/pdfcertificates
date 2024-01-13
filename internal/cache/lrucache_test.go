package cache

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testValue struct {
	value int
}

func (testValue) Size() uint64 {
	return 1
}

func TestLRUCacheImplementsInterface(t *testing.T) {
	assert.Implements(t, (*Cache[string, testValue])(nil), new(LRUCache[string, testValue]))
}

func TestLRUCacheAdd(t *testing.T) {
	t.Parallel()
	t.Run("Add multiple pairs with different keys", func(t *testing.T) {
		t.Parallel()
		amount := 4
		var expKeys []string
		var expValues []testValue

		got := NewLRUCache[string, testValue](0)
		for i := 0; i < amount; i++ {
			expKeys = append(expKeys, fmt.Sprintf("key %d", i))
			expValues = append(expValues, testValue{i})
			got.Add(expKeys[i], expValues[i])
		}

		assert.Equal(t, expKeys, got.Keys())
		assert.Equal(t, expValues, got.Values())
		assert.Equal(t, uint64(amount), got.Len())
		assert.Equal(t, uint64(amount)*testValue{}.Size(), got.Size())
	})
	t.Run("Add pairs with same key, replacing old value with eviction", func(t *testing.T) {
		t.Parallel()
		expKey := "key"
		oldValue := testValue{1}
		expValue := testValue{2}
		expCalled := false

		got := NewLRUCacheWithEviction[string, testValue](0, func(key string, value testValue) {
			expCalled = true
			assert.Equal(t, expKey, key)
			assert.Equal(t, oldValue, value)
		})
		got.Add(expKey, oldValue)
		got.Add(expKey, expValue)

		assert.Equal(t, []string{expKey}, got.Keys())
		assert.Equal(t, []testValue{expValue}, got.Values())
		assert.Equal(t, uint64(1), got.Len())
		assert.Equal(t, expValue.Size(), got.Size())
		assert.True(t, expCalled)
	})
	t.Run("Overflow capacity, evicting old pairs", func(t *testing.T) {
		t.Parallel()
		amount := 4
		expLen := amount / 2
		expCap := uint64(expLen) * testValue{}.Size()
		var expKeys, evictedKeys, cbKeys []string
		var expValues, evictedValues, cbValues []testValue
		expCalled := amount - expLen
		called := 0

		got := NewLRUCacheWithEviction[string, testValue](expCap, func(key string, value testValue) {
			called++
			cbKeys = append(cbKeys, key)
			cbValues = append(cbValues, value)
		})
		for i := 0; i < amount; i++ {
			key := fmt.Sprintf("key %d", i)
			value := testValue{i}
			if i < expLen {
				evictedKeys = append(evictedKeys, key)
				evictedValues = append(evictedValues, value)
			} else {
				expKeys = append(expKeys, key)
				expValues = append(expValues, value)
			}
			got.Add(key, value)
		}

		assert.Equal(t, expKeys, got.Keys())
		assert.Equal(t, expValues, got.Values())
		assert.Equal(t, uint64(expLen), got.Len())
		assert.Equal(t, expCap, got.Capacity())
		assert.Equal(t, expCap, got.Size())
		assert.Equal(t, evictedKeys, cbKeys)
		assert.Equal(t, evictedValues, cbValues)
		assert.Equal(t, expCalled, called)
	})
}

func TestLRUCacheGet(t *testing.T) {
	t.Parallel()
	t.Run("Return value for existing key, update recency", func(t *testing.T) {
		t.Parallel()
		amount := 4
		expKey := "existing key"
		expValue := testValue{math.MaxInt}
		var keys []string
		var values []testValue

		cache := NewLRUCache[string, testValue](0)
		cache.Add(expKey, expValue)
		for i := 0; i < amount; i++ {
			key := fmt.Sprintf("key %d", i)
			value := testValue{i}
			cache.Add(key, value)
			keys = append(keys, key)
			values = append(values, value)
		}
		assert.Equal(t, append([]string{expKey}, keys...), cache.Keys())
		assert.Equal(t, append([]testValue{expValue}, values...), cache.Values())

		got, ok := cache.Get(expKey)

		assert.True(t, ok)
		assert.Equal(t, expValue, got)
		assert.Equal(t, append(keys, expKey), cache.Keys())
		assert.Equal(t, append(values, expValue), cache.Values())
	})
	t.Run("Should not return value for alredy removed key", func(t *testing.T) {
		t.Parallel()
		key := "key"
		value := testValue{0}

		cache := NewLRUCache[string, testValue](0)
		cache.Add(key, value)
		cache.Remove(key)

		got, ok := cache.Get(key)

		assert.False(t, ok)
		assert.Empty(t, got)
	})
}

func TestLRUCacheGetOldest(t *testing.T) {
	t.Parallel()
	t.Run("Return oldest value, update recency", func(t *testing.T) {
		t.Parallel()
		amount := 4
		var keys []string
		var values []testValue

		cache := NewLRUCache[string, testValue](0)
		for i := 0; i < amount; i++ {
			keys = append(keys, fmt.Sprintf("key %d", i))
			values = append(values, testValue{i})
			cache.Add(keys[i], values[i])
		}
		assert.Equal(t, keys, cache.Keys())
		assert.Equal(t, values, cache.Values())

		got, ok := cache.GetOldest()

		assert.True(t, ok)
		assert.Equal(t, values[0], got)
		assert.Equal(t, append(keys[1:], keys[0]), cache.Keys())
		assert.Equal(t, append(values[1:], values[0]), cache.Values())
	})
	t.Run("Empty cache should return nothing", func(t *testing.T) {
		t.Parallel()
		cache := NewLRUCache[string, testValue](0)

		got, ok := cache.GetOldest()

		assert.False(t, ok)
		assert.Empty(t, got)
	})
}

func TestLRUCachePeek(t *testing.T) {
	t.Parallel()
	t.Run("Return value for existing key, recency should not be updated", func(t *testing.T) {
		t.Parallel()
		amount := 4
		var expKeys []string
		var expValues []testValue

		cache := NewLRUCache[string, testValue](0)
		for i := 0; i < amount; i++ {
			key := fmt.Sprintf("key %d", i)
			value := testValue{i}
			cache.Add(key, value)
			expKeys = append(expKeys, key)
			expValues = append(expValues, value)
		}

		got, ok := cache.Peek(expKeys[0])

		assert.True(t, ok)
		assert.Equal(t, expValues[0], got)
		assert.Equal(t, expKeys, cache.Keys())
		assert.Equal(t, expValues, cache.Values())
	})
	t.Run("Should not return value for already removed key", func(t *testing.T) {
		t.Parallel()
		key := "key"
		value := testValue{0}

		cache := NewLRUCache[string, testValue](0)
		cache.Add(key, value)
		cache.Remove(key)

		got, ok := cache.Peek(key)

		assert.False(t, ok)
		assert.Empty(t, got)
	})
}

func TestLRUCacheTouch(t *testing.T) {
	t.Parallel()
	t.Run("Update recency for existing key", func(t *testing.T) {
		t.Parallel()
		amount := 4
		var keys []string
		var values []testValue

		cache := NewLRUCache[string, testValue](0)
		for i := 0; i < amount; i++ {
			key := fmt.Sprintf("key %d", i)
			value := testValue{i}
			cache.Add(key, value)
			keys = append(keys, key)
			values = append(values, value)
		}
		assert.Equal(t, keys, cache.Keys())
		assert.Equal(t, values, cache.Values())

		cache.Touch(keys[0])

		assert.Equal(t, append(keys[1:], keys[0]), cache.Keys())
		assert.Equal(t, append(values[1:], values[0]), cache.Values())
	})
	t.Run("Should not affect recency if key doesn't exists", func(t *testing.T) {
		t.Parallel()
		amount := 4
		var keys []string
		var values []testValue

		cache := NewLRUCache[string, testValue](0)
		for i := 0; i < amount; i++ {
			key := fmt.Sprintf("key %d", i)
			value := testValue{i}
			cache.Add(key, value)
			keys = append(keys, key)
			values = append(values, value)
		}
		assert.Equal(t, keys, cache.Keys())
		assert.Equal(t, values, cache.Values())

		cache.Touch("does not exists")

		assert.Equal(t, keys, cache.Keys())
		assert.Equal(t, values, cache.Values())
	})
}

func TestLRUCacheRemove(t *testing.T) {
	t.Parallel()
	t.Run("Remove pair for existing key, eviction callback should be triggered", func(t *testing.T) {
		t.Parallel()
		amount := 4
		var expKeys []string
		var expValues []testValue
		expKey := "target key"
		expValue := testValue{math.MaxInt}
		expCalled := false

		got := NewLRUCacheWithEviction[string, testValue](0, func(key string, value testValue) {
			expCalled = true
			assert.Equal(t, expKey, key)
			assert.Equal(t, expValue, value)
		})
		for i := 0; i < amount; i++ {
			key := fmt.Sprintf("key %d", i)
			value := testValue{i}
			got.Add(key, value)
			expKeys = append(expKeys, key)
			expValues = append(expValues, value)
		}
		got.Add(expKey, expValue)

		got.Remove(expKey)

		assert.True(t, expCalled)
		assert.Equal(t, expKeys, got.Keys())
		assert.Equal(t, expValues, got.Values())
		assert.Equal(t, uint64(amount), got.Len())
		assert.Equal(t, uint64(amount)*testValue{}.Size(), got.Size())
	})
	t.Run("Should not affect cache if key doesn't exists", func(t *testing.T) {
		t.Parallel()
		amount := 4
		var expKeys []string
		var expValues []testValue

		cache := NewLRUCache[string, testValue](0)
		for i := 0; i < amount; i++ {
			key := fmt.Sprintf("key %d", i)
			value := testValue{i}
			cache.Add(key, value)
			expKeys = append(expKeys, key)
			expValues = append(expValues, value)
		}

		cache.Remove("does not exists")

		assert.Equal(t, expKeys, cache.Keys())
		assert.Equal(t, expValues, cache.Values())
		assert.Equal(t, uint64(amount), cache.Len())
		assert.Equal(t, uint64(amount)*testValue{}.Size(), cache.Size())
	})
}

func TestLRUCacheRemoveOldest(t *testing.T) {
	t.Parallel()
	t.Run("Remove oldest pair, eviction callback should be triggered", func(t *testing.T) {
		t.Parallel()
		amount := 4
		var expKeys []string
		var expValues []testValue
		expKey := "target key"
		expValue := testValue{math.MaxInt}
		expCalled := false

		got := NewLRUCacheWithEviction[string, testValue](0, func(key string, value testValue) {
			expCalled = true
			assert.Equal(t, expKey, key)
			assert.Equal(t, expValue, value)
		})
		got.Add(expKey, expValue)
		for i := 0; i < amount; i++ {
			key := fmt.Sprintf("key %d", i)
			value := testValue{i}
			got.Add(key, value)
			expKeys = append(expKeys, key)
			expValues = append(expValues, value)
		}

		got.Remove(expKey)

		assert.True(t, expCalled)
		assert.Equal(t, expKeys, got.Keys())
		assert.Equal(t, expValues, got.Values())
		assert.Equal(t, uint64(amount), got.Len())
		assert.Equal(t, uint64(amount)*testValue{}.Size(), got.Size())
	})
}

func TestLRUCachePurge(t *testing.T) {
	t.Parallel()
	t.Run("Purge cache, eviction callback should be triggered for each pair", func(t *testing.T) {
		t.Parallel()
		amount := 4
		var expKeys, evictedKeys []string
		var expValues, evictedValues []testValue
		evicted := 0

		got := NewLRUCacheWithEviction[string, testValue](0, func(key string, value testValue) {
			evicted++
			evictedKeys = append(evictedKeys, key)
			evictedValues = append(evictedValues, value)
		})
		for i := 0; i < amount; i++ {
			key := fmt.Sprintf("key %d", i)
			value := testValue{i}
			expKeys = append(expKeys, key)
			expValues = append(expValues, value)
			got.Add(key, value)
		}

		got.Purge()

		assert.Equal(t, amount, evicted)
		assert.Equal(t, expKeys, evictedKeys)
		assert.Equal(t, expValues, evictedValues)
		assert.Empty(t, got.Keys())
		assert.Empty(t, got.Values())
		assert.Zero(t, got.Len())
		assert.Zero(t, got.Size())
	})
}

func TestLRUCacheContains(t *testing.T) {
	t.Parallel()
	t.Run("True for existing key", func(t *testing.T) {
		t.Parallel()
		key := "existing key"
		value := testValue{0}

		cache := NewLRUCache[string, testValue](0)
		cache.Add(key, value)

		got := cache.Contains(key)

		assert.True(t, got)
	})
	t.Run("False for non existing key", func(t *testing.T) {
		t.Parallel()
		key := "existing key"
		value := testValue{0}

		cache := NewLRUCache[string, testValue](0)
		cache.Add(key, value)

		got := cache.Contains("does not exists")

		assert.False(t, got)
	})
}

func TestLRUCacheKeys(t *testing.T) {
	t.Parallel()
	t.Run("Return keys from old to new, should not affect recency", func(t *testing.T) {
		t.Parallel()
		amount := 4
		var expKeys []string

		got := NewLRUCache[string, testValue](uint64(amount) * testValue{}.Size())
		for i := 0; i < amount; i++ {
			expKeys = append(expKeys, fmt.Sprintf("key %d", i))
			got.Add(expKeys[i], testValue{i})
		}

		assert.Equal(t, expKeys, got.Keys())
		assert.Equal(t, expKeys, got.Keys())
	})
}

func TestLRUCacheValues(t *testing.T) {
	t.Parallel()
	t.Run("Return values from old to new, should not affect recency", func(t *testing.T) {
		t.Parallel()
		amount := 4
		var expValues []testValue

		got := NewLRUCache[string, testValue](uint64(amount) * testValue{}.Size())
		for i := 0; i < amount; i++ {
			expValues = append(expValues, testValue{i})
			got.Add(fmt.Sprintf("key %d", i), expValues[i])
		}

		assert.Equal(t, expValues, got.Values())
		assert.Equal(t, expValues, got.Values())
	})
}

func TestLRUCacheCapacity(t *testing.T) {
	t.Parallel()
	t.Run("Return cache capacity", func(t *testing.T) {
		t.Parallel()
		expCap := rand.Uint64()

		got := NewLRUCache[string, testValue](expCap)

		assert.Equal(t, expCap, got.Capacity())
	})
}

func TestLRUCacheLen(t *testing.T) {
	t.Parallel()
	t.Run("Return cache length", func(t *testing.T) {
		t.Parallel()
		expLen := rand.Intn(100)

		got := NewLRUCache[string, testValue](0)
		for i := 0; i < expLen; i++ {
			got.Add(fmt.Sprintf("key %d", i), testValue{i})
		}

		assert.Equal(t, uint64(expLen), got.Len())
	})
}

func TestLRUCacheSize(t *testing.T) {
	t.Parallel()
	t.Run("Return cache size", func(t *testing.T) {
		t.Parallel()
		amount := rand.Intn(100)
		expSize := uint64(amount) * testValue{}.Size()

		got := NewLRUCache[string, testValue](0)
		for i := 0; i < amount; i++ {
			got.Add(fmt.Sprintf("key %d", i), testValue{i})
		}

		assert.Equal(t, expSize, got.Size())
	})
}

func TestLRUCacheResize(t *testing.T) {
	t.Parallel()
	t.Run("Growing capacity should not affect cache", func(t *testing.T) {
		t.Parallel()
		amount := 4
		initCapacity := uint64(amount) * testValue{}.Size()
		expCapacity := initCapacity * 2
		expSize := uint64(amount) * testValue{}.Size()
		var expKeys []string
		var expValues []testValue
		called := false

		got := NewLRUCacheWithEviction[string, testValue](initCapacity, func(_ string, _ testValue) {
			called = true
		})
		for i := 0; i < amount; i++ {
			key := fmt.Sprintf("key %d", i)
			value := testValue{i}
			got.Add(key, value)
			expKeys = append(expKeys, key)
			expValues = append(expValues, value)
		}

		got.Resize(expCapacity)

		assert.Equal(t, expCapacity, got.Capacity())
		assert.Equal(t, expKeys, got.Keys())
		assert.Equal(t, expValues, got.Values())
		assert.Equal(t, uint64(amount), got.Len())
		assert.Equal(t, expSize, got.Size())
		assert.False(t, called)
	})
	t.Run("Shrinking capacity should remove old pairs with eviction", func(t *testing.T) {
		t.Parallel()
		initLen := 10
		initCapacity := uint64(initLen) * testValue{}.Size()
		expLen := initLen / 2
		expCapacity := uint64(expLen) * testValue{}.Size()
		expSize := uint64(expLen) * testValue{}.Size()
		var expKeys, evictedKeys, cbKeys []string
		var expValues, evictedValues, cbValues []testValue
		called := 0

		got := NewLRUCacheWithEviction[string, testValue](initCapacity, func(key string, value testValue) {
			called++
			cbKeys = append(cbKeys, key)
			cbValues = append(cbValues, value)
		})
		for i := 0; i < initLen; i++ {
			key := fmt.Sprintf("key %d", i)
			value := testValue{i}
			got.Add(key, value)
			if i < expLen {
				evictedKeys = append(evictedKeys, key)
				evictedValues = append(evictedValues, value)
			} else {
				expKeys = append(expKeys, key)
				expValues = append(expValues, value)
			}
		}

		got.Resize(expCapacity)

		assert.Equal(t, expCapacity, got.Capacity())
		assert.Equal(t, expKeys, got.Keys())
		assert.Equal(t, expValues, got.Values())
		assert.Equal(t, uint64(expLen), got.Len())
		assert.Equal(t, expSize, got.Size())
		assert.Equal(t, initLen-expLen, called)
		assert.Equal(t, evictedKeys, cbKeys)
		assert.Equal(t, evictedValues, cbValues)
	})
}
