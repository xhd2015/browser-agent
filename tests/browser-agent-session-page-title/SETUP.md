# Scenario

**Feature**: session page browser tab title `{sessionId} - Browser Agent`

```
# Inject SPA path (HTTP)
Test Client -> NewSessionRegistry + NewRegistryControlHandler -> httptest
Test Client -> GET /go?session=<id>
  -> HTML <title>{id} - Browser Agent</title>

# Fallback shell (source contract)
Test Client -> read browseragent/server.go writeFallbackSessionHTML
  -> title format {id} - Browser Agent

# React client
Test Client -> read react/src/ui/SessionPageApp.tsx
  -> document.title = sid + " - Browser Agent" when sid known
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Tree root is `tests/browser-agent-session-page-title/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Package `browseragent` already exports `NewSessionRegistry`,
  `NewRegistryControlHandler` (phase3+). Title rewrite is **not** implemented yet
  (classic TDD — RED until implementer).
- Registry leaves use isolated temp `BaseDir` per leaf; no real Chrome.
- Filesystem leaves read module tree via `ModuleRoot`.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate temp `BaseDir` for every leaf (registry leaves use it).
3. Default deterministic `SessionID` when empty.
4. Leave `Mode` and surface-specific fields for grouping/leaf Setup.

## Context

- Spec version **0.0.2**.
- Parallel-safe: temp dirs + httptest per leaf.
- Shared helpers below available to all descendant Assert/Setup packages.
- Classic TDD: assert **desired** title behavior; RED is expected pre-implement.

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
	req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	dir := t.TempDir()
	req.BaseDir = filepath.Join(dir, "browser-agent-base")
	if err := os.MkdirAll(req.BaseDir, 0o755); err != nil {
		return err
	}
	req.NoOpenChrome = true
	if req.SessionID == "" {
		req.SessionID = fmt.Sprintf("ba-title-%d", time.Now().UnixNano()%1e12)
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

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// assertPageTitleIsSessionFormat checks page title equals sessionId + " - Browser Agent".
func assertPageTitleIsSessionFormat(t *testing.T, sessionID, pageTitle, body string) {
	t.Helper()
	want := expectedSessionTitle(sessionID)
	got := strings.TrimSpace(pageTitle)
	if got == "" {
		// Fall back to raw body search for <title>…</title>
		got = extractHTMLTitle(body)
	}
	if got != want {
		// Also accept if body contains exact title tag text even if extract failed on attributes.
		if strings.Contains(body, want) && titleTagContains(body, want) {
			return
		}
		t.Fatalf("page title = %q, want %q; body snippet=%s",
			got, want, truncate(body, 600))
	}
}

func titleTagContains(htmlBody, wantInner string) bool {
	re := regexp.MustCompile(`(?is)<title[^>]*>\s*` + regexp.QuoteMeta(wantInner) + `\s*</title>`)
	return re.MatchString(htmlBody)
}

// assertNotStaticSPATitle fails if the sole page title is the old static SPA title.
func assertNotStaticSPATitle(t *testing.T, pageTitle, body string) {
	t.Helper()
	got := strings.TrimSpace(pageTitle)
	if got == "" {
		got = extractHTMLTitle(body)
	}
	staticSPA := "Browser Agent Session"
	if got == staticSPA {
		t.Fatalf("page title still static %q; want {sessionId} - Browser Agent", staticSPA)
	}
}

// Silence unused import when only helpers reference net/http in some packages.
var _ = http.StatusOK
```
