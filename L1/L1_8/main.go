package main

import "fmt"

func setBit(num int64, n int64) int64 {
	return num | (1 << n)
}
func clearBit(num int64, n int64) int64 {
	return num & ^(1 << n)
}
func main() {
	fmt.Println("Введите число и номер бита")
	var num int64
	var n int64
	_, err := fmt.Scan(&num, &n)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Если вы хотите установить бит в 1. введите 1, если в 0, введите 0")
	var flag bool
	_, err = fmt.Scan(&flag)
	if err != nil {
		fmt.Println(err)
	}
	if flag {
		ans := setBit(num, n)
		fmt.Println(ans)
	} else {
		ans := clearBit(num, n)
		fmt.Println(ans)
	}
}
