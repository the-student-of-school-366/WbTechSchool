package main

import (
	"fmt"
	"sync"
	"time"
)

func or(channels ...<-chan any) <-chan any {
	switch len(channels) {
	case 0:
		c := make(chan any)
		close(c)
		return c
	case 1:
		return channels[0]
	}

	orDone := make(chan any)
	var once sync.Once

	go func() {
		var wg sync.WaitGroup
		wg.Add(len(channels))

		for _, c := range channels {
			go func(ch <-chan any) {
				defer wg.Done()
				<-ch
				once.Do(func() {
					close(orDone)
				})
			}(c)
		}
		wg.Wait()
	}()

	return orDone
}

func main() {
	sig := func(after time.Duration) <-chan any {
		c := make(chan any)
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}

	start := time.Now()
	<-or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)
	fmt.Printf("done after %v", time.Since(start))

}
