# Scenario

**Feature**: auto-exclude control traffic; live entries push; server preview + clear

```
# Pure filter (extension capture gate)
# request URL under product control hosts must not enter entries
URL -> ShouldCaptureURL -> false for 127.0.0.1:43759 and localhost:43759
URL -> ShouldCaptureURL -> true for normal https / other ports

# Control Server live preview pipeline (mock extension push; no Chrome)
User -> browser-trace (NoOpenChrome) -> Control Server @ Addr
Mock Extension -> POST /v1/entries {session_id, entries, count}
  (periodic ~1s in product; harness posts on demand)
  (clear = POST empty entries while still recording)
Test Client -> GET /v1/entries?session=… -> {entries, count, updated_at}
Test Client -> GET /preview?session=… -> HTML live viewer (poll /v1/entries)
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Package `browsertrace` exports `Config` + `Run(ctx, Config) (*Result, error)` and
  serves at least `/v1/health` (existing).
- Implementer must add:
  - `ShouldCaptureURL(url string) bool`
  - `POST /v1/entries`, `GET /v1/entries`, `GET /preview`
- Each HTTP leaf uses an isolated temp `BaseDir` and a free loopback `Addr`.
- `SessionSuffix` is a stable known session id string for the live session.
- `NoOpenChrome` is always true for HTTP leaves.
- Ready/complete timeouts are short enough for CI.
- No real Chrome, no popup confirm UI, no extension process.
- Fixture entry URLs are **never** control hosts (exclude is client-side; server
  stores what is posted).
- Product control exclude hosts are fixed: `127.0.0.1:43759` and `localhost:43759`
  (not the ephemeral test bind port).

## Steps

1. Allocate a unique temp `BaseDir` for the leaf (HTTP leaves).
2. Set `SessionSuffix` to a deterministic id (e.g. `ecp-<random>`).
3. Set `NoOpenChrome = true`.
4. Default timeouts: ready 5s, complete 2s.
5. Leave `Mode`, pure URL, and HTTP probe/stage flags unset; grouping/leaf
   Setup narrows them.

## Context

- Wire protocol bind: loopback only (`127.0.0.1`).
- Parallel-safe: free ports + temp dirs per leaf.
- HTTP probes run while `browsertrace.Run` is still in its ready wait loop;
  Run is cancelled immediately after the final GET returns.
- Shared helpers below are available to all descendant Assert/Setup packages.
- Popup confirm text (`Discard all captured requests so far?`) and extension
  fallback `preview.html` are **out of scope** (implementer / manual).
- Sealed trees `browser-trace`, `browser-trace-session-page`, and
  `browser-trace-install-panel` remain regression surfaces for unrelated routes.

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
	req.BaseDir = filepath.Join(dir, "browser-trace-base")
	if err := os.MkdirAll(req.BaseDir, 0o755); err != nil {
		return err
	}
	req.NoOpenChrome = true
	if req.SessionSuffix == "" {
		req.SessionSuffix = fmt.Sprintf("ecp-%d", time.Now().UnixNano()%1e12)
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
	// Allow missing Content-Type if body is valid JSON (early impls).
	if ct == "" {
		return
	}
	if !strings.Contains(ct, "json") {
		t.Fatalf("Content-Type %q does not look like JSON; body=%s",
			resp.ContentType, truncate(resp.BodyString, 200))
	}
}

func assertHTMLContentType(t *testing.T, resp *Response) {
	t.Helper()
	ct := strings.ToLower(resp.ContentType)
	if ct == "" {
		// Accept empty if body looks like HTML.
		low := strings.ToLower(resp.BodyString)
		if strings.Contains(low, "<html") || strings.Contains(low, "<!doctype") ||
			strings.Contains(low, "<body") || strings.Contains(low, "<table") {
			return
		}
	}
	if !strings.Contains(ct, "html") {
		t.Fatalf("Content-Type %q does not look like HTML; body=%s",
			resp.ContentType, truncate(resp.BodyString, 200))
	}
}

func assertCaptureResult(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("ShouldCaptureURL run error: %v", err)
	}
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.CaptureCalled {
		t.Fatal("CaptureCalled is false; Run did not execute should-capture mode")
	}
	if resp.CaptureResult != req.WantCapture {
		t.Fatalf("ShouldCaptureURL(%q) = %v, want %v",
			req.CaptureURL, resp.CaptureResult, req.WantCapture)
	}
}

func assertNotFoundBody(t *testing.T, body string) {
	t.Helper()
	low := strings.ToLower(body)
	ok := strings.Contains(low, "not found") ||
		strings.Contains(low, "not_found") ||
		strings.Contains(low, "unknown session") ||
		strings.Contains(low, "no such session") ||
		strings.Contains(low, "session not found") ||
		strings.Contains(low, `"error"`)
	if !ok {
		t.Fatalf("404 body should indicate not-found/session error; body=%s",
			truncate(body, 400))
	}
}

// sampleEntries returns two minimal HAR-like entries with non-control URLs.
func sampleEntries() []map[string]any {
	return []map[string]any{
		{
			"startedDateTime": "2026-07-11T00:00:00.000Z",
			"request": map[string]any{
				"method": "GET",
				"url":    "https://api.example.com/v1/alpha",
			},
			"response": map[string]any{
				"status": 200,
			},
		},
		{
			"startedDateTime": "2026-07-11T00:00:01.000Z",
			"request": map[string]any{
				"method": "POST",
				"url":    "https://cdn.example.com/assets/app.js",
			},
			"response": map[string]any{
				"status": 200,
			},
		},
	}
}

func sampleEntryURLs() []string {
	return []string{
		"https://api.example.com/v1/alpha",
		"https://cdn.example.com/assets/app.js",
	}
}

func assertEntryURLsMatch(t *testing.T, resp *Response, want []string) {
	t.Helper()
	if resp.EntriesCount != len(want) {
		// Prefer explicit count field; also check len(entries).
		if len(resp.Entries) != len(want) {
			t.Fatalf("entries count = %d (len=%d), want %d; urls=%v body=%s",
				resp.EntriesCount, len(resp.Entries), len(want), resp.EntryURLs,
				truncate(resp.BodyString, 400))
		}
	}
	// Every wanted URL must appear in EntryURLs or raw body.
	for _, u := range want {
		found := false
		for _, got := range resp.EntryURLs {
			if got == u {
				found = true
				break
			}
		}
		if !found && strings.Contains(resp.BodyString, u) {
			found = true
		}
		if !found {
			t.Fatalf("missing entry URL %q; got urls=%v body=%s",
				u, resp.EntryURLs, truncate(resp.BodyString, 400))
		}
	}
}

func assertPreviewLiveMarkers(t *testing.T, body string) {
	t.Helper()
	low := strings.ToLower(body)
	// Poll script or entries API reference, or explicit preview root marker.
	hasPoll := strings.Contains(low, "/v1/entries") ||
		strings.Contains(low, "v1/entries") ||
		strings.Contains(low, "data-browser-trace-preview") ||
		strings.Contains(low, "id=\"browser-trace-preview\"") ||
		strings.Contains(low, "id='browser-trace-preview'") ||
		strings.Contains(low, "live preview") ||
		strings.Contains(low, "preview-entries")
	if !hasPoll {
		t.Fatalf("preview HTML should reference /v1/entries poll or preview root marker; body=%s",
			truncate(body, 600))
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// silence unused imports in leaves that only need a subset
var (
	_ = http.StatusOK
	_ = os.PathSeparator
)
```
