# Scenario

**Feature**: `HandleCLI session new` dispatch

```
HandleCLI ["session","new", flags...] -> EnsureDaemon + create + stdout; exit 0
```

## Preconditions

- Mode `ModeCLIDispatch`.
- Leaf sets `CLIDispatchOp`.
- `inject.WithSessionNewHooks` around HandleCLI; `--no-wait` (no real Chrome / no extension poll).

## Steps

1. Set `Mode = ModeCLIDispatch`.

## Context

- CLI must register `session new` in `cliSession` switch.
- Harness uses ephemeral `--addr` on loopback `:0`.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeCLIDispatch
	return nil
}```
