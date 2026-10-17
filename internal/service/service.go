//internal/service/service.go

package service

import (
	"context"
	"fmt"
	"time"
	"webaviable/internal/checker"
	"webaviable/storage"
)

type Service struct {
	storage Storage
	checker *checker.Checker
}

type Storage interface {
	SaveURL(alias string, url string) (int64, error)
	GetURL(alias string) (storage.URL, error)
	GetChecks(urlID int64) ([]storage.Check, error)
	SaveCheck(urlID int64, httpStatus int, responseTimeMs int64, checkedAt time.Time) (int64, error)
	GetAllURLs() ([]storage.URL, error)
	DeleteURL(alias string) error
}

func NewService(storage Storage, checker *checker.Checker) *Service {
	return &Service{
		storage: storage,
		checker: checker,
	}
}

func (s *Service) AddURL(alias string, url string) (int64, error) {
	return s.storage.SaveURL(alias, url)
}

func (s *Service) DeleteURL(alias string) error {
	return s.storage.DeleteURL(alias)
}

func (s *Service) GetChecks(alias string) ([]storage.Check, error) {
	url, err := s.storage.GetURL(alias)
	if err != nil {
		return nil, err
	}
	return s.storage.GetChecks(url.ID)
}

func (s *Service) GetURL(alias string) (storage.URL, error) {
	return s.storage.GetURL(alias)
}

func (s *Service) CheckURL(alias string, ctx context.Context) (checker.Result, error) {
	url, err := s.storage.GetURL(alias)
	if err != nil {
		return checker.Result{}, err
	}
	result := s.checker.Check(ctx, url.URL)
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if _, err := s.storage.SaveCheck(url.ID, result.StatusCode, result.ResponseTimeMs, result.CheckedAt); err != nil {
		return checker.Result{}, fmt.Errorf("save check for %q: %w", alias, err)
	}

	return result, nil
}

func (s *Service) GetAllURLs() ([]storage.URL, error) {
	return s.storage.GetAllURLs()
}
