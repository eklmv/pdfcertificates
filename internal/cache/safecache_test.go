package cache

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeCacheImplementsInterface(t *testing.T) {
	assert.Implements(t, (*Cache[string, testValue])(nil), new(SafeCache[string, testValue]))
}

func TestSafeCacheAdd(t *testing.T) {
	t.Parallel()
	t.Run("Add multiple pairs with different keys, concurrently", func(t *testing.T) {
		t.Parallel()
		workers := 10
		amount := 100
		expLen := workers * amount
		expSize := expLen * int(testValue{}.Size())
		expKeys := make([]string, expLen)
		expValues := make([]testValue, expLen)
		cache := NewLRUCache[string, testValue](0)

		got := NewSafeCache[string, testValue](cache)
		var complete sync.WaitGroup
		complete.Add(workers)
		for i := 0; i < workers; i++ {
			go func(i int) {
				defer complete.Done()
				for j := 0; j < amount; j++ {
					index := amount*i + j
					key := fmt.Sprintf("key %d", index)
					value := testValue{index}
					expKeys[index] = key
					expValues[index] = value
					got.Add(key, value)
				}
			}(i)
		}
		complete.Wait()

		assert.ElementsMatch(t, expKeys, got.Keys())
		assert.ElementsMatch(t, expValues, got.Values())
		assert.Equal(t, uint64(expLen), got.Len())
		assert.Equal(t, uint64(expSize), got.Size())
	})
	t.Run("Add pairs with same key, replacing old values with eviction, concurrently", func(t *testing.T) {
		t.Parallel()
		workers := 10
		amount := 100
		expLen := workers * amount
		expSize := expLen * int(testValue{}.Size())
		expEvicted := expLen / 2
		expKeys := make([]string, expLen)
		expValues := make([]testValue, expLen)
		evictedKeys := make([]string, expEvicted)
		evictedValues := make([]testValue, expEvicted)
		cbKeys := make([]string, expEvicted)
		cbValues := make([]testValue, expEvicted)
		var evicted atomic.Int32
		cache := NewLRUCacheWithEviction[string, testValue](0, func(key string, value testValue) {
			evicted.Add(1)
			cbKeys[value.value/2] = key
			cbValues[value.value/2] = value
		})

		got := NewSafeCache[string, testValue](cache)
		var complete sync.WaitGroup
		complete.Add(workers * 2)
		for i := 0; i < workers; i++ {
			go func(i int) {
				defer complete.Done()
				for j := 0; j < amount; j++ {
					index := amount*i + j
					key := fmt.Sprintf("key %d", index)
					value := testValue{index}
					expKeys[index] = key
					expValues[index] = value
					got.Add(key, value)
				}
			}(i)
			go func(i int) {
				defer complete.Done()
				for j := 0; j < amount; j++ {
					if j%2 != 0 {
						index := amount*i + j
						key := fmt.Sprintf("key %d", index)
						value := testValue{index}
						evictedKeys[index/2] = key
						evictedValues[index/2] = value
						got.Add(key, value)
					}
				}
			}(i)
		}
		complete.Wait()

		assert.ElementsMatch(t, expKeys, got.Keys())
		assert.ElementsMatch(t, expValues, got.Values())
		assert.ElementsMatch(t, evictedKeys, cbKeys)
		assert.ElementsMatch(t, evictedValues, cbValues)
		assert.Equal(t, uint64(expLen), got.Len())
		assert.Equal(t, uint64(expSize), got.Size())
		assert.Equal(t, expEvicted, int(evicted.Load()))
	})
	t.Run("Overflow capacity, evicting old pairs, concurrently", func(t *testing.T) {
		t.Parallel()
		workers := 10
		amount := 100
		expCap := workers * amount
		expLen := workers * amount
		expSize := expLen * int(testValue{}.Size())
		expEvicted := expLen / 2
		expKeys := make([]string, expLen)
		expValues := make([]testValue, expLen)
		evictedKeys := make([]string, expEvicted)
		evictedValues := make([]testValue, expEvicted)
		cbKeys := make([]string, expEvicted)
		cbValues := make([]testValue, expEvicted)
		var evicted atomic.Int32
		cache := NewLRUCacheWithEviction[string, testValue](uint64(expCap), func(key string, value testValue) {
			evicted.Add(1)
			cbKeys[value.value] = key
			cbValues[value.value] = value
		})

		got := NewSafeCache[string, testValue](cache)
		for i := 0; i < expEvicted; i++ {
			key := fmt.Sprintf("key %d", i)
			value := testValue{i}
			evictedKeys[i] = key
			evictedValues[i] = value
			got.Add(key, value)
		}
		var complete sync.WaitGroup
		complete.Add(workers)
		for i := 0; i < workers; i++ {
			go func(i int) {
				defer complete.Done()
				for j := 0; j < amount; j++ {
					index := amount*i + j
					key := fmt.Sprintf("key %d", index)
					value := testValue{index}
					expKeys[index] = key
					expValues[index] = value
					got.Add(key, value)
				}
			}(i)
		}
		complete.Wait()

		assert.ElementsMatch(t, expKeys, got.Keys())
		assert.ElementsMatch(t, expValues, got.Values())
		assert.ElementsMatch(t, evictedKeys, cbKeys)
		assert.ElementsMatch(t, evictedValues, cbValues)
		assert.Equal(t, uint64(expLen), got.Len())
		assert.Equal(t, uint64(expSize), got.Size())
		assert.Equal(t, uint64(expCap), got.Capacity())
		assert.Equal(t, expEvicted, int(evicted.Load()))
	})
}

