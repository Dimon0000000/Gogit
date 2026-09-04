package session

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sync"

	"github.com/charmbracelet/x/xpty"
)

var (
	errAlreadyStarted = errors.New("session has already started")
	errNotStarted     = errors.New("session has not started")
)

var _ ShellSession = (*ptySession)(nil)

type ptySession struct {
	pty    xpty.Pty
	cmd    *exec.Cmd
	ctx    context.Context
	cancel context.CancelFunc

	width  int
	height int

	closeOnce sync.Once
	closeErr  error
}

func New(cmd *exec.Cmd, width, height int) ShellSession {
	ctx, cancel := context.WithCancel(context.Background())

	return &ptySession{
		cmd:    cmd,
		ctx:    ctx,
		cancel: cancel,
		width:  width,
		height: height,
	}
}

func (p *ptySession) Start() error {
	if p.cmd == nil {
		return errors.New("shell command is nil")
	}

	if p.pty != nil {
		return errAlreadyStarted
	}

	terminal, err := xpty.NewPty(p.width, p.height)
	if err != nil {
		return fmt.Errorf("could not start pty: %w", err)
	}

	if err := terminal.Start(p.cmd); err != nil {
		_ = terminal.Close()
		return fmt.Errorf("could not start pty: %w", err)
	}

	p.pty = terminal
	return nil
}

func (p *ptySession) Read(data []byte) (int, error) {
	if p.pty == nil {
		return 0, errNotStarted
	}
	return p.pty.Read(data)
}

func (p *ptySession) Write(data []byte) (int, error) {
	if p.pty == nil {
		return 0, errNotStarted
	}
	return p.pty.Write(data)
}

func (p *ptySession) Resize(width, height int) error {
	if p.pty == nil {
		return errNotStarted
	}

	if err := p.pty.Resize(width, height); err != nil {
		return fmt.Errorf("could not resize pty: %w", err)
	}
	return nil
}

func (p *ptySession) Wait() error {
	if p.pty == nil || p.cmd == nil || p.cmd.Process == nil {
		return errNotStarted
	}

	if err := xpty.WaitProcess(p.ctx, p.cmd); err != nil {
		return fmt.Errorf("wait for shell: %w", err)
	}
	return nil
}

func (p *ptySession) Close() error {
	p.closeOnce.Do(func() {
		p.cancel()

		if p.pty != nil {
			p.closeErr = p.pty.Close()
		}
	})

	return p.closeErr
}
