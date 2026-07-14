# Scenario

**Feature**: pure popup recording stats from in-memory capture state (no Chrome)

```
# Background holds capture buffer + attached set + tab meta
Background Service Worker: Entries, AttachedTabIds, TabMeta

# Stats Builder is pure; Popup only displays the result
Test Client -> popupstats.BuildPopupStats(Input)
          -> PopupStats { count, tabsWatching, domainCount, tabs[] }

# Popup poll path (not executed here)
Popup ?- getState -> Background (enriched by same rules)
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Package `popupstats` exports `BuildPopupStats(Input) PopupStats` and the
  types listed in root `DOCTEST.md` (EntryValue, TabMeta, DomainCount, TabStats,
  PopupStats, Input).
- No Chrome process, no extension load, no filesystem side effects.
- Entry keys use `"${tabId}:${requestId}"` (integer tab id before first `:`).
- Host bucket for unparseable URLs is the literal string `opaque`.
- Leaves start with empty `Entries` / empty `AttachedTabIDs` / empty `TabMeta`
  unless a grouping or leaf Setup populates them.

## Steps

1. Initialize `Entries` to an empty map if nil.
2. Initialize `TabMeta` to an empty map if nil.
3. Leave `AttachedTabIDs` nil/empty; descendants set membership and fixtures.
4. Provide shared helpers for entry keys, adding URLs, tab meta, and common asserts.

## Context

- Parallel-safe: pure function, no shared mutable process state beyond Go maps
  owned by each leaf’s `Request`.
- Helpers below are available to all descendant Setup/Assert packages.
- Scenario **#6** (sum of per-tab `requestCount` equals global `count`) is
  asserted on multi-entry leaves via `assertRequestCountInvariant`.

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/browser-agent/popupstats"
)

// OpaqueHost is the single bucket name for invalid/unparseable request URLs.
const OpaqueHost = "opaque"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	if req.Entries == nil {
		req.Entries = make(map[string]popupstats.EntryValue)
	}
	if req.TabMeta == nil {
		req.TabMeta = make(map[int]popupstats.TabMeta)
	}
	// AttachedTabIDs intentionally left nil until grouping/leaf Setup.
	return nil
}

// entryKey builds the preferred capture key "${tabId}:${requestId}".
func entryKey(tabID int, requestID string) string {
	return fmt.Sprintf("%d:%s", tabID, requestID)
}

// addEntry records one capture for tabID with the given request URL.
func addEntry(req *Request, tabID int, requestID, url string) {
	if req.Entries == nil {
		req.Entries = make(map[string]popupstats.EntryValue)
	}
	req.Entries[entryKey(tabID, requestID)] = popupstats.EntryValue{URL: url}
}

// setTabMeta sets chrome.tabs-style metadata for tabID.
func setTabMeta(req *Request, tabID int, title, pageURL string, active bool) {
	if req.TabMeta == nil {
		req.TabMeta = make(map[int]popupstats.TabMeta)
	}
	req.TabMeta[tabID] = popupstats.TabMeta{
		Title:  title,
		URL:    pageURL,
		Active: active,
	}
}

// attachTabs replaces AttachedTabIDs with the given unique tab ids.
func attachTabs(req *Request, tabIDs ...int) {
	req.AttachedTabIDs = append([]int(nil), tabIDs...)
}

// findTab returns the TabStats for tabID, or reports failure.
func findTab(t *testing.T, stats popupstats.PopupStats, tabID int) popupstats.TabStats {
	t.Helper()
	for _, tab := range stats.Tabs {
		if tab.TabID == tabID {
			return tab
		}
	}
	t.Fatalf("tabId %d not found in stats.Tabs (%d rows)", tabID, len(stats.Tabs))
	return popupstats.TabStats{}
}

// assertNoRunError fails if Run returned an error (builder must not panic/err).
func assertNoRunError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("BuildPopupStats should not error; got %v", err)
	}
}

// assertGlobals checks count / tabsWatching / domainCount.
func assertGlobals(t *testing.T, stats popupstats.PopupStats, count, watching, domains int) {
	t.Helper()
	if stats.Count != count {
		t.Fatalf("count = %d, want %d", stats.Count, count)
	}
	if stats.TabsWatching != watching {
		t.Fatalf("tabsWatching = %d, want %d", stats.TabsWatching, watching)
	}
	if stats.DomainCount != domains {
		t.Fatalf("domainCount = %d, want %d", stats.DomainCount, domains)
	}
}

// assertRequestCountInvariant enforces sum(tab.requestCount) == count (#6).
func assertRequestCountInvariant(t *testing.T, stats popupstats.PopupStats) {
	t.Helper()
	sum := 0
	for _, tab := range stats.Tabs {
		if tab.RequestCount < 0 {
			t.Fatalf("tabId %d requestCount = %d, want >= 0", tab.TabID, tab.RequestCount)
		}
		sum += tab.RequestCount
	}
	if sum != stats.Count {
		t.Fatalf("invariant sum(tab.requestCount)=%d != count=%d", sum, stats.Count)
	}
}

// assertTabOrderIDs checks Tabs are in the expected tabId sequence (sort result).
func assertTabOrderIDs(t *testing.T, stats popupstats.PopupStats, wantIDs ...int) {
	t.Helper()
	if len(stats.Tabs) != len(wantIDs) {
		t.Fatalf("tabs len = %d, want %d", len(stats.Tabs), len(wantIDs))
	}
	for i, want := range wantIDs {
		if stats.Tabs[i].TabID != want {
			got := make([]int, len(stats.Tabs))
			for j, tab := range stats.Tabs {
				got[j] = tab.TabID
			}
			t.Fatalf("tabs order by tabId = %v, want %v (at index %d)", got, wantIDs, i)
		}
	}
}

// domainHosts returns host names from a domains slice (order preserved).
func domainHosts(domains []popupstats.DomainCount) []string {
	out := make([]string, len(domains))
	for i, d := range domains {
		out[i] = d.Host
	}
	return out
}
```
