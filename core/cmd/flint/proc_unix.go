//go:build !windows

package main

import (
	"os"
	"os/exec"
	"syscall"
)

func detachProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

func killProcess(proc *os.Process) error {
	return proc.Signal(syscall.SIGTERM)
}
