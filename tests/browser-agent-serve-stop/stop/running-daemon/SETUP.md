# Scenario

**Feature**: serve --stop kills a running daemon

```
RunDaemon -> health OK + server.json
HandleCLI serve --stop -> health down + meta gone + exit 0
```

## Preconditions

- Prior daemon healthy with `server.json`.
- Same `--base-dir` and `--addr` for CLI stop.

## Steps

1. Set `StopOp = StopOpRunningDaemon`.

## Context

- Must not start a second serve (no re-bind on same addr).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.StopOp = StopOpRunningDaemon
	return nil
}
```