//go:build !windows

package cmd

import (
	"os"
	"os/exec"
)

func systemShell() *exec.Cmd {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	return exec.Command(shell, "-i")
}
