package main

// Задача: обход бинарного дерева категорий, подсчёт повторений имён.
//
// out: { 'Машины': 3, 'Легковые': 2, ... }

import "fmt"

type Category struct {
	Name  string
	Left  *Category
	Right *Category
}

func countNames(root *Category) map[string]int {
	result := make(map[string]int)
	var dfs func(node *Category)
	dfs = func(node *Category) {
		if node == nil {
			return
		}
		result[node.Name]++
		dfs(node.Left)
		dfs(node.Right)
	}
	dfs(root)
	return result
}

func main() {
	root := &Category{
		Name: "Машины",
		Left: &Category{
			Name: "Легковые",
			Left: &Category{
				Name:  "Российские",
				Left:  &Category{Name: "Легковые"},
				Right: &Category{Name: "Красивые"},
			},
		},
		Right: &Category{
			Name: "Грузовые",
			Left: &Category{
				Name:  "Большие",
				Left:  &Category{Name: "Зеленые"},
				Right: &Category{Name: "Незеленые"},
			},
			Right: &Category{
				Name:  "Красивые",
				Left:  &Category{Name: "Красные"},
				Right: &Category{
					Name: "Некрасные",
					Left: &Category{
						Name:  "Машины",
						Right: &Category{Name: "Машины"},
					},
				},
			},
		},
	}

	counts := countNames(root)
	fmt.Println(counts)
	// map[Большие:1 Грузовые:1 Зеленые:1 Красивые:2 Красные:1 Легковые:2 Машины:3 Незеленые:1 Некрасные:1 Российские:1]
}
