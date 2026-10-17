//internal/checker/checker.go

package checker

import (
	"context"
	"net/http"
	"time"
)

type Checker struct {
	client *http.Client
}

type Result struct {
	StatusCode     int
	ResponseTimeMs int64
	CheckedAt      time.Time
	Error          error
}

func NewChecker(client *http.Client) *Checker {
	return &Checker{
		client: client,
	}
}

func (c *Checker) Check(ctx context.Context, url string) (result Result) {
	startTime := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Result{
			StatusCode: 0,
			CheckedAt:  startTime,
			Error:      err,
		}
	}
	resp, err := c.client.Do(req)
	totalTime := (time.Since(startTime)).Milliseconds()
	if err != nil {
		return Result{
			StatusCode:     0,
			ResponseTimeMs: totalTime,
			CheckedAt:      startTime,
			Error:          err,
		}
	}
	defer resp.Body.Close()

	return Result{
		StatusCode:     resp.StatusCode,
		ResponseTimeMs: totalTime,
		CheckedAt:      startTime,
		Error:          nil,
	}
}
