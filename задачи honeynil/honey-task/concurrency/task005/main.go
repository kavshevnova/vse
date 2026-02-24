// ЗАДАЧА 5: Worker Pool
package main

import (
	"fmt"
	"sync"
)

func worker(f func(int) int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		results <- f(job)
	}
}

const numJobs = 5
const numWorkers = 3

func main() {
	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)
	wg := sync.WaitGroup{}

	multiplier := func(x int) int {
		return x * 10
	}

	// Запускаем numWorkers воркеров
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go worker(multiplier, jobs, results, &wg)
	}

	// Отправляем задания
	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	close(jobs)

	// Закрываем results после завершения всех воркеров
	go func() {
		wg.Wait()
		close(results)
	}()

	// Читаем и выводим результаты
	for r := range results {
		fmt.Println(r)
	}
}
