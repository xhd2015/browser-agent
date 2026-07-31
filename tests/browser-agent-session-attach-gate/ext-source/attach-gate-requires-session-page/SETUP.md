# Scenario

**Feature**: attach gate requires an open same-window session control tab

```
withDebuggerForSession / attachDebuggerForSession
  -> gate: ≥1 open /go?session=S in target window
  -> else refuse attach (clear error); no chrome.debugger.attach
```

## Preconditions

- ExtSourceTarget = attach-gate-requires-session-page.

## Steps

1. Set `ExtSourceTarget = ExtSrcAttachGateRequiresSessionPage`.

## Context

- Regression: attach gate must remain (`hasOpenSessionPage` / refuse without control tab).
- Routing errors like “session page not bound (windowId missing)” are **not** the attach gate.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcAttachGateRequiresSessionPage
	return nil
}
```
