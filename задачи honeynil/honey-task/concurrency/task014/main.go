// ЗАДАЧА 14: Run конкурентных функций
package main

import (
	"errors"
	"sync"
	"sync/atomic"
)

type fn func() error

func main() {
	expErr := errors.New("error")

	funcs := []fn{
		func() error { return nil },
		func() error { return nil },
		func() error { return expErr },
		func() error { return nil },
	}

	if err := Run(funcs...); !errors.Is(err, expErr) {
		panic("wrong code")
	}
}

// Run запускает все функции конкурентно и возвращает первую ошибку (если есть).
func Run(fs ...fn) error {
	var (
		wg      sync.WaitGroup
		firstErr atomic.Value
	)
	for _, f := range fs {
		wg.Add(1)
		go func(f fn) {
			defer wg.Done()
			if err := f(); err != nil {
				firstErr.CompareAndSwap(nil, err)
			}
		}(f)
	}
	wg.Wait()
	if v := firstErr.Load(); v != nil {
		return v.(error)
	}
	return nil
}
