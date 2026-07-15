# browser-agent session page title

Classic-TDD tree for the **session page browser tab title** on package
`github.com/xhd2015/browser-agent/browseragent` + React session SPA:

| Surface | What is under test |
|---------|-------------------|
| GET `/go?session=<id>` (inject SPA path) | Served HTML `<title>` is `{id} - Browser Agent` |
| Fallback `writeFallbackSessionHTML` | Same title format in pure-Go shell (source / format contract) |
| React `SessionPageApp` | Sets `document.title` to `{sid} - Browser Agent` when sid known |

**No real Chrome.** **No Playwright.** **No real agent-run.** Registry
`httptest` for `/go`; filesystem source probes for fallback + React.

## Mode

**Classic TDD** — feature is **not implemented yet**. Static titles remain
`Browser Agent Session` (SPA source/embed) / `browser-agent session` (fallback
Go HTML). No `document.title` update in React. Expect **RED** leaves until
implementer lands changes. Do **not** implement production code from this tree.

## Product decisions (authoritative)

1. **Format**: `{sessionId} - Browser Agent` (spaces around hyphen; product
   suffix is **Browser Agent**, not bare `Agent`).
2. **Fallback shell**: yes — pure-Go `writeFallbackSessionHTML` uses the same
   format.
3. **Both paths**: server inject **and** React client set `document.title`.

Example: session id `sess-a1b2c3` → title `sess-a1b2c3 - Browser Agent`.

## Non-goals

- Popup title (`Browser Agent`).
- In-page `<h1>` copy.
- Fake-extension mock tab titles in other doctest trees.
- Storage / API / protocol changes.

## Version

0.0.2

# DSN (Domain Specific Notion)

**Session page** is the HTML served at `GET /go?session=<id>`. The browser tab
title must encode the live **session id** so operators can distinguish multiple
session windows.

```text
title = sessionId + " - Browser Agent"
# e.g. "sess-title - Browser Agent"
```

**Inject path** (primary, embed present):

```text
readEmbeddedSessionIndex() -> injectSessionBoot(html, sessionID, snap)
  rewrites / inserts <title>{sessionID} - Browser Agent</title>
  -> HTTP 200 text/html
```

**Fallback path** (embed unavailable):

```text
writeFallbackSessionHTML(w, sessionID, snap)
  <title>{sessionID} - Browser Agent</title>   # same format
```

**React client** (SPA after boot):

```text
SessionPageApp: when sid known
  document.title = sid + " - Browser Agent"
when sid empty/missing
  do NOT set broken " - Browser Agent"; leave static or skip update
```

**Test Client** in this tree:

- `go-html` leaves: `NewSessionRegistry` + `NewRegistryControlHandler` +
  `httptest.Server`; pre-create session; `GET /go?session=`.
- `go-src` leaf: read `browseragent/server.go` (or sibling) for fallback title
  format tokens (embed always present makes pure-HTTP fallback hard).
- `react-src` leaf: read `react/src/ui/SessionPageApp.tsx` (or apps entry) for
  `document.title` assignment.

## Decision Tree

```
browser-agent-session-page-title
├── go-html/                                   [GET /go inject SPA path]
│   └── title-includes-session-id/               T1 HTML <title>{id} - Browser Agent</title>
├── go-src/                                    [fallback shell source contract]
│   └── fallback-title-format/                   T2 writeFallbackSessionHTML title format
└── react-src/                                 [SessionPageApp document.title]
    └── sets-document-title/                     T3 document.title = sid + " - Browser Agent"
```

### Parameter significance (high → low)

1. **Surface / Mode** — HTTP inject (`/go`) vs Go fallback source vs React source
   (different `Run` branches).
2. **Outcome** — dynamic title format correct vs static sole title forbidden.
3. **Session presence** — known session id appears in title text; React empty sid
   must not produce a broken title.

## Test Index

| Leaf | Scenario |
|------|----------|
| `go-html/title-includes-session-id` | (T1) Pre-create `sess-title`; GET `/go?session=sess-title` → 200 HTML; `<title>sess-title - Browser Agent</title>`; not sole static `Browser Agent Session` |
| `go-src/fallback-title-format` | (T2) `writeFallbackSessionHTML` source uses `{sessionId} - Browser Agent` title format (not sole static `browser-agent session`) |
| `react-src/sets-document-title` | (T3) `SessionPageApp` assigns `document.title` with session id + ` - Browser Agent`; empty sid does not set broken `" - Browser Agent"` |

**Leaf count: 3**

## How to Run

```sh
# module root
doctest vet ./tests/browser-agent-session-page-title
doctest test ./tests/browser-agent-session-page-title
# expect RED until implementer lands feature
```

Module: `github.com/xhd2015/browser-agent`.  
Package under test: `…/browseragent` + `react/src/ui/SessionPageApp.tsx`.

### Implementer contract (authoritative for GREEN)

**Exact title format** (spaces around hyphen):

