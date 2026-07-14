# Scenario

**Feature**: --no-color disables ANSI on serve stderr

```
HandleCLI serve --no-color --stop -> stderr has no ESC [ sequences
```

## Preconditions

- Empty base-dir; idempotent `--stop` warning path.
- Explicit empty env.

## Steps

1. Set `ColorOp = ColorOpNoColorFlag`.

## Context

- Flag must override default TTY-auto behavior when forced off.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ColorOp = ColorOpNoColorFlag
	return nil
}
```