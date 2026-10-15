package service

import "webaviable/storage"

type Service struct {
	storage Storage
}

type Storage interface {
	SaveURL(alias string, url string) (int64, error)
	GetURL(alias string) (storage.URL, error)
	GetChecks(urlID int64) ([]storage.Check, error)
}

func NewService(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}
