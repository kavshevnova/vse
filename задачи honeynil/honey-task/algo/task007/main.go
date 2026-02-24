package main

// Задача: найти путь до подкатегории в дереве (DFS).
// Вывести путь от корня БЕЗ "root", через " > ".
//
// search_name = "OLED"
// out: "Бытовая техника > Телевизоры > OLED"

import (
	"fmt"
	"strings"
)

type Category struct {
	Name     string
	Children []*Category
}

// findPath возвращает срез имён от потомка root до target (не включая root).
func findPath(root *Category, target string) []string {
	var dfs func(node *Category, path []string) []string
	dfs = func(node *Category, path []string) []string {
		current := append(path, node.Name)
		if node.Name == target {
			return current
		}
		for _, child := range node.Children {
			if result := dfs(child, current); result != nil {
				return result
			}
		}
		return nil
	}
	path := dfs(root, nil)
	if len(path) <= 1 {
		return nil
	}
	return path[1:] // убираем "root"
}

func main() {
	root := &Category{
		Name: "root",
		Children: []*Category{
			{
				Name: "Бытовая техника",
				Children: []*Category{
					{
						Name: "Телевизоры",
						Children: []*Category{
							{Name: "ЭЛТ"},
							{Name: "LED"},
							{Name: "OLED"},
						},
					},
					{
						Name: "Холодильники",
						Children: []*Category{
							{Name: "Двухкамерные"},
							{Name: "Однокамерные"},
						},
					},
					{Name: "Утюги"},
				},
			},
			{
				Name: "Растения",
				Children: []*Category{
					{Name: "Комнатные"},
					{Name: "Садовые"},
				},
			},
		},
	}

	path := findPath(root, "OLED")
	fmt.Println(strings.Join(path, " > "))
	// Бытовая техника > Телевизоры > OLED

	path2 := findPath(root, "Садовые")
	fmt.Println(strings.Join(path2, " > "))
	// Растения > Садовые

	path3 := findPath(root, "Несуществующий")
	fmt.Println(path3) // []
}