func TestSafeCacheGet(t *testing.T) {
	t.Parallel()
	t.Run("Return values for existing keys, update recency, concurrently", func(t *testing.T) {
		t.Parallel()
		workers := 10
		amount := workers * 100
		getTill := amount / 2
		work := getTill / workers
		var keys []string
		var values []testValue
		cache := NewLRUCache[string, testValue](0)
		sCache := NewSafeCache[string, testValue](cache)
		for i := 0; i < amount; i++ {
			key := fmt.Sprintf("key %d", i)
			value := testValue{i}
			keys = append(keys, key)
			values = append(values, value)
			sCache.Add(key, value)
		}
		assert.Equal(t, keys, sCache.Keys())
		assert.Equal(t, values, sCache.Values())

		var complete sync.WaitGroup
		complete.Add(workers)
		for i := 0; i < workers; i++ {
			start := i * work
			end := start + work
			if i == workers-1 {
				end = getTill
			}
			go func(start, end int) {
				defer complete.Done()
				for j := start; j < end; j++ {
					got, ok := sCache.Get(keys[j])
					assert.True(t, ok)
					assert.Equal(t, values[j], got)
				}
			}(start, end)
		}
		complete.Wait()

		assert.ElementsMatch(t, keys[getTill:], sCache.Keys()[:getTill])
		assert.ElementsMatch(t, values[getTill:], sCache.Values()[:getTill])
	})
}

func TestSafeCacheGetOldest(t *testing.T) {
	t.Parallel()
	t.Run("Return oldest values concurrently, no same values should be returned, update recency", func(t *testing.T) {
		t.Parallel()
		workers := 10
		amount := workers * 100
		oldest := 10
		var keys []string
		var values []testValue
		oldestValues := make([]testValue, workers*oldest)
		cache := NewLRUCache[string, testValue](0)
		sCache := NewSafeCache[string, testValue](cache)
		for i := 0; i < amount; i++ {
			key := fmt.Sprintf("key %d", i)
			value := testValue{i + 1}
			keys = append(keys, key)
			values = append(values, value)
			sCache.Add(key, value)
		}

		var complete sync.WaitGroup
		complete.Add(workers)
		for i := 0; i < workers; i++ {
			go func(i int) {
				defer complete.Done()
				for j := 0; j < oldest; j++ {
					got, ok := sCache.GetOldest()
					assert.True(t, ok)
					assert.Contains(t, values[:oldest*workers], got)
					assert.NotContains(t, oldestValues, got)
					oldestValues[i*oldest+j] = got
				}
			}(i)
		}
		complete.Wait()

		assert.ElementsMatch(t, keys[:oldest*workers], sCache.Keys()[amount-oldest*workers:])
		assert.ElementsMatch(t, values[:oldest*workers], sCache.Values()[amount-oldest*workers:])
	})
}

func TestSafeCachePeek(t *testing.T) {
	t.Parallel()
	t.Run("Return values for existing keys concurrently, recency should not be updated", func(t *testing.T) {
		t.Parallel()
		workers := 10
		amount := workers * 100
		peekTill := amount / 2
		work := peekTill / workers
		var keys []string
		var values []testValue
		cache := NewLRUCache[string, testValue](0)
		sCache := NewSafeCache[string, testValue](cache)
		for i := 0; i < amount; i++ {
			key := fmt.Sprintf("key %d", i)
			value := testValue{i}
			keys = append(keys, key)
			values = append(values, value)
			sCache.Add(key, value)
		}
		assert.Equal(t, keys, sCache.Keys())
		assert.Equal(t, values, sCache.Values())

		var complete sync.WaitGroup
		complete.Add(workers)
		for i := 0; i < workers; i++ {
			start := i * work
			end := start + work
			if i == workers-1 {
				end = peekTill
			}
			go func(start, end int) {
				defer complete.Done()
				for j := start; j < end; j++ {
					got, ok := sCache.Peek(keys[j])
					assert.True(t, ok)
					assert.Equal(t, values[j], got)
				}
			}(start, end)
		}
		complete.Wait()

		assert.ElementsMatch(t, keys, sCache.Keys())
		assert.ElementsMatch(t, values, sCache.Values())
	})
}

