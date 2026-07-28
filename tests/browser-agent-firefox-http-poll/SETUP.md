# Scenario

**Feature**: Firefox-Ext background uses HTTP poll (`/v1/ext/hello|poll|result`) as transport/fallback (not WS-only)

```
# Firefox HTTP poll transport (Phase 2 design)
Content Script register
  -> Background sessions[S]
  -> POST /v1/ext/hello { session_id, browser_product: firefox, … }
  -> loop POST /v1/ext/poll { session_id, wait_ms }
       jobs[]   -> handleJob -> POST /v1/ext/result
       events[] -> prepare_reconnect (schedule close + reconnect)
  -> (optional) try WS first; on fail fall back to HTTP poll

# Anti WS-only
background must not only open WebSocket(/v1/ws) without HTTP poll path

# No real Firefox / no network / no live control server
# Server routes covered by tests/browser-agent-ext-http-poll
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Tree root is `tests/browser-agent-firefox-http-poll/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Server `POST /v1/ext/*` may already be GREEN; this tree is **RED** until
  Firefox **client** background uses those routes.
- No real Firefox; no network; no bound control port in this tree.
- Read-only FS probes under `Firefox-Ext-Browser-Agent/` (public preferred, build fallback).

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Leave `Mode` and target fields for grouping/leaf Setup.

## Context

- Spec version **0.0.2**.
- Server contract reference: `tests/browser-agent-ext-http-poll` +
  `browseragent/ext_http_poll.go`.
- Chrome remains WebSocket; this tree only probes **Firefox-Ext** sources.
- Parallel-safe: ModuleRoot reads only; no temp mutation required.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ModuleRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", ".."))
	return nil
}

func assertNoRunErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run transport/setup error: %v", err)
	}
}

func assertBackgroundPresent(t *testing.T, req *Request, resp *Response) string {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.FileExists || strings.TrimSpace(resp.CombinedText) == "" {
		t.Fatalf("Firefox background missing under ModuleRoot=%s; err=%q found=%v",
			req.ModuleRoot, resp.ErrText, resp.FoundPaths)
	}
	return resp.CombinedText
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
```
