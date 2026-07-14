# Scenario

**Feature**: browser-agent control plane without Chrome or agent-run

```
# Pure surfaces (no listen socket)
Test Client -> ResolveSessionID(flag, env) -> session id | error
Test Client -> JobQueue Enqueue/Dequeue/Wait/Complete
Test Client -> FormatSystemPrompt(sessionID) / DefaultAddr

# Integration: Control Server + fake Extension over WS
Test Client -> browseragent.Run(NoOpenChrome, NoAgentRun) -> Control Server @ Addr
Fake Extension -> WS /v1/ws hello|result
Test Client -> POST /v1/jobs | GET /v1/session | GET /go
Control Server -> JobResult | session JSON | SPA HTML
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Package `browseragent` will export (TDD red until implemented):
  - `ResolveSessionID`, `NewJobQueue` + job types, `FormatSystemPrompt`
  - `DefaultAddr` / `DefaultControlPort` (**43761**)
  - `Config` + `Run(ctx, Config)` control server with `/v1/health`, `/v1/session`,
    `/v1/jobs`, `/v1/ws`, `/` or `/go`
- Each server leaf uses an isolated temp `BaseDir` and free loopback `Addr`.
- `NoOpenChrome` and no real `agent-run` (harness sets `NoAgentRun`).
- **Disconnect policy v1**: fail inflight jobs on WS disconnect (no requeue).
- No real Chrome process.

## Steps

1. Allocate a unique temp `BaseDir` for leaves that may start a server.
2. Set `SessionID` to a deterministic id when empty.
3. Set `NoOpenChrome = true`.
4. Default `ReadyTimeout` to 5s for health wait.
5. Leave `Mode` and surface-specific fields for grouping/leaf Setup.

## Context

- Wire protocol bind: loopback only (`127.0.0.1`).
- Parallel-safe: free ports + temp dirs per leaf.
- Shared helpers below are available to all descendant Assert/Setup packages.
- Existing browser-trace doctests are external regression only (not this tree).

```go
import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	dir := t.TempDir()
	req.BaseDir = filepath.Join(dir, "browser-agent-base")
	if err := os.MkdirAll(req.BaseDir, 0o755); err != nil {
		return err
	}
	req.NoOpenChrome = true
	if req.SessionID == "" {
		req.SessionID = fmt.Sprintf("ba-%d", time.Now().UnixNano()%1e12)
	}
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 5 * time.Second
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
		t.Fatalf("HTTP status = %d, want %d; content-type=%q body=%s",
			resp.StatusCode, want, resp.ContentType, truncate(resp.BodyString, 400))
	}
}

func assertJSONContentType(t *testing.T, resp *Response) {
	t.Helper()
	ct := strings.ToLower(resp.ContentType)
	if !strings.Contains(ct, "json") && resp.StatusCode == http.StatusOK {
		// Soft: some early servers omit content-type; body must still parse via Run.
		if resp.Raw == nil && len(resp.Body) > 0 && resp.Body[0] != '{' {
			t.Fatalf("Content-Type %q does not look like JSON; body=%s",
				resp.ContentType, truncate(resp.BodyString, 200))
		}
	}
}

func assertHTMLContentType(t *testing.T, resp *Response) {
	t.Helper()
	ct := strings.ToLower(resp.ContentType)
	if !strings.Contains(ct, "html") {
		// Accept missing content-type if body looks like HTML.
		if !strings.Contains(strings.ToLower(resp.BodyString), "<html") &&
			!strings.Contains(strings.ToLower(resp.BodyString), "<!doctype") {
			t.Fatalf("Content-Type %q / body not HTML; body=%s",
				resp.ContentType, truncate(resp.BodyString, 200))
		}
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

func errText(resp *Response, err error) string {
	if err != nil {
		return err.Error()
	}
	if resp == nil {
		return ""
	}
	parts := []string{resp.ResolveErr, resp.JobResultError, resp.HTTPJobError, resp.CompleteErr, resp.RunErrText}
	return strings.Join(parts, " ")
}
```
