//go:build windows

package cmd

import (
	"fmt"
	"os/exec"
)

func systemShell(marker string) *exec.Cmd {
	script := fmt.Sprintf(`function global:prompt {'%s'}`, marker)

	if path, err := exec.LookPath("pwsh.exe"); err == nil {
		return exec.Command(
			path,
			"-NoLogo",
			"-NoProfile",
			"-NoExit",
			"-Command",
			script,
		)
	}

	return exec.Command(
		"powershell.exe",
		"-NoLogo",
		"-NoProfile",
		"-NoExit",
		"-Command",
		script,
	)
}
