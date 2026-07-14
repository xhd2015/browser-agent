# Scenario

**Feature**: tabs sorted by requestCount descending, tabId ascending tie-break (#4)

```
# Three attached tabs with distinct request counts + two equal-count tabs
Background: AttachedTabIds={2,5,9,20,30}
  tab 20 → 5 reqs; tab 30 → 3; tab 9 → 4; tab 2 → 2; tab 5 → 2
Test Client -> BuildPopupStats
          -> tabs order by requestCount desc, then tabId asc:
             20 (5), 9 (4), 30 (3), 2 (2), 5 (2)
```

## Preconditions

- Five attached tabs with controlled request counts.
- Hosts may be the same (`https://x.example.com/…`); ordering is by counts only.
- Equal counts for tab 2 and tab 5 → smaller tabId first (2 before 5).

## Steps

1. Attach tabs `2, 5, 9, 20, 30`.
2. Add entries:
   - tab 20: 5 requests
   - tab 9: 4 requests
   - tab 30: 3 requests
   - tab 2: 2 requests
   - tab 5: 2 requests
3. Minimal TabMeta titles for readability.

## Context

- Requirement documents tabId ascending as an acceptable stable tie-break;
  this leaf **requires** that tie-break so order is fully specified.

```go
import (
	"fmt"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	attachTabs(req, 2, 5, 9, 20, 30)

	addN := func(tabID, n int) {
		for i := 0; i < n; i++ {
			addEntry(req, tabID, fmt.Sprintf("t%d-r%d", tabID, i),
				fmt.Sprintf("https://x.example.com/%d/%d", tabID, i))
		}
	}
	addN(20, 5)
	addN(9, 4)
	addN(30, 3)
	addN(2, 2)
	addN(5, 2)

	setTabMeta(req, 20, "T20", "https://x.example.com/", false)
	setTabMeta(req, 9, "T9", "https://x.example.com/", false)
	setTabMeta(req, 30, "T30", "https://x.example.com/", true)
	setTabMeta(req, 2, "T2", "https://x.example.com/", false)
	setTabMeta(req, 5, "T5", "https://x.example.com/", false)
	return nil
}
```
