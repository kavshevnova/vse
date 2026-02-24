// ЗАДАЧА 4: RepeatFn и Take
package main

import (
	"context"
	"math/rand"
)

// repeatFn бесконечно вызывает fn и отправляет результат в канал.
func repeatFn(ctx context.Context, fn func() interface{}) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case out <- fn():
			}
		}
	}()
	return out
}

// take читает не более num значений из канала in.
func take(ctx context.Context, in <-chan interface{}, num int) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		defer close(out)
		for i := 0; i < num; i++ {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				out <- v
			}
		}
	}()
	return out
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	randFn := func() interface{} { return rand.Int() }

	var res []interface{}
	for num := range take(ctx, repeatFn(ctx, randFn), 3) {
		res = append(res, num)
	}

	if len(res) != 3 {
		panic("wrong code")
	}
}
