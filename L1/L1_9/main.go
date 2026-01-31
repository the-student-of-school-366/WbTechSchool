package main

import "fmt"

func generator(numbers []int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range numbers {
			out <- n
		}
	}()
	return out
}

func multplier(inp <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range inp {
			out <- n * n
		}
	}()
	return out
}
func main() {
	numbers := []int{1, 2, 3, 4, 5, 6, 7}
	stage1 := generator(numbers)
	stage2 := multplier(stage1)
	for n := range stage2 {
		fmt.Println(n)
	}
}
