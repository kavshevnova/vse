package main

// Задача: фильтрация продавцов по запрошенным городам.
// Вернуть только тех продавцов, у которых есть хотя бы один из запрошенных городов,
// и оставить только те города продавца, которые есть в запросе.
//
// out: { 1: [Москва], 2: [Москва, Казань], 4: [Москва, Казань, Тула] }
// (продавец 3 исключён — у него нет ни одного из запрошенных городов)

import "fmt"

func filterSellers(sellers map[int][]string, cities []string) map[int][]string {
	citySet := make(map[string]bool, len(cities))
	for _, c := range cities {
		citySet[c] = true
	}

	// Сохраняем оригинальный порядок городов из cities
	result := make(map[int][]string)
	for sellerID, sellerCities := range sellers {
		var matched []string
		for _, city := range cities { // итерируем по cities, чтобы сохранить порядок
			for _, sc := range sellerCities {
				if sc == city {
					matched = append(matched, city)
					break
				}
			}
		}
		if len(matched) > 0 {
			result[sellerID] = matched
		}
	}
	return result
}

func main() {
	sellers := map[int][]string{
		1: {"Москва", "Самара", "Ростов"},
		2: {"Москва", "Самара", "Ростов", "Казань", "Курган", "Пенза"},
		3: {"Самара", "Ростов", "Курган", "Пенза"},
		4: {"Москва", "Казань", "Тула"},
	}
	cities := []string{"Москва", "Казань", "Тула"}

	result := filterSellers(sellers, cities)
	fmt.Println(result)
	// map[1:[Москва] 2:[Москва Казань] 4:[Москва Казань Тула]]
}
