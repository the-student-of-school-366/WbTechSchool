package main

import "fmt"

func binSearch(nums []int, x int) int {
	l := 0
	r := len(nums) - 1
	for l <= r {
		mid := (l + r) / 2
		if nums[mid] == x {
			return mid
		} else if nums[mid] > x {
			r = mid - 1
		} else if nums[mid] < x {
			l = mid + 1
		}
	}
	return -1
}
func main() {
	nums := []int{1, 3, 5, 7, 9}
	fmt.Println(binSearch(nums, 9))
	fmt.Println(binSearch(nums, 100))
}
