# Scenario

**Feature**: last session-page leave detaches every tab in the session attach set

```
Last /go?session=S leaves window
  -> unregisterSession / leave path
  -> detachDebugger for every tabId in session attach set
  -> chrome.debugger no longer attached to any session tab
```

## Preconditions

- ExtSourceTarget = detach-on-session-leave.

## Steps

1. Set `ExtSourceTarget = ExtSrcDetachOnSessionLeave`.

## Context

- Policy B: leave must clear the **whole** attach set, not only a sticky singular
  `attachedTabId`.
- Detach-on-tab-switch inside `attachDebuggerForSession` does **not** satisfy this leaf
  (and switch-detach is obsolete).
- Sticky single-id one-shot detach **fails** this leaf until multi-set detach-all lands.

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
