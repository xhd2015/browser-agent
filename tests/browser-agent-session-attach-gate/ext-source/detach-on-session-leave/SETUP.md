# Scenario

**Feature**: last session-page leave detaches session debugger (banner gone)

```
Last /go?session=S leaves window
  -> unregisterSession / leave path
  -> detachDebugger for sessionAttachState.attachedTabId
  -> chrome.debugger no longer attached
```

## Preconditions

- ExtSourceTarget = detach-on-session-leave.

## Steps

1. Set `ExtSourceTarget = ExtSrcDetachOnSessionLeave`.

## Context

- Current master: `unregisterSession` closes WS only — **no** `chrome.debugger.detach`.
- Detach-on-tab-switch inside `attachDebuggerForSession` does **not** satisfy this leaf.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcDetachOnSessionLeave
	return nil
}
```
