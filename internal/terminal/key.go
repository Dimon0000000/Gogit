package terminal

// KeyType identifies a decoded terminal key.
type KeyType uint8

const (
	KeyRune KeyType = iota
	KeyEnter
	KeyBackspace
	KeyTab
	KeyEscape
	KeyCtrlC
	KeyCtrlD
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyDelete
)

// Key is one logical keyboard event. Rune is set only for KeyRune.
type Key struct {
	Type KeyType
	Rune rune
}
