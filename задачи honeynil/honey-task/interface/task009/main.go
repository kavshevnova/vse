package main

// Задача: CircuitBreaker — защита от каскадных сбоев (Closed → Open → HalfOpen).

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type State string

const (
	StateClosed   State = "closed"
	StateOpen     State = "open"
	StateHalfOpen State = "half-open"
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

type CircuitBreaker interface {
	Call(fn func() (interface{}, error)) (interface{}, error)
	State() State
	Reset()
}

type SimpleCircuitBreaker struct {
	mu           sync.Mutex
	state        State
	failures     int
	maxFailures  int
	timeout      time.Duration
	lastFailTime time.Time
}

func NewCircuitBreaker(maxFailures int, timeout time.Duration) *SimpleCircuitBreaker {
	return &SimpleCircuitBreaker{
		state:       StateClosed,
		maxFailures: maxFailures,
		timeout:     timeout,
	}
}

func (cb *SimpleCircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.currentState()
}

func (cb *SimpleCircuitBreaker) currentState() State {
	if cb.state == StateOpen && time.Since(cb.lastFailTime) > cb.timeout {
		cb.state = StateHalfOpen
	}
	return cb.state
}

func (cb *SimpleCircuitBreaker) Call(fn func() (interface{}, error)) (interface{}, error) {
	cb.mu.Lock()
	state := cb.currentState()
	if state == StateOpen {
		cb.mu.Unlock()
		return nil, ErrCircuitOpen
	}
	cb.mu.Unlock()

	result, err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailTime = time.Now()
		if cb.state == StateHalfOpen || cb.failures >= cb.maxFailures {
			cb.state = StateOpen
		}
		return nil, err
	}

	// Успех
	if cb.state == StateHalfOpen {
		cb.state = StateClosed
		cb.failures = 0
	}
	return result, nil
}

func (cb *SimpleCircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failures = 0
}

func main() {
	cb := NewCircuitBreaker(3, 1*time.Second)

	fail := func() (interface{}, error) { return nil, errors.New("service error") }
	ok := func() (interface{}, error) { return "ok", nil }

	for i := 0; i < 3; i++ {
		_, err := cb.Call(fail)
		fmt.Println("call:", err)
	}
	fmt.Println("state:", cb.State()) // open

	_, err := cb.Call(ok)
	fmt.Println("blocked:", err) // circuit open

	time.Sleep(1100 * time.Millisecond)
	fmt.Println("state after timeout:", cb.State()) // half-open

	v, _ := cb.Call(ok)
	fmt.Println("result:", v)                // ok
	fmt.Println("state after success:", cb.State()) // closed
}
