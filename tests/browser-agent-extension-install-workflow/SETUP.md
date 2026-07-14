# Scenario

**Feature**: system Chrome + manual Load unpacked extension install workflow

```
Canonical Layout -> EnsureCanonicalExtension() -> extensions/browser-agent/{ver}/
SessionNew -> system openChrome(sessionURL) + enriched stdout
install-chrome-extension -> canonical path (not baseDir/extension/)
open-managed-chrome -> managed profile + Chrome 137 stderr warning
serve / RunDaemon -> never LaunchFn / OpenChromeFn
session info disconnected -> install-chrome-extension + path hints
GET /v1/session -> extension_install_path canonical
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Tree root is `tests/browser-agent-extension-install-workflow/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Package `browseragent` will export canonical APIs (RED until implemented):
  `DefaultExtensionInstallLayout`, `EnsureCanonicalExtension`.
- `browseragent/inject` exports `SessionNewTestHooks` + `ManagedChromeTestHooks`.
- Canonical-path / install leaves set isolated `TestHome` via `t.Setenv("HOME", …)`.
- Daemon leaves use ephemeral `127.0.0.1:0` + temp `BaseDir`.
- No real Chrome; `LaunchFn` / `OpenChromeFn` record calls only.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate temp `BaseDir` and optional `TestHome` per leaf.
3. Default `ReadyTimeout = 5s`, `SessionID = "sess-ext-install-1"`.
4. Leave `Mode` and op-specific fields for grouping/leaf Setup.

## Context

- Spec version **0.0.2**.
- Canonical segment: `extensions/browser-agent/<version>/`.
- Parallel-safe: per-leaf temp dirs; no shared mutable HOME without `TestHome`.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	dir := t.TempDir()
	req.BaseDir = filepath.Join(dir, "browser-agent-base")
	if err := os.MkdirAll(req.BaseDir, 0o755); err != nil {
		return err
	}
	req.TestHome = filepath.Join(dir, "home")
	if err := os.MkdirAll(req.TestHome, 0o755); err != nil {
		return err
	}
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	if req.SessionID == "" {
		req.SessionID = "sess-ext-install-1"
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

func assertCanonicalPathSegment(t *testing.T, path string) {
	t.Helper()
	norm := filepath.ToSlash(path)
	if !strings.Contains(norm, "extensions/browser-agent/") {
		t.Fatalf("path should contain extensions/browser-agent/; got %q", path)
	}
}

func assertArgsNoManagedChrome(t *testing.T, args []string) {
	t.Helper()
	for _, a := range args {
		if a == "--user-data-dir" || strings.HasPrefix(a, "--user-data-dir=") {
			t.Fatalf("system chrome must not include --user-data-dir; args=%v", args)
		}
		if a == "--load-extension" || strings.HasPrefix(a, "--load-extension=") {
			t.Fatalf("system chrome must not include --load-extension; args=%v", args)
		}
	}
}

func assertArgsHasUserDataDir(t *testing.T, args []string) {
	t.Helper()
	for _, a := range args {
		if a == "--user-data-dir" || strings.HasPrefix(a, "--user-data-dir=") {
			return
		}
	}
	t.Fatalf("managed chrome args missing --user-data-dir; args=%v", args)
}

func assertArgsHasNewWindowAndURL(t *testing.T, args []string, wantURLSubstr string) {
	t.Helper()
	hasNewWindow := false
	hasURL := false
	for _, a := range args {
		if a == "--new-window" {
			hasNewWindow = true
		}
		if wantURLSubstr != "" && strings.Contains(a, wantURLSubstr) {
			hasURL = true
		}
	}
	if !hasNewWindow {
		t.Fatalf("args missing --new-window; args=%v", args)
	}
	if wantURLSubstr != "" && !hasURL {
		t.Fatalf("args missing session URL containing %q; args=%v", wantURLSubstr, args)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func findCanonicalExtensionUnderHome(testHome string) (path, version string, ok bool) {
	base := filepath.Join(testHome, ".browser-agent", "managed-chrome", "extensions", "browser-agent")
	entries, err := os.ReadDir(base)
	if err != nil {
		return "", "", false
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		verDir := filepath.Join(base, e.Name())
		if _, err := os.Stat(filepath.Join(verDir, "manifest.json")); err == nil {
			abs, _ := filepath.Abs(verDir)
			if abs == "" {
				abs = verDir
			}
			return abs, e.Name(), true
		}
	}
	return "", "", false
}
```