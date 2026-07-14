# Scenario

**Feature**: Phase 5 daemon host vs session serve split

```
# RunDaemon (most leaves)
Test Client -> RunDaemon(ctx, DaemonConfig) -> empty registry + server.json
Test Client -> GET /v1/health | /v1/sessions
ctx cancel -> RemoveDaemonMeta(server.json)

# Run compat (one leaf)
Test Client -> Run(ctx, Config{SessionID}) -> GET /v1/session?session=

# CLI serve (one leaf)
Test Client -> HandleCLI(serve --session-id) -> stderr deprecation warning
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` importable.
- Phase 5 APIs (`RunDaemon`, `DaemonConfig`) not implemented yet — tree is **RED**.
- Tree root is `tests/browser-agent-daemon-phase5/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- All leaves use isolated temp `BaseDir` and ephemeral `127.0.0.1:0` listen.
- No real Chrome; no agent-run.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate temp `BaseDir` for every leaf.
3. Default `ReadyTimeout = 5s`, `ShutdownWait = 3s`.
4. Leave `Mode` and surface-specific fields for grouping/leaf Setup.

## Context

- Spec version **0.0.2**.
- `server.json` path: `{BaseDir}/server.json`.
- After implementer, `doctest test ./tests/browser-agent/...` must stay GREEN.

```go
import (
	"net/http"
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
	req.NoOpenChrome = true
	req.NoAgentRun = true
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	if req.ShutdownWait == 0 {
		req.ShutdownWait = 3 * time.Second
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

func daemonMetaFieldsMatch(t *testing.T, meta browseragent.DaemonMeta, addr, baseDir string) {
	t.Helper()
	if meta.PID != os.Getpid() {
		t.Fatalf("DaemonMeta.PID=%d want os.Getpid()=%d", meta.PID, os.Getpid())
	}
	if meta.Addr != addr {
		t.Fatalf("DaemonMeta.Addr=%q want %q", meta.Addr, addr)
	}
	wantBaseURL := "http://" + addr
	if meta.BaseURL != wantBaseURL {
		t.Fatalf("DaemonMeta.BaseURL=%q want %q", meta.BaseURL, wantBaseURL)
	}
	if meta.BaseDir != baseDir {
		t.Fatalf("DaemonMeta.BaseDir=%q want %q", meta.BaseDir, baseDir)
	}
	if meta.StartedAt.IsZero() {
		t.Fatal("DaemonMeta.StartedAt is zero")
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

var _ = http.StatusOK
```