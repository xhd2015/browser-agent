# Scenario

**Feature**: GET `/go` HTML session dashboard with status UI hooks

```
# Session page opened in Chrome window (tests fetch HTML over HTTP only)
Test Client -> GET /go?session=<id>
Control Server -> text/html dashboard
  - shows session id
  - status UI root (id or data-browser-trace-status)
  - inline JS polls /v1/session
```

## Preconditions

- Probe target is the HTML page (not JSON).
- Valid session id is used (unknown-id HTML is out of scope for this tree).

## Steps

1. Set `Probe = ProbeGoHTML` (`"go"`).
2. Descendants set known session and any optional staging (default: no hello required for smoke).

## Context

- Smoke asserts only: no browser DOM execution; body string inspection is enough.
- Poller interval (~500ms–1s) need not be exact; the HTML must reference `/v1/session`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Probe = ProbeGoHTML
	req.ForceUnknownSession = false
	return nil
}
```
