# Scenario

**Feature**: GET `/go` install/update re-run guidance (close window + re-run)

```
# Session page surfaces re-run guidance after Load unpacked / Reload
User -> browser-trace (NoOpenChrome) -> Control Server
Test Client -> GET /go?session=<id>
Control Server -> HTML body
  (close Chrome window + re-run browser-trace after install/reload)
```

## Preconditions

- Mode is HTTP probe of `/go` (no mock extension required).
- Install panel may also be present; this grouping asserts **re-run** copy, not
  expand/collapse (covered by `browser-trace-install-panel`).

## Steps

1. Set `Mode = ModeGoHTML`.
2. Use a known `SessionSuffix` so probe URL matches the live session id.

## Context

- Requirement scenario #1.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeGoHTML
	if req.SessionSuffix == "" {
		req.SessionSuffix = "go-rerun-test"
	}
	return nil
}
```
