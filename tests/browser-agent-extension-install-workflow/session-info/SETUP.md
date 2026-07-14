# Scenario

**Feature**: session info disconnected install hints

```
POST /v1/sessions -> HandleCLI(session info) -> install-chrome-extension + path; no open-chrome
```

## Preconditions

- Disconnected session (no extension hello).
- `TestHome` for canonical path in hint text after implement.

## Steps

1. Set `Mode = session-info`.
2. Leaf sets `SessionInfoOp`.

## Context

- Human stdout (not --json).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeSessionInfo
	return nil
}
```