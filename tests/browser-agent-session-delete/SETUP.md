# Scenario

**Feature**: session delete removes waiting sessions; rejects connected; disk-only cleanup

```
RunDaemon(:0, BaseDir) -> server.json
POST /v1/sessions -> sess-xxxxxx (waiting_extension)

HandleCLI session delete --session-id ID --base-dir BaseDir   # no --addr
  -> exit 0; deleted message; dir gone; absent from GET /v1/sessions

Fake Extension -> GET /v1/ws?session=ID hello
HandleCLI session delete ... -> exit 1; extension connected; session remains

DELETE /v1/session?session=ID -> 200|204 (waiting) or 409 (connected)

Disk-only mkdir sessions/id (no registry) -> delete -> dir gone
session --help -> lists session delete
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` importable.
- `session delete` and `DELETE /v1/session` **not implemented** — tree is **RED**.
- Tree root is `tests/browser-agent-session-delete/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Daemon leaves use isolated temp `BaseDir` and ephemeral `127.0.0.1:0` listen.
- No real Chrome; no agent-run.
- Reuse phase4 fake extension WS hello for connected leaves.
- Reuse phase6 `startDaemonServer` / `RunDaemon` harness.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate temp `BaseDir` for every leaf.
3. Default `ReadyTimeout = 5s`, `MaxDispatchWait = 12s`.
4. Default `HelloVersion = 1.0.0`, `HelloFeatures = ["browser-agent"]`.
5. Default `CLIEnv` to empty map when nil.
6. Grouping/leaf Setup sets `Mode`, op fields, and session ids.

## Context

- Spec version **0.0.2**.
- `server.json` path: `{BaseDir}/server.json`.
- Sessions created via `POST /v1/sessions` unless disk-only leaf.
- After successful delete: `GET /v1/sessions` must not include id; `SessionDirExists` false.

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
	if req.MaxDispatchWait == 0 {
		req.MaxDispatchWait = 12 * time.Second
	}
	if req.CLIEnv == nil {
		req.CLIEnv = map[string]string{}
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
		t.Fatalf("ExitCode=%d want 0; CLIErr=%q stderr=%q stdout=%q",
			resp.ExitCode, resp.CLIErr, resp.Stderr, resp.Stdout)
	}
	if resp.CLIErr != "" {
		t.Fatalf("CLIErr=%q want empty; stderr=%q stdout=%q",
			resp.CLIErr, resp.Stderr, resp.Stdout)
	}
}

func assertExitOne(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ExitCode != 1 {
		t.Fatalf("ExitCode=%d want 1; CLIErr=%q stderr=%q stdout=%q",
			resp.ExitCode, resp.CLIErr, resp.Stderr, resp.Stdout)
	}
	if resp.CLIErr == "" {
		t.Fatalf("CLIErr empty want non-nil on failure; stderr=%q stdout=%q",
			resp.Stderr, resp.Stdout)
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

func combinedOutput(resp *Response) string {
	if resp == nil {
		return ""
	}
	return resp.Stdout + resp.Stderr + resp.CLIErr
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
```