package main

import (
	"fmt"
	"time"
)

func Sleep(seconds int) {
	timer := time.NewTimer(time.Second * time.Duration(seconds))
	<-timer.C
}

func main() {
	var seconds int
	fmt.Println("Введите количество секунд ожидания")
	fmt.Scan(&seconds)
	Sleep(seconds)
	fmt.Println("Мы проснулись!")
}
