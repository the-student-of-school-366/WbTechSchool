package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

func echo(args []string, out io.Writer) {
	fmt.Fprintln(out, strings.Join(args, " "))
}

func cd(path string) error {
	if path == "" {
		path = os.Getenv("HOME")
		if path == "" {
			path = os.Getenv("USERPROFILE")
		}
		if path == "" {
			return errors.New("cd: path is required")
		}
	}
	return os.Chdir(filepath.Clean(path))
}

func kill(args []string) error {
	if len(args) < 2 {
		return errors.New("kill: pid is required")
	}

	pid, err := strconv.Atoi(args[1])
	if err != nil {
		return err
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Signal(os.Kill)
}

func pwd(out io.Writer) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Fprintln(out, dir)
	return nil
}

func ps(out io.Writer) error {
	processes, err := process.Processes()
	if err != nil {
		return err
	}

	for _, proc := range processes {
		name, nameErr := proc.Name()
		if nameErr != nil {
			name = "<unknown>"
		}
		fmt.Fprintf(out, "PID: %d, Name: %s\n", proc.Pid, name)
	}

	return nil
}

func runBuiltin(args []string, out io.Writer, allowStateChange bool) (bool, error) {
	if len(args) == 0 {
		return false, nil
	}

	switch args[0] {
	case "cd":
		if !allowStateChange {
			return true, errors.New("cd cannot be used in a pipeline")
		}
		path := ""
		if len(args) > 1 {
			path = args[1]
		}
		return true, cd(path)
	case "pwd":
		return true, pwd(out)
	case "echo":
		echo(args[1:], out)
		return true, nil
	case "kill":
		return true, kill(args)
	case "ps":
		return true, ps(out)
	case "exit":
		if !allowStateChange {
			return true, errors.New("exit cannot be used in a pipeline")
		}
		os.Exit(0)
	}
	return false, nil
}
