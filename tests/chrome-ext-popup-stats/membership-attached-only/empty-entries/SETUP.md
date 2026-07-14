# Scenario

**Feature**: empty entries with one attached tab (requirement #1)

```
# Just started watching tab 1; no requests yet
Background: Entries={}, AttachedTabIds={1}, TabMeta[1]={title,url,active}
Test Client -> BuildPopupStats
          -> count=0, tabsWatching=1, domainCount=0
          -> tabs=[{tabId:1, requestCount:0, domainCount:0, domains=[], attached:true}]
```

## Preconditions

- Exactly one attached tab id `1`.
- Entries map is empty.
- TabMeta provides a real title/url and `active=true` for tab 1.

## Steps

1. Clear any entries (must remain empty).
2. Attach tab id `1` only.
3. Set TabMeta for tab 1: title `"Home"`, url `https://example.com/`, active true.

## Context

- Expect a single zero-request row so the popup can show “Tabs: 1” without an
  empty-list dead zone while recording has started.

```go
import (
	"testing"

	"github.com/xhd2015/browser-agent/popupstats"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Entries = make(map[string]popupstats.EntryValue)
	attachTabs(req, 1)
	setTabMeta(req, 1, "Home", "https://example.com/", true)
	return nil
}
```
