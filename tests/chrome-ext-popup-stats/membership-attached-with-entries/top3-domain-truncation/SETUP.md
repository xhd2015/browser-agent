# Scenario

**Feature**: per-tab domains array truncated to top 3 (requirement #3)

```
# One attached tab with 5 distinct hosts (counts 10,8,6,4,2)
Background: AttachedTabIds={10}; 30 entries across 5 hosts
Test Client -> BuildPopupStats
          -> tab.domainCount=5
          -> tab.domains length=3 (hosts a,b,c only; d and e omitted)
          -> global domainCount=5
```

## Preconditions

- Single attached tab id `10`.
- Five hosts with strictly decreasing counts so top-3 order is unambiguous:
  - `a.example.com` ×10
  - `b.example.com` ×8
  - `c.example.com` ×6
  - `d.example.com` ×4
  - `e.example.com` ×2
- Total entries = 30; global unique domains = 5.

## Steps

1. Attach tab `10` with title `"Busy"`.
2. Add 10+8+6+4+2 entries with the hosts above (unique request ids).

## Context

- `domains` is a **display** slice (top 3); `domainCount` remains the full unique
  host count so the UI can show “+N more” if desired.

```go
import (
	"fmt"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	attachTabs(req, 10)
	setTabMeta(req, 10, "Busy", "https://a.example.com/", true)

	hosts := []struct {
		host  string
		count int
	}{
		{"a.example.com", 10},
		{"b.example.com", 8},
		{"c.example.com", 6},
		{"d.example.com", 4},
		{"e.example.com", 2},
	}
	n := 0
	for _, h := range hosts {
		for i := 0; i < h.count; i++ {
			n++
			addEntry(req, 10, fmt.Sprintf("req-%d", n), "https://"+h.host+"/p/"+fmt.Sprint(i))
		}
	}
	return nil
}
```
