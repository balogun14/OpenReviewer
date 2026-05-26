package provider

import (
	"context"
	"errors"
	"time"

	"github.com/openreview-ai/openreview/internal/finding"
)

type Reviewer interface {
	Review(ctx context.Context, req Request) ([]finding.Finding, error)
}

type RetryConfig struct {
	MaxAttempts int
	Delay       time.Duration
}

type RetryProvider struct {
	next   Reviewer
	config RetryConfig
	sleep  func(context.Context, time.Duration) error
}

func WithRetry(next Reviewer, config RetryConfig) RetryProvider {
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = 1
	}
	if config.Delay <= 0 {
		config.Delay = 250 * time.Millisecond
	}

	return RetryProvider{
		next:   next,
		config: config,
		sleep: func(ctx context.Context, delay time.Duration) error {
			timer := time.NewTimer(delay)
			defer timer.Stop()

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				return nil
			}
		},
	}
}

func (p RetryProvider) Review(ctx context.Context, req Request) ([]finding.Finding, error) {
	var lastErr error

	for attempt := 1; attempt <= p.config.MaxAttempts; attempt++ {
		findings, err := p.next.Review(ctx, req)
		if err == nil {
			return findings, nil
		}

		lastErr = err
		if !isRetryable(err) || attempt == p.config.MaxAttempts {
			break
		}

		if sleepErr := p.sleep(ctx, p.config.Delay); sleepErr != nil {
			return nil, sleepErr
		}
	}

	return nil, lastErr
}

type RetryableError struct {
	Err error
}

func (e RetryableError) Error() string {
	return e.Err.Error()
}

func (e RetryableError) Unwrap() error {
	return e.Err
}

func isRetryable(err error) bool {
	var retryable RetryableError
	return errors.As(err, &retryable)
}
