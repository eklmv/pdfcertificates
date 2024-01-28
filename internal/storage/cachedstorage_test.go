package storage

import (
	"fmt"
	"testing"
	"time"

	"github.com/eklmv/pdfcertificates/internal/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCachedStorageImplementsInterface(t *testing.T) {
	assert.Implements(t, (*Storage)(nil), &CachedStorage{})
}

func TestCachedStorageAdd(t *testing.T) {
	t.Run("should cache certificate if underlying Add was successful", func(t *testing.T) {
		id := "00000000"
		cert := []byte("Hello, world!")
		timestamp := time.Now()
		m := NewMockStorage(t)
		cs := NewCachedStorage(m)

		m.EXPECT().Add(id, cert, timestamp).Return(nil).Once()

		err := cs.Add(id, cert, timestamp)

		require.NoError(t, err)
		assert.Equal(t, uint64(1), cs.c.Len())
		assert.Equal(t, cache.HashString(id), cs.c.Keys()[0])
		assert.Equal(t, cert, cs.c.Values()[0].file)
		m.AssertExpectations(t)
	})
	t.Run("should not affect cache if underlying Add failed", func(t *testing.T) {
		id := "00000000"
		cert := []byte("Hello, world!")
		timestamp := time.Now()
		m := NewMockStorage(t)
		cs := NewCachedStorage(m)

		m.EXPECT().Add(id, cert, timestamp).Return(fmt.Errorf("failed")).Once()

		err := cs.Add(id, cert, timestamp)

		require.Error(t, err)
		assert.Empty(t, cs.c.Len())
		assert.Empty(t, cs.c.Size())
		assert.Empty(t, cs.c.Keys())
		m.AssertExpectations(t)
	})
}

func TestCachedStorageGet(t *testing.T) {
	t.Run("return cached file if requested timestamp same or older, avoid underlying call", func(t *testing.T) {
		id := "00000000"
		exp := []byte("Hello, world!")
		timestamp := time.Now()
		older_timestamp := timestamp.Add(-1 * time.Hour)
		m := NewMockStorage(t)
		cs := NewCachedStorage(m)
		m.EXPECT().Add(id, exp, timestamp).Return(nil).Once()
		err := cs.Add(id, exp, timestamp)
		require.NoError(t, err)

		got, err := cs.Get(id, timestamp)

		require.NoError(t, err)
		assert.Equal(t, exp, got)

		got, err = cs.Get(id, older_timestamp)

		require.NoError(t, err)
		assert.Equal(t, exp, got)
		m.AssertExpectations(t)
	})
	t.Run("if requested timestamp is newer than stored make underlying call and cache result", func(t *testing.T) {
		id := "00000000"
		cert := []byte("Hello, world!")
		exp := []byte("Newer certificate")
		timestamp := time.Now()
		newer_timestamp := timestamp.Add(1 * time.Hour)
		m := NewMockStorage(t)
		cs := NewCachedStorage(m)
		m.EXPECT().Add(id, cert, timestamp).Return(nil).Once()
		err := cs.Add(id, cert, timestamp)
		require.NoError(t, err)

		m.EXPECT().Get(id, newer_timestamp).Return(exp, nil).Once()
		got, err := cs.Get(id, newer_timestamp)

		require.NoError(t, err)
		assert.Equal(t, exp, got)
		assert.Equal(t, uint64(1), cs.c.Len())
		assert.Equal(t, cache.HashString(id), cs.c.Keys()[0])
		assert.Equal(t, exp, cs.c.Values()[0].file)
		m.AssertExpectations(t)
	})
}

func TestCachedStorageDelete(t *testing.T) {
	t.Run("remove cached certificate and make underlying call", func(t *testing.T) {
		id := "00000000"
		cert := []byte("Hello, world!")
		timestamp := time.Now()
		m := NewMockStorage(t)
		cs := NewCachedStorage(m)
		m.EXPECT().Add(id, cert, timestamp).Return(nil).Once()
		err := cs.Add(id, cert, timestamp)
		require.NoError(t, err)

		m.EXPECT().Delete(id).Once()
		cs.Delete(id)

		assert.Empty(t, cs.c.Len())
		assert.Empty(t, cs.c.Size())
		assert.Empty(t, cs.c.Keys())
		assert.Empty(t, cs.c.Values())
		m.AssertExpectations(t)
	})
}

func TestCachedStorageExists(t *testing.T) {
	t.Run("return true if certificate cached with same or newer timestamp, avoid underlying call", func(t *testing.T) {
		id := "00000000"
		cert := []byte("Hello, world!")
		timestamp := time.Now()
		older_timestamp := timestamp.Add(-1 * time.Hour)
		m := NewMockStorage(t)
		cs := NewCachedStorage(m)
		m.EXPECT().Add(id, cert, timestamp).Return(nil).Once()
		err := cs.Add(id, cert, timestamp)
		require.NoError(t, err)

		got := cs.Exists(id, timestamp)
		assert.True(t, got)

		got = cs.Exists(id, older_timestamp)
		assert.True(t, got)

		m.AssertExpectations(t)
	})
	t.Run("if certificate cached with older timestamp make underlying call", func(t *testing.T) {
		id := "00000000"
		cert := []byte("Hello, world!")
		timestamp := time.Now()
		newer_timestamp := timestamp.Add(1 * time.Hour)
		m := NewMockStorage(t)
		cs := NewCachedStorage(m)
		m.EXPECT().Add(id, cert, timestamp).Return(nil).Once()
		err := cs.Add(id, cert, timestamp)
		require.NoError(t, err)

		m.EXPECT().Exists(id, newer_timestamp).Return(false).Once()
		got := cs.Exists(id, newer_timestamp)

		assert.False(t, got)
		m.AssertExpectations(t)
	})
	t.Run("if certificate not cached make underlying call", func(t *testing.T) {
		id := "00000000"
		timestamp := time.Now()
		m := NewMockStorage(t)
		cs := NewCachedStorage(m)

		m.EXPECT().Exists(id, timestamp).Return(true).Once()
		got := cs.Exists(id, timestamp)

		assert.True(t, got)
		m.AssertExpectations(t)
	})
}
