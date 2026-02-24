package main

import (
	"fmt"
	"math/rand"
	"strings"
)

// ТЗ: Функции для обработки больших объемов данных.
// Необходимо оптимизировать работу с памятью при выделении slice и map.
// Найти проблемы.

func ProcessUserData(userIDs []int) map[int]string {
	// Обрабатываем данные пользователей
	results := make(map[int]string)

	for _, id := range userIDs {
		userData := FetchUserData(id)
		results[id] = userData
	}

	return results
}

func FetchUserData(id int) string {
	return fmt.Sprintf("User data for ID: %d", id)
}

func GenerateReport(data []string) string {
	var report string

	report += "=== REPORT ===\n"
	report += "Generated at: 2024-01-01\n"
	report += "==============\n\n"

	for i, line := range data {
		report += fmt.Sprintf("%d. %s\n", i+1, line)
	}

	report += "\n==============\n"
	report += fmt.Sprintf("Total items: %d\n", len(data))

	return report
}

func FilterActiveUsers(users []map[string]interface{}) []map[string]interface{} {
	var active []map[string]interface{}

	for _, user := range users {
		if user["active"] == true {
			active = append(active, user)
		}
	}

	return active
}

func MergeDataSets(set1, set2, set3 []int) []int {
	var merged []int

	merged = append(merged, set1...)
	merged = append(merged, set2...)
	merged = append(merged, set3...)

	return merged
}

func BuildQueryString(params map[string]string) string {
	var parts []string

	for key, value := range params {
		parts = append(parts, key+"="+value)
	}

	return "?" + strings.Join(parts, "&")
}

func main() {
	// Тест 1: Обработка 10000 пользователей
	userIDs := make([]int, 10000)
	for i := 0; i < 10000; i++ {
		userIDs[i] = i
	}
	userDataMap := ProcessUserData(userIDs)
	fmt.Printf("Processed %d users\n", len(userDataMap))

	// Тест 2: Генерация отчета из 5000 строк
	reportData := make([]string, 5000)
	for i := 0; i < 5000; i++ {
		reportData[i] = fmt.Sprintf("Line %d with some data", i)
	}
	report := GenerateReport(reportData)
	fmt.Printf("Generated report with %d characters\n", len(report))

	// Тест 3: Фильтрация пользователей
	allUsers := make([]map[string]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		allUsers[i] = map[string]interface{}{
			"id":     i,
			"name":   fmt.Sprintf("User%d", i),
			"active": rand.Intn(2) == 1,
		}
	}
	activeUsers := FilterActiveUsers(allUsers)
	fmt.Printf("Found %d active users out of %d\n", len(activeUsers), len(allUsers))

	// Тест 4: Объединение больших датасетов
	set1 := make([]int, 1000)
	set2 := make([]int, 2000)
	set3 := make([]int, 1500)
	merged := MergeDataSets(set1, set2, set3)
	fmt.Printf("Merged %d elements\n", len(merged))

	// Тест 5: Построение query string
	params := make(map[string]string)
	for i := 0; i < 50; i++ {
		params[fmt.Sprintf("param%d", i)] = fmt.Sprintf("value%d", i)
	}
	query := BuildQueryString(params)
	fmt.Printf("Query string length: %d\n", len(query))
}
