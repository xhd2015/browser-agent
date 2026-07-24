# Chrome-Ext-Capture-API popup recording stats (pure builder)

Exercises the **pure popup stats builder** used while the API Capture extension
is **recording**. The builder turns in-memory capture state (entries, attached
tabs, tab metadata) into the enriched `getState` payload: global chips
(tabs watching, unique domains, request count) plus a per-tab breakdown
(request count, domain count, top-3 domains).

**No Chrome, no DOM, no network.** Leaves call a Go package
`popupstats.BuildPopupStats` with the same rules the extension JS must follow
(algorithm parity; JS may reimplement).

This tree does **not** cover popup HTML/CSS chrome, idle UI, or HAR save format.

Product defaults relevant here:

| Setting | Default |
|---------|---------|
| popup width | ~320–340px (UI only; not asserted here) |
| domains per tab shown | **top 3** by count desc |
| closed tabs with entries | remain listed until Stop (`attached=false`) |
| bad / unparseable URL host | single bucket `"opaque"` |
| tab sort | `requestCount` desc; tie-break `tabId` asc |
| title fallback | meta title → else `"Tab {id}"` if attached → else `"Closed tab"` |

## Version

0.0.2

# DSN (Domain Specific Notion)

**User** opens the extension **Popup** while **Recording** is active. The popup
polls `getState` (~1s) and shows summary chips + a “By tab” list.

**Background Service Worker** holds in-memory capture state:

- **Entries** — map key → `{ request.url, … }`; preferred key
  `"${tabId}:${requestId}"` (legacy unscoped keys optional)
- **AttachedTabIds** — set of tab ids with debugger attached (currently watched)
- **TabMeta** — map tabId → `{ title, url, active }` from Chrome tab info when known

**Stats Builder** (`BuildPopupStats`) is a pure function:

```
Entries + AttachedTabIds + TabMeta
  -> PopupStats {
       count,           # total entries
       tabsWatching,    # len(AttachedTabIds)
       domainCount,     # unique hosts across ALL entries (not sum of per-tab)
       tabs[]           # union: attached OR has ≥1 entry
     }
```

Per **Tab row**:

- `requestCount` — entries attributed to that tabId
- `domainCount` — unique hosts on that tab (may be >3)
- `domains` — top **3** `{host, count}` sorted by count desc
- `attached` — tabId ∈ AttachedTabIds
- `active` / `title` / `url` — from TabMeta (with title fallbacks)
- Host from `request.url` hostname; invalid/unparseable → `"opaque"`

**Popup UI** (out of scope for this tree) renders chips + scrollable tab list from
this payload while recording; idle stays simple (no empty stats noise).

**Test Client** builds `Request` fixtures in Setup and calls
`popupstats.BuildPopupStats` via root `Run` — no Chrome process.

## Decision Tree

```
chrome-ext-popup-stats (BuildPopupStats pure function)
├── membership-attached-only/              [tabs appear only because attached]
│   └── empty-entries/                       #1 count=0; watching=1; one zero-row
├── membership-attached-with-entries/      [attached tabs that also have captures]
│   ├── multi-host-unique-global/            #2 multi-tab multi-host; #6 sum invariant
│   ├── top3-domain-truncation/              #3 domains len≤3; domainCount can be >3
│   ├── sort-by-request-count-desc/          #4 tabs ordered by requestCount desc
│   └── invalid-url-opaque-host/             #7 bad URL → host "opaque"; no panic
└── membership-closed-with-entries/        [not attached but still has entries]
    └── not-attached-still-listed/           #5 closed tab listed; attached=false
```

### Parameter significance (high → low)

1. **Tab membership source** — attached-only vs attached+entries vs closed+entries
   (fundamentally changes which rows appear and `tabsWatching` vs list length).
2. **Domain / URL rules** (when entries exist) — multi-host uniqueness, top-3
   truncation, invalid host bucket.
