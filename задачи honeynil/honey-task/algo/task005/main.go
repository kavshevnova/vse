package main

// Задача: сгруппировать отзывы по рейтингу.
//
// out: { 5: ["Отлично!", "Все как в описании"], 4: ["Хороший товар"], ... }

import "fmt"

type Review struct {
	Text   string
	Rating int
}

func groupReviews(reviews []Review) map[int][]string {
	result := make(map[int][]string)
	for _, r := range reviews {
		result[r.Rating] = append(result[r.Rating], r.Text)
	}
	return result
}

func main() {
	reviews := []Review{
		{"Отлично!", 5},
		{"Хороший товар", 4},
		{"Ожидал большего", 3},
		{"Не оправдал ожиданий", 1},
		{"Все как в описании", 5},
		{"Не понравилось", 1},
	}
	groups := groupReviews(reviews)
	fmt.Println(groups)
	// map[1:[Не оправдал ожиданий Не понравилось] 3:[Ожидал большего] 4:[Хороший товар] 5:[Отлично! Все как в описании]]
}
