package storage

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/eklmv/pdfcertificates/internal/cache"
)

type FileSystem struct {
	c    cache.Cache[uint32, fsCertLink]
	path string
}

type fsCertLink struct {
	link      string
	timestamp time.Time
	size      uint64
}

func (cu fsCertLink) Size() uint64 {
	return cu.size
}

func NewFileSystem(absPath string) (*FileSystem, error) {
	absPath = fsEnsureTrailingSlash(absPath)
	_, err := os.Stat(absPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.Error("failed to initialize new file system storage",
			slog.String("absPath", absPath), slog.Any("error", err))
		return nil, err
	}
	if errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(absPath, 0777)
	}
	if err != nil {
		slog.Error("failed to initialize new file system storage",
			slog.String("absPath", absPath), slog.Any("error", err))
		return nil, err
	}
	fs := &FileSystem{path: absPath}
	fs.initCache()
	return fs, nil
}

func fsEnsureTrailingSlash(path string) string {
	r := regexp.MustCompile("/$")
	if r.MatchString(path) {
		return path
	}
	return path + "/"
}

func (fs *FileSystem) initCache() {
	fs.c = cache.NewSafeCache(cache.NewLRUCacheWithEviction[uint32, fsCertLink](0, fsOnEviction))
}

// TODO: replace naive single attempt with retry system
func fsOnEviction(_ uint32, value fsCertLink) {
	_ = os.Remove(value.link)
}

func toFileName(id string, timestamp time.Time) string {
	extension := ".pdf"
	nsec := timestamp.UnixNano()
	return id + "_" + strconv.FormatInt(nsec, 10) + extension
}

func fromFileName(fileName string) (id string, timestamp time.Time, err error) {
	r := regexp.MustCompile("[_.]")
	parts := r.Split(fileName, -1)
	if len(parts) != 3 {
		err = fmt.Errorf("failed to split file name: %s, %v", fileName, parts)
		slog.Error("failed to convert certificate file name to id, timestamp tuple",
			slog.String("fileName", fileName), slog.Any("error", err))
		return
	}
	id = parts[0]
	nsec, err := strconv.ParseInt(parts[1], 10, 0)
	if err != nil {
		slog.Error("failed to convert certificate file name to id, timestamp tuple",
			slog.String("fileName", fileName), slog.Any("error", err))
		return
	}
	timestamp = time.Unix(0, nsec)
	return
}

func (fs *FileSystem) Add(id string, cert []byte, timestamp time.Time) error {
	hash := cache.HashString(id)
	cl, ok := fs.c.Peek(hash)
	if ok && (cl.timestamp.Equal(timestamp) || cl.timestamp.After(timestamp)) {
		slog.Info("same or newer certificate already stored", slog.String("id", id),
			slog.Time("requested timestamp", timestamp), slog.Time("stored timestamp", cl.timestamp))
		return nil
	}
	link := fs.path + toFileName(id, timestamp)
	err := os.WriteFile(link, cert, 0666)
	if err != nil {
		slog.Info("failed to store certificate file", slog.String("id", id), slog.Any("cert", cert),
			slog.Time("timestamp", timestamp), slog.Any("error", err))
		return err
	}
	cl = fsCertLink{
		link:      link,
		timestamp: timestamp,
	}
	cl.size = cache.SizeOf(cl)
	fs.c.Add(hash, cl)
	return err
}

func (fs *FileSystem) Get(id string, timestamp time.Time) (cert []byte, err error) {
	hash := cache.HashString(id)
	cl, ok := fs.c.Peek(hash)
	if ok && (cl.timestamp.Equal(timestamp) || cl.timestamp.After(timestamp)) {
		cert, err := os.ReadFile(cl.link)
		if err != nil {
			slog.Error("failed to read certificate file", slog.String("id", id),
				slog.Time("requested timestamp", timestamp), slog.Time("stored timestamp", cl.timestamp),
				slog.String("link", cl.link), slog.Any("error", err))
			return nil, err
		}
		fs.c.Touch(hash)
		return cert, nil
	}
	err = CertificateFileNotFoundError
	slog.Error("requested certificate not found", slog.String("id", id),
		slog.Time("timestamp", timestamp), slog.Any("error", err))
	return nil, err
}

func (fs *FileSystem) Delete(id string, timestamp time.Time) {
	hash := cache.HashString(id)
	cl, ok := fs.c.Peek(hash)
	if ok && (cl.timestamp.Equal(timestamp) || cl.timestamp.Before(timestamp)) {
		fs.c.Remove(hash)
	}
}

func (fs *FileSystem) Exists(id string, timestamp time.Time) bool {
	hash := cache.HashString(id)
	cl, ok := fs.c.Peek(hash)
	if ok && (cl.timestamp.Equal(timestamp) || cl.timestamp.After(timestamp)) {
		return true
	}
	return false
}

func (fs *FileSystem) Load() error {
	list, err := os.ReadDir(fs.path)
	if err != nil {
		slog.Error("failed to load file system storage", slog.String("path", fs.path),
			slog.Any("error", err))
		return err
	}
	for _, entry := range list {
		if entry.Type().IsRegular() {
			id, timestamp, err := fromFileName(entry.Name())
			if err != nil {
				slog.Error("failed to load file system storage", slog.String("path", fs.path),
					slog.Any("error", err))
				return err
			}
			info, err := entry.Info()
			if err != nil {
				slog.Error("failed to load file system storage", slog.String("path", fs.path),
					slog.Any("error", err))
				return err
			}
			size := uint64(info.Size())
			hash := cache.HashString(id)
			fs.c.Add(hash, fsCertLink{
				link:      fs.path + entry.Name(),
				timestamp: timestamp,
				size:      size,
			})
		}
	}
	return nil
}
