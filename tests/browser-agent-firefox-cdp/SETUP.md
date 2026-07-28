# Scenario

**Feature**: Firefox-Ext background CDP job matrix without chrome.debugger (Phase 3 C)

```
# CDP job path (WebExtensions polyfill, no debugger)
Control Server
  -> WS job { type: cdp, params: { method, params } }
  -> Background handleJob
       case "cdp"
         Page.navigate      -> tabs.update({ url })
         Runtime.evaluate   -> eval path (executeScript)
         Target.createTarget (optional) -> create_tab
         other              -> error "not supported" + "firefox"
  -> WS result

# Anti Phase-2 generic unknown
cdp must not only fall through default: unknown job type

# No real Firefox / no network / no chrome.debugger
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Tree root is `tests/browser-agent-firefox-cdp/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Phase 2 core jobs may already be GREEN; this tree is **RED** until Phase 3
  dedicated `cdp` case + method matrix land.
- No real Firefox; no network; no CDP debugger attach.
- Read-only FS probes under `Firefox-Ext-Browser-Agent/` (public preferred, build fallback).

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Leave `Mode` and target fields for grouping/leaf Setup.

## Context

- Spec version **0.0.2**.
- Chrome agent `handleCdpJob` under `Chrome-Ext-Browser-Agent/public/background.js`
  is the **behavioral** reference for method names, but Firefox Phase 3 must
  polyfill with WebExtensions APIs instead of `chrome.debugger.sendCommand`.
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
