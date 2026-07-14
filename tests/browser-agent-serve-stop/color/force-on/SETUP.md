# Scenario

**Feature**: --color forces ANSI on serve stderr (pipe)

```
HandleCLI serve --color --stop (no daemon) -> stderr contains ESC [ ANSI sequences
```

## Preconditions

- Empty base-dir; idempotent `--stop` warning path.
- Explicit empty env (no `NO_COLOR`).

## Steps

1. Set `ColorOp = ColorOpForceOn`.

## Context

- Pipe is non-TTY; `--color` must opt in to coloring.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ColorOp = ColorOpForceOn
	return nil
}
```