package main

import "fmt"

func main() {
	var a, b int
	fmt.Scan(&a, &b)
	a = a + b
	b = a - b
	a = a - b
	fmt.Println(a, b)
}
