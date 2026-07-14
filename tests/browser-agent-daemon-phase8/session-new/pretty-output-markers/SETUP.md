# Scenario

**Feature**: SessionNew pretty stdout for operators

```
SessionNew(explicit id) -> stdout contains session-id, export hint, inspect recipes
```

## Preconditions

- Explicit session id `sess-pretty-8`.
- `SessionNewOp` pretty-output-markers.

## Steps

1. Set `SessionNewOp = SessionNewOpPrettyOutputMarkers`.
2. Set `SessionID = "sess-pretty-8"`.

## Context

- Export hint should mention `BROWSER_AGENT_SESSION_ID` or export wording.
- Inspect/interact recipes use nested `browser-agent session info|eval|…` form.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionNewOp = SessionNewOpPrettyOutputMarkers
	req.SessionID = "sess-pretty-8"
	return nil
}```
