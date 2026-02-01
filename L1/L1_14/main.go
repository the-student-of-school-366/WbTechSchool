package main

import "fmt"

func getType(v interface{}) {
	switch v.(type) {
	case int:
		fmt.Println("int", v)
	case string:
		fmt.Println("string", v)
	case bool:
		fmt.Println("bool", v)
	case chan int:
		fmt.Println("chan", v)
	}
}
func main() {
	integer := 2
	getType(integer)
	str := "It's over, Anakin, I have the high ground!"
	getType(str)
	bol := true
	getType(bol)
	chanInt := make(chan int)
	getType(chanInt)
}
