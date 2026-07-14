# Scenario

**Bug**: Jobs pinned to session control page tab instead of active user tab in session window

```
# static: read background.js pickTargetTabIdForSession
Test Client -> assert active+windowId priority, session-page fallback

# e2e: real browser two-tab scenario
RunDaemon -> POST /v1/sessions -> playwright-debug --extension
Tab1 /go?session=S; Tab2 user URL active -> eval returns user tab URL
```

## Preconditions

- **ModuleRoot** = workspace root (`filepath.Join(DOCTEST_ROOT, "..", "..")`).
- Fix in `pickTargetTabIdForSession` already landed — **GREEN** expected (coverage backfill).
- Ext-source leaves: no browser; read `Chrome-Ext-Browser-Agent/public/background.js`
  (also accept `build/` / `src/` fallbacks).
- E2e leaves: `playwright-debug` on PATH; Chromium for Playwright; skip when tool absent.
- Daemon phases 1–9 (`RunDaemon`, `POST /v1/sessions`, per-session extension WS).

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Grouping Setup sets `Mode` (`ext-source` or `e2e`).
3. E2e grouping allocates temp `BaseDir` and default timeouts.
4. Leaf Setup sets `ExtSourceTarget` or `PlaywrightOp` + `SessionID`.

## Context

- Spec version **0.0.2**.
- Complements phase9 `job-session-routing` (session-scoped pick) with **active-tab-in-window** priority.
- Playwright uses `--headed` (MV3 extensions may not load in classic headless).

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	return nil
}

func assertNoRunErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run transport/setup error: %v", err)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func hasActiveTabInWindowQuery(text string) bool {
	low := strings.ToLower(text)
	if !strings.Contains(low, "picktargettabidforsession") {
		return false
	}
	hasActive := strings.Contains(low, "active: true") || strings.Contains(low, "active:true")
	hasWindow := strings.Contains(low, "windowid") &&
		(strings.Contains(low, "entry.windowid") || strings.Contains(low, "windowid: entry"))
	return hasActive && hasWindow
}

func hasSessionPageFallback(text string) bool {
	low := strings.ToLower(text)
	hasTabId := strings.Contains(low, "entry.tabid")
	hasGoURL := strings.Contains(low, "/go?session=") || strings.Contains(low, "go?session=")
	return hasTabId && hasGoURL
}
```