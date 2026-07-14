# Scenario

**Feature**: WS `type=status` push updates server page count after hello

```
Fake Extension -> hello count=1
Fake Extension -> status count=2
GET /v1/session -> session_page_count=2
```

## Preconditions

- Mode is `extension`.
- Extension stays connected across hello + status push.

## Steps

1. Set `Mode = ModeExtension`.
2. Leaf sets `ExtensionOp = status-push-updates-count`.

## Context

- Tab open/close telemetry via status messages (ws.go handles).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeExtension
	return nil
}
```