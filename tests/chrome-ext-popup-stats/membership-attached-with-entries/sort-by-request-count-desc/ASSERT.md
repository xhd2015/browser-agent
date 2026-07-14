# Expected

Requirement scenario **#4** — tab sort by requestCount desc:

- `count` = 5+4+3+2+2 = **16**
- `tabsWatching` = 5
- `domainCount` = 1 (`x.example.com`)
- Tab order (tabId sequence): **`20, 9, 30, 2, 5`**
  - requestCounts: 5, 4, 3, 2, 2
  - tabs 2 and 5 tied at 2 → **tabId asc** (2 before 5)
- Each row’s `requestCount` matches the fixture
- Invariant sum == count

## Side Effects

- None.

## Errors

- Must not sort by tabId alone or by attachment order.
- Run must not error.

## Exit Code

- Not applicable.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunError(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	stats := resp.Stats
	assertGlobals(t, stats, 16, 5, 1)
	assertRequestCountInvariant(t, stats)
	assertTabOrderIDs(t, stats, 20, 9, 30, 2, 5)

	wantCounts := map[int]int{20: 5, 9: 4, 30: 3, 2: 2, 5: 2}
	for id, want := range wantCounts {
		tab := findTab(t, stats, id)
		if tab.RequestCount != want {
			t.Fatalf("tab %d requestCount = %d, want %d", id, tab.RequestCount, want)
		}
		if !tab.Attached {
			t.Fatalf("tab %d attached = false, want true", id)
		}
	}

	// Explicit tie-break check: among equal requestCount, tabId order.
	var twoCountIDs []int
	for _, tab := range stats.Tabs {
		if tab.RequestCount == 2 {
			twoCountIDs = append(twoCountIDs, tab.TabID)
		}
	}
	if len(twoCountIDs) != 2 || twoCountIDs[0] != 2 || twoCountIDs[1] != 5 {
		t.Fatalf("equal-count tab order = %v, want [2 5] (tabId asc)", twoCountIDs)
	}
}
```
