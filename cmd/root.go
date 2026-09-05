package cmd

import (
	"context"
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

func Run() error {
	return runPersistentShell()
}

func runPersistentShell() (resultErr error) {
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
		resultErr = errors.Join(resultErr, shellSession.Close())
	}()

	oldState, err := term.MakeRaw(inputFD)
	if err != nil {
		return fmt.Errorf("enable terminal raw mode: %w", err)
	}

	defer func() {
		resultErr = errors.Join(
			resultErr,
			term.Restore(inputFD, oldState),
		)
	}()

	resizeCtx, stopResize := context.WithCancel(context.Background())
	resizeDone := watchTerminalResize(resizeCtx, outputFD, shellSession, width, height)

	outputDone := make(chan error, 1)
	inputDone := make(chan error, 1)
	shellDone := make(chan error, 1)

	go func() {
		_, err := io.Copy(os.Stdout, shellSession)
		outputDone <- err
	}()

	go func() {
		_, err := io.Copy(shellSession, os.Stdin)
		inputDone <- err
	}()

	go func() {
		shellDone <- shellSession.Wait()
	}()

	var (
		runErr         error
		shellFinished  bool
		outputFinished bool
		inputFinished  bool
		resizeFinished bool
	)

	select {
	case runErr = <-shellDone:
		shellFinished = true
	case err := <-outputDone:
		outputFinished = true
		if err != nil {
			runErr = fmt.Errorf("copy PTY output: %w", err)
		}
	case err := <-inputDone:
		inputFinished = true
		if err != nil {
			runErr = fmt.Errorf("copy terminal input: %w", err)
		}
	case err := <-resizeDone:
		resizeFinished = true
		if err != nil {
			runErr = fmt.Errorf("watch terminal resize: %w", err)
		}
	}

	// Capture any I/O failure that was already reported before shutdown began.
	if !outputFinished {
		select {
		case err := <-outputDone:
			outputFinished = true
			if err != nil {
				runErr = errors.Join(runErr, fmt.Errorf("copy PTY output: %w", err))
			}
		default:
		}
	}
	if !inputFinished {
		select {
		case err := <-inputDone:
			inputFinished = true
			if err != nil {
				runErr = errors.Join(runErr, fmt.Errorf("copy terminal input: %w", err))
			}
		default:
		}
	}

	stopResize()
	if !resizeFinished {
		<-resizeDone
	}

	// Stop the resize watcher before closing the session so Resize and Close
	// cannot operate on the PTY concurrently.
	_ = shellSession.Close()

	if !shellFinished {
		// Close terminates and reaps the process; this receive only drains the
		// result produced by the dedicated waiter.
		<-shellDone
	}
	if !outputFinished {
		// The deliberate PTY close releases the output reader. Any error first
		// reported after this point is part of normal shutdown.
		shutdownOutputErr := <-outputDone
		_ = shutdownOutputErr
	}

	return runErr
}
