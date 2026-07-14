package browseragent

import (
	"io"
	"os"
)

// fileIsTTY reports whether f is a character device (terminal-ish).
// Buffers and non-files return false. Used for orange mismatch coloring.
func fileIsTTY(f *os.File) bool {
	if f == nil {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// writerIsTTY returns true when w is (or wraps) an *os.File that looks like a TTY.
func writerIsTTY(w io.Writer) bool {
	if w == nil {
		return false
	}
	if f, ok := w.(*os.File); ok {
		return fileIsTTY(f)
	}
	// Unwrap common multi-writers? Not required for tests (bytes.Buffer).
	return false
}
