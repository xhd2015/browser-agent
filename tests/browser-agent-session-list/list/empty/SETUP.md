# Scenario

**Feature**: list with zero sessions on running daemon

```
RunDaemon -> GET /v1/sessions -> []
HandleCLI session list --base-dir BaseDir
  -> exit 0; "0 sessions" or empty table
```

## Preconditions

- Daemon running; no sessions created.

## Steps

1. Set `ListOp = ListOpEmpty`.
2. `SessionIDsToCreate` empty.

## Context

- Human output should indicate zero sessions without error.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ListOp = ListOpEmpty
	req.SessionIDsToCreate = nil
	req.JSONMode = false
	return nil
}```
