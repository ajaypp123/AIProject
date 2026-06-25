package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/spark-debug-mcp/spark-debug-mcp/pkg/retry"
)

func TestDoSuccessFirstAttempt(t *testing.T) {
	calls := 0
	err := retry.Do(context.Background(), 3, 10*time.Millisecond, func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestDoSuccessAfterRetry(t *testing.T) {
	calls := 0
	err := retry.Do(context.Background(), 3, 10*time.Millisecond, func() error {
		calls++
		if calls < 2 {
			return errors.New("transient")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

func TestDoExhausted(t *testing.T) {
	err := retry.Do(context.Background(), 2, 10*time.Millisecond, func() error {
		return errors.New("persistent")
	})
	if err == nil {
		t.Error("expected error after exhausted retries")
	}
}

func TestDoContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := retry.Do(ctx, 3, 100*time.Millisecond, func() error {
		return errors.New("fail")
	})
	if err == nil {
		t.Error("expected context error")
	}
}
