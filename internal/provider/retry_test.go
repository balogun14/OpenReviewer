package provider

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/openreview-ai/openreview/internal/finding"
)

func TestRetryProviderRetriesRetryableErrors(t *testing.T) {
	attempts := 0
	retryProvider := WithRetry(reviewFunc(func(ctx context.Context, req Request) ([]finding.Finding, error) {
		attempts++
		if attempts == 1 {
			return nil, RetryableError{Err: errors.New("temporary")}
		}
		return []finding.Finding{{Title: "ok"}}, nil
	}), RetryConfig{
		MaxAttempts: 2,
		Delay:       time.Nanosecond,
	})

	findings, err := retryProvider.Review(context.Background(), Request{})
	if err != nil {
		t.Fatalf("Review returned error: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("expected two attempts, got %d", attempts)
	}
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %d", len(findings))
	}
}

func TestRetryProviderDoesNotRetryPermanentErrors(t *testing.T) {
	attempts := 0
	retryProvider := WithRetry(reviewFunc(func(ctx context.Context, req Request) ([]finding.Finding, error) {
		attempts++
		return nil, errors.New("permanent")
	}), RetryConfig{
		MaxAttempts: 3,
		Delay:       time.Nanosecond,
	})

	_, err := retryProvider.Review(context.Background(), Request{})
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != 1 {
		t.Fatalf("expected one attempt, got %d", attempts)
	}
}

type reviewFunc func(context.Context, Request) ([]finding.Finding, error)

func (f reviewFunc) Review(ctx context.Context, req Request) ([]finding.Finding, error) {
	return f(ctx, req)
}
