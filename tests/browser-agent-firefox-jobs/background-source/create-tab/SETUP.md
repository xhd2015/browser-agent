# Scenario

**Feature**: create_tab job uses browser.tabs.create in session window

```
handleJob type=create_tab
  -> browser.tabs.create({ windowId, url?, active? })
  -> result { type: "create_tab", tab_id, … }
```

## Preconditions

- BackgroundSourceTarget = create-tab.

## Steps

1. Set `BackgroundSourceTarget = BgSrcCreateTab`.

## Context

- `browser.tabs.create`, `chrome.tabs.create`, or `.tabs.create` all accepted.
- No Target.* CDP polyfill required in Phase 2 (Phase 3).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BackgroundSourceTarget = BgSrcCreateTab
	return nil
}
```
