# browser-trace install panel — always visible + expand/collapse

Exercises the **session-page install panel** on `GET /go?session=…`: the panel
must **always** be present in the HTML (never removed when the extension is
healthy), **auto-expanded** when the extension is not fully working, and
**auto-collapsed** when connected **and** `supports_browser_trace`.

Also covers the optional pure helper
`ShouldExpandInstallPanel(connected, supports bool) bool`
(`!(connected && supports)`).

This tree is **separate** from:

- `./tests/browser-trace/` — sealed capture lifecycle / HAR / deadlines
- `./tests/browser-trace-session-page/` — generic `/v1/session` + `/go` smoke
- `./tests/browser-trace-embed-extension/` — extract / install CLI / not-connected panel smoke

Those trees must stay green. This tree deepens always-visible + expand/collapse
without rewriting sealed leaves.

**No real Chrome**, no DOM automation. HTTP leaves drive `browsertrace.Run` with
`NoOpenChrome` and inspect the HTML string. Pure-helper leaves call the Go
function directly.

Product rules under test:

| State | Panel in DOM | Default expanded |
|-------|--------------|------------------|
| Not connected (no hello) | Yes | **Yes** |
| Connected, !supports | Yes | **Yes** |
| Connected + supports | Yes | **No** (collapsed; summary still visible) |

Recommended HTML shape: `<details id="browser-trace-install" … open?>` with
`data-browser-trace-install`, `data-default-open="true|false"`, path attributes,
and install body (path + `chrome://extensions` + embedded version). **Never**
`display:none` on the whole panel at serve time.

## Version

0.0.2

# DSN (Domain Specific Notion)

**User** opens the **Session Page** (`GET /go?session=<id>`) while a browser-trace
session is live. They need install/update steps available even after the
extension is healthy (to update / reload), without the panel disappearing.

**Control Server** keeps one in-memory **Session** and serves:

- `GET /go?session=…` — HTML dashboard that **always** includes the install
  panel (`#browser-trace-install` / `data-browser-trace-install`)
- `POST /v1/hello` — **Extension Agent** announces `version` + `features[]`
- `GET /v1/session` — poller JSON (connected / supports); not primary here

**Install Panel** (session-page UI):

- Always rendered in the HTML (even when install path is empty: still show
  steps + `chrome://extensions`)
- **Expanded** when `!(connected && supports_browser_trace)` at serve time
  (`open` and/or `data-default-open="true"`)
- **Collapsed** when `connected && supports_browser_trace`
  (`open` absent or `data-default-open="false"`; `<summary>` remains)
- Body keeps absolute path, `chrome://extensions` text, embedded version
- Client poll may re-sync expand/collapse unless the user toggled details
  (`data-user-toggled`); user-toggle freeze is client-only and **out of scope**
  for this HTTP/string tree (no browser DOM automation)
- **Must not** hide the whole panel with `display:none`

**Expand Policy** (optional pure helper for server + client parity):

```
ShouldExpandInstallPanel(connected, supports) = !(connected && supports)
```

**Test Client** starts the control server via `browsertrace.Run`
(`NoOpenChrome`, known `SessionSuffix`, temp `BaseDir`, free port), optionally
posts hello to stage connection/capability, then `GET /go` and asserts panel
presence + expand/collapse markers. Pure leaves call
`browsertrace.ShouldExpandInstallPanel` only.

## Decision Tree

```
browser-trace-install-panel
├── go-html/                                 [GET /go HTML install panel]
│   ├── not-connected/                         no hello → present + expanded + path/chrome
│   ├── hello-no-supports/                     hello OK, supports=false → present + expanded
│   └── hello-supports/                        hello OK, supports=true → present + collapsed
└── should-expand-helper/                    [pure ShouldExpandInstallPanel]
    ├── expand/                                expect true (not both connected+supports)
    │   ├── neither/                             connected=false, supports=false
    │   ├── supports-only/                       connected=false, supports=true
    │   └── connected-only/                      connected=true,  supports=false
    └── collapse/                              expect false (only when both true)
        └── both/                                connected=true, supports=true
```

### Parameter significance (high → low)

1. **Surface** — live `/go` HTML vs pure expand helper (different `Run` modes /
   contracts).
2. **Extension readiness** (HTML) — no hello vs hello without support vs
   connected+supports (drives expand vs collapse while panel always present).
3. **Truth-table cell** (pure helper) — the four `(connected, supports)` pairs,
   split by expected expand (true) vs collapse (false).

