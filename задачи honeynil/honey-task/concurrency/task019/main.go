// ЗАДАЧА 19: Once с каналами (без пакета sync)
package main

import (
	"fmt"
	"sync"
)

const goroutinesNumber = 10

// once — реализация sync.Once через канал.
type once struct {
	done chan struct{}
}

func new() *once {
	o := &once{done: make(chan struct{}, 1)}
	o.done <- struct{}{} // один "токен"
	return o
}

// do выполняет f только при первом вызове.
func (o *once) do(f func()) {
	select {
	case <-o.done: // получили токен — первый вызов
		f()
	default: // токен уже взят — ничего не делаем
	}
}

func funcToCall() {
	fmt.Printf("call")
}

func main() {
	wg := sync.WaitGroup{}
	so := new()

	wg.Add(goroutinesNumber)
	for i := 0; i < goroutinesNumber; i++ {
		go func(f func()) {
			defer wg.Done()
			so.do(f)
		}(funcToCall)
	}

	wg.Wait()
	fmt.Println()
}
