package main

// Задача: статистика серверов — распределить серверы по показателю стабильности.
//
// stats = [{server: 1, stability: 99}, {server: 2, stability: 97}, {server: 3, stability: 34}, {server: 4, stability: 97}, {server: 5, stability: 97.1}]
// out:  { 34: [3], 97: [2, 4], 99: [1], 97.1: [5] }

import "fmt"

type ServerStat struct {
	Server    int
	Stability float64
}

func groupByStability(stats []ServerStat) map[float64][]int {
	result := make(map[float64][]int)
	for _, s := range stats {
		result[s.Stability] = append(result[s.Stability], s.Server)
	}
	return result
}

func main() {
	stats := []ServerStat{
		{1, 99},
		{2, 97},
		{3, 34},
		{4, 97},
		{5, 97.1},
	}
	fmt.Println(groupByStability(stats))
	// map[34:[3] 97:[2 4] 97.1:[5] 99:[1]]
}
