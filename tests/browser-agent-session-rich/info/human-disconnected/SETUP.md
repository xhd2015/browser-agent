# Scenario

**Feature**: session info human default when disconnected

```
POST /v1/sessions (no WS)
HandleCLI session info (no --json) -> human sections + hints
```

## Preconditions

- `InfoOp = human-disconnected`.
- No `--json` flag.

## Steps

1. Set `InfoOp = InfoOpHumanDisconnected`.
2. Set `SessionID = sess-rich-info-disc`.

## Context

- Human output mentions session URL and delete hint; NOT raw JSON-only.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.InfoOp = InfoOpHumanDisconnected
	req.SessionID = "sess-rich-info-disc"
	req.JSONMode = false
	return nil
}
```