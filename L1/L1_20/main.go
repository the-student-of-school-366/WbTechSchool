package main

import "fmt"

func main() {
	inpString := "sun dog snow"
	inp := []rune(inpString)
	n := len(inp)
	right := n - 1
	for i := right; i > 0; i-- {
		if inp[i] == ' ' {
			for j := i + 1; j <= right; j++ {
				fmt.Print(string(inp[j]))
			}
			right = i - 1
			fmt.Print(" ")
		}
	}
	for i := 0; i <= right; i++ {
		fmt.Print(string(inp[i]))
	}
}
