# Scenario

**Feature**: WS status push updates server page count after hello

```
hello count=1 -> status count=2 -> GET /v1/session count=2
```

## Preconditions

- `ExtensionOp = status-push-updates-count`.

## Steps

1. Set `ExtensionOp = ExtensionOpStatusPush`.
2. Set `SessionID = sess-rich-status-push`.
3. Set `StatusPushCount = 2`.

## Context

- Tab open/close reflected via `type=status` messages.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtensionOp = ExtensionOpStatusPush
	req.SessionID = "sess-rich-status-push"
	req.StatusPushCount = 2
	return nil
}
```