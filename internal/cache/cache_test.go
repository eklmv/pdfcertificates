package cache

import (
	"strconv"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

type _testAlias int32

type _testStruct1 struct {
	f float64
	b bool
	i int32
}

type _testStruct2 struct {
	f float64
	b bool
	s string
}

type _testStruct3 struct {
	f float64
	b bool
	s string
	t _testStruct2
}

func TestSizeOf(t *testing.T) {
	intSize := uint64(strconv.IntSize / 8)
	tCases := []struct {
		name string
		a    any
		exp  uint64
	}{
		{"bool", true, 1},
		{"int32", int32(0), 4},
		{"int64", int64(0), 8},
		{"uint32", uint32(0), 4},
		{"uint64", uint64(0), 8},
		{"float32", float32(0), 4},
		{"float64", float64(0), 8},
		// string header
		{"empty string", "", uint64(intSize * 2)},
		{"string", "12345", uint64(intSize*2 + 5)},
		// slice header
		{"empty slice of byte", []byte(""), uint64(intSize * 3)},
		{"slice of byte", []byte("12345"), uint64(intSize*3 + 5)},
		{"type alias", _testAlias(0), 4},
		{"struct with unexported fields", _testStruct1{0, false, 0}, uint64(unsafe.Sizeof(_testStruct1{}))},
		{"struct with unexported string field", _testStruct2{0, false, "12345"}, uint64(unsafe.Sizeof(_testStruct2{}) + 5)},
		{"struct with unexported struct field", _testStruct3{0, false, "12345", _testStruct2{0, false, "12345"}}, uint64(unsafe.Sizeof(_testStruct3{}) + 10)},
		{"slice of structs", []_testStruct1{{}, {}}, uint64(unsafe.Sizeof(_testStruct1{})*2) + uint64(intSize*3)},
	}
	for _, tc := range tCases {
		t.Run(tc.name, func(t *testing.T) {
			got := SizeOf(tc.a)

			assert.Equal(t, int(tc.exp), int(got))
		})
	}
}
