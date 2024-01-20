package cache

import (
	"hash/fnv"
	"reflect"
	"unsafe"
)

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

func HashString(str string) uint32 {
	hasher := fnv.New32a()
	hasher.Write([]byte(str))
	return hasher.Sum32()
}

// TODO: add correct calculation for more complex types
func SizeOf(a any) uint64 {
	size := uint64(reflect.TypeOf(a).Size())
	switch reflect.TypeOf(a).Kind() {
	case reflect.Struct:
		s := reflect.ValueOf(a)
		if !s.CanAddr() {
			tmp := reflect.New(s.Type()).Elem()
			tmp.Set(s)
			s = tmp
		}
		for i := 0; i < s.NumField(); i++ {
			f := s.Field(i)
			if f.Kind() > reflect.Complex128 {
				size -= uint64(f.Type().Size())
				if f.CanInterface() {
					size += SizeOf(f.Interface())
				} else {
					uf := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
					size += SizeOf(uf.Interface())
				}
			}
		}
	case reflect.Slice:
		s := reflect.ValueOf(a)
		for i := 0; i < s.Len(); i++ {
			size += SizeOf(s.Index(i).Interface())
		}
	case reflect.String:
		size += uint64(reflect.ValueOf(a).Len())
	}
	return size
}
