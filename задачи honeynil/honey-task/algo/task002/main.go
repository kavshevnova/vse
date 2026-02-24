package main

// Задача: проверить, являются ли две строки анаграммами.
// Строки могут содержать только буквы латиницы и кириллицы.
//
// isAnagram("anagram", "nagaram") → true
// isAnagram("кит", "ток") → false

import "fmt"

func isAnagram(s, t string) bool {
	if len(s) != len(t) {
		return false
	}
	counts := make(map[rune]int)
	for _, r := range s {
		counts[r]++
	}
	for _, r := range t {
		counts[r]--
		if counts[r] < 0 {
			return false
		}
	}
	return true
}

func main() {
	fmt.Println(isAnagram("anagram", "nagaram")) // true
	fmt.Println(isAnagram("кит", "ток"))         // false
	fmt.Println(isAnagram("rat", "car"))          // false
}
