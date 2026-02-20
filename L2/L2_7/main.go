package main

import (
	"fmt"
	"math/rand"
	"time"
)

// asChan получает на вход произвольное количество чисел и асинхронно записывает их в канал
func asChan(vs ...int) <-chan int {
	c := make(chan int)
	go func() {
		for _, v := range vs {
			c <- v
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		}
		close(c)
	}()
	return c
}

// merge асинхронно объединяет данные из двух каналов в один
func merge(a, b <-chan int) <-chan int {
	c := make(chan int)
	go func() {
		for {
			select {
			case v, ok := <-a: // пытаемся прочесть число из канала a
				if ok { // если канал не закрыт, записываем число в канал c
					c <- v
				} else { // иначе помечаем канал a как nil, в смысле закрыт
					a = nil
				}
			case v, ok := <-b: // пытаемся прочесть число из канала b
				if ok { // если канал не закрыт, записываем число в канал c
					c <- v
				} else { // иначе помечаем канал b как nil, в смысле закрыт
					b = nil
				}
			}
			if a == nil && b == nil { // если оба входных канала закрыты, закрываем канал c
				close(c)
				return
			}
		}
	}()
	return c
}

func main() {
	rand.Seed(time.Now().Unix())
	a := asChan(1, 3, 5, 7)
	b := asChan(2, 4, 6, 8)
	c := merge(a, b)
	for v := range c {
		fmt.Print(v)
	}
}
