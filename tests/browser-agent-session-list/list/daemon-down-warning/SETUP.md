# Scenario

**Feature**: list when daemon unreachable prints warning and exits 0

```
no RunDaemon; empty BaseDir without server.json
HandleCLI session list --base-dir BaseDir
  -> stderr warning: daemon not running; stdout empty / "0 sessions"
```

## Preconditions

- No daemon; no `server.json` health meta.

## Steps

1. Set `ListOp = ListOpDaemonDown`.
2. `StartDaemon = false`.
3. `SessionIDsToCreate` empty.

## Context

- Approved default Q1: daemon unreachable → exit 0 + warning stderr + empty output.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ListOp = ListOpDaemonDown
	req.StartDaemon = false
	req.SessionIDsToCreate = nil
	req.JSONMode = false
	return nil
}```
