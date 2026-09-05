//go:build !windows

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Haruko386/Gogit/internal/session"
	"github.com/charmbracelet/x/term"
)

func watchTerminalResize(
	ctx context.Context,
	outputFD uintptr,
	shellSession session.ShellSession,
	_, _ int,
) <-chan error {
	done := make(chan error, 1)

	go func() {
		resizeSignals := make(chan os.Signal, 1)
		signal.Notify(resizeSignals, syscall.SIGWINCH)
		defer signal.Stop(resizeSignals)

		for {
			select {
			case <-ctx.Done():
				done <- nil
				return
			case <-resizeSignals:
				width, height, err := term.GetSize(outputFD)
				if err != nil {
					done <- fmt.Errorf("get terminal size: %w", err)
					return
				}
				if err := shellSession.Resize(width, height); err != nil {
					done <- err
					return
				}
			}
		}
	}()

	return done
}
