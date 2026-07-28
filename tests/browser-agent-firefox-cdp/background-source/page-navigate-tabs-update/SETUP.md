# Scenario

**Feature**: CDP Page.navigate maps to tabs.update on Firefox

```
handleCdpJob / cdp branch
  method == "Page.navigate"
    -> browser.tabs.update(tabId, { url: params.url })
    # no chrome.debugger sendCommand required
```

## Preconditions

- BackgroundSourceTarget = page-navigate-tabs-update.

## Steps

1. Set `BackgroundSourceTarget = BgSrcPageNavigateTabsUpdate`.

## Context

- Source markers only: `Page.navigate` string + `tabs.update` API token.
- URL field name in params is implementer choice (`url` preferred).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BackgroundSourceTarget = BgSrcPageNavigateTabsUpdate
	return nil
}
```
