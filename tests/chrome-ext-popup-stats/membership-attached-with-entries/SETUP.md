# Scenario

**Feature**: attached tabs that also have captured network entries

```
# Recording in progress: watched tabs accumulating requests
Background: AttachedTabIds non-empty; Entries non-empty for those tabs
Stats Builder -> per-tab requestCount/domainCount/domains
Stats Builder -> global count + unique domainCount
```

## Preconditions

- Membership factor: **attached + entries** — every tab under this branch is
  debugger-attached and has ≥1 entry (unless a leaf documents otherwise for a
  secondary assertion).
- Leaves specialize domain rules, ordering, or URL parsing.

## Steps

1. Leave concrete fixtures to leaves.
2. Leaves must call `attachTabs` and `addEntry` as needed.

## Context

- Covers requirement scenarios **#2, #3, #4, #6, #7** (domain uniqueness,
  top-3 truncation, sort, request-count invariant, opaque host).

```go
import (
	"testing"

	"github.com/xhd2015/browser-agent/popupstats"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	// Ensure maps exist so leaves can addEntry/setTabMeta without nil checks.
	if req.Entries == nil {
		req.Entries = make(map[string]popupstats.EntryValue)
	}
	if req.TabMeta == nil {
		req.TabMeta = make(map[int]popupstats.TabMeta)
	}
	// Leaves under this branch must attach at least one tab and add entries.
	return nil
}
```

