# Scenario

**Feature**: session list shows `—` when page count unknown

```
POST /v1/sessions (no telemetry)
HandleCLI session list -> Pages column shows —
```

## Preconditions

- `ListOp = unknown-pages-dash`.
- No fake extension / no hello telemetry.

## Steps

1. Set `ListOp = ListOpUnknownDash`.
2. Set `SessionID = sess-rich-notelemetry`.

## Context

- nil page count → display `—` (em dash), not `0`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ListOp = ListOpUnknownDash
	req.SessionID = "sess-rich-notelemetry"
	return nil
}
```