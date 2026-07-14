# Scenario

**Feature**: tab rows that appear only because the tab is attached (watching)

```
# Debugger attached; capture buffer may be empty
Background: AttachedTabIds non-empty; Entries may be empty
Stats Builder -> tabs include every attached id even when requestCount=0
Stats Builder -> tabsWatching = len(AttachedTabIds)
```

## Preconditions

- Membership factor: **attached-only** — tabs are in the output because they are
  watched, not because they already have captures.
- Child leaves set whether entries are empty and how many tabs are attached.

## Steps

1. Do not add entries at this grouping level (child may keep empty or add later).
2. Leave concrete tab ids and meta to the leaf.

## Context

- Primary product case for “just started recording, nothing captured yet”: chips
  show Tabs=N, Domains=0, Requests=0 with N zero-count rows.

```go
import (
	"testing"

	"github.com/xhd2015/browser-agent/popupstats"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	// Grouping marker only: leaf sets attachTabs + optional meta.
	// Ensure entries stay empty unless a future child adds some.
	if len(req.Entries) != 0 {
		// Reset so this branch stays attached-only / empty-buffer oriented.
		req.Entries = map[string]popupstats.EntryValue{}
	}
	return nil
}
```

