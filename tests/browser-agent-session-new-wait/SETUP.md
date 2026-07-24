# Scenario

**Feature**: `session new` waits for extension connection before returning

```
RunDaemon(:0, BaseDir) -> server.json
POST /v1/sessions -> session id

# with-wait mode — extension outcomes
Fake Extension -> hello { version, features }
Poll GET /v1/session?session=ID -> extension.connected, extension.features
  -> supported: "Extension connected" stderr, exit 0
  -> unsupported: error, exit 1
  -> timeout: stderr warning, exit 0

# skip-wait mode
NoOpenChrome=true or NoWait=true
  -> no wait, exit 0, elapsed < 1s
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` importable.
- Session new wait logic **not implemented** — tree is **RED**.
- Tree root is `tests/browser-agent-session-new-wait/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Daemon leaves use isolated temp `BaseDir` and ephemeral `127.0.0.1:0` listen.
- No real Chrome; no agent-run.
- Reuse phase4 fake extension WS harness; reuse session-rich `RunDaemon` pattern.
- `SessionNewConfig` will have new fields `WaitExtensionTimeout` and `NoWait`.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate temp `BaseDir` for every leaf.
3. Default `ReadyTimeout = 5s`, `MaxDispatchWait = 12s`.
4. Default `WaitExtensionTimeout = 30s`.
5. Default `HelloVersion = 1.0.0`, `HelloFeatures = ["browser-agent"]`.
6. Default `NoWait = false`, `NoOpenChrome = false`.
7. Grouping/leaf Setup sets `Mode` and op-specific fields.

## Context

- Spec version **0.0.2**.
- Polling interval: 500ms for extension connection.
- Timeout: 30s configurable via `SessionNewConfig.WaitExtensionTimeout`.
- On success: "Extension connected ✓" to stderr.
- On timeout: "warning: extension did not connect within 30s" to stderr.
- On unsupported: error "Error: extension does not support browser-agent".
- `--no-open-chrome` and `--no-wait` skip the wait entirely.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ModuleRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", ".."))
	dir := t.TempDir()
	req.BaseDir = filepath.Join(dir, "browser-agent-base")
	if err := os.MkdirAll(req.BaseDir, 0o755); err != nil {
		return err
	}
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	if req.MaxDispatchWait == 0 {
		req.MaxDispatchWait = 12 * time.Second
	}
	if req.WaitExtensionTimeout == 0 {
		req.WaitExtensionTimeout = 30 * time.Second
	}
	if req.HelloVersion == "" {
		req.HelloVersion = "1.0.0"
	}
	if req.HelloFeatures == nil {
		req.HelloFeatures = []string{"browser-agent"}
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
		t.Fatalf("ExitCode=%d want 0; ErrStr=%q stderr=%q stdout=%q",
			resp.ExitCode, resp.ErrStr, resp.Stderr, resp.Stdout)
	}
	if resp.ErrStr != "" {
		t.Fatalf("ErrStr=%q want empty; stderr=%q stdout=%q",
			resp.ErrStr, resp.Stderr, resp.Stdout)
	}
}

func assertExitNonZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ExitCode == 0 {
		t.Fatalf("ExitCode=%d want non-zero; ErrStr=%q stderr=%q stdout=%q",
			resp.ExitCode, resp.ErrStr, resp.Stderr, resp.Stdout)
	}
}

func assertContains(t *testing.T, haystack string, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Fatalf("expected text to contain %q; got:\n%s", needle, truncate(haystack, 800))
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
			t.Fatalf("expected text to NOT contain %q; got:\n%s", n, truncate(haystack, 800))
		}
	}
}

func combinedOutput(resp *Response) string {
	if resp == nil {
		return ""
	}
	return resp.Stdout + resp.Stderr + resp.ErrStr
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
```
