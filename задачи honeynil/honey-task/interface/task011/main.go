package main

// Задача: Retry — повторное выполнение с разными стратегиями задержки.

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

type RetryStrategy interface {
	NextDelay(attempt int) time.Duration
}

type Retry interface {
	Do(ctx context.Context, fn func() error) error
	DoWithData(ctx context.Context, fn func() (interface{}, error)) (interface{}, error)
}

// --- ConstantBackoff ---

type ConstantBackoff struct {
	Delay time.Duration
}

func (c *ConstantBackoff) NextDelay(attempt int) time.Duration { return c.Delay }

// --- ExponentialBackoff ---

type ExponentialBackoff struct {
	Initial time.Duration
	Max     time.Duration
}

func (e *ExponentialBackoff) NextDelay(attempt int) time.Duration {
	d := time.Duration(float64(e.Initial) * math.Pow(2, float64(attempt)))
	if e.Max > 0 && d > e.Max {
		return e.Max
	}
	return d
}

// --- RetryExecutor ---

type RetryExecutor struct {
	maxAttempts int
	strategy    RetryStrategy
}

func NewRetryExecutor(maxAttempts int, strategy RetryStrategy) *RetryExecutor {
	return &RetryExecutor{maxAttempts: maxAttempts, strategy: strategy}
}

func (r *RetryExecutor) Do(ctx context.Context, fn func() error) error {
	_, err := r.DoWithData(ctx, func() (interface{}, error) { return nil, fn() })
	return err
}

func (r *RetryExecutor) DoWithData(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	var lastErr error
	for attempt := 0; attempt < r.maxAttempts; attempt++ {
		if attempt > 0 {
			delay := r.strategy.NextDelay(attempt - 1)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}
		result, err := fn()
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func main() {
	attempts := 0
	fn := func() error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary error")
		}
		return nil
	}

	executor := NewRetryExecutor(5, &ConstantBackoff{Delay: 10 * time.Millisecond})
	err := executor.Do(context.Background(), fn)
	fmt.Printf("done in %d attempts, err=%v\n", attempts, err) // 3 attempts, err=nil

	// Exponential backoff
	executor2 := NewRetryExecutor(4, &ExponentialBackoff{Initial: 10 * time.Millisecond, Max: 100 * time.Millisecond})
	for i := 0; i < 4; i++ {
		fmt.Printf("delay[%d]: %v\n", i, executor2.strategy.NextDelay(i))
	}
}
