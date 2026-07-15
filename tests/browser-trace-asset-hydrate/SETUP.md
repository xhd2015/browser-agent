# Scenario

**Feature**: P6 browser-trace asset hydrate (extension completeness + EnsureAsset + assets CLI)

```
# Completeness
Test Client -> openFixtureFS(extension fixture | empty)
  -> browseragent.EmbedCompleteFS(fs, "extension")
  -> bool

# Download
Test Client -> httptest + XDG_CACHE_HOME temp
  -> browseragent.EnsureAsset(ctx, "browser-trace", version, "extension", cfg)
  -> complete cache dir under asset-cache/browser-trace/…

# CLI
Test Client -> browsertrace.HandleCLI(["assets", …], env, stdout, stderr)
  assets --help  -> ensure + status
  assets ensure  -> EnsureAsset extension via BROWSER_AGENT_ASSET_BASE_URL
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` and
  `github.com/xhd2015/browser-agent/browsertrace` importable.
- Tree root `tests/browser-trace-asset-hydrate/`; ModuleRoot =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- browser-trace product embeds **extension only** (no session-page SPA).
- CLI leaves call `browsertrace.HandleCLI` (not binary shell-out).
- Isolate with `XDG_CACHE_HOME` + `BROWSER_AGENT_ASSET_BASE_URL` + httptest.
- Avoid leaf slugs ending in `-js`.

## Steps

1. Resolve ModuleRoot from DOCTEST_ROOT.
2. Leave Mode empty at root.
3. Shared helpers for all leaves.

## Context

- Spec version **0.0.2**.
- Parallel-safe: temp XDG + httptest + env maps per leaf.
- Classic TDD: CLI expected RED until `HandleCLI` assets exists.

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

func assertExitZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ExitCode != 0 {
		t.Fatalf("ExitCode=%d want 0", resp.ExitCode)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func assertPathUnder(t *testing.T, child, parent string) {
	t.Helper()
	if child == "" || parent == "" {
		t.Fatalf("assertPathUnder empty: child=%q parent=%q", child, parent)
	}
	absChild, err1 := filepath.Abs(child)
	absParent, err2 := filepath.Abs(parent)
	if err1 != nil || err2 != nil {
		t.Fatalf("abs: child=%v parent=%v", err1, err2)
	}
	rel, err := filepath.Rel(absParent, absChild)
	if err != nil || strings.HasPrefix(rel, "..") {
		t.Fatalf("path %q is not under %q (rel=%q err=%v)", absChild, absParent, rel, err)
	}
}

func assertTrailingNewline(t *testing.T, s, label string) {
	t.Helper()
	if s == "" || !strings.HasSuffix(s, "\n") {
		t.Fatalf("%s must end with trailing newline; got %q", label, truncate(s, 200))
	}
}
```
