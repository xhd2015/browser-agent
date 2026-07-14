# Scenario

**Feature**: tabs that appear because they still have captures after detach/close

```
# User closed a tab mid-recording; entries remain until Stop
Background: Entries for tab C; C ∉ AttachedTabIds
Stats Builder -> still emits a row for C with attached=false
Stats Builder -> tabsWatching counts only currently attached ids
```

## Preconditions

- Membership factor: **closed / not attached but has entries**.
- Product rule: closed tabs with captures stay in the list until Stop.

## Steps

1. Leaves set attached set (may include other live tabs) and entries for a
   non-attached tab id.
2. Prefer missing TabMeta for the closed tab to exercise title fallback
   `"Closed tab"`.

## Context

- Covers requirement scenario **#5**.

```go
import (
	"testing"

	"github.com/xhd2015/browser-agent/popupstats"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	if req.Entries == nil {
		req.Entries = make(map[string]popupstats.EntryValue)
	}
	if req.TabMeta == nil {
		req.TabMeta = make(map[int]popupstats.TabMeta)
	}
	// Leaves must add entries for a tab that is NOT in AttachedTabIDs.
	return nil
}
```

