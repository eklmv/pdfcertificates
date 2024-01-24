package storage

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSystemImplementsInterface(t *testing.T) {
	assert.Implements(t, (*Storage)(nil), &FileSystem{})
}

func testDir(tb testing.TB) string {
	tb.Helper()
	out := "../../out/test/"
	path, err := filepath.Abs(out + tb.Name() + "_" + time.Now().Format("060102_150405"))
	if err != nil {
		slog.Error("failed to resolve absolute path to test directory", slog.String("path", path), slog.Any("error", err))
		tb.FailNow()
	}
	return path
}

func TestFileSystemAdd(t *testing.T) {
	t.Run("store certificate as pdf file on disk", func(t *testing.T) {
		path := testDir(t)
		cert := []byte("Hello, world!")
		id := "00000000"
		timestamp := time.Now()
		fs, err := NewFileSystem(path)
		require.NoError(t, err)

		err = fs.Add(id, cert, timestamp)
		require.NoError(t, err)
		assert.FileExists(t, fsEnsureTrailingSlash(path)+toFileName(id, timestamp))
	})
	t.Run("certificate with same id but older timestamp should not be stored", func(t *testing.T) {
		path := testDir(t)
		cert := []byte("Hello, world!")
		id := "00000000"
		timestamp := time.Now()
		older_timestamp := timestamp.Add(-1 * time.Hour)
		fs, err := NewFileSystem(path)
		require.NoError(t, err)

		err = fs.Add(id, cert, timestamp)
		require.NoError(t, err)

		err = fs.Add(id, cert, older_timestamp)
		require.NoError(t, err)

		assert.FileExists(t, fsEnsureTrailingSlash(path)+toFileName(id, timestamp))
		assert.NoFileExists(t, fsEnsureTrailingSlash(path)+toFileName(id, older_timestamp))
	})
	t.Run("certificate with same id but newer timestamp should be stored and old certificate removed", func(t *testing.T) {
		path := testDir(t)
		cert := []byte("Hello, world!")
		id := "00000000"
		timestamp := time.Now()
		newer_timestamp := timestamp.Add(1 * time.Hour)
		fs, err := NewFileSystem(path)
		require.NoError(t, err)

		err = fs.Add(id, cert, timestamp)
		require.NoError(t, err)

		err = fs.Add(id, cert, newer_timestamp)
		require.NoError(t, err)

		assert.NoFileExists(t, fsEnsureTrailingSlash(path)+toFileName(id, timestamp))
		assert.FileExists(t, fsEnsureTrailingSlash(path)+toFileName(id, newer_timestamp))
	})
}

func TestFileSystemGet(t *testing.T) {
	t.Run("return stored certificate", func(t *testing.T) {
		path := testDir(t)
		exp := []byte("Hello, world!")
		id := "00000000"
		timestamp := time.Now()
		fs, err := NewFileSystem(path)
		require.NoError(t, err)
		err = fs.Add(id, exp, timestamp)
		require.NoError(t, err)

		got, err := fs.Get(id, timestamp)

		require.NoError(t, err)
		assert.Equal(t, exp, got)
	})
	t.Run("return stored certificate with newer timestamp if older timestamp was requested", func(t *testing.T) {
		path := testDir(t)
		exp := []byte("Hello, world!")
		id := "00000000"
		timestamp := time.Now()
		older_timestamp := timestamp.Add(-1 * time.Hour)
		fs, err := NewFileSystem(path)
		require.NoError(t, err)
		err = fs.Add(id, exp, timestamp)
		require.NoError(t, err)

		got, err := fs.Get(id, older_timestamp)

		require.NoError(t, err)
		assert.Equal(t, exp, got)
	})
	t.Run("return not found error if requested timestamp is newer than stored timestamp", func(t *testing.T) {
		path := testDir(t)
		cert := []byte("Hello, world!")
		id := "00000000"
		timestamp := time.Now()
		newer_timestamp := timestamp.Add(1 * time.Hour)
		fs, err := NewFileSystem(path)
		require.NoError(t, err)
		err = fs.Add(id, cert, timestamp)
		require.NoError(t, err)

		got, err := fs.Get(id, newer_timestamp)

		require.Error(t, err)
		assert.ErrorIs(t, err, CertificateFileNotFoundError)
		assert.Empty(t, got)
	})
}

