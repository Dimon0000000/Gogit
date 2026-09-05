package protocol

import "testing"

func TestScannerFindsSplitMarker(t *testing.T) {
	scanner := NewScanner("__READY__")

	visible, readyCount := scanner.Push([]byte("command output\r\n__REA"))
	if string(visible) != "command output\r\n" {
		t.Fatalf("first visible output = %q", visible)
	}
	if readyCount != 0 {
		t.Fatalf("first ready count = %d", readyCount)
	}

	visible, readyCount = scanner.Push([]byte("DY__"))
	if len(visible) != 0 {
		t.Fatalf("second visible output = %q", visible)
	}
	if readyCount != 1 {
		t.Fatalf("second ready count = %d", readyCount)
	}
}

func TestScannerFindsMultipleMarkers(t *testing.T) {
	scanner := NewScanner("__READY__")

	visible, readyCount := scanner.Push(
		[]byte("__READY__one__READY__two"),
	)

	if string(visible) != "onetwo" {
		t.Fatalf("visible output = %q", visible)
	}
	if readyCount != 2 {
		t.Fatalf("ready count = %d", readyCount)
	}
}

func TestScannerDoesNotDiscardFalsePrefix(t *testing.T) {
	scanner := NewScanner("__READY__")

	visible, readyCount := scanner.Push([]byte("text __READX"))

	if string(visible) != "text __READX" {
		t.Fatalf("visible output = %q", visible)
	}
	if readyCount != 0 {
		t.Fatalf("ready count = %d", readyCount)
	}
}

func TestScannerFlushesPendingBytes(t *testing.T) {
	scanner := NewScanner("__READY__")

	visible, readyCount := scanner.Push([]byte("text __REA"))
	if string(visible) != "text " {
		t.Fatalf("visible output = %q", visible)
	}
	if readyCount != 0 {
		t.Fatalf("ready count = %d", readyCount)
	}

	remaining := scanner.Flush()
	if string(remaining) != "__REA" {
		t.Fatalf("remaining output = %q", remaining)
	}
}

func TestNewMarkerCreatesDifferentMarkers(t *testing.T) {
	first, err := NewMarker()
	if err != nil {
		t.Fatal(err)
	}

	second, err := NewMarker()
	if err != nil {
		t.Fatal(err)
	}

	if first == second {
		t.Fatalf("markers should be different: %q", first)
	}
}
