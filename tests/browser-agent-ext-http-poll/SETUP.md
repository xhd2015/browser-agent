# Scenario

**Feature**: Phase 1 daemon HTTP poll transport — hello / poll / result for extension attach without WebSocket

```
# pure wait_ms
NormalizeExtPollWaitMS(n) -> default 25000 / cap 30000

# HTTP attach + job delivery
registry.Create(session)
POST /v1/ext/hello {session_id} -> phase extension_connected
POST /v1/jobs (bg) stays Queued without WS
POST /v1/ext/poll {session_id, wait_ms} -> {jobs, events}
POST /v1/ext/result {session_id, job_id, ok} -> completes waiter
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` is importable.
- Tree root is `tests/browser-agent-ext-http-poll/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- **Classic TDD**: routes `POST /v1/ext/hello|poll|result` are **not** GREEN yet;
  `doctest test` should be RED (404 / missing body fields). Pure
  `NormalizeExtPollWaitMS` is a **nested** tree at `wait-ms/`.
- Registry HTTP leaves use `t.TempDir()` BaseDir + `httptest` via
  `NewRegistryControlHandler` + `registry.Create`.
- No real browser / no real extension process.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Default `SessionID` when empty so leaves can override.
3. Default hello identity (`version`, `features`).
4. Leave `Mode` empty at root (grouping/leaf Setup sets Mode).

## Context

- Spec version **0.0.2**.
- Default poll wait: **25000** ms; max **30000** ms.
- Phase 2 (Firefox extension poll loop) is **out of scope**.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ModuleRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", ".."))
	if req.SessionID == "" {
		req.SessionID = "sess-ext-poll-default"
	}
	if req.HelloVersion == "" {
		req.HelloVersion = "1.0.0"
	}
	if req.HelloFeatures == nil {
		req.HelloFeatures = []string{"browser-agent"}
	}
	if req.JobHTTPType == "" {
		req.JobHTTPType = "eval"
	}
	if req.JobHTTPParams == nil {
		req.JobHTTPParams = map[string]any{"expression": "1+1"}
	}
	if req.JobHTTPTimeoutMS <= 0 {
		req.JobHTTPTimeoutMS = 8000
	}
	return nil
}

func assertNoRunErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run transport error: %v", err)
	}
}

func assertExitZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ExitCode != 0 {
		t.Fatalf("ExitCode=%d want 0", resp.ExitCode)
	}
}

func assertJSONContentType(t *testing.T, ct string) {
	t.Helper()
	if ct == "" {
		return
	}
	if !strings.Contains(strings.ToLower(ct), "json") {
		t.Fatalf("Content-Type=%q want application/json", ct)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
```
