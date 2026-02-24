package main

// Задача: RateLimiter на алгоритме Token Bucket (потокобезопасный).

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type RateLimiter interface {
	Allow() bool
	Wait() error
	WaitN(n int) error
}

type TokenBucket struct {
	mu       sync.Mutex
	tokens   float64
	capacity float64
	rate     float64 // токенов в секунду
	lastTime time.Time
}

func NewTokenBucket(rate int, capacity int) *TokenBucket {
	return &TokenBucket{
		tokens:   float64(capacity),
		capacity: float64(capacity),
		rate:     float64(rate),
		lastTime: time.Now(),
	}
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastTime).Seconds()
	tb.tokens = min(tb.capacity, tb.tokens+elapsed*tb.rate)
	tb.lastTime = now
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.refill()
	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

func (tb *TokenBucket) Wait() error {
	return tb.WaitN(1)
}

func (tb *TokenBucket) WaitN(n int) error {
	return tb.WaitNContext(context.Background(), n)
}

func (tb *TokenBucket) WaitNContext(ctx context.Context, n int) error {
	for {
		tb.mu.Lock()
		tb.refill()
		if tb.tokens >= float64(n) {
			tb.tokens -= float64(n)
			tb.mu.Unlock()
			return nil
		}
		// Вычисляем время ожидания
		needed := float64(n) - tb.tokens
		waitDuration := time.Duration(needed/tb.rate*float64(time.Second)) + time.Millisecond
		tb.mu.Unlock()

		select {
		case <-ctx.Done():
			return errors.New("context cancelled")
		case <-time.After(waitDuration):
		}
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func main() {
	limiter := NewTokenBucket(5, 10) // 5 токенов/с, max 10
	for i := 0; i < 5; i++ {
		if limiter.Allow() {
			fmt.Printf("request %d allowed\n", i)
		} else {
			fmt.Printf("request %d denied\n", i)
		}
	}
	fmt.Println("waiting for 1 token...")
	limiter.Wait()
	fmt.Println("token acquired")
}
