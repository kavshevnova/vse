// Задача 6 – Порядок defer
// Как отработает код?
//
// ОТВЕТ:
//	1
//	0
//
// Объяснение:
// - defer fmt.Println(a): аргументы defer ВЫЧИСЛЯЮТСЯ СРАЗУ при регистрации
//   → a=0 в момент defer → выведет 0
// - defer func() { a++; fmt.Println(a) }(): замыкание захватывает переменную a по ссылке
//   → выполняется первым (LIFO) → a++ (0→1) → выводит 1
//
// Порядок выполнения defer — LIFO (Last In, First Out):
// 1. Сначала анонимная функция: a++ → a=1; println(1)
// 2. Потом fmt.Println(a=0): выводит 0 (значение было зафиксировано при defer)
package main

import "fmt"

func main() {
	a := 0
	defer fmt.Println(a)
	defer func() {
		a++
		fmt.Println(a)
	}()
}
