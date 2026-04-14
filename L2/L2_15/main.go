package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)
	defer signal.Stop(sigCh)

	go func() {
		for range sigCh {
			interruptRunningCommands()
			fmt.Println()
			fmt.Print("$ ")
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("$ ")

		if !scanner.Scan() {
			fmt.Println("exit")
			return
		}

		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}

		if err := executePipeline(text); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
