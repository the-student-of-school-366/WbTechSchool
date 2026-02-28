package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

func main() {
	var text string
	sc := bufio.NewScanner(os.Stdin)
	if sc.Scan() {
		text = sc.Text()
	}
	input := strings.Split(text, " ")

	tmpWords := make(map[string][]string)
	for _, v := range input {
		v = strings.ToLower(v)
		tmpWords[sortString(v)] = append(tmpWords[sortString(v)], v)
	}

	finalWords := make(map[string][]string)
	foundKeys := make(map[string]string)
	for _, v := range input {
		v = strings.ToLower(v)
		_, ok := foundKeys[sortString(v)]
		if !ok {
			finalWords[v] = append(finalWords[v], v)
			foundKeys[sortString(v)] = v
		} else {
			key := foundKeys[sortString(v)]
			finalWords[key] = append(finalWords[key], v)
		}
	}
	for k, v := range finalWords {
		if len(v) > 2 {
			fmt.Print(k, ": ")
			fmt.Println(v)
		}
	}

}

func sortString(s string) string {
	charArr := []rune(s)
	sort.Slice(charArr, func(i, j int) bool {
		return charArr[i] < charArr[j]
	})
	return string(charArr)
}
