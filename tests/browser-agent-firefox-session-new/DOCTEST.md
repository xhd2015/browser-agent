# browser-agent Firefox session new (P2)

Exercises **`session new --browser firefox`**: pure Firefox open argv, package
`SessionNew` with `Browser=firefox` + injectable `OpenFirefoxFn`, CLI
`--browser` flag (help + unknown error), and chrome-default smoke.

**Classic TDD — RED against current code** (no `BuildFirefoxOpenArgs`, no
`SessionNewConfig.Browser` / `OpenFirefoxFn`, no CLI `--browser` yet).

**P1 is GREEN** under `./tests/browser-agent-firefox-ext/` — do **not** rewrite
those sealed leaves. This tree is P2 only.

**No real Firefox. No real Chrome. No real network** beyond ephemeral local
daemon (`RunDaemon` on `127.0.0.1:0`). Open hooks record calls only.

### Out of scope

- managed-firefox / `--user-data-dir` / `--load-extension` for Firefox
- Full Firefox job runner / CDP / debugger
- browser-trace
- Vite multi-target popup
- Changing Chrome default when `--browser` is omitted

## Version

0.0.2

# DSN (Domain Specific Notion)

**Operator** creates a session and opens a browser. For Firefox they load the
temporary add-on manually via `about:debugging` (no managed profile).

### 1. Pure open argv

```text
BuildFirefoxOpenArgs(sessionURL) → []string  # argv without binary
  # always starts with -new-window
  # non-empty URL: contains sessionURL; MUST NOT contain --load-extension
  # or --user-data-dir
  # empty URL: [-new-window] only; still no load-extension / user-data-dir
```

Production `openFirefox(sessionURL)` is best-effort platform open; tests never
call it — they inject `OpenFirefoxFn`.

### 2. SessionNew package API

```text
SessionNewConfig {
  Browser string              // ""|"chrome" → chrome path; "firefox" → firefox
  NoOpenChrome bool           // skip open for any browser
  OpenChromeFn(sessionURL, extPath) error
  OpenFirefoxFn(sessionURL) error   // URL only (no extension path)
  Home string                 // WithHome isolation for ensure path
  NoWait bool                 // skip extension wait (tests always set true)
  …
}
```

When `Browser == "firefox"`:

1. Ensure **Firefox** extension via `EnsureCanonicalFirefoxExtensionWithHome(Home)`
   (path under `…/extensions/browser-agent-firefox/{ver}/`) — **not** Chrome
   managed-chrome / `EnsureCanonicalExtension`.
2. If not `NoOpenChrome`: call `OpenFirefoxFn(sessionURL)` (or inject / default
   `openFirefox`). **Never** call `OpenChromeFn` for the firefox path.
3. Stdout (formatSessionNewOutput firefox variant) includes:
   - session-id, Session URL, Control (as today)
   - browser: firefox (marker) and/or firefox-specific install guidance
   - extension path containing `browser-agent-firefox`
   - `about:debugging` / Load Temporary Add-on steps
   - **not** chrome://extensions as primary install path

When `Browser` empty or `"chrome"`: existing chrome ensure + `OpenChromeFn`
behavior (no regression).

Invalid browser (e.g. `"safari"`) → error (CLI: stderr + nonzero).

### 3. CLI

```text
browser-agent session new --browser firefox [flags]
  --browser chrome|firefox     default chrome (empty / omitted)
  --no-open-chrome             skip open for any browser; still print path
  --no-wait                    unchanged
```

- `HandleCLI session new --browser firefox` with inject `OpenFirefoxFn`
- Help documents `--browser`
- Unknown browser → error, nonzero exit

**Test Client** uses package APIs + `HandleCLI` + inject hooks; never launches
a real browser.

## Decision Tree

