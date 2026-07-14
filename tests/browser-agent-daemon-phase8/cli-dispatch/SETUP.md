# Scenario

**Feature**: `HandleCLI session new` dispatch

```
HandleCLI ["session","new", flags...] -> EnsureDaemon + create + stdout; exit 0
```

## Preconditions

- Mode `ModeCLIDispatch`.
- Leaf sets `CLIDispatchOp`.
- `SessionNewTestHooks` used for OpenChrome recording (no real Chrome).

## Steps

1. Set `Mode = ModeCLIDispatch`.

## Context

- CLI must register `session new` in `cliSession` switch.
- Harness uses ephemeral `--addr` on loopback `:0`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeCLIDispatch
	return nil
}```