func TestSafeCacheTouch(t *testing.T) {
	t.Parallel()
	t.Run("Update recency for existing keys concurrently", func(t *testing.T) {
		t.Parallel()
		workers := 10
		amount := workers * 100
		touchTill := amount / 2
		work := touchTill / workers
		var keys []string
		var values []testValue
		cache := NewLRUCache[string, testValue](0)
		sCache := NewSafeCache[string, testValue](cache)
		for i := 0; i < amount; i++ {
			key := fmt.Sprintf("key %d", i)
			value := testValue{i}
			keys = append(keys, key)
			values = append(values, value)
			sCache.Add(key, value)
		}
		assert.Equal(t, keys, sCache.Keys())
		assert.Equal(t, values, sCache.Values())

		var complete sync.WaitGroup
		complete.Add(workers)
		for i := 0; i < workers; i++ {
			start := i * work
			end := start + work
			if i == workers-1 {
				end = touchTill
			}
			go func(start, end int) {
				defer complete.Done()
				for j := start; j < end; j++ {
					sCache.Touch(keys[j])
				}
			}(start, end)
		}
		complete.Wait()

		assert.ElementsMatch(t, keys[touchTill:], sCache.Keys()[:touchTill])
		assert.ElementsMatch(t, values[touchTill:], sCache.Values()[:touchTill])
	})
}

func TestSafeCacheRemove(t *testing.T) {
	t.Parallel()
	t.Run("Remove pair for existing key concurrently, eviction callback should be triggered", func(t *testing.T) {
		t.Parallel()
		workers := 10
		amount := workers * 100
		removeTill := amount / 2
		work := removeTill / workers
		expLen := amount - removeTill
		expSize := expLen * int(testValue{}.Size())
		expEvicted := removeTill
		var keys []string
		var values []testValue
		cbKeys := make([]string, expEvicted)
		cbValues := make([]testValue, expEvicted)
		var evicted atomic.Int32
		cache := NewLRUCacheWithEviction[string, testValue](0, func(key string, value testValue) {
			evicted.Add(1)
			cbKeys[value.value] = key
			cbValues[value.value] = value
		})

		got := NewSafeCache[string, testValue](cache)
		for i := 0; i < amount; i++ {
			key := fmt.Sprintf("key %d", i)
			value := testValue{i}
			keys = append(keys, key)
			values = append(values, value)
			got.Add(key, value)
		}
		var complete sync.WaitGroup
		complete.Add(workers)
		for i := 0; i < workers; i++ {
			start := i * work
			end := start + work
			if i == workers-1 {
				end = removeTill
			}
			go func(start, end int) {
				defer complete.Done()
				for j := start; j < end; j++ {
					got.Remove(keys[j])
				}
			}(start, end)
		}
		complete.Wait()

		assert.Equal(t, keys[removeTill:], got.Keys())
		assert.Equal(t, values[removeTill:], got.Values())
		assert.ElementsMatch(t, keys[:removeTill], cbKeys)
		assert.ElementsMatch(t, values[:removeTill], cbValues)
		assert.Equal(t, expEvicted, int(evicted.Load()))
		assert.Equal(t, uint64(expLen), got.Len())
		assert.Equal(t, uint64(expSize), got.Size())
	})
}

