package main

// Задача: для каждого продавца вернуть количество участников, которых он опередил.
// Общее количество участников продавцу НЕИЗВЕСТНО (не возвращаем ранг, только кол-во побеждённых).
//
// sales = [8,1,2,2,3] → [4,0,1,1,3]
// sales = [5,5,5,5]   → [0,0,0,0]

import "fmt"

// beatCount возвращает для каждого продавца количество участников с меньшим результатом.
func beatCount(sales []int) []int {
	// Строим карту: значение → количество продавцов со строго меньшим результатом
	counts := make(map[int]int)
	for _, s := range sales {
		for _, other := range sales {
			if other < s {
				counts[s]++
			}
		}
	}
	// Т.к. counts[s] суммирует по всем элементам, нужно разделить на кол-во одинаковых s
	// Нет — просто строим уникальный счётчик для каждого значения
	// Пересчитаем: для каждого уникального значения v, сколько sales[j] < v
	unique := make(map[int]int)
	for _, s := range sales {
		if _, seen := unique[s]; !seen {
			cnt := 0
			for _, other := range sales {
				if other < s {
					cnt++
				}
			}
			unique[s] = cnt
		}
	}
	result := make([]int, len(sales))
	for i, s := range sales {
		result[i] = unique[s]
	}
	return result
}

func main() {
	fmt.Println(beatCount([]int{8, 1, 2, 2, 3})) // [4 0 1 1 3]
	fmt.Println(beatCount([]int{5, 5, 5, 5}))     // [0 0 0 0]
}
