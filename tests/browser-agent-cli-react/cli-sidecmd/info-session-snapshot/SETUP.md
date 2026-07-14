# Scenario

**Feature**: session info against live server returns session snapshot (B2)

```
Serve session X
HandleCLI session info --session-id X --addr <base>
  -> nil error
  -> stdout includes session_id and connected field (JSON or text)
  -> trailing \n
```

## Preconditions

- Sidecmd = info.
- Fake extension optional; connected may be false without hello — field still present.
- Nested `session info` only.

## Steps

1. Set SidecmdInfo.
2. Leave CLIArgs empty for harness injection of addr/session.

## Context

- Snapshot may come from GET /v1/session or an info job; either is fine.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Sidecmd = SidecmdInfo
	req.FakeExtension = false
	req.CLIArgs = nil
	return nil
}
```