func TestSafeCacheRemoveOldest(t *testing.T) {
	t.Parallel()
	t.Run("Remove oldest pairs concurrently, eviction callback should be triggered for each pair only once", func(t *testing.T) {
		workers := 10
		amount := workers * 100
		oldest := amount / 2
		work := oldest / workers
		expLen := amount - oldest
		expSize := expLen * int(testValue{}.Size())
		expEvicted := oldest
		var keys []string
		var values []testValue
		cbKeys := make([]string, expEvicted)
		cbValues := make([]testValue, expEvicted)
		var evicted atomic.Int32
		cache := NewLRUCacheWithEviction[string, testValue](0, func(key string, value testValue) {
			evicted.Add(1)
			assert.NotContains(t, cbKeys, key)
			assert.NotContains(t, cbValues, value)
			cbKeys[value.value-1] = key
			cbValues[value.value-1] = value
		})

		got := NewSafeCache[string, testValue](cache)
		for i := 0; i < amount; i++ {
			key := fmt.Sprintf("key %d", i)
			value := testValue{i + 1}
			keys = append(keys, key)
			values = append(values, value)
			got.Add(key, value)
		}
		var complete sync.WaitGroup
		complete.Add(workers)
		for i := 0; i < workers; i++ {
			start := i * work
			end := start + work
			if i == workers-1 {
				end = oldest
			}
			go func(start, end int) {
				defer complete.Done()
				for j := start; j < end; j++ {
					got.RemoveOldest()
				}
			}(start, end)
		}
		complete.Wait()

		assert.Equal(t, keys[oldest:], got.Keys())
		assert.Equal(t, values[oldest:], got.Values())
		assert.ElementsMatch(t, keys[:oldest], cbKeys)
		assert.ElementsMatch(t, values[:oldest], cbValues)
		assert.Equal(t, expEvicted, int(evicted.Load()))
		assert.Equal(t, uint64(expLen), got.Len())
		assert.Equal(t, uint64(expSize), got.Size())
	})
}

func TestSafeCachePurge(t *testing.T) {
	t.Parallel()
	t.Run("Purge cache concurrently, eviction callback should be triggered for each pair only once", func(t *testing.T) {
		t.Parallel()
		workers := 10
		amount := 1000
		var expKeys []string
		var expValues []testValue
		cbKeys := make([]string, amount)
		cbValues := make([]testValue, amount)
		var evicted atomic.Int32

		cache := NewLRUCacheWithEviction[string, testValue](0, func(key string, value testValue) {
			evicted.Add(1)
			assert.NotContains(t, cbKeys, key)
			assert.NotContains(t, cbValues, value)
			cbKeys[value.value-1] = key
			cbValues[value.value-1] = value
		})

		got := NewSafeCache[string, testValue](cache)
		for i := 0; i < amount; i++ {
			key := fmt.Sprintf("key %d", i)
			value := testValue{i + 1}
			got.Add(key, value)
			expKeys = append(expKeys, key)
			expValues = append(expValues, value)
		}
		var complete sync.WaitGroup
		complete.Add(workers)
		for i := 0; i < workers; i++ {
			go func() {
				defer complete.Done()
				got.Purge()
			}()
		}
		complete.Wait()

		assert.Equal(t, amount, int(evicted.Load()))
		assert.ElementsMatch(t, expKeys, cbKeys)
		assert.ElementsMatch(t, expValues, cbValues)
		assert.Empty(t, got.Keys())
		assert.Empty(t, got.Values())
		assert.Zero(t, got.Len())
		assert.Zero(t, got.Size())
	})
}

func TestSafeCacheContains(t *testing.T) {
	t.Parallel()
	t.Run("True for existing key, concurrent calls", func(t *testing.T) {
		t.Parallel()
		workers := 100
		key := "some key"
		value := testValue{0}
		cache := NewLRUCache[string, testValue](0)
		sCache := NewSafeCache[string, testValue](cache)
		sCache.Add(key, value)

		var complete sync.WaitGroup
		complete.Add(workers)
		for i := 0; i < workers; i++ {
			go func() {
				defer complete.Done()
				got := sCache.Contains(key)
				assert.True(t, got)
			}()
		}
		complete.Wait()
	})
}

func TestSafeCacheKeys(t *testing.T) {
	t.Parallel()
	t.Run("Return keys from old to new, concurrent calls, should not affect recency", func(t *testing.T) {
		t.Parallel()
		workers := 100
		amount := 1000
		var expKeys []string
		cache := NewLRUCache[string, testValue](0)
		sCache := NewSafeCache[string, testValue](cache)
		for i := 0; i < amount; i++ {
			key := fmt.Sprintf("key %d", i)
			value := testValue{0}
			sCache.Add(key, value)
			expKeys = append(expKeys, key)
		}

		var complete sync.WaitGroup
		complete.Add(workers)
		for i := 0; i < workers; i++ {
			go func() {
				defer complete.Done()
				got := sCache.Keys()
				assert.Equal(t, expKeys, got)
			}()
		}
		complete.Wait()
	})
}

