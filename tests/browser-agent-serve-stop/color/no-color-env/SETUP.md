# Scenario

**Feature**: NO_COLOR=1 disables ANSI on serve stderr

```
env NO_COLOR=1 + HandleCLI serve --stop -> stderr has no ESC [ sequences
```

## Preconditions

- Empty base-dir; no `--color` / `--no-color` flags.
- Re-inject `NO_COLOR=1` into explicit env map.

## Steps

1. Set `ColorOp = ColorOpNoColorEnv`.
2. Set `req.Env["NO_COLOR"] = "1"`.

## Context

- Only this leaf sets `NO_COLOR`; root Setup keeps ambient env out.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ColorOp = ColorOpNoColorEnv
	if req.Env == nil {
		req.Env = map[string]string{}
	}
	req.Env["NO_COLOR"] = "1"
	return nil
}
```