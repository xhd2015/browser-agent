# Scenario

**Feature**: browser-agent CDP job types + full CLI side-commands (no real Chrome)

```
# CLI dispatch / help / missing session
Operator -> HandleCLI(args, empty env, stdout, stderr)
  --help -> lists serve,session + nested info,eval,run,logs,screenshot,cdp + \n
  session run|logs|screenshot|cdp without session -> error names both sources

# Live job-type observation
Test Client -> browseragent.Run(NoOpenChrome, NoAgentRun) -> Control Server
Fake Extension -> WS /v1/ws hello; record first job; auto result
HandleCLI session eval|run|logs|screenshot|cdp --session-id --addr
  -> POST /v1/jobs with correct type+params
  -> Fake Extension observes type/params
  -> CLI stdout result + trailing \n

# Pure / FS
Test Client -> FormatSystemPrompt(sessionID)
Test Client -> IsKnownJobType(s)
Test Client -> read shell/embedded background.js CDP tokens
Test Client -> read react/src/lib/protocol/jobs.ts
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Tree root is `tests/browser-agent-cdp-jobs/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Package `browseragent` will export / extend (TDD red until implemented):
  - `HandleCLI` commands: `run`, `logs`, `screenshot`, `cdp` (+ existing)
  - `IsKnownJobType(string) bool`
  - existing `FormatSystemPrompt`, `Config` + `Run` for serve
- Each server leaf uses isolated temp `BaseDir` and free loopback `Addr`.
- No real Chrome; no real agent-run; fake WS only.
- Sealed trees must not be modified:
  `tests/browser-agent/`, `tests/browser-agent-cli-react/`,
  `tests/browser-agent-serve-runtime/`.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate a unique temp `BaseDir` for leaves that serve or write scripts.
3. Default `SessionID`, `NoOpenChrome`, `NoAgentRun`, short `ReadyTimeout`.
4. Leave `Mode` and surface-specific fields for grouping/leaf Setup.
5. Default CLI env to empty map semantics in Run when nil (no ambient session).

## Context

- Parallel-safe: temp dirs + free ports per leaf.
- Shared helpers below available to all descendant Assert/Setup packages.
- Prefer package APIs over building `cmd/browser-agent` binary.
- Fake WS records **first** job only for type/params asserts.

```go
import (
	"fmt"
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
	req.NoOpenChrome = true
	req.NoAgentRun = true
	if req.SessionID == "" {
		req.SessionID = fmt.Sprintf("ba-cdp-%d", time.Now().UnixNano()%1e12)
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

func assertStdoutTrailingNewline(t *testing.T, stdout string) {
	t.Helper()
	if stdout == "" {
		t.Fatal("stdout is empty")
	}
	if !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("stdout must end with \\n (POSIX); last bytes=%q", tail(stdout, 40))
	}
}

func assertPrintedTrailingNewline(t *testing.T, resp *Response) {
	t.Helper()
	body := resp.Stdout
	if strings.TrimSpace(body) == "" {
		body = resp.Stderr
	}
	if strings.TrimSpace(body) == "" {
		t.Fatal("printed usage/help empty on stdout+stderr")
	}
	if !strings.HasSuffix(body, "\n") {
		t.Fatalf("printed text must end with \\n; last bytes=%q", tail(body, 40))
	}
}

func combinedCLIText(resp *Response) string {
	return resp.Stdout + resp.Stderr
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

func assertSessionSourcesInText(t *testing.T, text string) {
	t.Helper()
	if !strings.Contains(text, "--session-id") {
		t.Fatalf("error/text must mention --session-id; got %q", text)
	}
	if !strings.Contains(text, "BROWSER_AGENT_SESSION_ID") {
		t.Fatalf("error/text must mention BROWSER_AGENT_SESSION_ID; got %q", text)
	}
}

func assertObservedJobType(t *testing.T, resp *Response, want string) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.WSJobReceived && resp.ObservedJobType == "" {
		t.Fatalf("fake WS did not observe a job (want type %q); raw=%s jobsSeen=%d cliErr=%q stdout=%q",
			want, truncate(resp.ObservedJobRaw, 300), resp.JobsSeen, resp.CLIErr, truncate(resp.Stdout, 200))
	}
	got := strings.ToLower(strings.TrimSpace(resp.ObservedJobType))
	if got != strings.ToLower(want) {
		t.Fatalf("observed job type = %q, want %q; params=%v raw=%s",
			resp.ObservedJobType, want, resp.ObservedJobParams, truncate(resp.ObservedJobRaw, 400))
	}
}

func paramString(params map[string]any, keys ...string) string {
	if params == nil {
		return ""
	}
	for _, k := range keys {
		if v, ok := params[k]; ok && v != nil {
			switch t := v.(type) {
			case string:
				return t
			default:
				return fmt.Sprint(t)
			}
		}
	}
	// Nested params.params for cdp
	if nested, ok := params["params"].(map[string]any); ok {
		for _, k := range keys {
			if v, ok := nested[k]; ok && v != nil {
				if s, ok := v.(string); ok {
					return s
				}
				return fmt.Sprint(v)
			}
		}
	}
	return ""
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
```
