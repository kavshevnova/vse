// ЗАДАЧА 8: ExecuteTaskWithTimeout
package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

const timeout = 100 * time.Millisecond

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := executeTaskWithTimeout(ctx)
	if err != nil {
		fmt.Println("timeout:", err)
		return
	}
	fmt.Println("task done")
}

// executeTaskWithTimeout выполняет задачу, завершаясь по отмене контекста.
func executeTaskWithTimeout(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		executeTask()
		close(done)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

func executeTask() {
	time.Sleep(time.Duration(rand.Intn(3)) * timeout)
}
