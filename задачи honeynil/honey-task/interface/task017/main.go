package main

// Задача: Connection Pool — переиспользование соединений, контроль lifetime.

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Connection interface {
	Execute(query string) (interface{}, error)
	Close() error
	IsValid() bool
}

type ConnectionFactory interface {
	Create() (Connection, error)
}

type Pool interface {
	Acquire(ctx context.Context) (Connection, error)
	Release(conn Connection) error
	Stats() PoolStats
	Close() error
}

type PoolStats struct {
	TotalConnections  int
	IdleConnections   int
	ActiveConnections int
}

// --- Mock connection ---

type mockConn struct {
	id      int
	healthy bool
}

func (c *mockConn) Execute(query string) (interface{}, error) {
	if !c.healthy {
		return nil, errors.New("connection is invalid")
	}
	return fmt.Sprintf("result of %q from conn#%d", query, c.id), nil
}
func (c *mockConn) Close() error  { c.healthy = false; return nil }
func (c *mockConn) IsValid() bool { return c.healthy }

type MockFactory struct {
	mu  sync.Mutex
	seq int
}

func (f *MockFactory) Create() (Connection, error) {
	f.mu.Lock()
	f.seq++
	id := f.seq
	f.mu.Unlock()
	return &mockConn{id: id, healthy: true}, nil
}

// --- ConnectionPool ---

type poolConn struct {
	conn      Connection
	createdAt time.Time
}

type ConnectionPool struct {
	factory     ConnectionFactory
	minSize     int
	maxSize     int
	idle        chan *poolConn
	mu          sync.Mutex
	totalOpen   int
	closed      bool
	maxLifetime time.Duration
}

func NewConnectionPool(factory ConnectionFactory, minSize, maxSize int) *ConnectionPool {
	p := &ConnectionPool{
		factory:     factory,
		minSize:     minSize,
		maxSize:     maxSize,
		idle:        make(chan *poolConn, maxSize),
		maxLifetime: 5 * time.Minute,
	}
	for i := 0; i < minSize; i++ {
		if conn, err := factory.Create(); err == nil {
			p.mu.Lock()
			p.totalOpen++
			p.mu.Unlock()
			p.idle <- &poolConn{conn: conn, createdAt: time.Now()}
		}
	}
	return p
}

func (p *ConnectionPool) Acquire(ctx context.Context) (Connection, error) {
	for {
		select {
		case pc := <-p.idle:
			if !pc.conn.IsValid() || time.Since(pc.createdAt) > p.maxLifetime {
				_ = pc.conn.Close()
				p.mu.Lock()
				p.totalOpen--
				p.mu.Unlock()
				continue
			}
			return pc.conn, nil
		default:
		}

		p.mu.Lock()
		if p.closed {
			p.mu.Unlock()
			return nil, errors.New("pool is closed")
		}
		if p.totalOpen < p.maxSize {
			p.totalOpen++
			p.mu.Unlock()
			conn, err := p.factory.Create()
			if err != nil {
				p.mu.Lock()
				p.totalOpen--
				p.mu.Unlock()
				return nil, err
			}
			return conn, nil
		}
		p.mu.Unlock()

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case pc := <-p.idle:
			if !pc.conn.IsValid() {
				_ = pc.conn.Close()
				p.mu.Lock()
				p.totalOpen--
				p.mu.Unlock()
				continue
			}
			return pc.conn, nil
		}
	}
}

func (p *ConnectionPool) Release(conn Connection) error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return conn.Close()
	}
	p.mu.Unlock()
	select {
	case p.idle <- &poolConn{conn: conn, createdAt: time.Now()}:
	default:
		_ = conn.Close()
		p.mu.Lock()
		p.totalOpen--
		p.mu.Unlock()
	}
	return nil
}

func (p *ConnectionPool) Stats() PoolStats {
	p.mu.Lock()
	defer p.mu.Unlock()
	idle := len(p.idle)
	return PoolStats{
		TotalConnections:  p.totalOpen,
		IdleConnections:   idle,
		ActiveConnections: p.totalOpen - idle,
	}
}

func (p *ConnectionPool) Close() error {
	p.mu.Lock()
	p.closed = true
	p.mu.Unlock()
	close(p.idle)
	for pc := range p.idle {
		_ = pc.conn.Close()
	}
	return nil
}

func main() {
	pool := NewConnectionPool(&MockFactory{}, 2, 5)
	fmt.Println("stats:", pool.Stats())

	ctx := context.Background()
	conn, _ := pool.Acquire(ctx)
	result, _ := conn.Execute("SELECT 1")
	fmt.Println("result:", result)
	pool.Release(conn)
	fmt.Println("stats after release:", pool.Stats())
	pool.Close()
}
