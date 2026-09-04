//go:build windows

package cmd

import "os/exec"

const promptScript = `
function global:prompt {
    "$([char]27)[36m(Gogit)$([char]27)[0m $($executionContext.SessionState.Path.CurrentLocation)> "
}
`

func systemShell() *exec.Cmd {
	if path, err := exec.LookPath("pwsh.exe"); err == nil {
		return exec.Command(
			path,
			"-NoLogo",
			"-NoProfile",
			"-NoExit",
			"-Command",
			promptScript,
		)
	}

	return exec.Command(
		"powershell.exe",
		"-NoLogo",
		"-NoProfile",
		"-NoExit",
		"-Command",
		promptScript,
	)
}
