# Scenario

**Feature**: Phase 6 graceful shutdown and kill-existing

```
# HTTP shutdown
RunDaemon -> POST /v1/shutdown -> 202 -> drain -> server.json gone

# Kill helper
KillExistingDaemon(baseDir) -> read server.json -> POST shutdown -> wait -> force kill?

# CLI
HandleCLI serve --kill-existing -> KillExistingDaemon -> RunDaemon
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` importable.
- Phase 6 APIs (`POST /v1/shutdown`, `ShutdownDaemon`, `KillExistingDaemon`,
  `serve --kill-existing`, `DaemonConfig.ShutdownGracePeriod`) not implemented
  yet — tree is **RED**.
- Tree root is `tests/browser-agent-daemon-phase6/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- All leaves use isolated temp `BaseDir` and ephemeral `127.0.0.1:0` listen.
- No real Chrome; no agent-run.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate temp `BaseDir` for every leaf.
3. Default `ReadyTimeout = 5s`, `ShutdownWait = 8s`, `KillTimeout = 10s`.
4. Leave `Mode` and surface-specific fields for grouping/leaf Setup.

## Context

- Spec version **0.0.2**.
- `server.json` path: `{BaseDir}/server.json`.
- Shutdown leaves use **HTTP POST** (not ctx cancel).
- Force-kill leaf sets `ShutdownGracePeriod` and short `KillTimeout`.

```go
import (
	"net/http"
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
	if req.ShutdownWait == 0 {
		req.ShutdownWait = 8 * time.Second
	}
	if req.KillTimeout == 0 {
		req.KillTimeout = 10 * time.Second
	}
	return nil
}

func assertNoRunErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run transport/setup error: %v", err)
	}
}

func assertHTTPStatus(t *testing.T, resp *Response, want int) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.StatusCode != want {
		t.Fatalf("HTTP status = %d, want %d; body=%s",
			resp.StatusCode, want, truncate(resp.BodyString, 400))
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

var _ = http.StatusAccepted
```