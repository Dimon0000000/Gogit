//go:build windows

package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/Haruko386/Gogit/internal/session"
	"github.com/charmbracelet/x/term"
)

func watchTerminalResize(
	ctx context.Context,
	outputFD uintptr,
	shellSession session.ShellSession,
	width, height int,
) <-chan error {
	done := make(chan error, 1)

	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				done <- nil
				return
			case <-ticker.C:
				newWidth, newHeight, err := term.GetSize(outputFD)
				if err != nil {
					done <- fmt.Errorf("get terminal size: %w", err)
					return
				}
				if newWidth == width && newHeight == height {
					continue
				}
				if err := shellSession.Resize(newWidth, newHeight); err != nil {
					done <- err
					return
				}
				width, height = newWidth, newHeight
			}
		}
	}()

	return done
}
