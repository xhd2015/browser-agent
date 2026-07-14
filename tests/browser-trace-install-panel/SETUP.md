# Scenario

**Feature**: session-page install panel always visible with expand/collapse policy

```
# CLI/library starts Control Server; tests never open real Chrome
User -> browser-trace (NoOpenChrome) -> Control Server @ Addr

# Optional staging as Extension Agent
Test Client ?-> POST /v1/hello {version, features}

# Surfaces under test
Test Client -> GET /go?session=… -> HTML with always-present install panel
  (expanded when !working; collapsed when connected+supports)

# Pure expand policy (optional package helper)
Test Client -> ShouldExpandInstallPanel(connected, supports) -> bool
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Package `browsertrace` exports `Config` + `Run(ctx, Config) (*Result, error)` and
  serves `/v1/health`, `/go`, `/v1/hello` (and related session routes).
- Optional (pure leaves): `ShouldExpandInstallPanel(connected, supports bool) bool`.
- Each leaf uses an isolated temp `BaseDir` and a free loopback `Addr` for HTTP.
- `SessionSuffix` is a stable known session id string.
- `NoOpenChrome` is always true for HTTP leaves.
- Ready/complete timeouts are short enough for CI.
- No real Chrome process and no browser DOM automation (HTML string inspection only).
- Capability rule: `supports_browser_trace` requires feature `browser-trace` **and**
  version ≥ `1.2.0` (same as session-page tree).

## Steps

1. Allocate a unique temp `BaseDir` for the leaf.
2. Set `SessionSuffix` to a deterministic id (e.g. `ip-<random>`).
3. Set `NoOpenChrome = true`.
4. Default timeouts: ready 5s, complete 2s.
5. Leave `Mode`, hello, and pure-helper flags unset; grouping/leaf Setup narrows them.

## Context

- Wire protocol bind: loopback only (`127.0.0.1`).
- Parallel-safe: free ports + temp dirs per leaf.
- `/go` probes run while `browsertrace.Run` is still in its ready wait loop;
  Run is cancelled immediately after the HTTP probe returns.
- Shared helpers below are available to all descendant Assert/Setup packages.
- User-toggle freeze (`data-user-toggled`) is product behavior but **not** covered
  here (would need browser automation).
- Sealed trees `browser-trace-embed-extension` (not-connected panel smoke) and
  `browser-trace-session-page` remain the regression surfaces for shared `/go`.

```go
import (
	"fmt"
	"net/http"
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
	if req.SessionSuffix == "" {
		req.SessionSuffix = fmt.Sprintf("ip-%d", time.Now().UnixNano()%1e12)
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

func assertHTMLContentType(t *testing.T, resp *Response) {
	t.Helper()
	ct := strings.ToLower(resp.ContentType)
	if !strings.Contains(ct, "html") {
		t.Fatalf("Content-Type %q does not look like HTML; body=%s",
			resp.ContentType, truncate(resp.BodyString, 200))
	}
}

// assertInstallPanelPresent requires the always-visible install panel marker.
func assertInstallPanelPresent(t *testing.T, body string) {
	t.Helper()
	if body == "" {
		t.Fatal("HTML body is empty")
	}
	low := strings.ToLower(body)
	hasPanel := strings.Contains(low, "data-browser-trace-install") ||
		strings.Contains(low, `id="browser-trace-install"`) ||
		strings.Contains(low, `id='browser-trace-install'`) ||
		strings.Contains(low, `id="browser-trace-install-panel"`)
	if !hasPanel {
		// Last-resort: class/id fragment (weaker; fail with clear message).
		if !strings.Contains(low, "browser-trace-install") {
			t.Fatalf("HTML missing install panel marker (data-browser-trace-install or id=browser-trace-install); body=%s",
				truncate(body, 700))
		}
	}
}

// assertPanelNotDisplayNone fails if the install panel root is server-hidden via display:none.
func assertPanelNotDisplayNone(t *testing.T, body string) {
	t.Helper()
	// Look for style=...display:none near the install panel open tag.
	// Conservative: flag common patterns on the panel element itself.
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?is)id=["']browser-trace-install["'][^>]*style=["'][^"']*display\s*:\s*none`),
		regexp.MustCompile(`(?is)data-browser-trace-install[^>]*style=["'][^"']*display\s*:\s*none`),
		regexp.MustCompile(`(?is)style=["'][^"']*display\s*:\s*none[^"']*["'][^>]*(id=["']browser-trace-install["']|data-browser-trace-install)`),
	}
	for _, re := range patterns {
		if re.FindString(body) != "" {
			t.Fatalf("install panel must not be display:none in server HTML; body=%s", truncate(body, 700))
		}
	}
}

// panelSnippet returns a lowercased window around the install panel open tag when possible.
func panelSnippet(body string) string {
	low := strings.ToLower(body)
	idx := strings.Index(low, "browser-trace-install")
	if idx < 0 {
		return low
	}
	start := idx - 80
	if start < 0 {
		start = 0
	}
	end := idx + 400
	if end > len(low) {
		end = len(low)
	}
	return low[start:end]
}

// assertPanelExpanded expects open / data-default-open=true signals.
func assertPanelExpanded(t *testing.T, body string) {
	t.Helper()
	snip := panelSnippet(body)
	// Prefer details open attribute or data-default-open="true".
	hasOpenAttr := strings.Contains(snip, " data-default-open=\"true\"") ||
		strings.Contains(snip, " data-default-open='true'") ||
		strings.Contains(snip, " data-default-open=true") ||
		// bare open on details — common: <details ... open> or open=""
		regexp.MustCompile(`(?i)<details[^>]*\sopen(\s|>|/|=)`).FindString(body) != "" ||
		regexp.MustCompile(`(?i)<details[^>]*\sopen=["'][^"']*["']`).FindString(body) != ""

	// Also accept open on the element that carries the panel id/data attr.
	if !hasOpenAttr {
		// data-default-open true anywhere near panel id
		if regexp.MustCompile(`(?is)(data-browser-trace-install|id=["']browser-trace-install["'])[^>]*data-default-open\s*=\s*["']?true`).FindString(body) != "" {
			hasOpenAttr = true
		}
	}
	if !hasOpenAttr {
		// Some implementations put open before id
		if regexp.MustCompile(`(?is)<details[^>]*\bopen\b[^>]*(data-browser-trace-install|id=["']browser-trace-install["'])`).FindString(body) != "" {
			hasOpenAttr = true
		}
	}
	if !hasOpenAttr {
		t.Fatalf("install panel should be expanded (open attr and/or data-default-open=true); panel snip=%s",
			truncate(snip, 500))
	}
}

// assertPanelCollapsed expects no open / data-default-open=false while panel remains.
func assertPanelCollapsed(t *testing.T, body string) {
	t.Helper()
	// Explicit false is a strong collapse signal.
	explicitFalse := regexp.MustCompile(`(?is)data-default-open\s*=\s*["']?false`).FindString(body) != ""

	// If details has open, that is expanded — fail unless we only see open elsewhere.
	openOnPanel := regexp.MustCompile(`(?is)<details[^>]*(data-browser-trace-install|id=["']browser-trace-install["'])[^>]*\bopen\b`).FindString(body) != "" ||
		regexp.MustCompile(`(?is)<details[^>]*\bopen\b[^>]*(data-browser-trace-install|id=["']browser-trace-install["'])`).FindString(body) != ""

	// Also catch non-details roots that incorrectly set open/default-open true.
	explicitTrue := regexp.MustCompile(`(?is)data-default-open\s*=\s*["']?true`).FindString(body) != ""

	if openOnPanel {
		t.Fatalf("install panel should be collapsed (no open on panel details); body snip=%s",
			truncate(panelSnippet(body), 500))
	}
	if explicitTrue && !explicitFalse {
		// data-default-open=true conflicts with collapsed expectation.
		t.Fatalf("install panel collapsed expected but data-default-open=true present; snip=%s",
			truncate(panelSnippet(body), 500))
	}
	// Accept either explicit false or simply no open / no default-open true.
	// Require panel still present (caller should already assert).
	_ = explicitFalse
}

// assertInstallGuidance checks path + chrome://extensions (regression #5).
func assertInstallGuidance(t *testing.T, req *Request, body string) {
	t.Helper()
	if !strings.Contains(body, "chrome://extensions") {
		t.Fatalf("HTML must include chrome://extensions as text; body=%s", truncate(body, 600))
	}
	low := strings.ToLower(body)
	hasPath := strings.Contains(body, "/extension/") ||
		strings.Contains(body, req.BaseDir) ||
		strings.Contains(low, "data-extension-path") ||
		strings.Contains(low, "data-install-path")
	if !hasPath {
		re := regexp.MustCompile(`/extension/[^"'<\s]+`)
		if re.FindString(body) != "" {
			hasPath = true
		}
	}
	if !hasPath {
		t.Fatalf("HTML must show extension install path guidance; body=%s", truncate(body, 700))
	}
}

func assertSessionIDInBody(t *testing.T, req *Request, resp *Response) {
	t.Helper()
	wantID := resp.RealSessionID
	if wantID == "" {
		wantID = req.SessionSuffix
	}
	if !strings.Contains(resp.BodyString, wantID) {
		t.Fatalf("HTML missing session id %q; body=%s", wantID, truncate(resp.BodyString, 400))
	}
}

func assertExpandResult(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("ShouldExpandInstallPanel transport/run error: %v", err)
	}
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.ExpandCalled {
		t.Fatal("ExpandCalled is false; Run did not execute should-expand mode")
	}
	if resp.ExpandResult != req.WantExpand {
		t.Fatalf("ShouldExpandInstallPanel(%v, %v) = %v, want %v",
			req.Connected, req.Supports, resp.ExpandResult, req.WantExpand)
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
