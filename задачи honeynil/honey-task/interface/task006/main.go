package main

// Задача: Worker Pool — ограниченный параллелизм + graceful shutdown.

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Task func() error

type WorkerPool interface {
	Submit(task Task) error
	Start(ctx context.Context) error
	Shutdown() error
	ShutdownNow() error
}

type SimpleWorkerPool struct {
	workerCount int
	queue       chan Task
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	once        sync.Once
	closed      bool
	mu          sync.Mutex
}

func NewWorkerPool(workerCount int, queueSize int) *SimpleWorkerPool {
	return &SimpleWorkerPool{
		workerCount: workerCount,
		queue:       make(chan Task, queueSize),
	}
}

func (p *SimpleWorkerPool) Start(ctx context.Context) error {
	p.ctx, p.cancel = context.WithCancel(ctx)
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for {
				select {
				case task, ok := <-p.queue:
					if !ok {
						return
					}
					task()
				case <-p.ctx.Done():
					return
				}
			}
		}()
	}
	return nil
}

func (p *SimpleWorkerPool) Submit(task Task) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return errors.New("pool is closed")
	}
	p.queue <- task
	return nil
}

// Shutdown ждёт выполнения всех задач в очереди.
func (p *SimpleWorkerPool) Shutdown() error {
	p.once.Do(func() {
		p.mu.Lock()
		p.closed = true
		p.mu.Unlock()
		close(p.queue)
		p.wg.Wait()
		p.cancel()
	})
	return nil
}

// ShutdownNow немедленно останавливает пул.
func (p *SimpleWorkerPool) ShutdownNow() error {
	p.once.Do(func() {
		p.mu.Lock()
		p.closed = true
		p.mu.Unlock()
		p.cancel()
		// Дренируем очередь
		for len(p.queue) > 0 {
			<-p.queue
		}
		close(p.queue)
		p.wg.Wait()
	})
	return nil
}

func main() {
	pool := NewWorkerPool(3, 10)
	pool.Start(context.Background())

	var mu sync.Mutex
	results := []int{}
	for i := 0; i < 5; i++ {
		n := i
		pool.Submit(func() error {
			mu.Lock()
			results = append(results, n)
			mu.Unlock()
			return nil
		})
	}
	pool.Shutdown()
	fmt.Println("processed:", len(results), "tasks")
}
