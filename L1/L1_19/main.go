package main

import "fmt"

func main() {
	fmt.Println("Введите строку для разворота")
	var inp string
	fmt.Scanln(&inp)
	runeArray := []rune(inp)
	n := len(runeArray)
	output := make([]rune, n)
	for i, x := range runeArray {
		output[n-i-1] = x
	}
	fmt.Println(string(output))
}
