# Scenario

**Feature**: closed tab with entries, not attached, still listed (#5)

```
# Tab 7 closed (no longer attached) but still has 2 captures
# Tab 1 still watching with 0 captures
Background:
  AttachedTabIds={1}
  Entries for tab 7 only (two URLs on closed.example.com)
  TabMeta only for tab 1 (tab 7 meta missing)
Test Client -> BuildPopupStats
          -> tabsWatching=1
          -> tabs include both 7 (attached=false, title "Closed tab") and 1 (attached=true)
          -> tab 7 sorts first (requestCount 2 > 0)
```

## Preconditions

- Attached set is only tab `1` (still watching, empty buffer).
- Entries only for tab `7` (closed / not attached).
- No TabMeta for tab `7` → title fallback `"Closed tab"`.
- TabMeta for tab `1` present.

## Steps

1. Attach only tab `1`; set its meta title `"Live"`, active true.
2. Add two entries for tab `7` on `https://closed.example.com/…`.
3. Do **not** put `7` in AttachedTabIDs; do **not** set TabMeta for `7`.

## Context

- Watching chip counts attached only; list length can exceed `tabsWatching`
  when closed tabs still hold captures.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	attachTabs(req, 1)
	setTabMeta(req, 1, "Live", "https://live.example.com/", true)

	// Closed tab 7: entries only, no attach, no meta.
	addEntry(req, 7, "c1", "https://closed.example.com/one")
	addEntry(req, 7, "c2", "https://closed.example.com/two")
	return nil
}
```