3. **Tab ordering** — `requestCount` desc (+ `tabId` asc tie-break).
4. **Title/meta fallbacks** — asserted where closed/missing meta matters (#5).

## Test Index

| Leaf | Scenario (requirement #) |
|------|--------------------------|
| `membership-attached-only/empty-entries` | (#1) Empty entries, 1 attached tab → count=0, tabsWatching=1, domainCount=0, one row with 0 requests |
| `membership-attached-with-entries/multi-host-unique-global` | (#2) Two tabs, multi-host URLs → per-tab counts/domains; global domainCount is unique hosts. Also (#6) sum(tab.requestCount)==count |
| `membership-attached-with-entries/top3-domain-truncation` | (#3) ≥4 hosts on one tab → `domains` length 3; per-tab `domainCount` >3 |
| `membership-attached-with-entries/sort-by-request-count-desc` | (#4) Tabs sorted by requestCount desc; equal counts use tabId asc |
| `membership-closed-with-entries/not-attached-still-listed` | (#5) Closed tab with entries, not attached → still listed; attached=false; tabsWatching excludes it |
| `membership-attached-with-entries/invalid-url-opaque-host` | (#7) Invalid/bad URL → host `"opaque"`; no panic; domainCount counts the bucket |

## How to Run

```sh
cd tests/chrome-ext-popup-stats
doctest vet .
doctest test -v .
# or from repo root:
doctest vet ./tests/chrome-ext-popup-stats
doctest test ./tests/chrome-ext-popup-stats
```

Requires package `github.com/xhd2015/browser-agent/popupstats` exporting
`BuildPopupStats` (and supporting types). Leaves fail to compile / run red until
the implementer adds that package (TDD red → green). Extension JS should mirror
the same rules when enriching `getState`.

### Expected pure API (implementer)

Prose contract (authoritative rules; types sketched in the Run harness):

```go
package popupstats

type EntryValue struct {
    URL string // request.url
}

type TabMeta struct {
    Title  string
    URL    string
    Active bool
}

type DomainCount struct {
    Host  string
    Count int
}

type TabStats struct {
    TabID        int
    Title        string
    URL          string
    Active       bool
    Attached     bool
    RequestCount int
    DomainCount  int
    Domains      []DomainCount // top 3 only
}

type PopupStats struct {
    Count        int
    TabsWatching int
    DomainCount  int // unique hosts across all entries
    Tabs         []TabStats
}

type Input struct {
    Entries        map[string]EntryValue // key "${tabId}:${requestId}" preferred
    AttachedTabIDs []int
    TabMeta        map[int]TabMeta
}

func BuildPopupStats(in Input) PopupStats
```

**Rules recap:**

| Rule | Detail |
|------|--------|
| Host | Parse hostname from `request.url`; invalid → `"opaque"` |
| Global `domainCount` | Unique hosts across **all** entries |
| `count` | `len(entries)` |
| `tabsWatching` | `len(attachedTabIds)` (unique ids) |
| Per-tab rows | Union of attached **or** ≥1 entry |
| Sort tabs | `requestCount` desc, then `tabId` asc |
| Domains per tab | count desc; top **3** in `domains`; full unique count in `domainCount` |
| Title | TabMeta.Title if non-empty; else if attached `"Tab {id}"`; else `"Closed tab"` |
| Entry→tab | Preferred key `"tabId:requestId"`; tabId is the integer before the first `:` |

```go
import (
	"testing"

	"github.com/xhd2015/browser-agent/popupstats"
)

// Request is narrowed root→leaf by Setup functions.
// Mirrors popupstats.Input so leaves build fixtures without Chrome.
type Request struct {
	// Entries is the capture buffer: key "${tabId}:${requestId}" → {URL}.
	// Empty map / nil means no captures.
	Entries map[string]popupstats.EntryValue

	// AttachedTabIDs are debugger-attached tab ids (watching).
	// Order does not matter; implementer should treat as a set (unique).
	AttachedTabIDs []int

	// TabMeta is optional chrome.tabs-style metadata keyed by tab id.
	TabMeta map[int]popupstats.TabMeta
}

// Response holds the pure builder output.
type Response struct {
	Stats popupstats.PopupStats
}

// Run invokes the pure stats builder. No I/O, no Chrome.
func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	if req.Entries == nil {
		req.Entries = map[string]popupstats.EntryValue{}
	}
	if req.TabMeta == nil {
		req.TabMeta = map[int]popupstats.TabMeta{}
	}
	// Copy attached slice so builder mutation cannot leak across leaves.
	attached := append([]int(nil), req.AttachedTabIDs...)

	stats := popupstats.BuildPopupStats(popupstats.Input{
		Entries:        req.Entries,
		AttachedTabIDs: attached,
		TabMeta:        req.TabMeta,
	})
	return &Response{Stats: stats}, nil
}
```
