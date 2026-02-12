package main

import (
	"fmt"
	"strings"
)

func IsAllCharactersAreUnique(str string) bool {
	str = strings.ToLower(str)
	m := make(map[string]int)
	for _, v := range str {
		m[string(v)]++
	}
	for k := range m {
		if m[k] > 1 {
			return false
		}
	}
	return true
}

func main() {
	var input string
	fmt.Println("Введите строку")
	fmt.Scan(&input)
	if IsAllCharactersAreUnique(input) {
		fmt.Println("Все символы уникальны,йей!")
	} else {
		fmt.Println("Символы НЕ уникальны, увы")
	}
}
