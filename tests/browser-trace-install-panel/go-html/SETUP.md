# Scenario

**Feature**: GET `/go` always includes install panel; expand/collapse from session state

```
# Session page HTML (tests fetch over HTTP only; no DOM)
Test Client -> browsertrace.Run (NoOpenChrome) -> extract on start
Test Client ?-> POST /v1/hello {version, features}
Test Client -> GET /go?session=<id>
Control Server -> HTML with install panel always in DOM:
  - expanded when !(connected && supports)
  - collapsed when connected && supports
  - path + chrome://extensions + markers
```

## Preconditions

- Mode is go HTML probe.
- Session id is the live suffix (known session).
- Product extracts on Run so path/version can appear in the panel.

## Steps

1. Set `Mode = ModeGoHTML` (`"go"`).
2. NoOpenChrome true; known SessionSuffix from root Setup.
3. Descendants stage hello (or not) and assert expand vs collapse.

## Context

- HTML string inspection only — no browser automation.
- Complements embed-extension `go-install-panel/not-connected-panel` (smoke)
  and session-page `go-page` (status UI smoke) without overlapping their leaves.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeGoHTML
	req.NoOpenChrome = true
	return nil
}
```
