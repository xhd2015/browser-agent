# Scenario

**Feature**: Firefox connect agent (register + WS hello + keepalive) and bundle embed stage (Phase 1 A + E1)

```
# Content script register (source contract)
Session Page /go?session=S
  -> Firefox contentScript reads session_id
  -> browser.runtime.sendMessage({type:"register", session_id:S, control_port, …})
  -> keeps window.__BROWSER_AGENT_EXT__

# Background WS hello + keepalive
Background on register
  -> sessions[S] = {ws, …}
  -> WebSocket(ws://host:port/v1/ws?session=S)
  -> hello { browser_product:"firefox", features include browser-agent }
  -> keepalive ping / reconnect (not no-op-only alarm)

# Bundle E1
Test Client -> BuildFirefoxExtensionShell + Bundle/StageFirefox
  -> {Root}/browseragent/embedded/extension-firefox/manifest.json

# Chrome regression
Test Client -> BuildExtensionShell -> Chrome-Ext-Browser-Agent/build only
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Tree root is `tests/browser-agent-firefox-connect/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Firefox-Ext public is currently a **P1 stub** (marker + no-op alarm) — leaves are
  **RED** until connect path + embed stage land.
- No real Firefox; no network; no Vite/npm in this tree.
- Bundle leaf uses temp `BundleRoot` / `ShellRoot` (no live embed mutation).
- Chrome regression stages minimal Chrome-Ext public under temp `ShellRoot`.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Leave `Mode` and op-specific fields for grouping/leaf Setup.

## Context

- Spec version **0.0.2**.
- Chrome agent protocol under `Chrome-Ext-Browser-Agent/public/` is the port source
  of truth for register / WS hello / keepalive (not re-tested here).
- Parallel-safe: per-leaf temp dirs; ModuleRoot reads only for ext-source leaves.

```go
import (
	"os"
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

// stageFirefoxPublic writes a minimal Firefox-Ext-Browser-Agent/public under root.
// Used by bundle-stages leaf (isolated temp root; no ModuleRoot mutation).
func stageFirefoxPublic(t *testing.T, root string) string {
	t.Helper()
	public := filepath.Join(root, "Firefox-Ext-Browser-Agent", "public")
	if err := os.MkdirAll(public, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{
  "manifest_version": 3,
  "name": "Browser Agent Firefox Fixture",
  "version": "1.0.1",
  "browser_specific_settings": {
    "gecko": { "id": "browser-agent@xhd2015" }
  },
  "background": { "scripts": ["background.js"] },
  "permissions": ["tabs", "storage", "alarms"],
  "host_permissions": [
    "http://127.0.0.1/*",
    "http://localhost/*",
    "ws://127.0.0.1/*",
    "ws://localhost/*"
  ],
  "content_scripts": [{
    "matches": ["http://127.0.0.1/*", "http://localhost/*"],
    "js": ["contentScript.js"],
    "run_at": "document_start"
  }]
}
`
	files := map[string]string{
		"manifest.json":    manifest,
		"background.js":    "// firefox fixture background\n",
		"contentScript.js": "// firefox fixture content\n",
		"popup.html":       "<!doctype html><title>ff-fixture</title>\n",
		"popup.js":         "// firefox fixture popup\n",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(public, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return public
}

// stageChromePublic writes a minimal Chrome-Ext-Browser-Agent/public under root.
func stageChromePublic(t *testing.T, root string) string {
	t.Helper()
	public := filepath.Join(root, "Chrome-Ext-Browser-Agent", "public")
	if err := os.MkdirAll(public, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"manifest_version":3,"name":"Chrome Fixture","version":"1.0.0"}`
	if err := os.WriteFile(filepath.Join(public, "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(public, "background.js"), []byte("// chrome fixture\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return public
}

// copyTreeIfExists copies src→dst recursively when src exists (best-effort fixture stage).
func copyTreeIfExists(t *testing.T, src, dst string) error {
	t.Helper()
	st, err := os.Stat(src)
	if err != nil || !st.IsDir() {
		return err
	}
	return filepath.Walk(src, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dst, 0o755)
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

// hasHelloBrowserProductFirefox reports hello + firefox product markers in background source.
func hasHelloBrowserProductFirefox(text string) bool {
	low := strings.ToLower(text)
	if !strings.Contains(low, "hello") {
		return false
	}
	if strings.Contains(low, "browser_product") {
		return true
	}
	// Accept literal product string without key if clearly firefox identity in hello payload.
	return strings.Contains(low, "firefox")
}

// hasPerSessionWSURL mirrors daemon-phase9 helper for Firefox background.
func hasPerSessionWSURL(text string) bool {
	low := strings.ToLower(text)
	if strings.Contains(text, "/v1/ws?session=") {
		return true
	}
	if strings.Contains(low, "?session=") && strings.Contains(low, "/v1/ws") {
		return true
	}
	if strings.Contains(low, "/v1/ws") && strings.Contains(low, "session") &&
		(strings.Contains(text, "?session=") || strings.Contains(text, "&session=") ||
			strings.Contains(text, "+ \"?session=\"") || strings.Contains(text, "`?session=")) {
		return true
	}
	return false
}

// hasRealKeepalive reports reconnect/keepalive/ping that is not only a no-op alarm stub.
func hasRealKeepalive(text string) bool {
	low := strings.ToLower(text)
	hasPing := strings.Contains(low, "\"ping\"") || strings.Contains(low, "type: \"ping\"") ||
		strings.Contains(low, "type:\"ping\"") || strings.Contains(low, "type = \"ping\"") ||
		(strings.Contains(low, "ping") && (strings.Contains(low, "pong") || strings.Contains(low, "keepalive")))
	hasKeepaliveWord := strings.Contains(low, "keepalive") || strings.Contains(low, "keep_alive") ||
		strings.Contains(low, "keepalivetimer") || strings.Contains(low, "keepalive_ms")
	hasReconnect := strings.Contains(low, "reconnect") || strings.Contains(low, "connectsession")
	// Real keepalive: interval/timer ping while open, or reconnect on dead sockets with ping.
	if hasPing && (hasKeepaliveWord || hasReconnect || strings.Contains(low, "setinterval")) {
		return true
	}
	if hasKeepaliveWord && hasReconnect {
		return true
	}
	// Alarm that actually reconnects (connectSession / WebSocket) is acceptable.
	if strings.Contains(low, "alarms") && hasReconnect && strings.Contains(low, "websocket") {
		return true
	}
	return false
}

// isNoOpOnlyAlarmStub detects the P1 stub pattern: alarm with empty/no-op body only.
func isNoOpOnlyAlarmStub(text string) bool {
	low := strings.ToLower(text)
	if !strings.Contains(low, "alarm") {
		return false
	}
	// Stub marker from current Firefox background.
	if strings.Contains(low, "no-op") || strings.Contains(low, "noop") {
		// If real keepalive also present, not "only" stub.
		if hasRealKeepalive(text) {
			return false
		}
		return true
	}
	// Alarm without WebSocket / ping / reconnect is insufficient alone.
	if !strings.Contains(low, "websocket") && !hasRealKeepalive(text) {
		return strings.Contains(low, "periodinminutes")
	}
	return false
}
```
