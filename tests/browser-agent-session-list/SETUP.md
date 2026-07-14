# Scenario

**Feature**: `browser-agent session list` lists live sessions from daemon registry

```
RunDaemon(:0, BaseDir) -> server.json
POST /v1/sessions -> session snapshots

HandleCLI session list --base-dir BaseDir [--json]
  -> human table or JSON array

no daemon -> warning stderr; exit 0; empty output
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` importable.
- `session list` **not implemented** — tree is **RED**.
- Tree root is `tests/browser-agent-session-list/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Daemon leaves use isolated temp `BaseDir` and ephemeral `127.0.0.1:0` listen.
- No real Chrome; no agent-run.
- Reuse phase4 fake extension WS hello for `phase-and-connected`.
- Reuse `startDaemonServer` / `RunDaemon` harness from session-delete / session-addr-resolve.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate temp `BaseDir` for every leaf.
3. Default `ReadyTimeout = 5s`, `MaxDispatchWait = 12s`.
4. Default `HelloVersion = 1.0.0`, `HelloFeatures = ["browser-agent"]`.
5. Default `CLIEnv` to empty map when nil.
6. Default `OmitAddr = true`, `PassBaseDir = true`, `StartDaemon = true`.
7. Grouping/leaf Setup sets `Mode`, `ListOp`, and session seeds.

## Context

- Spec version **0.0.2**.
- `server.json` path: `{BaseDir}/server.json`.
- Sessions created via `POST /v1/sessions` unless empty leaf.
- Human columns: **Session ID, Phase, Connected**; trailing `N sessions`.
- Daemon down: stderr `warning: daemon not running in <base-dir>`; exit 0.

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
	req.OmitAddr = true
	req.PassBaseDir = true
	req.StartDaemon = true
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
	return resp.Stdout + resp.Stderr + resp.CLIErr
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func sortedIDs(ids []string) []string {
	out := append([]string(nil), ids...)
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

func stdoutSessionIDPositions(stdout string, ids []string) map[string]int {
	pos := make(map[string]int, len(ids))
	low := strings.ToLower(stdout)
	for _, id := range ids {
		pos[id] = strings.Index(low, strings.ToLower(id))
	}
	return pos
}```