```
browser-agent-firefox-session-new
├── firefox-open-args/                         [BuildFirefoxOpenArgs pure]
│   ├── with-url/                                non-empty URL in argv; no managed flags
│   └── empty-url/                               empty URL; still no load-ext / user-data-dir
├── session-new/                               [SessionNew package API]
│   ├── browser-firefox/                         Browser=firefox
│   │   ├── open-records-url/                      OpenFirefoxFn once with sessionURL
│   │   └── no-open-still-prints-path/             NoOpenChrome: no open; still path
│   └── browser-default-chrome/                  Browser empty → OpenChromeFn only
└── cli/                                       [HandleCLI]
    ├── session-new-browser-firefox/             --browser firefox + inject open
    ├── help-mentions-browser/                   help documents --browser
    └── unknown-browser/                         unknown browser → error nonzero
```

### Parameter significance (high → low)

1. **Surface / Mode** — pure args vs SessionNew package API vs CLI.
2. **Browser** (within session-new) — firefox vs default/chrome.
3. **Open policy** (within browser-firefox) — open vs `--no-open-chrome`.
4. **URL emptiness** (within firefox-open-args) — with-url vs empty-url.
5. **CLI op** — session-new-firefox vs help vs unknown-browser.

## Test Index

| Leaf | Scenario |
|------|----------|
| `firefox-open-args/with-url` | `BuildFirefoxOpenArgs(url)` includes URL; no `--load-extension` / `--user-data-dir` |
| `firefox-open-args/empty-url` | Empty URL → still no load-extension / user-data-dir |
| `session-new/browser-firefox/open-records-url` | `Browser=firefox`: OpenFirefoxFn once with sessionURL; no OpenChromeFn; stdout has firefox path + about:debugging |
| `session-new/browser-firefox/no-open-still-prints-path` | `NoOpenChrome`: OpenFirefoxFn not called; stdout still has path + about:debugging |
| `session-new/browser-default-chrome` | Browser empty: OpenChromeFn once; OpenFirefoxFn never; stdout not firefox-primary |
| `cli/session-new-browser-firefox` | `HandleCLI session new --browser firefox`: OpenFirefoxFn + path + about:debugging |
| `cli/help-mentions-browser` | Help documents `--browser` (chrome\|firefox) |
| `cli/unknown-browser` | `--browser safari` → error, nonzero |

**Leaf count: 8**

## How to Run

```sh
doctest vet ./tests/browser-agent-firefox-session-new
doctest test ./tests/browser-agent-firefox-session-new
# expect RED until implementer lands Browser / OpenFirefoxFn / BuildFirefoxOpenArgs / CLI --browser
```

Sibling trees that must stay GREEN after implement (orchestrator verify):

```sh
doctest test ./tests/browser-agent-firefox-ext
doctest test ./tests/browser-agent-daemon-phase8
doctest test ./tests/browser-agent-open-chrome
```

Module: `github.com/xhd2015/browser-agent`. Package under test: `browseragent`
(+ `browseragent/inject` for CLI `WithSessionNewHooks`).

### Implementer contract (authoritative for GREEN)

```text
// Pure argv (no binary):
func BuildFirefoxOpenArgs(sessionURL string) []string
// non-empty URL: argv contains sessionURL
// MUST NOT contain --load-extension or --user-data-dir (any form)

// SessionNewConfig extensions:
Browser string              // "" or "chrome" → chrome path; "firefox" → firefox path
OpenFirefoxFn func(sessionURL string) error
// NoOpenChrome skips open for chrome and firefox
// Home: EnsureCanonicalFirefoxExtensionWithHome when Browser=firefox

// When Browser == "firefox":
//   1. EnsureCanonicalFirefoxExtensionWithHome(Home) (or EnsureCanonicalFirefoxExtension)
//   2. if !NoOpenChrome: OpenFirefoxFn(sessionURL) | inject | openFirefox(sessionURL)
//   3. stdout includes browser-agent-firefox path + about:debugging / Load Temporary
//   4. MUST NOT call OpenChromeFn
// When Browser == "" or "chrome": existing chrome path unchanged
// Invalid Browser → error (CLI: message containing "unknown browser", nonzero)

// inject.SessionNewHooks:
OpenFirefoxFn func(sessionURL string) error
// SessionNewOpenFirefoxFn() snapshot (parallel-safe via WithSessionNewHooks)

// CLI:
//   session new --browser chrome|firefox   (default chrome when omitted)
//   --no-open-chrome still skips open; path still printed
//   fullHelp / session new --help documents --browser
//   unknown browser → Error: unknown browser "…" (or equivalent), nonzero
```

