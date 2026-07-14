# Scenario

**Feature**: serve --status rich output (version + extension + connected)

```
QueryDaemonStatus(baseDir) -> DaemonStatus (rich fields, read-only)
FormatDaemonStatus(w, st)  -> version + extension + Connected table
HandleCLI serve --status   -> query + format + exit 0
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` importable.
- Rich status fields on `DaemonStatus` / formatter **not implemented** — tree is **RED**.
- Tree root is `tests/browser-agent-serve-status-rich/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Running leaves use isolated temp `BaseDir`, ephemeral `127.0.0.1:0` listen,
  and isolated `HOME` (`TestHome`) for canonical extension paths.
- No real Chrome; no agent-run.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate temp `BaseDir` and `TestHome` for every leaf.
3. Set `t.Setenv("HOME", TestHome)` before any extension extract.
4. Default `ReadyTimeout = 5s`, `SessionID = "sess-rich-status"`.
5. Default `DaemonVersion = browseragent.ClientVersion()`.
6. Leave `Mode` and surface-specific fields for grouping/leaf Setup.

## Context

- Spec version **0.0.2**.
- `server.json` path: `{BaseDir}/server.json`.
- All status probes must be **read-only** — meta bytes unchanged after query/CLI.
- Canonical extension under `{TestHome}/.browser-agent/managed-chrome/extensions/browser-agent/{ver}/`.

```go
import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/browser-agent/browseragent"
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
	t.Setenv("HOME", req.TestHome)
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	if req.SessionID == "" {
		req.SessionID = "sess-rich-status"
	}
	if strings.TrimSpace(req.DaemonVersion) == "" {
		req.DaemonVersion = browseragent.ClientVersion()
	}
	return nil
}

func assertNoRunErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run transport/setup error: %v", err)
	}
}

func assertMetaUnchanged(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.DaemonMetaBeforeHit != resp.DaemonMetaAfterHit {
		t.Fatalf("server.json existence changed: before=%v after=%v path=%s",
			resp.DaemonMetaBeforeHit, resp.DaemonMetaAfterHit, resp.DaemonMetaPath)
	}
	if !bytes.Equal(resp.DaemonMetaBefore, resp.DaemonMetaAfter) {
		t.Fatalf("server.json mutated:\nbefore=%q\nafter=%q",
			truncate(string(resp.DaemonMetaBefore), 400),
			truncate(string(resp.DaemonMetaAfter), 400))
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

func assertStringSlicesEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("session ids len=%d want %d; got=%v want=%v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("session ids[%d]=%q want %q; got=%v want=%v", i, got[i], want[i], got, want)
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

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
```