func TestSafeCacheValues(t *testing.T) {
	t.Parallel()
	t.Run("Return values from old to new, concurrent calls, should not affect recency", func(t *testing.T) {
		t.Parallel()
		workers := 100
		amount := 1000
		var expValues []testValue
		cache := NewLRUCache[string, testValue](0)
		sCache := NewSafeCache[string, testValue](cache)
		for i := 0; i < amount; i++ {
			key := fmt.Sprintf("key %d", i)
			value := testValue{i}
			sCache.Add(key, value)
			expValues = append(expValues, value)
		}

		var complete sync.WaitGroup
		complete.Add(workers)
		for i := 0; i < workers; i++ {
			go func() {
				defer complete.Done()
				got := sCache.Values()
				assert.Equal(t, expValues, got)
			}()
		}
		complete.Wait()
	})
}

func TestSafeCacheCapacity(t *testing.T) {
	t.Parallel()
	t.Run("Return cache capacicty, concurrent calls", func(t *testing.T) {
		t.Parallel()
		workers := 100
		expCap := rand.Intn(1000)
		cache := NewLRUCache[string, testValue](uint64(expCap))
		sCache := NewSafeCache[string, testValue](cache)

		var complete sync.WaitGroup
		complete.Add(workers)
		for i := 0; i < workers; i++ {
			go func() {
				defer complete.Done()
				got := sCache.Capacity()
				assert.Equal(t, uint64(expCap), got)
			}()
		}
		complete.Wait()
	})
}

func TestSafeCacheLen(t *testing.T) {
	t.Parallel()
	t.Run("Return cache lenght, concurrent calls", func(t *testing.T) {
		t.Parallel()
		workers := 100
		amount := rand.Intn(1000) + 10
		expLen := amount
		cache := NewLRUCache[string, testValue](0)
		sCache := NewSafeCache[string, testValue](cache)
		for i := 0; i < amount; i++ {
			key := fmt.Sprintf("key %d", i)
			value := testValue{i}
			sCache.Add(key, value)
		}

		var complete sync.WaitGroup
		complete.Add(workers)
		for i := 0; i < workers; i++ {
			go func() {
				defer complete.Done()
				got := sCache.Len()
				assert.Equal(t, uint64(expLen), got)
			}()
		}
		complete.Wait()
	})
}

func TestSafeCacheSize(t *testing.T) {
	t.Parallel()
	t.Run("Return cache size, concurrent calls", func(t *testing.T) {
		t.Parallel()
		workers := 100
		amount := rand.Intn(1000) + 10
		expSize := uint64(amount) * testValue{}.Size()
		cache := NewLRUCache[string, testValue](0)
		sCache := NewSafeCache[string, testValue](cache)
		for i := 0; i < amount; i++ {
			key := fmt.Sprintf("key %d", i)
			value := testValue{i}
			sCache.Add(key, value)
		}

		var complete sync.WaitGroup
		complete.Add(workers)
		for i := 0; i < workers; i++ {
			go func() {
				defer complete.Done()
				got := sCache.Size()
				assert.Equal(t, expSize, got)
			}()
		}
		complete.Wait()
	})
}

func TestSafeCacheResize(t *testing.T) {
	t.Parallel()
	t.Run("Concurrent calls to shrink capacity should remove old pars with eviction only once", func(t *testing.T) {
		t.Parallel()
		workers := 10
		initLen := 1000
		initCap := uint64(initLen) * testValue{}.Size()
		expLen := initLen / 2
		expCap := uint64(expLen) * testValue{}.Size()
		oldest := initLen - expLen
		var keys []string
		var values []testValue
		cbKeys := make([]string, oldest)
		cbValues := make([]testValue, oldest)
		var evicted atomic.Int32
		cache := NewLRUCacheWithEviction[string, testValue](initCap, func(key string, value testValue) {
			evicted.Add(1)
			assert.NotContains(t, cbKeys, key)
			assert.NotContains(t, cbValues, value)
			cbKeys[value.value-1] = key
			cbValues[value.value-1] = value
		})

		got := NewSafeCache[string, testValue](cache)
		for i := 0; i < initLen; i++ {
			key := fmt.Sprintf("key %d", i)
			value := testValue{i + 1}
			got.Add(key, value)
			keys = append(keys, key)
			values = append(values, value)
		}
		var complete sync.WaitGroup
		complete.Add(workers)
		for i := 0; i < workers; i++ {
			go func() {
				defer complete.Done()
				got.Resize(expCap)
			}()
		}
		complete.Wait()

		assert.Equal(t, expCap, got.Capacity())
		assert.Equal(t, uint64(expLen), got.Len())
		assert.Equal(t, oldest, int(evicted.Load()))
		assert.Equal(t, keys[:oldest], cbKeys)
		assert.Equal(t, values[:oldest], cbValues)
		assert.Equal(t, keys[oldest:], got.Keys())
		assert.Equal(t, values[oldest:], got.Values())
	})
}
