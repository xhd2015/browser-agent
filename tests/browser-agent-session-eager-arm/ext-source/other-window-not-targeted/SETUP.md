# Scenario

**Feature**: create + eager attach scoped to session `entry.windowId` only

```
createTabInSession
  -> chrome.tabs.create({ windowId: entry.windowId, ... })
  -> eager attach new tab in that window only

tabs.onUpdated / onCreated eager attach
  -> only if tab.windowId === entry.windowId for armed session
  -> tabs in other windows: no proactive auto-attach for this session
```

## Preconditions

- ExtSourceTarget = other-window-not-targeted.

## Steps

1. Set `ExtSourceTarget = ExtSrcOtherWindowNotTargeted`.

## Context

- Create already scopes to `entry.windowId` (regression); this leaf also requires
  **eager attach** to be window-scoped — create-only without eager attach **fails**
  (P2 not landed).
- Still requires session bound (`windowId` from `/go`) — no attach without arming.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcOtherWindowNotTargeted
	return nil
}
```
