# Scenario

**Feature**: session new --browser firefox (open URL-only + print firefox ext path)

```
# Pure argv
Test Client -> BuildFirefoxOpenArgs(sessionURL)
  -> argv contains URL only; no --load-extension / --user-data-dir

# Package SessionNew (firefox)
SessionNew(Browser=firefox, OpenFirefoxFn, Home=TestHome, NoWait)
  -> EnsureCanonicalFirefoxExtensionWithHome
  -> OpenFirefoxFn(sessionURL) once  # unless NoOpenChrome
  -> stdout: browser-agent-firefox path + about:debugging

# Package SessionNew (default chrome smoke)
SessionNew(Browser="", OpenChromeFn, OpenFirefoxFn, NoWait)
  -> OpenChromeFn once; OpenFirefoxFn never

# CLI
Operator -> session new --browser firefox -> inject OpenFirefoxFn + path + about:debugging
Operator -> session new --help -> documents --browser
Operator -> session new --browser safari -> unknown browser error, nonzero
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Tree root is `tests/browser-agent-firefox-session-new/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- P1 APIs already GREEN: `EnsureCanonicalFirefoxExtensionWithHome`,
  `install-firefox-extension` (not re-tested as primary here).
- P2 APIs (RED until implementer):
  - `BuildFirefoxOpenArgs(sessionURL string) []string`
  - `SessionNewConfig.Browser`, `SessionNewConfig.OpenFirefoxFn`
  - CLI `--browser chrome|firefox`; inject `SessionNewHooks.OpenFirefoxFn`
- No real Firefox/Chrome; session-new leaves use ephemeral `RunDaemon` + `NoWait`.
- Per-leaf `TestHome` / `BaseDir` isolation (parallel-safe).

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate temp `BaseDir` and `TestHome` per leaf.
3. Default `SessionID = "sess-ff-new-1"`, `ReadyTimeout = 5s`, `NoWait = true`.
4. Leave `Mode` and op-specific fields for grouping/leaf Setup.

## Context

- Spec version **0.0.2**.
- Canonical Firefox segment: `extensions/browser-agent-firefox/<version>/`.
- Chrome default when `--browser` omitted remains unchanged.
- Shared assertion helpers below available to all descendant Assert packages.

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
	req.TestHome = filepath.Join(dir, "home")
	if err := os.MkdirAll(req.BaseDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(req.TestHome, 0o755); err != nil {
		return err
	}
	if req.SessionID == "" {
		req.SessionID = "sess-ff-new-1"
	}
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	req.NoWait = true
	return nil
}

func assertNoRunErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run transport/setup error: %v", err)
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
			t.Fatalf("expected text NOT to contain %q; got:\n%s", n, truncate(haystack, 800))
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// assertNoManagedFirefoxFlags fails if argv contains --load-extension or --user-data-dir.
func assertNoManagedFirefoxFlags(t *testing.T, args []string) {
	t.Helper()
	for _, a := range args {
		if a == "--load-extension" || strings.HasPrefix(a, "--load-extension=") {
			t.Fatalf("Firefox open args must not include --load-extension; args=%v", args)
		}
		if a == "--user-data-dir" || strings.HasPrefix(a, "--user-data-dir=") {
			t.Fatalf("Firefox open args must not include --user-data-dir; args=%v", args)
		}
	}
}

// assertFirefoxSessionNewStdout checks path + about:debugging markers.
func assertFirefoxSessionNewStdout(t *testing.T, stdout string) {
	t.Helper()
	if stdout == "" {
		t.Fatal("stdout is empty")
	}
	assertContainsFold(t, stdout,
		"browser-agent-firefox",
		"about:debugging",
	)
	// Load Temporary is preferred; accept "temporary add-on" wording variants.
	low := strings.ToLower(stdout)
	if !strings.Contains(low, "load temporary") && !strings.Contains(low, "temporary add-on") && !strings.Contains(low, "temporary addon") {
		t.Fatalf("stdout should mention Load Temporary Add-on (or temporary add-on); got:\n%s", truncate(stdout, 800))
	}
	// Primary path must not be Chrome Load unpacked instructions.
	assertNotContainsFold(t, stdout, "chrome://extensions")
	if !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("stdout must end with \\n; last bytes=%q", tail(stdout, 40))
	}
}

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
```
