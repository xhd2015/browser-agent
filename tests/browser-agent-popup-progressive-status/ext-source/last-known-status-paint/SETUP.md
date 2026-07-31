# Scenario

**Feature**: popup paints last-known status from storage (or SW query) promptly

```
popup open
  -> chrome.storage.session|local.get(last status)  # and/or sendMessage getStatus
  -> paint phase rows from last-known
  -> in parallel: live health + status refresh
  -> never block first status paint solely on network
```

## Preconditions

- ExtSourceTarget = last-known-status-paint.

## Steps

1. Set `ExtSourceTarget = ExtSrcLastKnownStatusPaint`.

## Context

- Current popup only `fetch`es health — no storage read, no SW status query —
  **RED** until implementer adds last-known path.
- `chrome.storage.session.get` / `local.get` preferred; `runtime.sendMessage`
  status query is acceptable when SW holds last-known.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcLastKnownStatusPaint
	return nil
}
```