func TestFileSystemDelete(t *testing.T) {
	t.Run("delete stored certificate file", func(t *testing.T) {
		path := testDir(t)
		cert := []byte("Hello, world!")
		id := "00000000"
		timestamp := time.Now()
		fs, err := NewFileSystem(path)
		require.NoError(t, err)
		err = fs.Add(id, cert, timestamp)
		require.NoError(t, err)
		require.FileExists(t, fsEnsureTrailingSlash(path)+toFileName(id, timestamp))

		fs.Delete(id, timestamp)

		require.NoError(t, err)
		assert.NoFileExists(t, fsEnsureTrailingSlash(path)+toFileName(id, timestamp))
	})
	t.Run("delete stored certificate file if requested timestamp is newer", func(t *testing.T) {
		path := testDir(t)
		cert := []byte("Hello, world!")
		id := "00000000"
		timestamp := time.Now()
		newer_timestamp := timestamp.Add(1 * time.Hour)
		fs, err := NewFileSystem(path)
		require.NoError(t, err)
		err = fs.Add(id, cert, timestamp)
		require.NoError(t, err)
		require.FileExists(t, fsEnsureTrailingSlash(path)+toFileName(id, timestamp))

		fs.Delete(id, newer_timestamp)

		require.NoError(t, err)
		assert.NoFileExists(t, fsEnsureTrailingSlash(path)+toFileName(id, timestamp))
	})
	t.Run("should not delete stored certificate if requested timestamp is older", func(t *testing.T) {
		path := testDir(t)
		cert := []byte("Hello, world!")
		id := "00000000"
		timestamp := time.Now()
		older_timestamp := timestamp.Add(-1 * time.Hour)
		fs, err := NewFileSystem(path)
		require.NoError(t, err)
		err = fs.Add(id, cert, timestamp)
		require.NoError(t, err)
		require.FileExists(t, fsEnsureTrailingSlash(path)+toFileName(id, timestamp))

		fs.Delete(id, older_timestamp)

		require.NoError(t, err)
		assert.FileExists(t, fsEnsureTrailingSlash(path)+toFileName(id, timestamp))
	})
}

func TestFileSystemExists(t *testing.T) {
	t.Run("return true if requested certificate exists with same or newer timestamp", func(t *testing.T) {
		path := testDir(t)
		cert := []byte("Hello, world!")
		id := "00000000"
		timestamp := time.Now()
		older_timestamp := timestamp.Add(-1 * time.Hour)
		fs, err := NewFileSystem(path)
		require.NoError(t, err)
		err = fs.Add(id, cert, timestamp)
		require.NoError(t, err)

		got := fs.Exists(id, timestamp)
		assert.True(t, got)

		got = fs.Exists(id, older_timestamp)
		assert.True(t, got)
	})
	t.Run("return false if requested certificate doesn't exists", func(t *testing.T) {
		path := testDir(t)
		fs, err := NewFileSystem(path)
		require.NoError(t, err)

		got := fs.Exists("00000000", time.Now())

		assert.False(t, got)
	})
	t.Run("return false if requested certificate exists with older timestamp", func(t *testing.T) {
		path := testDir(t)
		cert := []byte("Hello, world!")
		id := "00000000"
		timestamp := time.Now()
		newer_timestamp := timestamp.Add(1 * time.Hour)
		fs, err := NewFileSystem(path)
		require.NoError(t, err)
		err = fs.Add(id, cert, timestamp)
		require.NoError(t, err)

		got := fs.Exists(id, timestamp)
		assert.True(t, got)

		got = fs.Exists(id, newer_timestamp)
		assert.False(t, got)
	})
}

func TestFileSystemLoad(t *testing.T) {
	t.Run("load all certificate files from directory", func(t *testing.T) {
		path := testDir(t)
		amount := 10
		var ids []string
		var timestamps []time.Time
		var expCerts [][]byte
		writeFs, err := NewFileSystem(path)
		require.NoError(t, err)
		for i := 0; i < amount; i++ {
			id := strconv.Itoa(i)
			cert := []byte(fmt.Sprintf("certificate: %s", id))
			timestamp := time.Now()
			ids = append(ids, id)
			expCerts = append(expCerts, cert)
			timestamps = append(timestamps, timestamp)
			err = writeFs.Add(id, cert, timestamp)
			require.NoError(t, err)
		}

		fs, err := NewFileSystem(path)
		require.NoError(t, err)

		err = fs.Load()
		assert.NoError(t, err)

		var gotCerts [][]byte
		for i, id := range ids {
			got, err := fs.Get(id, timestamps[i])
			assert.NoError(t, err)
			gotCerts = append(gotCerts, got)
		}
		assert.Equal(t, expCerts, gotCerts)
	})
}
