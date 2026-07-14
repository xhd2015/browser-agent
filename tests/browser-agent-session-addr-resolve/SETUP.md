# Scenario

**Bug**: session side-commands omit `--addr` and hit default `:43761` while daemon
listens on ephemeral port in `server.json`

```
RunDaemon(:0, BaseDir) -> server.json (addr != 43761)
POST /v1/sessions -> sess-xxxxxx

HandleCLI session info|eval --session-id X --base-dir BaseDir   # no --addr
  -> should read server.json (after fix)
  -> RED: hits http://127.0.0.1:43761 -> 404 session not found
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` importable.
- Fix not landed — `info/from-server-json` and `eval/from-server-json` are **RED**.
- Tree root is `tests/browser-agent-session-addr-resolve/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Daemon leaves use isolated temp `BaseDir` and ephemeral `127.0.0.1:0` listen.
- No real Chrome; no agent-run.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate temp `BaseDir` for every leaf.
3. Default `ReadyTimeout = 5s`, `MaxDispatchWait = 12s`.
4. Default `CLIEnv` to empty map when nil.
5. Grouping/leaf Setup sets `Sidecmd`, `AddrSource`, and argv flags.

## Context

- Spec version **0.0.2**.
- `server.json` path: `{BaseDir}/server.json`.
- Reuse `startDaemonServer` / `RunDaemon` harness from phase6 DOCTEST.md.
- Sessions created via `POST /v1/sessions` (not `browseragent.Run` single-session).

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
	req.StartDaemon = true
	req.PassBaseDir = true
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

func assertStdoutTrailingNewline(t *testing.T, stdout string) {
	t.Helper()
	if stdout == "" {
		return
	}
	if !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("stdout must end with trailing \\n; got %q", truncate(stdout, 200))
	}
}

func combinedOutput(resp *Response) string {
	if resp == nil {
		return ""
	}
	return resp.Stdout + resp.Stderr + resp.CLIErr + resp.ErrText
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

```