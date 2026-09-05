package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/charmbracelet/x/xpty"
)

var (
	errAlreadyStarted = errors.New("session has already started")
	errNotStarted     = errors.New("session has not started")
	errSessionClosed  = errors.New("session has been closed")
)

var _ ShellSession = (*ptySession)(nil)

type ptySession struct {
	mu      sync.RWMutex
	pty     xpty.Pty
	cmd     *exec.Cmd
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
	closed  bool

	width  int
	height int

	closeOnce sync.Once
	closeErr  error
	waitOnce  sync.Once
	waitErr   error
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
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return errSessionClosed
	}

	if p.cmd == nil {
		return errors.New("shell command is nil")
	}

	if p.started {
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
	p.started = true
	return nil
}

func (p *ptySession) Read(data []byte) (int, error) {
	p.mu.RLock()
	terminal := p.pty
	started := p.started
	p.mu.RUnlock()

	if !started || terminal == nil {
		return 0, errNotStarted
	}
	return terminal.Read(data)
}

func (p *ptySession) Write(data []byte) (int, error) {
	p.mu.RLock()
	terminal := p.pty
	started := p.started
	p.mu.RUnlock()

	if !started || terminal == nil {
		return 0, errNotStarted
	}
	return terminal.Write(data)
}

func (p *ptySession) Resize(width, height int) error {
	p.mu.RLock()
	terminal := p.pty
	started := p.started
	closed := p.closed
	p.mu.RUnlock()

	if closed {
		return errSessionClosed
	}
	if !started || terminal == nil {
		return errNotStarted
	}

	if err := terminal.Resize(width, height); err != nil {
		return fmt.Errorf("could not resize pty: %w", err)
	}
	return nil
}

func (p *ptySession) Wait() error {
	p.mu.RLock()
	started := p.started
	cmd := p.cmd
	ctx := p.ctx
	p.mu.RUnlock()

	if !started || cmd == nil || cmd.Process == nil {
		return errNotStarted
	}

	p.waitOnce.Do(func() {
		if err := xpty.WaitProcess(ctx, cmd); err != nil {
			p.waitErr = fmt.Errorf("wait for shell: %w", err)
		}
	})

	return p.waitErr
}

func (p *ptySession) Close() error {
	p.closeOnce.Do(func() {
		p.mu.Lock()
		p.closed = true
		started := p.started
		terminal := p.pty
		cmd := p.cmd
		cancel := p.cancel
		p.mu.Unlock()

		var killErr error
		if started && cmd != nil && cmd.Process != nil {
			killErr = cmd.Process.Kill()
			if errors.Is(killErr, os.ErrProcessDone) {
				killErr = nil
			}
		}

		cancel()

		if started {
			// Wait exactly once so the child is reaped on every close path.
			_ = p.Wait()
		}

		var terminalErr error
		if terminal != nil {
			terminalErr = terminal.Close()
		}

		p.closeErr = errors.Join(killErr, terminalErr)
	})

	return p.closeErr
}
