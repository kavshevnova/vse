// ЗАДАЧА 3: Generator и Squarer с контекстом
package main

import (
	"context"
	"fmt"
)

func main() {
	ctx := context.Background()
	pipeline := squarer(ctx, generator(ctx, 1, 2, 3))
	for x := range pipeline {
		fmt.Println(x) // 1, 4, 9
	}
}

// generator последовательно отправляет числа в канал, завершается по ctx.Done().
func generator(ctx context.Context, in ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, v := range in {
			select {
			case <-ctx.Done():
				return
			case out <- v:
			}
		}
	}()
	return out
}

// squarer читает числа, возводит в квадрат, отправляет в новый канал.
func squarer(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			select {
			case <-ctx.Done():
				return
			case out <- v * v:
			}
		}
	}()
	return out
}
