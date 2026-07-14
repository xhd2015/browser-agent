# Scenario

**Feature**: product hardening — extension manifest permission contract + session info control vs browser

```
# Manifest contract
Test Client -> ValidateExtensionManifestJSON(fixture|embedded|shell)
  -> nil when debugger+tabs+alarms+storage + 43761 hosts + broad host
  -> error mentioning missing permission name when absent

# Session info CLI
Operator -> HandleCLI(["session","info", …], env)
  -> always control GET /v1/session
  -> if extension.connected: merge info job (tabs)
  -> if not connected: no fabricated tabs; browser unavailable signal
  -> missing session id: error mentions --session-id and BROWSER_AGENT_SESSION_ID
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` is importable.
- Tree root is `tests/browser-agent-product-hardening/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- RED until implementer exports `ValidateExtensionManifestJSON` and hardens
  `session info` to merge browser `info` job when connected.
- No real Chrome / agent-run; fake WS only for connected session-info leaf.
- Ambient `BROWSER_AGENT_SESSION_ID` process env is ignored when CLIEnv map is set.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Leave Mode empty at root (grouping/leaf Setup sets Mode).
3. Default live-serve knobs (`NoOpenChrome`, `NoAgentRun`, short ReadyTimeout).
4. Helpers below are shared by all leaves.

## Context

- Spec version **0.0.2**.
- Pure fixtures live beside pure-json leaves as `manifest.json`.
- Production FS sources: `browseragent/embedded/extension/manifest.json` and
  `Chrome-Ext-Browser-Agent/public/manifest.json`.
- Session info JSON field names are flexible; asserts accept nested
  `browser.tabs` or equivalent documented shapes.

```go
import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	req.NoOpenChrome = true
	req.NoAgentRun = true
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	if req.MaxDispatchWait == 0 {
		req.MaxDispatchWait = 3 * time.Second
	}
	return nil
}

func assertNoRunErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run transport error: %v", err)
	}
}

func assertExitZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ExitCode != 0 {
		t.Fatalf("ExitCode=%d, want 0; CLIErr=%q ValidateErr=%q stdout=%s stderr=%s",
			resp.ExitCode, resp.CLIErr, resp.ValidateErr,
			truncate(resp.Stdout, 300), truncate(resp.Stderr, 300))
	}
}

func combinedCLIText(resp *Response) string {
	if resp == nil {
		return ""
	}
	return resp.Stdout + resp.Stderr + resp.CLIErr + resp.ErrText
}

func assertPrintedTrailingNewline(t *testing.T, text string) {
	t.Helper()
	if text == "" {
		t.Fatal("expected printed body")
	}
	if !strings.HasSuffix(text, "\n") {
		t.Fatalf("printed body must end with \\n; got tail %q", tail(text, 40))
	}
}

func assertSessionResolveErrorText(t *testing.T, text string) {
	t.Helper()
	if !strings.Contains(text, "--session-id") {
		t.Fatalf("error/text must mention --session-id; got %q", text)
	}
	if !strings.Contains(text, "BROWSER_AGENT_SESSION_ID") {
		t.Fatalf("error/text must mention BROWSER_AGENT_SESSION_ID; got %q", text)
	}
}

func readLeafManifest(t *testing.T, name string) []byte {
	t.Helper()
	// ASSERT/SETUP run with leaf cwd; prefer sibling manifest.json.
	path := name
	if name == "" {
		path = "manifest.json"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read leaf manifest %s: %v", path, err)
	}
	return data
}

func mustValidJSONObject(t *testing.T, data []byte) {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("fixture must be JSON object: %v", err)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

func stdoutLooksLikeJSONObject(s string) bool {
	tr := strings.TrimSpace(s)
	return strings.HasPrefix(tr, "{") && strings.HasSuffix(tr, "}")
}

// fabricatedPageTabs reports whether stdout contains a tabs array with
// http(s) page URLs — the anti-pattern when extension is disconnected.
func fabricatedPageTabs(stdout string, parsed map[string]any) bool {
	if parsed != nil {
		tabs := findTabsSlice(parsed)
		for _, item := range tabs {
			tm, ok := item.(map[string]any)
			if !ok {
				continue
			}
			u, _ := tm["url"].(string)
			if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
				return true
			}
		}
	}
	// Heuristic fallback: JSON-ish "tabs" near example.com / http URLs.
	low := strings.ToLower(stdout)
	if strings.Contains(low, `"tabs"`) &&
		(strings.Contains(low, "http://") || strings.Contains(low, "https://")) {
		// Allow install/hint URLs that mention control host without a tabs array of pages.
		// If tabs key exists with http urls in raw text, treat as fabricated.
		return true
	}
	return false
}
```
