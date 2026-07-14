# Scenario

**Feature**: RunDaemon blocking host with empty registry

```
RunDaemon(ctx, DaemonConfig) -> registry with zero sessions
Test Client -> GET /v1/health | /v1/sessions | read server.json
ctx cancel -> server.json removed
```

## Preconditions

- Mode `ModeRunDaemon`.
- Leaf sets `RunDaemonOp`.

## Steps

1. Set `Mode = ModeRunDaemon`.

## Context

- No Chrome, no agent-run, no auto-session create.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeRunDaemon
	return nil
}
```