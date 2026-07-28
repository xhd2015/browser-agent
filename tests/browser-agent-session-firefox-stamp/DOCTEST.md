# browser-agent session firefox stamp (Phase 2)

Classic-TDD tree for **stamping Firefox on session create**: when the operator
creates a session with `Browser=firefox` (`session new --browser firefox`),
the durable session identity and install UX must record Firefox — not the
default Chrome extension tree.

| Surface | What is under test |
|---------|-------------------|
| `SessionNew` + `meta.json` | `browser: "firefox"` + `extension_install_path` under `browser-agent-firefox` |
| `GET /v1/session` | Live snapshot reflects the same Firefox install path |
| Session page boot | `injectSessionBoot` / boot JSON uses `browser: "firefox"` when snap path indicates firefox |
| `session info` Next steps | Disconnected hints use Firefox install (`install-firefox-extension` / about:debugging), not Chrome Load unpacked |

**Depends on (already GREEN):** P1 firefox ext extract
(`EnsureCanonicalFirefoxExtensionWithHome`), P2 session-new open/stdout
(`SessionNewConfig.Browser`, `OpenFirefoxFn`). This tree is **Phase 2 stamp only**
— do not re-test open argv or about:debugging stdout recipes as primary.

**Classic TDD — RED against current code.** Today `SessionRegistry.Create`
always calls `EnsureCanonicalExtension()` (Chrome) and writes meta without a
firefox `browser` field. `SessionNew(Browser=firefox)` prints the firefox path
on stdout but does **not** stamp meta / live snap for inject or session info.

**No real Firefox. No real Chrome. No real network** beyond ephemeral local
daemon (`RunDaemon` on `127.0.0.1:0`). Prefer package API + temp `HOME` /
`SessionNewConfig.Home`. Injectable open hooks only (NoOpenChrome for stamp leaves).

### Plan phase

| Phase | Scope |
|-------|--------|
| **1** | Client install detect (session-page-browser-detect) — sealed separately |
| **2 (this)** | Server Create / SessionNew stamps firefox extension path + browser in meta/boot/info |

### Out of scope

- Connect/register fix for Firefox content script
- Re-testing `BuildFirefoxOpenArgs` / OpenFirefoxFn call counts (firefox-session-new)
- Managed-firefox profile / `--load-extension`

## Version

0.0.2

# DSN (Domain Specific Notion)

**Operator** runs session create with browser firefox. The control plane must
**stamp** the session so every downstream surface (disk meta, snapshot, session
page boot, session info) agrees the install tree is Firefox.

```text
# Operator path (preferred test surface)
SessionNew(Browser=firefox, Home=TestHome, NoOpenChrome, NoWait)
  -> ensure firefox extension under TestHome
     …/extensions/browser-agent-firefox/{version}/
  -> create session (POST /v1/sessions and/or Create with browser)
  -> durable stamp:
       meta.json.browser == "firefox"   (or equivalent field)
       meta.json.extension_install_path contains browser-agent-firefox
       meta path MUST NOT be managed-chrome / Chrome-only tree

# Live snapshot
GET /v1/session?session=<id>
  -> extension_install_path same firefox segment
  -> optional browsers/browser field includes firefox when stamped

# Boot inject (session page)
GET /go?session=<id>  (or injectSessionBoot)
  -> #browser-agent-boot / FormatSessionBootJSONWithBrowser
     JSON "browser":"firefox" when snap path/browsers indicate firefox
  -> window.__BROWSER_AGENT.browser == "firefox"

# Session info (disconnected)
session info / FormatSessionInfo(snap)
  -> Next steps: install-firefox-extension and/or about:debugging
  -> Load path shows browser-agent-firefox
  -> MUST NOT primary-recommend install-chrome-extension / chrome://extensions
```

**Chrome regression:** `SessionNew` with default/empty browser still stamps the
Chrome extension segment (not forced to firefox).

**Implementation freedom** (any one is fine if outcomes hold):

- Pass browser into `Create` / `CreateSessionResult`
- `SessionNew` updates meta after Create
- Registry `Create` accepts optional Browser field
- POST `/v1/sessions` body may include `"browser":"firefox"`

**Test Client** uses package APIs + temp dirs; never launches a real browser.

## Decision Tree

