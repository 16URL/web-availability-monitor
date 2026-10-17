//internal/scheduler/scheduler.go

package scheduler

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
	"webaviable/internal/service"
)

type Scheduler struct {
	interval time.Duration
	service  *service.Service
	slog     *slog.Logger
	worker   int
}

func NewScheduler(svc *service.Service, interval time.Duration, log *slog.Logger, workers int) *Scheduler {
	if workers <= 0 {
		workers = 1
	}
	return &Scheduler{service: svc, interval: interval, slog: log, worker: workers}
}

func (s *Scheduler) Start(ctx context.Context) {
	s.checkAll(ctx)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkAll(ctx)
		}
	}
}

func (s *Scheduler) checkAll(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	var wg sync.WaitGroup

	items, err := s.service.GetAllURLs()
	if err != nil {
		s.slog.Error("failed to get URLs", slog.Any("error", err))
		return
	}

	jobs := make(chan string, 10)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(jobs)
		for _, u := range items {
			select {
			case <-ctx.Done():
				return
			case jobs <- u.Alias:
				continue
			}
		}
	}()

	for i := 0; i < s.worker; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case alias, ok := <-jobs:
					if !ok {
						return
					}
					_, err := s.service.CheckURL(alias, ctx)
					if err == nil {
						continue
					}
					if ctx.Err() != nil {
						return
					}
					if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
						continue
					}
					s.slog.Error("failed to check URL", slog.Any("error", err), slog.String("alias", alias))
				}
			}
		}()
	}
	wg.Wait()
}
