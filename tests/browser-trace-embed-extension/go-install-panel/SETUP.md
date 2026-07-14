# Scenario

**Feature**: GET /go HTML shows install panel when extension not connected

```
# Session page in Chrome (tests fetch HTML only)
Test Client -> browsertrace.Run (NoOpenChrome) -> extract on start
Test Client -> GET /go?session=<id>
Control Server -> HTML with install panel:
  - absolute extension path
  - chrome://extensions as text
  - stable panel markers (data-browser-trace-install / id)
```

## Preconditions

- Mode is go HTML probe.
- Not connected (no hello) for the primary panel leaf.
- chrome: URL links from http origin are blocked — text + copy affordance is enough.

## Steps

1. Set `Mode = ModeGoHTML` (`"go"`).
2. NoOpenChrome true; known SessionSuffix.

## Context

- Smoke string asserts only; no browser DOM automation.
- Complements go-page smoke in browser-trace-session-page (which does not assert install panel).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeGoHTML
	req.NoOpenChrome = true
	req.DoHello = false
	return nil
}
```
