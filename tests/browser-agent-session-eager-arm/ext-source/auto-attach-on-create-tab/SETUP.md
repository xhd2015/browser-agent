# Scenario

**Feature**: auto-attach new capturable tab after create_tab in session window

```
createTabInSession(S, { url })
  -> chrome.tabs.create({ windowId: entry.windowId, url, active })
  -> if capturable: attachDebuggerForSession(S, newTabId)
  -> return { type: create_tab, tab_id, ... }
```

## Preconditions

- ExtSourceTarget = auto-attach-on-create-tab.

## Steps

1. Set `ExtSourceTarget = ExtSrcAutoAttachOnCreateTab`.

## Context

- Current `createTabInSession` only creates the tab and returns `tab_id` — **RED**
  until post-create attach lands.
- Shared create path also serves `Target.createTarget`; eager attach should apply
  on the shared path so both job types arm the new tab.
- Non-capturable create URLs are covered by `skip-non-capturable`.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcAutoAttachOnCreateTab
	return nil
}
```
