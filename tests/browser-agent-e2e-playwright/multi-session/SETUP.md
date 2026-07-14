# Scenario

**Feature**: Multi-session isolation in real browser

```
RunDaemon -> POST /v1/sessions (session A + session B, no Chrome)
playwright-debug -> two tabs -> each /go?session=<id>
both sessions connect independently
```

## Preconditions

- Harness creates **both** sessions before playwright script runs.
- Script receives `sessionIdA` argv[4], `sessionIdB` argv[5].

## Steps

1. Leaf sets `PlaywrightOp`, `SessionID`, `SessionIDB`.

## Context

- Validates per-session extension WS isolation across tabs.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	return nil
}
```