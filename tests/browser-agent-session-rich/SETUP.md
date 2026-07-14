# Scenario

**Feature**: enriched session snapshots, extension telemetry, human-default session info

```
RunDaemon(:0, BaseDir) -> server.json
POST /v1/sessions -> session id

Fake Extension -> hello|status { session_page_count, browser_product, session_pages }
GET /v1/session?session=ID -> enriched snapshot fields

HandleCLI session list|info --base-dir BaseDir [--json]
  -> wider list columns + hints; human info default
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` importable.
- Session-rich fields and CLI UX **not implemented** — tree is **RED**.
- Tree root is `tests/browser-agent-session-rich/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Daemon leaves use isolated temp `BaseDir` and ephemeral `127.0.0.1:0` listen.
- No real Chrome; no agent-run.
- Reuse phase4 fake extension WS harness; reuse session-list/delete `RunDaemon` pattern.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate temp `BaseDir` for every leaf.
3. Default `ReadyTimeout = 5s`, `MaxDispatchWait = 12s`.
4. Default `HelloVersion = 1.0.0`, `HelloFeatures = ["browser-agent"]`.
5. Default `CLIEnv` to empty map when nil.
6. Default `OmitAddr = false`, `PassBaseDir = true`.
7. Grouping/leaf Setup sets `Mode` and op-specific fields.

## Context

- Spec version **0.0.2**.
- Unknown page count → display `—` (not 0).
- Derived status: count 0 → `no_session_page`; count >1 → `multiple_pages`;
  count nil → `unknown`.
- `session info` human default; `--json` optional for machine output.

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

func assertHTTPStatus(t *testing.T, resp *Response, want int) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.SessionStatusCode != want {
		t.Fatalf("GET /v1/session status=%d want %d; body=%s",
			resp.SessionStatusCode, want, truncate(resp.SessionBodyString, 500))
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

func stdoutLooksLikeJSONObject(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")
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

func intPtr(n int) *int {
	return &n
}

```