```
browser-agent-session-firefox-stamp
├── session-new/                               [SessionNew package API stamps]
│   ├── browser-firefox/                         Browser=firefox
│   │   ├── meta-browser-and-path/                 meta.json browser + firefox path
│   │   └── v1-session-snapshot/                   GET /v1/session firefox path
│   └── browser-chrome/                          Browser empty (default chrome)
│       └── meta-not-firefox-path/                 chrome path; not firefox-forced
├── boot-inject/                               [session page boot after stamp]
│   └── go-page-boot-firefox/                    /go boot JSON browser=firefox
└── session-info/                              [human next steps]
    └── firefox-disconnected-install-hint/       install-firefox + path; not chrome primary
```

### Parameter significance (high → low)

1. **Surface** — durable stamp (session-new) vs boot inject vs session-info
   (largest behavioral fan-out).
2. **Browser** (session-new only) — firefox stamp vs chrome regression.
3. **Observation channel** (firefox stamp) — on-disk `meta.json` vs live
   `GET /v1/session`.
4. **Downstream consumer** — go-page boot vs session info text.

## Test Index

| Leaf | Scenario |
|------|----------|
| `session-new/browser-firefox/meta-browser-and-path` | After `SessionNew(Browser=firefox, Home=TestHome, NoOpen, NoWait)`: `meta.json` has `browser` firefox (or equivalent) and `extension_install_path` containing `browser-agent-firefox` (not managed-chrome) |
| `session-new/browser-firefox/v1-session-snapshot` | Same create: `GET /v1/session` `extension_install_path` contains `browser-agent-firefox` (absolute path preferred) |
| `session-new/browser-chrome/meta-not-firefox-path` | Default chrome create: `extension_install_path` is Chrome segment (contains `browser-agent` install tree, not `browser-agent-firefox`) |
| `boot-inject/go-page-boot-firefox` | After firefox SessionNew: `GET /go?session=` HTML boot JSON / `__BROWSER_AGENT` has `"browser":"firefox"` |
| `session-info/firefox-disconnected-install-hint` | After firefox SessionNew (disconnected): session info Next steps mention firefox install (`install-firefox-extension` and/or `about:debugging`) + firefox path; not primary chrome://extensions / install-chrome-extension |

**Leaf count: 5**

## How to Run

```sh
doctest vet ./tests/browser-agent-session-firefox-stamp
doctest test ./tests/browser-agent-session-firefox-stamp
# expect RED until Create/SessionNew stamps firefox path + browser;
# session info firefox next steps may remain RED until FormatSessionInfo branch
```

Sibling trees (orchestrator verify after implement; do not rewrite):

```sh
doctest test ./tests/browser-agent-firefox-session-new
doctest test ./tests/browser-agent-firefox-ext
doctest test ./tests/browser-agent-session-page-browser-detect
```

Module: `github.com/xhd2015/browser-agent`. Package under test: `browseragent`.

### Implementer contract (authoritative for GREEN)

```text
// After SessionNew with Browser=firefox (Home optional isolation):
//   1. sessions/<id>/meta.json:
//        "browser": "firefox"   (field name browser preferred; browsers:["firefox"] also ok if documented)
//        "extension_install_path": path containing "browser-agent-firefox"
//        path MUST NOT be under managed-chrome Chrome-only tree as the install path
//   2. In-memory / GET /v1/session snapshot.extension_install_path matches firefox segment
//   3. injectSessionBoot / FormatSessionBootJSONWithBrowser path:
//        when snap indicates firefox (path segment or browser field / browsers list),
//        boot JSON and __BROWSER_AGENT.browser == "firefox"
//   4. FormatSessionInfo disconnected Next steps for firefox snaps:
//        install-firefox-extension and/or about:debugging + load path
//        MUST NOT primary-recommend install-chrome-extension / chrome://extensions

// Chrome default SessionNew (Browser ""|"chrome"):
//   extension_install_path remains Chrome browser-agent tree (not browser-agent-firefox)

// Implementation may:
//   - registry.Create(id, opts) with Browser
//   - POST /v1/sessions {"session_id","browser"}
//   - SessionNew patch meta + setExtensionInstallPath after Create
// Any approach is fine if outcomes above hold.
```

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/browser-agent/browseragent"
)

// Mode — top-level surface.
const (
	ModeSessionNew  = "session-new"
	ModeBootInject  = "boot-inject"
	ModeSessionInfo = "session-info"
)

// SessionNewBrowser — browser selection under session-new.
const (
	SessionNewBrowserFirefox = "firefox"
	SessionNewBrowserChrome  = "chrome"
)

