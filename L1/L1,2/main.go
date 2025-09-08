package main

import (
	"fmt"
	"sync"
)

// sqr функция для вычисления квадрата числа
func sqr(x int, wg *sync.WaitGroup) {
	defer wg.Done() // defer выполнится в конце функции. Уменьшаем счетчик на 1
	fmt.Println(x * x)
}

func main() {
	//создаем waitGroup чтобы дождаться пока все горутины выполнятся
	var wg sync.WaitGroup
	arr := [5]int{2, 4, 6, 8, 10}
	for _, num := range arr {
		wg.Add(1) //Увеличиваем счетчик на 1
		go sqr(num, &wg)
	}
	wg.Wait() // ждем пока счетчик не станет равным 0
}
