# Scenario

**Feature**: Phase 7 read-only daemon status (`serve --status`)

```
QueryDaemonStatus(baseDir) -> DaemonStatus (no writes)
FormatDaemonStatus(w, st)  -> pretty table
HandleCLI serve --status   -> query + format + exit 0
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` importable.
- Phase 7 APIs (`DaemonStatus`, `QueryDaemonStatus`, `FormatDaemonStatus`,
  `serve --status`) not implemented yet — tree is **RED**.
- Tree root is `tests/browser-agent-daemon-phase7/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Running leaves use isolated temp `BaseDir` and ephemeral `127.0.0.1:0` listen.
- No real Chrome; no agent-run.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate temp `BaseDir` for every leaf.
3. Default `ReadyTimeout = 5s`, `SessionID = "sess-status-7"`.
4. Leave `Mode` and surface-specific fields for grouping/leaf Setup.

## Context

- Spec version **0.0.2**.
- `server.json` path: `{BaseDir}/server.json`.
- All status probes must be **read-only** — meta bytes unchanged after query/CLI.
- Stale-pid leaf writes fixture meta once in Run (not via QueryDaemonStatus).

```go
import (
	"bytes"
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
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	if req.SessionID == "" {
		req.SessionID = "sess-status-7"
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

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
```