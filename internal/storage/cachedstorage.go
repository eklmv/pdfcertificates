package storage

import (
	"log/slog"
	"time"

	"github.com/eklmv/pdfcertificates/internal/cache"
)

type CachedStorage struct {
	Storage
	c cache.Cache[uint32, certFile]
}

type certFile struct {
	file      []byte
	timestamp time.Time
	size      uint64
}

func (cf certFile) Size() uint64 {
	return cf.size
}

func NewCachedStorage(storage Storage) *CachedStorage {
	c := cache.NewSafeCache(cache.NewLRUCache[uint32, certFile](0))
	return &CachedStorage{Storage: storage, c: c}
}

func (cs *CachedStorage) Add(id string, cert []byte, timestamp time.Time) error {
	err := cs.Storage.Add(id, cert, timestamp)
	if err == nil {
		hash := cache.HashString(id)
		cf := certFile{
			file:      cert,
			timestamp: timestamp,
		}
		cf.size = cache.SizeOf(cf)
		cs.c.Add(hash, cf)
	}
	return err
}

func (cs *CachedStorage) Get(id string, timestamp time.Time) (cert []byte, err error) {
	hash := cache.HashString(id)
	cf, ok := cs.c.Peek(hash)
	if ok && (cf.timestamp.Equal(timestamp) || cf.timestamp.After(timestamp)) {
		cs.c.Touch(hash)
		slog.Info("sucessful storage cache hit", slog.String("id", id), slog.Time("timestamp", timestamp))
		return cf.file, nil
	}
	cert, err = cs.Storage.Get(id, timestamp)
	if err == nil {
		cf := certFile{
			file:      cert,
			timestamp: timestamp,
		}
		cf.size = cache.SizeOf(cf)
		cs.c.Add(hash, cf)
	}
	return
}

func (cs *CachedStorage) Delete(id string, timestamp time.Time) {
	hash := cache.HashString(id)
	cf, ok := cs.c.Peek(hash)
	if ok && (cf.timestamp.Equal(timestamp) || cf.timestamp.Before(timestamp)) {
		cs.c.Remove(hash)
	}
	cs.Storage.Delete(id, timestamp)
}

func (cs *CachedStorage) Exists(id string, timestamp time.Time) bool {
	hash := cache.HashString(id)
	cf, ok := cs.c.Peek(hash)
	if ok && (cf.timestamp.Equal(timestamp) || cf.timestamp.After(timestamp)) {
		return true
	}
	return cs.Storage.Exists(id, timestamp)
}
