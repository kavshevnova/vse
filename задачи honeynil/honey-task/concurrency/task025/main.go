// Задача 5 – Срезы и race condition с append
// Исправьте код чтобы он работал корректно
//
// ПРОБЛЕМА: DATA RACE на data (go run -race выявит)
// - 100 горутин одновременно вызывают data = append(data, val)
// - append не потокобезопасен; возможна потеря данных или паника
//
// ИСПРАВЛЕНИЕ 1: mutex вокруг append
// ИСПРАВЛЕНИЕ 2: использовать sync.Mutex или channel для сбора результатов
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	var mu sync.Mutex // ИСПРАВЛЕНИЕ: добавить mutex
	data := []int{}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			mu.Lock()         // ИСПРАВЛЕНИЕ: защитить append
			data = append(data, val)
			mu.Unlock()       // ИСПРАВЛЕНИЕ
		}(i)
	}

	wg.Wait()
	fmt.Println("Length:", len(data))
	fmt.Println("Expected: 100")
}
