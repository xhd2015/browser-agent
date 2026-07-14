# Expected

Requirement scenario **#7** — invalid / bad URL host bucket:

- Run returns `(resp, nil)` — **no error**, no panic
- `count` = 5
- `tabsWatching` = 1
- Global `domainCount` = **2** (`ok.example.com` + `opaque`)
- Tab 3: `requestCount` = 5, `domainCount` = 2
- Domains include:
  - `ok.example.com` with count **2**
  - `opaque` with count **3** (three bad URLs)
- Domains sorted by count desc: `opaque` (3) before `ok.example.com` (2)
- Invariant sum == count

## Side Effects

- None.

## Errors

- Must not surface parse errors to the caller.
- Must not invent separate buckets per distinct bad string.

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
	assertGlobals(t, stats, 5, 1, 2)
	assertRequestCountInvariant(t, stats)

	tab := findTab(t, stats, 3)
	if tab.RequestCount != 5 {
		t.Fatalf("requestCount = %d, want 5", tab.RequestCount)
	}
	if tab.DomainCount != 2 {
		t.Fatalf("domainCount = %d, want 2 (ok + opaque)", tab.DomainCount)
	}
	if len(tab.Domains) != 2 {
		t.Fatalf("domains len = %d, want 2; got %v", len(tab.Domains), domainHosts(tab.Domains))
	}

	var opaqueCount, okCount int
	var foundOpaque, foundOK bool
	for _, d := range tab.Domains {
		switch d.Host {
		case OpaqueHost:
			foundOpaque = true
			opaqueCount = d.Count
		case "ok.example.com":
			foundOK = true
			okCount = d.Count
		default:
			t.Fatalf("unexpected host %q in domains", d.Host)
		}
	}
	if !foundOpaque {
		t.Fatalf("missing host %q in domains %v", OpaqueHost, domainHosts(tab.Domains))
	}
	if !foundOK {
		t.Fatalf("missing host ok.example.com in domains %v", domainHosts(tab.Domains))
	}
	if opaqueCount != 3 {
		t.Fatalf("opaque count = %d, want 3", opaqueCount)
	}
	if okCount != 2 {
		t.Fatalf("ok.example.com count = %d, want 2", okCount)
	}
	// count desc: opaque (3) first
	if tab.Domains[0].Host != OpaqueHost || tab.Domains[0].Count != 3 {
		t.Fatalf("domains[0] = {%q,%d}, want {%q,3}", tab.Domains[0].Host, tab.Domains[0].Count, OpaqueHost)
	}
}
```