// SessionNewFirefoxOp — observation channel under browser-firefox.
const (
	SessionNewFirefoxOpMetaBrowserAndPath = "meta-browser-and-path"
	SessionNewFirefoxOpV1SessionSnapshot  = "v1-session-snapshot"
)

// SessionNewChromeOp — chrome regression.
const (
	SessionNewChromeOpMetaNotFirefoxPath = "meta-not-firefox-path"
)

// BootInjectOp — boot inject leaves.
const (
	BootInjectOpGoPageBootFirefox = "go-page-boot-firefox"
)

// SessionInfoOp — session info leaves.
const (
	SessionInfoOpFirefoxDisconnectedInstallHint = "firefox-disconnected-install-hint"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string
	BaseDir    string
	TestHome   string // isolated HOME for firefox ensure (SessionNewConfig.Home)
	Addr       string
	SessionID  string

	SessionNewBrowser   string
	SessionNewFirefoxOp string
	SessionNewChromeOp  string
	BootInjectOp        string
	SessionInfoOp       string

	// Browser passed to SessionNewConfig.Browser ("" = chrome default).
	Browser string

	// Whether Run should also fetch /v1/session and/or /go after SessionNew.
	FetchV1Session bool
	FetchGoPage    bool
	FetchSessionInfo bool

	ReadyTimeout time.Duration
}

// Response holds outcomes for all modes.
type Response struct {
	SessionNewErr string
	Stdout        string
	Stderr        string
	SessionID     string
	BaseURL       string
	Addr          string

	// On-disk meta.json after create.
	MetaPath               string
	MetaJSON               string
	MetaBrowser            string
	MetaExtensionInstallPath string

	// GET /v1/session
	V1Status                   int
	V1Body                     string
	V1ExtensionInstallPath     string
	V1BrowserField             string

	// GET /go?session=
	GoStatus int
	GoHTML   string
	GoBootBrowser string

	// session info human stdout
	SessionInfoStdout string
	SessionInfoErr    string
}

