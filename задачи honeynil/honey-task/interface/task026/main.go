package main

// Задача: Distributed Rate Limiter — Token Bucket, Sliding Window, иерархические квоты.

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type RateLimitAlgorithm string

const (
	TokenBucket   RateLimitAlgorithm = "token_bucket"
	LeakyBucket   RateLimitAlgorithm = "leaky_bucket"
	FixedWindow   RateLimitAlgorithm = "fixed_window"
	SlidingWindow RateLimitAlgorithm = "sliding_window"
)

type RateLimit struct {
	Rate      int
	Period    time.Duration
	Burst     int
	Algorithm RateLimitAlgorithm
}

type QuotaLevel string

const (
	QuotaGlobal  QuotaLevel = "global"
	QuotaPerUser QuotaLevel = "per_user"
	QuotaPerIP   QuotaLevel = "per_ip"
	QuotaPerKey  QuotaLevel = "per_key"
)

type Request struct {
	UserID    string
	IP        string
	APIKey    string
	Resource  string
	Timestamp time.Time
}

type RateLimitResult struct {
	Allowed      bool
	Remaining    int
	ResetAt      time.Time
	RetryAfter   time.Duration
	CurrentUsage int64
	QuotaLimit   int64
}

type DistributedRateLimiter interface {
	Allow(ctx context.Context, req Request) (RateLimitResult, error)
	Reserve(ctx context.Context, req Request, count int) (RateLimitResult, error)
	Wait(ctx context.Context, req Request) error
	GetUsage(ctx context.Context, level QuotaLevel, identifier string) (int64, error)
	ResetQuota(ctx context.Context, level QuotaLevel, identifier string) error
}

type QuotaManager interface {
	SetLimit(ctx context.Context, level QuotaLevel, identifier string, limit RateLimit) error
	GetLimit(ctx context.Context, level QuotaLevel, identifier string) (*RateLimit, error)
	DeleteLimit(ctx context.Context, level QuotaLevel, identifier string) error
	ListLimits(ctx context.Context) (map[string]RateLimit, error)
}

// --- TokenBucketLimiter ---

type tokenBucket struct {
	mu        sync.Mutex
	tokens    float64
	maxTokens float64
	refillRate float64 // tokens per nanosecond
	lastRefill time.Time
}

func newTokenBucket(limit RateLimit) *tokenBucket {
	burst := limit.Burst
	if burst < limit.Rate {
		burst = limit.Rate
	}
	return &tokenBucket{
		tokens:     float64(burst),
		maxTokens:  float64(burst),
		refillRate: float64(limit.Rate) / float64(limit.Period),
		lastRefill: time.Now(),
	}
}

func (b *tokenBucket) take(n float64) (bool, float64, time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(b.lastRefill)
	b.tokens = min(b.maxTokens, b.tokens+float64(elapsed)*b.refillRate)
	b.lastRefill = now
	if b.tokens >= n {
		b.tokens -= n
		return true, b.tokens, 0
	}
	need := n - b.tokens
	retryAfter := time.Duration(need / b.refillRate)
	return false, b.tokens, retryAfter
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

type TokenBucketLimiter struct {
	mu      sync.RWMutex
	buckets map[string]*tokenBucket
	limits  map[string]RateLimit
}

func NewTokenBucketLimiter() *TokenBucketLimiter {
	return &TokenBucketLimiter{
		buckets: make(map[string]*tokenBucket),
		limits:  make(map[string]RateLimit),
	}
}

func (l *TokenBucketLimiter) keyFor(req Request) string {
	return fmt.Sprintf("%s:%s:%s", req.UserID, req.IP, req.Resource)
}

func (l *TokenBucketLimiter) bucketFor(key string) *tokenBucket {
	l.mu.Lock()
	defer l.mu.Unlock()
	if b, ok := l.buckets[key]; ok {
		return b
	}
	limit, ok := l.limits[key]
	if !ok {
		limit = RateLimit{Rate: 100, Period: time.Second, Burst: 200}
	}
	b := newTokenBucket(limit)
	l.buckets[key] = b
	return b
}

func (l *TokenBucketLimiter) SetLimit(_ context.Context, level QuotaLevel, id string, limit RateLimit) error {
	key := fmt.Sprintf("%s:%s", level, id)
	l.mu.Lock()
	l.limits[key] = limit
	delete(l.buckets, key) // reset bucket
	l.mu.Unlock()
	return nil
}

func (l *TokenBucketLimiter) Allow(_ context.Context, req Request) (RateLimitResult, error) {
	key := l.keyFor(req)
	b := l.bucketFor(key)
	allowed, remaining, retryAfter := b.take(1)
	return RateLimitResult{
		Allowed:    allowed,
		Remaining:  int(remaining),
		ResetAt:    time.Now().Add(retryAfter),
		RetryAfter: retryAfter,
	}, nil
}

func (l *TokenBucketLimiter) Reserve(_ context.Context, req Request, count int) (RateLimitResult, error) {
	key := l.keyFor(req)
	b := l.bucketFor(key)
	allowed, remaining, retryAfter := b.take(float64(count))
	return RateLimitResult{
		Allowed:    allowed,
		Remaining:  int(remaining),
		RetryAfter: retryAfter,
	}, nil
}

func (l *TokenBucketLimiter) Wait(ctx context.Context, req Request) error {
	for {
		result, err := l.Allow(ctx, req)
		if err != nil {
			return err
		}
		if result.Allowed {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(result.RetryAfter):
		}
	}
}

func (l *TokenBucketLimiter) GetUsage(_ context.Context, _ QuotaLevel, identifier string) (int64, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if b, ok := l.buckets[identifier]; ok {
		b.mu.Lock()
		used := int64(b.maxTokens - b.tokens)
		b.mu.Unlock()
		return used, nil
	}
	return 0, nil
}

func (l *TokenBucketLimiter) ResetQuota(_ context.Context, _ QuotaLevel, identifier string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.buckets, identifier)
	return nil
}

