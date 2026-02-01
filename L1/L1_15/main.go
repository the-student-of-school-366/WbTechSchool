package main

import (
	"fmt"
	"strings"
)

var justString string

func createHugeString(n int) string {
	return strings.Repeat("a", n)

}
func someFunc() string {
	v := createHugeString(1 << 10)
	justString = strings.Clone(v[:100])
	return justString
}

func main() {
	someFunc()
	fmt.Println(justString)

}
