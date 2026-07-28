# Scenario

**Feature**: stamp firefox browser + extension path on session create (Phase 2)

```
# SessionNew firefox stamps durable identity
SessionNew(Browser=firefox, Home=TestHome, NoOpenChrome, NoWait)
  -> meta.json browser=firefox + extension_install_path …/browser-agent-firefox/…
  -> GET /v1/session extension_install_path same firefox segment

# Boot inject follows stamp
GET /go?session=<id> -> boot JSON / __BROWSER_AGENT.browser == "firefox"

# Session info next steps follow stamp
session info (disconnected) -> install-firefox-extension / about:debugging + path

# Chrome regression
SessionNew(Browser="", …) -> path is Chrome browser-agent tree (not browser-agent-firefox)
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Tree root is `tests/browser-agent-session-firefox-stamp/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- P1 firefox extract and P2 SessionNew Browser/OpenFirefoxFn are available;
  this tree asserts **stamp** outcomes only.
- No real Firefox/Chrome; stamp leaves use `NoOpenChrome` + `NoWait`.
- Per-leaf `TestHome` / `BaseDir` isolation (parallel-safe).

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate temp `BaseDir` and `TestHome` per leaf.
3. Default `SessionID = "sess-ff-stamp-1"`, `ReadyTimeout = 5s`.
4. Leave `Mode` and op-specific fields for grouping/leaf Setup.

## Context

- Spec version **0.0.2**.
- Canonical Firefox segment: `extensions/browser-agent-firefox/<version>/`.
- Chrome install segment: `extensions/browser-agent/` (may sit under managed-chrome).
- Shared assertion helpers below available to all descendant Assert packages.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ModuleRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", ".."))
	dir := t.TempDir()
	req.BaseDir = filepath.Join(dir, "browser-agent-base")
	req.TestHome = filepath.Join(dir, "home")
	if err := os.MkdirAll(req.BaseDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(req.TestHome, 0o755); err != nil {
		return err
	}
	if req.SessionID == "" {
		req.SessionID = "sess-ff-stamp-1"
	}
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 5 * time.Second
	}
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

func assertNotContainsFold(t *testing.T, haystack string, needles ...string) {
	t.Helper()
	low := strings.ToLower(haystack)
	for _, n := range needles {
		if strings.Contains(low, strings.ToLower(n)) {
			t.Fatalf("expected text NOT to contain %q; got:\n%s", n, truncate(haystack, 800))
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// assertFirefoxInstallPath requires browser-agent-firefox segment and rejects
// managed-chrome as the install tree.
func assertFirefoxInstallPath(t *testing.T, path string) {
	t.Helper()
	if strings.TrimSpace(path) == "" {
		t.Fatal("extension_install_path is empty")
	}
	norm := filepath.ToSlash(path)
	if !strings.Contains(norm, "browser-agent-firefox") {
		t.Fatalf("extension_install_path must contain browser-agent-firefox; got %q", path)
	}
	if strings.Contains(norm, "managed-chrome") {
		t.Fatalf("extension_install_path must not be under managed-chrome; got %q", path)
	}
	// Chrome-only segment without -firefox must not be the sole identity.
	if strings.Contains(norm, "/browser-agent/") && !strings.Contains(norm, "browser-agent-firefox") {
		t.Fatalf("extension_install_path looks like Chrome browser-agent tree; got %q", path)
	}
}

// assertChromeInstallPath requires Chrome browser-agent segment, not firefox.
func assertChromeInstallPath(t *testing.T, path string) {
	t.Helper()
	if strings.TrimSpace(path) == "" {
		t.Fatal("extension_install_path is empty (chrome create should stamp a path)")
	}
	norm := filepath.ToSlash(path)
	if strings.Contains(norm, "browser-agent-firefox") {
		t.Fatalf("chrome session must not stamp browser-agent-firefox path; got %q", path)
	}
	// Accept extensions/browser-agent/ (managed-chrome or otherwise).
	if !strings.Contains(norm, "browser-agent") {
		t.Fatalf("chrome extension_install_path should mention browser-agent; got %q", path)
	}
}

// assertMetaBrowserFirefox accepts browser field or browsers list signal already
// normalized into MetaBrowser by Run.
func assertMetaBrowserFirefox(t *testing.T, metaBrowser, metaJSON string) {
	t.Helper()
	if strings.EqualFold(strings.TrimSpace(metaBrowser), "firefox") {
		return
	}
	// Raw meta fallback: "browser":"firefox"
	low := strings.ToLower(metaJSON)
	if strings.Contains(low, `"browser"`) && strings.Contains(low, "firefox") {
		return
	}
	t.Fatalf("meta.json must stamp browser firefox (field browser or browsers); MetaBrowser=%q meta=%s",
		metaBrowser, truncate(metaJSON, 600))
}
```