// --- SlidingWindowLimiter ---

type windowEntry struct {
	timestamp time.Time
}

type SlidingWindowLimiter struct {
	mu      sync.RWMutex
	windows map[string][]windowEntry
	limits  map[string]RateLimit
}

func NewSlidingWindowLimiter() *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		windows: make(map[string][]windowEntry),
		limits:  make(map[string]RateLimit),
	}
}

func (l *SlidingWindowLimiter) Allow(_ context.Context, req Request) (RateLimitResult, error) {
	key := fmt.Sprintf("%s:%s", req.UserID, req.Resource)
	l.mu.Lock()
	defer l.mu.Unlock()

	limit, ok := l.limits[key]
	if !ok {
		limit = RateLimit{Rate: 100, Period: time.Second}
	}

	now := time.Now()
	cutoff := now.Add(-limit.Period)
	entries := l.windows[key]
	// prune old entries
	n := 0
	for _, e := range entries {
		if e.timestamp.After(cutoff) {
			entries[n] = e
			n++
		}
	}
	entries = entries[:n]

	if len(entries) >= limit.Rate {
		resetAt := entries[0].timestamp.Add(limit.Period)
		return RateLimitResult{
			Allowed:    false,
			Remaining:  0,
			ResetAt:    resetAt,
			RetryAfter: time.Until(resetAt),
			QuotaLimit: int64(limit.Rate),
		}, nil
	}
	entries = append(entries, windowEntry{timestamp: now})
	l.windows[key] = entries
	return RateLimitResult{
		Allowed:      true,
		Remaining:    limit.Rate - len(entries),
		CurrentUsage: int64(len(entries)),
		QuotaLimit:   int64(limit.Rate),
	}, nil
}

func (l *SlidingWindowLimiter) Reserve(ctx context.Context, req Request, count int) (RateLimitResult, error) {
	return l.Allow(ctx, req)
}
func (l *SlidingWindowLimiter) Wait(ctx context.Context, req Request) error {
	for {
		result, err := l.Allow(ctx, req)
		if err != nil {
			return err
		}
		if result.Allowed {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(result.RetryAfter):
		}
	}
}
func (l *SlidingWindowLimiter) GetUsage(_ context.Context, _ QuotaLevel, identifier string) (int64, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return int64(len(l.windows[identifier])), nil
}
func (l *SlidingWindowLimiter) ResetQuota(_ context.Context, _ QuotaLevel, identifier string) error {
	l.mu.Lock()
	delete(l.windows, identifier)
	l.mu.Unlock()
	return nil
}

// --- HierarchicalQuotaManager ---

type quotaKey struct{ level QuotaLevel; id string }

type HierarchicalQuotaManager struct {
	mu     sync.RWMutex
	limits map[quotaKey]RateLimit
}

func NewHierarchicalQuotaManager() *HierarchicalQuotaManager {
	return &HierarchicalQuotaManager{limits: make(map[quotaKey]RateLimit)}
}

func (m *HierarchicalQuotaManager) SetLimit(_ context.Context, level QuotaLevel, id string, limit RateLimit) error {
	m.mu.Lock()
	m.limits[quotaKey{level, id}] = limit
	m.mu.Unlock()
	return nil
}

func (m *HierarchicalQuotaManager) GetLimit(_ context.Context, level QuotaLevel, id string) (*RateLimit, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// Check specific, then global
	for _, k := range []quotaKey{{level, id}, {QuotaGlobal, "default"}} {
		if l, ok := m.limits[k]; ok {
			lCopy := l
			return &lCopy, nil
		}
	}
	return nil, fmt.Errorf("no limit for %s:%s", level, id)
}

func (m *HierarchicalQuotaManager) DeleteLimit(_ context.Context, level QuotaLevel, id string) error {
	m.mu.Lock()
	delete(m.limits, quotaKey{level, id})
	m.mu.Unlock()
	return nil
}

func (m *HierarchicalQuotaManager) ListLimits(_ context.Context) (map[string]RateLimit, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]RateLimit, len(m.limits))
	for k, v := range m.limits {
		out[fmt.Sprintf("%s:%s", k.level, k.id)] = v
	}
	return out, nil
}

func main() {
	limiter := NewTokenBucketLimiter()
	ctx := context.Background()

	req := Request{UserID: "user-1", IP: "1.2.3.4", Resource: "/api/search"}

	for i := 0; i < 5; i++ {
		result, _ := limiter.Allow(ctx, req)
		fmt.Printf("request %d: allowed=%v remaining=%d\n", i+1, result.Allowed, result.Remaining)
	}

	swLimiter := NewSlidingWindowLimiter()
	swLimiter.limits[fmt.Sprintf("%s:%s", req.UserID, req.Resource)] = RateLimit{Rate: 3, Period: time.Second}
	for i := 0; i < 5; i++ {
		result, _ := swLimiter.Allow(ctx, req)
		fmt.Printf("sliding window request %d: allowed=%v remaining=%d\n", i+1, result.Allowed, result.Remaining)
	}
}
