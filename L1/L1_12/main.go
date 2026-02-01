package main

import "fmt"

func main() {
	animals := []string{"cat", "cat", "dog", "cat", "tree"}
	map1 := make(map[string]int)
	for _, v := range animals {
		map1[v]++
	}
	for k := range map1 {
		fmt.Println(k)
	}
}
