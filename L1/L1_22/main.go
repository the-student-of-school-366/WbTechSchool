package main

import (
	"fmt"
	"math/big"
)

func main() {
	fmt.Println("Введите два ОЧЕНЬ больших числа")
	var inp1, inp2 string
	fmt.Scanln(&inp1)
	fmt.Scanln(&inp2)
	a := new(big.Int)
	a.SetString(inp1, 10)
	b := new(big.Int)
	b.SetString(inp2, 10)
	result := new(big.Int)
	fmt.Println("Сумма двух чтсел:", result.Add(a, b))
	fmt.Println("Разность двух чтсел:", result.Sub(a, b))
	fmt.Println("Произведение двух чтсел:", result.Mul(a, b))
	if b.Sign() != 0 {
		fmt.Println("Частное двух чтсел:", result.Div(a, b))
	} else {
		fmt.Println("Бро, не дели на ноль, плиз")
	}
}
