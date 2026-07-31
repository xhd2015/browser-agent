# Scenario

**Feature**: auto-attach capturable tabs on same-window open/navigate while armed

```
tabs.onUpdated / tabs.onCreated (while session armed)
  -> tab.windowId === entry.windowId
  -> isCapturableTabURL(url)
  -> attachDebuggerForSession(sessionId, tabId)
```

## Preconditions

- ExtSourceTarget = auto-attach-on-same-window-navigate.

## Steps

1. Set `ExtSourceTarget = ExtSrcAutoAttachOnSameWindowNavigate`.

## Context

- Current `tabs.onUpdated` only `maybeRegisterGoTab` + leave recount — **RED**
  until user-tab auto-attach lands.
- `maybeRegisterGoTab` / control-tab register alone does **not** satisfy (covered by
  `auto-attach-on-go-register`).
- Scope: same armed window only (other-window covered by sibling leaf).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcAutoAttachOnSameWindowNavigate
	return nil
}
```
