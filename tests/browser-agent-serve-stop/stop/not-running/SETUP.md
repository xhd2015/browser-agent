# Scenario

**Feature**: serve --stop when no daemon is running

```
Empty base-dir (no server.json)
HandleCLI serve --stop -> warning stderr -> exit 0
```

## Preconditions

- No prior `RunDaemon`; `server.json` absent.

## Steps

1. Set `StopOp = StopOpNotRunning`.

## Context

- Idempotent operator path: warning on stderr, success exit code.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.StopOp = StopOpNotRunning
	return nil
}
```