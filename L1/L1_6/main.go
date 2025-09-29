package main

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// остановка через канал
func workerChannel(done <-chan bool) {
	for {
		select {
		default:
			fmt.Println("workerChannel running")
			time.Sleep(time.Second)
		case <-done:
			fmt.Println("workerChannel done")
			return
		}
	}
}

// остановка через контекст
func workerContext(ctx context.Context) {
	for {
		select {
		default:
			fmt.Println("workerContext running")
			time.Sleep(time.Second)
		case <-ctx.Done():
			fmt.Println("workerContext done")
			return
		}
	}
}

// остановка через контекст WithTimeout
func workerContextWithTimeOut(ctx context.Context) {
	for {
		select {
		default:
			fmt.Println("workerWithTimeOut running")
			time.Sleep(time.Second)
		case <-ctx.Done():
			fmt.Println("workerWithTimeOut done")
			return
		}
	}
}

// остановка через runtime.Goexit()
func workerGoExit() {
	for {
		defer fmt.Println("workerGoExit done")
		fmt.Println("workerGoExit running")
		time.Sleep(time.Second)
		runtime.Goexit()
	}
}

func main() {

	//остановка по условию
	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(100 * time.Millisecond)
			fmt.Println("WorkerCondition: ", i)
		}
		fmt.Println("WorkerCondition done")
	}()

	//остановка через канал
	done := make(chan bool)
	go workerChannel(done)
	time.Sleep(time.Second)
	done <- true
	time.Sleep(time.Second)

	//остановка через контекст
	ctx, cancel := context.WithCancel(context.Background())
	go workerContext(ctx)
	time.Sleep(time.Second)
	cancel()

	//остановка через ContextWithTimeOut
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	go workerContextWithTimeOut(ctx)
	time.Sleep(time.Second)

	// остановка через runtime.Goexit()
	go workerGoExit()

	time.Sleep(2 * time.Second)
	fmt.Println("Program finished")

}
