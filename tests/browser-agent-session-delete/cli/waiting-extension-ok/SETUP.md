# Scenario

**Feature**: delete waiting session via CLI succeeds

```
POST /v1/sessions (no WS) -> waiting_extension
HandleCLI session delete --session-id ID --base-dir BaseDir
  -> exit 0; deleted message; dir gone; not in list
```

## Preconditions

- No fake extension connection.

## Steps

1. Set `CLIOp = waiting-extension-ok`.
2. `ConnectExtension = false`.

## Context

- Addr resolves from `server.json`; `--addr` omitted.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CLIOp = CLIOpWaitingOK
	req.ConnectExtension = false
	return nil
}
```