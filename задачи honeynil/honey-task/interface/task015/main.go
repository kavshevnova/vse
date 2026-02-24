package main

// Задача: Middleware Chain — цепочка обработчиков с возможностью прерывания.

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type Context interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	Next() error
	Abort(err error)
	IsAborted() bool
	Error() error
}

type Handler func(ctx Context) error

type Middleware interface {
	Handle(ctx Context) error
}

type MiddlewareChain interface {
	Use(middleware Middleware) MiddlewareChain
	Execute(ctx Context) error
}

// --- SimpleContext ---

type SimpleContext struct {
	mu          sync.RWMutex
	data        map[string]interface{}
	chain       []Middleware
	index       int
	aborted     bool
	err         error
	goCtx       context.Context
}

func NewSimpleContext(goCtx context.Context) *SimpleContext {
	return &SimpleContext{data: make(map[string]interface{}), goCtx: goCtx}
}

func (c *SimpleContext) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.data[key]
	return v, ok
}

func (c *SimpleContext) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

func (c *SimpleContext) Next() error {
	c.index++
	if c.index < len(c.chain) {
		return c.chain[c.index].Handle(c)
	}
	return nil
}

func (c *SimpleContext) Abort(err error) { c.aborted = true; c.err = err }
func (c *SimpleContext) IsAborted() bool { return c.aborted }
func (c *SimpleContext) Error() error    { return c.err }

// --- Chain ---

type Chain struct {
	middlewares []Middleware
}

func (ch *Chain) Use(m Middleware) MiddlewareChain {
	ch.middlewares = append(ch.middlewares, m)
	return ch
}

func (ch *Chain) Execute(ctx Context) error {
	if sc, ok := ctx.(*SimpleContext); ok {
		sc.chain = ch.middlewares
		sc.index = -1
	}
	return ctx.Next()
}

// --- LoggingMiddleware ---

type LoggingMiddleware struct{}

func (m *LoggingMiddleware) Handle(ctx Context) error {
	start := time.Now()
	err := ctx.Next()
	log.Printf("request processed in %v, err=%v", time.Since(start), err)
	return err
}

// --- RecoveryMiddleware ---

type RecoveryMiddleware struct{}

func (m *RecoveryMiddleware) Handle(ctx Context) error {
	defer func() {
		if r := recover(); r != nil {
			ctx.Abort(fmt.Errorf("panic recovered: %v", r))
		}
	}()
	return ctx.Next()
}

// --- TimeoutMiddleware ---

type TimeoutMiddleware struct{ Timeout time.Duration }

func (m *TimeoutMiddleware) Handle(ctx Context) error {
	done := make(chan error, 1)
	go func() { done <- ctx.Next() }()
	select {
	case err := <-done:
		return err
	case <-time.After(m.Timeout):
		ctx.Abort(fmt.Errorf("timeout after %v", m.Timeout))
		return ctx.Error()
	}
}

// --- AuthMiddleware ---

type AuthMiddleware struct{}

func (m *AuthMiddleware) Handle(ctx Context) error {
	token, ok := ctx.Get("token")
	if !ok || token == "" {
		ctx.Abort(fmt.Errorf("unauthorized"))
		return ctx.Error()
	}
	return ctx.Next()
}

func main() {
	chain := &Chain{}
	chain.Use(&LoggingMiddleware{}).
		Use(&RecoveryMiddleware{}).
		Use(&AuthMiddleware{})

	// Без токена
	goCtx := context.Background()
	ctx1 := NewSimpleContext(goCtx)
	err := chain.Execute(ctx1)
	fmt.Println("no token:", err)

	// С токеном
	ctx2 := NewSimpleContext(goCtx)
	ctx2.Set("token", "Bearer abc123")
	err = chain.Execute(ctx2)
	fmt.Println("with token:", err)
}
