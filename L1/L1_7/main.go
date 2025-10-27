package main

import (
	"fmt"
	"sync"
	"time"
)

type myMap struct {
	mu sync.RWMutex
	m  map[int]int
}

func NewMyMap() *myMap {
	return &myMap{m: make(map[int]int)}
}

func (myMap *myMap) Set(key int, value int) {
	myMap.mu.Lock()
	myMap.m[key] = value
	myMap.mu.Unlock()
}
func (myMap *myMap) Get(key int) int {
	myMap.mu.RLock()
	defer myMap.mu.RUnlock()
	return myMap.m[key]
}

func main() {
	mapa := NewMyMap()
	mapa.Set(1, 1)
	wg := sync.WaitGroup{}
	for i := 0; i < 15; i++ {
		wg.Add(1)
		go func(i int) {
			mapa.Set(i, i)
			time.Sleep(time.Millisecond * 100)
			value := mapa.Get(i)
			fmt.Printf("Goroutine %d: got value %d\n", i, value)
			wg.Done()
		}(i)
	}
	wg.Wait()
	fmt.Println("Completed")
}
