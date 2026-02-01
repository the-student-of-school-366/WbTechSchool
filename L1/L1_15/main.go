package main

import "strings"

var justString string

func createHugeString(n int) string {
	return strings.Repeat("a", n)

}
func someFunc() {
	v := createHugeString(1 << 10)
	justString = strings.Clone(v[:100])
}

func main() {
	someFunc()
}
