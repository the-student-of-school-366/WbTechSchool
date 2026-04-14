package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var (
	runningMu   sync.Mutex
	runningCmds []*exec.Cmd
)

func setRunningCmds(cmds []*exec.Cmd) {
	runningMu.Lock()
	defer runningMu.Unlock()
	runningCmds = cmds
}

func clearRunningCmds() {
	runningMu.Lock()
	defer runningMu.Unlock()
	runningCmds = nil
}

func interruptRunningCommands() {
	runningMu.Lock()
	defer runningMu.Unlock()

	for _, cmd := range runningCmds {
		if cmd == nil || cmd.Process == nil {
			continue
		}
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			_ = cmd.Process.Kill()
		}
	}
}

func executePipeline(line string) error {
	parts := strings.Split(line, "|")
	commands := make([][]string, 0, len(parts))
	for _, part := range parts {
		args := strings.Fields(strings.TrimSpace(part))
		if len(args) == 0 {
			return errors.New("empty command in pipeline")
		}
		commands = append(commands, args)
	}

	if len(commands) == 1 {
		if ok, err := runBuiltin(commands[0], os.Stdout, true); ok {
			return err
		}
		cmd := exec.Command(commands[0][0], commands[0][1:]...)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		setRunningCmds([]*exec.Cmd{cmd})
		defer clearRunningCmds()
		return cmd.Run()
	}

	cmds := make([]*exec.Cmd, len(commands))
	for i, args := range commands {
		if ok, _ := runBuiltin(args, os.Stdout, false); ok {
			return errors.New("builtins cannot be used inside pipelines")
		}
		cmds[i] = exec.Command(args[0], args[1:]...)
	}

	for i := 0; i < len(cmds)-1; i++ {
		pipe, err := cmds[i].StdoutPipe()
		if err != nil {
			return err
		}
		cmds[i+1].Stdin = pipe
	}

	cmds[0].Stdin = os.Stdin
	cmds[len(cmds)-1].Stdout = os.Stdout
	for _, cmd := range cmds {
		cmd.Stderr = os.Stderr
	}

	setRunningCmds(cmds)
	defer clearRunningCmds()

	for _, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			return err
		}
	}

	var firstErr error
	for _, cmd := range cmds {
		if err := cmd.Wait(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