```go
import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xhd2015/browser-agent/browseragent"
	inj "github.com/xhd2015/browser-agent/browseragent/inject"
)

// Mode — top-level API surface.
const (
	ModeFirefoxOpenArgs = "firefox-open-args"
	ModeSessionNew      = "session-new"
	ModeCLI             = "cli"
)

// FirefoxOpenArgsOp — BuildFirefoxOpenArgs probes.
const (
	FirefoxOpenArgsOpWithURL  = "with-url"
	FirefoxOpenArgsOpEmptyURL = "empty-url"
)

// SessionNewBrowser — Browser field under session-new.
const (
	SessionNewBrowserFirefox       = "firefox"
	SessionNewBrowserDefaultChrome = "default-chrome"
)

// SessionNewFirefoxOp — open policy within browser-firefox.
const (
	SessionNewFirefoxOpOpenRecordsURL       = "open-records-url"
	SessionNewFirefoxOpNoOpenStillPrintsPath = "no-open-still-prints-path"
)

// CLIOp — HandleCLI probes.
const (
	CLIOpSessionNewBrowserFirefox = "session-new-browser-firefox"
	CLIOpHelpMentionsBrowser      = "help-mentions-browser"
	CLIOpUnknownBrowser           = "unknown-browser"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string
	BaseDir    string
	TestHome   string // isolated HOME for firefox ensure (WithHome)
	Addr       string
	SessionID  string

	FirefoxOpenArgsOp string
	SessionNewBrowser string
	SessionNewFirefoxOp string
	CLIOp             string

	// Session URL for pure BuildFirefoxOpenArgs.
	URL string

	// SessionNew / CLI browser selection.
	Browser      string // passed to SessionNewConfig.Browser or --browser
	NoOpenChrome bool
	NoWait       bool

	// Unknown browser value for CLI error leaf.
	UnknownBrowser string

	ReadyTimeout time.Duration
}

// Response holds outcomes for all modes.
type Response struct {
	// firefox-open-args
	FirefoxArgs []string

	// session-new / cli open recording
	OpenFirefoxCallCount int
	OpenFirefoxURL       string
	OpenChromeCallCount  int
	OpenChromeSessionURL string
	OpenChromeExtPath    string

	// shared IO
	Stdout        string
	Stderr        string
	SessionNewErr string
	CLIErr        string
	ExitCode      int

	SessionID       string
	SessionOnServer bool
	ServerSessionIDs []string
	BaseURL         string
	Addr            string
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
	if !req.NoWait {
		// All session-new / CLI leaves that create sessions skip extension wait.
		req.NoWait = true
	}

	switch req.Mode {
	case ModeFirefoxOpenArgs:
		return runFirefoxOpenArgsMode(t, req)
	case ModeSessionNew:
		return runSessionNewMode(t, req)
	case ModeCLI:
		return runCLIMode(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runFirefoxOpenArgsMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.FirefoxOpenArgsOp == "" {
		t.Fatal("FirefoxOpenArgsOp must be set")
	}
	url := req.URL
	switch req.FirefoxOpenArgsOp {
	case FirefoxOpenArgsOpWithURL:
		if url == "" {
			url = "http://127.0.0.1:43761/go?session=sess-ff-args-1"
		}
	case FirefoxOpenArgsOpEmptyURL:
		url = ""
	default:
		return nil, fmt.Errorf("unknown FirefoxOpenArgsOp %q", req.FirefoxOpenArgsOp)
	}
	args := browseragent.BuildFirefoxOpenArgs(url)
	return &Response{FirefoxArgs: args}, nil
}

func runSessionNewMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.SessionNewBrowser == "" {
		t.Fatal("SessionNewBrowser must be set")
	}
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set")
	}
	if req.TestHome == "" {
		t.Fatal("TestHome must be set")
	}

	var (
		ffCount, chromeCount int
		ffURL, chromeURL, chromeExt string
		mu sync.Mutex
	)
	recordFirefox := func(sessionURL string) error {
		mu.Lock()
		defer mu.Unlock()
		ffCount++
		ffURL = sessionURL
		return nil
	}
	recordChrome := func(sessionURL, extPath string) error {
		mu.Lock()
		defer mu.Unlock()
		chromeCount++
		chromeURL = sessionURL
		chromeExt = extPath
		return nil
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	addr := ln.Addr().String()
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
		_, err := browseragent.RunDaemon(ctx, cfg)
		done <- err
	}()
	t.Cleanup(cancel)

	baseURL := "http://" + addr
	if err := waitHealth(baseURL, req.ReadyTimeout); err != nil {
		return nil, err
	}

	browser := ""
	noOpen := req.NoOpenChrome
	switch req.SessionNewBrowser {
	case SessionNewBrowserFirefox:
		browser = "firefox"
		if req.SessionNewFirefoxOp == SessionNewFirefoxOpNoOpenStillPrintsPath {
			noOpen = true
		}
	case SessionNewBrowserDefaultChrome:
		browser = "" // default chrome
	default:
		return nil, fmt.Errorf("unknown SessionNewBrowser %q", req.SessionNewBrowser)
	}

	var stdout, stderr bytes.Buffer
	snCfg := browseragent.SessionNewConfig{
		BaseDir:      req.BaseDir,
		Addr:         addr,
		SessionID:    req.SessionID,
		Browser:      browser,
		NoOpenChrome: noOpen,
		NoWait:       true,
		Home:         req.TestHome,
		OpenChromeFn: recordChrome,
		OpenFirefoxFn: recordFirefox,
		Stdout:       &stdout,
		Stderr:       &stderr,
	}
	snErr := browseragent.SessionNew(snCfg)
	resp := &Response{
		Stdout:  stdout.String(),
		Stderr:  stderr.String(),
		BaseURL: baseURL,
		Addr:    addr,
	}
	if snErr != nil {
		resp.SessionNewErr = snErr.Error()
	}
	mu.Lock()
	resp.OpenFirefoxCallCount = ffCount
	resp.OpenFirefoxURL = ffURL
	resp.OpenChromeCallCount = chromeCount
	resp.OpenChromeSessionURL = chromeURL
	resp.OpenChromeExtPath = chromeExt
	mu.Unlock()

	if snErr != nil {
		return resp, snErr
	}
	sid := req.SessionID
	if sid == "" {
		sid = extractSessionIDFromStdout(resp.Stdout)
	}
	resp.SessionID = sid
	return resp, nil
}

func runCLIMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.CLIOp == "" {
		t.Fatal("CLIOp must be set")
	}

	switch req.CLIOp {
	case CLIOpHelpMentionsBrowser:
		return runCLIHelp(t, req)
	case CLIOpUnknownBrowser:
		return runCLIUnknownBrowser(t, req)
	case CLIOpSessionNewBrowserFirefox:
		return runCLISessionNewFirefox(t, req)
	default:
		return nil, fmt.Errorf("unknown CLIOp %q", req.CLIOp)
	}
}

func runCLIHelp(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	// session new --help returns fullHelp (same as other session new help leaves).
	cliErr := browseragent.HandleCLI(
		[]string{"session", "new", "--help"},
		map[string]string{},
		&stdout,
		&stderr,
	)
	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if cliErr != nil {
		resp.CLIErr = cliErr.Error()
		resp.ExitCode = 1
	}
	// Help may return nil even when printing help.
	return resp, cliErr
}

func runCLIUnknownBrowser(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	browser := req.UnknownBrowser
	if browser == "" {
		browser = "safari"
	}
	var stdout, stderr bytes.Buffer
	args := []string{
		"session", "new",
		"--browser", browser,
		"--base-dir", req.BaseDir,
		"--no-wait",
		"--no-open-chrome",
	}
	cliErr := browseragent.HandleCLI(args, map[string]string{}, &stdout, &stderr)
	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if cliErr != nil {
		resp.CLIErr = cliErr.Error()
		resp.ExitCode = 1
	} else {
		resp.ExitCode = 0
	}
	// Expected error path: return nil transport error so Assert checks CLIErr/ExitCode.
	return resp, nil
}

func runCLISessionNewFirefox(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set")
	}
	if req.TestHome == "" {
		t.Fatal("TestHome must be set")
	}

	var (
		ffCount int
		ffURL   string
		mu      sync.Mutex
	)
	hooks := &inj.SessionNewHooks{
		OpenFirefoxFn: func(sessionURL string) error {
			mu.Lock()
			defer mu.Unlock()
			ffCount++
			ffURL = sessionURL
			return nil
		},
		OpenChromeFn: func(sessionURL, extPath string) error {
			mu.Lock()
			defer mu.Unlock()
			// Chrome open must not run for --browser firefox.
			return fmt.Errorf("OpenChromeFn must not be called for browser=firefox")
		},
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	addr := ln.Addr().String()
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
		_, err := browseragent.RunDaemon(ctx, cfg)
		done <- err
	}()
	t.Cleanup(cancel)

	baseURL := "http://" + addr
	if err := waitHealth(baseURL, req.ReadyTimeout); err != nil {
		return nil, err
	}

	host, portStr, _ := net.SplitHostPort(addr)
	args := []string{
		"session", "new",
		"--browser", "firefox",
		"--base-dir", req.BaseDir,
		"--host", host,
		"--server-port", portStr,
		"--no-wait",
	}
	if req.SessionID != "" {
		args = append(args, "--session-id", req.SessionID)
	}

	var stdout, stderr bytes.Buffer
	// Isolate HOME so EnsureCanonicalFirefoxExtension lands under TestHome when
	// CLI path uses process home (package API uses cfg.Home; CLI may set env).
	env := map[string]string{"HOME": req.TestHome}
	cliErr := inj.WithSessionNewHooks(hooks, func() error {
		return browseragent.HandleCLI(args, env, &stdout, &stderr)
	})
	resp := &Response{
		Stdout:  stdout.String(),
		Stderr:  stderr.String(),
		BaseURL: baseURL,
		Addr:    addr,
	}
	if cliErr != nil {
		resp.CLIErr = cliErr.Error()
		resp.ExitCode = 1
	}
	mu.Lock()
	resp.OpenFirefoxCallCount = ffCount
	resp.OpenFirefoxURL = ffURL
	mu.Unlock()

	if cliErr != nil {
		return resp, cliErr
	}
	sid := req.SessionID
	if sid == "" {
		sid = extractSessionIDFromStdout(resp.Stdout)
	}
	resp.SessionID = sid
	return resp, nil
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
	// Prefer "session-id: <id>" line from formatSessionNewOutput.
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "session-id:") {
			return strings.TrimSpace(line[len("session-id:"):])
		}
	}
	// Fallback: first sess- token.
	for _, tok := range strings.Fields(stdout) {
		if strings.HasPrefix(tok, "sess-") {
			return strings.Trim(tok, `"'`)
		}
	}
	return ""
}
```
