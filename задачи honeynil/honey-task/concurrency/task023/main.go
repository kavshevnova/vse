// Задача 3 – Канал-стопер done
// Как отработает код?
//
// ОТВЕТ:
//	processed: cmd.1
//	processed: cmd.2
//	stopped
//
// Объяснение:
// - ch и done — небуферизованные каналы
// - Запускается горутина с select на ch и done
// - main отправляет cmd.1 → горутина печатает "processed: cmd.1"
// - main отправляет cmd.2 → горутина печатает "processed: cmd.2"
// - main отправляет в done → горутина выбирает case <-done → "stopped" и return
// - Программа завершается корректно (no leak, no deadlock)
//
// ЗАМЕЧАНИЕ: порядок "cmd.2" и "stopped" ГАРАНТИРОВАН, т.к. cmd.2 отправляется ДО done.
package main

import (
	"fmt"
)

func main() {
	ch := make(chan string)
	done := make(chan struct{})

	go func() {
		for {
			select {
			case msg := <-ch:
				fmt.Println("processed:", msg)
			case <-done:
				fmt.Println("stopped")
				return
			}
		}
	}()

	ch <- "cmd.1"
	ch <- "cmd.2"
	done <- struct{}{} // корректное завершение
}
