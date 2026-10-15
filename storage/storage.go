// internal/storage/storage.go

package storage

import (
	"errors"
	"time"
)

var (
	ErrURLNotFound = errors.New("url not found")
	ErrURLExists   = errors.New("url exists")
)

type Check struct {
	ID             int64
	URLID          int64
	HTTPStatus     int
	ResponseTimeMs int64
	CheckedAt      time.Time
}

type Storage interface {
	SaveURL(alias string, url string) (int64, error)
	SaveCheck(urlID int64, httpStatus int, responseTimeMs int64, checkedAt time.Time) (int64, error)
	GetChecks(urlID int64) ([]Check, error)
	GetURL(alias string) (URL, error)
	GetAllURLs() ([]URL, error)
	Close() error
}

type URL struct {
	ID    int64
	Alias string
	URL   string
}
