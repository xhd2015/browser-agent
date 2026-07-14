# Expected

Requirement scenario **#1** — empty entries, one attached tab:

- `count` = 0
- `tabsWatching` = 1
- `domainCount` = 0
- Exactly **one** tab row
- Row tabId `1`: `attached=true`, `active=true`, `requestCount=0`, `domainCount=0`
- `domains` empty (nil or len 0)
- Title is `"Home"` (from TabMeta)
- URL is `https://example.com/` (from TabMeta)
- Request-count invariant holds (0 == 0)

## Side Effects

- None (pure function).

## Errors

- Run must not return an error.
- Must not invent non-zero counts from an empty buffer.

## Exit Code

- Not applicable (library call).

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
	assertGlobals(t, stats, 0, 1, 0)
	assertRequestCountInvariant(t, stats)

	if len(stats.Tabs) != 1 {
		t.Fatalf("tabs len = %d, want 1", len(stats.Tabs))
	}
	tab := findTab(t, stats, 1)
	if !tab.Attached {
		t.Fatal("tab 1 attached = false, want true")
	}
	if !tab.Active {
		t.Fatal("tab 1 active = false, want true")
	}
	if tab.RequestCount != 0 {
		t.Fatalf("tab 1 requestCount = %d, want 0", tab.RequestCount)
	}
	if tab.DomainCount != 0 {
		t.Fatalf("tab 1 domainCount = %d, want 0", tab.DomainCount)
	}
	if len(tab.Domains) != 0 {
		t.Fatalf("tab 1 domains len = %d, want 0", len(tab.Domains))
	}
	if tab.Title != "Home" {
		t.Fatalf("tab 1 title = %q, want %q", tab.Title, "Home")
	}
	if tab.URL != "https://example.com/" {
		t.Fatalf("tab 1 url = %q, want %q", tab.URL, "https://example.com/")
	}
}
```
