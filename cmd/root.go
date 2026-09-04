package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/Haruko386/Gogit/internal/session"
	"github.com/charmbracelet/x/term"
)

const (
	fallbackTerminalWidth  = 80
	fallbackTerminalHeight = 24
)

func Run() {
	if err := runPersistentShell(); err != nil {
		fmt.Fprintf(os.Stderr, "Gogit: %v\n", err)
	}
}

func runPersistentShell() error {
	inputFD := os.Stdin.Fd()
	outputFD := os.Stdout.Fd()

	if !term.IsTerminal(inputFD) {
		return errors.New("standard input is not a terminal")
	}

	width, height, err := term.GetSize(outputFD)
	if err != nil {
		width = fallbackTerminalWidth
		height = fallbackTerminalHeight
	}

	shellSession := session.New(
		systemShell(),
		width,
		height,
	)

	if err := shellSession.Start(); err != nil {
		return fmt.Errorf("start shell: %w", err)
	}

	defer func() {
		_ = shellSession.Close()
	}()

	oldState, err := term.MakeRaw(inputFD)
	if err != nil {
		return fmt.Errorf("enable terminal raw mode: %w", err)
	}

	defer func() {
		_ = term.Restore(inputFD, oldState)
	}()

	outputDone := make(chan struct{})

	go func() {
		defer close(outputDone)
		_, _ = io.Copy(os.Stdout, shellSession)
	}()

	go func() {
		_, _ = io.Copy(shellSession, os.Stdin)
	}()

	waitErr := shellSession.Wait()
	closeErr := shellSession.Close()

	<-outputDone

	return errors.Join(waitErr, closeErr)
}
