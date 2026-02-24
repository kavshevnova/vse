// ЗАДАЧА 16: Inc цепочка каналов
package main

// inc читает из in, прибавляет 1 и пишет в новый канал.
func inc(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			out <- v + 1
		}
	}()
	return out
}

func main() {
	n := 10

	// Создаём цепочку из n каналов: каждый добавляет 1
	first := make(chan int, 1)
	first <- 0 // начальное значение
	close(first)

	last := (<-chan int)(first)
	for i := 0; i < n; i++ {
		last = inc(last)
	}

	result := <-last
	if n != result {
		panic("wrong code")
	}
}
