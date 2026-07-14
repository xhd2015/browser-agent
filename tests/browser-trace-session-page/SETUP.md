# Scenario

**Feature**: session page live status via Control Server HTTP (no Chrome)

```
# CLI/library starts Control Server; tests never open real Chrome
User -> browser-trace (NoOpenChrome) -> Control Server @ Addr

# Optional staging as Extension Agent
Test Client ?-> POST /v1/hello {version, features}
Test Client ?-> POST /v1/status {state: recording, entry_count, window_id}

# Surfaces under test
Test Client -> GET /v1/session?session=…  -> JSON phase/extension/recording/ready/hint
Test Client -> GET /go?session=…          -> HTML status dashboard + poller JS
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Package `browsertrace` exports `Config` + `Run(ctx, Config) (*Result, error)` and
  serves the control protocol on `cfg.Addr` (including `/v1/health`, `/go`,
  `/v1/hello`, `/v1/status`, and **new** `/v1/session`).
- Each leaf uses an isolated temp `BaseDir` and a free loopback `Addr`.
- `SessionSuffix` is set to a stable known session id string (product uses suffix
  as session id).
- `NoOpenChrome` is always true.
- Ready/complete timeouts are short enough for CI; descendants may tighten further.
- No real Chrome process and no real extension load.
- Capability rule under test: `supports_browser_trace` requires feature
  `browser-trace` **and** version ≥ `1.2.0`. Version alone is not enough.

## Steps

1. Allocate a unique temp `BaseDir` for the leaf.
2. Set `SessionSuffix` to a deterministic id (e.g. `sess-page-<random>`).
3. Set `NoOpenChrome = true`.
4. Default timeouts: ready 5s, complete 2s (enough to health-check, stage, probe, cancel).
5. Leave `Probe`, hello, and status flags unset; grouping/leaf Setup narrows them.

## Context

- Wire protocol bind: loopback only (`127.0.0.1`).
- Parallel-safe: free ports + temp dirs per leaf.
- Probes run while `browsertrace.Run` is still in its ready/recording wait loop;
  Run is cancelled immediately after the HTTP probe returns.
- Shared helpers below are available to all descendant Assert/Setup packages.
- This tree does **not** assert HAR save or CLI stdout lifecycle (see
  `./tests/browser-trace/`).

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
	dir := t.TempDir()
	req.BaseDir = filepath.Join(dir, "browser-trace-base")
	if err := os.MkdirAll(req.BaseDir, 0o755); err != nil {
		return err
	}
	req.NoOpenChrome = true
	if req.SessionSuffix == "" {
		// Stable, unique-enough session id for this leaf process.
		req.SessionSuffix = fmt.Sprintf("sess-%d", time.Now().UnixNano()%1e12)
	}
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	if req.CompleteTimeout == 0 {
		req.CompleteTimeout = 2 * time.Second
	}
	return nil
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
	if !strings.Contains(ct, "json") {
		t.Fatalf("Content-Type %q does not look like JSON; body=%s",
			resp.ContentType, truncate(resp.BodyString, 200))
	}
}

func assertHTMLContentType(t *testing.T, resp *Response) {
	t.Helper()
	ct := strings.ToLower(resp.ContentType)
	if !strings.Contains(ct, "html") {
		t.Fatalf("Content-Type %q does not look like HTML; body=%s",
			resp.ContentType, truncate(resp.BodyString, 200))
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

func assertSessionIDMatch(t *testing.T, resp *Response, want string) {
	t.Helper()
	if resp.SessionID != want {
		t.Fatalf("session_id = %q, want %q; raw phase=%q", resp.SessionID, want, resp.Phase)
	}
}

func assertPhase(t *testing.T, resp *Response, want string) {
	t.Helper()
	if resp.Phase != want {
		t.Fatalf("phase = %q, want %q; body=%s", resp.Phase, want, truncate(resp.BodyString, 400))
	}
}

func assertExtensionConnected(t *testing.T, resp *Response, want bool) {
	t.Helper()
	if resp.ExtensionConnected != want {
		t.Fatalf("extension.connected = %v, want %v; body=%s",
			resp.ExtensionConnected, want, truncate(resp.BodyString, 400))
	}
}

func assertSupportsBrowserTrace(t *testing.T, resp *Response, want bool) {
	t.Helper()
	if resp.SupportsBrowserTrace != want {
		t.Fatalf("extension.supports_browser_trace = %v, want %v; version=%q features=%v body=%s",
			resp.SupportsBrowserTrace, want, resp.ExtensionVersion, resp.ExtensionFeatures,
			truncate(resp.BodyString, 400))
	}
}

func assertHintNonEmpty(t *testing.T, resp *Response) {
	t.Helper()
	if strings.TrimSpace(resp.Hint) == "" {
		t.Fatalf("hint is empty; body=%s", truncate(resp.BodyString, 400))
	}
}

func assertHintMentions(t *testing.T, resp *Response, needles ...string) {
	t.Helper()
	assertHintNonEmpty(t, resp)
	assertContainsFold(t, resp.Hint, needles...)
}

func assertRecording(t *testing.T, resp *Response, active bool, entryCount int) {
	t.Helper()
	if resp.RecordingActive != active {
		t.Fatalf("recording.active = %v, want %v; body=%s",
			resp.RecordingActive, active, truncate(resp.BodyString, 400))
	}
	if resp.EntryCount != entryCount {
		t.Fatalf("recording.entry_count = %d, want %d; body=%s",
			resp.EntryCount, entryCount, truncate(resp.BodyString, 400))
	}
}

func assertReadyCountdownPresent(t *testing.T, resp *Response) {
	t.Helper()
	// Product ready timeout default is 30000ms; tests may use shorter ReadyTimeout.
	// deadline_ms should be positive; remaining_ms should be >= 0 and <= deadline.
	if resp.ReadyDeadlineMS <= 0 {
		t.Fatalf("ready.deadline_ms = %d, want > 0; body=%s",
			resp.ReadyDeadlineMS, truncate(resp.BodyString, 400))
	}
	if resp.ReadyRemainingMS < 0 {
		t.Fatalf("ready.remaining_ms = %d, want >= 0", resp.ReadyRemainingMS)
	}
	if resp.ReadyElapsedMS < 0 {
		t.Fatalf("ready.elapsed_ms = %d, want >= 0", resp.ReadyElapsedMS)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func featuresContain(feats []string, want string) bool {
	for _, f := range feats {
		if f == want {
			return true
		}
	}
	return false
}
```
