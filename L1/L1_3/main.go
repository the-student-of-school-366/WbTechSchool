package main

import (
	"fmt"
	"strconv"
	"time"
)

func worker(id int, jobs <-chan int, results chan<- string) {
	for j := range jobs {
		fmt.Println("worker", id, "выел:", j)
		time.Sleep(time.Second)
		results <- "Worker " + strconv.Itoa(id) + " закончил"
	}
}

func main() {
	fmt.Println("Введите число горутин-воркеров:")
	var n int
	_, err := fmt.Scan(&n)
	if err != nil {
		fmt.Println(err)
	}
	jobs := make(chan int, 100)
	results := make(chan string, 100)

	for i := 0; i < n; i++ {
		go worker(i, jobs, results)
	}

	for i := 0; i < 2*n; i++ {
		jobs <- i
	}

	for i := 0; i < n; i++ {
		s := <-results
		fmt.Println(s)
	}
}
