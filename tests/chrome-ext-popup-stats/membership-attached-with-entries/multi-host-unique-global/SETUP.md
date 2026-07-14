# Scenario

**Feature**: two attached tabs, multi-host URLs, global unique domains (#2 + #6)

```
# Tab 1: api×2, cdn×1  |  Tab 2: api×1, auth×1
Background: AttachedTabIds={1,2}; Entries across api/cdn/auth
Test Client -> BuildPopupStats
          -> count=5, tabsWatching=2, domainCount=3 (api,cdn,auth unique)
          -> tab1 requestCount=3 domains api,cdn; tab2 requestCount=2 domains api,auth
          -> sum(requestCount)==count
```

## Preconditions

- Two attached tabs: `1` and `2`.
- Hosts used: `api.example.com`, `cdn.example.com`, `auth.example.com`.
- Tab 1 has more requests than tab 2 (sort places 1 before 2).

## Steps

1. Attach tabs `1` and `2`.
2. Add entries:
   - tab1: `https://api.example.com/a`, `https://api.example.com/b`, `https://cdn.example.com/x`
   - tab2: `https://api.example.com/c`, `https://auth.example.com/login`
3. Set TabMeta titles `"App"` / `"Auth"`; tab1 active.

## Context

- Global `domainCount` is **unique hosts across all entries** (3), not
  sum of per-tab domainCounts (2+2=4).
- Also carries requirement **#6** invariant assertion.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	attachTabs(req, 1, 2)

	addEntry(req, 1, "r1", "https://api.example.com/a")
	addEntry(req, 1, "r2", "https://api.example.com/b")
	addEntry(req, 1, "r3", "https://cdn.example.com/x")

	addEntry(req, 2, "r4", "https://api.example.com/c")
	addEntry(req, 2, "r5", "https://auth.example.com/login")

	setTabMeta(req, 1, "App", "https://app.example.com/", true)
	setTabMeta(req, 2, "Auth", "https://auth.example.com/", false)
	return nil
}
```
