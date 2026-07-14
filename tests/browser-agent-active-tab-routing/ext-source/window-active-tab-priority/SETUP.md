# Scenario

**Bug**: `pickTargetTabIdForSession` ignored active tab in session window

```
Background pickTargetTabIdForSession(S)
  -> chrome.tabs.query({ active: true, windowId: entry.windowId })
  -> fallback entry.tabId (session page)
  -> fallback URL /go?session=S
```

## Preconditions

- `ExtSourceTarget` = window-active-tab-priority.

## Steps

1. Set `ExtSourceTarget = ExtSrcWindowActiveTabPriority`.

## Context

- Does **not** require global `lastFocusedWindow`-only routing.
- Post-fix: active capturable tab in session window wins over registered session page tab.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcWindowActiveTabPriority
	return nil
}
```