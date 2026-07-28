# Scenario

**Feature**: info job lists session tabs with Chrome-like tabs shape

```
handleJob type=info
  -> handleInfoJob (preferred) or inline info branch
  -> tabs.query / list capturable tabs in session window
  -> data.tabs[] with id, url, title (… optional active/role/index)
```

## Preconditions

- BackgroundSourceTarget = info-tabs-shape.

## Steps

1. Set `BackgroundSourceTarget = BgSrcInfoTabsShape`.

## Context

- Shape markers are source-level (no runtime job execution).
- Prefer function name `handleInfoJob` when implementer factors Chrome-style handlers.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BackgroundSourceTarget = BgSrcInfoTabsShape
	return nil
}
```