## Test Index

| Leaf | Scenario (requirement #) |
|------|--------------------------|
| `go-html/not-connected` | (#1, #5) No hello → panel present, expanded, path + `chrome://extensions` |
| `go-html/hello-no-supports` | (#3) Hello without supports → panel present, expanded |
| `go-html/hello-supports` | (#2) Hello + supports → panel **still present**, collapsed; markers remain |
| `should-expand-helper/expand/neither` | (#4) `(false,false)` → expand true |
| `should-expand-helper/expand/supports-only` | (#4) `(false,true)` → expand true |
| `should-expand-helper/expand/connected-only` | (#4) `(true,false)` → expand true |
| `should-expand-helper/collapse/both` | (#4) `(true,true)` → expand false |

## How to Run

```sh
cd tests/browser-trace-install-panel
doctest vet .
doctest test -v .
# or from repo root:
doctest vet ./tests/browser-trace-install-panel
doctest test ./tests/browser-trace-install-panel
# regression (shared /go + session surfaces must stay green):
doctest test ./tests/browser-trace-embed-extension
doctest test ./tests/browser-trace-session-page
```

Requires package `github.com/xhd2015/browser-agent/browsertrace`:

- Existing `Config` + `Run` HTTP session control server
- `/go` always injects install panel with expand/collapse markers (TDD red until
  product stops omitting the panel when connected+supports and stops
  `display:none` hide)
- Optional export: `ShouldExpandInstallPanel(connected, supports bool) bool`

### Expected `/go` install panel contract (implementer)

Prose contract (authoritative; harness inspects HTML text only):

- Panel root always present when session HTML is served, with at least one of:
  - `data-browser-trace-install`
  - `id="browser-trace-install"` / `id='browser-trace-install'`
- Prefer `<details>` with optional `<summary>`; body keeps install steps.
- Expand defaults from **current** session extension state at serve time:
  - expand when `!(helloOK && supports_browser_trace)` (connected ≈ successful hello)
  - collapse when connected **and** supports
- Markers for expand state (accept any strong signal):
  - expanded: `open` attribute on the panel `<details>` **and/or**
    `data-default-open="true"` (or `data-default-open=true`)
  - collapsed: no `open` attribute **and/or** `data-default-open="false"`
- Attributes for install path / version (when extract succeeded):
  `data-extension-path`, `data-install-path`, `data-embedded-version`
- Body text includes `chrome://extensions` (copyable text, not only a chrome:
  link) and path guidance when path is known.
- Server-rendered HTML must **not** set the panel to `display:none` / `display: none`.
- Client poll may set `details.open` from the same expand policy unless the user
  toggled; user-toggle is **not** asserted in this tree.

### Pure helper contract

```go
// ShouldExpandInstallPanel reports whether the install panel should be open.
// Expand unless the extension is connected and supports browser-trace.
func ShouldExpandInstallPanel(connected, supports bool) bool {
    return !(connected && supports)
}
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
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/browser-agent/browsertrace"
)

// Mode values — surface under test (set by grouping SETUP).
const (
	ModeGoHTML       = "go"
	ModeShouldExpand = "should-expand"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	// Mode selects which operation Run executes: ModeGoHTML or ModeShouldExpand.
	Mode string

	// BaseDir is session parent directory (temp per leaf). Required for ModeGoHTML.
	BaseDir string
	// Addr is host:port for Control Server. Empty → free loopback port in Run.
	Addr string
	// ReadyTimeout / CompleteTimeout override product defaults for fast tests.
	ReadyTimeout    time.Duration
	CompleteTimeout time.Duration
	// NoOpenChrome skips launching Chrome (always true for HTTP leaves).
	NoOpenChrome bool
	// SessionSuffix is the known session id (browsertrace uses suffix as id).
	SessionSuffix string

	// DoHello, when true, POST /v1/hello before the /go probe.
	DoHello bool
	// HelloVersion is the version field on hello (e.g. "1.2.0").
	HelloVersion string
	// HelloFeatures is the features array on hello.
	HelloFeatures []string

	// Pure helper inputs (ModeShouldExpand).
	Connected bool
	Supports  bool
	// WantExpand is the expected ShouldExpandInstallPanel result (set by leaf).
	WantExpand bool
}

// Response holds HTTP probe fields and/or pure helper result.
type Response struct {
	// HTTP probe
	StatusCode  int
	ContentType string
	Body        []byte
	BodyString  string

	RealSessionID string
	BaseURL       string
	ProbeURL      string
	RunExitCode   int
	RunErrText    string

	// Pure helper
	ExpandResult bool
	// ExpandCalled is true when ModeShouldExpand ran the helper.
	ExpandCalled bool
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	switch req.Mode {
	case ModeGoHTML:
		return runGoHTML(t, req)
	case ModeShouldExpand:
		return runShouldExpand(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runShouldExpand(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	// Pure: browsertrace.ShouldExpandInstallPanel(connected, supports).
	got := browsertrace.ShouldExpandInstallPanel(req.Connected, req.Supports)
	return &Response{
		ExpandResult: got,
		ExpandCalled: true,
	}, nil
}

func runGoHTML(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set by Setup")
	}
	if req.SessionSuffix == "" {
		t.Fatal("SessionSuffix must be set by Setup (known session id)")
	}
	req.NoOpenChrome = true
	if req.ReadyTimeout <= 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	if req.CompleteTimeout <= 0 {
		req.CompleteTimeout = 2 * time.Second
	}

	addr := req.Addr
	if addr == "" {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return nil, err
		}
		addr = ln.Addr().String()
		_ = ln.Close()
		req.Addr = addr
	}
	baseURL := "http://" + addr
	realSessionID := req.SessionSuffix

	runCtx, runCancel := context.WithCancel(context.Background())
	defer runCancel()

	var stdout, stderr bytes.Buffer
	cfg := browsertrace.Config{
		Addr:            addr,
		BaseDir:         req.BaseDir,
		ReadyTimeout:    req.ReadyTimeout,
		CompleteTimeout: req.CompleteTimeout,
		NoOpenChrome:    true,
		SessionSuffix:   req.SessionSuffix,
		Stdout:          &stdout,
		Stderr:          &stderr,
	}

	type runOutcome struct {
		result *browsertrace.Result
		err    error
	}
	runDone := make(chan runOutcome, 1)
	go func() {
		res, err := browsertrace.Run(runCtx, cfg)
		runDone <- runOutcome{result: res, err: err}
	}()

	if err := waitHealth(baseURL, req.ReadyTimeout); err != nil {
		runCancel()
		<-runDone
		return nil, fmt.Errorf("control server never became healthy at %s: %w", baseURL, err)
	}

	if req.DoHello {
		if err := postHello(baseURL, req.HelloVersion, req.HelloFeatures); err != nil {
			runCancel()
			<-runDone
			return nil, fmt.Errorf("POST /v1/hello: %w", err)
		}
		time.Sleep(30 * time.Millisecond)
	}

	probeURL := baseURL + "/go?session=" + realSessionID
	httpResp, body, err := doGET(probeURL)
	runCancel()
	out := <-runDone

	resp := &Response{
		RealSessionID: realSessionID,
		BaseURL:       baseURL,
		ProbeURL:      probeURL,
		RunExitCode:   -1,
	}
	if out.result != nil {
		resp.RunExitCode = out.result.ExitCode
	} else if out.err != nil {
		resp.RunExitCode = 1
	}
	if out.err != nil {
		resp.RunErrText = out.err.Error()
	}
	if err != nil {
		return resp, err
	}
	resp.StatusCode = httpResp.StatusCode
	resp.ContentType = httpResp.Header.Get("Content-Type")
	resp.Body = body
	resp.BodyString = string(body)
	return resp, nil
}

func waitHealth(baseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var last error
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/v1/health", nil)
		if err != nil {
			cancel()
			return err
		}
		res, err := http.DefaultClient.Do(req)
		if err == nil {
			io.Copy(io.Discard, res.Body)
			res.Body.Close()
			cancel()
			if res.StatusCode == http.StatusOK {
				return nil
			}
			last = fmt.Errorf("health status %d", res.StatusCode)
		} else {
			last = err
			cancel()
		}
		time.Sleep(20 * time.Millisecond)
	}
	if last == nil {
		last = fmt.Errorf("timeout waiting for health")
	}
	return last
}

func postHello(baseURL, version string, features []string) error {
	payload := map[string]any{
		"version":  version,
		"features": features,
	}
	if features == nil {
		payload["features"] = []string{}
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	res, err := http.Post(baseURL+"/v1/hello", "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	io.Copy(io.Discard, res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("hello status %d", res.StatusCode)
	}
	return nil
}

func doGET(url string) (*http.Response, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return res, nil, err
	}
	return res, body, nil
}

// Silence unused import guard for strings in some generations.
var _ = strings.TrimSpace
```
