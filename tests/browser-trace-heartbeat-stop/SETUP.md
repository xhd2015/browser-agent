# Scenario

**Feature**: install re-run guidance on `/go` and recording heartbeat stop

```
# Session page tells user to close window and re-run after install/reload
User -> GET /go -> HTML (close Chrome window + re-run browser-trace)

# Recording heartbeat: status/entries refresh lastHeartbeatAt
Mock Extension -> POST /v1/status | POST /v1/entries -> lastHeartbeatAt refresh
# Silence > HeartbeatTimeout → partial save + exit 0 warning
Control Server -> recording.har (from previewEntries) + meta (heartbeat_lost, partial)
browser-trace stdout -> "{sessionDir}\n"
browser-trace stderr -> warning (heartbeat / unusual|acceptable)

# Normal complete path still works with continuous heartbeats
Mock Extension -> continuous status/entries -> POST /v1/complete -> exit 0
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Package `browsertrace` implements:
  - `Config.HeartbeatTimeout` (zero → 10s default)
  - Recording-phase heartbeat_lost save + exit 0 + stderr warning
  - `/go` HTML install re-run copy (close window + re-run `browser-trace`)
- Each leaf uses an isolated temp `BaseDir` and ephemeral free `127.0.0.1:port`.
- `NoOpenChrome` is always true in this tree.
- Heartbeat-lost leaves inject short `HeartbeatTimeout` (e.g. 200ms).
- No real Chrome process and no real extension load.
- Ready-phase timeout behavior is **not** under test here.

## Steps

1. Allocate a unique temp `BaseDir` for the leaf.
2. Leave `Addr` empty so `Run` picks a free loopback port.
3. Set `NoOpenChrome = true`.
4. Default ready/complete timeouts to short-but-safe values; descendants may shorten further.
5. Leave `Mode` / `ExtensionScript` empty for grouping nodes to set.

## Context

- Parallel-safe: free ports + temp dirs.
- Shared helpers assert exit codes, stdout trailing newline, meta partial fields,
  HAR URL presence, and flexible stderr warning tokens.
- `DOCTEST_SESSION_ID` available if a future leaf needs session-scoped caches;
  this tree uses the library API only.

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
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	if req.CompleteTimeout == 0 {
		req.CompleteTimeout = 5 * time.Second
	}
	// HeartbeatTimeout left 0 by default → product default inside package;
	// heartbeat-lost leaves inject a short value.
	return nil
}

// sessionDirNameRe matches YYYY-MM-DD-HH-MM-SS-<suffix> (suffix non-empty).
var sessionDirNameRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-\d{2}-\d{2}-\d{2}-.+$`)

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

func assertStdoutSessionPath(t *testing.T, resp *Response) {
	t.Helper()
	if resp.Stdout == "" {
		t.Fatal("stdout is empty")
	}
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("stdout must end with trailing newline; last bytes %q",
			resp.Stdout[len(resp.Stdout)-min(8, len(resp.Stdout)):])
	}
	if resp.SessionDir == "" {
		t.Fatal("SessionDir empty; cannot verify path on stdout")
	}
	if !strings.Contains(resp.Stdout, resp.SessionDir) &&
		!strings.Contains(resp.Stdout, filepath.Base(resp.SessionDir)) {
		t.Fatalf("stdout should mention session path %q; stdout=%q",
			resp.SessionDir, resp.Stdout)
	}
}

func assertHeartbeatWarning(t *testing.T, stderr string) {
	t.Helper()
	low := strings.ToLower(stderr)
	if !strings.Contains(low, "warning") {
		t.Fatalf("stderr should contain warning token; stderr=%q", stderr)
	}
	if !strings.Contains(low, "heartbeat") {
		t.Fatalf("stderr should mention heartbeat; stderr=%q", stderr)
	}
	if !strings.Contains(low, "unusual") && !strings.Contains(low, "acceptable") {
		t.Fatalf("stderr should mention unusual or acceptable; stderr=%q", stderr)
	}
}

func assertNoHeartbeatWarningRequired(t *testing.T, stderr string) {
	t.Helper()
	// Normal complete must not require heartbeat_lost messaging.
	// Soft check: if both "heartbeat" and "warning" appear with lost/timeout language, fail.
	low := strings.ToLower(stderr)
	if strings.Contains(low, "heartbeat_lost") ||
		(strings.Contains(low, "heartbeat") && strings.Contains(low, "unusual")) ||
		(strings.Contains(low, "heartbeat") && strings.Contains(low, "acceptable") && strings.Contains(low, "warning")) {
		t.Fatalf("normal complete path should not emit heartbeat_lost warning; stderr=%q", stderr)
	}
}

func assertMetaHeartbeatLost(t *testing.T, metaJSON []byte) {
	t.Helper()
	if len(metaJSON) == 0 {
		t.Fatal("meta.json is empty or missing")
	}
	var meta map[string]any
	if err := json.Unmarshal(metaJSON, &meta); err != nil {
		t.Fatalf("parse meta.json: %v\n%s", err, metaJSON)
	}
	sr, _ := meta["stop_reason"].(string)
	if sr == "" || !strings.Contains(strings.ToLower(sr), "heartbeat") {
		t.Fatalf("meta.stop_reason = %q, want reason containing heartbeat (e.g. heartbeat_lost)", sr)
	}
	// partial may be bool true or string "true"
	switch v := meta["partial"].(type) {
	case bool:
		if !v {
			t.Fatalf("meta.partial = false, want true")
		}
	case string:
		if !strings.EqualFold(v, "true") {
			t.Fatalf("meta.partial = %q, want true", v)
		}
	default:
		t.Fatalf("meta.partial missing or wrong type %T (want true)", meta["partial"])
	}
}

func assertHARContainsURLs(t *testing.T, harJSON []byte, urls ...string) {
	t.Helper()
	if len(harJSON) == 0 {
		t.Fatal("recording.har is empty or missing")
	}
	harStr := string(harJSON)
	for _, u := range urls {
		if !strings.Contains(harStr, u) {
			t.Fatalf("HAR missing URL %q; har=%s", u, harStr)
		}
	}
}

func assertHAREmptyOrMinimal(t *testing.T, harJSON []byte) {
	t.Helper()
	if len(harJSON) == 0 {
		// Missing file is also acceptable for empty snapshot if product only writes meta;
		// prefer a valid empty HAR when present — require file for implementer clarity.
		t.Fatal("recording.har is empty or missing (want valid HAR with empty entries)")
	}
	var doc struct {
		Log struct {
			Entries []json.RawMessage `json:"entries"`
		} `json:"log"`
	}
	if err := json.Unmarshal(harJSON, &doc); err != nil {
		t.Fatalf("parse recording.har: %v\n%s", err, harJSON)
	}
	if len(doc.Log.Entries) != 0 {
		t.Fatalf("HAR entries = %d, want 0 for empty snapshot", len(doc.Log.Entries))
	}
}

func assertSessionArtifactsExist(t *testing.T, resp *Response) {
	t.Helper()
	if resp.SessionDir == "" {
		t.Fatal("SessionDir empty")
	}
	if _, err := os.Stat(filepath.Join(resp.SessionDir, "meta.json")); err != nil {
		t.Fatalf("meta.json missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(resp.SessionDir, "recording.har")); err != nil {
		t.Fatalf("recording.har missing: %v", err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
```
