package main

import "fmt"

func main() {
	arr1 := []int{1, 3, 5, 7, 9}
	arr2 := []int{0, 1, 2, 3, 4, 5}
	map1 := make(map[int]int)
	var ans []int
	for _, v := range arr1 {
		map1[v]++
	}
	for _, v := range arr2 {
		if map1[v] > 0 {
			ans = append(ans, v)
		}
	}
	fmt.Println(ans)
}