// Run executes the scenario selected by req.Mode and leaf Setup narrowing.
func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	if req == nil {
		t.Fatal("req is nil")
	}
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	if req.ModuleRoot == "" {
		req.ModuleRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", ".."))
	}
	if req.ReadyTimeout <= 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	if req.SessionID == "" {
		req.SessionID = "sess-ff-stamp-1"
	}
	if req.BaseDir == "" || req.TestHome == "" {
		t.Fatal("BaseDir and TestHome must be set by root Setup")
	}

	switch req.Mode {
	case ModeSessionNew:
		return runSessionNewStamp(t, req)
	case ModeBootInject:
		return runBootInject(t, req)
	case ModeSessionInfo:
		return runSessionInfo(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runSessionNewStamp(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	browser := req.Browser
	switch req.SessionNewBrowser {
	case SessionNewBrowserFirefox:
		browser = "firefox"
	case SessionNewBrowserChrome:
		browser = ""
	default:
		if req.SessionNewBrowser != "" {
			return nil, fmt.Errorf("unknown SessionNewBrowser %q", req.SessionNewBrowser)
		}
	}

	baseURL, addr, cleanup, err := startEphemeralDaemon(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp, err := doSessionNew(t, req, baseURL, addr, browser)
	if err != nil {
		return resp, err
	}
	if err := loadMetaIntoResponse(t, req, resp); err != nil {
		return resp, err
	}
	if req.FetchV1Session || req.SessionNewFirefoxOp == SessionNewFirefoxOpV1SessionSnapshot {
		if err := fetchV1Session(t, resp); err != nil {
			return resp, err
		}
	}
	return resp, nil
}

func runBootInject(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.BootInjectOp != BootInjectOpGoPageBootFirefox {
		return nil, fmt.Errorf("unknown BootInjectOp %q", req.BootInjectOp)
	}
	baseURL, addr, cleanup, err := startEphemeralDaemon(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp, err := doSessionNew(t, req, baseURL, addr, "firefox")
	if err != nil {
		return resp, err
	}
	if err := loadMetaIntoResponse(t, req, resp); err != nil {
		return resp, err
	}
	if err := fetchGoPage(t, resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func runSessionInfo(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.SessionInfoOp != SessionInfoOpFirefoxDisconnectedInstallHint {
		return nil, fmt.Errorf("unknown SessionInfoOp %q", req.SessionInfoOp)
	}
	baseURL, addr, cleanup, err := startEphemeralDaemon(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp, err := doSessionNew(t, req, baseURL, addr, "firefox")
	if err != nil {
		return resp, err
	}
	if err := loadMetaIntoResponse(t, req, resp); err != nil {
		return resp, err
	}

	// Package-adjacent: CLI session info against the same daemon (human FormatSessionInfo).
	host, portStr, _ := net.SplitHostPort(addr)
	var stdout, stderr bytes.Buffer
	args := []string{
		"session", "info",
		"--session-id", resp.SessionID,
		"--base-dir", req.BaseDir,
		"--host", host,
		"--server-port", portStr,
	}
	cliErr := browseragent.HandleCLI(args, map[string]string{"HOME": req.TestHome}, &stdout, &stderr)
	resp.SessionInfoStdout = stdout.String()
	resp.Stderr += stderr.String()
	if cliErr != nil {
		resp.SessionInfoErr = cliErr.Error()
	}
	return resp, nil
}

func startEphemeralDaemon(t *testing.T, req *Request) (baseURL, addr string, cleanup func(), err error) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", "", nil, err
	}
	addr = ln.Addr().String()
	_ = ln.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cfg := browseragent.DaemonConfig{
		Addr:    addr,
		BaseDir: req.BaseDir,
		Stdout:  io.Discard,
		Stderr:  io.Discard,
	}
	done := make(chan error, 1)
	go func() {
		_, e := browseragent.RunDaemon(ctx, cfg)
		done <- e
	}()
	cleanup = func() {
		cancel()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
		}
	}
	baseURL = "http://" + addr
	if err := waitHealth(baseURL, req.ReadyTimeout); err != nil {
		cleanup()
		return "", "", nil, err
	}
	return baseURL, addr, cleanup, nil
}

func doSessionNew(t *testing.T, req *Request, baseURL, addr, browser string) (*Response, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	snCfg := browseragent.SessionNewConfig{
		BaseDir:      req.BaseDir,
		Addr:         addr,
		SessionID:    req.SessionID,
		Browser:      browser,
		NoOpenChrome: true, // stamp leaves never open a browser
		NoWait:       true,
		Home:         req.TestHome,
		OpenChromeFn: func(sessionURL, extPath string) error {
			return fmt.Errorf("OpenChromeFn must not be called in stamp leaves")
		},
		OpenFirefoxFn: func(sessionURL string) error {
			return fmt.Errorf("OpenFirefoxFn must not be called when NoOpenChrome")
		},
		Stdout: &stdout,
		Stderr: &stderr,
	}
	snErr := browseragent.SessionNew(snCfg)
	resp := &Response{
		Stdout:    stdout.String(),
		Stderr:    stderr.String(),
		BaseURL:   baseURL,
		Addr:      addr,
		SessionID: req.SessionID,
	}
	if snErr != nil {
		resp.SessionNewErr = snErr.Error()
		return resp, snErr
	}
	if resp.SessionID == "" {
		resp.SessionID = extractSessionIDFromStdout(resp.Stdout)
	}
	return resp, nil
}

func loadMetaIntoResponse(t *testing.T, req *Request, resp *Response) error {
	t.Helper()
	sid := resp.SessionID
	if sid == "" {
		sid = req.SessionID
	}
	metaPath := filepath.Join(req.BaseDir, "sessions", sid, "meta.json")
	resp.MetaPath = metaPath
	raw, err := os.ReadFile(metaPath)
	if err != nil {
		// Leave empty fields; Assert fails with clear message.
		return nil
	}
	resp.MetaJSON = string(raw)
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	resp.MetaBrowser = firstStringField(m, "browser", "Browser")
	// browsers: ["firefox"] accepted as equivalent stamp signal when browser missing.
	if resp.MetaBrowser == "" {
		if arr, ok := m["browsers"].([]any); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok && strings.EqualFold(strings.TrimSpace(s), "firefox") {
					resp.MetaBrowser = "firefox"
					break
				}
			}
		}
	}
	resp.MetaExtensionInstallPath = firstStringField(m, "extension_install_path", "extensionInstallPath")
	return nil
}

func fetchV1Session(t *testing.T, resp *Response) error {
	t.Helper()
	u := strings.TrimRight(resp.BaseURL, "/") + "/v1/session?session=" + resp.SessionID
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	resp.V1Status = res.StatusCode
	resp.V1Body = string(body)
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return nil
	}
	resp.V1ExtensionInstallPath = firstStringField(m, "extension_install_path", "extensionInstallPath")
	if resp.V1ExtensionInstallPath == "" {
		if be, ok := m["bundled_extension"].(map[string]any); ok {
			resp.V1ExtensionInstallPath = firstStringField(be, "path", "Path")
		}
	}
	resp.V1BrowserField = firstStringField(m, "browser", "Browser")
	if resp.V1BrowserField == "" {
		if arr, ok := m["browsers"].([]any); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok && strings.EqualFold(strings.TrimSpace(s), "firefox") {
					resp.V1BrowserField = "firefox"
					break
				}
			}
		}
	}
	return nil
}

