// ЗАДАЧА 10: Tee
package main

import (
	"context"
	"reflect"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	i := 0
	inc := func() interface{} {
		i++
		return i
	}

	out1, out2 := tee(ctx, take(ctx, repeatFn(ctx, inc), 3))

	var res1, res2 []interface{}
	for val1 := range out1 {
		res1 = append(res1, val1)
		res2 = append(res2, <-out2)
	}

	exp := []interface{}{1, 2, 3}
	if !reflect.DeepEqual(res1, exp) || !reflect.DeepEqual(res2, exp) {
		panic("wrong code")
	}
}

// tee разветвляет канал in в два канала с одинаковыми данными.
func tee(ctx context.Context, in <-chan interface{}) (_, _ <-chan interface{}) {
	out1 := make(chan interface{})
	out2 := make(chan interface{})
	go func() {
		defer close(out1)
		defer close(out2)
		for v := range orDone(ctx, in) {
			// Сохраняем значение и отправляем в оба канала
			v1, v2 := v, v
			// Используем локальные переменные чтобы каждая отправка была независима
			for _, ch := range []chan<- interface{}{out1, out2} {
				select {
				case ch <- v1:
				case <-ctx.Done():
					return
				}
				_ = v2
				v1, v2 = v2, v1 // поменяем местами чтобы v2 отправился второй раз
			}
		}
	}()
	return out1, out2
}

func orDone(ctx context.Context, in <-chan interface{}) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- v:
				case <-ctx.Done():
				}
			}
		}
	}()
	return out
}

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
