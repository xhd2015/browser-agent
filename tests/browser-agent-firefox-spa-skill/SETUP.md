# Scenario

**Feature**: Firefox session-page install UX + skill docs (Phase 4 D)

```
# React install panel
Test Client -> read react/src/ui/InstallGuideline.tsx
  when firefox: about:debugging + temporary add-on + browser-agent-firefox
  when chrome: chrome://extensions preserved

# SessionPageApp wiring
Test Client -> read react/src/ui/SessionPageApp.tsx (+ InstallGuideline)
  wires browser firefox into install UI

# Go inject / fallback
Test Client -> read browseragent/server.go injectSessionBoot / writeFallbackSessionHTML
  firefox install markers vs chrome://extensions preserved

# Skill docs
Test Client -> read browseragent/SKILL.md (cmd/browser-agent/SKILL.md)
  install-firefox-extension + about:debugging + session new --browser firefox
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Tree root is `tests/browser-agent-firefox-spa-skill/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- No real browser; no network; no npm — read-only FS probes.
- Phase 1–3 Firefox trees may be GREEN; this tree is **RED** for Firefox SPA /
  skill markers until Phase 4 implementer lands them.
- Chrome-preservation leaves must stay GREEN (do not remove Chrome install path).

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Leave `Mode` and surface-specific probe fields for grouping/leaf Setup.

## Context

- Spec version **0.0.2**.
- Parallel-safe: ModuleRoot reads only; no temp mutation required.
- Shared helpers below available to all descendant Assert/Setup packages.

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

func assertFilePresent(t *testing.T, req *Request, resp *Response, label string) string {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.FileExists || strings.TrimSpace(resp.CombinedText) == "" {
		t.Fatalf("%s source missing under ModuleRoot=%s; err=%q found=%v",
			label, req.ModuleRoot, resp.ErrText, resp.FoundPaths)
	}
	return resp.CombinedText
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// hasAboutDebugging reports about:debugging presence (URL fragment optional).
func hasAboutDebugging(s string) bool {
	return strings.Contains(s, "about:debugging")
}

// hasTemporaryAddon reports temporary-add-on install wording.
func hasTemporaryAddon(s string) bool {
	low := strings.ToLower(s)
	return strings.Contains(low, "temporary add-on") ||
		strings.Contains(low, "temporary addon") ||
		strings.Contains(low, "load temporary add-on") ||
		strings.Contains(low, "load temporary addon")
}

// hasFirefoxExtPath reports canonical firefox extension install segment.
func hasFirefoxExtPath(s string) bool {
	return strings.Contains(s, "browser-agent-firefox")
}

// hasChromeExtensionsURL reports Chrome install primary path.
func hasChromeExtensionsURL(s string) bool {
	return strings.Contains(s, "chrome://extensions")
}
```
