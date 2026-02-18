package main

import (
	"fmt"
)

func main() {
	var s = []string{"1", "2", "3"} //массив A
	modifySlice(s)
	fmt.Println(s)
}

func modifySlice(i []string) { //копируем структуру слайса: cap, len, pointer (указываем на массив A)
	i[0] = "3"         // меняем массив A
	i = append(i, "4") //меняет cap и len, из-за этого теперь i указывает на массив B
	i[1] = "5"         // меняем массив B
	i = append(i, "6") // меняем массив B
}
