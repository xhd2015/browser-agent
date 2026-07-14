# Scenario

**Feature**: cli-color on serve operator stderr

```
HandleCLI serve operator paths -> colored or plain stderr per --color / --no-color / NO_COLOR
```

## Preconditions

- Mode `ModeColor`.
- Explicit `req.Env` map (ambient `NO_COLOR` stripped at root).
- Non-TTY pipe: `--color` forces ANSI when color is on.

## Steps

1. Set `Mode = ModeColor`.

## Context

- Uses `serve --stop` against empty base-dir for deterministic stderr (warning path).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeColor
	return nil
}
```