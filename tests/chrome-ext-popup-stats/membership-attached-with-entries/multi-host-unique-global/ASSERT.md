# Expected

Requirement scenario **#2** — two tabs, multi-host URLs:

| Field | Value |
|-------|-------|
| `count` | 5 |
| `tabsWatching` | 2 |
| `domainCount` (global) | **3** (api, cdn, auth) — not 4 |
| tabs length | 2 |
| order | tab 1 then tab 2 (requestCount 3 > 2) |

Tab **1**:

- `requestCount` = 3, `domainCount` = 2, `attached` = true, title `"App"`
- `domains` top hosts include `api.example.com` (count 2) then `cdn.example.com` (count 1)

Tab **2**:

- `requestCount` = 2, `domainCount` = 2, `attached` = true, title `"Auth"`
- hosts `api.example.com` and `auth.example.com` (each count 1; order by count desc, tie host order implementation-defined among equal counts — both must appear)

Requirement **#6** — `sum(tab.requestCount) == count` (5).

## Side Effects

- None.

## Errors

- Must not treat global domainCount as sum of per-tab domainCounts.
- Run must not error.

## Exit Code

- Not applicable.

```go
import (
	"testing"

	"github.com/xhd2015/browser-agent/popupstats"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunError(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	stats := resp.Stats
	assertGlobals(t, stats, 5, 2, 3)
	assertRequestCountInvariant(t, stats)
	assertTabOrderIDs(t, stats, 1, 2)

	t1 := findTab(t, stats, 1)
	if t1.RequestCount != 3 {
		t.Fatalf("tab1 requestCount = %d, want 3", t1.RequestCount)
	}
	if t1.DomainCount != 2 {
		t.Fatalf("tab1 domainCount = %d, want 2", t1.DomainCount)
	}
	if !t1.Attached {
		t.Fatal("tab1 attached = false, want true")
	}
	if t1.Title != "App" {
		t.Fatalf("tab1 title = %q, want App", t1.Title)
	}
	assertDomainCount(t, t1, "api.example.com", 2)
	assertDomainCount(t, t1, "cdn.example.com", 1)
	if len(t1.Domains) != 2 {
		t.Fatalf("tab1 domains len = %d, want 2", len(t1.Domains))
	}
	// Sorted by count desc: api (2) before cdn (1).
	if t1.Domains[0].Host != "api.example.com" || t1.Domains[0].Count != 2 {
		t.Fatalf("tab1 domains[0] = {%q,%d}, want api.example.com/2", t1.Domains[0].Host, t1.Domains[0].Count)
	}

	t2 := findTab(t, stats, 2)
	if t2.RequestCount != 2 {
		t.Fatalf("tab2 requestCount = %d, want 2", t2.RequestCount)
	}
	if t2.DomainCount != 2 {
		t.Fatalf("tab2 domainCount = %d, want 2", t2.DomainCount)
	}
	if !t2.Attached {
		t.Fatal("tab2 attached = false, want true")
	}
	assertDomainCount(t, t2, "api.example.com", 1)
	assertDomainCount(t, t2, "auth.example.com", 1)
}

func assertDomainCount(t *testing.T, tab popupstats.TabStats, host string, want int) {
	t.Helper()
	for _, d := range tab.Domains {
		if d.Host == host {
			if d.Count != want {
				t.Fatalf("tab %d host %q count = %d, want %d", tab.TabID, host, d.Count, want)
			}
			return
		}
	}
	t.Fatalf("tab %d missing domain %q in %v", tab.TabID, host, domainHosts(tab.Domains))
}
```