```text
sessionId + " - Browser Agent"
```

**Go — `browseragent/server.go`**

1. `injectSessionBoot` — rewrite or insert `<title>{sessionID} - Browser Agent</title>`
   (session id HTML-escaped when placed in markup).
2. `writeFallbackSessionHTML` — same title string (not static
   `browser-agent session`).
3. `handleGo` continues preferring embed + inject; fallback only when embed missing.

**React — `react/src/ui/SessionPageApp.tsx`**

```ts
// when sid is non-empty:
document.title = `${sid} - Browser Agent`;
// when sid empty: skip update (do not set " - Browser Agent")
```

Prefer `useEffect` keyed on `sid`. Optional: keep/extend package unit tests in
`browseragent/session_page_inject_test.go` (not required by this tree).

**Non-goals remain**: popup title, `<h1>` body copy, protocol/storage.

```go
import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/browser-agent/browseragent"
)

// Mode values — top-level surface under test.
const (
	ModeGoHTML   = "go-html"
	ModeGoSrc    = "go-src"
	ModeReactSrc = "react-src"
)

// GoSrcProbe for ModeGoSrc.
const (
	GoSrcFallbackTitle = "fallback-title-format"
)

// ReactProbe for ModeReactSrc.
const (
	ReactProbeDocumentTitle = "sets-document-title"
)

// TitleSuffix is the exact product suffix including spaces around the hyphen.
const TitleSuffix = " - Browser Agent"

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	// Mode selects the surface Run executes.
	Mode string

	// ModuleRoot is module directory (filesystem leaves).
	// Root Setup sets from DOCTEST_ROOT/../..
	ModuleRoot string

	// BaseDir is temp parent for registry fixtures.
	BaseDir string

	// Session / server (go-html)
	Addr                string
	SessionID           string
	PreCreateSessionIDs []string
	GoSessionQuery      string
	GoUnknownSession    bool
	UnknownSessionID    string
	ReadyTimeout        time.Duration
	NoOpenChrome        bool

	// --- go-src ---
	GoSrcProbe string

	// --- react-src ---
	ReactProbe string
}

// Response holds outcomes for all modes (fields used per mode).
type Response struct {
	// HTTP (go-html)
	StatusCode  int
	ContentType string
	Body        []byte
	BodyString  string
	BaseURL     string
	ProbeURL    string
	// RealSessionID is the session id used for the probe (post-create).
	RealSessionID string
	// PageTitle is the first <title>…</title> text when parseable.
	PageTitle string

	// Filesystem probes
	FoundPaths   []string
	FileExists   bool
	CombinedText string
	FileContents map[string]string
	ErrText      string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	if req.ModuleRoot == "" {
		req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	}
	switch req.Mode {
	case ModeGoHTML:
		return runGoHTML(t, req)
	case ModeGoSrc:
		return runGoSrc(t, req)
	case ModeReactSrc:
		return runReactSrc(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

// --- GET /go HTML (registry control handler) ---

func runGoHTML(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	srv, cleanup, err := startRegistryHTTPServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{BaseURL: srv.BaseURL, RealSessionID: srv.PrimarySessionID}

	sessionID := srv.PrimarySessionID
	if req.GoUnknownSession {
		sessionID = req.UnknownSessionID
		if sessionID == "" {
			sessionID = "does-not-exist"
		}
	} else if req.GoSessionQuery != "" {
		sessionID = req.GoSessionQuery
	}
	resp.RealSessionID = sessionID

	probeURL := srv.BaseURL + "/go?session=" + url.QueryEscape(sessionID)
	status, ct, rawBody, err := doGET(probeURL)
	if err != nil {
		return resp, err
	}
	resp.StatusCode = status
	resp.ContentType = ct
	resp.Body = rawBody
	resp.BodyString = string(rawBody)
	resp.ProbeURL = probeURL
	resp.PageTitle = extractHTMLTitle(resp.BodyString)
	return resp, nil
}

// --- Go source: fallback / inject title format ---

func runGoSrc(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.GoSrcProbe == "" {
		req.GoSrcProbe = GoSrcFallbackTitle
	}
	root := req.ModuleRoot
	resp := &Response{FileContents: map[string]string{}}

	switch req.GoSrcProbe {
	case GoSrcFallbackTitle:
		candidates := []string{
			filepath.Join(root, "browseragent", "server.go"),
			filepath.Join(root, "browseragent", "session_page.go"),
			filepath.Join(root, "browseragent", "go_html.go"),
		}
		// Prefer server.go; also accept any sibling that defines writeFallbackSessionHTML.
		path, data, ok := firstExistingFile(candidates)
		if !ok {
			// Walk browseragent/*.go for writeFallbackSessionHTML.
			dir := filepath.Join(root, "browseragent")
			entries, _ := os.ReadDir(dir)
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
					continue
				}
				p := filepath.Join(dir, e.Name())
				b, err := os.ReadFile(p)
				if err != nil {
					continue
				}
				if strings.Contains(string(b), "writeFallbackSessionHTML") {
					path, data, ok = p, b, true
					break
				}
			}
		}
		resp.FileExists = ok
		if ok {
			resp.FoundPaths = []string{path}
			resp.FileContents[path] = string(data)
			resp.CombinedText = string(data)
		} else {
			resp.ErrText = "writeFallbackSessionHTML source not found under browseragent/"
		}
		return resp, nil
	default:
		return nil, fmt.Errorf("unknown GoSrcProbe %q", req.GoSrcProbe)
	}
}

// --- React source: document.title ---

func runReactSrc(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ReactProbe == "" {
		req.ReactProbe = ReactProbeDocumentTitle
	}
	root := req.ModuleRoot
	reactRoot := filepath.Join(root, "react")
	if st, err := os.Stat(reactRoot); err != nil || !st.IsDir() {
		alt := filepath.Join(root, "project-api-capture-react")
		if st2, err2 := os.Stat(alt); err2 == nil && st2.IsDir() {
			reactRoot = alt
		}
	}
	resp := &Response{FileContents: map[string]string{}}

	switch req.ReactProbe {
	case ReactProbeDocumentTitle:
		candidates := []string{
			filepath.Join(reactRoot, "src", "ui", "SessionPageApp.tsx"),
			filepath.Join(reactRoot, "src", "ui", "SessionPageApp.ts"),
			filepath.Join(reactRoot, "src", "ui", "SessionPageApp.jsx"),
			filepath.Join(reactRoot, "src", "apps", "session-page", "SessionPageApp.tsx"),
			filepath.Join(reactRoot, "src", "apps", "session-page", "App.tsx"),
			filepath.Join(reactRoot, "src", "apps", "session-page", "main.tsx"),
		}
		path, data, ok := firstExistingFile(candidates)
		// If main.tsx only, also try to combine SessionPageApp from ui/.
		if ok {
			resp.FileExists = true
			resp.FoundPaths = []string{path}
			resp.FileContents[path] = string(data)
			resp.CombinedText = string(data)
			// Augment with ui/SessionPageApp if main was chosen and ui exists.
			uiPath := filepath.Join(reactRoot, "src", "ui", "SessionPageApp.tsx")
			if path != uiPath {
				if b, err := os.ReadFile(uiPath); err == nil {
					resp.FoundPaths = append(resp.FoundPaths, uiPath)
					resp.FileContents[uiPath] = string(b)
					resp.CombinedText = resp.CombinedText + "\n" + string(b)
				}
			}
		} else {
			resp.ErrText = "SessionPageApp source not found under react/src/ui or apps/session-page"
		}
		return resp, nil
	default:
		return nil, fmt.Errorf("unknown ReactProbe %q", req.ReactProbe)
	}
}

// --- registry httptest harness (phase3 pattern) ---

type registryHTTPServer struct {
	BaseURL          string
	PrimarySessionID string
	registry         *browseragent.SessionRegistry
	server           *httptest.Server
}

func startRegistryHTTPServer(t *testing.T, req *Request) (*registryHTTPServer, func(), error) {
	t.Helper()
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set by root Setup")
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, nil, err
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	if req.Addr != "" {
		addr = req.Addr
	}

	reg := browseragent.NewSessionRegistry(req.BaseDir, addr)
	for _, id := range req.PreCreateSessionIDs {
		if _, err := reg.Create(id); err != nil {
			return nil, nil, fmt.Errorf("registry pre-create %q: %w", id, err)
		}
	}

	handler := browseragent.NewRegistryControlHandler(reg)
	ts := httptest.NewServer(handler)
	baseURL := ts.URL

	primary := req.SessionID
	if primary == "" && len(req.PreCreateSessionIDs) > 0 {
		primary = req.PreCreateSessionIDs[0]
	}

	out := &registryHTTPServer{
		BaseURL:          baseURL,
		PrimarySessionID: primary,
		registry:         reg,
		server:           ts,
	}
	cleanup := func() { ts.Close() }
	return out, cleanup, nil
}

func doGET(rawURL string) (int, string, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, "", nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, "", nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, res.Header.Get("Content-Type"), nil, err
	}
	return res.StatusCode, res.Header.Get("Content-Type"), body, nil
}

func firstExistingFile(paths []string) (string, []byte, bool) {
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err == nil {
			return p, data, true
		}
	}
	return "", nil, false
}

// extractHTMLTitle returns the first HTML title text (trimmed), case-insensitive tag.
func extractHTMLTitle(htmlBody string) string {
	re := regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	m := re.FindStringSubmatch(htmlBody)
	if len(m) < 2 {
		return ""
	}
	return strings.TrimSpace(m[1])
}

// expectedSessionTitle builds the product title for a session id.
func expectedSessionTitle(sessionID string) string {
	return sessionID + TitleSuffix
}

// Silence unused import when only helpers reference net/http in some packages.
var _ = http.StatusOK
```
