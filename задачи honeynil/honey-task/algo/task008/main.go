package main

// Задача: восстановить маршрут из неупорядоченных пар городов (направление неизвестно).
// Каждый город посещён ровно один раз; начало и конец — разные города.
// Города с нечётным числом соединений (degree=1) — это начало или конец маршрута.
//
// [("Москва", "Белград")] → ["Москва", "Белград"]
// [("Москва", "Белград"), ("Москва", "Ереван")] → ["Ереван", "Москва", "Белград"]

import "fmt"

// Flight — пара городов (направление неизвестно).
type Flight = [2]string

// GetRoute восстанавливает маршрут из набора перелётов.
// Решение: строим граф, находим вершины со степенью 1 (начало/конец),
// затем обходим граф в правильном порядке.
func GetRoute(flights []Flight) []string {
	// Строим список смежности (граф без направления)
	adj := make(map[string][]string)
	for _, f := range flights {
		adj[f[0]] = append(adj[f[0]], f[1])
		adj[f[1]] = append(adj[f[1]], f[0])
	}

	// Вершины со степенью 1 — начало и конец маршрута
	var endpoints []string
	for city, neighbors := range adj {
		if len(neighbors) == 1 {
			endpoints = append(endpoints, city)
		}
	}

	start := endpoints[0]

	// DFS/обход без рекурсии (граф — простая цепочка)
	visited := make(map[string]bool)
	route := []string{start}
	visited[start] = true

	for {
		current := route[len(route)-1]
		moved := false
		for _, next := range adj[current] {
			if !visited[next] {
				visited[next] = true
				route = append(route, next)
				moved = true
				break
			}
		}
		if !moved {
			break
		}
	}
	return route
}

func main() {
	fmt.Println(GetRoute([]Flight{{"Москва", "Белград"}}))
	// [Москва Белград]

	fmt.Println(GetRoute([]Flight{{"Москва", "Белград"}, {"Москва", "Ереван"}}))
	// [Ереван Москва Белград] или [Белград Москва Ереван]

	fmt.Println(GetRoute([]Flight{{"Рим", "Белград"}, {"Москва", "Белград"}, {"Москва", "Ереван"}}))
	// [Ереван Москва Белград Рим] или обратный
}
