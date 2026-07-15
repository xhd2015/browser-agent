# Scenario

**Feature**: P1–P5 asset hydrate + P7 release names and operator docs

```
# P1–P4 package APIs
# P5 CLI HandleCLI assets ensure|status|help
# P7
Test Client -> AssetReleaseNames(version) -> archive basenames
Test Client -> read docs/README/SKILL -> hydrate operator guide
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` importable.
- Tree root `tests/browser-agent-asset-hydrate/`; ModuleRoot =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- **P1–P5** may be GREEN; **P7** Classic TDD (RED until release names + docs).
- CLI leaves call `HandleCLI` (not binary shell-out).
- Isolate with `XDG_CACHE_HOME` + `BROWSER_AGENT_ASSET_BASE_URL` + httptest.
- Avoid leaf slugs ending in `-js`.

## Steps

1. Resolve ModuleRoot from DOCTEST_ROOT.
2. Leave Mode empty at root.
3. Shared helpers for all leaves.

## Context

- Spec version **0.0.2**.
- Parallel-safe: temp XDG + httptest + env maps per leaf.

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
	_ = os.RemoveAll(filepath.Join(DOCTEST_ROOT, "completeness", "session-page", "html-without-js"))
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

func assertCompleteTrue(t *testing.T, resp *Response, kind string) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.Complete {
		t.Fatalf("EmbedCompleteFS(kind=%q) = false, want true; FSRoot=%s",
			kind, resp.FSRoot)
	}
}

func assertCompleteFalse(t *testing.T, resp *Response, kind string) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.Complete {
		t.Fatalf("EmbedCompleteFS(kind=%q) = true, want false; FSRoot=%s",
			kind, resp.FSRoot)
	}
}

func assertHTMLHasSessionRoot(t *testing.T, html string) {
	t.Helper()
	if strings.TrimSpace(html) == "" {
		t.Fatal("resolve HTML is empty")
	}
	low := strings.ToLower(html)
	if !strings.Contains(low, "data-browser-agent-root") &&
		!strings.Contains(low, `id="root"`) &&
		!strings.Contains(low, `id='root'`) {
		t.Fatalf("resolve HTML missing session root marker; body=%s",
			truncate(html, 400))
	}
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
