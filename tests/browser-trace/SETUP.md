# Scenario

**Feature**: browser-trace control server session with mock extension (no Chrome)

```
# CLI starts Control Server; tests never open real Chrome
User -> browser-trace (NoOpenChrome) -> Control Server @ Addr

# Mock Extension speaks the same HTTP wire protocol as Chrome-Ext-Capture-API
Mock Extension -> POST /v1/hello
Mock Extension -> GET /v1/commands (long-poll start|stop)
Mock Extension -> POST /v1/status (recording)
Mock Extension -> POST /v1/complete (HAR + stop_reason)

# On success, Storage writes session dir under BaseDir
Control Server -> BaseDir/YYYY-MM-DD-HH-MM-SS-suffix/{meta.json,recording.har}
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Package `browsertrace` implements `Config` + `Run(ctx, Config) (*Result, error)` as
  expected by root `DOCTEST.md` `Run` (created by implementer during TDD green).
- Each leaf uses an isolated temp `BaseDir` under the test temp directory.
- Each leaf uses an ephemeral free `127.0.0.1:port` unless a leaf sets a specific
  `Addr` (port-conflict leaf occupies that address first).
- `NoOpenChrome` is always true in this tree.
- Product default timeouts are 30s; timeout-path leaves inject short durations
  (e.g. 300–800ms) so tests stay fast.
- No real Chrome process and no real extension load.

## Steps

1. Allocate a unique temp `BaseDir` for the leaf.
2. Leave `Addr` empty so `Run` picks a free loopback port (conflict leaf overrides).
3. Set `NoOpenChrome = true`.
4. Set default timeouts to product values; descendants shorten them for timeout paths.
5. Default `ExtensionScript` to `none` and `StopMode` to `none`; descendants override.

## Context

- Wire protocol bind: loopback only (`127.0.0.1`), never `0.0.0.0`.
- Parallel-safe: free ports + temp dirs; conflict scenario binds a real listener.
- `DOCTEST_SESSION_ID` is available for future session-scoped binary caches; this
  tree runs against the library API and does not build a CLI binary yet.
- Shared helpers below are available to all descendant Assert/Setup packages.

```go
import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
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
	// Product defaults; timeout leaves override to short values.
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
	return nil
}

// sessionDirNameRe matches YYYY-MM-DD-HH-MM-SS-<suffix> (suffix non-empty).
var sessionDirNameRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-\d{2}-\d{2}-\d{2}-.+$`)

func assertExitNonZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ExitCode == 0 {
		t.Fatalf("exit code = 0, want ≠ 0; stderr=%q err=%q stdout=%q", resp.Stderr, resp.ErrText, resp.Stdout)
	}
}

func assertExitZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q err=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.ErrText, resp.Stdout)
	}
}

func combinedErrText(resp *Response) string {
	return strings.ToLower(resp.Stderr + "\n" + resp.ErrText + "\n" + resp.Stdout)
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

func assertSessionDirPattern(t *testing.T, sessionDir, baseDir string) {
	t.Helper()
	if sessionDir == "" {
		t.Fatal("SessionDir is empty")
	}
	rel, err := filepath.Rel(baseDir, sessionDir)
	if err != nil {
		t.Fatal(err)
	}
	// Must be a single path segment under baseDir.
	if strings.Contains(rel, string(filepath.Separator)) || strings.HasPrefix(rel, "..") {
		t.Fatalf("session dir %q is not a direct child of base %q (rel=%q)", sessionDir, baseDir, rel)
	}
	name := filepath.Base(sessionDir)
	if !sessionDirNameRe.MatchString(name) {
		t.Fatalf("session dir name %q does not match YYYY-MM-DD-HH-MM-SS-<suffix>", name)
	}
}

func assertHARHasMergedEntries(t *testing.T, harJSON []byte, minEntries int) {
	t.Helper()
	if len(harJSON) == 0 {
		t.Fatal("recording.har is empty or missing")
	}
	var doc struct {
		Log struct {
			Entries []json.RawMessage `json:"entries"`
		} `json:"log"`
	}
	if err := json.Unmarshal(harJSON, &doc); err != nil {
		t.Fatalf("parse recording.har: %v\n%s", err, harJSON)
	}
	if len(doc.Log.Entries) < minEntries {
		t.Fatalf("HAR entries = %d, want >= %d", len(doc.Log.Entries), minEntries)
	}
}

func assertMetaPresent(t *testing.T, metaJSON []byte) {
	t.Helper()
	if len(metaJSON) == 0 {
		t.Fatal("meta.json is empty or missing")
	}
	var meta map[string]any
	if err := json.Unmarshal(metaJSON, &meta); err != nil {
		t.Fatalf("parse meta.json: %v\n%s", err, metaJSON)
	}
}

func assertStdoutTrailingNewline(t *testing.T, stdout string) {
	t.Helper()
	if stdout == "" {
		t.Fatal("stdout is empty")
	}
	if !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("stdout must end with trailing newline; last bytes %q", stdout[len(stdout)-min(8, len(stdout)):])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func harFileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.Size() > 0
}
```
