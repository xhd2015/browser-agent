# Expected

Requirement scenario **#3** — top-3 domains truncation:

- `count` = 30
- `tabsWatching` = 1
- `domainCount` (global) = 5
- One tab row for tabId `10`
- Per-tab `domainCount` = **5** (full unique hosts)
- Per-tab `domains` length = **3** (not 5)
- `domains` order by count desc:
  1. `a.example.com` count 10
  2. `b.example.com` count 8
  3. `c.example.com` count 6
- Hosts `d.example.com` and `e.example.com` must **not** appear in `domains`
- Invariant sum(requestCount)==count

## Side Effects

- None.

## Errors

- Must not drop domainCount to 3 just because domains is truncated.
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
	assertGlobals(t, stats, 30, 1, 5)
	assertRequestCountInvariant(t, stats)

	if len(stats.Tabs) != 1 {
		t.Fatalf("tabs len = %d, want 1", len(stats.Tabs))
	}
	tab := findTab(t, stats, 10)
	if tab.RequestCount != 30 {
		t.Fatalf("requestCount = %d, want 30", tab.RequestCount)
	}
	if tab.DomainCount != 5 {
		t.Fatalf("tab.domainCount = %d, want 5 (full unique, not truncated)", tab.DomainCount)
	}
	if len(tab.Domains) != 3 {
		t.Fatalf("domains len = %d, want 3; hosts=%v", len(tab.Domains), domainHosts(tab.Domains))
	}

	want := []struct {
		host  string
		count int
	}{
		{"a.example.com", 10},
		{"b.example.com", 8},
		{"c.example.com", 6},
	}
	for i, w := range want {
		if tab.Domains[i].Host != w.host || tab.Domains[i].Count != w.count {
			t.Fatalf("domains[%d] = {%q,%d}, want {%q,%d}",
				i, tab.Domains[i].Host, tab.Domains[i].Count, w.host, w.count)
		}
	}
	for _, banned := range []string{"d.example.com", "e.example.com"} {
		for _, d := range tab.Domains {
			if d.Host == banned {
				t.Fatalf("domains should omit %q (outside top 3); got %v", banned, domainHosts(tab.Domains))
			}
		}
	}
}
```
