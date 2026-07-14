# Scenario

**Feature**: CLI status against running daemon

```
RunDaemon + session -> HandleCLI serve --status -> stdout table -> exit 0
```

## Preconditions

- `CLIStatusOp` running-exit-0.

## Steps

1. Set `CLIStatusOp = CLIStatusOpRunningExit0`.

## Context

- CLI must return immediately (not block in RunDaemon).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CLIStatusOp = CLIStatusOpRunningExit0
	return nil
}
```