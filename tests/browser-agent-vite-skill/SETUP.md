# Scenario

**Feature**: Vite session-page embed + skill packaging for browser-agent (no npm/Chrome)

```
# Embed FS
Test Client -> browseragent.SessionPageFS() -> index.html root mount

# HTTP SPA + assets
Test Client -> browseragent.Run(NoOpenChrome, NoAgentRun) -> Control Server
Test Client -> GET /go | / | /assets/session-page.js
  -> HTML with boot/product/install markers | asset 200

# Skill CLI (package API)
Operator -> HandleCLI(["skill", "--list"|"--show"|…], env, stdout, stderr)
  -> skill name | SKILL.md body | help/error

# Boot pure helper
Test Client -> FormatSessionBootJSON(sessionID)
  -> {"session_id","product":"browser-agent","control_port":43761}
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Tree root is `tests/browser-agent-vite-skill/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Package `browseragent` will export (TDD red until implemented):
  - `SessionPageFS() fs.FS`
  - `FormatSessionBootJSON(sessionID string) string`
  - `HandleCLI` gains `skill` subcommand (writers, not raw os.Stdout-only)
  - existing `Config` + `Run` for serve (NoOpenChrome / NoAgentRun)
- Committed fixture under `browseragent/embedded/session-page/` (no npm).
- Each HTTP leaf uses isolated temp `BaseDir` and free loopback `Addr`.
- Sealed prior browser-agent trees must not be modified.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate a unique temp `BaseDir` for leaves that serve.
3. Default `SessionID`, `NoOpenChrome`, `NoAgentRun`, short `ReadyTimeout`.
4. Leave `Mode` and surface-specific fields for grouping/leaf Setup.

## Context

- Parallel-safe: temp dirs + free ports per leaf.
- Prefer package APIs over building `cmd/browser-agent` binary.
- Skill output must land on HandleCLI stdout/stderr buffers.

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
	req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	dir := t.TempDir()
	req.BaseDir = filepath.Join(dir, "browser-agent-vite-base")
	if err := os.MkdirAll(req.BaseDir, 0o755); err != nil {
		return err
	}
	req.NoOpenChrome = true
	req.NoAgentRun = true
	if req.SessionID == "" {
		req.SessionID = fmt.Sprintf("ba-vite-%d", time.Now().UnixNano()%1e12)
	}
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	if req.MaxDispatchWait == 0 {
		req.MaxDispatchWait = 3 * time.Second
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

func assertHTMLContentType(t *testing.T, resp *Response) {
	t.Helper()
	ct := strings.ToLower(resp.ContentType)
	if !strings.Contains(ct, "html") {
		if !strings.Contains(strings.ToLower(resp.BodyString), "<html") &&
			!strings.Contains(strings.ToLower(resp.BodyString), "<!doctype") {
			t.Fatalf("Content-Type %q / body not HTML; body=%s",
				resp.ContentType, truncate(resp.BodyString, 200))
		}
	}
}

func assertStdoutTrailingNewline(t *testing.T, stdout string) {
	t.Helper()
	if stdout == "" {
		t.Fatal("stdout is empty")
	}
	if !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("stdout must end with \\n (POSIX); last bytes=%q", tail(stdout, 40))
	}
}

func assertExitZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q err=%q stdout=%q cliErr=%q",
			resp.ExitCode, resp.Stderr, resp.ErrText, resp.Stdout, resp.CLIErr)
	}
}

func assertCLINilErr(t *testing.T, resp *Response) {
	t.Helper()
	if resp.CLIErr != "" {
		t.Fatalf("HandleCLI error = %q; stdout=%q stderr=%q",
			resp.CLIErr, resp.Stdout, resp.Stderr)
	}
	if resp.DispatchTimedOut {
		t.Fatal("HandleCLI timed out")
	}
}

func combinedCLIText(resp *Response) string {
	return resp.Stdout + resp.Stderr
}

func hasRootMount(body string) bool {
	low := strings.ToLower(body)
	return strings.Contains(low, `id="root"`) ||
		strings.Contains(low, `id='root'`) ||
		strings.Contains(low, "data-browser-agent-root") ||
		strings.Contains(low, "browser-agent-root")
}

func hasInstallMarkers(body string) bool {
	low := strings.ToLower(body)
	return strings.Contains(low, "chrome://extensions") ||
		strings.Contains(low, "load unpacked") ||
		strings.Contains(low, "data-browser-agent-install") ||
		strings.Contains(low, "browser-agent-install") ||
		strings.Contains(low, "installguideline") ||
		strings.Contains(low, "install-guideline")
}

func hasBootOrProductMarkers(body string) bool {
	low := strings.ToLower(body)
	hasProduct := strings.Contains(low, "browser-agent")
	hasPort := strings.Contains(body, "43761")
	hasBoot := strings.Contains(low, "browser-agent-boot") ||
		strings.Contains(low, "data-session-id") ||
		strings.Contains(low, "data-control-port") ||
		strings.Contains(low, "__browser_agent") ||
		strings.Contains(low, "control_port") ||
		strings.Contains(low, "controlport")
	return hasProduct && hasPort && hasBoot
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

var (
	_ = http.StatusOK
)
```
