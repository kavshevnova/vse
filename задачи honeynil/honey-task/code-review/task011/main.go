package main

import (
	"fmt"
	"time"
)

// ТЗ: Система обработки заказов. Воркеры должны обрабатывать заказы из очереди,
// а главная функция должна дождаться обработки всех заказов перед завершением.
// Найти и исправить проблемы.

type Order struct {
	ID       int
	Customer string
	Amount   float64
}

func main() {
	orders := []Order{
		{ID: 1, Customer: "Alice", Amount: 100.50},
		{ID: 2, Customer: "Bob", Amount: 250.00},
		{ID: 3, Customer: "Charlie", Amount: 75.25},
		{ID: 4, Customer: "Diana", Amount: 450.00},
		{ID: 5, Customer: "Eve", Amount: 320.75},
	}

	orderChan := make(chan Order)
	resultChan := make(chan string)

	for i := 1; i <= 3; i++ {
		go worker(i, orderChan, resultChan)
	}

	go func() {
		for _, order := range orders {
			orderChan <- order
		}
	}()

	processedCount := 0
	for result := range resultChan {
		fmt.Println(result)
		processedCount++
		if processedCount == len(orders) {
			break
		}
	}

	fmt.Println("All orders processed!")
}

func worker(id int, orders chan Order, results chan string) {
	for order := range orders {
		// Симулируем обработку заказа
		time.Sleep(100 * time.Millisecond)

		result := fmt.Sprintf("Worker %d processed order %d for %s: $%.2f",
			id, order.ID, order.Customer, order.Amount)

		results <- result
	}
}
