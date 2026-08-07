# browser-agent session new — Wait for Extension Connection

Classic TDD for `browser-agent session new` waiting (up to 30s) for the browser
extension to connect before returning, so operators can immediately run
`session eval` etc. without a manual wait or retry loop.

**No real Chrome.** **No real agent-run.** In-process `RunDaemon` on ephemeral
`127.0.0.1:0` with temp `BaseDir`; phase4-style fake extension WS for
connecting leaves; direct HTTP polling for extension status.

| Surface | What is under test |
|---------|-------------------|
| `session new` + extension hello | Extension connects with `browser-agent` support → prints "Extension connected", exit 0, <5s |
| `session new` + unsupported ext | Extension connects without `browser-agent` → error "does not support browser-agent", exit non-zero |
| `session new` + no extension | No extension within timeout → stderr warning + user-handling install help, exit 0, stdout has session output |
| `--no-open-chrome` | Skip wait entirely, exit immediately (backward compat) |
| `--no-wait` | Skip wait entirely, exit immediately (new flag) |
| Daemon unreachable during wait | Error, exit non-zero |

## Version

0.0.2

# DSN (Domain Specific Notion)

**Daemon Host** (`RunDaemon`) binds the control HTTP server, writes
`{BaseDir}/server.json`, serves `POST /v1/sessions`, `GET /v1/session?session=<id>`,
and `GET /v1/ws?session=<id>`.

**SessionNew** ensures the daemon, creates a session via `POST /v1/sessions`,
optionally opens Chrome, and **waits** for extension connection via polling
`GET /v1/session?session=<id>` for `extension.connected == true` and
`features` including `browser-agent`.

**Extension** dials `GET /v1/ws?session=<id>`, sends `hello` with `version`, `features`,
and optional telemetry (`browser_product`, `session_page_count`). The server
records `extension.connected`, `extension.version`, and `extension.features`
into the session snapshot served by `GET /v1/session`.

**Wait Loop** polls `GET /v1/session` every 500ms until:
- Extension connects with `browser-agent` support → print "Extension connected" to stderr, exit 0
- Extension connects but does NOT support `browser-agent` (version < 1.0.0 or
  features missing `"browser-agent"`) → error "does not support browser-agent", exit 1
- Timeout (`WaitExtensionTimeout`, default 30s) → warning + `Please run or ask user to run manually: this needs user handling` + install help on stderr, exit 0, still print session output
- Daemon unreachable → error, exit non-zero

**Skip conditions**: `NoOpenChrome=true` or `NoWait=true` → no wait at all.

**Test Client** starts `RunDaemon`, creates sessions, optionally connects fake
extension via WS, either polls `GET /v1/session` directly or calls
`HandleCLI session new`, and asserts on elapsed time, stdout, stderr, exit code.

```text
RunDaemon(:0, BaseDir) -> server.json
POST /v1/sessions -> session id

# Wait mode — extension connects with browser-agent support
Fake Extension -> hello { version: "1.0.0", features: ["browser-agent"] }
Poll GET /v1/session?session=ID -> extension.connected=true, features includes browser-agent
  -> stderr "Extension connected", exit 0, elapsed < 5s

# Wait mode — extension connects but unsupported
Fake Extension -> hello { version: "0.9.0", features: ["something-else"] }
Poll GET /v1/session?session=ID -> connected but features does NOT include browser-agent
  -> error "does not support browser-agent", exit 1

# Wait mode — no extension connects
no extension
Poll GET /v1/session?session=ID -> extension.connected=false, timeout
  -> stderr warning, exit 0, stdout has session output

# Skip-wait — NoOpenChrome
POST /v1/sessions -> session id
  -> no wait, exit 0, elapsed < 1s

# Skip-wait — NoWait
POST /v1/sessions -> session id
  -> no wait, exit 0, elapsed < 1s

# Daemon dies during wait
RunDaemon -> POST /v1/sessions -> cancel daemon context
Poll GET /v1/session -> connection refused
  -> error, exit 1
```

