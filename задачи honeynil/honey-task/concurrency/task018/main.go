// ЗАДАЧА 18: Produce и Main с сигналами
package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	produceCount = 3
	produceStop  = 10
)

// produce бесконечно пишет числа в pipe, завершается по сигналу quit.
func produce(pipe chan<- int, quit <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	i := 0
	for {
		select {
		case <-quit:
			time.Sleep(3 * time.Second)
			fmt.Println("produce finished")
			return
		default:
			pipe <- i
			i++
		}
	}
}

func main() {
	pipe := make(chan int, 100)
	quit := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(produceCount)
	for i := 0; i < produceCount; i++ {
		go produce(pipe, quit, &wg)
	}

	// Читаем из канала пока не получим produceStop
	for v := range pipe {
		fmt.Println(v)
		if v >= produceStop {
			break
		}
	}

	// Сигнализируем всем produce завершиться
	close(quit)

	wg.Wait()
	fmt.Println("main finished")
}
