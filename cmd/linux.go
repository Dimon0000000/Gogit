//go:build !windows

package cmd

import (
	"os"
	"os/exec"
)

func systemShell(marker string) *exec.Cmd {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	command := exec.Command(shell, "-i")
	command.Env = append(os.Environ(), "PS1="+marker)
	return command
}
