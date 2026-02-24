// ЗАДАЧА 11: Bridge
package main

import (
	"context"
	"reflect"
)

func main() {
	genVals := func() <-chan <-chan interface{} {
		out := make(chan (<-chan interface{}))
		go func() {
			defer close(out)
			for i := 0; i < 3; i++ {
				stream := make(chan interface{}, 1)
				stream <- i
				close(stream)
				out <- stream
			}
		}()
		return out
	}

	var res []interface{}
	for v := range bridge(context.Background(), genVals()) {
		res = append(res, v)
	}

	if !reflect.DeepEqual(res, []interface{}{0, 1, 2}) {
		panic("wrong code")
	}
}

// bridge последовательно вычитывает каналы из ins и перенаправляет их данные.
func bridge(ctx context.Context, ins <-chan <-chan interface{}) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		defer close(out)
		for {
			var inner <-chan interface{}
			select {
			case <-ctx.Done():
				return
			case ch, ok := <-ins:
				if !ok {
					return
				}
				inner = ch
			}
			for v := range orDone(ctx, inner) {
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
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
