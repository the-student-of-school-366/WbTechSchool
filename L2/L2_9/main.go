package main

import (
	"fmt"
	"strings"
)

func multString(inp string) (string, error) {
	if inp == "" {
		return "", nil
	}
	var out strings.Builder
	var prev rune = rune(inp[0])
	hasPrev := false
	for i, v := range inp {
		if v >= 48 && v <= 57 {
			if prev == '\\' {
				if i == len(inp)-1 {
					out.WriteString(string(v))
				}
				hasPrev = true
			} else {
				if !hasPrev {
					return "", fmt.Errorf("две цифры не могут идти подряд")
				}
				out.WriteString(strings.Repeat(string(prev), int(v)-'0'))
				hasPrev = false
			}
		} else {
			if hasPrev {
				out.WriteString(string(prev))
			}
			if i == len(inp)-1 {
				out.WriteString(string(v))
			}
			hasPrev = true
		}
		prev = v
	}
	return out.String(), nil
}

func main() {
	var inp string
	fmt.Scanln(&inp)
	out, err := multString(inp)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(out)
	}
}
