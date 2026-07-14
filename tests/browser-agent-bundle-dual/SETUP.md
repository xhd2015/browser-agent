# Scenario

**Feature**: browser-agent Bundle fixture pipeline + dual-product coexistence (no npm/Chrome)

```
# Bundle (isolated temp Root)
Test Client -> browseragent.Bundle(UseFixture, Root=temp, Fixture* from ModuleRoot)
  -> ExtensionDir (manifest + 43761) + SessionPageDir (index root mount)
  -> second Bundle: idempotent stable paths

# Dual ProductConfig (Go)
Test Client -> ProductBrowserAgent | ProductBrowserTrace
  -> ports 43761 vs 43759; features; distinct ids

# Dual shells on ModuleRoot disk
Test Client -> read Chrome-Ext-Browser-Agent | Chrome-Ext-Capture-API
  -> manifest ports/names + contentScript page markers

# Dual React products
Test Client -> read react/src/products/{browser-agent,browser-trace}.ts
  -> controlPort 43761 | 43759
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Tree root is `tests/browser-agent-bundle-dual/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Package `browseragent` will export (TDD red until implemented):
  - `Bundle(opts BundleOptions) (*BundleResult, error)` with fixture + Root isolation
  - existing `ProductBrowserAgent` / `ProductBrowserTrace`
- Fixture sources already present under ModuleRoot (embedded mini or
  `tests/browser-agent-cli-react/testdata/mini-extension` +
  `browseragent/embedded/session-page`).
- No npm, no Chrome, no network.
- Sealed prior browser-agent trees must not be modified.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Leave `Mode` and surface-specific fields for grouping/leaf Setup.
3. Bundle leaves allocate **temp BundleRoot** (not ModuleRoot) for isolation.

## Context

- Parallel-safe: Bundle uses per-leaf temp Root; FS leaves are read-only.
- Prefer package `Bundle` over `go run ./script/browser-agent/bundle`.
- Dual product assertions document non-collision: agent **43761**, trace **43759**.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	if st, err := os.Stat(req.ModuleRoot); err != nil || !st.IsDir() {
		t.Fatalf("ModuleRoot %s not a directory: %v", req.ModuleRoot, err)
	}
	return nil
}

func assertNoRunErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run transport/setup error: %v", err)
	}
}

func assertExitZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0; err=%q", resp.ExitCode, resp.ErrText)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func hasRootMount(body string) bool {
	low := strings.ToLower(body)
	return strings.Contains(low, `id="root"`) ||
		strings.Contains(low, `id='root'`) ||
		strings.Contains(low, "data-browser-agent-root") ||
		strings.Contains(low, "browser-agent-root")
}

func resolveFixtureExtensionDir(moduleRoot string) string {
	candidates := []string{
		filepath.Join(moduleRoot, "browseragent", "embedded", "extension"),
		filepath.Join(moduleRoot, "tests", "browser-agent-cli-react", "testdata", "mini-extension"),
		filepath.Join(moduleRoot, "tests", "browser-agent-bundle-dual", "testdata", "mini-extension"),
	}
	for _, c := range candidates {
		if st, err := os.Stat(filepath.Join(c, "manifest.json")); err == nil && !st.IsDir() {
			return c
		}
	}
	return candidates[0]
}

func resolveFixtureSessionPageDir(moduleRoot string) string {
	candidates := []string{
		filepath.Join(moduleRoot, "browseragent", "embedded", "session-page"),
		filepath.Join(moduleRoot, "tests", "browser-agent-bundle-dual", "testdata", "mini-session-page"),
		filepath.Join(moduleRoot, "tests", "browser-agent-vite-skill", "testdata", "session-page"),
	}
	for _, c := range candidates {
		for _, name := range []string{"index.html", "session-page.html"} {
			if st, err := os.Stat(filepath.Join(c, name)); err == nil && !st.IsDir() {
				return c
			}
		}
	}
	return candidates[0]
}
```
