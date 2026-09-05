package protocol

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

const markerPrefix = "__GOGIT_READY_"

// NewMarker creates a marker that identifies the shell prompt belonging to
// this Gogit process.
func NewMarker() (string, error) {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate shell marker: %w", err)
	}

	return markerPrefix + hex.EncodeToString(random) + "__", nil
}

// Scanner removes shell-ready markers from arbitrarily chunked PTY output.
type Scanner struct {
	marker  []byte
	pending []byte
}

// NewScanner creates a scanner for one shell session.
func NewScanner(marker string) *Scanner {
	if marker == "" {
		panic("protocol marker must not be empty")
	}

	return &Scanner{
		marker: []byte(marker),
	}
}

// Push accepts the next PTY output chunk.
//
// visible contains bytes that may be printed to the user's terminal.
// readyCount reports how many complete prompt markers were found.
func (s *Scanner) Push(data []byte) (visible []byte, readyCount int) {
	buffer := make([]byte, 0, len(s.pending)+len(data))
	buffer = append(buffer, s.pending...)
	buffer = append(buffer, data...)
	s.pending = nil

	for len(buffer) > 0 {
		index := bytes.Index(buffer, s.marker)
		if index >= 0 {
			visible = append(visible, buffer[:index]...)
			buffer = buffer[index+len(s.marker):]
			readyCount++
			continue
		}

		keep := matchingSuffixLength(buffer, s.marker)
		visible = append(visible, buffer[:len(buffer)-keep]...)

		if keep > 0 {
			s.pending = append(s.pending, buffer[len(buffer)-keep:]...)
		}
		break
	}

	return visible, readyCount
}

// Flush releases bytes that were temporarily held because they looked like
// the beginning of a marker.
func (s *Scanner) Flush() []byte {
	remaining := append([]byte(nil), s.pending...)
	s.pending = nil
	return remaining
}

// matchingSuffixLength returns the longest suffix of data that is also a
// prefix of marker.
func matchingSuffixLength(data, marker []byte) int {
	maxLength := min(len(data), len(marker)-1)

	for length := maxLength; length > 0; length-- {
		if bytes.Equal(
			data[len(data)-length:],
			marker[:length],
		) {
			return length
		}
	}

	return 0
}
