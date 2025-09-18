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

	wg := new(sync.WaitGroup)

	fmt.Println("Введите число горутин-воркеров:")
	var n int
	_, err := fmt.Scan(&n)
	if err != nil {
		fmt.Println(err)
	}
	//создаем буферизированный канал
	jobs := make(chan int, 100)

	//запускаем n воркеров
	for i := 0; i < n; i++ {
		wg.Add(1)
		go worker(i, jobs, wg)
	}

	//записываем данные в канал
	for i := 1; i <= n; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
}
