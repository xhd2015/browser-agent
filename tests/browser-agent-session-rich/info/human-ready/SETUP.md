# Scenario

**Feature**: session info human output when status is ready

```
hello { session_page_count: 1 } + connected
HandleCLI session info (no --json) -> human Status ready
```

## Preconditions

- `InfoOp = human-ready`.
- Fake extension hello with one session page.

## Steps

1. Set `InfoOp = InfoOpHumanReady`.
2. Set `SessionID = sess-rich-info-ready`.

## Context

- connected + supports BA + count=1 → status `ready` in human output.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.InfoOp = InfoOpHumanReady
	req.SessionID = "sess-rich-info-ready"
	req.JSONMode = false
	return nil
}
```