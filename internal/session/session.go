package session

type ShellSession interface {
	Start() error
	Read(data []byte) (int, error)
	Write(data []byte) (int, error)
	Resize(width, height int) error
	Wait() error
	Close() error
}
