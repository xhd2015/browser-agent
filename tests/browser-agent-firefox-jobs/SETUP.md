# Scenario

**Feature**: Firefox-Ext background core jobs without chrome.debugger (Phase 2 B)

```
# Job path (WebExtensions APIs, no debugger)
Control Server
  -> WS job { type: info|eval|run|logs|screenshot|create_tab }
  -> Background handleJob
       info        -> tabs list shape
       create_tab  -> tabs.create
       eval|run    -> executeScript
       screenshot  -> captureVisibleTab
       logs        -> best-effort entries (empty OK)
  -> WS result

# Anti Phase-1 stub
handleJob must not be only "not implemented: firefox job runner (phase 1…)"

# Manifest
scripting permission when using scripting.executeScript;
host_permissions for page injection/capture beyond localhost-only
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Tree root is `tests/browser-agent-firefox-jobs/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Phase 1 connect is GREEN; `handleJob` may still be Phase-1 stub — this tree is
  **RED** until Phase 2 job handlers land.
- No real Firefox; no network; no CDP attach.
- Read-only FS probes under `Firefox-Ext-Browser-Agent/` (public preferred, build fallback).

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Leave `Mode` and target fields for grouping/leaf Setup.

## Context

- Spec version **0.0.2**.
- Chrome agent job handlers under `Chrome-Ext-Browser-Agent/public/background.js`
  are the **behavioral** reference, but Firefox Phase 2 must use WebExtensions
  APIs (`tabs` / `scripting`) instead of `chrome.debugger`.
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

func assertManifestPresent(t *testing.T, req *Request, resp *Response) string {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.FileExists || strings.TrimSpace(resp.CombinedText) == "" {
		t.Fatalf("Firefox manifest missing under ModuleRoot=%s; err=%q found=%v",
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
