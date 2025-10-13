package main

import (
	"context"
	"fmt"
	"time"
)

const N = 5

// worker отправляет в канал числа, ждем сигнала завершения от Context
func worker(ctx context.Context, dataChan chan<- int) {
	defer close(dataChan)
	i := 0
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Получен сигнал завершения работы.")
			return
		case dataChan <- i:
			fmt.Printf("Воркер отправил %d\n", i)
			i++
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func main() {
	fmt.Printf("Программа будет работать %d секунд.\n\n", N)

	// Создаем канал для обмена данными.
	dataChan := make(chan int)

	//Контекст, который автоматически завершится через 5 секунд
	ctx, cancel := context.WithTimeout(context.Background(), N*time.Second)
	defer cancel()

	// Запускаем  горутину.
	go worker(ctx, dataChan)

	for value := range dataChan {
		fmt.Printf("Читатель: получил %d\n", value)
	}
}
