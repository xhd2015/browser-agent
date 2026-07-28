# Scenario

**Feature**: Firefox-Ext-Browser-Agent package scaffold + build shell + canonical extract + install help (P1)

```
# Package layout (repo source of truth)
ModuleRoot/Firefox-Ext-Browser-Agent/public -> manifest (MV3 + gecko.id, no debugger)

# Build shell
Test Client -> BuildFirefoxExtensionShell(root) -> Firefox-Ext-Browser-Agent/build/

# Canonical extract (not managed-chrome)
Test Client -> EnsureCanonicalFirefoxExtensionWithHome(TestHome)
  -> {TestHome}/.browser-agent/extensions/browser-agent-firefox/{ver}/

# Install help CLI
Operator -> install-firefox-extension -> path + about:debugging + Load Temporary Add-on
Operator -> install-firefox-extension -> step 3 open folder; step 4 select manifest.json; about:addons caution; warning:
Operator -> install-firefox-extension --color|--no-color -> ANSI / plain / conflict
Operator -> browser-agent --help -> lists install-firefox-extension

# Chrome unchanged
Test Client -> BuildExtensionShell (Chrome-Ext only)
Operator -> install-chrome-extension still works
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Tree root is `tests/browser-agent-firefox-ext/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Package `browseragent` will export Firefox APIs (RED until implemented):
  `BuildFirefoxExtensionShell`, `EnsureCanonicalFirefoxExtension`,
  `EnsureCanonicalFirefoxExtensionWithHome`, `InstallFirefoxExtension`,
  `InstallFirefoxExtensionWithHome`.
- CLI will gain `install-firefox-extension`; `install-chrome-extension` stays.
- No real Firefox; no network; no Vite/npm in this tree.
- Ensure/install leaves isolate with `TestHome` (WithHome / process-env helpers).
- Build-shell leaves stage a minimal package under temp `ShellRoot` (no shared disk mutation).

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate temp `TestHome` per leaf.
3. Leave `Mode` and op-specific fields for grouping/leaf Setup.

## Context

- Spec version **0.0.2**.
- Canonical Firefox segment: `extensions/browser-agent-firefox/<version>/`.
- Chrome canonical remains under managed-chrome (`extensions/browser-agent/`).
- Parallel-safe: per-leaf temp dirs; no bare process HOME mutation.

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
	dir := t.TempDir()
	req.TestHome = filepath.Join(dir, "home")
	if err := os.MkdirAll(req.TestHome, 0o755); err != nil {
		return err
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

func assertFirefoxCanonicalPathSegment(t *testing.T, path string) {
	t.Helper()
	norm := filepath.ToSlash(path)
	if !strings.Contains(norm, "extensions/browser-agent-firefox/") {
		t.Fatalf("path should contain extensions/browser-agent-firefox/; got %q", path)
	}
	if strings.Contains(norm, "managed-chrome") {
		t.Fatalf("Firefox extension path must not be under managed-chrome; got %q", path)
	}
}

func assertStdoutTrailingNewline(t *testing.T, stdout string) {
	t.Helper()
	if stdout == "" {
		t.Fatal("stdout is empty")
	}
	if !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("stdout must end with \\n; last bytes=%q", tail(stdout, 40))
	}
}

// stageFirefoxPublic writes a minimal Firefox-Ext-Browser-Agent/public under root.
// Used by build-shell public-to-build (isolated temp root; no ModuleRoot mutation).
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
  "host_permissions": ["http://127.0.0.1/*", "http://localhost/*"]
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
```
