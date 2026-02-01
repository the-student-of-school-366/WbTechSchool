package main

import "fmt"

func main() {
	tempreture := make(map[int][]float32)
	var n int
	_, err := fmt.Scan(&n)
	if err != nil {
		fmt.Println(err)
	}
	for i := 0; i < n; i++ {
		var num float32
		_, err = fmt.Scan(&num)
		if err != nil {
			fmt.Println(err)
		}
		base := int(num / 1)
		base = base - base%10
		tempreture[base] = append(tempreture[base], num)
	}
	for k, v := range tempreture {
		fmt.Println("Group of: ", k)
		for _, j := range v {
			fmt.Print(j, "\t")
		}
		fmt.Println()
	}

}
