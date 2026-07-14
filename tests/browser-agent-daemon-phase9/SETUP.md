# Scenario

**Feature**: Phase 9 extension per-session WS + tab routing (static source contracts)

```
# no Chrome — read extension sources under ModuleRoot
Test Client -> read Chrome-Ext-Browser-Agent public/ background.js, contentScript.js, manifest.json
Test Client -> assert per-session WS, register map, job routing, go-page content_scripts
```

## Preconditions

- Tree is **RED** until implementer lands phase 9 extension changes.
- **ModuleRoot** = workspace root (`filepath.Join(DOCTEST_ROOT, "..", "..")`).
- No real browser; filesystem reads only.
- Prefer `Chrome-Ext-Browser-Agent/public/` sources (also accept build/src fallbacks).

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Leave `Mode` and surface-specific fields for grouping/leaf Setup.

## Context

- Spec version **0.0.2**.
- Embedded copy under `browseragent/embedded/extension/` should mirror `public/`
  when build copies from shell (not asserted in this tree; implementer responsibility).

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

func assertContainsFold(t *testing.T, haystack string, needles ...string) {
	t.Helper()
	low := strings.ToLower(haystack)
	for _, n := range needles {
		if !strings.Contains(low, strings.ToLower(n)) {
			t.Fatalf("expected text to contain %q; got:\n%s", n, truncate(haystack, 800))
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func hasPerSessionWSURL(text string) bool {
	low := strings.ToLower(text)
	if strings.Contains(text, "/v1/ws?session=") {
		return true
	}
	if strings.Contains(low, "?session=") && strings.Contains(low, "/v1/ws") {
		return true
	}
	// Template building session into WS path, e.g. `/v1/ws?session=${id}`
	if strings.Contains(low, "/v1/ws") && strings.Contains(low, "session") &&
		(strings.Contains(text, "?session=") || strings.Contains(text, "&session=") ||
			strings.Contains(text, "+ \"?session=\"") || strings.Contains(text, "`?session=")) {
		return true
	}
	return false
}

func hasRegisterHandler(text string) bool {
	low := strings.ToLower(text)
	return strings.Contains(low, "register") &&
		(strings.Contains(low, "onmessage") || strings.Contains(low, "runtime.onmessage") ||
			strings.Contains(low, "sendmessage"))
}

func hasSessionsMap(text string) bool {
	low := strings.ToLower(text)
	return strings.Contains(low, "sessions") &&
		(strings.Contains(text, "Map") || strings.Contains(low, "sessions[") ||
			strings.Contains(low, "sessions.get") || strings.Contains(low, "sessions.set"))
}
```