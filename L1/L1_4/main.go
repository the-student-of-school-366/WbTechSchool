package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func worker(ctx context.Context, wg *sync.WaitGroup) {
	fmt.Println("Before start")
	defer wg.Done()
	select {
	case <-ctx.Done():
		fmt.Println("Worker ended")
		return
	default:
		fmt.Println("Worker started")
		time.Sleep(1 * time.Second)
	}

}

func main() {

	// Создаем канал для получения сигналов
	fmt.Println("Program started")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	wg := new(sync.WaitGroup)

	ctx, cancel := context.WithCancel(context.Background())
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go worker(ctx, wg)
	}

	// Блокируемся до получения сигнала
	go func() {
		sig := <-sigChan
		fmt.Printf("\nПолучен сигнал: %v\n", sig)
		fmt.Println("Завершаю программу...")
		cancel()
	}()

	wg.Wait()
	fmt.Println("Program ended")
}
