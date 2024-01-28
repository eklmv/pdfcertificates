package storage

import (
	"errors"
	"time"
)

type Storage interface {
	Add(id string, cert []byte, timestamp time.Time) error
	Get(id string, timestamp time.Time) (cert []byte, err error)
	Delete(id string)
	Exists(id string, timestamp time.Time) bool
	Load() error
}

var CertificateFileNotFoundError = errors.New("certificate file not found")
