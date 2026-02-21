package main

import (
	"fmt"
	"os"

	"github.com/beevik/ntp"
)

func main() {
	currTime, err := ntp.Time("0.beevik-ntp.pool.ntp.org")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(currTime)
}
