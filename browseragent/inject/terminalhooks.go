package inject

import "io"

// IsTerminalFn, when non-nil, overrides TTY detection for serve --stop confirm (tests).
var IsTerminalFn func(io.Reader) bool