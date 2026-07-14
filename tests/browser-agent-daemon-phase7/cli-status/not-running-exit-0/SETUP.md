# Scenario

**Feature**: CLI status when daemon not running

```
no server.json -> HandleCLI serve --status -> stdout not-running -> exit 0
```

## Preconditions

- `CLIStatusOp` not-running-exit-0.
- No daemon started.

## Steps

1. Set `CLIStatusOp = CLIStatusOpNotRunningExit0`.

## Context

- Exit 0 even when not running (operator-friendly).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CLIStatusOp = CLIStatusOpNotRunningExit0
	return nil
}
```