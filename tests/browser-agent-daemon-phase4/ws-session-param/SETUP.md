# Scenario

**Feature**: `GET /v1/ws` session query param when registry has 2+ sessions

```
Test Client -> registry with sess-p4-a + sess-p4-b
Test Client -> GET /v1/ws (no session) -> 400
Test Client -> GET /v1/ws?session=sess-p4-a -> WebSocket upgrade ok
```

## Preconditions

- Two sessions pre-created in registry.
- Mode is `ws-session-param`.

## Steps

1. Set `Mode = ModeWSSessionParam`.
2. Leaves set `WSSessionOp` and dial options.

## Context

- Single-session fallback (missing param ok) is backward-compat for `Run`; not in this tree.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeWSSessionParam
	return nil
}
```