func fetchGoPage(t *testing.T, resp *Response) error {
	t.Helper()
	u := strings.TrimRight(resp.BaseURL, "/") + "/go?session=" + resp.SessionID
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	resp.GoStatus = res.StatusCode
	resp.GoHTML = string(body)
	resp.GoBootBrowser = extractBootBrowser(resp.GoHTML)
	return nil
}

func extractBootBrowser(html string) string {
	// Prefer application/json boot script body.
	const marker = `id="browser-agent-boot"`
	idx := strings.Index(html, marker)
	if idx >= 0 {
		// Find following > ... </script>
		gt := strings.Index(html[idx:], ">")
		if gt >= 0 {
			start := idx + gt + 1
			end := strings.Index(html[start:], "</script>")
			if end >= 0 {
				raw := strings.TrimSpace(html[start : start+end])
				var m map[string]any
				if json.Unmarshal([]byte(raw), &m) == nil {
					if b, ok := m["browser"].(string); ok {
						return strings.ToLower(strings.TrimSpace(b))
					}
				}
			}
		}
	}
	// Fallback: window.__BROWSER_AGENT = { ... browser: "firefox" }
	if i := strings.Index(html, "__BROWSER_AGENT"); i >= 0 {
		chunk := html[i:]
		// Prefer quoted JSON-like "browser":"firefox" inside the object chunk.
		if j := strings.Index(chunk, `"browser"`); j >= 0 {
			rest := chunk[j+len(`"browser"`):]
			colon := strings.Index(rest, ":")
			if colon >= 0 {
				rest = strings.TrimSpace(rest[colon+1:])
				for _, q := range []string{`"`, `'`} {
					if strings.HasPrefix(rest, q) {
						rest2 := rest[1:]
						if e := strings.Index(rest2, q); e >= 0 {
							return strings.ToLower(strings.TrimSpace(rest2[:e]))
						}
					}
				}
			}
		}
		// JS style browser: "firefox"
		lowChunk := strings.ToLower(chunk)
		if k := strings.Index(lowChunk, "browser"); k >= 0 {
			sub := chunk[k:]
			colon := strings.Index(sub, ":")
			if colon >= 0 {
				rest := strings.TrimSpace(sub[colon+1:])
				for _, q := range []string{`"`, `'`} {
					if strings.HasPrefix(rest, q) {
						rest2 := rest[1:]
						if e := strings.Index(rest2, q); e >= 0 {
							return strings.ToLower(strings.TrimSpace(rest2[:e]))
						}
					}
				}
			}
		}
	}
	// Last resort: FormatSessionBootJSON-like substring "browser":"firefox"
	if strings.Contains(html, `"browser":"firefox"`) || strings.Contains(html, `"browser": "firefox"`) {
		return "firefox"
	}
	if strings.Contains(html, `"browser":"chrome"`) || strings.Contains(html, `"browser": "chrome"`) {
		return "chrome"
	}
	return ""
}

func waitHealth(baseURL string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}
	for time.Now().Before(deadline) {
		resp, err := client.Get(strings.TrimRight(baseURL, "/") + "/v1/health")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("health check timed out for %s", baseURL)
}

func extractSessionIDFromStdout(stdout string) string {
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "session-id:") {
			return strings.TrimSpace(line[len("session-id:"):])
		}
	}
	for _, tok := range strings.Fields(stdout) {
		if strings.HasPrefix(tok, "sess-") {
			return strings.Trim(tok, `"'`)
		}
	}
	return ""
}

func firstStringField(m map[string]any, keys ...string) string {
	if m == nil {
		return ""
	}
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch t := v.(type) {
			case string:
				return strings.TrimSpace(t)
			}
		}
	}
	return ""
}
```
