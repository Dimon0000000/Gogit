package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/Haruko386/Gogit/internal/editor"
	"github.com/Haruko386/Gogit/internal/history"
	"github.com/Haruko386/Gogit/internal/protocol"
	"github.com/Haruko386/Gogit/internal/session"
	"github.com/Haruko386/Gogit/internal/suggest"
	"github.com/Haruko386/Gogit/internal/terminal"
)

const (
	gogitPrompt      = "\x1b[36m(Gogit)\x1b[0m "
	gogitPromptWidth = 8
)

type streamEvent struct {
	data []byte
	err  error
}

func runShellUI(shellSession session.ShellSession, marker string, resizeDone <-chan error) (resizeFinished bool, resultErr error) {
	inputEvents := readStream(os.Stdin)
	outputEvents := readStream(shellSession)
	shellDone := make(chan error, 1)

	go func() {
		shellDone <- shellSession.Wait()
	}()

	var (
		decoder        terminal.Decoder
		lineEditor     editor.Editor
		commandHistory history.History
		renderer       terminal.Renderer
		markerScan     = protocol.NewScanner(marker)
		selected       int
		editing        bool
	)

	render := func() error {
		suggestions := suggest.Suggest(
			lineEditor.Line(),
			lineEditor.Cursor(),
		)

		if selected < 0 || selected >= len(suggestions) {
			selected = 0
		}

		return writeOutput(renderer.Render(terminal.View{
			Prompt:      gogitPrompt,
			PromptWidth: gogitPromptWidth,
			Line:        lineEditor.Line(),
			Cursor:      lineEditor.Cursor(),
			Suggestions: suggestions,
			Selected:    selected,
		}))
	}

	for {
		select {
		case event := <-inputEvents:
			if event.err != nil {
				if errors.Is(event.err, io.EOF) {
					return false, nil
				}
				return false, fmt.Errorf(
					"read terminal input: %w",
					event.err,
				)
			}

			if !editing {
				if err := writeAll(shellSession, event.data); err != nil {
					return false, fmt.Errorf(
						"write PTY input: %w",
						err,
					)
				}
				continue
			}

			keys := decoder.Feed(event.data)
			changed := false

		keyLoop:
			for _, key := range keys {
				switch key.Type {
				case terminal.KeyRune:
					lineEditor.Insert(key.Rune)
					selected = 0
					changed = true

				case terminal.KeyBackspace:
					changed = lineEditor.Backspace() || changed
					selected = 0

				case terminal.KeyDelete:
					changed = lineEditor.Delete() || changed
					selected = 0

				case terminal.KeyLeft:
					changed = lineEditor.MoveLeft() || changed
					selected = 0

				case terminal.KeyRight:
					changed = lineEditor.MoveRight() || changed
					selected = 0

				case terminal.KeyHome:
					changed = lineEditor.MoveHome() || changed
					selected = 0

				case terminal.KeyEnd:
					changed = lineEditor.MoveEnd() || changed
					selected = 0

				case terminal.KeyUp:
					// FIXME: 选择历史指令和选择Tab填充时有冲突(我目前觉得只有空白时才能选择历史；当然后续这个需要复杂的设计)
					suggestions := suggest.Suggest(
						lineEditor.Line(),
						lineEditor.Cursor(),
					)
					if len(suggestions) > 0 {
						selected--
						if selected < 0 {
							selected = len(suggestions) - 1
						}
						changed = true
					}

					command, ok := commandHistory.Previous(lineEditor.Line())
					if ok {
						lineEditor.SetLine(command)
						selected = 0
						changed = true
					}

				case terminal.KeyDown:
					suggestions := suggest.Suggest(
						lineEditor.Line(),
						lineEditor.Cursor(),
					)
					if len(suggestions) > 0 {
						selected++
						if selected >= len(suggestions) {
							selected = 0
						}
						changed = true
					}

					command, ok := commandHistory.Next()
					if ok {
						lineEditor.SetLine(command)
						selected = 0
						changed = true
					}

				case terminal.KeyTab:
					suggestions := suggest.Suggest(
						lineEditor.Line(),
						lineEditor.Cursor(),
					)
					if len(suggestions) == 0 {
						continue
					}
					if selected >= len(suggestions) {
						selected = 0
					}

					context, ok := suggest.ParseContext(
						lineEditor.Line(),
						lineEditor.Cursor(),
					)
					if !ok {
						continue
					}

					lineEditor.Replace(
						context.TokenStart,
						context.TokenEnd,
						suggestions[selected].Value,
					)
					selected = 0
					changed = true

				case terminal.KeyCtrlC:
					if err := writeOutput(renderer.Clear()); err != nil {
						return false, err
					}
					if err := writeOutput("^C\r\n"); err != nil {
						return false, err
					}

					lineEditor.Clear()
					selected = 0
					changed = true

				case terminal.KeyCtrlD:
					if lineEditor.Line() == "" {
						if err := writeOutput(renderer.Clear()); err != nil {
							return false, err
						}
						return false, nil
					}

					changed = lineEditor.Delete() || changed

				case terminal.KeyEnter:
					if err := writeOutput(renderer.Clear()); err != nil {
						return false, err
					}

					command := lineEditor.Line()
					// save command to history
					commandHistory.Add(command)
					command += "\r"

					if err := writeAll(
						shellSession,
						[]byte(command),
					); err != nil {
						return false, fmt.Errorf(
							"submit command: %w",
							err,
						)
					}

					lineEditor.Clear()
					commandHistory.Reset()
					selected = 0
					editing = false
					changed = false
					break keyLoop
				}
			}

			if editing && changed {
				if err := render(); err != nil {
					return false, err
				}
			}

		case event := <-outputEvents:
			if len(event.data) > 0 {
				visible, readyCount := markerScan.Push(event.data)

				if len(visible) > 0 && editing {
					if err := writeOutput(renderer.Clear()); err != nil {
						return false, err
					}
				}

				if len(visible) > 0 {
					if err := writeOutput(string(visible)); err != nil {
						return false, err
					}
				}

				if readyCount > 0 {
					editing = true
					lineEditor.Clear()
					selected = 0
				}

				if editing && (len(visible) > 0 || readyCount > 0) {
					if err := render(); err != nil {
						return false, err
					}
				}
			}

			if event.err != nil {
				remaining := markerScan.Flush()
				if len(remaining) > 0 {
					if err := writeOutput(string(remaining)); err != nil {
						return false, err
					}
				}

				if errors.Is(event.err, io.EOF) {
					outputEvents = nil
					continue
				}

				return false, fmt.Errorf(
					"read PTY output: %w",
					event.err,
				)
			}

		case err := <-shellDone:
			return false, err

		case err := <-resizeDone:
			if err != nil {
				return true, fmt.Errorf(
					"watch terminal resize: %w",
					err,
				)
			}
			return true, nil
		}
	}
}

func readStream(reader io.Reader) <-chan streamEvent {
	events := make(chan streamEvent, 1)

	go func() {
		buffer := make([]byte, 4096)

		for {
			count, err := reader.Read(buffer)

			if count > 0 {
				data := append([]byte(nil), buffer[:count]...)
				events <- streamEvent{data: data}
			}

			if err != nil {
				events <- streamEvent{err: err}
				return
			}
		}
	}()

	return events
}

func writeAll(writer io.Writer, data []byte) error {
	for len(data) > 0 {
		count, err := writer.Write(data)
		if err != nil {
			return err
		}
		if count == 0 {
			return io.ErrShortWrite
		}

		data = data[count:]
	}

	return nil
}

func writeOutput(value string) error {
	if err := writeAll(os.Stdout, []byte(value)); err != nil {
		return fmt.Errorf("write terminal output: %w", err)
	}
	return nil
}
