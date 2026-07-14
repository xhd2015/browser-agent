# Scenario

**Feature**: browser-trace lifecycle logging with mock extension (no Chrome)

```
# CLI starts Control Server; Lifecycle Logger writes progress to stderr / log file
User -> browser-trace (NoOpenChrome) -> Control Server @ Addr
browser-trace -> Lifecycle Logger -> stderr (info milestones)
browser-trace -> Lifecycle Logger -> {sessionDir}/browser-trace.log  # unless Quiet|NoLogFile

# Mock Extension speaks wire protocol; success path completes quickly
Mock Extension -> POST /v1/hello
Mock Extension -> POST /v1/status (recording)
Mock Extension -> POST /v1/complete

# Success stdout remains machine-readable (path only)
browser-trace stdout -> "{sessionDir}\n"
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Package `browsertrace` implements logging Config fields (`Verbose`, `Quiet`,
  `NoLogFile`, `ReadyHeartbeat`) and emits milestones as described in root
  `DOCTEST.md` (implementer TDD green).
- Each leaf uses an isolated temp `BaseDir` and ephemeral free `127.0.0.1:port`.
- `NoOpenChrome` is always true.
- Product default ready heartbeat is 5s; heartbeat leaf injects a short interval.
- No real Chrome process and no real extension load.

## Steps

1. Allocate a unique temp `BaseDir` for the leaf.
2. Leave `Addr` empty so `Run` picks a free loopback port.
3. Set `NoOpenChrome = true`.
4. Default timeouts to product values; descendants shorten for timeout paths.
5. Default log flags off (`Verbose=false`, `Quiet=false`, `NoLogFile=false`);
   descendants override.
6. Default `ExtensionScript` / `StopMode` to none; success leaves set record+complete.

## Context

- Parallel-safe: free ports + temp dirs.
- Shared helpers assert stdout contract, milestone tokens (case-insensitive),
  and log-file presence.
- Token matching is intentional: log line prefix/format is not rigidly pinned.

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
	dir := t.TempDir()
	req.BaseDir = filepath.Join(dir, "browser-trace-base")
	if err := os.MkdirAll(req.BaseDir, 0o755); err != nil {
		return err
	}
	req.NoOpenChrome = true
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 30 * time.Second
	}
	if req.CompleteTimeout == 0 {
		req.CompleteTimeout = 30 * time.Second
	}
	if req.ExtensionScript == "" {
		req.ExtensionScript = ExtNone
	}
	if req.StopMode == "" {
		req.StopMode = StopNone
	}
	// Logging flags default false; ReadyHeartbeat 0 → product default in Run.
	return nil
}

func assertExitZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q err=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.ErrText, resp.Stdout)
	}
}

func assertExitNonZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ExitCode == 0 {
		t.Fatalf("exit code = 0, want ≠ 0; stderr=%q err=%q stdout=%q",
			resp.Stderr, resp.ErrText, resp.Stdout)
	}
}

// assertStdoutSessionPathOnly enforces the stable success stdout contract:
// exactly sessionDir + "\n" (no banners, no extra blank lines).
func assertStdoutSessionPathOnly(t *testing.T, resp *Response) {
	t.Helper()
	if resp.SessionDir == "" {
		t.Fatal("SessionDir empty")
	}
	want := resp.SessionDir + "\n"
	if resp.Stdout != want {
		t.Fatalf("stdout contract violated: want exact %q, got %q", want, resp.Stdout)
	}
}

func combinedErrText(resp *Response) string {
	return resp.Stderr + "\n" + resp.ErrText + "\n" + resp.Stdout
}

func assertContainsFold(t *testing.T, haystack string, needles ...string) {
	t.Helper()
	low := strings.ToLower(haystack)
	for _, n := range needles {
		if !strings.Contains(low, strings.ToLower(n)) {
			t.Fatalf("expected text to contain %q; got:\n%s", n, haystack)
		}
	}
}

func assertNoneContainsFold(t *testing.T, haystack string, needles ...string) {
	t.Helper()
	low := strings.ToLower(haystack)
	for _, n := range needles {
		if strings.Contains(low, strings.ToLower(n)) {
			t.Fatalf("expected text NOT to contain info token %q; got:\n%s", n, haystack)
		}
	}
}

// assertDefaultInfoMilestones checks required default (non-quiet) success
// progress tokens on stderr. Matches substrings, not exact line format.
func assertDefaultInfoMilestones(t *testing.T, stderr string) {
	t.Helper()
	low := strings.ToLower(stderr)
	// 1. listen / addr
	hasListen := strings.Contains(low, "listen") || strings.Contains(low, "listening") ||
		strings.Contains(low, "addr") || strings.Contains(low, "127.0.0.1")
	if !hasListen {
		t.Fatalf("stderr missing listen/addr milestone; stderr:\n%s", stderr)
	}
	// 2. session URL or session id / go?session=
	hasSession := strings.Contains(low, "session") || strings.Contains(low, "/go?") ||
		strings.Contains(low, "go?session")
	if !hasSession {
		t.Fatalf("stderr missing session URL/id milestone; stderr:\n%s", stderr)
	}
	// 3. ready waiting (or ready + timeout language)
	hasReady := strings.Contains(low, "ready") || strings.Contains(low, "waiting")
	if !hasReady {
		t.Fatalf("stderr missing ready/waiting milestone; stderr:\n%s", stderr)
	}
	// 4. recording started
	hasRec := strings.Contains(low, "recording") || strings.Contains(low, "record")
	if !hasRec {
		t.Fatalf("stderr missing recording-started milestone; stderr:\n%s", stderr)
	}
}

// infoMilestoneTokens are tokens that Quiet mode must not emit as progress.
// Errors may still mention "timeout" etc.; Quiet success should avoid these.
var infoMilestoneTokens = []string{
	"listening",
	"listen on",
	"ready wait",
	"waiting for",
	"recording started",
	"start recording",
}

func countFold(haystack, needle string) int {
	low := strings.ToLower(haystack)
	n := strings.ToLower(needle)
	if n == "" {
		return 0
	}
	c := 0
	for {
		i := strings.Index(low, n)
		if i < 0 {
			return c
		}
		c++
		low = low[i+len(n):]
	}
}
```
