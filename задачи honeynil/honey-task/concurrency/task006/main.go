// ЗАДАЧА 6: MergeSorted
package main

import "fmt"

// mergeSorted объединяет два отсортированных канала в один отсортированный.
func mergeSorted(cs ...<-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		// Читаем текущие значения из каждого канала
		vals := make([]int, len(cs))
		valid := make([]bool, len(cs))
		chans := make([]<-chan int, len(cs))
		copy(chans, cs)

		// Инициализируем первые значения
		for i, ch := range chans {
			if v, ok := <-ch; ok {
				vals[i] = v
				valid[i] = true
			}
		}

		for {
			// Находим минимальный валидный
			minIdx := -1
			for i, v := range valid {
				if v {
					if minIdx == -1 || vals[i] < vals[minIdx] {
						minIdx = i
					}
				}
			}
			if minIdx == -1 {
				break
			}
			out <- vals[minIdx]
			// Читаем следующее из того же канала
			if v, ok := <-chans[minIdx]; ok {
				vals[minIdx] = v
			} else {
				valid[minIdx] = false
			}
		}
	}()
	return out
}

func fillChanA(c chan int) {
	c <- 1
	c <- 2
	c <- 4
	close(c)
}

func fillChanB(c chan int) {
	c <- -1
	c <- 4
	c <- 5
	close(c)
}

func main() {
	a, b := make(chan int), make(chan int)
	go fillChanA(a)
	go fillChanB(b)

	c := mergeSorted(a, b)

	for val := range c {
		fmt.Printf("%d ", val)
	}
	fmt.Println()
}

// Вывод: -1 1 2 4 4 5
