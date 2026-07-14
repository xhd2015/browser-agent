# Scenario

**Feature**: CLI rich status against running daemon

```
RunDaemon + session -> HandleCLI serve --status -> rich stdout -> exit 0
```

## Preconditions

- `CLIStatusOp` running-exit-0.

## Steps

1. Set `CLIStatusOp = CLIStatusOpRunningExit0`.

## Context

- Phase7 parity plus version/extension/Connected markers.

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