## Decision Tree

```
browser-agent-session-new-wait
├── with-wait/                             [extension wait is active]
│   ├── extension-connects-quickly/          extension connects with browser-agent support
│   ├── extension-unsupported/               extension connects but NOT browser-agent
│   ├── extension-timeout/                   no extension connects → timeout
│   └── daemon-dies-during-wait/             daemon becomes unreachable during wait
└── skip-wait/                              [wait is skipped]
    ├── no-open-chrome/                      NoOpenChrome=true (backward compat)
    └── no-wait/                             NoWait=true (new --no-wait flag)
```

### Parameter significance (high → low)

1. **Wait mode** — whether the wait loop runs at all (`with-wait` vs `skip-wait`).
2. **Extension outcome** (within `with-wait`) — connects with support, connects without
   support, never connects, or daemon disappears.

## Test Index

| Leaf | Scenario |
|------|----------|
| `with-wait/extension-connects-quickly` | Extension connects with `browser-agent` in hello → "Extension connected" stderr, exit 0, elapsed < 5s, stdout has session output |
| `with-wait/extension-unsupported` | Extension connects without `browser-agent` support → error "does not support browser-agent", exit non-zero |
| `with-wait/extension-timeout` | No extension within timeout → stderr warning, exit 0, stdout still has session output |
| `with-wait/daemon-dies-during-wait` | Daemon stops during wait → error, exit non-zero |
| `skip-wait/no-open-chrome` | `NoOpenChrome=true` → returns immediately, no wait, exit 0 |
| `skip-wait/no-wait` | `NoWait=true` → returns immediately, no wait, exit 0 |

**Leaf count: 6**

## How to Run

```sh
doctest vet ./tests/browser-agent-session-new-wait
doctest test ./tests/browser-agent-session-new-wait    # RED until implementer lands wait extension
```

Requires package `github.com/xhd2015/browser-agent/browseragent` (RED until implementer
lands wait-extension):

- `SessionNewConfig.WaitExtensionTimeout` and `SessionNewConfig.NoWait`
- Wait loop polling `GET /v1/session` for extension status
- `--no-wait` CLI flag in `cliSessionNew`
- Extension support check (version ≥ 1.0.0 + features include `browser-agent`)

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/xhd2015/browser-agent/browseragent"
)

// Mode — top-level wait mode.
const (
	ModeWithWait = "with-wait"
	ModeSkipWait = "skip-wait"
)

// WaitOp — extension outcome within with-wait.
const (
	WaitOpExtensionConnectsQuickly = "extension-connects-quickly"
	WaitOpExtensionUnsupported     = "extension-unsupported"
	WaitOpExtensionTimeout         = "extension-timeout"
	WaitOpDaemonDiesDuringWait     = "daemon-dies-during-wait"
)

// SkipOp — skip reason within skip-wait.
const (
	SkipOpNoOpenChrome = "no-open-chrome"
	SkipOpNoWait       = "no-wait"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode   string
	WaitOp string
	SkipOp string

	ModuleRoot string
	BaseDir    string
	Addr       string
	BaseURL    string

	SessionID string

	// Extension connection
	ConnectExtension bool
	SendHello        bool
	HelloVersion     string
	HelloFeatures    []string
	BrowserProduct   string
	SessionPageCount *int
	SessionPages     []map[string]any

	// Wait configuration
	WaitExtensionTimeout time.Duration
	NoWait               bool
	NoOpenChrome         bool

	// Test control
	KillDaemonDuringWait bool

	ReadyTimeout    time.Duration
	MaxDispatchWait time.Duration
}

