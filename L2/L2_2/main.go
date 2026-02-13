package main

import "fmt"

func test() (x int) {
	defer func() {
		x++
	}()
	x = 1
	return //в начале x передается return, потом выполняется defer
}

func anotherTest() int {
	var x int
	defer func() {
		x++
	}()
	x = 1
	return x //КОПИРОВАНИЕ x для return происходит до выполнения defer,
}

func main() {
	fmt.Println(test())        // 2
	fmt.Println(anotherTest()) // 1
	//Вывод: 2 1
}
