package main

import "fmt"

func main() {
	numbers := make([]int, 7, 7)
	for i := 0; i < len(numbers); i++ {
		numbers[i] = i * i
	}
	fmt.Println("Слайс до удаления элемента: ", numbers)
	i := 3
	copy(numbers[i:], numbers[i+1:])
	numbers = numbers[:len(numbers)-1]
	fmt.Println("Слайс после удаления элемента: ", numbers)
}
