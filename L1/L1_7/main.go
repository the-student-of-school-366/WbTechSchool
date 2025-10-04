package main

import (
	"fmt"
	"sync"
	"time"
)

// worker принимает из канала ввода число, выводит его в консоль, имитирует работу
func worker(id int, jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs {
		fmt.Println("worker", id, "выел:", j)
		time.Sleep(time.Second)

	}
}

func main() {

}
