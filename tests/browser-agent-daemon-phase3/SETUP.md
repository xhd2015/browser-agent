# Scenario

**Feature**: Phase 3 multi-session HTTP control plane via SessionRegistry

```
# Registry-backed httptest (most leaves)
Test Client -> NewSessionRegistry(baseDir, addr)
Test Client -> NewRegistryControlHandler(registry) -> httptest.Server
Test Client -> POST/GET /v1/sessions | /v1/session | /v1/jobs | /go

# Regression bridge (one leaf)
Test Client -> browseragent.Run(NoOpenChrome, NoAgentRun) -> registry-backed serve
Test Client -> GET /v1/session?session=<live id>
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` importable.
- Phase 2 `SessionRegistry` exists; Phase 3 adds `NewRegistryControlHandler` and
  multi-session route behavior (tests are **RED** until implementer lands).
- Tree root is `tests/browser-agent-daemon-phase3/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Registry leaves use isolated temp `BaseDir` per leaf; no real Chrome.
- Fake extension WS helpers are defined in root `DOCTEST.md` for reuse (not used
  by default phase 3 leaves).

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate temp `BaseDir` for every leaf.
3. Default deterministic `SessionID` when empty.
4. Leave `Mode` and surface-specific fields for grouping/leaf Setup.

## Context

- Spec version **0.0.2**.
- `httptest.Server` URL host is used for registry `addr` metadata unless leaf sets `Addr`.
- After implementer, `doctest test ./tests/browser-agent/...` must remain GREEN.

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
	req.BaseDir = filepath.Join(dir, "browser-agent-base")
	if err := os.MkdirAll(req.BaseDir, 0o755); err != nil {
		return err
	}
	req.NoOpenChrome = true
	if req.SessionID == "" {
		req.SessionID = fmt.Sprintf("ba-p3-%d", time.Now().UnixNano()%1e12)
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
	if !strings.Contains(ct, "json") && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if resp.Raw == nil && len(resp.Body) > 0 && resp.Body[0] != '{' && resp.Body[0] != '[' {
			t.Fatalf("Content-Type %q does not look like JSON; body=%s",
				resp.ContentType, truncate(resp.BodyString, 200))
		}
	}
}

func assertHTMLContentType(t *testing.T, resp *Response) {
	t.Helper()
	ct := strings.ToLower(resp.ContentType)
	if !strings.Contains(ct, "html") {
		low := strings.ToLower(resp.BodyString)
		if !strings.Contains(low, "<html") && !strings.Contains(low, "<!doctype") {
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

func assertDisconnectedHint(t *testing.T, hint, sessionID string) {
	t.Helper()
	if strings.TrimSpace(hint) == "" {
		t.Fatal("hint is empty; want disconnected guidance mentioning /go?session=")
	}
	assertContainsFold(t, hint, "/go?session="+sessionID, "/go?session=")
	// Accept paraphrases for "keep open" / "do not navigate away".
	low := strings.ToLower(hint)
	hasKeep := strings.Contains(low, "keep") && strings.Contains(low, "open")
	hasDontNav := strings.Contains(low, "do not") || strings.Contains(low, "don't") ||
		strings.Contains(low, "not close") || strings.Contains(low, "same window") ||
		strings.Contains(low, "navigat")
	if !hasKeep && !hasDontNav {
		t.Fatalf("hint missing keep-open / do-not-navigate guidance; hint=%q", truncate(hint, 400))
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func sessionURLContainsID(baseURL, sessionID string) string {
	return "/go?session=" + sessionID
}

// Silence unused import when only helpers reference net/http in some packages.
var _ = http.StatusOK
```