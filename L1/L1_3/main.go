package main

import (
	"fmt"
	"strconv"
	"time"
)

// worker принимает из канала ввода число, выводит его в консоль, имитирует работу, и вывод сообщение в канал вывода
func worker(id int, jobs <-chan int, results chan<- string) {
	for j := range jobs {
		fmt.Println("worker", id, "выел:", j)
		time.Sleep(time.Second)
		results <- "Worker " + strconv.Itoa(id) + " закончил"
	}
}

func main() {

	//получаем n
	fmt.Println("Введите число горутин-воркеров:")
	var n int
	_, err := fmt.Scan(&n)
	if err != nil {
		fmt.Println(err)
	}
	//создаем буферизированные каналы ввода и вывода
	jobs := make(chan int, 100)
	results := make(chan string, 100)

	//запускаем n воркеров
	for i := 0; i < n; i++ {
		go worker(i, jobs, results)
	}

	//записываем данные в канал
	for i := 0; i < 2*n; i++ {
		jobs <- i
	}
	close(jobs)
	//выводим данные из канала
	for i := 0; i < 2*n; i++ {
		s := <-results
		fmt.Println(s)
	}
}