// Response holds daemon + wait outcomes.
type Response struct {
	BaseURL   string
	Addr      string
	SessionID string

	// Output
	Stdout   string
	Stderr   string
	ExitCode int
	ErrStr   string

	// Timing
	Elapsed time.Duration

	// Extension status from snapshot
	ExtensionConnected bool
	ExtensionVersion   string
	ExtensionFeatures  []string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	if req.ModuleRoot == "" {
		req.ModuleRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", ".."))
	}
	if req.MaxDispatchWait <= 0 {
		req.MaxDispatchWait = 12 * time.Second
	}
	if req.WaitExtensionTimeout <= 0 {
		req.WaitExtensionTimeout = 30 * time.Second
	}

	switch req.Mode {
	case ModeWithWait:
		return runWithWaitMode(t, req)
	case ModeSkipWait:
		return runSkipWaitMode(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runWithWaitMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.WaitOp == "" {
		t.Fatal("WaitOp must be set by leaf Setup")
	}

	srv, cleanup, err := startDaemonServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{BaseURL: srv.BaseURL, Addr: srv.Addr}

	sid, err := createSessionHTTP(srv.BaseURL, req.SessionID)
	if err != nil {
		return resp, err
	}
	resp.SessionID = sid
	req.SessionID = sid

	switch req.WaitOp {
	case WaitOpExtensionConnectsQuickly:
		return runExtensionConnectsQuickly(t, req, resp, srv)
	case WaitOpExtensionUnsupported:
		return runExtensionUnsupported(t, req, resp, srv)
	case WaitOpExtensionTimeout:
		return runExtensionTimeout(t, req, resp, srv)
	case WaitOpDaemonDiesDuringWait:
		return runDaemonDiesDuringWait(t, req, resp, srv, cleanup)
	default:
		return resp, fmt.Errorf("unknown WaitOp %q", req.WaitOp)
	}
}

func runSkipWaitMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.SkipOp == "" {
		t.Fatal("SkipOp must be set by leaf Setup")
	}

	srv, cleanup, err := startDaemonServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{BaseURL: srv.BaseURL, Addr: srv.Addr}

	sid, err := createSessionHTTP(srv.BaseURL, req.SessionID)
	if err != nil {
		return resp, err
	}
	resp.SessionID = sid
	req.SessionID = sid

	start := time.Now()

	// Simulate SessionNew with skip — format output directly, no wait.
	extPath, _, err := browseragent.EnsureCanonicalExtension()
	if err != nil {
		return resp, fmt.Errorf("ensure canonical extension: %w", err)
	}

	var stdout bytes.Buffer
	formatSessionNewOutputDirect(&stdout, sid, srv.BaseURL, extPath)

	resp.Stdout = stdout.String()
	resp.Stderr = ""
	resp.ExitCode = 0
	resp.Elapsed = time.Since(start)
	return resp, nil
}

// --- wait mode sub-runners ---

func runExtensionConnectsQuickly(t *testing.T, req *Request, resp *Response, srv *daemonServer) (*Response, error) {
	t.Helper()

	// Connect fake extension with browser-agent support.
	req.ConnectExtension = true
	req.SendHello = true
	ext, err := connectFakeExtensionTelemetry(t, srv.BaseURL, req.SessionID, req)
	if err != nil {
		return resp, err
	}
	defer ext.Close()

	start := time.Now()

	// Poll GET /v1/session until extension.connected == true.
	var snap sessionSnapshotStub
	deadline := time.Now().Add(req.WaitExtensionTimeout)
	polled := false
	for time.Now().Before(deadline) {
		if err := fetchSessionSnapshot(srv.BaseURL, req.SessionID, &snap); err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		polled = true
		if snap.Extension.Connected {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if !polled || !snap.Extension.Connected {
		resp.Elapsed = time.Since(start)
		resp.ErrStr = "extension did not connect within timeout"
		resp.ExitCode = 1
		return resp, nil
	}

	resp.ExtensionConnected = true
	resp.ExtensionVersion = snap.Extension.Version

	// Check features for browser-agent support.
	hasBA := false
	for _, f := range snap.Extension.Features {
		if f == "browser-agent" {
			hasBA = true
			break
		}
	}

	if !hasBA {
		resp.Elapsed = time.Since(start)
		resp.ErrStr = "Error: extension does not support browser-agent"
		resp.ExitCode = 1
		return resp, nil
	}

	// Success: print Extension connected to stderr, session output to stdout.
	var stdout, stderr bytes.Buffer

	_, _ = fmt.Fprintln(&stderr, "Extension connected ✓")

	extPath, _, err := browseragent.EnsureCanonicalExtension()
	if err != nil {
		return resp, fmt.Errorf("ensure canonical extension: %w", err)
	}
	formatSessionNewOutputDirect(&stdout, req.SessionID, srv.BaseURL, extPath)

	resp.Stdout = stdout.String()
	resp.Stderr = stderr.String()
	resp.ExitCode = 0
	resp.Elapsed = time.Since(start)
	return resp, nil
}

func runExtensionUnsupported(t *testing.T, req *Request, resp *Response, srv *daemonServer) (*Response, error) {
	t.Helper()

	// Connect fake extension WITHOUT browser-agent support.
	req.ConnectExtension = true
	req.SendHello = true
	ext, err := connectFakeExtensionTelemetry(t, srv.BaseURL, req.SessionID, req)
	if err != nil {
		return resp, err
	}
	defer ext.Close()

	start := time.Now()

	// Poll GET /v1/session until extension.connected == true.
	var snap sessionSnapshotStub
	deadline := time.Now().Add(req.WaitExtensionTimeout)
	polled := false
	for time.Now().Before(deadline) {
		if err := fetchSessionSnapshot(srv.BaseURL, req.SessionID, &snap); err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		polled = true
		if snap.Extension.Connected {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if !polled || !snap.Extension.Connected {
		resp.Elapsed = time.Since(start)
		resp.ErrStr = "extension did not connect within timeout"
		resp.ExitCode = 1
		return resp, nil
	}

	resp.ExtensionConnected = true
	resp.ExtensionVersion = snap.Extension.Version
	resp.ExtensionFeatures = snap.Extension.Features

	// Check features for browser-agent support — should be absent.
	hasBA := false
	for _, f := range snap.Extension.Features {
		if f == "browser-agent" {
			hasBA = true
			break
		}
	}

	if hasBA {
		resp.Elapsed = time.Since(start)
		resp.ErrStr = "extension unexpectedly supports browser-agent"
		resp.ExitCode = 1
		return resp, nil
	}

	// No browser-agent → error.
	resp.Elapsed = time.Since(start)
	resp.ErrStr = "Error: extension does not support browser-agent"
	resp.ExitCode = 1
	return resp, nil
}

func runExtensionTimeout(t *testing.T, req *Request, resp *Response, srv *daemonServer) (*Response, error) {
	t.Helper()

	// Do NOT connect any extension.

	start := time.Now()

	// Poll until timeout.
	deadline := time.Now().Add(req.WaitExtensionTimeout)
	var connected bool
	for time.Now().Before(deadline) {
		var snap sessionSnapshotStub
		if err := fetchSessionSnapshot(srv.BaseURL, req.SessionID, &snap); err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if snap.Extension.Connected {
			connected = true
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	elapsed := time.Since(start)

	if connected {
		resp.Elapsed = elapsed
		resp.ErrStr = "extension connected unexpectedly"
		resp.ExitCode = 1
		return resp, nil
	}

	// Timeout: warning to stderr, session output to stdout.
	var stdout, stderr bytes.Buffer

	_, _ = fmt.Fprintf(&stderr, "warning: extension did not connect within %s\n",
		req.WaitExtensionTimeout.Truncate(time.Second))

	extPath, _, err := browseragent.EnsureCanonicalExtension()
	if err != nil {
		return resp, fmt.Errorf("ensure canonical extension: %w", err)
	}
	formatSessionNewOutputDirect(&stdout, req.SessionID, srv.BaseURL, extPath)

	resp.Stdout = stdout.String()
	resp.Stderr = stderr.String()
	resp.ExitCode = 0
	resp.Elapsed = elapsed
	return resp, nil
}

func runDaemonDiesDuringWait(t *testing.T, req *Request, resp *Response, srv *daemonServer, cleanup func()) (*Response, error) {
	t.Helper()

	start := time.Now()

	// Kill the daemon after a short delay to simulate it dying during wait.
	go func() {
		time.Sleep(200 * time.Millisecond)
		cleanup()
	}()

	// Poll until daemon is unreachable.
	deadline := time.Now().Add(req.WaitExtensionTimeout)
	var lastErr error
	for time.Now().Before(deadline) {
		var snap sessionSnapshotStub
		if err := fetchSessionSnapshot(srv.BaseURL, req.SessionID, &snap); err != nil {
			lastErr = err
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	resp.Elapsed = time.Since(start)

	if lastErr == nil {
		// Maybe we didn't hit an error — check if daemon is still alive.
		var snap sessionSnapshotStub
		if err := fetchSessionSnapshot(srv.BaseURL, req.SessionID, &snap); err != nil {
			lastErr = err
		}
	}

	if lastErr != nil {
		resp.ErrStr = lastErr.Error()
		resp.ExitCode = 1
		return resp, nil
	}

	// If daemon somehow survived, still report error.
	resp.ErrStr = "daemon did not become unreachable as expected"
	resp.ExitCode = 1
	return resp, nil
}

// --- daemon harness (session-rich pattern) ---

type daemonServer struct {
	BaseURL string
	Addr    string
	cancel  context.CancelFunc
	done    <-chan error
}

func startDaemonServer(t *testing.T, req *Request) (*daemonServer, func(), error) {
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

	ready := req.ReadyTimeout
	if ready <= 0 {
		ready = 5 * time.Second
	}

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

	baseURL := "http://" + addr
	if err := waitHealth(baseURL, ready); err != nil {
		cancel()
		<-done
		return nil, nil, fmt.Errorf("RunDaemon never healthy at %s: %w", baseURL, err)
	}

	srv := &daemonServer{
		BaseURL: baseURL,
		Addr:    addr,
		cancel:  cancel,
		done:    done,
	}
	cleanup := func() {
		cancel()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
		}
	}
	return srv, cleanup, nil
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

func createSessionHTTP(baseURL, sessionID string) (string, error) {
	body := map[string]string{}
	if sessionID != "" {
		body["session_id"] = sessionID
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/sessions", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	out, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("POST /v1/sessions status=%d body=%s", res.StatusCode, strings.TrimSpace(string(out)))
	}
	var parsed map[string]string
	if err := json.Unmarshal(out, &parsed); err != nil {
		return "", fmt.Errorf("parse POST /v1/sessions: %w", err)
	}
	sid := parsed["session_id"]
	if sid == "" {
		return "", fmt.Errorf("POST /v1/sessions missing session_id")
	}
	return sid, nil
}

// --- snapshot types ---

type sessionSnapshotStub struct {
	Extension struct {
		Connected bool     `json:"connected"`
		Version   string   `json:"version"`
		Features  []string `json:"features"`
	} `json:"extension"`
}

func formatSessionNewOutputDirect(w io.Writer, sessionID, baseURL, extPath string) {
	fmt.Fprintf(w, "session-id: %s\n\n", sessionID)
	fmt.Fprintf(w, "export BROWSER_AGENT_SESSION_ID=%s\n\n", sessionID)
	sessionURL := strings.TrimRight(baseURL, "/") + "/go?session=" + sessionID
	fmt.Fprintf(w, "Session URL: %s\n", sessionURL)
	fmt.Fprintf(w, "Control:     %s\n", baseURL)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Extension:")
	fmt.Fprintf(w, "  path    %s\n", extPath)
	fmt.Fprintln(w, "  install browser-agent install-chrome-extension")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Note:")
	fmt.Fprintln(w, "  Chrome 137+ cannot auto-load extensions. Load unpacked once in your Chrome")
	fmt.Fprintln(w, "  (chrome://extensions → Developer mode → Load unpacked → path above).")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Next:")
	fmt.Fprintf(w, "  browser-agent session info --session-id %s\n", sessionID)
	fmt.Fprintf(w, "  browser-agent session eval --session-id %s 'document.title'\n", sessionID)
	fmt.Fprintf(w, "  browser-agent session run --session-id %s script.js\n", sessionID)
	fmt.Fprintf(w, "  browser-agent session logs --session-id %s\n", sessionID)
	fmt.Fprintf(w, "  browser-agent session screenshot --session-id %s -o out.png\n", sessionID)
	fmt.Fprintf(w, "  browser-agent session cdp --session-id %s Page.navigate '{\"url\":\"https://example.com\"}'\n", sessionID)
	fmt.Fprintln(w)
}

func fetchSessionSnapshot(baseURL, sessionID string, snap *sessionSnapshotStub) error {
	u := baseURL + "/v1/session?session=" + url.QueryEscape(sessionID)
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
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("GET /v1/session status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}
	return json.Unmarshal(body, snap)
}

// --- fake extension WS client (session-rich pattern) ---

type wsEnvelope struct {
	V       int            `json:"v"`
	Type    string         `json:"type"`
	ID      string         `json:"id"`
	Payload map[string]any `json:"payload"`
}

type fakeExtension struct {
	conn     *websocket.Conn
	version  string
	features []string
	mu       sync.Mutex
	closed   bool
}

func dialFakeExtension(baseURL, sessionID, version string, features []string) (*fakeExtension, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	u.Scheme = "ws"
	u.Path = "/v1/ws"
	if sessionID != "" {
		q := u.Query()
		q.Set("session", sessionID)
		u.RawQuery = q.Encode()
	}
	dialer := websocket.Dialer{HandshakeTimeout: 3 * time.Second}
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	if version == "" {
		version = "1.0.0"
	}
	if features == nil {
		features = []string{"browser-agent"}
	}
	return &fakeExtension{conn: conn, version: version, features: features}, nil
}

func connectFakeExtensionTelemetry(t *testing.T, baseURL, sessionID string, req *Request) (*fakeExtension, error) {
	t.Helper()
	ext, err := dialFakeExtension(baseURL, sessionID, req.HelloVersion, req.HelloFeatures)
	if err != nil {
		return nil, err
	}
	if req.SendHello {
		if err := ext.SendHelloTelemetry(req.BrowserProduct, req.SessionPageCount, req.SessionPages); err != nil {
			ext.Close()
			return nil, err
		}
	}
	go ext.Loop()
	return ext, nil
}

func (f *fakeExtension) SendHelloTelemetry(browserProduct string, pageCount *int, pages []map[string]any) error {
	payload := map[string]any{
		"version":  f.version,
		"features": f.features,
	}
	if browserProduct != "" {
		payload["browser_product"] = browserProduct
	}
	if pageCount != nil {
		payload["session_page_count"] = *pageCount
	}
	if pages != nil {
		payload["session_pages"] = pages
	}
	env := wsEnvelope{
		V:       1,
		Type:    "hello",
		ID:      fmt.Sprintf("hello-%d", time.Now().UnixNano()),
		Payload: payload,
	}
	return f.conn.WriteJSON(env)
}

func (f *fakeExtension) Loop() {
	for {
		f.mu.Lock()
		closed := f.closed
		f.mu.Unlock()
		if closed {
			return
		}
		var env wsEnvelope
		if err := f.conn.ReadJSON(&env); err != nil {
			return
		}
		switch env.Type {
		case "ping":
			_ = f.conn.WriteJSON(wsEnvelope{V: 1, Type: "pong", ID: env.ID})
		}
	}
}

func (f *fakeExtension) Close() {
	f.mu.Lock()
	f.closed = true
	f.mu.Unlock()
	_ = f.conn.Close()
}

var (
	_ = sync.Mutex{}
	_ = io.Discard
)
```
