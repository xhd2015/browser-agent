# Scenario

**Feature**: delete unknown session id fails with not found

```
RunDaemon -> no session for unknown id
HandleCLI session delete --session-id unknown -> exit 1; not found
```

## Preconditions

- Unknown session id never created.

## Steps

1. Set `CLIOp = not-found`.
2. Set `UnknownSessionID = sess-not-found-delete`.

## Context

- Daemon running so addr resolves; target id absent from registry and disk.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CLIOp = CLIOpNotFound
	req.UnknownSessionID = "sess-not-found-delete"
	req.ConnectExtension = false
	return nil
}
```