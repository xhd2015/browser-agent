# Scenario

**Feature**: Phase 8 `session new` + `EnsureDaemon`

```
EnsureDaemon(baseDir, addr) -> reuse | spawn + wait healthy + server.json
SessionNew(cfg)            -> POST /v1/sessions + OpenChromeFn + pretty stdout
HandleCLI session new      -> CLI dispatch (no agent-run)
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` importable.
- Phase 8 APIs (`EnsureDaemon`, `SessionNew`, `session new` CLI) not implemented
  yet — tree is **RED**.
- Tree root is `tests/browser-agent-daemon-phase8/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Running leaves use isolated temp `BaseDir` and ephemeral `127.0.0.1:0` listen.
- No real Chrome; no agent-run; hooks record calls only.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate temp `BaseDir` for every leaf.
3. Default `ReadyTimeout = 5s`, `SessionID = "sess-new-8"` for explicit-id leaves.
4. Leave `Mode` and surface-specific fields for grouping/leaf Setup.

## Context

- Spec version **0.0.2**.
- `server.json` path: `{BaseDir}/server.json`.
- Auto-generated ids: `^sess-[a-z0-9]{6}$` (Phase 1 `GenerateSessionID`).
- Duplicate create surfaces as error text containing duplicate/exists/409.

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
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	if req.SessionID == "" {
		req.SessionID = "sess-new-8"
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
}```
