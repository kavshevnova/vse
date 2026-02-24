package main

// Задача: Distributed Lock — InMemoryLock, RetryableLock, AutoRefreshLock.

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Lock interface {
	Acquire(ctx context.Context, resource string, ttl time.Duration) (bool, error)
	Release(ctx context.Context, resource string) error
	Refresh(ctx context.Context, resource string, ttl time.Duration) error
	IsLocked(ctx context.Context, resource string) (bool, error)
}

type DistributedLock interface {
	Lock
	AcquireWithRetry(ctx context.Context, resource string, ttl time.Duration, retries int) error
	WithLock(ctx context.Context, resource string, ttl time.Duration, fn func() error) error
}

// --- InMemoryLock ---

type lockEntry struct {
	expiresAt time.Time
}

type InMemoryLock struct {
	mu    sync.Mutex
	locks map[string]lockEntry
}

func NewInMemoryLock() *InMemoryLock {
	return &InMemoryLock{locks: make(map[string]lockEntry)}
}

func (l *InMemoryLock) Acquire(_ context.Context, resource string, ttl time.Duration) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if e, ok := l.locks[resource]; ok && time.Now().Before(e.expiresAt) {
		return false, nil
	}
	l.locks[resource] = lockEntry{expiresAt: time.Now().Add(ttl)}
	return true, nil
}

func (l *InMemoryLock) Release(_ context.Context, resource string) error {
	l.mu.Lock()
	delete(l.locks, resource)
	l.mu.Unlock()
	return nil
}

func (l *InMemoryLock) Refresh(_ context.Context, resource string, ttl time.Duration) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.locks[resource]; !ok {
		return fmt.Errorf("lock %q not held", resource)
	}
	l.locks[resource] = lockEntry{expiresAt: time.Now().Add(ttl)}
	return nil
}

func (l *InMemoryLock) IsLocked(_ context.Context, resource string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.locks[resource]
	return ok && time.Now().Before(e.expiresAt), nil
}

// --- RetryableLock ---

type RetryableLock struct {
	inner    Lock
	interval time.Duration
}

func NewRetryableLock(inner Lock, retryInterval time.Duration) *RetryableLock {
	return &RetryableLock{inner: inner, interval: retryInterval}
}

func (r *RetryableLock) Acquire(ctx context.Context, resource string, ttl time.Duration) (bool, error) {
	return r.inner.Acquire(ctx, resource, ttl)
}
func (r *RetryableLock) Release(ctx context.Context, resource string) error {
	return r.inner.Release(ctx, resource)
}
func (r *RetryableLock) Refresh(ctx context.Context, resource string, ttl time.Duration) error {
	return r.inner.Refresh(ctx, resource, ttl)
}
func (r *RetryableLock) IsLocked(ctx context.Context, resource string) (bool, error) {
	return r.inner.IsLocked(ctx, resource)
}

func (r *RetryableLock) AcquireWithRetry(ctx context.Context, resource string, ttl time.Duration, retries int) error {
	for i := 0; i <= retries; i++ {
		ok, err := r.inner.Acquire(ctx, resource, ttl)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(r.interval):
		}
	}
	return fmt.Errorf("failed to acquire lock on %q after %d retries", resource, retries)
}

func (r *RetryableLock) WithLock(ctx context.Context, resource string, ttl time.Duration, fn func() error) error {
	if err := r.AcquireWithRetry(ctx, resource, ttl, 10); err != nil {
		return err
	}
	defer r.inner.Release(ctx, resource)
	return fn()
}

// --- AutoRefreshLock ---

type AutoRefreshLock struct {
	DistributedLock
	inner Lock
}

func NewAutoRefreshLock(inner Lock, interval time.Duration) *AutoRefreshLock {
	retryable := NewRetryableLock(inner, 50*time.Millisecond)
	return &AutoRefreshLock{DistributedLock: retryable, inner: inner}
}

func (a *AutoRefreshLock) AcquireAutoRefresh(ctx context.Context, resource string, ttl time.Duration) (context.CancelFunc, error) {
	ok, err := a.inner.Acquire(ctx, resource, ttl)
	if err != nil || !ok {
		return nil, fmt.Errorf("could not acquire lock: %w", err)
	}
	refreshCtx, cancel := context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(ttl / 2)
		defer ticker.Stop()
		for {
			select {
			case <-refreshCtx.Done():
				a.inner.Release(context.Background(), resource)
				return
			case <-ticker.C:
				a.inner.Refresh(refreshCtx, resource, ttl)
			}
		}
	}()
	return cancel, nil
}

func main() {
	base := NewInMemoryLock()
	retryable := NewRetryableLock(base, 10*time.Millisecond)

	ctx := context.Background()
	err := retryable.WithLock(ctx, "my-resource", 1*time.Second, func() error {
		fmt.Println("critical section executing")
		return nil
	})
	fmt.Println("WithLock err:", err)

	ok, _ := base.Acquire(ctx, "resource-2", 1*time.Second)
	fmt.Println("first acquire:", ok)
	ok, _ = base.Acquire(ctx, "resource-2", 1*time.Second)
	fmt.Println("second acquire (should be false):", ok)
}
