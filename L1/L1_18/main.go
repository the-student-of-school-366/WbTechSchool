package main

import (
	"fmt"
	"sync"
)

type myCounter struct {
	value int
	mu    sync.RWMutex
}

func newMyCounter() *myCounter {
	return &myCounter{}
}

func main() {
	counter := newMyCounter()
	wg := sync.WaitGroup{}
	for i := 0; i < 20; i++ {
		wg.Add(1)
		fmt.Printf("Worker %d is running\n", i)
		go func() {
			defer wg.Done()
			counter.mu.Lock()
			counter.value++
			counter.mu.Unlock()
			fmt.Printf("Worker %d DONEEE\n", i)
		}()
	}
	wg.Wait()
	fmt.Printf("Counter is %d\n", counter.value)
}
