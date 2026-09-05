//go:build !windows

package cmd

import (
	"os"
	"os/exec"

	"path/filepath"
)

func systemShell(marker string) *exec.Cmd {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	arguments := []string{"-i"}
	switch filepath.Base(shell) {
	case "bash":
		arguments = []string{"--norc", "-i"}
	case "zsh":
		arguments = []string{"-f", "-i"}
	}
	command := exec.Command(shell, arguments...)
	command.Env = append(os.Environ(), "PS1="+marker)
	return command
}
