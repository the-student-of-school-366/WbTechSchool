package main

import "fmt"

func quicksort(arr []int, l, r int) {
	if l < r {
		q := partition(arr, l, r)
		quicksort(arr, l, q)
		quicksort(arr, q+1, r)
	}
}

func partition(arr []int, l, r int) int {
	i := l
	j := r
	v := arr[(l+r)/2]
	for i <= j {
		for arr[i] < v {
			i++
		}
		for arr[j] > v {
			j--
		}
		if i >= j {
			break
		}
		arr[i], arr[j] = arr[j], arr[i]
	}
	return j
}
func main() {
	arr := []int{8, 7, 5, 10, 9, 1, 3, 2, 4}
	quicksort(arr, 0, len(arr)-1)
	fmt.Println(arr)
}
