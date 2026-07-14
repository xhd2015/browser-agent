# Expected

Requirement scenario **#5** — closed tab with entries, not attached:

| Field | Value |
|-------|-------|
| `count` | 2 |
| `tabsWatching` | **1** (only tab 1) |
| `domainCount` | 1 (`closed.example.com`) |
| tabs length | **2** (union of attached ∪ has-entries) |

Tab **7** (closed):

- Present in `tabs`
- `attached` = **false**
- `requestCount` = 2, `domainCount` = 1
- domains include `closed.example.com` count 2
- `title` = **`"Closed tab"`** (no meta, not attached)
- `active` = false (no meta)

Tab **1** (watching, empty):

- `attached` = true
- `requestCount` = 0
- title `"Live"`

Order: tab 7 before tab 1 (requestCount 2 > 0).

Invariant sum(requestCount)==count (2+0==2).

## Side Effects

- None.

## Errors

- Must not drop closed tabs that still have entries.
- Must not count closed tabs toward `tabsWatching`.
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
	assertGlobals(t, stats, 2, 1, 1)
	assertRequestCountInvariant(t, stats)
	assertTabOrderIDs(t, stats, 7, 1)

	closed := findTab(t, stats, 7)
	if closed.Attached {
		t.Fatal("closed tab 7 attached = true, want false")
	}
	if closed.RequestCount != 2 {
		t.Fatalf("tab 7 requestCount = %d, want 2", closed.RequestCount)
	}
	if closed.DomainCount != 1 {
		t.Fatalf("tab 7 domainCount = %d, want 1", closed.DomainCount)
	}
	if closed.Title != "Closed tab" {
		t.Fatalf("tab 7 title = %q, want %q", closed.Title, "Closed tab")
	}
	if closed.Active {
		t.Fatal("tab 7 active = true, want false")
	}
	if len(closed.Domains) != 1 || closed.Domains[0].Host != "closed.example.com" || closed.Domains[0].Count != 2 {
		t.Fatalf("tab 7 domains = %v, want [closed.example.com:2]", closed.Domains)
	}

	live := findTab(t, stats, 1)
	if !live.Attached {
		t.Fatal("live tab 1 attached = false, want true")
	}
	if live.RequestCount != 0 {
		t.Fatalf("tab 1 requestCount = %d, want 0", live.RequestCount)
	}
	if live.Title != "Live" {
		t.Fatalf("tab 1 title = %q, want Live", live.Title)
	}
}
```
