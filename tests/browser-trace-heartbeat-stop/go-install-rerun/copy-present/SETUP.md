# Scenario

**Feature**: `/go` HTML includes close-window + re-run `browser-trace` copy

```
# Probe session page without extension connected
Test Client -> GET /go?session=go-rerun-test
Control Server -> 200 text/html
HTML -> close (Chrome) window + re-run browser-trace
       after Load unpacked / Reload / install
# Optional marker
HTML ?-> data-install-rerun-guidance
```

## Preconditions

- Server is up with known session id (`SessionSuffix`).
- No hello required for re-run copy (guidance is always useful after install).

## Steps

1. Keep ModeGoHTML and SessionSuffix from parent.
2. No mock extension; cancel Run after GET.

## Context

- Requirement scenario #1 — install/update re-run guidance on session page.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	// Parent already set Mode + SessionSuffix.
	req.SessionSuffix = "go-rerun-test"
	return nil
}
```
