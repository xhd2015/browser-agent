# Scenario

**Feature**: `devtools://` pages are not capturable (skip-list alignment)

```
IsCapturableTabURL("devtools://devtools/bundled/inspector.html") -> false
ShouldAttemptAttach(open gates, devtools URL) -> false
```

## Preconditions

- DevTools UI tabs must be skipped (same as `attachAllTabsInWindow`).
- Gates open.

## Steps

1. Set `URL = "devtools://devtools/bundled/inspector.html"`.

## Context

- Requirement goal #1 lists `devtools://` among non-capturable prefixes.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.URL = "devtools://devtools/bundled/inspector.html"
	return nil
}
```
