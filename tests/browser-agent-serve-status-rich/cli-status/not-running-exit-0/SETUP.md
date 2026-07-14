# Scenario

**Feature**: CLI rich status when daemon not running

```
HandleCLI serve --status -> extension block + not running -> exit 0
```

## Preconditions

- `CLIStatusOp` not-running-exit-0.

## Steps

1. Set `CLIStatusOp = CLIStatusOpNotRunningExit0`.

## Context

- No `server.json`; extension block still shown.

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